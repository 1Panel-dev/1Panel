#!/usr/bin/env bash

set -Eeuo pipefail

RUN_ID="$(date -u +%Y%m%dT%H%M%SZ)-$$"
INSTALL_DIR=""
OUTPUT_DIR="${OUTPUT_DIR:-${PWD}/openresty-module-diagnostics/${RUN_ID}}"
CONTAINER=""
FULL_CONFIG=0
CREATE_ARCHIVE=1
NGINX_TEST=1
UNEXPECTED_FAILURE=""
declare -a FAILED_CHECKS=()

usage() {
    cat <<'EOF'
Usage: diagnose-install.sh INSTALL_DIR [options]

Collect a mostly read-only diagnostic report for an installed 1Panel OpenResty.
The only container command with behavior is `nginx -t`; no reload is performed.

Options:
  --output PATH       Result directory
  --container NAME    Override the container discovered from Docker Compose
  --full-config       Retain full `nginx -T` output (may contain sensitive data)
  --no-nginx-test     Do not execute nginx -t/-T in the running container
  --no-archive        Do not create a .tar.gz report
  -h, --help          Show this help
EOF
}

log() {
    printf '[%s] %s\n' "$(date -u +%H:%M:%S)" "$*" | tee -a "${OUTPUT_DIR}/run.log"
}

mark_failed() {
    FAILED_CHECKS+=("$1")
    log "CHECK FAILED: $1"
}

mark_passed() {
    log "CHECK PASSED: $1"
}

require_command() {
    if ! command -v "$1" >/dev/null 2>&1; then
        printf 'Required command not found: %s\n' "$1" >&2
        exit 2
    fi
}

on_error() {
    local status="$1" line="$2" command="$3"
    UNEXPECTED_FAILURE="line ${line}: ${command} (exit ${status})"
    return "${status}"
}

finalize() {
    local status=$?
    set +e
    if [[ -n "${UNEXPECTED_FAILURE}" ]]; then
        printf '%s\n' "${UNEXPECTED_FAILURE}" >"${OUTPUT_DIR}/unexpected-failure.txt"
    fi
    {
        printf 'install_dir=%s\n' "${INSTALL_DIR}"
        printf 'container=%s\n' "${CONTAINER}"
        printf 'failed_checks=%s\n' "${#FAILED_CHECKS[@]}"
        local check
        for check in "${FAILED_CHECKS[@]:-}"; do
            [[ -n "${check}" ]] && printf 'failure=%s\n' "${check}"
        done
    } >"${OUTPUT_DIR}/summary.txt"

    if [[ "${CREATE_ARCHIVE}" -eq 1 ]]; then
        local archive="${OUTPUT_DIR%/}.tar.gz"
        tar -czf "${archive}" -C "$(dirname -- "${OUTPUT_DIR}")" "$(basename -- "${OUTPUT_DIR}")" 2>/dev/null || true
        printf 'Diagnostic archive: %s\n' "${archive}"
    fi
    printf 'Diagnostic directory: %s\n' "${OUTPUT_DIR}"

    if [[ "${status}" -eq 0 && "${#FAILED_CHECKS[@]}" -gt 0 ]]; then
        status=1
    fi
    exit "${status}"
}

if [[ $# -eq 0 ]]; then
    usage >&2
    exit 2
fi
if [[ "$1" == "-h" || "$1" == "--help" ]]; then
    usage
    exit 0
fi

INSTALL_DIR="$1"
shift
while [[ $# -gt 0 ]]; do
    case "$1" in
        --output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        --container)
            CONTAINER="$2"
            shift 2
            ;;
        --full-config)
            FULL_CONFIG=1
            shift
            ;;
        --no-nginx-test)
            NGINX_TEST=0
            shift
            ;;
        --no-archive)
            CREATE_ARCHIVE=0
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

[[ -d "${INSTALL_DIR}" ]] || {
    printf 'Install directory not found: %s\n' "${INSTALL_DIR}" >&2
    exit 2
}
INSTALL_DIR="$(cd -- "${INSTALL_DIR}" && pwd -P)"
if [[ -d "${OUTPUT_DIR}" ]] && find "${OUTPUT_DIR}" -mindepth 1 -print -quit | grep -q .; then
    printf 'Output directory must be empty: %s\n' "${OUTPUT_DIR}" >&2
    exit 2
fi
mkdir -p "${OUTPUT_DIR}"
OUTPUT_DIR="$(cd -- "${OUTPUT_DIR}" && pwd -P)"

trap 'on_error "$?" "$LINENO" "$BASH_COMMAND"' ERR
trap finalize EXIT

