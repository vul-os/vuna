package plans

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Plan struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Interval string `json:"interval"`
	Amount   int    `json:"amount"`
}

type PaystackPlanManager struct {
	secretKey string
}

func NewPaystackPlanManager(secretKey string) *PaystackPlanManager {
	return &PaystackPlanManager{
		secretKey: secretKey,
	}
}

func (manager *PaystackPlanManager) CreatePlan(p Plan) (*Plan, error) {
	url := "https://api.paystack.co/plan"
	method := "POST"

	payload, err := json.Marshal(p)
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

	var plan Plan
	err = json.NewDecoder(resp.Body).Decode(&plan)
	if err != nil {
		return nil, err
	}

	return &plan, nil
}

func (manager *PaystackPlanManager) GetPlan(planID string) (*Plan, error) {
	url := fmt.Sprintf("https://api.paystack.co/plan/%s", planID)
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

	var plan Plan
	err = json.NewDecoder(resp.Body).Decode(&plan)
	if err != nil {
		return nil, err
	}

	return &plan, nil
}

func (manager *PaystackPlanManager) UpdatePlan(planID string, p Plan) error {
	url := fmt.Sprintf("https://api.paystack.co/plan/%s", planID)
	method := "PUT"

	payload, err := json.Marshal(p)
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

func (manager *PaystackPlanManager) DeletePlan(planID string) error {
	url := fmt.Sprintf("https://api.paystack.co/plan/%s", planID)
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
