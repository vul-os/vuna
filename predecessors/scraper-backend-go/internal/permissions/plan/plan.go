package plan

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type PaystackSubscriptionResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Plan struct {
			PlanCode string `json:"plan_code"`
		} `json:"plan"`
	} `json:"data"`
}

type PaystackCustomerResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Subscriptions []struct {
			SubscriptionCode string `json:"subscription_code"`
		} `json:"subscriptions"`
	} `json:"data"`
}

func (s *UserPlanService) GetMaxProductsForUser(user_email string) (int, error) {
	req, err := http.NewRequest("GET",
		fmt.Sprintf("https://api.paystack.co/customer/%s", user_email), nil)
	if err != nil {
		return 0, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.PaystackKey))

	httpClient := http.Client{}
	resp, err := httpClient.Do(req)
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

	subscriptionCode := customerResponse.Data.Subscriptions[0].SubscriptionCode
	fmt.Println(subscriptionCode)
	// Fetch subscription data
	subscriptionReq, err := http.NewRequest("GET",
		fmt.Sprintf("https://api.paystack.co/subscription/%s", subscriptionCode), nil)
	if err != nil {
		return 0, err
	}

	subscriptionReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.PaystackKey))
	subscriptionResp, err := httpClient.Do(subscriptionReq)
	if err != nil {
		return 0, err
	}
	defer subscriptionResp.Body.Close()

	if subscriptionResp.StatusCode != http.StatusOK {
		return 0, errors.New("Paystack API returned non-200 status code")
	}

	subscriptionBody, err := ioutil.ReadAll(subscriptionResp.Body)
	if err != nil {
		return 0, err
	}

	var subscriptionResponse PaystackSubscriptionResponse
	err = json.Unmarshal(subscriptionBody, &subscriptionResponse)
	if err != nil {
		return 0, err
	}

	planCode := subscriptionResponse.Data.Plan.PlanCode

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
