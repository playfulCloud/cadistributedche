#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"
CACHE_KEY="manual:$(date +%s)"
TTL_KEY="${CACHE_KEY}:ttl"
MISSING_KEY="${CACHE_KEY}:missing"

request() {
  local method="$1"
  local path="$2"
  local body="${3:-}"
  local expected_status="$4"

  local response_file
  response_file="$(mktemp)"

  local status
  if [[ -n "$body" ]]; then
    status="$(curl -sS -o "$response_file" -w "%{http_code}" -X "$method" "$BASE_URL$path" --data "$body")"
  else
    status="$(curl -sS -o "$response_file" -w "%{http_code}" -X "$method" "$BASE_URL$path")"
  fi

  printf "\n%s %s -> %s\n" "$method" "$path" "$status"
  cat "$response_file"

  if [[ "$status" != "$expected_status" ]]; then
    printf "\nexpected status %s, got %s\n" "$expected_status" "$status" >&2
    rm -f "$response_file"
    exit 1
  fi

  rm -f "$response_file"
}

printf "Running manual cache API checks against %s\n" "$BASE_URL"
printf "Start the server first with: go run ./cmd/cadistributedche\n"

request PUT "/cache/$CACHE_KEY" "first-value" "201"
request GET "/cache/$CACHE_KEY" "" "200"
request PUT "/cache/$CACHE_KEY" "second-value" "200"
request GET "/cache/$CACHE_KEY" "" "200"
request DELETE "/cache/$CACHE_KEY" "" "204"
request GET "/cache/$CACHE_KEY" "" "404"

request GET "/cache/$MISSING_KEY" "" "404"
request PUT "/cache/$TTL_KEY?ttl=2s" "expires-soon" "201"
request GET "/cache/$TTL_KEY" "" "200"

printf "\nWaiting for TTL expiration...\n"
sleep 3

request GET "/cache/$TTL_KEY" "" "404"
request PUT "/cache/$CACHE_KEY?ttl=invalid" "bad-ttl" "400"


printf "\n Metrics: \n"
request GET "/cache/metrics" "" "200"

printf "\nManual cache API checks completed successfully.\n"
