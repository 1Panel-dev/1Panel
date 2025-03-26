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

# 检测操作系统类型
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
    OS_LIKE=$(echo $ID_LIKE | awk '{print $1}')  # 获取类似的发行版信息
else
    echo "无法检测操作系统类型"
    exit 1
fi

# 安装防火墙
FIREWALL=""
if [ "$OS" == "ubuntu" ] || [ "$OS" == "debian" ]; then
    echo "检测到 Debian/Ubuntu 系统，正在安装 ufw..."
    FIREWALL="ufw"
    apt-get update
    apt-get install -y ufw
elif [ "$OS" == "centos" ] || [ "$OS" == "rhel" ] || [ "$OS_LIKE" == "rhel" ]; then
    echo "检测到 Red Hat/CentOS 系统，正在安装 firewall..."
    FIREWALL="firewalld"
    yum update
    yum install -y firewalld
else
    echo "不支持的操作系统: $OS"
    exit 1
fi

read -p "请输入需要放行的端口(多个端口用空格分隔，如 80 443 22): " PORTS

# 验证端口输入
if [ -z "$PORTS" ]; then
    echo "错误：未输入任何端口"
    exit 1
fi

case $FIREWALL in
    firewalld)
        echo "配置firewalld..."
        systemctl start firewalld
        systemctl enable firewalld
        
        for port in $PORTS; do
            firewall-cmd --zone=public --permanent --add-port="$port/tcp"
        done
        
        firewall-cmd --reload
        echo "已放行以下TCP端口: $PORTS"
        ;;
        
    ufw)
        echo "配置ufw..."
        ufw --force enable
        
        for port in $PORTS; do
            ufw allow "$port/tcp"
        done
        
        echo "已放行以下TCP端口: $PORTS"
        ;;
esac

# 检查防火墙是否正常运行
if [ "$FIREWALL" = "firewalld" ]; then
    if firewall-cmd --state &> /dev/null; then
        echo "firewalld 安装完成并正常运行！"
    else
        echo "firewalld 安装或配置出现问题，请检查日志"
        exit 1
    fi
else 
    if ufw status | grep -q "Status: active"; then
        echo "ufw 安装完成并正常运行！"
    else
        echo "ufw 安装或配置出现问题，请检查日志"
        exit 1
    fi
fi

exit 0`
}

func loadInstallFTP() string {
	return `#!/bin/bash

# 检测操作系统类型
if [ -f /etc/os-release ]; then
  . /etc/os-release
  OS=$ID
  OS_LIKE=$(echo $ID_LIKE | awk '{print $1}')  # 获取类似的发行版信息
else
  echo "无法检测操作系统类型"
  exit 1
fi

# 安装 Pure-FTPd
if [ "$OS" == "ubuntu" ] || [ "$OS" == "debian" ]; then
  echo "检测到 Debian/Ubuntu 系统，正在安装 Pure-FTPd..."
  apt-get update
  apt-get install -y pure-ftpd
elif [ "$OS" == "centos" ] || [ "$OS" == "rhel" ] || [ "$OS_LIKE" == "rhel" ]; then
  echo "检测到 Red Hat/CentOS 系统，正在安装 Pure-FTPd..."
  yum install -y epel-release
  yum install -y pure-ftpd
else
  echo "不支持的操作系统: $OS"
  exit 1
fi

# 配置 Pure-FTPd
echo "配置 Pure-FTPd..."
PURE_FTPD_CONF="/etc/pure-ftpd/pure-ftpd.conf"
if [ -f "$PURE_FTPD_CONF" ]; then
    cp "$PURE_FTPD_CONF" "$PURE_FTPD_CONF.bak"
    echo "PureDB /etc/pure-ftpd/pureftpd.pdb" >> "$PURE_FTPD_CONF"
    sed -i 's/^# UnixAuthentication/UnixAuthentication/' "$PURE_FTPD_CONF"
    echo "NoAnonymous yes" >> "$PURE_FTPD_CONF"
    echo "PassivePortRange 39000 40000" >> "$PURE_FTPD_CONF"
    echo "ChrootEveryone yes" >> "$PURE_FTPD_CONF"
    echo "VerboseLog yes" >> "$PURE_FTPD_CONF"
else
    echo '/etc/pure-ftpd/pureftpd.pdb' > /etc/pure-ftpd/conf/PureDB
    echo yes > /etc/pure-ftpd/conf/VerboseLog 
    echo yes > /etc/pure-ftpd/conf/NoAnonymous
    echo '39000 40000' > /etc/pure-ftpd/conf/PassivePortRange
    ln -s /etc/pure-ftpd/conf/PureDB /etc/pure-ftpd/auth/50puredb
