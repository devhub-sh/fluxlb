#!/bin/bash

# FluxLB Setup Script

echo "🚀 FluxLB Setup"
echo "==============="

# Build Go binary
echo ""
echo "1. Building Go binary..."
go build -o fluxlb
if [ $? -ne 0 ]; then
    echo "❌ Failed to build Go binary"
    exit 1
fi
echo "✅ Go binary built successfully"

# Check if Node.js is installed
if ! command -v node &> /dev/null; then
    echo ""
    echo "⚠️  Node.js is not installed. Skipping React build."
    echo "   The application will use the legacy HTML dashboard."
    echo "   To install Node.js, visit: https://nodejs.org/"
else
    # Build React frontend
    echo ""
    echo "2. Building React frontend..."
    cd frontend
    if [ ! -d "node_modules" ]; then
        echo "   Installing npm dependencies..."
        npm install
    fi
    npm run build
    if [ $? -ne 0 ]; then
        echo "❌ Failed to build React frontend"
        cd ..
        exit 1
    fi
    cd ..
    echo "✅ React frontend built successfully"
fi

# Generate self-signed certificates if they don't exist
echo ""
echo "3. Checking HTTPS certificates..."
if [ ! -d "certs" ]; then
    mkdir -p certs
fi

if [ ! -f "certs/server.crt" ] || [ ! -f "certs/server.key" ]; then
    echo "   Generating self-signed certificates for development..."
    openssl req -x509 -newkey rsa:4096 -keyout certs/server.key -out certs/server.crt -days 365 -nodes -subj "/CN=localhost" 2>/dev/null
    if [ $? -eq 0 ]; then
        echo "✅ Certificates generated"
    else
        echo "⚠️  Failed to generate certificates. HTTPS will not be available."
        echo "   Install OpenSSL to enable HTTPS support."
    fi
else
    echo "✅ Certificates already exist"
fi

echo ""
echo "🎉 Setup complete!"
echo ""
echo "To start FluxLB:"
echo "  ./fluxlb -config config.json"
echo ""
echo "Dashboard will be available at:"
echo "  HTTP:  http://localhost:8080/dashboard"
echo "  HTTPS: https://localhost:8443/dashboard (if enabled)"
echo ""
echo "Default credentials:"
echo "  Username: admin"
echo "  Password: admin123"
echo ""
