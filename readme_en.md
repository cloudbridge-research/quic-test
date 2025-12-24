# quic-test

Professional QUIC protocol testing platform for network engineers, researchers, and educators.

[![CI](https://github.com/cloudbridge-research/quic-test/actions/workflows/pipeline.yml/badge.svg)](https://github.com/cloudbridge-research/quic-test/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/cloudbridge-research/quic-test)](https://goreportcard.com/report/github.com/cloudbridge-research/quic-test)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/docker/v/mlanies/quic-test?label=docker)](https://hub.docker.com/r/mlanies/quic-test)

**English** | [Русский](readme.md)

## What is this?

`quic-test` is a professional platform for testing and analyzing QUIC protocol performance. Designed for educational and research purposes, with focus on reproducibility and detailed analytics.

**Key Features:**
- Web GUI interface for less technical users
- Measure latency, jitter, throughput for QUIC and TCP
- Emulate various network conditions (loss, delay, bandwidth)
- Real-time TUI visualization
- Prometheus metrics export
- WebTransport and HTTP/3 load testing
- Forward Error Correction (FEC) with SIMD optimization
- Post-Quantum Cryptography simulation
- BBRv3 congestion control with dual-scale bandwidth estimation

## Quick Start

### GUI Interface (Recommended for beginners)

```bash
# Build GUI
make build

# Start GUI server
make gui
# or
./quic-gui --addr=:8080 --api-addr=:8081

# Open browser
open http://localhost:8080
```

**GUI Features:**
- Create tests through web forms
- Real-time monitoring of active tests
- Test history with detailed metrics
- Stop tests with one click
- Ready-made presets for various scenarios

### Docker (recommended for production)

```bash
# Run with GUI
docker run -p 8080:8080 -p 8081:8081 -p 9000:9000/udp mlanies/quic-test:latest gui

# Run client test
docker run mlanies/quic-test:latest --mode=client --server=demo.quic.tech:4433

# Run server
docker run -p 4433:4433/udp mlanies/quic-test:latest --mode=server
```

### Command Line Interface

```bash
# Build from source
git clone https://github.com/cloudbridge-research/quic-test
cd quic-test

# Build FEC library (optional, for better performance)
cd internal/fec && make && cd ../..

# Build all components
make build

# Run basic test
./quic-test --mode=client --server=demo.quic.tech:4433
```

## Basic Usage

```bash
# Simple latency/throughput test
./quic-test --mode=client --server=localhost:4433 --duration=30s

# Compare QUIC vs TCP
./quic-test --mode=client --compare-tcp --duration=60s

# Emulate mobile network
./quic-test --profile=mobile --duration=30s

# TUI monitoring
./cmd/tui/tui --server=localhost:4433

# WebTransport testing
make test-webtransport

# HTTP/3 load testing
make test-http3
```

## Architecture

```
quic-test/
├── cmd/
│   ├── gui/                # Web GUI interface
│   ├── tui/                # Terminal UI monitoring
│   ├── experimental/       # Experimental features
│   ├── quic-client/        # QUIC client
│   ├── quic-server/        # QUIC server
│   ├── dashboard/          # Dashboard
│   ├── masque/             # MASQUE VPN tests
│   ├── ice/                # ICE/STUN/TURN tests
│   └── security-test/      # Security tests
├── client/                 # QUIC client (legacy)
├── server/                 # QUIC server (legacy)
├── internal/
│   ├── gui/                # GUI server and API
│   ├── quic/               # QUIC logic
│   ├── fec/                # Forward Error Correction (C++/AVX2)
│   ├── congestion/         # BBRv2/BBRv3 algorithms
│   ├── webtransport/       # WebTransport support
│   ├── http3/              # HTTP/3 load testing
│   ├── pqc/                # Post-Quantum Crypto simulation
│   ├── metrics/            # Prometheus metrics
│   └── ai/                 # AI integration
├── web/                    # Web GUI static files
│   ├── static/css/         # CSS styles
│   └── static/js/          # JavaScript
└── docs/                   # Documentation
```

**Details:** [docs/architecture.md](docs/architecture.md)

## Features

### Stable Features

- **Web GUI interface** — user-friendly web interface for creating and monitoring tests
- **QUIC client/server** — based on quic-go with extensions
- **RTT, jitter, throughput measurements** — detailed performance analytics
- **Network profile emulation** — mobile, satellite, fiber, WiFi
- **TUI visualization** — real-time terminal monitoring
- **Prometheus export** — integration with monitoring systems
- **BBRv2 congestion control** — modern congestion control algorithm

### Experimental Features

- **BBRv3 congestion control** — with dual-scale bandwidth estimation and 2% loss threshold
- **Forward Error Correction (FEC)** — with AVX2/SIMD optimization
- **WebTransport support** — WebTransport connection testing
- **HTTP/3 load testing** — HTTP/3 load testing
- **Post-Quantum Cryptography** — PQC algorithm simulation (ML-KEM, Dilithium)
- **MASQUE VPN testing** — VPN over QUIC tests
- **ICE/STUN/TURN tests** — NAT traversal testing

### Planned Features (Roadmap)

- Automatic anomaly detection
- Multi-cloud deployment
- Extended AI integration
- QUIC v2 support

**Full roadmap:** [docs/roadmap.md](docs/roadmap.md)

## Documentation

- **[MEI Collaboration Report](docs/MEI_COLLABORATION_REPORT.md)** — project metrics and internship program
- **[Student Guide](docs/STUDENT_GUIDE_EN.md)** — terminology, TCP vs QUIC, RFC documents
- **[API Reference](docs/API_REFERENCE.md)** — complete REST API reference
- **[CLI Reference](docs/cli.md)** — command reference
- **[Architecture](docs/architecture.md)** — detailed architecture
- **[Education](docs/education.md)** — lab materials for universities
- **[AI Integration](docs/ai-routing-integration.md)** — AI Routing Lab integration
- **[Case Studies](docs/case-studies.md)** — test results with methodology
- **[TUI User Guide](docs/TUI_USER_GUIDE.md)** — TUI interface guide

## GUI Interface

Web GUI provides a user-friendly interface for users without deep technical knowledge:

### Main GUI Features:
- **Dashboard** — overview of active tests and system status
- **New Test** — create tests through web forms with validation
- **Test History** — view all executed tests
- **Test Details** — detailed view of test metrics and logs
- **Real-time Updates** — automatic test status updates

### API Endpoints:
- `POST /api/tests` — create new test
- `GET /api/tests` — get test list
- `GET /api/tests/{id}` — get test details
- `DELETE /api/tests/{id}` — stop test
- `GET /api/metrics/current` — current aggregated metrics
- `GET /api/metrics/prometheus` — metrics in Prometheus format

**Details:** [docs/API_REFERENCE.md](docs/API_REFERENCE.md)

## For Universities

Designed with education and career preparation in mind. Includes ready-to-use lab materials, educational resources, and internship program.

### Educational Resources:
- **[Student Guide](docs/STUDENT_GUIDE_EN.md)** — terminology, TCP vs QUIC comparison, RFC documents
- **Practical lab assignments** with step-by-step instructions
- **Ready-made test scenarios** for various network conditions

### Lab Assignments:
- **Lab #1:** QUIC Basics — handshake, 0-RTT, connection migration
- **Lab #2:** Congestion Control — BBRv2 vs BBRv3 comparison
- **Lab #3:** Performance — QUIC vs TCP under various conditions
- **Lab #4:** Forward Error Correction — FEC impact on performance
- **Lab #5:** Post-Quantum Cryptography — PQC algorithm testing

### CloudBridge Research Internship Program

**Available opportunities for MEI students:**

**Career Tracks:**
- Junior Network Engineer (80,000 - 120,000 RUB/month)
- Protocol Research Developer (120,000 - 180,000 RUB/month)
- DevOps/Infrastructure Engineer (100,000 - 160,000 RUB/month)
- AI/ML Engineer (140,000 - 200,000 RUB/month)

**Internship Conditions:**
- Summer internship: 40,000 RUB/month (3 months)
- Thesis practice: 50,000 RUB/month (6 months)
- Hybrid work format (office + remote)
- Employment opportunity after successful completion

**Details:** [docs/education.md](docs/education.md) | [Collaboration Report](docs/MEI_COLLABORATION_REPORT.md)

## AI Routing Lab Integration

`quic-test` exports metrics to Prometheus, which are used in [AI Routing Lab](https://github.com/cloudbridge-research/ai-routing-lab) for training route prediction models.

**Example:**
```bash
# Run with Prometheus export
./quic-test --mode=server --prometheus-port=9090

# AI Routing Lab collects metrics
curl http://localhost:9090/metrics
```

**Details:** [docs/ai-routing-integration.md](docs/ai-routing-integration.md)

## Development

```bash
# Run tests
make test

# Full test suite
make all

# Smoke test
make smoke

# Build Docker image
make docker-build

# Run in Docker
make docker-run

# Linting
golangci-lint run

# Build status
make status
```

### Available Make Commands:
- `make build` — build all binaries
- `make gui` — start GUI server
- `make test` — basic functional tests
- `make bench-rtt` — RTT benchmarks
- `make bench-loss` — packet loss benchmarks
- `make soak-2h` — 2-hour stress test
- `make regression` — full regression test suite
- `make performance` — performance tests

## License

MIT License. See [LICENSE](LICENSE).

## Contacts

- **GitHub:** [cloudbridge-research/quic-test](https://github.com/cloudbridge-research/quic-test)
- **Blog:** [cloudbridge-research.ru](https://cloudbridge-research.ru)
- **Email:** info@cloudbridge-research.ru
- **Docker Hub:** [mlanies/quic-test](https://hub.docker.com/r/mlanies/quic-test)