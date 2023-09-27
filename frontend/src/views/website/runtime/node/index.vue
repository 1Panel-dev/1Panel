<template>
    <div>
        <RouterMenu />
        <LayoutContent :title="'Node.js'" v-loading="loading">
            <template #prompt>
                <el-alert type="info" :closable="false">
                    <template #default>
                        <span><span v-html="$t('runtime.statusHelper')"></span></span>
                    </template>
                </el-alert>
            </template>
            <template #toolbar>
                <el-button type="primary" @click="openCreate">
                    {{ $t('runtime.create') }}
                </el-button>
            </template>
            <template #main>
                <ComplexTable :pagination-config="paginationConfig" :data="items" @search="search()">
                    <el-table-column :label="$t('commons.table.name')" fix prop="name" min-width="120px">
                        <template #default="{ row }">
                            <Tooltip :text="row.name" @click="openDetail(row)" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('runtime.codeDir')" prop="codeDir">
                        <template #default="{ row }">
                            <el-button type="primary" link @click="toFolder(row.codeDir)">
                                <el-icon>
                                    <FolderOpened />
                                </el-icon>
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('runtime.version')" prop="version"></el-table-column>
                    <el-table-column
                        :label="$t('runtime.externalPort')"
                        prop="params.PANEL_APP_PORT_HTTP"
                    ></el-table-column>
                    <el-table-column :label="$t('commons.table.status')" prop="status">
                        <template #default="{ row }">
                            <el-popover
                                v-if="row.status === 'error'"
                                placement="bottom"
                                :width="400"
                                trigger="hover"
                                :content="row.message"
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
                    <el-table-column
                        prop="createdAt"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                        show-overflow-tooltip
                        min-width="120"
                        fix
                    />
                    <fu-table-operations
                        :ellipsis="10"
                        width="250px"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        fixed="right"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>
        <OperateNode ref="operateRef" @close="search" />
        <Delete ref="deleteRef" @close="search" />
    </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, reactive, ref } from 'vue';
import { Runtime } from '@/api/interface/runtime';
import { OperateRuntime, SearchRuntimes } from '@/api/modules/runtime';
import { dateFormat } from '@/utils/util';
import OperateNode from '@/views/website/runtime/node/operate/index.vue';
import Status from '@/components/status/index.vue';
import Delete from '@/views/website/runtime/delete/index.vue';
import i18n from '@/lang';
import RouterMenu from '../index.vue';
import router from '@/routers/router';

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
    type: 'node',
});
let timer: NodeJS.Timer | null = null;

const buttons = [
    {
        label: i18n.global.t('container.stop'),
        click: function (row: Runtime.Runtime) {
            operateRuntime('down', row.id);
        },
        disabled: function (row: Runtime.Runtime) {
            return disabledRuntime(row) || row.status === 'stopped';
        },
    },
    {
        label: i18n.global.t('container.start'),
        click: function (row: Runtime.Runtime) {
            operateRuntime('up', row.id);
        },
        disabled: function (row: Runtime.Runtime) {
            return disabledRuntime(row) || row.status === 'running';
        },
    },
    {
        label: i18n.global.t('container.restart'),
        click: function (row: Runtime.Runtime) {
            operateRuntime('restart', row.id);
        },
        disabled: function (row: Runtime.Runtime) {
            return disabledRuntime(row);
        },
    },
    {
        label: i18n.global.t('commons.button.edit'),
        click: function (row: Runtime.Runtime) {
            openDetail(row);
        },
        disabled: function (row: Runtime.Runtime) {
            return disabledRuntime(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: function (row: Runtime.Runtime) {
            openDelete(row);
        },
    },
];
const loading = ref(false);
const items = ref<Runtime.RuntimeDTO[]>([]);
const operateRef = ref();
const deleteRef = ref();

const disabledRuntime = (row: Runtime.Runtime) => {
    return row.status === 'starting' || row.status === 'recreating';
};

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
    operateRef.value.acceptParams({ type: 'node', mode: 'create' });
};

const openDetail = (row: Runtime.Runtime) => {
    operateRef.value.acceptParams({ type: row.type, mode: 'edit', id: row.id });
};

const openDelete = async (row: Runtime.Runtime) => {
    deleteRef.value.acceptParams(row.id, row.name);
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
    search();
    timer = setInterval(() => {
        search();
    }, 10000 * 3);
});

onUnmounted(() => {
    clearInterval(Number(timer));
    timer = null;
});
</script>

<style lang="scss" scoped></style>
