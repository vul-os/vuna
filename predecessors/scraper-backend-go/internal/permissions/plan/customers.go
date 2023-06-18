package plan

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
)

type PaystackCustomerResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Subscriptions []Subscription `json:"subscriptions"`
	} `json:"data"`
}


type Subscription struct {
	SubscriptionCode string `json:"subscription_code"`
}

func (s *UserPlanService) GetCustomer(email string) ([]Subscription, error) {
	req, err := http.NewRequest("GET",
		fmt.Sprintf("https://api.paystack.co/customer/%s", email), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.PaystackKey))

	httpClient := http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Paystack API returned non-200 status code")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var customerResponse PaystackCustomerResponse
	err = json.Unmarshal(body, &customerResponse)
	if err != nil {
		return nil, err
	}
	if len(customerResponse.Data.Subscriptions) == 0 {
		return nil, errors.New("No subscriptions found for the user")
	}

	return customerResponse.Data.Subscriptions, nil
}
