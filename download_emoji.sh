#!/bin/bash
# Download Noto Color Emoji PNG files (136x128)

set -e

NOTO_VERSION="v2.042"
DOWNLOAD_URL="https://github.com/googlefonts/noto-emoji/archive/refs/tags/${NOTO_VERSION}.tar.gz"
TEMP_DIR="noto_temp"
TARGET_DIR="emojis"

echo "Downloading Noto Color Emoji ${NOTO_VERSION} PNG files..."

# Create temp directory
mkdir -p "${TEMP_DIR}"

# Download tarball
curl -L "${DOWNLOAD_URL}" -o "${TEMP_DIR}/noto-emoji.tar.gz"

# Extract
echo "Extracting..."
tar -xzf "${TEMP_DIR}/noto-emoji.tar.gz" -C "${TEMP_DIR}"

# Move PNG files from png/136 to target directory
echo "Moving PNG files to ${TARGET_DIR}..."
mkdir -p "${TARGET_DIR}"
find "${TEMP_DIR}" -name "*.png" -path "*/png/136/*" -exec cp {} "${TARGET_DIR}/" \;

# Cleanup
echo "Cleaning up..."
rm -rf "${TEMP_DIR}"

EMOJI_COUNT=$(ls -1 "${TARGET_DIR}"/*.png 2>/dev/null | wc -l)
echo "Done! Downloaded ${EMOJI_COUNT} emoji PNG files."
