#!/bin/bash

# Script to install QUIC Test CA certificate into system trust store

set -e

CA_CERT="certs/ca.crt"
CA_NAME="quic-test-ca"

if [ ! -f "$CA_CERT" ]; then
    echo "Error: CA certificate not found at $CA_CERT"
    echo "Please run a test first to generate the CA certificate"
    exit 1
fi

echo "Installing QUIC Test CA certificate into system trust store..."

# Detect OS and install accordingly
if [ -f /etc/debian_version ]; then
    # Debian/Ubuntu
    echo "Detected Debian/Ubuntu system"
    sudo cp "$CA_CERT" "/usr/local/share/ca-certificates/${CA_NAME}.crt"
    sudo update-ca-certificates
    echo "CA certificate installed successfully on Debian/Ubuntu"
    
elif [ -f /etc/redhat-release ]; then
    # RedHat/CentOS/Fedora
    echo "Detected RedHat/CentOS/Fedora system"
    sudo cp "$CA_CERT" "/etc/pki/ca-trust/source/anchors/${CA_NAME}.crt"
    sudo update-ca-trust
    echo "CA certificate installed successfully on RedHat/CentOS/Fedora"
    
elif [ -f /etc/arch-release ]; then
    # Arch Linux
    echo "Detected Arch Linux system"
    sudo trust anchor "$CA_CERT"
    echo "CA certificate installed successfully on Arch Linux"
    
else
    echo "Unsupported OS. Please manually install the CA certificate:"
    echo "CA certificate location: $(pwd)/$CA_CERT"
    exit 1
fi

echo ""
echo "CA certificate installed successfully!"
echo "You can now use HTTPS connections without certificate warnings."
echo ""
echo "To verify the installation:"
echo "  openssl verify -CAfile $CA_CERT certs/server.crt"
echo ""
echo "To remove the CA certificate later:"
if [ -f /etc/debian_version ]; then
    echo "  sudo rm /usr/local/share/ca-certificates/${CA_NAME}.crt"
    echo "  sudo update-ca-certificates --fresh"
elif [ -f /etc/redhat-release ]; then
    echo "  sudo rm /etc/pki/ca-trust/source/anchors/${CA_NAME}.crt"
    echo "  sudo update-ca-trust"
elif [ -f /etc/arch-release ]; then
    echo "  sudo trust anchor --remove $CA_CERT"
fi