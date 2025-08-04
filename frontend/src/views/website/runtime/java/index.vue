<template>
    <div>
        <RouterMenu />
        <DockerStatus v-model:isActive="isActive" v-model:isExist="isExist" />
        <LayoutContent v-loading="loading" v-if="isExist" :class="{ mask: !isActive }">
            <template #leftToolBar>
                <el-button type="primary" @click="openCreate">
                    {{ $t('runtime.create') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <TableRefresh @search="search()" />
                <TableSetting title="java-runtime-refresh" @search="search()" />
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
                            <el-text type="primary" class="cursor-pointer" @click="openDetail(row)">
                                {{ row.name }}
                            </el-text>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('runtime.codeDir')" prop="codeDir" min-width="120px">
                        <template #default="{ row }">
                            <el-button type="primary" link @click="routerToFileWithPath(row.codeDir)">
                                <el-icon>
                                    <FolderOpened />
                                </el-icon>
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('app.version')" prop="version"></el-table-column>
                    <el-table-column :label="$t('runtime.externalPort')" prop="port" min-width="110px">
                        <template #default="{ row }">
                            <PortJump :row="row" :jump="goDashboard" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.status')" prop="status">
                        <template #default="{ row }">
                            <RuntimeStatus :row="row" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.button.log')" prop="path" min-width="90px">
                        <template #default="{ row }">
                            <el-button @click="openLog(row)" link type="primary" v-if="row.status != 'Stopped'">
                                {{ $t('website.check') }}
                            </el-button>
                            <span v-else>-</span>
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
        <OperateJava ref="operateRef" @close="search" />
        <Delete ref="deleteRef" @close="search" />
        <ComposeLogs ref="composeLogRef" />
        <PortJumpDialog ref="dialogPortJumpRef" />
        <AppResources ref="checkRef" @close="search" />
        <Terminal ref="terminalRef" />
    </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { Runtime } from '@/api/interface/runtime';
import { RuntimeDeleteCheck, SearchRuntimes, SyncRuntime } from '@/api/modules/runtime';
import { dateFormat } from '@/utils/util';
import OperateJava from '@/views/website/runtime/java/operate/index.vue';
import Delete from '@/views/website/runtime/delete/index.vue';
import i18n from '@/lang';
import RouterMenu from '../index.vue';
import ComposeLogs from '@/components/log/compose/index.vue';
import PortJumpDialog from '@/components/port-jump/index.vue';
import AppResources from '@/views/website/runtime/php/check/index.vue';
import RuntimeStatus from '@/views/website/runtime/components/runtime-status.vue';
import PortJump from '@/views/website/runtime/components/port-jump.vue';
import Terminal from '@/views/website/runtime/components/terminal.vue';
import DockerStatus from '@/views/container/docker-status/index.vue';
import { disabledButton } from '@/utils/runtime';
import { GlobalStore } from '@/store';
import { operateRuntime, updateRuntimeRemark } from '../common/utils';
import { routerToFileWithPath } from '@/utils/router';
const globalStore = GlobalStore();
const mobile = computed(() => {
    return globalStore.isMobile();
});

const loading = ref(false);
const items = ref<Runtime.RuntimeDTO[]>([]);
const operateRef = ref();
const deleteRef = ref();
const dialogPortJumpRef = ref();
const composeLogRef = ref();
const checkRef = ref();
const terminalRef = ref();
const isActive = ref(false);
const isExist = ref(false);

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
    type: 'java',
});
const buttons = [
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
        label: i18n.global.t('commons.button.edit'),
        click: function (row: Runtime.Runtime) {
            openDetail(row);
        },
        disabled: function (row: Runtime.Runtime) {
            return disabledButton(row, 'edit');
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
    operateRef.value.acceptParams({ type: 'java', mode: 'create' });
};

const openDetail = (row: Runtime.Runtime) => {
    operateRef.value.acceptParams({ type: row.type, mode: 'edit', id: row.id });
};

const openDelete = (row: Runtime.Runtime) => {
    RuntimeDeleteCheck(row.id).then(async (res) => {
        const items = res.data;
        if (res.data && res.data.length > 0) {
            checkRef.value.acceptParams({ items: items, key: 'website', installID: row.id });
        } else {
            deleteRef.value.acceptParams(row.id, row.name);
        }
    });
};

const openLog = (row: any) => {
    composeLogRef.value.acceptParams({
        compose: row.path + '/docker-compose.yml',
        resource: row.name,
        container: row.container,
    });
};

const openTerminal = (row: Runtime.Runtime) => {
    const container = row.params['CONTAINER_NAME'];
    terminalRef.value.acceptParams({ containerID: container, container: container });
};

const goDashboard = async (port: any, protocol: string) => {
    dialogPortJumpRef.value.acceptParams({ port: port, protocol: protocol });
};

onMounted(() => {
    sync();
    search();
});
</script>

<style lang="scss" scoped></style>
