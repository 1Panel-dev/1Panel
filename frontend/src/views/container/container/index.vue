<template>
    <div v-loading="loading">
        <docker-status v-model:isActive="isActive" v-model:isExist="isExist" @search="search" />

        <LayoutContent :title="$t('menu.container', 2)" v-if="isExist" :class="{ mask: !isActive }">
            <template #search v-if="tags.length !== 0">
                <div class="card-interval" v-if="isExist && isActive">
                    <div v-for="item in tags" :key="item.key" class="inline">
                        <el-button
                            v-if="item.count"
                            class="tag-button"
                            :class="activeTag === item.key ? '' : 'no-active'"
                            @click="searchWithStatus(item.key)"
                            :type="activeTag === item.key ? 'primary' : ''"
                            :plain="activeTag !== item.key"
                        >
                            {{ item.key === 'all' ? $t('commons.table.all') : $t('commons.status.' + item.key) }} *
                            {{ item.count }}
                        </el-button>
                    </div>
                </div>
            </template>

            <template #leftToolBar>
                <el-button v-permission type="primary" @click="onContainerOperate('')">
                    {{ $t('commons.button.create') }}
                </el-button>
                <el-button v-permission type="primary" plain @click="onImportCreate()">
                    {{ $t('commons.button.import') }}
                </el-button>
                <el-button v-permission type="primary" plain @click="onClean()">
                    {{ $t('container.containerPrune') }}
                </el-button>
                <el-button-group class="button-group">
                    <el-button v-permission :disabled="checkStatus('start', null)" @click="onOperate('start', null)">
                        {{ $t('commons.operate.start') }}
                    </el-button>
                    <el-button v-permission :disabled="checkStatus('stop', null)" @click="onOperate('stop', null)">
                        {{ $t('commons.operate.stop') }}
                    </el-button>
                    <el-button
                        v-permission
                        :disabled="checkStatus('restart', null)"
                        @click="onOperate('restart', null)"
                    >
                        {{ $t('commons.button.restart') }}
                    </el-button>
                    <el-button v-permission :disabled="checkStatus('kill', null)" @click="onOperate('kill', null)">
                        {{ $t('container.kill') }}
                    </el-button>
                    <el-button v-permission :disabled="checkStatus('pause', null)" @click="onOperate('pause', null)">
                        {{ $t('container.pause') }}
                    </el-button>
                    <el-button
                        v-permission
                        :disabled="checkStatus('unpause', null)"
                        @click="onOperate('unpause', null)"
                    >
                        {{ $t('container.unpause') }}
                    </el-button>
                    <el-button v-permission :disabled="checkStatus('remove', null)" @click="onOperate('remove', null)">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </el-button-group>
            </template>
            <template #rightToolBar>
                <TableSearch @search="search()" v-model:searchName="searchName" />
                <el-tooltip
                    :content="includeAppStore ? $t('container.includeAppstore') : $t('container.excludeAppstore')"
                >
                    <el-button
                        :type="includeAppStore ? '' : 'primary'"
                        @click="searchWithAppShow(!includeAppStore)"
                        :icon="includeAppStore ? 'View' : 'Hide'"
                    />
                </el-tooltip>
                <TableRefresh @search="search()" />
                <TableSetting title="container-refresh" @search="refresh()" />
                <fu-table-column-select
                    :columns="columns"
                    trigger="hover"
                    :title="$t('commons.table.selectColumn')"
                    popper-class="popper-class"
                    :only-icon="true"
                />
            </template>
            <template #main>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    :data="data"
                    row-key="containerID"
                    @sort-change="search"
                    @search="search"
                    @cell-mouse-enter="showFavorite"
                    @cell-mouse-leave="hideFavorite"
                    :row-style="{ height: '65px' }"
                    style="width: 100%"
                    :columns="columns"
                    localKey="containerColumn"
                    :heightDiff="300"
                >
                    <el-table-column type="selection" />
                    <el-table-column
                        :label="$t('commons.table.name')"
                        min-width="250"
                        prop="name"
                        sortable="custom"
                        fix
                        :fixed="isMobile ? false : 'left'"
                        show-overflow-tooltip
                    >
                        <template #default="{ row, $index }">
                            <el-text type="primary" class="cursor-pointer" @click="onInspect(row)">
                                {{ row.name }}
                            </el-text>

                            <div class="float-right">
                                <el-tooltip
                                    :content="row.isPinned ? $t('website.cancelFavorite') : $t('website.favorite')"
                                    v-if="row.isPinned || hoveredRowIndex === $index"
                                >
                                    <el-button
                                        link
                                        size="large"
                                        :icon="row.isPinned ? 'StarFilled' : 'Star'"
                                        type="warning"
                                        v-permission
                                        @click="changePinned(row, true)"
                                    />
                                </el-tooltip>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('container.image')"
                        show-overflow-tooltip
                        min-width="180"
                        prop="imageName"
                    />
                    <el-table-column :label="$t('commons.table.status')" min-width="150" prop="state">
                        <template #default="{ row }">
                            <el-dropdown
                                placement="bottom"
                                @visible-change="
                                    (visible) => handleStatusDropdownVisibleChange(row.containerID, visible)
                                "
                            >
                                <Status v-permission :status="row.state" :operate="true" />
                                <template #dropdown>
                                    <el-dropdown-menu v-if="activeDropdownContainerId === row.containerID">
                                        <el-dropdown-item
                                            :disabled="checkStatus('start', row)"
                                            @click="onOperate('start', row)"
                                        >
                                            {{ $t('commons.operate.start') }}
                                        </el-dropdown-item>
                                        <el-dropdown-item
                                            :disabled="checkStatus('stop', row)"
                                            @click="onOperate('stop', row)"
                                        >
                                            {{ $t('commons.operate.stop') }}
                                        </el-dropdown-item>
                                        <el-dropdown-item
                                            :disabled="checkStatus('restart', row)"
                                            @click="onOperate('restart', row)"
                                        >
                                            {{ $t('commons.button.restart') }}
                                        </el-dropdown-item>
                                        <el-dropdown-item
                                            :disabled="checkStatus('kill', row)"
                                            @click="onOperate('kill', row)"
                                        >
                                            {{ $t('container.kill') }}
                                        </el-dropdown-item>
                                        <el-dropdown-item
                                            :disabled="checkStatus('pause', row)"
                                            @click="onOperate('pause', row)"
                                        >
                                            {{ $t('container.pause') }}
                                        </el-dropdown-item>
                                        <el-dropdown-item
                                            :disabled="checkStatus('unpause', row)"
                                            @click="onOperate('unpause', row)"
                                        >
                                            {{ $t('container.unpause') }}
                                        </el-dropdown-item>
                                    </el-dropdown-menu>
                                </template>
                            </el-dropdown>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('container.source')"
                        show-overflow-tooltip
                        prop="resource"
                        min-width="120"
                    >
                        <template #default="{ row }">
                            <div v-if="row.hasLoad">
                                <div class="source-font">CPU: {{ row.cpuPercent.toFixed(2) }}%</div>
                                <div class="float-left source-font">
                                    {{ $t('monitor.memory') }}: {{ row.memoryPercent.toFixed(2) }}%
                                </div>
                                <el-popover placement="right" width="500px" class="float-right">
                                    <template #reference>
                                        <svg-icon iconName="p-xiangqing" class="svg-icon"></svg-icon>
                                    </template>
                                    <template #default>
                                        <el-descriptions direction="vertical" border :column="3" size="small">
                                            <el-descriptions-item :label="$t('container.cpuUsage')">
                                                {{ computeCPU(row.cpuTotalUsage) }}
                                            </el-descriptions-item>
                                            <el-descriptions-item :label="$t('container.cpuTotal')">
                                                {{ computeCPU(row.systemUsage) }}
                                            </el-descriptions-item>
                                            <el-descriptions-item :label="$t('container.core')">
                                                {{ row.percpuUsage }}
                                            </el-descriptions-item>

                                            <el-descriptions-item :label="$t('container.memUsage')">
                                                {{ computeSizeForDocker(row.memoryUsage) }}
                                            </el-descriptions-item>
                                            <el-descriptions-item :label="$t('container.memCache')">
                                                {{ computeSizeForDocker(row.memoryCache) }}
                                            </el-descriptions-item>
                                            <el-descriptions-item :label="$t('container.memTotal')">
                                                {{ computeSizeForDocker(row.memoryLimit) }}
                                            </el-descriptions-item>
                                            <el-descriptions-item v-if="row.hasLoadSize">
                                                <template #label>
                                                    {{ $t('container.sizeRw') }}
                                                    <el-tooltip :content="$t('container.sizeRwHelper')">
                                                        <el-icon class="icon-item"><InfoFilled /></el-icon>
                                                    </el-tooltip>
                                                </template>
                                                {{ computeSize2(row.sizeRw) }}
                                            </el-descriptions-item>
                                            <el-descriptions-item
                                                :label="$t('container.sizeRootFs')"
                                                v-if="row.hasLoadSize"
                                            >
                                                <template #label>
                                                    {{ $t('container.sizeRootFs') }}
                                                    <el-tooltip :content="$t('container.sizeRootFsHelper')">
                                                        <el-icon class="icon-item"><InfoFilled /></el-icon>
                                                    </el-tooltip>
                                                </template>
                                                {{ computeSize2(row.sizeRootFs) }}
                                            </el-descriptions-item>
                                        </el-descriptions>

                                        <el-button
                                            class="mt-2"
                                            v-if="!row.hasLoadSize"
                                            size="small"
                                            link
                                            type="primary"
                                            @click="loadSize(row)"
                                        >
                                            {{ $t('container.loadSize') }}
                                        </el-button>
                                    </template>
                                </el-popover>
                            </div>
                            <div v-if="!row.hasLoad">
                                <el-button link loading></el-button>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('container.ip')"
                        :width="isMobile ? 120 : 'auto'"
                        min-width="120"
                        prop="network"
                    >
                        <template #default="{ row }">
                            <div v-if="row.network">
                                <div v-for="(item, index) in row.network" :key="index">{{ item }}</div>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('container.related')"
                        show-overflow-tooltip
                        min-width="210"
                        prop="appName"
                    >
                        <template #default="{ row }">
                            <el-button v-if="row.appName != '' || row.websites != null" link icon="Position" />
                            <el-text
                                v-if="row.appName != ''"
                                link
                                class="cursor-pointer"
                                size="small"
                                @click="routerToName('AppInstalled')"
                            >
                                {{ $t('app.app') }}: {{ row.appName }} [{{ row.appInstallName }}]
                            </el-text>
                            <el-text
                                v-if="row.websites != null"
                                link
                                class="cursor-pointer"
                                size="small"
                                @click="routerToName('Website')"
                            >
                                {{ $t('menu.website') }}:
                                {{ row.websites.join(',') }}
                            </el-text>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('commons.table.port')"
                        :width="isMobile ? 260 : 'auto'"
                        min-width="200"
                        prop="ports"
                    >
                        <template #default="{ row }">
                            <div v-if="row.ports">
                                <div v-for="(item, index) in row.ports" :key="index">
                                    <div v-if="row.expand || (!row.expand && index < 3)">
                                        <el-tooltip :hide-after="20" :content="item" placement="top">
                                            <el-button
                                                v-if="item.indexOf('->') !== -1"
                                                @click="goDashboard(item)"
                                                class="tagMargin"
                                                icon="Position"
                                                plain
                                                size="small"
                                            >
                                                {{ item.length > 25 ? item.substring(0, 25) + '...' : item }}
                                            </el-button>
                                            <el-button v-else class="tagMargin" plain size="small">
                                                {{ item }}
                                            </el-button>
                                        </el-tooltip>
                                    </div>
                                </div>
                                <div v-if="!row.expand && row.ports.length > 3">
                                    <el-button type="primary" link @click="row.expand = true">
                                        {{ $t('commons.button.expand') }}...
                                    </el-button>
                                </div>
                                <div v-if="row.expand && row.ports.length > 3">
                                    <el-button type="primary" link @click="row.expand = false">
                                        {{ $t('commons.button.collapse') }}
                                    </el-button>
                                </div>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column
                        min-width="200"
                        :label="$t('commons.table.description')"
                        prop="description"
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <fu-input-rw-switch
                                v-model="row.description"
                                v-permission
                                @enter="changePinned(row, false)"
                                @blur="changePinned(row, false)"
                            />
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('container.upTime')"
                        min-width="200"
                        show-overflow-tooltip
                        prop="runTime"
                    />
                    <fu-table-operations
                        fix
                        width="220px"
                        :ellipsis="2"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        :fixed="isMobile ? false : 'right'"
                        prop="operate"
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <OpDialog ref="opRef" @search="search" @submit="onSubmitOperate" />

        <ContainerInspectDialog ref="containerInspectRef" />
        <PruneDialog @search="search" ref="dialogPruneRef" />

        <RenameDialog @search="search" ref="dialogRenameRef" />
        <ContainerLogDialog ref="dialogContainerLogRef" :highlightDiff="210" />
        <UpgradeDialog @search="search" ref="dialogUpgradeRef" />
        <CommitDialog @search="search" ref="dialogCommitRef" />
        <MonitorDialog ref="dialogMonitorRef" />
        <TerminalDialog ref="dialogTerminalRef" />
        <ContainerFileDrawer ref="dialogFileBrowserRef" />

        <PortJumpDialog ref="dialogPortJumpRef" />
        <Backups ref="dialogBackupRef" />
        <Uploads ref="uploadRef" @close="search" />
        <TaskLog ref="taskLogRef" width="70%" @close="search" />
    </div>
