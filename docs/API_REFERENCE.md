# API Reference

Complete API documentation for the QUIC Test Suite.

## Table of Contents

1. [Overview](#overview)
2. [Authentication](#authentication)
3. [Response Format](#response-format)
4. [Error Handling](#error-handling)
5. [Test Management API](#test-management-api)
6. [Metrics API](#metrics-api)
7. [Configuration API](#configuration-api)
8. [System API](#system-api)
9. [WebSocket API](#websocket-api)
10. [WebTransport API](#webtransport-api)
11. [HTTP/3 Load Testing API](#http3-load-testing-api)
12. [Examples](#examples)
13. [SDKs and Libraries](#sdks-and-libraries)

## Overview

The QUIC Test Suite provides a comprehensive REST API for programmatic access to all testing functionality. The API follows RESTful principles and returns JSON responses.

**Base URL:** `http://localhost:8081/api`

**API Version:** v1

**Content-Type:** `application/json`

## Authentication

Currently, the API does not require authentication for local development. For production deployments, implement one of the following:

- API Keys via `Authorization: Bearer <token>` header
- OAuth2 with JWT tokens
- mTLS client certificates

## Response Format

All API responses follow a consistent JSON structure:

```json
{
  "success": true,
  "data": { ... },
  "error": null,
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req_123456789"
}
```

### Success Response
```json
{
  "success": true,
  "data": {
    "id": "test_1704110400",
    "status": "running"
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Error Response
```json
{
  "success": false,
  "error": "Invalid configuration: duration must be positive",
  "error_code": "INVALID_CONFIG",
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## Error Handling

### HTTP Status Codes

| Code | Description |
|------|-------------|
| 200 | OK - Request successful |
| 201 | Created - Resource created successfully |
| 400 | Bad Request - Invalid request parameters |
| 401 | Unauthorized - Authentication required |
| 403 | Forbidden - Insufficient permissions |
| 404 | Not Found - Resource not found |
| 409 | Conflict - Resource already exists |
| 422 | Unprocessable Entity - Validation failed |
| 429 | Too Many Requests - Rate limit exceeded |
| 500 | Internal Server Error - Server error |
| 503 | Service Unavailable - Service temporarily unavailable |

### Error Codes

| Code | Description |
|------|-------------|
| `INVALID_CONFIG` | Test configuration validation failed |
| `TEST_NOT_FOUND` | Specified test ID does not exist |
| `TEST_NOT_RUNNING` | Operation requires test to be running |
| `RESOURCE_LIMIT` | Maximum resource limit reached |
| `NETWORK_ERROR` | Network connectivity issue |
| `TIMEOUT` | Operation timed out |

## Test Management API

### Create Test

Create and start a new test session.

**Endpoint:** `POST /api/tests`

**Request Body:**
```json
{
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
  "fec_redundancy": 0.10,
  "pqc_enabled": false,
  "pqc_algorithm": "ml-kem-768",
  "emulate_latency": "50ms",
  "emulate_loss": 0.01,
  "emulate_dup": 0.005,
  "network_profile": "mobile",
  "scenario": "standard"
}
```

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `mode` | string | Yes | Test mode: `test`, `client`, `server` |
| `duration` | string | Yes | Test duration (e.g., `60s`, `5m`, `1h`) |
| `connections` | integer | Yes | Number of QUIC connections (1-100) |
| `streams` | integer | Yes | Streams per connection (1-100) |
| `addr` | string | No | Server address (default: `localhost:9000`) |
| `packet_size` | integer | No | Packet size in bytes (64-65535, default: 1200) |
| `rate` | integer | No | Packet rate per second (1-10000, default: 100) |
| `congestion_control` | string | No | Algorithm: `cubic`, `bbr`, `bbrv2`, `bbrv3`, `reno` |
| `prometheus` | boolean | No | Enable Prometheus metrics export |
| `fec_enabled` | boolean | No | Enable Forward Error Correction |
| `fec_redundancy` | float | No | FEC redundancy rate (0.05-0.20, default: 0.10) |
| `pqc_enabled` | boolean | No | Enable Post-Quantum Crypto simulation |
| `pqc_algorithm` | string | No | PQC algorithm: `ml-kem-512`, `ml-kem-768`, `dilithium-2`, `hybrid` |
| `emulate_latency` | string | No | Additional latency (e.g., `50ms`) |
| `emulate_loss` | float | No | Packet loss rate (0.0-1.0) |
| `emulate_dup` | float | No | Packet duplication rate (0.0-1.0) |
| `network_profile` | string | No | Network profile: `fiber`, `mobile`, `satellite`, `wifi` |
| `scenario` | string | No | Test scenario: `quick`, `standard`, `intensive`, `endurance` |

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "test_1704110400",
    "status": "running",
    "start_time": "2024-01-01T12:00:00Z",
    "config": {
      "mode": "test",
      "duration": "60s",
      "connections": 2,
      "streams": 4
    }
  }
}
```

### Get Test Status

Retrieve current status and metrics for a specific test.

**Endpoint:** `GET /api/tests/{id}`

**Path Parameters:**
- `id` (string, required): Test ID

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "test_1704110400",
    "status": "running",
    "start_time": "2024-01-01T12:00:00Z",
    "end_time": null,
    "config": {
      "mode": "test",
      "duration": "60s",
      "connections": 2,
      "streams": 4,
      "congestion_control": "bbrv3"
    },
    "metrics": {
      "latency_ms": 45.2,
      "throughput_mbps": 125.8,
      "packet_loss": 0.01,
      "connections": 2,
      "streams": 8,
      "bytes_sent": 1048576,
      "bytes_received": 1048576,
      "retransmits": 12,
      "errors": 0,
      "rtt_p50_ms": 42.1,
      "rtt_p95_ms": 58.3,
      "rtt_p99_ms": 67.8,
      "jitter_ms": 3.2,
      "handshake_time_ms": 15.6,
      "bbr_phase": "ProbeBW",
      "bbr_bandwidth_mbps": 130.2,
      "fec_packets_sent": 1024,
      "fec_recovered": 8,
      "pqc_handshake_size": 1184,
      "pqc_handshake_time_ms": 12.3
    },
    "logs": [
      "[12:00:01] Test started",
      "[12:00:02] Server started, beginning client test",
      "[12:00:05] BBRv3 entered ProbeBW phase",
      "[12:00:10] FEC recovered 3 packets"
    ]
  }
}
```

### List Tests

Retrieve list of all tests with optional filtering and pagination.

**Endpoint:** `GET /api/tests`

**Query Parameters:**
- `status` (string, optional): Filter by status (`running`, `completed`, `failed`, `stopped`)
- `mode` (string, optional): Filter by test mode (`test`, `client`, `server`)
- `limit` (integer, optional): Maximum results per page (default: 50, max: 200)
- `offset` (integer, optional): Number of results to skip (default: 0)
- `sort` (string, optional): Sort field (`start_time`, `duration`, `status`)
- `order` (string, optional): Sort order (`asc`, `desc`, default: `desc`)

**Response:**
```json
{
  "success": true,
  "data": {
    "tests": [
      {
        "id": "test_1704110400",
        "status": "completed",
        "start_time": "2024-01-01T12:00:00Z",
        "end_time": "2024-01-01T12:01:00Z",
        "config": {
          "mode": "test",
          "duration": "60s"
        }
      }
    ],
    "total": 25,
    "limit": 50,
    "offset": 0,
    "has_more": false
  }
}
```

### Stop Test

Stop a running test.

**Endpoint:** `DELETE /api/tests/{id}`

**Path Parameters:**
- `id` (string, required): Test ID

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Test stopped successfully",
    "stopped_at": "2024-01-01T12:00:30Z"
  }
}
```

### Update Test Configuration

Update configuration of a running test (limited parameters).

**Endpoint:** `PATCH /api/tests/{id}`

**Path Parameters:**
- `id` (string, required): Test ID

**Request Body:**
```json
{
  "rate": 200,
  "emulate_loss": 0.02
}
```

**Updatable Parameters:**
- `rate`: Packet rate per second
- `emulate_loss`: Packet loss rate
- `emulate_latency`: Additional latency
- `emulate_dup`: Packet duplication rate

## Metrics API

### Get Current Metrics

Get current aggregated metrics from all active tests.

**Endpoint:** `GET /api/metrics/current`

**Response:**
```json
{
  "success": true,
  "data": {
    "timestamp": "2024-01-01T12:00:00Z",
    "active_tests": 3,
    "total_connections": 6,
    "avg_latency_ms": 45.2,
    "total_throughput_mbps": 378.4,
    "avg_packet_loss": 0.008,
    "total_errors": 2,
    "system_metrics": {
      "cpu_usage": 15.2,
      "memory_usage_mb": 256,
      "network_rx_mbps": 180.5,
      "network_tx_mbps": 185.2
    }
  }
}
```

### Get Historical Metrics

Get historical metrics data for analysis and visualization.

**Endpoint:** `GET /api/metrics/history`

**Query Parameters:**
- `test_id` (string, optional): Specific test ID
- `start_time` (string, optional): Start time (ISO 8601 format)
- `end_time` (string, optional): End time (ISO 8601 format)
- `interval` (string, optional): Data interval (`1s`, `5s`, `1m`, `5m`, default: `5s`)
- `metrics` (string, optional): Comma-separated list of metrics to include

**Response:**
```json
{
  "success": true,
  "data": {
    "test_id": "test_1704110400",
    "start_time": "2024-01-01T12:00:00Z",
    "end_time": "2024-01-01T12:01:00Z",
    "interval": "5s",
    "metrics": [
      {
        "timestamp": "2024-01-01T12:00:00Z",
        "latency_ms": 45.2,
        "throughput_mbps": 125.8,
        "packet_loss": 0.01,
        "rtt_p95_ms": 58.3,
        "bbr_phase": "Startup"
      },
      {
        "timestamp": "2024-01-01T12:00:05Z",
        "latency_ms": 47.1,
        "throughput_mbps": 128.3,
        "packet_loss": 0.008,
        "rtt_p95_ms": 56.7,
        "bbr_phase": "ProbeBW"
      }
    ]
  }
}
```

### Get Prometheus Metrics

Get metrics in Prometheus format for scraping.

**Endpoint:** `GET /api/metrics/prometheus`

**Response:** (Content-Type: `text/plain`)
```
# HELP quic_test_active_tests Number of active tests
# TYPE quic_test_active_tests gauge
quic_test_active_tests 3

# HELP quic_test_latency_ms Current latency in milliseconds
# TYPE quic_test_latency_ms gauge
quic_test_latency_ms{test_id="test_1704110400"} 45.20

# HELP quic_test_throughput_mbps Current throughput in Mbps
# TYPE quic_test_throughput_mbps gauge
quic_test_throughput_mbps{test_id="test_1704110400"} 125.80

# HELP quic_test_packet_loss Packet loss rate
# TYPE quic_test_packet_loss gauge
quic_test_packet_loss{test_id="test_1704110400"} 0.0100

# HELP quic_test_connections Number of active connections
# TYPE quic_test_connections gauge
quic_test_connections{test_id="test_1704110400"} 2

# HELP quic_test_rtt_seconds RTT histogram
# TYPE quic_test_rtt_seconds histogram
quic_test_rtt_seconds_bucket{test_id="test_1704110400",le="0.01"} 0
quic_test_rtt_seconds_bucket{test_id="test_1704110400",le="0.05"} 1250
quic_test_rtt_seconds_bucket{test_id="test_1704110400",le="0.1"} 1890
quic_test_rtt_seconds_bucket{test_id="test_1704110400",le="+Inf"} 2000
quic_test_rtt_seconds_sum{test_id="test_1704110400"} 90.4
quic_test_rtt_seconds_count{test_id="test_1704110400"} 2000
```

### Export Test Results

Export test results in various formats.

**Endpoint:** `GET /api/tests/{id}/export`

**Path Parameters:**
- `id` (string, required): Test ID

**Query Parameters:**
- `format` (string, required): Export format (`json`, `csv`, `markdown`, `prometheus`)
- `include_raw` (boolean, optional): Include raw packet data (default: false)

**Response:** (Content varies by format)

## Configuration API

### Get Network Presets

Get available network configuration presets.

**Endpoint:** `GET /api/config/presets`

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "name": "fiber",
      "description": "Fiber optic connection (low latency, high bandwidth)",
      "latency": "5ms",
      "bandwidth": "1000Mbps",
      "loss": "0.01%",
      "jitter": "1ms"
    },
    {
      "name": "mobile",
      "description": "4G/LTE mobile connection",
      "latency": "50ms",
      "bandwidth": "50Mbps",
      "loss": "1%",
      "jitter": "10ms"
    },
    {
      "name": "satellite",
      "description": "Satellite connection (high latency)",
      "latency": "600ms",
      "bandwidth": "25Mbps",
      "loss": "2%",
      "jitter": "50ms"
    }
  ]
}
```

### Get Test Profiles

Get available test configuration profiles.

**Endpoint:** `GET /api/config/profiles`

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "name": "quick",
      "description": "Quick performance test (30 seconds)",
      "duration": "30s",
      "connections": 1,
      "streams": 2,
      "rate": 100
    },
    {
      "name": "standard",
      "description": "Standard performance test (2 minutes)",
      "duration": "120s",
      "connections": 2,
      "streams": 4,
      "rate": 200
    },
    {
      "name": "intensive",
      "description": "Intensive load test (5 minutes)",
      "duration": "300s",
      "connections": 4,
      "streams": 8,
      "rate": 500
    }
  ]
}
```

### Get Congestion Control Algorithms

Get available congestion control algorithms and their parameters.

**Endpoint:** `GET /api/config/congestion-control`

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "name": "bbrv3",
      "description": "BBRv3 with dual-scale bandwidth estimation",
      "parameters": {
        "loss_threshold": 0.02,
        "beta": 0.7,
        "headroom_fraction": 0.15,
        "startup_pacing_gain": 2.77,
        "drain_pacing_gain": 0.35
      },
      "recommended_for": ["high-bandwidth", "long-distance"]
    },
    {
      "name": "bbrv2",
      "description": "BBRv2 with improved fairness",
      "parameters": {
        "probe_bw_gain": 1.25,
        "probe_rtt_duration": "200ms"
      },
      "recommended_for": ["general-purpose", "mixed-traffic"]
    }
  ]
}
```

## System API

### Get System Status

Get system status and resource usage information.

**Endpoint:** `GET /api/system/status`

**Response:**
```json
{
  "success": true,
  "data": {
    "uptime": "2h15m30s",
    "version": "1.0.0",
    "build_time": "2024-01-01T00:00:00Z",
    "git_commit": "abc123def456",
    "active_tests": 3,
    "total_tests": 25,
    "system_resources": {
      "cpu_usage": 15.2,
      "memory_usage_mb": 256,
      "memory_total_mb": 8192,
      "disk_usage_mb": 1024,
      "disk_total_mb": 102400,
      "network_interfaces": [
        {
          "name": "eth0",
          "rx_bytes": 1048576000,
          "tx_bytes": 1048576000,
          "rx_packets": 1000000,
          "tx_packets": 1000000
        }
      ]
    },
    "features": {
      "fec_simd": true,
      "bbrv3": true,
      "pqc_simulation": true,
      "webtransport": true,
      "http3_load_testing": true
    }
  }
}
```

### Health Check

Simple health check endpoint for monitoring.

**Endpoint:** `GET /api/system/health`

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "timestamp": "2024-01-01T12:00:00Z",
    "checks": {
      "api_server": "ok",
      "test_manager": "ok",
      "metrics_collector": "ok",
      "database": "ok"
    },
    "response_time_ms": 2.5
  }
}
```

## WebSocket API

### Real-time Metrics Stream

Connect to WebSocket for real-time metrics updates.

**Endpoint:** `ws://localhost:8081/api/ws/metrics`

**Query Parameters:**
- `test_id` (string, optional): Subscribe to specific test
- `interval` (string, optional): Update interval (`1s`, `5s`, default: `1s`)

**Message Format:**
```json
{
  "type": "metrics_update",
  "test_id": "test_1704110400",
  "timestamp": "2024-01-01T12:00:00Z",
  "data": {
    "latency_ms": 45.2,
    "throughput_mbps": 125.8,
    "packet_loss": 0.01,
    "connections": 2,
    "bbr_phase": "ProbeBW"
  }
}
```

**Event Types:**
- `metrics_update`: Regular metrics update
- `test_started`: New test started
- `test_completed`: Test completed
- `test_failed`: Test failed
- `alert`: System alert or warning

## WebTransport API

### Create WebTransport Session

Create a new WebTransport session for testing.

**Endpoint:** `POST /api/webtransport/sessions`

**Request Body:**
```json
{
  "url": "https://example.com:4433/webtransport",
  "duration": "60s",
  "streams": 4,
  "datagrams": true,
  "certificate_hash": "sha256:abcd1234...",
  "alpn": ["wt"]
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "session_id": "wt_session_123",
    "status": "connecting",
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

### Get WebTransport Session Status

**Endpoint:** `GET /api/webtransport/sessions/{id}`

**Response:**
```json
{
  "success": true,
  "data": {
    "session_id": "wt_session_123",
    "status": "connected",
    "created_at": "2024-01-01T12:00:00Z",
    "connected_at": "2024-01-01T12:00:02Z",
    "metrics": {
      "streams_opened": 4,
      "streams_closed": 0,
      "datagrams_sent": 1000,
      "datagrams_received": 995,
      "bytes_sent": 1048576,
      "bytes_received": 1045000
    }
  }
}
```

## HTTP/3 Load Testing API

### Create HTTP/3 Load Test

Create a new HTTP/3 load testing session.

**Endpoint:** `POST /api/http3/load-tests`

**Request Body:**
```json
{
  "target_url": "https://example.com:443",
  "duration": "300s",
  "concurrent_connections": 10,
  "requests_per_connection": 100,
  "request_pattern": "sequential",
  "headers": {
    "User-Agent": "QUIC-Test-Suite/1.0"
  },
  "body_size": 1024,
  "think_time": "100ms"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "load_test_id": "http3_load_123",
    "status": "starting",
    "created_at": "2024-01-01T12:00:00Z",
    "estimated_completion": "2024-01-01T12:05:00Z"
  }
}
```

### Get HTTP/3 Load Test Results

**Endpoint:** `GET /api/http3/load-tests/{id}`

**Response:**
```json
{
  "success": true,
  "data": {
    "load_test_id": "http3_load_123",
    "status": "completed",
    "created_at": "2024-01-01T12:00:00Z",
    "completed_at": "2024-01-01T12:05:00Z",
    "results": {
      "total_requests": 1000,
      "successful_requests": 995,
      "failed_requests": 5,
      "avg_response_time_ms": 125.5,
      "p95_response_time_ms": 250.0,
      "p99_response_time_ms": 400.0,
      "requests_per_second": 66.7,
      "bytes_transferred": 10485760,
      "error_rate": 0.005,
      "status_codes": {
        "200": 995,
        "500": 3,
        "timeout": 2
      }
    }
  }
}
```

## Examples

### Start a Basic Test

```bash
curl -X POST http://localhost:8081/api/tests \
  -H "Content-Type: application/json" \
  -d '{
    "mode": "test",
    "duration": "60s",
    "connections": 2,
    "streams": 4,
    "congestion_control": "bbrv3"
  }'
```

### Monitor Test Progress

```bash
# Get test status
curl http://localhost:8081/api/tests/test_1704110400

# Get real-time metrics
curl http://localhost:8081/api/metrics/current
```

### Export Results

```bash
# Export as JSON
curl "http://localhost:8081/api/tests/test_1704110400/export?format=json" > results.json

# Export as CSV
curl "http://localhost:8081/api/tests/test_1704110400/export?format=csv" > results.csv

# Get Prometheus metrics
curl http://localhost:8081/api/metrics/prometheus
```

### JavaScript Example

```javascript
// Start a test
async function startTest() {
  const response = await fetch('/api/tests', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      mode: 'test',
      duration: '60s',
      connections: 2,
      streams: 4,
      congestion_control: 'bbrv3',
      fec_enabled: true,
      fec_redundancy: 0.10
    })
  });
  
  const result = await response.json();
  if (result.success) {
    return result.data.id;
  } else {
    throw new Error(result.error);
  }
}

