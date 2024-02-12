#!/bin/bash
# Instalar a Última Versão Estável do 1Panel

osCheck=`uname -a`
if [[ $osCheck =~ 'x86_64' ]];then
    architecture="amd64"
elif [[ $osCheck =~ 'arm64' ]] || [[ $osCheck =~ 'aarch64' ]];then
    architecture="arm64"
elif [[ $osCheck =~ 'armv7l' ]];then
    architecture="armv7"
elif [[ $osCheck =~ 'ppc64le' ]];then
    architecture="ppc64le"
elif [[ $osCheck =~ 's390x' ]];then
    architecture="s390x"
else
    echo "Arquitetura do sistema não suportada atualmente. Consulte a documentação oficial para selecionar um sistema suportado."
    exit 1
fi

if [[ ! ${INSTALL_MODE} ]];then
    INSTALL_MODE="stable"
else
    if [[ ${INSTALL_MODE} != "dev" && ${INSTALL_MODE} != "stable" ]];then
        echo "Por favor, insira um modo de instalação válido (dev ou stable)"
        exit 1
    fi
fi

VERSION=$(curl -s https://resource.fit2cloud.com/1panel/package/${INSTALL_MODE}/latest)
HASH_FILE_URL="https://resource.fit2cloud.com/1panel/package/${INSTALL_MODE}/${VERSION}/release/checksums.txt"

if [[ "x${VERSION}" == "x" ]];then
    echo "Falha ao obter a versão mais recente. Por favor, tente novamente mais tarde."
    exit 1
fi

package_file_name="1panel-${VERSION}-linux-${architecture}.tar.gz"
package_download_url="https://resource.fit2cloud.com/1panel/package/${INSTALL_MODE}/${VERSION}/release/${package_file_name}"
expected_hash=$(curl -s "$HASH_FILE_URL" | grep "$package_file_name" | awk '{print $1}')

if [ -f ${package_file_name} ];then
    actual_hash=$(sha256sum "$package_file_name" | awk '{print $1}')
    if [[ "$expected_hash" == "$actual_hash" ]];then
        echo "O pacote de instalação já existe. Pulando o download."
        rm -rf 1panel-${VERSION}-linux-${architecture}
        tar zxvf ${package_file_name}
        cd 1panel-${VERSION}-linux-${architecture}
        /bin/bash install.sh
        exit 0
    else
        echo "O pacote de instalação já existe, mas o hash não corresponde. Começando o download novamente."
        rm -f ${package_file_name}
    fi
fi

echo "Iniciando o download da versão ${VERSION} do 1Panel."
echo "URL de download do pacote: ${package_download_url}"

curl -LOk -o ${package_file_name} ${package_download_url}
curl -sfL https://resource.fit2cloud.com/installation-log.sh | sh -s 1p install ${VERSION}
if [ ! -f ${package_file_name} ];then
    echo "Falha ao baixar o pacote de instalação. Por favor, tente novamente mais tarde."
    exit 1
fi

tar zxvf ${package_file_name}
if [ $? != 0 ];then
    echo "Falha ao baixar o pacote de instalação. Por favor, tente novamente mais tarde."
    rm -f ${package_file_name}
    exit 1
fi
cd 1panel-${VERSION}-linux-${architecture}

/bin/bash install.sh
