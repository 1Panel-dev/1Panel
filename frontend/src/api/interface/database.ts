export namespace Database {
    export interface MysqlDBInfo {
        id: number;
        createdAt: Date;
        name: string;
        format: string;
        username: string;
        password: string;
        permission: string;
        permissionIPs: string;
        description: string;
    }
    export interface MysqlDBCreate {
        name: string;
        format: string;
        username: string;
        password: string;
        permission: string;
        permissionIPs: string;
        description: string;
    }
    export interface MysqlVariables {
        binlog_cache_size: number;
        innodb_buffer_pool_size: number;
        innodb_log_buffer_size: number;
        join_buffer_size: number;
        key_buffer_size: number;
        max_connections: number;
        max_heap_table_size: number;
        query_cache_size: number;
        query_cache_type: string;
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
}