// Monitor test with WebSocket
function monitorTest(testId) {
  const ws = new WebSocket(`ws://localhost:8081/api/ws/metrics?test_id=${testId}`);
  
  ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    if (data.type === 'metrics_update') {
      console.log('Metrics update:', data.data);
      updateUI(data.data);
    }
  };
  
  ws.onerror = (error) => {
    console.error('WebSocket error:', error);
  };
  
  return ws;
}

// Get historical data
async function getHistoricalMetrics(testId, startTime, endTime) {
  const params = new URLSearchParams({
    test_id: testId,
    start_time: startTime,
    end_time: endTime,
    interval: '5s'
  });
  
  const response = await fetch(`/api/metrics/history?${params}`);
  const result = await response.json();
  
  if (result.success) {
    return result.data.metrics;
  } else {
    throw new Error(result.error);
  }
}
```

### Python Example

```python
import requests
import json
import time

class QUICTestClient:
    def __init__(self, base_url="http://localhost:8081/api"):
        self.base_url = base_url
        self.session = requests.Session()
    
    def start_test(self, config):
        """Start a new test"""
        response = self.session.post(
            f"{self.base_url}/tests",
            json=config
        )
        response.raise_for_status()
        return response.json()["data"]
    
    def get_test_status(self, test_id):
        """Get test status and metrics"""
        response = self.session.get(f"{self.base_url}/tests/{test_id}")
        response.raise_for_status()
        return response.json()["data"]
    
    def wait_for_completion(self, test_id, timeout=300):
        """Wait for test to complete"""
        start_time = time.time()
        while time.time() - start_time < timeout:
            status = self.get_test_status(test_id)
            if status["status"] in ["completed", "failed", "stopped"]:
                return status
            time.sleep(5)
        raise TimeoutError(f"Test {test_id} did not complete within {timeout}s")
    
    def export_results(self, test_id, format="json"):
        """Export test results"""
        response = self.session.get(
            f"{self.base_url}/tests/{test_id}/export",
            params={"format": format}
        )
        response.raise_for_status()
        return response.text if format != "json" else response.json()

