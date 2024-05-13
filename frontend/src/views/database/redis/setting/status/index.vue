<template>
    <div v-if="statusShow" class="database-status">
        <el-form label-position="top">
            <span class="title">{{ $t('database.baseParam') }}</span>
            <el-divider class="divider" />
            <el-row class="content">
                <el-col :xs="8" :sm="6" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">uptime_in_days</span>
                        </template>
                        <span class="status-count">{{ redisStatus.uptime_in_days }}</span>
                        <span class="input-help">{{ $t('database.uptimeInDays') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="8" :sm="6" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">tcp_port</span>
                        </template>
                        <span class="status-count">{{ redisStatus.tcp_port }}</span>
                        <span class="input-help">{{ $t('database.tcpPort') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="8" :sm="6" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">connected_clients</span>
                        </template>
                        <span class="status-count">{{ redisStatus.connected_clients }}</span>
                        <span class="input-help">{{ $t('database.connectedClients') }}</span>
                    </el-form-item>
                </el-col>
            </el-row>

            <span class="title">{{ $t('database.performanceParam') }}</span>
            <el-divider class="divider" />
            <el-row class="content">
                <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">used_memory_rss</span>
                        </template>
                        <span class="status-count">{{ redisStatus.used_memory_rss }}</span>
                        <span class="input-help">{{ $t('database.usedMemoryRss') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">used_memory</span>
                        </template>
                        <span class="status-count">{{ redisStatus.used_memory }}</span>
                        <span class="input-help">{{ $t('database.usedMemory') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">used_memory_peak</span>
                        </template>
                        <span class="status-count">{{ redisStatus.used_memory_peak }}</span>
                        <span class="input-help">{{ $t('database.usedMemoryPeak') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">mem_fragmentation_ratio</span>
                        </template>
                        <span class="status-count">{{ redisStatus.mem_fragmentation_ratio }}</span>
                        <span class="input-help">{{ $t('database.tmpTableToDBHelper') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="8" :sm="12" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">total_connections_received</span>
                        </template>
                        <span class="status-count">{{ redisStatus.total_connections_received }}</span>
                        <span class="input-help">{{ $t('database.totalConnectionsReceived') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">total_commands_processed</span>
                        </template>
                        <span class="status-count">{{ redisStatus.total_commands_processed }}</span>
                        <span class="input-help">{{ $t('database.totalCommandsProcessed') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">instantaneous_ops_per_sec</span>
                        </template>
                        <span class="status-count">{{ redisStatus.instantaneous_ops_per_sec }}</span>
                        <span class="input-help">{{ $t('database.instantaneousOpsPerSec') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">keyspace_hits</span>
                        </template>
                        <span class="status-count">{{ redisStatus.keyspace_hits }}</span>
                        <span class="input-help">{{ $t('database.keyspaceHits') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">keyspace_misses</span>
                        </template>
                        <span class="status-count">{{ redisStatus.keyspace_misses }}</span>
                        <span class="input-help">{{ $t('database.keyspaceMisses') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">hit</span>
                        </template>
                        <span class="status-count">{{ redisStatus.hit }}</span>
                        <span class="input-help">{{ $t('database.hit') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">latest_fork_usec</span>
                        </template>
                        <span class="status-count">{{ redisStatus.latest_fork_usec }}</span>
                        <span class="input-help">{{ $t('database.latestForkUsec') }}</span>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
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

const database = ref();
const statusShow = ref(false);

interface DialogProps {
    database: string;
    status: string;
}
const acceptParams = (prop: DialogProps): void => {
    statusShow.value = true;
    database.value = prop.database;
    if (prop.status === 'Running') {
        loadStatus();
    }
};

const loadStatus = async () => {
    const res = await loadRedisStatus(database.value);
    let hit = (
        (Number(res.data.keyspace_hits) / (Number(res.data.keyspace_hits) + Number(res.data.keyspace_misses))) *
        100
    ).toFixed(2);

    redisStatus.uptime_in_days = res.data.uptime_in_days;
    redisStatus.tcp_port = res.data.tcp_port;
    redisStatus.connected_clients = res.data.connected_clients;
    redisStatus.used_memory_rss = (Number(res.data.used_memory_rss) / 1024 / 1024).toFixed(2) + ' MB';
    redisStatus.used_memory_peak = (Number(res.data.used_memory_peak) / 1024 / 1024).toFixed(2) + ' MB';
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
});
</script>

<style lang="scss" scoped>
.database-status {
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
    .content {
        margin-left: 50px;
    }
}
</style>