</template>

<script lang="ts" setup>
import PruneDialog from '@/views/container/container/prune/index.vue';
import RenameDialog from '@/views/container/container/rename/index.vue';
import UpgradeDialog from '@/views/container/container/upgrade/index.vue';
import CommitDialog from '@/views/container/container/commit/index.vue';
import MonitorDialog from '@/views/container/container/monitor/index.vue';
import TerminalDialog from '@/views/container/container/terminal/index.vue';
import ContainerFileDrawer from '@/views/container/container/file-browser/index.vue';
import ContainerInspectDialog from '@/views/container/container/inspect/index.vue';
import PortJumpDialog from '@/components/port-jump/index.vue';
import TaskLog from '@/components/log/task/index.vue';
import Backups from '@/components/backup/index.vue';
import Uploads from '@/components/upload/index.vue';
import DockerStatus from '@/views/container/docker-status/index.vue';
import ContainerLogDialog from '@/components/log/container-drawer/index.vue';
import Status from '@/components/status/index.vue';
import { reactive, onMounted, ref } from 'vue';
import {
    containerItemStats,
    containerListStats,
    containerOperator,
    inspect,
    loadContainerStatus,
    searchContainer,
} from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import i18n from '@/lang';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { routerToName, routerToNameWithQuery } from '@/utils/router';
import router from '@/routers';
import { computeSize2, computeSizeForDocker, computeCPU } from '@/utils/size';
import { newUUID } from '@/utils/id';
import { updateCommonDescription } from '@/api/modules/setting';

