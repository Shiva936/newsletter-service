# Service Architecture

## 🏗️ **System Overview**

Newsletter Service is built with a modern, microservice-ready architecture that separates concerns between API handling, background processing, and data management. The system is designed for high availability, scalability, and maintainability.

## 📐 **Architecture Diagram**

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLIENT APPLICATIONS                     │
└─────────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
┌─────────────────────────────────────────────────────────────────┐
│                         LOAD BALANCER                          │
└─────────────────────────────────────────────────────────────────┘
                                  │
                ┌─────────────────┼─────────────────┐
                ▼                 ▼                 ▼
    ┌──────────────────┐ ┌──────────────────┐ ┌──────────────────┐
    │   WEB API (1)    │ │   WEB API (2)    │ │   WEB API (N)    │
    │   Port: 8080     │ │   Port: 8081     │ │   Port: 808N     │
    └──────────────────┘ └──────────────────┘ └──────────────────┘
                │                 │                 │
                └─────────────────┼─────────────────┘
                                  │
                                  ▼
    ┌─────────────────────────────────────────────────────────────┐
    │                    SHARED RESOURCES                         │
    │  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐│
    │  │   PostgreSQL    │ │      Redis      │ │   Email Queue   ││
    │  │   (Database)    │ │    (Cache)      │ │  (Background)   ││
    │  └─────────────────┘ └─────────────────┘ └─────────────────┘│
    └─────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
    ┌─────────────────────────────────────────────────────────────┐
    │                   WORKER SERVICES                           │
    │  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐│
    │  │   Worker (1)    │ │   Worker (2)    │ │   Worker (N)    ││
    │  │  (Scheduler)    │ │  (Scheduler)    │ │  (Scheduler)    ││
    │  └─────────────────┘ └─────────────────┘ └─────────────────┘│
    └─────────────────────────────────────────────────────────────┘
                                  │
                                  ▼
    ┌─────────────────────────────────────────────────────────────┐
    │                EMAIL PROVIDER LAYER                         │
    │                                                             │
    │  ┌─────────────────────────────────────────────────────────┐ │
    │  │                PROVIDER FACTORY                         │ │
    │  │  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────┐│ │
    │  │  │ Load Balancer   │ │ Health Monitor  │ │ Rate Limiter││ │
    │  │  └─────────────────┘ └─────────────────┘ └─────────────┘│ │
    │  └─────────────────────────────────────────────────────────┐ │
    │                          │                                  │
    │     ┌────────────────────┼────────────────────┐             │
    │     ▼                    ▼                    ▼             │
    │  ┌──────────────┐ ┌──────────────┐ ┌──────────────┐       │
    │  │SMTP Provider │ │API Provider  │ │API Provider  │       │
    │  │(Primary)     │ │(SendGrid)    │ │(Mailtrap)    │       │
    │  └──────────────┘ └──────────────┘ └──────────────┘       │
    └─────────────────────────────────────────────────────────────┘
```

## 🏗️ **Core Components**

### **1. Web API Layer**

**Purpose**: Handle HTTP requests and provide REST API endpoints

**Components**:
- **Gin Router**: HTTP routing and middleware
- **Handlers**: Request processing and response formatting
- **Middleware**: Authentication, rate limiting, logging
- **Validation**: Input validation and error handling

**Responsibilities**:
- Subscriber management (CRUD operations)
- Topic management (CRUD operations)
- Content management (CRUD operations)
- Notification triggering (scheduler endpoints)
- Health checks and monitoring

**Key Files**:
- `cmd/web/main.go` - Application entry point
- `internal/handlers/` - HTTP handlers
- `internal/router/` - Route definitions and middleware
- `internal/services/` - Business logic layer

### **2. Worker Service Layer**

**Purpose**: Background processing for scheduled tasks and email delivery

**Components**:
- **Notification Scheduler**: Processes pending notifications
- **Email Dispatcher**: Manages email sending through providers
- **Retry Handler**: Handles failed email delivery retries
- **Worker Pool**: Manages concurrent processing

**Responsibilities**:
- Monitor for scheduled content
- Process email queues
- Handle provider failover
- Retry failed deliveries
- Update delivery status

**Key Files**:
- `cmd/worker/main.go` - Worker entry point
- `internal/schedulers/` - Background job processing
- `internal/providers/` - Email provider management

### **3. Data Layer**

#### **PostgreSQL Database**

**Purpose**: Persistent storage for application data

**Schema Design**:
```sql
-- Core entities
Topics: id, name, description, created_at, updated_at
Subscribers: id, name, email, is_active, created_at, updated_at
Contents: id, title, body, topic_id, scheduled_time, status, created_at
Subscriptions: id, subscriber_id, topic_id, created_at
Email_Logs: id, subscriber_id, content_id, email_address, status, sent_at, error_message
```

**Key Features**:
- GORM ORM for object-relational mapping
- Automatic migrations
- Connection pooling
- Transaction support

#### **Redis Cache**

**Purpose**: Caching and session management

**Usage**:
- Rate limiting counters
- Session storage
- Temporary data caching
- Provider health status cache

### **4. Email Provider Layer**

#### **Provider Factory Pattern**

**Purpose**: Abstraction layer for multiple email providers

**Components**:
- **Provider Interface**: Common contract for all providers
- **Load Balancer**: Distributes load across providers
- **Health Monitor**: Tracks provider availability and performance
- **Batch Manager**: Handles async batching for non-bulk providers

#### **Supported Providers**:

1. **SMTP Providers**
   - Traditional SMTP servers (Gmail, Outlook, etc.)
   - Individual email sending
   - Fallback for reliability

2. **API Providers**
   - **SendGrid**: High-volume transactional emails
   - **Mailtrap**: Testing and development
   - **Generic API**: Extensible for other providers

#### **Provider Selection Logic**:
```go
// Selection priority:
1. Health Check (is provider healthy?)
2. Priority Level (lower number = higher priority)
3. Load Balancing Strategy (round-robin, weighted, least-load)
4. Rate Limiting (within provider limits?)
5. Bulk Capability (for large recipient lists)
```

## 🔄 **Data Flow Architecture**

### **1. Content Creation Flow**

```
Client Request → Web API → Content Service → Database → Response
     │
     └→ Scheduled Time Reached → Worker → Email Dispatch → Providers
