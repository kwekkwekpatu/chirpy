# Build stage
FROM golang:1.23.1 AS builder

# Install tools
RUN go install github.com/pressly/goose/v3/cmd/goose@latest
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
RUN apt-get update && apt-get install -y postgresql-client

# Run stage
FROM debian:stable-slim

# Create necessary directories
RUN mkdir -p /public
RUN mkdir -p /app/scripts
RUN mkdir -p /app/sql

# Copy and set permissions for wait-for-it.sh
COPY ./scripts/wait-for-it.sh /app/scripts/
RUN chmod +x /app/scripts/wait-for-it.sh

# Copy SQL files and directories
COPY ./sql /app/sql

# Copy goose from the builder stage
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY --from=builder /go/bin/sqlc /usr/local/bin/sqlc
COPY sqlc.yaml /app/sqlc.yaml

# Copy your application files
COPY chirpy /app/chirpy
COPY index.html /public/index.html
COPY .env /app/

RUN chmod +x /app/chirpy

# Expose the port the app runs on
EXPOSE 8080
ENV PORT=8080

# Set chirpy as the entrypoint
ENTRYPOINT ["/app/chirpy"]
CMD []
