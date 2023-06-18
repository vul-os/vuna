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

func (s *UserPlanService) CreateTransaction(w http.ResponseWriter,
	r *http.Request) {
	user, ok := r.Context().Value("user").(scraperAuth.User)
	if !ok {
		http.Error(w, "Failed to retrieve user from context", http.StatusInternalServerError)
		return
	}
	email := user.Email
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
