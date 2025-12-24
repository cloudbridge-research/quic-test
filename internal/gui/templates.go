package gui

import (
	"fmt"
	"net/http"
	"time"
)

// renderIndexHTML renders the main dashboard page
func (s *Server) renderIndexHTML(w http.ResponseWriter, data interface{}) {
	d := data.(struct {
		Title       string
		ActiveTests int
		TotalTests  int
		Uptime      time.Duration
	})
	
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <link rel="stylesheet" href="/static/css/style.css">
    <script src="/static/js/dashboard.js"></script>
</head>
<body>
    <nav class="navbar">
        <div class="nav-brand">
            <h1>QUIC Test Suite</h1>
        </div>
        <div class="nav-links">
            <a href="/" class="active">Dashboard</a>
            <a href="/test/new">New Test</a>
            <a href="/tests">Test History</a>
            <a href="/docs">Documentation</a>
            <a href="/api-docs">API Docs</a>
        </div>
    </nav>

    <main class="container">
        <div class="dashboard-header">
            <h2>Dashboard Overview</h2>
            <div class="status-indicators">
                <div class="status-card">
                    <h3>Active Tests</h3>
                    <div class="status-value">%d</div>
                </div>
                <div class="status-card">
                    <h3>Total Tests</h3>
                    <div class="status-value">%d</div>
                </div>
                <div class="status-card">
                    <h3>Uptime</h3>
                    <div class="status-value">%s</div>
                </div>
            </div>
        </div>

        <div class="dashboard-grid">
            <div class="card">
                <h3>Quick Start</h3>
                <p>Create and run QUIC performance tests with ease.</p>
                <div class="quick-actions">
                    <a href="/test/new" class="btn btn-primary">New Test</a>
                    <a href="/tests" class="btn btn-secondary">View History</a>
                </div>
            </div>

            <div class="card">
                <h3>Recent Activity</h3>
                <div id="recent-activity">
                    <p>Loading recent test activity...</p>
                </div>
            </div>

            <div class="card">
                <h3>Active Tests</h3>
                <div id="active-tests-list">
                    <p>Loading active tests...</p>
                </div>
            </div>

            <div class="card">
                <h3>System Status</h3>
                <div class="system-status">
                    <div class="status-item">
                        <span class="status-label">GUI Server</span>
                        <span class="status-indicator online">Online</span>
                    </div>
                    <div class="status-item">
                        <span class="status-label">API Server</span>
                        <span class="status-indicator online">Online</span>
                    </div>
                </div>
            </div>

            <div class="card">
                <h3>Documentation</h3>
                <p>Learn how to use the QUIC test suite effectively.</p>
                <div class="doc-links">
                    <a href="/docs">User Guide</a>
                    <a href="/api-docs">API Reference</a>
                </div>
            </div>
        </div>
    </main>

    <script>
        // Load recent activity
        fetch('/api/tests')
            .then(response => response.json())
            .then(result => {
                const container = document.getElementById('recent-activity');
                if (!result.success || result.data.tests.length === 0) {
                    container.innerHTML = '<p>No recent tests</p>';
                    return;
                }
                
                const recentTests = result.data.tests.slice(-5).reverse();
                const html = recentTests.map(test => 
                    '<div class="activity-item">' +
                    '<a href="/test/' + test.id + '" class="test-id">' + test.id + '</a>' +
                    '<span class="test-status status-' + test.status + '">' + test.status + '</span>' +
                    '<span class="test-time">' + new Date(test.start_time).toLocaleTimeString() + '</span>' +
                    '</div>'
                ).join('');
                container.innerHTML = html;
            })
            .catch(error => {
                console.error('Failed to load recent activity:', error);
                document.getElementById('recent-activity').innerHTML = '<p>Failed to load activity</p>';
            });
    </script>
</body>
</html>`, d.Title, d.ActiveTests, d.TotalTests, d.Uptime.String())
	
	w.Write([]byte(html))
}

// renderNewTestHTML renders the new test creation page
func (s *Server) renderNewTestHTML(w http.ResponseWriter, data interface{}) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Create New Test - QUIC Test Suite</title>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <nav class="navbar">
        <div class="nav-brand">
            <h1>QUIC Test Suite</h1>
        </div>
        <div class="nav-links">
            <a href="/">Dashboard</a>
            <a href="/test/new" class="active">New Test</a>
            <a href="/tests">Test History</a>
            <a href="/docs">Documentation</a>
            <a href="/api-docs">API Docs</a>
        </div>
    </nav>

    <main class="container">
        <div class="page-header">
            <h2>Create New Test</h2>
            <p>Configure and start a new QUIC performance test</p>
        </div>

        <form id="test-form" class="test-form">
            <div class="form-section">
                <h3>Basic Configuration</h3>
                <div class="form-grid">
                    <div class="form-group">
                        <label for="mode">Test Mode</label>
                        <select id="mode" name="mode" required>
                            <option value="test">Integrated (Server + Client)</option>
                            <option value="client">Client Only</option>
                            <option value="server">Server Only</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label for="duration">Duration</label>
                        <input type="text" id="duration" name="duration" value="60s" placeholder="e.g., 60s, 5m">
                    </div>
                    <div class="form-group">
                        <label for="connections">Connections</label>
                        <input type="number" id="connections" name="connections" value="2" min="1" max="100">
                    </div>
                    <div class="form-group">
                        <label for="streams">Streams per Connection</label>
                        <input type="number" id="streams" name="streams" value="4" min="1" max="100">
                    </div>
                </div>
            </div>

            <div class="form-section">
                <h3>Network Configuration</h3>
                <div class="form-grid">
                    <div class="form-group">
                        <label for="server-addr">Server Address</label>
                        <input type="text" id="server-addr" name="addr" value="localhost:9000" placeholder="host:port">
                    </div>
                    <div class="form-group">
                        <label for="packet-size">Packet Size (bytes)</label>
                        <input type="number" id="packet-size" name="packet_size" value="1200" min="64" max="65535">
                    </div>
                    <div class="form-group">
                        <label for="rate">Packet Rate (pps)</label>
                        <input type="number" id="rate" name="rate" value="100" min="1" max="10000">
                    </div>
                    <div class="form-group">
                        <label for="congestion-control">Congestion Control</label>
                        <select id="congestion-control" name="congestion_control">
                            <option value="">Default</option>
                            <option value="cubic">CUBIC</option>
                            <option value="bbr">BBR</option>
                            <option value="bbrv2">BBRv2</option>
                            <option value="bbrv3">BBRv3</option>
                            <option value="reno">NewReno</option>
                        </select>
                    </div>
                </div>
            </div>

            <div class="form-section">
                <h3>Network Emulation</h3>
                <div class="form-grid">
                    <div class="form-group">
                        <label for="emulate-latency">Additional Latency</label>
                        <input type="text" id="emulate-latency" name="emulate_latency" placeholder="e.g., 50ms">
                    </div>
                    <div class="form-group">
                        <label for="emulate-loss">Packet Loss Rate</label>
                        <input type="number" id="emulate-loss" name="emulate_loss" step="0.001" min="0" max="1" placeholder="0.01 = 1%">
                    </div>
                    <div class="form-group">
                        <label for="emulate-dup">Packet Duplication Rate</label>
                        <input type="number" id="emulate-dup" name="emulate_dup" step="0.001" min="0" max="1" placeholder="0.01 = 1%">
                    </div>
                </div>
            </div>

            <div class="form-section">
                <h3>Advanced Options</h3>
                <div class="form-grid">
                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="prometheus" name="prometheus">
                            Enable Prometheus Metrics
                        </label>
                    </div>
                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="fec-enabled" name="fec_enabled">
                            Enable Forward Error Correction
                        </label>
                    </div>
                    <div class="form-group">
                        <label for="fec-redundancy">FEC Redundancy Rate</label>
                        <input type="number" id="fec-redundancy" name="fec_redundancy" step="0.01" min="0.05" max="0.20" value="0.10" placeholder="0.10 = 10%">
                    </div>
                    <div class="form-group">
                        <label>
                            <input type="checkbox" id="pqc-enabled" name="pqc_enabled">
                            Enable Post-Quantum Crypto Simulation
                        </label>
                    </div>
                </div>
            </div>

            <div class="form-actions">
                <button type="button" id="load-preset" class="btn btn-secondary">Load Preset</button>
                <button type="submit" class="btn btn-primary">Start Test</button>
            </div>
        </form>

        <div id="preset-modal" class="modal" style="display: none;">
            <div class="modal-content">
                <div class="modal-header">
                    <h3>Load Test Preset</h3>
                    <button type="button" class="modal-close">&times;</button>
                </div>
                <div class="modal-body">
                    <div class="preset-categories">
                        <div class="preset-category">
                            <h4>Network Presets</h4>
                            <div id="network-presets"></div>
                        </div>
                        <div class="preset-category">
                            <h4>Test Profiles</h4>
                            <div id="test-profiles"></div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </main>

    <script src="/static/js/new-test.js"></script>
</body>
</html>`
	
	w.Write([]byte(html))
}

