#!/bin/bash
# Download OpenMoji color PNG files (72x72)

set -e

OPENMOJI_VERSION="15.0.0"
DOWNLOAD_URL="https://github.com/hfg-gmuend/openmoji/releases/download/${OPENMOJI_VERSION}/openmoji-72x72-color.zip"
TEMP_DIR="openmoji_temp"
TARGET_DIR="emojis"

echo "Downloading OpenMoji ${OPENMOJI_VERSION} color PNG files..."

# Create temp directory
mkdir -p "${TEMP_DIR}"

# Download zip file
curl -L "${DOWNLOAD_URL}" -o "${TEMP_DIR}/openmoji.zip"

# Extract
echo "Extracting..."
unzip -q "${TEMP_DIR}/openmoji.zip" -d "${TEMP_DIR}"

# Move PNG files to target directory
echo "Moving PNG files to ${TARGET_DIR}..."
mkdir -p "${TARGET_DIR}"
mv "${TEMP_DIR}"/*.png "${TARGET_DIR}/"

# Cleanup
echo "Cleaning up..."
rm -rf "${TEMP_DIR}"

echo "Done! Downloaded $(ls -1 ${TARGET_DIR} | wc -l) emoji PNG files."
