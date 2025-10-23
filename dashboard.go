package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"
)

const dashboardHTML = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>FluxLB Dashboard</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            margin: 0;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: #333;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        h1 {
            text-align: center;
            color: white;
            margin-bottom: 30px;
            font-size: 2.5em;
            text-shadow: 2px 2px 4px rgba(0,0,0,0.3);
        }
        .backend-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
            gap: 20px;
        }
        .backend-card {
            background: white;
            border-radius: 12px;
            padding: 24px;
            box-shadow: 0 4px 6px rgba(0,0,0,0.1);
            transition: transform 0.2s;
        }
        .backend-card:hover {
            transform: translateY(-5px);
            box-shadow: 0 8px 12px rgba(0,0,0,0.2);
        }
        .backend-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 16px;
            padding-bottom: 12px;
            border-bottom: 2px solid #f0f0f0;
        }
        .backend-url {
            font-weight: bold;
            font-size: 1.1em;
            color: #667eea;
            word-break: break-all;
        }
        .status {
            padding: 6px 12px;
            border-radius: 20px;
            font-size: 0.85em;
            font-weight: bold;
            text-transform: uppercase;
        }
        .status-up {
            background-color: #10b981;
            color: white;
        }
        .status-down {
            background-color: #ef4444;
            color: white;
        }
        .metrics {
            display: grid;
            gap: 12px;
        }
        .metric {
            display: flex;
            justify-content: space-between;
            padding: 8px 0;
        }
        .metric-label {
            color: #666;
            font-weight: 500;
        }
        .metric-value {
            font-weight: bold;
            color: #333;
        }
        .refresh-info {
            text-align: center;
            color: white;
            margin-top: 20px;
            font-size: 0.9em;
        }
    </style>
    <script>
        function refreshDashboard() {
            fetch('/api/metrics')
                .then(response => response.json())
                .then(data => {
                    location.reload();
                });
        }

        // Auto-refresh every 5 seconds
        setInterval(refreshDashboard, 5000);
    </script>
</head>
<body>
    <div class="container">
        <h1>🚀 FluxLB Dashboard</h1>
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
                <div class="metrics">
                    <div class="metric">
                        <span class="metric-label">📊 Requests</span>
                        <span class="metric-value">{{.RequestCount}}</span>
                    </div>
                    <div class="metric">
                        <span class="metric-label">⚡ Avg Latency</span>
                        <span class="metric-value">{{.AvgLatencyMs}}</span>
                    </div>
                    <div class="metric">
                        <span class="metric-label">⏱️ Uptime</span>
                        <span class="metric-value">{{.UptimeStr}}</span>
                    </div>
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
	lb       *LoadBalancer
	template *template.Template
}

// NewDashboard creates a new dashboard instance
func NewDashboard(lb *LoadBalancer) (*Dashboard, error) {
	tmpl, err := template.New("dashboard").Parse(dashboardHTML)
	if err != nil {
		return nil, err
	}

	return &Dashboard{
		lb:       lb,
		template: tmpl,
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
	d.template.Execute(w, views)
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
