#!/usr/bin/env bash
# Wait for Pritunl API to be ready

URL=${1:-https://localhost/state}
TIMEOUT=${2:-60}
INTERVAL=5

echo "Waiting for Pritunl API at '$URL'..."

elapsed=0
while [ $elapsed -lt $TIMEOUT ]; do
    if curl -sk "$URL" >/dev/null 2>&1; then
        echo "Pritunl API is ready after ${elapsed}s"
        exit 0
    fi
    echo "Attempt $((elapsed / INTERVAL + 1)): API not ready, waiting ${INTERVAL}s..."
    sleep $INTERVAL
    elapsed=$((elapsed + INTERVAL))
done

echo "Timeout: Pritunl API not ready after ${TIMEOUT}s"
exit 1
