# Advanced Real-Time Transaction Monitoring System

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![React](https://img.shields.io/badge/React-18.0+-61DAFB?style=for-the-badge&logo=react)](https://reactjs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15+-336791?style=for-the-badge&logo=postgresql)](https://postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=for-the-badge&logo=docker)](https://docker.com/)

A modern, full-stack fraud detection and transaction monitoring system built with Go, React, and best practices.

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   React Frontend │    │   Go Backend    │    │   PostgreSQL    │
│                 │◄──►│                 │◄──►│                 │
│ • Dashboard     │    │ • REST API      │    │ • Transactions  │
│ • Charts        │    │ • WebSocket     │    │ • Users         │
│ • Real-time UI  │    │ • Fraud Engine  │    │ • Audit Logs    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         ▲                       ▲                       ▲
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   WebSocket     │    │     Redis       │    │     Docker      │
│                 │    │                 │    │                 │
│ • Real-time     │    │ • Caching       │    │ • Containerized │
│ • Notifications │    │ • Sessions      │    │ • Scalable      │
│ • Live Updates  │    │ • Rate Limiting │    │ • Production    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.21+ (for local development)
- Node.js 18+ (for local development)
- PostgreSQL 15+ (if running locally)

### Using Docker (Recommended)

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd transaction-monitoring
   ```

2. **Start the entire stack**
   ```bash
   ./scripts/setup.sh
   ```

3. **Access the applications**
   - Frontend Dashboard: http://localhost:3000
   - Backend API: http://localhost:8080
   - Login: admin/admin123

### Local Development

1. **Database Setup**
   ```bash
   docker-compose up -d postgres redis
   psql -h localhost -U transaction_user -d transactions_db -f backend/database/migrations.sql
   ```

2. **Backend Setup**
   ```bash
   cd backend
   cp .env.example .env
   go mod tidy
   go run .
   ```

3. **Frontend Setup**
   ```bash
   cd frontend
   npm install
   npm start
   ```

## API Documentation

### Authentication
```bash
POST /api/v1/auth/login
Content-Type: application/json
{
  "username": "admin",
  "password": "admin123"
}

GET /api/v1/auth/profile
Authorization: Bearer <token>
```

### Transactions
```bash
POST /api/v1/transactions
Authorization: Bearer <token>
{
  "user_id": 1,
  "amount": 100.50,
  "currency": "USD",
  "merchant_id": "MERCHANT_001",
  "location": "New York"
}

GET /api/v1/transactions?limit=50&status=approved
Authorization: Bearer <token>

PUT /api/v1/transactions/{id}/review
Authorization: Bearer <token>
{
  "status": "approved",
  "notes": "Manual review completed"
}
```

### Alerts
```bash
GET /api/v1/alerts?severity=high&status=active
Authorization: Bearer <token>

PUT /api/v1/alerts/{id}/resolve
Authorization: Bearer <token>
{
  "status": "resolved",
  "notes": "False positive"
}
```

## Configuration

### Environment Variables

**Backend (.env)**
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=transaction_user
DB_PASSWORD=transaction_password
DB_NAME=transactions_db
REDIS_URL=redis://localhost:6379
JWT_SECRET=your-super-secret-jwt-key
SERVER_PORT=8080
ENVIRONMENT=development
```

**Frontend (.env.local)**
```env
REACT_APP_API_URL=http://localhost:8080/api/v1
REACT_APP_WS_URL=ws://localhost:8080/ws
```

## Testing

### Backend Tests
```bash
cd backend
go test ./... -v
go test -bench=. ./...
```

### Frontend Tests
```bash
cd frontend
npm test
npm run test:coverage
```

## Deployment

### Production Deployment
```bash
docker-compose up -d
./scripts/setup.sh
```

### Kubernetes Deployment
```bash
kubectl apply -f k8s/
```
