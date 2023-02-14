<template>
    <div v-loading="loading">
        <div class="app-content" style="margin-top: 20px">
            <el-card class="app-card">
                <el-row :gutter="20">
                    <el-col :lg="3" :xl="2">
                        <div>
                            <el-tag effect="dark" type="success">{{ composeName }}</el-tag>
                        </div>
                    </el-col>
                    <el-col :lg="8" :xl="12">
                        <div v-if="createdBy === '1Panel'">
                            <el-button link type="primary" @click="onComposeOperate('start')">
                                {{ $t('container.start') }}
                            </el-button>
                            <el-divider direction="vertical" />
                            <el-button link type="primary" @click="onComposeOperate('stop')">
                                {{ $t('container.stop') }}
                            </el-button>
                            <el-divider direction="vertical" />
                            <el-button link type="primary" @click="onComposeOperate('down')">
                                {{ $t('container.remove') }}
                            </el-button>
                        </div>
                        <div v-else>
                            <el-alert
                                style="margin-top: -5px"
                                :closable="false"
                                show-icon
                                :title="$t('container.composeDetailHelper')"
                                type="info"
                            />
                        </div>
                    </el-col>
                </el-row>
            </el-card>
        </div>
        <LayoutContent
            v-loading="loading"
            style="margin-top: 30px"
            back-name="Compose"
            :title="$t('container.containerList')"
            :reload="true"
        >
            <template #toolbar>
                <el-button-group>
                    <el-button :disabled="checkStatus('start')" @click="onOperate('start')">
                        {{ $t('container.start') }}
                    </el-button>
                    <el-button :disabled="checkStatus('stop')" @click="onOperate('stop')">
                        {{ $t('container.stop') }}
                    </el-button>
                    <el-button :disabled="checkStatus('restart')" @click="onOperate('restart')">
                        {{ $t('container.restart') }}
                    </el-button>
                    <el-button :disabled="checkStatus('kill')" @click="onOperate('kill')">
                        {{ $t('container.kill') }}
                    </el-button>
                    <el-button :disabled="checkStatus('pause')" @click="onOperate('pause')">
                        {{ $t('container.pause') }}
                    </el-button>
                    <el-button :disabled="checkStatus('unpause')" @click="onOperate('unpause')">
                        {{ $t('container.unpause') }}
                    </el-button>
                    <el-button :disabled="checkStatus('remove')" @click="onOperate('remove')">
                        {{ $t('container.remove') }}
                    </el-button>
                </el-button-group>
            </template>
            <template #main>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    :data="data"
                    @search="search"
                >
                    <el-table-column type="selection" fix />
                    <el-table-column
                        :label="$t('commons.table.name')"
                        show-overflow-tooltip
                        min-width="100"
                        prop="name"
                        fix
                    >
                        <template #default="{ row }">
                            <el-link @click="onInspect(row.containerID)" type="primary">{{ row.name }}</el-link>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('container.image')"
                        show-overflow-tooltip
                        min-width="100"
                        prop="imageName"
                    />
                    <el-table-column :label="$t('commons.table.status')" min-width="50" prop="state" fix>
                        <template #default="{ row }">
                            <Status :key="row.state" :status="row.state"></Status>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('container.upTime')" min-width="100" prop="runTime" fix />
                    <el-table-column
                        prop="createTime"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                        show-overflow-tooltip
                    />
                    <fu-table-operations
                        width="220"
                        :ellipsis="10"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        fix
                    />
                </ComplexTable>

                <CodemirrorDialog ref="mydetail" />

                <ReNameDialog @search="search" ref="dialogReNameRef" />
                <ContainerLogDialog ref="dialogContainerLogRef" />
                <CreateDialog @search="search" ref="dialogCreateRef" />
                <MonitorDialog ref="dialogMonitorRef" />
                <TerminalDialog ref="dialogTerminalRef" />
            </template>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import LayoutContent from '@/layout/layout-content.vue';
