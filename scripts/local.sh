#!/bin/bash

set -e

MIGRATION_DIR="./migration/sql"
DB_URL="postgres://postgres:postgres@localhost:5432/newsletter_db?sslmode=disable"

function run_migration() {
  local command=${1:-up}
  
  # Override with environment variable if set
  if [ ! -z "$DATABASE_URL" ]; then
    DB_URL="$DATABASE_URL"
  fi

  # Check if Goose is installed
  if ! command -v goose &> /dev/null; then
    echo "Goose is not installed. Installing..."
    go install github.com/pressly/goose/v3/cmd/goose@latest
  fi

  echo "Running migration: $command"
  echo "Database: $DB_URL"
  echo "Migration dir: $MIGRATION_DIR"

  case $command in
    up)
      goose -dir "$MIGRATION_DIR" postgres "$DB_URL" up
      ;;
    down)
      goose -dir "$MIGRATION_DIR" postgres "$DB_URL" down
      ;;
    status)
      goose -dir "$MIGRATION_DIR" postgres "$DB_URL" status
      ;;
    version)
      goose -dir "$MIGRATION_DIR" postgres "$DB_URL" version
      ;;
    reset)
      goose -dir "$MIGRATION_DIR" postgres "$DB_URL" reset
      ;;
    *)
      echo "Migration usage: $0 migrate [up|down|status|version|reset]"
      exit 1
      ;;
  esac

  echo "Migration completed successfully!"
}

function setup() {
  echo "Starting local setup for newsletter service..."

  # Network for containers (optional)
  docker network create newsletter-net 2>/dev/null || true

  # Start Postgres container
  echo "Starting PostgreSQL..."
  if ! docker ps -q -f name=newsletter-postgres | grep -q .; then
    docker run -d --name newsletter-postgres \
      --network newsletter-net \
      -e POSTGRES_USER=postgres \
      -e POSTGRES_PASSWORD=postgres \
      -e POSTGRES_DB=newsletter_db \
      -p 5432:5432 \
      postgres:15-alpine
    echo "PostgreSQL started on port 5432"
  else
    echo "PostgreSQL container already running"
  fi

  # Start Redis container
  echo "Starting Redis..."
  if ! docker ps -q -f name=newsletter-redis | grep -q .; then
    docker run -d --name newsletter-redis \
      --network newsletter-net \
      -p 6379:6379 \
      redis:7-alpine
    echo "Redis started on port 6379"
  else
    echo "Redis container already running"
  fi

  echo "Waiting for Postgres and Redis to be ready..."
  sleep 10

  # Run migrations using Goose directly
  echo "Running database migrations..."
  run_migration up

  echo "Building and running backend server container..."
  docker build -f scripts/Dockerfile.web -t newsletter-web .
  docker run -d --name newsletter-web --network newsletter-net \
    -e DATABASE_HOST=newsletter-postgres \
    -e REDIS_HOST=newsletter-redis \
    -p 8080:8080 newsletter-web

  echo "Building and running worker container..."
  docker build -f scripts/Dockerfile.worker -t newsletter-worker .
  docker run -d --name newsletter-worker --network newsletter-net \
    -e DATABASE_HOST=newsletter-postgres \
    -e REDIS_HOST=newsletter-redis \
    newsletter-worker

  echo "Backend server running on http://localhost:8080"
  echo "To stop containers, run: ./scripts/local.sh clean"
  echo "To run migrations, use: ./scripts/local.sh migrate [up|down|status|version|reset]"
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
elif [[ "$1" == "migrate" ]]; then
  run_migration "$2"
else
  echo "Usage: $0 {setup|clean|migrate [up|down|status|version|reset]}"
  echo ""
  echo "Commands:"
  echo "  setup   - Start local development environment with containers and run migrations"
  echo "  clean   - Stop and remove all containers"
  echo "  migrate - Run database migrations (default: up)"
  echo ""
  echo "Migration examples:"
  echo "  $0 migrate up      # Run all pending migrations"
  echo "  $0 migrate down    # Rollback last migration"
  echo "  $0 migrate status  # Check migration status"
  exit 1
fi
