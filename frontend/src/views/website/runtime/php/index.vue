<template>
    <div>
        <RouterMenu />
        <DockerStatus v-model:isActive="isActive" v-model:isExist="isExist" />
        <LayoutContent v-loading="loading" v-if="isExist" :class="{ mask: !isActive }">
            <template #leftToolBar>
                <el-button type="primary" @click="openCreate">
                    {{ $t('runtime.create') }}
                </el-button>

                <el-button type="primary" plain @click="openExtensions">
                    {{ $t('php.extensions') }}
                </el-button>

                <el-button type="primary" plain @click="onOpenBuildCache()">
                    {{ $t('container.cleanBuildCache') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <TableRefresh @search="search()" />
                <TableSetting title="php-runtime-refresh" @search="search()" />
            </template>
            <template #main>
                <ComplexTable :pagination-config="paginationConfig" :data="items" @search="search()" :heightDiff="260">
                    <el-table-column
                        :label="$t('commons.table.name')"
                        fix
                        prop="name"
                        min-width="120px"
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <el-text
                                type="primary"
                                class="cursor-pointer"
                                @click="openDetail(row)"
                                v-if="row.status != 'building'"
                            >
                                {{ row.name }}
                            </el-text>
                            <el-text type="info" class="cursor-pointer" v-else>
                                {{ row.name }}
                            </el-text>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('home.dir')" prop="codeDir" width="80px">
                        <template #default="{ row }">
                            <el-button type="primary" link @click="routerToFileWithPath(row.path)">
                                <el-icon>
                                    <FolderOpened />
                                </el-icon>
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('app.source')" prop="resource">
                        <template #default="{ row }">
                            <span v-if="row.resource == 'appstore'">{{ $t('menu.apps') }}</span>
                            <span v-if="row.resource == 'local'">{{ $t('commons.table.local') }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('app.version')" prop="version">
                        <template #default="{ row }">{{ row.params['PHP_VERSION'] }}</template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('container.image')"
                        prop="image"
                        show-overflow-tooltip
                    ></el-table-column>
                    <el-table-column :label="$t('commons.table.port')" prop="port">
                        <template #default="{ row }">
                            {{ row.port }}
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.status')" prop="status" width="100px">
                        <template #default="{ row }">
                            <RuntimeStatus :row="row" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.button.log')" prop="">
                        <template #default="{ row }">
                            <el-button @click="openLog(row)" link type="primary" :disabled="row.resource == 'local'">
                                {{ $t('website.check') }}
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('website.remark')" prop="remark" min-width="150px">
                        <template #default="{ row }">
                            <fu-read-write-switch>
                                <template #read>
                                    <MsgInfo :info="row.remark" :width="'150'" />
                                </template>
                                <template #default="{ read }">
                                    <el-input v-model="row.remark" @blur="updateRuntimeRemark(row, read)" />
                                </template>
                            </fu-read-write-switch>
                        </template>
                    </el-table-column>
                    <el-table-column
                        prop="createdAt"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                        show-overflow-tooltip
                        width="180"
                        fix
                    />
                    <fu-table-operations
                        :ellipsis="mobile ? 0 : 3"
                        :width="mobile ? 'auto' : 200"
                        :buttons="buttons"
                        fixed="right"
                        :label="$t('commons.table.operate')"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <CreateRuntime ref="createRef" @close="search" @submit="openCreateLog" />
        <OpDialog ref="opRef" @search="search" />
        <Log ref="logRef" @close="search" :heightDiff="200" />
        <Extensions ref="extensionsRef" @close="search" />
        <AppResources ref="checkRef" @close="search" />
        <ExtManagement ref="extManagementRef" />
        <ComposeLogs ref="composeLogRef" :highlightDiff="200" />
        <Config ref="configRef" />
        <Supervisor ref="supervisorRef" />
        <Terminal ref="terminalRef" />
        <TaskLog ref="taskLogRef" width="70%" @close="search" />
    </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { Runtime } from '@/api/interface/runtime';
import { DeleteRuntime, RuntimeDeleteCheck, SearchRuntimes } from '@/api/modules/runtime';
import { dateFormat, newUUID } from '@/utils/util';
import { ElMessageBox } from 'element-plus';
import { containerPrune } from '@/api/modules/container';
import TaskLog from '@/components/log/task/index.vue';
import i18n from '@/lang';
import ExtManagement from './extension-management/index.vue';
import Extensions from './extension-template/index.vue';
import AppResources from '@/views/website/runtime/php/check/index.vue';
import CreateRuntime from '@/views/website/runtime/php/create/index.vue';
import RouterMenu from '../index.vue';
import Log from '@/components/log/file-drawer/index.vue';
import ComposeLogs from '@/components/log/compose/index.vue';
import Config from '@/views/website/runtime/php/config/index.vue';
import Supervisor from '@/views/website/runtime/php/supervisor/index.vue';
import RuntimeStatus from '@/views/website/runtime/components/runtime-status.vue';
import Terminal from '@/views/website/runtime/components/terminal.vue';
import { disabledButton } from '@/utils/runtime';
import { GlobalStore } from '@/store';
import DockerStatus from '@/views/container/docker-status/index.vue';
import { operateRuntime, updateRuntimeRemark } from '../common/utils';
import { routerToFileWithPath } from '@/utils/router';
const globalStore = GlobalStore();
const mobile = computed(() => {
    return globalStore.isMobile();
});

const taskLogRef = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'runtime-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('runtime-page-size')) || 20,
    total: 0,
});
let req = reactive<Runtime.RuntimeReq>({
    name: '',
    page: 1,
    pageSize: 40,
    type: 'php',
});
const opRef = ref();
const logRef = ref();
const extensionsRef = ref();
const extManagementRef = ref();
const checkRef = ref();
const createRef = ref();
const loading = ref(false);
const items = ref<Runtime.RuntimeDTO[]>([]);
const composeLogRef = ref();
const configRef = ref();
const supervisorRef = ref();
const terminalRef = ref();
const isActive = ref(false);
const isExist = ref(false);

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: function (row: Runtime.Runtime) {
            openDetail(row);
        },
        disabled: function (row: Runtime.Runtime) {
            return disabledButton(row, 'edit');
        },
    },
    {
        label: i18n.global.t('runtime.extension'),
        click: function (row: Runtime.Runtime) {
            openExtensionsManagement(row);
        },
        disabled: function (row: Runtime.Runtime) {
            return disabledButton(row, 'extension');
        },
    },
    {
        label: i18n.global.t('menu.terminal'),
        click: function (row: Runtime.Runtime) {
            openTerminal(row);
        },
        disabled: function (row: Runtime.Runtime) {
            return disabledButton(row, 'config');
        },
    },
    {
        label: i18n.global.t('commons.operate.stop'),
        click: function (row: Runtime.Runtime) {
            operateRuntime('down', row.id, loading, search);
        },
        disabled: function (row: Runtime.Runtime) {
            return disabledButton(row, 'stop');
        },
    },
    {
        label: i18n.global.t('commons.operate.start'),
        click: function (row: Runtime.Runtime) {
            operateRuntime('up', row.id, loading, search);
        },
        disabled: function (row: Runtime.Runtime) {
            return disabledButton(row, 'start');
        },
    },
    {
        label: i18n.global.t('commons.button.restart'),
        click: function (row: Runtime.Runtime) {
            operateRuntime('restart', row.id, loading, search);
        },
        disabled: function (row: Runtime.Runtime) {
            return disabledButton(row, 'restart');
        },
    },
    {
        label: i18n.global.t('menu.config'),
        click: function (row: Runtime.Runtime) {
            openConfig(row);
        },
        disabled: function (row: Runtime.Runtime) {
            return disabledButton(row, 'config');
        },
    },
    {
        label: i18n.global.t('menu.supervisor'),
        click: function (row: Runtime.Runtime) {
            openSupervisor(row);
        },
        disabled: function (row: Runtime.Runtime) {
            return disabledButton(row, 'config');
        },
    },

    {
        label: i18n.global.t('commons.button.delete'),
        click: function (row: Runtime.Runtime) {
            openDelete(row);
        },
    },
];

