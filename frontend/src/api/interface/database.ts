import { ReqPage } from '.';

export namespace Database {
    export interface Search extends ReqPage {
        mysqlName: string;
    }
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
    }
    export interface MysqlDBCreate {
        name: string;
        mysqlName: string;
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
        mysqlName: string;
        operation: string;
        value: string;
    }

    // redis
    export interface SearchRedisWithPage extends ReqPage {
        redisName: string;
        db: number;
    }
    export interface RedisBaseReq {
        redisName: string;
        db: number;
    }
    export interface RedisConfUpdate {
        redisName: string;
        db: number;
        paramName: string;
        value: string;
    }
    export interface RedisData {
        redisName: string;
        db: number;
        key: string;
        value: string;
        type: string;
        length: number;
        expiration: number;
    }
    export interface RedisDataSet {
        redisName: string;
        db: number;
        key: string;
        value: string;
        expiration: number;
    }
    export interface RedisDelBatch {
        redisName: string;
        db: number;
        names: Array<string>;
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
        timeout: number;
        maxclients: number;
        databases: number;
        requirepass: string;
        maxmemory: number;

        dir: string;
        appendonly: string;
        appendfsync: string;
        save: string;
    }
}
