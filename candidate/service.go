package candidate

import "context"

const (
	dailyQuota = 10
)

var (
	getLastCachedCandidateRepo      = getLastCachedCandidate
	getUserByIDRepo                 = getUserByID
	getUserInterestedCandidatesRepo = getUserInterestedCandidates
	getCandidateRepo                = getCandidate
	getCachedCandidateIDsRepo       = getCachedCandidateIDs
	checkUserInterestRepo           = checkUserInterest
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
		var candidates []User
		var excludedIDs []uint64
		candidates, err = getUserInterestedCandidatesRepo(ctx)
		if err != nil {
			return
		}

		for _, c := range candidates {
			excludedIDs = append(excludedIDs, c.ID)
		}
		userID := ctx.Value("user_id").(uint64)
		candidate, err = getCandidateRepo(ctx, append(excludedIDs, userID))
	}
	return
}

func swipeCandidate(ctx context.Context, candidateID uint64, liked bool) (nextCandidate User, err error) {
	ids, err := getCachedCandidateIDsRepo(ctx)
	if err != nil {
		return
	}

	verified := ctx.Value("verified").(bool)
	if !verified && len(ids) >= dailyQuota {
		err = ErrExceedQuota
		return
	}

	for _, v := range ids {
		if v == candidateID {
			err = ErrAlreadySwiped
			return
		}
	}

	alreadyLiked, err := checkUserInterestRepo(ctx, candidateID)
	if err != nil {
		return
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

	err = cacheCandidateIDRepo(ctx, candidateID)
	if err != nil {
		return
	}

	if verified || len(ids)+1 < dailyQuota {
		userID := ctx.Value("user_id").(uint64)
		nextCandidate, err = getCandidateRepo(ctx, append(ids, candidateID, userID))
	} else {
		err = ErrExceedQuota
	}

	return
}

func subscribePremium(ctx context.Context) (err error) {
	userID := ctx.Value("user_id").(uint64)
	user, err := getUserByIDRepo(ctx, userID)
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
	candidates, err = getUserInterestedCandidates(ctx)
	return
}
