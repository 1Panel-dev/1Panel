#!/bin/bash

command -v wget >/dev/null || { 
  echo "wget not found, please install it and try again ."
  exit 1
}

if [ ! -f "1pctl" ]; then 
  wget https://github.com/1Panel-dev/installer/raw/main/1pctl
fi

if [ ! -f "1panel.service" ]; then 
  wget https://github.com/1Panel-dev/installer/raw/main/1panel.service
fi

sed -i 's@ORIGINAL_INSTALLED_VERSION=.*@ORIGINAL_INSTALLED_VERSION=v{{ .Version }}@g' 1pctl
chmod 755 1pctl