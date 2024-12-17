<template>
    <div>
        <el-card v-if="showStopped" class="mask-prompt">
            <span>{{ $t('tool.supervisor.notStartWarn') }}</span>
        </el-card>
        <LayoutContent :title="$t('tool.supervisor.list', 2)" v-loading="loading">
            <template #app>
                <SuperVisorStatus
                    @setting="setting"
                    v-model:loading="loading"
                    @get-status="getStatus"
                    v-model:mask-show="maskShow"
                />
            </template>
            <template v-if="showTable" #toolbar>
                <el-button type="primary" @click="openCreate">
                    {{ $t('commons.button.create') + $t('tool.supervisor.list').toLowerCase() }}
                </el-button>
            </template>
            <template #main v-if="showTable">
                <ComplexTable :data="data" :class="{ mask: !supervisorStatus.isRunning }" v-loading="dataLoading">
                    <el-table-column
                        :label="$t('commons.table.name')"
                        fix
                        prop="name"
                        min-width="80px"
                        show-overflow-tooltip
                    ></el-table-column>
                    <el-table-column
                        :label="$t('tool.supervisor.command')"
                        prop="command"
                        min-width="100px"
                        fix
                        show-overflow-tooltip
                    ></el-table-column>
                    <el-table-column
                        :label="$t('tool.supervisor.dir')"
                        prop="dir"
                        min-width="100px"
                        fix
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <el-button text type="primary" @click="toFolder(row.dir)">
                                {{ row.dir }}
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('tool.supervisor.user')"
                        prop="user"
                        show-overflow-tooltip
                        min-width="120"
                    ></el-table-column>
                    <el-table-column
                        :label="$t('tool.supervisor.numprocs')"
                        prop="numprocs"
                        min-width="120"
                    ></el-table-column>
                    <el-table-column :label="$t('tool.supervisor.manage')" min-width="120">
                        <template #default="{ row }">
                            <div v-if="row.status && row.status.length > 0 && row.hasLoad">
                                <el-button
                                    v-if="checkStatus(row.status) === 'RUNNING'"
                                    link
                                    type="success"
                                    :icon="VideoPlay"
                                    @click="operate('stop', row.name)"
                                >
                                    {{ $t('commons.status.running') }}
                                </el-button>
                                <el-button
                                    v-else-if="checkStatus(row.status) === 'WARNING'"
                                    link
                                    type="warning"
                                    :icon="RefreshRight"
                                    @click="operate('restart', row.name)"
                                >
                                    {{ $t('commons.status.unhealthy') }}
                                </el-button>
                                <el-button
                                    v-else
                                    link
                                    type="danger"
                                    :icon="VideoPause"
                                    @click="operate('start', row.name)"
                                >
                                    {{ $t('commons.status.stopped') }}
                                </el-button>
                            </div>
                            <div v-if="!row.hasLoad">
                                <el-button link loading></el-button>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.status')" min-width="60px">
                        <template #default="{ row }">
                            <div v-if="row.hasLoad">
                                <el-popover placement="bottom" :width="600" trigger="hover">
                                    <template #reference>
                                        <el-button type="primary" link v-if="row.status.length > 1">
                                            {{ $t('website.check') }}
                                        </el-button>
                                        <el-button type="primary" link v-else>
                                            <span>{{ $t('tool.supervisor.' + row.status[0].status) }}</span>
                                        </el-button>
                                    </template>
                                    <el-table :data="row.status">
                                        <el-table-column
                                            property="name"
                                            :label="$t('commons.table.name')"
                                            fix
                                            show-overflow-tooltip
                                        />
                                        <el-table-column
                                            property="status"
                                            :label="$t('tool.supervisor.statusCode')"
                                            width="100px"
                                        />
                                        <el-table-column property="PID" label="PID" width="100px" />
                                        <el-table-column
                                            property="uptime"
                                            :label="$t('tool.supervisor.uptime')"
                                            width="100px"
                                        />
                                        <el-table-column
                                            property="msg"
                                            :label="$t('tool.supervisor.msg')"
                                            fix
                                            show-overflow-tooltip
                                        />
                                    </el-table>
                                </el-popover>
                            </div>
                            <div v-if="!row.hasLoad">
                                <el-button link loading></el-button>
                            </div>
                        </template>
                    </el-table-column>
                    <fu-table-operations
                        :ellipsis="6"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        :fixed="mobile ? false : 'right'"
                        min-width="300"
                        fix
                    />
                </ComplexTable>
            </template>
            <ConfigSuperVisor v-if="setSuperVisor" />
        </LayoutContent>
        <Create ref="createRef" @close="search"></Create>
        <File ref="fileRef" @search="search"></File>
    </div>
</template>

