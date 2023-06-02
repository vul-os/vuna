package subscriptions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Subscription struct {
	ID       string `json:"id"`
	Customer string `json:"customer"`
	Plan     string `json:"plan"`
}

type PaystackSubscriptionManager struct {
	secretKey string
}

func NewPaystackSubscriptionManager(secretKey string) *PaystackSubscriptionManager {
	return &PaystackSubscriptionManager{
		secretKey: secretKey,
	}
}

func (manager *PaystackSubscriptionManager) CreateSubscription(s Subscription) (*Subscription, error) {
	url := "https://api.paystack.co/subscription"
	method := "POST"

	payload, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(method, url, strings.NewReader(string(payload)))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+manager.secretKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var subscription Subscription
	err = json.NewDecoder(resp.Body).Decode(&subscription)
	if err != nil {
		return nil, err
	}

	return &subscription, nil
}

func (manager *PaystackSubscriptionManager) GetSubscription(subscriptionID string) (*Subscription, error) {
	url := fmt.Sprintf("https://api.paystack.co/subscription/%s", subscriptionID)
	method := "GET"

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+manager.secretKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var subscription Subscription
	err = json.NewDecoder(resp.Body).Decode(&subscription)
	if err != nil {
		return nil, err
	}

	return &subscription, nil
}

func (manager *PaystackSubscriptionManager) UpdateSubscription(subscriptionID string, s Subscription) error {
	url := fmt.Sprintf("https://api.paystack.co/subscription/%s", subscriptionID)
	method := "PUT"

	payload, err := json.Marshal(s)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, url, strings.NewReader(string(payload)))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+manager.secretKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (manager *PaystackSubscriptionManager) DeleteSubscription(subscriptionID string) error {
	url := fmt.Sprintf("https://api.paystack.co/subscription/%s", subscriptionID)
	method := "DELETE"

	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+manager.secretKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}
