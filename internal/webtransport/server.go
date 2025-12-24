package webtransport

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/quic-go/quic-go/http3"
)

// Server represents a WebTransport server
type Server struct {
	config   *ServerConfig
	server   *http3.Server
	sessions map[string]*ServerSession
	metrics  *ServerMetrics
	mu       sync.RWMutex
}

// ServerConfig holds WebTransport server configuration
type ServerConfig struct {
	Addr      string      `json:"addr"`
	TLSConfig *tls.Config `json:"-"`
	CertFile  string      `json:"cert_file,omitempty"`
	KeyFile   string      `json:"key_file,omitempty"`
}

// ServerSession represents a server-side WebTransport session
type ServerSession struct {
	ID          string                 `json:"session_id"`
	ClientAddr  string                 `json:"client_addr"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	LastActive  time.Time              `json:"last_active"`
	Streams     map[string]*StreamInfo `json:"streams"`
	Metrics     map[string]interface{} `json:"metrics"`
	mu          sync.RWMutex
}

// ServerMetrics holds server-side WebTransport metrics
type ServerMetrics struct {
	ActiveSessions    int64   `json:"active_sessions"`
	TotalSessions     int64   `json:"total_sessions"`
	TotalStreams      int64   `json:"total_streams"`
	TotalDatagrams    int64   `json:"total_datagrams"`
	BytesReceived     int64   `json:"bytes_received"`
	BytesSent         int64   `json:"bytes_sent"`
	AvgSessionTime    float64 `json:"avg_session_time_ms"`
	ErrorCount        int64   `json:"error_count"`
	LastError         string  `json:"last_error,omitempty"`
	
	mu sync.RWMutex
}

// NewServer creates a new WebTransport server
func NewServer(config *ServerConfig) *Server {
	return &Server{
		config:   config,
		sessions: make(map[string]*ServerSession),
		metrics:  &ServerMetrics{},
	}
}

// Start starts the WebTransport server
func (s *Server) Start(ctx context.Context) error {
	// Configure TLS
	tlsConfig := s.config.TLSConfig
	if tlsConfig == nil {
		if s.config.CertFile == "" || s.config.KeyFile == "" {
			// Generate self-signed certificate for testing
			tlsConfig = s.generateSelfSignedTLS()
		} else {
			cert, err := tls.LoadX509KeyPair(s.config.CertFile, s.config.KeyFile)
			if err != nil {
				return fmt.Errorf("failed to load TLS certificate: %w", err)
			}
			
			tlsConfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
				NextProtos:   []string{"wt", "h3"},
			}
		}
	}
	
	// Create HTTP/3 server
	mux := http.NewServeMux()
	mux.HandleFunc("/webtransport", s.handleWebTransport)
	mux.HandleFunc("/health", s.handleHealth)
	
	s.server = &http3.Server{
		Addr:      s.config.Addr,
		Handler:   mux,
		TLSConfig: tlsConfig,
	}
	
	fmt.Printf("Starting WebTransport server on %s\n", s.config.Addr)
	
	// Start server in background
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.metrics.mu.Lock()
			s.metrics.ErrorCount++
			s.metrics.LastError = fmt.Sprintf("Server error: %v", err)
			s.metrics.mu.Unlock()
		}
	}()
	
	// Wait for context cancellation
	<-ctx.Done()
	
	// Graceful shutdown
	return s.Stop()
}

// Stop stops the WebTransport server
func (s *Server) Stop() error {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		return s.server.Close()
	}
	return nil
}

// handleWebTransport handles WebTransport connection requests
func (s *Server) handleWebTransport(w http.ResponseWriter, r *http.Request) {
	// Check for WebTransport upgrade
	if r.Header.Get("Connection") != "Upgrade" || 
	   r.Header.Get("Upgrade") != "webtransport" {
		http.Error(w, "Not a WebTransport request", http.StatusBadRequest)
		return
	}
	
	// Create new session
	sessionID := fmt.Sprintf("server_session_%d", time.Now().UnixNano())
	
	session := &ServerSession{
		ID:         sessionID,
		ClientAddr: r.RemoteAddr,
		Status:     "connected",
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
		Streams:    make(map[string]*StreamInfo),
		Metrics:    make(map[string]interface{}),
	}
	
	s.mu.Lock()
	s.sessions[sessionID] = session
	s.metrics.ActiveSessions++
	s.metrics.TotalSessions++
	s.mu.Unlock()
	
	// Accept WebTransport connection
	w.Header().Set("Sec-WebTransport-Http3-Draft", "draft02")
	w.WriteHeader(http.StatusOK)
	
	// Handle session
	s.handleSession(r.Context(), session)
}

// handleSession handles a WebTransport session
func (s *Server) handleSession(ctx context.Context, session *ServerSession) {
	defer func() {
		// Clean up session
		s.mu.Lock()
		delete(s.sessions, session.ID)
		s.metrics.ActiveSessions--
		s.mu.Unlock()
		
		session.mu.Lock()
		session.Status = "closed"
		session.mu.Unlock()
	}()
	
	// Simulate session handling
	// In a real implementation, this would handle actual WebTransport streams and datagrams
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			session.mu.Lock()
			session.LastActive = time.Now()
			
			// Simulate receiving data
			session.Metrics["bytes_received"] = s.metrics.BytesReceived
			session.Metrics["streams_count"] = len(session.Streams)
			session.mu.Unlock()
			
			// Update server metrics
			s.metrics.mu.Lock()
			s.metrics.BytesReceived += 1024 // Simulate 1KB received per second
			s.metrics.BytesSent += 1024     // Simulate 1KB sent per second
			s.metrics.mu.Unlock()
		}
	}
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := map[string]interface{}{
		"status":          "healthy",
		"active_sessions": s.metrics.ActiveSessions,
		"total_sessions":  s.metrics.TotalSessions,
		"uptime":          time.Since(time.Now()).String(),
	}
	
	fmt.Fprintf(w, `{"status":"healthy","active_sessions":%d,"total_sessions":%d}`,
		s.metrics.ActiveSessions, s.metrics.TotalSessions)
}

// GetSessions returns all active sessions
func (s *Server) GetSessions() map[string]*ServerSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Return a copy
	sessions := make(map[string]*ServerSession)
	for id, session := range s.sessions {
		sessions[id] = session
	}
	
	return sessions
}

// GetSession returns a specific session
func (s *Server) GetSession(sessionID string) *ServerSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	return s.sessions[sessionID]
}

// GetMetrics returns server metrics
func (s *Server) GetMetrics() *ServerMetrics {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()
	
	// Return a copy
	metrics := *s.metrics
	return &metrics
}

// generateSelfSignedTLS generates a self-signed TLS certificate for testing
func (s *Server) generateSelfSignedTLS() *tls.Config {
	// This is a simplified implementation
	// In production, use proper certificate generation
	return &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"wt", "h3"},
	}
}