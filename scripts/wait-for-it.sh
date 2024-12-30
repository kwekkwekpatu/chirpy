#!/bin/sh
set -e

HOST="$1"
PORT="$2"

echo "wait-for-it.sh: Waiting for ${HOST}:${PORT}"

retries=10
count=0

until nc -zv "$HOST" "$PORT"; do
    if [ "$count" -ge "$retries" ]; then
        echo "wait-for-it.sh: Failed to connect after ${retries} retries"
        exit 1
    fi
    echo "wait-for-it.sh: Still waiting for database at ${HOST}:${PORT}..."
    sleep 3
    count=$((count + 1))
done

echo "wait-for-it.sh: Database is up at ${HOST}:${PORT}!"
