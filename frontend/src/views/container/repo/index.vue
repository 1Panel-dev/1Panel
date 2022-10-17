<template>
    <div>
        <Submenu activeName="repo" />
        <el-card style="margin-top: 20px">
            <ComplexTable :pagination-config="paginationConfig" v-model:selects="selects" :data="data" @search="search">
                <template #toolbar>
                    <el-button type="primary" @click="onOpenDialog('create')">
                        {{ $t('commons.button.create') }}
                    </el-button>
                    <el-button type="danger" plain :disabled="selects.length === 0" @click="onBatchDelete(null)">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </template>
                <el-table-column type="selection" fix></el-table-column>
                <el-table-column :label="$t('commons.table.name')" prop="name" min-width="60" />
                <el-table-column
                    :label="$t('container.downloadUrl')"
                    show-overflow-tooltip
                    prop="downloadUrl"
                    min-width="100"
                    fix
                />
                <el-table-column :label="$t('container.protocol')" prop="protocol" min-width="60" fix />
                <el-table-column :label="$t('commons.table.createdAt')" min-width="80" fix>
                    <template #default="{ row }">
                        {{ dateFromat(0, 0, row.createdAt) }}
                    </template>
                </el-table-column>
                <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" />
            </ComplexTable>
        </el-card>
        <OperatorDialog @search="search" ref="dialogRef" />
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import OperatorDialog from '@/views/container/repo/operator/index.vue';
import Submenu from '@/views/container/index.vue';
import { reactive, onMounted, ref } from 'vue';
import { dateFromat } from '@/utils/util';
import { Container } from '@/api/interface/container';
import { deleteImageRepo, searchImageRepo } from '@/api/modules/container';
import { useDeleteData } from '@/hooks/use-delete-data';
import i18n from '@/lang';

const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    page: 1,
    pageSize: 10,
    total: 0,
});

const search = async () => {
    let params = {
        page: paginationConfig.page,
        pageSize: paginationConfig.pageSize,
    };
    await searchImageRepo(params).then((res) => {
        if (res.data) {
            data.value = res.data.items;
        }
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

const onBatchDelete = async (row: Container.RepoInfo | null) => {
    let ids: Array<number> = [];
    if (row) {
        ids.push(row.id);
    } else {
        selects.value.forEach((item: Container.RepoInfo) => {
            ids.push(item.id);
        });
    }
    await useDeleteData(deleteImageRepo, { ids: ids }, 'commons.msg.delete', true);
    search();
};

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        disabled: (row: Container.RepoInfo) => {
            return row.downloadUrl === 'docker.io';
        },
        click: (row: Container.RepoInfo) => {
            onOpenDialog('edit', row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        disabled: (row: Container.RepoInfo) => {
            return row.downloadUrl === 'docker.io';
        },
        click: (row: Container.RepoInfo) => {
            onBatchDelete(row);
        },
    },
];

onMounted(() => {
    search();
});
</script>
