#!/bin/bash

docker compose down
docker compose up -d

keycloakServer=http://localhost
url="${keycloakServer}:9000/health"
echo "Checking service availability at $url (CTRL+C to exit)"
while true; do
    response=$(curl -s -o /dev/null -w "%{http_code}" $url)
    if [ $response -eq 200 ]; then
        break
    fi
    sleep 1
done
echo "Service is now available at ${keycloakServer}:8080"

ARGS=()
if [ $# -gt 0 ]; then
    ARGS+=("-run")
    ARGS+=("^($@)$")
fi

go test -failfast -race -cover -coverprofile=coverage.out -covermode=atomic -p 10 -cpu 1,2 -parallel 1 -bench . -benchmem ${ARGS[@]}

docker compose down
