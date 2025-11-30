package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"social/internal/env"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

/*
Seeder for the social project database.

Usage examples:
  # Uses DATABASE_URL from env (.env is loaded automatically) and seeds some data
  go run ./cmd/seed \
    -users 25 \
    -minPosts 1 -maxPosts 4 \
    -minComments 0 -maxComments 6 \
    -maxFollows 8 \
    -seed 42

Flags:
  -users        Number of users to create
  -minPosts     Minimum posts per user
  -maxPosts     Maximum posts per user (inclusive)
  -minComments  Minimum comments per post
  -maxComments  Maximum comments per post (inclusive)
  -maxFollows   Maximum number of followings each user will have (0..max)
  -seed         Random seed (int64). If -1, a time-based seed is used
  -clear        If set, clears tables before inserting (DANGEROUS: deletes all data)
  -batch        Batch size for prepared statement inserts (default 1000)

Environment:
  DATABASE_URL  Postgres DSN (e.g. postgres://user:pass@localhost:5432/social?sslmode=disable)
*/

func init() { _ = godotenv.Load() }

func main() {
	var (
		userCount   = flag.Int("users", 20, "number of users to create")
		minPosts    = flag.Int("minPosts", 1, "minimum posts per user")
		maxPosts    = flag.Int("maxPosts", 3, "maximum posts per user")
		minComments = flag.Int("minComments", 0, "minimum comments per post")
		maxComments = flag.Int("maxComments", 5, "maximum comments per post")
		maxFollows  = flag.Int("maxFollows", 5, "maximum followings per user")
		seedVal     = flag.Int64("seed", -1, "random seed; -1 for time-based")
		clear       = flag.Bool("clear", false, "truncate tables before seeding (destructive)")
		batchSize   = flag.Int("batch", 1000, "batch size for prepared statement inserts")
	)
	flag.Parse()

	if *userCount <= 0 {
		log.Fatal("users must be > 0")
	}
	if *maxPosts < *minPosts {
		log.Fatal("maxPosts must be >= minPosts")
	}
	if *maxComments < *minComments {
		log.Fatal("maxComments must be >= minComments")
	}
	if *maxFollows < 0 {
		log.Fatal("maxFollows must be >= 0")
	}

	// RNG
	if *seedVal == -1 {
		*seedVal = time.Now().UnixNano()
	}
	r := rand.New(rand.NewSource(*seedVal))
	log.Printf("seeding with seed=%d", *seedVal)

	// DB connect
	dsn := env.GetString("DATABASE_URL_SEED", "")
	if dsn == "" {
		log.Fatal("DATABASE_URL_SEED is not set")
	}
	pool, err := openDB(dsn)
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()

	if *clear {
		if err := clearAll(ctx, pool); err != nil {
			log.Fatalf("failed to clear tables: %v", err)
		}
		log.Println("cleared tables users, posts, comments, followers")
	}

	// Use a single connection and transaction + prepared statements for speed
	conn, err := pool.Acquire(ctx)
	if err != nil {
		log.Fatalf("acquire conn: %v", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		log.Fatalf("begin tx: %v", err)
	}
	// If we exit with error before commit, rollback
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// Generate users
	users := make([]uuid.UUID, 0, *userCount)
	usernames := uniqueStrings(r, *userCount, func() string { return username(r) })
	emails := make([]string, *userCount)
	for i := 0; i < *userCount; i++ {
		emails[i] = fmt.Sprintf("%s@example.com", usernames[i])
	}
	log.Printf("creating %d users...", *userCount)
	{
		batch := &pgx.Batch{}
		queued := 0
		flush := func() {
			if queued == 0 {
				return
			}
			br := tx.SendBatch(ctx, batch)
			if err := br.Close(); err != nil {
				log.Fatalf("batch users: %v", err)
			}
			batch = &pgx.Batch{}
			queued = 0
		}
		for i := 0; i < *userCount; i++ {
			id := uuid.New()
			pwd := []byte("password") // same password for simplicity
			batch.Queue("INSERT INTO users (id, username, email, password) VALUES ($1,$2,$3,$4)", id, usernames[i], emails[i], pwd)
			queued++
			users = append(users, id)
			if queued >= *batchSize {
				flush()
			}
		}
		flush()
	}

	// Generate posts per user
	tagPool := []string{"golang", "programming", "dev", "life", "music", "travel", "food", "movies", "books", "tech", "tutorial"}
	postIDs := make([]uuid.UUID, 0, *userCount*(*maxPosts))
	log.Println("creating posts...")
	{
		batch := &pgx.Batch{}
		queued := 0
		flush := func() {
			if queued == 0 {
				return
			}
			br := tx.SendBatch(ctx, batch)
			if err := br.Close(); err != nil {
				log.Fatalf("batch posts: %v", err)
			}
			batch = &pgx.Batch{}
			queued = 0
		}
		for _, uid := range users {
			pc := *minPosts + r.Intn(*maxPosts-*minPosts+1)
			for i := 0; i < pc; i++ {
				pid := uuid.New()
				title := sentence(r, 3, 7)
				content := paragraphs(r, 1+r.Intn(3))
				tags := sampleStrings(r, tagPool, 1+r.Intn(3))
				batch.Queue("INSERT INTO posts (id, title, content, user_id, tags) VALUES ($1,$2,$3,$4,$5)", pid, title, content, uid, tags)
				queued++
				postIDs = append(postIDs, pid)
				if queued >= *batchSize {
					flush()
				}
			}
		}
		flush()
	}

	// Generate comments for each post
	log.Println("creating comments...")
	{
		batch := &pgx.Batch{}
		queued := 0
		flush := func() {
			if queued == 0 {
				return
			}
			br := tx.SendBatch(ctx, batch)
			if err := br.Close(); err != nil {
				log.Fatalf("batch comments: %v", err)
			}
			batch = &pgx.Batch{}
			queued = 0
		}
		for _, pid := range postIDs {
			cc := *minComments + r.Intn(*maxComments-*minComments+1)
			for i := 0; i < cc; i++ {
				cid := uuid.New()
				content := sentence(r, 6, 16)
				commenter := users[r.Intn(len(users))]
				batch.Queue("INSERT INTO comments (id, content, post_id, user_id) VALUES ($1,$2,$3,$4)", cid, content, pid, commenter)
				queued++
				if queued >= *batchSize {
					flush()
				}
			}
		}
		flush()
	}

	// Generate followers relations
	if *maxFollows > 0 && len(users) > 1 {
		log.Println("creating followers...")
		batch := &pgx.Batch{}
		queued := 0
		flush := func() {
			if queued == 0 {
				return
			}
			br := tx.SendBatch(ctx, batch)
			if err := br.Close(); err != nil {
				log.Fatalf("batch followers: %v", err)
			}
			batch = &pgx.Batch{}
			queued = 0
		}
		seen := make(map[string]struct{}, len(users)*(*maxFollows))
		for _, uid := range users {
			followCount := r.Intn(*maxFollows + 1)
			if followCount == 0 {
				continue
			}
			candidates := make([]uuid.UUID, 0, len(users)-1)
			for _, other := range users {
				if other != uid {
					candidates = append(candidates, other)
				}
			}
			r.Shuffle(len(candidates), func(i, j int) { candidates[i], candidates[j] = candidates[j], candidates[i] })
			if followCount > len(candidates) {
				followCount = len(candidates)
			}
			for i := 0; i < followCount; i++ {
				f := candidates[i]
				if uid == f {
					continue
				}
				key := uid.String() + "->" + f.String()
				if _, ok := seen[key]; ok {
					continue
				}
				batch.Queue("INSERT INTO followers (user_id, follower_id) VALUES ($1,$2) ON CONFLICT DO NOTHING", uid, f)
				queued++
				seen[key] = struct{}{}
				if queued >= *batchSize {
					flush()
				}
			}
		}
		flush()
	}

	if err := tx.Commit(ctx); err != nil {
		log.Fatalf("commit: %v", err)
	}

	// Quick summary counts to verify inserts landed
	var uCnt, pCnt, cCnt, fCnt int
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&uCnt); err != nil {
		log.Printf("count users error: %v", err)
	}
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM posts").Scan(&pCnt); err != nil {
		log.Printf("count posts error: %v", err)
	}
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM comments").Scan(&cCnt); err != nil {
		log.Printf("count comments error: %v", err)
	}
	if err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM followers").Scan(&fCnt); err != nil {
		log.Printf("count followers error: %v", err)
	}
	log.Printf("seeding completed successfully: users=%d posts=%d comments=%d followers=%d", uCnt, pCnt, cCnt, fCnt)
}

