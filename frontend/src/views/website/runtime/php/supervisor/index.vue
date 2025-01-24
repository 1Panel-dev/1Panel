<template>
    <DrawerPro v-model="open" :header="$t('tool.supervisor.list')" size="60%" :back="handleClose">
        <template #content>
            <ComplexTable :data="data" v-loading="loading">
                <template #toolbar>
                    <el-button type="primary" @click="openCreate">
                        {{ $t('commons.button.create') + $t('tool.supervisor.list') }}
                    </el-button>
                </template>
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
                ></el-table-column>
                <el-table-column
                    :label="$t('tool.supervisor.user')"
                    prop="user"
                    show-overflow-tooltip
                    min-width="60px"
                ></el-table-column>
                <el-table-column
                    :label="$t('tool.supervisor.numprocs')"
                    prop="numprocs"
                    min-width="60px"
                ></el-table-column>
                <el-table-column :label="$t('tool.supervisor.manage')" min-width="80px">
                    <template #default="{ row }">
                        <div v-if="row.status && row.status.length > 0 && row.hasLoad">
                            <Status
                                v-if="checkStatus(row.status) === 'RUNNING'"
                                status="running"
                                @click="operate('stop', row.name)"
                            />
                            <Status
                                v-else-if="checkStatus(row.status) === 'WARNING'"
                                status="unhealthy"
                                @click="operate('restart', row.name)"
                            />
                            <Status v-else status="stopped" @click="operate('start', row.name)" />
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
                    width="280px"
                    fix
                />
            </ComplexTable>
        </template>
    </DrawerPro>
    <File ref="fileRef" @search="search" />
    <Create ref="createRef" @close="search" />
</template>

<script setup lang="ts">
import { ref } from '@vue/runtime-core';
import { computed } from 'vue';
import Create from './create/index.vue';
import File from './file/index.vue';
import { GetSupervisorProcess, operateSupervisorProcess } from '@/api/modules/runtime';
import { GlobalStore } from '@/store';
import i18n from '@/lang';
import { HostTool } from '@/api/interface/host-tool';
import { MsgSuccess } from '@/utils/message';
const globalStore = GlobalStore();

const loading = ref(false);
const fileRef = ref();
const data = ref();
const createRef = ref();
const dataLoading = ref(false);
const open = ref(false);
const runtimeID = ref(0);

const handleClose = () => {
    open.value = false;
};

const acceptParams = async (id: number) => {
    runtimeID.value = id;
    search();
    open.value = true;
};

const openCreate = () => {
    createRef.value.acceptParams('create', undefined, runtimeID.value);
};

const search = async () => {
    let needLoadStatus = false;
    dataLoading.value = true;
    try {
        const res = await GetSupervisorProcess(runtimeID.value);
        data.value = res.data;
        for (const process of data.value) {
            if (process.status && process.status.length > 0) {
                process.hasLoad = true;
            } else {
                process.hasLoad = false;
                needLoadStatus = true;
            }
        }
        if (needLoadStatus) {
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
        const res = await GetSupervisorProcess(runtimeID.value);
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
        if (needLoadStatus) {
            setTimeout(loadStatus, 20000);
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
                operateSupervisorProcess({ operate: operation, name: name, id: runtimeID.value })
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

const getFile = (name: string, file: string, runtimeID: number) => {
    fileRef.value.acceptParams(name, file, 'get', runtimeID);
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
        show: function (row: HostTool.SupersivorProcess) {
            return row.name != 'php-fpm';
        },
    },
    {
        label: i18n.global.t('website.proxyFile'),
        click: function (row: HostTool.SupersivorProcess) {
            getFile(row.name, 'config', runtimeID.value);
        },
        show: function (row: HostTool.SupersivorProcess) {
            return row.name != 'php-fpm';
        },
    },
    {
        label: i18n.global.t('website.log'),
        click: function (row: HostTool.SupersivorProcess) {
            getFile(row.name, 'out.log', runtimeID.value);
        },
        show: function (row: HostTool.SupersivorProcess) {
            return row.name != 'php-fpm';
        },
    },
    {
        label: i18n.global.t('commons.button.restart'),
        click: function (row: HostTool.SupersivorProcess) {
            operate('restart', row.name);
        },
        show: function (row: HostTool.SupersivorProcess) {
            return row.name != 'php-fpm';
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: function (row: HostTool.SupersivorProcess) {
            operate('delete', row.name);
        },
        show: function (row: HostTool.SupersivorProcess) {
            return row.name != 'php-fpm';
        },
    },
];

defineExpose({
    acceptParams,
});
</script>
