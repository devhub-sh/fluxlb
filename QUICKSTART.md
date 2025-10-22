# FluxLB Quick Start Guide

This guide will help you get FluxLB up and running in minutes.

## Prerequisites

- Go 1.16 or higher installed
- Terminal access

## Quick Start (5 minutes)

### Step 1: Build FluxLB

```bash
git clone https://github.com/devhub-sh/fluxlb.git
cd fluxlb
go build -o fluxlb
```

### Step 2: Start Test Backend Servers

Open 3 separate terminal windows and run:

**Terminal 1:**
```bash
go run examples/test_backend.go 8081
```

**Terminal 2:**
```bash
go run examples/test_backend.go 8082
```

**Terminal 3:**
```bash
go run examples/test_backend.go 8083
```

### Step 3: Start FluxLB

In a 4th terminal:
```bash
./fluxlb -config config.json
```

You should see:
```
2025/10/22 11:20:31 FluxLB starting with 3 backends on port 8080
2025/10/22 11:20:31 Added backend: http://localhost:8081
2025/10/22 11:20:31 Added backend: http://localhost:8082
2025/10/22 11:20:31 Added backend: http://localhost:8083
2025/10/22 11:20:31 FluxLB listening on http://localhost:8080
2025/10/22 11:20:31 Dashboard available at http://localhost:8080/dashboard
```

### Step 4: Test the Load Balancer

Open a 5th terminal and send some requests:

```bash
# Send 6 requests - should round-robin between backends
for i in {1..6}; do curl http://localhost:8080/; done
```

Expected output (requests alternate between backends):
```
Response from backend on port 8081
Response from backend on port 8082
Response from backend on port 8083
Response from backend on port 8081
Response from backend on port 8082
Response from backend on port 8083
```

### Step 5: View the Dashboard

Open your browser and navigate to:
```
http://localhost:8080/dashboard
```

You'll see:
- ✅ All 3 backends with green "UP" status
- 📊 Request counts for each backend
- ⚡ Average latency per backend
- ⏱️ Uptime for each backend
- Auto-refresh every 5 seconds

### Step 6: Test Health Check (Optional)

Stop one of the backend servers (press Ctrl+C in one terminal).

Wait 10 seconds for the health check to detect the failure.

Send more requests:
```bash
for i in {1..6}; do curl http://localhost:8080/; done
```

You'll notice:
- Requests only go to the 2 healthy backends
- Dashboard shows the stopped backend with red "DOWN" status

## API Endpoints

- `GET /` - Proxied to backend servers (load balanced)
- `GET /dashboard` - Web monitoring dashboard
- `GET /api/metrics` - JSON metrics API

Example metrics API call:
```bash
curl http://localhost:8080/api/metrics | jq
```

## Configuration

Edit `config.json` to customize:

```json
{
  "port": 8080,
  "health_check_path": "/health",
  "health_check_interval_seconds": 10,
  "backends": [
    { "url": "http://localhost:8081" },
    { "url": "http://localhost:8082" },
    { "url": "http://localhost:8083" }
  ]
}
```

## Running Tests

```bash
# Run all tests
go test -v

# Run with coverage
go test -v -cover
```

## Next Steps

- Customize `config.json` for your backend servers
- Deploy to production
- Monitor your backends via the dashboard
- Use the metrics API for integration with monitoring tools

## Troubleshooting

**Problem:** "connection refused" errors
- **Solution:** Make sure all backend servers are running and listening on the configured ports

**Problem:** Dashboard shows all backends as "DOWN"
- **Solution:** Check that your backends have a `/health` endpoint that returns 200 OK

**Problem:** Requests not load balancing
- **Solution:** Verify all backends are marked as "UP" in the dashboard

## Support

For issues or questions, please open an issue on GitHub.
