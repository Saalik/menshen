#!/bin/bash
set -e

echo "Installing Menshen..."

# Build the binary if it doesn't exist
if [ ! -f "menshen" ]; then
    echo "Building menshen..."
    go build -o menshen ./src
fi

# Detect OS
OS="$(uname -s)"
case "${OS}" in
    Linux*)     machine=Linux;;
    Darwin*)    machine=Mac;;
    *)          echo "Unsupported OS: ${OS}"; exit 1;;
esac

echo "Detected OS: $machine"

# Create standard configuration directory
sudo mkdir -p /etc/menshen

if [ ! -f "/etc/menshen/config.yaml" ]; then
    echo "Installing default configuration..."
    sudo cp config.yaml /etc/menshen/config.yaml
else
    echo "Configuration /etc/menshen/config.yaml already exists. Skipping."
fi

# Install binary
sudo cp menshen /usr/local/bin/menshen
sudo chmod +x /usr/local/bin/menshen

if [ "$machine" == "Linux" ]; then
    echo "Setting up Systemd service..."
    
    # Create user if not exists
    if ! id "menshen" &>/dev/null; then
        sudo useradd -r -s /bin/false -d /opt/menshen menshen
    fi

    # Create working directory
    sudo mkdir -p /opt/menshen/repos
    sudo chown -R menshen:menshen /opt/menshen
    sudo chown -R menshen:menshen /etc/menshen
    
    # Install systemd service
    cat << 'EOF' | sudo tee /etc/systemd/system/menshen.service > /dev/null
[Unit]
Description=Menshen Git Server
After=network.target

[Service]
Type=simple
User=menshen
Group=menshen
WorkingDirectory=/opt/menshen
ExecStart=/usr/local/bin/menshen
Restart=always

[Install]
WantedBy=multi-user.target
EOF

    sudo systemctl daemon-reload
    sudo systemctl enable menshen
    echo "Menshen installed. Start it using: sudo systemctl start menshen"

elif [ "$machine" == "Mac" ]; then
    echo "Setting up Launchd daemon..."
    
    # Create working directory
    sudo mkdir -p /opt/menshen/repos
    sudo chown -R $USER /opt/menshen
    sudo chown -R $USER /etc/menshen
    
    PLIST_PATH="$HOME/Library/LaunchAgents/com.menshen.server.plist"
    cat << EOF > "$PLIST_PATH"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.menshen.server</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/menshen</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>WorkingDirectory</key>
    <string>/opt/menshen</string>
    <key>StandardOutPath</key>
    <string>/opt/menshen/menshen.log</string>
    <key>StandardErrorPath</key>
    <string>/opt/menshen/menshen.error.log</string>
</dict>
</plist>
EOF
    
    launchctl load "$PLIST_PATH"
    echo "Menshen installed and loaded into launchd."
fi

echo "Installation complete!"
