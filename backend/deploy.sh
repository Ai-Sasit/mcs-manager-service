#!/bin/bash

# Configuration
APP_NAME="mc-manage-backend"
SERVICE_NAME="mc-manage"
INSTALL_DIR="/opt/mc-manage/backend"
BIN_NAME="mc-manage-backend"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check for root
function check_root() {
    if [[ $EUID -ne 0 ]]; then
        echo -e "${RED}Error: This script must be run as root (sudo).${NC}"
        exit 1
    fi
}

function show_menu() {
    clear
    echo "=========================================="
    echo "   MC Manage Backend Deployment Menu"
    echo "=========================================="
    echo "1) Full Setup / Update (Build & Install)"
    echo "2) Start Service"
    echo "3) Stop Service"
    echo "4) Restart Service"
    echo "5) Check Status"
    echo "6) View Logs (Tailing)"
    echo "7) Build Only"
    echo "8) Exit"
    echo "=========================================="
    echo -n "Choose an option [1-8]: "
}

function build_app() {
    echo -e "${YELLOW}Building Go binary...${NC}"
    if ! command -v go &> /dev/null; then
        echo -e "${RED}Error: Go is not installed. Please install Go first.${NC}"
        return 1
    fi
    go build -o "$BIN_NAME" main.go
    if [[ $? -eq 0 ]]; then
        echo -e "${GREEN}Build successful!${NC}"
    else
        echo -e "${RED}Build failed!${NC}"
        return 1
    fi
}

function setup_service() {
    check_root
    echo -e "${YELLOW}Setting up systemd service...${NC}"
    
    # Ensure install dir exists
    mkdir -p "$INSTALL_DIR"
    
    # Copy binary
    cp "$BIN_NAME" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/$BIN_NAME"
    
    # Copy .env if exists
    if [[ -f ".env" ]]; then
        cp ".env" "$INSTALL_DIR/"
        echo -e "${GREEN}Copied .env file.${NC}"
    fi

    # Create/Copy service file
    if [[ -f "mc-manage.service" ]]; then
        cp "mc-manage.service" "/etc/systemd/system/$SERVICE_NAME.service"
    else
        # Inline template if file missing
        cat <<EOF > "/etc/systemd/system/$SERVICE_NAME.service"
[Unit]
Description=Minecraft Management Backend
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/$BIN_NAME
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
    fi

    systemctl daemon-reload
    systemctl enable "$SERVICE_NAME"
    echo -e "${GREEN}Service setup complete and enabled!${NC}"
}

function start_service() {
    check_root
    echo -e "${YELLOW}Starting $SERVICE_NAME...${NC}"
    systemctl start "$SERVICE_NAME"
    echo -e "${GREEN}Done.${NC}"
}

function stop_service() {
    check_root
    echo -e "${YELLOW}Stopping $SERVICE_NAME...${NC}"
    systemctl stop "$SERVICE_NAME"
    echo -e "${GREEN}Done.${NC}"
}

function restart_service() {
    check_root
    echo -e "${YELLOW}Restarting $SERVICE_NAME...${NC}"
    systemctl restart "$SERVICE_NAME"
    echo -e "${GREEN}Done.${NC}"
}

function check_status() {
    echo -e "${YELLOW}Service Status:${NC}"
    systemctl status "$SERVICE_NAME"
}

function view_logs() {
    echo -e "${YELLOW}Tailing logs... (Press Ctrl+C to stop)${NC}"
    journalctl -u "$SERVICE_NAME" -f
}

# Main loop
while true; do
    show_menu
    read choice
    case $choice in
        1)
            build_app && setup_service && restart_service
            ;;
        2)
            start_service
            ;;
        3)
            stop_service
            ;;
        4)
            restart_service
            ;;
        5)
            check_status
            ;;
        6)
            view_logs
            ;;
        7)
            build_app
            ;;
        8)
            echo "Exiting..."
            exit 0
            ;;
        *)
            echo -e "${RED}Invalid option!${NC}"
            ;;
    esac
    echo -n "Press Enter to continue..."
    read
done
