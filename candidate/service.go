package candidate

import (
	"context"

	"minder/auth"
)

const (
	dailyQuota = 10
)

var (
	getLastCachedCandidateRepo      = getLastCachedCandidate
	getUserByIDRepo                 = getUserByID
	getUserInterestedCandidatesRepo = getUserInterestedCandidates
	getCandidateRepo                = getCandidate
	getCachedCandidateIDsRepo       = getCachedCandidateIDs
	insertUserInterestRepo          = insertUserInterest
	cacheCandidateIDRepo            = cacheCandidateID
	updateUserRepo                  = updateUser
)

func getCandidateFeed(ctx context.Context) (candidate User, err error) {
	candidateID, err := getLastCachedCandidateRepo(ctx)
	if err != nil && err != ErrNoCachedCandidate {
		return
	}

	if candidateID > 0 {
		candidate, err = getUserByIDRepo(ctx, candidateID)
	} else {
		var likedCandidates []User
		var likedIDs []uint64
		likedCandidates, err = getUserInterestedCandidatesRepo(ctx)
		if err != nil {
			return
		}

		for _, c := range likedCandidates {
			likedIDs = append(likedIDs, c.ID)
		}
		userID := ctx.Value(auth.ClaimsKeyUserID).(float64)
		candidate, err = getCandidateRepo(ctx, append(likedIDs, uint64(userID)))
		_ = cacheCandidateIDRepo(ctx, candidate.ID)
	}
	return
}

func swipeCandidate(ctx context.Context, candidateID uint64, liked bool) (nextCandidate User, err error) {
	cachedIDs, err := getCachedCandidateIDsRepo(ctx)
	if err != nil {
		return
	}

	verified := ctx.Value(auth.ClaimsKeyVerified).(bool)
	if !verified && len(cachedIDs) >= dailyQuota && candidateID != cachedIDs[len(cachedIDs)-1] {
		err = ErrExceedQuota
		return
	}

	for i, v := range cachedIDs {
		if v == candidateID && i < len(cachedIDs)-1 {
			err = ErrAlreadySwiped
			return
		}
	}

	likedIDs := []uint64{}
	alreadyLiked := false
	likedCandidates, err := getUserInterestedCandidatesRepo(ctx)
	if err != nil {
		return
	}

	for _, c := range likedCandidates {
		likedIDs = append(likedIDs, c.ID)
		if c.ID == candidateID {
			alreadyLiked = true
		}
	}

	if alreadyLiked {
		err = ErrAlreadySwiped
		return
	}

	if liked {
		err = insertUserInterestRepo(ctx, candidateID)
		if err != nil {
			return
		}
	}

	if candidateID != cachedIDs[len(cachedIDs)-1] {
		err = cacheCandidateIDRepo(ctx, candidateID)
		if err != nil {
			return
		}
		cachedIDs = append(cachedIDs, candidateID)
	}

	if verified || len(cachedIDs) < dailyQuota {
		userID := ctx.Value(auth.ClaimsKeyUserID).(float64)
		excludedIDs := append(cachedIDs, likedIDs...)
		nextCandidate, err = getCandidateRepo(ctx, append(excludedIDs, uint64(userID)))
		_ = cacheCandidateIDRepo(ctx, nextCandidate.ID)
	} else {
		err = ErrExceedQuota
	}

	return
}

func subscribePremium(ctx context.Context) (err error) {
	userID := ctx.Value(auth.ClaimsKeyUserID).(float64)
	user, err := getUserByIDRepo(ctx, uint64(userID))
	if err != nil {
		return
	}

	if user.Verified {
		err = ErrAlreadyVerified
		return
	}

	user.Verified = true
	err = updateUserRepo(ctx, user)
	return
}

func getUserInterest(ctx context.Context) (candidates []User, err error) {
	candidates, err = getUserInterestedCandidatesRepo(ctx)
	return
}