```

### **2. Email Delivery Flow**

```
Scheduler Timer → Worker Service → Content Lookup → Subscriber Lookup
       │
       ▼
Provider Factory → Load Balancer → Health Check → Provider Selection
       │
       ▼
Email Dispatch → Provider API/SMTP → Delivery Status → Log Update
```

### **3. Multi-Provider Failover**

```
Primary Provider → Health Check → [FAIL] → Secondary Provider
       │                                           │
       ▼                                           ▼
   [SUCCESS]                                   Health Check
       │                                           │
       ▼                                           ▼
  Log Success                                  [SUCCESS/FAIL]
```

## ⚙️ **Configuration Architecture**

### **Environment-Based Configuration**

```toml
# env/default.toml - Base configuration
[database]
host = "localhost"
port = 5432

[providers]
enabled = ["smtp_primary", "sendgrid"]
load_balancing = "round_robin"

# Dynamic provider mapping
[providers.smtp.smtp_primary]
host = "smtp.example.com"
priority = 1

[providers.api.sendgrid]
token = "SG.api-key"
bulk_enabled = true
priority = 2
```

### **Dynamic Provider Configuration**

The system supports runtime provider configuration changes:

- **Hot-reload**: Configuration changes without restart
- **Provider addition**: Add new email providers dynamically
- **Load balancing**: Switch strategies without downtime
- **Health monitoring**: Automatic provider health detection

## 🔒 **Security Architecture**

### **Authentication & Authorization**

- **Basic Authentication**: For administrative endpoints
- **API Key Authentication**: For programmatic access
- **Rate Limiting**: Protection against abuse
- **Input Validation**: Comprehensive request validation

### **Data Protection**

- **Email Encryption**: Sensitive email content protection
- **Database Security**: Connection encryption and credentials
- **Provider Keys**: Secure storage of API keys
- **Audit Logging**: Comprehensive activity logging

## 📊 **Monitoring & Observability**

### **Health Checks**

```
/health                 - Basic service health
/health/detailed        - Detailed component status
/providers/health       - Email provider health
/providers/stats        - Provider performance metrics
```

### **Logging Strategy**

- **Structured Logging**: JSON formatted logs
- **Log Levels**: DEBUG, INFO, WARN, ERROR
- **Context Propagation**: Request tracking across services
- **Error Tracking**: Detailed error information and stack traces

### **Metrics Collection**

- **Email Delivery Rates**: Success/failure statistics
- **Provider Performance**: Response times and error rates
- **System Performance**: CPU, memory, and database metrics
- **Business Metrics**: Subscriber growth, content engagement

## 🚀 **Scalability Architecture**

### **Horizontal Scaling**

- **Stateless Design**: Web API can scale horizontally
- **Load Balancing**: Multiple API instances behind load balancer
- **Worker Scaling**: Multiple worker instances for background processing
- **Database Scaling**: Read replicas and connection pooling

### **Performance Optimizations**

- **Connection Pooling**: Database and Redis connection reuse
- **Batch Processing**: Efficient bulk email handling
- **Async Processing**: Non-blocking email delivery
- **Caching Strategy**: Redis caching for frequently accessed data

### **Resource Management**

- **Worker Pools**: Configurable concurrency limits
- **Memory Management**: Efficient goroutine handling
- **Rate Limiting**: Provider-specific rate controls
- **Circuit Breakers**: Provider failure protection

## 🔧 **Development Architecture**

### **Code Organization**

```
cmd/
├── web/        - Web API application
└── worker/     - Background worker application

internal/
├── config/     - Configuration management
├── handlers/   - HTTP request handlers
├── services/   - Business logic layer
├── providers/  - Email provider implementations
├── connections/ - Database and Redis connections
├── router/     - HTTP routing and middleware
└── schedulers/ - Background job processing

migration/
└── sql/        - Database migration scripts

env/
└── default.toml - Configuration files
```

### **Design Patterns**

- **Factory Pattern**: Provider creation and management
- **Strategy Pattern**: Load balancing strategies
- **Observer Pattern**: Health monitoring
- **Repository Pattern**: Data access abstraction
- **Dependency Injection**: Service composition

### **Testing Strategy**

- **Unit Tests**: Individual component testing
- **Integration Tests**: Service interaction testing
- **End-to-End Tests**: Complete workflow testing
- **Load Tests**: Performance and scalability testing

## 🐳 **Deployment Architecture**

### **Containerization**

- **Docker Images**: Separate images for web and worker
- **Multi-stage Builds**: Optimized production images
- **Health Checks**: Container health monitoring
- **Resource Limits**: CPU and memory constraints

### **Infrastructure Patterns**

- **12-Factor App**: Cloud-native application design
- **Microservice Ready**: Separation of concerns
- **Stateless Services**: Horizontal scalability
- **External Configuration**: Environment-based config

### **Production Deployment**

- **Container Orchestration**: Kubernetes or Docker Swarm
- **Service Discovery**: Dynamic service location
- **Load Balancing**: Traffic distribution
- **Auto-scaling**: Resource-based scaling policies

---

This architecture provides a robust, scalable foundation for newsletter delivery with enterprise-grade features including multi-provider email support, comprehensive monitoring, and production-ready deployment patterns.