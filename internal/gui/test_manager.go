package gui

import (
	"context"
	"fmt"
	"time"

	"quic-test/internal"
)

// StartTest starts a new test session
func (tm *TestManager) StartTest(config internal.TestConfig) *TestSession {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	// Generate unique test ID
	testID := fmt.Sprintf("test_%d", time.Now().Unix())
	
	session := &TestSession{
		ID:        testID,
		Config:    config,
		Status:    "running",
		StartTime: time.Now(),
		Metrics:   make(map[string]interface{}),
		Logs:      make([]string, 0),
	}
	
	tm.activeTests[testID] = session
	
	// Start test in background
	go tm.runTest(session)
	
	return session
}

// StopTest stops a running test
func (tm *TestManager) StopTest(testID string) error {
	tm.mu.RLock()
	session, exists := tm.activeTests[testID]
	tm.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("test not found: %s", testID)
	}
	
	session.mu.Lock()
	defer session.mu.Unlock()
	
	if session.Status != "running" {
		return fmt.Errorf("test is not running: %s", testID)
	}
	
	session.Status = "stopped"
	now := time.Now()
	session.EndTime = &now
	session.addLog("Test stopped by user")
	
	return nil
}

// GetTest retrieves a test session by ID
func (tm *TestManager) GetTest(testID string) *TestSession {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	return tm.activeTests[testID]
}

// GetAllTests returns all test sessions
func (tm *TestManager) GetAllTests() []*TestSession {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	tests := make([]*TestSession, 0, len(tm.activeTests))
	for _, session := range tm.activeTests {
		tests = append(tests, session)
	}
	
	return tests
}

// GetActiveTestCount returns the number of running tests
func (tm *TestManager) GetActiveTestCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	count := 0
	for _, session := range tm.activeTests {
		session.mu.RLock()
		if session.Status == "running" {
			count++
		}
		session.mu.RUnlock()
	}
	
	return count
}

// GetTotalTestCount returns the total number of tests
func (tm *TestManager) GetTotalTestCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	return len(tm.activeTests)
}

// runTest executes a test session
func (tm *TestManager) runTest(session *TestSession) {
	defer func() {
		if r := recover(); r != nil {
			session.mu.Lock()
			session.Status = "failed"
			now := time.Now()
			session.EndTime = &now
			session.addLog(fmt.Sprintf("Test failed with panic: %v", r))
			session.mu.Unlock()
		}
	}()
	
	session.addLogSafe("Starting test execution")
	
	// Create context with timeout
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Monitor for stop requests
	go func() {
		for {
			session.mu.RLock()
			status := session.Status
			session.mu.RUnlock()
			
			if status == "stopped" {
				cancel()
				return
			}
			
			time.Sleep(100 * time.Millisecond)
		}
	}()
	
	// Run the actual test based on mode
	switch session.Config.Mode {
	case "server":
		tm.runServerTest(ctx, session)
	case "client":
		tm.runClientTest(ctx, session)
	case "test":
		tm.runIntegratedTest(ctx, session)
	default:
		session.mu.Lock()
		session.Status = "failed"
		now := time.Now()
		session.EndTime = &now
		session.addLog(fmt.Sprintf("Unknown test mode: %s", session.Config.Mode))
		session.mu.Unlock()
		return
	}
	
	// Mark test as completed if not already stopped/failed
	session.mu.Lock()
	if session.Status == "running" {
		session.Status = "completed"
		now := time.Now()
		session.EndTime = &now
		session.addLog("Test completed successfully")
	}
	session.mu.Unlock()
}

// runServerTest runs a server-only test
func (tm *TestManager) runServerTest(ctx context.Context, session *TestSession) {
	session.addLogSafe("Starting QUIC server")
	
	// This would integrate with the actual server implementation
	// For now, simulate server operation
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			session.addLogSafe("Server test stopped")
			return
		case <-ticker.C:
			// Update metrics (simulated)
			session.updateMetrics(map[string]interface{}{
				"connections": 0,
				"bytes_received": 0,
				"uptime": time.Since(session.StartTime).Seconds(),
			})
		}
	}
}

// runClientTest runs a client-only test
func (tm *TestManager) runClientTest(ctx context.Context, session *TestSession) {
	session.addLogSafe("Starting QUIC client test")
	
	// This would integrate with the actual client implementation
	// For now, simulate client operation
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	startTime := time.Now()
	
	for {
		select {
		case <-ctx.Done():
			session.addLogSafe("Client test stopped")
			return
		case <-ticker.C:
			elapsed := time.Since(startTime)
			
			// Check if duration limit reached
			if session.Config.Duration > 0 && elapsed >= session.Config.Duration {
				session.addLogSafe("Test duration reached")
				return
			}
			
			// Update metrics (simulated)
			session.updateMetrics(map[string]interface{}{
				"latency_ms": 50.0 + (10.0 * (0.5 - float64(time.Now().UnixNano()%1000)/1000.0)),
				"throughput_mbps": 100.0 + (20.0 * (0.5 - float64(time.Now().UnixNano()%1000)/1000.0)),
				"packet_loss": 0.01,
				"connections": session.Config.Connections,
				"elapsed_seconds": elapsed.Seconds(),
			})
		}
	}
}

// runIntegratedTest runs both server and client
func (tm *TestManager) runIntegratedTest(ctx context.Context, session *TestSession) {
	session.addLogSafe("Starting integrated test (server + client)")
	
	// Start server in background
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		tm.runServerTest(ctx, session)
	}()
	
	// Wait a bit for server to start
	time.Sleep(2 * time.Second)
	session.addLogSafe("Server started, beginning client test")
	
	// Run client test
	tm.runClientTest(ctx, session)
	
	// Wait for server to finish
	<-serverDone
	session.addLogSafe("Integrated test completed")
}

// Helper methods for TestSession
func (ts *TestSession) addLog(message string) {
	// Note: This method assumes the caller already holds the mutex
	// If called without mutex, it should be called as addLogSafe
	timestamp := time.Now().Format("15:04:05")
	logEntry := fmt.Sprintf("[%s] %s", timestamp, message)
	ts.Logs = append(ts.Logs, logEntry)
	
	// Keep only last 100 log entries
	if len(ts.Logs) > 100 {
		ts.Logs = ts.Logs[len(ts.Logs)-100:]
	}
}

// addLogSafe adds a log entry with mutex protection
func (ts *TestSession) addLogSafe(message string) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	ts.addLog(message)
}

func (ts *TestSession) updateMetrics(metrics map[string]interface{}) {
	ts.mu.Lock()
	defer ts.mu.Unlock()
	
	for key, value := range metrics {
		ts.Metrics[key] = value
	}
}

// GetMetrics returns a copy of current metrics
func (ts *TestSession) GetMetrics() map[string]interface{} {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	
	metrics := make(map[string]interface{})
	for key, value := range ts.Metrics {
		metrics[key] = value
	}
	
	return metrics
}

// GetLogs returns a copy of current logs
func (ts *TestSession) GetLogs() []string {
	ts.mu.RLock()
	defer ts.mu.RUnlock()
	
	logs := make([]string, len(ts.Logs))
	copy(logs, ts.Logs)
	
	return logs
}