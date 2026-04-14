#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

PLACEHOLDER_LOGO="${REPO_DIR}/frontend/public/favicon.png"
START_JAVA_SCRIPT="${REPO_DIR}/jerinte_1.0.0.2/jerinte_1.0.0.2/base_framework/startJava.sh"

APP_KEY=""
APP_NAME=""
APP_VERSION="1.0.0"
APP_TYPE="app"
APP_OUTPUT_ROOT="${REPO_DIR}/build/generated-local-apps"
APP_JAR=""
APP_SERVER_NAME=""
APP_INTERNAL_PORT="8080"
APP_HOST_PORT="18080"
APP_JAVA_IMAGE="alibaba_dragonwell_jdk_anolis:21"
APP_TIMEZONE="Asia/Shanghai"
APP_JAVA_OPTS="-Xms512m -Xmx512m -XX:+HeapDumpOnOutOfMemoryError -Dfile.encoding=utf-8"
APP_APPLICATION_OPS=""
APP_CONFIG_FILE=""
APP_CONFIG_TARGET=""
APP_LOGO=""
APP_DESC_ZH=""
APP_DESC_EN=""
APP_MEMORY_REQUIRED="512"
ARCHIVE_OUTPUT="false"

WEB_DIR=""
WEB_ROUTE_PREFIX=""
WEB_API_PREFIX="/api/"
WEB_API_PROXY_PASS=""
WEB_ROOT_NAME=""

declare -a APP_TAGS=()
declare -a APP_ENVS=()
declare -a APP_FIXED_ENVS=()

usage() {
    cat <<'EOF'
Usage:
  generate_1panel_java_app.sh \
    --jar /path/to/app.jar \
    --app-key jerinte-foo \
    --app-name "Jerinte Foo" \
    [options]

Required:
  --jar PATH                 Path to the jar file
  --app-key KEY              1Panel app key
  --app-name NAME            1Panel app display name

Optional:
  --version VERSION          App version, default: 1.0.0
  --output-root DIR          Output root, default: build/generated-local-apps
  --server-name NAME         Jar file name inside container, default: app-key
  --internal-port PORT       Java service port inside container, default: 8080
  --host-port PORT           1Panel form default host port, default: 18080
  --java-image IMAGE         Default Java image
  --timezone TZ              Default timezone
  --java-opts OPTS           JAVA_OPS content
  --application-ops OPTS     APPLICATION_OPS content
  --config-file PATH         Extra config file mounted into container
  --config-target PATH       Target mount path for config file, default: /app/<basename>
  --logo PATH                App logo
  --desc-zh TEXT             Chinese description
  --desc-en TEXT             English description
  --tag TAG                  Repeatable app tag
  --env KEY=VALUE            Repeatable editable env var, auto-exposed in formFields
  --fixed-env KEY=VALUE      Repeatable hidden env var, written to .env only
  --memory-required MB       Memory required hint, default: 512
  --archive                  Also create <app-key>-<version>.tar.gz

Web options:
  --web-dir DIR              Optional static frontend directory
  --route-prefix PATH        Route prefix such as /drg, default: /<app-key>
  --api-prefix PATH          Nginx API prefix, default: /api/
  --api-proxy-pass URL       Nginx proxy_pass target, default: direct to api service
  --web-root-name NAME       Mounted web root name inside nginx

Examples:
  Backend only:
    generate_1panel_java_app.sh \
      --jar ./foo.jar \
      --app-key foo-service \
      --app-name "Foo Service" \
      --internal-port 8088 \
      --host-port 18088 \
      --env REDIS_HOST=base-redis \
      --env REDIS_PORT=6379

  Backend + frontend:
    generate_1panel_java_app.sh \
      --jar ./foo.jar \
      --app-key foo-web \
      --app-name "Foo Web" \
      --internal-port 8080 \
      --host-port 18080 \
      --web-dir ./dist \
      --route-prefix /foo \
      --api-prefix /foo/api/ \
      --api-proxy-pass http://foo-web-api:8080/
EOF
}

