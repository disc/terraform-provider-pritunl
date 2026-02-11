#!/usr/bin/env bash
# Wait for Pritunl to create the default admin user in MongoDB
# This must happen before mongo.js runs, otherwise the update won't match any documents

CONTAINER=${1:-tf_pritunl_acc_test}
TIMEOUT=${2:-60}
INTERVAL=3

echo "Waiting for Pritunl admin user to be created..."

elapsed=0
while [ $elapsed -lt $TIMEOUT ]; do
    COUNT=$(docker exec "$CONTAINER" mongo pritunl --quiet --eval 'db.administrators.count()' 2>/dev/null)

    if [ "$COUNT" -gt 0 ] 2>/dev/null; then
        echo "Admin user found after ${elapsed}s"
        exit 0
    fi

    echo "Attempt $((elapsed / INTERVAL + 1)): admin count=$COUNT, waiting ${INTERVAL}s..."
    sleep $INTERVAL
    elapsed=$((elapsed + INTERVAL))
done

echo "Timeout: admin user not found after ${TIMEOUT}s"
exit 1