# Usage example
client = QUICTestClient()

# Start test
test_config = {
    "mode": "test",
    "duration": "60s",
    "connections": 2,
    "streams": 4,
    "congestion_control": "bbrv3",
    "fec_enabled": True,
    "fec_redundancy": 0.10
}

test = client.start_test(test_config)
print(f"Started test: {test['id']}")

# Wait for completion
final_status = client.wait_for_completion(test["id"])
print(f"Test completed with status: {final_status['status']}")

# Export results
results = client.export_results(test["id"], "json")
print(f"Final metrics: {results['data']['metrics']}")
```

## SDKs and Libraries

### Official SDKs

- **JavaScript/TypeScript**: `@quic-test/client`
- **Python**: `quic-test-client`
- **Go**: `github.com/twogc/quic-test/client`

### Community Libraries

- **Java**: `quic-test-java-client`
- **C#**: `QuicTest.Client`
- **Rust**: `quic-test-rs`

### Installation

```bash
# JavaScript/Node.js
npm install @quic-test/client

# Python
pip install quic-test-client

# Go
go get github.com/twogc/quic-test/client
```

## Rate Limiting

The API implements rate limiting to prevent abuse:

- **Default limit**: 100 requests per minute per IP
- **Test creation**: 10 tests per minute per IP
- **Metrics queries**: 1000 requests per minute per IP

Rate limit headers are included in responses:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1704110460
```

## Versioning

The API uses semantic versioning. Current version: `v1`

- **Major version**: Breaking changes
- **Minor version**: New features, backward compatible
- **Patch version**: Bug fixes, backward compatible

Version is included in response headers:
```
X-API-Version: 1.0.0
```

## Support

For API support and questions:

- **Documentation**: [https://github.com/twogc/quic-test/docs](https://github.com/twogc/quic-test/docs)
- **Issues**: [https://github.com/twogc/quic-test/issues](https://github.com/twogc/quic-test/issues)
- **Discussions**: [https://github.com/twogc/quic-test/discussions](https://github.com/twogc/quic-test/discussions)