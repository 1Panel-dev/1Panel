<template>
    <div v-loading="loading">
        <LayoutContent
            back-name="Compose"
            :title="$t('container.containerList') + ' [ ' + composeName + ' ]'"
            :reload="true"
        >
            <template #main>
                <el-button-group>
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
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    style="margin-top: 20px"
                    :data="data"
                    @search="search"
                >
                    <el-table-column type="selection" fix />
                    <el-table-column
                        :label="$t('commons.table.name')"
                        min-width="100"
                        prop="name"
                        fix
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <el-button text type="primary" @click="onInspect(row.containerID)">
                                {{ row.name }}
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('container.image')"
                        show-overflow-tooltip
                        min-width="100"
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

                <CodemirrorDrawer ref="myDetail" />
                <OpDialog ref="opRef" @search="search" />

                <ContainerLogDialog ref="dialogContainerLogRef" />
                <MonitorDialog ref="dialogMonitorRef" />
                <TerminalDialog ref="dialogTerminalRef" />
            </template>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import MonitorDialog from '@/views/container/container/monitor/index.vue';
import ContainerLogDialog from '@/components/log/container-drawer/index.vue';
import TerminalDialog from '@/views/container/container/terminal/index.vue';
import CodemirrorDrawer from '@/components/codemirror-pro/drawer.vue';
import Status from '@/components/status/index.vue';
import { dateFormat } from '@/utils/util';
import { containerOperator, inspect, searchContainer } from '@/api/modules/container';
import i18n from '@/lang';
import { Container } from '@/api/interface/container';
import router from '@/routers';

const composeName = ref();
const dialogContainerLogRef = ref();
const opRef = ref();

const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    cacheSizeKey: 'container-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const loading = ref(false);

const search = async () => {
    let params = {
        name: '',
        state: 'all',
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        filters: 'com.docker.compose.project=' + composeName.value,
        orderBy: 'createdAt',
        order: 'null',
    };
    loading.value = true;
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

const detailInfo = ref();
const myDetail = ref();
const onInspect = async (id: string) => {
    const res = await inspect({ id: id, type: 'container' });
    detailInfo.value = JSON.stringify(JSON.parse(res.data), null, 2);
    let param = {
        header: i18n.global.t('commons.button.view'),
        detailInfo: detailInfo.value,
        mode: 'yaml',
    };
    myDetail.value!.acceptParams(param);
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

const dialogMonitorRef = ref();
const onMonitor = (row: any) => {
    dialogMonitorRef.value!.acceptParams({ containerID: row.containerID, container: row.name });
};

const dialogTerminalRef = ref();
const onTerminal = (row: any) => {
    const title = i18n.global.t('menu.container') + ' ' + row.name;
    dialogTerminalRef.value!.acceptParams({ containerID: row.containerID, title: title });
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
        label: i18n.global.t('menu.monitor'),
        disabled: (row: Container.ContainerInfo) => {
            return row.state !== 'running';
        },
        click: (row: Container.ContainerInfo) => {
            onMonitor(row);
        },
    },
    {
        label: i18n.global.t('commons.button.log'),
        click: (row: Container.ContainerInfo) => {
            dialogContainerLogRef.value!.acceptParams({ containerID: row.containerID, container: row.name });
        },
    },
];

onMounted(() => {
    if (router.currentRoute.value.query?.name) {
        composeName.value = router.currentRoute.value.query.name;
    }
    search();
});
</script>

<style lang="scss" scoped>
.app-content {
    height: 50px;
}
</style>
