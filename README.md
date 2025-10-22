# FluxLB

A lightweight, extensible HTTP load balancer written in Go with a built-in web dashboard to visualize backend health, request count, and response times.

## Features

- **Round-Robin Load Balancing**: Distributes requests evenly across backend servers
- **Health Checks**: Automatic health monitoring of backend servers
- **Live Metrics**: Real-time tracking of:
  - Request count per backend
  - Average latency per backend
  - Uptime per backend
- **Web Dashboard**: Beautiful, auto-refreshing web interface for monitoring
- **JSON Configuration**: Simple configuration via JSON file
- **Graceful Shutdown**: Clean shutdown with connection draining

## Installation

### Prerequisites
- Go 1.16 or higher

### Build from Source

```bash
git clone https://github.com/devhub-sh/fluxlb.git
cd fluxlb
go build -o fluxlb
```

## Configuration

Create a `config.json` file:

```json
{
  "port": 8080,
  "health_check_path": "/health",
  "health_check_interval_seconds": 10,
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
- `health_check_path`: URL path for health checks (default: /health)
- `health_check_interval_seconds`: Interval between health checks in seconds (default: 10)
- `backends`: Array of backend server URLs

## Usage

### Start the Load Balancer

```bash
./fluxlb -config config.json
```

The load balancer will start and:
- Listen for incoming requests on the configured port
- Forward requests to backend servers using round-robin
- Perform health checks on all backends
- Provide a dashboard at `/dashboard`

### Access the Dashboard

Open your browser and navigate to:
```
http://localhost:8080/dashboard
```

The dashboard displays:
- Backend server URLs
- Health status (UP/DOWN)
- Request count
- Average latency
- Uptime

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

- `GET /` - Proxied to backend servers (load balanced)
- `GET /dashboard` - Web dashboard
- `GET /api/metrics` - JSON metrics API

### Example: Get Metrics via API

```bash
curl http://localhost:8080/api/metrics
```

Response:
```json
[
  {
    "url": "http://localhost:8081",
    "alive": true,
    "request_count": 42,
    "avg_latency_ns": 15000000,
    "uptime_ns": 300000000000
  },
  ...
]
```

## Architecture

FluxLB consists of several key components:

1. **Load Balancer**: Core routing logic with round-robin algorithm
2. **Backend**: Represents a backend server with metrics tracking
3. **Health Checker**: Periodic health monitoring of backends
4. **Dashboard**: Web interface for visualization
5. **Config Loader**: JSON configuration parser

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

## License

See [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