import ReNameDialog from '@/views/container/container/rename/index.vue';
import CreateDialog from '@/views/container/container/create/index.vue';
import MonitorDialog from '@/views/container/container/monitor/index.vue';
import ContainerLogDialog from '@/views/container/container/log/index.vue';
import TerminalDialog from '@/views/container/container/terminal/index.vue';
import CodemirrorDialog from '@/components/codemirror-dialog/codemirror.vue';
import ComplexTable from '@/components/complex-table/index.vue';
import Status from '@/components/status/index.vue';
import { dateFormat } from '@/utils/util';
import { composeOperator, ContainerOperator, inspect, searchContainer } from '@/api/modules/container';
import { ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import { Container } from '@/api/interface/container';
import { MsgSuccess } from '@/utils/message';

const composeName = ref();
const composePath = ref();
const filters = ref();
const createdBy = ref();

const emit = defineEmits<{ (e: 'back'): void }>();
interface DialogProps {
    createdBy: string;
    name: string;
    path: string;
    filters: string;
}
const acceptParams = (props: DialogProps): void => {
    composePath.value = props.path;
    composeName.value = props.name;
    filters.value = props.filters;
    createdBy.value = props.createdBy;
    search();
};

const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const loading = ref(false);

const search = async () => {
    let filterItem = filters.value;
    let params = {
        name: '',
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        filters: filterItem,
    };
    await searchContainer(params).then((res) => {
        data.value = res.data.items || [];
        paginationConfig.total = res.data.total;
    });
};

const detailInfo = ref();
const mydetail = ref();
const onInspect = async (id: string) => {
    const res = await inspect({ id: id, type: 'container' });
    detailInfo.value = JSON.stringify(JSON.parse(res.data), null, 2);
    let param = {
        header: i18n.global.t('commons.button.view'),
        detailInfo: detailInfo.value,
    };
    mydetail.value!.acceptParams(param);
};

const checkStatus = (operation: string) => {
    if (selects.value.length < 1) {
        return true;
    }
    switch (operation) {
        case 'start':
            for (const item of selects.value) {
                if (item.state === 'running') {
                    return true;
                }
            }
            return false;
        case 'stop':
            for (const item of selects.value) {
                if (item.state === 'stopped' || item.state === 'exited') {
                    return true;
                }
            }
            return false;
    }
};

const onOperate = async (operation: string) => {
    ElMessageBox.confirm(
        i18n.global.t('container.operatorHelper', [i18n.global.t('container.' + operation)]),
        i18n.global.t('container.' + operation),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    ).then(() => {
        let ps = [];
        for (const item of selects.value) {
            const param = {
                name: item.name,
                operation: operation,
                newName: '',
            };
            ps.push(ContainerOperator(param));
        }
        Promise.all(ps)
            .then(() => {
                search();
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                search();
            });
    });
};

const onComposeOperate = async (operation: string) => {
    ElMessageBox.confirm(
        i18n.global.t('container.composeOperatorHelper', [composeName.value, i18n.global.t('container.' + operation)]),
        i18n.global.t('container.' + operation),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    ).then(async () => {
        let params = {
            name: composeName.value,
            path: composePath.value,
            operation: operation,
        };
        loading.value = true;
        await composeOperator(params)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                if (operation === 'down') {
                    emit('back');
                } else {
                    search();
                }
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const dialogMonitorRef = ref();
const onMonitor = (containerID: string) => {
    dialogMonitorRef.value!.acceptParams({ containerID: containerID });
};

const dialogTerminalRef = ref();
const onTerminal = (containerID: string) => {
    dialogTerminalRef.value!.acceptParams({ containerID: containerID });
};

const dialogContainerLogRef = ref();
const dialogReNameRef = ref();

const buttons = [
    {
        label: i18n.global.t('file.terminal'),
        disabled: (row: Container.ContainerInfo) => {
            return row.state !== 'running';
        },
        click: (row: Container.ContainerInfo) => {
            onTerminal(row.containerID);
        },
    },
    {
        label: i18n.global.t('container.monitor'),
        disabled: (row: Container.ContainerInfo) => {
            return row.state !== 'running';
        },
        click: (row: Container.ContainerInfo) => {
            onMonitor(row.containerID);
        },
    },
    {
        label: i18n.global.t('container.rename'),
        click: (row: Container.ContainerInfo) => {
            dialogReNameRef.value!.acceptParams({ containerID: row.containerID, container: row.name });
        },
    },
    {
        label: i18n.global.t('commons.button.log'),
        click: (row: Container.ContainerInfo) => {
            dialogContainerLogRef.value!.acceptParams({ containerID: row.containerID });
        },
    },
];

defineExpose({
    acceptParams,
});
</script>

<style lang="scss">
.app-card {
    font-size: 14px;
    height: 60px;
}

.app-content {
    height: 50px;
}

body {
    margin: 0;
}
</style>
