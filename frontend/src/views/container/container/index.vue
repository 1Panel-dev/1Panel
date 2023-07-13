<template>
    <div v-loading="loading">
        <el-card v-if="dockerStatus != 'Running'" class="mask-prompt">
            <span>{{ $t('container.serviceUnavailable') }}</span>
            <el-button type="primary" class="bt" link @click="goSetting">【 {{ $t('container.setting') }} 】</el-button>
            <span>{{ $t('container.startIn') }}</span>
        </el-card>
        <LayoutContent :title="$t('container.container')" :class="{ mask: dockerStatus != 'Running' }">
            <template #toolbar>
                <el-row>
                    <el-col :xs="24" :sm="16" :md="16" :lg="16" :xl="16">
                        <el-button type="primary" @click="onOpenDialog('create')">
                            {{ $t('container.create') }}
                        </el-button>
                        <el-button type="primary" plain @click="onClean()">
                            {{ $t('container.containerPrune') }}
                        </el-button>
                        <el-button-group style="margin-left: 10px">
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
                    </el-col>
                    <el-col :xs="24" :sm="8" :md="8" :lg="8" :xl="8">
                        <TableSetting @search="search()" />
                        <div class="search-button">
                            <el-input
                                v-model="searchName"
                                clearable
                                @clear="search()"
                                suffix-icon="Search"
                                @keyup.enter="search()"
                                @change="search()"
                                :placeholder="$t('commons.button.search')"
                            ></el-input>
                        </div>
                    </el-col>
                </el-row>
            </template>
            <template #main>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    :data="data"
                    @sort-change="search"
                    @search="search"
                >
                    <el-table-column type="selection" fix />
                    <el-table-column :label="$t('commons.table.name')" min-width="80" prop="name" sortable fix>
                        <template #default="{ row }">
                            <Tooltip @click="onInspect(row.containerID)" :text="row.name" />
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('container.image')"
                        show-overflow-tooltip
                        min-width="80"
                        prop="imageName"
                    />
                    <el-table-column :label="$t('commons.table.status')" min-width="60" prop="state" sortable fix>
                        <template #default="{ row }">
                            <Status :key="row.state" :status="row.state"></Status>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('container.source')" show-overflow-tooltip min-width="75" fix>
                        <template #default="{ row }">
                            <div v-if="row.hasLoad">
                                <div>CPU: {{ row.cpuPercent.toFixed(2) }}%</div>
                                <div>{{ $t('monitor.memory') }}: {{ row.memoryPercent.toFixed(2) }}%</div>
                            </div>
                            <div v-if="!row.hasLoad">
                                <el-button link loading></el-button>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.port')" min-width="120" prop="ports" fix>
                        <template #default="{ row }">
                            <div v-if="row.ports">
                                <div v-for="(item, index) in row.ports" :key="index">
                                    <div v-if="row.expand || (!row.expand && index < 3)">
                                        <el-button
                                            v-if="item.indexOf('->') !== -1"
                                            @click="goDashboard(item)"
                                            class="tagMargin"
                                            icon="Position"
                                            type="primary"
                                            plain
                                            size="small"
                                        >
                                            {{ item }}
                                        </el-button>
                                        <el-button v-else class="tagMargin" type="primary" plain size="small">
                                            {{ item }}
                                        </el-button>
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
                        min-width="70"
                        show-overflow-tooltip
                        prop="runTime"
                        fix
                    />
                    <fu-table-operations
                        width="300px"
                        :ellipsis="4"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <CodemirrorDialog ref="mydetail" />

        <ReNameDialog @search="search" ref="dialogReNameRef" />
        <ContainerLogDialog ref="dialogContainerLogRef" />
        <OperateDialog @search="search" ref="dialogOperateRef" />
        <UpgraeDialog @search="search" ref="dialogUpgradeRef" />
        <MonitorDialog ref="dialogMonitorRef" />
        <TerminalDialog ref="dialogTerminalRef" />

        <PortJumpDialog ref="dialogPortJumpRef" />
    </div>
