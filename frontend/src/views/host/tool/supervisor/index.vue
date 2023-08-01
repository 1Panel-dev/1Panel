<template>
    <div>
        <ToolRouter />
        <el-card v-if="!isRunningSuperVisor && maskShow" class="mask-prompt">
            <span>{{ $t('firewall.firewallNotStart') }}</span>
        </el-card>
        <LayoutContent :title="$t('tool.supervisor.list')" v-loading="loading">
            <template #app>
                <SuperVisorStatus
                    @setting="setting"
                    v-model:loading="loading"
                    @is-exist="isExist"
                    @is-running="isRunning"
                    v-model:mask-show="maskShow"
                />
            </template>
            <template v-if="isExistSuperVisor && !setSuperVisor" #toolbar>
                <el-button type="primary" @click="openCreate">
                    {{ $t('commons.button.create') + $t('tool.supervisor.list') }}
                </el-button>
            </template>
            <template #main v-if="isExistSuperVisor && !setSuperVisor">
                <ComplexTable :data="data" :class="{ mask: !isRunningSuperVisor }">
                    <el-table-column :label="$t('commons.table.name')" fix prop="name"></el-table-column>
                    <el-table-column :label="$t('tool.supervisor.command')" prop="command"></el-table-column>
                    <el-table-column :label="$t('tool.supervisor.dir')" prop="dir"></el-table-column>
                    <el-table-column :label="$t('tool.supervisor.user')" prop="user"></el-table-column>
                    <el-table-column :label="$t('tool.supervisor.numprocs')" prop="numprocs"></el-table-column>
                    <el-table-column :label="$t('commons.table.status')">
                        <template #default="{ row }">
                            <div v-if="row.status">
                                <el-popover placement="left" :width="600" trigger="hover">
                                    <template #reference>
                                        <el-button type="primary" link v-if="row.status.length > 1">
                                            {{ $t('website.check') }}
                                        </el-button>
                                        <el-button type="primary" link v-else>
                                            <span>{{ row.status[0].status }}</span>
                                        </el-button>
                                    </template>
                                    <el-table :data="row.status">
                                        <el-table-column
                                            property="name"
                                            :label="$t('commons.table.name')"
                                            width="300"
                                        />
                                        <el-table-column property="status" :label="$t('commons.table.status')" />
                                        <el-table-column property="PID" label="PID" />
                                        <el-table-column property="uptime" :label="$t('tool.supervisor.uptime')" />
                                    </el-table>
                                </el-popover>
                            </div>
                        </template>
                    </el-table-column>
                    <fu-table-operations
                        :ellipsis="6"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        :fixed="mobile ? false : 'right'"
                        width="350px"
                        fix
                    />
                </ComplexTable>
            </template>
            <ConfigSuperVisor v-if="setSuperVisor" />
        </LayoutContent>
        <Create ref="createRef" @close="search"></Create>
        <File ref="fileRef"></File>
    </div>
</template>

<script setup lang="ts">
import ToolRouter from '@/views/host/tool/index.vue';
import SuperVisorStatus from './status/index.vue';
import { ref } from '@vue/runtime-core';
import ConfigSuperVisor from './config/index.vue';
import { computed, onMounted } from 'vue';
import Create from './create/index.vue';
import File from './file/index.vue';
import { GetSupervisorProcess, OperateSupervisorProcess } from '@/api/modules/host-tool';
import { GlobalStore } from '@/store';
import i18n from '@/lang';
import { HostTool } from '@/api/interface/host-tool';
import { MsgSuccess } from '@/utils/message';
const globalStore = GlobalStore();

const loading = ref(false);
const setSuperVisor = ref(false);
const isExistSuperVisor = ref(false);
const isRunningSuperVisor = ref(true);
const createRef = ref();
const fileRef = ref();
const data = ref();
const maskShow = ref(true);

const setting = () => {
    setSuperVisor.value = true;
};

const isExist = (isExist: boolean) => {
    isExistSuperVisor.value = isExist;
};

const isRunning = (running: boolean) => {
    isRunningSuperVisor.value = running;
};

const openCreate = () => {
    createRef.value.acceptParams();
};

const search = async () => {
    loading.value = true;
    try {
        const res = await GetSupervisorProcess();
        console.log(res);
        data.value = res.data;
    } catch (error) {}
    loading.value = false;
};

const mobile = computed(() => {
    return globalStore.isMobile();
});

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
    createRef.value.acceptParams('edit', row);
};

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: function (row: HostTool.SupersivorProcess) {
            edit(row);
        },
    },
    {
        label: i18n.global.t('website.proxyFile'),
        click: function (row: HostTool.SupersivorProcess) {
            getFile(row.name, 'config');
        },
    },
    {
        label: i18n.global.t('website.log'),
        click: function (row: HostTool.SupersivorProcess) {
            getFile(row.name, 'out.log');
        },
    },
    {
        label: i18n.global.t('app.start'),
        click: function (row: HostTool.SupersivorProcess) {
            operate('start', row.name);
        },
        disabled: (row: any) => {
            if (row.status == undefined) {
                return true;
            } else {
                return row.status && row.status[0].status == 'RUNNING';
            }
        },
    },
    {
        label: i18n.global.t('app.stop'),
        click: function (row: HostTool.SupersivorProcess) {
            operate('stop', row.name);
        },
        disabled: (row: any) => {
            if (row.status == undefined) {
                return true;
            }
            return row.status && row.status[0].status != 'RUNNING';
        },
    },
    {
        label: i18n.global.t('commons.button.restart'),
        click: function (row: HostTool.SupersivorProcess) {
            operate('restart', row.name);
        },
        disabled: (row: any): boolean => {
            if (row.status == undefined) {
                return true;
            }
            return row.status && row.status[0].status != 'RUNNING';
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
