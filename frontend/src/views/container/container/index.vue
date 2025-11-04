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
                <el-button type="primary" @click="onContainerOperate('')">
                    {{ $t('container.create') }}
                </el-button>
                <el-button type="primary" plain @click="onClean()">
                    {{ $t('container.containerPrune') }}
                </el-button>
                <el-button-group class="button-group">
                    <el-button :disabled="checkStatus('start', null)" @click="onOperate('start', null)">
                        {{ $t('commons.operate.start') }}
                    </el-button>
                    <el-button :disabled="checkStatus('stop', null)" @click="onOperate('stop', null)">
                        {{ $t('commons.operate.stop') }}
                    </el-button>
                    <el-button :disabled="checkStatus('restart', null)" @click="onOperate('restart', null)">
                        {{ $t('commons.button.restart') }}
                    </el-button>
                    <el-button :disabled="checkStatus('kill', null)" @click="onOperate('kill', null)">
                        {{ $t('container.kill') }}
                    </el-button>
                    <el-button :disabled="checkStatus('pause', null)" @click="onOperate('pause', null)">
                        {{ $t('container.pause') }}
                    </el-button>
                    <el-button :disabled="checkStatus('unpause', null)" @click="onOperate('unpause', null)">
                        {{ $t('container.unpause') }}
                    </el-button>
                    <el-button :disabled="checkStatus('remove', null)" @click="onOperate('remove', null)">
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
                    @sort-change="search"
                    @search="search"
                    :row-style="{ height: '65px' }"
                    style="width: 100%"
                    :columns="columns"
                    localKey="containerColumn"
                    :heightDiff="300"
                >
                    <el-table-column type="selection" />
                    <el-table-column
                        :label="$t('commons.table.name')"
                        :width="mobile ? 300 : 200"
                        min-width="100"
                        prop="name"
                        sortable
                        fix
                        :fixed="mobile ? false : 'left'"
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <el-text type="primary" class="cursor-pointer" @click="onInspect(row.containerID)">
                                {{ row.name }}
                            </el-text>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('container.image')"
                        show-overflow-tooltip
                        min-width="150"
                        prop="imageName"
                    />
                    <el-table-column :label="$t('commons.table.status')" min-width="100" prop="state" sortable>
                        <template #default="{ row }">
                            <el-dropdown placement="bottom">
                                <Status :key="row.state" :status="row.state" :operate="true"></Status>
                                <template #dropdown>
                                    <el-dropdown-menu>
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

                                            <el-descriptions-item>
                                                <template #label>
                                                    {{ $t('container.sizeRw') }}
                                                    <el-tooltip :content="$t('container.sizeRwHelper')">
                                                        <el-icon class="icon-item"><InfoFilled /></el-icon>
                                                    </el-tooltip>
                                                </template>
                                                {{ computeSize2(row.sizeRw) }}
                                            </el-descriptions-item>
                                            <el-descriptions-item :label="$t('container.sizeRootFs')">
                                                <template #label>
                                                    {{ $t('container.sizeRootFs') }}
                                                    <el-tooltip :content="$t('container.sizeRootFsHelper')">
                                                        <el-icon class="icon-item"><InfoFilled /></el-icon>
                                                    </el-tooltip>
                                                </template>
                                                {{ computeSize2(row.sizeRootFs) }}
                                            </el-descriptions-item>
                                        </el-descriptions>
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
                        :width="mobile ? 120 : 'auto'"
                        min-width="120"
                        prop="network"
                    >
                        <template #default="{ row }">
                            <div v-if="row.network">
                                <div v-for="(item, index) in row.network" :key="index">{{ item }}</div>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('container.related')" min-width="210" prop="appName">
                        <template #default="{ row }">
                            <div>
                                <el-tooltip
                                    v-if="row.appName != ''"
                                    :hide-after="20"
                                    :content="$t('app.app') + ': ' + row.appName + '[' + row.appInstallName + ']'"
                                    placement="top"
                                >
                                    <el-button icon="Position" plain size="small" @click="routerToName('AppInstalled')">
                                        {{ $t('app.app') }}: {{ row.appName }} [{{ row.appInstallName }}]
                                    </el-button>
                                </el-tooltip>
                            </div>
                            <div>
                                <el-tooltip
                                    v-if="row.websites != null"
                                    :hide-after="20"
                                    :content="row.websites.join(',')"
                                    placement="top"
                                    class="mt-1"
                                >
                                    <el-button icon="Position" plain size="small" @click="routerToName('Website')">
                                        {{ $t('menu.website') }}:
                                        {{ row.websites.join(',') }}
                                    </el-button>
                                </el-tooltip>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('commons.table.port')"
                        :width="mobile ? 260 : 'auto'"
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
                        :label="$t('container.upTime')"
                        min-width="200"
                        show-overflow-tooltip
                        prop="runTime"
                    />
                    <fu-table-operations
                        fix
                        width="200px"
                        :ellipsis="2"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        :fixed="mobile ? false : 'right'"
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

        <PortJumpDialog ref="dialogPortJumpRef" />
        <TaskLog ref="taskLogRef" width="70%" />
    </div>
</template>