const { currentNode, isAdminOrNodeAdmin, isMobile } = useGlobalStore();

const isActive = ref(false);
const isExist = ref(false);

const loading = ref(false);
const data = ref<any[]>([]);
const selects = ref<any>([]);
const paginationConfig = reactive({
    cacheSizeKey: 'container-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('container-page-size')) || 20,
    total: 0,
    state: 'all',
    orderBy: 'createdAt',
    order: 'null',
});
const searchName = ref();
const dialogUpgradeRef = ref();
const dialogCommitRef = ref();
const dialogPortJumpRef = ref();
const dialogBackupRef = ref();
const opRef = ref();
const includeAppStore = ref(true);
const columns = ref([]);

const batchNames = ref();
const batchOp = ref();
const taskLogRef = ref();

const tags = ref([]);
const activeTag = ref('all');

const hoveredRowIndex = ref(-1);
const activeDropdownContainerId = ref('');
const statFields = [
    'cpuTotalUsage',
    'systemUsage',
    'cpuPercent',
    'percpuUsage',
    'memoryCache',
    'memoryUsage',
    'memoryLimit',
    'memoryPercent',
] as const;

const assignFields = (target: Record<string, any>, source: Record<string, any>, skipKeys: string[] = []) => {
    const skipSet = new Set(skipKeys);
    for (const [key, value] of Object.entries(source)) {
        if (skipSet.has(key)) {
            continue;
        }
        if (target[key] !== value) {
            target[key] = value;
        }
    }
};

