# Use the official Golang image to build the binary.
FROM golang:1.20-alpine AS builder

# Set the working directory inside the container.
WORKDIR /app

# Copy the go.mod and go.sum files.
COPY go.mod go.sum ./

# Download dependencies.
RUN go mod download

# Copy the source code into the container.
ADD . .

# Build the Go binary.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/main ./cmd/main.go

# Start a new stage from scratch.
FROM scratch

# Copy the prebuilt binary from the builder stage.
COPY --from=builder /app/main /main

COPY --from=builder /app/swagger-docs /swagger-docs

COPY .env .

# Run the compiled binary.
ENTRYPOINT ["/main"]
