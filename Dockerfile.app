# Build stage
FROM golang:1.23.1 AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Install tools
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o chirpy

# Run stage
FROM debian:stable-slim

# Install system packages
RUN apt-get update && apt-get install -y \
    postgresql-client \
    netcat-traditional \
    bash

# Create necessary directories
RUN mkdir -p /public
RUN mkdir -p /app
RUN mkdir -p /app/sql

# Copy and set permissions for wait-for-it.sh
COPY ./scripts/wait-for-it.sh /app/scripts/
RUN chmod +x /app/scripts/wait-for-it.sh
COPY scripts/goose-start.sh /app/scripts/
RUN chmod +x /app/scripts/goose-start.sh

# Copy SQL files and directories
COPY ./sql /app/sql

# Copy sqlc from the builder stage
COPY --from=builder /go/bin/sqlc /usr/local/bin/sqlc
COPY --from=builder /app/chirpy /app/chirpy
COPY sqlc.yaml /app/sqlc.yaml

# Copy your application files
COPY index.html /public/index.html
COPY .env /app/.env

RUN chmod +x /app/chirpy

# Expose the port the app runs on
EXPOSE 8080
ENV PORT=8080
