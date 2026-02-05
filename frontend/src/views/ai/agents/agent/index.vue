<template>
    <div>
        <RouterMenu />
        <LayoutContent :title="$t('aiTools.agents.agentList')">
            <template #leftToolBar>
                <el-button type="primary" @click="openCreate">{{ $t('aiTools.agents.createAgent') }}</el-button>
            </template>
            <template #rightToolBar>
                <TableSearch v-model:searchName="searchName" @search="search" />
                <TableRefresh @search="search" />
            </template>
            <template #main>
                <ComplexTable :data="items" :pagination-config="paginationConfig" @search="search">
                    <el-table-column :label="$t('commons.table.name')" prop="name" min-width="160" />
                    <el-table-column :label="$t('aiTools.agents.appVersion')" prop="appVersion" width="120" />
                    <el-table-column :label="$t('aiTools.agents.webuiPort')" prop="webUIPort" width="140">
                        <template #default="{ row }">
                            <el-button icon="Position" plain size="small" @click="jumpWebUI(row)">
                                {{ row.webUIPort }}
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('aiTools.agents.bridgePort')" prop="bridgePort" width="120" />
                    <el-table-column :label="$t('aiTools.agents.provider')" prop="provider" width="120" />
                    <el-table-column :label="$t('aiTools.model.model')" prop="model" min-width="180" />
                    <el-table-column :label="$t('aiTools.agents.token')" width="120">
                        <template #default="{ row }">
                            <CopyButton :content="row.token" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.status')" prop="status" width="120">
                        <template #default="{ row }">
                            <Status :status="row.status" />
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('commons.table.date')"
                        prop="createdAt"
                        width="180"
                        :formatter="dateFormat"
                        show-overflow-tooltip
                    />
                    <fu-table-operations
                        :buttons="buttons"
                        min-width="300"
                        :label="$t('commons.table.operate')"
                        fixed="right"
                        :ellipsis="5"
                    />
                </ComplexTable>
            </template>
        </LayoutContent>
        <AddDialog ref="addRef" @search="search" @task="openTaskLog" />
        <TaskLog ref="taskLogRef" @close="search" />
        <DeleteDialog ref="deleteRef" @close="search" />
        <ComposeLogs ref="composeLogRef" />
        <TerminalDialog ref="dialogTerminalRef" />
        <PortJumpDialog ref="dialogPortJumpRef" />
    </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { pageAgents } from '@/api/modules/ai';
import { installedOp } from '@/api/modules/app';
import { AI } from '@/api/interface/ai';
import { SearchWithPage } from '@/api/interface';
import { dateFormat, newUUID } from '@/utils/util';

import RouterMenu from '@/views/ai/agents/index.vue';
import AddDialog from '@/views/ai/agents/agent/add/index.vue';
import DeleteDialog from '@/views/ai/agents/agent/delete/index.vue';
import TaskLog from '@/components/log/task/index.vue';
import ComposeLogs from '@/components/log/compose/index.vue';
import TerminalDialog from '@/views/container/container/terminal/index.vue';
import i18n from '@/lang';
import PortJumpDialog from '@/components/port-jump/index.vue';

const items = ref<AI.AgentItem[]>([]);
const addRef = ref();
const taskLogRef = ref();
const deleteRef = ref();
const composeLogRef = ref();
const dialogTerminalRef = ref();
const dialogPortJumpRef = ref();
const searchName = ref('');

const buttons = [
    {
        label: i18n.global.t('commons.operate.start'),
        click: (row: AI.AgentItem) => onOperate(row, 'start'),
        disabled: (row: AI.AgentItem) => row.status === 'Running',
    },
    {
        label: i18n.global.t('commons.operate.stop'),
        click: (row: AI.AgentItem) => onOperate(row, 'stop'),
        disabled: (row: AI.AgentItem) => row.status !== 'Running',
    },
    {
        label: i18n.global.t('commons.operate.restart'),
        click: (row: AI.AgentItem) => onOperate(row, 'restart'),
    },
    {
        label: i18n.global.t('commons.button.log'),
        click: (row: AI.AgentItem) => openLog(row),
    },
    {
        label: i18n.global.t('menu.terminal'),
        click: (row: AI.AgentItem) => openTerminal(row),
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: AI.AgentItem) => onDelete(row),
    },
];

const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const search = async () => {
    const req: SearchWithPage = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        info: searchName.value || '',
    };
    const res = await pageAgents(req);
    items.value = res.data.items || [];
    paginationConfig.total = res.data.total || 0;
};

const openCreate = () => {
    if (addRef.value?.open) {
        addRef.value.open();
    }
};

const openTaskLog = (taskID: string) => {
    if (taskLogRef.value?.openWithTaskID) {
        taskLogRef.value.openWithTaskID(taskID);
    }
};

const onOperate = async (row: AI.AgentItem, operate: string) => {
    await ElMessageBox.confirm(
        i18n.global.t('app.operatorHelper', [i18n.global.t('commons.operate.' + operate)]),
        i18n.global.t('commons.operate.' + operate),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    );
    const taskID = newUUID();
    await installedOp({ installId: row.appInstallId, operate, taskID });
    await search();
};

const openLog = (row: AI.AgentItem) => {
    if (row.status === 'Installing') {
        taskLogRef.value?.openWithResourceID('App', 'TaskInstall', row.appInstallId);
        return;
    }
    composeLogRef.value?.acceptParams({
        compose: `${row.path}/docker-compose.yml`,
        resource: row.name,
        container: row.containerName,
    });
};

const openTerminal = (row: AI.AgentItem) => {
    const title = i18n.global.t('aiTools.agents.agent') + ' ' + row.name;
    dialogTerminalRef.value?.acceptParams({ containerID: row.containerName, title });
};

const jumpWebUI = (row: AI.AgentItem) => {
    if (dialogPortJumpRef.value?.acceptParams) {
        dialogPortJumpRef.value.acceptParams({
            port: row.webUIPort,
            protocol: 'http',
            query: `token=${row.token}`,
        });
    }
};

const onDelete = (row: AI.AgentItem) => {
    deleteRef.value?.acceptParams(row.id, row.name);
};

onMounted(async () => {
    await search();
});
</script>
