package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"razorpay-recovery-intelligence/backend/internal/domain"
)

type MLPredictionRequest struct {
	TransactionAmount   float64 `json:"transaction_amount"`
	PaymentMethod       string  `json:"payment_method"`
	FailureReason       string  `json:"failure_reason"`
	AttemptCount        int     `json:"attempt_count"`
	CustomerSuccessRate float64 `json:"customer_success_rate"`
	CustomerFailureRate float64 `json:"customer_failure_rate"`
	CustomerValue       float64 `json:"customer_value"`
	HourOfDay           int     `json:"hour_of_day"`
	DayOfWeek           int     `json:"day_of_week"`
}

type MLPredictionResponse struct {
	RecoveryProbability float64             `json:"recovery_probability"`
	Confidence          float64             `json:"confidence"`
	ModelVersion        string              `json:"model_version"`
	SHAPFactors         []domain.SHAPFactor `json:"shap_factors"`
}

type MLClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewMLClient(baseURL string, timeoutSec int) *MLClient {
	if timeoutSec <= 0 {
		timeoutSec = 5
	}
	return &MLClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: time.Duration(timeoutSec) * time.Second,
		},
	}
}

func (c *MLClient) PredictRecovery(ctx context.Context, req MLPredictionRequest) (*MLPredictionResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal prediction request: %w", err)
	}

	url := fmt.Sprintf("%s/predict", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		log.Printf("[MLClient] Call to %s failed: %v. Using statistical baseline fallback.", url, err)
		return c.fallbackPrediction(req), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[MLClient] Non-200 response (%d) from ML service. Using statistical fallback.", resp.StatusCode)
		return c.fallbackPrediction(req), nil
	}

	var predResp MLPredictionResponse
	if err := json.NewDecoder(resp.Body).Decode(&predResp); err != nil {
		return c.fallbackPrediction(req), nil
	}

	return &predResp, nil
}

func (c *MLClient) fallbackPrediction(req MLPredictionRequest) *MLPredictionResponse {
	baseP := 0.50
	switch req.FailureReason {
	case string(domain.FailureBankTimeout):
		baseP = 0.70 - (float64(req.AttemptCount) * 0.10)
	case string(domain.FailureNetworkError):
		baseP = 0.82 - (float64(req.AttemptCount) * 0.12)
	case string(domain.FailureInsufficientFunds):
		baseP = 0.45 - (float64(req.AttemptCount) * 0.08)
	case string(domain.FailureCardExpired):
		baseP = 0.10
	case string(domain.FailureCustomerAbandonment):
		baseP = 0.58
	case string(domain.FailurePaymentMethodFailure):
		baseP = 0.65 - (float64(req.AttemptCount) * 0.08)
	case string(domain.FailureTechnicalError):
		baseP = 0.68 - (float64(req.AttemptCount) * 0.10)
	}

	p := baseP*0.65 + (req.CustomerSuccessRate * 0.35)
	if p < 0.05 {
		p = 0.05
	}
	if p > 0.95 {
		p = 0.95
	}

	conf := 0.5 + (abs(p-0.5) * 1.0)
	if conf > 0.95 {
		conf = 0.95
	}

	return &MLPredictionResponse{
		RecoveryProbability: round(p, 4),
		Confidence:          round(conf, 4),
		ModelVersion:        "v1.0.0-baseline-fallback",
		SHAPFactors: []domain.SHAPFactor{
			{Feature: "failure_reason", Impact: 0.22, Direction: "positive", Description: "Failure characteristic pattern"},
			{Feature: "customer_success_rate", Impact: 0.18, Direction: "positive", Description: "Customer historical reliability"},
			{Feature: "attempt_count", Impact: -0.12, Direction: "negative", Description: "Previous failed attempts penalty"},
		},
	}
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func round(v float64, decimals int) float64 {
	var mult float64 = 1
	for i := 0; i < decimals; i++ {
		mult *= 10
	}
	return float64(int(v*mult+0.5)) / mult
}
