#!/usr/bin/env bash

HOST=${HOST:-"localhost:8080"}

# Importing only account fixtures
for fixture in ./fixtures/fetch_*.json; do
  echo "Importing $fixture"

  curl -s -X POST "http://${HOST}/v1/organisation/accounts" \
     --header "Content-Type: application/json" \
     --data-raw "$(cat $fixture)" > /dev/null || true;
done;