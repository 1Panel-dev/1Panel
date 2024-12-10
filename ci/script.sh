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

if [ ! -f "install.sh" ]; then 
  wget https://github.com/1Panel-dev/installer/raw/main/install.sh
fi

if [ ! -f "GeoIP.mmdb" ]; then 
  wget https://resource.1panel.hk/geo/GeoIP.mmdb
fi

if [ ! -f "lang.tar.gz" ]; then 
  wget https://resource.1panel.hk/language/lang.tar.gz
  tar zxvf lang.tar.gz
  rm -rf lang.tar.gz
fi

chmod 755 1pctl install.sh
