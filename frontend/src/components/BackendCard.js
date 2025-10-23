import React from 'react';
import './BackendCard.css';

function BackendCard({ metric, onRemove }) {
  const formatLatency = (ns) => {
    const ms = ns / 1000000;
    return `${ms.toFixed(2)} ms`;
  };

  const formatUptime = (ns) => {
    const seconds = Math.floor(ns / 1000000000);
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;

    if (hours > 0) {
      return `${hours}h ${minutes}m ${secs}s`;
    } else if (minutes > 0) {
      return `${minutes}m ${secs}s`;
    }
    return `${secs}s`;
  };

  return (
    <div className="backend-card">
      <div className="backend-header">
        <div className="backend-url">{metric.url}</div>
        <span className={`status ${metric.alive ? 'status-up' : 'status-down'}`}>
          {metric.alive ? 'UP' : 'DOWN'}
        </span>
      </div>
      
      <div className="metrics">
        <div className="metric">
          <span className="metric-label">📊 Requests</span>
          <span className="metric-value">{metric.request_count}</span>
        </div>
        <div className="metric">
          <span className="metric-label">⚡ Avg Latency</span>
          <span className="metric-value">{formatLatency(metric.avg_latency_ns)}</span>
        </div>
        <div className="metric">
          <span className="metric-label">🔌 Active Connections</span>
          <span className="metric-value">{metric.active_connections}</span>
        </div>
        <div className="metric">
          <span className="metric-label">📈 Req/sec</span>
          <span className="metric-value">{metric.requests_per_sec.toFixed(2)}</span>
        </div>
        <div className="metric">
          <span className="metric-label">⏱️ Uptime</span>
          <span className="metric-value">{formatUptime(metric.uptime_ns)}</span>
        </div>
        <div className="metric">
          <span className="metric-label">⏲️ Time Quanta</span>
          <span className="metric-value">{formatLatency(metric.time_quanta_ns)}</span>
        </div>
      </div>
      
      <div className="backend-actions">
        <button 
          className="remove-btn" 
          onClick={() => onRemove(metric.url)}
          title="Remove backend"
        >
          Remove
        </button>
      </div>
    </div>
  );
}

export default BackendCard;
