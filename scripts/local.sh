#!/bin/bash

set -e

function setup() {
  echo "Starting local setup for newsletter service..."

  # Network for containers (optional)
  docker network create newsletter-net || true

  # Start Postgres container
  echo "Starting PostgreSQL..."
  docker run -d --name newsletter-postgres \
    --network newsletter-net \
    -e POSTGRES_USER=postgres \
    -e POSTGRES_PASSWORD=postgres \
    -e POSTGRES_DB=newsletter_db \
    -p 5432:5432 \
    postgres:15-alpine

  echo "PostgreSQL started on port 5432"

  # Start Redis container
  echo "Starting Redis..."
  docker run -d --name newsletter-redis \
    --network newsletter-net \
    -p 6379:6379 \
    redis:7-alpine

  echo "Redis started on port 6379"

  # Wait for Postgres & Redis to be ready
  echo "Waiting for Postgres and Redis to be ready..."
  sleep 10

  # Create DB schema (optional, can use migrations instead)
  echo "Setting up DB schema..."
  docker exec -i newsletter-postgres psql -U postgres -d newsletter_db <<EOF
CREATE TABLE IF NOT EXISTS subscribers (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    subscribed_topics TEXT[]
);

CREATE TABLE IF NOT EXISTS content (
    id SERIAL PRIMARY KEY,
    topic VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    scheduled_time TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS topics (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL
);
EOF

  echo "Database schema created."

  echo "Local environment is ready."
  echo "Use these connection details in your app:"
  echo "Postgres: postgres://postgres:postgres@localhost:5432/newsletter_db?sslmode=disable"
  echo "Redis: redis://localhost:6379"

  echo "Building and running backend server container..."
  docker build -f Dockerfile.web -t newsletter-web .
  docker run -d --name newsletter-web --network newsletter-net -p 8080:8080 newsletter-web

  echo "Building and running worker container..."
  docker build -f Dockerfile.worker -t newsletter-worker .
  docker run -d --name newsletter-worker --network newsletter-net newsletter-worker

  echo "Backend server running on http://localhost:8080"
  echo "To stop containers, run: ./scripts/local.sh clean"
}

function clean() {
  echo "Stopping and removing containers..."
  docker stop newsletter-postgres newsletter-redis || true
  docker rm newsletter-postgres newsletter-redis || true
  echo "Local environment cleaned up."
}

if [[ "$1" == "setup" ]]; then
  setup
elif [[ "$1" == "clean" ]]; then
  clean
else
  echo "Usage: $0 {setup|clean}"
  exit 1
fi