const syncContainerRows = (containers: Record<string, any>[]) => {
    const currentMap = new Map(data.value.map((item) => [item.containerID, item]));
    data.value = containers.map((container) => {
        const current = currentMap.get(container.containerID);
        if (!current) {
            return container;
        }
        assignFields(current, container);
        return current;
    });
};

const applyStatsToRows = (stats: Record<string, any>[]) => {
    if (stats.length === 0 || data.value.length === 0) {
        return;
    }
    const statsMap = new Map(stats.map((item) => [item.containerID, item]));
    for (const container of data.value) {
        const stat = statsMap.get(container.containerID);
        if (!stat) {
            continue;
        }
        if (!container.hasLoad) {
            container.hasLoad = true;
        }
        for (const field of statFields) {
            if (container[field] !== stat[field]) {
                container[field] = stat[field];
            }
        }
    }
};

const updateTags = (status: Record<string, any>) => {
    const nextTags = [];
    if (status.containerCount) {
        nextTags.push({ key: 'all', count: status.containerCount });
    }
    if (status.running) {
        nextTags.push({ key: 'running', count: status.running });
    }
    if (status.paused) {
        nextTags.push({ key: 'paused', count: status.paused });
    }
    if (status.restarting) {
        nextTags.push({ key: 'restarting', count: status.restarting });
    }
    if (status.removing) {
        nextTags.push({ key: 'removing', count: status.removing });
    }
    if (status.created) {
        nextTags.push({ key: 'created', count: status.created });
    }
    if (status.dead) {
        nextTags.push({ key: 'dead', count: status.dead });
    }
    if (status.exited) {
        nextTags.push({ key: 'exited', count: status.exited });
    }
    tags.value = nextTags;
};

