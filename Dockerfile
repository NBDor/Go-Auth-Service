# Use the official Go image as a parent image
FROM golang:1.22-alpine

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files first (if they exist)
COPY go.mod go.sum* ./

# Download dependencies (if any)
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN go build -o main ./cmd/server

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./main"]