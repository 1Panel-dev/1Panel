<template>
    <div>
        <RouterMenu />
        <LayoutContent :title="'PHP'" v-loading="loading">
            <template #prompt>
                <el-alert type="info" :closable="false">
                    <template #title>
                        <span>{{ $t('runtime.systemRestartHelper') }}</span>
                    </template>
                </el-alert>
            </template>
            <template #toolbar>
                <el-button type="primary" @click="openCreate">
                    {{ $t('runtime.create') }}
                </el-button>

                <el-button @click="openExtensions">
                    {{ $t('php.extensions') }}
                </el-button>
            </template>
            <template #main>
                <ComplexTable :pagination-config="paginationConfig" :data="items" @search="search()">
                    <el-table-column :label="$t('commons.table.name')" fix prop="name" min-width="120px">
                        <template #default="{ row }">
                            <Tooltip :text="row.name" @click="openDetail(row)" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('runtime.resource')" prop="resource">
                        <template #default="{ row }">
                            <span>{{ $t('runtime.' + toLowerCase(row.resource)) }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('runtime.version')" prop="version"></el-table-column>
                    <el-table-column :label="$t('runtime.image')" prop="image" show-overflow-tooltip></el-table-column>
                    <el-table-column :label="$t('commons.table.status')" prop="status">
                        <template #default="{ row }">
                            <el-popover
                                v-if="row.status === 'error' || row.status === 'systemRestart'"
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
                    <el-table-column :label="$t('website.log')" prop="">
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
                        :ellipsis="10"
                        width="120px"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        fixed="right"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <CreateRuntime ref="createRef" @close="search" @submit="openCreateLog" />
        <OpDialog ref="opRef" @search="search" />
        <Log ref="logRef" @close="search" />
        <Extensions ref="extensionsRef" @close="search" />
    </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, reactive, ref } from 'vue';
import { Runtime } from '@/api/interface/runtime';
import { DeleteRuntime, SearchRuntimes } from '@/api/modules/runtime';
import { dateFormat, toLowerCase } from '@/utils/util';
import CreateRuntime from '@/views/website/runtime/php/create/index.vue';
import Status from '@/components/status/index.vue';
import i18n from '@/lang';
import RouterMenu from '../index.vue';
import Log from '@/components/log-dialog/index.vue';
import Extensions from './extensions/index.vue';

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
let timer: NodeJS.Timer | null = null;
const opRef = ref();
const logRef = ref();
const extensionsRef = ref();

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: function (row: Runtime.Runtime) {
            openDetail(row);
        },
        disabled: function (row: Runtime.Runtime) {
            return row.status === 'building';
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
const createRef = ref();

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

const openLog = (row: Runtime.RuntimeDTO) => {
    logRef.value.acceptParams({ id: row.id, type: 'php', tail: row.status == 'building' });
};

const openCreateLog = (id: number) => {
    logRef.value.acceptParams({ id: id, type: 'php', tail: true });
};

const openExtensions = () => {
    extensionsRef.value.acceptParams();
};

const openDelete = async (row: Runtime.Runtime) => {
    opRef.value.acceptParams({
        title: i18n.global.t('commons.msg.deleteTitle'),
        names: [row.name],
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('website.runtime'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: DeleteRuntime,
        params: { id: row.id, forceDelete: true },
    });
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

<style lang="scss" scoped>
.open-warn {
    color: $primary-color;
    cursor: pointer;
}
</style>
