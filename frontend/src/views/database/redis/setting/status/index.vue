<template>
    <div v-if="statusShow">
        <el-row>
            <el-col :span="1"><br /></el-col>
            <el-col :span="12">
                <table style="margin-top: 20px; width: 100%" class="myTable">
                    <tr>
                        <td>uptime_in_days</td>
                        <td>{{ redisStatus!.uptime_in_days }}</td>
                        <td>{{ $t('database.uptimeInDays') }}</td>
                    </tr>
                    <tr>
                        <td>tcp_port</td>
                        <td>{{ redisStatus!.tcp_port }}</td>
                        <td>{{ $t('database.tcpPort') }}</td>
                    </tr>
                    <tr>
                        <td>connected_clients</td>
                        <td>{{ redisStatus!.connected_clients }}</td>
                        <td>{{ $t('database.connectedClients') }}</td>
                    </tr>
                    <tr>
                        <td>used_memory_rss</td>
                        <td>{{ redisStatus!.used_memory_rss }}</td>
                        <td>{{ $t('database.usedMemoryRss') }}</td>
                    </tr>
                    <tr>
                        <td>used_memory</td>
                        <td>{{ redisStatus!.used_memory }}</td>
                        <td>{{ $t('database.usedMemory') }}</td>
                    </tr>
                    <tr>
                        <td>mem_fragmentation_ratio</td>
                        <td>{{ redisStatus!.mem_fragmentation_ratio }}</td>
                        <td>{{ $t('database.tmpTableToDBHelper') }}</td>
                    </tr>
                    <tr>
                        <td>total_connections_received</td>
                        <td>{{ redisStatus!.total_connections_received }}</td>
                        <td>{{ $t('database.totalConnectionsReceived') }}</td>
                    </tr>
                    <tr>
                        <td>total_commands_processed</td>
                        <td>{{ redisStatus!.total_commands_processed }}</td>
                        <td>{{ $t('database.totalCommandsProcessed') }}</td>
                    </tr>
                    <tr>
                        <td>instantaneous_ops_per_sec</td>
                        <td>{{ redisStatus!.instantaneous_ops_per_sec }}</td>
                        <td>{{ $t('database.instantaneousOpsPerSec') }}</td>
                    </tr>
                    <tr>
                        <td>keyspace_hits</td>
                        <td>{{ redisStatus!.keyspace_hits }}</td>
                        <td>{{ $t('database.keyspaceHits') }}</td>
                    </tr>
                    <tr>
                        <td>keyspace_misses</td>
                        <td>{{ redisStatus!.keyspace_misses }}</td>
                        <td>{{ $t('database.keyspaceMisses') }}</td>
                    </tr>
                    <tr>
                        <td>hit</td>
                        <td>{{ redisStatus!.hit }}</td>
                        <td>{{ $t('database.hit') }}</td>
                    </tr>
                    <tr>
                        <td>latest_fork_usec</td>
                        <td>{{ redisStatus!.latest_fork_usec }}</td>
                        <td>{{ $t('database.latestForkUsec') }}</td>
                    </tr>
                </table>
            </el-col>
        </el-row>
    </div>
</template>

<script lang="ts" setup>
import { loadRedisStatus } from '@/api/modules/database';
import { reactive, ref } from 'vue';

const redisStatus = reactive({
    tcp_port: '',
    uptime_in_days: '',
    connected_clients: '',
    used_memory: '',
    used_memory_rss: '',
    used_memory_peak: '',
    mem_fragmentation_ratio: '',
    total_connections_received: '',
    total_commands_processed: '',
    instantaneous_ops_per_sec: '',
    keyspace_hits: '',
    keyspace_misses: '',
    hit: '',
    latest_fork_usec: '',
});

const statusShow = ref(false);

interface DialogProps {
    status: string;
}
const acceptParams = (prop: DialogProps): void => {
    statusShow.value = true;
    if (prop.status === 'Running') {
        loadStatus();
    }
};
const onClose = (): void => {
    statusShow.value = false;
};

const loadStatus = async () => {
    const res = await loadRedisStatus();
    let hit = (
        (Number(res.data.keyspace_hits) / (Number(res.data.keyspace_hits) + Number(res.data.keyspace_misses))) *
        100
    ).toFixed(2);

    redisStatus.uptime_in_days = res.data.uptime_in_days;
    redisStatus.tcp_port = res.data.tcp_port;
    redisStatus.connected_clients = res.data.connected_clients;
    redisStatus.used_memory_rss = (Number(res.data.used_memory_rss) / 1024 / 1024).toFixed(2) + ' MB';
    redisStatus.used_memory = (Number(res.data.used_memory) / 1024 / 1024).toFixed(2) + ' MB';
    redisStatus.mem_fragmentation_ratio = res.data.mem_fragmentation_ratio;
    redisStatus.total_connections_received = res.data.total_connections_received;
    redisStatus.total_commands_processed = res.data.total_commands_processed;
    redisStatus.instantaneous_ops_per_sec = res.data.instantaneous_ops_per_sec;
    redisStatus.keyspace_hits = res.data.keyspace_hits;
    redisStatus.keyspace_misses = res.data.keyspace_misses;
    redisStatus.hit = hit;
    redisStatus.latest_fork_usec = res.data.latest_fork_usec;
};

defineExpose({
    acceptParams,
    onClose,
});
</script>
