# Micro Payment Gateway

This is a Go-based microservice designed to handle payment transactions such as deposits, withdrawals, and callbacks. The service utilizes circuit breakers and payment gateway routing to ensure high availability and resilience. It supports both HTTP and TCP protocols, including ISO8583 message processing over TCP.

## Table of Contents

- [Micro Payment Gateway](#micro-payment-gateway)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Prerequisites](#prerequisites)
  - [Setup](#setup)
    - [1. Clone the Repository](#1-clone-the-repository)
    - [2. Configure Environment Variables](#2-configure-environment-variables)
      - [Example `.env` file](#example-env-file)
    - [3. Install Go Dependencies](#3-install-go-dependencies)
  - [Running the Service](#running-the-service)
    - [Using Docker](#using-docker)
        - [1. Build and Run with Docker Compose](#1-build-and-run-with-docker-compose)
        - [2. Accessing the Service](#2-accessing-the-service)
    - [Running Locally](#running-locally)
  - [Testing](#testing)
    - [Running Unit Tests](#running-unit-tests)
  - [Environment Variables](#environment-variables)
  - [Project Structure](#project-structure)

## Features

- **Deposit and Withdrawal Operations:** Supports secure deposit and withdrawal transactions.
- **Payment Gateway Routing:** Dynamically routes transactions through multiple payment gateways based on availability and performance.
- **Circuit Breaker Pattern:** Implements circuit breakers to handle failures gracefully and maintain system stability.
- **HTTP and TCP Support:** Provides RESTful APIs over HTTP and supports ISO8583 message processing over TCP.
- **Logging:** Structured and configurable logging using Uber's Zap library.
- **Configuration Management:** Easy configuration through environment variables and `.env` files.
- **Dockerized Deployment:** Supports containerization using Docker and orchestration through Docker Compose.

## Prerequisites

Before setting up the project, ensure you have the following installed on your system:

- [Go 1.20+](https://golang.org/dl/)
- [Docker](https://www.docker.com/products/docker-desktop/)
- [Docker Compose](https://docs.docker.com/compose/install/)
- [Git](https://git-scm.com/downloads)

## Setup

### 1. Clone the Repository

```bash
git clone https://github.com/yourusername/micro-payment-gateway.git
cd micro-payment-gateway
```

### 2. Configure Environment Variables
Create a `.env` file in the root directory and add the necessary environment variables as described in the Environment Variables section.

#### Example `.env` file
```env
SERVER_ADDRESS=0.0.0.0:8080
TCP_PORT=0.0.0.0:9090
```

### 3. Install Go Dependencies
Ensure you have Go installed and your `GOPATH` is correctly set.

```bash
go mod tidy
```

This command will download all the necessary dependencies for the project.

## Running the Service
You can run the Micro Payment Gateway service using Docker or directly on your local machine.

### Using Docker

##### 1. Build and Run with Docker Compose
Ensure Docker and Docker Compose are installed and running on your system.

```bash
docker compose up --build
```

This command will build the Docker image and start the containers as defined in the docker-compose.yml file.

##### 2. Accessing the Service

- HTTP API: Available at http://localhost:8080
- TCP Server: Listens on localhost:9090
- Open API doc available at http://localhost:8080/swagger

### Running Locally

1. Start the HTTP Server along with TCP server
   
```bash
go run cmd/main.go
```

## Testing

### Running Unit Tests

```bash
go test  ./internal/service  -v
go test  ./internal/http/handler  -v
```

with coverage

```bash
go test  ./internal/service  -cover
go test  ./internal/http/handler  -cover
```


## Environment Variables
| Variable       | Description                               |
| -------------- | ----------------------------------------- |
| SERVER_ADDRESS | The address and port for the HTTP server. |
| TCP_PORT       | The address and port for the TCP server.  |

## Project Structure

```bash
micro-payment-gateway/
├── Dockerfile                    # Dockerfile for building the service
├── README.md                     # Project documentation
├── cmd/
│   └── main.go                   # HTTP server entry point
├── design.md                     # Design documentation
├── docker-compose.yml            # Docker Compose configuration
├── Makefile                      # Helper Makefile
├── docs/
│   └── swagger.yaml              # Swagger/OpenAPI specification
├── go.mod                        # Go module dependencies
├── go.sum                        # Go module checksum file
├── internal/
│   ├── app/
│   │   └── config/
│   │       └── config.go         # Configuration loading
│   ├── http/
│   │   ├── handler/
│   │   │   ├── handler.go        # HTTP request handlers
│   │   │   └── handler_test.go   # Handler tests
│   │   └── routes.go             # HTTP routes
│   ├── logger/
│   │   └── logger.go             # Logging setup
│   ├── service/
│   │   ├── database/
│   │   │   └── wallet.go         # Wallet database interactions
│   │   ├── gateway/
│   │   │   ├── gateway.go        # Payment gateway interface and factory
│   │   │   ├── pga.go            # Implementation for Payment Gateway A
│   │   │   └── pgb.go            # Implementation for Payment Gateway B
│   │   ├── model/
│   │   │   └── model.go          # Data models
│   │   ├── service.go            # Core business logic
│   │   └── service_test.go       # Service tests
│   └── tcp/
│       ├── iso8583.go            # ISO8583 message processing
│       └── server.go             # TCP server implementation
├── openapitools.json             # OpenAPI Generator configuration
├── pkg/                          # External packages and utilities
└── swagger-docs/
    └── index.html                # Swagger UI HTML entry point
```