#!/bin/bash

set -e

download_resources() {
  command -v wget >/dev/null || {
    echo "wget not found, please install it and try again."
    exit 1
  }

  if [ ! -f "1pctl" ]; then
    wget https://github.com/1Panel-dev/installer/raw/v2/1pctl
  fi

  if [ ! -f "install.sh" ]; then
    wget https://github.com/1Panel-dev/installer/raw/v2/install.sh
  fi

  if [ ! -f "1panel-core.service" ]; then
    wget https://github.com/1Panel-dev/installer/raw/v2/initscript/1panel-core.service
  fi

  if [ ! -f "1panel-agent.service" ]; then
    wget https://github.com/1Panel-dev/installer/raw/v2/initscript/1panel-agent.service
  fi

  if [ ! -d "lang" ]; then
    mkdir -p lang && cd lang
    for lang in en fa pt-BR ru zh; do
      wget -q https://github.com/1Panel-dev/installer/raw/v2/lang/$lang.sh
    done
    cd ..
  fi

  if [ ! -f "GeoIP.mmdb" ]; then
    wget https://resource.fit2cloud.com/1panel/package/v2/geo/GeoIP.mmdb
  fi

  chmod 755 1pctl install.sh
}

compress_binary() {
  local binary_path="$1"
  local arch="$2"

  echo "Attempting to compress: $binary_path for arch: $arch"

  if [ "$arch" = "s390x" ]; then
    echo "Skipping UPX compression for s390x"
    return
  fi

  if ! command -v upx >/dev/null; then
    echo "Installing upx..."
    if [ "$(uname -m)" = "s390x" ]; then
      echo "UPX not supported on s390x"
      return
    fi

    sudo apt-get update
    sudo apt-get install -y upx-ucl
  fi

  upx --best --lzma "$binary_path" && echo "[ok] Compressed: $binary_path" || echo "[warn] Failed to compress: $binary_path"
}

if [ "$1" = "compress_binary" ]; then
  compress_binary "$2" "$3"
else
  download_resources
fi
