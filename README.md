# FluxLB

A lightweight, extensible HTTPS load balancer written in Go with a modern React dashboard for visualizing backend health, metrics, and managing backends dynamically.

## Features

- **Smart Round-Robin Load Balancing**: Time-quanta-based scheduling for optimal distribution
- **HTTPS Support**: Secure reverse proxy with TLS/SSL support
- **Authentication**: Session-based login system for dashboard access
- **Health Checks**: Automatic health monitoring of backend servers
- **Live Metrics**: Real-time tracking of:
  - Request count per backend
  - Average latency per backend
  - Requests per second
  - Active connections
  - Time quanta (processing time)
  - Uptime per backend
- **React Dashboard**: Modern, interactive web interface for monitoring and management
- **Dynamic Backend Management**: Add/remove backends from the dashboard
- **JSON Configuration**: Simple configuration via JSON file
- **Graceful Shutdown**: Clean shutdown with connection draining

## Installation

### Prerequisites
- Go 1.16 or higher
- Node.js 14+ and npm (for building the React dashboard)

### Build from Source

```bash
git clone https://github.com/devhub-sh/fluxlb.git
cd fluxlb
go build -o fluxlb
```

### Build React Dashboard (Optional)

For the full React dashboard experience:

```bash
cd frontend
npm install
npm run build
cd ..
```

The Go application will automatically detect and serve the React build if available, otherwise it falls back to a simple HTML dashboard.

## Configuration

Create a `config.json` file:

```json
{
  "port": 8080,
  "https_port": 8443,
  "enable_https": false,
  "cert_file": "certs/server.crt",
  "key_file": "certs/server.key",
  "health_check_path": "/health",
  "health_check_interval_seconds": 10,
  "auth": {
    "enabled": true,
    "username": "admin",
    "password": "admin123"
  },
  "backends": [
    {
      "url": "http://localhost:8081"
    },
    {
      "url": "http://localhost:8082"
    },
    {
      "url": "http://localhost:8083"
    }
  ]
}
```

### Configuration Options

- `port`: Port on which the load balancer listens (default: 8080)
- `https_port`: HTTPS port (default: 8443)
- `enable_https`: Enable HTTPS support (default: false)
- `cert_file`: Path to TLS certificate file
- `key_file`: Path to TLS private key file
- `health_check_path`: URL path for health checks (default: /health)
- `health_check_interval_seconds`: Interval between health checks in seconds (default: 10)
- `auth.enabled`: Enable authentication (default: true)
- `auth.username`: Dashboard username
- `auth.password`: Dashboard password
- `backends`: Array of backend server URLs

### HTTPS Setup

To enable HTTPS, you need to generate TLS certificates:

```bash
# Create certs directory
mkdir -p certs

# Generate self-signed certificate for testing
openssl req -x509 -newkey rsa:4096 -keyout certs/server.key -out certs/server.crt -days 365 -nodes -subj "/CN=localhost"

# Update config.json
# Set "enable_https": true
```

For production, use certificates from a trusted Certificate Authority (CA) like Let's Encrypt.

## Usage

### Start the Load Balancer

```bash
./fluxlb -config config.json
```

The load balancer will start and:
- Listen for incoming requests on the configured port
- Forward requests to backend servers using smart round-robin
- Perform health checks on all backends
- Provide a dashboard at `/dashboard`

### Access the Dashboard

Open your browser and navigate to:
```
http://localhost:8080/dashboard
```

**Default credentials:**
- Username: `admin`
- Password: `admin123`

The dashboard provides:
- Real-time metrics for all backends
- Add/remove backend servers dynamically
- Health status monitoring
- Request statistics and latency tracking
- Active connection monitoring

The dashboard auto-refreshes every 5 seconds.

### Testing with Mock Backends

For testing purposes, you can use the included test backend:

```bash
# Terminal 1 - Start first backend
go run examples/test_backend.go 8081

# Terminal 2 - Start second backend
go run examples/test_backend.go 8082

# Terminal 3 - Start third backend
go run examples/test_backend.go 8083

# Terminal 4 - Start FluxLB
./fluxlb -config config.json

# Terminal 5 - Send test requests
curl http://localhost:8080/
```

### API Endpoints

- `POST /api/login` - Authenticate and create session
- `POST /api/logout` - Logout and clear session
- `GET /api/metrics` - JSON metrics API (authenticated)
- `POST /api/backends/add` - Add a new backend (authenticated)
- `POST /api/backends/remove` - Remove a backend (authenticated)
- `GET /api/backends` - List all backends (authenticated)
- `GET /dashboard` - Web dashboard (authenticated)
- `GET /health` - Health check endpoint
- `GET /` - Proxied to backend servers (load balanced)

### Example: Get Metrics via API

```bash
# Login first
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  -c cookies.txt

# Get metrics
curl http://localhost:8080/api/metrics -b cookies.txt
```

Response:
```json
[
  {
    "url": "http://localhost:8081",
    "alive": true,
    "request_count": 42,
    "avg_latency_ns": 15000000,
    "uptime_ns": 300000000000,
    "active_connections": 2,
    "requests_per_sec": 0.14,
    "time_quanta_ns": 12000000
  }
]
```

### Example: Add Backend via API

```bash
curl -X POST http://localhost:8080/api/backends/add \
  -H "Content-Type: application/json" \
  -d '{"url":"http://localhost:8084"}' \
  -b cookies.txt
```

## Architecture

FluxLB consists of several key components:

1. **Load Balancer**: Core routing logic with smart round-robin algorithm
2. **Backend**: Represents a backend server with metrics tracking
3. **Health Checker**: Periodic health monitoring of backends
4. **Dashboard**: React-based web interface for visualization and management
5. **Auth Manager**: Session-based authentication system
6. **API Handler**: REST API for backend management
7. **Config Loader**: JSON configuration parser

### Smart Scheduling Algorithm

FluxLB uses a time-quanta-based scheduling algorithm that considers:
- Average processing time (time quanta) of each backend
- Current number of active connections
- Average latency

The algorithm selects the backend with the lowest score:
```
score = (time_quanta × (1 + active_connections)) + avg_latency
```

This ensures requests are distributed to the fastest and least-loaded backends.

## Development

### Run Tests

```bash
go test ./...
```

### Format Code

```bash
go fmt ./...
```

### Lint Code

```bash
go vet ./...
```

### Development Mode (React)

To develop the React dashboard with hot reload:

```bash
cd frontend
npm start
```

The React dev server will proxy API requests to the Go backend.

## Security Considerations

- Change default credentials in production
- Use HTTPS in production with valid certificates
- Keep session tokens secure (HttpOnly cookies)
- Regularly update dependencies
- Use strong passwords for authentication

## License

See [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
