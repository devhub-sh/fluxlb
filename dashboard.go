package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

const loginHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>FluxLB Login</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 0;
            padding: 0;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
        }
        .login-container {
            background: white;
            border-radius: 8px;
            padding: 40px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
            width: 100%;
            max-width: 400px;
        }
        h1 {
            text-align: center;
            color: #333;
            margin-bottom: 10px;
        }
        .subtitle {
            text-align: center;
            color: #666;
            margin-bottom: 30px;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            margin-bottom: 5px;
            color: #333;
            font-weight: bold;
        }
        input {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 4px;
            box-sizing: border-box;
        }
        button {
            width: 100%;
            padding: 12px;
            background: #667eea;
            color: white;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
            font-weight: bold;
        }
        button:hover {
            background: #5568d3;
        }
        .error {
            color: #ef4444;
            margin-top: 10px;
            text-align: center;
        }
        .hint {
            text-align: center;
            color: #999;
            margin-top: 20px;
            font-size: 14px;
        }
    </style>
</head>
<body>
    <div class="login-container">
        <h1>🚀 FluxLB</h1>
        <p class="subtitle">Load Balancer Dashboard</p>
        <form id="loginForm">
            <div class="form-group">
                <label for="username">Username</label>
                <input type="text" id="username" name="username" required>
            </div>
            <div class="form-group">
                <label for="password">Password</label>
                <input type="password" id="password" name="password" required>
            </div>
            <button type="submit">Login</button>
            <div id="error" class="error"></div>
        </form>
        <p class="hint">Default: admin / admin123</p>
    </div>
    <script>
        document.getElementById('loginForm').addEventListener('submit', async (e) => {
            e.preventDefault();
            const username = document.getElementById('username').value;
            const password = document.getElementById('password').value;
            const errorDiv = document.getElementById('error');
            
            try {
                const response = await fetch('/api/login', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    credentials: 'include',
                    body: JSON.stringify({ username, password })
                });
                const data = await response.json();
                
                if (data.success) {
                    window.location.href = '/dashboard';
                } else {
                    errorDiv.textContent = data.message || 'Login failed';
                }
            } catch (err) {
                errorDiv.textContent = 'Connection error. Please try again.';
            }
        });
    </script>
</body>
</html>
`

const dashboardHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>FluxLB Dashboard</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 0;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: #333;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 30px;
        }
        h1 {
            color: white;
            margin: 0;
        }
        .logout-btn {
            padding: 10px 20px;
            background: white;
            color: #667eea;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-weight: bold;
        }
        .logout-btn:hover {
            background: #f0f0f0;
        }
        .backend-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
            gap: 20px;
        }
        .backend-card {
            background: white;
            border-radius: 8px;
            padding: 20px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
        }
        .backend-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 15px;
            padding-bottom: 10px;
            border-bottom: 1px solid #eee;
        }
        .backend-url {
            font-weight: bold;
            color: #667eea;
            word-break: break-all;
        }
        .status {
            padding: 5px 10px;
            border-radius: 4px;
            font-size: 12px;
            font-weight: bold;
        }
        .status-up {
            background: #10b981;
            color: white;
        }
        .status-down {
            background: #ef4444;
            color: white;
        }
        .metric {
            display: flex;
            justify-content: space-between;
            padding: 5px 0;
        }
        .metric-label {
            color: #666;
        }
        .metric-value {
            font-weight: bold;
        }
        .refresh-info {
            text-align: center;
            color: white;
            margin-top: 20px;
        }
    </style>
    <script>
        function refreshDashboard() {
            fetch('/api/metrics')
                .then(response => {
                    if (response.status === 401) {
                        window.location.href = '/login';
                        return;
                    }
                    return response.json();
                })
                .then(data => {
                    if (data) location.reload();
                });
        }

        function logout() {
            fetch('/api/logout', { method: 'POST' })
                .then(() => {
                    window.location.href = '/login';
                });
        }

        setInterval(refreshDashboard, 5000);
    </script>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🚀 FluxLB Dashboard</h1>
            <button class="logout-btn" onclick="logout()">Logout</button>
        </div>
        <div class="backend-grid">
            {{range .}}
            <div class="backend-card">
                <div class="backend-header">
                    <div class="backend-url">{{.URL}}</div>
                    {{if .Alive}}
                    <span class="status status-up">UP</span>
                    {{else}}
                    <span class="status status-down">DOWN</span>
                    {{end}}
                </div>
                <div class="metric">
                    <span class="metric-label">Requests</span>
                    <span class="metric-value">{{.RequestCount}}</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Avg Latency</span>
                    <span class="metric-value">{{.AvgLatencyMs}}</span>
                </div>
                <div class="metric">
                    <span class="metric-label">Uptime</span>
                    <span class="metric-value">{{.UptimeStr}}</span>
                </div>
            </div>
            {{end}}
        </div>
        <div class="refresh-info">
            Auto-refreshing every 5 seconds...
        </div>
    </div>
</body>
</html>
`

// Dashboard provides the web dashboard for monitoring
type Dashboard struct {
	lb             *LoadBalancer
	dashboardTmpl  *template.Template
	loginTmpl      *template.Template
}

// NewDashboard creates a new dashboard instance
func NewDashboard(lb *LoadBalancer) (*Dashboard, error) {
	dashboardTmpl, err := template.New("dashboard").Parse(dashboardHTML)
	if err != nil {
		return nil, err
	}

	loginTmpl, err := template.New("login").Parse(loginHTML)
	if err != nil {
		return nil, err
	}

	return &Dashboard{
		lb:            lb,
		dashboardTmpl: dashboardTmpl,
		loginTmpl:     loginTmpl,
	}, nil
}

// MetricsView represents the metrics in a view-friendly format
type MetricsView struct {
	URL          string
	Alive        bool
	RequestCount int64
	AvgLatencyMs string
	UptimeStr    string
}

// ServeHTTP handles dashboard requests
func (d *Dashboard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	metrics := d.lb.GetMetrics()

	views := make([]MetricsView, 0, len(metrics))
	for _, m := range metrics {
		view := MetricsView{
			URL:          m.URL,
			Alive:        m.Alive,
			RequestCount: m.RequestCount,
			AvgLatencyMs: fmt.Sprintf("%.2f ms", float64(m.AvgLatency.Microseconds())/1000.0),
			UptimeStr:    formatDuration(m.Uptime),
		}
		views = append(views, view)
	}

	w.Header().Set("Content-Type", "text/html")
	if err := d.dashboardTmpl.Execute(w, views); err != nil {
		http.Error(w, "Error rendering dashboard", http.StatusInternalServerError)
	}
}

// ServeLogin handles login page requests
func (d *Dashboard) ServeLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	if err := d.loginTmpl.Execute(w, nil); err != nil {
		http.Error(w, "Error rendering login page", http.StatusInternalServerError)
	}
}

// ServeMetricsAPI serves metrics as JSON
func (d *Dashboard) ServeMetricsAPI(w http.ResponseWriter, r *http.Request) {
	metrics := d.lb.GetMetrics()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}
