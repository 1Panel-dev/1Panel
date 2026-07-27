# 1Panel v2 Source Code Setup Guide

## This guide is intended for **WSL** and **Linux** environments only

> **Prerequisites**
> - [1Panel Source Code](https://github.com/1Panel-dev/1Panel)
> - [Golang 1.26.1+](https://go.dev/doc/install)
> - [Node.js 20.2.0+](https://nodejs.org/en/download/)
> - Docker

---

# 1. Install Dependencies

In the `/core` directory (relative to the project root), run the following command to install backend dependencies:
```bash
go mod tidy
```

In the `/agent` directory (relative to the project root), run the following command to install agent dependencies:
```bash
go mod tidy
```

In the `/frontend` directory (relative to the project root), run the following command to install frontend dependencies:
```bash
npm install
```

# 2. Build the Frontend

In the `/frontend` directory, run the following command to build the frontend:
```bash
npm run build:pro
```

# 3. Configure the Backend Environment

Add an `app.yaml` file to the `/opt/1panel/conf` directory and adjust it according to your environment:
```yaml
base:
  install_dir: /opt
  mode: dev
  is_demo: false
  is_offline: false
  is_fxplay: false
  is_enterprise: false
  port: 9999
  username: admin
  password: admin123
  version: v2.0.0

log:
  level: debug
  time_zone: Asia/Shanghai
  log_name: 1Panel-Core
  log_suffix: .log
  max_backup: 10
```

Create the `1pctl` script in the `/usr/local/bin` directory (this script provides environment variables required by 1Panel):
```bash
sudo tee /usr/local/bin/1pctl >/dev/null <<'EOF'
#!/usr/bin/env bash
cat <<'CFG'
BASE_DIR=/opt/1panel
ORIGINAL_PORT=9999
ORIGINAL_VERSION=v2.1.3
ORIGINAL_ENTRANCE=
ORIGINAL_USERNAME=admin
ORIGINAL_PASSWORD=admin123
LANGUAGE=zh-CN
PANEL_EDITION=cn
CFG
exit 0
EOF
sudo chmod +x /usr/local/bin/1pctl
```

Create the required runtime directories and change the ownership of `/etc/1panel` to your current user:
```bash
sudo mkdir -p /opt/1panel/conf /opt/1panel/1panel/{db,log,tmp} /etc/1panel
sudo chmod -R 777 /opt/1panel
sudo chown -R $USER:$USER /etc/1panel
```

# 4. Run the Application

In the `/core` directory, run the following command to start the core service:
```bash
go run ./cmd/server
```

In the `/agent` directory, run the following command to start the agent service:
```bash
go run ./cmd/server
```

# 5. Access the Application

Open your browser and navigate to **localhost:9999/1panel**.

Log in using the username `admin` and password `admin123`.