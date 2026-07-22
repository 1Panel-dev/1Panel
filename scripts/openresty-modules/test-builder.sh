#!/usr/bin/env bash

set -Eeuo pipefail
export DOCKER_BUILDKIT="${DOCKER_BUILDKIT:-1}"

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd -P)"
REPO_ROOT="$(cd -- "${SCRIPT_DIR}/../.." && pwd -P)"
DEFAULT_APPSTORE_ROOT="$(cd -- "${REPO_ROOT}/.." && pwd -P)/appstore"
RUN_ID="$(date -u +%Y%m%dT%H%M%SZ)-$$"

APPSTORE_ROOT="${APPSTORE_ROOT:-${DEFAULT_APPSTORE_ROOT}}"
VERSIONS_CSV="1.27.1.2-5-1-focal,1.29.2.5-0-noble,1.31.1.1-0-noble"
MODULES_CSV=""
OUTPUT_DIR="${OUTPUT_DIR:-${PWD}/openresty-module-test-results/${RUN_ID}}"
MIRROR="${MIRROR:-}"
SKIP_PULL=0
NO_CACHE=0
KEEP_DOCKER=0
KEEP_CONTEXT=0
RUN_SOURCE_CHECKS=0
STRICT_INDIVIDUAL=0
CLEANUP_READY=0

declare -a CREATED_CONTAINERS=()
declare -a CREATED_IMAGES=()
declare -a VERSIONS=()
declare -a REQUESTED_MODULES=()

usage() {
    cat <<'EOF'
Usage: test-builder.sh [options]

Build and load-test OpenResty dynamic modules directly from the appstore tree.

Options:
  --appstore PATH       Appstore repository root (default: sibling appstore repo)
  --versions CSV        App versions to test
  --modules CSV         Module names to test (default: every catalog module)
  --output PATH         Persistent result directory
  --mirror URL          apt mirror for module build packages (CONTAINER_PACKAGE_URL)
  --skip-pull           Use local runtime images without pulling
  --no-cache            Pass --no-cache to every module Docker build
  --strict-individual   Fail when a module cannot load by itself
  --source-checks       Run Go module tests and go vet before Docker tests
  --keep-docker         Keep temporary images and containers
  --keep-context        Keep copied Docker build contexts
  -h, --help            Show this help

Environment equivalents: APPSTORE_ROOT, OUTPUT_DIR, MIRROR, DOCKER_BUILDKIT.
EOF
}

log() {
    printf '[%s] %s\n' "$(date -u +%H:%M:%S)" "$*" | tee -a "${OUTPUT_DIR}/run.log"
}

die() {
    log "ERROR: $*"
    return 1
}

require_command() {
    command -v "$1" >/dev/null 2>&1 || die "required command not found: $1"
}

split_csv() {
    local value="$1"
    local -n destination="$2"
    IFS=',' read -r -a destination <<<"${value}"
}

safe_name() {
    local value="$1"
    local base digest
    base="$(printf '%s' "${value}" | sed -E 's/[^a-zA-Z0-9._-]+/-/g; s/^-+//; s/-+$//' | cut -c1-48)"
    [[ -n "${base}" ]] || base="module"
    digest="$(printf '%s' "${value}" | sha256sum | awk '{print substr($1,1,8)}')"
    printf '%s-%s' "${base}" "${digest}"
}

run_logged() {
    local log_file="$1"
    shift
    mkdir -p "$(dirname -- "${log_file}")"
    set +e
    "$@" > >(tee "${log_file}") 2>&1
    local status=$?
    set -e
    return "${status}"
}

docker_rm_container() {
    local name="$1"
    docker rm -f "${name}" >/dev/null 2>&1 || true
}

cleanup() {
    local status=$?
    if [[ "${CLEANUP_READY}" -eq 0 ]]; then
        exit "${status}"
    fi
    if [[ "${KEEP_DOCKER}" -eq 0 ]]; then
        local item
        for item in "${CREATED_CONTAINERS[@]:-}"; do
            [[ -n "${item}" ]] && docker_rm_container "${item}"
        done
        for item in "${CREATED_IMAGES[@]:-}"; do
            [[ -n "${item}" ]] && docker image rm -f "${item}" >/dev/null 2>&1 || true
        done
    fi
    if [[ "${KEEP_CONTEXT}" -eq 0 && -d "${OUTPUT_DIR}/work" ]]; then
        find "${OUTPUT_DIR}/work" -mindepth 2 -maxdepth 2 -type d -name context -prune -exec rm -rf -- {} + 2>/dev/null || true
    fi
    exit "${status}"
}