preflight() {
    [[ "$(uname -s)" == "Linux" ]] || {
        printf 'This diagnostic script must run on Linux.\n' >&2
        exit 2
    }
    require_command docker
    require_command jq
    require_command python3
    require_command sha256sum
    require_command tar

    docker version >"${OUTPUT_DIR}/docker-version.txt" 2>&1
    docker info >"${OUTPUT_DIR}/docker-info.txt" 2>&1
    uname -a >"${OUTPUT_DIR}/uname.txt"
    cp /etc/os-release "${OUTPUT_DIR}/os-release.txt" 2>/dev/null || true
    df -h >"${OUTPUT_DIR}/disk-free.txt"
    free -h >"${OUTPUT_DIR}/memory.txt" 2>&1 || true
    log "Inspecting ${INSTALL_DIR}"
}

collect_filesystem_state() {
    [[ -d "${INSTALL_DIR}/modules" ]] || mark_failed "module artifact directory is missing"
    [[ -d "${INSTALL_DIR}/conf/modules-enabled" ]] || mark_failed "managed module config directory is missing"
    find "${INSTALL_DIR}/modules" -maxdepth 5 -printf '%M\t%u:%g\t%s\t%TY-%Tm-%TdT%TH:%TM:%TS\t%p\n' \
        >"${OUTPUT_DIR}/module-files.txt" 2>&1 || true
    find "${INSTALL_DIR}/conf/modules-enabled" -maxdepth 1 -type f -printf '%f\n' \
        >"${OUTPUT_DIR}/managed-config-files.txt" 2>&1 || true
    grep -RnsE '^[[:space:]]*load_module[[:space:]]+' "${INSTALL_DIR}/conf/modules-enabled" \
        >"${OUTPUT_DIR}/load-module-directives.txt" 2>&1 || true
    grep -E '^(RESTY_|CONTAINER_NAME=|PANEL_APP_PORT_HTTP=)' "${INSTALL_DIR}/.env" \
        >"${OUTPUT_DIR}/relevant-env.txt" 2>/dev/null || true

    if [[ -f "${INSTALL_DIR}/build/module.json" ]]; then
        jq 'map(if has("script") then .script = "<redacted>" else . end)' \
            "${INSTALL_DIR}/build/module.json" >"${OUTPUT_DIR}/module-state.json"
    else
        mark_failed "module state file is missing"
    fi
    if [[ -f "${INSTALL_DIR}/build/module.catalog.json" ]]; then
        jq 'map(if has("script") then .script = "<redacted>" else . end)' \
            "${INSTALL_DIR}/build/module.catalog.json" >"${OUTPUT_DIR}/module-catalog.json"
    fi
}

validate_artifacts() {
    local state="${INSTALL_DIR}/build/module.json"
    [[ -f "${state}" ]] || return 0
    if python3 - "${state}" "${INSTALL_DIR}/modules" >"${OUTPUT_DIR}/artifact-validation.tsv" <<'PY'
import hashlib
import json
import os
import pathlib
import sys

state_path = pathlib.Path(sys.argv[1])
modules_root = pathlib.Path(sys.argv[2]).resolve()
modules = json.loads(state_path.read_text(encoding="utf-8"))
failed = False
print("module\tbuild_status\ttarget_key\tartifact\texpected\tactual\tresult")
for module in modules:
    for build in module.get("builds") or []:
        target_key = (build.get("target") or {}).get("key", "")
        for artifact in build.get("artifacts") or []:
            relative = artifact.get("path", "")
            expected = artifact.get("checksum", "")
            result = "OK"
            actual = ""
            try:
                pure = pathlib.PurePosixPath(relative)
                if not relative or pure.is_absolute() or ".." in pure.parts or "\\" in relative:
                    raise ValueError("unsafe-path")
                candidate = modules_root / pathlib.Path(*pure.parts)
                if candidate.is_symlink():
                    raise ValueError("symlink-not-allowed")
                full_path = candidate.resolve(strict=True)
                if modules_root not in full_path.parents:
                    raise ValueError("outside-module-root")
                if not full_path.is_file():
                    raise ValueError("not-regular-file")
                digest = hashlib.sha256()
                with full_path.open("rb") as handle:
                    for chunk in iter(lambda: handle.read(1024 * 1024), b""):
                        digest.update(chunk)
                actual = digest.hexdigest()
                if actual.lower() != expected.lower():
                    raise ValueError("checksum-mismatch")
            except Exception as error:
                result = str(error)
                failed = True
            print("\t".join([
                module.get("name", ""), build.get("status", ""), target_key,
                relative, expected, actual, result,
            ]))
sys.exit(1 if failed else 0)
PY
    then
        mark_passed "artifact paths and checksums"
    else
        mark_failed "artifact paths or checksums"
    fi
}

