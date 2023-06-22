package plan

import (
	"errors"
	"time"
)

func (s *UserPlanService) GetMaxProductsForUser(user_email string) (int, error) {
	planCode, _, _, err := s.GetFirstSubscriptionCodeAndToken(user_email)

	if err != nil {
		return 0, err
	}

	maxProducts := 0
	switch planCode {
	case "PLN_c2lqr775xgi0ffm":
		maxProducts = 10000
	case "plan_code_2":
		maxProducts = 200
	case "plan_code_3":
		maxProducts = 300
	// Add more case blocks for other plans
	default:
		return 0, errors.New("Unrecognized plan code, free plan")
	}

	return maxProducts, nil
}

func (s *UserPlanService) GetMaxProductsForTrialUser(userID string) (int, error) {
	// Get the user creation date
	userCreatedDate, err := s.GetUserCreationDate(userID)
	if err != nil {
		return 0, err
	}

	// Calculate the trial expiration date (current trial date + 7 days)
	trialExpirationDate := userCreatedDate.AddDate(0, 0, 7)

	// Get the current date
	currentDate := time.Now()

	// Check if the trial has expired
	if currentDate.After(trialExpirationDate) {
		return 0, errors.New("trial has expired")
	}

	return 10000, nil
}