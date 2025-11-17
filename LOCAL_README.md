# Newsletter Service - Local Development

A comprehensive guide for setting up and running the newsletter service locally for development purposes.

## 🏠 Local Development Setup

This guide covers setting up the newsletter service on your local machine for development, testing, and debugging.

### Prerequisites

- **Go 1.22+**: [Download and install Go](https://golang.org/dl/)
- **Docker & Docker Compose**: [Install Docker](https://docs.docker.com/get-docker/)
- **PostgreSQL**: Either via Docker or local installation
- **Redis**: Either via Docker or local installation
- **Git**: For version control

### 🚀 Quick Start (Automated Setup)

The fastest way to get started is using our automated setup script:

The fastest way to get started is using our automated setup script. The script works on both Linux/Unix and Windows systems:

**For Linux/Mac/Unix systems:**

```bash
# Clone the repository
git clone https://github.com/your-username/newsletter-service.git
cd newsletter-service

# Make setup script executable
chmod +x scripts/local.sh

# Start everything (PostgreSQL, Redis, migrations, web, worker)
./scripts/local.sh setup

# When done, clean up
./scripts/local.sh clean
```

**For Windows systems:**

_Prerequisites: Git for Windows must be installed (includes Git Bash)_

```bash
# Open Git Bash (not Command Prompt or PowerShell)
# Navigate to your project directory
cd /d/Projects/newsletter-service

# Make setup script executable
chmod +x scripts/local.sh

# Start everything (PostgreSQL, Redis, migrations, web, worker)
./scripts/local.sh setup

# When done, clean up
./scripts/local.sh clean
```

**Alternative Windows method (if Git Bash is not preferred):**

If you prefer to use Command Prompt or PowerShell, you can run the bash script through Git Bash programmatically:

**Command Prompt:**

```cmd
# Setup
"<path_to_git>\bin\bash.exe" -c "./scripts/local.sh setup"

# Clean up
"<path_to_git>\bin\bash.exe" -c "./scripts/local.sh clean"
```

**PowerShell:**

```powershell
# Setup (use & operator for PowerShell)
& "<path_to_git>\bin\bash.exe" -c "./scripts/local.sh setup"

# Clean up
& "<path_to_git>\bin\bash.exe" -c "./scripts/local.sh clean"
```

**Note:** Adjust the Git installation path if different on your system. Find your Git path with `where git` and replace `cmd\git.exe` with `bin\bash.exe`.

**What this does:**

- Starts PostgreSQL container on port 5432
- Starts Redis container on port 6379
- Runs database migrations automatically
- Builds and starts web server on port 8080
- Builds and starts background worker
- Sets up Docker network for inter-service communication

**📝 Note for Windows users:** Replace `<path_to_git>` with your actual Git installation path. To find it:

```cmd
# In Command Prompt or PowerShell
where git
# Look for the git.exe path, then replace 'cmd\git.exe' with 'bin\bash.exe'
# Example: If git is at 'D:\Program Files\Git\cmd\git.exe'
# Then use: 'D:\Program Files\Git\bin\bash.exe'
# Common locations: 'C:\Program Files\Git' or 'C:\Program Files (x86)\Git'
```

### 🔧 Manual Setup (Step by Step)

If you prefer manual control or need to customize the setup:

#### 1. Start Infrastructure Services

**Option A: Using Docker Compose (Recommended)**

```bash
# Start PostgreSQL and Redis
docker-compose up -d postgres redis

# Verify services are running
docker-compose ps
```

**Option B: Individual Docker Containers**

```bash
# Create network
docker network create newsletter-net

# Start PostgreSQL
docker run -d --name newsletter-postgres \
  --network newsletter-net \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=newsletter_db \
  -p 5432:5432 \
  postgres:15-alpine

# Start Redis
docker run -d --name newsletter-redis \
  --network newsletter-net \
  -p 6379:6379 \
  redis:7-alpine
```

**Option C: Local Installation**

```bash
# Install and start PostgreSQL (Ubuntu/Debian)
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql

# Install and start Redis
sudo apt install redis-server
sudo systemctl start redis-server

# Create database
sudo -u postgres createdb newsletter_db
```

#### 2. Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install Goose for migrations
go install github.com/pressly/goose/v3/cmd/goose@latest

# Verify installation
goose --version
```

#### 3. Configure Environment

**Option A: Use default configuration**

```bash
# Copy default config (already configured for local development)
cp env/default.toml env/local.toml

# Default settings work for Docker setup:
# - Database: localhost:5432/newsletter_db
# - Redis: localhost:6379
# - Auth: admin/changeme
# - Scheduler: scheduler/scheduler123
```

**Option B: Environment variables**

```bash
# Set environment variables (optional override)
export DATABASE_HOST=localhost
export DATABASE_USER=postgres
export DATABASE_PASSWORD=postgres
export DATABASE_NAME=newsletter_db

export REDIS_HOST=localhost
export REDIS_PORT=6379

export AUTH_USERNAME=admin
export AUTH_PASSWORD=changeme

export SCHEDULER_USERNAME=scheduler
export SCHEDULER_PASSWORD=scheduler123
export SCHEDULER_ENABLED=true

export RATE_LIMIT_ENABLED=true
export RATE_LIMIT_STORAGE=redis
```

#### 4. Run Database Migrations

**Option A: Using local script**

**Linux/Mac/Unix:**

```bash
# Run migrations
./scripts/local.sh migrate up

# Check migration status
./scripts/local.sh migrate status

# Rollback if needed
./scripts/local.sh migrate down
```

**Windows (Git Bash):**

```bash
# Open Git Bash and run:
./scripts/local.sh migrate up

# Check migration status
./scripts/local.sh migrate status

# Rollback if needed
./scripts/local.sh migrate down
```

**Windows (Command Prompt/PowerShell):**

```cmd
# Command Prompt:
"<path_to_git>\bin\bash.exe" -c "./scripts/local.sh migrate up"
"<path_to_git>\bin\bash.exe" -c "./scripts/local.sh migrate status"
"<path_to_git>\bin\bash.exe" -c "./scripts/local.sh migrate down"
```

```powershell
# PowerShell (use & operator):
& "<path_to_git>\bin\bash.exe" -c "./scripts/local.sh migrate up"
& "<path_to_git>\bin\bash.exe" -c "./scripts/local.sh migrate status"
& "<path_to_git>\bin\bash.exe" -c "./scripts/local.sh migrate down"
```

**Option B: Direct Goose commands**

```bash
# Run all pending migrations
goose -dir ./migration/sql postgres "postgres://postgres:postgres@localhost:5432/newsletter_db?sslmode=disable" up

# Check status
goose -dir ./migration/sql postgres "postgres://postgres:postgres@localhost:5432/newsletter_db?sslmode=disable" status
```

**Option C: Using Go wrapper**

```bash
# Run migrations using Go command
go run cmd/migration/migration.go up
```

#### 5. Start Services

**Development mode (single terminal):**

```bash
# Start web server
go run cmd/web/main.go

# In another terminal, start worker
go run cmd/worker/main.go
```

**Production-like mode (using Docker):**

```bash
# Build and run web service
docker build -f scripts/Dockerfile.web -t newsletter-web .
docker run -d --name newsletter-web --network newsletter-net -p 8080:8080 newsletter-web

# Build and run worker service
docker build -f scripts/Dockerfile.worker -t newsletter-worker .
docker run -d --name newsletter-worker --network newsletter-net newsletter-worker
```

### 🧪 Testing the Setup

#### 1. Health Checks

```bash
# Test main service health
curl http://localhost:8080/health

# Test scheduler health (with auth)
curl -u scheduler:scheduler123 http://localhost:8080/scheduler/v1/health
```

#### 2. API Testing

```bash
# Create a topic
curl -u admin:changeme -X POST http://localhost:8080/api/v1/topics \
  -H "Content-Type: application/json" \
  -d '{"name":"Tech News","description":"Technology updates"}'

# List topics
curl -u admin:changeme http://localhost:8080/api/v1/topics

# Create a subscriber
curl -u admin:changeme -X POST http://localhost:8080/api/v1/subscribers \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","name":"Test User"}'

# List subscribers
curl -u admin:changeme http://localhost:8080/api/v1/subscribers
```

#### 3. Rate Limiting Test

```bash
# Test rate limiting (should work first few times, then get blocked)
for i in {1..20}; do
  curl -u admin:changeme http://localhost:8080/api/v1/topics
  echo "Request $i"
  sleep 0.1
done
```

### 🛠️ Development Tools

#### Database Management

```bash
# Connect to PostgreSQL
docker exec -it newsletter-postgres psql -U postgres -d newsletter_db

# Or if using local PostgreSQL
psql -U postgres -d newsletter_db

# Common SQL commands
\dt              # List tables
\d topics        # Describe topics table
SELECT * FROM topics;
```

#### Redis Management

```bash
# Connect to Redis
docker exec -it newsletter-redis redis-cli

# Or if using local Redis
redis-cli

# Common Redis commands
KEYS *                     # List all keys
GET rate_limit:ip:127.0.0.1  # Check rate limit bucket
FLUSHALL                   # Clear all data (dev only!)
```

#### Log Monitoring

```bash
# Watch application logs
go run cmd/web/main.go 2>&1 | tee app.log

# Watch Docker container logs
docker logs -f newsletter-web
docker logs -f newsletter-worker

# Watch all infrastructure logs
docker-compose logs -f
```

### 🔄 Development Workflow

#### Making Changes

1. **Code Changes**: Edit source files
2. **Restart Service**:

   ```bash
   # Kill running process (Ctrl+C)
   # Restart
   go run cmd/web/main.go
   ```

3. **Database Changes**:

   ```bash
   # Create new migration
   goose -dir ./migration/sql create add_new_field sql

   # Edit the generated file
   # Run migration
   ./scripts/local.sh migrate up
   ```

4. **Configuration Changes**: Edit `env/default.toml` and restart

#### Debugging

```bash
# Run with verbose logging
export GIN_MODE=debug
go run cmd/web/main.go

# Use Go debugger
go install github.com/go-delve/delve/cmd/dlv@latest
dlv debug cmd/web/main.go
```

#### Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/services/topic/...
```

### 🧹 Cleanup Commands

```bash
# Stop and remove all containers (using script)
./scripts/local.sh clean

# Manual cleanup
docker stop newsletter-postgres newsletter-redis newsletter-web newsletter-worker
docker rm newsletter-postgres newsletter-redis newsletter-web newsletter-worker
docker network rm newsletter-net

# Clean up Docker images
docker rmi newsletter-web newsletter-worker

# Reset database (if needed)
docker volume rm newsletter-service_postgres-data
```

### 📁 Project Structure

```
newsletter-service/
├── cmd/
│   ├── web/main.go           # Web server entry point
│   ├── worker/main.go        # Background worker
│   └── migration/main.go     # Migration tool
├── internal/
│   ├── config/               # Configuration management
│   ├── database/             # Database connection
│   ├── services/             # Business logic
│   ├── handlers/             # HTTP handlers (separated by domain)
│   ├── router/               # Route definitions + middleware
│   └── dtos/                 # Data transfer objects
├── migration/
│   └── sql/                  # Database migration files
├── scripts/
│   ├── local.sh              # Local development script
│   ├── Dockerfile.web        # Web service container
│   └── Dockerfile.worker     # Worker service container
├── env/
│   └── default.toml          # Default configuration
├── go.mod                    # Go dependencies
└── Procfile                  # Heroku deployment config
```

### 🔧 Configuration Reference

#### env/default.toml

```toml
environment = "development"

[auth]
username = "admin"
password = "changeme"

[scheduler]
username = "scheduler"
password = "scheduler123"
enabled = true

[database]
host = "localhost"
port = 5432
user = "postgres"
password = "postgres"
name = "newsletter_db"
sslmode = "disable"
auto_migrate = false

[redis]
host = "localhost"
port = 6379
password = ""
db = 0

[rate_limit]
enabled = true
storage = "redis"

[rate_limit.default]
enabled = true
bucket_size = 100
refill_size = 10
refill_duration = "1m"
identify_by = "ip"
```

### 🚨 Common Issues & Solutions

#### 1. Port Already in Use

```bash
# Find process using port 8080
lsof -ti:8080

# Kill process
kill -9 $(lsof -ti:8080)

# Or use different port
PORT=8081 go run cmd/web/main.go
```

#### 2. Database Connection Failed

```bash
# Check if PostgreSQL container is running
docker ps | grep postgres

# Check connectivity
telnet localhost 5432

# Reset database container
docker stop newsletter-postgres
docker rm newsletter-postgres
./scripts/local.sh setup
```

#### 3. Redis Connection Failed

```bash
# Check if Redis container is running
docker ps | grep redis

# Test Redis connectivity
redis-cli ping

# Fallback to memory storage
export RATE_LIMIT_STORAGE=memory
```

#### 4. Migration Errors

```bash
# Check current migration status
./scripts/local.sh migrate status

# Reset and re-run migrations
./scripts/local.sh migrate reset
./scripts/local.sh migrate up

# Manual SQL fix (if needed)
docker exec -it newsletter-postgres psql -U postgres -d newsletter_db
```

### 📚 Additional Resources

- **API Documentation**: See [OpenAPI Specification](./api-docs.yaml) for complete API reference
- **Enterprise Features**: See `ENTERPRISE_FEATURES.md` for advanced features
- **Production Deployment**: See main `README.md` for Heroku deployment
- **Go Documentation**: [golang.org/doc](https://golang.org/doc/)
- **Gin Framework**: [gin-gonic.com](https://gin-gonic.com/)
- **GORM**: [gorm.io](https://gorm.io/)
- **Goose Migrations**: [github.com/pressly/goose](https://github.com/pressly/goose)
