package plan

import (
	"context"
	"fmt"
	"time"
)

func (s *UserPlanService) GetUserCreationDate(userID string) (time.Time, error) {
	// Retrieve the user record from Firebase Auth
	fmt.Println("here222")
	user, err := s.AuthClient.GetUser(context.Background(), userID)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get user record: %v", err)
	}
	fmt.Println("here222 ", user)

	// Get the creation timestamp of the user
	creationTime := time.Unix(user.UserMetadata.CreationTimestamp/1000, 0)
	fmt.Println(creationTime)
	return creationTime, nil
}
