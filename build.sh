#!/bin/bash

# Build script for ShellCast

echo "Building ShellCast..."

# Check if Go is installed
if ! command -v go &> /dev/null
then
    echo "Error: Go is not installed"
    exit 1
fi

# Ensure all files exist
for file in config.go shellcast.go interactive.go main.go; do
    if [[ ! -f $file ]]; then
        echo "Error: Required file $file not found"
        exit 1
    fi
done

# Build the application
go build -o shellcast config.go shellcast.go interactive.go main.go

# Check if build was successful
if [[ $? -eq 0 && -f shellcast ]]; then
    echo "Build successful: $(pwd)/shellcast"
    echo ""
    echo "Usage examples:"
    echo "  ./shellcast -interactive"
    echo "  ./shellcast -rtmp rtmp://server/app ls -la"
    echo "  ./shellcast -theme hacker -timestamp on -record command"
    echo "  ./shellcast -split \"ls -la\" \"top -n 1\""
    chmod +x shellcast
else
    echo "Build failed"
    exit 1
fi
