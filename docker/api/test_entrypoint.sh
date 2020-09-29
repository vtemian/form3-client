#!/usr/bin/dumb-init /bin/sh
set -e

dockerize -wait "tcp://${VAULT_HOST}:${VAULT_PORT}" -wait "tcp://${PSQL_HOST}:${SQL_PORT}"

exec /app/entrypoint.sh