</template>

<script lang="ts" setup>
import Tooltip from '@/components/tooltip/index.vue';
import TableSetting from '@/components/table-setting/index.vue';
import ReNameDialog from '@/views/container/container/rename/index.vue';
import OperateDialog from '@/views/container/container/operate/index.vue';
import UpgraeDialog from '@/views/container/container/upgrade/index.vue';
import MonitorDialog from '@/views/container/container/monitor/index.vue';
import ContainerLogDialog from '@/views/container/container/log/index.vue';
import TerminalDialog from '@/views/container/container/terminal/index.vue';
import CodemirrorDialog from '@/components/codemirror-dialog/index.vue';
import PortJumpDialog from '@/components/port-jump/index.vue';
import Status from '@/components/status/index.vue';
import { reactive, onMounted, ref } from 'vue';
import {
    containerListStats,
    containerOperator,
    containerPrune,
    inspect,
    loadContainerInfo,
    loadDockerStatus,
    searchContainer,
} from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import { ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import router from '@/routers';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { computeSize } from '@/utils/util';

const loading = ref();
const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const searchName = ref();
const dialogUpgradeRef = ref();
const dialogPortJumpRef = ref();

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
    if (!port || port.indexOf(':') === -1 || port.indexOf('->') === -1) {
        MsgWarning(i18n.global.t('commons.msg.errPort'));
        return;
    }
    let portEx = port.match(/:(\d+)/)[1];
    dialogPortJumpRef.value.acceptParams({ port: portEx });
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

const detailInfo = ref();
const mydetail = ref();

const dialogContainerLogRef = ref();
const dialogReNameRef = ref();

const search = async (column?: any) => {
    let filterItem = props.filters ? props.filters : '';
    let params = {
        name: searchName.value,
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        filters: filterItem,
        orderBy: column?.order ? column.prop : 'created_at',
        order: column?.order ? column.order : 'null',
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
                container.cpuPercent = item.cpuPercent;
                container.memoryPercent = item.memoryPercent;
                break;
            }
        }
    }
};

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
    detailInfo.value = JSON.stringify(JSON.parse(res.data), null, 2);
    let param = {
        header: i18n.global.t('commons.button.view'),
        detailInfo: detailInfo.value,
    };
    mydetail.value!.acceptParams(param);
};

const onClean = () => {
    ElMessageBox.confirm(i18n.global.t('container.containerPruneHelper'), i18n.global.t('container.containerPrune'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        let params = {
            pruneType: 'container',
            withTagAll: false,
        };
        await containerPrune(params)
            .then((res) => {
                loading.value = false;
                MsgSuccess(
                    i18n.global.t('container.cleanSuccessWithSpace', [
                        res.data.deletedNumber,
                        computeSize(res.data.spaceReclaimed),
                    ]),
                );
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    });
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
const onOperate = async (operation: string, row: Container.ContainerInfo | null) => {
    let opList = row ? [row] : selects.value;
    let msg = i18n.global.t('container.operatorHelper', [i18n.global.t('container.' + operation)]);
    for (const item of opList) {
        if (item.isFromApp) {
            msg = i18n.global.t('container.operatorAppHelper', [i18n.global.t('container.' + operation)]);
            break;
        }
    }
    ElMessageBox.confirm(msg, i18n.global.t('container.' + operation), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(() => {
        let ps = [];
        for (const item of opList) {
            const param = {
                name: item.name,
                operation: operation,
                newName: '',
            };
            ps.push(containerOperator(param));
        }
        loading.value = true;
        Promise.all(ps)
            .then(() => {
                loading.value = false;
                search();
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
                search();
            });
    });
};

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: Container.ContainerInfo) => {
            onEdit(row.containerID);
        },
    },
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
            dialogReNameRef.value!.acceptParams({ container: row.name });
        },
        disabled: (row: any) => {
            return row.isFromCompose;
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
    loadStatus();
});
</script>

<style scoped lang="scss">
.tagMargin {
    margin-top: 2px;
}
</style>
