package payment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const PaystackBaseURL = "https://api.paystack.co"

type PaystackService struct {
	SecretKey string
	Client    *http.Client
}

func NewPaystackService(secretKey string) *PaystackService {
	return &PaystackService{
		SecretKey: secretKey,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type WebhookEvent struct {
	Event string `json:"event"`
	Data  struct {
		Reference string          `json:"reference"`
		Status    string          `json:"status"`
		Amount    int64           `json:"amount"`
		ID        int64           `json:"id"`
		Metadata  json.RawMessage `json:"metadata"`
	} `json:"data"`
}

// TransactionInitRequest is the payload sent to Paystack
type TransactionInitRequest struct {
	Email     string `json:"email"`
	Amount    int64  `json:"amount"` // Amount in kobo (Naira * 100)
	Currency  string `json:"currency"`
	Reference string `json:"reference"`
	Callback  string `json:"callback_url"`
}

// TransactionInitResponse is what Paystack sends back
type TransactionInitResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

func (s *PaystackService) InitializeTransaction(req TransactionInitRequest) (*TransactionInitResponse, error) {
	url := fmt.Sprintf("%s/transaction/initialize", PaystackBaseURL)

	body, _ := json.Marshal(req)
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+s.SecretKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result TransactionInitResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if !result.Status {
		return nil, fmt.Errorf("paystack error: %s", result.Message)
	}

	return &result, nil
}
