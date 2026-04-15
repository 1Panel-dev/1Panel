# 1Panel 服务编译与安装脚本说明

仓库已提供脚本：

[`scripts/build_install_1panel_service.sh`](/Users/kemengkai/Documents/代码项目/go/1Panel/scripts/build_install_1panel_service.sh)

这个脚本用于在 Linux 机器上完成：

1. 前端编译
2. `1panel-core` 编译
3. `1panel-agent` 编译
4. 二进制安装到 `/usr/local/bin`
5. 生成 `/opt/1panel/conf/app.yaml`
6. 生成最小可用 `/usr/local/bin/1pctl`
7. 生成并启用 `systemd` 服务

## 默认安装方式

```bash
sudo ./scripts/build_install_1panel_service.sh
```

默认行为：

- 前端会执行 `npm install && npm run build:pro`
- Go 二进制会按当前机器架构编译为 Linux 版本
- 安装目录记录为 `/opt`
- Core 服务端口默认 `9999`
- Agent 配置端口默认 `9998`
- 配置文件写入 `/opt/1panel/conf/app.yaml`

## 常用参数

```bash
sudo ./scripts/build_install_1panel_service.sh \
  --panel-port 8080 \
  --username admin \
  --password 'StrongPass123'
```

可选参数：

- `--install-dir DIR`
- `--panel-port PORT`
- `--agent-port PORT`
- `--bind-address ADDR`
- `--username NAME`
- `--password PASS`
- `--language LANG`
- `--edition cn|intl`
- `--entrance PATH`
- `--ssl disable|Enable|Mux`
- `--skip-frontend`
- `--skip-npm-install`
- `--skip-enable`
- `--skip-start`

## 典型场景

### 1. 已经装过前端依赖，只想重新编译并安装

```bash
sudo ./scripts/build_install_1panel_service.sh --skip-npm-install
```

### 2. 只想安装 Go 服务，不重新打前端

```bash
sudo ./scripts/build_install_1panel_service.sh --skip-frontend
```

### 3. 只生成服务文件，不立刻启动

```bash
sudo ./scripts/build_install_1panel_service.sh --skip-start
```

## 输出位置

- Core 二进制：`/usr/local/bin/1panel-core`
- Agent 二进制：`/usr/local/bin/1panel-agent`
- 兼容脚本：`/usr/local/bin/1pctl`
- 配置文件：`/opt/1panel/conf/app.yaml`
- systemd：
  - `/etc/systemd/system/1panel-core.service`
  - `/etc/systemd/system/1panel-agent.service`

## 注意事项

1. 这个脚本默认面向 `systemd` 机器。
2. 当前代码在 `dev` 模式下会强依赖 `/opt/1panel/conf/app.yaml`，所以脚本固定写到这个位置。
3. `agent` 社区模式下会从 `/usr/local/bin/1pctl` 读取 `BASE_DIR` 等参数，所以脚本会顺手生成一个最小可用的 `1pctl`。
4. 如果你的服务器没有 Node.js / npm，就不能执行前端编译步骤，需要先准备前端构建环境，或者使用 `--skip-frontend`。
5. 如果你只是开发调试，建议先在测试机执行，不要直接在生产机上覆盖原有服务。
