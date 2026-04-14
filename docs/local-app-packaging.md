# 1Panel 本地 Docker 服务打包说明

1Panel OSS 已经支持“本地应用”模式，不需要先追平专业版。运维可以把应用目录放到 `1panel/resource/apps/local/`，然后在应用商店点击“同步本地应用”完成导入。

## 目录规范

每个应用目录至少需要这些文件：

- `APP_ROOT/data.yml`：应用元数据，定义名称、Key、标签、描述。
- `APP_ROOT/logo.png`：应用图标。
- `APP_ROOT/README.md`：安装说明。
- `APP_ROOT/VERSION/data.yml`：安装表单，定义镜像、端口、密码等参数。
- `APP_ROOT/VERSION/docker-compose.yml`：完整可运行的 Compose。

注意：版本目录下必须是“完整 Compose”，不能只放一个 Compose 片段。

## Jerinte 当前包需要改造的点

Jerinte 原包可以复用，但要按 1Panel 规则改造：

1. `pda/ptrp/drg/medical` 的 `docker-compose-*-web.yml` 只是给共享 Nginx 追加 volume，不是完整服务定义，不能直接作为 1Panel 应用。
2. 原包大量使用 `${BASE_DIR}/...` 绝对路径和 `/opt/...` 软链，1Panel 更适合改为版本目录内的相对路径，比如 `./config/...`、`./data/...`。
3. 业务包如果要独立安装，必须挂到共享网络，例如 `1panel-network`，这样才能访问基础框架里的 `base-mysql`、`base-redis`、`base-register`、`base-gateway`。
4. MySQL 这类组件不能再依赖外部安装脚本隐式注入环境变量，必须在 Compose 里显式声明 `MYSQL_ROOT_PASSWORD` 等参数。
5. `ptrp` 原始 Nginx 片段里有明显路径错误：`drg-ui` / `/drg/index.html`，打包时应修正为 `ptrp-ui` / `/ptrp/index.html`。

## 推荐拆包方式

- `jerinte-base-framework`
  包含 MySQL、Redis、Nacos、UPMS、Auth、Gateway、XXL-Job。
- `jerinte-pda`
  独立包含 PDA API + PDA Web。
- `jerinte-drg`
  独立包含 DRG API + DRG Web。
- `jerinte-ptrp`
  独立包含 PTRP API + PTRP Web。
- `jerinte-medical`
  独立包含 Medical API + Medical Web。

这样可以先安装基础框架，再按业务需要单独安装服务。

## 现成脚手架

仓库已提供脚本：

```bash
./scripts/build_jerinte_local_apps.sh \
  "./jerinte_1.0.0.2/jerinte_1.0.0.2" \
  "./build/jerinte-local-apps"
```

生成后，把 `./build/jerinte-local-apps/*` 复制到目标 1Panel 服务器的本地应用目录：

```bash
cp -R ./build/jerinte-local-apps/* /opt/1panel/resource/apps/local/
```

然后在 1Panel 页面操作：

1. 进入“应用商店”
2. 点击“同步本地应用”
3. 勾选“本地应用”
4. 先安装 `Jerinte 基础框架`
5. 再安装需要的业务应用

## 给一个 Jar 直接生成 1Panel 应用包

仓库还提供了通用脚本：

```bash
./scripts/generate_1panel_java_app.sh \
  --jar ./foo.jar \
  --app-key foo-service \
  --app-name "Foo Service" \
  --version 1.0.0 \
  --internal-port 8088 \
  --host-port 18088 \
  --env REDIS_HOST=base-redis \
  --env REDIS_PORT=6379 \
  --env REDIS_PASSWORD=Password123@redis \
  --archive
```

这个脚本适合“给一个 jar，补少量参数，直接生成 1Panel 本地应用包”的场景。

如果服务还有前端静态资源，可以继续传：

```bash
  --web-dir ./dist \
  --route-prefix /foo \
  --api-prefix /foo/api/ \
  --api-proxy-pass http://foo-service-api:8088/
```

如果服务还需要额外挂载配置文件：

```bash
  --config-file ./application-prod.yml \
  --config-target /app/application-prod.yml \
  --application-ops "--spring.config.additional-location=/app/application-prod.yml"
```

## 通用脚本的边界

“只有 jar”时，脚本可以自动生成后端型应用包；但对于类似 DRG 的复杂服务，光有 jar 仍然不够，需要补充这些信息：

1. 服务内部监听端口
2. 是否带前端静态资源
3. 前端访问路径，比如 `/drg`
4. API 代理路径和目标地址
5. 业务依赖变量，比如 MySQL、Redis、Nacos、HIS 数据库等

所以最现实的标准输入是：

- `jar`
- 可选 `dist`
- 可选应用配置文件
- 一组环境变量默认值

这样已经足够做到“跑一个脚本就生成 1Panel 应用包”。

## 以后如何制作自己的 Docker 服务安装包

以后新服务按这套规则做就行：

1. 以“一个可独立运行的 Compose 项目”为一个 1Panel 应用。
2. 把所有宿主机绝对路径改为相对路径。
3. 依赖基础服务时，统一接入 `1panel-network`。
4. 所有需要运维填写的变量，都放进版本目录 `data.yml` 的 `formFields`。
5. 如果原包需要脚本预处理，放到 `VERSION/scripts/init.sh`，1Panel 安装时会自动执行。
6. Java 类服务尽量统一成：
   - `/app/<server-name>.jar`
   - `/logs`
   - `startJava.sh`
   - `.env` 中维护默认参数
7. 如果前端不是单独的 Nginx 容器，而是依赖基础框架共享 Nginx，那在 1Panel 里最好改造成“业务包自带 web 容器”；否则它不能作为独立应用安装。