write_debug_bundle() {
    local status="$1" line="$2" command="$3"
    {
        printf 'exit_status=%s\n' "${status}"
        printf 'line=%s\n' "${line}"
        printf 'command=%s\n' "${command}"
        printf 'run_id=%s\n' "${RUN_ID}"
    } >"${OUTPUT_DIR}/failure.txt"

    docker ps -a --no-trunc >"${OUTPUT_DIR}/docker-ps.txt" 2>&1 || true
    docker image ls --digests --no-trunc >"${OUTPUT_DIR}/docker-images.txt" 2>&1 || true
    docker system df >"${OUTPUT_DIR}/docker-system-df.txt" 2>&1 || true

    local item
    for item in "${CREATED_CONTAINERS[@]:-}"; do
        [[ -n "${item}" ]] || continue
        docker inspect "${item}" >"${OUTPUT_DIR}/container-${item}.json" 2>&1 || true
        docker logs "${item}" >"${OUTPUT_DIR}/container-${item}.log" 2>&1 || true
    done

    local archive="${OUTPUT_DIR%/}.tar.gz"
    tar --exclude='*/context' -czf "${archive}" -C "$(dirname -- "${OUTPUT_DIR}")" "$(basename -- "${OUTPUT_DIR}")" 2>/dev/null || true
    printf 'Debug bundle: %s\n' "${archive}" >&2
}

on_error() {
    local status="$1" line="$2" command="$3"
    set +e
    log "FAILED at line ${line}: ${command} (exit ${status})"
    write_debug_bundle "${status}" "${line}" "${command}"
    return "${status}"
}

trap cleanup EXIT

while [[ $# -gt 0 ]]; do
    case "$1" in
        --appstore)
            APPSTORE_ROOT="$2"
            shift 2
            ;;
        --versions)
            VERSIONS_CSV="$2"
            shift 2
            ;;
        --modules)
            MODULES_CSV="$2"
            shift 2
            ;;
        --output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        --mirror)
            MIRROR="$2"
            shift 2
            ;;
        --skip-pull)
            SKIP_PULL=1
            shift
            ;;
        --no-cache)
            NO_CACHE=1
            shift
            ;;
        --strict-individual)
            STRICT_INDIVIDUAL=1
            shift
            ;;
        --source-checks)
            RUN_SOURCE_CHECKS=1
            shift
            ;;
        --keep-docker)
            KEEP_DOCKER=1
            shift
            ;;
        --keep-context)
            KEEP_CONTEXT=1
            shift
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            printf 'Unknown option: %s\n' "$1" >&2
            usage >&2
            exit 2
            ;;
    esac
done

if [[ -d "${OUTPUT_DIR}" ]] && find "${OUTPUT_DIR}" -mindepth 1 -print -quit | grep -q .; then
    printf 'Output directory must be empty: %s\n' "${OUTPUT_DIR}" >&2
    exit 2
fi
mkdir -p "${OUTPUT_DIR}/work"
OUTPUT_DIR="$(cd -- "${OUTPUT_DIR}" && pwd -P)"
CLEANUP_READY=1
trap 'on_error "$?" "$LINENO" "$BASH_COMMAND"' ERR
APPSTORE_ROOT="$(cd -- "${APPSTORE_ROOT}" && pwd -P)"
split_csv "${VERSIONS_CSV}" VERSIONS
if [[ "${#VERSIONS[@]}" -eq 0 || -z "${VERSIONS[0]}" ]]; then
    die "at least one OpenResty version is required"
fi
if [[ -n "${MODULES_CSV}" ]]; then
    split_csv "${MODULES_CSV}" REQUESTED_MODULES
fi

