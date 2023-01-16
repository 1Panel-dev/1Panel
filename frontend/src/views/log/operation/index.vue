<template>
    <div>
        <Submenu activeName="operation" />
        <el-card style="margin-top: 20px">
            <LayoutContent :header="$t('logs.operation')">
                <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search">
                    <template #toolbar>
                        <el-button type="primary" @click="onClean()">
                            {{ $t('logs.deleteLogs') }}
                        </el-button>
                    </template>

                    <el-table-column :label="$t('logs.resource')" prop="group" fix>
                        <template #default="{ row }">
                            {{ $t('logs.detail.' + row.group) }}
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('logs.operate')" min-width="150px" prop="detailZH" />
                    <el-table-column :label="$t('logs.status')" prop="status">
                        <template #default="{ row }">
                            <el-tag v-if="row.status === 'Success'" class="ml-2" type="success">
                                {{ $t('commons.status.success') }}
                            </el-tag>
                            <div v-else>
                                <el-popover
                                    placement="top-start"
                                    :title="$t('commons.table.message')"
                                    :width="400"
                                    trigger="hover"
                                    :content="row.message"
                                >
                                    <template #reference>
                                        <el-tag class="ml-2" type="danger">{{ $t('commons.status.failed') }}</el-tag>
                                    </template>
                                </el-popover>
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
            </LayoutContent>
        </el-card>

        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmitClean"></ConfirmDialog>
    </div>
</template>

<script setup lang="ts">
import ComplexTable from '@/components/complex-table/index.vue';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { dateFormat } from '@/utils/util';
import LayoutContent from '@/layout/layout-content.vue';
import { cleanLogs, getOperationLogs } from '@/api/modules/log';
import Submenu from '@/views/log/index.vue';
import { onMounted, reactive, ref } from '@vue/runtime-core';
import i18n from '@/lang';
import { ElMessage } from 'element-plus';

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
    const res = await getOperationLogs(params);
    data.value = res.data.items;
    paginationConfig.total = res.data.total;
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
    await cleanLogs({ logType: 'operation' });
    search();
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
};

onMounted(() => {
    search();
});
</script>
