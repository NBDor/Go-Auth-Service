FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Use a smaller image for the final container
FROM alpine:latest  

WORKDIR /root/

# Install necessary packages
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Command to run
CMD ["./main"]