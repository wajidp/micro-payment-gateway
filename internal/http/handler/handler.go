package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/wajidp/micro-payment-gateway/internal/service"
	"github.com/wajidp/micro-payment-gateway/internal/service/model"
)

// Handler which holds the processor repo
type Handler struct {
	service service.PaymentProcessorRepo
}

// struct to handle cash in cash out request to seperate from app model
type HandlerRequest struct {
	UserID      string `json:"userId"`
	Currency    string `json:"currency"`
	Amount      int64  `json:"amount"`
	Exponent    int    `json:"exponent"`
	CountryCode string `json:"country_code"`
}

// NewHandler create the handler
func NewHandler(_service service.PaymentProcessorRepo) *Handler {
	return &Handler{
		service: _service,
	}
}

// handleRequest is a helper method that processes both deposit and withdrawal requests
func (h *Handler) handleRequest(c *gin.Context, reqType model.RequestType) {
	var req HandlerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"details": err.Error(), "message": "Bad Request"})
		return
	}

	paymentReq := &model.PaymentRequest{
		UserID:      req.UserID,
		Currency:    req.Currency,
		Amount:      req.Amount,
		Exponent:    req.Exponent,
		CountryCode: req.CountryCode,
		Type:        reqType,
	}

	var (
		response *model.PaymentResponse
		err      error
	)

	//check the req type & invokes the respective methods
	if reqType == model.Deposit {
		response, err = h.service.Deposit(paymentReq)
	} else {
		response, err = h.service.Withdraw(paymentReq)
	}

	// return error based on scenario
	if err != nil {
		switch {
		//if validation error return bad request
		case errors.Is(err, model.ErrValidation):
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation failed",
				"details": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to process request",
				"details": err.Error(),
			})
		}
		return
	}

	//on success return accepted
	c.JSON(http.StatusAccepted, response)
}

// Deposit handles deposit requests
func (h *Handler) Deposit(c *gin.Context) {
	h.handleRequest(c, model.Deposit)
}

// Withdraw handles withdrawal requests
func (h *Handler) Withdraw(c *gin.Context) {
	h.handleRequest(c, model.Withdraw)
}

//HandleCallback to handle callback request
func (h *Handler) HandleCallback(c *gin.Context) {
	var callbackReq *model.CallbackRequest

	if err := c.ShouldBindJSON(&callbackReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"details": err.Error(), "message": "Bad Request"})
		return
	}

	err := h.service.HandleCallback(callbackReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"details": err.Error(), "message": "Error"})
		return
	}
	//on transaction approved return ok
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
