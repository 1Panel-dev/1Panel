package helper

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/global"
)

func LoadScript() {
	groups := []model.Group{
		{Name: "install", Type: "script", IsDefault: false},
		{Name: "docker", Type: "script", IsDefault: false},
		{Name: "firewall", Type: "script", IsDefault: false},
		{Name: "supervisor", Type: "script", IsDefault: false},
		{Name: "clamav", Type: "script", IsDefault: false},
		{Name: "ftp", Type: "script", IsDefault: false},
		{Name: "fail2ban", Type: "script", IsDefault: false}}
	_ = global.DB.Where("`type` = ?", "script").Delete(&model.Group{}).Error
	_ = global.DB.Create(&groups).Error

	_ = global.DB.Where("is_system = ?", 1).Delete(model.ScriptLibrary{}).Error
	list := []model.ScriptLibrary{
		{Name: "Install Docker", IsSystem: true, Groups: fmt.Sprintf("%v,%v", groups[1].ID, groups[0].ID), Script: "bash <(curl -sSL https://linuxmirrors.cn/docker.sh)"},

		{Name: "Install Firewall", IsSystem: true, Groups: fmt.Sprintf("%v,%v", groups[2].ID, groups[0].ID), Script: loadInstallFirewall()},

		{Name: "Install Supervisor", IsSystem: true, Groups: fmt.Sprintf("%v,%v", groups[3].ID, groups[0].ID), Script: loadInstallSupervisor()},

		{Name: "Install ClamAV", IsSystem: true, Groups: fmt.Sprintf("%v,%v", groups[4].ID, groups[0].ID), Script: loadInstallClamAV()},

		{Name: "Install Pure-FTPd", IsSystem: true, Groups: fmt.Sprintf("%v,%v", groups[5].ID, groups[0].ID), Script: loadInstallFTP()},

		{Name: "Install Fail2ban", IsSystem: true, Groups: fmt.Sprintf("%v,%v", groups[6].ID, groups[0].ID), Script: loadInstallFail2ban()},
	}

	_ = global.DB.Create(&list).Error
}

func loadInstallFirewall() string {
	return `#!/bin/bash

# 防火墙 安装配置脚本
# 支持 Ubuntu/Debian/CentOS/RHEL/Alpine/Arch Linux

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

FIREWALL=""


# 检测操作系统
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        VERSION=$VERSION_ID
    elif type lsb_release >/dev/null 2>&1; then
        OS=$(lsb_release -si | tr '[:upper:]' '[:lower:]')
        VERSION=$(lsb_release -sr)
    elif [ -f /etc/redhat-release ]; then
        OS="rhel"
        VERSION=$(grep -oE '[0-9]+\.[0-9]+' /etc/redhat-release)
    elif [ -f /etc/alpine-release ]; then
        OS="alpine"
        VERSION=$(cat /etc/alpine-release)
    else
        OS=$(uname -s | tr '[:upper:]' '[:lower:]')
        VERSION=$(uname -r)
    fi
}

# 安装防火墙
install_firewall() {
    if [ "$OS" == "ubuntu" ] || [ "$OS" == "debian" ]; then
        echo -e "${GREEN}检测到操作系统: $OS $VERSION，正在安装 ufw...${NC}"
        FIREWALL="ufw"
        apt-get update
        apt-get install -y ufw
    elif [ "$OS" == "centos" ] || [ "$OS" == "rhel" ] || [ "$OS_LIKE" == "rhel" ]; then
        echo -e "${GREEN}检测到操作系统: $OS $VERSION，正在安装 firewall...${NC}"
        FIREWALL="firewalld"
        yum update
        yum install -y firewalld
    else
        echo -e "${RED}不支持的操作系统${NC}"
        exit 1
    fi
}

# 初始化并启动
start_with_init() {
    read -p "请输入需要放行的端口(多个端口用空格分隔，如 80 443 22): " PORTS

    # 验证端口输入
    if [ -z "$PORTS" ]; then
        echo -e "${RED}错误：未输入任何端口${NC}"
        exit 1
    fi

    case $FIREWALL in
        firewalld)
            echo -e "${GREEN}配置firewalld...${NC}"
            echo "初始化并启动firewalld..."
            systemctl start firewalld
            systemctl enable firewalld
            
            for port in $PORTS; do
                firewall-cmd --zone=public --permanent --add-port="$port/tcp"
            done
            
            firewall-cmd --reload
            echo -e "${GREEN}已放行以下TCP端口: $PORTS ${NC}"
            ;;
            
        ufw)
            echo -e "${GREEN}初始化并启动ufw...${NC}"
            ufw --force enable
            
            for port in $PORTS; do
                ufw allow "$port/tcp"
            done
            
            echo -e "${GREEN}已放行以下TCP端口: $PORTS ${NC}"
            ;;
    esac
}

# 检查防火墙是否正常运行
check_install() {
    if [ "$FIREWALL" = "firewalld" ]; then
        if command -v firewall-cmd &> /dev/null; then
            systemctl status firewalld || true
        fi
    else 
        if command -v ufw &> /dev/null; then
            ufw status || true
        fi
    fi

    echo -e "${GREEN}$FIREWALL 安装完成并启动${NC}"
}

# 主函数
main() {
    detect_os
    install_firewall
    start_with_init
    check_install
}

main "$@"`
}

