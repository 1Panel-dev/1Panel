# Redis Commander

Redis web management tool written in node.js


# Install and Run

```
$ npm install -g redis-commander
$ redis-commander
```

Installation via `yarn` is currently not supported. Please use `npm` as package manager.

Or run Redis Commander as Docker image `ghcr.io/joeferner/redis-commander` ~~rediscommander/redis-commander~~ (instructions see below).

Multi-Arch images built are available at `ghcr.io/joeferner/redis-commander:latest`.
(https://github.com/joeferner/redis-commander/pkgs/container/redis-commander)

Remark: new version are not published to Dockerhub right now.

# Features

Web-UI to display and edit data within multiple different Redis servers.

It has support for the following data types to view, add, update and delete data:
* Strings
* Lists
* Sets
* Sorted Set
* Streams (Basic support based on HFXBus project from https://github.com/exocet-engineering/hfx-bus, only view/add/delete data)
* ReJSON documents (Basic support, only for viewing values of ReJSON type keys)

# Usage

```
$ redis-commander --help
Options:
  --redis-port                         The port to find redis on.                  [string]
  --redis-host                         The host to find redis on.                  [string]
  --redis-socket                       The unix-socket to find redis on.           [string]
  --redis-username                     The redis username.                         [string]
  --redis-password                     The redis password.                         [string]
  --redis-db                           The redis database.                         [string]
  --redis-label                        The label to display for the connection.    [string]
  --redis-tls                          Use TLS for connection to redis server or sentinel. [boolean] [default: false]
  --redis-optional                     Set to true if no permanent auto-reconnect shall be done if server is down [boolean] [default: false]
  --sentinel-port                      The port to find redis sentinel on.         [string]
  --sentinel-host                      The host to find redis sentinel on.         [string]
  --sentinels                          Comma separated list of sentinels with host:port. [string]
  --sentinel-name                      The redis sentinel group name to use.       [string]  [default: mymaster]
  --sentinel-username                  The username for sentinel instance.         [string]
  --sentinel-password                  The password for sentinel instance.         [string]
  --http-auth-username, --http-u       The http authorisation username.            [string]
  --http-auth-password, --http-p       The http authorisation password.            [string]
  --http-auth-password-hash, --http-h  The http authorisation password hash.       [string]
  --address, -a                        The address to run the server on.           [string]  [default: 0.0.0.0]
  --port, -p                           The port to run the server on.              [string]  [default: 8081]
  --url-prefix, -u                     The url prefix to respond on.               [string]  [default: ""]
  --root-pattern, --rp                 The root pattern of the redis keys.         [string]  [default: "*"]
  --read-only                          Start app in read-only mode.                [boolean] [default: false]
  --trust-proxy                        App is run behind proxy (enable Express "trust proxy") [boolean|string] [default: false]
  --nosave, --ns                       Do not save new connections to config file. [boolean] [default: true]
  --noload, --nl                       Do not load connections from config.        [boolean] [default: false]
  --use-scan, --sc                     Use scan instead of keys.                   [boolean] [default: false]
  --clear-config, --cc                 Clear configuration file.
  --migrate-config                     Migrate old configuration file in $HOME to new style.
  --scan-count, --sc                   The size of each separate scan.             [integer] [default: 100]
  --no-log-data                        Do not log data values from redis store.    [boolean] [default: false]
  --open                               Open web-browser with Redis-Commander.      [boolean] [default: false]
  --folding-char, --fc                 Character to fold keys at in tree view.     [character] [default: ":"]
  --test, -t                           Test final configuration (file, env-vars, command line)
```

The connection can be established either via direct connection to redis server or indirect
via a sentinel instance. Most of this command line parameters map onto configuration params read from
the config file - see [docs/configuration.md](docs/configuration.md) and [docs/connections.md](docs/connections.md).

## Configuration

Redis Commander can be configured by configuration files, environment variables or using command line
parameters. The different types of config values overwrite each other, only the last (most important)
value is used.

For configuration files the `node-config` module (https://github.com/lorenwest/node-config) is used, with default to json syntax.

The order of precedence for all configuration values (from least to most important) is:

- Configuration files

  `default.json` - this file contains all default values and SHOULD NOT be changed

  `local.json` - optional file, all local overwrites for values inside default.json should be placed here as well
  as a list of redis connections to use at startup

  `local-<NODE_ENV>.json` - Do not add anything else than connections to this file! Redis Commander will overwrite this whenever a
  connection is added or removed via user interface. Inside docker container this file is used to store
  all connections parsed from REDIS_HOSTS env var.
  This file overwrites all connections defined inside `local.json`

  There are some more possible files available to use - please check the node-config Wiki
  for an complete list of all possible file names (https://github.com/lorenwest/node-config/wiki/Configuration-Files)

- Environment variables - the full list of env vars possible (except the docker specific ones)
  can be get from the file `config/custom-environment-variables.json` together with their mapping
  to the respective configuration key.

- Command line parameters - Overwrites everything

To check the final configuration created from files, env-vars set and command line param overwrites
start redis commander with additional param "--test". All invalid configuration keys will be listed
in the output. The config test does not check if hostnames or ip addresses can be resolved.

More information can be found in the documentation at [docs/configuration.md](docs/configuration.md)
and [docs/connections.md](docs/connections.md).

## Environment Variables

These environment variables can be used starting Redis Commander as normal
application or inside docker container (defined inside file `config/custom-environment-variables.json`)
and at [docs/configuration.md](docs/configuration.md):

```
HTTP_USER
HTTP_PASSWORD
HTTP_PASSWORD_HASH
ADDRESS
PORT
READ_ONLY
URL_PREFIX
SIGNIN_PATH
ROOT_PATTERN
NOSAVE
NO_LOG_DATA
FOLDING_CHAR
VIEW_JSON_DEFAULT
USE_SCAN
SCAN_COUNT
FLUSH_ON_IMPORT
REDIS_CONNECTION_NAME
REDIS_LABEL
CLIENT_MAX_BODY_SIZE
BINARY_AS_HEX
```

## Docker

All environment variables listed at "Environment Variables" can be used running image
with Docker. The following additional environment variables are available too (defined inside
docker startup script):

```
HTTP_PASSWORD_FILE
HTTP_PASSWORD_HASH_FILE
REDIS_PORT
REDIS_HOST
REDIS_SOCKET
REDIS_TLS
REDIS_USERNAME
REDIS_PASSWORD
REDIS_PASSWORD_FILE
REDIS_DB
REDIS_HOSTS
REDIS_OPTIONAL
SENTINEL_PORT
SENTINEL_HOST
SENTINEL_NAME
SENTINEL_USERNAME
SENTINEL_PASSWORD
SENTINEL_PASSWORD_FILE
SENTINELS
K8S_SIGTERM
```
A (partial) description for the mapping onto the cli params and into the config files can be found
at the [docs/connections.md](docs/connections.md) file.

The `K8S_SIGTERM` variable (default "0") can be set to "1" to work around kubernetes specifics
to allow pod replacement with zero downtime. More information on how kubernetes handles termination of old pods and the
setup of new ones can be found within the thread [https://github.com/kubernetes/contrib/issues/1140#issuecomment-290836405]

Hosts can be optionally specified with a comma separated string by setting the `REDIS_HOSTS` environment variable.

After running the container, `redis-commander` will be available at [localhost:8081](http://localhost:8081).

### Valid host strings

the `REDIS_HOSTS` environment variable is a comma separated list of host definitions,
where each host should follow one of these templates:

`hostname`

`label:hostname`

`label:hostname:port`

`label:hostname:port:dbIndex`

`label:hostname:port:dbIndex:password`

Connection strings defined with `REDIS_HOSTS` variable do not support TLS connections.
If remote redis server needs TLS write all connections into a config file instead
of using `REDIS_HOSTS` (see [docs/connections.md](docs/connections.md) at the end
within the more complex examples).

### With docker-compose

```
version: '3'
services:
  redis:
    container_name: redis
    hostname: redis
    image: redis

  redis-commander:
    container_name: redis-commander
    hostname: redis-commander
    image: ghcr.io/joeferner/redis-commander:latest
    restart: always
    environment:
    - REDIS_HOSTS=local:redis:6379
    ports:
    - "8081:8081"
```

### Without docker-compose

#### Simplest

If you're running redis on `localhost:6379`, this is all you need to get started.

```
docker run --rm --name redis-commander -d -p 8081:8081 \
  ghcr.io/joeferner/redis-commander:latest
```

#### Specify single host

```
docker run --rm --name redis-commander -d -p 8081:8081 \
  --env REDIS_HOSTS=10.10.20.30 \
  ghcr.io/joeferner/redis-commander:latest
```

#### Specify multiple hosts with labels

```
docker run --rm --name redis-commander -d -p 8081:8081 \
  --env REDIS_HOSTS=local:localhost:6379,myredis:10.10.20.30 \
  ghcr.io/joeferner/redis-commander:latest
```

## Kubernetes

An example deployment can be found at [k8s/redis-commander/deployment.yaml](k8s/redis-commander/deployment.yaml).

If you already have a cluster running with `redis` in the default namespace, deploy `redis-commander` with `kubectl apply -f k8s/redis-commander`. If you don't have `redis` running yet, you can deploy a simple pod with `kubectl apply -f k8s/redis`.

Alternatively, you can add a container to a deployment's spec like this:

```
containers:
- name: redis-commander
  image: ghcr.io/joeferner/redis-commander
  env:
  - name: REDIS_HOSTS
    value: instance1:redis:6379
  ports:
  - name: redis-commander
    containerPort: 8081
```

known issues with Kubernetes:

* using REDIS_HOSTS works only with a password-less redis db. You must specify REDIS_HOST on a password
  protected redis db


## Helm chart

You can install the application on any Kubernetes cluster using Helm.
There is no helm repo available currently, therefore local checkout of helm sources inside
this repo is needed:

```
helm -n myspace install redis-web-ui ./k8s/helm-chart/redis-commander
```

More [Documentation](k8s/helm-chart/README.md) about this Helm chart and its values.

## OpenShift V3

To use the stock Node.js image builder do the following.

1. Open Catalog and select the Node.js template
1. Specify the name of the application and the URL to the [redis-command github repository](https://github.com/joeferner/redis-commander.git)
1. Click the ```advanced options``` link
1. (optional) specify the hostname for the route - _if one is not specified it will be generated_
1. In the Deployment Configuration section
    * Add ```REDIS_HOST``` environment variable whose value is the name of the redis service - e.g., ```redis```
    * Add ```REDIS_PORT``` environment variable whose value is the port exposed of the redis service - e.g., ```6379```
    * Add value from secret generated by the [redis template](https://github.com/sclorg/redis-container/blob/master/examples/redis-persistent-template.json):
        * name: ```REDIS_PASSWORD```
        * resource: ```redis```
        * key: ```database-password```
1. (optional) specify a label such as ```appl=redis-commander-dev1```
    * _this label will be applied on all objects created allowing for easy deletion later via:_
   ```
   oc delete all --selector appl=redis-commander-dev1
   ```

## Helper Scripts
### Generate BCrypted password hash

Redis commander allows setting either a plain text password for http authentication or an already bcrypted
password hash.
To generate a hashed password the script `bin/bcrypt-password.js` can be used. The parameter "-p" to set password should be given.

Usage example:
```
$ git clone https://github.com/joeferner/redis-commander.git
$ cd redis-commander/bin
$ node bcrypt-password.js -p myplainpass
$2b$10BQPbC8dlxeEqB/nXOkyjr.tlafGZ28J3ug8sWIMRoeq5LSVOXpl3W
```

This generated hash can be set inside the config file as "server.httpAuth.passwordHash", as env var "HTTP_PASSWORD_HASH"
or on the command line as `--http-auth-password-hash`.
Running inside docker image a file containing this password hash can be set via env var
`HTTP_PASSWORD_HASH_FILE`

## Build images based on this one

To use this images as a base image for other images you need to call "apk update" inside your Dockerfile
before adding other apk packages with "apk add foo". Afterwards, to reduce your image size, you may
remove all temporary apk configs too again as this Dockerfile does.
