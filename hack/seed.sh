#!/usr/bin/env bash

HOST=${TEST_API_HOST:-"http://localhost:8080"}

echo "Deleting all accounts"
for id in $(curl "${HOST}/v1/organisation/accounts" | jq ".data[] | .id"); do
  curl -X DELETE "${HOST}/v1/organisation/accounts/${id}?version=0" || true;
done;

# Importing only account fixtures
for fixture in ./fixtures/fetch_*.json; do
  echo "Importing $fixture"

  curl -X POST "${HOST}/v1/organisation/accounts" \
     --header "Content-Type: application/json" \
     --data-raw "$(cat $fixture)" || true;
done;
