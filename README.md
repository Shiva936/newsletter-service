# Newsletter Service

A robust, scalable newsletter service designed to send scheduled content to topic-based subscribers with enterprise-grade multi-provider email support.

## 🎯 **What is Newsletter Service?**

Newsletter Service is a comprehensive solution for managing and delivering newsletters at scale. It provides a complete backend infrastructure for:

- **Subscriber Management**: Organize users and their topic preferences
- **Content Creation**: Schedule newsletters with topic-based organization
- **Automated Delivery**: Background processing for reliable email delivery
- **Multi-Provider Email**: Enterprise-grade email delivery with failover
- **Monitoring & Analytics**: Track delivery status and provider health

## 🚀 **Key Features**

### **Core Capabilities**
- 📧 **Topic-Based Subscriptions**: Users subscribe to specific content categories
- ⏰ **Scheduled Delivery**: Automatic sending at specified times
- 🔄 **Multi-Provider Email**: SMTP and API providers with automatic failover
- 📊 **Bulk Email Support**: Efficient handling of large subscriber lists
- 🎯 **Load Balancing**: Intelligent distribution across email providers
- 📈 **Health Monitoring**: Real-time provider statistics and health checks

### **Enterprise Features**
- ⚖️ **Rate Limiting**: Configurable limits per provider
- 🔁 **Retry Mechanisms**: Automatic retry for failed deliveries
- 📝 **Email Tracking**: Comprehensive delivery status logging
- 🏗️ **Async Processing**: Worker pools for optimal performance
- 🐳 **Container Ready**: Docker containerization for easy deployment

## 🏗️ **Architecture Overview**

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web API       │    │   Worker        │    │   Database      │
│   (REST APIs)   │────│   (Scheduler)   │────│   (PostgreSQL)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       └───────────────────────┼─── Redis Cache
         │                                               │
         └─── Multi-Provider Email System ────────────────┘
              ├── SMTP Providers (Primary, Backup)
              ├── API Providers (SendGrid, Mailtrap)
              └── Load Balancer + Health Monitoring
```

## 🛠️ **Technology Stack**

- **Backend**: Go (Gin framework)
- **Database**: PostgreSQL with GORM
- **Cache**: Redis
- **Email**: SMTP + API providers (SendGrid, Mailtrap)
- **Containerization**: Docker & Docker Compose
- **Configuration**: TOML-based configuration

## 📚 **Documentation**

### **Setup & Deployment**
- 🏠 [**Local Setup Guide**](LOCAL_SETUP.md) - Quick start with Docker and local development
- ☁️ [**Live Deployment Guide**](LIVE_DEPLOYMENT.md) - Production deployment with Heroku and Upstash
- 🏗️ [**Service Architecture**](SERVICE_ARCHITECTURE.md) - Detailed technical architecture and design

### **Quick Start**

1. **Local Development**
```bash
# Clone and start services
git clone <repository-url>
cd newsletter-service
chmod +x scripts/local.sh
./scripts/local.sh
```

2. **Verify Setup**
```bash
curl http://localhost:8080/health
```

3. **Create Your First Newsletter**
```bash
# Create a topic
curl -X POST http://localhost:8080/topics \
  -H "Content-Type: application/json" \
  -d '{"name":"Tech News","description":"Latest technology updates"}'

# Add a subscriber
curl -X POST http://localhost:8080/subscribers \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","subscribed_topics":["Tech News"]}'

# Schedule content
curl -X POST http://localhost:8080/contents \
  -H "Content-Type: application/json" \
  -d '{"title":"Weekly Update","body":"This week in tech...","topic_id":1,"scheduled_time":"2025-11-18T10:00:00Z"}'
```

## 📊 **Use Cases**

### **Business Applications**
- 📰 **Company Newsletters**: Regular updates to employees and customers
- 🛍️ **Marketing Campaigns**: Product announcements and promotional content
- 📈 **Investment Updates**: Financial reports and market analysis
- 🎓 **Educational Content**: Course updates and learning materials

### **Technical Applications**
- 🔔 **System Notifications**: Infrastructure alerts and status updates
- 📊 **Report Distribution**: Automated report delivery to stakeholders
- 🚨 **Alert Systems**: Critical system notifications and warnings
- 📅 **Event Reminders**: Scheduled event notifications and updates

## 🌟 **Why Choose Newsletter Service?**

### **Reliability**
- **Multi-Provider Architecture**: Never depend on a single email service
- **Automatic Failover**: Seamless switching between providers
- **Health Monitoring**: Real-time provider status tracking
- **Retry Mechanisms**: Automatic retry for failed deliveries

### **Scalability**
- **Bulk Email Support**: Handle thousands of subscribers efficiently
- **Load Balancing**: Distribute load across multiple providers
- **Async Processing**: Non-blocking email delivery
- **Worker Pools**: Configurable concurrency for optimal performance

### **Developer Friendly**
- **REST API**: Complete API for integration
- **Docker Ready**: Easy containerized deployment
- **Comprehensive Docs**: Detailed setup and API documentation
- **Health Checks**: Built-in monitoring endpoints

## 🚀 **Getting Started**

1. **For Local Development**: Follow the [Local Setup Guide](LOCAL_SETUP.md)
2. **For Production Deployment**: Check the [Live Deployment Guide](LIVE_DEPLOYMENT.md)
3. **For Understanding Architecture**: Read the [Service Architecture](SERVICE_ARCHITECTURE.md)

## 📞 **Support**

- 📝 **Issues**: Create an issue in the GitHub repository
- 📖 **Documentation**: Check the detailed guides in this repository
- 🔧 **Configuration**: Review the configuration examples in each guide

---

**Built for reliable, scalable newsletter delivery** 🚀