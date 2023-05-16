<template>
    <div>
        <FireRouter />

        <LayoutContent v-loading="loading" :title="$t('ssh.loginLogs')">
            <template #toolbar>
                <el-row>
                    <el-col :span="16">
                        <el-tag type="success">{{ $t('commons.status.success') }}： {{ successfulCount }}</el-tag>
                        <el-tag type="danger" style="margin-left: 5px">
                            {{ $t('commons.status.failed') }}： {{ faliedCount }}
                        </el-tag>
                    </el-col>
                    <el-col :span="8">
                        <TableSetting @search="search()" />
                        <div class="search-button">
                            <el-input
                                v-model="searchInfo"
                                clearable
                                @clear="search()"
                                suffix-icon="Search"
                                @keyup.enter="search()"
                                @change="search()"
                                :placeholder="$t('commons.button.search')"
                            ></el-input>
                        </div>
                    </el-col>
                </el-row>
            </template>

            <template #search>
                <el-select v-model="searchStatus" @change="search()" clearable>
                    <template #prefix>{{ $t('commons.table.status') }}</template>
                    <el-option :label="$t('commons.table.all')" value="All"></el-option>
                    <el-option :label="$t('commons.status.success')" value="Success"></el-option>
                    <el-option :label="$t('commons.status.failed')" value="Failed"></el-option>
                </el-select>
            </template>
            <template #main>
                <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search">
                    <el-table-column min-width="40" :label="$t('logs.loginIP')" prop="ip">
                        <template #default="{ row }">{{ row.address }}:{{ row.port }}</template>
                    </el-table-column>
                    <el-table-column min-width="40" :label="$t('ssh.loginMode')" prop="authMode">
                        <template #default="{ row }">{{ $t('ssh.' + row.authMode) }}</template>
                    </el-table-column>
                    <el-table-column min-width="40" :label="$t('ssh.loginUser')" prop="user" />
                    <el-table-column min-width="40" :label="$t('logs.loginStatus')" prop="status">
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
import FireRouter from '@/views/host/ssh/index.vue';
import ComplexTable from '@/components/complex-table/index.vue';
import TableSetting from '@/components/table-setting/index.vue';
import LayoutContent from '@/layout/layout-content.vue';
import { dateFormat } from '@/utils/util';
import { onMounted, reactive, ref } from '@vue/runtime-core';
import { loadSSHLogs } from '@/api/modules/host';

const loading = ref();
const data = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const searchInfo = ref();
const searchStatus = ref('All');
const successfulCount = ref(0);
const faliedCount = ref(0);

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
            data.value = res.data.logs || [];
            faliedCount.value = res.data.failedCount;
            successfulCount.value = res.data.successfulCount;
            paginationConfig.total = res.data.failedCount + res.data.successfulCount;
        })
        .catch(() => {
            loading.value = false;
        });
};

onMounted(() => {
    search();
});
</script>