// renderTestListHTML renders the test history page
func (s *Server) renderTestListHTML(w http.ResponseWriter, data interface{}) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Test History - QUIC Test Suite</title>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <nav class="navbar">
        <div class="nav-brand">
            <h1>QUIC Test Suite</h1>
        </div>
        <div class="nav-links">
            <a href="/">Dashboard</a>
            <a href="/test/new">New Test</a>
            <a href="/tests" class="active">Test History</a>
            <a href="/docs">Documentation</a>
            <a href="/api-docs">API Docs</a>
        </div>
    </nav>

    <main class="container">
        <div class="page-header">
            <h2>Test History</h2>
            <p>View and manage your test results</p>
        </div>

        <div class="test-list-container">
            <div class="test-list-header">
                <div class="list-controls">
                    <input type="text" id="search-tests" placeholder="Search tests..." class="search-input">
                    <select id="filter-status" class="filter-select">
                        <option value="">All Status</option>
                        <option value="running">Running</option>
                        <option value="completed">Completed</option>
                        <option value="failed">Failed</option>
                        <option value="stopped">Stopped</option>
                    </select>
                </div>
            </div>

            <div id="test-list" class="test-list">
                <p>Loading test history...</p>
            </div>
        </div>
    </main>

    <script>
        function loadTestList() {
            fetch('/api/tests')
                .then(response => response.json())
                .then(result => {
                    const container = document.getElementById('test-list');
                    
                    const tests = result.success && result.data ? result.data.tests : [];
                    
                    if (tests.length === 0) {
                        container.innerHTML = '<div class="empty-state"><p>No tests found</p><a href="/test/new" class="btn btn-primary">Create First Test</a></div>';
                        return;
                    }
                    
                    // Sort tests by start time (newest first)
                    tests.sort((a, b) => new Date(b.start_time) - new Date(a.start_time));
                    
                    const html = tests.map(test => {
                        const startTime = new Date(test.start_time).toLocaleString();
                        const duration = test.end_time ? 
                            Math.round((new Date(test.end_time) - new Date(test.start_time)) / 1000) + 's' : 
                            'Running';
                        
                        return '<div class="test-item">' +
                            '<div class="test-header">' +
                            '<h3><a href="/test/' + test.id + '">' + test.id + '</a></h3>' +
                            '<span class="test-status status-' + test.status + '">' + test.status + '</span>' +
                            '</div>' +
                            '<div class="test-details">' +
                            '<span class="test-mode">' + test.config.mode + '</span>' +
                            '<span class="test-time">' + startTime + '</span>' +
                            '<span class="test-duration">' + duration + '</span>' +
                            '</div>' +
                            '</div>';
                    }).join('');
                    
                    container.innerHTML = html;
                })
                .catch(error => {
                    console.error('Failed to load test list:', error);
                    document.getElementById('test-list').innerHTML = '<p>Failed to load test history</p>';
                });
        }
        
        // Load test list on page load
        loadTestList();
        
        // Auto-refresh every 5 seconds
        setInterval(loadTestList, 5000);
        
        // Search and filter functionality
        document.getElementById('search-tests').addEventListener('input', filterTests);
        document.getElementById('filter-status').addEventListener('change', filterTests);
        
        function filterTests() {
            const searchTerm = document.getElementById('search-tests').value.toLowerCase();
            const statusFilter = document.getElementById('filter-status').value;
            const testItems = document.querySelectorAll('.test-item');
            
            testItems.forEach(item => {
                const testId = item.querySelector('h3 a').textContent.toLowerCase();
                const testStatus = item.querySelector('.test-status').textContent;
                
                const matchesSearch = testId.includes(searchTerm);
                const matchesStatus = !statusFilter || testStatus === statusFilter;
                
                item.style.display = (matchesSearch && matchesStatus) ? 'block' : 'none';
            });
        }
    </script>