func loadInstallFTP() string {
	return `#!/bin/bash

# Pure-FTPd 安装配置脚本
# 支持 Ubuntu/Debian/CentOS/RHEL/Alpine/Arch Linux

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color


# 检测操作系统
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        VERSION=$VERSION_ID
    elif type lsb_release >/dev/null 2>&1; then
        OS=$(lsb_release -si | tr '[:upper:]' '[:lower:]')
        VERSION=$(lsb_release -sr)
    elif [ -f /etc/redhat-release ]; then
        OS="rhel"
        VERSION=$(grep -oE '[0-9]+\.[0-9]+' /etc/redhat-release)
    elif [ -f /etc/alpine-release ]; then
        OS="alpine"
        VERSION=$(cat /etc/alpine-release)
    else
        OS=$(uname -s | tr '[:upper:]' '[:lower:]')
        VERSION=$(uname -r)
    fi
}

# 安装Pure-FTPd
install_pureftpd() {
    echo -e "${GREEN}检测到操作系统: $OS $VERSION${NC}"

    case "$OS" in
        ubuntu|debian)
            apt-get update
            apt-get install -y pure-ftpd
            ;;
        centos|rhel|fedora)
            if [ "$OS" = "rhel" ] && [ "${VERSION%%.*}" -ge 8 ]; then
                dnf install -y epel-release
                dnf install -y pure-ftpd
            else
                yum install -y epel-release
                yum install -y pure-ftpd
            fi
            ;;
        alpine)
            apk add --no-cache pure-ftpd
            ;;
        arch)
            pacman -Sy --noconfirm pure-ftpd
            ;;
        *)
            echo -e "${RED}不支持的操作系统${NC}"
            exit 1
            ;;
    esac

    if ! command -v pure-ftpd &> /dev/null; then
        echo -e "${RED}Pure-FTPd安装失败${NC}"
        exit 1
    fi
}

# 配置Pure-FTPd
configure_pureftpd() {
    echo -e "${GREEN}配置Pure-FTPd...${NC}"
    
    PURE_FTPD_CONF="/etc/pure-ftpd/pure-ftpd.conf"
    if [ -f "$PURE_FTPD_CONF" ]; then
        cp "$PURE_FTPD_CONF" "$PURE_FTPD_CONF.bak"
        sed -i 's/^NoAnonymous[[:space:]]\+no$/NoAnonymous yes/' "$PURE_FTPD_CONF"
        sed -i 's/^PAMAuthentication[[:space:]]\+yes$/PAMAuthentication no/' "$PURE_FTPD_CONF"
        sed -i 's/^# PassivePortRange[[:space:]]\+30000 50000$/PassivePortRange 39000 40000/' "$PURE_FTPD_CONF"
        sed -i 's/^VerboseLog[[:space:]]\+no$/VerboseLog yes/' "$PURE_FTPD_CONF"
        sed -i 's/^# PureDB[[:space:]]\+\/etc\/pure-ftpd\/pureftpd\.pdb[[:space:]]*$/PureDB \/etc\/pure-ftpd\/pureftpd.pdb/' "$PURE_FTPD_CONF"
    else
        touch /etc/pure-ftpd/pureftpd.pdb
        chmod 644 /etc/pure-ftpd/pureftpd.pdb
        echo '/etc/pure-ftpd/pureftpd.pdb' > /etc/pure-ftpd/conf/PureDB
        echo yes > /etc/pure-ftpd/conf/VerboseLog 
        echo yes > /etc/pure-ftpd/conf/NoAnonymous
        echo '39000 40000' > /etc/pure-ftpd/conf/PassivePortRange
        echo 'no' > /etc/pure-ftpd/conf/PAMAuthentication
        echo 'no' > /etc/pure-ftpd/conf/UnixAuthentication
        echo 'clf:/var/log/pure-ftpd/transfer.log' > /etc/pure-ftpd/conf/AltLog
        ln -s /etc/pure-ftpd/conf/PureDB /etc/pure-ftpd/auth/50puredb
    fi
}

# 启动服务
start_service() {
    echo -e "${GREEN}启动Pure-FTPd服务...${NC}"
    
    case "$OS" in
        ubuntu|debian)
            systemctl enable pure-ftpd
            systemctl restart pure-ftpd
            ;;
        centos|rhel|fedora)
            systemctl enable pure-ftpd
            systemctl restart pure-ftpd
            ;;
        alpine)
            rc-update add pure-ftpd
            rc-service pure-ftpd start
            ;;
        arch)
            systemctl enable pure-ftpd
            systemctl restart pure-ftpd
            ;;
        *)
            echo -e "${YELLOW}无法自动启动服务，请手动启动${NC}"
            ;;
    esac

    # 验证服务状态
    if command -v systemctl &> /dev/null; then
        systemctl status pure-ftpd || true
    else
        rc-service pure-ftpd status || true
    fi
}



# 主函数
main() {
    detect_os
    install_pureftpd
    configure_pureftpd
    start_service
}

main "$@"`
}

