<template>
    <div v-loading="loading">
        <el-card v-if="dockerStatus != 'Running'" class="mask-prompt">
            <span>{{ $t('container.serviceUnavailable') }}</span>
            <el-button type="primary" link class="bt" @click="goSetting">【 {{ $t('container.setting') }} 】</el-button>
            <span>{{ $t('container.startIn') }}</span>
        </el-card>

        <LayoutContent :title="$t('container.repo')" :class="{ mask: dockerStatus != 'Running' }">
            <template #toolbar>
                <div class="flex justify-between gap-2 flex-wrap sm:flex-row">
                    <div class="flex flex-wrap gap-3">
                        <el-button type="primary" @click="onOpenDialog('add')">
                            {{ $t('container.createRepo') }}
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
                    <el-table-column :label="$t('commons.table.name')" prop="name" min-width="60" />
                    <el-table-column
                        :label="$t('container.downloadUrl')"
                        show-overflow-tooltip
                        prop="downloadUrl"
                        min-width="100"
                        fix
                    />
                    <el-table-column :label="$t('commons.table.protocol')" prop="protocol" min-width="60" fix />
                    <el-table-column :label="$t('commons.table.status')" prop="status" min-width="60" fix>
                        <template #default="{ row }">
                            <el-tag v-if="row.status === 'Success'" type="success">
                                {{ $t('commons.status.success') }}
                            </el-tag>
                            <el-tooltip v-else effect="dark" placement="bottom">
                                <template #content>
                                    <div style="width: 300px; word-break: break-all">{{ row.message }}</div>
                                </template>
                                <el-tag type="danger">{{ $t('commons.status.failed') }}</el-tag>
                            </el-tooltip>
                        </template>
                    </el-table-column>
                    <el-table-column
                        prop="createdAt"
                        :label="$t('commons.table.createdAt')"
                        min-width="80"
                        fix
                        :formatter="dateFormat"
                    />
                    <fu-table-operations width="200px" :buttons="buttons" :label="$t('commons.table.operate')" />
                </ComplexTable>
            </template>
        </LayoutContent>

        <OpDialog ref="opRef" @search="search" />
        <OperatorDialog @search="search" ref="dialogRef" />
    </div>
</template>

<script lang="ts" setup>
import OperatorDialog from '@/views/container/repo/operator/index.vue';
import { reactive, onMounted, ref } from 'vue';
import { dateFormat } from '@/utils/util';
import { Container } from '@/api/interface/container';
import { checkRepoStatus, deleteImageRepo, loadDockerStatus, searchImageRepo } from '@/api/modules/container';
import i18n from '@/lang';
import router from '@/routers';

const loading = ref();
const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    cacheSizeKey: 'image-repo-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const searchName = ref();

const opRef = ref();

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

const search = async () => {
    let params = {
        info: searchName.value,
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    loading.value = true;
    await searchImageRepo(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};
const dialogRef = ref();
const onOpenDialog = async (
    title: string,
    rowData: Partial<Container.RepoInfo> = {
        auth: true,
        protocol: 'http',
    },
) => {
    let params = {
        title,
        rowData: { ...rowData },
    };
    dialogRef.value!.acceptParams(params);
};

const onDelete = async (row: Container.RepoInfo) => {
    let ids = [row.id];
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: [row.name],
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('container.repo'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: deleteImageRepo,
        params: { ids: ids },
    });
};

const onCheckConn = async (row: Container.RepoInfo) => {
    loading.value = true;
    await checkRepoStatus(row.id)
        .then(() => {
            loading.value = false;
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};

const buttons = [
    {
        label: i18n.global.t('commons.button.sync'),
        disabled: (row: Container.RepoInfo) => {
            return row.id === 1;
        },
        click: (row: Container.RepoInfo) => {
            onCheckConn(row);
        },
    },
    {
        label: i18n.global.t('commons.button.edit'),
        disabled: (row: Container.RepoInfo) => {
            return row.id === 1;
        },
        click: (row: Container.RepoInfo) => {
            onOpenDialog('edit', row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        disabled: (row: Container.RepoInfo) => {
            return row.id === 1;
        },
        click: (row: Container.RepoInfo) => {
            onDelete(row);
        },
    },
];

onMounted(() => {
    loadStatus();
});
</script>
