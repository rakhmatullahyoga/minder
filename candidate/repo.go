package candidate

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

const (
	cachePrefix = "candidate:"
)

var (
	db    *sqlx.DB
	cache *redis.Client
)

func SetDB(newdb *sqlx.DB) (err error) {
	if newdb == nil {
		err = ErrEmptyDB
		return
	}

	db = newdb
	return
}

func SetCache(newCache *redis.Client) (err error) {
	if newCache == nil {
		err = ErrEmptyCache
		return
	}

	cache = newCache
	return
}

func getCandidate(ctx context.Context, excludedIDs []int) (user User, err error) {
	err = db.GetContext(ctx, &user, getCandidateQuery(excludedIDs))
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrNoCandidateAvailable
		}
	}

	return
}

func getCandidateQuery(excludedIDs []int) (query string) {
	var strIDs []string
	query = "SELECT id, name, email, verified FROM users"
	for _, v := range excludedIDs {
		strIDs = append(strIDs, fmt.Sprintf("%d", v))
	}

	if len(excludedIDs) > 0 {
		query += fmt.Sprintf(" WHERE id NOT IN (%s)", strings.Join(strIDs, ","))
	}
	query += " ORDER BY created_at DESC LIMIT 1"
	return
}

func getCachedCandidateIDs(ctx context.Context) (ids []int, err error) {
	userID := ctx.Value("user_id").(string)
	idsStr, err := cache.LRange(ctx, cachePrefix+userID, 0, -1).Result()
	if err != nil {
		return
	}
	for _, v := range idsStr {
		id, _ := strconv.Atoi(v)
		ids = append(ids, id)
	}
	return
}