const handleStatusDropdownVisibleChange = (containerID: string, visible: boolean) => {
    if (visible) {
        activeDropdownContainerId.value = containerID;
        return;
    }
    if (activeDropdownContainerId.value === containerID) {
        activeDropdownContainerId.value = '';
    }
};

const goDashboard = async (port: any) => {
    if (port.indexOf('127.0.0.1') !== -1) {
        MsgWarning(i18n.global.t('container.unExposedPort'));
        return;
    }
    if (port.indexOf(':') === -1) {
        MsgWarning(i18n.global.t('commons.msg.errPort'));
        return;
    }
    let portEx = port.match(/:(\d+)/)[1];

    let matches = port.match(new RegExp(':', 'g'));
    let ip = matches && matches.length > 1 ? 'ipv6' : 'ipv4';
    dialogPortJumpRef.value.acceptParams({ port: portEx, ip: ip });
};

interface Filters {
    filters?: string;
}
const props = withDefaults(defineProps<Filters>(), {
    filters: '',
});

const containerInspectRef = ref();

const dialogContainerLogRef = ref();
const dialogRenameRef = ref();
const dialogPruneRef = ref();

const search = async (column?: any) => {
    if (!isActive.value || !isExist.value) {
        return;
    }
    localStorage.setItem('includeAppStore', includeAppStore.value ? 'true' : 'false');
    let filterItem = (router.currentRoute.value.query?.filters as string) || '';
    paginationConfig.orderBy = column?.order ? column.prop : paginationConfig.orderBy;
    paginationConfig.order = column?.order ? column.order : paginationConfig.order;
    let params = {
        name: searchName.value,
        state: paginationConfig.state || 'all',
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        filters: filterItem,
        orderBy: paginationConfig.orderBy,
        order: paginationConfig.order,
        excludeAppStore: !includeAppStore.value,
    };
    loading.value = true;
    const [containerResult, statsResult, statusResult] = await Promise.allSettled([
        searchContainer(params),
        containerListStats(),
        loadContainerStatus(),
    ]);
    loading.value = false;

    if (containerResult.status === 'fulfilled') {
        const containers = containerResult.value.data.items || [];
        syncContainerRows(containers);
        paginationConfig.total = containerResult.value.data.total;
    }

    if (statsResult.status === 'fulfilled') {
        applyStatsToRows(statsResult.value.data || []);
    }

    if (statusResult.status === 'fulfilled') {
        updateTags(statusResult.value.data || {});
    }
};

