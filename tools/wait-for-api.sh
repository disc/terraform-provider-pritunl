#!/usr/bin/env bash
# Wait for Pritunl API to be ready
#
# A reachable HTTPS port is not enough to tell the API apart from a Pritunl that
# is still starting: pritunl-web serves that port as soon as it comes up and
# answers every request on its own, with a 401 "Missing token" or "Not
# initialized", until the Pritunl backend behind it has handed it the web
# secret, and with a 502 while that backend is restarting. The API is only
# usable once the request makes it through to the backend, which answers an
# unauthenticated one with a 401 of its own.

URL=${1:-https://localhost/state}
TIMEOUT=${2:-60}
INTERVAL=5

echo "Waiting for Pritunl API at '$URL'..."

elapsed=0
while [ $elapsed -lt $TIMEOUT ]; do
    # the status code is appended to the body, an unreachable endpoint leaves
    # both of them empty
    response=$(curl -sk --max-time "$INTERVAL" -w '%{http_code}' "$URL" 2>/dev/null)

    case "$response" in
        *"Missing token"*|*"Not initialized"*|*502)
            ;;
        ?*)
            echo "Pritunl API is ready after ${elapsed}s"
            exit 0
            ;;
    esac

    echo "Attempt $((elapsed / INTERVAL + 1)): API not ready, waiting ${INTERVAL}s..."
    sleep $INTERVAL
    elapsed=$((elapsed + INTERVAL))
done

echo "Timeout: Pritunl API not ready after ${TIMEOUT}s"
exit 1
