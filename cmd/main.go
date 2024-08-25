package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/wajidp/micro-payment-gateway/internal/app/config"
	"github.com/wajidp/micro-payment-gateway/internal/http"
	"github.com/wajidp/micro-payment-gateway/internal/logger"
	"github.com/wajidp/micro-payment-gateway/internal/service"
	"github.com/wajidp/micro-payment-gateway/internal/service/model"
	"github.com/wajidp/micro-payment-gateway/internal/tcp"
)

func main() {
	//inits config
	if err := config.InitConfig(); err != nil {
		log.Fatalf("%v - %v", "Cannot Instantaniate Config", err.Error())
	}

	//setup logger
	logger.SetUp()
	gin.SetMode(gin.ReleaseMode)
	//inits default gin router
	router := gin.Default()
	//create service
	service := service.NewPaymentProcessor(model.PgRoutingMasters)
	//register routes
	http.RegisterRoutes(router, service)

	//start the tcp server for iso8583 implementation
	go tcp.NewTCPServer(service).Start(config.AppConfig.TcpPort)

	logger.Infof("Starting HTTP Server %s", config.AppConfig.ServerAddress)
	//run the server
	if err := router.Run(config.AppConfig.ServerAddress); err != nil {
		log.Fatal(fmt.Sprintf("Failed to start server %v", err))
	}

}