const searchWithStatus = (item: string) => {
    activeTag.value = item;
    paginationConfig.state = activeTag.value;
    search();
};

const searchWithAppShow = (item: any) => {
    includeAppStore.value = item;
    search();
};

const showFavorite = (row: any) => {
    hoveredRowIndex.value = data.value.findIndex((item) => item === row);
};
const hideFavorite = () => {
    hoveredRowIndex.value = -1;
};
const changePinned = (row: any, isPinned: boolean) => {
    let params = {
        id: row.containerID,
        type: 'container',
        detailType: '',
        isPinned: row.isPinned,
        description: row.description || '',
    };
    if (isPinned) {
        params.isPinned = !row.isPinned;
    }
    updateCommonDescription(params).then(() => {
        search();
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    });
};

const refresh = async () => {
    let filterItem = props.filters ? props.filters : '';
    let params = {
        name: searchName.value,
        state: paginationConfig.state || 'all',
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        filters: filterItem,
        orderBy: paginationConfig.orderBy,
        order: paginationConfig.order,
    };
    const [containerResult, statsResult] = await Promise.all([searchContainer(params), containerListStats()]);
    syncContainerRows(containerResult.data.items || []);
    applyStatsToRows(statsResult.data || []);
};

const loadSize = async (row: any) => {
    containerItemStats(row.containerID).then((res) => {
        row.sizeRw = res.data.sizeRw || 0;
        row.sizeRootFs = res.data.sizeRootFs || 0;
        row.hasLoadSize = true;
    });
};

const onContainerOperate = async (container: string) => {
    routerToNameWithQuery('ContainerCreate', { name: container });
};

const onBackup = (row: Container.ContainerInfo) => {
    dialogBackupRef.value!.acceptParams({
        type: 'container',
        name: row.name,
        detailName: '',
    });
};

const uploadRef = ref();
const onImportCreate = () => {
    uploadRef.value!.acceptParams({
        type: 'container',
        name: '',
        detailName: '',
        remark: '.tar.gz',
        node: currentNode.value,
    });
};

const dialogMonitorRef = ref();
const onMonitor = (row: any) => {
    dialogMonitorRef.value!.acceptParams({ containerID: row.containerID, container: row.name });
};

const dialogTerminalRef = ref();
const onTerminal = (row: any) => {
    const title = i18n.global.t('menu.container') + ' ' + row.name;
    dialogTerminalRef.value!.acceptParams({ containerID: row.containerID, title: title });
};
const dialogFileBrowserRef = ref();
const onOpenFileBrowser = async (row: any) => {
    const title = i18n.global.t('menu.container') + ' ' + row.name;
    let workingDir = '/';
    try {
        const res = await inspect({ id: row.containerID, type: 'container', detail: '' });
        const data = typeof res.data === 'string' ? JSON.parse(res.data) : res.data;
        if (data?.Config?.WorkingDir) {
            workingDir = data.Config.WorkingDir;
        }
    } catch (e) {
        /* fallback to root */
    }
    dialogFileBrowserRef.value!.acceptParams({ containerID: row.containerID, title: title, workingDir: workingDir });
};

const onInspect = async (row: any) => {
    const res = await inspect({ id: row.containerID, type: 'container', detail: '' });
    containerInspectRef.value!.acceptParams({ data: res.data, ports: row.ports });
};

const onClean = () => {
    dialogPruneRef.value!.acceptParams();
};

const checkStatus = (operation: string, row: Container.ContainerInfo | null) => {
    let opList = row ? [row] : selects.value;
    if (opList.length < 1) {
        return true;
    }
    switch (operation) {
        case 'start':
            for (const item of opList) {
                if (item.state === 'running') {
                    return true;
                }
            }
            return false;
        case 'stop':
            for (const item of opList) {
                if (item.state === 'stopped' || item.state === 'exited') {
                    return true;
                }
            }
            return false;
        case 'pause':
            for (const item of opList) {
                if (item.state === 'paused' || item.state === 'exited') {
                    return true;
                }
            }
            return false;
        case 'unpause':
            for (const item of opList) {
                if (item.state !== 'paused') {
                    return true;
                }
            }
            return false;
    }
};

