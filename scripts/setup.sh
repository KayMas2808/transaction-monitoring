#!/bin/bash

set -e

echo "Setting up Transaction Monitoring System..."

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_dependencies() {
    print_status "Checking dependencies..."
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        print_success "Go $GO_VERSION is installed"
    else
        print_warning "Go is not installed. You'll need it for local backend development."
    fi
    
    if command -v node &> /dev/null; then
        NODE_VERSION=$(node --version)
        print_success "Node.js $NODE_VERSION is installed"
    else
        print_warning "Node.js is not installed. You'll need it for local frontend development."
    fi
    
    print_success "All required dependencies are available"
}

create_directories() {
    print_status "Creating necessary directories..."
    
    mkdir -p logs
    
    print_success "Directories created"
}

setup_environment() {
    print_status "Setting up environment files..."
    
    if [ ! -f backend/.env ]; then
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
        print_success "Created backend/.env file"
    else
        print_warning "backend/.env already exists, skipping..."
    fi
    
    if [ ! -f frontend/.env.local ]; then
        cat > frontend/.env.local << EOF
REACT_APP_API_URL=http://localhost:8080/api/v1
REACT_APP_WS_URL=ws://localhost:8080/ws
EOF
        print_success "Created frontend/.env.local file"
    else
        print_warning "frontend/.env.local already exists, skipping..."
    fi
}

start_services() {
    print_status "Building and starting services..."
    
    docker-compose pull
    docker-compose build
    docker-compose up -d
    
    print_success "Services started successfully"
}

wait_for_services() {
    print_status "Waiting for services to be ready..."
    
    print_status "Waiting for PostgreSQL..."
    until docker-compose exec postgres pg_isready -U transaction_user > /dev/null 2>&1; do
        sleep 1
    done
    print_success "PostgreSQL is ready"
    
    print_status "Waiting for backend API..."
    until curl -f http://localhost:8080/health > /dev/null 2>&1; do
        sleep 2
    done
    print_success "Backend API is ready"
    
    print_status "Checking frontend..."
    if curl -f http://localhost:3000 > /dev/null 2>&1; then
        print_success "Frontend is ready"
    else
        print_warning "Frontend may still be starting up..."
    fi
}

show_access_info() {
echo "[SUCCESS] Transaction Monitoring System is ready!"
echo ""
echo "Access the application here:"
echo "Application URL:      http://localhost:3000"
echo ""
echo "Default login credentials:"
echo "Username: admin"
echo "Password: admin123"

}

case "${1:-full}" in
    "deps")
        check_dependencies
        ;;
    "env")
        setup_environment
        ;;
    "build")
        start_services
        ;;
    "full")
        check_dependencies
        create_directories
        setup_environment
        start_services
        wait_for_services
        show_access_info
        ;;
    "clean")
        print_status "Cleaning up..."
        docker-compose down -v
        docker system prune -f
        print_success "Cleanup completed"
        ;;
    *)
        echo "Usage: $0 [deps|env|build|full|clean]"
        echo "  deps  - Check dependencies only"
        echo "  env   - Setup environment files only"
        echo "  build - Build and start services only"
        echo "  full  - Complete setup (default)"
        echo "  clean - Clean up everything"
        exit 1
        ;;
esac 
