package models

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const InviteExpirationTime = time.Hour * 12

type InvitesInterface interface {
	GetInvite(context.Context, *Invite) error
	Delete(context.Context, uuid.UUID) error
}

type Invite struct {
	UserID      uuid.UUID `json:"user_id"`
	InviteToken uuid.UUID `json:"invite_token"`
	ExpiresAt   time.Time `json:"expires_at"`
	SentCount   int       `json:"sent_count"`
	CreatedAt   time.Time `json:"created_at"`
	LastSeenAt  time.Time `json:"last_seen_at"`
}

type InvitesModel struct {
	pool *pgxpool.Pool
}

func (i *InvitesModel) Delete(ctx context.Context, userID uuid.UUID) error {
	statement := `
		DELETE FROM user_invites
		WHERE user_id = $1
	`
	ctx, cancel := context.WithTimeout(ctx, maxQueryDuration)
	defer cancel()

	_, err := i.pool.Exec(ctx, statement, userID)
	return err
}

func (i *InvitesModel) GetInvite(ctx context.Context, invite *Invite) error {
	statement := `
		SELECT user_id, invite_token, expires_at, sent_count, created_at, last_seen_at
		FROM user_invites ui
		WHERE ui.invite_token = $1
		`
	ctx, cancel := context.WithTimeout(ctx, maxQueryDuration)
	defer cancel()

	return i.pool.QueryRow(ctx, statement, invite.InviteToken).Scan(
		&invite.UserID,
		&invite.InviteToken,
		&invite.ExpiresAt,
		&invite.SentCount,
		&invite.CreatedAt,
		&invite.LastSeenAt,
	)
}

func createInvite(ctx context.Context, tx pgx.Tx, invite *Invite) error {
	statement := `
			INSERT INTO user_invites(user_id, invite_token, expires_at)
			VALUES ($1, $2, $3)
			RETURNING sent_count, created_at, last_seen_at
		`
	ctx, cancel := context.WithTimeout(ctx, maxQueryDuration)
	defer cancel()

	return tx.QueryRow(
		ctx,
		statement,
		invite.UserID,
		invite.InviteToken,
		invite.ExpiresAt,
	).Scan(
		&invite.SentCount,
		&invite.CreatedAt,
		&invite.LastSeenAt,
	)
}
