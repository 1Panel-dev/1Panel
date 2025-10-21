<template>
    <div>
        <RouterMenu />
        <LayoutContent>
            <template #leftToolBar>
                <div class="flex flex-wrap gap-3">
                    <el-button type="primary" @click="openCreate">
                        {{ $t('commons.button.create') }}
                    </el-button>
                </div>
            </template>
            <template #rightToolBar>
                <TableRefresh @search="search()" />
            </template>
            <template #main>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    :data="data"
                    @search="search"
                    v-loading="loading"
                >
                    <el-table-column
                        :label="$t('commons.table.name')"
                        min-width="120"
                        prop="name"
                        show-overflow-tooltip
                    />
                    <el-table-column :label="$t('app.version')" min-width="100" prop="version" show-overflow-tooltip />
                    <el-table-column :label="$t('commons.table.port')" min-width="80" prop="port">
                        <template #default="{ row }">
                            <PortJump :row="row" :jump="goDashboard" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.status')" min-width="100" prop="status">
                        <template #default="{ row }">
                            <Status :key="row.status" :status="row.status"></Status>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.button.log')" width="120px">
                        <template #default="{ row }">
                            <el-button
                                @click="openLog(row)"
                                link
                                type="primary"
                                :disabled="
                                    row.status !== 'Running' && row.status !== 'Rrror' && row.status !== 'Restarting'
                                "
                            >
                                {{ $t('website.check') }}
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column
                        prop="createdAt"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                        show-overflow-tooltip
                        width="180"
                        fix
                    />
                    <fu-table-operations
                        :ellipsis="mobile ? 0 : 5"
                        :min-width="mobile ? 'auto' : 300"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        fixed="right"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>
        <OpDialog ref="opRef" @search="search" />
        <OperateDialog @search="search" ref="dialogRef" />
        <ComposeLogs ref="composeLogRef" />
        <PortJumpDialog ref="dialogPortJumpRef" />
    </div>
</template>

<script lang="ts" setup>
import OperateDialog from './operate/index.vue';
import RouterMenu from '@/views/ai/model/index.vue';
import ComposeLogs from '@/components/log/compose/index.vue';
import PortJumpDialog from '@/components/port-jump/index.vue';
import PortJump from '@/views/website/runtime/components/port-jump.vue';

import { reactive, onMounted, ref } from 'vue';
import { dateFormat } from '@/utils/util';
import { AI } from '@/api/interface/ai';
import { deleteTensorRTLLM, operateTensorRTLLM, pageTensorRTLLM } from '@/api/modules/ai';
import { ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const mobile = computed(() => {
    return globalStore.isMobile();
});

const loading = ref();
const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const searchName = ref();
const opRef = ref();
const dialogRef = ref();
const composeLogRef = ref();
const dialogPortJumpRef = ref();

const search = async () => {
    const params = {
        name: searchName.value,
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    loading.value = true;
    await pageTensorRTLLM(params)
        .then((res) => {
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total || 0;
        })
        .finally(() => {
            loading.value = false;
        });
};

const openCreate = () => {
    dialogRef.value.openCreate();
};

const openEdit = (row: AI.TensorRTLLM) => {
    dialogRef.value.openEdit(row);
};

const goDashboard = async (port: any, protocol: string) => {
    dialogPortJumpRef.value.acceptParams({ port: port, protocol: protocol });
};

const openLog = (row: AI.McpServer) => {
    composeLogRef.value.acceptParams({
        compose: row.dir + '/docker-compose.yml',
        resource: row.name,
        container: row.containerName,
    });
};

const operate = async (row: AI.TensorRTLLM, operation: string) => {
    ElMessageBox.confirm(
        i18n.global.t('commons.msg.operatorHelper', ['LLM', i18n.global.t('commons.operate.' + operation)]),
        i18n.global.t('commons.operate.' + operation),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    ).then(async () => {
        loading.value = true;
        await operateTensorRTLLM({ id: row.id, operate: operation })
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                search();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};
const deleteLLM = async (row: AI.TensorRTLLM) => {
    try {
        opRef.value.acceptParams({
            title: i18n.global.t('commons.button.delete'),
            names: [row.name],
            msg: i18n.global.t('commons.msg.operatorHelper', ['LLM', i18n.global.t('commons.button.delete')]),
            api: deleteTensorRTLLM,
            params: { id: row.id },
        });
    } catch (error) {}
};

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: AI.TensorRTLLM) => {
            openEdit(row);
        },
    },
    {
        label: i18n.global.t('commons.button.start'),
        disabled: (row: AI.TensorRTLLM) => {
            return row.status === 'running';
        },
        click: (row: AI.TensorRTLLM) => {
            operate(row, 'start');
        },
    },
    {
        label: i18n.global.t('commons.button.stop'),
        disabled: (row: AI.TensorRTLLM) => {
            return row.status !== 'running';
        },
        click: (row: AI.TensorRTLLM) => {
            operate(row, 'stop');
        },
    },
    {
        label: i18n.global.t('commons.button.restart'),
        disabled: (row: AI.TensorRTLLM) => {
            return row.status !== 'running';
        },
        click: (row: AI.TensorRTLLM) => {
            operate(row, 'restart');
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: AI.TensorRTLLM) => {
            deleteLLM(row);
        },
    },
];

onMounted(() => {
    search();
});
</script>
