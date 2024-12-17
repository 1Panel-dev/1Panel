<template>
    <div v-loading="loading">
        <div v-show="isOnDetail">
            <ComposeDetail ref="composeDetailRef" />
        </div>
        <el-card v-if="dockerStatus != 'Running'" class="mask-prompt">
            <span>{{ $t('container.serviceUnavailable') }}</span>
            <el-button type="primary" link class="bt" @click="goSetting">【 {{ $t('container.setting') }} 】</el-button>
            <span>{{ $t('container.startIn') }}</span>
        </el-card>

        <LayoutContent
            v-if="!isOnDetail"
            :title="$t('container.compose', 2)"
            :class="{ mask: dockerStatus != 'Running' }"
        >
            <template #toolbar>
                <div class="flex justify-between gap-2 flex-wrap sm:flex-row">
                    <div class="flex flex-wrap gap-3">
                        <el-button type="primary" @click="onOpenDialog()">
                            {{ $t('container.createCompose') }}
                        </el-button>
                    </div>
                    <div class="flex flex-wrap gap-3">
                        <TableSetting @search="search()" />
                        <TableSearch @search="search()" v-model:searchName="searchName" />
                    </div>
                </div>
            </template>
            <template #main>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    :data="data"
                    @search="search"
                >
                    <el-table-column
                        :label="$t('commons.table.name')"
                        width="170"
                        prop="name"
                        sortable
                        fix
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <el-text type="primary" class="cursor-pointer" @click="loadDetail(row)">
                                {{ row.name }}
                            </el-text>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('container.from')" prop="createdBy" min-width="120" fix>
                        <template #default="{ row }">
                            <span v-if="row.createdBy === ''">{{ $t('container.local') }}</span>
                            <span v-if="row.createdBy === 'Apps'">{{ $t('container.apps') }}</span>
                            <span v-if="row.createdBy === '1Panel'">1Panel</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('container.composeDirectory')" min-width="100" fix>
                        <template #default="{ row }">
                            <el-button type="primary" link @click="toComposeFolder(row)">
                                <el-icon>
                                    <FolderOpened />
                                </el-icon>
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('container.containerStatus')" min-width="200" fix>
                        <template #default="scope">
                            <div>
                                {{ getContainerStatus(scope.row.containers) }}
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.createdAt')" prop="createdAt" min-width="180" fix />
                    <fu-table-operations
                        width="200px"
                        :ellipsis="10"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        :fixed="mobile ? false : 'right'"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <EditDialog @search="search" ref="dialogEditRef" />
        <CreateDialog @search="search" ref="dialogRef" />
        <DeleteDialog @search="search" ref="dialogDelRef" />
    </div>
</template>

<script lang="ts" setup>
import { reactive, onMounted, ref, computed } from 'vue';
import EditDialog from '@/views/container/compose/edit/index.vue';
import CreateDialog from '@/views/container/compose/create/index.vue';
import DeleteDialog from '@/views/container/compose/delete/index.vue';
import ComposeDetail from '@/views/container/compose/detail/index.vue';
import { loadContainerLog, loadDockerStatus, searchCompose } from '@/api/modules/container';
import i18n from '@/lang';
import { Container } from '@/api/interface/container';
import router from '@/routers';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const mobile = computed(() => {
    return globalStore.isMobile();
});

const data = ref();
const selects = ref<any>([]);
const loading = ref(false);

const isOnDetail = ref(false);

const paginationConfig = reactive({
    cacheSizeKey: 'container-compose-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const searchName = ref();

const dockerStatus = ref('Running');
const loadStatus = async () => {
    loading.value = true;
    await loadDockerStatus()
        .then((res) => {
            loading.value = false;
            dockerStatus.value = res.data;
            if (dockerStatus.value === 'Running') {
                search();
            }
        })
        .catch(() => {
            dockerStatus.value = 'Failed';
            loading.value = false;
        });
};
const goSetting = async () => {
    router.push({ name: 'ContainerSetting' });
};

const toComposeFolder = async (row: Container.ComposeInfo) => {
    router.push({ path: '/hosts/files', query: { path: row.workdir } });
};

const search = async () => {
    let params = {
        info: searchName.value,
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    loading.value = true;
    await searchCompose(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .finally(() => {
            loading.value = false;
        });
};

const composeDetailRef = ref();
const loadDetail = async (row: Container.ComposeInfo) => {
    let params = {
        createdBy: row.createdBy,
        name: row.name,
        path: row.path,
        filters: 'com.docker.compose.project=' + row.name,
    };
    isOnDetail.value = true;
    composeDetailRef.value!.acceptParams(params);
};
const getContainerStatus = (containers) => {
    const safeContainers = containers || [];
    const runningCount = safeContainers.filter((container) => container.state.toLowerCase() === 'running').length;
    const totalCount = safeContainers.length;
    const statusText = runningCount > 0 ? 'Running' : 'Exited';
    if (statusText === 'Exited') {
        return i18n.global.t('container.exited');
    } else {
        return i18n.global.t('container.running') + ` (${runningCount}/${totalCount})`;
    }
};

const dialogRef = ref();
const onOpenDialog = async () => {
    dialogRef.value!.acceptParams();
};

const dialogDelRef = ref();
const onDelete = async (row: Container.ComposeInfo) => {
    const param = {
        name: row.name,
        path: row.path,
    };
    dialogDelRef.value.acceptParams(param);
};

const dialogEditRef = ref();
const onEdit = async (row: Container.ComposeInfo) => {
    const res = await loadContainerLog('compose-detail', row.name);
    let params = {
        name: row.name,
        path: row.path,
        content: res.data,
        env: row.env,
        createdBy: row.createdBy,
    };
    dialogEditRef.value!.acceptParams(params);
};

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: Container.ComposeInfo) => {
            onEdit(row);
        },
        disabled: (row: any) => {
            return row.createdBy === 'Local';
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Container.ComposeInfo) => {
            onDelete(row);
        },
        disabled: (row: any) => {
            return row.createdBy !== '1Panel';
        },
    },
];
onMounted(() => {
    loadStatus();
});
</script>
