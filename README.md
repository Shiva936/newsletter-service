# Newsletter Service - Cloud Deployment

A comprehensive newsletter management service with enterprise-grade features including rate limiting, scheduler authentication, and proper migration management.

## 🚀 Cloud Deployment (Heroku + Upstash)

This guide covers deploying the newsletter service to production using Heroku for hosting and Upstash for Redis.

### Prerequisites

- **Heroku Account**: [Sign up at heroku.com](https://signup.heroku.com/)
- **Heroku CLI**: [Install Heroku CLI](https://devcenter.heroku.com/articles/heroku-cli)
- **Upstash Account**: [Sign up at upstash.com](https://console.upstash.com/login)
- **Git**: Repository should be ready for deployment

### 1. Create Heroku Application

```bash
# Login to Heroku
heroku login

# Create new app (replace with your app name)
heroku create newsletter-service-prod

# Add Heroku Postgres addon
heroku addons:create heroku-postgresql:hobby-dev

# Verify addons
heroku addons
```

### 2. Setup Upstash Redis

1. **Create Redis Database**:

   - Go to [Upstash Console](https://console.upstash.com/)
   - Click "Create Database"
   - Choose region closest to your Heroku app
   - Select "Global" for multi-region or "Regional" for single region
   - Copy the Redis URL

2. **Get Connection Details**:
   - Copy the `UPSTASH_REDIS_REST_URL`
   - Copy the `UPSTASH_REDIS_REST_TOKEN`
   - Or use the standard Redis URL format

### 3. Configure Environment Variables

```bash
# Database (automatically set by Heroku Postgres addon)
# DATABASE_URL is set automatically

# Redis Configuration
heroku config:set REDIS_HOST=your-redis-host.upstash.io
heroku config:set REDIS_PORT=6379
heroku config:set REDIS_PASSWORD=your-redis-password
heroku config:set REDIS_DB=0

# Authentication
heroku config:set AUTH_USERNAME=admin
heroku config:set AUTH_PASSWORD=your-secure-password

# Scheduler Authentication
heroku config:set SCHEDULER_USERNAME=scheduler
heroku config:set SCHEDULER_PASSWORD=your-scheduler-password
heroku config:set SCHEDULER_ENABLED=true

# Rate Limiting
heroku config:set RATE_LIMIT_ENABLED=true
heroku config:set RATE_LIMIT_STORAGE=redis

# Database Configuration
heroku config:set DATABASE_AUTO_MIGRATE=false

# Optional: Environment
heroku config:set ENVIRONMENT=production
```

### 4. Deploy Application

```bash
# Add Heroku remote (if not already added)
heroku git:remote -a newsletter-service-prod

# Deploy to Heroku
git push heroku main

# The release phase will automatically run migrations
# (defined in Procfile: release: goose -dir ./migration/sql postgres "$DATABASE_URL" up)
```

### 5. Scale Services

```bash
# Scale web dynos (API server)
heroku ps:scale web=1

# Scale worker dynos (background tasks)
heroku ps:scale worker=1

# Check dyno status
heroku ps
```

### 6. Verify Deployment

```bash
# Check logs
heroku logs --tail

# Test health endpoint
curl https://your-app-name.herokuapp.com/health

# Test scheduler health
curl -u scheduler:your-scheduler-password \
  https://your-app-name.herokuapp.com/scheduler/v1/health

# Test API endpoint
curl -u admin:your-password \
  https://your-app-name.herokuapp.com/api/v1/topics
```

### 7. Database Management

```bash
# Check migration status
heroku run 'goose -dir ./migration/sql postgres "$DATABASE_URL" status'

# Run specific migration command if needed
heroku run 'goose -dir ./migration/sql postgres "$DATABASE_URL" up'

# Access database directly
heroku pg:psql
```

### 8. Monitoring and Scaling

```bash
# View application metrics
heroku logs --tail

# Scale based on load
heroku ps:scale web=2 worker=2

# Monitor Redis usage in Upstash console
# Monitor Postgres usage
heroku pg:info
```

## 🔧 Configuration

### Required Environment Variables

| Variable             | Description                                | Example                 |
| -------------------- | ------------------------------------------ | ----------------------- |
| `DATABASE_URL`       | PostgreSQL connection (auto-set by Heroku) | `postgres://...`        |
| `REDIS_HOST`         | Upstash Redis host                         | `your-redis.upstash.io` |
| `REDIS_PASSWORD`     | Upstash Redis password                     | `your-password`         |
| `AUTH_USERNAME`      | API basic auth username                    | `admin`                 |
| `AUTH_PASSWORD`      | API basic auth password                    | `secure-password`       |
| `SCHEDULER_USERNAME` | Scheduler auth username                    | `scheduler`             |
| `SCHEDULER_PASSWORD` | Scheduler auth password                    | `scheduler-password`    |

### Optional Environment Variables

| Variable                | Description          | Default       |
| ----------------------- | -------------------- | ------------- |
| `ENVIRONMENT`           | Environment name     | `development` |
| `RATE_LIMIT_ENABLED`    | Enable rate limiting | `true`        |
| `RATE_LIMIT_STORAGE`    | Storage backend      | `redis`       |
| `DATABASE_AUTO_MIGRATE` | Auto-run migrations  | `false`       |

## 📖 API Documentation

For complete API documentation including endpoints, request/response schemas, authentication details, and interactive testing, see:

**[OpenAPI Specification](./api-docs.yaml)**

### Quick API Overview

- **Main API**: `https://your-app.herokuapp.com/api/v1/*`
  - Authentication: Basic Auth with `AUTH_USERNAME:AUTH_PASSWORD`
  - Endpoints: Topics, Subscribers, Subscriptions, Content, Email Logs
- **Scheduler API**: `https://your-app.herokuapp.com/scheduler/v1/*`

  - Authentication: Basic Auth with `SCHEDULER_USERNAME:SCHEDULER_PASSWORD`
  - Endpoints: Notifications, Health checks

- **Health Check**: `https://your-app.herokuapp.com/health` (no auth required)

## 🛡️ Security Features

- **Rate Limiting**: Configurable per-route limits using Redis
- **Authentication Separation**: Different credentials for API vs Scheduler
- **Environment Isolation**: Production-safe configuration
- **Migration Safety**: Controlled schema management with Goose

## 🚨 Troubleshooting

### Common Issues

1. **Migration Failures**:

   ```bash
   heroku logs --tail | grep migration
   heroku run 'goose -dir ./migration/sql postgres "$DATABASE_URL" status'
   ```

2. **Redis Connection Issues**:

   ```bash
   heroku config:get REDIS_HOST
   heroku logs --tail | grep redis
   ```

3. **Authentication Problems**:

   ```bash
   heroku config:get AUTH_USERNAME
   heroku config:get AUTH_PASSWORD
   ```

4. **Rate Limiting Issues**:
   ```bash
   heroku config:set RATE_LIMIT_ENABLED=false  # Temporary disable
   ```

### Performance Optimization

1. **Scale Resources**:

   ```bash
   heroku ps:scale web=2 worker=2
   heroku addons:create heroku-postgresql:standard-0  # Upgrade DB
   ```

2. **Monitor Performance**:
   - Use Heroku metrics dashboard
   - Monitor Upstash Redis metrics
   - Set up log monitoring

## 📝 Support

- **GitHub Issues**: [Create an issue](https://github.com/your-username/newsletter-service/issues)
- **API Documentation**: See [OpenAPI Specification](./api-docs.yaml) for detailed API reference
- **Local Development**: See `LOCAL_README.md` for local development setup
- **Enterprise Features**: See `ENTERPRISE_FEATURES.md` for advanced configuration
