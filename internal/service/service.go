package service

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sony/gobreaker"
	"github.com/wajidp/micro-payment-gateway/internal/logger"
	"github.com/wajidp/micro-payment-gateway/internal/service/database"
	"github.com/wajidp/micro-payment-gateway/internal/service/gateway"
	"github.com/wajidp/micro-payment-gateway/internal/service/model"
	"go.uber.org/zap"
)

// Action constants
const (
	ActionDeposit  = "Deposit"
	ActionWithdraw = "Withdraw"
)

type PaymentProcessorRepo interface {
	Deposit(request *model.PaymentRequest) (*model.PaymentResponse, error)
	Withdraw(request *model.PaymentRequest) (*model.PaymentResponse, error)
	HandleCallback(callback *model.CallbackRequest) error
}

type PaymentProcessor struct {
	Factory          gateway.GatewayFactoryInterface
	WalletRepo       model.WalletRepository
	CircuitBreakers  map[string]*gobreaker.CircuitBreaker
	PgRoutingMasters []*model.PgRoutingMaster
}

func NewPaymentProcessor(pgmasters []*model.PgRoutingMaster) PaymentProcessorRepo {
	return &PaymentProcessor{
		Factory:          gateway.NewGatewayFactory(),
		CircuitBreakers:  make(map[string]*gobreaker.CircuitBreaker),
		WalletRepo:       database.NewUserWalletRepo(),
		PgRoutingMasters: pgmasters,
	}
}

// processPayment handles both Deposit and Withdraw operations with circuit breaker and PG switching
func (p *PaymentProcessor) processPayment(request *model.PaymentRequest, action string) (*model.PaymentResponse, error) {

	//masking sensitive information
	logger.Info(fmt.Sprintf("%s Request ", action), zap.Object("transaction", request))
	// Perform common validations
	if err := validateRequest(request); err != nil {
		return nil, err
	}

	// creates a new id
	id := uuid.New().String()
	txn := &model.Transaction{
		ID:       id,
		UserID:   request.UserID,
		Amount:   request.Amount,
		Currency: request.Currency,
		Type:     action,
		State:    model.StateAuthorized,
	}
	request.TransactionID = id

	var lastError error

	// Iterate over all available payment gateways
	for _, pgm := range p.PgRoutingMasters {

		logger.Debugf("Trying PG: %s", pgm.PaymentGateway)

		// Get the circuit breaker for the payment gateway
		cb, exists := p.CircuitBreakers[pgm.PaymentGateway]
		if !exists {
			cb = gobreaker.NewCircuitBreaker(gobreaker.Settings{
				Name:     pgm.PaymentGateway,
				Timeout:  2 * time.Second,
				Interval: 500 * time.Millisecond,

				OnStateChange: func(name string, from, to gobreaker.State) {
					logger.Infof("Circuit breaker state changed from %s to %s for %s", from, to, name)
				},
				ReadyToTrip: p.tripLogic,
			})
			p.CircuitBreakers[pgm.PaymentGateway] = cb
		}

		// Check if the circuit breaker is open
		if cb.State() == gobreaker.StateOpen {
			logger.Infof("Circuit breaker open for PG: %s, skipping...", pgm.PaymentGateway)
			continue
		}

		// Get the appropriate payment gateway instance
		pg, err := p.Factory.GetPaymentGatewayInstance(pgm.PaymentGateway)
		if err != nil {
			lastError = err
			continue
		}

		// Execute the action
		operation := func() (interface{}, error) {
			if action == ActionDeposit {
				return pg.Deposit(request)
			}
			return pg.Withdraw(request)
		}

		result, err := cb.Execute(operation)
		if err != nil {
			logger.Infof("%s operation failed for PG %s: %v", action, pgm.PaymentGateway, err)
			lastError = err
			continue
		}
		if err := p.WalletRepo.UpdateTransaction(txn); err != nil {
			lastError = err
			return nil, err
		}

		// If successful, return the response
		return result.(*model.PaymentResponse), nil
	}

	// If all gateways failed, return the last error
	if lastError != nil {
		return nil, fmt.Errorf("%s operation failed: %v", action, lastError)
	}
	return nil, fmt.Errorf("%s operation failed", action)
}

// Deposit handles deposit requests
func (p *PaymentProcessor) Deposit(request *model.PaymentRequest) (*model.PaymentResponse, error) {
	res, err := p.processPayment(request, ActionDeposit)
	if err != nil {
		return res, err
	}

	return res, nil
}

// Withdraw handles withdrawal requests
func (p *PaymentProcessor) Withdraw(request *model.PaymentRequest) (*model.PaymentResponse, error) {

	wallet, err := p.WalletRepo.GetWallet(request.UserID)
	if err != nil {
		return nil, err
	}
	if wallet.Balance < request.Amount {
		return nil, model.WrapError(model.ErrValidation, "validation error: insufficient funds")
	}
	res, err := p.processPayment(request, gateway.ActionWithdraw)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// HandleCallback processes the callback and updates the transaction and wallet accordingly
func (p *PaymentProcessor) HandleCallback(callback *model.CallbackRequest) error {
	// Retrieve the transaction by ID
	txn, err := p.WalletRepo.GetTransaction(callback.TransactionID)
	if err != nil {
		return err
	}

	// Update the transaction state based on the callback status
	if callback.State == model.StateApproved {
		txn.State = model.StateApproved

		// Update the user's wallet since the transaction is approved
		wallet, err := p.WalletRepo.GetWallet(txn.UserID)
		if err != nil {
			return err
		}

		if txn.Type == ActionDeposit {
			wallet.Balance += txn.Amount
		} else if txn.Type == ActionWithdraw {
			if wallet.Balance < txn.Amount {
				return errors.New("insufficient funds")
			}
			wallet.Balance -= txn.Amount
		}

		if err := p.WalletRepo.UpdateWallet(txn.UserID, wallet); err != nil {
			return err
		}
	} else if callback.State == model.StateFailed {
		txn.State = model.StateFailed
	}
	if err := p.WalletRepo.UpdateTransaction(txn); err != nil {
		return err
	}
	return nil
}

// tripLogic defines when the circuit breaker should trip
func (p *PaymentProcessor) tripLogic(counts gobreaker.Counts) bool {
	failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
	logger.Infof("Breaker Trip Check: Requests=%d, Failures=%d, FailureRatio=%.2f", counts.Requests, counts.TotalFailures, failureRatio)
	return counts.Requests >= 3 && failureRatio >= 0.6
}

// validateRequest performs common validations on the PaymentRequest
func validateRequest(request *model.PaymentRequest) error {
	if !validateAccount(request.UserID) {
		return model.WrapError(model.ErrValidation, "invalid account ID")
	}
	if !validateCurrency(request.Currency) {
		return model.WrapError(model.ErrValidation, "invalid currency")
	}
	if !validateAmount(request.Amount, request.Exponent) {
		return model.WrapError(model.ErrValidation, "invalid amount")
	}
	return nil
}

// validateAccount checks if the account ID is valid (stub implementation)
func validateAccount(userID string) bool {
	return userID != ""
}

// validateCurrency checks if the currency is supported
func validateCurrency(currency string) bool {
	return model.SupportedCurrencies[currency]
}

// validateAmount checks if the amount and exponent are valid
func validateAmount(amount int64, exponent int) bool {
	return amount > 0
}
