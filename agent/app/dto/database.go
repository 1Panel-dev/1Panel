package dto

import "time"

// common
type DBConfUpdateByFile struct {
	Type     string `json:"type" validate:"required,oneof=mysql mariadb postgresql redis"`
	Database string `json:"database" validate:"required"`
	File     string `json:"file"`
}
type ChangeDBInfo struct {
	ID       uint   `json:"id"`
	From     string `json:"from" validate:"required,oneof=local remote"`
	Type     string `json:"type" validate:"required,oneof=mysql mariadb postgresql"`
	Database string `json:"database" validate:"required"`
	Value    string `json:"value" validate:"required"`
}

type DBBaseInfo struct {
	Name          string `json:"name"`
	ContainerName string `json:"containerName"`
	Port          int64  `json:"port"`
}

// mysql
type MysqlDBSearch struct {
	PageInfo
	Info     string `json:"info"`
	Database string `json:"database" validate:"required"`
	OrderBy  string `json:"orderBy" validate:"required,oneof=name created_at"`
	Order    string `json:"order" validate:"required,oneof=null ascending descending"`
}

type MysqlDBInfo struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	Name        string    `json:"name"`
	From        string    `json:"from"`
	MysqlName   string    `json:"mysqlName"`
	Format      string    `json:"format"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	Permission  string    `json:"permission"`
	IsDelete    bool      `json:"isDelete"`
	Description string    `json:"description"`
}

type MysqlOption struct {
	ID       uint   `json:"id"`
	From     string `json:"from"`
	Type     string `json:"type"`
	Database string `json:"database"`
	Name     string `json:"name"`
}

type MysqlDBCreate struct {
	Name        string `json:"name" validate:"required"`
	From        string `json:"from" validate:"required,oneof=local remote"`
	Database    string `json:"database" validate:"required"`
	Format      string `json:"format" validate:"required,oneof=utf8mb4 utf8 gbk big5"`
	Username    string `json:"username" validate:"required"`
	Password    string `json:"password" validate:"required"`
	Permission  string `json:"permission" validate:"required"`
	Description string `json:"description"`
}

type BindUser struct {
	Database   string `json:"database" validate:"required"`
	DB         string `json:"db" validate:"required"`
	Username   string `json:"username" validate:"required"`
	Password   string `json:"password" validate:"required"`
	Permission string `json:"permission" validate:"required"`
}

type MysqlLoadDB struct {
	From     string `json:"from" validate:"required,oneof=local remote"`
	Type     string `json:"type" validate:"required,oneof=mysql mariadb"`
	Database string `json:"database" validate:"required"`
}

type MysqlDBDeleteCheck struct {
	ID       uint   `json:"id" validate:"required"`
	Type     string `json:"type" validate:"required,oneof=mysql mariadb"`
	Database string `json:"database" validate:"required"`
}

type MysqlDBDelete struct {
	ID           uint   `json:"id" validate:"required"`
	Type         string `json:"type" validate:"required,oneof=mysql mariadb"`
	Database     string `json:"database" validate:"required"`
	ForceDelete  bool   `json:"forceDelete"`
	DeleteBackup bool   `json:"deleteBackup"`
}

type MysqlStatus struct {
	AbortedClients               string `json:"Aborted_clients"`
	AbortedConnects              string `json:"Aborted_connects"`
	BytesReceived                string `json:"Bytes_received"`
	BytesSent                    string `json:"Bytes_sent"`
	ComCommit                    string `json:"Com_commit"`
	ComRollback                  string `json:"Com_rollback"`
	Connections                  string `json:"Connections"`
	CreatedTmpDiskTables         string `json:"Created_tmp_disk_tables"`
	CreatedTmpTables             string `json:"Created_tmp_tables"`
	InnodbBufferPoolPagesDirty   string `json:"Innodb_buffer_pool_pages_dirty"`
	InnodbBufferPoolReadRequests string `json:"Innodb_buffer_pool_read_requests"`
	InnodbBufferPoolReads        string `json:"Innodb_buffer_pool_reads"`
	KeyReadRequests              string `json:"Key_read_requests"`
	KeyReads                     string `json:"Key_reads"`
	KeyWriteEequests             string `json:"Key_write_requests"`
	KeyWrites                    string `json:"Key_writes"`
	MaxUsedConnections           string `json:"Max_used_connections"`
	OpenTables                   string `json:"Open_tables"`
	OpenedFiles                  string `json:"Opened_files"`
	OpenedTables                 string `json:"Opened_tables"`
	QcacheHits                   string `json:"Qcache_hits"`
	QcacheInserts                string `json:"Qcache_inserts"`
	Questions                    string `json:"Questions"`
	SelectFullJoin               string `json:"Select_full_join"`
	SelectRangeCheck             string `json:"Select_range_check"`
	SortMergePasses              string `json:"Sort_merge_passes"`
	TableLocksWaited             string `json:"Table_locks_waited"`
	ThreadsCached                string `json:"Threads_cached"`
	ThreadsConnected             string `json:"Threads_connected"`
	ThreadsCreated               string `json:"Threads_created"`
	ThreadsRunning               string `json:"Threads_running"`
	Uptime                       string `json:"Uptime"`
	Run                          string `json:"Run"`
	File                         string `json:"File"`
	Position                     string `json:"Position"`
}

type MysqlVariables struct {
	BinlogCacheSize      string `json:"binlog_cache_size"`
	InnodbBufferPoolSize string `json:"innodb_buffer_pool_size"`
	InnodbLogBufferSize  string `json:"innodb_log_buffer_size"`
	JoinBufferSize       string `json:"join_buffer_size"`
	KeyBufferSize        string `json:"key_buffer_size"`
	MaxConnections       string `json:"max_connections"`
	MaxHeapTableSize     string `json:"max_heap_table_size"`
	QueryCacheSize       string `json:"query_cache_size"`
	QueryCacheType       string `json:"query_cache_type"`
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

type MysqlVariablesUpdate struct {
	Type      string                       `json:"type" validate:"required,oneof=mysql mariadb"`
	Database  string                       `json:"database" validate:"required"`
	Variables []MysqlVariablesUpdateHelper `json:"variables"`
}

type MysqlVariablesUpdateHelper struct {
	Param string      `json:"param"`
	Value interface{} `json:"value"`
}

// redis
type ChangeRedisPass struct {
	Database string `json:"database" validate:"required"`
	Value    string `json:"value"`
}

type RedisConfUpdate struct {
	Database   string `json:"database" validate:"required"`
	Timeout    string `json:"timeout"`
	Maxclients string `json:"maxclients"`
	Maxmemory  string `json:"maxmemory"`
}
type RedisConfPersistenceUpdate struct {
	Database    string `json:"database" validate:"required"`
	Type        string `json:"type" validate:"required,oneof=aof rbd"`
	Appendonly  string `json:"appendonly"`
	Appendfsync string `json:"appendfsync"`
	Save        string `json:"save"`
}

type RedisConf struct {
	Database      string `json:"database" validate:"required"`
	Name          string `json:"name"`
	Port          int64  `json:"port"`
	ContainerName string `json:"containerName"`
	Timeout       string `json:"timeout"`
	Maxclients    string `json:"maxclients"`
	Requirepass   string `json:"requirepass"`
	Maxmemory     string `json:"maxmemory"`
}

type RedisPersistence struct {
	Database    string `json:"database" validate:"required"`
	Appendonly  string `json:"appendonly"`
	Appendfsync string `json:"appendfsync"`
	Save        string `json:"save"`
}

type RedisStatus struct {
	Database                 string `json:"database" validate:"required"`
	TcpPort                  string `json:"tcp_port"`
	UptimeInDays             string `json:"uptime_in_days"`
	ConnectedClients         string `json:"connected_clients"`
	UsedMemory               string `json:"used_memory"`
	UsedMemoryRss            string `json:"used_memory_rss"`
	UsedMemoryPeak           string `json:"used_memory_peak"`
	MemFragmentationRatio    string `json:"mem_fragmentation_ratio"`
	TotalConnectionsReceived string `json:"total_connections_received"`
	TotalCommandsProcessed   string `json:"total_commands_processed"`
	InstantaneousOpsPerSec   string `json:"instantaneous_ops_per_sec"`
	KeyspaceHits             string `json:"keyspace_hits"`
	KeyspaceMisses           string `json:"keyspace_misses"`
	LatestForkUsec           string `json:"latest_fork_usec"`
}

type DatabaseFileRecords struct {
	Database  string `json:"database" validate:"required"`
	FileName  string `json:"fileName"`
	FileDir   string `json:"fileDir"`
	CreatedAt string `json:"createdAt"`
	Size      int    `json:"size"`
}
type RedisBackupRecover struct {
	Database string `json:"database" validate:"required"`
	FileName string `json:"fileName"`
	FileDir  string `json:"fileDir"`
}

// database
type DatabaseSearch struct {
	PageInfo
	Info    string `json:"info"`
	Type    string `json:"type"`
	OrderBy string `json:"orderBy" validate:"required,oneof=name created_at"`
	Order   string `json:"order" validate:"required,oneof=null ascending descending"`
}

type DatabaseInfo struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	Name      string    `json:"name" validate:"max=256"`
	From      string    `json:"from"`
	Type      string    `json:"type"`
	Version   string    `json:"version"`
	Address   string    `json:"address"`
	Port      uint      `json:"port"`
	Username  string    `json:"username"`
	Password  string    `json:"password"`

	SSL        bool   `json:"ssl"`
	RootCert   string `json:"rootCert"`
	ClientKey  string `json:"clientKey"`
	ClientCert string `json:"clientCert"`
	SkipVerify bool   `json:"skipVerify"`

	Description string `json:"description"`
}

type DatabaseOption struct {
	ID       uint   `json:"id"`
	Type     string `json:"type"`
	From     string `json:"from"`
	Database string `json:"database"`
	Version  string `json:"version"`
	Address  string `json:"address"`
}

type DatabaseItem struct {
	ID       uint   `json:"id"`
	From     string `json:"from"`
	Database string `json:"database"`
	Name     string `json:"name"`
}

type DatabaseCreate struct {
	Name     string `json:"name" validate:"required,max=256"`
	Type     string `json:"type" validate:"required"`
	From     string `json:"from" validate:"required,oneof=local remote"`
	Version  string `json:"version" validate:"required"`
	Address  string `json:"address"`
	Port     uint   `json:"port"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password"`

	SSL        bool   `json:"ssl"`
	RootCert   string `json:"rootCert"`
	ClientKey  string `json:"clientKey"`
	ClientCert string `json:"clientCert"`
	SkipVerify bool   `json:"skipVerify"`

	Description string `json:"description"`
}

type DatabaseUpdate struct {
	ID       uint   `json:"id"`
	Type     string `json:"type" validate:"required"`
	Version  string `json:"version" validate:"required"`
	Address  string `json:"address"`
	Port     uint   `json:"port"`
	Username string `json:"username" validate:"required"`
	Password string `json:"password"`

	SSL        bool   `json:"ssl"`
	RootCert   string `json:"rootCert"`
	ClientKey  string `json:"clientKey"`
	ClientCert string `json:"clientCert"`
	SkipVerify bool   `json:"skipVerify"`

	Description string `json:"description"`
}

type DatabaseDelete struct {
	ID           uint `json:"id" validate:"required"`
	ForceDelete  bool `json:"forceDelete"`
	DeleteBackup bool `json:"deleteBackup"`
}
