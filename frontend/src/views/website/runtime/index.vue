<template>
    <div>
        <RouterButton
            :buttons="[
                {
                    label: 'PHP',
                    path: '/runtimes/php',
                },
            ]"
        />
        <LayoutContent :title="$t('runtime.runtime')" v-loading="loading">
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
                        width="120px"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        fixed="right"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>
        <CreateRuntime ref="createRef" @close="search" />
    </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, reactive, ref } from 'vue';
import { Runtime } from '@/api/interface/runtime';
import { DeleteRuntime, SearchRuntimes } from '@/api/modules/runtime';
import { dateFormat, toLowerCase } from '@/utils/util';
import CreateRuntime from '@/views/website/runtime/create/index.vue';
import Status from '@/components/status/index.vue';
import i18n from '@/lang';
import { useDeleteData } from '@/hooks/use-delete-data';

const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
let req = reactive<Runtime.RuntimeReq>({
    name: '',
    page: 1,
    pageSize: 40,
});
let timer: NodeJS.Timer | null = null;

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: function (row: Runtime.Runtime) {
            openDetail(row);
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
    createRef.value.acceptParams({ type: row.type, mode: 'edit', id: row.id });
};

const openDelete = async (row: Runtime.Runtime) => {
    await useDeleteData(DeleteRuntime, { id: row.id }, 'commons.msg.delete');
    search();
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
