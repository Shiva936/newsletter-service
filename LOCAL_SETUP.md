# Local Setup Guide

Complete guide to set up Newsletter Service on your local development environment using the automated setup script.

## 🎯 **Quick Start**

The fastest way to get Newsletter Service running locally is using our automated setup script:

```bash
# Clone the repository
git clone <repository-url>
cd newsletter-service

# Make setup script executable and run
chmod +x scripts/local.sh
./scripts/local.sh
```

**That's it!** The script handles everything automatically. Continue reading for detailed explanations and manual setup options.

## 📋 **Prerequisites**

### **Required Software**
- **Docker**: [Install Docker Desktop](https://www.docker.com/products/docker-desktop/)
- **Docker Compose**: Included with Docker Desktop
- **Git**: For cloning the repository
- **curl**: For testing APIs (usually pre-installed)

### **Optional (for manual development)**
- **Go 1.21+**: [Install Go](https://golang.org/dl/)
- **PostgreSQL 13+**: [Install PostgreSQL](https://www.postgresql.org/download/)
- **Redis 6+**: [Install Redis](https://redis.io/download)

### **System Requirements**
- **RAM**: 4GB minimum, 8GB recommended
- **Storage**: 2GB free space
- **CPU**: 2+ cores recommended
- **OS**: Linux, macOS, or Windows with Docker support

## 🚀 **Automated Setup with local.sh**

### **What the Script Does**

The `scripts/local.sh` script automates the entire setup process:

1. **🐳 Creates Docker network** for service communication
2. **🗄️ Starts PostgreSQL database** with proper configuration
3. **🔄 Starts Redis cache** for session management
4. **📦 Builds application containers** (web API and worker)
5. **🔧 Runs database migrations** to create schema
6. **✅ Verifies all services** are healthy and running
7. **📊 Provides connection details** and test commands

### **Script Execution**

```bash
# Navigate to project directory
cd newsletter-service

# Make script executable
chmod +x scripts/local.sh

# Run automated setup
./scripts/local.sh
```

### **Expected Output**

```bash
🚀 Starting Newsletter Service Local Development Setup...

✅ Docker network 'newsletter-network' created successfully
✅ PostgreSQL container started successfully
✅ Redis container started successfully  
✅ Web API container started successfully
✅ Worker container started successfully
✅ Database migrations completed successfully
✅ All services are healthy!

🌟 Newsletter Service is ready!

📊 Service Information:
- Web API: http://localhost:8080
- PostgreSQL: localhost:5432
- Redis: localhost:6379
- Database: newsletter_db
- Username: postgres / Password: postgres

🧪 Test the setup:
curl http://localhost:8080/health
```

## 🔍 **What Gets Created**

### **Docker Containers**

1. **newsletter-db** (PostgreSQL)
   - **Image**: `postgres:13`
   - **Port**: `5432:5432`
   - **Database**: `newsletter_db`
   - **Credentials**: `postgres/postgres`

2. **newsletter-redis** (Redis)
   - **Image**: `redis:6-alpine`
   - **Port**: `6379:6379`
   - **Configuration**: Default Redis settings

3. **newsletter-web** (Web API)
   - **Build**: Custom Go application
   - **Port**: `8080:8080`
   - **Features**: REST API, authentication, rate limiting

4. **newsletter-worker** (Background Worker)
   - **Build**: Custom Go application
   - **Features**: Scheduled processing, email delivery

### **Docker Network**

- **Name**: `newsletter-network`
- **Purpose**: Allows containers to communicate securely
- **Type**: Bridge network for local development

### **Database Schema**

The migration automatically creates:

```sql
-- Topics for organizing content
topics: id, name, description, created_at, updated_at

-- Subscribers with email addresses
subscribers: id, name, email, is_active, created_at, updated_at

-- Newsletter content with scheduling
contents: id, title, body, topic_id, scheduled_time, status, created_at

-- Many-to-many topic subscriptions
subscriptions: id, subscriber_id, topic_id, created_at

-- Email delivery tracking
email_logs: id, subscriber_id, content_id, email_address, status, sent_at, error_message
```

## ✅ **Verification Steps**

### **1. Check Service Health**

```bash
# Basic health check
curl http://localhost:8080/health

# Expected response:
{"status":"healthy","timestamp":"2025-11-17T10:00:00Z"}
```

### **2. Verify Database Connection**

```bash
# Detailed health check
curl http://localhost:8080/health/detailed

# Expected response includes database status
{
  "status": "healthy",
  "database": "connected",
  "redis": "connected",
  "timestamp": "2025-11-17T10:00:00Z"
}
```

### **3. Test Container Status**

```bash
# Check all containers are running
docker ps --format \"table {{.Names}}\\t{{.Status}}\\t{{.Ports}}\"

# Expected output:
NAMES               STATUS              PORTS
newsletter-web      Up X minutes        0.0.0.0:8080->8080/tcp
newsletter-worker   Up X minutes        
newsletter-redis    Up X minutes        0.0.0.0:6379->6379/tcp
newsletter-db       Up X minutes        0.0.0.0:5432->5432/tcp
```

## 📚 **API Testing**

### **Create Your First Newsletter**

#### **1. Create a Topic**
```bash
curl -X POST http://localhost:8080/topics \\
  -H \"Content-Type: application/json\" \\
  -d '{
    \"name\": \"Technology\",
    \"description\": \"Latest technology news and updates\"
  }'

# Response:
{
  \"id\": 1,
  \"name\": \"Technology\", 
  \"description\": \"Latest technology news and updates\",
  \"created_at\": \"2025-11-17T10:00:00Z\",
  \"updated_at\": \"2025-11-17T10:00:00Z\"
}
```

#### **2. Add a Subscriber**
```bash
curl -X POST http://localhost:8080/subscribers \\
  -H \"Content-Type: application/json\" \\
  -d '{
    \"name\": \"John Doe\",
    \"email\": \"john@example.com\",
    \"subscribed_topics\": [\"Technology\"]
  }'

# Response:
{
  \"id\": 1,
  \"name\": \"John Doe\",
  \"email\": \"john@example.com\",
  \"is_active\": true,
  \"subscribed_topics\": [\"Technology\"],
  \"created_at\": \"2025-11-17T10:00:00Z\",
  \"updated_at\": \"2025-11-17T10:00:00Z\"
}
```

#### **3. Create Newsletter Content**
```bash
curl -X POST http://localhost:8080/contents \\
  -H \"Content-Type: application/json\" \\
  -d '{
    \"title\": \"Weekly Tech Update\",
    \"body\": \"This week in technology: AI advances, new frameworks, and industry insights.\",
    \"topic_id\": 1,
    \"scheduled_time\": \"2025-11-18T10:00:00Z\"
  }'

# Response:
{
  \"id\": 1,
  \"title\": \"Weekly Tech Update\",
  \"body\": \"This week in technology...\",
  \"topic_id\": 1,
  \"scheduled_time\": \"2025-11-18T10:00:00Z\",
  \"status\": \"pending\",
  \"created_at\": \"2025-11-17T10:00:00Z\"
}
```

#### **4. Test Email Delivery (Manual Trigger)**
```bash
# First, encode scheduler credentials (scheduler:scheduler123)
echo -n \"scheduler:scheduler123\" | base64
# Result: c2NoZWR1bGVyOnNjaGVkdWxlcjEyMw==

# Trigger notification manually
curl -X POST http://localhost:8080/notifications/send \\
  -H \"Content-Type: application/json\" \\
  -H \"Authorization: Basic c2NoZWR1bGVyOnNjaGVkdWxlcjEyMw==\" \\
  -d '{\"content_id\": 1}'

# Response:
{\"message\": \"Notifications sent successfully\"}
```

#### **5. Check Email Logs**
```bash
curl http://localhost:8080/notifications/logs

# Response:
[
  {
    \"id\": 1,
    \"subscriber_id\": 1,
    \"content_id\": 1,
    \"email_address\": \"john@example.com\",
    \"status\": \"sent\",
    \"sent_at\": \"2025-11-17T10:05:00Z\"
  }
]
```

## 🔧 **Development Workflow**

### **Making Code Changes**

1. **Edit source code** in your preferred editor
2. **Rebuild containers** to apply changes:
   ```bash
   docker-compose down
   docker-compose up --build -d
   ```

3. **Check logs** for any issues:
   ```bash
   docker logs newsletter-web
   docker logs newsletter-worker
   ```

### **Database Operations**

#### **Access Database**
```bash
# Connect to PostgreSQL
docker exec -it newsletter-db psql -U postgres -d newsletter_db

# List tables
\\dt

# Query subscribers
SELECT * FROM subscribers;

# Exit
\\q
```

#### **Reset Database**
```bash
# Stop services
docker-compose down

# Remove database volume (CAUTION: destroys data)
docker volume rm newsletter-service_postgres_data

# Restart services (will recreate database)
./scripts/local.sh
```

### **Redis Operations**

#### **Access Redis**
```bash
# Connect to Redis CLI
docker exec -it newsletter-redis redis-cli

# Check keys
KEYS *

# Get cached data
GET some_key

# Exit
exit
```

## 📊 **Monitoring & Logs**

### **View Application Logs**

```bash
# Web API logs
docker logs newsletter-web --follow

# Worker service logs  
docker logs newsletter-worker --follow

# Database logs
docker logs newsletter-db

# Redis logs
docker logs newsletter-redis
```

### **Health Monitoring**

```bash
# Check provider health
curl http://localhost:8080/providers/health

# Check provider statistics
curl http://localhost:8080/providers/stats

# Monitor email delivery
curl http://localhost:8080/notifications/logs
```

## ⚙️ **Configuration**

### **Environment Configuration**

The local setup uses the default configuration from `env/default.toml`:

```toml
# Database configuration
[database]
host = \"localhost\"
port = 5432
user = \"postgres\"
password = \"postgres\"
name = \"newsletter_db\"

# Redis configuration  
[redis]
host = \"localhost\"
port = 6379
password = \"\"
db = 0

# Email provider configuration
[providers]
enabled = [\"smtp_primary\", \"mailtrap\"]
load_balancing = \"round_robin\"
```

### **Email Provider Setup**

For local development, you can configure email providers:

#### **Gmail SMTP (for testing)**
```toml
[providers.smtp.smtp_primary]
host = \"smtp.gmail.com\"
port = 587
username = \"your-email@gmail.com\"
password = \"your-app-password\"
from = \"your-email@gmail.com\"
```

#### **Mailtrap (for testing)**
```toml
[providers.api.mailtrap]
endpoint = \"https://bulk.api.mailtrap.io/api/send\"
token = \"your-mailtrap-token\"
from = \"test@yourdomain.com\"
bulk_enabled = true
```

## 🛑 **Stopping Services**

### **Stop All Services**
```bash
# Stop and remove containers
docker-compose down

# Stop and remove containers + volumes (destroys data)
docker-compose down -v

# Stop and remove containers + images
docker-compose down --rmi all
```

### **Restart Services**
```bash
# Quick restart
docker-compose restart

# Full restart with rebuild
docker-compose down && ./scripts/local.sh
```

## 🐛 **Troubleshooting**

### **Common Issues**

#### **1. Port Already in Use**
```bash
# Error: port 8080 already in use
# Solution: Kill process using port
lsof -ti:8080 | xargs kill -9

# Or use different port in docker-compose.yml
ports:
  - \"8081:8080\"
```

#### **2. Docker Network Issues**
```bash
# Error: network conflicts
# Solution: Remove existing network
docker network rm newsletter-network

# Re-run setup script
./scripts/local.sh
```

#### **3. Database Connection Failed**
```bash
# Check if PostgreSQL container is running
docker ps | grep newsletter-db

# Check database logs
docker logs newsletter-db

# Restart database
docker restart newsletter-db
```

#### **4. Permission Denied**
```bash
# Make script executable
chmod +x scripts/local.sh

# Check Docker permissions (Linux)
sudo usermod -aG docker $USER
newgrp docker
```

### **Reset Everything**
```bash
# Nuclear option: remove all containers and data
docker-compose down -v
docker system prune -f
docker volume prune -f

# Start fresh
./scripts/local.sh
```

## 🔧 **Manual Setup (Alternative)**

If you prefer manual setup or the script fails:

### **1. Start Infrastructure**
```bash
# Create network
docker network create newsletter-network

# Start PostgreSQL
docker run -d \\
  --name newsletter-db \\
  --network newsletter-network \\
  -e POSTGRES_DB=newsletter_db \\
  -e POSTGRES_USER=postgres \\
  -e POSTGRES_PASSWORD=postgres \\
  -p 5432:5432 \\
  postgres:13

# Start Redis
docker run -d \\
  --name newsletter-redis \\
  --network newsletter-network \\
  -p 6379:6379 \\
  redis:6-alpine
```

### **2. Build and Run Application**
```bash
# Build web service
docker build -f scripts/Dockerfile.web -t newsletter-web .

# Build worker service  
docker build -f scripts/Dockerfile.worker -t newsletter-worker .

# Run web service
docker run -d \\
  --name newsletter-web \\
  --network newsletter-network \\
  -p 8080:8080 \\
  newsletter-web

# Run worker service
docker run -d \\
  --name newsletter-worker \\
  --network newsletter-network \\
  newsletter-worker
```

### **3. Run Migrations**
```bash
# Run database migrations
docker exec newsletter-web go run migration/migrate.go
```

## 📈 **Performance Tips**

### **Optimize for Development**
- Use Docker BuildKit for faster builds
- Enable Docker layer caching
- Use bind mounts for live code reloading
- Allocate adequate resources to Docker

### **Resource Allocation**
```bash
# Increase Docker memory (Docker Desktop)
# Settings → Resources → Advanced → Memory: 4GB+

# Monitor resource usage
docker stats
```

## ✅ **Success Checklist**

- [ ] All containers running (`docker ps`)
- [ ] Health check passing (`curl http://localhost:8080/health`)
- [ ] Can create topics, subscribers, and content
- [ ] Database accessible and populated
- [ ] Redis caching working
- [ ] Email providers configured (optional)
- [ ] Worker processing notifications

---

Your local Newsletter Service development environment is ready! 🎉

**Next Steps**: Start building features, test API endpoints, or deploy to production using the [Live Deployment Guide](LIVE_DEPLOYMENT.md).