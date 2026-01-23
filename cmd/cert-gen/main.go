package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"quic-test/internal/ca"
)

func main() {
	var (
		hosts    = flag.String("hosts", "localhost,127.0.0.1,::1", "Comma-separated list of hosts for the certificate")
		certPath = flag.String("cert", "certs/server.crt", "Output certificate path")
		keyPath  = flag.String("key", "certs/server.key", "Output key path")
		caPath   = flag.String("ca-cert", "certs/ca.crt", "CA certificate path")
		caKey    = flag.String("ca-key", "certs/ca.key", "CA key path")
		initCA   = flag.Bool("init-ca", false, "Initialize new CA")
	)
	flag.Parse()

	if *initCA {
		// Initialize CA
		caInstance := ca.NewCA(*caPath, *caKey)
		if err := caInstance.Initialize(); err != nil {
			fmt.Fprintf(os.Stderr, "Error initializing CA: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("CA initialized successfully\n")
		fmt.Printf("CA certificate: %s\n", *caPath)
		fmt.Printf("CA key: %s\n", *caKey)
		return
	}

	// Generate server certificate
	caInstance := ca.NewCA(*caPath, *caKey)
	if err := caInstance.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading CA: %v\n", err)
		os.Exit(1)
	}

	hostList := strings.Split(*hosts, ",")
	for i, host := range hostList {
		hostList[i] = strings.TrimSpace(host)
	}

	if err := caInstance.GenerateServerCertificate(hostList, *certPath, *keyPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating server certificate: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Server certificate generated successfully\n")
	fmt.Printf("Certificate: %s\n", *certPath)
	fmt.Printf("Key: %s\n", *keyPath)
	fmt.Printf("Hosts: %v\n", hostList)
}