const onOperate = async (op: string, row: Container.ContainerInfo | null) => {
    let opList = row ? [row] : selects.value;
    let msg = i18n.global.t('container.operatorHelper', [i18n.global.t('container.' + op)]);
    batchNames.value = [];
    batchOp.value = op;
    for (const item of opList) {
        batchNames.value.push(item.name);
        if (item.isFromApp) {
            msg =
                op == 'remove'
                    ? i18n.global.t('container.containerDeleteHelper', [i18n.global.t('container.' + op)])
                    : i18n.global.t('container.operatorAppHelper', [i18n.global.t('container.' + op)]);
        }
    }
    const successMsg = `${i18n.global.t('container.' + op)}${i18n.global.t('commons.status.success')}`;
    opRef.value.acceptParams({
        title: i18n.global.t('container.' + op),
        names: batchNames.value,
        msg: msg,
        api: null,
        params: null,
        successMsg,
    });
};

const onSubmitOperate = async () => {
    loading.value = true;
    let taskID = newUUID();
    await containerOperator({ names: batchNames.value, operation: batchOp.value, taskID: taskID })
        .then(() => {
            loading.value = false;
            search();
            openTaskLog(taskID);
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};

const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const buttons = [
    {
        label: i18n.global.t('menu.terminal'),
        disabled: (row: Container.ContainerInfo) => {
            return row.state !== 'running' || !isAdminOrNodeAdmin.value;
        },
        click: (row: Container.ContainerInfo) => {
            onTerminal(row);
        },
    },
    {
        label: i18n.global.t('commons.button.log'),
        click: (row: Container.ContainerInfo) => {
            dialogContainerLogRef.value!.acceptParams({ containerID: row.containerID, container: row.name });
        },
    },
    {
        label: i18n.global.t('home.dir'),
        permission: true,
        disabled: (row: Container.ContainerInfo) => {
            return row.state !== 'running';
        },
        click: (row: Container.ContainerInfo) => {
            onOpenFileBrowser(row);
        },
    },
    {
        label: i18n.global.t('commons.button.edit'),
        permission: true,
        click: (row: Container.ContainerInfo) => {
            onContainerOperate(row.name);
        },
    },
    {
        label: i18n.global.t('commons.button.upgrade'),
        permission: true,
        click: (row: Container.ContainerInfo) => {
            dialogUpgradeRef.value!.acceptParams({ container: row.name, image: row.imageName, fromApp: row.isFromApp });
        },
    },
    {
        label: i18n.global.t('commons.button.backup'),
        permission: true,
        click: (row: Container.ContainerInfo) => {
            onBackup(row);
        },
    },
    {
        label: i18n.global.t('menu.monitor'),
        disabled: (row: Container.ContainerInfo) => {
            return row.state !== 'running';
        },
        click: (row: Container.ContainerInfo) => {
            onMonitor(row);
        },
    },
    {
        label: i18n.global.t('container.rename'),
        permission: true,
        click: (row: Container.ContainerInfo) => {
            dialogRenameRef.value!.acceptParams({ container: row.name });
        },
        disabled: (row: any) => {
            return row.isFromCompose;
        },
    },
    {
        label: i18n.global.t('container.makeImage'),
        permission: true,
        click: (row: Container.ContainerInfo) => {
            dialogCommitRef.value!.acceptParams({ containerID: row.containerID, containerName: row.name });
        },
        disabled: (row: any) => {
            return checkStatus('commit', row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
        click: (row: Container.ContainerInfo) => {
            onOperate('remove', row);
        },
        disabled: (row: any) => {
            return checkStatus('remove', row);
        },
    },
];

onMounted(() => {
    let includeItem = localStorage.getItem('includeAppStore');
    includeAppStore.value = !includeItem || includeItem === 'true';
});
</script>

<style scoped lang="scss">
.tagMargin {
    margin-top: 2px;
}
.source-font {
    font-size: 12px;
}
.svg-icon {
    margin-top: -3px;
    font-size: 6px;
    cursor: pointer;
}
.button-group .el-button {
    margin-left: -1px !important;
    position: relative !important;
    z-index: 1 !important;
}
.tag-button {
    margin-top: -5px;
    margin-right: 10px;
    &.no-active {
        background: none;
        border: none;
    }
}
.button-cell {
    width: 100%;
    max-width: 150px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
</style>
