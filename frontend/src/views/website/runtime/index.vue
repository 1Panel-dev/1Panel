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
        <LayoutContent :title="$t('runtime.runtime')" v-loading="loading" :class="{ mask: !versionExist }">
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
                    <el-table-column :label="$t('runtime.status')" prop="status">
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
        <el-card width="30%" v-if="!versionExist" class="mask-prompt">
            <span>
                {{ $t('runtime.openrestryWarn') }}
                <span class="open-warn" @click="goRouter()">
                    <el-icon><Position /></el-icon>
                    {{ $t('runtime.toupgrade') }}
                </span>
            </span>
        </el-card>
        <CreateRuntime ref="createRef" @close="search" />
    </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, reactive, ref } from 'vue';
import { Runtime } from '@/api/interface/runtime';
import { DeleteRuntime, SearchRuntimes } from '@/api/modules/runtime';
import { dateFormat, toLowerCase } from '@/utils/util';
import RouterButton from '@/components/router-button/index.vue';
import ComplexTable from '@/components/complex-table/index.vue';
import LayoutContent from '@/layout/layout-content.vue';
import CreateRuntime from '@/views/website/runtime/create/index.vue';
import Status from '@/components/status/index.vue';
import i18n from '@/lang';
import { useDeleteData } from '@/hooks/use-delete-data';
import { CheckAppInstalled } from '@/api/modules/app';
import router from '@/routers';

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
const versionExist = ref(false);

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

const onCheck = async () => {
    try {
        const res = await CheckAppInstalled('openresty');
        if (res.data && res.data.version) {
            if (compareVersions(res.data.version, '1.21.4')) {
                versionExist.value = true;
            }
        }
    } catch (error) {}
};

function compareVersions(version1: string, version2: string): boolean {
    const v1 = version1.split('.');
    const v2 = version2.split('.');
    const len = Math.max(v1.length, v2.length);

    for (let i = 0; i < len; i++) {
        const num1 = parseInt(v1[i] || '0');
        const num2 = parseInt(v2[i] || '0');

        if (num1 !== num2) {
            return num1 > num2 ? true : false;
        }
    }

    return false;
}

const goRouter = async () => {
    router.push({ name: 'AppUpgrade' });
};

onMounted(() => {
    search();
    timer = setInterval(() => {
        search();
    }, 10000 * 3);
    onCheck();
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
