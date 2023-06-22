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
		SubscriptionCode string `json:"subscription_code"`
		EmailToken       string `json:"email_token"`
		Plan             struct {
			PlanCode string `json:"plan_code"`
		} `json:"plan"`
	} `json:"data"`
}

func (s *UserPlanService) GetFirstSubscriptionCodeAndToken(email string) (string, string, string, error) {
	subscriptions, err := s.GetCustomer(email)
	if err != nil {
		return "", "", "", err
	}
	subscriptionCode := subscriptions[0].SubscriptionCode
	fmt.Println(fmt.Sprintf("https://api.paystack.co/subscription/%s", subscriptionCode))
	// Fetch subscription data
	subscriptionReq, err := http.NewRequest("GET",
		fmt.Sprintf("https://api.paystack.co/subscription/%s", subscriptionCode), nil)
	if err != nil {
		return "", "", "", err
	}
	httpClient := http.Client{}
	subscriptionReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.PaystackKey))
	subscriptionResp, err := httpClient.Do(subscriptionReq)
	if err != nil {
		return "", "", "", err
	}
	defer subscriptionResp.Body.Close()

	if subscriptionResp.StatusCode != http.StatusOK {
		return "", "", "", errors.New("Paystack API returned non-200 status code")
	}

	subscriptionBody, err := ioutil.ReadAll(subscriptionResp.Body)
	if err != nil {
		return "", "", "", err
	}

	var subscriptionResponse PaystackSubscriptionResponse
	err = json.Unmarshal(subscriptionBody, &subscriptionResponse)
	if err != nil {
		return "", "", "", err
	}
	fmt.Println(subscriptionResponse)
	planCode := subscriptionResponse.Data.Plan.PlanCode
	subCode := subscriptionResponse.Data.SubscriptionCode
	emailToken := subscriptionResponse.Data.EmailToken
	return planCode, subCode, emailToken, nil
}
