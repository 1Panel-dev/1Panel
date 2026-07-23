# OpenResty Dynamic Module Linux Tests

These scripts test the local dynamic-module build path and collect diagnostics
from an installed 1Panel OpenResty instance. Run them on a disposable Linux
host with Docker access before testing on a production installation.

## Requirements

- Bash 4.3 or newer
- Docker Engine with the Compose v2 plugin
- `jq`, `python3`, `file`, `binutils`, `tar`, and GNU coreutils
- Internet access for runtime images and Ubuntu build packages
- Go, only when `--source-checks` is used

On Debian or Ubuntu:

```bash
sudo apt-get update
sudo apt-get install -y jq python3 file binutils tar
```

Make the scripts executable:

```bash
chmod +x scripts/openresty-modules/*.sh
```

## Builder Test

Start with one version and one small module:

```bash
./scripts/openresty-modules/test-builder.sh \
  --appstore ../appstore \
  --versions 1.31.1.1-0-noble \
  --modules ngx_brotli \
  --source-checks
```

Test every catalog module against all refactored OpenResty versions:

```bash
./scripts/openresty-modules/test-builder.sh --appstore ../appstore
```

Bypass Docker's module build cache when reproducing a compiler problem:

```bash
./scripts/openresty-modules/test-builder.sh \
  --appstore ../appstore \
  --versions 1.31.1.1-0-noble \
  --modules geoip2 \
  --no-cache \
  --keep-context \
  --keep-docker
```

The builder test performs these phases for every selected version:

1. Validate appstore JSON, shell scripts, Compose mounts, and Nginx include.
2. Pull and identify the exact target runtime image.
3. Convert catalog options to dynamic configure options.
4. Build every module with `Dockerfile.modules` and copy `/out` locally.
5. Record SHA-256, ELF metadata, compiler output, and runtime dependencies.
6. Validate individual modules for debugging. Individual failures are warnings
   by default because modules may depend on an earlier module.
7. Validate all modules together in catalog `loadOrder`.
8. Start an isolated OpenResty master, add module configs, and hot reload.
9. Inject a missing module, prove `nginx -t` rejects it, restore the config,
   and prove the running process remains healthy.

Use `--strict-individual` when every selected module is expected to load alone.

Use `--mirror URL` (environment variable `MIRROR`) to pass an apt mirror as
`CONTAINER_PACKAGE_URL` to module builds, matching the 1Panel module build.

Results are written to:

```text
openresty-module-test-results/<run-id>/
```

Important files:

- `summary.tsv`: result per OpenResty version
- `work/<version>/logs/build-*.log`: complete BuildKit output
- `work/<version>/logs/load-combined.log`: authoritative ABI/load-order test
- `work/<version>/artifacts.tsv`: module paths, checksums, and sizes
- `work/<version>/image-inspect.json`: exact target image identity
- `work/<version>/runtime/`: reload and rollback test logs
- `<result-dir>.tar.gz`: automatically created when an unexpected failure occurs

## Installed Instance Diagnostics

Find the OpenResty installation directory first. A common path is similar to:

```text
/opt/1panel/apps/openresty/openresty
```

Run the diagnostic collector:

```bash
./scripts/openresty-modules/diagnose-install.sh \
  /opt/1panel/apps/openresty/openresty
```

Override container discovery when needed:

```bash
./scripts/openresty-modules/diagnose-install.sh \
  /opt/1panel/apps/openresty/openresty \
  --container 1Panel-openresty
```

The collector checks:

- `module.json` artifact paths and SHA-256 checksums
- managed `load_module` files and host/container path mapping
- read-only Compose mounts
- current container image ID versus enabled module target image IDs
- container state, Nginx build options, `nginx -t`, loaded module directives,
  module checksums inside the container, `ldd`, and recent logs

The default report does not retain full `nginx -T` output. Use
`--full-config` only on a test host because the resulting archive may contain
credentials or private site configuration.

Module scripts are redacted from the copied state files by default. Container
logs and error strings can still contain site names, URLs, or command output;
review an archive before sharing it outside your team.

## Final Manual Matrix

Run this matrix through the 1Panel UI on a disposable installation. Collect a
diagnostic archive after each important transition.

1. Install the oldest selected OpenResty version with every module disabled.
2. Switch one module to `auto`, enable it, and build it locally.
3. Confirm `buildStatus=ready`, `compatibility=compatible`, and `nginx -t`.
4. Force rebuild it. Confirm the artifact path changes and the old config is
   replaced only after the new artifact passes validation.
5. Enable all catalog modules and verify catalog load order with the builder
   test and the installed-instance collector.
6. On the test host, make one module script return a failure. Confirm the old
   managed config and old ready artifact remain active.
7. Restore the module definition and rebuild successfully.
8. Upgrade OpenResty. Confirm every enabled dynamic module has a ready build
   whose target image ID matches the new running container.
9. Restart the container and host. Run the diagnostic collector again to prove
   the persisted mounts and configs remain valid.
10. Switch a module to `static`, rebuild, then switch it back to `dynamic` and
    verify that all enabled dynamic modules are regenerated for the new image.

## Failure Triage

| Symptom | First evidence to inspect |
| --- | --- |
| Docker build fails | `logs/build-<module>.log`, `inputs/<module>/` |
| `.so` missing | build log and the `module-output` stage `/out` checks |
| Individual load fails, combined passes | module dependency and `loadOrder` |
| Combined load fails | ABI mismatch, duplicate module, missing shared library |
| `ldd` shows `not found` | bundled `lib/`, RPATH, or future runtime packages |
| Checksum mismatch | interrupted copy, manual modification, stale state file |
| Target image mismatch | module was not rebuilt after image upgrade/rebuild |
| Builder passes, installed `nginx -t` fails | Compose mounts or managed config |
| Reload fails but old process runs | inspect rollback logs and old config snapshot |

Do not edit generated module state or managed config files while a 1Panel app
task is running. Preserve the result directory and archive before retrying a
failed build.
