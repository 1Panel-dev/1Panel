package constant

import "slices"

const (
	ContainerPrefix = "1Panel-"

	AppNormal   = "Normal"
	AppTakeDown = "TakeDown"

	AppOpenresty         = "openresty"
	AppOpenclaw          = "openclaw"
	AppCopaw             = "copaw"
	AppHermesAgent       = "hermes-agent"
	AppDeepSeekHarness   = "deepseek-harness"
	AppOctop             = "octop"
	AppMysql             = "mysql"
	AppMariaDB           = "mariadb"
	AppPostgresql        = "postgresql"
	AppRedis             = "redis"
	AppPostgres          = "postgres"
	AppMongodb           = "mongodb"
	AppMemcached         = "memcached"
	AppMysqlCluster      = "mysql-cluster"
	AppPostgresqlCluster = "postgresql-cluster"
	AppRedisCluster      = "redis-cluster"

	AppResourceLocal  = "local"
	AppResourceRemote = "remote"
	AppResourceCustom = "custom"

	CPUS          = "CPUS"
	MemoryLimit   = "MEMORY_LIMIT"
	HostIP        = "HOST_IP"
	ContainerName = "CONTAINER_NAME"
)

func ClusterAppKeys() []string {
	return []string{AppMysqlCluster, AppPostgresqlCluster, AppRedisCluster}
}

func IsClusterApp(key string) bool {
	return slices.Contains(ClusterAppKeys(), key)
}

type AppOperate string

var (
	Start    AppOperate = "start"
	Stop     AppOperate = "stop"
	Restart  AppOperate = "restart"
	Delete   AppOperate = "delete"
	Sync     AppOperate = "sync"
	Backup   AppOperate = "backup"
	Update   AppOperate = "update"
	Rebuild  AppOperate = "rebuild"
	Upgrade  AppOperate = "upgrade"
	Reload   AppOperate = "reload"
	Favorite AppOperate = "favorite"
)
