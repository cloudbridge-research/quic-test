#!/bin/sh

# Docker entrypoint script for QUIC Test Suite

if [ "$1" = "gui" ]; then
    echo "Starting QUIC Test GUI..."
    exec ./quic-gui --addr=:8080 --api-addr=:8081
elif [ "$1" = "dashboard" ]; then
    echo "Starting Dashboard..."
    exec ./dashboard
elif [ "$1" = "client" ]; then
    echo "Starting QUIC Client..."
    exec ./quic-client "${@:2}"
elif [ "$1" = "server" ]; then
    echo "Starting QUIC Server..."
    exec ./quic-server "${@:2}"
else
    # Default: run main quic-test binary
    exec ./quic-test "$@"
fi