<script lang="ts" setup>
import PruneDialog from '@/views/container/container/prune/index.vue';
import RenameDialog from '@/views/container/container/rename/index.vue';
import UpgradeDialog from '@/views/container/container/upgrade/index.vue';
import CommitDialog from '@/views/container/container/commit/index.vue';
import MonitorDialog from '@/views/container/container/monitor/index.vue';
import TerminalDialog from '@/views/container/container/terminal/index.vue';
import ContainerInspectDialog from '@/views/container/container/inspect/index.vue';
import PortJumpDialog from '@/components/port-jump/index.vue';
import TaskLog from '@/components/log/task/index.vue';
import DockerStatus from '@/views/container/docker-status/index.vue';
import ContainerLogDialog from '@/components/log/container-drawer/index.vue';
import Status from '@/components/status/index.vue';
import { reactive, onMounted, ref, computed } from 'vue';
import {
    containerListStats,
    containerOperator,
    inspect,
    loadContainerStatus,
    searchContainer,
} from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import i18n from '@/lang';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { GlobalStore } from '@/store';
import { routerToName, routerToNameWithQuery } from '@/utils/router';
import router from '@/routers';
import { computeSize2, computeSizeForDocker, computeCPU, newUUID } from '@/utils/util';
const globalStore = GlobalStore();

const mobile = computed(() => {
    return globalStore.isMobile();
});
const isActive = ref(false);
const isExist = ref(false);

const loading = ref(false);
const data = ref();
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
const opRef = ref();
const includeAppStore = ref(true);
const columns = ref([]);

const batchNames = ref();
const batchOp = ref();
const taskLogRef = ref();

const tags = ref([]);
const activeTag = ref('all');

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
    loadStats();
    loadContainerCount();
    await searchContainer(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
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

const loadContainerCount = async () => {
    await loadContainerStatus().then((res) => {
        tags.value = [];
        if (res.data.containerCount) {
            tags.value.push({ key: 'all', count: res.data.containerCount });
        }
        if (res.data.running) {
            tags.value.push({ key: 'running', count: res.data.running });
        }
        if (res.data.paused) {
            tags.value.push({ key: 'paused', count: res.data.paused });
        }
        if (res.data.restarting) {
            tags.value.push({ key: 'restarting', count: res.data.restarting });
        }
        if (res.data.removing) {
            tags.value.push({ key: 'removing', count: res.data.removing });
        }
        if (res.data.created) {
            tags.value.push({ key: 'created', count: res.data.created });
        }
        if (res.data.dead) {
            tags.value.push({ key: 'dead', count: res.data.dead });
        }
        if (res.data.exited) {
            tags.value.push({ key: 'exited', count: res.data.exited });
        }
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
    loadStats();
    const res = await searchContainer(params);
    let containers = res.data.items || [];
    for (const container of containers) {
        for (const c of data.value) {
            c.hasLoad = true;
            if (container.containerID == c.containerID) {
                for (let key in container) {
                    if (key !== 'cpuPercent' && key !== 'memoryPercent') {
                        c[key] = container[key];
                    }
                }
            }
        }
    }
};

const loadStats = async () => {
    const res = await containerListStats();
    let stats = res.data || [];
    if (stats.length === 0) {
        return;
    }
    for (const container of data.value) {
        for (const item of stats) {
            if (container.containerID === item.containerID) {
                container.hasLoad = true;
                container.cpuTotalUsage = item.cpuTotalUsage;
                container.systemUsage = item.systemUsage;
                container.cpuPercent = item.cpuPercent;
                container.percpuUsage = item.percpuUsage;
                container.memoryCache = item.memoryCache;
                container.memoryUsage = item.memoryUsage;
                container.memoryLimit = item.memoryLimit;
                container.memoryPercent = item.memoryPercent;
                break;
            }
        }
    }
};

const onContainerOperate = async (container: string) => {
    routerToNameWithQuery('ContainerCreate', { name: container });
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

const onInspect = async (id: string) => {
    const res = await inspect({ id: id, type: 'container' });
    containerInspectRef.value!.acceptParams({ data: res.data });
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
            msg = i18n.global.t('container.operatorAppHelper', [i18n.global.t('container.' + op)]);
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
            return row.state !== 'running';
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
        label: i18n.global.t('commons.button.edit'),
        click: (row: Container.ContainerInfo) => {
            onContainerOperate(row.name);
        },
    },
    {
        label: i18n.global.t('commons.button.upgrade'),
        click: (row: Container.ContainerInfo) => {
            dialogUpgradeRef.value!.acceptParams({ container: row.name, image: row.imageName, fromApp: row.isFromApp });
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
        click: (row: Container.ContainerInfo) => {
            dialogRenameRef.value!.acceptParams({ container: row.name });
        },
        disabled: (row: any) => {
            return row.isFromCompose;
        },
    },
    {
        label: i18n.global.t('container.makeImage'),
        click: (row: Container.ContainerInfo) => {
            dialogCommitRef.value!.acceptParams({ containerID: row.containerID, containerName: row.name });
        },
        disabled: (row: any) => {
            return checkStatus('commit', row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
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
</style>
