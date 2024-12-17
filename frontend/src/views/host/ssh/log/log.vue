<template>
    <div>
        <LayoutContent v-loading="loading" :title="'SSH ' + $t('ssh.loginLogs', 2)">
            <template #prompt>
                <el-alert type="info" :title="$t('ssh.sshAlert2')" :closable="false" />
                <div class="mt-2"><el-alert type="info" :title="$t('ssh.sshAlert')" :closable="false" /></div>
            </template>
            <template #search>
                <el-select v-model="searchStatus" @change="search()" class="p-w-200">
                    <template #prefix>{{ $t('commons.table.status') }}</template>
                    <el-option :label="$t('commons.table.all')" value="All"></el-option>
                    <el-option :label="$t('commons.status.success')" value="Success"></el-option>
                    <el-option :label="$t('commons.status.failed')" value="Failed"></el-option>
                </el-select>
            </template>
            <template #toolbar>
                <div class="flex justify-between gap-2 flex-wrap sm:flex-row">
                    <div><!-- 占位 --></div>
                    <div class="flex flex-wrap gap-3">
                        <TableSetting @search="search()" />
                        <TableSearch @search="search()" v-model:searchName="searchInfo" />
                    </div>
                </div>
            </template>

            <template #main>
                <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search">
                    <el-table-column min-width="80" :label="$t('logs.loginIP')" prop="address" />
                    <el-table-column min-width="60" :label="$t('ssh.belong')" prop="area" />
                    <el-table-column min-width="60" :label="$t('commons.table.port')" prop="port" />
                    <el-table-column min-width="60" :label="$t('ssh.loginMode')" prop="authMode">
                        <template #default="{ row }">
                            <span v-if="row.authMode">{{ $t('ssh.' + row.authMode) }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column min-width="60" :label="$t('commons.table.user')" prop="user" />
                    <el-table-column min-width="60" :label="$t('logs.loginStatus')" prop="status">
                        <template #default="{ row }">
                            <div v-if="row.status === 'Success'">
                                <el-tag type="success">{{ $t('commons.status.success') }}</el-tag>
                            </div>
                            <div v-else>
                                <el-tooltip class="box-item" effect="dark" :content="row.message" placement="top-start">
                                    <el-tag type="danger">{{ $t('commons.status.failed') }}</el-tag>
                                </el-tooltip>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column
                        prop="date"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                        show-overflow-tooltip
                    />
                </ComplexTable>
            </template>
        </LayoutContent>
    </div>
</template>

<script setup lang="ts">
import { dateFormat } from '@/utils/util';
import { onMounted, reactive, ref } from 'vue';
import { loadSSHLogs } from '@/api/modules/host';

const loading = ref();
const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'ssh-log-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const searchInfo = ref();
const searchStatus = ref('All');

const search = async () => {
    let params = {
        info: searchInfo.value,
        status: searchStatus.value,
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    loading.value = true;
    await loadSSHLogs(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data?.logs || [];
            if (searchStatus.value === 'Success') {
                paginationConfig.total = res.data.successfulCount;
            }
            if (searchStatus.value === 'Failed') {
                paginationConfig.total = res.data.failedCount;
            }
            if (searchStatus.value === 'All') {
                paginationConfig.total = res.data.failedCount + res.data.successfulCount;
            }
        })
        .catch(() => {
            loading.value = false;
        });
};

onMounted(() => {
    search();
});
</script>
