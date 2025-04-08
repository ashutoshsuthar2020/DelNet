# Use the official Golang image as a base
FROM golang:1.23.2 AS builder

# Set the working directory
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download the dependencies
RUN go mod tidy

# Copy the source code
COPY . .

# Build the application for Linux
RUN GOOS=linux GOARCH=amd64 go build -o main .

# Use a minimal base image for production
FROM alpine:latest

# Install necessary libraries
RUN apk add --no-cache libc6-compat

# Set the working directory
WORKDIR /root/

# Copy the compiled binary from the builder
COPY --from=builder /app/main .

# Ensure the binary is executable
RUN chmod +x ./main

# Expose the application port
EXPOSE 8089

# Run the compiled binary
CMD ["./main"]