fi

# 设置开机自启动并启动服务
if command -v systemctl &> /dev/null; then
  echo "设置 Pure-FTPd 开机自启动..."
  systemctl enable pure-ftpd

  echo "启动 Pure-FTPd 服务..."
  systemctl start pure-ftpd

  if systemctl is-active --quiet pure-ftpd; then
    echo "Pure-FTPd 已成功安装并启动"
  else
    echo "Pure-FTPd 启动失败，请检查日志"
    exit 1
  fi
else
  echo "systemctl 不可用，请手动启动 Pure-FTPd"
  exit 1
fi

exit 0`
}

func loadInstallClamAV() string {
	return `#!/bin/bash

# 检测操作系统类型
if [ -f /etc/os-release ]; then
    . /etc/os-release
    OS=$ID
    OS_VER=$VERSION_ID
elif type lsb_release >/dev/null 2>&1; then
    OS=$(lsb_release -si | tr '[:upper:]' '[:lower:]')
    OS_VER=$(lsb_release -sr)
else
    echo "无法识别操作系统"
    exit 1
fi

# 安装ClamAV
install_clamav() {
    case $OS in
        ubuntu|debian)
            apt-get update
            apt-get install -y clamav clamav-daemon
            ;;
        centos|rhel|rocky|almalinux)
            if [[ $OS_VER == 7* ]]; then
                yum install -y epel-release
                yum install -y clamav clamd clamav-update
            else
                dnf install -y clamav clamd clamav-update
            fi
            ;;
        *)
            echo "不支持的OS: $OS"
            exit 1
            ;;
    esac
}

# 配置 clamd
configure_clamd() {
    CLAMD_CONF=""
    if [ -f "/etc/clamd.d/scan.conf" ]; then
        CLAMD_CONF="/etc/clamd.d/scan.conf"
    elif [ -f "/etc/clamav/clamd.conf" ]; then
        CLAMD_CONF="/etc/clamav/clamd.conf"
    else
        echo "未找到 freshclam 配置文件，请手动配置"
        exit 1
    fi

    echo "配置 clamd $CLAMD_CONF..."
    # 备份原始配置文件
    cp "$CLAMD_CONF" "$CLAMD_CONF.bak"

    # 修改配置文件
    sed -i 's|^LogFileMaxSize .*|LogFileMaxSize 2M|' "$CLAMD_CONF"
    sed -i 's|^PidFile .*|PidFile /run/clamd.scan/clamd.pid|' "$CLAMD_CONF"
    sed -i 's|^DatabaseDirectory .*|DatabaseDirectory /var/lib/clamav|' "$CLAMD_CONF"
    sed -i 's|^LocalSocket .*|LocalSocket /run/clamd.scan/clamd.sock|' "$CLAMD_CONF"
}

# 配置 freshclam
configure_freshclam() {
    FRESHCLAM_CONF=""
    if [ -f "/etc/freshclam.conf" ]; then
        FRESHCLAM_CONF="/etc/freshclam.conf"
    elif [ -f "/etc/clamav/freshclam.conf" ]; then
        FRESHCLAM_CONF="/etc/clamav/freshclam.conf"
    else
        echo "未找到 freshclam 配置文件，请手动配置"
        exit 1
    fi

    echo "freshclam.con $FRESHCLAM_CONF..."
    # 备份原始配置文件
    cp "$FRESHCLAM_CONF" "$FRESHCLAM_CONF.bak"

    # 修改配置文件
    sed -i 's|^DatabaseDirectory .*|DatabaseDirectory /var/lib/clamav|' "$FRESHCLAM_CONF"
    sed -i 's|^PidFile .*|PidFile /var/run/freshclam.pid|' "$FRESHCLAM_CONF"
    sed -i 's/DatabaseMirror db.local.clamav.net/DatabaseMirror database.clamav.net/' "$FRESHCLAM_CONF"
    sed -i 's|^Checks .*|Checks 12|' "$FRESHCLAM_CONF"
}

