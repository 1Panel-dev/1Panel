<template>
    <div>
        <el-row :gutter="20">
            <el-col :span="1"><br /></el-col>
            <el-col :span="6">
                <table style="width: 100%" class="myTable">
                    <tr>
                        <td>{{ $t('database.runTime') }}</td>
                        <td>{{ mysqlStatus.run }}</td>
                    </tr>
                    <tr>
                        <td>{{ $t('database.connections') }}</td>
                        <td>{{ mysqlStatus.connections }}</td>
                    </tr>
                    <tr>
                        <td>{{ $t('database.bytesSent') }}</td>
                        <td>{{ mysqlStatus!.bytesSent }}</td>
                    </tr>
                    <tr>
                        <td>{{ $t('database.bytesReceived') }}</td>
                        <td>{{ mysqlStatus!.bytesReceived }}</td>
                    </tr>
                </table>
            </el-col>
            <el-col :span="6">
                <table style="width: 100%" class="myTable">
                    <tr>
                        <td>{{ $t('database.queryPerSecond') }}</td>
                        <td>{{ mysqlStatus!.queryPerSecond }}</td>
                    </tr>
                    <tr>
                        <td>{{ $t('database.queryPerSecond') }}</td>
                        <td>{{ mysqlStatus!.txPerSecond }}</td>
                    </tr>
                    <tr>
                        <td>File</td>
                        <td>{{ mysqlStatus!.file }}</td>
                    </tr>
                    <tr>
                        <td>Position</td>
                        <td>{{ mysqlStatus!.position }}</td>
                    </tr>
                </table>
            </el-col>
        </el-row>
        <el-row>
            <el-col :span="1"><br /></el-col>
            <el-col :span="12">
                <table style="margin-top: 20px; width: 100%" class="myTable">
                    <tr>
                        <td>{{ $t('database.queryPerSecond') }}</td>
                        <td>{{ mysqlStatus!.connInfo }}</td>
                        <td>{{ $t('database.connInfoHelper') }}</td>
                    </tr>
                    <tr>
                        <td>{{ $t('database.threadCacheHit') }}</td>
                        <td>{{ mysqlStatus!.threadCacheHit }}</td>
                        <td>{{ $t('database.threadCacheHitHelper') }}</td>
                    </tr>
                    <tr>
                        <td>{{ $t('database.indexHit') }}</td>
                        <td>{{ mysqlStatus!.indexHit }}</td>
                        <td>{{ $t('database.indexHitHelper') }}</td>
                    </tr>
                    <tr>
                        <td>{{ $t('database.innodbIndexHit') }}</td>
                        <td>{{ mysqlStatus!.innodbIndexHit }}</td>
                        <td>{{ $t('database.innodbIndexHitHelper') }}</td>
                    </tr>
                    <tr>
                        <td>{{ $t('database.cacheHit') }}</td>
                        <td>{{ mysqlStatus!.cacheHit }}</td>
                        <td>{{ $t('database.cacheHitHelper') }}</td>
                    </tr>
                    <tr>
                        <td>{{ $t('database.tmpTableToDB') }}</td>
                        <td>{{ mysqlStatus!.tmpTableToDB }}</td>
                        <td>{{ $t('database.tmpTableToDBHelper') }}</td>
                    </tr>
                    <tr>
                        <td>{{ $t('database.openTables') }}</td>
                        <td>{{ mysqlStatus!.openTables }}</td>
                        <td>{{ $t('database.openTablesHelper') }}</td>
                    </tr>
                    <tr>
                        <td>{{ $t('database.selectFullJoin') }}</td>
                        <td>{{ mysqlStatus!.selectFullJoin }}</td>
                        <td>{{ $t('database.selectFullJoinHelper') }}</td>
                    </tr>
                    <tr>
                        <td>{{ $t('database.selectRangeCheck') }}</td>
                        <td>{{ mysqlStatus!.selectRangeCheck }}</td>
                        <td>{{ $t('database.selectRangeCheckHelper') }}</td>
                    </tr>
                    <tr>
                        <td>{{ $t('database.sortMergePasses') }}</td>
                        <td>{{ mysqlStatus!.sortMergePasses }}</td>
                        <td>{{ $t('database.sortMergePassesHelper') }}</td>
                    </tr>
                    <tr>
                        <td>{{ $t('database.tableLocksWaited') }}</td>
                        <td>{{ mysqlStatus!.tableLocksWaited }}</td>
                        <td>{{ $t('database.tableLocksWaitedHelper') }}</td>
                    </tr>
                </table>
            </el-col>
        </el-row>
    </div>
</template>
<script lang="ts" setup>
import { loadMysqlStatus } from '@/api/modules/database';
import { computeSize } from '@/utils/util';
import { reactive, ref } from 'vue';

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

const mysqlName = ref();
interface DialogProps {
    mysqlName: string;
}
const acceptParams = (params: DialogProps): void => {
    mysqlName.value = params.mysqlName;
    loadStatus();
};

const loadStatus = async () => {
    const res = await loadMysqlStatus(mysqlName.value);
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
