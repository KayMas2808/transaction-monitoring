# Quick Start Guide

## Method 1: One-Command Setup (Recommended)

### Prerequisites
- Docker and Docker Compose installed
- Git

### Steps
1. **Clone the repository**
   ```bash
   git clone <your-repo-url>
   cd transaction-monitoring
   ```

2. **Run the setup script**
   ```bash
   ./scripts/setup.sh
   ```

3. **Access the application**
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - Login with: `admin` / `admin123`

That's it! The script will automatically:
- Check dependencies
- Set up environment files
- Start all services with Docker
- Wait for services to be ready

## Method 2: Manual Docker Setup

### Prerequisites
- Docker and Docker Compose installed

### Steps
1. **Clone and navigate**
   ```bash
   git clone <your-repo-url>
   cd transaction-monitoring
   ```

2. **Create environment files**
   ```bash
   # Backend environment
   cat > backend/.env << EOF
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=transaction_user
   DB_PASSWORD=transaction_password
   DB_NAME=transactions_db
   REDIS_URL=redis://localhost:6379
   JWT_SECRET=your-super-secret-jwt-key-change-in-production
   SERVER_PORT=8080
   ENVIRONMENT=development
   EOF

   # Frontend environment
   cat > frontend/.env.local << EOF
   REACT_APP_API_URL=http://localhost:8080/api/v1
   REACT_APP_WS_URL=ws://localhost:8080/ws
   EOF
   ```

3. **Start services**
   ```bash
   docker-compose up -d
   ```

4. **Wait for services to start** (about 1-2 minutes)

5. **Access the application**
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080
   - Login with: `admin` / `admin123`

## Method 3: Local Development

### Prerequisites
- Go 1.21+
- Node.js 18+
- PostgreSQL
- Redis

### Steps
1. **Start infrastructure**
   ```bash
   docker-compose up -d postgres redis
   ```

2. **Setup database**
   ```bash
   psql -h localhost -U transaction_user -d transactions_db -f backend/database/migrations.sql
   ```

3. **Start backend**
   ```bash
   cd backend
   cp .env.example .env  # Edit as needed
   go mod tidy
   go run .
   ```

4. **Start frontend** (in new terminal)
   ```bash
   cd frontend
   npm install
   npm start
   ```

5. **Access the application**
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080

## Default Credentials

- **Username**: `admin`
- **Password**: `admin123`
- **Role**: Admin (full access)

## Testing the System

### 1. Create a Test Transaction
```bash
curl -X POST http://localhost:8080/api/v1/transactions \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "amount": 15000.00,
    "currency": "USD",
    "merchant_id": "TEST_MERCHANT",
    "location": "New York"
  }'
```

### 2. View Transactions
- Go to http://localhost:3000
- Login with admin credentials
- Check the dashboard for real-time updates

### 3. Test Fraud Detection
Create a high-value transaction (>$10,000) to trigger fraud alerts.

## Stopping the System

```bash
# Stop all services
docker-compose down

# Stop and remove all data
docker-compose down -v

# Or use the setup script
./scripts/setup.sh clean
```

## API Documentation
- Health check: http://localhost:8080/health
- Login: POST http://localhost:8080/api/v1/auth/login
- Transactions: GET http://localhost:8080/api/v1/transactions
- Alerts: GET http://localhost:8080/api/v1/alerts 