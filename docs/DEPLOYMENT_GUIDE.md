# Deployment Guide

## 🚀 Deployment Options

### **1. Development Setup**
Quick setup for local development

### **2. Docker Deployment**
Containerized deployment with Docker Compose

### **3. Production Deployment**
Production-ready deployment with optimizations

### **4. Kubernetes Deployment**
Scalable deployment on Kubernetes

---

## 🛠️ Development Setup

### **Prerequisites**
- Go 1.21+
- Node.js 18+
- MongoDB 6.0+
- Redis 7.0+

### **Quick Start**
```bash
# Clone repository
git clone <repository-url>
cd realtime-chat-system

# Install dependencies
go mod download
cd frontend && npm install

# Setup environment
cp .env.example .env
cp frontend/.env.example frontend/.env

# Start databases
docker-compose up -d mongodb redis

# Start backend
go run cmd/server/main.go

# Start frontend (new terminal)
cd frontend && npm run dev
```

### **Environment Configuration**
```bash
# .env
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=chat_system
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRATION=24h
SERVER_PORT=8080
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s
SERVER_SHUTDOWN_TIMEOUT=30s
```

```bash
# frontend/.env
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
```

---

## 🐳 Docker Deployment

### **Full Stack with Docker Compose**
```yaml
# docker-compose.yml
version: '3.8'

services:
  # Backend
  backend:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - MONGODB_URI=mongodb://mongodb:27017
      - REDIS_ADDR=redis:6379
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - mongodb
      - redis
    restart: unless-stopped

  # Frontend
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    ports:
      - "3000:80"
    environment:
      - VITE_API_URL=http://localhost:8080
      - VITE_WS_URL=ws://localhost:8080
    restart: unless-stopped

  # MongoDB
  mongodb:
    image: mongo:6.0
    ports:
      - "27017:27017"
    volumes:
      - mongodb_data:/data/db
    restart: unless-stopped

  # Redis
  redis:
    image: redis:7.0-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    restart: unless-stopped

volumes:
  mongodb_data:
  redis_data:
```

### **Backend Dockerfile**
```dockerfile
# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server cmd/server/main.go

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/server .

EXPOSE 8080
CMD ["./server"]
```

### **Frontend Dockerfile**
```dockerfile
# Build stage
FROM node:18-alpine AS builder

WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

COPY . .
RUN npm run build

# Runtime stage
FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### **Nginx Configuration**
```nginx
# nginx.conf
events {
    worker_connections 1024;
}

