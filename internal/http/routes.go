package http

import (
	"github.com/gin-gonic/gin"
	"github.com/wajidp/micro-payment-gateway/internal/http/handler"
	"github.com/wajidp/micro-payment-gateway/internal/service"
)

// RegisterRoutes register routes
func RegisterRoutes(router *gin.Engine, service service.PaymentProcessorRepo) {

	handler := handler.NewHandler(service)
	router.POST("/deposit", handler.Deposit)
	router.POST("/withdraw", handler.Withdraw)
	router.POST("/callback", handler.HandleCallback)
	// Serve the swagger-docs directory as static files
	router.Static("/swagger", "./swagger-docs")

}
