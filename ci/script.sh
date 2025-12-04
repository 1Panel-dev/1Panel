#!/bin/bash

command -v wget >/dev/null || { 
  echo "wget not found, please install it and try again ."
  exit 1
}

if [ ! -f "1pctl" ]; then 
  wget https://github.com/1Panel-dev/installer/raw/v1/1pctl
fi

if [ ! -f "install.sh" ]; then 
  wget https://github.com/1Panel-dev/installer/raw/v1/install.sh
fi

if [ ! -d "initscript" ]; then
  wget https://github.com/1Panel-dev/installer/raw/v1/initscript/1panel.service
  mkdir -p initscript && cd initscript
  for file in 1panel.service 1paneld.init 1paneld.openrc 1paneld.procd; do
    wget -q https://github.com/1Panel-dev/installer/raw/v1/initscript/$file
  done
  cd ..
fi

if [ ! -d "lang" ]; then
  mkdir -p lang && cd lang
  for lang in en fa pt-BR ru zh; do
    wget -q https://github.com/1Panel-dev/installer/raw/v1/lang/$lang.sh
  done
  cd ..
fi

if [ ! -f "GeoIP.mmdb" ]; then 
  wget https://resource.fit2cloud.com/1panel/package/geo/GeoIP.mmdb
fi

chmod 755 1pctl install.sh
