<template>
    <div>
        <el-card width="30%" v-if="dockerStatus != 'Running'" class="mask-prompt">
            <span style="font-size: 14px">{{ $t('container.serviceUnavailable') }}</span>
            <el-button type="primary" link style="font-size: 14px; margin-bottom: 5px" @click="goSetting">
                【 {{ $t('container.setting') }} 】
            </el-button>
            <span style="font-size: 14px">{{ $t('container.startIn') }}</span>
        </el-card>
        <LayoutContent
            v-loading="loading"
            :title="$t('container.container')"
            :class="{ mask: dockerStatus != 'Running' }"
        >
            <template #toolbar>
                <el-row>
                    <el-col :span="16">
                        <el-button type="primary" @click="onCreate()">
                            {{ $t('container.createContainer') }}
                        </el-button>
                        <el-button-group style="margin-left: 10px">
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
                    </el-col>
                    <el-col :span="8">
                        <TableSetting @search="search()" />
                        <div class="search-button">
                            <el-input
                                v-model="searchName"
                                clearable
                                @clear="search()"
                                suffix-icon="Search"
                                @keyup.enter="search()"
                                @blur="search()"
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
                    @search="search"
                >
                    <el-table-column type="selection" fix />
                    <el-table-column
                        :label="$t('commons.table.name')"
                        show-overflow-tooltip
                        min-width="80"
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
                        min-width="80"
                        prop="imageName"
                    />
                    <el-table-column :label="$t('commons.table.status')" min-width="50" prop="state" fix>
                        <template #default="{ row }">
                            <Status :key="row.state" :status="row.state"></Status>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('container.upTime')" min-width="80" prop="runTime" fix />
                    <el-table-column
                        prop="createTime"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                        show-overflow-tooltip
                    />
                    <fu-table-operations
                        width="220px"
                        :ellipsis="10"
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
        <CreateDialog @search="search" ref="dialogCreateRef" />
        <MonitorDialog ref="dialogMonitorRef" />
        <TerminalDialog ref="dialogTerminalRef" />
    </div>
</template>

<script lang="ts" setup>
import LayoutContent from '@/layout/layout-content.vue';
import ComplexTable from '@/components/complex-table/index.vue';
import TableSetting from '@/components/table-setting/index.vue';
import ReNameDialog from '@/views/container/container/rename/index.vue';
import CreateDialog from '@/views/container/container/create/index.vue';
import MonitorDialog from '@/views/container/container/monitor/index.vue';
import ContainerLogDialog from '@/views/container/container/log/index.vue';
import TerminalDialog from '@/views/container/container/terminal/index.vue';
import CodemirrorDialog from '@/components/codemirror-dialog/codemirror.vue';
import Status from '@/components/status/index.vue';
import { reactive, onMounted, ref } from 'vue';
import { dateFormat } from '@/utils/util';
import { ContainerOperator, inspect, loadDockerStatus, searchContainer } from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import { ElMessage, ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import router from '@/routers';

const loading = ref();
const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const searchName = ref();

const dockerStatus = ref();
const loadStatus = async () => {
    const res = await loadDockerStatus();
    dockerStatus.value = res.data;
    if (dockerStatus.value === 'Running') {
        search();
    }
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

const search = async () => {
    let filterItem = props.filters ? props.filters : '';
    let params = {
        name: searchName.value,
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        filters: filterItem,
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

const dialogCreateRef = ref();
const onCreate = () => {
    dialogCreateRef.value!.acceptParams();
};

const dialogMonitorRef = ref();
const onMonitor = (containerID: string) => {
    dialogMonitorRef.value!.acceptParams({ containerID: containerID });
};

const dialogTerminalRef = ref();
const onTerminal = (containerID: string) => {
    dialogTerminalRef.value!.acceptParams({ containerID: containerID });
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
        case 'pause':
            for (const item of selects.value) {
                if (item.state === 'paused' || item.state === 'exited') {
                    return true;
                }
            }
            return false;
        case 'unpause':
            for (const item of selects.value) {
                if (item.state !== 'paused') {
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
        loading.value = true;
        Promise.all(ps)
            .then(() => {
                loading.value = false;
                search();
                ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
                search();
            });
    });
};

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
            console.log(row.name);
            dialogReNameRef.value!.acceptParams({ container: row.name });
        },
    },
    {
        label: i18n.global.t('commons.button.log'),
        click: (row: Container.ContainerInfo) => {
            dialogContainerLogRef.value!.acceptParams({ containerID: row.containerID });
        },
    },
];

onMounted(() => {
    loadStatus();
});
</script>
