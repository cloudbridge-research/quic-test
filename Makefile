# QUIC Experimental Test Suite Makefile
# ====================================

.PHONY: help build test clean bench-rtt bench-loss bench-pps soak-2h

# Default target
help:
	@echo "QUIC Experimental Test Suite"
	@echo "=========================="
	@echo ""
	@echo "Available targets:"
	@echo "  build        - Build the QUIC test binary"
	@echo "  test         - Run basic functionality tests"
	@echo "  clean        - Clean build artifacts and test results"
	@echo "  bench-rtt    - Run RTT sensitivity benchmarks"
	@echo "  bench-loss   - Run loss rate benchmarks"
	@echo "  bench-pps    - Run packet rate benchmarks"
	@echo "  soak-2h      - Run 2-hour soak test"
	@echo "  regression   - Run full regression test suite"
	@echo "  real-world   - Run real-world scenario tests"
	@echo ""

# Build all binaries
build:
	@echo "Building QUIC test suite..."
	go build -o quic-test ./
	go build -o quic-gui ./cmd/gui/
	@echo "Build completed"

# Build the experimental QUIC test binary
build-experimental:
	@echo "Building QUIC test experimental binary..."
	go build -o quic-test-experimental ./cmd/experimental/
	@echo "Experimental build completed"

# Run basic functionality tests
test: build
	@echo "Running basic functionality tests..."
	@mkdir -p test-results
	@./scripts/regression_test_script.sh --duration 30 --cleanup
	@echo "Basic tests completed"

# Clean build artifacts and test results
clean:
	@echo "Cleaning build artifacts and test results..."
	rm -f quic-test-experimental quic-test quic-gui
	rm -rf test-results/
	rm -rf regression-results/
	rm -rf performance-results/
	rm -rf real-world-results/
	@echo "Cleanup completed"

# Run RTT sensitivity benchmarks
bench-rtt: build
	@echo "Running RTT sensitivity benchmarks..."
	@mkdir -p test-results/bench-rtt
	@./scripts/rtt_test_script.sh \
		--rtt 5,10,25,50,100,200 \
		--algorithms cubic,bbrv2 \
		--duration 60 \
		--output test-results/bench-rtt \
		--cleanup
	@echo "RTT benchmarks completed"

# Run loss rate benchmarks
bench-loss: build
	@echo "Running loss rate benchmarks..."
	@mkdir -p test-results/bench-loss
	@./scripts/real_world_test_script.sh \
		--duration 120 \
		--output test-results/bench-loss \
		--cleanup
	@echo "Loss rate benchmarks completed"

# Run packet rate benchmarks
bench-pps: build
	@echo "Running packet rate benchmarks..."
	@mkdir -p test-results/bench-pps
	@./scripts/load_test_script.sh \
		--load 100,300,600,1000,2000 \
		--connections 1,2,4,8 \
		--algorithms cubic,bbrv2 \
		--duration 120 \
		--output test-results/bench-pps \
		--cleanup
	@echo "Packet rate benchmarks completed"

# Run 2-hour soak test
soak-2h: build
	@echo "Running 2-hour soak test..."
	@mkdir -p test-results/soak-2h
	@echo "Starting long-term stability test..."
	@nohup ./quic-test-experimental \
		--mode server \
		--cc bbrv2 \
		--qlog test-results/soak-2h/server-qlog \
		--verbose \
		--metrics-interval 10s \
		> test-results/soak-2h/server.log 2>&1 &
	@SERVER_PID=$$!; \
	sleep 5; \
	timeout 7200s ./quic-test-experimental \
		--mode client \
		--addr 127.0.0.1:9000 \
		--cc bbrv2 \
		--qlog test-results/soak-2h/client-qlog \
		--duration 7200s \
		--connections 4 \
		--streams 2 \
		--rate 500 \
		--packet-size 1200 \
		--verbose \
		> test-results/soak-2h/client.log 2>&1; \
	kill $$SERVER_PID 2>/dev/null || true; \
	wait $$SERVER_PID 2>/dev/null || true
	@echo "Soak test completed"

# Run full regression test suite
regression: build
	@echo "Running full regression test suite..."
	@./scripts/run_regression_tests.sh --full --cleanup
	@echo "Regression tests completed"

