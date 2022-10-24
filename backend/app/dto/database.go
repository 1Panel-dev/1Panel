package dto

import "time"

type MysqlDBInfo struct {
	ID          uint      `json:"id"`
	CreatedAt   time.Time `json:"createdAt"`
	Name        string    `json:"name"`
	Format      string    `json:"format"`
	Username    string    `json:"username"`
	Password    string    `json:"password"`
	Permission  string    `json:"permission"`
	Description string    `json:"description"`
}

type MysqlDBCreate struct {
	Name        string `json:"name" validate:"required"`
	Version     string `json:"version" validate:"required,oneof=5.7.39 8.0.30"`
	Format      string `json:"format" validate:"required,oneof=utf8mb4 utf-8 gbk big5"`
	Username    string `json:"username" validate:"required"`
	Password    string `json:"password" validate:"required"`
	Permission  string `json:"permission" validate:"required"`
	Description string `json:"description"`
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
	Tmp_tableSize        string `json:"tmp_table_size"`
}

type ChangeDBInfo struct {
	ID        uint   `json:"id" validate:"required"`
	Operation string `json:"operation" validate:"required,oneof=password privilege"`
	Value     string `json:"value" validate:"required"`
}

type BatchDeleteByName struct {
	Names []string `json:"names" validate:"required"`
}
