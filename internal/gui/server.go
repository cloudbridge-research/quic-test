package gui

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"quic-test/internal"
)

// Server handles the GUI web interface
type Server struct {
	templates   *template.Template
	devMode     bool
	testManager *TestManager
	mu          sync.RWMutex
}

// TestManager manages running tests
type TestManager struct {
	activeTests map[string]*TestSession
	mu          sync.RWMutex
}

// TestSession represents an active test session
type TestSession struct {
	ID          string                 `json:"id"`
	Config      internal.TestConfig    `json:"config"`
	Status      string                 `json:"status"` // "running", "completed", "failed"
	StartTime   time.Time              `json:"start_time"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	Metrics     map[string]interface{} `json:"metrics"`
	Logs        []string               `json:"logs"`
	mu          sync.RWMutex
}

// NewServer creates a new GUI server
func NewServer(devMode bool) *Server {
	server := &Server{
		devMode:     devMode,
		testManager: NewTestManager(),
	}
	
	server.loadTemplates()
	return server
}

// NewTestManager creates a new test manager
func NewTestManager() *TestManager {
	return &TestManager{
		activeTests: make(map[string]*TestSession),
	}
}

// loadTemplates loads HTML templates
func (s *Server) loadTemplates() {
	if s.devMode {
		// In dev mode, reload templates on each request
		return
	}
	
	// Load templates from embedded files or filesystem
	tmpl := template.New("")
	
	// Add template functions
	tmpl.Funcs(template.FuncMap{
		"formatDuration": func(d time.Duration) string {
			return d.String()
		},
		"formatTime": func(t time.Time) string {
			return t.Format("15:04:05")
		},
		"formatBytes": func(bytes int64) string {
			const unit = 1024
			if bytes < unit {
				return fmt.Sprintf("%d B", bytes)
			}
			div, exp := int64(unit), 0
			for n := bytes / unit; n >= unit; n /= unit {
				div *= unit
				exp++
			}
			return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
		},
	})
	
	s.templates = tmpl
}

// RegisterRoutes registers HTTP routes
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	// Static files
	mux.HandleFunc("/static/", s.handleStatic)
	
	// API proxy - forward /api/ requests to API server
	mux.HandleFunc("/api/", s.handleAPIProxy)
	
	// Main pages
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/test/new", s.handleNewTest)
	mux.HandleFunc("/test/", s.handleTestDetails)
	mux.HandleFunc("/tests", s.handleTestList)
	mux.HandleFunc("/docs", s.handleDocs)
	mux.HandleFunc("/api-docs", s.handleAPIDocs)
	
	// API endpoints for GUI (legacy)
	mux.HandleFunc("/api/gui/tests", s.handleAPITests)
	mux.HandleFunc("/api/gui/test/start", s.handleAPITestStart)
	mux.HandleFunc("/api/gui/test/stop", s.handleAPITestStop)
	mux.HandleFunc("/api/gui/test/status", s.handleAPITestStatus)
	mux.HandleFunc("/api/gui/presets", s.handleAPIPresets)
}

// handleIndex serves the main dashboard page
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	
	data := struct {
		Title       string
		ActiveTests int
		TotalTests  int
		Uptime      time.Duration
	}{
		Title:       "QUIC Test Dashboard",
		ActiveTests: s.testManager.GetActiveTestCount(),
		TotalTests:  s.testManager.GetTotalTestCount(),
		Uptime:      time.Since(startTime),
	}
	
	s.renderTemplate(w, "index.html", data)
}

// handleNewTest serves the new test creation page
func (s *Server) handleNewTest(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title    string
		Presets  []NetworkPreset
		Profiles []TestProfile
	}{
		Title:    "Create New Test",
		Presets:  getNetworkPresets(),
		Profiles: getTestProfiles(),
	}
	
	s.renderTemplate(w, "new-test.html", data)
}

// handleTestDetails serves individual test details page
func (s *Server) handleTestDetails(w http.ResponseWriter, r *http.Request) {
	testID := strings.TrimPrefix(r.URL.Path, "/test/")
	if testID == "" {
		http.NotFound(w, r)
		return
	}
	
	// Get test data from API server
	apiURL := fmt.Sprintf("http://localhost:8081/api/tests/%s", testID)
	resp, err := http.Get(apiURL)
	if err != nil {
		http.Error(w, "Failed to fetch test data", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode == http.StatusNotFound {
		http.NotFound(w, r)
		return
	}
	
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to fetch test data", http.StatusInternalServerError)
		return
	}
	
	var apiResponse struct {
		Success bool         `json:"success"`
		Data    *TestSession `json:"data"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		http.Error(w, "Failed to parse test data", http.StatusInternalServerError)
		return
	}
	
	if !apiResponse.Success || apiResponse.Data == nil {
		http.NotFound(w, r)
		return
	}
	
	data := struct {
		Title   string
		Session *TestSession
	}{
		Title:   "Test Details - " + testID,
		Session: apiResponse.Data,
	}
	
	s.renderTemplate(w, "test-details.html", data)
}