func openDB(dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 20
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}
	return pool, nil
}

func clearAll(ctx context.Context, pool *pgxpool.Pool) error {
	// order matters because of FK constraints
	stmts := []string{
		"DELETE FROM followers",
		"DELETE FROM comments",
		"DELETE FROM posts",
		"DELETE FROM users",
	}
	for _, s := range stmts {
		if _, err := pool.Exec(ctx, s); err != nil {
			return err
		}
	}
	return nil
}

func insertUser(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, username, email string, password []byte) error {
	stmt := `INSERT INTO users (id, username, email, password) VALUES ($1,$2,$3,$4)`
	_, err := pool.Exec(ctx, stmt, id, username, email, password)
	return err
}

func insertPost(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, title, content string, userID uuid.UUID, tags []string) error {
	stmt := `INSERT INTO posts (id, title, content, user_id, tags) VALUES ($1,$2,$3,$4,$5)`
	_, err := pool.Exec(ctx, stmt, id, title, content, userID, tags)
	return err
}

func insertComment(ctx context.Context, pool *pgxpool.Pool, id uuid.UUID, content string, postID, userID uuid.UUID) error {
	stmt := `INSERT INTO comments (id, content, post_id, user_id) VALUES ($1,$2,$3,$4)`
	_, err := pool.Exec(ctx, stmt, id, content, postID, userID)
	return err
}

