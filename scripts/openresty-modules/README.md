# OpenResty Dynamic Module Diagnostics

This script collects diagnostics from an installed 1Panel OpenResty instance.

## Requirements

- Bash 4.3 or newer
- Docker Engine with the Compose v2 plugin
- `jq`, `python3`, `file`, `binutils`, `tar`, and GNU coreutils

On Debian or Ubuntu:

```bash
sudo apt-get update
sudo apt-get install -y jq python3 file binutils tar
```

Make the scripts executable:

```bash
chmod +x scripts/openresty-modules/*.sh
```

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

Do not edit generated module state or managed config files while a 1Panel app
task is running. Review the diagnostic archive before sharing it.
