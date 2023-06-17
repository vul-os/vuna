package plan

import (
	"net/http"
	"firebase.google.com/go/auth"
)

type UserPlanService struct {
	httpClient  *http.Client
	authClient  *auth.Client
	paystackKey string
}

func NewUserPlanService(
		authClient *auth.Client, 
		paystackKey string,
) *UserPlanService {
	return &UserPlanService{
		authClient:  authClient,
		paystackKey: paystackKey,
	}
}