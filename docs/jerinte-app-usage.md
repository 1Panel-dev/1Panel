# Jerinte 与自定义 Java 服务接入 1Panel 使用说明

这份文档专门记录当前仓库里两个打包脚本的用途和用法，避免后续遗忘。

## 目标

当前方案不是改 1Panel 核心，而是把自己的服务整理成 1Panel 可识别的“本地应用包”，再通过 1Panel 自带的“同步本地应用”功能导入安装。

适用两类场景：

1. `Jerinte` 整套服务打包
2. 新增单个 Java 服务或类似 DRG 的业务服务接入

## 相关脚本

### 1. Jerinte 专项脚本

文件：

[`scripts/build_jerinte_local_apps.sh`](/Users/kemengkai/Documents/代码项目/go/1Panel/scripts/build_jerinte_local_apps.sh)

作用：

- 把 `jerinte_1.0.0.2` 原始交付包转换成 1Panel 可识别的本地应用包
- 会生成以下应用：
  - `jerinte-base-framework`
  - `jerinte-pda`
  - `jerinte-drg`
  - `jerinte-ptrp`
  - `jerinte-medical`

使用方式：

```bash
./scripts/build_jerinte_local_apps.sh \
  "./jerinte_1.0.0.2/jerinte_1.0.0.2" \
  "./build/jerinte-local-apps"
```

参数说明：

- 第 1 个参数：Jerinte 原始包主目录
- 第 2 个参数：输出目录
- 第 3 个参数：可选，版本号，默认 `1.0.0.2`

执行结果：

- 输出目录下会生成多个本地应用目录
- 每个目录都符合 1Panel 本地应用规范

### 2. 通用 Java 服务脚本

文件：

[`scripts/generate_1panel_java_app.sh`](/Users/kemengkai/Documents/代码项目/go/1Panel/scripts/generate_1panel_java_app.sh)

作用：

- 给一个 `jar`，生成一个 1Panel 可识别的本地应用包
- 适合以后新接入的单体 Java 服务
- 也适合“jar + dist + 配置文件”的业务服务

最小示例：

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

如果带前端静态资源：

```bash
./scripts/generate_1panel_java_app.sh \
  --jar ./foo.jar \
  --app-key foo-web \
  --app-name "Foo Web" \
  --version 1.0.0 \
  --internal-port 8080 \
  --host-port 18080 \
  --web-dir ./dist \
  --route-prefix /foo \
  --api-prefix /foo/api/ \
  --api-proxy-pass http://base-gateway:9999/foo/ \
  --config-file ./application-prod.yml \
  --config-target /app/application-prod.yml \
  --application-ops "--spring.config.additional-location=/app/application-prod.yml" \
  --env MYSQL_HOST=base-mysql \
  --env MYSQL_PORT=3306 \
  --env MYSQL_USER=root \
  --env MYSQL_PWD=Password123@mysql
```

## 生成后如何导入 1Panel

把生成结果复制到 1Panel 服务器本地应用目录：

```bash
cp -R ./build/jerinte-local-apps/* /opt/1panel/resource/apps/local/
```

或者复制通用脚本生成的单应用目录：

```bash
cp -R ./build/generated-local-apps/foo-service /opt/1panel/resource/apps/local/
```

然后在 1Panel 页面中操作：

1. 进入“应用商店”
2. 如果手里是 `.tar.gz` 应用包，点击“上传本地应用包”
3. 如果应用目录已经在服务器本地应用目录中，点击“同步本地应用”
4. 勾选“本地应用”
5. 安装对应应用

对于 Jerinte：

1. 先安装 `Jerinte 基础框架`
2. 再按需安装 `PDA / DRG / PTRP / Medical`

## 以后新增服务怎么接入

### 场景 A：只有一个 jar

直接用通用脚本。

至少要提供这些信息：

- jar 文件路径
- app key
- app name
- 服务内部端口
- 默认对外端口
- 依赖环境变量，比如 Redis / MySQL / Nacos

### 场景 B：jar + 配置文件

在场景 A 基础上增加：

- `--config-file`
- `--config-target`
- `--application-ops`

适合 Spring Boot 外置配置。

### 场景 C：jar + 前端 dist

在场景 B 基础上增加：

- `--web-dir`
- `--route-prefix`
- `--api-prefix`
- `--api-proxy-pass`

适合类似 DRG/PDA 这种业务服务。

## 重要规则

以后新服务做成 1Panel 应用包时，尽量遵守这些规则：

1. 一个服务对应一个完整 Compose 项目，不要只给 Compose 片段。
2. 不要依赖 `/opt/...` 这种宿主机绝对路径，尽量改为版本目录相对路径。
3. 如果依赖基础框架，统一接入 `1panel-network`。
4. 运维需要填写的变量，要通过 `data.yml` 的表单字段暴露出来。
5. 如果模板配置文件里用了 `${VAR}`，要配合 `scripts/init.sh` 在安装时渲染。
6. 如果业务必须走共享网关或共享 Nginx，优先改造成“业务包自带 Web 容器”，这样更适合独立安装和维护。

## 常见问题

### 1. 为什么不能直接拿原始 docker-compose-files 用？

因为你们原包里很多业务目录只有 Web 的 Compose 片段，只是给共享 Nginx 增加 volume，不是完整可运行服务，1Panel 无法直接识别。

### 2. 为什么 Jerinte 要拆成基础框架和业务包？

因为基础服务和业务服务生命周期不同。拆开之后：

- 基础框架只装一次
- 业务服务可以按需单独装、单独升级、单独停启

### 3. 只有 jar 能不能自动推断全部参数？

不能。脚本可以生成基础骨架，但像：

- 数据库地址
- Redis 密码
- Nacos 地址
- 前端路由
- API 代理规则

这些业务信息还是需要明确提供。

## 建议

后面每新增一个类似 DRG 的服务，最好同时准备一份最小接入信息：

```text
服务名:
jar 文件:
内部端口:
默认外部端口:
是否有前端 dist:
前端访问路径:
API 代理路径:
是否依赖 MySQL:
是否依赖 Redis:
是否依赖 Nacos:
是否依赖基础网关:
是否需要外置配置文件:
```

有了这份信息，基本就能直接用通用脚本生成可识别应用包。
