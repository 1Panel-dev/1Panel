import { ReqPage } from '.';

export namespace Database {
    export interface SearchBackupRecord extends ReqPage {
        mysqlName: string;
        dbName: string;
    }
    export interface Backup {
        mysqlName: string;
        dbName: string;
    }
    export interface Recover {
        mysqlName: string;
        dbName: string;
        backupName: string;
    }
    export interface RecoverByUpload {
        mysqlName: string;
        dbName: string;
        fileName: string;
        fileDir: string;
    }
    export interface MysqlDBInfo {
        id: number;
        createdAt: Date;
        name: string;
        format: string;
        username: string;
        password: string;
        permission: string;
        description: string;
    }
    export interface BaseInfo {
        name: string;
        port: number;
        password: string;
        remoteConn: boolean;
        mysqlKey: string;
        containerName: string;
    }
    export interface MysqlConfUpdateByFile {
        mysqlName: string;
        file: string;
    }
    export interface MysqlDBCreate {
        name: string;
        format: string;
        username: string;
        password: string;
        permission: string;
        description: string;
    }
    export interface MysqlVariables {
        mysqlName: string;
        binlog_cache_size: number;
        innodb_buffer_pool_size: number;
        innodb_log_buffer_size: number;
        join_buffer_size: number;
        key_buffer_size: number;
        max_connections: number;
        query_cache_size: number;
        read_buffer_size: number;
        read_rnd_buffer_size: number;
        sort_buffer_size: number;
        table_open_cache: number;
        thread_cache_size: number;
        thread_stack: number;
        tmp_table_size: number;

        slow_query_log: string;
        long_query_time: number;
    }
    export interface VariablesUpdate {
        param: string;
        value: any;
    }
    export interface MysqlStatus {
        Aborted_clients: number;
        Aborted_connects: number;
        Bytes_received: number;
        Bytes_sent: number;
        Com_commit: number;
        Com_rollback: number;
        Connections: number;
        Created_tmp_disk_tables: number;
        Created_tmp_tables: number;
        Innodb_buffer_pool_pages_dirty: number;
        Innodb_buffer_pool_read_requests: number;
        Innodb_buffer_pool_reads: number;
        Key_read_requests: number;
        Key_reads: number;
        Key_write_requests: number;
        Key_writes: number;
        Max_used_connections: number;
        Open_tables: number;
        Opened_files: number;
        Opened_tables: number;
        Qcache_hits: number;
        Qcache_inserts: number;
        Questions: number;
        Select_full_join: number;
        Select_range_check: number;
        Sort_merge_passes: number;
        Table_locks_waited: number;
        Threads_cached: number;
        Threads_connected: number;
        Threads_created: number;
        Threads_running: number;
        Uptime: number;
        Run: number;
        File: string;
        Position: number;
    }
    export interface ChangeInfo {
        id: number;
        operation: string;
        value: string;
    }

    // redis
    export interface RedisConfUpdate {
        timeout: string;
        maxclients: string;
        requirepass: string;
        maxmemory: string;
    }
    export interface RedisConfPersistenceUpdate {
        type: string;
        appendonly: string;
        appendfsync: string;
        save: string;
    }
    export interface RedisConfUpdateByFile {
        file: string;
        restartNow: boolean;
    }
    export interface RedisStatus {
        tcp_port: string;
        uptime_in_days: string;
        connected_clients: string;
        used_memory: string;
        used_memory_rss: string;
        used_memory_peak: string;
        mem_fragmentation_ratio: string;
        total_connections_received: string;
        total_commands_processed: string;
        instantaneous_ops_per_sec: string;
        keyspace_hits: string;
        keyspace_misses: string;
        latest_fork_usec: string;
    }
    export interface RedisConf {
        name: string;
        port: number;
        timeout: number;
        maxclients: number;
        requirepass: string;
        maxmemory: number;
    }
    export interface RedisPersistenceConf {
        appendonly: string;
        appendfsync: string;
        save: string;
    }
    export interface FileRecord {
        fileName: string;
        fileDir: string;
        createdAt: string;
        size: string;
    }
    export interface RedisRecover {
        fileName: string;
        fileDir: string;
    }
}
