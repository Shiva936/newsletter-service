# Live Deployment Guide

Complete guide for deploying Newsletter Service to production using Heroku and Upstash Redis.

## 🎯 **Deployment Overview**

This guide covers production deployment using:
- **Heroku**: Application hosting and PostgreSQL database
- **Upstash**: Managed Redis for caching and session management
- **Environment Variables**: Secure configuration management
- **Health Monitoring**: Production readiness checks

## 📋 **Prerequisites**

### **Required Accounts**
- **Heroku Account**: [Sign up at heroku.com](https://signup.heroku.com/)
- **Upstash Account**: [Sign up at upstash.com](https://console.upstash.com/login)
- **Email Provider Account**: SendGrid, Gmail, or similar

### **Required Tools**
- **Heroku CLI**: [Install Heroku CLI](https://devcenter.heroku.com/articles/heroku-cli)
- **Git**: For code deployment
- **curl**: For testing APIs

## 🚀 **Step 1: Heroku Application Setup**

### **1.1 Login to Heroku**
```bash
# Login to your Heroku account
heroku login

# Verify login
heroku auth:whoami
```

### **1.2 Create Heroku Application**
```bash
# Create new app (replace with your app name)
heroku create newsletter-service-prod

# Add Heroku Git remote
heroku git:remote -a newsletter-service-prod

# Verify app creation
heroku apps:info newsletter-service-prod
```

### **1.3 Add PostgreSQL Database**
```bash
# Add Heroku Postgres addon
heroku addons:create heroku-postgresql:hobby-dev -a newsletter-service-prod

# Verify addon
heroku addons -a newsletter-service-prod

# Get database URL
heroku config:get DATABASE_URL -a newsletter-service-prod
```

## ☁️ **Step 2: Upstash Redis Setup**

### **2.1 Create Redis Database**

1. **Login to Upstash Console**:
   - Go to [Upstash Console](https://console.upstash.com/)
   - Login with your account

2. **Create Database**:
   - Click **\"Create Database\"**
   - **Name**: `newsletter-service-redis`
   - **Region**: Choose closest to your Heroku region
   - **Type**: Select **\"Global\"** or **\"Regional\"**
   - Click **\"Create\"**

3. **Get Connection Details**:
   - Copy the **Redis URL** from the database details
   - Format: `redis://default:password@host:port`

### **2.2 Configure Upstash Redis**
```bash
# Add Redis URL to Heroku
heroku config:set REDIS_URL=\"redis://default:your-password@your-host:port\" -a newsletter-service-prod

# Verify configuration
heroku config -a newsletter-service-prod
```

## ⚙️ **Step 3: Environment Configuration**

### **3.1 Email Provider Setup**

#### **Option A: Gmail SMTP (Free)**
```bash
# Gmail SMTP configuration
heroku config:set SMTP_HOST=\"smtp.gmail.com\" -a newsletter-service-prod
heroku config:set SMTP_PORT=\"587\" -a newsletter-service-prod
heroku config:set SMTP_USERNAME=\"your-email@gmail.com\" -a newsletter-service-prod
heroku config:set SMTP_PASSWORD=\"your-app-password\" -a newsletter-service-prod
heroku config:set SMTP_FROM=\"your-email@gmail.com\" -a newsletter-service-prod
```

**Gmail App Password Setup**:
1. Enable 2-factor authentication
2. Go to Google Account settings
3. Security → App passwords
4. Generate app password for \"Mail\"
5. Use this password in SMTP_PASSWORD

#### **Option B: SendGrid API (Scalable)**
```bash
# SendGrid configuration
heroku config:set SENDGRID_API_KEY=\"SG.your-api-key\" -a newsletter-service-prod
heroku config:set SENDGRID_FROM=\"noreply@yourdomain.com\" -a newsletter-service-prod
```

**SendGrid Setup**:
1. [Sign up for SendGrid](https://sendgrid.com/free/)
2. Verify your sender email/domain
3. Create API key with \"Mail Send\" permissions
4. Use API key in configuration

### **3.2 Application Configuration**
```bash
# Basic application settings
heroku config:set ENV=\"production\" -a newsletter-service-prod
heroku config:set PORT=\"8080\" -a newsletter-service-prod

# Authentication
heroku config:set AUTH_USERNAME=\"admin\" -a newsletter-service-prod
heroku config:set AUTH_PASSWORD=\"secure-admin-password\" -a newsletter-service-prod
heroku config:set SCHEDULER_USERNAME=\"scheduler\" -a newsletter-service-prod
heroku config:set SCHEDULER_PASSWORD=\"secure-scheduler-password\" -a newsletter-service-prod

# Worker configuration
heroku config:set MAX_ASYNC_PROCESS=\"10\" -a newsletter-service-prod
```

### **3.3 Provider Configuration**
```bash
# Multi-provider email configuration
heroku config:set PROVIDERS_ENABLED=\"smtp_primary,sendgrid\" -a newsletter-service-prod
heroku config:set LOAD_BALANCING=\"round_robin\" -a newsletter-service-prod

# Provider priorities
heroku config:set SMTP_PRIORITY=\"1\" -a newsletter-service-prod
heroku config:set SENDGRID_PRIORITY=\"2\" -a newsletter-service-prod

# Rate limiting
heroku config:set SMTP_MAX_EMAILS_HOUR=\"1000\" -a newsletter-service-prod
heroku config:set SENDGRID_MAX_EMAILS_HOUR=\"10000\" -a newsletter-service-prod
```

## 📦 **Step 4: Application Deployment**

### **4.1 Prepare Deployment Files**

Create `Procfile` in project root:
```bash
# Procfile
web: go run cmd/web/main.go
worker: go run cmd/worker/main.go
```

Create `heroku.yml` for container deployment (optional):
```yaml
# heroku.yml
build:
  docker:
    web: scripts/Dockerfile.web
    worker: scripts/Dockerfile.worker
run:
  web: go run cmd/web/main.go
  worker: go run cmd/worker/main.go
```

### **4.2 Deploy Application**
```bash
# Commit your changes
git add .
git commit -m \"Production deployment setup\"

# Deploy to Heroku
git push heroku master

# Check deployment logs
heroku logs --tail -a newsletter-service-prod
```

### **4.3 Scale Services**
```bash
# Scale web service
heroku ps:scale web=1 -a newsletter-service-prod

# Scale worker service  
heroku ps:scale worker=1 -a newsletter-service-prod

# Check running processes
heroku ps -a newsletter-service-prod
```

## 🗄️ **Step 5: Database Setup**

### **5.1 Run Migrations**
```bash
# Run database migrations
heroku run go run migration/migrate.go -a newsletter-service-prod

# Verify database structure
heroku pg:psql -a newsletter-service-prod -c \"\\dt\"
```

### **5.2 Create Initial Data (Optional)**
```bash
# Connect to database
heroku pg:psql -a newsletter-service-prod

# Create sample topic
INSERT INTO topics (name, description, created_at, updated_at) 
VALUES ('General', 'General newsletter content', NOW(), NOW());

# Exit database
\\q
```

## ✅ **Step 6: Verification & Testing**

### **6.1 Health Checks**
```bash
# Get app URL
heroku info -a newsletter-service-prod

# Test basic health
curl https://your-app-name.herokuapp.com/health

# Test detailed health
curl https://your-app-name.herokuapp.com/health/detailed

# Test provider health
curl https://your-app-name.herokuapp.com/providers/health
```

### **6.2 API Testing**
```bash
# Set your app URL
export APP_URL=\"https://your-app-name.herokuapp.com\"

# Test topic creation
curl -X POST $APP_URL/topics \\
  -H \"Content-Type: application/json\" \\
  -d '{\"name\":\"Tech News\",\"description\":\"Technology updates\"}'

# Test subscriber creation
curl -X POST $APP_URL/subscribers \\
  -H \"Content-Type: application/json\" \\
  -d '{\"name\":\"John Doe\",\"email\":\"john@example.com\",\"subscribed_topics\":[\"Tech News\"]}'

# Test content creation
curl -X POST $APP_URL/contents \\
  -H \"Content-Type: application/json\" \\
  -d '{\"title\":\"Test Newsletter\",\"body\":\"Test content\",\"topic_id\":1,\"scheduled_time\":\"2025-11-18T10:00:00Z\"}'
```

### **6.3 Email Testing**
```bash
# Encode scheduler credentials
echo -n \"scheduler:your-scheduler-password\" | base64

# Test notification trigger
curl -X POST $APP_URL/notifications/send \\
  -H \"Content-Type: application/json\" \\
  -H \"Authorization: Basic <base64-credentials>\" \\
  -d '{\"content_id\":1}'

# Check email logs
curl $APP_URL/notifications/logs
```

## 📊 **Step 7: Monitoring Setup**

### **7.1 Log Management**
```bash
# View real-time logs
heroku logs --tail -a newsletter-service-prod

# View specific service logs
heroku logs --source=app --tail -a newsletter-service-prod

# Search logs
heroku logs --grep=\"ERROR\" -a newsletter-service-prod
```

### **7.2 Performance Monitoring**
```bash
# Check app metrics
heroku ps -a newsletter-service-prod

# Monitor database performance
heroku pg:info -a newsletter-service-prod

# Check Redis metrics (from Upstash dashboard)
```

### **7.3 Alerts Setup**

**Heroku Alerts**:
```bash
# Add Heroku scheduler for monitoring
heroku addons:create scheduler:standard -a newsletter-service-prod

# Schedule health checks
heroku addons:open scheduler -a newsletter-service-prod
```

**Upstash Monitoring**:
- Enable alerts in Upstash dashboard
- Set memory usage alerts
- Configure connection count alerts

## 🔧 **Step 8: Production Optimization**

### **8.1 Performance Configuration**
```bash
# Optimize worker configuration
heroku config:set MAX_ASYNC_PROCESS=\"20\" -a newsletter-service-prod

# Configure database connections
heroku config:set DB_MAX_OPEN_CONNS=\"10\" -a newsletter-service-prod
heroku config:set DB_MAX_IDLE_CONNS=\"5\" -a newsletter-service-prod

# Redis connection optimization
heroku config:set REDIS_POOL_SIZE=\"10\" -a newsletter-service-prod
```

### **8.2 Scaling Configuration**
```bash
# Scale for higher load
heroku ps:scale web=2 worker=2 -a newsletter-service-prod

# Configure auto-scaling (Professional plans)
heroku ps:autoscale enable web --min=1 --max=10 -a newsletter-service-prod
```

## 🔒 **Step 9: Security Hardening**

### **9.1 Environment Security**
```bash
# Use strong passwords
heroku config:set AUTH_PASSWORD=\"$(openssl rand -base64 32)\" -a newsletter-service-prod
heroku config:set SCHEDULER_PASSWORD=\"$(openssl rand -base64 32)\" -a newsletter-service-prod

# Enable HTTPS redirect
heroku config:set FORCE_HTTPS=\"true\" -a newsletter-service-prod
```

### **9.2 Access Control**
```bash
# Restrict Heroku access
heroku access -a newsletter-service-prod

# Add team members (if needed)
heroku access:add email@example.com -a newsletter-service-prod
```

## 🔄 **Step 10: Maintenance & Updates**

### **10.1 Regular Maintenance**
```bash
# Check for updates
git pull origin master
git push heroku master

# Restart application
heroku restart -a newsletter-service-prod

# Check application health
curl https://your-app-name.herokuapp.com/health
```

### **10.2 Backup Strategy**
```bash
# Create database backup
heroku pg:backups:capture -a newsletter-service-prod

# List available backups
heroku pg:backups -a newsletter-service-prod

# Download backup (if needed)
heroku pg:backups:download -a newsletter-service-prod
```

## 📱 **Step 11: Custom Domain (Optional)**

### **11.1 Add Custom Domain**
```bash
# Add your domain
heroku domains:add newsletter.yourdomain.com -a newsletter-service-prod

# Get DNS target
heroku domains -a newsletter-service-prod
```

### **11.2 SSL Certificate**
```bash
# Add SSL certificate
heroku certs:auto:enable -a newsletter-service-prod

# Verify SSL
heroku certs -a newsletter-service-prod
```

## 🎯 **Production Checklist**

### **✅ Pre-Launch**
- [ ] Database migrations completed
- [ ] Environment variables configured
- [ ] Email providers tested
- [ ] Health checks passing
- [ ] Authentication working
- [ ] Worker service running

### **✅ Post-Launch**
- [ ] Monitor application logs
- [ ] Check email delivery rates
- [ ] Verify provider health
- [ ] Test failover scenarios
- [ ] Monitor resource usage
- [ ] Set up alerts and notifications

## 🐛 **Troubleshooting**

### **Common Issues**

1. **Application Not Starting**
   ```bash
   heroku logs --tail -a newsletter-service-prod
   # Check for configuration errors
   ```

2. **Database Connection Issues**
   ```bash
   heroku pg:info -a newsletter-service-prod
   # Verify DATABASE_URL is set
   ```

3. **Email Delivery Problems**
   ```bash
   curl $APP_URL/providers/health
   # Check provider configuration
   ```

4. **Worker Not Processing**
   ```bash
   heroku ps -a newsletter-service-prod
   # Ensure worker is scaled up
   ```

### **Support Resources**
- **Heroku Documentation**: [devcenter.heroku.com](https://devcenter.heroku.com/)
- **Upstash Documentation**: [docs.upstash.com](https://docs.upstash.com/)
- **Application Logs**: `heroku logs --tail`

## 💰 **Cost Optimization**

### **Free Tier Usage**
- **Heroku**: 550-1000 free dyno hours/month
- **PostgreSQL**: hobby-dev plan (free)
- **Upstash**: 10K commands/day (free tier)

### **Paid Plans**
- **Heroku Hobby**: $7/month per dyno
- **PostgreSQL Standard-0**: $9/month
- **Upstash Pro**: $0.2 per 100K commands

---

Your Newsletter Service is now live and ready for production use! 🎉

**Next Steps**: Monitor performance, set up alerts, and scale based on usage patterns.