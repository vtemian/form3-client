#!/usr/bin/dumb-init /bin/sh
set -e

dockerize -wait "${VAULT_ADDR}" -wait "${PSQL_HOST}:${PSQL_PORT}"

exec /app/entrypoint.sh