die() {
    printf '%s\n' "$*" >&2
    exit 1
}

require_file() {
    local file="$1"
    [[ -f "${file}" ]] || die "missing file: ${file}"
}

escape_yaml() {
    printf '%s' "$1" | sed 's/"/\\"/g'
}

humanize_label() {
    local key="$1"
    key="${key//_/ }"
    key="$(printf '%s' "${key}" | tr '[:upper:]' '[:lower:]')"
    awk '{
        for (i=1; i<=NF; i++) {
            $i=toupper(substr($i,1,1)) substr($i,2)
        }
        print
    }' <<<"${key}"
}

quote_compose() {
    local value="$1"
    value="${value//\\/\\\\}"
    value="${value//\"/\\\"}"
    printf '"%s"' "${value}"
}

guess_field_type() {
    local key="$1"
    if [[ "${key}" =~ (PASSWORD|PASSWD|PWD|TOKEN|SECRET|KEY)$ ]]; then
        printf 'password\n'
        return 0
    fi
    if [[ "${key}" =~ PORT ]]; then
        printf 'number\n'
        return 0
    fi
    printf 'text\n'
}

write_dynamic_env_yaml() {
    local item key
    if ((${#APP_ENVS[@]})); then
        for item in "${APP_ENVS[@]}"; do
            key="${item%%=*}"
            printf '      %s: "${%s}"\n' "${key}" "${key}"
        done
    fi
    if ((${#APP_FIXED_ENVS[@]})); then
        for item in "${APP_FIXED_ENVS[@]}"; do
            key="${item%%=*}"
            printf '      %s: "${%s}"\n' "${key}" "${key}"
        done
    fi
}

write_optional_config_mount() {
    if [[ -n "${APP_CONFIG_FILE}" ]]; then
        printf '      - ./config/%s:%s:ro\n' "$(basename "${APP_CONFIG_FILE}")" "${APP_CONFIG_TARGET}"
    fi
}

write_root_data_yml() {
    local app_dir="$1"
    local desc_zh="$2"
    local desc_en="$3"
    cat > "${app_dir}/data.yml" <<EOF
additionalProperties:
  name: ${APP_NAME}
  key: ${APP_KEY}
  type: ${APP_TYPE}
  tags:
$(for tag in "${APP_TAGS[@]}"; do printf '    - %s\n' "${tag}"; done)
  shortDescZh: ${APP_NAME}
  shortDescEn: ${APP_NAME}
  description:
    zh: "$(escape_yaml "${desc_zh}")"
    en: "$(escape_yaml "${desc_en}")"
  crossVersionUpdate: true
  limit: 1
  recommend: 100
  website: ""
  github: ""
  document: ""
  architectures:
    - amd64
    - arm64
  memoryRequired: ${APP_MEMORY_REQUIRED}
  gpuSupport: false
  batchInstallSupport: false
EOF
}

write_version_data_yml() {
    local version_dir="$1"
    cat > "${version_dir}/data.yml" <<EOF
additionalProperties:
  formFields:
    - type: text
      labelZh: Java 镜像
      labelEn: Java image
      required: true
      default: ${APP_JAVA_IMAGE}
      envKey: JAVA_IMAGES
    - type: text
      labelZh: 时区
      labelEn: Timezone
      required: true
      default: ${APP_TIMEZONE}
      envKey: TIMEZONE
    - type: number
      labelZh: 对外端口
      labelEn: Host port
      required: true
      default: ${APP_HOST_PORT}
      envKey: PANEL_APP_PORT_HTTP
      rule: port
EOF

    local item key value field_type label
    if ((${#APP_ENVS[@]})); then
        for item in "${APP_ENVS[@]}"; do
            key="${item%%=*}"
            value="${item#*=}"
            field_type="$(guess_field_type "${key}")"
            label="$(humanize_label "${key}")"
            cat >> "${version_dir}/data.yml" <<EOF
    - type: ${field_type}
      labelZh: ${label}
      labelEn: ${label}
      required: true
      default: $(if [[ "${field_type}" == "number" ]]; then printf '%s' "${value}"; else printf '%s' "${value}"; fi)
      envKey: ${key}
$(if [[ "${field_type}" == "number" ]]; then printf '      rule: port\n'; fi)
EOF
        done
    fi
}

write_env_file() {
    local version_dir="$1"
    {
        printf 'JAVA_IMAGES=%s\n' "${APP_JAVA_IMAGE}"
        printf 'TIMEZONE=%s\n' "${APP_TIMEZONE}"
        printf 'PANEL_APP_PORT_HTTP=%s\n' "${APP_HOST_PORT}"
        if ((${#APP_ENVS[@]})); then
            for item in "${APP_ENVS[@]}"; do
                printf '%s\n' "${item}"
            done
        fi
        if ((${#APP_FIXED_ENVS[@]})); then
            for item in "${APP_FIXED_ENVS[@]}"; do
                printf '%s\n' "${item}"
            done
        fi
    } > "${version_dir}/.env"
}

write_readme() {
    local app_dir="$1"
    cat > "${app_dir}/README.md" <<EOF
# ${APP_NAME}

This package was generated by \`scripts/generate_1panel_java_app.sh\`.

Install steps:

1. Copy this app directory into the 1Panel local app directory.
2. Open 1Panel App Store.
3. Click "Sync Local App".
4. Install ${APP_NAME}.

Container behavior:

- Jar file will run as \`/app/${APP_SERVER_NAME}.jar\`
- Host port form field defaults to \`${APP_HOST_PORT}\`
- Internal service port defaults to \`${APP_INTERNAL_PORT}\`
EOF
}

write_backend_compose() {
    local version_dir="$1"
    local api_service_name="${APP_KEY}"
    if [[ "${api_service_name}" != *-api ]]; then
        api_service_name="${api_service_name}-api"
    fi
    cat > "${version_dir}/docker-compose.yml" <<EOF
services:
  ${api_service_name}:
    image: \${JAVA_IMAGES}
    container_name: ${api_service_name}
    environment:
      PROJECT_NAME: ${APP_KEY}
      SERVER_PORT: "${APP_INTERNAL_PORT}"
      JAVA_OPS: $(quote_compose "${APP_JAVA_OPTS}")
      SERVER_NAME: ${APP_SERVER_NAME}
      APPLICATION_OPS: $(quote_compose "${APP_APPLICATION_OPS}")
      TIMEZONE: \${TIMEZONE}
$(write_dynamic_env_yaml)
    command: /startJava.sh
    ports:
      - \${PANEL_APP_PORT_HTTP}:${APP_INTERNAL_PORT}
    volumes:
      - ./data:/app
      - ./logs:/logs
      - ./startJava.sh:/startJava.sh:ro
$(write_optional_config_mount)
    restart: always

networks:
  default:
    name: 1panel-network
    external: true
EOF
}

write_web_compose() {
    local version_dir="$1"
    local api_service_name="${APP_KEY}"
    local web_service_name="${APP_KEY}"
    if [[ "${api_service_name}" != *-api ]]; then
        api_service_name="${api_service_name}-api"
    fi
    if [[ "${web_service_name}" != *-web ]]; then
        web_service_name="${web_service_name}-web"
    fi
    cat > "${version_dir}/docker-compose.yml" <<EOF
services:
  ${api_service_name}:
    image: \${JAVA_IMAGES}
    container_name: ${api_service_name}
    environment:
      PROJECT_NAME: ${APP_KEY}
      SERVER_PORT: "${APP_INTERNAL_PORT}"
      JAVA_OPS: $(quote_compose "${APP_JAVA_OPTS}")
      SERVER_NAME: ${APP_SERVER_NAME}
      APPLICATION_OPS: $(quote_compose "${APP_APPLICATION_OPS}")
      TIMEZONE: \${TIMEZONE}
$(write_dynamic_env_yaml)
    command: /startJava.sh
    volumes:
      - ./data:/app
      - ./logs:/logs
      - ./startJava.sh:/startJava.sh:ro
$(write_optional_config_mount)
    restart: always

  ${web_service_name}:
    image: nginx:1.27.1
    container_name: ${web_service_name}
    depends_on:
      - ${api_service_name}
    ports:
      - \${PANEL_APP_PORT_HTTP}:80
    volumes:
      - ./web:/usr/share/nginx/${WEB_ROOT_NAME}:ro
      - ./config/nginx/default.conf:/etc/nginx/conf.d/default.conf:ro
    restart: always

networks:
  default:
    name: 1panel-network
    external: true
EOF
}

write_nginx_conf() {
    local version_dir="$1"
    local api_service_name="${APP_KEY}"
    if [[ "${api_service_name}" != *-api ]]; then
        api_service_name="${api_service_name}-api"
    fi
    local route_prefix="${WEB_ROUTE_PREFIX}"
    local api_prefix="${WEB_API_PREFIX}"
    local proxy_pass="${WEB_API_PROXY_PASS}"

    if [[ -z "${proxy_pass}" ]]; then
        proxy_pass="http://${api_service_name}:${APP_INTERNAL_PORT}/"
    fi

    mkdir -p "${version_dir}/config/nginx"
    cat > "${version_dir}/config/nginx/default.conf" <<EOF
server {
    listen 80;
    server_name _;
    client_max_body_size 100m;

    location ${route_prefix} {
        alias /usr/share/nginx/${WEB_ROOT_NAME};
        index index.html;
        try_files \$uri \$uri/ ${route_prefix}/index.html;
    }

    location ^~${api_prefix} {
        proxy_pass ${proxy_pass};
        proxy_connect_timeout 60s;
        proxy_read_timeout 120s;
        proxy_send_timeout 120s;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto http;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host \$http_host;
    }

    location / {
        return 404;
    }
}
EOF
}

create_archive() {
    local app_dir="$1"
    local archive_path="${APP_OUTPUT_ROOT}/${APP_KEY}-${APP_VERSION}.tar.gz"
    tar -C "${APP_OUTPUT_ROOT}" -czf "${archive_path}" "$(basename "${app_dir}")"
    printf 'created archive %s\n' "${archive_path}"
}

while [[ $# -gt 0 ]]; do
    case "$1" in
        --jar) APP_JAR="$2"; shift 2 ;;
        --app-key) APP_KEY="$2"; shift 2 ;;
        --app-name) APP_NAME="$2"; shift 2 ;;
        --version) APP_VERSION="$2"; shift 2 ;;
        --output-root) APP_OUTPUT_ROOT="$2"; shift 2 ;;
        --server-name) APP_SERVER_NAME="$2"; shift 2 ;;
        --internal-port) APP_INTERNAL_PORT="$2"; shift 2 ;;
        --host-port) APP_HOST_PORT="$2"; shift 2 ;;
        --java-image) APP_JAVA_IMAGE="$2"; shift 2 ;;
        --timezone) APP_TIMEZONE="$2"; shift 2 ;;
        --java-opts) APP_JAVA_OPTS="$2"; shift 2 ;;
        --application-ops) APP_APPLICATION_OPS="$2"; shift 2 ;;
        --config-file) APP_CONFIG_FILE="$2"; shift 2 ;;
        --config-target) APP_CONFIG_TARGET="$2"; shift 2 ;;
        --logo) APP_LOGO="$2"; shift 2 ;;
        --desc-zh) APP_DESC_ZH="$2"; shift 2 ;;
        --desc-en) APP_DESC_EN="$2"; shift 2 ;;
        --tag) APP_TAGS+=("$2"); shift 2 ;;
        --env) APP_ENVS+=("$2"); shift 2 ;;
        --fixed-env) APP_FIXED_ENVS+=("$2"); shift 2 ;;
        --memory-required) APP_MEMORY_REQUIRED="$2"; shift 2 ;;
        --web-dir) WEB_DIR="$2"; shift 2 ;;
        --route-prefix) WEB_ROUTE_PREFIX="$2"; shift 2 ;;
        --api-prefix) WEB_API_PREFIX="$2"; shift 2 ;;
        --api-proxy-pass) WEB_API_PROXY_PASS="$2"; shift 2 ;;
        --web-root-name) WEB_ROOT_NAME="$2"; shift 2 ;;
        --archive) ARCHIVE_OUTPUT="true"; shift ;;
        -h|--help) usage; exit 0 ;;
        *) die "unknown argument: $1" ;;
    esac
done

[[ -n "${APP_KEY}" ]] || die "--app-key is required"
[[ -n "${APP_NAME}" ]] || die "--app-name is required"
[[ -n "${APP_JAR}" ]] || die "--jar is required"
require_file "${APP_JAR}"
require_file "${PLACEHOLDER_LOGO}"
require_file "${START_JAVA_SCRIPT}"
if [[ -n "${APP_CONFIG_FILE}" ]]; then
    require_file "${APP_CONFIG_FILE}"
fi
if [[ -n "${WEB_DIR}" ]]; then
    [[ -d "${WEB_DIR}" ]] || die "web dir not found: ${WEB_DIR}"
fi

if [[ -z "${APP_SERVER_NAME}" ]]; then
    APP_SERVER_NAME="${APP_KEY##*/}"
fi
if [[ -z "${APP_LOGO}" ]]; then
    APP_LOGO="${PLACEHOLDER_LOGO}"
fi
if [[ -z "${APP_DESC_ZH}" ]]; then
    APP_DESC_ZH="安装 ${APP_NAME} 本地应用包。"
fi
if [[ -z "${APP_DESC_EN}" ]]; then
    APP_DESC_EN="Install the ${APP_NAME} local app package."
fi
if [[ ${#APP_TAGS[@]} -eq 0 ]]; then
    APP_TAGS=("custom")
fi
if [[ -z "${APP_CONFIG_TARGET}" && -n "${APP_CONFIG_FILE}" ]]; then
    APP_CONFIG_TARGET="/app/$(basename "${APP_CONFIG_FILE}")"
fi
if [[ -n "${WEB_DIR}" && -z "${WEB_ROUTE_PREFIX}" ]]; then
    WEB_ROUTE_PREFIX="/${APP_KEY#jerinte-}"
fi
if [[ -n "${WEB_DIR}" && -z "${WEB_ROOT_NAME}" ]]; then
    WEB_ROOT_NAME="${APP_KEY#jerinte-}-ui"
fi

APP_DIR="${APP_OUTPUT_ROOT}/${APP_KEY}"
VERSION_DIR="${APP_DIR}/${APP_VERSION}"

rm -rf "${APP_DIR}"
mkdir -p "${VERSION_DIR}/data" "${VERSION_DIR}/logs"

cp "${APP_LOGO}" "${APP_DIR}/logo.png"
cp "${START_JAVA_SCRIPT}" "${VERSION_DIR}/startJava.sh"
cp "${APP_JAR}" "${VERSION_DIR}/data/${APP_SERVER_NAME}.jar"
if [[ -n "${APP_CONFIG_FILE}" ]]; then
    mkdir -p "${VERSION_DIR}/config"
    cp "${APP_CONFIG_FILE}" "${VERSION_DIR}/config/$(basename "${APP_CONFIG_FILE}")"
fi
if [[ -n "${WEB_DIR}" ]]; then
    mkdir -p "${VERSION_DIR}/web"
    cp -R "${WEB_DIR}/." "${VERSION_DIR}/web/"
fi

write_root_data_yml "${APP_DIR}" "${APP_DESC_ZH}" "${APP_DESC_EN}"
write_version_data_yml "${VERSION_DIR}"
write_env_file "${VERSION_DIR}"
write_readme "${APP_DIR}"

if [[ -n "${WEB_DIR}" ]]; then
    write_web_compose "${VERSION_DIR}"
    write_nginx_conf "${VERSION_DIR}"
else
    write_backend_compose "${VERSION_DIR}"
fi

printf 'generated local app package in %s\n' "${APP_DIR}"
if [[ "${ARCHIVE_OUTPUT}" == "true" ]]; then
    create_archive "${APP_DIR}"
fi
