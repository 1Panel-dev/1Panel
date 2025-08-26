<template>
    <div v-loading="loading">
        <docker-status
            v-model:isActive="isActive"
            v-model:isExist="isExist"
            v-model:loading="loading"
            @search="search"
        />

        <LayoutContent v-if="isExist" :title="$t('container.repo', 2)" :class="{ mask: !isActive }">
            <template #leftToolBar>
                <el-button type="primary" @click="onOpenDialog('add')">
                    {{ $t('container.createRepo') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <TableSearch @search="search()" v-model:searchName="searchName" />
                <TableRefresh @search="search()" />
                <TableSetting title="repo-refresh" @search="search()" />
            </template>
            <template #main>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    :data="data"
                    @search="search"
                    :heightDiff="300"
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
                            <Status :status="row.status" :msg="row.message" />
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

        <OpDialog ref="opRef" @search="search" @submit="submitDelete" />
        <OperatorDialog @search="search" ref="dialogRef" />
        <ConfirmDialog ref="confirmDialog" @confirm="submitDelete" />
    </div>
</template>

<script lang="ts" setup>
import OperatorDialog from '@/views/container/repo/operator/index.vue';
import { reactive, ref } from 'vue';
import { dateFormat } from '@/utils/util';
import { Container } from '@/api/interface/container';
import { checkRepoStatus, deleteImageRepo, searchImageRepo } from '@/api/modules/container';
import DockerStatus from '@/views/container/docker-status/index.vue';
import i18n from '@/lang';

const loading = ref();
const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    cacheSizeKey: 'image-repo-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('image-repo-page-size')) || 10,
    total: 0,
});
const searchName = ref();

const opRef = ref();
const confirmDialog = ref();
const currentRepo = ref();

const isActive = ref();
const isExist = ref();

const search = async () => {
    if (!isActive.value || !isExist.value) {
        return;
    }
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
    currentRepo.value = row.id;
    if (row.protocol === 'http') {
        let params = {
            header: i18n.global.t('container.repo'),
            operationInfo: i18n.global.t('container.httpRepoHelper'),
            submitInputInfo: i18n.global.t('database.restartNow'),
        };
        confirmDialog.value!.acceptParams(params);
        return;
    }
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: [row.name],
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('container.repo'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: null,
        params: null,
    });
};

const submitDelete = async () => {
    loading.value = true;
    await deleteImageRepo(currentRepo.value)
        .then(() => {
            loading.value = false;
            search();
        })
        .catch(() => {
            loading.value = false;
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
</script>
