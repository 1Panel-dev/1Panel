import { ReqPage } from '.';

export namespace Database {
    export interface SearchDBWithPage {
        info: string;
        database: string;
        page: number;
        pageSize: number;
        orderBy?: string;
        order?: string;
    }
    export interface SearchBackupRecord extends ReqPage {
        mysqlName: string;
        dbName: string;
    }
    export interface MysqlDBInfo {
        id: number;
        createdAt: Date;
        name: string;
        mysqlName: string;
        from: string;
        format: string;
        username: string;
        password: string;
        permission: string;
        isDelete: string;
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
    export interface DBConfUpdate {
        type: string;
        database: string;
        file: string;
    }
    export interface MysqlDBCreate {
        name: string;
        from: string;
        database: string;
        format: string;
        username: string;
        password: string;
        permission: string;
        description: string;
    }

    export interface BindUser {
        database: string;
        db: string;
        username: string;
        password: string;
        permission: string;
    }

    export interface MysqlLoadDB {
        from: string;
        type: string;
        database: string;
    }
    export interface MysqlDBDeleteCheck {
        id: number;
        type: string;
        database: string;
    }

    export interface MysqlDBDelete {
        id: number;
        type: string;
        database: string;
        forceDelete: boolean;
        deleteBackup: boolean;
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
        type: string;
        database: string;
        variables: Array<VariablesUpdateHelper>;
    }
    export interface VariablesUpdateHelper {
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
    export interface PgLoadDB {
        from: string;
        type: string;
        database: string;
    }
    export interface PgBind {
        name: string;
        database: string;
        username: string;
        password: string;
        superUser: boolean;
    }
    export interface PgChangePrivileges {
        name: string;
        database: string;
        username: string;
        superUser: boolean;
    }
    export interface PostgresqlDBDelete {
        id: number;
        type: string;
        database: string;
        forceDelete: boolean;
        deleteBackup: boolean;
    }
    export interface PostgresqlDBDeleteCheck {
        id: number;
        type: string;
        database: string;
    }
    export interface PostgresqlDBInfo {
        id: number;
        createdAt: Date;
        name: string;
        postgresqlName: string;
        from: string;
        format: string;
        username: string;
        password: string;
        description: string;
    }
    export interface PostgresqlConfUpdateByFile {
        type: string;
        database: string;
        file: string;
    }
    export interface PostgresqlDBCreate {
        name: string;
        from: string;
        database: string;
        format: string;
        username: string;
        password: string;
        superUser: boolean;
        description: string;
    }
    export interface PostgresqlDBInfo {
        id: number;
        createdAt: Date;
        name: string;
        mysqlName: string;
        from: string;
        format: string;
        username: string;
        password: string;
        superUser: boolean;
        isDelete: string;
        description: string;
    }
    export interface ChangeInfo {
        id: number;
        from: string;
        type: string;
        database: string;
        value: string;
    }

    // redis
    export interface RedisConfUpdate {
        database: string;
        timeout: string;
        maxclients: string;
        maxmemory: string;
    }
    export interface RedisConfPersistenceUpdate {
        database: string;
        type: string;
        appendonly: string;
        appendfsync: string;
        save: string;
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
        maxmemory: string;
    }
    export interface RedisPersistenceConf {
        appendonly: string;
        appendfsync: string;
        save: string;
    }

    // remote
    export interface DatabaseInfo {
        id: number;
        createdAt: Date;
        name: string;
        type: string;
        version: string;
        from: string;
        address: string;
        port: number;
        username: string;
        password: string;

        ssl: boolean;
        hasCA: boolean;
        rootCert: string;
        clientKey: string;
        clientCert: string;
        skipVerify: boolean;

        description: string;
    }
    export interface SearchDatabasePage {
        info: string;
        type: string;
        page: number;
        pageSize: number;
        orderBy?: string;
        order?: string;
    }
    export interface DatabaseOption {
        id: number;
        from: string;
        type: string;
        database: string;
        version: string;
        address: string;
    }
    export interface DbItem {
        id: number;
        from: string;
        database: string;
        name: string;
    }
    export interface DatabaseCreate {
        name: string;
        version: string;
        from: string;
        address: string;
        port: number;
        username: string;
        password: string;

        ssl: boolean;
        rootCert: string;
        clientKey: string;
        clientCert: string;
        skipVerify: boolean;

        description: string;
    }
    export interface DatabaseUpdate {
        id: number;
        version: string;
        address: string;
        port: number;
        username: string;
        password: string;

        ssl: boolean;
        rootCert: string;
        clientKey: string;
        clientCert: string;
        skipVerify: boolean;

        description: string;
    }
    export interface DatabaseDelete {
        id: number;
        forceDelete: boolean;
        deleteBackup: boolean;
    }
}
