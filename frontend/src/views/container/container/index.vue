<template>
    <div v-loading="loading">
        <docker-status v-model:isActive="isActive" @search="search" />

        <div class="mt-5" v-if="isActive">
            <el-tag @click="searchWithStatus('all')" v-if="countItem.all" effect="plain" size="large">
                {{ $t('commons.table.all') }} * {{ countItem.all }}
            </el-tag>
            <el-tag
                @click="searchWithStatus('running')"
                v-if="countItem.running"
                effect="plain"
                size="large"
                class="ml-2"
            >
                {{ $t('commons.status.running') }} * {{ countItem.running }}
            </el-tag>
            <el-tag
                @click="searchWithStatus('created')"
                v-if="countItem.created"
                effect="plain"
                size="large"
                class="ml-2"
            >
                {{ $t('commons.status.created') }} * {{ countItem.created }}
            </el-tag>
            <el-tag
                @click="searchWithStatus('paused')"
                v-if="countItem.paused"
                effect="plain"
                size="large"
                class="ml-2"
            >
                {{ $t('commons.status.paused') }} * {{ countItem.paused }}
            </el-tag>
            <el-tag
                @click="searchWithStatus('restarting')"
                v-if="countItem.restarting"
                effect="plain"
                size="large"
                class="ml-2"
            >
                {{ $t('commons.status.restarting') }} * {{ countItem.restarting }}
            </el-tag>
            <el-tag
                @click="searchWithStatus('removing')"
                v-if="countItem.removing"
                effect="plain"
                size="large"
                class="ml-2"
            >
                {{ $t('commons.status.removing') }} * {{ countItem.removing }}
            </el-tag>
            <el-tag
                @click="searchWithStatus('exited')"
                v-if="countItem.exited"
                effect="plain"
                size="large"
                class="ml-2"
            >
                {{ $t('commons.status.exited') }} * {{ countItem.exited }}
            </el-tag>
            <el-tag @click="searchWithStatus('dead')" v-if="countItem.dead" effect="plain" size="large" class="ml-2">
                {{ $t('commons.status.dead') }} * {{ countItem.dead }}
            </el-tag>
        </div>

        <LayoutContent :title="$t('container.container')" :class="{ mask: !isActive }">
            <template #leftToolBar>
                <el-button type="primary" @click="onContainerOperate('')">
                    {{ $t('container.create') }}
                </el-button>
                <el-button type="primary" plain @click="onClean()">
                    {{ $t('container.containerPrune') }}
                </el-button>
                <el-button-group>
                    <el-button :disabled="checkStatus('start', null)" @click="onOperate('start', null)">
                        {{ $t('container.start') }}
                    </el-button>
                    <el-button :disabled="checkStatus('stop', null)" @click="onOperate('stop', null)">
                        {{ $t('container.stop') }}
                    </el-button>
                    <el-button :disabled="checkStatus('restart', null)" @click="onOperate('restart', null)">
                        {{ $t('container.restart') }}
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
                        {{ $t('container.remove') }}
                    </el-button>
                </el-button-group>
            </template>
            <template #rightToolBar>
                <el-checkbox v-model="includeAppStore" @change="search()">
                    {{ $t('container.includeAppstore') }}
                </el-checkbox>
                <TableSearch @search="search()" v-model:searchName="searchName" />
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
                    <el-table-column :label="$t('commons.table.status')" min-width="100" prop="state">
                        <template #default="{ row }">
                            <el-dropdown placement="bottom">
                                <Status :key="row.state" :status="row.state"></Status>
                                <template #dropdown>
                                    <el-dropdown-menu>
                                        <el-dropdown-item
                                            :disabled="checkStatus('start', row)"
                                            @click="onOperate('start', row)"
                                        >
                                            {{ $t('container.start') }}
                                        </el-dropdown-item>
                                        <el-dropdown-item
                                            :disabled="checkStatus('stop', row)"
                                            @click="onOperate('stop', row)"
                                        >
                                            {{ $t('container.stop') }}
                                        </el-dropdown-item>
                                        <el-dropdown-item
                                            :disabled="checkStatus('restart', row)"
                                            @click="onOperate('restart', row)"
                                        >
                                            {{ $t('container.restart') }}
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
                                        <el-row>
                                            <el-col :span="8">
                                                <el-statistic
                                                    :title="$t('container.cpuUsage')"
                                                    :value="loadCPUValue(row.cpuTotalUsage)"
                                                    :precision="2"
                                                >
                                                    <template #suffix>{{ loadCPUUnit(row.cpuTotalUsage) }}</template>
                                                </el-statistic>
                                            </el-col>
                                            <el-col :span="8">
                                                <el-statistic
                                                    :title="$t('container.cpuTotal')"
                                                    :value="loadCPUValue(row.systemUsage)"
                                                    :precision="2"
                                                >
                                                    <template #suffix>{{ loadCPUUnit(row.systemUsage) }}</template>
                                                </el-statistic>
                                            </el-col>
                                            <el-col :span="8">
                                                <el-statistic :title="$t('container.core')" :value="row.percpuUsage" />
                                            </el-col>
                                        </el-row>

                                        <el-row class="mt-4">
                                            <el-col :span="8">
                                                <el-statistic
                                                    :title="$t('container.memUsage')"
                                                    :value="loadMemValue(row.memoryUsage)"
                                                    :precision="2"
                                                >
                                                    <template #suffix>{{ loadMemUnit(row.memoryUsage) }}</template>
                                                </el-statistic>
                                            </el-col>
                                            <el-col :span="8">
                                                <el-statistic
                                                    :title="$t('container.memCache')"
                                                    :value="loadMemValue(row.memoryCache)"
                                                    :precision="2"
                                                >
                                                    <template #suffix>{{ loadMemUnit(row.memoryCache) }}</template>
                                                </el-statistic>
                                            </el-col>
                                            <el-col :span="8">
                                                <el-statistic
                                                    :title="$t('container.memTotal')"
                                                    :value="loadMemValue(row.memoryLimit)"
                                                    :precision="2"
                                                >
                                                    <template #suffix>{{ loadMemUnit(row.memoryLimit) }}</template>
                                                </el-statistic>
                                            </el-col>
                                        </el-row>
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
                        :width="mobile ? 80 : 'auto'"
                        min-width="100"
                        prop="network"
                    >
                        <template #default="{ row }">
                            <div v-if="row.network">
                                <div v-for="(item, index) in row.network" :key="index">{{ item }}</div>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('container.related')" min-width="200" prop="appName">
                        <template #default="{ row }">
                            <div>
                                <el-tooltip
                                    v-if="row.appName != ''"
                                    :hide-after="20"
                                    :content="$t('app.app') + ': ' + row.appName + '[' + row.appInstallName + ']'"
                                    placement="top"
                                >
                                    <el-button
                                        icon="Position"
                                        plain
                                        size="small"
                                        @click="router.push({ name: 'AppInstalled' })"
                                    >
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
                                    <el-button
                                        icon="Position"
                                        plain
                                        size="small"
                                        @click="router.push({ name: 'Website' })"
                                    >
                                        {{ $t('website.website') }}:
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

        <OpDialog ref="opRef" @search="search" />

        <CodemirrorDialog ref="myDetail" />
        <PruneDialog @search="search" ref="dialogPruneRef" />

        <RenameDialog @search="search" ref="dialogRenameRef" />
        <ContainerLogDialog ref="dialogContainerLogRef" />
        <UpgradeDialog @search="search" ref="dialogUpgradeRef" />
        <CommitDialog @search="search" ref="dialogCommitRef" />
        <MonitorDialog ref="dialogMonitorRef" />
        <TerminalDialog ref="dialogTerminalRef" />

        <PortJumpDialog ref="dialogPortJumpRef" />
    </div>