<script setup lang="ts">
import SuperVisorStatus from './status/index.vue';
import { ref } from 'vue';
import ConfigSuperVisor from './config/index.vue';
import { computed, onMounted } from 'vue';
import Create from './create/index.vue';
import File from './file/index.vue';
import { GetSupervisorProcess, OperateSupervisorProcess } from '@/api/modules/host-tool';
import { GlobalStore } from '@/store';
import i18n from '@/lang';
import { HostTool } from '@/api/interface/host-tool';
import { MsgSuccess } from '@/utils/message';
import { VideoPlay, VideoPause, RefreshRight } from '@element-plus/icons-vue';
import router from '@/routers';
const globalStore = GlobalStore();

const loading = ref(false);
const setSuperVisor = ref(false);
const createRef = ref();
const fileRef = ref();
const data = ref();
const maskShow = ref(true);
const supervisorStatus = ref({
    maskShow: true,
    isExist: false,
    isRunning: false,
    init: true,
});
const dataLoading = ref(false);

const setting = () => {
    setSuperVisor.value = true;
};

const getStatus = (status: any) => {
    supervisorStatus.value = status;
    search();
};

const toFolder = (folder: string) => {
    router.push({ path: '/hosts/files', query: { path: folder } });
};

const showStopped = computed((): boolean => {
    if (supervisorStatus.value.init || setSuperVisor.value) {
        return false;
    }
    if (supervisorStatus.value.isExist && !supervisorStatus.value.isRunning && maskShow.value) {
        return true;
    }
    return false;
});

const showTable = computed((): boolean => {
    if (supervisorStatus.value.init || setSuperVisor.value || !supervisorStatus.value.isExist) {
        return false;
    }
    if (supervisorStatus.value.isExist && !setSuperVisor.value) {
        return true;
    }
    return true;
});

const openCreate = () => {
    createRef.value.acceptParams();
};

const search = async () => {
    if (!supervisorStatus.value.isExist) {
        return;
    }

    let needLoadStatus = false;
    dataLoading.value = true;
    try {
        const res = await GetSupervisorProcess();
        data.value = res.data;
        for (const process of data.value) {
            if (process.status && process.status.length > 0) {
                process.hasLoad = true;
            } else {
                process.hasLoad = false;
                needLoadStatus = true;
            }
        }
        if (supervisorStatus.value.isRunning && needLoadStatus) {
            setTimeout(loadStatus, 1000);
        }
    } catch (error) {
    } finally {
        dataLoading.value = false;
    }
};

const loadStatus = async () => {
    let needLoadStatus = false;
    try {
        const res = await GetSupervisorProcess();
        const stats = res.data || [];
        for (const process of data.value) {
            for (const item of stats) {
                if (process.name === item.name) {
                    if (item.status && item.status.length > 0) {
                        process.status = item.status;
                        process.hasLoad = true;
                    } else {
                        needLoadStatus = true;
                    }
                }
            }
        }
        if (supervisorStatus.value.isRunning && needLoadStatus) {
            setTimeout(loadStatus, 2000);
        }
    } catch (error) {}
};

const mobile = computed(() => {
    return globalStore.isMobile();
});

const checkStatus = (status: HostTool.ProcessStatus[]): string => {
    if (!status || status.length === 0) return 'STOPPED';

    const statusCounts = status.reduce((acc, curr) => {
        acc[curr.status] = (acc[curr.status] || 0) + 1;
        return acc;
    }, {} as Record<string, number>);

    if (statusCounts['STARTING']) return 'STARTING';
    if (statusCounts['RUNNING'] === status.length) return 'RUNNING';
    if (statusCounts['RUNNING'] > 0) return 'WARNING';
    return 'STOPPED';
};

const operate = async (operation: string, name: string) => {
    try {
        ElMessageBox.confirm(
            i18n.global.t('tool.supervisor.operatorHelper', [name, i18n.global.t('app.' + operation)]),
            i18n.global.t('app.' + operation),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                type: 'info',
            },
        )
            .then(() => {
                loading.value = true;
                OperateSupervisorProcess({ operate: operation, name: name })
                    .then(() => {
                        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                        search();
                    })
                    .catch(() => {})
                    .finally(() => {
                        loading.value = false;
                    });
            })
            .catch(() => {});
    } catch (error) {}
};

const getFile = (name: string, file: string) => {
    fileRef.value.acceptParams(name, file, 'get');
};

const edit = (row: HostTool.SupersivorProcess) => {
    createRef.value.acceptParams('update', row);
};

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: function (row: HostTool.SupersivorProcess) {
            edit(row);
        },
    },
    {
        label: i18n.global.t('website.sourceFile'),
        click: function (row: HostTool.SupersivorProcess) {
            getFile(row.name, 'config');
        },
    },
    {
        label: i18n.global.t('commons.button.log'),
        click: function (row: HostTool.SupersivorProcess) {
            getFile(row.name, 'out.log');
        },
    },
    {
        label: i18n.global.t('commons.button.restart'),
        click: function (row: HostTool.SupersivorProcess) {
            operate('restart', row.name);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: function (row: HostTool.SupersivorProcess) {
            operate('delete', row.name);
        },
    },
];

onMounted(() => {
    search();
});
</script>
