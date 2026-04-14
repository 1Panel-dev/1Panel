#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

SOURCE_DIR="${1:-${REPO_DIR}/jerinte_1.0.0.2/jerinte_1.0.0.2}"
OUTPUT_DIR="${2:-${REPO_DIR}/build/jerinte-local-apps}"
VERSION="${3:-1.0.0.2}"

PLACEHOLDER_LOGO="${REPO_DIR}/frontend/public/favicon.png"

die() {
    printf '%s\n' "$*" >&2
    exit 1
}

require_file() {
    local path="$1"
    [[ -f "${path}" ]] || die "missing file: ${path}"
}

read_kv() {
    local key="$1"
    local file="$2"
    awk -F= -v k="${key}" '$1 == k { print substr($0, index($0, "=") + 1) }' "${file}" | tail -n 1
}

replace_line() {
    local key="$1"
    local value="$2"
    local file="$3"
    local tmp
    tmp="$(mktemp)"
    awk -v k="${key}" -v v="${value}" '
        BEGIN { done = 0 }
        $0 ~ "^" k "=" {
            print k "=" v
            done = 1
            next
        }
        { print }
        END {
            if (!done) {
                print k "=" v
            }
        }
    ' "${file}" > "${tmp}"
    mv "${tmp}" "${file}"
}

copy_logo() {
    local app_dir="$1"
    cp "${PLACEHOLDER_LOGO}" "${app_dir}/logo.png"
}

write_root_data_yml() {
    local app_dir="$1"
    local app_name="$2"
    local app_key="$3"
    local short_desc="$4"
    local zh_desc="$5"
    local en_desc="$6"
    local memory_required="$7"

    cat > "${app_dir}/data.yml" <<EOF
additionalProperties:
  name: ${app_name}
  key: ${app_key}
  type: app
  tags:
    - jerinte
  shortDescZh: ${short_desc}
  shortDescEn: ${short_desc}
  description:
    zh: ${zh_desc}
    en: ${en_desc}
  crossVersionUpdate: true
  limit: 1
  recommend: 100
  website: ""
  github: ""
  document: ""
  architectures:
    - amd64
    - arm64
  memoryRequired: ${memory_required}
  gpuSupport: false
  batchInstallSupport: false
EOF
}

write_root_readme() {
    local app_dir="$1"
    local title="$2"
    local body="$3"

    cat > "${app_dir}/README.md" <<EOF
# ${title}

${body}
EOF
}

write_base_version_data_yml() {
    local version_dir="$1"
    local java_image="$2"
    local mysql_image="$3"
    local mysql_pwd="$4"
    local redis_pwd="$5"
    local docker_subnet="$6"

    cat > "${version_dir}/data.yml" <<EOF
additionalProperties:
  formFields:
    - type: text
      labelZh: Java 镜像
      labelEn: Java image
      required: true
      default: ${java_image}
      envKey: JAVA_IMAGES
    - type: text
      labelZh: MySQL 镜像
      labelEn: MySQL image
      required: true
      default: ${mysql_image}
      envKey: MYSQL_IMAGES
    - type: text
      labelZh: 时区
      labelEn: Timezone
      required: true
      default: Asia/Shanghai
      envKey: TIMEZONE
    - type: text
      labelZh: Docker 网段
      labelEn: Docker subnet
      required: true
      default: ${docker_subnet}
      envKey: DOCKER_SUBNET
    - type: number
      labelZh: MySQL 对外端口
      labelEn: MySQL host port
      required: true
      default: 3306
      envKey: MYSQL_PORT
      rule: port
    - type: password
      labelZh: MySQL Root 密码
      labelEn: MySQL root password
      required: true
      default: ${mysql_pwd}
      envKey: MYSQL_PWD
    - type: text
      labelZh: MySQL 用户名
      labelEn: MySQL username
      required: true
      default: root
      envKey: MYSQL_USER
    - type: number
      labelZh: Redis 对外端口
      labelEn: Redis host port
      required: true
      default: 6379
      envKey: REDIS_PORT
      rule: port
    - type: password
      labelZh: Redis 密码
      labelEn: Redis password
      required: true
      default: ${redis_pwd}
      envKey: REDIS_PASSWORD
    - type: number
      labelZh: Nacos HTTP 端口
      labelEn: Nacos HTTP port
      required: true
      default: 8848
      envKey: REGISTER_PORT
      rule: port
    - type: number
      labelZh: Nacos gRPC 端口
      labelEn: Nacos gRPC port
      required: true
      default: 9848
      envKey: REGISTER_RAFT_PORT
      rule: port
    - type: number
      labelZh: Nacos 集群端口
      labelEn: Nacos peer port
      required: true
      default: 9849
      envKey: REGISTER_PEER_PORT
      rule: port
    - type: number
      labelZh: UPMS 对外端口
      labelEn: UPMS host port
      required: true
      default: 4000
      envKey: UPMS_PORT
      rule: port
    - type: number
      labelZh: Auth 对外端口
      labelEn: Auth host port
      required: true
      default: 3000
      envKey: AUTH_PORT
      rule: port
    - type: number
      labelZh: Gateway 对外端口
      labelEn: Gateway host port
      required: true
      default: 9999
      envKey: GATEWAY_PORT
      rule: port
    - type: number
      labelZh: XXL-Job 对外端口
      labelEn: XXL-Job host port
      required: true
      default: 9080
      envKey: XXL_PORT
      rule: port
EOF
}

