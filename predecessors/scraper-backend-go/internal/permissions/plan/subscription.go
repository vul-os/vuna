package plan

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	scraperAuth "github.com/exolutionza/scraper-backend-go/internal/auth"
)

type CreateTransactionRequest struct {
	Plan string `json:"plan"`
}

type PaystackTransactionRequest struct {
	Email  string `json:"email"`
	Amount string `json:"amount"`
	Plan   string `json:"plan"`
}

type PaystackTransactionResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

func (s *UserPlanService) CreateSubscription(w http.ResponseWriter,
	r *http.Request) {
	user, ok := r.Context().Value("user").(scraperAuth.User)
	if !ok {
		http.Error(w, "Failed to retrieve user from context", http.StatusInternalServerError)
		return
	}
	email := user.Email

	subscriptions, err := s.GetCustomer(email)
	fmt.Println(subscriptions)
	if err == nil || len(subscriptions) > 0 {
		http.Error(w, "User already is subscribed", http.StatusInternalServerError)
		return
	}

	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	transactionReq := PaystackTransactionRequest{
		Email:  email,
		Amount: "1",
		Plan:   req.Plan,
	}
	reqBody, err := json.Marshal(transactionReq)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to marshal transaction request: %v", err),
			http.StatusInternalServerError)
		return
	}

	httpReq, err := http.NewRequest("POST",
		"https://api.paystack.co/transaction/initialize",
		bytes.NewBuffer(reqBody))
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to create Paystack transaction request: %v", err),
			http.StatusInternalServerError)
		return
	}
	fmt.Println(s.PaystackKey)
	httpReq.Header.Set("Authorization", "Bearer "+s.PaystackKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpClient := http.Client{}

	resp, err := httpClient.Do(httpReq)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to send Paystack transaction request: %v", err),
			http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Paystack API returned non-200 status code",
			http.StatusInternalServerError)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Failed to read Paystack transaction response: %v", err),
			http.StatusInternalServerError)
		return
	}

	var transactionResp PaystackTransactionResponse
	err = json.Unmarshal(body, &transactionResp)
	if err != nil {
		http.Error(w,
			fmt.Sprintf(
				"Failed to unmarshal Paystack transaction response: %v",
				err),
			http.StatusInternalServerError)
		return
	}

	if !transactionResp.Status {
		http.Error(w, fmt.Sprintf("Paystack transaction creation failed: %s",
			transactionResp.Message), http.StatusInternalServerError)
		return
	}

	respData := struct {
		AuthorizationURL string `json:"authorization_url"`
	}{
		AuthorizationURL: transactionResp.Data.AuthorizationURL,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(respData); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err),
			http.StatusInternalServerError)
		return
	}
}

type DisableSubscriptionRequest struct {
	Code  string `json:"code"`
	Token string `json:"token"`
}

type DisableSubscriptionResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

func (s *UserPlanService) DisableSubscription(w http.ResponseWriter, r *http.Request) {
	// Retrieve user email from context
	user, ok := r.Context().Value("user").(scraperAuth.User)
	if !ok {
		http.Error(w, "Failed to retrieve user from context", http.StatusInternalServerError)
		return
	}
	email := user.Email

	// Get subscription code and token from the customer's subscriptions
	_, subscriptionCode, emailToken, err := s.GetFirstSubscriptionCodeAndToken(email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create disable subscription request body
	disableReq := DisableSubscriptionRequest{
		Code:  subscriptionCode,
		Token: emailToken,
	}
	fmt.Println(disableReq)
	reqBody, err := json.Marshal(disableReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to marshal disable subscription request: %v", err), http.StatusInternalServerError)
		return
	}

	// Create HTTP request to disable the subscription
	httpReq, err := http.NewRequest("POST", "https://api.paystack.co/subscription/disable", bytes.NewBuffer(reqBody))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create Paystack disable subscription request: %v", err), http.StatusInternalServerError)
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+s.PaystackKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// Send the HTTP request
	httpClient := http.Client{}
	resp, err := httpClient.Do(httpReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send Paystack disable subscription request: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to read Paystack disable subscription response: %v", err), http.StatusInternalServerError)
		return
	}

	// Unmarshal the response into a DisableSubscriptionResponse struct
	var disableResp DisableSubscriptionResponse
	err = json.Unmarshal(body, &disableResp)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to unmarshal Paystack disable subscription response: %v", err), http.StatusInternalServerError)
		return
	}

	// Check if the subscription was disabled successfully
	if !disableResp.Status {
		http.Error(w, fmt.Sprintf("Failed to disable subscription: %s", disableResp.Message), http.StatusInternalServerError)
		return
	}

	// Send the success response
	response := struct {
		Status  bool   `json:"status"`
		Message string `json:"message"`
	}{
		Status:  true,
		Message: "Subscription disabled successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode response: %v", err), http.StatusInternalServerError)
		return
	}
}