func loadInstallClamAV() string {
	return `#!/bin/bash

# ClamAV 安装启动脚本
# 支持 Ubuntu/Debian/CentOS/RHEL/Alpine/Arch Linux

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# 检测操作系统
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        VERSION=$VERSION_ID
    elif type lsb_release >/dev/null 2>&1; then
        OS=$(lsb_release -si | tr '[:upper:]' '[:lower:]')
        VERSION=$(lsb_release -sr)
    elif [ -f /etc/redhat-release ]; then
        OS="rhel"
        VERSION=$(grep -oE '[0-9]+\.[0-9]+' /etc/redhat-release)
    elif [ -f /etc/alpine-release ]; then
        OS="alpine"
        VERSION=$(cat /etc/alpine-release)
    else
        OS=$(uname -s | tr '[:upper:]' '[:lower:]')
        VERSION=$(uname -r)
    fi
}

# 安装ClamAV
install_clamav() {
    echo -e "${GREEN}检测到操作系统: $OS $VERSION${NC}"

    case "$OS" in
        ubuntu|debian)
            apt-get update
            apt-get install -y clamav clamav-daemon clamav-freshclam
            ;;
        centos|rhel|fedora)
            if [ "$OS" = "rhel" ] && [ "${VERSION%%.*}" -ge 8 ]; then
                dnf install -y epel-release
                dnf install -y clamav clamd clamav-update
            else
                yum install -y epel-release
                yum install -y clamav clamd clamav-update
            fi
            ;;
        alpine)
            apk add --no-cache clamav clamav-libunrar clamav-daemon clamav-freshclam
            ;;
        arch)
            pacman -Sy --noconfirm clamav
            ;;
        *)
            echo -e "${RED}不支持的操作系统${NC}"
            exit 1
            ;;
    esac

    if ! command -v clamscan &> /dev/null; then
        echo -e "${RED}ClamAV安装失败${NC}"
        exit 1
    fi
}

# clamd
configure_clamd() {
    echo -e "${GREEN}配置clamd...${NC}"
    
    # 备份原始配置
    CLAMD_CONF=""
    if [ -f "/etc/clamd.d/scan.conf" ]; then
        CLAMD_CONF="/etc/clamd.d/scan.conf"
    elif [ -f "/etc/clamav/clamd.conf" ]; then
        CLAMD_CONF="/etc/clamav/clamd.conf"
    else
        echo "未找到 freshclam 配置文件，请手动配置"
        exit 1
    fi
    cp "$CLAMD_CONF" "$CLAMD_CONF.bak"

    # 禁用检查新版本以避免权限问题
    sed -i 's|^LogFileMaxSize .*|LogFileMaxSize 2M|' "$CLAMD_CONF"
    sed -i 's|^PidFile .*|PidFile /run/clamd.scan/clamd.pid|' "$CLAMD_CONF"
    sed -i 's|^DatabaseDirectory .*|DatabaseDirectory /var/lib/clamav|' "$CLAMD_CONF"
    sed -i 's|^LocalSocket .*|LocalSocket /run/clamd.scan/clamd.sock|' "$CLAMD_CONF"
}

# 配置freshclam
configure_freshclam() {
    echo -e "${GREEN}配置freshclam...${NC}"
    
    # 备份原始配置
    FRESHCLAM_CONF=""
    if [ -f "/etc/freshclam.conf" ]; then
        FRESHCLAM_CONF="/etc/freshclam.conf"
    elif [ -f "/etc/clamav/freshclam.conf" ]; then
        FRESHCLAM_CONF="/etc/clamav/freshclam.conf"
    else
        echo "未找到 freshclam 配置文件，请手动配置"
        exit 1
    fi
    cp "$FRESHCLAM_CONF" "$FRESHCLAM_CONF.bak"

    # 禁用检查新版本以避免权限问题
    sed -i 's|^DatabaseDirectory .*|DatabaseDirectory /var/lib/clamav|' "$FRESHCLAM_CONF"
    sed -i 's|^PidFile .*|PidFile /var/run/freshclam.pid|' "$FRESHCLAM_CONF"
    sed -i '/^DatabaseMirror/d' "$FRESHCLAM_CONF"
    echo "DatabaseMirror database.clamav.net" | sudo tee -a "$FRESHCLAM_CONF"
    sed -i 's|^Checks .*|Checks 12|' "$FRESHCLAM_CONF"
}

# 下载病毒数据库
download_database() {
    systemctl stop clamav-freshclam
    echo -e "${GREEN}开始下载病毒数据库...${NC}"
    
    MAX_RETRIES=5
    RETRY_DELAY=60
    ATTEMPT=1
    
    while [ $ATTEMPT -le $MAX_RETRIES ]; do
        echo -e "${YELLOW}尝试 $ATTEMPT/$MAX_RETRIES: 运行freshclam...${NC}"
        
        if freshclam --verbose; then
            echo -e "${GREEN}成功下载病毒数据库${NC}"
            return 0
        fi
        
        if [ $ATTEMPT -lt $MAX_RETRIES ]; then
            echo -e "${YELLOW}下载失败，等待 $RETRY_DELAY 秒后重试...${NC}"
            sleep $RETRY_DELAY
        fi
        
        ATTEMPT=$((ATTEMPT+1))
    done
    
    echo -e "${RED}错误: 无法在 $MAX_RETRIES 次尝试后下载病毒数据库${NC}" >&2
    exit 1
}

# 启动ClamAV服务
start_services() {
    echo -e "${GREEN}启动ClamAV服务...${NC}"
    
    case "$OS" in
        ubuntu|debian)
            systemctl enable --now clamav-daemon
            systemctl enable --now clamav-freshclam
            ;;
        centos|rhel|fedora)
            systemctl enable --now clamd@scan
            systemctl enable --now clamav-freshclam
            ;;
        alpine)
            rc-update add clamd boot
            rc-update add freshclam boot
            rc-service clamd start
            rc-service freshclam start
            ;;
        arch)
            systemctl enable --now clamav-daemon
            systemctl enable --now clamav-freshclam
            ;;
        *)
            echo -e "${YELLOW}无法自动启动服务，请手动启动${NC}"
            ;;
    esac
    
    # 验证服务状态
    if command -v systemctl &> /dev/null; then
        systemctl status clamav-daemon || true
        systemctl status clamav-freshclam || true
    fi
    
    echo -e "${GREEN}ClamAV安装完成并启动${NC}"
}

# 主函数
main() {
    detect_os
    install_clamav
    configure_clamd
    configure_freshclam
    download_database
    start_services
}

main "$@"`
}

