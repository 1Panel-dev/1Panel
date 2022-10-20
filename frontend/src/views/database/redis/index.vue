<template>
    <div>
        <Submenu activeName="redis" />
        <ComplexTable
            :pagination-config="paginationConfig"
            v-model:selects="selects"
            @search="search"
            style="margin-top: 20px"
            :data="data"
        >
            <template #toolbar>
                <el-button type="primary" @click="onOpenDialog()">{{ $t('commons.button.create') }}</el-button>
                <el-button type="danger" plain :disabled="selects.length === 0" @click="onBatchDelete(null)">
                    {{ $t('commons.button.delete') }}
                </el-button>
            </template>
            <el-table-column type="selection" fix />
            <el-table-column :label="$t('commons.table.name')" prop="name" />
            <el-table-column :label="$t('auth.username')" prop="username" />
            <el-table-column :label="$t('auth.password')" prop="password" />
            <el-table-column :label="$t('commons.table.description')" prop="description" />
            <el-table-column
                prop="createdAt"
                :label="$t('commons.table.date')"
                :formatter="dateFromat"
                show-overflow-tooltip
            />
            <fu-table-operations type="icon" :buttons="buttons" :label="$t('commons.table.operate')" fix />
        </ComplexTable>

        <OperatrDialog @search="search" ref="dialogRef" />
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import OperatrDialog from '@/views/database/create/index.vue';
import Submenu from '@/views/database/index.vue';
import { dateFromat } from '@/utils/util';
import { onMounted, reactive, ref } from 'vue';
import { deleteMysqlDB, searchMysqlDBs } from '@/api/modules/database';
import i18n from '@/lang';
import { Cronjob } from '@/api/interface/cronjob';
import { useDeleteData } from '@/hooks/use-delete-data';

const selects = ref<any>([]);

const data = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const dialogRef = ref();
const onOpenDialog = async () => {
    dialogRef.value!.acceptParams();
};

const search = async () => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    const res = await searchMysqlDBs(params);
    data.value = res.data.items || [];
    paginationConfig.total = res.data.total;
};

const onBatchDelete = async (row: Cronjob.CronjobInfo | null) => {
    let ids: Array<number> = [];
    if (row) {
        ids.push(row.id);
    } else {
        selects.value.forEach((item: Cronjob.CronjobInfo) => {
            ids.push(item.id);
        });
    }
    await useDeleteData(deleteMysqlDB, { ids: ids }, 'commons.msg.delete', true);
    search();
};

const buttons = [
    {
        label: i18n.global.t('commons.button.delete'),
        icon: 'Delete',
        click: (row: Cronjob.CronjobInfo) => {
            onBatchDelete(row);
        },
    },
];

onMounted(() => {
    search();
});
</script>
