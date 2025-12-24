package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"quic-test/client"
	"quic-test/internal"
	"quic-test/server"
)

func main() {
	// Add --version flag
	version := flag.Bool("version", false, "Show program version")
	
	fmt.Println("\033[1;36m==========================================\033[0m")
	fmt.Println("\033[1;36m    2GC Network Protocol Suite\033[0m")
	fmt.Println("\033[1;36m==========================================\033[0m")
	fmt.Println("Comprehensive testing of QUIC, MASQUE, ICE/STUN/TURN and other network protocols")
	mode := flag.String("mode", "test", "Mode: server | client | test")
	addr := flag.String("addr", ":9000", "Address for connection or listening")
	streams := flag.Int("streams", 1, "Number of streams per connection")
	connections := flag.Int("connections", 1, "Number of QUIC connections")
	duration := flag.Duration("duration", 0, "Test duration (0 - until manual termination)")
	packetSize := flag.Int("packet-size", 1200, "Packet size (bytes)")
	rate := flag.Int("rate", 100, "Packet sending rate (per second)")
	reportPath := flag.String("report", "", "Path to report file (optional)")
	reportFormat := flag.String("report-format", "md", "Report format: csv | md | json")
	certPath := flag.String("cert", "", "Path to TLS certificate (optional)")
	keyPath := flag.String("key", "", "Path to TLS key (optional)")
	pattern := flag.String("pattern", "random", "Data pattern: random | zeroes | increment")
	noTLS := flag.Bool("no-tls", false, "Disable TLS (for testing)")
	prometheus := flag.Bool("prometheus", false, "Export Prometheus metrics on /metrics")
	quicBottom := flag.Bool("quic-bottom", false, "Start QUIC Bottom for metrics visualization")
	emulateLoss := flag.Float64("emulate-loss", 0, "Packet loss probability (0..1)")
	emulateLatency := flag.Duration("emulate-latency", 0, "Additional latency before packet sending (e.g., 20ms)")
	emulateDup := flag.Float64("emulate-dup", 0, "Packet duplication probability (0..1)")
	
	// FEC flags
	fecEnabled := flag.Bool("enable-fec", false, "Enable Forward Error Correction")
	fecRate := flag.Float64("fec-rate", 0.10, "FEC redundancy level (0.05-0.20, e.g. 0.05=5%, 0.10=10%, 0.20=20%)")
	// Alias for backward compatibility
	fecEnabledAlias := flag.Bool("fec", false, "Alias for --enable-fec")
	fecRedundancyAlias := flag.Float64("fec-redundancy", 0.10, "Alias for --fec-rate")
	
	// PQC flags
	pqcEnabled := flag.Bool("pqc", false, "Enable Post-Quantum Cryptography (simulation)")
	pqcAlgorithm := flag.String("pqc-algorithm", "ml-kem-768", "PQC algorithm: ml-kem-512, ml-kem-768, dilithium-2, hybrid, baseline")
	
	// SLA flags
	slaRttP95 := flag.Duration("sla-rtt-p95", 0, "SLA: maximum RTT p95 (e.g., 100ms)")
	slaLoss := flag.Float64("sla-loss", 0, "SLA: maximum packet loss (0..1, e.g., 0.01 for 1%)")
	slaThroughput := flag.Float64("sla-throughput", 0, "SLA: minimum throughput (KB/s)")
	slaErrors := flag.Int64("sla-errors", 0, "SLA: maximum number of errors")
	
	// QUIC tuning flags
	cc := flag.String("cc", "", "Congestion control algorithm: cubic, bbr, bbrv2, bbrv3, reno")
	maxIdleTimeout := flag.Duration("max-idle-timeout", 0, "Maximum connection idle timeout")
	handshakeTimeout := flag.Duration("handshake-timeout", 0, "Handshake timeout")
	keepAlive := flag.Duration("keep-alive", 0, "Keep-alive interval")
	maxStreams := flag.Int64("max-streams", 0, "Maximum number of streams")
	maxStreamData := flag.Int64("max-stream-data", 0, "Maximum stream data size")
	enable0RTT := flag.Bool("enable-0rtt", false, "Enable 0-RTT")
	enableKeyUpdate := flag.Bool("enable-key-update", false, "Enable key update")
	enableDatagrams := flag.Bool("enable-datagrams", false, "Enable datagrams")
	maxIncomingStreams := flag.Int64("max-incoming-streams", 0, "Maximum number of incoming streams")
	maxIncomingUniStreams := flag.Int64("max-incoming-uni-streams", 0, "Maximum number of incoming unidirectional streams")
	
	// Test scenarios
	scenario := flag.String("scenario", "", "Predefined scenario: wifi, lte, sat, dc-eu, ru-eu, loss-burst, reorder")
	listScenarios := flag.Bool("list-scenarios", false, "Show list of available scenarios")
	
	// Network profiles
	networkProfile := flag.String("network-profile", "", "Network profile: wifi, lte, 5g, satellite, ethernet, fiber, datacenter")
	listProfiles := flag.Bool("list-profiles", false, "Show list of available network profiles")
	
	flag.Parse()

	// Handle --version flag
	if *version {
		internal.PrintVersion()
		os.Exit(0)
	}

	cfg := internal.TestConfig{
		Mode:           *mode,
		Addr:           *addr,
		Streams:        *streams,
		Connections:    *connections,
		Duration:       *duration,
		PacketSize:     *packetSize,
		Rate:           *rate,
		ReportPath:     *reportPath,
		ReportFormat:   *reportFormat,
		CertPath:       *certPath,
		KeyPath:        *keyPath,
		Pattern:        *pattern,
		NoTLS:          *noTLS,
		Prometheus:     *prometheus,
		EmulateLoss:    *emulateLoss,
		EmulateLatency: *emulateLatency,
		EmulateDup:     *emulateDup,
		SlaRttP95:      *slaRttP95,
		SlaLoss:        *slaLoss,
		SlaThroughput:  *slaThroughput,
		SlaErrors:      *slaErrors,
		CongestionControl: *cc,
		MaxIdleTimeout:    *maxIdleTimeout,
		HandshakeTimeout:  *handshakeTimeout,
		KeepAlive:         *keepAlive,
		MaxStreams:        *maxStreams,
		MaxStreamData:      *maxStreamData,
		Enable0RTT:        *enable0RTT,
		EnableKeyUpdate:   *enableKeyUpdate,
		EnableDatagrams:   *enableDatagrams,
		MaxIncomingStreams: *maxIncomingStreams,
		MaxIncomingUniStreams: *maxIncomingUniStreams,
		FECEnabled:       *fecEnabled || *fecEnabledAlias,
		FECRedundancy:    func() float64 {
			if *fecEnabled || *fecEnabledAlias {
				if *fecRedundancyAlias != 0.10 {
					return *fecRedundancyAlias
				}
				return *fecRate
			}
			return 0
		}(),
		PQCEnabled:       *pqcEnabled,
		PQCAlgorithm:     *pqcAlgorithm,
	}

	fmt.Printf("mode=%s, addr=%s, connections=%d, streams=%d, duration=%s, packet-size=%d, rate=%d, report=%s, report-format=%s, cert=%s, key=%s, pattern=%s, no-tls=%v, prometheus=%v\n",
		cfg.Mode, cfg.Addr, cfg.Connections, cfg.Streams, cfg.Duration.String(), cfg.PacketSize, cfg.Rate, cfg.ReportPath, cfg.ReportFormat, cfg.CertPath, cfg.KeyPath, cfg.Pattern, cfg.NoTLS, cfg.Prometheus)
	
	// Print SLA configuration if set
	internal.PrintSLAConfig(cfg)
	
	// Print QUIC configuration if set
	internal.PrintQUICConfig(cfg)
	
	// Start QUIC Bottom if requested
	if *quicBottom {
		fmt.Println("Starting QUIC Bottom for real-time metrics visualization...")
		go func() {
			// Start QUIC Bottom in background mode
			cmd := exec.Command("./quic-bottom/target/release/quic-bottom-real")
			cmd.Dir = "."
			if err := cmd.Run(); err != nil {
				fmt.Printf("Failed to start QUIC Bottom: %v\n", err)
			}
		}()
		
		// Wait a bit for QUIC Bottom to start
		time.Sleep(2 * time.Second)
		fmt.Println("QUIC Bottom started on port 8080")
	}

	// Handle scenarios
	if *listScenarios {
		fmt.Println("Available Test Scenarios:")
		scenarios := internal.ListScenarios()
		for _, name := range scenarios {
			scenario, _ := internal.GetScenario(name)
			fmt.Printf("  - %s: %s\n", name, scenario.Description)
		}
		os.Exit(0)
	}
	
	// Handle network profiles
	if *listProfiles {
		fmt.Println("Available Network Profiles:")
		profiles := internal.ListNetworkProfiles()
		for _, name := range profiles {
			profile, _ := internal.GetNetworkProfile(name)
			fmt.Printf("  - %s: %s\n", name, profile.Description)
		}
		os.Exit(0)
	}
	
	if *scenario != "" {
		scenarioConfig, err := internal.GetScenario(*scenario)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}
		
		// Apply scenario configuration
		cfg = scenarioConfig.Config
		fmt.Printf("Running scenario: %s\n", scenarioConfig.Name)
	}
	
	if *networkProfile != "" {
		profile, err := internal.GetNetworkProfile(*networkProfile)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}
		
		// Apply network profile
		internal.ApplyNetworkProfile(&cfg, profile)
		internal.PrintNetworkProfile(profile)
		internal.PrintProfileRecommendations(profile)
	}

	// Initialize QUIC Bottom (use 127.0.0.1 instead of localhost to avoid IPv6 issues)
	internal.InitBottomBridge("http://127.0.0.1:8080", 100*time.Millisecond)
	internal.EnableBottomBridge()

	// Handle signals for graceful shutdown
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func(cancelFunc context.CancelFunc) {
		<-sigs
		fmt.Println("\nReceived termination signal, shutting down...")
		cancelFunc() // Correct termination
	}(cancel)

	switch cfg.Mode {
	case "server":
		fmt.Println("Starting in server mode...")
		server.Run(cfg)
	case "client":
		fmt.Println("Starting in client mode...")
		client.Run(cfg)
	case "test":
		fmt.Println("Starting in test mode (server+client)...")
		runTestMode(cfg)
	default:
		fmt.Println("Unknown mode", cfg.Mode)
		os.Exit(1)
	}
}

// runTestMode starts server and client for testing
func runTestMode(cfg internal.TestConfig) {
	// Start server in goroutine
	serverDone := make(chan struct{})
	go func() {
		defer close(serverDone)
		server.Run(cfg)
	}()

	// Wait for server to start
	time.Sleep(3 * time.Second)

	// Start client
	client.Run(cfg)

	// Give server time to shutdown gracefully (maximum 5 seconds)
	serverTimeout := time.NewTimer(5 * time.Second)
	select {
	case <-serverDone:
		serverTimeout.Stop()
	case <-serverTimeout.C:
		fmt.Println("Server shutdown timeout, exiting...")
	}
}
