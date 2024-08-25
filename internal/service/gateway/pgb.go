package gateway

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"

	"github.com/wajidp/micro-payment-gateway/internal/logger"
	"github.com/wajidp/micro-payment-gateway/internal/service/model"
)

// Envelope is the SOAP envelope that contains the Body
type Envelope struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    Body     `xml:"Body"`
}

// Body contains the SOAP body, which can have different types of responses
type Body struct {
	XMLName          xml.Name          `xml:"Body"`
	DepositResponse  *DepositResponse  `xml:"depositResponse,omitempty"`
	WithdrawResponse *WithdrawResponse `xml:"withdrawResponse,omitempty"`
}

// DepositResponse handles the deposit response
type DepositResponse struct {
	XMLName xml.Name `xml:"depositResponse"`
	Return  Return   `xml:"return"`
}

// WithdrawResponse handles the withdraw response
type WithdrawResponse struct {
	XMLName xml.Name `xml:"withdrawResponse"`
	Return  Return   `xml:"return"`
}

// Return contains the status and message returned from the SOAP service
type Return struct {
	Status  string `xml:"status"`
	Message string `xml:"message"`
}

// PGSB represents the payment gateway service B
type PGSB struct {
	httpClient *http.Client
	url        string
}

// NewPGSB creates a new PGSB instance
func NewPGSB() *PGSB {
	return &PGSB{
		httpClient: &http.Client{},
		url:        "http://pgsb.com/soap",
	}
}

// ProcessPayment is a generic method for processing both Deposit and Withdraw operations
func (pg *PGSB) ProcessPayment(request *model.PaymentRequest, action string) (*model.PaymentResponse, error) {
	logger.Infof("Preparing PGSB %s request..", action)

	// Create the SOAP/XML request body
	soapRequest := fmt.Sprintf(`
	<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:ws="http://pgsb.com/">
	   <soapenv:Header/>
	   <soapenv:Body>
	      <ws:PaymentRequest>
	         <ws:UserID>%s</ws:UserID>
	         <ws:Currency>%s</ws:Currency>
	         <ws:Amount>%d</ws:Amount>
	         <ws:Exponent>%d</ws:Exponent>
	         <ws:CountryCode>%s</ws:CountryCode>
	      </ws:PaymentRequest>
	   </soapenv:Body>
	</soapenv:Envelope>`, request.UserID, request.Currency, request.Amount, request.Exponent, request.CountryCode)

	logger.Infof("PGSB request --> %s", soapRequest)
	// Make the HTTP request using the common utility function
	respBody, err := makeHTTPRequest(pg.httpClient, pg.url, action, "text/xml", soapRequest)
	if err != nil {
		return nil, err
	}
	logger.Infof("PGSB response --> %s", string(respBody))
	// Parse the XML response
	r, err := pg.parseResponse(respBody, action)
	if err != nil {
		return nil, model.WrapError(model.ErrHttpResponseFailure, string(respBody))
	}
	r.TransactionID = request.TransactionID
	return r, nil
}

// Deposit handles deposit requests to the PGSB
func (pg *PGSB) Deposit(request *model.PaymentRequest) (*model.PaymentResponse, error) {
	return pg.ProcessPayment(request, ActionDeposit)
}

// Withdraw handles withdrawal requests to the PGSB
func (pg *PGSB) Withdraw(request *model.PaymentRequest) (*model.PaymentResponse, error) {
	return pg.ProcessPayment(request, ActionWithdraw)
}

// parseResponse parses the SOAP response based on the action
func (pg *PGSB) parseResponse(xmlData []byte, action string) (*model.PaymentResponse, error) {
	var envelope Envelope
	if err := xml.Unmarshal(xmlData, &envelope); err != nil {
		return nil, fmt.Errorf("error parsing XML: %w", err)
	}

	var status, message string

	switch action {
	case ActionDeposit:
		if envelope.Body.DepositResponse != nil {
			status = envelope.Body.DepositResponse.Return.Status
			message = envelope.Body.DepositResponse.Return.Message
		} else {
			return nil, errors.New("no valid deposit response found")
		}
	case ActionWithdraw:
		if envelope.Body.WithdrawResponse != nil {
			status = envelope.Body.WithdrawResponse.Return.Status
			message = envelope.Body.WithdrawResponse.Return.Message
		} else {
			return nil, errors.New("no valid withdraw response found")
		}
	default:
		return nil, errors.New("invalid action")
	}

	return &model.PaymentResponse{
		Status:  status,
		Message: message,
	}, nil
}