</template>

<script lang="ts" setup>
import PruneDialog from '@/views/container/container/prune/index.vue';
import RenameDialog from '@/views/container/container/rename/index.vue';
import UpgradeDialog from '@/views/container/container/upgrade/index.vue';
import CommitDialog from '@/views/container/container/commit/index.vue';
import MonitorDialog from '@/views/container/container/monitor/index.vue';
import ContainerLogDialog from '@/views/container/container/log/index.vue';
import TerminalDialog from '@/views/container/container/terminal/index.vue';
import CodemirrorDialog from '@/components/codemirror-dialog/index.vue';
import PortJumpDialog from '@/components/port-jump/index.vue';
import DockerStatus from '@/views/container/docker-status/index.vue';
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
import router from '@/routers';
import { MsgWarning } from '@/utils/message';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const mobile = computed(() => {
    return globalStore.isMobile();
});
const isActive = ref(false);

const loading = ref(false);
const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    cacheSizeKey: 'container-page-size',
    currentPage: 1,
    pageSize: 10,
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
const includeAppStore = ref();
const columns = ref([]);

const countItem = reactive({
    all: 0,
    created: 0,
    running: 0,
    paused: 0,
    restarting: 0,
    removing: 0,
    exited: 0,
    dead: 0,
});

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

const myDetail = ref();

