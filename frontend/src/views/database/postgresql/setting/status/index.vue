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
                        <span class="status-count">{{ postgresqlStatus.uptime }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">{{ $t('database.connections') }}</span>
                        </template>
                        <span class="status-count">{{ postgresqlStatus.max_connections }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">{{ $t('database.version') }}</span>
                        </template>
                        <span class="status-count">{{ postgresqlStatus.version }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">AUTOVACUUM</span>
                        </template>
                        <span class="status-count">{{ postgresqlStatus.autovacuum }}</span>
                    </el-form-item>
                </el-col>
            </el-row>

            <span class="title">{{ $t('database.performanceParam') }}</span>
            <el-divider class="divider" />
            <el-row type="flex" style="margin-left: 50px" justify="center" :gutter="20">
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">{{ $t('database.connInfo') }}</span>
                        </template>
                        <span class="status-count">{{ postgresqlStatus.current_connections }}</span>
                        <span class="input-help">{{ $t('database.connInfoHelper') }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">{{ $t('database.cacheHit') }}</span>
                        </template>
                        <span class="status-count">{{ postgresqlStatus.hit_ratio }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item style="width: 25%"></el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item style="width: 25%"></el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">SHARED_BUFFERS</span>
                        </template>
                        <span class="status-count">{{ postgresqlStatus.shared_buffers }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">BUFFERS_CLEAN</span>
                        </template>
                        <span class="status-count">{{ postgresqlStatus.buffers_clean }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">MAXWRITTEN_CLEAN</span>
                        </template>
                        <span class="status-count">{{ postgresqlStatus.maxwritten_clean }}</span>
                    </el-form-item>
                </el-col>
                <el-col :xs="24" :sm="24" :md="6" :lg="6" :xl="6">
                    <el-form-item>
                        <template #label>
                            <span class="status-label">BUFFERS_BACKEND_FSYNC</span>
                        </template>
                        <span class="status-count">{{ postgresqlStatus.buffers_backend_fsync }}</span>
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
    </div>
</template>
<script lang="ts" setup>
import { loadPostgresqlStatus } from '@/api/modules/database';
import { reactive } from 'vue';

const postgresqlStatus = reactive({
    uptime: '',
    version: '',
    max_connections: '',
    autovacuum: '',
    current_connections: '',
    hit_ratio: '',
    shared_buffers: '',
    buffers_clean: '',
    maxwritten_clean: '',
    buffers_backend_fsync: '',
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
    const res = await loadPostgresqlStatus(currentDB.type, currentDB.database);
    postgresqlStatus.uptime = res.data.uptime;
    postgresqlStatus.version = res.data.version;
    postgresqlStatus.max_connections = res.data.max_connections;
    postgresqlStatus.autovacuum = res.data.autovacuum;
    postgresqlStatus.current_connections = res.data.current_connections;
    postgresqlStatus.hit_ratio = res.data.hit_ratio;
    postgresqlStatus.shared_buffers = res.data.shared_buffers;
    postgresqlStatus.buffers_clean = res.data.buffers_clean;
    postgresqlStatus.maxwritten_clean = res.data.maxwritten_clean;
    postgresqlStatus.buffers_backend_fsync = res.data.buffers_backend_fsync;
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
