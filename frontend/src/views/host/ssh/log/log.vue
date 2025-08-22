<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('ssh.loginLogs', 2)">
            <template #prompt>
                <el-alert type="info" :title="$t('ssh.sshAlert2')" :closable="false" />
                <div class="mt-2"><el-alert type="info" :title="$t('ssh.sshAlert')" :closable="false" /></div>
            </template>
            <template #leftToolBar>
                <el-button type="primary" @click="onExport">
                    {{ $t('commons.button.export') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <el-select v-model="searchStatus" @change="search()" class="p-w-200">
                    <template #prefix>{{ $t('commons.table.status') }}</template>
                    <el-option :label="$t('commons.table.all')" value="All"></el-option>
                    <el-option :label="$t('commons.status.success')" value="Success"></el-option>
                    <el-option :label="$t('commons.status.failed')" value="Failed"></el-option>
                </el-select>
                <TableSearch @search="search()" v-model:searchName="searchInfo" />
                <TableRefresh @search="search()" />
                <TableSetting title="ssh-log-refresh" @search="search()" />
            </template>

            <template #main>
                <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search" :heightDiff="400">
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

        <DialogPro v-model="open" :title="$t('commons.button.export')" size="mini">
            <el-form class="mt-5" ref="backupForm" @submit.prevent v-loading="loading">
                <el-form-item :label="$t('commons.table.status')">
                    <el-select v-model="exportConfig.status" class="w-full">
                        <el-option :label="$t('commons.table.all')" value="All"></el-option>
                        <el-option :label="$t('commons.status.success')" value="Success"></el-option>
                        <el-option :label="$t('commons.status.failed')" value="Failed"></el-option>
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('container.lines')">
                    <el-select class="tailClass" v-model.number="exportConfig.count">
                        <el-option :value="-1" :label="$t('commons.table.all')" />
                        <el-option :value="100" :label="100" />
                        <el-option :value="200" :label="200" />
                        <el-option :value="500" :label="500" />
                        <el-option :value="1000" :label="1000" />
                    </el-select>
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="open = false" :disabled="loading">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button type="primary" @click="onSubmitExport" :disabled="loading">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </DialogPro>
    </div>
</template>

<script setup lang="ts">
import { dateFormat, downloadFile } from '@/utils/util';
import { onMounted, reactive, ref } from 'vue';
import { exportSSHLogs, loadSSHLogs } from '@/api/modules/host';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const loading = ref();
const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'ssh-log-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const open = ref();
const exportConfig = reactive({
    count: 100,
    status: 'All',
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
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const onExport = async () => {
    open.value = true;
    exportConfig.status = 'All';
    exportConfig.count = -1;
};

const onSubmitExport = async () => {
    let params = {
        info: '',
        status: exportConfig.status,
        page: 1,
        pageSize: exportConfig.count,
    };
    await exportSSHLogs(params)
        .then((res) => {
            if (res.data) {
                downloadFile(res.data, globalStore.currentNode);
            }
            open.value = false;
        })
        .catch(() => {
            loading.value = false;
        });
};

onMounted(() => {
    search();
});
</script>
