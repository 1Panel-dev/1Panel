<template>
    <div>
        <RouterButton :buttons="headerButtons" />
        <DockerStatus v-model:isActive="isActive" v-model:isExist="isExist" />
        <LayoutContent v-loading="loading" v-if="isExist" :class="{ mask: !isActive }">
            <template #leftToolBar>
                <el-button v-permission type="primary" @click="openCreate" :disabled="noApp">
                    {{ $t('commons.button.create') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <TableSearch v-model:searchName="searchName" @search="search" />
                <TableRefresh @search="search" />
            </template>
            <template #main>
                <NoApp v-if="noApp" />
                <ComplexTable :data="items" :pagination-config="paginationConfig" @search="search" v-if="!noApp">
                    <el-table-column
                        :label="$t('commons.table.name')"
                        show-overflow-tooltip
                        prop="name"
                        min-width="120"
                    />
                    <el-table-column :label="$t('commons.table.type')" prop="agentType" min-width="150">
                        <template #default="{ row }">
                            <div class="agent-type-cell">
                                <img
                                    class="agent-type-icon"
                                    :src="getAgentTypeIcon(row.agentType)"
                                    :alt="getAgentTypeLabel(row.agentType)"
                                />
                                <span>{{ getAgentTypeLabel(row.agentType) }}</span>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.status')" prop="status" width="120">
                        <template #default="{ row }">
                            <el-dropdown placement="bottom">
                                <Status v-permission :status="row.status" :operate="true" />
                                <template #dropdown>
                                    <el-dropdown-menu>
                                        <fu-dropdown-item
                                            v-permission
                                            :disabled="checkStatus('start', row)"
                                            @click="onOperate(row, 'start')"
                                        >
                                            {{ $t('commons.operate.start') }}
                                        </fu-dropdown-item>
                                        <fu-dropdown-item
                                            v-permission
                                            :disabled="checkStatus('stop', row)"
                                            @click="onOperate(row, 'stop')"
                                        >
                                            {{ $t('commons.operate.stop') }}
                                        </fu-dropdown-item>
                                        <fu-dropdown-item
                                            v-permission
                                            :disabled="checkStatus('restart', row)"
                                            @click="onOperate(row, 'restart')"
                                        >
                                            {{ $t('commons.button.restart') }}
                                        </fu-dropdown-item>
                                    </el-dropdown-menu>
                                </template>
                            </el-dropdown>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('aiTools.agents.appVersion')" prop="appVersion" min-width="140">
                        <template #default="{ row }">
                            <div class="version-cell">
                                <span>{{ row.appVersion }}</span>
                                <el-button
                                    v-permission
                                    v-if="row.upgradable"
                                    link
                                    type="primary"
                                    @click="openUpgrade(row)"
                                >
                                    {{ $t('commons.button.upgrade') }}
                                </el-button>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('aiTools.model.model')"
                        show-overflow-tooltip
                        prop="provider"
                        min-width="150"
                    >
                        <template #default="{ row }">
                            <template v-if="row.agentType === 'openclaw' || row.agentType === 'hermes-agent'">
                                <span>{{ getAgentProviderDisplayName(row.provider, row.providerName) }}</span>
                                <div>
                                    <span>{{ row.model }}</span>
                                </div>
                            </template>
                            <span v-else>-</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.port')" prop="webUIPort" min-width="180">
                        <template #default="{ row }">
                            <el-button icon="Position" plain size="small" @click="jumpWebUI(row)">
                                {{ $t('aiTools.agents.webuiPort') }}: {{ row.webUIPort }}
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('menu.website')" min-width="180">
                        <template #default="{ row }">
                            <div v-if="row.websiteId > 0" class="website-link-cell">
                                <el-popover
                                    placement="right"
                                    trigger="hover"
                                    :width="420"
                                    @before-enter="loadWebsiteDomains(row.websiteId)"
                                >
                                    <template #reference>
                                        <el-text
                                            type="primary"
                                            class="cursor-pointer website-link-cell__name"
                                            @click="openWebsite(row)"
                                        >
                                            {{ getWebsiteDisplayName(row) }}
                                        </el-text>
                                    </template>
                                    <table v-if="getWebsiteBaseUrls(row).length > 0">
                                        <tbody>
                                            <tr v-for="url in getWebsiteBaseUrls(row)" :key="url">
                                                <td>
                                                    <el-button type="primary" link @click="openWebsiteUrl(url, row)">
                                                        {{ url }}
                                                    </el-button>
                                                </td>
                                                <td>
                                                    <CopyButton :content="url" />
                                                </td>
                                            </tr>
                                        </tbody>
                                    </table>
                                    <el-empty v-else :image-size="48" :description="$t('commons.msg.noneData')" />
                                </el-popover>
                                <el-button
                                    v-if="canUnbindWebsite(row)"
                                    link
                                    type="primary"
                                    class="website-link-cell__unbind"
                                    v-permission
                                    @click="onUnbindWebsite(row)"
                                >
                                    {{ $t('commons.button.unbind') }}
                                </el-button>
                            </div>
                            <el-button v-else link type="primary" v-permission @click="openBindWebsite(row)">
                                {{ $t('commons.button.bind') }}
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('website.remark')" prop="remark" min-width="150">
                        <template #default="{ row }">
                            <fu-read-write-switch v-permission>
                                <template #read>
                                    <MsgInfo :info="row.remark" :width="'150'" />
                                </template>
                                <template #default="{ read }">
                                    <el-input v-model="row.remark" @blur="handleUpdateRemark(row, read)" />
                                </template>
                            </fu-read-write-switch>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('runtime.workDir')" min-width="90">
                        <template #default="{ row }">
                            <el-button
                                v-permission:view="'host_file_view'"
                                type="primary"
                                link
                                @click="openWorkDir(row)"
                            >
                                <el-icon>
                                    <FolderOpened />
                                </el-icon>
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column label="Token" min-width="120">
                        <template #default="{ row }">
                            <el-space v-if="supportsAgentToken(row.agentType)">
                                <CopyButton :content="row.token" />
                                <el-button v-permission link type="primary" @click="onResetToken(row)">
                                    {{ $t('commons.button.reset') }}
                                </el-button>
                            </el-space>
                            <span v-else>-</span>
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
                        min-width="200"
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
        <AppResources ref="checkRef" @close="search" />
        <ConfigDrawer ref="configRef" @updated="search" />
        <OverviewDrawer ref="overviewRef" />
        <BindWebsiteDialog ref="bindWebsiteRef" @success="search" />
        <AppUpgrade ref="upgradeRef" @close="search" />
        <ComposeLogs ref="composeLogRef" />
        <AgentTerminalDialog ref="dialogTerminalRef" />
        <HermesChatDialog ref="hermesChatRef" />
        <PortJumpDialog ref="dialogPortJumpRef" />
    </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { deleteAgentCheck, pageAgents, resetAgentToken, unbindAgentWebsite, updateAgentRemark } from '@/api/modules/ai';
import { checkAppInstalled, installedOp, searchApp, searchAppInstalled } from '@/api/modules/app';
import { AI } from '@/api/interface/ai';
import { App } from '@/api/interface/app';
import { Website } from '@/api/interface/website';
import { SearchWithPage } from '@/api/interface';
import { dateFormat } from '@/utils/date';
import { newUUID } from '@/utils/id';
import { MsgSuccess } from '@/utils/message';

import AddDialog from '@/views/ai/agents/agent/add/index.vue';
import DeleteDialog from '@/views/ai/agents/agent/delete/index.vue';
import AppResources from '@/views/app-store/installed/check/index.vue';
import ConfigDrawer from '@/views/ai/agents/agent/config/index.vue';
import OverviewDrawer from '@/views/ai/agents/agent/components/overview.vue';
import BindWebsiteDialog from '@/views/ai/agents/agent/website/index.vue';
import AppUpgrade from '@/views/app-store/installed/upgrade/index.vue';
import TaskLog from '@/components/log/task/index.vue';
import ComposeLogs from '@/components/log/compose/index.vue';
import AgentTerminalDialog from '@/views/ai/agents/agent/components/terminal.vue';
import HermesChatDialog from '@/views/ai/agents/agent/components/hermes-chat.vue';
import i18n from '@/lang';
import PortJumpDialog from '@/components/port-jump/index.vue';
import DockerStatus from '@/views/container/docker-status/index.vue';
import {
    getAgentProviderDisplayName,
    getOpenclawAccessScheme,
    supportsAgentModelConfig,
    supportsAgentToken,
} from '@/utils/agent';
import { routerToFileWithPath } from '@/utils/router';
import { listDomains } from '@/api/modules/website';
import NoApp from '@/views/app-store/apps/no-app/index.vue';
import openclawIcon from '@/assets/images/ai-agent-openclaw.svg';
import copawIcon from '@/assets/images/ai-agent-copaw.svg';
import hermesIcon from '@/assets/images/ai-agent-hermes-agent.svg';
import { GlobalStore } from '@/store';

const items = ref<AI.AgentItem[]>([]);
const loading = ref(false);
const addRef = ref();
const taskLogRef = ref();
const deleteRef = ref();
const checkRef = ref();
const configRef = ref();
const overviewRef = ref();
const bindWebsiteRef = ref();
const upgradeRef = ref();
const composeLogRef = ref();
const dialogTerminalRef = ref();
const hermesChatRef = ref();
const dialogPortJumpRef = ref();
const route = useRoute();
const router = useRouter();
const isActive = ref(false);
const isExist = ref(false);
const noApp = ref(false);
const searchName = ref('');
const defaultHttpsPort = ref(443);
const openrestyPortLoaded = ref(false);
const websiteDomainsMap = ref<Record<number, Website.Domain[]>>({});
const globalStore = GlobalStore();

const headerButtons = [
    {
        label: i18n.global.t('aiTools.agents.agent'),
        path: '/ai/agents/agent',
    },
];

const buttons = [
    {
        label: i18n.global.t('menu.config'),
        click: (row: AI.AgentItem) => openConfig(row),
        show: (row: AI.AgentItem) => supportsAgentModelConfig(row.agentType),
        disabled: (row: AI.AgentItem) => row.status !== 'Running',
    },
    {
        label: i18n.global.t('aiTools.agents.hermesChatAction'),
        click: (row: AI.AgentItem) => openHermesChat(row),
        disabled: () => !globalStore.isAdminOrNodeAdmin,
        show: (row: AI.AgentItem) => row.agentType === 'hermes-agent' && row.status === 'Running',
    },
    {
        label: i18n.global.t('commons.button.log'),
        click: (row: AI.AgentItem) => openLog(row),
    },
    {
        label: i18n.global.t('menu.terminal'),
        click: (row: AI.AgentItem) => openTerminal(row),
        disabled: () => !globalStore.isAdminOrNodeAdmin,
    },
    {
        label: i18n.global.t('menu.home'),
        click: (row: AI.AgentItem) => openOverview(row),
        show: (row: AI.AgentItem) => row.agentType === 'openclaw',
    },
    {
        label: i18n.global.t('commons.operate.start'),
        permission: true,
        click: (row: AI.AgentItem) => onOperate(row, 'start'),
        disabled: (row: AI.AgentItem) => row.status === 'Running',
    },
    {
        label: i18n.global.t('commons.operate.stop'),
        permission: true,
        click: (row: AI.AgentItem) => onOperate(row, 'stop'),
        disabled: (row: AI.AgentItem) => row.status !== 'Running',
    },
    {
        label: i18n.global.t('commons.operate.restart'),
        permission: true,
        click: (row: AI.AgentItem) => onOperate(row, 'restart'),
    },
    {
        label: i18n.global.t('commons.button.upgrade'),
        permission: true,
        click: (row: AI.AgentItem) => openUpgrade(row),
        disabled: (row: AI.AgentItem) => !row.upgradable,
    },
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
        click: (row: AI.AgentItem) => onDelete(row),
    },
];

const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const checkNoApp = async () => {
    try {
        const appRes = await searchApp({
            page: 1,
            pageSize: 1,
            name: '',
            tags: [],
            resource: 'all',
            showCurrentArch: false,
        });
        noApp.value = (appRes.data.total || 0) === 0;
    } catch {
        noApp.value = false;
    }
};

const search = async () => {
    loading.value = true;
    try {
        const req: SearchWithPage = {
            page: paginationConfig.currentPage,
            pageSize: paginationConfig.pageSize,
            info: searchName.value || '',
        };
        const [res] = await Promise.all([pageAgents(req), checkNoApp(), loadOpenrestyHttpsPort()]);
        items.value = res.data.items || [];
        paginationConfig.total = res.data.total || 0;
        await preloadWebsiteDomains(items.value);
    } finally {
        loading.value = false;
    }
};

const getAgentTypeLabel = (agentType: AI.AgentType) => {
    switch (agentType) {
        case 'copaw':
            return i18n.global.t('aiTools.agents.copawType');
        case 'hermes-agent':
            return i18n.global.t('aiTools.agents.hermesType');
        default:
            return i18n.global.t('aiTools.agents.openclawType');
    }
};

const getAgentTypeIcon = (agentType: AI.AgentType) => {
    switch (agentType) {
        case 'copaw':
            return copawIcon;
        case 'hermes-agent':
            return hermesIcon;
        default:
            return openclawIcon;
    }
};

const openCreate = (agentType?: AI.AgentType) => {
    if (noApp.value) {
        return;
    }
    const targetType =
        agentType === 'copaw' || agentType === 'hermes-agent' || agentType === 'openclaw' ? agentType : 'openclaw';
    if (addRef.value?.open) {
        addRef.value.open(targetType);
    }
};

const openCreateFromQuery = async () => {
    const shouldOpen = route.query.open === 'create';
    if (!shouldOpen) {
        return;
    }
    const agentType =
        route.query.agentType === 'copaw'
            ? 'copaw'
            : route.query.agentType === 'hermes-agent'
              ? 'hermes-agent'
              : 'openclaw';
    openCreate(agentType);
    const nextQuery = { ...route.query };
    delete nextQuery.open;
    delete nextQuery.agentType;
    await router.replace({ path: route.path, query: nextQuery });
};

const openTaskLog = (taskID: string) => {
    if (taskLogRef.value?.openWithTaskID) {
        taskLogRef.value.openWithTaskID(taskID);
    }
};

const checkStatus = (operate: string, row: AI.AgentItem) => {
    const status = row.status.toLowerCase();
    switch (operate) {
        case 'start':
            return status === 'running' || status === 'starting' || status === 'restarting';
        case 'stop':
            return status !== 'running';
        case 'restart':
            return status === 'starting';
        default:
            return false;
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
    openTerminalDialog({
        containerID: row.containerName,
        title,
        users: row.agentType === 'hermes-agent' ? ['hermes', 'root'] : ['node', 'root'],
        shell: '/bin/bash',
        initCmd: '',
    });
};

const openHermesChat = (row: AI.AgentItem) => {
    hermesChatRef.value?.acceptParams({
        agentId: row.id,
        containerID: row.containerName,
        title: i18n.global.t('aiTools.agents.hermesChatDialogTitle', [row.name]),
    });
};

const openTerminalDialog = (params: {
    containerID: string;
    title: string;
    users: string[];
    shell: string;
    initCmd?: string;
}) => {
    dialogTerminalRef.value?.acceptParams(params);
};

const openWorkDir = (row: AI.AgentItem) => {
    if (!row.path) {
        return;
    }
    routerToFileWithPath(`${row.path}/data`);
};

const jumpWebUI = (row: AI.AgentItem) => {
    if (dialogPortJumpRef.value?.acceptParams) {
        dialogPortJumpRef.value.acceptParams({
            port: row.webUIPort,
            protocol: row.agentType === 'openclaw' ? getOpenclawAccessScheme(row.appVersion) : 'http',
            path: row.agentType === 'openclaw' ? '/' : undefined,
            hash: row.agentType === 'openclaw' ? `token=${row.token}` : undefined,
        });
    }
};

const onDelete = async (row: AI.AgentItem) => {
    const res = await deleteAgentCheck({ agentId: row.id });
    if ((res.data || []).length > 0) {
        checkRef.value?.acceptParams({
            items: res.data,
            installID: row.appInstallId,
            key: row.agentType,
        });
        return;
    }
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

const handleUpdateRemark = async (row: AI.AgentItem, read: Function) => {
    read();
    await updateAgentRemark({ id: row.id, remark: row.remark });
    MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
};

const openConfig = (row: AI.AgentItem) => {
    configRef.value?.open(row);
};

const openOverview = (row: AI.AgentItem) => {
    overviewRef.value?.acceptParams(row);
};

const openBindWebsite = (row: AI.AgentItem) => {
    bindWebsiteRef.value?.acceptParams(row);
};

const canUnbindWebsite = (row: AI.AgentItem) => {
    return row.websiteType === 'proxy' || row.websiteType === 'static';
};

const onUnbindWebsite = async (row: AI.AgentItem) => {
    await ElMessageBox.confirm(
        i18n.global.t('aiTools.mcp.operatorHelper', [
            i18n.global.t('menu.website'),
            i18n.global.t('commons.button.unbind'),
        ]),
        i18n.global.t('commons.button.unbind'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    );
    await unbindAgentWebsite({ agentId: row.id });
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    await search();
};

const loadOpenrestyHttpsPort = async () => {
    if (openrestyPortLoaded.value) {
        return;
    }
    openrestyPortLoaded.value = true;
    try {
        const res = await checkAppInstalled('openresty', '');
        defaultHttpsPort.value = res.data.httpsPort || 443;
    } catch {
        defaultHttpsPort.value = 443;
    }
};

const sortWebsiteDomains = (domains: Website.Domain[]) => {
    return [...domains].sort((a, b) => a.id - b.id);
};

const loadWebsiteDomains = async (websiteId: number) => {
    if (!websiteId) {
        return [];
    }
    if (websiteDomainsMap.value[websiteId]) {
        return websiteDomainsMap.value[websiteId];
    }
    const res = await listDomains(websiteId);
    const domains = sortWebsiteDomains(res.data || []);
    websiteDomainsMap.value = {
        ...websiteDomainsMap.value,
        [websiteId]: domains,
    };
    return domains;
};

const preloadWebsiteDomains = async (rows: AI.AgentItem[]) => {
    const websiteIDs = [...new Set(rows.map((row) => row.websiteId).filter((websiteId) => websiteId > 0))];
    await Promise.allSettled(websiteIDs.map((websiteId) => loadWebsiteDomains(websiteId)));
};

const formatWebsiteHost = (domain: string) => {
    const host = String(domain || '')
        .trim()
        .replace(/^\[|\]$/g, '');
    return host.includes(':') ? `[${host}]` : host;
};

const buildWebsiteBaseUrl = (domain: Website.Domain, row: AI.AgentItem) => {
    const protocol = (row.websiteProtocol || 'http').toLowerCase();
    let url = `${protocol}://${formatWebsiteHost(domain.domain)}`;
    if (protocol === 'http') {
        if (domain.port && domain.port !== 80) {
            url = `${url}:${domain.port}`;
        }
        return url;
    }
    let port = domain.port;
    if (!domain.ssl) {
        port = defaultHttpsPort.value || 443;
    }
    if (port && port !== 443) {
        url = `${url}:${port}`;
    }
    return url;
};

const appendWebsiteAgentToken = (url: string, row: AI.AgentItem) => {
    if (supportsAgentToken(row.agentType) && row.token) {
        const target = new URL(url);
        target.hash = `token=${row.token}`;
        return target.toString();
    }
    return url;
};

const getWebsiteBaseUrls = (row: AI.AgentItem) => {
    const domains = websiteDomainsMap.value[row.websiteId] || [];
    return domains.map((domain) => buildWebsiteBaseUrl(domain, row));
};

const getWebsiteDisplayName = (row: AI.AgentItem) => {
    const domains = websiteDomainsMap.value[row.websiteId] || [];
    return domains[0]?.domain || row.websitePrimaryDomain || '-';
};

const openUrl = (url: string) => {
    window.open(url);
};

const openWebsiteUrl = (url: string, row: AI.AgentItem) => {
    openUrl(appendWebsiteAgentToken(url, row));
};

const openWebsite = async (row: AI.AgentItem) => {
    const domains = await loadWebsiteDomains(row.websiteId);
    if (domains.length === 0) {
        return;
    }
    openWebsiteUrl(buildWebsiteBaseUrl(domains[0], row), row);
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
    await openCreateFromQuery();
});
</script>

<style scoped>
.version-cell {
    display: inline-flex;
    align-items: center;
    gap: 4px;
}

.agent-type-cell {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    white-space: nowrap;
}

.agent-type-icon {
    width: 16px;
    height: 16px;
    flex: 0 0 16px;
    object-fit: contain;
}

.website-link-cell {
    display: inline-flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
}

.website-link-cell__name {
    white-space: nowrap;
}

.website-link-cell__unbind {
    margin-left: 0;
}
</style>
