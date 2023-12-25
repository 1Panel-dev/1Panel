package dto

import "time"

type PostgresqlDBSearch struct {
	PageInfo
	Info     string `json:"info"`
	Database string `json:"database" validate:"required"`
	OrderBy  string `json:"orderBy"`
	Order    string `json:"order"`
}

type PostgresqlDBInfo struct {
	ID             uint      `json:"id"`
	CreatedAt      time.Time `json:"createdAt"`
	Name           string    `json:"name"`
	From           string    `json:"from"`
	PostgresqlName string    `json:"postgresqlName"`
	Format         string    `json:"format"`
	Username       string    `json:"username"`
	Password       string    `json:"password"`
	Permission     string    `json:"permission"`
	BackupCount    int       `json:"backupCount"`
	Description    string    `json:"description"`
}

type PostgresqlOption struct {
	ID       uint   `json:"id"`
	From     string `json:"from"`
	Type     string `json:"type"`
	Database string `json:"database"`
	Name     string `json:"name"`
}

type PostgresqlDBCreate struct {
	Name        string `json:"name" validate:"required"`
	From        string `json:"from" validate:"required,oneof=local remote"`
	Database    string `json:"database" validate:"required"`
	Format      string `json:"format"`
	Username    string `json:"username" validate:"required"`
	Password    string `json:"password" validate:"required"`
	Permission  string `json:"permission" validate:"required"`
	Description string `json:"description"`
}

type PostgresqlLoadDB struct {
	From     string `json:"from" validate:"required,oneof=local remote"`
	Type     string `json:"type" validate:"required,oneof=postgresql"`
	Database string `json:"database" validate:"required"`
}

type PostgresqlDBDeleteCheck struct {
	ID       uint   `json:"id" validate:"required"`
	Type     string `json:"type" validate:"required,oneof=postgresql"`
	Database string `json:"database" validate:"required"`
}

type PostgresqlDBDelete struct {
	ID           uint   `json:"id" validate:"required"`
	Type         string `json:"type" validate:"required,oneof=postgresql"`
	Database     string `json:"database" validate:"required"`
	ForceDelete  bool   `json:"forceDelete"`
	DeleteBackup bool   `json:"deleteBackup"`
}

type PostgresqlStatus struct {
	Uptime              string `json:"uptime"`
	Version             string `json:"version"`
	MaxConnections      string `json:"max_connections"`
	Autovacuum          string `json:"autovacuum"`
	CurrentConnections  string `json:"current_connections"`
	HitRatio            string `json:"hit_ratio"`
	SharedBuffers       string `json:"shared_buffers"`
	BuffersClean        string `json:"buffers_clean"`
	MaxwrittenClean     string `json:"maxwritten_clean"`
	BuffersBackendFsync string `json:"buffers_backend_fsync"`
}

type PostgresqlVariables struct {
	BinlogCachSize       string `json:"binlog_cache_size"`
	InnodbBufferPoolSize string `json:"innodb_buffer_pool_size"`
	InnodbLogBufferSize  string `json:"innodb_log_buffer_size"`
	JoinBufferSize       string `json:"join_buffer_size"`
	KeyBufferSize        string `json:"key_buffer_size"`
	MaxConnections       string `json:"max_connections"`
	MaxHeapTableSize     string `json:"max_heap_table_size"`
	QueryCacheSize       string `json:"query_cache_size"`
	QueryCache_type      string `json:"query_cache_type"`
	ReadBufferSize       string `json:"read_buffer_size"`
	ReadRndBufferSize    string `json:"read_rnd_buffer_size"`
	SortBufferSize       string `json:"sort_buffer_size"`
	TableOpenCache       string `json:"table_open_cache"`
	ThreadCacheSize      string `json:"thread_cache_size"`
	ThreadStack          string `json:"thread_stack"`
	TmpTableSize         string `json:"tmp_table_size"`

	SlowQueryLog  string `json:"slow_query_log"`
	LongQueryTime string `json:"long_query_time"`
}

type PostgresqlVariablesUpdate struct {
	Type      string                            `json:"type" validate:"required,oneof=postgresql"`
	Database  string                            `json:"database" validate:"required"`
	Variables []PostgresqlVariablesUpdateHelper `json:"variables"`
}

type PostgresqlVariablesUpdateHelper struct {
	Param string      `json:"param"`
	Value interface{} `json:"value"`
}
type PostgresqlConfUpdateByFile struct {
	Type     string `json:"type" validate:"required,oneof=postgresql mariadb"`
	Database string `json:"database" validate:"required"`
	File     string `json:"file"`
}
