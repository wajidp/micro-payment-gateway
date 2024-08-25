package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/h2non/gock"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/wajidp/micro-payment-gateway/internal/service"
	"github.com/wajidp/micro-payment-gateway/internal/service/model"
)

// initGock initializes gock for mocking HTTP requests.
func initGock() {
	gock.DisableNetworking()

	// Mocking Gateway A (PGA) with JSON over HTTP
	gock.New("http://pgsa.com").
		Post("/deposit").
		Reply(200).
		JSON(map[string]string{"status": "success", "message": "Transaction processed successfully"})

	gock.New("http://pgsa.com").
		Post("/withdraw").
		Reply(200).
		JSON(map[string]string{"status": "success", "message": "Transaction processed successfully"})

	// Mocking Gateway B (PGB) with SOAP/XML over HTTP
	gock.New("http://pgsb.com").
		Post("/deposit").
		Reply(200).
		XML(`
			<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">
				<soapenv:Body>
					<ns2:depositResponse xmlns:ns2="http://pgb.com/">
						<return>
							<status>success</status>
							<message>Transaction processed successfully</message>
						</return>
					</ns2:depositResponse>
				</soapenv:Body>
			</soapenv:Envelope>
		`)

	gock.New("http://pgsb.com").
		Post("/withdraw").
		Reply(200).
		XML(`
			<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/">
				<soapenv:Body>
					<ns2:withdrawResponse xmlns:ns2="http://pgb.com/">
						<return>
							<status>success</status>
							<message>Transaction processed successfully</message>
						</return>
					</ns2:withdrawResponse>
				</soapenv:Body>
			</soapenv:Envelope>
		`)
}

// newTestServer initializes a new Gin server and handler for testing.
func newTestServer() *gin.Engine {
	processor := service.NewPaymentProcessor(model.PgRoutingMasters)
	handler := NewHandler(processor)

	router := gin.Default()
	router.POST("/deposit", handler.Deposit)
	router.POST("/withdraw", handler.Withdraw)
	router.POST("/callback", handler.HandleCallback)
	return router
}

// performRequest performs a HTTP request to the Gin server.
func performRequest(router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	reqBody, _ := json.Marshal(body)
	req, _ := http.NewRequest(method, path, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// TestHandler_Deposit tests the successful processing of a deposit request.
func TestHandler_Deposit(t *testing.T) {
	defer gock.Off()
	initGock()

	router := newTestServer()

	depositRequest := &model.PaymentRequest{
		UserID:      "123",
		Amount:      10000,
		Currency:    "USD",
		CountryCode: "US",
	}

	w := performRequest(router, "POST", "/deposit", depositRequest)
	assert.Equal(t, http.StatusAccepted, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])

	callbackRequest := &model.CallbackRequest{
		TransactionID: cast.ToString(response["id"]),
		State:         model.StateApproved,
	}

	w = performRequest(router, "POST", "/callback", callbackRequest)
	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
}

// TestHandler_Withdraw_Failure tests the failure scenario for a withdrawal request
// due to insufficient funds in the user's wallet.
func TestHandler_Withdraw_Failure(t *testing.T) {
	defer gock.Off()
	initGock()

	router := newTestServer()

	withdrawRequest := &model.PaymentRequest{
		UserID:      "123",
		Amount:      10000,
		Currency:    "USD",
		CountryCode: "US",
	}

	w := performRequest(router, "POST", "/withdraw", withdrawRequest)
	// Assert the response for insufficient funds
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["details"], "insufficient funds")
}

// TestHandler_Deposit_MissingUserID verifies that the deposit handler returns an error
// when the UserID is missing from the request.
func TestHandler_Deposit_MissingUserID(t *testing.T) {
	defer gock.Off()
	initGock()

	router := newTestServer()

	depositRequest := &model.PaymentRequest{
		Amount:      10000,
		Currency:    "USD",
		CountryCode: "US",
	}

	w := performRequest(router, "POST", "/deposit", depositRequest)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["details"], "validation error: invalid account ID")
}

// TestHandler_Deposit_InvalidAmount verifies that the deposit handler returns an error
// when an invalid amount is provided in the request.
func TestHandler_Deposit_InvalidAmount(t *testing.T) {
	defer gock.Off()
	initGock()

	router := newTestServer()

	depositRequest := &model.PaymentRequest{
		UserID:      "123",
		Amount:      -10000, // Invalid amount
		Currency:    "USD",
		CountryCode: "US",
	}

	w := performRequest(router, "POST", "/deposit", depositRequest)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["details"], "validation error: invalid amount")
}

// TestHandler_Deposit_UnsupportedCurrency verifies that the deposit handler returns an error
// when an unsupported currency is provided in the request.
func TestHandler_Deposit_UnsupportedCurrency(t *testing.T) {
	defer gock.Off()
	initGock()

	router := newTestServer()

	depositRequest := &model.PaymentRequest{
		UserID:      "123",
		Amount:      10000,
		Currency:    "XYZ", // Unsupported currency
		CountryCode: "US",
	}

	w := performRequest(router, "POST", "/deposit", depositRequest)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["details"], "validation error: invalid currency")
}
