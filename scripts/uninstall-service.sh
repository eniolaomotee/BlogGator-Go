#!/bin/bash

set -e

echo "=== Gator Service Uninstaller ==="
echo

# Detect OS
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    OS="linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    OS="macos"
else
    echo "Unsupported OS: $OSTYPE"
    exit 1
fi

if [ "$OS" == "linux" ]; then
    echo "Uninstalling systemd service..."
    
    # Stop service
    sudo systemctl stop gator.service 2>/dev/null || true
    
    # Disable service
    sudo systemctl disable gator.service 2>/dev/null || true
    
    # Remove service file
    sudo rm -f /etc/systemd/system/gator.service
    
    # Reload systemd
    sudo systemctl daemon-reload
    
    echo "✓ Service uninstalled"
    
elif [ "$OS" == "macos" ]; then
    echo "Uninstalling launchd service..."
    
    PLIST_FILE="$HOME/Library/LaunchAgents/com.gator.aggregator.plist"
    
    # Unload service
    launchctl unload $PLIST_FILE 2>/dev/null || true
    
    # Remove plist file
    rm -f $PLIST_FILE
    
    echo "✓ Service uninstalled"
fi

echo
echo "Service uninstallation complete!"
echo "Note: Log files in ~/.gator/logs have been preserved"