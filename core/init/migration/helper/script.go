package helper

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/global"
)

func LoadScript() {
	groups := []model.Group{
		{Name: "install", Type: "script", IsDefault: false},
		{Name: "uninstall", Type: "script", IsDefault: false},
		{Name: "docker", Type: "script", IsDefault: false},
		{Name: "firewalld", Type: "script", IsDefault: false},
		{Name: "ufw", Type: "script", IsDefault: false},
		{Name: "supervisor", Type: "script", IsDefault: false},
		{Name: "clamav", Type: "script", IsDefault: false},
		{Name: "ftp", Type: "script", IsDefault: false},
		{Name: "fail2ban", Type: "script", IsDefault: false}}
	_ = global.DB.Create(&groups).Error

	_ = global.DB.Where("is_system = ?", 1).Delete(model.ScriptLibrary{}).Error
	list := []model.ScriptLibrary{
		{Name: "Install Docker", IsSystem: true, Groups: fmt.Sprintf("%v,%v", groups[2].ID, groups[0].ID), Script: "bash <(curl -sSL https://linuxmirrors.cn/docker.sh)"},

		{Name: "Install Firewalld", IsSystem: true, Groups: fmt.Sprintf("%v,%v", groups[3].ID, groups[0].ID), Script: loadInstallFirewalld()},
		{Name: "Uninstall Firewalld", IsSystem: true, Groups: fmt.Sprintf("%v,%v", groups[3].ID, groups[1].ID), Script: loadUninstallFirewalld()},

		{Name: "Install Ufw", IsSystem: true, Groups: fmt.Sprintf("%v,%v", groups[4].ID, groups[0].ID), Script: loadInstallUfw()},
		{Name: "Uninstall Ufw", IsSystem: true, Groups: fmt.Sprintf("%v,%v", groups[4].ID, groups[1].ID), Script: loadUninstallUfw()},

		{Name: "Install Supervisor", IsSystem: true, Groups: fmt.Sprintf("%v,%v", groups[5].ID, groups[0].ID), Script: loadInstallSupervisor()},
		{Name: "Uninstall Supervisor", IsSystem: true, Groups: fmt.Sprintf("%v,%v", groups[2].ID, groups[1].ID), Script: loadUninstallSupervisor()},

		{Name: "Install ClamAV", IsSystem: true, Groups: fmt.Sprintf("%v,%v", groups[6].ID, groups[0].ID), Script: loadInstallClamAV()},
		{Name: "Uninstall ClamAV", IsSystem: true, Groups: fmt.Sprintf("%v,%v", groups[6].ID, groups[1].ID), Script: loadUninstallClamAV()},

		{Name: "Install Pure-FTPd", IsSystem: true, Groups: fmt.Sprintf("%v,%v", groups[7].ID, groups[0].ID), Script: loadInstallFTP()},
		{Name: "Uninstall Pure-FTPd", IsSystem: true, Groups: fmt.Sprintf("%v,%v", groups[7].ID, groups[1].ID), Script: loadUninstallFTP()},

		{Name: "Install Fail2ban", IsSystem: true, Groups: fmt.Sprintf("%v,%v", groups[8].ID, groups[0].ID), Script: loadInstallFail2ban()},
		{Name: "Uninstall Fail2ban", IsSystem: true, Groups: fmt.Sprintf("%v,%v", groups[8].ID, groups[1].ID), Script: loadUninstallFail2ban()},
	}

	_ = global.DB.Create(&list).Error
}