validate_managed_configs() {
    if python3 - "${INSTALL_DIR}/conf/modules-enabled" "${INSTALL_DIR}/modules" \
        >"${OUTPUT_DIR}/managed-config-validation.tsv" <<'PY'
import pathlib
import re
import sys

config_root = pathlib.Path(sys.argv[1])
modules_root = pathlib.Path(sys.argv[2]).resolve()
container_prefix = "/usr/local/openresty/nginx/modules/1panel/"
pattern = re.compile(r"^\s*load_module\s+([^;]+);", re.MULTILINE)
failed = False
print("config\tcontainer_path\thost_path\tresult")
if config_root.exists():
    for config in sorted(config_root.glob("1panel-module-*.conf")):
        content = config.read_text(encoding="utf-8")
        for value in pattern.findall(content):
            container_path = value.strip().strip('"\'')
            result = "OK"
            host_path = ""
            try:
                if not container_path.startswith(container_prefix):
                    raise ValueError("unexpected-container-path")
                relative = pathlib.PurePosixPath(container_path[len(container_prefix):])
                if ".." in relative.parts:
                    raise ValueError("unsafe-path")
                resolved = (modules_root / pathlib.Path(*relative.parts)).resolve(strict=True)
                if modules_root not in resolved.parents or not resolved.is_file():
                    raise ValueError("missing-artifact")
                host_path = str(resolved)
            except Exception as error:
                result = str(error)
                failed = True
            print("\t".join([config.name, container_path, host_path, result]))
sys.exit(1 if failed else 0)
PY
    then
        mark_passed "managed load_module configs"
    else
        mark_failed "managed load_module configs"
    fi
}

collect_compose_state() {
    local compose_file="${INSTALL_DIR}/docker-compose.yml"
    if [[ ! -f "${compose_file}" ]]; then
        mark_failed "docker-compose.yml is missing"
        return 0
    fi

    if (cd "${INSTALL_DIR}" && docker compose config --format json) >"${OUTPUT_DIR}/compose.json" 2>"${OUTPUT_DIR}/compose-config.log"; then
        mark_passed "docker compose config"
    else
        mark_failed "docker compose config"
        return 0
    fi

    if jq -e '
        [.services[].volumes[]?]
        | (any(.[]; (.target | rtrimstr("/")) == "/usr/local/openresty/nginx/modules/1panel" and .read_only == true))
          and (any(.[]; (.target | rtrimstr("/")) == "/usr/local/openresty/nginx/conf/modules-enabled" and .read_only == true))
    ' "${OUTPUT_DIR}/compose.json" >/dev/null; then
        mark_passed "read-only module mounts"
    else
        mark_failed "read-only module mounts"
    fi

    (cd "${INSTALL_DIR}" && docker compose ps -a --format json) >"${OUTPUT_DIR}/compose-ps.json" 2>&1 || true
    if [[ -z "${CONTAINER}" ]]; then
        local cid
        cid="$(cd "${INSTALL_DIR}" && docker compose ps -q 2>/dev/null | head -n 1)"
        if [[ -n "${cid}" ]]; then
            CONTAINER="$(docker inspect --format '{{.Name}}' "${cid}" | sed 's#^/##')"
        else
            CONTAINER="$(jq -r '[.services[] | select((.image // "") | test("openresty"; "i")) | .container_name][0] // empty' \
                "${OUTPUT_DIR}/compose.json")"
        fi
    fi
}

compare_target_identity() {
    local state="${INSTALL_DIR}/build/module.json"
    local current_image_id="$1"
    [[ -f "${state}" ]] || return 0
    if python3 - "${state}" "${current_image_id}" "${INSTALL_DIR}/conf/modules-enabled" \
        >"${OUTPUT_DIR}/target-identity.tsv" <<'PY'
import json
import pathlib
import re
import sys

modules = json.load(open(sys.argv[1], encoding="utf-8"))
current = sys.argv[2]
config_root = pathlib.Path(sys.argv[3])
pattern = re.compile(r"^\s*load_module\s+([^;]+);", re.MULTILINE)
prefix = "/usr/local/openresty/nginx/modules/1panel/"
loaded = set()
if config_root.exists():
    for config in config_root.glob("1panel-module-*.conf"):
        for value in pattern.findall(config.read_text(encoding="utf-8")):
            container_path = value.strip().strip('"\'')
            if container_path.startswith(prefix):
                loaded.add(container_path[len(prefix):])
failed = False
print("module\tenabled\tmode\tready_image_ids\tmanaged_artifacts\tresult")
for module in modules:
    mode = module.get("buildMode") or "static"
    if module.get("deleted") or not module.get("enable") or mode == "static":
        continue
    ready = [build for build in module.get("builds") or [] if build.get("status") == "ready"]
    digests = sorted({(build.get("target") or {}).get("imageDigest", "") for build in ready})
    candidates = []
    if not ready:
        result = "NO-READY-BUILD"
        failed = True
    elif current in digests:
        result = "MATCH"
        candidates = [build for build in ready if (build.get("target") or {}).get("imageDigest") == current]
    elif not any(digests):
        result = "UNKNOWN-NO-IMAGE-DIGEST"
        candidates = ready
    else:
        result = "MISMATCH"
        failed = True

    managed = []
    if candidates:
        for build in candidates:
            paths = [artifact.get("path", "") for artifact in build.get("artifacts") or []]
            if paths and all(path in loaded for path in paths):
                managed = paths
                break
        if not managed:
            result += "+NOT-IN-MANAGED-CONFIG"
            failed = True
    print("\t".join([
        module.get("name", ""), str(module.get("enable", False)),
        mode, ",".join(digests), ",".join(managed), result,
    ]))
sys.exit(1 if failed else 0)
PY
    then
        mark_passed "enabled module target image identity"
    else
        mark_failed "enabled module target image identity"
    fi
}