preflight() {
    [[ "$(uname -s)" == "Linux" ]] || die "this integration test must run on Linux"
    require_command docker
    require_command jq
    require_command python3
    require_command sha256sum
    require_command sed
    require_command awk
    require_command tar
    require_command file
    require_command readelf

    docker version >"${OUTPUT_DIR}/docker-version.txt" 2>&1
    docker info >"${OUTPUT_DIR}/docker-info.txt" 2>&1
    docker compose version >"${OUTPUT_DIR}/docker-compose-version.txt" 2>&1
    uname -a >"${OUTPUT_DIR}/uname.txt"
    cp /etc/os-release "${OUTPUT_DIR}/os-release.txt" 2>/dev/null || true
    df -h >"${OUTPUT_DIR}/disk-free.txt"
    free -h >"${OUTPUT_DIR}/memory.txt" 2>&1 || true

    [[ -d "${APPSTORE_ROOT}/apps/openresty" ]] || die "invalid appstore root: ${APPSTORE_ROOT}"
    log "Results: ${OUTPUT_DIR}"
    log "Appstore: ${APPSTORE_ROOT}"
    log "Docker architecture: $(docker info --format '{{.Architecture}}')"
}

run_source_checks() {
    [[ "${RUN_SOURCE_CHECKS}" -eq 1 ]] || return 0
    require_command go
    log "Running Go dynamic-module tests"
    run_logged "${OUTPUT_DIR}/go-test.log" bash -c \
        "cd '${REPO_ROOT}/agent' && go test ./app/service -run 'NginxModule|DynamicModule' -count=1 -v"
    log "Running go vet"
    run_logged "${OUTPUT_DIR}/go-vet.log" bash -c "cd '${REPO_ROOT}/agent' && go vet ./..."
}

validate_template() {
    local version="$1"
    local app_dir="${APPSTORE_ROOT}/apps/openresty/${version}"
    local catalog="${app_dir}/build/module.catalog.json"

    [[ "${version}" =~ ^[a-zA-Z0-9._-]+$ ]] || die "unsafe version value: ${version}"
    [[ -f "${app_dir}/build/Dockerfile.modules" ]] || die "missing Dockerfile.modules for ${version}"
    [[ -f "${catalog}" ]] || die "missing module catalog for ${version}"
    jq -e 'type == "array" and length > 0 and all(.[]; .name and .params and .provider == "local")' \
        "${catalog}" >/dev/null
    bash -n "${app_dir}/scripts/init.sh"
    bash -n "${app_dir}/scripts/upgrade.sh"
    grep -Fq 'conf/modules-enabled:/usr/local/openresty/nginx/conf/modules-enabled/:ro' "${app_dir}/docker-compose.yml"
    grep -Fq './modules:/usr/local/openresty/nginx/modules/1panel/:ro' "${app_dir}/docker-compose.yml"
    grep -Fq 'include /usr/local/openresty/nginx/conf/modules-enabled/*.conf;' "${app_dir}/conf/nginx.conf"

    mkdir -p "${OUTPUT_DIR}/work/${version}/website/conf.d" "${OUTPUT_DIR}/work/${version}/website/stream.d"
    (
        cd "${app_dir}"
        CONTAINER_NAME="openresty-template-check" \
        WEBSITE_DIR="${OUTPUT_DIR}/work/${version}/website" \
        PANEL_APP_PORT_HTTP=18080 \
        docker compose config -q
    ) >"${OUTPUT_DIR}/work/${version}/compose-config.log" 2>&1
}

