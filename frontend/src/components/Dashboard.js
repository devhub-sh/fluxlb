import React, { useState, useEffect } from 'react';
import './Dashboard.css';
import BackendCard from './BackendCard';
import AddBackendModal from './AddBackendModal';

function Dashboard({ onLogout }) {
  const [metrics, setMetrics] = useState([]);
  const [loading, setLoading] = useState(true);
  const [showAddModal, setShowAddModal] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    fetchMetrics();
    const interval = setInterval(fetchMetrics, 5000);
    return () => clearInterval(interval);
  }, []);

  const fetchMetrics = async () => {
    try {
      const response = await fetch('/api/metrics', {
        credentials: 'include'
      });

      if (response.status === 401) {
        onLogout();
        return;
      }

      if (response.ok) {
        const data = await response.json();
        setMetrics(data);
        setError('');
      }
    } catch (err) {
      setError('Failed to fetch metrics');
    } finally {
      setLoading(false);
    }
  };

  const handleAddBackend = async (url) => {
    try {
      const response = await fetch('/api/backends/add', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({ url }),
      });

      const data = await response.json();

      if (data.success) {
        setShowAddModal(false);
        fetchMetrics();
      } else {
        alert(data.message || 'Failed to add backend');
      }
    } catch (err) {
      alert('Failed to add backend');
    }
  };

  const handleRemoveBackend = async (url) => {
    if (!window.confirm(`Remove backend ${url}?`)) {
      return;
    }

    try {
      const response = await fetch('/api/backends/remove', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        credentials: 'include',
        body: JSON.stringify({ url }),
      });

      const data = await response.json();

      if (data.success) {
        fetchMetrics();
      } else {
        alert(data.message || 'Failed to remove backend');
      }
    } catch (err) {
      alert('Failed to remove backend');
    }
  };

  if (loading) {
    return <div className="loading">Loading dashboard...</div>;
  }

  return (
    <div className="dashboard">
      <div className="dashboard-header">
        <h1 className="dashboard-title">🚀 FluxLB Dashboard</h1>
        <div className="header-actions">
          <button className="add-backend-btn" onClick={() => setShowAddModal(true)}>
            + Add Backend
          </button>
          <button className="logout-btn" onClick={onLogout}>
            Logout
          </button>
        </div>
      </div>

      {error && <div className="error-banner">{error}</div>}

      <div className="metrics-summary">
        <div className="summary-card">
          <div className="summary-label">Total Backends</div>
          <div className="summary-value">{metrics.length}</div>
        </div>
        <div className="summary-card">
          <div className="summary-label">Healthy</div>
          <div className="summary-value">
            {metrics.filter(m => m.alive).length}
          </div>
        </div>
        <div className="summary-card">
          <div className="summary-label">Total Requests</div>
          <div className="summary-value">
            {metrics.reduce((sum, m) => sum + m.request_count, 0)}
          </div>
        </div>
        <div className="summary-card">
          <div className="summary-label">Active Connections</div>
          <div className="summary-value">
            {metrics.reduce((sum, m) => sum + m.active_connections, 0)}
          </div>
        </div>
      </div>

      <div className="backend-grid">
        {metrics.map((metric) => (
          <BackendCard
            key={metric.url}
            metric={metric}
            onRemove={handleRemoveBackend}
          />
        ))}
      </div>

      {metrics.length === 0 && (
        <div className="empty-state">
          <p>No backends configured</p>
          <button onClick={() => setShowAddModal(true)}>Add your first backend</button>
        </div>
      )}

      {showAddModal && (
        <AddBackendModal
          onAdd={handleAddBackend}
          onClose={() => setShowAddModal(false)}
        />
      )}

      <div className="refresh-info">
        Auto-refreshing every 5 seconds...
      </div>
    </div>
  );
}

export default Dashboard;
