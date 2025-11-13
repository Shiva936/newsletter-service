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

  echo "Waiting for Postgres and Redis to be ready..."
  sleep 10

  # Run migration container
  echo "Running migration container to setup DB schema..."
  docker build -f scripts/Dockerfile.migration -t newsletter-service-migration .
  docker run --rm --name newsletter-migration --network newsletter-net -e DB_STRING="postgres://postgres:postgres@newsletter-postgres:5432/newsletter_db?sslmode=disable" newsletter-service-migration

  echo "Migration completed."

  echo "Building and running backend server container..."
  docker build -f scripts/Dockerfile.web -t newsletter-web .
  docker run -d --name newsletter-web --network newsletter-net -p 8080:8080 newsletter-web

  echo "Building and running worker container..."
  docker build -f scripts/Dockerfile.worker -t newsletter-worker .
  docker run -d --name newsletter-worker --network newsletter-net newsletter-worker

  echo "Backend server running on http://localhost:8080"
  echo "To stop containers, run: ./scripts/local.sh clean"
}

function clean() {
  echo "Stopping and removing containers..."
  docker stop newsletter-postgres newsletter-redis newsletter-web newsletter-worker || true
  docker rm newsletter-postgres newsletter-redis newsletter-web newsletter-worker || true
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
