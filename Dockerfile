# Use the official Go image
FROM golang:1.20 AS build

# Set the working directory
WORKDIR /app

# Copy source code
COPY . .

# Build the Go application
RUN go build -o main ./cmd/main.go

# Start a fresh image
FROM ubuntu:latest

# Set the working directory
WORKDIR /app

# Copy the built binary
COPY --from=build /app/main .

# Expose the port
EXPOSE 8080

RUN apt-get update && apt-get install -y ca-certificates

# Run the app
CMD ["/app/main"]