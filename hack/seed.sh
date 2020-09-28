#!/usr/bin/env bash

HOST=${TEST_API_HOST:-"http://localhost:8080"}

echo "Deleting all accounts"
for id in $(curl -s "${HOST}/v1/organisation/accounts" | jq ".data[] | .id"); do
  curl -s -X DELETE "${HOST}/v1/organisation/accounts/${id}?version=0" > /dev/null || true;
done;

# Importing only account fixtures
for fixture in ./fixtures/fetch_*.json; do
  echo "Importing $fixture"

  curl -s -X POST "${HOST}/v1/organisation/accounts" \
     --header "Content-Type: application/json" \
     --data-raw "$(cat $fixture)" > /dev/null || true;
done;
