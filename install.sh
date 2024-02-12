#!/bin/bash

function log() {
    message="[1Panel Log]: $1 "
    echo -e "${message}" 2>&1 | tee -a ${CURRENT_DIR}/install.log
}

log "======================= Iniciando a Instalação ======================="

function Check_Root() {
  if [[ $EUID -ne 0 ]]; then
    echo "Por favor, execute este script como root ou utilize sudo."
    exit 1
  fi
}

function Prepare_System(){
    if which 1panel >/dev/null 2>&1; then
        log "O Painel de Administração de Servidores Linux 1Panel já está instalado. Não é necessário instalar novamente."
        exit 1
    fi
}

function Set_Dir(){
    if read -t 120 -p "Defina o diretório de instalação do 1Panel (padrão é /opt so mecher se souber oque ta fazendo):" PANEL_BASE_DIR;then
        if [[ "$PANEL_BASE_DIR" != "" ]];then
            if [[ "$PANEL_BASE_DIR" != /* ]];then
                log "Por favor, insira o caminho completo para o diretório."
                Set_Dir
            fi

            if [[ ! -d $PANEL_BASE_DIR ]];then
                mkdir -p $PANEL_BASE_DIR
                log "O diretório de instalação selecionado é $PANEL_BASE_DIR"
            fi
        else
            PANEL_BASE_DIR=/opt
            log "O diretório de instalação selecionado é $PANEL_BASE_DIR"
        fi
    else
        PANEL_BASE_DIR=/opt
        log "(O tempo para definir o diretório expirou, utilizando o diretório padrão /opt)"
    fi
}

function Install_Docker(){
    if which docker >/dev/null 2>&1; then
        log "Docker está instalado. Pulando a instalação."
        log "Iniciando o Docker."
        systemctl start docker 2>&1 | tee -a ${CURRENT_DIR}/install.log
    else
        log "... Instalando o Docker online."

        if [[ $(curl -s ipinfo.io/country) == "CN" ]]; then
            sources=(
                "https://mirrors.aliyun.com/docker-ce"
                "https://mirrors.tencent.com/docker-ce"
                "https://mirrors.163.com/docker-ce"
                "https://mirrors.cernet.edu.cn/docker-ce"
            )

            get_average_delay() {
                local source=$1
                local total_delay=0
                local iterations=3

                for ((i = 0; i < iterations; i++)); do
                    delay=$(curl -o /dev/null -s -w "%{time_total}\n" "$source")
                    total_delay=$(awk "BEGIN {print $total_delay + $delay}")
                done

                average_delay=$(awk "BEGIN {print $total_delay / $iterations}")
                echo "$average_delay"
            }

            min_delay=${#sources[@]}
            selected_source=""

            for source in "${sources[@]}"; do
                average_delay=$(get_average_delay "$source")

                if (( $(awk 'BEGIN { print '"$average_delay"' < '"$min_delay"' }') )); then
                    min_delay=$average_delay
                    selected_source=$source
                fi
            done

            if [ -n "$selected_source" ]; then
                echo "Selecionando a fonte com o menor atraso $selected_source, atraso de $min_delay segundos"
                export DOWNLOAD_URL="$selected_source"
                curl -fsSL "https://get.docker.com" -o get-docker.sh
                sh get-docker.sh 2>&1 | tee -a ${CURRENT_DIR}/install.log

                log "... Iniciando o Docker."
                systemctl enable docker; systemctl daemon-reload; systemctl start docker 2>&1 | tee -a ${CURRENT_DIR}/install.log

                docker_config_folder="/etc/docker"
                if [[ ! -d "$docker_config_folder" ]];then
                    mkdir -p "$docker_config_folder"
                fi

                docker version >/dev/null 2>&1
                if [[ $? -ne 0 ]]; then
                    log "A instalação do Docker falhou."
                    exit 1
                else
                    log "Docker instalado com sucesso."
                fi
            else
                log "Não foi possível selecionar uma fonte para instalação."
                exit 1
            fi
        else
            log "Fora da China, nenhuma alteração de fonte é necessária."
            export DOWNLOAD_URL="https://download.docker.com"
            curl -fsSL "https://get.docker.com" -o get-docker.sh
            sh get-docker.sh 2>&1 | tee -a ${CURRENT_DIR}/install.log

            log "... Iniciando o Docker."
            systemctl enable docker; systemctl daemon-reload; systemctl start docker 2>&1 | tee -a ${CURRENT_DIR}/install.log

            docker_config_folder="/etc/docker"
            if [[ ! -d "$docker_config_folder" ]];then
                mkdir -p "$docker_config_folder"
            fi

            docker version >/dev/null 2>&1
            if [[ $? -ne 0 ]]; then
                log "A instalação do Docker falhou."
                exit 1
            else
                log "Docker instalado com sucesso."
            fi
        fi
    fi
}

function Install_Compose(){
    docker-compose version >/dev/null 2>&1
    if [[ $? -ne 0 ]]; then
        log "... Instalando o docker-compose online"
        
        arch=$(uname -m)
        if [ "$arch" == 'armv7l' ]; then
            arch='armv7'
        fi
        curl -L https://resource.fit2cloud.com/docker/compose/releases/download/v2.22.0/docker-compose-$(uname -s | tr A-Z a-z)-$arch -o /usr/local/bin/docker-compose 2>&1 | tee -a ${CURRENT_DIR}/install.log
        if [[ ! -f /usr/local/bin/docker-compose ]];then
            log "Falha ao baixar docker-compose. Por favor, tente novamente."
            exit 1
        fi
        chmod +x /usr/local/bin/docker-compose
        ln -s /usr/local/bin/docker-compose /usr/bin/docker-compose

        docker-compose version >/dev/null 2>&1
        if [[ $? -ne 0 ]]; then
            log "Falha na instalação do docker-compose."
            exit 1
        else
            log "docker-compose instalado com sucesso."
        fi
    else
        compose_v=`docker-compose -v`
        if [[ $compose_v =~ 'docker-compose' ]];then
            read -p "Foi detectado que uma versão mais antiga do Docker Compose está instalada (deve ser superior ou igual a v2.0.0), você deseja atualizar [s/n] : " UPGRADE_DOCKER_COMPOSE
            if [[ "$UPGRADE_DOCKER_COMPOSE" == "s" ]] || [[ "$UPGRADE_DOCKER_COMPOSE" == "S" ]]; then
                rm -rf /usr/local/bin/docker-compose /usr/bin/docker-compose
                Install_Compose
            else
                log "A versão atual do Docker Compose é $compose_v, o que pode afetar o funcionamento adequado da Loja de Aplicativos."
            fi
        else
            log "Foi detectado que o Docker Compose está instalado. Pulando a instalação."
        fi
    fi
}

function Set_Port(){
    DEFAULT_PORT=`expr $RANDOM % 55535 + 10000`

    while true; do
        read -p "Defina a porta do 1Panel (padrão é $DEFAULT_PORT):" PANEL_PORT

        if [[ "$PANEL_PORT" == "" ]];then
            PANEL_PORT=$DEFAULT_PORT
        fi

        if ! [[ "$PANEL_PORT" =~ ^[1-9][0-9]{0,4}$ && "$PANEL_PORT" -le 65535 ]]; then
            echo "Erro: A porta deve estar entre 1 e 65535"
            continue
        fi

        log "Porta definida: $PANEL_PORT"
        break
    done
}

function Set_Firewall(){
    if which firewall-cmd >/dev/null 2>&1; then
        if systemctl status firewalld | grep -q "Active: active" >/dev/null 2>&1;then
            log "Abrindo a porta $PANEL_PORT no firewall."
            firewall-cmd --zone=public --add-port=$PANEL_PORT/tcp --permanent
            firewall-cmd --reload
        else
            log "Firewall não está ativado, ignorando a abertura de porta."
        fi
    fi

    if which ufw >/dev/null 2>&1; then
        if systemctl status ufw | grep -q "Active: active" >/dev/null 2>&1;then
            log "Abrindo a porta $PANEL_PORT no firewall."
            ufw allow $PANEL_PORT/tcp
            ufw reload
        else
            log "Firewall não está ativado, ignorando a abertura de porta."
        fi
    fi
}

function Set_Username(){
    DEFAULT_USERNAME=`cat /dev/urandom | head -n 16 | md5sum | head -c 10`

    while true; do
        read -p "Defina o nome de usuário do 1Panel (padrão é $DEFAULT_USERNAME):" PANEL_USERNAME

        if [[ "$PANEL_USERNAME" == "" ]];then
            PANEL_USERNAME=$DEFAULT_USERNAME
        fi

        if [[ ! "$PANEL_USERNAME" =~ ^[a-zA-Z0-9_]{3,30}$ ]]; then
            echo "Erro: O nome de usuário só pode conter letras, números e sublinhados, com comprimento de 3 a 30 caracteres."
            continue
        fi

        log "Nome de usuário definido: $PANEL_USERNAME"
        break
    done
}

function Set_Password(){
    DEFAULT_PASSWORD=`cat /dev/urandom | head -n 16 | md5sum | head -c 10`

    while true; do
        echo "Defina a senha do 1Panel (padrão é $DEFAULT_PASSWORD):"
        read -s PANEL_PASSWORD

        if [[ "$PANEL_PASSWORD" == "" ]];then
            PANEL_PASSWORD=$DEFAULT_PASSWORD
        fi

        if [[ ! "$PANEL_PASSWORD" =~ ^[a-zA-Z0-9_!@#$%*,.?]{8,30}$ ]]; then
            echo "Erro: A senha deve conter apenas letras, números e caracteres especiais (!@#$%*_,.?), com comprimento entre 8 e 30 caracteres."
            continue
        fi

        break
    done
}

function Init_Panel(){
    log "Configurando o Serviço 1Panel"

    RUN_BASE_DIR=$PANEL_BASE_DIR/1panel
    mkdir -p $RUN_BASE_DIR
    rm -rf $RUN_BASE_DIR/*

    cd ${CURRENT_DIR}

    cp ./1panel /usr/local/bin && chmod +x /usr/local/bin/1panel
    if [[ ! -f /usr/bin/1panel ]]; then
        ln -s /usr/local/bin/1panel /usr/bin/1panel >/dev/null 2>&1
    fi

    cp ./1pctl /usr/local/bin && chmod +x /usr/local/bin/1pctl
    sed -i -e "s#BASE_DIR=.*#BASE_DIR=${PANEL_BASE_DIR}#g" /usr/local/bin/1pctl
    sed -i -e "s#ORIGINAL_PORT=.*#ORIGINAL_PORT=${PANEL_PORT}#g" /usr/local/bin/1pctl
    sed -i -e "s#ORIGINAL_USERNAME=.*#ORIGINAL_USERNAME=${PANEL_USERNAME}#g" /usr/local/bin/1pctl
    ESCAPED_PANEL_PASSWORD=$(echo "$PANEL_PASSWORD" | sed 's/[!@#$%*_,.?]/\\&/g')
    sed -i -e "s#ORIGINAL_PASSWORD=.*#ORIGINAL_PASSWORD=${ESCAPED_PANEL_PASSWORD}#g" /usr/local/bin/1pctl
    PANEL_ENTRANCE=`cat /dev/urandom | head -n 16 | md5sum | head -c 10`
    sed -i -e "s#ORIGINAL_ENTRANCE=.*#ORIGINAL_ENTRANCE=${PANEL_ENTRANCE}#g" /usr/local/bin/1pctl
    if [[ ! -f /usr/bin/1pctl ]]; then
        ln -s /usr/local/bin/1pctl /usr/bin/1pctl >/dev/null 2>&1
    fi

    cp ./1panel.service /etc/systemd/system

    systemctl enable 1panel; systemctl daemon-reload 2>&1 | tee -a ${CURRENT_DIR}/install.log

    log "Iniciando o Serviço 1Panel"
    systemctl start 1panel | tee -a ${CURRENT_DIR}/install.log

    for b in {1..30}
    do
        sleep 3
        service_status=`systemctl status 1panel 2>&1 | grep Active`
        if [[ $service_status == *running* ]];then
            log "Serviço 1Panel iniciado com sucesso!"
            break;
        else
            log "Erro ao iniciar o serviço 1Panel!"
            exit 1
        fi
    done
}

function Get_Ip(){
    active_interface=$(ip route get 8.8.8.8 | awk 'NR==1 {print $5}')
    if [[ -z $active_interface ]]; then
        LOCAL_IP="127.0.0.1"
    else
        LOCAL_IP=`ip -4 addr show dev "$active_interface" | grep -oP '(?<=inet\s)\d+(\.\d+){3}'`
    fi

    PUBLIC_IP=`curl -s https://api64.ipify.org`
    if [[ -z "$PUBLIC_IP" ]]; then
        PUBLIC_IP="N/A"
    fi
    if echo "$PUBLIC_IP" | grep -q ":"; then
        PUBLIC_IP=[${PUBLIC_IP}]
        1pctl listen-ip ipv6
    fi
}

function Show_Result(){
    log ""
    log "================= Obrigado por sua paciência, a instalação foi concluída =================="
    log ""
    log "Por favor, acesse o painel pelo navegador:"
    log "Endereço público: http://$PUBLIC_IP:$PANEL_PORT/$PANEL_ENTRANCE"
    log "Endereço local: http://$LOCAL_IP:$PANEL_PORT/$PANEL_ENTRANCE"
    log "Usuário do painel: $PANEL_USERNAME"
    log "Senha do painel: $PANEL_PASSWORD"
    log ""
    log "Site do Projeto: https://1panel.cn"
    log "Documentação: https://1panel.cn/docs"
    log "Repositório de Código: https://github.com/1Panel-dev/1Panel"
    log ""
    log "Se estiver usando um servidor em nuvem, abra a porta $PANEL_PORT no grupo de segurança."
    log ""
    log "================================================================"
}

function main(){
    Check_Root
    Prepare_System
    Set_Dir
    Install_Docker
    Install_Compose
    Set_Port
    Set_Firewall
    Set_Username
    Set_Password
    Init_Panel
    Get_Ip
    Show_Result
}
main