write_module_inputs() {
    local catalog="$1" module="$2" context="$3" input_dir="$4"
    local module_script params dynamic_params packages

    module_script="$(jq -er --arg name "${module}" '.[] | select(.name == $name) | .script' "${catalog}")"
    params="$(jq -er --arg name "${module}" '.[] | select(.name == $name) | .params' "${catalog}")"
    packages="$(jq -er --arg name "${module}" '.[] | select(.name == $name) | (.packages // []) | join(" ")' "${catalog}")"
    dynamic_params="${params//--add-module=/--add-dynamic-module=}"

    mkdir -p "${input_dir}"
    printf '%s\n' "${module_script}" >"${input_dir}/script.txt"
    printf '%s\n' "${params}" >"${input_dir}/params.original.txt"
    printf '%s\n' "${dynamic_params}" >"${input_dir}/params.dynamic.txt"
    printf '%s\n' "${packages}" >"${input_dir}/packages.txt"
    printf '#!/bin/bash\nset -e\n%s\n' "${module_script}" >"${context}/tmp/module-pre.sh"

    python3 - "${dynamic_params}" >"${context}/tmp/module-config.args" <<'PY'
import shlex
import sys

params = sys.argv[1]
args = shlex.split(params, posix=True)
if not args:
    raise SystemExit("dynamic module parameters are empty")
if not any(arg.startswith("--add-dynamic-module=") or "=dynamic" in arg for arg in args):
    raise SystemExit("module does not declare a dynamic configure option")
for arg in args:
    if not arg.startswith("--"):
        raise SystemExit(f"unsupported configure argument: {arg!r}")
    if any(char in arg for char in "\x00\r\n;&|<>"):
        raise SystemExit(f"unsafe configure argument: {arg!r}")
    print(arg)
PY
}

validate_load_directives() {
    local image="$1" modules_root="$2" directives_file="$3" config_path="$4" log_file="$5"
    {
        cat "${directives_file}"
        printf 'error_log stderr notice;\npid /tmp/nginx.pid;\nevents {}\nhttp {}\n'
    } >"${config_path}"

    run_logged "${log_file}" docker run --rm --network none \
        -v "${modules_root}:/usr/local/openresty/nginx/modules/1panel:ro" \
        -v "${config_path}:/tmp/1panel-module-test.conf:ro" \
        --entrypoint /usr/local/openresty/nginx/sbin/nginx \
        "${image}" -t -c /tmp/1panel-module-test.conf
}

build_module() {
    local version="$1" module="$2" image="$3" app_dir="$4" version_dir="$5" context="$6" sequence="$7"
    local catalog="${app_dir}/build/module.catalog.json"
    local module_key module_dir input_dir tag build_log cid packages artifact relative checksum
    local -a artifacts=()
    module_key="$(safe_name "${module}")"
    module_dir="${version_dir}/modules/${module_key}/${RUN_ID}"
    input_dir="${version_dir}/inputs/${module_key}"
    tag="1panel/openresty-module-test:${module_key}-$(safe_name "${version}")-${RUN_ID}"
    tag="${tag:0:127}"
    build_log="${version_dir}/logs/build-${module_key}.log"

    log "[${version}] building ${module}"
    write_module_inputs "${catalog}" "${module}" "${context}" "${input_dir}"
    packages="$(cat "${input_dir}/packages.txt")"

    local -a build_args=(
        build --progress=plain --target module-output
        -f "${context}/Dockerfile.modules"
        -t "${tag}"
        --build-arg "PANEL_OPENRESTY_VERSION=${version}"
        --build-arg "RESTY_ADD_PACKAGE_BUILDDEPS=${packages}"
    )
    [[ "${NO_CACHE}" -eq 0 ]] || build_args+=(--no-cache)
    [[ -z "${MIRROR}" ]] || build_args+=(--build-arg "CONTAINER_PACKAGE_URL=${MIRROR}")
    build_args+=("${context}")

    run_logged "${build_log}" docker "${build_args[@]}"
    CREATED_IMAGES+=("${tag}")

    cid="1panel-module-copy-${module_key}-${RUN_ID}"
    cid="${cid:0:63}"
    docker create --name "${cid}" "${tag}" /bin/true >"${input_dir}/container-id.txt"
    CREATED_CONTAINERS+=("${cid}")
    mkdir -p "${module_dir}"
    docker cp "${cid}:/out/." "${module_dir}"
    docker_rm_container "${cid}"

    mapfile -t artifacts < <(find "${module_dir}" -maxdepth 1 -type f -name '*.so' -print | sort)
    [[ "${#artifacts[@]}" -gt 0 ]] || die "${module} produced no top-level .so files"

    : >"${input_dir}/load-directives.conf"
    for artifact in "${artifacts[@]}"; do
        relative="${artifact#${version_dir}/modules/}"
        [[ "${relative}" =~ ^[a-zA-Z0-9_./+-]+\.so$ ]] || \
            die "unsafe module artifact path: ${relative}"
        checksum="$(sha256sum "${artifact}" | awk '{print $1}')"
        printf '%s\t%s\t%s\t%s\n' "${module}" "${relative}" "${checksum}" "$(stat -c '%s' "${artifact}")" \
            >>"${version_dir}/artifacts.tsv"
        printf 'load_module /usr/local/openresty/nginx/modules/1panel/%s;\n' "${relative}" \
            >>"${input_dir}/load-directives.conf"
        file "${artifact}" >>"${input_dir}/file.txt"
        readelf -d "${artifact}" >>"${input_dir}/readelf-dynamic.txt" 2>&1 || true
    done

    if ! validate_load_directives "${image}" "${version_dir}/modules" \
        "${input_dir}/load-directives.conf" "${input_dir}/individual-nginx.conf" \
        "${version_dir}/logs/load-${module_key}.log"; then
        printf '%s\tindividual-load-failed\n' "${module}" >>"${version_dir}/status.tsv"
        if [[ "${STRICT_INDIVIDUAL}" -eq 1 ]]; then
            die "${module} failed individual load validation"
        fi
        log "[${version}] ${module} cannot load alone; combined validation will decide"
    else
        printf '%s\tindividual-load-ok\n' "${module}" >>"${version_dir}/status.tsv"
    fi

    cat "${input_dir}/load-directives.conf" >>"${version_dir}/combined-load-directives.conf"
    cp "${input_dir}/load-directives.conf" \
        "${version_dir}/ordered-configs/$(printf '%04d' "${sequence}")-${module_key}.conf"
}