</body>
</html>`
	
	w.Write([]byte(html))
}

// renderDocsHTML renders the documentation page
func (s *Server) renderDocsHTML(w http.ResponseWriter, data interface{}) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Documentation - QUIC Test Suite</title>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <nav class="navbar">
        <div class="nav-brand">
            <h1>QUIC Test Suite</h1>
        </div>
        <div class="nav-links">
            <a href="/">Dashboard</a>
            <a href="/test/new">New Test</a>
            <a href="/tests">Test History</a>
            <a href="/docs" class="active">Documentation</a>
            <a href="/api-docs">API Docs</a>
        </div>
    </nav>

    <main class="container">
        <div class="docs-layout">
            <aside class="docs-sidebar">
                <nav class="docs-nav">
                    <h3>Documentation</h3>
                    <ul>
                        <li><a href="#getting-started">Getting Started</a></li>
                        <li><a href="#test-configuration">Test Configuration</a></li>
                        <li><a href="#network-emulation">Network Emulation</a></li>
                        <li><a href="#congestion-control">Congestion Control</a></li>
                        <li><a href="#advanced-features">Advanced Features</a></li>
                        <li><a href="#metrics-analysis">Metrics Analysis</a></li>
                        <li><a href="#troubleshooting">Troubleshooting</a></li>
                    </ul>
                </nav>
            </aside>

            <div class="docs-content">
                <section id="getting-started">
                    <h2>Getting Started</h2>
                    <p>The QUIC Test Suite is a comprehensive platform for testing and analyzing QUIC protocol performance. This guide will help you get started with creating and running your first tests.</p>
                    
                    <h3>Quick Start</h3>
                    <ol>
                        <li>Navigate to the <a href="/test/new">New Test</a> page</li>
                        <li>Select a test mode (Integrated, Client, or Server)</li>
                        <li>Configure basic parameters (duration, connections, streams)</li>
                        <li>Click "Start Test" to begin</li>
                        <li>Monitor results in real-time on the test details page</li>
                    </ol>
                </section>

                <section id="test-configuration">
                    <h2>Test Configuration</h2>
                    
                    <h3>Test Modes</h3>
                    <ul>
                        <li><strong>Integrated:</strong> Runs both server and client for complete testing</li>
                        <li><strong>Client:</strong> Connects to an external QUIC server</li>
                        <li><strong>Server:</strong> Runs a QUIC server waiting for connections</li>
                    </ul>
                    
                    <h3>Basic Parameters</h3>
                    <ul>
                        <li><strong>Duration:</strong> How long to run the test (e.g., 60s, 5m, 1h)</li>
                        <li><strong>Connections:</strong> Number of parallel QUIC connections</li>
                        <li><strong>Streams:</strong> Number of streams per connection</li>
                        <li><strong>Packet Rate:</strong> Packets per second to send</li>
                        <li><strong>Packet Size:</strong> Size of each packet in bytes</li>
                    </ul>
                </section>

                <section id="network-emulation">
                    <h2>Network Emulation</h2>
                    <p>Simulate various network conditions to test QUIC performance under different scenarios.</p>
                    
                    <h3>Available Parameters</h3>
                    <ul>
                        <li><strong>Additional Latency:</strong> Add artificial delay to packets</li>
                        <li><strong>Packet Loss Rate:</strong> Percentage of packets to drop (0.01 = 1%)</li>
                        <li><strong>Packet Duplication:</strong> Percentage of packets to duplicate</li>
                    </ul>
                    
                    <h3>Common Network Profiles</h3>
                    <ul>
                        <li><strong>Fiber:</strong> Low latency (5ms), high bandwidth, minimal loss</li>
                        <li><strong>Mobile:</strong> Medium latency (50ms), moderate bandwidth, some loss</li>
                        <li><strong>Satellite:</strong> High latency (600ms), limited bandwidth, higher loss</li>
                        <li><strong>WiFi:</strong> Variable latency (20ms), good bandwidth, occasional loss</li>
                    </ul>
                </section>

                <section id="congestion-control">
                    <h2>Congestion Control</h2>
                    <p>Test different congestion control algorithms to understand their behavior and performance characteristics.</p>
                    
                    <h3>Available Algorithms</h3>
                    <ul>
                        <li><strong>CUBIC:</strong> Traditional TCP-like algorithm</li>
                        <li><strong>BBR:</strong> Bottleneck Bandwidth and RTT algorithm</li>
                        <li><strong>BBRv2:</strong> Improved version with better fairness</li>
                        <li><strong>BBRv3:</strong> Latest experimental version with dual-scale bandwidth estimation</li>
                        <li><strong>NewReno:</strong> Classic TCP NewReno algorithm</li>
                    </ul>
                    
                    <h3>BBRv3 Features</h3>
                    <ul>
                        <li>Dual-scale bandwidth model (fast/slow)</li>
                        <li>2% loss threshold</li>
                        <li>Adaptive pacing gains</li>
                        <li>Improved bufferbloat handling</li>
                    </ul>
                </section>

                <section id="advanced-features">
                    <h2>Advanced Features</h2>
                    
                    <h3>Forward Error Correction (FEC)</h3>
                    <p>Enable FEC to improve performance in lossy networks by adding redundant data that can recover lost packets.</p>
                    <ul>
                        <li>Configurable redundancy rate (5-20%)</li>
                        <li>SIMD-optimized implementation for high performance</li>
                        <li>Automatic fallback to software implementation</li>
                    </ul>
                    
                    <h3>Post-Quantum Cryptography Simulation</h3>
                    <p>Test the impact of post-quantum cryptographic algorithms on QUIC performance.</p>
                    <ul>
                        <li>ML-KEM-512/768 key encapsulation</li>
                        <li>Dilithium-2 digital signatures</li>
                        <li>Hybrid classical/post-quantum modes</li>
                    </ul>
                    
                    <h3>Prometheus Metrics</h3>
                    <p>Export detailed metrics to Prometheus for monitoring and analysis.</p>
                    <ul>
                        <li>Real-time performance metrics</li>
                        <li>HDR histogram data for accurate percentiles</li>
                        <li>Integration with Grafana dashboards</li>
                    </ul>
                </section>

                <section id="metrics-analysis">
                    <h2>Metrics Analysis</h2>
                    
                    <h3>Key Metrics</h3>
                    <ul>
                        <li><strong>Latency:</strong> Round-trip time for packets</li>
                        <li><strong>Throughput:</strong> Data transfer rate</li>
                        <li><strong>Packet Loss:</strong> Percentage of lost packets</li>
                        <li><strong>Jitter:</strong> Variation in latency</li>
                        <li><strong>Retransmissions:</strong> Number of packet retransmissions</li>
                    </ul>
                    
                    <h3>Understanding Results</h3>
                    <ul>
                        <li>Lower latency and jitter indicate better responsiveness</li>
                        <li>Higher throughput shows better bandwidth utilization</li>
                        <li>Lower packet loss and retransmissions indicate better reliability</li>
                        <li>Compare results across different configurations to identify optimal settings</li>
                    </ul>
                </section>

                <section id="troubleshooting">
                    <h2>Troubleshooting</h2>
                    
                    <h3>Common Issues</h3>
                    <ul>
                        <li><strong>Test fails to start:</strong> Check server address and port availability</li>
                        <li><strong>No metrics displayed:</strong> Ensure test is running and metrics are enabled</li>
                        <li><strong>High error rates:</strong> Check network configuration and firewall settings</li>
                        <li><strong>Poor performance:</strong> Verify system resources and network capacity</li>
                    </ul>
                    
                    <h3>Getting Help</h3>
                    <ul>
                        <li>Check test logs for detailed error messages</li>
                        <li>Review configuration parameters for correctness</li>
                        <li>Consult the API documentation for programmatic access</li>
                        <li>Use the integrated test mode for initial testing</li>
                    </ul>
                </section>
            </div>
        </div>
    </main>
</body>
</html>`
	
	w.Write([]byte(html))
}