http {
    include       /etc/nginx/mime.types;
    default_type  application/octet-stream;

    server {
        listen 80;
        server_name localhost;
        root /usr/share/nginx/html;
        index index.html;

        # Frontend routes
        location / {
            try_files $uri $uri/ /index.html;
        }

        # API proxy
        location /api/ {
            proxy_pass http://backend:8080;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }

        # WebSocket proxy
        location /ws {
            proxy_pass http://backend:8080;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }
    }
}
```

### **Deploy with Docker Compose**
```bash
# Production deployment
docker-compose -f docker-compose.prod.yml up -d

# Check logs
docker-compose logs -f backend
docker-compose logs -f frontend

# Scale services
docker-compose up -d --scale backend=3

# Update services
docker-compose pull
docker-compose up -d --force-recreate
```

---

## 🏭 Production Deployment

### **Production Environment Setup**

#### **1. Server Requirements**
```
Minimum:
- CPU: 2 cores
- RAM: 4GB
- Storage: 20GB SSD
- Network: 100Mbps

Recommended:
- CPU: 4+ cores
- RAM: 8GB+
- Storage: 50GB+ SSD
- Network: 1Gbps
```

#### **2. Security Configuration**
```bash
# Generate secure JWT secret
JWT_SECRET=$(openssl rand -base64 64)

# Setup SSL certificates (Let's Encrypt)
sudo apt install certbot
sudo certbot certonly --standalone -d your-domain.com

# Firewall configuration
sudo ufw allow 22    # SSH
sudo ufw allow 80    # HTTP
sudo ufw allow 443   # HTTPS
sudo ufw enable
```

#### **3. Production Environment Variables**
```bash
# .env.production
MONGODB_URI=mongodb://username:password@mongodb-host:27017/chat_system?authSource=admin
REDIS_ADDR=redis-host:6379
REDIS_PASSWORD=secure-redis-password
JWT_SECRET=your-super-secure-jwt-secret-64-chars-long
JWT_EXPIRATION=24h
SERVER_PORT=8080
LOG_LEVEL=info
```

#### **4. Database Setup**
```bash
# MongoDB production setup
# Create admin user
mongo admin --eval "
  db.createUser({
    user: 'admin',
    pwd: 'secure-password',
    roles: ['userAdminAnyDatabase', 'dbAdminAnyDatabase']
  })
"

# Create application user
mongo chat_system --eval "
  db.createUser({
    user: 'chatapp',
    pwd: 'app-password',
    roles: ['readWrite']
  })
"

# Redis production setup
# Enable authentication
echo "requirepass secure-redis-password" >> /etc/redis/redis.conf
sudo systemctl restart redis
```

### **Production Docker Compose**
```yaml
# docker-compose.prod.yml
version: '3.8'

services:
  # Reverse Proxy
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.prod.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/ssl/certs
    depends_on:
      - backend
      - frontend
    restart: unless-stopped

  # Backend (multiple instances)
  backend:
    build:
      context: .
      dockerfile: Dockerfile.prod
    environment:
      - MONGODB_URI=${MONGODB_URI}
      - REDIS_ADDR=${REDIS_ADDR}
      - REDIS_PASSWORD=${REDIS_PASSWORD}
      - JWT_SECRET=${JWT_SECRET}
      - LOG_LEVEL=info
    depends_on:
      - mongodb
      - redis
    restart: unless-stopped
    deploy:
      replicas: 3

  # Frontend
  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile.prod
    environment:
      - VITE_API_URL=https://your-domain.com
      - VITE_WS_URL=wss://your-domain.com
    restart: unless-stopped

  # MongoDB with persistence
  mongodb:
    image: mongo:6.0
    environment:
      - MONGO_INITDB_ROOT_USERNAME=${MONGO_ROOT_USER}
      - MONGO_INITDB_ROOT_PASSWORD=${MONGO_ROOT_PASSWORD}
    volumes:
      - mongodb_data:/data/db
      - ./mongo-init.js:/docker-entrypoint-initdb.d/mongo-init.js
    restart: unless-stopped

  # Redis with persistence
  redis:
    image: redis:7.0-alpine
    command: redis-server --requirepass ${REDIS_PASSWORD} --appendonly yes
    volumes:
      - redis_data:/data
    restart: unless-stopped

volumes:
  mongodb_data:
    driver: local
  redis_data:
    driver: local
```

### **Production Nginx Configuration**
```nginx
# nginx.prod.conf
events {
    worker_connections 1024;
}

http {
    upstream backend {
        server backend:8080;
    }

    # Rate limiting
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req_zone $binary_remote_addr zone=auth:10m rate=5r/m;

    server {
        listen 80;
        server_name your-domain.com;
        return 301 https://$server_name$request_uri;
    }

    server {
        listen 443 ssl http2;
        server_name your-domain.com;

        # SSL Configuration
        ssl_certificate /etc/ssl/certs/fullchain.pem;
        ssl_certificate_key /etc/ssl/certs/privkey.pem;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;

        # Security headers
        add_header X-Frame-Options DENY;
        add_header X-Content-Type-Options nosniff;
        add_header X-XSS-Protection "1; mode=block";

        # Frontend
        location / {
            root /usr/share/nginx/html;
            try_files $uri $uri/ /index.html;
            
            # Caching
            location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
                expires 1y;
                add_header Cache-Control "public, immutable";
            }
        }

        # API with rate limiting
        location /api/ {
            limit_req zone=api burst=20 nodelay;
            proxy_pass http://backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }

        # Auth endpoints with stricter rate limiting
        location /api/auth/ {
            limit_req zone=auth burst=5 nodelay;
            proxy_pass http://backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
        }

        # WebSocket with sticky sessions
        location /ws {
            proxy_pass http://backend;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection "upgrade";
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            
            # Sticky sessions for WebSocket
            ip_hash;
        }
    }
}
```

---

## ☸️ Kubernetes Deployment

### **Namespace**
```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: chat-system
```

### **ConfigMap**
```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: chat-config
  namespace: chat-system
data:
  MONGODB_URI: "mongodb://mongodb:27017/chat_system"
  REDIS_ADDR: "redis:6379"
  SERVER_PORT: "8080"
  LOG_LEVEL: "info"
```

### **Secrets**
```yaml
# secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: chat-secrets
  namespace: chat-system
type: Opaque
data:
  JWT_SECRET: <base64-encoded-secret>
  REDIS_PASSWORD: <base64-encoded-password>
  MONGO_ROOT_PASSWORD: <base64-encoded-password>
```

### **MongoDB Deployment**
```yaml
# mongodb.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: mongodb
  namespace: chat-system
spec:
  serviceName: mongodb
  replicas: 1
  selector:
    matchLabels:
      app: mongodb
  template:
    metadata:
      labels:
        app: mongodb
    spec:
      containers:
      - name: mongodb
        image: mongo:6.0
        ports:
        - containerPort: 27017
        env:
        - name: MONGO_INITDB_ROOT_PASSWORD
          valueFrom:
            secretKeyRef:
              name: chat-secrets
              key: MONGO_ROOT_PASSWORD
        volumeMounts:
        - name: mongodb-data
          mountPath: /data/db
  volumeClaimTemplates:
  - metadata:
      name: mongodb-data
    spec:
      accessModes: ["ReadWriteOnce"]
      resources:
        requests:
          storage: 10Gi

---
apiVersion: v1
kind: Service
metadata:
  name: mongodb
  namespace: chat-system
spec:
  selector:
    app: mongodb
  ports:
  - port: 27017
    targetPort: 27017
```

### **Redis Deployment**
```yaml
# redis.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis
  namespace: chat-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
      - name: redis
        image: redis:7.0-alpine
        ports:
        - containerPort: 6379
        env:
        - name: REDIS_PASSWORD
          valueFrom:
            secretKeyRef:
              name: chat-secrets
              key: REDIS_PASSWORD
        command:
        - redis-server
        - --requirepass
        - $(REDIS_PASSWORD)
        - --appendonly
        - "yes"
        volumeMounts:
        - name: redis-data
          mountPath: /data
      volumes:
      - name: redis-data
        persistentVolumeClaim:
          claimName: redis-pvc

---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: redis-pvc
  namespace: chat-system
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: 5Gi

---
apiVersion: v1
kind: Service
metadata:
  name: redis
  namespace: chat-system
spec:
  selector:
    app: redis
  ports:
  - port: 6379
    targetPort: 6379
```

### **Backend Deployment**
```yaml
# backend.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: chat-backend
  namespace: chat-system
spec:
  replicas: 3
  selector:
    matchLabels:
      app: chat-backend
  template:
    metadata:
      labels:
        app: chat-backend
    spec:
      containers:
      - name: backend
        image: your-registry/chat-backend:latest
        ports:
        - containerPort: 8080
        envFrom:
        - configMapRef:
            name: chat-config
        - secretRef:
            name: chat-secrets
        resources:
          requests:
            memory: "256Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5

---
apiVersion: v1
kind: Service
metadata:
  name: chat-backend
  namespace: chat-system
spec:
  selector:
    app: chat-backend
  ports:
  - port: 8080
    targetPort: 8080
  type: ClusterIP
```

### **Frontend Deployment**
```yaml
# frontend.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: chat-frontend
  namespace: chat-system
spec:
  replicas: 2
  selector:
    matchLabels:
      app: chat-frontend
  template:
    metadata:
      labels:
        app: chat-frontend
    spec:
      containers:
      - name: frontend
        image: your-registry/chat-frontend:latest
        ports:
        - containerPort: 80
        resources:
          requests:
            memory: "64Mi"
            cpu: "50m"
          limits:
            memory: "128Mi"
            cpu: "100m"

---
apiVersion: v1
kind: Service
metadata:
  name: chat-frontend
  namespace: chat-system
spec:
  selector:
    app: chat-frontend
  ports:
  - port: 80
    targetPort: 80
  type: ClusterIP
```

### **Ingress**
```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: chat-ingress
  namespace: chat-system
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/websocket-services: "chat-backend"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
spec:
  tls:
  - hosts:
    - your-domain.com
    secretName: chat-tls
  rules:
  - host: your-domain.com
    http:
      paths:
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: chat-backend
            port:
              number: 8080
      - path: /ws
        pathType: Prefix
        backend:
          service:
            name: chat-backend
            port:
              number: 8080
      - path: /
        pathType: Prefix
        backend:
          service:
            name: chat-frontend
            port:
              number: 80
```

### **Deploy to Kubernetes**
```bash
# Apply all configurations
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f secrets.yaml
kubectl apply -f mongodb.yaml
kubectl apply -f redis.yaml
kubectl apply -f backend.yaml
kubectl apply -f frontend.yaml
kubectl apply -f ingress.yaml

# Check deployment status
kubectl get pods -n chat-system
kubectl get services -n chat-system
kubectl get ingress -n chat-system

# View logs
kubectl logs -f deployment/chat-backend -n chat-system
kubectl logs -f deployment/chat-frontend -n chat-system

# Scale deployment
kubectl scale deployment chat-backend --replicas=5 -n chat-system
```

---

## 📊 Monitoring & Logging

### **Health Checks**
```bash
# Backend health
curl http://localhost:8080/health

# Database connectivity
curl http://localhost:8080/health | jq '.database'

# Redis connectivity
curl http://localhost:8080/health | jq '.redis'
```

### **Log Aggregation**
```yaml
# fluentd-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluentd-config
data:
  fluent.conf: |
    <source>
      @type tail
      path /var/log/containers/*chat*.log
      pos_file /var/log/fluentd-containers.log.pos
      tag kubernetes.*
      format json
    </source>
    
    <match kubernetes.**>
      @type elasticsearch
      host elasticsearch
      port 9200
      index_name chat-logs
    </match>
```

### **Metrics Collection**
```yaml
# prometheus-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: prometheus-config
data:
  prometheus.yml: |
    global:
      scrape_interval: 15s
    
    scrape_configs:
    - job_name: 'chat-backend'
      static_configs:
      - targets: ['chat-backend:8080']
      metrics_path: /metrics
```

---

## 🔧 Maintenance

### **Backup Strategy**
```bash
# MongoDB backup
mongodump --uri="mongodb://username:password@host:27017/chat_system" --out=/backup/$(date +%Y%m%d)

# Redis backup
redis-cli --rdb /backup/redis-$(date +%Y%m%d).rdb

# Automated backup script
#!/bin/bash
BACKUP_DIR="/backup/$(date +%Y%m%d)"
mkdir -p $BACKUP_DIR

# MongoDB backup
mongodump --uri="$MONGODB_URI" --out="$BACKUP_DIR/mongodb"

# Redis backup
redis-cli --rdb "$BACKUP_DIR/redis.rdb"

# Compress and upload to S3
tar -czf "$BACKUP_DIR.tar.gz" "$BACKUP_DIR"
aws s3 cp "$BACKUP_DIR.tar.gz" s3://your-backup-bucket/
```

### **Update Process**
```bash
# Rolling update (Kubernetes)
kubectl set image deployment/chat-backend backend=your-registry/chat-backend:v2.0.0 -n chat-system
kubectl rollout status deployment/chat-backend -n chat-system

# Rollback if needed
kubectl rollout undo deployment/chat-backend -n chat-system

# Docker Compose update
docker-compose pull
docker-compose up -d --force-recreate --no-deps backend
```

### **Database Migrations**
```bash
# Run migrations
go run cmd/migrate/main.go up

# Rollback migrations
go run cmd/migrate/main.go down 1
```

---

## 🚨 Troubleshooting

### **Common Issues**

#### **1. WebSocket Connection Failed**
```bash
# Check backend logs
kubectl logs -f deployment/chat-backend -n chat-system

# Verify WebSocket endpoint
curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" -H "Sec-WebSocket-Key: test" -H "Sec-WebSocket-Version: 13" http://localhost:8080/ws
```

#### **2. Database Connection Issues**
```bash
# Test MongoDB connection
mongo "mongodb://username:password@host:27017/chat_system"

# Test Redis connection
redis-cli -h host -p 6379 -a password ping
```

#### **3. High Memory Usage**
```bash
# Check memory usage
kubectl top pods -n chat-system

# Analyze Go memory
go tool pprof http://localhost:8080/debug/pprof/heap
```

#### **4. Performance Issues**
```bash
# Check CPU usage
kubectl top pods -n chat-system

# Analyze Go CPU profile
go tool pprof http://localhost:8080/debug/pprof/profile
```

### **Debug Commands**
```bash
# View all resources
kubectl get all -n chat-system

# Describe problematic pod
kubectl describe pod <pod-name> -n chat-system

# Get pod logs
kubectl logs <pod-name> -n chat-system --previous

# Execute commands in pod
kubectl exec -it <pod-name> -n chat-system -- /bin/sh

# Port forward for debugging
kubectl port-forward service/chat-backend 8080:8080 -n chat-system
```

---

## 📈 Performance Tuning

### **Backend Optimization**
```go
// Connection pool tuning
clientOptions := options.Client().
    ApplyURI(uri).
    SetMaxPoolSize(100).        // Increase for high load
    SetMinPoolSize(10).
    SetMaxConnIdleTime(30 * time.Second).
    SetConnectTimeout(10 * time.Second).
    SetServerSelectionTimeout(5 * time.Second)
```

### **Redis Optimization**
```bash
# redis.conf optimizations
maxmemory 2gb
maxmemory-policy allkeys-lru
tcp-keepalive 300
timeout 300
```

### **Nginx Optimization**
```nginx
# nginx.conf optimizations
worker_processes auto;
worker_connections 4096;
keepalive_timeout 65;
keepalive_requests 1000;

# Enable gzip compression
gzip on;
gzip_types text/plain text/css application/json application/javascript;
```

---

## 🔐 Security Checklist

- [ ] Use HTTPS/WSS in production
- [ ] Implement rate limiting
- [ ] Use strong JWT secrets
- [ ] Enable database authentication
- [ ] Set up firewall rules
- [ ] Regular security updates
- [ ] Monitor for suspicious activity
- [ ] Backup encryption
- [ ] Access logging
- [ ] Input validation and sanitization