const search = async () => {
    req.page = paginationConfig.currentPage;
    req.pageSize = paginationConfig.pageSize;
    loading.value = true;
    try {
        const res = await SearchRuntimes(req);
        items.value = res.data.items;
        paginationConfig.total = res.data.total;
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

const openCreate = () => {
    createRef.value.acceptParams({ type: 'php', mode: 'create' });
};

const openDetail = (row: Runtime.Runtime) => {
    createRef.value.acceptParams({ type: row.type, mode: 'edit', id: row.id, appID: row.appID });
};

const openConfig = (row: Runtime.Runtime) => {
    configRef.value.acceptParams(row);
};

const openSupervisor = (row: Runtime.Runtime) => {
    supervisorRef.value.acceptParams(row.id);
};

const openTerminal = (row: Runtime.Runtime) => {
    const container = row.params['CONTAINER_NAME'];
    terminalRef.value.acceptParams({ containerID: container, container: container, user: 'www-data' });
};

const openLog = (row: Runtime.RuntimeDTO) => {
    if (row.status == 'Running') {
        composeLogRef.value.acceptParams({
            compose: row.path + '/docker-compose.yml',
            resource: row.name,
            container: row.container,
        });
    } else {
        logRef.value.acceptParams({ id: row.id, type: 'php', tail: row.status == 'Building' });
    }
};

const openCreateLog = (id: number) => {
    logRef.value.acceptParams({ id: id, type: 'php', tail: true });
};

const openExtensions = () => {
    extensionsRef.value.acceptParams();
};

const openExtensionsManagement = (row: Runtime.Runtime) => {
    extManagementRef.value.acceptParams(row);
};

const openDelete = async (row: Runtime.Runtime) => {
    RuntimeDeleteCheck(row.id).then(async (res) => {
        const items = res.data;
        if (res.data && res.data.length > 0) {
            checkRef.value.acceptParams({ items: items, key: 'website', installID: row.id });
        } else {
            opRef.value.acceptParams({
                title: i18n.global.t('commons.button.delete'),
                names: [row.name],
                msg: i18n.global.t('commons.msg.operatorHelper', [
                    i18n.global.t('website.runtime'),
                    i18n.global.t('commons.button.delete'),
                ]),
                api: DeleteRuntime,
                params: { id: row.id, forceDelete: true },
            });
        }
    });
};

const onOpenBuildCache = () => {
    ElMessageBox.confirm(i18n.global.t('container.delBuildCacheHelper'), i18n.global.t('container.cleanBuildCache'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        let params = {
            taskID: newUUID(),
            pruneType: 'buildcache',
            withTagAll: false,
        };
        await containerPrune(params)
            .then(() => {
                loading.value = false;
                openTaskLog(params.taskID);
            })
            .catch(() => {
                loading.value = false;
            });
    });
};
const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

onMounted(() => {
    search();
});
</script>

<style lang="scss" scoped>
.open-warn {
    color: $primary-color;
    cursor: pointer;
}
</style>
