#!/usr/bin/env bash
set -euo pipefail

helper_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
mkdir -p "${helper_dir}/bin"

for arch in amd64 arm64; do
    target="${helper_dir}/bin/1panel-kvm-${arch}"
    if [ -s "${target}" ]; then
        continue
    fi
    download="${target}.download"
    trap 'rm -f "${download}"' EXIT
    curl --fail --location --retry 3 --connect-timeout 15 --max-time 300 \
        "https://resource.fit2cloud.com/1panel/resource/kvm/1panel-kvm-${arch}" \
        --output "${download}"
    test -s "${download}"
    chmod 755 "${download}"
    mv "${download}" "${target}"
done
