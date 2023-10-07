<template>
    <div>
        <el-form label-position="top">
            <span class="title">{{ $t('database.baseParam') }}</span>
            <el-divider class="divider" />
            <el-row type="flex" justify="center" style="margin-left: 50px" :gutter="20">
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">{{ $t('database.runTime') }}</span>
                        </template>
                        <span class="status-count">{{ mysqlStatus.run }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">{{ $t('database.connections') }}</span>
                        </template>
                        <span class="status-count">{{ mysqlStatus.connections }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">{{ $t('database.bytesSent') }}</span>
                        </template>
                        <span class="status-count">{{ mysqlStatus.bytesSent }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">{{ $t('database.bytesReceived') }}</span>
                        </template>
                        <span class="status-count">{{ mysqlStatus.bytesReceived }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">{{ $t('database.queryPerSecond') }}</span>
                        </template>
                        <span class="status-count">{{ mysqlStatus.queryPerSecond }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">{{ $t('database.txPerSecond') }}</span>
                        </template>
                        <span class="status-count">{{ mysqlStatus.txPerSecond }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">File</span>
                        </template>
                        <span class="status-count">{{ mysqlStatus.file }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">Position</span>
                        </template>
                        <span class="status-count">{{ mysqlStatus.position }}</span>
                    </el-form-item>
                </el-col>
            </el-row>

            <span class="title">{{ $t('database.performanceParam') }}</span>
            <el-divider class="divider" />
            <el-row type="flex" style="margin-left: 50px" justify="center" :gutter="20">
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">{{ $t('database.queryPerSecond') }}</span>
                        </template>
                        <span class="status-count">{{ mysqlStatus.queryPerSecond }}</span>
                        <span class="input-help">{{ $t('database.connInfoHelper') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">{{ $t('database.threadCacheHit') }}</span>
                        </template>
                        <span class="status-count">{{ mysqlStatus.threadCacheHit }}</span>
                        <span class="input-help">{{ $t('database.threadCacheHitHelper') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">{{ $t('database.indexHit') }}</span>
                        </template>
                        <span class="status-count">{{ mysqlStatus.indexHit }}</span>
                        <span class="input-help">{{ $t('database.indexHitHelper') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">{{ $t('database.innodbIndexHit') }}</span>
                        </template>
                        <span class="status-count">{{ mysqlStatus.innodbIndexHit }}</span>
                        <span class="input-help">{{ $t('database.innodbIndexHitHelper') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">{{ $t('database.cacheHit') }}</span>
                        </template>
                        <span class="status-count">{{ mysqlStatus.cacheHit }}</span>
                        <span class="input-help">{{ $t('database.cacheHitHelper') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">{{ $t('database.tmpTableToDB') }}</span>
                        </template>
                        <span class="status-count">{{ mysqlStatus.tmpTableToDB }}</span>
                        <span class="input-help">{{ $t('database.tmpTableToDBHelper') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">{{ $t('database.openTables') }}</span>
                        </template>
                        <span class="status-count">{{ mysqlStatus.openTables }}</span>
                        <span class="input-help">{{ $t('database.openTablesHelper') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">{{ $t('database.selectFullJoin') }}</span>
                        </template>
                        <span class="status-count">{{ mysqlStatus.selectFullJoin }}</span>
                        <span class="input-help">{{ $t('database.selectFullJoinHelper') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">{{ $t('database.selectRangeCheck') }}</span>
                        </template>
                        <span class="status-count">{{ mysqlStatus.selectRangeCheck }}</span>
                        <span class="input-help">{{ $t('database.selectRangeCheckHelper') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">{{ $t('database.sortMergePasses') }}</span>
                        </template>
                        <span class="status-count">{{ mysqlStatus.sortMergePasses }}</span>
                        <span class="input-help">{{ $t('database.sortMergePassesHelper') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">{{ $t('database.tableLocksWaited') }}</span>
                        </template>
                        <span class="status-count">{{ mysqlStatus.tableLocksWaited }}</span>
                        <span class="input-help">{{ $t('database.tableLocksWaitedHelper') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item style="width: 25%"></el-form-item>
                </el-col>
            </el-row>
        </el-form>
    </div>
</template>
<script lang="ts" setup>
import { loadMysqlStatus } from '@/api/modules/database';
import { computeSize } from '@/utils/util';
import { reactive } from 'vue';

let mysqlStatus = reactive({
    run: 0,
    connections: 0,
    bytesSent: '',
    bytesReceived: '',

    queryPerSecond: '',
    txPerSecond: '',
    file: '',
    position: 0,

    connInfo: '',
    threadCacheHit: '',
    indexHit: '',
    innodbIndexHit: '',
    cacheHit: '',
    tmpTableToDB: '',
    openTables: 0,
    selectFullJoin: 0,
    selectRangeCheck: 0,
    sortMergePasses: 0,
    tableLocksWaited: 0,
});

const currentDB = reactive({
    type: '',
    database: '',
});

interface DialogProps {
    type: string;
    database: string;
}

const acceptParams = (params: DialogProps): void => {
    currentDB.type = params.type;
    currentDB.database = params.database;
    loadStatus();
};

const loadStatus = async () => {
    const res = await loadMysqlStatus(currentDB.type, currentDB.database);
    let queryPerSecond = res.data.Questions / res.data.Uptime;
    let txPerSecond = (res.data!.Com_commit + res.data.Com_rollback) / res.data.Uptime;

    let threadCacheHit = (1 - res.data.Threads_created / res.data.Connections) * 100;
    let cacheHit = (res.data.Qcache_hits / (res.data.Qcache_hits + res.data.Qcache_inserts)) * 100;
    let indexHit = (1 - res.data.Key_reads / res.data.Key_read_requests) * 100;
    let innodbIndexHit = (1 - res.data.Innodb_buffer_pool_reads / res.data.Innodb_buffer_pool_read_requests) * 100;
    let tmpTableToDB = (res.data.Created_tmp_disk_tables / res.data.Created_tmp_tables) * 100;

    mysqlStatus.run = res.data.Run;
    mysqlStatus.connections = res.data.Connections;
    mysqlStatus.bytesSent = res.data.Bytes_sent ? computeSize(res.data.Bytes_sent) : '0';
    mysqlStatus.bytesReceived = res.data.Bytes_received ? computeSize(res.data.Bytes_received) : '0';

    mysqlStatus.queryPerSecond = isNaN(queryPerSecond) || queryPerSecond === 0 ? '0' : queryPerSecond.toFixed(2);
    mysqlStatus.txPerSecond = isNaN(txPerSecond) || txPerSecond === 0 ? '0' : txPerSecond.toFixed(2);
    mysqlStatus.file = res.data.File;
    mysqlStatus.position = res.data.Position;

    mysqlStatus.connInfo = res.data.Threads_running + '/' + res.data.Max_used_connections;
    mysqlStatus.threadCacheHit = isNaN(threadCacheHit) || threadCacheHit === 0 ? '0' : threadCacheHit.toFixed(2) + '%';
    mysqlStatus.indexHit = isNaN(indexHit) || indexHit === 0 ? '0' : indexHit.toFixed(2) + '%';
    mysqlStatus.innodbIndexHit = isNaN(innodbIndexHit) || innodbIndexHit === 0 ? '0' : innodbIndexHit.toFixed(2) + '%';
    mysqlStatus.cacheHit = isNaN(cacheHit) || cacheHit === 0 ? 'OFF' : cacheHit.toFixed(2) + '%';
    mysqlStatus.tmpTableToDB = isNaN(tmpTableToDB) || tmpTableToDB === 0 ? '0' : tmpTableToDB.toFixed(2) + '%';
    mysqlStatus.openTables = res.data.Open_tables;
    mysqlStatus.selectFullJoin = res.data.Select_full_join;
    mysqlStatus.selectRangeCheck = res.data.Select_range_check;
    mysqlStatus.sortMergePasses = res.data.Sort_merge_passes;
    mysqlStatus.tableLocksWaited = res.data.Table_locks_waited;
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.divider {
    display: block;
    height: 1px;
    width: 100%;
    margin: 12px 0;
    border-top: 1px var(--el-border-color) var(--el-border-style);
}
.title {
    font-size: 20px;
    font-weight: 500;
    margin-left: 50px;
}
</style>
