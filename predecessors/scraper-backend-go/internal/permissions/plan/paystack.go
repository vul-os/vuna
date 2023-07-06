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

type Plan struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	PlanCode    string `json:"plan_code"`
	Description string `json:"description"`
	Amount      int    `json:"amount"`
	Interval    string `json:"interval"`
	SendInvoices bool   `json:"send_invoices"`
	SendSMS     bool   `json:"send_sms"`
	Currency    string `json:"currency"`
}

type PaystackPlanResponse struct {
	Status  bool `json:"status"`
	Message string `json:"message"`
	Data struct {
		// Other fields...
		Plan `json:"plan"`
	} `json:"data"`
}

func (s *UserPlanService) GetPlan(sub string) (Plan, error) {
	req, err := http.NewRequest("GET",
		fmt.Sprintf("https://api.paystack.co/subscription/%s", sub), nil)
	if err != nil {
		return Plan{}, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.PaystackKey))

	httpClient := http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return Plan{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Plan{}, errors.New("Paystack API returned non-200 status code")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Plan{}, err
	}

	var plansResp PaystackPlanResponse
	err = json.Unmarshal(body, &plansResp)
	if err != nil {
		return Plan{}, err
	}
	if plansResp.Data.PlanCode == "" {
		return Plan{}, errors.New("No plans found for the user")
	}

	return plansResp.Data.Plan, nil
}

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