write_business_version_data_yml() {
    local version_dir="$1"
    local java_image="$2"
    local http_port="$3"
    local extra_fields="$4"

    cat > "${version_dir}/data.yml" <<EOF
additionalProperties:
  formFields:
    - type: text
      labelZh: Java 镜像
      labelEn: Java image
      required: true
      default: ${java_image}
      envKey: JAVA_IMAGES
    - type: text
      labelZh: 时区
      labelEn: Timezone
      required: true
      default: Asia/Shanghai
      envKey: TIMEZONE
    - type: number
      labelZh: Web 访问端口
      labelEn: Web host port
      required: true
      default: ${http_port}
      envKey: PANEL_APP_PORT_HTTP
      rule: port
${extra_fields}
EOF
}

normalize_base_conf() {
    local file="$1"
    replace_line "BASE_DIR" "/workdir" "${file}"
    replace_line "BASE_PATH" "/app" "${file}"
    replace_line "JAVA_BASE_PATH" "/app" "${file}"
    replace_line "DOCKER_COMPOSE_PATH" "/workdir" "${file}"
    replace_line "TIMEZONE" "\${TIMEZONE}" "${file}"
    replace_line "MYSQL_HOST" "\${MYSQL_HOST}" "${file}"
    replace_line "MYSQL_PORT" "\${MYSQL_PORT}" "${file}"
    replace_line "MYSQL_USER" "\${MYSQL_USER}" "${file}"
    replace_line "MYSQL_PWD" "\${MYSQL_PWD}" "${file}"
    replace_line "MYSQL_ROOT_PASSWORD" "\${MYSQL_PWD}" "${file}"
    replace_line "REDIS_HOST" "\${REDIS_HOST}" "${file}"
    replace_line "REDIS_PORT" "\${REDIS_PORT}" "${file}"
    replace_line "REDIS_PASSWORD" "\${REDIS_PASSWORD}" "${file}"
    replace_line "NACOS_HOST" "\${NACOS_HOST}" "${file}"
    replace_line "NACOS_PORT" "\${NACOS_PORT}" "${file}"
}

normalize_business_conf() {
    local file="$1"
    replace_line "BASE_DIR" "/workdir" "${file}"
    replace_line "BASE_PATH" "/app" "${file}"
    replace_line "JAVA_BASE_PATH" "/app" "${file}"
    replace_line "DOCKER_COMPOSE_PATH" "/workdir" "${file}"
    replace_line "TIMEZONE" "\${TIMEZONE}" "${file}"
    replace_line "REDIS_HOST" "\${REDIS_HOST}" "${file}"
    replace_line "REDIS_PORT" "\${REDIS_PORT}" "${file}"
    replace_line "REDIS_PASSWORD" "\${REDIS_PASSWORD}" "${file}"
    replace_line "NACOS_HOST" "\${NACOS_HOST}" "${file}"
    replace_line "NACOS_PORT" "\${NACOS_PORT}" "${file}"
    replace_line "MYSQL_HOST" "\${MYSQL_HOST}" "${file}"
    replace_line "MYSQL_PORT" "\${MYSQL_PORT}" "${file}"
    replace_line "MYSQL_USER" "\${MYSQL_USER}" "${file}"
    replace_line "MYSQL_PWD" "\${MYSQL_PWD}" "${file}"
    replace_line "HIS_DB_SCHEMA" "\${HIS_DB_SCHEMA}" "${file}"
    replace_line "HIS_DB_HOST" "\${HIS_DB_HOST}" "${file}"
    replace_line "HIS_DB_PORT" "\${HIS_DB_PORT}" "${file}"
    replace_line "HIS_DB_SERVER_NAME" "\${HIS_DB_SERVER_NAME}" "${file}"
    replace_line "HIS_DB_USER" "\${HIS_DB_USER}" "${file}"
    replace_line "HIS_DB_PWD" "\${HIS_DB_PWD}" "${file}"
    replace_line "HIS_DB_PASSWORD" "\${HIS_DB_PWD}" "${file}"
    replace_line "PTRP_DB_HOST" "\${PTRP_DB_HOST}" "${file}"
    replace_line "PTRP_DB_PORT" "\${PTRP_DB_PORT}" "${file}"
    replace_line "PTRP_DB_USER" "\${PTRP_DB_USER}" "${file}"
    replace_line "PTRP_DB_PWD" "\${PTRP_DB_PWD}" "${file}"
}

