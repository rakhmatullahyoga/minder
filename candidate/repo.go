package candidate

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"minder/auth"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

const (
	cachePrefix = "candidate"

	getUserByIdQuery             = "SELECT id, name, email, verified FROM users WHERE id = ?"
	getInterestedCandidatesQuery = "SELECT u.id, u.name, u.email, u.verified FROM users u JOIN user_interests ui ON ui.liked_user_id = u.id WHERE ui.user_id = ?"
	insertUserInterestQuery      = "INSERT INTO user_interests (user_id, liked_user_id) VALUES (?, ?)"
	updateUserQuery              = "UPDATE users SET name = ?, email = ?, verified = ? WHERE id = ?"
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

func getLastCachedCandidate(ctx context.Context) (candidateID uint64, err error) {
	userID := ctx.Value(auth.ClaimsKeyUserID)
	candidateIdStr, err := cache.LIndex(ctx, fmt.Sprintf("%s:%v", cachePrefix, userID), -1).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			err = ErrNoCachedCandidate
		}
		return
	}

	candidateID, err = strconv.ParseUint(candidateIdStr, 10, 64)
	if err != nil {
		err = ErrNoCachedCandidate
		return
	}
	return
}

func getUserByID(ctx context.Context, userID uint64) (user User, err error) {
	err = db.GetContext(ctx, &user, getUserByIdQuery, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrNoCandidateAvailable
		}
	}

	return
}

func getUserInterestedCandidates(ctx context.Context) (candidates []User, err error) {
	userID := ctx.Value(auth.ClaimsKeyUserID).(float64)
	err = db.SelectContext(ctx, &candidates, getInterestedCandidatesQuery, uint64(userID))
	if err == sql.ErrNoRows {
		err = nil
	}
	return
}

func updateUser(ctx context.Context, user User) (err error) {
	_, err = db.ExecContext(ctx, updateUserQuery, user.Name, user.Email, user.Verified, user.ID)
	return
}

func getCandidate(ctx context.Context, excludedIDs []uint64) (user User, err error) {
	err = db.GetContext(ctx, &user, getCandidateQuery(excludedIDs))
	if err != nil {
		if err == sql.ErrNoRows {
			err = ErrNoCandidateAvailable
		}
	}

	return
}

func getCandidateQuery(excludedIDs []uint64) (query string) {
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

func getCachedCandidateIDs(ctx context.Context) (ids []uint64, err error) {
	userID := ctx.Value(auth.ClaimsKeyUserID)
	idsStr, err := cache.LRange(ctx, fmt.Sprintf("%s:%v", cachePrefix, userID), 0, -1).Result()
	if err != nil {
		return
	}

	for _, v := range idsStr {
		id, _ := strconv.ParseUint(v, 10, 64)
		ids = append(ids, id)
	}
	return
}

func insertUserInterest(ctx context.Context, candidateID uint64) (err error) {
	userID := ctx.Value(auth.ClaimsKeyUserID).(float64)
	_, err = db.ExecContext(ctx, insertUserInterestQuery, uint64(userID), candidateID)
	return
}

func cacheCandidateID(ctx context.Context, candidateID uint64) (err error) {
	userID := ctx.Value(auth.ClaimsKeyUserID)
	key := fmt.Sprintf("%s:%v", cachePrefix, userID)
	_, err = cache.RPush(ctx, key, candidateID).Result()
	if err != nil {
		return
	}

	ttl, _ := cache.TTL(ctx, key).Result()
	if ttl < 0 {
		year, month, day := time.Now().AddDate(0, 0, 1).Date()
		tomorrow := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
		_, _ = cache.ExpireAt(ctx, key, tomorrow).Result()
	}
	return
}
