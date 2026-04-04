#!/bin/bash
set -euo pipefail

APP_NAME="twitch-tts"
INSTALL_DIR="$HOME/.local/bin"
ICON_DIR="$HOME/.local/share/icons"
DESKTOP_DIR="$HOME/.local/share/applications"

echo "Removing ${APP_NAME}..."
rm -f "$INSTALL_DIR/${APP_NAME}"
rm -f "$ICON_DIR/${APP_NAME}.png"
rm -f "$DESKTOP_DIR/${APP_NAME}.desktop"

echo "Done!"