mark_templates() {
    local version_dir="$1"
    shift
    for conf in "$@"; do
        if [[ -f "${version_dir}/config/${conf}" ]]; then
            mv "${version_dir}/config/${conf}" "${version_dir}/config/${conf}.tmpl"
        fi
    done
}

write_template_init_script() {
    local version_dir="$1"
    mkdir -p "${version_dir}/scripts"
    cat > "${version_dir}/scripts/init.sh" <<'EOF'
#!/usr/bin/env bash

set -euo pipefail

WORKDIR="$(cd "$(dirname "$0")/.." && pwd)"
ENV_FILE="${WORKDIR}/.env"

if [[ ! -f "${ENV_FILE}" ]]; then
    exit 0
fi

render_template() {
    local input="$1"
    local output="$2"
    cp "${input}" "${output}"
    while IFS='=' read -r key value; do
        if [[ -z "${key}" || "${key}" =~ ^# ]]; then
            continue
        fi
        local escaped="${value//\\/\\\\}"
        escaped="${escaped//&/\\&}"
        escaped="${escaped//|/\\|}"
        sed -i.bak "s|\${${key}}|${escaped}|g" "${output}"
        rm -f "${output}.bak"
    done < <(grep -E '^[A-Za-z_][A-Za-z0-9_]*=' "${ENV_FILE}")
}

for template in "${WORKDIR}"/config/*.tmpl; do
    if [[ -f "${template}" ]]; then
        render_template "${template}" "${template%.tmpl}"
    fi
done
EOF
    chmod +x "${version_dir}/scripts/init.sh"
}

write_base_env() {
    local version_dir="$1"
    local java_image="$2"
    local mysql_image="$3"
    local mysql_pwd="$4"
    local redis_pwd="$5"
    local docker_subnet="$6"

    cat > "${version_dir}/.env" <<EOF
JAVA_IMAGES=${java_image}
MYSQL_IMAGES=${mysql_image}
TIMEZONE=Asia/Shanghai
DOCKER_SUBNET=${docker_subnet}
MYSQL_PORT=3306
MYSQL_HOST=base-mysql
MYSQL_USER=root
MYSQL_PWD=${mysql_pwd}
REDIS_PORT=6379
REDIS_HOST=base-redis
REDIS_PASSWORD=${redis_pwd}
NACOS_HOST=base-register
NACOS_PORT=8848
REGISTER_PORT=8848
REGISTER_RAFT_PORT=9848
REGISTER_PEER_PORT=9849
UPMS_PORT=4000
AUTH_PORT=3000
GATEWAY_PORT=9999
XXL_PORT=9080
EOF
}

write_business_env() {
    local version_dir="$1"
    local java_image="$2"
    local default_http_port="$3"
    local extra_env="$4"

    cat > "${version_dir}/.env" <<EOF
JAVA_IMAGES=${java_image}
TIMEZONE=Asia/Shanghai
PANEL_APP_PORT_HTTP=${default_http_port}
REDIS_HOST=base-redis
REDIS_PORT=6379
REDIS_PASSWORD=${REDIS_PASSWORD_DEFAULT}
NACOS_HOST=base-register
NACOS_PORT=8848
MYSQL_HOST=base-mysql
MYSQL_PORT=3306
MYSQL_USER=root
MYSQL_PWD=${MYSQL_PASSWORD_DEFAULT}
${extra_env}
EOF
}

write_base_compose() {
    local version_dir="$1"

    cat > "${version_dir}/docker-compose.yml" <<'EOF'
services:
  base-mysql:
    image: ${MYSQL_IMAGES}
    container_name: base-mysql
    environment:
      MYSQL_ROOT_PASSWORD: ${MYSQL_PWD}
    healthcheck:
      test: ["CMD-SHELL", "mysqladmin ping -uroot -p$$MYSQL_ROOT_PASSWORD"]
      interval: 30s
      timeout: 10s
      retries: 6
      start_period: 20s
    ports:
      - ${MYSQL_PORT}:3306
    volumes:
      - ./data/mysql:/var/lib/mysql
    command:
      - --character-set-server=utf8mb4
      - --collation-server=utf8mb4_general_ci
      - --pid-file=/var/lib/mysql/mysql.pid
    restart: always

  base-redis:
    image: redis:7.4.0
    container_name: base-redis
    command:
      - redis-server
      - --appendonly
      - "yes"
      - --requirepass
      - ${REDIS_PASSWORD}
    healthcheck:
      test: ["CMD-SHELL", "redis-cli -a ${REDIS_PASSWORD} ping"]
      interval: 10s
      timeout: 5s
      retries: 10
    ports:
      - ${REDIS_PORT}:6379
    volumes:
      - ./data/redis:/data
    restart: always

  base-register:
    image: ${JAVA_IMAGES}
    container_name: base-register
    env_file:
      - ./config/register.conf
    command: /startJava.sh
    ports:
      - ${REGISTER_PORT}:8848
      - ${REGISTER_RAFT_PORT}:9848
      - ${REGISTER_PEER_PORT}:9849
    healthcheck:
      test: ["CMD-SHELL", "wget -q -O - http://127.0.0.1:8848/nacos/ || exit 1"]
      interval: 10s
      timeout: 10s
      retries: 20
    volumes:
      - ./data/register:/app
      - ./data/register/logs:/root/nacos/logs
      - ./startJava.sh:/startJava.sh:ro
    depends_on:
      base-mysql:
        condition: service_healthy
    restart: always

  base-upms:
    image: ${JAVA_IMAGES}
    container_name: base-upms
    env_file:
      - ./config/upms.conf
    command: /startJava.sh
    ports:
      - ${UPMS_PORT}:4000
    healthcheck:
      test: ["CMD-SHELL", "wget -q -O - http://127.0.0.1:4000/actuator/health || exit 1"]
      interval: 10s
      timeout: 10s
      retries: 20
    volumes:
      - ./data/upms:/app
      - ./data/upms/logs:/logs
      - ./startJava.sh:/startJava.sh:ro
    depends_on:
      base-register:
        condition: service_healthy
      base-mysql:
        condition: service_healthy
    restart: always

  base-auth:
    image: ${JAVA_IMAGES}
    container_name: base-auth
    env_file:
      - ./config/auth.conf
    command: /startJava.sh
    ports:
      - ${AUTH_PORT}:3000
    healthcheck:
      test: ["CMD-SHELL", "wget -q -O - http://127.0.0.1:3000/actuator/health || exit 1"]
      interval: 10s
      timeout: 10s
      retries: 20
    volumes:
      - ./data/auth:/app
      - ./data/auth/logs:/logs
      - ./startJava.sh:/startJava.sh:ro
    depends_on:
      base-upms:
        condition: service_healthy
    restart: always

  base-gateway:
    image: ${JAVA_IMAGES}
    container_name: base-gateway
    env_file:
      - ./config/gateway.conf
    command: /startJava.sh
    ports:
      - ${GATEWAY_PORT}:9999
    healthcheck:
      test: ["CMD-SHELL", "wget -q -O - http://127.0.0.1:9999/actuator/health || exit 1"]
      interval: 10s
      timeout: 10s
      retries: 20
    volumes:
      - ./data/gateway:/app
      - ./data/gateway/logs:/logs
      - ./startJava.sh:/startJava.sh:ro
    depends_on:
      base-upms:
        condition: service_healthy
    restart: always

  base-xxl:
    image: ${JAVA_IMAGES}
    container_name: base-xxl
    env_file:
      - ./config/xxl_job.conf
    command: /startJava.sh
    ports:
      - ${XXL_PORT}:9080
    healthcheck:
      test: ["CMD-SHELL", "wget -q -O - http://127.0.0.1:9080/xxl-job-admin/toLogin || exit 1"]
      interval: 10s
      timeout: 10s
      retries: 20
    volumes:
      - ./data/xxl_job:/app
      - ./data/xxl_job/logs:/logs
      - ./startJava.sh:/startJava.sh:ro
    depends_on:
      base-upms:
        condition: service_healthy
    restart: always

networks:
  default:
    name: 1panel-network
    external: true
EOF
}

write_business_compose() {
    local version_dir="$1"
    local api_service_name="$2"
    local api_container_name="$3"
    local api_port="$4"
    local api_config_file="$5"
    local web_service_name="$6"
    local web_mount_source="$7"
    local web_mount_target="$8"

    cat > "${version_dir}/docker-compose.yml" <<EOF
services:
  ${api_service_name}:
    image: \${JAVA_IMAGES}
    container_name: ${api_container_name}
    env_file:
      - ./config/${api_config_file}
    command: /startJava.sh
    volumes:
      - ./data/api:/app
      - ./data/api/logs:/logs
      - ./startJava.sh:/startJava.sh:ro
      - ./config/application-dev.yml:/app/application-dev.yml:ro
    restart: always

  ${web_service_name}:
    image: nginx:1.27.1
    container_name: ${web_service_name}
    depends_on:
      - ${api_service_name}
    ports:
      - \${PANEL_APP_PORT_HTTP}:80
    volumes:
      - ${web_mount_source}:${web_mount_target}:ro
      - ./config/nginx/default.conf:/etc/nginx/conf.d/default.conf:ro
    restart: always

networks:
  default:
    name: 1panel-network
    external: true
EOF
}

write_nginx_default_conf() {
    local output_file="$1"
    local fragment_file="$2"

    cat > "${output_file}" <<'EOF'
server {
    listen 80;
    server_name _;
    client_max_body_size 100m;

EOF
    cat "${fragment_file}" >> "${output_file}"
    cat >> "${output_file}" <<'EOF'

    location / {
        return 404;
    }
}
EOF
}

resolve_service_dir() {
    local service_name="$1"
    if [[ -d "${SOURCE_DIR}/${service_name}" ]]; then
        printf '%s\n' "${SOURCE_DIR}/${service_name}"
        return 0
    fi
    if [[ -d "$(dirname "${SOURCE_DIR}")/${service_name}" ]]; then
        printf '%s\n' "$(dirname "${SOURCE_DIR}")/${service_name}"
        return 0
    fi
    die "missing service directory: ${service_name}"
}

prepare_base_framework() {
    local app_dir="${OUTPUT_DIR}/jerinte-base-framework"
    local version_dir="${app_dir}/${VERSION}"

    mkdir -p "${version_dir}"
    cp -R "${SOURCE_DIR}/base_framework/config" "${version_dir}/"
    cp -R "${SOURCE_DIR}/base_framework/data" "${version_dir}/"
    cp "${SOURCE_DIR}/base_framework/startJava.sh" "${version_dir}/startJava.sh"

    for conf in auth.conf gateway.conf mysql.conf register.conf upms.conf xxl_job.conf; do
        normalize_base_conf "${version_dir}/config/${conf}"
    done
    mark_templates "${version_dir}" auth.conf gateway.conf register.conf upms.conf xxl_job.conf
    write_template_init_script "${version_dir}"

    write_root_data_yml \
        "${app_dir}" \
        "Jerinte 基础框架" \
        "jerinte-base-framework" \
        "安装 Jerinte 共享基础服务" \
        "安装 Jerinte 共享基础服务，包含 MySQL、Redis、Nacos、Gateway、UPMS、Auth、XXL-Job。" \
        "Install the shared Jerinte base stack including MySQL, Redis, Nacos, Gateway, UPMS, Auth, and XXL-Job." \
        "2048"
    write_root_readme \
        "${app_dir}" \
        "Jerinte 基础框架" \
        "先安装这个应用，再安装具体业务应用。业务应用默认通过共享的 \`1panel-network\` 访问本应用提供的 \`base-mysql\`、\`base-redis\`、\`base-register\` 和 \`base-gateway\`。"
    copy_logo "${app_dir}"
    write_base_env "${version_dir}" "${JAVA_IMAGE_DEFAULT}" "${MYSQL_IMAGE_DEFAULT}" "${MYSQL_PASSWORD_DEFAULT}" "${REDIS_PASSWORD_DEFAULT}" "${DOCKER_SUBNET_DEFAULT}"
    write_base_version_data_yml "${version_dir}" "${JAVA_IMAGE_DEFAULT}" "${MYSQL_IMAGE_DEFAULT}" "${MYSQL_PASSWORD_DEFAULT}" "${REDIS_PASSWORD_DEFAULT}" "${DOCKER_SUBNET_DEFAULT}"
    write_base_compose "${version_dir}"
}

prepare_business_app() {
    local source_name="$1"
    local app_key="$2"
    local display_name="$3"
    local short_desc="$4"
    local zh_desc="$5"
    local en_desc="$6"
    local api_config="$7"
    local api_service="$8"
    local api_container="$9"
    local api_port="${10}"
    local web_service="${11}"
    local web_source="${12}"
    local web_target="${13}"
    local default_http_port="${14}"
    local extra_env="${15}"
    local extra_fields="${16}"
    local patch_ptrp="${17:-false}"

    local app_dir="${OUTPUT_DIR}/${app_key}"
    local version_dir="${app_dir}/${VERSION}"
    local source_dir
    source_dir="$(resolve_service_dir "${source_name}")"

    mkdir -p "${version_dir}"
    cp -R "${source_dir}/config" "${version_dir}/"
    cp -R "${source_dir}/data" "${version_dir}/"
    cp "${SOURCE_DIR}/base_framework/startJava.sh" "${version_dir}/startJava.sh"

    if [[ -f "${version_dir}/config/${api_config}" ]]; then
        normalize_business_conf "${version_dir}/config/${api_config}"
    fi
    if [[ "${patch_ptrp}" == "true" ]]; then
        replace_line "SERVER_NAME" "jerinte-ptrp-biz" "${version_dir}/config/${api_config}"
    fi
    mark_templates "${version_dir}" "${api_config}"
    write_template_init_script "${version_dir}"

    local nginx_fragment="${version_dir}/config/nginx/default.conf.fragment"
    local nginx_target="${version_dir}/config/nginx/default.conf"

    mkdir -p "${version_dir}/config/nginx"
    if [[ "${source_name}" == "ptrp" ]]; then
        cat > "${nginx_fragment}" <<'EOF'
    location /ptrp {
        alias /usr/share/nginx/ptrp-ui;
        index index.html;
        try_files $uri $uri/ /ptrp/index.html;
    }

    location ^~/ptrp/api/ {
        proxy_pass http://jerinte-ptrp-api:6081/api/;
        proxy_connect_timeout 60s;
        proxy_read_timeout 120s;
        proxy_send_timeout 120s;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto http;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $http_host;
    }
EOF
    elif [[ "${source_name}" == "medical" ]]; then
        cat > "${nginx_fragment}" <<'EOF'
    location /medical {
        alias /usr/share/nginx/medical-web;
        index index.html;
        try_files $uri $uri/ /medical/index.html;
    }

    location ^~/medical/api/ {
        proxy_pass http://base-gateway:9999/medical/;
        proxy_connect_timeout 60s;
        proxy_read_timeout 120s;
        proxy_send_timeout 120s;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto http;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $http_host;
    }
EOF
    elif [[ "${source_name}" == "pda" ]]; then
        cat > "${nginx_fragment}" <<'EOF'
    location /pda {
        alias /usr/share/nginx/pda-ui;
        index index.html;
        try_files $uri $uri/ /pda/index.html;
    }

    location ^~/pda/pda-api/ {
        proxy_pass http://jerinte-pda-api:8080/;
        proxy_connect_timeout 60s;
        proxy_read_timeout 120s;
        proxy_send_timeout 120s;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto http;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $http_host;
    }
EOF
    else
        cp "${source_dir}/config/nginx/"* "${nginx_fragment}"
    fi
    write_nginx_default_conf "${nginx_target}" "${nginx_fragment}"
    rm -f "${nginx_fragment}"

    write_root_data_yml "${app_dir}" "${display_name}" "${app_key}" "${short_desc}" "${zh_desc}" "${en_desc}" "1024"
    write_root_readme "${app_dir}" "${display_name}" "先确认 Jerinte 基础框架已安装并处于运行状态，再安装当前业务应用。该应用会加入共享的 \`1panel-network\`。"
    copy_logo "${app_dir}"
    write_business_env "${version_dir}" "${JAVA_IMAGE_DEFAULT}" "${default_http_port}" "${extra_env}"
    write_business_version_data_yml "${version_dir}" "${JAVA_IMAGE_DEFAULT}" "${default_http_port}" "${extra_fields}"
    write_business_compose "${version_dir}" "${api_service}" "${api_container}" "${api_port}" "${api_config}" "${web_service}" "${web_source}" "${web_target}"
}

require_file "${SOURCE_DIR}/install.conf"
require_file "${PLACEHOLDER_LOGO}"

JAVA_IMAGE_DEFAULT="$(read_kv "JAVA_IMAGES" "${SOURCE_DIR}/install.conf")"
MYSQL_IMAGE_DEFAULT="$(read_kv "MYSQL_IMAGES" "${SOURCE_DIR}/install.conf")"
MYSQL_PASSWORD_DEFAULT="$(read_kv "MYSQL_ROOT_PASSWORD" "${SOURCE_DIR}/install.conf")"
REDIS_PASSWORD_DEFAULT="$(read_kv "REDIS_PASSWORD" "${SOURCE_DIR}/install.conf")"
DOCKER_SUBNET_DEFAULT="$(read_kv "DOCKER_SUBNET" "${SOURCE_DIR}/install.conf")"

rm -rf "${OUTPUT_DIR}"
mkdir -p "${OUTPUT_DIR}"

prepare_base_framework

prepare_business_app \
    "pda" \
    "jerinte-pda" \
    "Jerinte PDA" \
    "安装 Jerinte PDA 业务应用" \
    "安装 Jerinte PDA 业务应用，默认接入基础框架提供的 Redis、Nacos 与 Gateway。" \
    "Install the Jerinte PDA business app. It connects to the base stack through shared Redis, Nacos, and Gateway services." \
    "pda.conf" \
    "jerinte-pda-api" \
    "jerinte-pda-api" \
    "8080" \
    "jerinte-pda-web" \
    "./data/web/dist" \
    "/usr/share/nginx/pda-ui" \
    "18080" \
    $'HIS_DB_SCHEMA=tracemed\nHIS_DB_HOST=192.168.4.46\nHIS_DB_PORT=1521\nHIS_DB_SERVER_NAME=orcltest\nHIS_DB_USER=tracemed\nHIS_DB_PWD=Med_168' \
    $'    - type: text\n      labelZh: HIS Schema\n      labelEn: HIS schema\n      required: true\n      default: tracemed\n      envKey: HIS_DB_SCHEMA\n    - type: text\n      labelZh: HIS 主机\n      labelEn: HIS host\n      required: true\n      default: 192.168.4.46\n      envKey: HIS_DB_HOST\n    - type: number\n      labelZh: HIS 端口\n      labelEn: HIS port\n      required: true\n      default: 1521\n      envKey: HIS_DB_PORT\n      rule: port\n    - type: text\n      labelZh: HIS 服务名\n      labelEn: HIS service name\n      required: true\n      default: orcltest\n      envKey: HIS_DB_SERVER_NAME\n    - type: text\n      labelZh: HIS 用户名\n      labelEn: HIS username\n      required: true\n      default: tracemed\n      envKey: HIS_DB_USER\n    - type: password\n      labelZh: HIS 密码\n      labelEn: HIS password\n      required: true\n      default: Med_168\n      envKey: HIS_DB_PWD\n    - type: text\n      labelZh: Redis 主机\n      labelEn: Redis host\n      required: true\n      default: base-redis\n      envKey: REDIS_HOST\n    - type: number\n      labelZh: Redis 端口\n      labelEn: Redis port\n      required: true\n      default: 6379\n      envKey: REDIS_PORT\n      rule: port\n    - type: password\n      labelZh: Redis 密码\n      labelEn: Redis password\n      required: true\n      default: '"${REDIS_PASSWORD_DEFAULT}"'\n      envKey: REDIS_PASSWORD\n    - type: text\n      labelZh: Nacos 主机\n      labelEn: Nacos host\n      required: true\n      default: base-register\n      envKey: NACOS_HOST\n    - type: number\n      labelZh: Nacos 端口\n      labelEn: Nacos port\n      required: true\n      default: 8848\n      envKey: NACOS_PORT\n      rule: port' \
    "false"

prepare_business_app \
    "drg" \
    "jerinte-drg" \
    "Jerinte DRG" \
    "安装 Jerinte DRG 业务应用" \
    "安装 Jerinte DRG 业务应用，默认连接基础框架的 MySQL、Redis 和 Nacos。" \
    "Install the Jerinte DRG business app. It connects to the base MySQL, Redis, and Nacos services." \
    "drg-ingroup.conf" \
    "jerinte-drg-api" \
    "jerinte-drg-api" \
    "6060" \
    "jerinte-drg-web" \
    "./data/drg-ui/dist" \
    "/usr/share/nginx/drg-ui" \
    "16060" \
    $'MYSQL_HOST=base-mysql\nMYSQL_PORT=3306\nMYSQL_USER=root\nMYSQL_PWD='"${MYSQL_PASSWORD_DEFAULT}"$'\nHIS_DB_SCHEMA=tracemed\nHIS_DB_HOST=192.168.4.46\nHIS_DB_PORT=1521\nHIS_DB_SERVER_NAME=orcltest\nHIS_DB_USER=tracemed\nHIS_DB_PWD=Med_168' \
    $'    - type: text\n      labelZh: MySQL 主机\n      labelEn: MySQL host\n      required: true\n      default: base-mysql\n      envKey: MYSQL_HOST\n    - type: number\n      labelZh: MySQL 端口\n      labelEn: MySQL port\n      required: true\n      default: 3306\n      envKey: MYSQL_PORT\n      rule: port\n    - type: text\n      labelZh: MySQL 用户名\n      labelEn: MySQL username\n      required: true\n      default: root\n      envKey: MYSQL_USER\n    - type: password\n      labelZh: MySQL 密码\n      labelEn: MySQL password\n      required: true\n      default: '"${MYSQL_PASSWORD_DEFAULT}"'\n      envKey: MYSQL_PWD\n    - type: text\n      labelZh: HIS Schema\n      labelEn: HIS schema\n      required: true\n      default: tracemed\n      envKey: HIS_DB_SCHEMA\n    - type: text\n      labelZh: HIS 主机\n      labelEn: HIS host\n      required: true\n      default: 192.168.4.46\n      envKey: HIS_DB_HOST\n    - type: number\n      labelZh: HIS 端口\n      labelEn: HIS port\n      required: true\n      default: 1521\n      envKey: HIS_DB_PORT\n      rule: port\n    - type: text\n      labelZh: HIS 服务名\n      labelEn: HIS service name\n      required: true\n      default: orcltest\n      envKey: HIS_DB_SERVER_NAME\n    - type: text\n      labelZh: HIS 用户名\n      labelEn: HIS username\n      required: true\n      default: tracemed\n      envKey: HIS_DB_USER\n    - type: password\n      labelZh: HIS 密码\n      labelEn: HIS password\n      required: true\n      default: Med_168\n      envKey: HIS_DB_PWD\n    - type: text\n      labelZh: Redis 主机\n      labelEn: Redis host\n      required: true\n      default: base-redis\n      envKey: REDIS_HOST\n    - type: number\n      labelZh: Redis 端口\n      labelEn: Redis port\n      required: true\n      default: 6379\n      envKey: REDIS_PORT\n      rule: port\n    - type: password\n      labelZh: Redis 密码\n      labelEn: Redis password\n      required: true\n      default: '"${REDIS_PASSWORD_DEFAULT}"'\n      envKey: REDIS_PASSWORD\n    - type: text\n      labelZh: Nacos 主机\n      labelEn: Nacos host\n      required: true\n      default: base-register\n      envKey: NACOS_HOST\n    - type: number\n      labelZh: Nacos 端口\n      labelEn: Nacos port\n      required: true\n      default: 8848\n      envKey: NACOS_PORT\n      rule: port' \
    "false"

prepare_business_app \
    "ptrp" \
    "jerinte-ptrp" \
    "Jerinte PTRP" \
    "安装 Jerinte PTRP 业务应用" \
    "安装 Jerinte PTRP 业务应用，默认连接基础框架的 Redis 和 Nacos。" \
    "Install the Jerinte PTRP business app. It connects to the base Redis and Nacos services." \
    "ptrp.conf" \
    "jerinte-ptrp-api" \
    "jerinte-ptrp-api" \
    "6081" \
    "jerinte-ptrp-web" \
    "./data/web/dist" \
    "/usr/share/nginx/ptrp-ui" \
    "16081" \
    $'PTRP_DB_HOST=127.0.0.1\nPTRP_DB_PORT=1521\nPTRP_DB_USER=ptrp\nPTRP_DB_PWD=ptrp123' \
    $'    - type: text\n      labelZh: PTRP 数据库主机\n      labelEn: PTRP DB host\n      required: true\n      default: 127.0.0.1\n      envKey: PTRP_DB_HOST\n    - type: number\n      labelZh: PTRP 数据库端口\n      labelEn: PTRP DB port\n      required: true\n      default: 1521\n      envKey: PTRP_DB_PORT\n      rule: port\n    - type: text\n      labelZh: PTRP 数据库用户\n      labelEn: PTRP DB user\n      required: true\n      default: ptrp\n      envKey: PTRP_DB_USER\n    - type: password\n      labelZh: PTRP 数据库密码\n      labelEn: PTRP DB password\n      required: true\n      default: ptrp123\n      envKey: PTRP_DB_PWD\n    - type: text\n      labelZh: Redis 主机\n      labelEn: Redis host\n      required: true\n      default: base-redis\n      envKey: REDIS_HOST\n    - type: number\n      labelZh: Redis 端口\n      labelEn: Redis port\n      required: true\n      default: 6379\n      envKey: REDIS_PORT\n      rule: port\n    - type: password\n      labelZh: Redis 密码\n      labelEn: Redis password\n      required: true\n      default: '"${REDIS_PASSWORD_DEFAULT}"'\n      envKey: REDIS_PASSWORD\n    - type: text\n      labelZh: Nacos 主机\n      labelEn: Nacos host\n      required: true\n      default: base-register\n      envKey: NACOS_HOST\n    - type: number\n      labelZh: Nacos 端口\n      labelEn: Nacos port\n      required: true\n      default: 8848\n      envKey: NACOS_PORT\n      rule: port' \
    "true"

prepare_business_app \
    "medical" \
    "jerinte-medical" \
    "Jerinte Medical" \
    "安装 Jerinte Medical 业务应用" \
    "安装 Jerinte Medical 业务应用，默认连接基础框架的 Gateway、Redis 和 Nacos。" \
    "Install the Jerinte Medical business app. It connects to the base Gateway, Redis, and Nacos services." \
    "medical.conf" \
    "jerinte-medical-api" \
    "jerinte-medical-api" \
    "6666" \
    "jerinte-medical-web" \
    "./data/web/dist" \
    "/usr/share/nginx/medical-web" \
    "16666" \
    $'HIS_DB_SCHEMA=tracemed\nHIS_DB_HOST=192.168.4.46\nHIS_DB_PORT=1521\nHIS_DB_SERVER_NAME=orcltest\nHIS_DB_USER=tracemed\nHIS_DB_PWD=Med_168' \
    $'    - type: text\n      labelZh: HIS Schema\n      labelEn: HIS schema\n      required: true\n      default: tracemed\n      envKey: HIS_DB_SCHEMA\n    - type: text\n      labelZh: HIS 主机\n      labelEn: HIS host\n      required: true\n      default: 192.168.4.46\n      envKey: HIS_DB_HOST\n    - type: number\n      labelZh: HIS 端口\n      labelEn: HIS port\n      required: true\n      default: 1521\n      envKey: HIS_DB_PORT\n      rule: port\n    - type: text\n      labelZh: HIS 服务名\n      labelEn: HIS service name\n      required: true\n      default: orcltest\n      envKey: HIS_DB_SERVER_NAME\n    - type: text\n      labelZh: HIS 用户名\n      labelEn: HIS username\n      required: true\n      default: tracemed\n      envKey: HIS_DB_USER\n    - type: password\n      labelZh: HIS 密码\n      labelEn: HIS password\n      required: true\n      default: Med_168\n      envKey: HIS_DB_PWD\n    - type: text\n      labelZh: Redis 主机\n      labelEn: Redis host\n      required: true\n      default: base-redis\n      envKey: REDIS_HOST\n    - type: number\n      labelZh: Redis 端口\n      labelEn: Redis port\n      required: true\n      default: 6379\n      envKey: REDIS_PORT\n      rule: port\n    - type: password\n      labelZh: Redis 密码\n      labelEn: Redis password\n      required: true\n      default: '"${REDIS_PASSWORD_DEFAULT}"'\n      envKey: REDIS_PASSWORD\n    - type: text\n      labelZh: Nacos 主机\n      labelEn: Nacos host\n      required: true\n      default: base-register\n      envKey: NACOS_HOST\n    - type: number\n      labelZh: Nacos 端口\n      labelEn: Nacos port\n      required: true\n      default: 8848\n      envKey: NACOS_PORT\n      rule: port' \
    "false"

printf 'generated local apps in %s\n' "${OUTPUT_DIR}"
