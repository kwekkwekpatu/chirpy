#!/bin/sh
echo "Starting goose service..."
pwd
ls -la /app/sql/schema
goose -v -dir /app/sql/schema postgres "${DB_URL}" up
