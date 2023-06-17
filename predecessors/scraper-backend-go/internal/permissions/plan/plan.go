package plan

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type PaystackCustomerResponse struct {
	Status  bool `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Email         string `json:"email"`
		CustomerCode  string `json:"customer_code"`
		Subscriptions []struct {
			PlanCode string `json:"plan_code"`
		} `json:"subscriptions"`
	} `json:"data"`
}

func (s *UserPlanService) GetMaxProductsForUser(user_id string) (int, error) {
	user, err := s.authClient.GetUser(context.Background(), user_id)
	if err != nil {
		return 0, fmt.Errorf("error getting user info: %v", err)
	}

	req, err := http.NewRequest("GET", 
		fmt.Sprintf("https://api.paystack.co/customer/%s", user.Email), nil)
	if err != nil {
		return 0, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.paystackKey))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, errors.New("Paystack API returned non-200 status code")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var customerResponse PaystackCustomerResponse
	err = json.Unmarshal(body, &customerResponse)
	if err != nil {
		return 0, err
	}

	if len(customerResponse.Data.Subscriptions) == 0 {
		return 0, errors.New("No subscriptions found for the user")
	}

	planCode := customerResponse.Data.Subscriptions[0].PlanCode

	maxProducts := 0

	switch planCode {
	case "plan_code_1": // Replace these with the actual plan codes
		maxProducts = 100
	case "plan_code_2":
		maxProducts = 200
	case "plan_code_3":
		maxProducts = 300
	// Add more case blocks for other plans
	default:
		return 0, errors.New("Unrecognized plan code")
	}

	return maxProducts, nil
}
