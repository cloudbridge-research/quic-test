package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"quic-test/internal/gui"
)

func main() {
	var (
		addr     = flag.String("addr", ":8080", "GUI server address")
		apiAddr  = flag.String("api-addr", ":8081", "API server address")
		certPath = flag.String("cert", "", "TLS certificate path (optional)")
		keyPath  = flag.String("key", "", "TLS key path (optional)")
		dev      = flag.Bool("dev", false, "Development mode (auto-reload)")
	)
	flag.Parse()

	fmt.Println("QUIC Test GUI Server")
	fmt.Println("===================")
	fmt.Printf("GUI Address: %s\n", *addr)
	fmt.Printf("API Address: %s\n", *apiAddr)
	fmt.Printf("Development Mode: %v\n", *dev)

	// Create GUI server
	guiServer := gui.NewServer(*dev)
	
	// Create API server
	apiServer := gui.NewAPIServer()

	// Setup HTTP servers
	guiMux := http.NewServeMux()
	guiServer.RegisterRoutes(guiMux)
	
	apiMux := http.NewServeMux()
	apiServer.RegisterRoutes(apiMux)

	guiHTTPServer := &http.Server{
		Addr:    *addr,
		Handler: guiMux,
	}

	apiHTTPServer := &http.Server{
		Addr:    *apiAddr,
		Handler: apiMux,
	}

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		fmt.Println("\nShutting down servers...")
		
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := guiHTTPServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("GUI server shutdown error: %v", err)
		}
		
		if err := apiHTTPServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("API server shutdown error: %v", err)
		}
		
		cancel()
	}()

	// Start servers
	go func() {
		fmt.Printf("Starting API server on %s\n", *apiAddr)
		if err := apiHTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("API server failed: %v", err)
		}
	}()

	fmt.Printf("Starting GUI server on %s\n", *addr)
	fmt.Printf("Open http://localhost%s in your browser\n", *addr)
	
	var err error
	if *certPath != "" && *keyPath != "" {
		fmt.Println("Using HTTPS")
		err = guiHTTPServer.ListenAndServeTLS(*certPath, *keyPath)
	} else {
		err = guiHTTPServer.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		log.Fatalf("GUI server failed: %v", err)
	}

	<-ctx.Done()
	fmt.Println("Servers stopped")
}