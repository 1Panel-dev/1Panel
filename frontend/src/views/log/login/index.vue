<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('logs.login')">
            <template #toolbar>
                <el-button type="primary" plain @click="onClean()">
                    {{ $t('logs.deleteLogs') }}
                </el-button>
            </template>
            <template #main>
                <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search">
                    <el-table-column min-width="40" :label="$t('logs.loginIP')" prop="ip" />
                    <el-table-column min-width="40" :label="$t('logs.loginAddress')" prop="address" />
                    <el-table-column :label="$t('logs.loginAgent')" show-overflow-tooltip prop="agent" />
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
import ComplexTable from '@/components/complex-table/index.vue';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import LayoutContent from '@/layout/layout-content.vue';
import { dateFormat } from '@/utils/util';
import { cleanLogs, getLoginLogs } from '@/api/modules/log';
import { onMounted, reactive, ref } from '@vue/runtime-core';
import i18n from '@/lang';
import { ElMessage } from 'element-plus';

const loading = ref();
const data = ref();
const confirmDialogRef = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 15,
    total: 0,
});

const search = async () => {
    let params = {
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
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
};

onMounted(() => {
    search();
});
</script>

<style scoped lang="scss">
.pre {
    white-space: pre-wrap;
    white-space: -moz-pre-wrap;
    white-space: -pre-wrap;
    white-space: -o-pre-wrap;
    word-wrap: break-word;
}
</style>
