# 🚀 Quick Deployment Guide

## Deploy Your Chat App in 10 Minutes!

This guide will help you deploy the Real-time Chat System to production quickly.

---

## Prerequisites

- Git repository (GitHub, GitLab, etc.)
- MongoDB Atlas account (free tier available)
- Vercel/Netlify account (for frontend)
- Railway/Render account (for backend) OR VPS server

---

## Option 1: Cloud Deployment (Recommended)

### Step 1: Deploy Backend to Railway

1. **Create Railway Account**
   - Go to [railway.app](https://railway.app)
   - Sign up with GitHub

2. **Create New Project**
   ```bash
   # Push your code to GitHub first
   git add .
   git commit -m "Ready for deployment"
   git push origin main
   ```

3. **Deploy on Railway**
   - Click "New Project"
   - Select "Deploy from GitHub repo"
   - Choose your repository
   - Railway will auto-detect Go application

4. **Add MongoDB**
   - In Railway project, click "New"
   - Select "Database" → "MongoDB"
   - Copy the connection string

5. **Set Environment Variables**
   ```
   MONGODB_URI=mongodb://...  (from Railway MongoDB)
   JWT_SECRET=your-super-secret-key-change-this
   PORT=8080
   CORS_ORIGINS=https://your-frontend-url.vercel.app
   ```

6. **Deploy**
   - Railway will automatically deploy
   - Note your backend URL: `https://your-app.railway.app`

### Step 2: Deploy Frontend to Vercel

1. **Create Vercel Account**
   - Go to [vercel.com](https://vercel.com)
   - Sign up with GitHub

2. **Import Project**
   - Click "New Project"
   - Import your GitHub repository
   - Select `frontend` as root directory

3. **Configure Build Settings**
   ```
   Framework Preset: Vite
   Build Command: npm run build
   Output Directory: dist
   Install Command: npm install
   ```

4. **Set Environment Variables**
   ```
   VITE_API_URL=https://your-backend.railway.app
   VITE_WS_URL=wss://your-backend.railway.app
   ```

5. **Deploy**
   - Click "Deploy"
   - Wait for build to complete
   - Your app is live! 🎉

---

## Option 2: VPS Deployment (Advanced)

### Backend Deployment

1. **SSH to your server**
   ```bash
   ssh user@your-server.com
   ```

2. **Install Dependencies**
   ```bash
   # Install Go
   wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
   sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
   export PATH=$PATH:/usr/local/go/bin
   
   # Install MongoDB
   # Follow: https://docs.mongodb.com/manual/installation/
   ```

3. **Clone and Build**
   ```bash
   git clone <your-repo>
   cd chat-app
   go build -o server cmd/server/main.go
   ```

4. **Create Systemd Service**
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
   WorkingDirectory=/path/to/chat-app
   ExecStart=/path/to/chat-app/server
   Restart=always
   Environment="MONGODB_URI=mongodb://localhost:27017/chat"
   Environment="JWT_SECRET=your-secret"
   Environment="PORT=8080"

   [Install]
   WantedBy=multi-user.target
   ```

5. **Start Service**
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable chat-backend
   sudo systemctl start chat-backend
   ```

6. **Setup Nginx Reverse Proxy**
   ```nginx
   server {
       listen 80;
       server_name api.yourdomain.com;

       location / {
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

### Frontend Deployment

1. **Build Frontend**
   ```bash
   cd frontend
   npm install
   npm run build
   ```

2. **Copy to Web Server**
   ```bash
   sudo cp -r dist/* /var/www/html/
   ```

3. **Configure Nginx**
   ```nginx
   server {
       listen 80;
       server_name yourdomain.com;
       root /var/www/html;
       index index.html;

       location / {
           try_files $uri $uri/ /index.html;
       }
   }
   ```

4. **Setup SSL (Let's Encrypt)**
   ```bash
   sudo apt install certbot python3-certbot-nginx
   sudo certbot --nginx -d yourdomain.com
   ```

---

## Option 3: Docker Deployment

### Using Docker Compose

1. **Create docker-compose.yml** (already exists)
   ```yaml
   version: '3.8'
   services:
     mongodb:
       image: mongo:6
       ports:
         - "27017:27017"
       volumes:
         - ./redis_data:/data/db

     backend:
       build: .
       ports:
         - "8080:8080"
       environment:
         - MONGODB_URI=mongodb://mongodb:27017/chat
         - JWT_SECRET=your-secret
       depends_on:
         - mongodb

     frontend:
       build: ./frontend
       ports:
         - "80:80"
       depends_on:
         - backend
   ```

2. **Deploy**
   ```bash
   docker-compose up -d
   ```

---

## MongoDB Atlas Setup (Free Tier)

1. **Create Account**
   - Go to [mongodb.com/cloud/atlas](https://www.mongodb.com/cloud/atlas)
   - Sign up for free

2. **Create Cluster**
   - Choose "Shared" (Free tier)
   - Select region closest to your users
   - Click "Create Cluster"

3. **Setup Database Access**
   - Go to "Database Access"
   - Add new database user
   - Save username and password

4. **Setup Network Access**
   - Go to "Network Access"
   - Add IP Address: `0.0.0.0/0` (allow from anywhere)
   - Or add specific IPs for better security

5. **Get Connection String**
   - Click "Connect" on your cluster
   - Choose "Connect your application"
   - Copy the connection string
   - Replace `<password>` with your database password

---

## Environment Variables Reference

### Backend (.env)
```bash
# Database
MONGODB_URI=mongodb://localhost:27017/chat
# or MongoDB Atlas:
# MONGODB_URI=mongodb+srv://user:pass@cluster.mongodb.net/chat

# JWT
JWT_SECRET=your-super-secret-key-min-32-chars

# Server
PORT=8080
HOST=0.0.0.0

# CORS
CORS_ORIGINS=http://localhost:5173,https://yourdomain.com

# Optional
LOG_LEVEL=info
```

### Frontend (.env)
```bash
# API URL (no trailing slash)
VITE_API_URL=http://localhost:8080
# or production:
# VITE_API_URL=https://api.yourdomain.com

# WebSocket URL (optional, defaults to API_URL)
VITE_WS_URL=ws://localhost:8080
# or production:
# VITE_WS_URL=wss://api.yourdomain.com
```

---

## Post-Deployment Checklist

### Security
- [ ] Change JWT_SECRET to a strong random string
- [ ] Enable HTTPS/SSL certificates
- [ ] Configure CORS properly
- [ ] Set up firewall rules
- [ ] Use environment variables (never commit secrets)
- [ ] Enable MongoDB authentication
- [ ] Restrict MongoDB network access

### Performance
- [ ] Enable gzip compression
- [ ] Set up CDN for static assets
- [ ] Configure caching headers
- [ ] Monitor server resources
- [ ] Set up database indexes (already done in code)

### Monitoring
- [ ] Set up error logging (Sentry, LogRocket)
- [ ] Monitor uptime (UptimeRobot, Pingdom)
- [ ] Track performance (New Relic, DataDog)
- [ ] Set up alerts for downtime

### Backup
- [ ] Enable MongoDB automated backups
- [ ] Set up regular database snapshots
- [ ] Document recovery procedures

---

## Testing Your Deployment

1. **Test Backend**
   ```bash
   # Health check
   curl https://your-backend.railway.app/health
   
   # Register user
   curl -X POST https://your-backend.railway.app/api/auth/register \
     -H "Content-Type: application/json" \
     -d '{"username":"test","email":"test@example.com","password":"password123"}'
   ```

2. **Test Frontend**
   - Open https://your-frontend.vercel.app
   - Register a new user
   - Create a room
   - Send messages
   - Check real-time updates

3. **Test WebSocket**
   - Open two browser windows
   - Login as different users
   - Send messages
   - Verify real-time delivery

---

## Troubleshooting

### Backend Issues

**Problem**: Can't connect to MongoDB
```bash
# Check MongoDB connection string
echo $MONGODB_URI

# Test MongoDB connection
mongosh "your-connection-string"
```

**Problem**: CORS errors
```bash
# Check CORS_ORIGINS environment variable
# Make sure it includes your frontend URL
CORS_ORIGINS=https://your-frontend.vercel.app
```

**Problem**: WebSocket connection fails
```bash
# Check if WebSocket upgrade is allowed
# Nginx: Make sure proxy_set_header Upgrade is set
# Railway/Render: WebSocket should work by default
```

### Frontend Issues

**Problem**: API calls fail
```bash
# Check VITE_API_URL in .env
# Make sure it points to your backend
# Check browser console for CORS errors
```

**Problem**: WebSocket won't connect
```bash
# Check VITE_WS_URL
# Use wss:// for HTTPS sites
# Check browser console for errors
```

**Problem**: Build fails
```bash
# Clear cache and rebuild
rm -rf node_modules dist
npm install
npm run build
```

---

## Quick Commands

### Start Development
```bash
# Backend
make docker-up
make run

# Frontend
cd frontend
npm run dev
```

### Build for Production
```bash
# Backend
go build -o server cmd/server/main.go

# Frontend
cd frontend
npm run build
```

### Check Logs
```bash
# Railway: View in dashboard
# VPS: 
sudo journalctl -u chat-backend -f

# Docker:
docker-compose logs -f
```

---

## Support & Resources

- **Documentation**: See `docs/` folder
- **Issues**: Check GitHub issues
- **MongoDB Atlas**: [docs.mongodb.com](https://docs.mongodb.com)
- **Railway**: [docs.railway.app](https://docs.railway.app)
- **Vercel**: [vercel.com/docs](https://vercel.com/docs)

---

## 🎉 Congratulations!

Your chat app is now live! Share the URL with your users and start chatting!

**Frontend**: https://your-app.vercel.app  
**Backend**: https://your-api.railway.app

---

**Need help?** Check the full deployment guide in `docs/DEPLOYMENT.md`
