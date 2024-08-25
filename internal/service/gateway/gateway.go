package gateway

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/wajidp/micro-payment-gateway/internal/service/model"
)

// Constants for actions
const (
	ActionDeposit  = "deposit"
	ActionWithdraw = "withdraw"
)

type PaymentGateway interface {
	Deposit(request *model.PaymentRequest) (*model.PaymentResponse, error)
	Withdraw(request *model.PaymentRequest) (*model.PaymentResponse, error)
}

type GatewayFactoryInterface interface {
	GetPaymentGatewayInstance(provider string) (PaymentGateway, error)
}

type GatewayFactory struct{}

func NewGatewayFactory() *GatewayFactory {
	return &GatewayFactory{}
}

func (f *GatewayFactory) GetPaymentGatewayInstance(provider string) (PaymentGateway, error) {
	switch provider {
	case "PGA":
		return NewPGSA(), nil
	case "PGB":
		return NewPGSB(), nil

	default:
		return nil, errors.New("payment gateway not implemented")
	}
}

// makeHTTPRequest is a common function to handle HTTP requests for both JSON and XML requests
func makeHTTPRequest(httpClient *http.Client, url, action, contentType, requestBody string) ([]byte, error) {
	// Create HTTP request
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/%s", url, action), strings.NewReader(requestBody))
	if err != nil {
		return nil, model.WrapError(model.ErrHttpRequestFailure, err.Error())
	}
	req.Header.Set("Content-Type", contentType)

	// Send HTTP request
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, model.WrapError(model.ErrHttpRequestFailure, err.Error())
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, model.WrapError(model.ErrHttpResponseFailure, err.Error())
	}

	if !IsSuccessStatus(resp.StatusCode) {
		return nil, model.WrapError(model.ErrHttpResponseFailure, string(respBody))
	}

	return respBody, nil
}

// isSuccessStatus checks if the status code indicates success (2xx)
func IsSuccessStatus(statusCode int) bool {
	return statusCode >= 200 && statusCode < 300
}
