package tcp

import (
	"fmt"
	"net"

	"github.com/wajidp/micro-payment-gateway/internal/logger"
	"github.com/wajidp/micro-payment-gateway/internal/service"
	"github.com/wajidp/micro-payment-gateway/internal/service/model"
)

//TCPServer which wraps the service
type TCPServer struct {
	service service.PaymentProcessorRepo
}

func NewTCPServer(_service service.PaymentProcessorRepo) *TCPServer {
	return &TCPServer{service: _service}
}

// StartTCPServer starts the TCP server to accept ISO8583 messages
func (s *TCPServer) Start(port string) error {
	//creates listener
	listener, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("failed to start TCP server: %v", err)
	}
	defer listener.Close()

	logger.Infof("TCP server listening on %s", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Infof("failed to accept connection: %v", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

func (s *TCPServer) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Read incoming message
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		logger.Infof("failed to read from connection: %v", err)
		return
	}

	rawMessage := buf[:n]
	logger.Infof("Received message: %s", rawMessage)

	user, amt, err := parseISO8583Message(rawMessage)
	// Create a PaymentRequest based on the ISO8583 message
	paymentReq := &model.PaymentRequest{
		UserID:   user,
		Currency: "USD",
		Amount:   amt,
	}
	// Call the service layer to process the payment
	response, err := s.service.Deposit(paymentReq)
	if err != nil {
		logger.Infof("Failed to process payment: %v", err)
		conn.Write([]byte("Payment processing failed"))
		return
	}

	logger.Infof("Payment processed successfully: %v", response)
	conn.Write([]byte("Payment processed successfully"))

}
