#!/bin/bash

HOST=$1
PORT=$2

until nc -z "$HOST" "$PORT"; do
  echo "Waiting for database..."
  sleep 1
done

echo "Database is up!"