func loadInstallFirewalld() string {
	return `#!/bin/bash

# 检查是否具有 sudo 权限
if [ "$EUID" -ne 0 ]; then
  echo "请使用 sudo 或以 root 用户运行此脚本"
  exit 1
fi

# 检测操作系统类型
if [ -f /etc/os-release ]; then
  . /etc/os-release
  OS=$ID
  OS_LIKE=$(echo $ID_LIKE | awk '{print $1}')  # 获取类似的发行版信息
else
  echo "无法检测操作系统类型"
  exit 1
fi

# 安装 firewalld
if [ "$OS" == "ubuntu" ] || [ "$OS" == "debian" ]; then
  echo "检测到 Debian/Ubuntu 系统，正在安装 firewalld..."
  apt-get update
  apt-get install -y firewalld
elif [ "$OS" == "centos" ] || [ "$OS" == "rhel" ] || [ "$OS_LIKE" == "rhel" ]; then
  echo "检测到 Red Hat/CentOS 系统，正在安装 firewalld..."
  yum install -y firewalld
else
  echo "不支持的操作系统: $OS"
  exit 1
fi

# 设置开机自启动并启动服务
if command -v systemctl &> /dev/null; then
  echo "设置 firewalld 开机自启动..."
  systemctl enable firewalld

  echo "启动 firewalld 服务..."
  systemctl start firewalld

  if systemctl is-active --quiet firewalld; then
    echo "firewalld 已成功安装并启动"
  else
    echo "firewalld 启动失败，请检查日志"
    exit 1
  fi
else
  echo "systemctl 不可用，请手动启动 firewalld"
  exit 1
fi

# 放行 SSH 端口（默认 22）
echo "在 firewalld 中放行 SSH 端口..."
firewall-cmd --zone=public --add-service=ssh --permanent
firewall-cmd --reload

# 检查 SSH 端口是否放行
if firewall-cmd --zone=public --query-service=ssh &> /dev/null; then
  echo "SSH 端口已成功放行"
else
  echo "SSH 端口放行失败，请手动检查"
  exit 1
fi

# 检查 firewalld 是否正常运行
if firewall-cmd --state &> /dev/null; then
  echo "firewalld 安装完成并正常运行！"
else
  echo "firewalld 安装或配置出现问题，请检查日志"
  exit 1
fi

exit 0`
}
func loadUninstallFirewalld() string {
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

# 停止 firewalld 服务
if command -v systemctl &> /dev/null; then
  echo "停止 firewalld 服务..."
  systemctl stop firewalld
  systemctl disable firewalld
elif command -v service &> /dev/null; then
  echo "停止 firewalld 服务..."
  service firewalld stop
  if command -v chkconfig &> /dev/null; then
    chkconfig firewalld off
  fi
else
  echo "无法停止 firewalld 服务，请手动停止"
fi

# 卸载 firewalld
if [ "$OS" == "ubuntu" ] || [ "$OS" == "debian" ]; then
  echo "检测到 Debian/Ubuntu 系统，正在卸载 firewalld..."
  apt-get remove --purge -y firewalld
  apt-get autoremove -y
elif [ "$OS" == "centos" ] || [ "$OS" == "rhel" ] || [ "$OS_LIKE" == "rhel" ]; then
  echo "检测到 Red Hat/CentOS 系统，正在卸载 firewalld..."
  yum remove -y firewalld
  yum autoremove -y
else
  echo "不支持的操作系统: $OS"
  exit 1
fi

# 删除 firewalld 配置文件
echo "删除 firewalld 配置文件..."
rm -rf /etc/firewalld
rm -f /etc/sysconfig/firewalld

# 检查是否卸载成功
if ! command -v firewall-cmd &> /dev/null; then
  echo "firewalld 已成功卸载"
else
  echo "firewalld 卸载失败，请手动检查"
  exit 1
fi

exit 0`
}

func loadInstallUfw() string {
	return `
#!/bin/bash

# 检测操作系统类型
if [ -f /etc/os-release ]; then
  . /etc/os-release
  OS=$ID
else
  echo "无法检测操作系统类型"
  exit 1
fi

# 仅支持 Debian/Ubuntu 系统
if [ "$OS" != "ubuntu" ] && [ "$OS" != "debian" ]; then
  echo "此脚本仅支持 Debian/Ubuntu 系统"
  exit 1
fi

# 安装 ufw
echo "检测到 Debian/Ubuntu 系统，正在安装 ufw..."
apt-get update
apt-get install -y ufw

# 启用 ufw
echo "启用 ufw..."
ufw enable

# 放行 SSH 端口（默认 22）
echo "放行 SSH 端口（22）..."
ufw allow 22/tcp

# 检查 SSH 端口是否放行
if ufw status | grep -q "22/tcp"; then
  echo "SSH 端口已成功放行"
else
  echo "SSH 端口放行失败，请手动检查"
  exit 1
fi

# 检查 ufw 是否正常运行
if ufw status | grep -q "Status: active"; then
  echo "ufw 安装完成并正常运行！"
else
  echo "ufw 安装或配置出现问题，请检查日志"
  exit 1
fi

exit 0`
}
func loadUninstallUfw() string {
	return `#!/bin/bash

# 检测操作系统类型
if [ -f /etc/os-release ]; then
  . /etc/os-release
  OS=$ID
else
  echo "无法检测操作系统类型"
  exit 1
fi

# 仅支持 Debian/Ubuntu 系统
if [ "$OS" != "ubuntu" ] && [ "$OS" != "debian" ]; then
  echo "此脚本仅支持 Debian/Ubuntu 系统"
  exit 1
fi

# 停止并禁用 ufw
echo "停止并禁用 ufw..."
ufw disable

# 卸载 ufw
echo "卸载 ufw..."
apt-get remove --purge -y ufw
apt-get autoremove -y

# 删除 ufw 配置文件
echo "删除 ufw 配置文件..."
rm -rf /etc/ufw
rm -f /etc/default/ufw

# 检查是否卸载成功
if ! command -v ufw &> /dev/null; then
  echo "ufw 已成功卸载"
else
  echo "ufw 卸载失败，请手动检查"
  exit 1
fi

exit 0`
}

func loadInstallFTP() string {
	return `#!/bin/bash

# 检查是否具有 sudo 权限
if [ "$EUID" -ne 0 ]; then
  echo "请使用 sudo 或以 root 用户运行此脚本"
  exit 1
fi

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
  # 备份原始配置文件
  cp "$PURE_FTPD_CONF" "$PURE_FTPD_CONF.bak"

  # 启用本地用户登录
  sed -i 's/^# UnixAuthentication/UnixAuthentication/' "$PURE_FTPD_CONF"

  # 拒绝匿名登录
  echo "NoAnonymous yes" >> "$PURE_FTPD_CONF"

  # 启用被动模式
  echo "PassivePortRange 30000 30100" >> "$PURE_FTPD_CONF"

  # 限制用户访问家目录
  echo "ChrootEveryone yes" >> "$PURE_FTPD_CONF"

  # 启用日志记录
  echo "VerboseLog yes" >> "$PURE_FTPD_CONF"
else
  echo "未找到 Pure-FTPd 配置文件，请手动配置"
  exit 1
fi

# 创建必要的目录和文件
echo "创建必要的目录和文件..."
mkdir -p /run/pure-ftpd
chown pure-ftpd:pure-ftpd /run/pure-ftpd
mkdir -p /var/log/pure-ftpd
chown pure-ftpd:pure-ftpd /var/log/pure-ftpd
touch /var/log/pure-ftpd/pure-ftpd.log
chown pure-ftpd:pure-ftpd /var/log/pure-ftpd/pure-ftpd.log

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

# 检查 Pure-FTPd 是否正常运行
if pure-ftpd --version &> /dev/null; then
  echo "Pure-FTPd 安装完成并正常运行！"
else
  echo "Pure-FTPd 安装或配置出现问题，请检查日志"
  exit 1
fi

exit 0`
}
func loadUninstallFTP() string {
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

# 停止 Pure-FTPd 服务
if command -v systemctl &> /dev/null; then
  echo "停止 Pure-FTPd 服务..."
  systemctl stop pure-ftpd
  systemctl disable pure-ftpd
elif command -v service &> /dev/null; then
  echo "停止 Pure-FTPd 服务..."
  service pure-ftpd stop
  if command -v chkconfig &> /dev/null; then
    chkconfig pure-ftpd off
  fi
else
  echo "无法停止 Pure-FTPd 服务，请手动停止"
fi

# 卸载 Pure-FTPd
if [ "$OS" == "ubuntu" ] || [ "$OS" == "debian" ]; then
  echo "检测到 Debian/Ubuntu 系统，正在卸载 Pure-FTPd..."
  apt-get remove --purge -y pure-ftpd
  apt-get autoremove -y
elif [ "$OS" == "centos" ] || [ "$OS" == "rhel" ] || [ "$OS_LIKE" == "rhel" ]; then
  echo "检测到 Red Hat/CentOS 系统，正在卸载 Pure-FTPd..."
  yum remove -y pure-ftpd
  yum autoremove -y
else
  echo "不支持的操作系统: $OS"
  exit 1
fi

# 删除 Pure-FTPd 配置文件、日志文件和用户数据库
echo "删除 Pure-FTPd 配置文件、日志文件和用户数据库..."
rm -rf /etc/pure-ftpd
rm -rf /var/log/pure-ftpd
rm -rf /var/lib/pure-ftpd

# 检查是否卸载成功
if ! command -v pure-ftpd &> /dev/null; then
  echo "Pure-FTPd 已成功卸载"
else
  echo "Pure-FTPd 卸载失败，请手动检查"
  exit 1
fi

exit 0`
}