# 服务管理
setup_service() {
    case $OS in
        ubuntu|debian)
            systemctl stop clamav-freshclam
            systemctl start clamav-daemon
            systemctl enable clamav-daemon
            systemctl start clamav-freshclam
            systemctl enable clamav-freshclam
            ;;
        centos|rhel|rocky|almalinux)
            if [[ $OS_VER == 7* ]]; then
                systemctl stop freshclam
                systemctl start clamd@scan
                systemctl enable clamd@scan
                systemctl start freshclam
                systemctl enable freshclam
            else
                systemctl stop clamav-freshclam
                systemctl start clamd@scan
                systemctl enable clamd@scan
                systemctl start clamav-freshclam
                systemctl enable clamav-freshclam
            fi
            ;;
    esac
}

# 主执行流程
echo "正在安装 ClamAV..."
install_clamav

echo -e "配置 clamd..."
configure_clamd

echo -e "配置 freshclam..."
configure_freshclam

echo -e "设置服务..."
setup_service

echo -e "安装完成！"

echo 0`
}

func loadInstallFail2ban() string {
	return `#!/bin/bash

# 检测操作系统类型
if [ -f /etc/os-release ]; then
  . /etc/os-release
  OS=$ID
  OS_LIKE=$(echo $ID_LIKE | awk '{print $1}')  # 获取类似的发行版信息
else
  echo "无法检测操作系统类型"
  exit 1
fi

# 安装 fail2ban
if [ "$OS" == "ubuntu" ] || [ "$OS" == "debian" ]; then
  echo "检测到 Debian/Ubuntu 系统，正在安装 fail2ban..."
  apt-get update
  apt-get install -y fail2ban
elif [ "$OS" == "centos" ] || [ "$OS" == "rhel" ] || [ "$OS_LIKE" == "rhel" ]; then
  echo "检测到 Red Hat/CentOS 系统，正在安装 fail2ban..."
  yum install -y epel-release
  yum install -y fail2ban
else
  echo "不支持的操作系统: $OS"
  exit 1
fi

# 配置 fail2ban
echo "配置 fail2ban..."
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

# 设置开机自启动并启动服务
if command -v systemctl &> /dev/null; then
  echo "设置 fail2ban 开机自启动..."
  systemctl enable fail2ban

  echo "启动 fail2ban 服务..."
  systemctl start fail2ban

  if systemctl is-active --quiet fail2ban; then
    echo "fail2ban 已成功安装并启动"
  else
    echo "fail2ban 启动失败，请检查日志"
    exit 1
  fi
else
  echo "systemctl 不可用，请手动启动 fail2ban"
  exit 1
fi

# 检查 fail2ban 是否正常运行
if fail2ban-client status &> /dev/null; then
  echo "fail2ban 安装完成并正常运行！"
else
  echo "fail2ban 安装或配置出现问题，请检查日志"
  exit 1
fi

exit 0`
}

func loadInstallSupervisor() string {
	return `#!/bin/bash

# 检测操作系统类型
if [ -f /etc/os-release ]; then
  . /etc/os-release
  OS=$ID
  OS_LIKE=$(echo $ID_LIKE | awk '{print $1}')  # 获取类似的发行版信息
else
  echo "无法检测操作系统类型"
  exit 1
fi

# 安装 Supervisor
if [ "$OS" == "ubuntu" ] || [ "$OS" == "debian" ]; then
  echo "检测到 Debian/Ubuntu 系统，正在安装 Supervisor..."
  apt-get update
  apt-get install -y supervisor
elif [ "$OS" == "centos" ] || [ "$OS" == "rhel" ] || [ "$OS_LIKE" == "rhel" ]; then
  echo "检测到 Red Hat/CentOS 系统，正在安装 Supervisor..."
  yum install -y epel-release
  yum install -y supervisor
else
  echo "不支持的操作系统: $OS"
  exit 1
fi

# 创建日志目录
echo "创建日志目录..."
mkdir -p /var/log/supervisor
chmod -R 755 /var/log/supervisor

# 设置开机自启动并启动服务
if command -v systemctl &> /dev/null; then
  echo "设置 Supervisor 开机自启动..."
  systemctl enable supervisord

  echo "启动 Supervisor 服务..."
  systemctl start supervisord

  if systemctl is-active --quiet supervisord; then
    echo "Supervisor 已成功安装并启动"
  else
    echo "Supervisor 启动失败，请检查日志"
    exit 1
  fi
else
  echo "systemctl 不可用，请手动启动 Supervisor"
  exit 1
fi

# 检查 Supervisor 是否正常运行
if supervisorctl status &> /dev/null; then
  echo "Supervisor 安装完成并正常运行！"
else
  echo "Supervisor 安装或配置出现问题，请检查日志"
  exit 1
fi

exit 0`
}
