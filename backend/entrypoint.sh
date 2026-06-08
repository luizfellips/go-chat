#!/bin/sh
set -e

echo "Running migrations..."
/app/server migrate

if [ "${SEED_DEMO_USERS:-false}" = "true" ]; then
  echo "Seeding demo users (if needed)..."
  /app/server seed || true
else
  echo "Skipping seed (SEED_DEMO_USERS is not true)"
fi

echo "Starting server..."
exec /app/server
