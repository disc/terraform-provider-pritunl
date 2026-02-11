#!/usr/bin/env bash
# Wait for Pritunl API to accept authenticated requests
# After mongo.js sets auth_api=true, the settings runner (60s interval)
# must detect the change and restart pritunl-web with WEB_STRICT=false

TOKEN=${1:-tfacctest_token}
SECRET=${2:-tfacctest_secret}
URL=${3:-https://localhost/state}
TIMEOUT=${4:-120}
INTERVAL=3
PATH_PART="/state"

echo "Waiting for authenticated API access at '$URL'..."

elapsed=0
while [ $elapsed -lt $TIMEOUT ]; do
    TIMESTAMP=$(date +%s)
    NONCE=$(openssl rand -hex 16)
    AUTH_STRING="${TOKEN}&${TIMESTAMP}&${NONCE}&GET&${PATH_PART}"
    SIGNATURE=$(printf '%s' "$AUTH_STRING" | openssl dgst -sha256 -hmac "$SECRET" -binary | openssl base64 -A)

    HTTP_CODE=$(curl -sk -o /dev/null -w "%{http_code}" \
        -H "Auth-Token: $TOKEN" \
        -H "Auth-Timestamp: $TIMESTAMP" \
        -H "Auth-Nonce: $NONCE" \
        -H "Auth-Signature: $SIGNATURE" \
        -H "Content-Type: application/json" \
        "$URL" 2>/dev/null)

    if [ "$HTTP_CODE" = "200" ]; then
        echo "Authenticated API access ready after ${elapsed}s"
        exit 0
    fi

    echo "Attempt $((elapsed / INTERVAL + 1)): HTTP $HTTP_CODE, waiting ${INTERVAL}s..."
    sleep $INTERVAL
    elapsed=$((elapsed + INTERVAL))
done

echo "Timeout: authenticated API access not ready after ${TIMEOUT}s"
exit 1
