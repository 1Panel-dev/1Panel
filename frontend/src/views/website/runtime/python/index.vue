<template>
    <div>
        <RouterMenu />
        <LayoutContent :title="'Python'" v-loading="loading">
            <template #prompt>
                <el-alert type="info" :closable="false">
                    <template #title>
                        <span class="input-help whitespace-break-spaces">
                            {{ $t('runtime.statusHelper') }}
                        </span>
                    </template>
                </el-alert>
            </template>
            <template #toolbar>
                <div class="flex flex-wrap gap-3">
                    <el-button type="primary" @click="openCreate">
                        {{ $t('runtime.create') }}
                    </el-button>

                    <el-button type="primary" plain @click="onOpenBuildCache()">
                        {{ $t('container.cleanBuildCache') }}
                    </el-button>
                </div>
            </template>
            <template #main>
                <ComplexTable :pagination-config="paginationConfig" :data="items" @search="search()">
                    <el-table-column
                        :label="$t('commons.table.name')"
                        fix
                        prop="name"
                        min-width="120px"
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <el-text type="primary" class="cursor-pointer" @click="openDetail(row)">
                                {{ row.name }}
                            </el-text>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('website.runDir')" prop="codeDir" min-width="120px">
                        <template #default="{ row }">
                            <el-button type="primary" link @click="toFolder(row.codeDir)">
                                <el-icon>
                                    <FolderOpened />
                                </el-icon>
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('runtime.version')" prop="version"></el-table-column>
                    <el-table-column :label="$t('runtime.externalPort')" prop="port" min-width="110px">
                        <template #default="{ row }">
                            {{ row.port }}
                            <el-button link :icon="Promotion" @click="goDashboard(row.port, 'http')"></el-button>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.status')" prop="status">
                        <template #default="{ row }">
                            <el-popover
                                v-if="row.status === 'error'"
                                placement="bottom"
                                :width="400"
                                trigger="hover"
                                :content="row.message"
                                popper-class="max-h-[300px] overflow-auto"
                            >
                                <template #reference>
                                    <Status :key="row.status" :status="row.status"></Status>
                                </template>
                            </el-popover>
                            <div v-else>
                                <Status :key="row.status" :status="row.status"></Status>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.button.log')" prop="path" min-width="90px">
                        <template #default="{ row }">
                            <el-button @click="openLog(row)" link type="primary">{{ $t('website.check') }}</el-button>
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
                        :ellipsis="mobile ? 0 : 10"
                        :min-width="mobile ? 'auto' : 300"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        fixed="right"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>
        <Operate ref="operateRef" @close="search" />
        <Delete ref="deleteRef" @close="search" />
        <ComposeLogs ref="composeLogRef" />
        <PortJumpDialog ref="dialogPortJumpRef" />
        <AppResources ref="checkRef" @close="search" />
    </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, reactive, ref, computed } from 'vue';
import { Runtime } from '@/api/interface/runtime';
import { OperateRuntime, RuntimeDeleteCheck, SearchRuntimes, SyncRuntime } from '@/api/modules/runtime';
import { dateFormat } from '@/utils/util';
import Operate from '@/views/website/runtime/python/operate/index.vue';
import Status from '@/components/status/index.vue';
import Delete from '@/views/website/runtime/delete/index.vue';
import i18n from '@/lang';
import RouterMenu from '../index.vue';
import router from '@/routers/router';
import ComposeLogs from '@/components/compose-log/index.vue';
import { Promotion } from '@element-plus/icons-vue';
import PortJumpDialog from '@/components/port-jump/index.vue';
import AppResources from '@/views/website/runtime/php/check/index.vue';
import { ElMessageBox } from 'element-plus';
import { containerPrune } from '@/api/modules/container';
import { MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';

let timer: NodeJS.Timer | null = null;
const loading = ref(false);
const items = ref<Runtime.RuntimeDTO[]>([]);
const operateRef = ref();
const deleteRef = ref();
const dialogPortJumpRef = ref();
const composeLogRef = ref();
const checkRef = ref();

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
const req = reactive<Runtime.RuntimeReq>({
    name: '',
    page: 1,
    pageSize: 40,
    type: 'python',
});
const buttons = [
    {
        label: i18n.global.t('container.stop'),
        click: function (row: Runtime.Runtime) {
            operateRuntime('down', row.id);
        },
        disabled: function (row: Runtime.Runtime) {
            return row.status === 'recreating' || row.status === 'stopped';
        },
    },
    {
        label: i18n.global.t('container.start'),
        click: function (row: Runtime.Runtime) {
            operateRuntime('up', row.id);
        },
        disabled: function (row: Runtime.Runtime) {
            return row.status === 'starting' || row.status === 'recreating' || row.status === 'running';
        },
    },
    {
        label: i18n.global.t('container.restart'),
        click: function (row: Runtime.Runtime) {
            operateRuntime('restart', row.id);
        },
        disabled: function (row: Runtime.Runtime) {
            return row.status === 'recreating';
        },
    },
    {
        label: i18n.global.t('commons.button.edit'),
        click: function (row: Runtime.Runtime) {
            openDetail(row);
        },
        disabled: function (row: Runtime.Runtime) {
            return row.status === 'recreating';
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

const sync = () => {
    SyncRuntime();
};

const openCreate = () => {
    operateRef.value.acceptParams({ type: 'python', mode: 'create' });
};

const openDetail = (row: Runtime.Runtime) => {
    operateRef.value.acceptParams({ type: row.type, mode: 'edit', id: row.id });
};

const openDelete = async (row: Runtime.Runtime) => {
    RuntimeDeleteCheck(row.id).then(async (res) => {
        const items = res.data;
        if (res.data && res.data.length > 0) {
            checkRef.value.acceptParams({ items: items, key: 'website', installID: row.id });
        } else {
            deleteRef.value.acceptParams(row.id, row.name);
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

const openLog = (row: any) => {
    composeLogRef.value.acceptParams({ compose: row.path + '/docker-compose.yml', resource: row.name });
};

const goDashboard = async (port: any, protocol: string) => {
    dialogPortJumpRef.value.acceptParams({ port: port, protocol: protocol });
};

const operateRuntime = async (operate: string, ID: number) => {
    try {
        const action = await ElMessageBox.confirm(
            i18n.global.t('runtime.operatorHelper', [i18n.global.t('commons.operate.' + operate)]),
            i18n.global.t('commons.operate.' + operate),
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

const toFolder = (folder: string) => {
    router.push({ path: '/hosts/files', query: { path: folder } });
};

onMounted(() => {
    sync();
    search();
    timer = setInterval(() => {
        search();
        sync();
    }, 1000 * 10);
});

onUnmounted(() => {
    clearInterval(Number(timer));
    timer = null;
});
</script>

<style lang="scss" scoped></style>
