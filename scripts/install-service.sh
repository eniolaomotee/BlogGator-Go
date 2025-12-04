#!/bin/bash

set -e

echo "=== Gator Service Installer ==="
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

# Get gator executable path
GATOR_PATH=$(which gator)
if [ -z "$GATOR_PATH" ]; then
    echo "Error: gator not found in PATH"
    echo "Please install gator first or add it to your PATH"
    exit 1
fi

echo "Found gator at: $GATOR_PATH"
echo "OS detected: $OS"
echo

if [ "$OS" == "linux" ]; then
    # Linux (systemd)
    echo "Installing systemd service..."
    
    SERVICE_FILE="/etc/systemd/system/gator.service"
    
    # Create service file
    sudo tee $SERVICE_FILE > /dev/null <<EOF
[Unit]
Description=Gator RSS Aggregator Service
After=network.target

[Service]
Type=simple
User=$USER
WorkingDirectory=$HOME
ExecStart=$GATOR_PATH service start
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

Environment="DB_URL=$DB_URL"

[Install]
WantedBy=multi-user.target
EOF

    # Reload systemd
    sudo systemctl daemon-reload
    
    # Enable service
    sudo systemctl enable gator.service
    
    echo "✓ Service installed"
    echo
    echo "To start the service:"
    echo "  sudo systemctl start gator"
    echo
    echo "To check status:"
    echo "  sudo systemctl status gator"
    echo
    echo "To view logs:"
    echo "  sudo journalctl -u gator -f"
    
elif [ "$OS" == "macos" ]; then
    # macOS (launchd)
    echo "Installing launchd service..."
    
    PLIST_FILE="$HOME/Library/LaunchAgents/com.gator.aggregator.plist"
    
    # Create LaunchAgents directory if it doesn't exist
    mkdir -p "$HOME/Library/LaunchAgents"
    
    # Create plist file
    cat > $PLIST_FILE <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.gator.aggregator</string>
    
    <key>ProgramArguments</key>
    <array>
        <string>$GATOR_PATH</string>
        <string>service</string>
        <string>start</string>
    </array>
    
    <key>RunAtLoad</key>
    <true/>
    
    <key>KeepAlive</key>
    <true/>
    
    <key>StandardOutPath</key>
    <string>$HOME/.gator/logs/gator-service.log</string>
    
    <key>StandardErrorPath</key>
    <string>$HOME/.gator/logs/gator-service-error.log</string>
    
    <key>EnvironmentVariables</key>
    <dict>
        <key>DB_URL</key>
        <string>$DB_URL</string>
    </dict>
</dict>
</plist>
EOF

     # Create logs directory
    mkdir -p "$HOME/.gator/logs"
    
    # Load the service
    launchctl load $PLIST_FILE
    
    echo "✓ Service installed"
    echo
    echo "To start the service:"
    echo "  launchctl start com.gator.aggregator"
    echo
    echo "To stop the service:"
    echo "  launchctl stop com.gator.aggregator"
    echo
    echo "To view logs:"
    echo "  tail -f $HOME/.gator/logs/gator-service.log"
fi

echo
echo "Service installation complete!"