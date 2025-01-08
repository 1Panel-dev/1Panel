<template>
    <div v-loading="loading">
        <el-card v-if="dockerStatus != 'Running'" class="mask-prompt">
            <span>{{ $t('container.serviceUnavailable') }}</span>
            <el-button type="primary" class="bt" link @click="goSetting">【 {{ $t('container.setting') }} 】</el-button>
            <span>{{ $t('container.startIn') }}</span>
        </el-card>
        <LayoutContent :title="$t('container.container', 2)" :class="{ mask: dockerStatus != 'Running' }">
            <template #rightButton>
                <div class="flex justify-end flex-col sm:flex-row">
                    <div class="mr-10">
                        <el-checkbox v-model="includeAppStore" @change="search()">
                            {{ $t('container.includeAppstore') }}
                        </el-checkbox>
                    </div>
                    <fu-table-column-select
                        :columns="columns"
                        trigger="hover"
                        :title="$t('commons.table.selectColumn')"
                        popper-class="popper-class"
                    />
                </div>
            </template>
            <template #toolbar>
                <div class="flex w-full flex-col gap-4 md:justify-between md:flex-row">
                    <div class="flex flex-wrap gap-4">
                        <el-button type="primary" @click="onOpenDialog('create')">
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
                    </div>
                    <div class="flex flex-row gap-2 md:flex-col lg:flex-row">
                        <TableSetting title="container-refresh" @search="refresh()" />
                        <TableSearch @search="search()" v-model:searchName="searchName" />
                    </div>
                </div>
            </template>
            <template #search>
                <el-select v-model="searchState" @change="search()" clearable class="p-w-200">
                    <template #prefix>{{ $t('commons.table.status') }}</template>
                    <el-option :label="$t('commons.table.all')" value="all"></el-option>
                    <el-option :label="$t('commons.status.created')" value="created"></el-option>
                    <el-option :label="$t('commons.status.running')" value="running"></el-option>
                    <el-option :label="$t('commons.status.paused')" value="paused"></el-option>
                    <el-option :label="$t('commons.status.restarting')" value="restarting"></el-option>
                    <el-option :label="$t('commons.status.removing')" value="removing"></el-option>
                    <el-option :label="$t('commons.status.exited')" value="exited"></el-option>
                    <el-option :label="$t('commons.status.dead')" value="dead"></el-option>
                </el-select>
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
                            <Status :key="row.state" :status="row.state"></Status>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('container.source')"
                        show-overflow-tooltip
                        prop="resource"
                        min-width="150"
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
                    <el-table-column :label="$t('container.related')" min-width="210" prop="appName">
                        <template #default="{ row }">
                            <div class="cell-button-class">
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
                            <div class="cell-button-class">
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
                            <div v-if="row.ports" class="cell-button-class">
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
        <OperateDialog @search="search" ref="dialogOperateRef" />
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
import OperateDialog from '@/views/container/container/operate/index.vue';
import UpgradeDialog from '@/views/container/container/upgrade/index.vue';
import CommitDialog from '@/views/container/container/commit/index.vue';
import MonitorDialog from '@/views/container/container/monitor/index.vue';
import ContainerLogDialog from '@/views/container/container/log/index.vue';
import TerminalDialog from '@/views/container/container/terminal/index.vue';
import CodemirrorDialog from '@/components/codemirror-dialog/index.vue';
import PortJumpDialog from '@/components/port-jump/index.vue';
import Status from '@/components/status/index.vue';
import { reactive, onMounted, ref, computed } from 'vue';
import {
    containerListStats,
    containerOperator,
    inspect,
    loadContainerInfo,
    loadDockerStatus,
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

const loading = ref(false);
const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    cacheSizeKey: 'container-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
    orderBy: 'created_at',
    order: 'null',
});
const searchName = ref();
const searchState = ref('all');
const dialogUpgradeRef = ref();
const dialogCommitRef = ref();
const dialogPortJumpRef = ref();
const opRef = ref();
const includeAppStore = ref();
const columns = ref([]);

const dockerStatus = ref('Running');
const loadStatus = async () => {
    loading.value = true;
    await loadDockerStatus()
        .then((res) => {
            loading.value = false;
            dockerStatus.value = res.data;
            if (dockerStatus.value === 'Running') {
                search();
            }
        })
        .catch(() => {
            dockerStatus.value = 'Failed';
            loading.value = false;
        });
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

const goSetting = async () => {
    router.push({ name: 'ContainerSetting' });
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
    localStorage.setItem('includeAppStore', includeAppStore.value ? 'true' : 'false');
    let filterItem = props.filters ? props.filters : '';
    paginationConfig.orderBy = column?.order ? column.prop : paginationConfig.orderBy;
    paginationConfig.order = column?.order ? column.order : paginationConfig.order;
    let params = {
        name: searchName.value,
        state: searchState.value || 'all',
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        filters: filterItem,
        orderBy: paginationConfig.orderBy,
        order: paginationConfig.order,
        excludeAppStore: !includeAppStore.value,
    };
    loading.value = true;
    loadStats();
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

const refresh = async () => {
    let filterItem = props.filters ? props.filters : '';
    let params = {
        name: searchName.value,
        state: searchState.value || 'all',
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

const dialogOperateRef = ref();
const onEdit = async (container: string) => {
    const res = await loadContainerInfo(container);
    if (res.data) {
        onOpenDialog('edit', res.data);
    }
};
const onOpenDialog = async (
    title: string,
    rowData: Partial<Container.ContainerHelper> = {
        cmd: [],
        cmdStr: '',
        publishAllPorts: false,
        exposedPorts: [],
        cpuShares: 1024,
        nanoCPUs: 0,
        memory: 0,
        memoryItem: 0,
        volumes: [],
        labels: [],
        env: [],
        restartPolicy: 'no',
    },
) => {
    let params = {
        title,
        rowData: { ...rowData },
    };
    dialogOperateRef.value!.acceptParams(params);
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
        label: i18n.global.t('container.containerTerminal'),
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
            onEdit(row.containerID);
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
        label: i18n.global.t('container.start'),
        click: (row: Container.ContainerInfo) => {
            onOperate('start', row);
        },
        disabled: (row: any) => {
            return checkStatus('start', row);
        },
    },
    {
        label: i18n.global.t('container.stop'),
        click: (row: Container.ContainerInfo) => {
            onOperate('stop', row);
        },
        disabled: (row: any) => {
            return checkStatus('stop', row);
        },
    },
    {
        label: i18n.global.t('container.restart'),
        click: (row: Container.ContainerInfo) => {
            onOperate('restart', row);
        },
        disabled: (row: any) => {
            return checkStatus('restart', row);
        },
    },
    {
        label: i18n.global.t('container.kill'),
        click: (row: Container.ContainerInfo) => {
            onOperate('kill', row);
        },
        disabled: (row: any) => {
            return checkStatus('kill', row);
        },
    },
    {
        label: i18n.global.t('container.pause'),
        click: (row: Container.ContainerInfo) => {
            onOperate('pause', row);
        },
        disabled: (row: any) => {
            return checkStatus('pause', row);
        },
    },
    {
        label: i18n.global.t('container.unpause'),
        click: (row: Container.ContainerInfo) => {
            onOperate('unpause', row);
        },
        disabled: (row: any) => {
            return checkStatus('unpause', row);
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
    loadStatus();
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
.cell-button-class {
    button,
    :deep(span) {
        max-width: 100%;
        overflow: hidden;
    }
}
</style>