collect_container_state() {
    if [[ -z "${CONTAINER}" ]]; then
        mark_failed "OpenResty container could not be discovered"
        return 0
    fi
    if ! docker inspect "${CONTAINER}" >/dev/null 2>&1; then
        mark_failed "container ${CONTAINER} does not exist"
        return 0
    fi

    docker inspect "${CONTAINER}" | jq '.[0] | {
        Id, Name, Image, State,
        Config: {Image: .Config.Image},
        Mounts: [.Mounts[] | {Type, Source, Destination, RW}]
    }' >"${OUTPUT_DIR}/container.json"
    docker logs --tail 1000 --timestamps "${CONTAINER}" >"${OUTPUT_DIR}/container.log" 2>&1 || true

    local running image_id image_name
    running="$(docker inspect --format '{{.State.Running}}' "${CONTAINER}")"
    image_id="$(docker inspect --format '{{.Image}}' "${CONTAINER}")"
    image_name="$(docker inspect --format '{{.Config.Image}}' "${CONTAINER}")"
    docker image inspect "${image_id}" | jq '.[0] | {Id, RepoTags, RepoDigests, Architecture, Os, Created}' \
        >"${OUTPUT_DIR}/runtime-image.json" 2>&1 || true
    printf 'container=%s\nrunning=%s\nimage_name=%s\nimage_id=%s\n' \
        "${CONTAINER}" "${running}" "${image_name}" "${image_id}" >"${OUTPUT_DIR}/runtime.txt"
    compare_target_identity "${image_id}"

    if [[ "${running}" != "true" ]]; then
        mark_failed "container ${CONTAINER} is not running"
        return 0
    fi
    mark_passed "container is running"

    docker exec "${CONTAINER}" /usr/local/openresty/nginx/sbin/nginx -V \
        >"${OUTPUT_DIR}/nginx-version.txt" 2>&1 || mark_failed "nginx -V"
    docker exec "${CONTAINER}" /bin/sh -c \
        'find /usr/local/openresty/nginx/modules/1panel -type f -name "*.so" -exec sha256sum {} \; | sort' \
        >"${OUTPUT_DIR}/container-artifact-checksums.txt" 2>&1 || true
    docker exec "${CONTAINER}" /bin/sh -c \
        'for f in $(find /usr/local/openresty/nginx/modules/1panel -type f -name "*.so" | sort); do echo "### $f"; ldd "$f" || true; done' \
        >"${OUTPUT_DIR}/container-artifact-ldd.txt" 2>&1 || true

    if [[ "${NGINX_TEST}" -eq 1 ]]; then
        if docker exec "${CONTAINER}" /usr/local/openresty/nginx/sbin/nginx -t \
            >"${OUTPUT_DIR}/nginx-test.txt" 2>&1; then
            mark_passed "running container nginx -t"
        else
            mark_failed "running container nginx -t"
        fi

        local full_output="${OUTPUT_DIR}/nginx-T.full.tmp"
        docker exec "${CONTAINER}" /usr/local/openresty/nginx/sbin/nginx -T >"${full_output}" 2>&1 || true
        if [[ "${FULL_CONFIG}" -eq 1 ]]; then
            mv "${full_output}" "${OUTPUT_DIR}/nginx-T.full.txt"
            log "WARNING: nginx-T.full.txt may contain credentials or private configuration"
        else
            grep -nE 'load_module|modules-enabled|nginx version:|configure arguments:' "${full_output}" \
                >"${OUTPUT_DIR}/nginx-T-modules.txt" 2>/dev/null || true
            rm -f "${full_output}"
        fi
    fi
}

main() {
    preflight
    collect_filesystem_state
    validate_artifacts
    validate_managed_configs
    collect_compose_state
    collect_container_state

    if [[ "${#FAILED_CHECKS[@]}" -gt 0 ]]; then
        log "Diagnostics completed with ${#FAILED_CHECKS[@]} failed checks"
        return 0
    fi
    log "Diagnostics completed without failed checks"
}

main "$@"