const dialogContainerLogRef = ref();
const dialogRenameRef = ref();
const dialogPruneRef = ref();

const search = async (column?: any) => {
    if (!isActive.value) {
        return;
    }
    localStorage.setItem('includeAppStore', includeAppStore.value ? 'true' : 'false');
    let filterItem = props.filters ? props.filters : '';
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

const searchWithStatus = (item: any) => {
    paginationConfig.state = item;
    search();
};

const loadContainerCount = async () => {
    await loadContainerStatus().then((res) => {
        countItem.all = res.data.all;
        countItem.running = res.data.running;
        countItem.paused = res.data.paused;
        countItem.restarting = res.data.restarting;
        countItem.removing = res.data.removing;
        countItem.created = res.data.created;
        countItem.dead = res.data.dead;
        countItem.exited = res.data.exited;
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

const loadCPUUnit = (t: number) => {
    const num = 1000;
    if (t < num) return ' ns';
    if (t < Math.pow(num, 2)) return ' μs';
    if (t < Math.pow(num, 3)) return ' ms';
    return ' s';
};
function loadCPUValue(t: number) {
    const num = 1000;
    if (t < num) return t;
    if (t < Math.pow(num, 2)) return Number((t / num).toFixed(2));
    if (t < Math.pow(num, 3)) return Number((t / Math.pow(num, 2)).toFixed(2));
    return Number((t / Math.pow(num, 3)).toFixed(2));
}
const loadMemUnit = (t: number) => {
    if (t == 0) {
        return '';
    }
    const num = 1024;
    if (t < num) return ' B';
    if (t < Math.pow(num, 2)) return ' KiB';
    if (t < Math.pow(num, 3)) return ' MiB';
    return ' GiB';
};
function loadMemValue(t: number) {
    const num = 1024;
    if (t < num) return t;
    if (t < Math.pow(num, 2)) return Number((t / num).toFixed(2));
    if (t < Math.pow(num, 3)) return Number((t / Math.pow(num, 2)).toFixed(2));
    return Number((t / Math.pow(num, 3)).toFixed(2));
}

const onContainerOperate = async (containerID: string) => {
    router.push({ name: 'ContainerCreate', query: { containerID: containerID } });
};

const dialogMonitorRef = ref();
const onMonitor = (row: any) => {
    dialogMonitorRef.value!.acceptParams({ containerID: row.containerID, container: row.name });
};

const dialogTerminalRef = ref();
const onTerminal = (row: any) => {
    dialogTerminalRef.value!.acceptParams({ containerID: row.containerID, container: row.name });
};

const onInspect = async (id: string) => {
    const res = await inspect({ id: id, type: 'container' });
    let detailInfo = JSON.stringify(JSON.parse(res.data), null, 2);
    let param = {
        header: i18n.global.t('commons.button.view'),
        detailInfo: detailInfo,
    };
    myDetail.value!.acceptParams(param);
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
    let names = [];
    for (const item of opList) {
        names.push(item.name);
        if (item.isFromApp) {
            msg = i18n.global.t('container.operatorAppHelper', [i18n.global.t('container.' + op)]);
        }
    }
    const successMsg = `${i18n.global.t('container.' + op)}${i18n.global.t('commons.status.success')}`;
    opRef.value.acceptParams({
        title: i18n.global.t('container.' + op),
        names: names,
        msg: msg,
        api: containerOperator,
        params: { names: names, operation: op },
        successMsg,
    });
};

const buttons = [
    {
        label: i18n.global.t('file.terminal'),
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
            onContainerOperate(row.containerID);
        },
    },
    {
        label: i18n.global.t('commons.button.upgrade'),
        click: (row: Container.ContainerInfo) => {
            dialogUpgradeRef.value!.acceptParams({ container: row.name, image: row.imageName, fromApp: row.isFromApp });
        },
    },
    {
        label: i18n.global.t('container.monitor'),
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
        label: i18n.global.t('container.remove'),
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
</style>
