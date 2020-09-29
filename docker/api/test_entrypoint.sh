#!/usr/bin/dumb-init /bin/sh
set -e

dockerize -wait "${VAULT_ADDR}" -wait "tcp://${PSQL_HOST}:${SQL_PORT}"

exec /app/entrypoint.sh
