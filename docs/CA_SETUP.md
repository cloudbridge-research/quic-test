# Certificate Authority Setup

This document describes how to use the built-in Certificate Authority (CA) functionality in the QUIC Test Suite.

## Overview

The QUIC Test Suite includes a built-in Certificate Authority that automatically generates and manages TLS certificates for testing. This eliminates certificate warnings and provides a more realistic testing environment.

## Features

- Automatic CA certificate generation
- Server certificate generation with multiple SANs (Subject Alternative Names)
- Support for localhost, IP addresses, and custom domains
- System trust store integration
- Command-line certificate generation tool

## Quick Start

### Automatic Certificate Generation

When you run a test without specifying certificates, the system automatically:

1. Creates a CA certificate (if it doesn't exist)
2. Generates a server certificate for localhost, 127.0.0.1, and ::1
3. Uses these certificates for secure QUIC connections

```bash
./quic-test -mode=test -addr=localhost:9000 -duration=10s
```

Output:
```
Created new CA certificate: certs/ca.crt
Generated server certificate for hosts [localhost 127.0.0.1 ::1]: certs/server.crt
```

### Manual Certificate Generation

Use the `cert-gen` tool to create certificates for specific hosts:

```bash
# Initialize CA (optional - done automatically)
./cert-gen -init-ca

# Generate certificate for specific hosts
./cert-gen -hosts="example.com,*.example.com,api.example.com" \
           -cert="certs/example.crt" \
           -key="certs/example.key"
```

## Certificate Files

The system creates the following files in the `certs/` directory:

- `ca.crt` - CA certificate (public)
- `ca.key` - CA private key
- `server.crt` - Default server certificate
- `server.key` - Default server private key

## Installing CA in System Trust Store

To eliminate certificate warnings in browsers and system tools:

### Linux (Ubuntu/Debian)
```bash
sudo cp certs/ca.crt /usr/local/share/ca-certificates/quic-test-ca.crt
sudo update-ca-certificates
```

### Linux (RedHat/CentOS/Fedora)
```bash
sudo cp certs/ca.crt /etc/pki/ca-trust/source/anchors/quic-test-ca.crt
sudo update-ca-trust
```

### Linux (Arch)
```bash
sudo trust anchor certs/ca.crt
```

### Automated Installation
Use the provided script:
```bash
./scripts/install-ca.sh
```

## Browser Configuration

### Firefox
1. Go to Settings → Privacy & Security → Certificates
2. Click "View Certificates"
3. Go to "Authorities" tab
4. Click "Import..." and select `certs/ca.crt`
5. Check "Trust this CA to identify websites"

### Chrome/Chromium
1. Go to Settings → Privacy and security → Security
2. Click "Manage certificates"
3. Go to "Authorities" tab
4. Click "Import" and select `certs/ca.crt`

## Certificate Verification

Verify that certificates are properly signed:

```bash
# Verify server certificate against CA
openssl verify -CAfile certs/ca.crt certs/server.crt

# View certificate details
openssl x509 -in certs/server.crt -text -noout

# Check certificate chain
openssl s_client -connect localhost:9000 -CAfile certs/ca.crt
```

## Advanced Usage

### Custom Certificate Paths

Specify custom certificate paths:

```bash
./quic-test -mode=server -cert=custom/server.crt -key=custom/server.key
```

### Multiple Domains

Generate certificates for multiple domains:

```bash
./cert-gen -hosts="localhost,example.com,*.example.com,192.168.1.100" \
           -cert="certs/multi-domain.crt" \
           -key="certs/multi-domain.key"
```

### Certificate Rotation

Regenerate certificates periodically:

```bash
# Remove old certificates
rm certs/server.crt certs/server.key

# Run test to generate new certificates
./quic-test -mode=test -addr=localhost:9000 -duration=1s
```

## Security Considerations

- CA private key (`ca.key`) should be protected
- Certificates are valid for 1 year by default
- Use different CAs for different environments (dev/staging/prod)
- Remove test CA from production systems

## Troubleshooting

### Certificate Not Trusted
- Ensure CA is installed in system trust store
- Check browser certificate settings
- Verify certificate chain with `openssl verify`

### Permission Errors
- Ensure write permissions to `certs/` directory
- Use `sudo` for system trust store installation

### Certificate Expired
- Remove old certificates and regenerate
- Check system clock synchronization

## Integration with Testing

The CA system integrates seamlessly with all test modes:

```bash
# Server mode with automatic certificates
./quic-test -mode=server -addr=localhost:9000

# Client mode connecting to CA-signed server
./quic-test -mode=client -addr=localhost:9000

# Integrated test with automatic certificate setup
./quic-test -mode=test -addr=localhost:9000 -connections=5 -duration=30s
```

This provides a complete testing environment with proper TLS security without manual certificate management.