func loadInstallFail2ban() string {
	return `#!/bin/bash

# Fail2ban 安装配置脚本
# 支持 Ubuntu/Debian/CentOS/RHEL/Alpine/Arch Linux

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color


# 检测操作系统
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        VERSION=$VERSION_ID
    elif type lsb_release >/dev/null 2>&1; then
        OS=$(lsb_release -si | tr '[:upper:]' '[:lower:]')
        VERSION=$(lsb_release -sr)
    elif [ -f /etc/redhat-release ]; then
        OS="rhel"
        VERSION=$(grep -oE '[0-9]+\.[0-9]+' /etc/redhat-release)
    elif [ -f /etc/alpine-release ]; then
        OS="alpine"
        VERSION=$(cat /etc/alpine-release)
    else
        OS=$(uname -s | tr '[:upper:]' '[:lower:]')
        VERSION=$(uname -r)
    fi
}

# 安装Fail2ban
install_fail2ban() {
    echo -e "${GREEN}检测到操作系统: $OS $VERSION${NC}"

    case "$OS" in
        ubuntu|debian)
            apt-get update
            apt-get install -y fail2ban
            ;;
        centos|rhel|fedora)
            if [ "$OS" = "rhel" ] && [ "${VERSION%%.*}" -ge 8 ]; then
                dnf install -y epel-release
                dnf install -y fail2ban
            else
                yum install -y epel-release
                yum install -y fail2ban
            fi
            ;;
        alpine)
            apk add --no-cache fail2ban
            ;;
        arch)
            pacman -Sy --noconfirm fail2ban
            ;;
        *)
            echo -e "${RED}不支持的操作系统${NC}"
            exit 1
            ;;
    esac

    sleep 2
    if command -v systemctl &> /dev/null; then
        systemctl status fail2ban --no-pager || true
    else
        rc-service fail2ban status || true
    fi

    fail2ban-client status
}

# 配置Fail2ban
configure_fail2ban() {
    echo -e "${GREEN}配置Fail2ban...${NC}"
    
    FAIL2BAN_CONF="/etc/fail2ban/jail.local"
    LOG_FILE=""
    BAN_ACTION=""

    if systemctl is-active --quiet firewalld 2>/dev/null; then
        BAN_ACTION="firewallcmd-ipset"
    elif systemctl is-active --quiet ufw 2>/dev/null || service ufw status 2>/dev/null | grep -q "active"; then
        BAN_ACTION="ufw"
    else
        BAN_ACTION="iptables-allports"
    fi

    if [ -f /var/log/secure ]; then
        LOG_FILE="/var/log/secure"
    else
        LOG_FILE="/var/log/auth.log"
    fi

    cat <<EOF > "$FAIL2BAN_CONF"
#DEFAULT-START
[DEFAULT]
bantime = 600
findtime = 300
maxretry = 5
banaction = $BAN_ACTION
action = %(action_mwl)s
#DEFAULT-END

[sshd]
ignoreip = 127.0.0.1/8
enabled = true
filter = sshd
port = 22
maxretry = 5
findtime = 300
bantime = 600
banaction = $BAN_ACTION
action = %(action_mwl)s
logpath = $LOG_FILE
EOF
}

# 启动服务
start_service() {
    echo -e "${GREEN}启动Fail2ban服务...${NC}"
    
    case "$OS" in
        ubuntu|debian)
            systemctl enable fail2ban
            systemctl restart fail2ban
            ;;
        centos|rhel|fedora)
            systemctl enable fail2ban
            systemctl restart fail2ban
            ;;
        alpine)
            rc-update add fail2ban
            rc-service fail2ban start
            ;;
        arch)
            systemctl enable fail2ban
            systemctl restart fail2ban
            ;;
        *)
            echo -e "${YELLOW}无法自动启动服务，请手动启动${NC}"
            ;;
    esac

    # 验证服务状态
    if command -v systemctl &> /dev/null; then
        systemctl status fail2ban || true
    else
        rc-service fail2ban status || true
    fi
}



# 主函数
main() {
    detect_os
    install_fail2ban
    configure_fail2ban
    start_service
}

main "$@"`
}

