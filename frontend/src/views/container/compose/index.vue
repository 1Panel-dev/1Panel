<template>
    <div v-loading="loading">
        <Submenu activeName="compose" />
        <el-card style="margin-top: 20px">
            <ComplexTable :pagination-config="paginationConfig" v-model:selects="selects" :data="data" @search="search">
                <template #toolbar>
                    <el-button icon="Plus" type="primary" @click="onOpenDialog()">
                        {{ $t('commons.button.create') }}
                    </el-button>
                </template>
                <el-table-column
                    :label="$t('commons.table.name')"
                    show-overflow-tooltip
                    min-width="100"
                    prop="name"
                    fix
                >
                    <template #default="{ row }">
                        <el-link @click="goContainer(row.name)" type="primary">{{ row.name }}</el-link>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('container.from')" prop="createdBy" min-width="80" fix />
                <el-table-column :label="$t('container.containerNumber')" prop="containerNumber" min-width="80" fix />
                <el-table-column :label="$t('commons.table.createdAt')" prop="createdAt" min-width="80" fix />
                <fu-table-operations
                    width="200px"
                    :ellipsis="10"
                    :buttons="buttons"
                    :label="$t('commons.table.operate')"
                    fix
                />
            </ComplexTable>
        </el-card>

        <EditDialog ref="dialogEditRef" />
        <CreateDialog @search="search" ref="dialogRef" />
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import { reactive, onMounted, ref } from 'vue';
import CreateDialog from '@/views/container/compose/create/index.vue';
import EditDialog from '@/views/container/compose/edit/index.vue';
import Submenu from '@/views/container/index.vue';
import { composeOperator, searchCompose } from '@/api/modules/container';
import i18n from '@/lang';
import { ElMessage } from 'element-plus';
import router from '@/routers';
import { Container } from '@/api/interface/container';
import { useDeleteData } from '@/hooks/use-delete-data';
import { LoadFile } from '@/api/modules/files';

const data = ref();
const selects = ref<any>([]);
const loading = ref(false);

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

const goContainer = async (name: string) => {
    router.push({ name: 'ComposeDetail', params: { filters: 'com.docker.compose.project=' + name } });
};

const dialogRef = ref();
const onOpenDialog = async () => {
    dialogRef.value!.acceptParams();
};

const onDelete = async (row: Container.ComposeInfo) => {
    const param = {
        name: row.name,
        path: row.path,
        operation: 'down',
    };
    await useDeleteData(composeOperator, param, 'commons.msg.delete');
    search();
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
};

const dialogEditRef = ref();
const onEdit = async (row: Container.ComposeInfo) => {
    const res = await LoadFile({ path: row.path });
    let params = {
        path: row.path,
        content: res.data,
    };
    dialogEditRef.value!.acceptParams(params);
};

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: Container.ComposeInfo) => {
            onEdit(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Container.ComposeInfo) => {
            onDelete(row);
        },
    },
];
onMounted(() => {
    search();
});
</script>
