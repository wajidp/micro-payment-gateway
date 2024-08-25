package gateway

import (
	"encoding/json"
	"net/http"

	"github.com/wajidp/micro-payment-gateway/internal/logger"
	"github.com/wajidp/micro-payment-gateway/internal/service/model"
)

type PGSA struct {
	httpClient *http.Client
	url        string
}

func NewPGSA() *PGSA {
	return &PGSA{
		httpClient: &http.Client{},
		url:        "http://pgsa.com",
	}
}

// processPayment handles both deposit and withdraw operations
func (pga *PGSA) processPayment(request *model.PaymentRequest, action string) (*model.PaymentResponse, error) {

	// Convert the request object into JSON
	requestBody, _ := json.Marshal(request)
	logger.Debugf("PGSA request --> %s", string(requestBody))
	respBody, err := makeHTTPRequest(pga.httpClient, pga.url, action, "application/json", string(requestBody))
	if err != nil {
		return nil, err
	}

	logger.Debugf("PGSA response --> %s", string(respBody))

	// decoding
	var paymentResponse model.PaymentResponse

	if err := json.Unmarshal(respBody, &paymentResponse); err != nil {
		return nil, model.WrapError(model.ErrHttpResponseFailure, string(respBody))
	}

	paymentResponse.TransactionID = request.TransactionID
	return &paymentResponse, nil
}

// Deposit handles deposit requests
func (pga *PGSA) Deposit(request *model.PaymentRequest) (*model.PaymentResponse, error) {
	return pga.processPayment(request, "deposit")
}

// Withdraw handles withdrawal requests
func (pga *PGSA) Withdraw(request *model.PaymentRequest) (*model.PaymentResponse, error) {
	return pga.processPayment(request, "withdraw")
}