func insertFollow(ctx context.Context, pool *pgxpool.Pool, userID, followerID uuid.UUID) error {
	if userID == followerID {
		return nil
	}
	stmt := `INSERT INTO followers (user_id, follower_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`
	_, err := pool.Exec(ctx, stmt, userID, followerID)
	return err
}

// --- simple fake data helpers (no external deps) ---

var words = []string{
	"go", "fast", "social", "network", "build", "code", "deploy", "scale", "learn", "fun",
	"today", "music", "movie", "book", "daily", "life", "travel", "food", "tech", "dev",
}

func username(r *rand.Rand) string {
	adjectives := []string{"blue", "fast", "happy", "smart", "lucky", "quiet", "brave", "kind", "wild", "calm"}
	nouns := []string{"tiger", "eagle", "fox", "panda", "otter", "lion", "koala", "whale", "hawk", "wolf"}
	return fmt.Sprintf("%s_%s%d", adjectives[r.Intn(len(adjectives))], nouns[r.Intn(len(nouns))], 100+r.Intn(900))
}

func sentence(r *rand.Rand, min, max int) string {
	wc := min + r.Intn(max-min+1)
	s := make([]string, wc)
	for i := 0; i < wc; i++ {
		s[i] = words[r.Intn(len(words))]
	}
	out := strings.Join(s, " ")
	return strings.ToUpper(out[:1]) + out[1:] + "."
}

func paragraphs(r *rand.Rand, n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteString(sentence(r, 8, 16))
		b.WriteString("\n\n")
	}
	return b.String()
}

func sampleStrings(r *rand.Rand, src []string, n int) []string {
	if n <= 0 {
		return nil
	}
	if n > len(src) {
		n = len(src)
	}
	out := make([]string, len(src))
	copy(out, src)
	r.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out[:n]
}

func uniqueStrings(r *rand.Rand, n int, gen func() string) []string {
	out := make([]string, 0, n)
	seen := make(map[string]struct{}, n*2)
	for len(out) < n {
		s := gen()
		if _, ok := seen[s]; ok {
			// make a variant to avoid collision
			s = fmt.Sprintf("%s%d", s, r.Intn(10000))
			if _, ok2 := seen[s]; ok2 {
				continue
			}
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
