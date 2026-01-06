#!/usr/bin/env bash
# Wait for MongoDB to be ready inside a Docker container

CONTAINER=${1:-tf_pritunl_acc_test}
TIMEOUT=${2:-60}
INTERVAL=5

echo "Waiting for MongoDB in container '$CONTAINER'..."

elapsed=0
while [ $elapsed -lt $TIMEOUT ]; do
    if docker exec "$CONTAINER" mongo --eval "db.runCommand('ping').ok" --quiet >/dev/null 2>&1; then
        echo "MongoDB is ready after ${elapsed}s"
        exit 0
    fi
    echo "Attempt $((elapsed / INTERVAL + 1)): MongoDB not ready, waiting ${INTERVAL}s..."
    sleep $INTERVAL
    elapsed=$((elapsed + INTERVAL))
done

echo "Timeout: MongoDB not ready after ${TIMEOUT}s"
exit 1
