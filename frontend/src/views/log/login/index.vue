<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('logs.login')">
            <template #search>
                <LogRouter current="LoginLog" />
            </template>
            <template #leftToolBar>
                <el-button type="primary" plain @click="onClean()">
                    {{ $t('logs.deleteLogs') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <el-select v-model="searchStatus" @change="search()" clearable class="p-w-200">
                    <template #prefix>{{ $t('commons.table.status') }}</template>
                    <el-option :label="$t('commons.table.all')" value=""></el-option>
                    <el-option :label="$t('commons.status.success')" value="Success"></el-option>
                    <el-option :label="$t('commons.status.failed')" value="Failed"></el-option>
                </el-select>
                <TableSearch @search="search()" v-model:searchName="searchIP" />
                <TableRefresh @search="search()" />
                <TableSetting title="login-log-refresh" @search="search()" />
            </template>
            <template #main>
                <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search" :heightDiff="370">
                    <el-table-column :label="$t('logs.loginIP')" prop="ip" />
                    <el-table-column :label="$t('logs.loginAddress')" prop="address" />
                    <el-table-column :label="$t('logs.loginAgent')" show-overflow-tooltip prop="agent" />
                    <el-table-column :label="$t('logs.loginStatus')" prop="status">
                        <template #default="{ row }">
                            <Status :status="row.status" :msg="row.message" />
                        </template>
                    </el-table-column>
                    <el-table-column
                        prop="createdAt"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                        show-overflow-tooltip
                    />
                </ComplexTable>
            </template>
        </LayoutContent>
        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmitClean"></ConfirmDialog>
    </div>
</template>

<script setup lang="ts">
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import LogRouter from '@/views/log/router/index.vue';
import { dateFormat } from '@/utils/util';
import { cleanLogs, getLoginLogs } from '@/api/modules/log';
import { onMounted, reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const loading = ref();
const data = ref();
const confirmDialogRef = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'login-log-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const searchIP = ref<string>('');
const searchStatus = ref<string>('');

const search = async () => {
    let params = {
        ip: searchIP.value,
        status: searchStatus.value,
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    loading.value = true;
    await getLoginLogs(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items;
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const onClean = async () => {
    let params = {
        header: i18n.global.t('logs.deleteLogs'),
        operationInfo: i18n.global.t('commons.msg.delete'),
        submitInputInfo: i18n.global.t('logs.deleteLogs'),
    };
    confirmDialogRef.value!.acceptParams(params);
};

const onSubmitClean = async () => {
    await cleanLogs({ logType: 'login' });
    search();
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
};

onMounted(() => {
    search();
});
</script>