runtime_reload_test() {
    local version="$1" image="$2" version_dir="$3"
    local runtime_dir="${version_dir}/runtime" container="1panel-module-runtime-$(safe_name "${version}")-${RUN_ID}"
    local -a module_configs=()
    container="${container:0:63}"
    mkdir -p "${runtime_dir}/modules-enabled"

    cat >"${runtime_dir}/nginx.conf" <<'EOF'
error_log stderr notice;
pid /tmp/nginx.pid;
include /tmp/modules-enabled/*.conf;
events {}
http {}
EOF
    printf '# empty initial module set\n' >"${runtime_dir}/modules-enabled/0000-empty.conf"

    mapfile -t module_configs < <(find "${version_dir}/ordered-configs" -type f -name '*.conf' -print | sort)
    [[ "${#module_configs[@]}" -gt 0 ]] || die "no module configs available for runtime test"

    log "[${version}] starting runtime reload test"
    docker run -d --name "${container}" --network none \
        -v "${version_dir}/modules:/usr/local/openresty/nginx/modules/1panel:ro" \
        -v "${runtime_dir}/modules-enabled:/tmp/modules-enabled:ro" \
        -v "${runtime_dir}/nginx.conf:/tmp/1panel-runtime-nginx.conf:ro" \
        --entrypoint /usr/local/openresty/nginx/sbin/nginx \
        "${image}" -c /tmp/1panel-runtime-nginx.conf -g 'daemon off;' \
        >"${runtime_dir}/container-id.txt"
    CREATED_CONTAINERS+=("${container}")

    local attempt
    for ((attempt = 1; attempt <= 20; attempt++)); do
        if [[ "$(docker inspect --format '{{.State.Running}}' "${container}" 2>/dev/null || true)" == "true" ]]; then
            break
        fi
        sleep 1
    done
    if [[ "$(docker inspect --format '{{.State.Running}}' "${container}" 2>/dev/null || true)" != "true" ]]; then
        docker logs "${container}" >"${runtime_dir}/startup-failure.log" 2>&1 || true
        die "runtime container failed to start"
    fi

    local config
    for config in "${module_configs[@]}"; do
        cp "${config}" "${runtime_dir}/modules-enabled/$(basename -- "${config}")"
    done
    run_logged "${runtime_dir}/nginx-test.log" docker exec "${container}" \
        /usr/local/openresty/nginx/sbin/nginx -t -c /tmp/1panel-runtime-nginx.conf
    run_logged "${runtime_dir}/nginx-reload.log" docker exec "${container}" \
        /usr/local/openresty/nginx/sbin/nginx -s reload -c /tmp/1panel-runtime-nginx.conf

    printf 'load_module /usr/local/openresty/nginx/modules/1panel/not-found.so;\n' \
        >"${runtime_dir}/modules-enabled/9999-invalid.conf"
    if docker exec "${container}" /usr/local/openresty/nginx/sbin/nginx \
        -t -c /tmp/1panel-runtime-nginx.conf >"${runtime_dir}/expected-invalid.log" 2>&1; then
        die "nginx -t unexpectedly accepted a missing module"
    fi
    rm -f "${runtime_dir}/modules-enabled/9999-invalid.conf"
    run_logged "${runtime_dir}/rollback-nginx-test.log" docker exec "${container}" \
        /usr/local/openresty/nginx/sbin/nginx -t -c /tmp/1panel-runtime-nginx.conf
    [[ "$(docker inspect --format '{{.State.Running}}' "${container}")" == "true" ]] || \
        die "runtime container stopped during rollback test"

    docker logs "${container}" >"${runtime_dir}/container.log" 2>&1 || true
    docker exec "${container}" /bin/sh -c \
        'for f in /usr/local/openresty/nginx/modules/1panel/*/*/*.so; do echo "### $f"; ldd "$f" || true; done' \
        >"${runtime_dir}/ldd.txt" 2>&1 || true
    docker_rm_container "${container}"
}

test_version() {
    local version="$1"
    local app_dir="${APPSTORE_ROOT}/apps/openresty/${version}"
    local version_dir="${OUTPUT_DIR}/work/${version}"
    local context="${version_dir}/context"
    local catalog="${app_dir}/build/module.catalog.json"
    local image="1panel/openresty:${version}"
    local -a modules=()

    validate_template "${version}"
    mkdir -p "${version_dir}/logs" "${version_dir}/inputs" "${version_dir}/modules" \
        "${version_dir}/ordered-configs"
    cp -a "${app_dir}/build" "${context}"
    : >"${version_dir}/artifacts.tsv"
    : >"${version_dir}/status.tsv"
    : >"${version_dir}/combined-load-directives.conf"

    if [[ "${#REQUESTED_MODULES[@]}" -gt 0 ]]; then
        modules=("${REQUESTED_MODULES[@]}")
    else
        mapfile -t modules < <(jq -r 'sort_by([.loadOrder // 50, .name])[] | .name' "${catalog}")
    fi

    if [[ "${SKIP_PULL}" -eq 0 ]]; then
        log "[${version}] pulling ${image}"
        run_logged "${version_dir}/logs/image-pull.log" docker pull "${image}"
    fi
    docker image inspect "${image}" >"${version_dir}/image-inspect.json"
    run_logged "${version_dir}/logs/nginx-version.log" docker run --rm \
        --entrypoint /usr/local/openresty/nginx/sbin/nginx "${image}" -V
    sha256sum "${app_dir}/build/Dockerfile.modules" >"${version_dir}/builder.sha256"
    find "${app_dir}/build/tmp" -maxdepth 1 -type f -print0 | sort -z | xargs -0 sha256sum \
        >"${version_dir}/build-inputs.sha256"

    local module sequence=0
    for module in "${modules[@]}"; do
        sequence=$((sequence + 1))
        jq -e --arg name "${module}" 'any(.[]; .name == $name)' "${catalog}" >/dev/null || \
            die "module ${module} is not present in ${version} catalog"
        build_module "${version}" "${module}" "${image}" "${app_dir}" "${version_dir}" "${context}" "${sequence}"
    done

    log "[${version}] validating the combined load order"
    validate_load_directives "${image}" "${version_dir}/modules" \
        "${version_dir}/combined-load-directives.conf" "${version_dir}/combined-nginx.conf" \
        "${version_dir}/logs/load-combined.log"
    runtime_reload_test "${version}" "${image}" "${version_dir}"
    printf '%s\tPASS\n' "${version}" >>"${OUTPUT_DIR}/summary.tsv"
    log "[${version}] PASS"
}

main() {
    preflight
    run_source_checks
    printf 'version\tresult\n' >"${OUTPUT_DIR}/summary.tsv"
    local version
    for version in "${VERSIONS[@]}"; do
        test_version "${version}"
    done
    log "All requested OpenResty module tests passed"
    log "Summary: ${OUTPUT_DIR}/summary.tsv"
}

main "$@"