# Run real-world scenario tests
real-world: build
	@echo "Running real-world scenario tests..."
	@./scripts/real_world_test_script.sh --duration 120 --cleanup
	@echo "Real-world tests completed"

# Run all performance tests
performance: build
	@echo "Running all performance tests..."
	@./scripts/run_performance_tests.sh --full --cleanup
	@echo "Performance tests completed"

# Generate reports
reports:
	@echo "Generating test reports..."
	@./scripts/run_regression_tests.sh --analysis-only
	@./scripts/run_performance_tests.sh --analysis-only
	@echo "Reports generated"

# Install system dependencies
deps:
	@echo "Installing system dependencies..."
	sudo apt-get update
	sudo apt-get install -y iproute2 jq bc
	@echo "Dependencies installed"

# Configure system for optimal performance
config:
	@echo "Configuring system for optimal performance..."
	@echo 'net.core.rmem_max = 134217728' | sudo tee -a /etc/sysctl.conf
	@echo 'net.core.wmem_max = 134217728' | sudo tee -a /etc/sysctl.conf
	@echo 'net.core.netdev_max_backlog = 5000' | sudo tee -a /etc/sysctl.conf
	@sudo sysctl -p
	@echo "System configured"

# Run quick smoke test
smoke: build
	@echo "Running quick smoke test..."
	@mkdir -p test-results/smoke
	@nohup ./quic-test-experimental \
		--mode server \
		--cc bbrv2 \
		--verbose \
		> test-results/smoke/server.log 2>&1 &
	@SERVER_PID=$$!; \
	sleep 2; \
	timeout 10s ./quic-test-experimental \
		--mode client \
		--addr 127.0.0.1:9000 \
		--cc bbrv2 \
		--duration 10s \
		--connections 1 \
		--rate 100 \
		--verbose \
		> test-results/smoke/client.log 2>&1; \
	kill $$SERVER_PID 2>/dev/null || true; \
	wait $$SERVER_PID 2>/dev/null || true
	@echo "Smoke test completed"

# Run comprehensive test suite
all: clean build test bench-rtt bench-loss bench-pps regression real-world performance reports
	@echo "All tests completed successfully!"

# Show test status
status:
	@echo "Test Status"
	@echo "=============="
	@if [ -f "quic-test-experimental" ]; then echo "Binary: Built"; else echo "Binary: Not built"; fi
	@if [ -f "quic-test" ]; then echo "Main binary: Built"; else echo "Main binary: Not built"; fi
	@if [ -f "quic-gui" ]; then echo "GUI binary: Built"; else echo "GUI binary: Not built"; fi
	@if [ -d "test-results" ]; then echo "Test results: Available"; else echo "Test results: Not available"; fi
	@if [ -d "regression-results" ]; then echo "Regression results: Available"; else echo "Regression results: Not available"; fi
	@if [ -d "performance-results" ]; then echo "Performance results: Available"; else echo "Performance results: Not available"; fi

# Start GUI server
gui: build
	@echo "Starting QUIC Test GUI..."
	./quic-gui --addr=:8080 --api-addr=:8081

# Run tests with GUI
test-gui: build
	@echo "Starting test with GUI monitoring..."
	./quic-gui --addr=:8080 --api-addr=:8081 &
	@GUI_PID=$$!; \
	sleep 3; \
	echo "GUI started at http://localhost:8080"; \
	echo "Press Ctrl+C to stop"; \
	wait $$GUI_PID

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t quic-test:latest .
	@echo "Docker image built"

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 -p 8081:8081 -p 9000:9000/udp quic-test:latest

# Test WebTransport functionality
test-webtransport: build
	@echo "Testing WebTransport functionality..."
	@mkdir -p test-results/webtransport
	./quic-test --mode=webtransport-test --duration=30s --output=test-results/webtransport/
	@echo "WebTransport test completed"

# Test HTTP/3 load testing
test-http3: build
	@echo "Testing HTTP/3 load testing..."
	@mkdir -p test-results/http3
	./quic-test --mode=http3-load-test --target=https://example.com --duration=60s --output=test-results/http3/
	@echo "HTTP/3 load test completed"