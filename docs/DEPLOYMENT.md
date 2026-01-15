# Deployment Guide

## Overview

This guide covers deploying the Real-time Chat System to production.

## Prerequisites

- Server with Ubuntu 20.04+ or similar
- Domain name (optional but recommended)
- SSL certificate (Let's Encrypt recommended)
- MongoDB instance (Atlas or self-hosted)

## Backend Deployment

### Option 1: Docker Deployment (Recommended)

1. **Create Dockerfile**

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/server .
COPY .env.example .env

EXPOSE 8080
CMD ["./server"]
```

2. **Build and run**

```bash
docker build -t chat-backend .
docker run -p 8080:8080 --env-file .env chat-backend
```

### Option 2: Systemd Service

1. **Build the binary**

```bash
make build
```

2. **Create systemd service**

```bash
sudo nano /etc/systemd/system/chat-backend.service
```

```ini
[Unit]
Description=Chat Backend Service
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/chat-backend
ExecStart=/opt/chat-backend/bin/server
Restart=always
RestartSec=5
Environment="MONGODB_URI=mongodb://localhost:27017"
Environment="JWT_SECRET=your-secret-key"

[Install]
WantedBy=multi-user.target
```

3. **Start service**

```bash
sudo systemctl daemon-reload
sudo systemctl enable chat-backend
sudo systemctl start chat-backend
```

### Option 3: Cloud Platforms

#### AWS EC2
1. Launch EC2 instance (t3.small or larger)
2. Install Go and MongoDB
3. Clone repository
4. Build and run with systemd

#### Google Cloud Run
1. Build Docker image
2. Push to Google Container Registry
3. Deploy to Cloud Run
4. Configure environment variables

#### Heroku
1. Create Heroku app
2. Add MongoDB addon
3. Push code
4. Configure environment variables

## Frontend Deployment

### Option 1: Vercel (Recommended)

1. **Install Vercel CLI**

```bash
npm install -g vercel
```

2. **Deploy**

```bash
cd frontend
vercel
```

3. **Configure environment variables in Vercel dashboard**

### Option 2: Netlify

1. **Build the app**

```bash
cd frontend
npm run build
```

2. **Deploy to Netlify**

```bash
npm install -g netlify-cli
netlify deploy --prod --dir=dist
```

### Option 3: Nginx Static Hosting

1. **Build the app**

```bash
cd frontend
npm run build
```

2. **Copy to web server**

```bash
sudo cp -r dist/* /var/www/html/chat/
```

3. **Configure Nginx**

```nginx
server {
    listen 80;
    server_name chat.example.com;

    root /var/www/html/chat;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

    location /ws {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Host $host;
    }
}
```

## Database Setup

### MongoDB Atlas (Recommended)

1. Create account at https://www.mongodb.com/cloud/atlas
2. Create cluster
3. Create database user
4. Whitelist IP addresses
5. Get connection string
6. Update `MONGODB_URI` in environment variables

### Self-hosted MongoDB

1. **Install MongoDB**

```bash
wget -qO - https://www.mongodb.org/static/pgp/server-6.0.asc | sudo apt-key add -
echo "deb [ arch=amd64,arm64 ] https://repo.mongodb.org/apt/ubuntu focal/mongodb-org/6.0 multiverse" | sudo tee /etc/apt/sources.list.d/mongodb-org-6.0.list
sudo apt-get update
sudo apt-get install -y mongodb-org
```

2. **Start MongoDB**

```bash
sudo systemctl start mongod
sudo systemctl enable mongod
```

3. **Secure MongoDB**

```bash
mongosh
use admin
db.createUser({
  user: "chatadmin",
  pwd: "secure-password",
  roles: ["readWriteAnyDatabase"]
})
```

## Environment Variables

### Backend (.env)

```env
# Server
SERVER_PORT=8080
SERVER_READ_TIMEOUT=10s
SERVER_WRITE_TIMEOUT=10s
SERVER_SHUTDOWN_TIMEOUT=30s

# Database
MONGODB_URI=mongodb://username:password@host:27017
MONGODB_DATABASE=chat
MONGODB_TIMEOUT=10s

# JWT
JWT_SECRET=your-very-secure-secret-key-change-this
JWT_EXPIRATION=24h
```

### Frontend (.env)

```env
VITE_API_URL=https://api.chat.example.com
VITE_WS_URL=wss://api.chat.example.com
```

## SSL/TLS Setup

### Let's Encrypt with Certbot

```bash
sudo apt-get install certbot python3-certbot-nginx
sudo certbot --nginx -d chat.example.com
```

## Monitoring

### Backend Logging

Logs are written to stdout. Capture with:

```bash
# Systemd
sudo journalctl -u chat-backend -f

# Docker
docker logs -f container-name
```

### Health Checks

```bash
curl http://localhost:8080/health
```

### Monitoring Tools

- **Prometheus**: Metrics collection
- **Grafana**: Visualization
- **Sentry**: Error tracking
- **Datadog**: Full-stack monitoring

## Scaling

### Horizontal Scaling

For multiple instances, you'll need:

1. **Redis for presence tracking**
   - Replace in-memory presence with Redis
   - Use Redis Pub/Sub for message broadcasting

2. **Load balancer**
   - Nginx or HAProxy
   - Sticky sessions for WebSocket

3. **Shared session store**
   - Redis for JWT token blacklist (if needed)

### Vertical Scaling

- Increase server resources (CPU, RAM)
- Optimize MongoDB indexes
- Enable MongoDB connection pooling

## Backup

### MongoDB Backup

```bash
# Backup
mongodump --uri="mongodb://username:password@host:27017/chat" --out=/backup/$(date +%Y%m%d)

# Restore
mongorestore --uri="mongodb://username:password@host:27017/chat" /backup/20260114
```

### Automated Backups

```bash
# Add to crontab
0 2 * * * /usr/local/bin/backup-mongodb.sh
```

## Security Checklist

- [ ] Use HTTPS/WSS in production
- [ ] Set strong JWT secret
- [ ] Enable MongoDB authentication
- [ ] Whitelist IP addresses
- [ ] Use environment variables for secrets
- [ ] Enable rate limiting
- [ ] Set up firewall rules
- [ ] Regular security updates
- [ ] Monitor for suspicious activity
- [ ] Implement CORS properly

## Performance Optimization

### Backend
- Enable gzip compression
- Use connection pooling
- Implement caching (Redis)
- Optimize database queries
- Add indexes to MongoDB

### Frontend
- Enable code splitting
- Optimize bundle size
- Use CDN for static assets
- Enable browser caching
- Compress images

## Troubleshooting

### Backend won't start
- Check MongoDB connection
- Verify environment variables
- Check port availability
- Review logs

### WebSocket connection fails
- Verify WSS configuration
- Check firewall rules
- Verify proxy settings
- Check CORS configuration

### High memory usage
- Check for memory leaks
- Optimize MongoDB queries
- Implement pagination
- Add connection limits

## Rollback Plan

1. Keep previous version binary
2. Database migrations should be reversible
3. Test rollback procedure
4. Document rollback steps

## Post-Deployment

1. Verify all endpoints work
2. Test WebSocket connections
3. Check database connectivity
4. Monitor error rates
5. Set up alerts
6. Document any issues

## Support

For deployment issues:
1. Check logs
2. Review documentation
3. Check GitHub issues
4. Contact support team

---

**Last Updated**: January 14, 2026
