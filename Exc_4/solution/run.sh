#!/bin/sh

# This script builds the order service image, starts the PostgreSQL database
# and then starts the order service. Both containers use the host network,
# so they can reach each other via 127.0.0.1 (DB_HOST in debug.env).
# All configuration is read from debug.env.

# Stop the script as soon as a command fails.
set -e

# docker build
# Builds the image from the Dockerfile in this folder and names it "orderservice".
docker build -t orderservice .

# Remove containers from a previous run, so the names can be reused.
# "|| true" ignores the error if the containers do not exist yet.
docker rm -f orderservice postgres-db 2>/dev/null || true

# docker run db
# The named volume "postgres-data" keeps the database files outside of the
# container, so the data survives when the container is removed.
# Docker creates the volume automatically if it does not exist yet.
# The postgres image reads POSTGRES_DB, POSTGRES_USER and POSTGRES_PASSWORD
# from debug.env to create the database and user, and PGPORT for the port.
docker run -d \
  --name postgres-db \
  --network host \
  --env-file debug.env \
  -v postgres-data:/var/lib/postgresql/18/docker \
  postgres:18

# Wait until the database accepts connections, otherwise the order service
# would fail to connect on startup. pg_isready is part of the postgres image.
echo "Waiting for database..."
until docker exec postgres-db sh -c 'pg_isready -h 127.0.0.1 -p "$PGPORT" -U "$POSTGRES_USER"' >/dev/null 2>&1; do
  sleep 1
done

# docker run orderservice
# The order service reads the same environment variables to build its
# connection string. With the host network, the server is reachable at
# http://localhost:3000 without any port mapping.
docker run -d \
  --name orderservice \
  --network host \
  --env-file debug.env \
  orderservice

echo "Order service is running on http://localhost:3000"