// handleTestList serves the test list page
func (s *Server) handleTestList(w http.ResponseWriter, r *http.Request) {
	// Get test data from API server
	apiURL := "http://localhost:8081/api/tests"
	resp, err := http.Get(apiURL)
	if err != nil {
		http.Error(w, "Failed to fetch test data", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to fetch test data", http.StatusInternalServerError)
		return
	}
	
	var apiResponse struct {
		Success bool `json:"success"`
		Data    struct {
			Tests []*TestSession `json:"tests"`
		} `json:"data"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		http.Error(w, "Failed to parse test data", http.StatusInternalServerError)
		return
	}
	
	tests := []*TestSession{}
	if apiResponse.Success && apiResponse.Data.Tests != nil {
		tests = apiResponse.Data.Tests
	}
	
	data := struct {
		Title string
		Tests []*TestSession
	}{
		Title: "Test History",
		Tests: tests,
	}
	
	s.renderTemplate(w, "test-list.html", data)
}

// handleDocs serves the documentation page
func (s *Server) handleDocs(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title string
	}{
		Title: "Documentation",
	}
	
	s.renderTemplate(w, "docs.html", data)
}

// handleAPIDocs serves the API documentation page
func (s *Server) handleAPIDocs(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title string
	}{
		Title: "API Documentation",
	}
	
	s.renderTemplate(w, "api-docs.html", data)
}

// handleStatic serves static files
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/static/")
	
	// Security check
	if strings.Contains(path, "..") {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}
	
	// Add cache control headers
	if s.devMode {
		// Disable caching in development mode
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
	} else {
		// Enable caching in production
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}
	
	// Serve from embedded files or filesystem
	staticPath := filepath.Join("web", "static", path)
	http.ServeFile(w, r, staticPath)
}

// handleAPIProxy proxies API requests to the API server
func (s *Server) handleAPIProxy(w http.ResponseWriter, r *http.Request) {
	// Create proxy URL to API server
	apiURL := "http://localhost:8081" + r.URL.Path
	if r.URL.RawQuery != "" {
		apiURL += "?" + r.URL.RawQuery
	}
	
	// Create new request
	proxyReq, err := http.NewRequest(r.Method, apiURL, r.Body)
	if err != nil {
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		return
	}
	
	// Copy headers
	for name, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}
	
	// Make request to API server with longer timeout for DELETE requests
	timeout := 5 * time.Second
	if r.Method == "DELETE" {
		timeout = 30 * time.Second // Longer timeout for stop operations
	}
	
	client := &http.Client{
		Timeout: timeout,
	}
	
	resp, err := client.Do(proxyReq)
	if err != nil {
		fmt.Printf("Proxy request failed: %v\n", err)
		http.Error(w, "Failed to proxy request: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	
	// Copy response headers
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	
	// Set status code
	w.WriteHeader(resp.StatusCode)
	
	// Copy response body
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		fmt.Printf("Error copying response body: %v\n", err)
	}
}

// renderTemplate renders an HTML template
func (s *Server) renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	if s.devMode {
		// Reload templates in dev mode
		s.loadTemplates()
	}
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	// For now, serve a simple HTML response
	// In production, this would use the loaded templates
	s.renderSimpleHTML(w, name, data)
}

// renderSimpleHTML renders a simple HTML response (temporary implementation)
func (s *Server) renderSimpleHTML(w http.ResponseWriter, name string, data interface{}) {
	switch name {
	case "index.html":
		s.renderIndexHTML(w, data)
	case "new-test.html":
		s.renderNewTestHTML(w, data)
	case "test-details.html":
		s.renderTestDetailsHTML(w, data)
	case "test-list.html":
		s.renderTestListHTML(w, data)
	case "docs.html":
		s.renderDocsHTML(w, data)
	case "api-docs.html":
		s.renderAPIDocsHTML(w, data)
	default:
		http.Error(w, "Template not found", http.StatusNotFound)
	}
}

var startTime = time.Now()

// API handlers for GUI operations
func (s *Server) handleAPITests(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		tests := s.testManager.GetAllTests()
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tests)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleAPITestStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var config internal.TestConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	session := s.testManager.StartTest(config)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

func (s *Server) handleAPITestStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	testID := r.URL.Query().Get("id")
	if testID == "" {
		http.Error(w, "Missing test ID", http.StatusBadRequest)
		return
	}
	
	if err := s.testManager.StopTest(testID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleAPITestStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	testID := r.URL.Query().Get("id")
	if testID == "" {
		http.Error(w, "Missing test ID", http.StatusBadRequest)
		return
	}
	
	session := s.testManager.GetTest(testID)
	if session == nil {
		http.Error(w, "Test not found", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

func (s *Server) handleAPIPresets(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	presets := struct {
		NetworkPresets []NetworkPreset `json:"network_presets"`
		TestProfiles   []TestProfile   `json:"test_profiles"`
	}{
		NetworkPresets: getNetworkPresets(),
		TestProfiles:   getTestProfiles(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(presets)
}

// Helper types and functions
type NetworkPreset struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Latency     string `json:"latency"`
	Bandwidth   string `json:"bandwidth"`
	Loss        string `json:"loss"`
}

type TestProfile struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Duration    string `json:"duration"`
	Connections int    `json:"connections"`
	Streams     int    `json:"streams"`
	Rate        int    `json:"rate"`
}

func getNetworkPresets() []NetworkPreset {
	return []NetworkPreset{
		{
			Name:        "fiber",
			Description: "Fiber optic connection (low latency, high bandwidth)",
			Latency:     "5ms",
			Bandwidth:   "1000Mbps",
			Loss:        "0.01%",
		},
		{
			Name:        "mobile",
			Description: "4G/LTE mobile connection",
			Latency:     "50ms",
			Bandwidth:   "50Mbps",
			Loss:        "1%",
		},
		{
			Name:        "satellite",
			Description: "Satellite connection (high latency)",
			Latency:     "600ms",
			Bandwidth:   "25Mbps",
			Loss:        "2%",
		},
		{
			Name:        "wifi",
			Description: "WiFi connection",
			Latency:     "20ms",
			Bandwidth:   "100Mbps",
			Loss:        "0.5%",
		},
	}
}

func getTestProfiles() []TestProfile {
	return []TestProfile{
		{
			Name:        "quick",
			Description: "Quick performance test (30 seconds)",
			Duration:    "30s",
			Connections: 1,
			Streams:     2,
			Rate:        100,
		},
		{
			Name:        "standard",
			Description: "Standard performance test (2 minutes)",
			Duration:    "120s",
			Connections: 2,
			Streams:     4,
			Rate:        200,
		},
		{
			Name:        "intensive",
			Description: "Intensive load test (5 minutes)",
			Duration:    "300s",
			Connections: 4,
			Streams:     8,
			Rate:        500,
		},
		{
			Name:        "endurance",
			Description: "Long-running endurance test (30 minutes)",
			Duration:    "1800s",
			Connections: 2,
			Streams:     4,
			Rate:        100,
		},
	}
}