func loadInstallSupervisor() string {
	return `#!/bin/bash

# Supervisor 安装管理脚本
# 功能：自动安装 + 基础配置 + 进程管理模板
# 支持 Ubuntu/Debian/CentOS/RHEL/Alpine/Arch Linux

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# 检测操作系统
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        VERSION=$VERSION_ID
    elif type lsb_release >/dev/null 2>&1; then
        OS=$(lsb_release -si | tr '[:upper:]' '[:lower:]')
        VERSION=$(lsb_release -sr)
    elif [ -f /etc/redhat-release ]; then
        OS="rhel"
        VERSION=$(grep -oE '[0-9]+\.[0-9]+' /etc/redhat-release)
    elif [ -f /etc/alpine-release ]; then
        OS="alpine"
        VERSION=$(cat /etc/alpine-release)
    else
        OS=$(uname -s | tr '[:upper:]' '[:lower:]')
        VERSION=$(uname -r)
    fi
}

# 安装Supervisor
install_supervisor() {
    echo -e "${GREEN}检测到操作系统: $OS $VERSION${NC}"

    case "$OS" in
        ubuntu|debian)
            apt-get update
            apt-get install -y supervisor
            ;;
        centos|rhel|fedora)
            if [ "$OS" = "rhel" ] && [ "${VERSION%%.*}" -ge 8 ]; then
                dnf install -y supervisor
            else
                yum install -y supervisor
            fi
            ;;
        alpine)
            apk add --no-cache supervisor
            mkdir -p /etc/supervisor.d
            ;;
        arch)
            pacman -Sy --noconfirm supervisor
            ;;
        *)
            echo -e "${RED}不支持的操作系统，尝试pip安装...${NC}"
            if ! command -v pip &> /dev/null; then
                python -m ensurepip --upgrade
            fi
            pip install supervisor
            ;;
    esac
}

# 启动服务
start_service() {
    echo -e "${GREEN}启动Supervisor服务...${NC}"
    
    case "$OS" in
        ubuntu|debian)
            systemctl enable supervisor
            systemctl restart supervisor
            ;;
        centos|rhel|fedora)
            systemctl enable supervisor
            systemctl restart supervisor
            ;;
        alpine)
            rc-update add supervisor
            rc-service supervisor start
            ;;
        arch)
            systemctl enable supervisor
            systemctl restart supervisor
            ;;
        *)
            echo -e "${YELLOW}无法自动启动服务，请手动启动${NC}"
            ;;
    esac

    # 验证服务状态
   if ! command -v supervisord &> /dev/null; then
        echo -e "${RED}Supervisor安装失败${NC}"
        exit 1
    fi
}

# 主函数
main() {
    detect_os
    install_supervisor
    start_service
}

main "$@"`
}
