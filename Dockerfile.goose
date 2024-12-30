FROM golang:1.23.1

# Set working directory
WORKDIR /app

# Install tools
RUN go install github.com/pressly/goose/v3/cmd/goose@latest
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Install system packages
RUN apt-get update && apt-get install -y \
    postgresql-client \
    netcat-traditional \
    bash

# Create necessary directories
RUN mkdir -p /app/scripts
RUN mkdir -p /app/sql

# Copy and set permissions for wait-for-it.sh
COPY ./scripts/wait-for-it.sh /app/scripts/
RUN chmod +x /app/scripts/wait-for-it.sh
COPY scripts/goose-start.sh /app/scripts/
RUN chmod +x /app/scripts/goose-start.sh

# Copy SQL files and directories
COPY ./sql /app/sql
COPY .env /app/.env
COPY sqlc.yaml /app/sqlc.yaml