// renderAPIDocsHTML renders the API documentation page
func (s *Server) renderAPIDocsHTML(w http.ResponseWriter, data interface{}) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>API Documentation - QUIC Test Suite</title>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <nav class="navbar">
        <div class="nav-brand">
            <h1>QUIC Test Suite</h1>
        </div>
        <div class="nav-links">
            <a href="/">Dashboard</a>
            <a href="/test/new">New Test</a>
            <a href="/tests">Test History</a>
            <a href="/docs">Documentation</a>
            <a href="/api-docs" class="active">API Docs</a>
        </div>
    </nav>

    <main class="container">
        <div class="docs-layout">
            <aside class="docs-sidebar">
                <nav class="docs-nav">
                    <h3>API Reference</h3>
                    <ul>
                        <li><a href="#overview">Overview</a></li>
                        <li><a href="#authentication">Authentication</a></li>
                        <li><a href="#test-management">Test Management</a></li>
                        <li><a href="#metrics-api">Metrics API</a></li>
                        <li><a href="#websocket-api">WebSocket API</a></li>
                        <li><a href="#examples">Examples</a></li>
                    </ul>
                </nav>
            </aside>

            <div class="docs-content">
                <section id="overview">
                    <h2>API Overview</h2>
                    <p>The QUIC Test Suite provides a comprehensive REST API for programmatic access to all testing functionality. The API is designed to be simple, consistent, and powerful.</p>
                    
                    <h3>Base URL</h3>
                    <pre><code>http://localhost:8081/api</code></pre>
                    
                    <h3>Response Format</h3>
                    <p>All API responses are in JSON format with consistent structure:</p>
                    <pre><code>{
  "success": true,
  "data": { ... },
  "error": null,
  "timestamp": "2024-01-01T12:00:00Z"
}</code></pre>
                </section>

                <section id="authentication">
                    <h2>Authentication</h2>
                    <p>Currently, the API does not require authentication. In production deployments, consider implementing API keys or OAuth2.</p>
                </section>

                <section id="test-management">
                    <h2>Test Management</h2>
                    
                    <h3>Start Test</h3>
                    <div class="api-endpoint">
                        <div class="method post">POST</div>
                        <div class="path">/api/tests</div>
                    </div>
                    <p>Create and start a new test session.</p>
                    
                    <h4>Request Body</h4>
                    <pre><code>{
  "mode": "test",
  "duration": "60s",
  "connections": 2,
  "streams": 4,
  "addr": "localhost:9000",
  "packet_size": 1200,
  "rate": 100,
  "congestion_control": "bbrv3",
  "prometheus": true,
  "fec_enabled": false,
  "fec_redundancy": 0.10
}</code></pre>
                    
                    <h4>Response</h4>
                    <pre><code>{
  "success": true,
  "data": {
    "id": "test_1704110400",
    "status": "running",
    "start_time": "2024-01-01T12:00:00Z",
    "config": { ... }
  }
}</code></pre>
                    
                    <h3>Get Test Status</h3>
                    <div class="api-endpoint">
                        <div class="method get">GET</div>
                        <div class="path">/api/tests/{id}</div>
                    </div>
                    <p>Retrieve current status and metrics for a test.</p>
                    
                    <h4>Response</h4>
                    <pre><code>{
  "success": true,
  "data": {
    "id": "test_1704110400",
    "status": "running",
    "start_time": "2024-01-01T12:00:00Z",
    "metrics": {
      "latency_ms": 45.2,
      "throughput_mbps": 125.8,
      "packet_loss": 0.01,
      "connections": 2
    },
    "logs": [
      "[12:00:01] Test started",
      "[12:00:02] Server started, beginning client test"
    ]
  }
}</code></pre>
                    
                    <h3>Stop Test</h3>
                    <div class="api-endpoint">
                        <div class="method delete">DELETE</div>
                        <div class="path">/api/tests/{id}</div>
                    </div>
                    <p>Stop a running test.</p>
                    
                    <h4>Response</h4>
                    <pre><code>{
  "success": true,
  "data": {
    "message": "Test stopped successfully"
  }
}</code></pre>
                    
                    <h3>List Tests</h3>
                    <div class="api-endpoint">
                        <div class="method get">GET</div>
                        <div class="path">/api/tests</div>
                    </div>
                    <p>Retrieve list of all tests.</p>
                    
                    <h4>Query Parameters</h4>
                    <ul>
                        <li><code>status</code> - Filter by status (running, completed, failed, stopped)</li>
                        <li><code>limit</code> - Maximum number of results (default: 50)</li>
                        <li><code>offset</code> - Number of results to skip (default: 0)</li>
                    </ul>
                </section>

                <section id="metrics-api">
                    <h2>Metrics API</h2>
                    
                    <h3>Get Current Metrics</h3>
                    <div class="api-endpoint">
                        <div class="method get">GET</div>
                        <div class="path">/api/metrics/current</div>
                    </div>
                    <p>Get current aggregated metrics from all active tests.</p>
                    
                    <h3>Get Historical Metrics</h3>
                    <div class="api-endpoint">
                        <div class="method get">GET</div>
                        <div class="path">/api/metrics/history</div>
                    </div>
                    <p>Get historical metrics data for analysis.</p>
                    
                    <h4>Query Parameters</h4>
                    <ul>
                        <li><code>test_id</code> - Specific test ID</li>
                        <li><code>start_time</code> - Start time (ISO 8601)</li>
                        <li><code>end_time</code> - End time (ISO 8601)</li>
                        <li><code>interval</code> - Data interval (1s, 5s, 1m, etc.)</li>
                    </ul>
                    
                    <h3>Prometheus Metrics</h3>
                    <div class="api-endpoint">
                        <div class="method get">GET</div>
                        <div class="path">/api/metrics/prometheus</div>
                    </div>
                    <p>Get metrics in Prometheus format for scraping.</p>
                </section>

                <section id="websocket-api">
                    <h2>WebSocket API</h2>
                    <p>Real-time metrics streaming via WebSocket connection.</p>
                    
                    <h3>Connection</h3>
                    <pre><code>ws://localhost:8081/api/ws/metrics</code></pre>
                    
                    <h3>Message Format</h3>
                    <pre><code>{
  "type": "metrics_update",
  "test_id": "test_1704110400",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "latency_ms": 45.2,
    "throughput_mbps": 125.8,
    "packet_loss": 0.01
  }
}</code></pre>
                </section>

                <section id="examples">
                    <h2>Examples</h2>
                    
                    <h3>Start a Basic Test</h3>
                    <pre><code>curl -X POST http://localhost:8081/api/tests \
  -H "Content-Type: application/json" \
  -d '{
    "mode": "test",
    "duration": "60s",
    "connections": 2,
    "streams": 4
  }'</code></pre>
                    
                    <h3>Monitor Test Progress</h3>
                    <pre><code>curl http://localhost:8081/api/tests/test_1704110400</code></pre>
                    
                    <h3>Get Prometheus Metrics</h3>
                    <pre><code>curl http://localhost:8081/api/metrics/prometheus</code></pre>
                    
                    <h3>JavaScript Example</h3>
                    <pre><code>// Start a test
