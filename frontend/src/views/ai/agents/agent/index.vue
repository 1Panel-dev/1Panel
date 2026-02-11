<template>
    <div>
        <RouterMenu />
        <DockerStatus v-model:isActive="isActive" v-model:isExist="isExist" />
        <LayoutContent v-loading="loading" v-if="isExist" :class="{ mask: !isActive }">
            <template #leftToolBar>
                <el-button type="primary" @click="openCreate">{{ $t('aiTools.agents.createAgent') }}</el-button>
            </template>
            <template #rightToolBar>
                <TableSearch v-model:searchName="searchName" @search="search" />
                <TableRefresh @search="search" />
            </template>
            <template #main>
                <ComplexTable :data="items" :pagination-config="paginationConfig" @search="search">
                    <el-table-column
                        :label="$t('commons.table.name')"
                        show-overflow-tooltip
                        prop="name"
                        min-width="120"
                    />
                    <el-table-column :label="$t('commons.table.status')" prop="status" width="120">
                        <template #default="{ row }">
                            <Status :status="row.status" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('aiTools.agents.appVersion')" prop="appVersion" min-width="140">
                        <template #default="{ row }">
                            <span>{{ row.appVersion }}</span>
                            <el-button v-if="row.upgradable" link type="primary" class="ml-1" @click="openUpgrade(row)">
                                {{ $t('commons.button.upgrade') }}
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('aiTools.model.model')"
                        show-overflow-tooltip
                        prop="provider"
                        min-width="120"
                    >
                        <template #default="{ row }">
                            {{ row.providerName || row.provider }}
                            <div>
                                <span>{{ row.model }}</span>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.port')" prop="webUIPort" min-width="150">
                        <template #default="{ row }">
                            <el-button icon="Position" plain size="small" @click="jumpWebUI(row)">
                                {{ $t('aiTools.agents.webuiPort') }}: {{ row.webUIPort }}
                            </el-button>
                            <div class="mt-0.5">
                                <el-button plain size="small">
                                    {{ $t('aiTools.agents.bridgePort') }} {{ row.bridgePort }}
                                </el-button>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('aiTools.agents.token')" min-width="80">
                        <template #default="{ row }">
                            <el-space>
                                <CopyButton :content="row.token" />
                                <el-button link type="primary" @click="onResetToken(row)">
                                    {{ $t('commons.button.reset') }}
                                </el-button>
                            </el-space>
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
                        min-width="220"
                        :label="$t('commons.table.operate')"
                        fixed="right"
                        :ellipsis="3"
                    />
                </ComplexTable>
            </template>
        </LayoutContent>
        <AddDialog ref="addRef" @search="search" @task="openTaskLog" />
        <TaskLog ref="taskLogRef" @close="search" />
        <DeleteDialog ref="deleteRef" @close="search" />
        <ConfigDrawer ref="configRef" @updated="search" />
        <AppUpgrade ref="upgradeRef" @close="search" />
        <ComposeLogs ref="composeLogRef" />
        <TerminalDialog ref="dialogTerminalRef" />
        <PortJumpDialog ref="dialogPortJumpRef" />
    </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { pageAgents, resetAgentToken } from '@/api/modules/ai';
import { installedOp, searchAppInstalled } from '@/api/modules/app';
import { AI } from '@/api/interface/ai';
import { App } from '@/api/interface/app';
import { SearchWithPage } from '@/api/interface';
import { dateFormat, newUUID } from '@/utils/util';
import { MsgSuccess } from '@/utils/message';

import RouterMenu from '@/views/ai/agents/index.vue';
import AddDialog from '@/views/ai/agents/agent/add/index.vue';
import DeleteDialog from '@/views/ai/agents/agent/delete/index.vue';
import ConfigDrawer from '@/views/ai/agents/agent/config/index.vue';
import AppUpgrade from '@/views/app-store/installed/upgrade/index.vue';
import TaskLog from '@/components/log/task/index.vue';
import ComposeLogs from '@/components/log/compose/index.vue';
import TerminalDialog from '@/views/container/container/terminal/index.vue';
import i18n from '@/lang';
import PortJumpDialog from '@/components/port-jump/index.vue';
import DockerStatus from '@/views/container/docker-status/index.vue';

const items = ref<AI.AgentItem[]>([]);
const loading = ref(false);
const addRef = ref();
const taskLogRef = ref();
const deleteRef = ref();
const configRef = ref();
const upgradeRef = ref();
const composeLogRef = ref();
const dialogTerminalRef = ref();
const dialogPortJumpRef = ref();
const isActive = ref(false);
const isExist = ref(false);
const searchName = ref('');

const buttons = [
    {
        label: i18n.global.t('menu.config'),
        click: (row: AI.AgentItem) => openConfig(row),
    },
    {
        label: i18n.global.t('menu.terminal'),
        click: (row: AI.AgentItem) => openTerminal(row),
    },
    {
        label: i18n.global.t('commons.button.log'),
        click: (row: AI.AgentItem) => openLog(row),
    },
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
        label: i18n.global.t('commons.button.upgrade'),
        click: (row: AI.AgentItem) => openUpgrade(row),
        disabled: (row: AI.AgentItem) => !row.upgradable,
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
    loading.value = true;
    try {
        const req: SearchWithPage = {
            page: paginationConfig.currentPage,
            pageSize: paginationConfig.pageSize,
            info: searchName.value || '',
        };
        const res = await pageAgents(req);
        items.value = res.data.items || [];
        paginationConfig.total = res.data.total || 0;
    } finally {
        loading.value = false;
    }
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

const onResetToken = async (row: AI.AgentItem) => {
    await ElMessageBox.confirm(
        i18n.global.t('aiTools.mcp.operatorHelper', ['token', i18n.global.t('commons.button.reset')]),
        i18n.global.t('commons.button.reset'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    );
    await resetAgentToken({ id: row.id });
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    await search();
};

const openConfig = (row: AI.AgentItem) => {
    configRef.value?.open(row);
};

const openUpgrade = async (row: AI.AgentItem) => {
    const res = await searchAppInstalled({ page: 1, pageSize: 200, name: row.name, update: true });
    const appInstall = (res.data.items || []).find((item: App.AppInstallDto) => item.id === row.appInstallId);
    if (!appInstall) {
        return;
    }
    upgradeRef.value?.acceptParams(appInstall, 'upgrade');
};

onMounted(async () => {
    await search();
});
</script>
