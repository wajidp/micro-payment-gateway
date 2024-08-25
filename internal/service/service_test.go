package service_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
	"github.com/wajidp/micro-payment-gateway/internal/service"
	"github.com/wajidp/micro-payment-gateway/internal/service/database"
	"github.com/wajidp/micro-payment-gateway/internal/service/model"
)

// TestPaymentProcessor_Resilience_CircuitBreakerAndFallback verifies the circuit breaker mechanism
// in the PaymentProcessor. The test simulates multiple failures from a payment gateway (PGSA)
// and checks if the circuit breaker trips after the configured number of failures.
// It also verifies that the circuit breaker resets after a timeout, allowing requests to be retried.
func TestPaymentProcessor_Resilience_CircuitBreakerAndFallback(t *testing.T) {
	defer gock.Off() // Ensure gock is disabled after test

	// Ensure PgRoutingMasters is configured to select PGSA for the test
	pgms := []*model.PgRoutingMaster{
		{Currency: "USD", CountryCode: "US", PaymentGateway: "PGA", Active: true, MaxRetryCount: 3, Priority: 0}, // PGSA is selected for USD in the US
	}

	// Initialize the wallet repository and set up the wallet
	walletRepo := database.NewUserWalletRepo()
	wallet := &model.Wallet{
		Balance: 200, // Set an initial balance
	}
	walletRepo.UpdateWallet("123", wallet)

	processor := service.NewPaymentProcessor(pgms)
	// Replace the real WalletRepo with the test wallet
	processor.(*service.PaymentProcessor).WalletRepo = walletRepo

	// Simulate a sequence of failed requests to trip the circuit breaker for PGSA
	gock.New("http://pgsa.com").
		Post("/deposit").
		Times(4). // Simulate 4 failed requests
		Reply(http.StatusInternalServerError).
		JSON(map[string]string{"status": "error", "message": "Internal Server Error"})

	request := &model.PaymentRequest{
		UserID:      "123",
		Amount:      100,
		Currency:    "USD",
		CountryCode: "US",
	}

	for i := 0; i < 4; i++ {
		_, err := processor.Deposit(request)
		if i < 3 {
			assert.Error(t, err, "Expected error for failed requests")
		} else {
			assert.Equal(t, fmt.Errorf("Deposit operation failed"), err, "Expected circuit breaker to open on the 4th attempt")
		}
	}

	// Wait for the circuit breaker to reset
	time.Sleep(4 * time.Second)

	gock.Off()
	// Ensure that the circuit breaker resets and allows requests again
	gock.New("http://pgsa.com").
		Post("/deposit").
		Reply(http.StatusOK).
		JSON(map[string]string{"status": "success", "message": "Transaction processed successfully"})

	response, err := processor.Deposit(request)
	gock.Off()
	assert.NoError(t, err, "Expected no error after circuit breaker resets")
	assert.Contains(t, "success", response.Status, "Expected successful response after reset")
}

// TestPaymentProcessor_Resilience_FallbackToPGB ensures that the PaymentProcessor correctly falls back
// to a secondary payment gateway (PGB) when the primary gateway (PGA) fails. The test simulates
// failures from PGA and verifies that the processor successfully processes the deposit using PGB.
func TestPaymentProcessor_Resilience_FallbackToPGB(t *testing.T) {
	defer gock.Off() // Ensure gock is disabled after test

	// Ensure PgRoutingMasters is configured for fallback to PGB
	pgms := []*model.PgRoutingMaster{
		{Currency: "USD", CountryCode: "US", PaymentGateway: "PGA", Active: true, MaxRetryCount: 3, Priority: 0},
		{Currency: "USD", CountryCode: "US", PaymentGateway: "PGB", Active: true, MaxRetryCount: 3, Priority: 1}, // PGB as a fallback
	}

	// Initialize the wallet repository and set up the wallet
	walletRepo := database.NewUserWalletRepo()
	wallet := &model.Wallet{
		Balance: 200, // Set an initial balance
	}
	walletRepo.UpdateWallet("123", wallet)

	// Initialize the payment processor
	processor := service.NewPaymentProcessor(pgms)
	// Replace the real WalletRepo with the test wallet
	processor.(*service.PaymentProcessor).WalletRepo = walletRepo

	// Simulate PGA failure
	gock.New("http://pgsa.com").
		Post("/deposit").
		Times(3).
		Reply(http.StatusInternalServerError).
		JSON(map[string]string{"status": "error", "message": "Internal Server Error"})

	// Simulate PGB success
	gock.New("http://pgsb.com").
		Post("/deposit").
		Reply(http.StatusOK).
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

	request := &model.PaymentRequest{
		UserID:      "123",
		Amount:      100,
		Currency:    "USD",
		CountryCode: "US",
	}

	// Attempt deposit, expecting it to fall back to PGB after PGA fails
	response, err := processor.Deposit(request)

	gock.Off()
	fmt.Println(response, err)
	assert.NoError(t, err, "Expected fallback to PGB with no error")
	assert.Contains(t, "success", response.Status, "Expected successful response from PGB")
}