const response = await fetch('/api/tests', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    mode: 'test',
    duration: '60s',
    connections: 2,
    streams: 4,
    congestion_control: 'bbrv3'
  })
});

const result = await response.json();
const testId = result.data.id;

// Monitor progress
const ws = new WebSocket('ws://localhost:8081/api/ws/metrics');
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  if (data.test_id === testId) {
    console.log('Metrics update:', data.data);
  }
};</code></pre>
                </section>
            </div>
        </div>
    </main>
</body>
</html>`
	
	w.Write([]byte(html))
}
// renderTestDetailsHTML renders the test details page
func (s *Server) renderTestDetailsHTML(w http.ResponseWriter, data interface{}) {
	d := data.(struct {
		Title   string
		Session *TestSession
	})
	
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <link rel="stylesheet" href="/static/css/style.css">
</head>
<body>
    <nav class="navbar">
        <div class="nav-brand">
            <h1>QUIC Test Suite</h1>
        </div>
        <div class="nav-links">
            <a href="/">Dashboard</a>
            <a href="/test/new">New Test</a>
            <a href="/tests">Test History</a>
            <a href="/docs">Documentation</a>
            <a href="/api-docs">API Docs</a>
        </div>
    </nav>

    <main class="container">
        <div class="page-header">
            <h2>Test Details</h2>
            <div class="test-actions">
                <button id="refresh-btn" class="btn btn-secondary">Refresh</button>
                <button id="stop-btn" class="btn btn-danger" style="display: none;">Stop Test</button>
                <a href="/tests" class="btn btn-secondary">Back to List</a>
            </div>
        </div>

        <div class="test-overview">
            <div class="test-info-card">
                <h3>Test Information</h3>
                <div class="info-grid">
                    <div class="info-item">
                        <label>Test ID:</label>
                        <span id="test-id">%s</span>
                    </div>
                    <div class="info-item">
                        <label>Status:</label>
                        <span id="test-status" class="status-indicator status-%s">%s</span>
                    </div>
                    <div class="info-item">
                        <label>Mode:</label>
                        <span id="test-mode">%s</span>
                    </div>
                    <div class="info-item">
                        <label>Started:</label>
                        <span id="test-start-time">%s</span>
                    </div>
                    <div class="info-item">
                        <label>Duration:</label>
                        <span id="test-duration">%s</span>
                    </div>
                    <div class="info-item">
                        <label>Server Address:</label>
                        <span id="test-addr">%s</span>
                    </div>
                </div>
            </div>

            <div class="test-config-card">
                <h3>Configuration</h3>
                <div class="config-grid">
                    <div class="config-item">
                        <label>Connections:</label>
                        <span>%d</span>
                    </div>
                    <div class="config-item">
                        <label>Streams:</label>
                        <span>%d</span>
                    </div>
                    <div class="config-item">
                        <label>Packet Size:</label>
                        <span>%d bytes</span>
                    </div>
                    <div class="config-item">
                        <label>Rate:</label>
                        <span>%d pps</span>
                    </div>
                    <div class="config-item">
                        <label>Prometheus:</label>
                        <span>%s</span>
                    </div>
                </div>
            </div>
        </div>

        <div class="metrics-section">
            <div class="metrics-card">
                <h3>Current Metrics</h3>
                <div class="metrics-grid" id="current-metrics">
                    <div class="metric-item">
                        <label>Latency:</label>
                        <span id="metric-latency">Loading...</span>
                    </div>
                    <div class="metric-item">
                        <label>Throughput:</label>
                        <span id="metric-throughput">Loading...</span>
                    </div>
                    <div class="metric-item">
                        <label>Packet Loss:</label>
                        <span id="metric-packet-loss">Loading...</span>
                    </div>
                    <div class="metric-item">
                        <label>Connections:</label>
                        <span id="metric-connections">Loading...</span>
                    </div>
                    <div class="metric-item">
                        <label>Elapsed Time:</label>
                        <span id="metric-elapsed">Loading...</span>
                    </div>
                </div>
            </div>

            <div class="logs-card">
                <h3>Test Logs</h3>
                <div class="logs-container" id="test-logs">
                    <p>Loading logs...</p>
                </div>
            </div>
        </div>
    </main>

    <script>
        const testId = '%s';
        let refreshInterval;

        function updateTestDetails() {
            fetch('/api/tests/' + testId)
                .then(response => response.json())
                .then(result => {
                    if (result.success && result.data) {
                        const test = result.data;
                        
                        // Update status
                        const statusElement = document.getElementById('test-status');
                        statusElement.textContent = test.status;
                        statusElement.className = 'status-indicator status-' + test.status;
                        
                        // Show/hide stop button
                        const stopBtn = document.getElementById('stop-btn');
                        if (test.status === 'running') {
                            stopBtn.style.display = 'inline-block';
                        } else {
                            stopBtn.style.display = 'none';
                        }
                        
                        // Update metrics
                        if (test.metrics) {
                            document.getElementById('metric-latency').textContent = 
                                test.metrics.latency_ms ? test.metrics.latency_ms.toFixed(1) + ' ms' : 'N/A';
                            document.getElementById('metric-throughput').textContent = 
                                test.metrics.throughput_mbps ? test.metrics.throughput_mbps.toFixed(1) + ' Mbps' : 'N/A';
                            document.getElementById('metric-packet-loss').textContent = 
                                test.metrics.packet_loss ? (test.metrics.packet_loss * 100).toFixed(2) + '%%' : 'N/A';
                            document.getElementById('metric-connections').textContent = 
                                test.metrics.connections || '0';
                            document.getElementById('metric-elapsed').textContent = 
                                test.metrics.elapsed_seconds ? test.metrics.elapsed_seconds.toFixed(1) + ' s' : 'N/A';
                        }
                        
                        // Update logs
                        if (test.logs && test.logs.length > 0) {
                            const logsHtml = test.logs.map(log => 
                                '<div class="log-entry">' + log + '</div>'
                            ).join('');
                            document.getElementById('test-logs').innerHTML = logsHtml;
                        }
                        
                        // Stop auto-refresh if test is completed
                        if (test.status !== 'running' && refreshInterval) {
                            clearInterval(refreshInterval);
                            refreshInterval = null;
                        }
                    }
                })
                .catch(error => {
                    console.error('Failed to update test details:', error);
                });
        }

        function stopTest() {
            if (confirm('Are you sure you want to stop this test?')) {
                fetch('/api/tests/' + testId, { method: 'DELETE' })
                    .then(response => response.json())
                    .then(result => {
                        if (result.success) {
                            updateTestDetails();
                        } else {
                            alert('Failed to stop test: ' + (result.error || 'Unknown error'));
                        }
                    })
                    .catch(error => {
                        console.error('Failed to stop test:', error);
                        alert('Failed to stop test');
                    });
            }
        }

        // Event listeners
        document.getElementById('refresh-btn').addEventListener('click', updateTestDetails);
        document.getElementById('stop-btn').addEventListener('click', stopTest);

        // Initial load and auto-refresh
        updateTestDetails();
        refreshInterval = setInterval(updateTestDetails, 2000); // Refresh every 2 seconds

        // Clean up interval on page unload
        window.addEventListener('beforeunload', () => {
            if (refreshInterval) {
                clearInterval(refreshInterval);
            }
        });
    </script>
</body>
</html>`, d.Title, d.Session.ID, d.Session.Status, d.Session.Status, d.Session.Config.Mode, 
		d.Session.StartTime.Format("2006-01-02 15:04:05"), 
		func() string {
			if d.Session.Config.Duration > 0 {
				return fmt.Sprintf("%.0fs", d.Session.Config.Duration.Seconds())
			}
			return "Unlimited"
		}(),
		d.Session.Config.Addr, d.Session.Config.Connections, d.Session.Config.Streams, 
		d.Session.Config.PacketSize, d.Session.Config.Rate,
		func() string {
			if d.Session.Config.Prometheus {
				return "Enabled"
			}
			return "Disabled"
		}(),
		d.Session.ID)
	
	w.Write([]byte(html))
}