func loadInstallClamAV() string {
	return `#!/bin/bash

# 检查是否具有 sudo 权限
if [ "$EUID" -ne 0 ]; then
  echo "请使用 sudo 或以 root 用户运行此脚本"
  exit 1
fi

# 检测操作系统类型
if [ -f /etc/os-release ]; then
  . /etc/os-release
  OS=$ID
  OS_LIKE=$(echo $ID_LIKE | awk '{print $1}')  # 获取类似的发行版信息
else
  echo "无法检测操作系统类型"
  exit 1
fi

# 安装 ClamAV
if [ "$OS" == "ubuntu" ] || [ "$OS" == "debian" ]; then
  echo "检测到 Debian/Ubuntu 系统，正在安装 ClamAV..."
  apt-get update
  apt-get install -y clamav clamav-daemon clamav-freshclam
elif [ "$OS" == "centos" ] || [ "$OS" == "rhel" ] || [ "$OS_LIKE" == "rhel" ]; then
  echo "检测到 Red Hat/CentOS 系统，正在安装 ClamAV..."
  yum install -y epel-release
  yum install -y clamav clamd clamav-update
else
  echo "不支持的操作系统: $OS"
  exit 1
fi

# 配置 clamd
CLAMD_CONF="/etc/clamd.d/scan.conf"
if [ -f "$CLAMD_CONF" ]; then
  echo "配置 clamd..."
  # 备份原始配置文件
  cp "$CLAMD_CONF" "$CLAMD_CONF.bak"

  # 修改配置文件
  sed -i 's|^#LogFile .*|LogFile /var/log/clamd.scan|' "$CLAMD_CONF"
  sed -i 's|^#LogFileMaxSize .*|LogFileMaxSize 2M|' "$CLAMD_CONF"
  sed -i 's|^#PidFile .*|PidFile /run/clamd.scan/clamd.pid|' "$CLAMD_CONF"
  sed -i 's|^#DatabaseDirectory .*|DatabaseDirectory /var/lib/clamav|' "$CLAMD_CONF"
  sed -i 's|^#LocalSocket .*|LocalSocket /run/clamd.scan/clamd.sock|' "$CLAMD_CONF"
else
  echo "未找到 clamd 配置文件，请手动配置"
  exit 1
fi

# 配置 freshclam
FRESHCLAM_CONF="/etc/freshclam.conf"
if [ -f "$FRESHCLAM_CONF" ]; then
  echo "配置 freshclam..."
  # 备份原始配置文件
  cp "$FRESHCLAM_CONF" "$FRESHCLAM_CONF.bak"

  # 修改配置文件
  sed -i 's|^#DatabaseDirectory .*|DatabaseDirectory /var/lib/clamav|' "$FRESHCLAM_CONF"
  sed -i 's|^#UpdateLogFile .*|UpdateLogFile /var/log/freshclam.log|' "$FRESHCLAM_CONF"
  sed -i 's|^#PidFile .*|PidFile /var/run/freshclam.pid|' "$FRESHCLAM_CONF"
  sed -i 's|^#DatabaseMirror .*|DatabaseMirror database.clamav.net|' "$FRESHCLAM_CONF"
  sed -i 's|^#Checks .*|Checks 12|' "$FRESHCLAM_CONF"
else
  echo "未找到 freshclam 配置文件，请手动配置"
  exit 1
fi

# 创建必要的目录和文件
echo "创建必要的目录和文件..."
mkdir -p /run/clamd.scan
chown clamav:clamav /run/clamd.scan
mkdir -p /var/log/clamav
chown clamav:clamav /var/log/clamav
touch /var/log/clamd.scan /var/log/freshclam.log
chown clamav:clamav /var/log/clamd.scan /var/log/freshclam.log

# 设置开机自启动并启动服务
if command -v systemctl &> /dev/null; then
  echo "设置 ClamAV 开机自启动..."
  systemctl enable clamav-daemon
  systemctl enable clamav-freshclam

  echo "启动 ClamAV 服务..."
  systemctl start clamav-daemon
  systemctl start clamav-freshclam

  if systemctl is-active --quiet clamav-daemon && systemctl is-active --quiet clamav-freshclam; then
    echo "ClamAV 已成功安装并启动"
  else
    echo "ClamAV 启动失败，请检查日志"
    exit 1
  fi
else
  echo "systemctl 不可用，请手动启动 ClamAV"
  exit 1
fi

# 更新病毒数据库
echo "更新 ClamAV 病毒数据库..."
freshclam

# 检查 ClamAV 是否正常运行
if clamscan --version &> /dev/null; then
  echo "ClamAV 安装完成并正常运行！"
else
  echo "ClamAV 安装或配置出现问题，请检查日志"
  exit 1
fi

exit 0`
}
func loadUninstallClamAV() string {
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

# 停止 ClamAV 服务
if command -v systemctl &> /dev/null; then
  echo "停止 ClamAV 服务..."
  if [ "$OS" == "ubuntu" ] || [ "$OS" == "debian" ]; then
    systemctl stop clamav-daemon
    systemctl disable clamav-daemon
  elif [ "$OS" == "centos" ] || [ "$OS" == "rhel" ] || [ "$OS_LIKE" == "rhel" ]; then
    systemctl stop clamd@scan
    systemctl disable clamd@scan
  fi
elif command -v service &> /dev/null; then
  echo "停止 ClamAV 服务..."
  service clamav-daemon stop
  if command -v chkconfig &> /dev/null; then
    chkconfig clamav-daemon off
  fi
else
  echo "无法停止 ClamAV 服务，请手动停止"
fi

# 卸载 ClamAV
if [ "$OS" == "ubuntu" ] || [ "$OS" == "debian" ]; then
  echo "检测到 Debian/Ubuntu 系统，正在卸载 ClamAV..."
  apt-get remove --purge -y clamav clamav-daemon
  apt-get autoremove -y
elif [ "$OS" == "centos" ] || [ "$OS" == "rhel" ] || [ "$OS_LIKE" == "rhel" ]; then
  echo "检测到 Red Hat/CentOS 系统，正在卸载 ClamAV..."
  yum remove -y clamav clamd clamav-update
  yum autoremove -y
else
  echo "不支持的操作系统: $OS"
  exit 1
fi

# 删除 ClamAV 配置文件、病毒数据库和日志文件
echo "删除 ClamAV 配置文件、病毒数据库和日志文件..."
rm -rf /etc/clamav
rm -rf /var/lib/clamav
rm -rf /var/log/clamav
rm -rf /var/run/clamav
rm -f /etc/cron.daily/freshclam
rm -f /etc/logrotate.d/clamav

# 检查是否卸载成功
if ! command -v clamscan &> /dev/null; then
  echo "ClamAV 已成功卸载"
else
  echo "ClamAV 卸载失败，请手动检查"
  exit 1
fi

exit 0`
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
if [ ! -f "$FAIL2BAN_CONF" ]; then
  echo "创建自定义配置文件 $FAIL2BAN_CONF..."
  cat <<EOF > "$FAIL2BAN_CONF"
#DEFAULT-START
[DEFAULT]
bantime = 600
findtime = 300
maxretry = 5
banaction = $banaction
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
banaction = $banaction
action = %(action_mwl)s
logpath = $logpath
EOF
else
  echo "自定义配置文件已存在，跳过创建"
fi

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
func loadUninstallFail2ban() string {
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

# 停止 fail2ban 服务
if command -v systemctl &> /dev/null; then
  echo "停止 fail2ban 服务..."
  systemctl stop fail2ban
  systemctl disable fail2ban
elif command -v service &> /dev/null; then
  echo "停止 fail2ban 服务..."
  service fail2ban stop
  if command -v chkconfig &> /dev/null; then
    chkconfig fail2ban off
  fi
else
  echo "无法停止 fail2ban 服务，请手动停止"
fi

# 卸载 fail2ban
if [ "$OS" == "ubuntu" ] || [ "$OS" == "debian" ]; then
  echo "检测到 Debian/Ubuntu 系统，正在卸载 fail2ban..."
  apt-get remove --purge -y fail2ban
  apt-get autoremove -y
elif [ "$OS" == "centos" ] || [ "$OS" == "rhel" ] || [ "$OS_LIKE" == "rhel" ]; then
  echo "检测到 Red Hat/CentOS 系统，正在卸载 fail2ban..."
  yum remove -y fail2ban
  yum autoremove -y
else
  echo "不支持的操作系统: $OS"
  exit 1
fi

# 删除 fail2ban 配置文件和数据
echo "删除 fail2ban 配置文件和数据..."
rm -rf /etc/fail2ban
rm -rf /var/lib/fail2ban
rm -rf /var/log/fail2ban*

# 检查是否卸载成功
if ! command -v fail2ban-client &> /dev/null; then
  echo "fail2ban 已成功卸载"
else
  echo "fail2ban 卸载失败，请手动检查"
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
func loadUninstallSupervisor() string {
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

# 停止 Supervisor 服务
if command -v systemctl &> /dev/null; then
  echo "停止 Supervisor 服务..."
  systemctl stop supervisord
  systemctl disable supervisord
elif command -v service &> /dev/null; then
  echo "停止 Supervisor 服务..."
  service supervisord stop
  if command -v chkconfig &> /dev/null; then
    chkconfig supervisord off
  fi
else
  echo "无法停止 Supervisor 服务，请手动停止"
fi

# 卸载 Supervisor
if [ "$OS" == "ubuntu" ] || [ "$OS" == "debian" ]; then
  echo "检测到 Debian/Ubuntu 系统，正在卸载 Supervisor..."
  apt-get remove --purge -y supervisor
  apt-get autoremove -y
elif [ "$OS" == "centos" ] || [ "$OS" == "rhel" ] || [ "$OS_LIKE" == "rhel" ]; then
  echo "检测到 Red Hat/CentOS 系统，正在卸载 Supervisor..."
  yum remove -y supervisor
  yum autoremove -y
else
  echo "不支持的操作系统: $OS"
  exit 1
fi

# 删除 Supervisor 配置文件、日志文件和进程管理文件
echo "删除 Supervisor 配置文件、日志文件和进程管理文件..."
rm -rf /etc/supervisor
rm -rf /var/log/supervisor
rm -rf /var/run/supervisor
rm -f /etc/default/supervisor
rm -f /etc/init.d/supervisor

# 检查是否卸载成功
if ! command -v supervisorctl &> /dev/null; then
  echo "Supervisor 已成功卸载"
else
  echo "Supervisor 卸载失败，请手动检查"
  exit 1
fi

exit 0`
}
