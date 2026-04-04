#!/bin/bash
set -euo pipefail

APP_NAME="twitch-tts"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
INSTALL_DIR="$HOME/.local/bin"
ICON_DIR="$HOME/.local/share/icons"
DESKTOP_DIR="$HOME/.local/share/applications"

if [ ! -f "$SCRIPT_DIR/${APP_NAME}" ]; then
  echo "Error: ${APP_NAME} binary not found in ${SCRIPT_DIR}"
  echo "Run this script from the extracted release directory."
  exit 1
fi

echo "Installing ${APP_NAME}..."

mkdir -p "$INSTALL_DIR"
cp "$SCRIPT_DIR/${APP_NAME}" "$INSTALL_DIR/${APP_NAME}"
chmod +x "$INSTALL_DIR/${APP_NAME}"

mkdir -p "$ICON_DIR"
cp "$SCRIPT_DIR/appicon.png" "$ICON_DIR/${APP_NAME}.png"

mkdir -p "$DESKTOP_DIR"
cat > "$DESKTOP_DIR/${APP_NAME}.desktop" <<EOF
[Desktop Entry]
Name=Twitch TTS
Comment=Twitch chat TTS reader using VOICEVOX
Exec=${INSTALL_DIR}/${APP_NAME}
Icon=${ICON_DIR}/${APP_NAME}.png
Terminal=false
Type=Application
Categories=AudioVideo;Audio;Network;
Keywords=twitch;tts;voicevox;streaming;
EOF

echo "Installed to ${INSTALL_DIR}/${APP_NAME}"
echo "You can launch 'Twitch TTS' from your app launcher."
