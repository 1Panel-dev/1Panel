<template>
    <div>
        <RouterMenu />
        <LayoutContent :title="'PHP'" v-loading="loading">
            <template #prompt></template>
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
                <ComplexTable :pagination-config="paginationConfig" :data="items" @search="search()" :heightDiff="350">
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
                    <el-table-column :label="$t('commons.table.status')" prop="status">
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
                    <el-table-column
                        prop="createdAt"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                        show-overflow-tooltip
                        min-width="120"
                        fix
                    />
                    <fu-table-operations
                        :ellipsis="mobile ? 0 : 5"
                        :min-width="mobile ? 'auto' : 300"
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
        <Log ref="logRef" @close="search" :heightDiff="280" />
        <Extensions ref="extensionsRef" @close="search" />
        <AppResources ref="checkRef" @close="search" />
        <ExtManagement ref="extManagementRef" />
        <ComposeLogs ref="composeLogRef" />
        <Config ref="configRef" />
        <Supervisor ref="supervisorRef" />
    </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { Runtime } from '@/api/interface/runtime';
import { DeleteRuntime, OperateRuntime, RuntimeDeleteCheck, SearchRuntimes } from '@/api/modules/runtime';
import { dateFormat } from '@/utils/util';
import { ElMessageBox } from 'element-plus';
import { containerPrune } from '@/api/modules/container';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';
import ExtManagement from './extension-management/index.vue';
import Extensions from './extension-template/index.vue';
import AppResources from '@/views/website/runtime/php/check/index.vue';
import CreateRuntime from '@/views/website/runtime/php/create/index.vue';
import RouterMenu from '../index.vue';
import Log from '@/components/log/drawer/index.vue';
import ComposeLogs from '@/components/log/compose/index.vue';
import Config from '@/views/website/runtime/php/config/index.vue';
import Supervisor from '@/views/website/runtime/php/supervisor/index.vue';
import RuntimeStatus from '@/views/website/runtime/components/runtime-status.vue';
import { disabledButton } from '@/utils/runtime';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();
const mobile = computed(() => {
    return globalStore.isMobile();
});

const paginationConfig = reactive({
    cacheSizeKey: 'runtime-page-size',
    currentPage: 1,
    pageSize: 10,
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

const buttons = [
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
        label: i18n.global.t('commons.operate.stop'),
        click: function (row: Runtime.Runtime) {
            operateRuntime('down', row.id);
        },
        disabled: function (row: Runtime.Runtime) {
            return disabledButton(row, 'stop');
        },
    },
    {
        label: i18n.global.t('commons.operate.start'),
        click: function (row: Runtime.Runtime) {
            operateRuntime('up', row.id);
        },
        disabled: function (row: Runtime.Runtime) {
            return disabledButton(row, 'start');
        },
    },
    {
        label: i18n.global.t('commons.button.restart'),
        click: function (row: Runtime.Runtime) {
            operateRuntime('restart', row.id);
        },
        disabled: function (row: Runtime.Runtime) {
            return disabledButton(row, 'restart');
        },
    },
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

const openLog = (row: Runtime.RuntimeDTO) => {
    if (row.status == 'Running') {
        composeLogRef.value.acceptParams({ compose: row.path + '/docker-compose.yml', resource: row.name });
    } else {
        logRef.value.acceptParams({ id: row.id, type: 'php', tail: row.status == 'Building', heightDiff: 220 });
    }
};

const openCreateLog = (id: number) => {
    logRef.value.acceptParams({ id: id, type: 'php', tail: true, heightDiff: 220 });
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

const operateRuntime = async (operate: string, ID: number) => {
    try {
        const action = await ElMessageBox.confirm(
            i18n.global.t('runtime.operatorHelper', [i18n.global.t('commons.button.' + operate)]),
            i18n.global.t('commons.button.' + operate),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                type: 'info',
            },
        );
        if (action === 'confirm') {
            loading.value = true;
            await OperateRuntime({ operate: operate, ID: ID });
            search();
        }
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

const onOpenBuildCache = () => {
    ElMessageBox.confirm(i18n.global.t('container.delBuildCacheHelper'), i18n.global.t('container.cleanBuildCache'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        let params = {
            pruneType: 'buildcache',
            withTagAll: false,
        };
        await containerPrune(params)
            .then((res) => {
                loading.value = false;
                MsgSuccess(i18n.global.t('container.cleanSuccess', [res.data.deletedNumber]));
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    });
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
