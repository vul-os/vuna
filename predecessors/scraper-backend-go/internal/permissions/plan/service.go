package plan

import (
	"firebase.google.com/go/v4/auth"
)

type UserPlanService struct {
	AuthClient  *auth.Client
	PaystackKey string
}

func NewUserPlanService(
		authClient *auth.Client, 
		paystackKey string,
) *UserPlanService {
	return &UserPlanService{
		AuthClient:  authClient,
		PaystackKey: paystackKey,
	}
}