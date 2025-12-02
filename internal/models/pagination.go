package models

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PaginatedFeedQuery struct {
	Limit  int       `json:"limit" validate:"gte=1,lte=20"`
	Offset int       `json:"offset" validate:"gte=0"`
	Sort   string    `json:"sort" validate:"oneof=ASC DESC"`
	Tags   []string  `json:"tags" validate:"max=5"`
	Search string    `json:"search" validate:"max=100"`
	From   time.Time `json:"from" validate:"lte"`
	To     time.Time `json:"to" validate:""`
}

func (pg PaginatedFeedQuery) Parse(r *http.Request) (PaginatedFeedQuery, error) {
	queryParams := r.URL.Query()

	limitString := queryParams.Get("limit")
	if limitString != "" {
		limit, err := strconv.Atoi(limitString)
		if err != nil {
			return PaginatedFeedQuery{}, err
		}
		pg.Limit = limit
	}

	offsetString := queryParams.Get("offset")
	if offsetString != "" {
		offset, err := strconv.Atoi(offsetString)
		if err != nil {
			return PaginatedFeedQuery{}, err
		}
		pg.Offset = offset
	}

	sort := queryParams.Get("sort")
	if sort != "" {
		pg.Sort = sort
	}

	tags := queryParams.Get("tags")
	if tags != "" {
		pg.Tags = strings.Split(tags, ",")
	}

	search := queryParams.Get("search")
	if search != "" {
		pg.Search = search
	}

	from := queryParams.Get("from")
	if from != "" {
		parsedFrom, err := time.Parse(time.RFC3339, from)
		if err == nil {
			pg.From = parsedFrom
		}
	}

	to := queryParams.Get("to")
	if to != "" {
		parsedTo, err := time.Parse(time.RFC3339, to)
		if err == nil {
			pg.To = parsedTo
		}
	}

	return pg, nil
}
