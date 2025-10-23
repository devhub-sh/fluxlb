package main

import (
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

type Backend struct {
	/*
		 * @ Represents a backend server in the load balancer
			* with its URL, health status, and reverse proxy
	*/
	URL          *url.URL
	Alive        bool
	mu           sync.RWMutex
	ReverseProxy *httputil.ReverseProxy

	/*
		 * @ Metrics for monitoring
			* such as total requests and total latency
			* since the backend was added
	*/
	RequestCount      int64
	TotalLatency      time.Duration
	StartTime         time.Time
	ActiveConnections int64
	LastRequestTime   time.Time

	// Time quanta for scheduling (average processing time)
	TimeQuanta time.Duration
}

/*
 * @ Represents the metrics of a backend server
 * for monitoring purposes
 * including URL, alive status, request count,
 * average latency, and uptime
 */
type BackendMetrics struct {
	URL               string        `json:"url"`
	Alive             bool          `json:"alive"`
	RequestCount      int64         `json:"request_count"`
	AvgLatency        time.Duration `json:"avg_latency_ns"`
	Uptime            time.Duration `json:"uptime_ns"`
	ActiveConnections int64         `json:"active_connections"`
	RequestsPerSec    float64       `json:"requests_per_sec"`
	TimeQuanta        time.Duration `json:"time_quanta_ns"`
}

/*
 		* @ Creates a new Backend instance
   			* with the given URL string
      		* and initializes its reverse proxy
*/
func NewBackend(urlStr string) (*Backend, error) {
	url, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	return &Backend{
		URL:          url,
		Alive:        true,
		ReverseProxy: httputil.NewSingleHostReverseProxy(url),
		StartTime:    time.Now(),
	}, nil
}

/*
* @ Sets the backend's alive status
 */

func (b *Backend) SetAlive(alive bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.Alive = alive
}

func (b *Backend) IsAlive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Alive
}

/*
 * @ Updates the backend's metrics
 * Add request increments the request count and latency
 */

func (b *Backend) AddRequest(latency time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.RequestCount++
	b.TotalLatency += latency
	b.LastRequestTime = time.Now()

	// Update time quanta (exponential moving average)
	if b.TimeQuanta == 0 {
		b.TimeQuanta = latency
	} else {
		// EMA with alpha = 0.3
		b.TimeQuanta = time.Duration(float64(b.TimeQuanta)*0.7 + float64(latency)*0.3)
	}
}

func (b *Backend) IncrementConnections() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ActiveConnections++
}

func (b *Backend) DecrementConnections() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.ActiveConnections > 0 {
		b.ActiveConnections--
	}
}

func (b *Backend) GetMetrics() BackendMetrics {
	b.mu.RLock()
	defer b.mu.RUnlock()

	uptime := time.Since(b.StartTime)
	var avgLatency time.Duration
	if b.RequestCount > 0 {
		avgLatency = b.TotalLatency / time.Duration(b.RequestCount)
	}

	// Calculate requests per second
	var reqPerSec float64
	if uptime.Seconds() > 0 {
		reqPerSec = float64(b.RequestCount) / uptime.Seconds()
	}

	return BackendMetrics{
		URL:               b.URL.String(),
		Alive:             b.Alive,
		RequestCount:      b.RequestCount,
		AvgLatency:        avgLatency,
		Uptime:            uptime,
		ActiveConnections: b.ActiveConnections,
		RequestsPerSec:    reqPerSec,
		TimeQuanta:        b.TimeQuanta,
	}

}

func (b *Backend) GetTimeQuanta() time.Duration {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.TimeQuanta == 0 {
		return time.Millisecond * 100 // Default
	}
	return b.TimeQuanta
}

func (b *Backend) GetActiveConnections() int64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.ActiveConnections
}
