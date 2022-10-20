<template>
    <div>
        <Submenu activeName="mysql" />
        <el-dropdown size="large" split-button style="margin-top: 20px; margin-bottom: 5px">
            Mysql 版本 {{ version }}
            <template #dropdown>
                <el-dropdown-menu v-model="version">
                    <el-dropdown-item @click="version = '5.7.39'">5.7.39</el-dropdown-item>
                    <el-dropdown-item @click="version = '8.0.30'">8.0.30</el-dropdown-item>
                </el-dropdown-menu>
            </template>
        </el-dropdown>
        <el-button style="margin-top: 20px; margin-left: 10px" size="large" icon="Setting" @click="onOpenDialog()">
            {{ $t('database.setting') }}
        </el-button>
        <el-card>
            <ComplexTable :pagination-config="paginationConfig" v-model:selects="selects" @search="search" :data="data">
                <template #toolbar>
                    <el-button type="primary" @click="onOpenDialog()">{{ $t('commons.button.create') }}</el-button>
                    <el-button @click="onOpenDialog()">{{ $t('database.rootPassword') }}</el-button>
                    <el-button @click="onOpenDialog()">phpMyAdmin</el-button>
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
                <fu-table-operations
                    width="300px"
                    :buttons="buttons"
                    :ellipsis="10"
                    :label="$t('commons.table.operate')"
                    fix
                />
            </ComplexTable>
        </el-card>

        <OperatrDialog @search="search" ref="dialogRef" />
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import OperatrDialog from '@/views/database/mysql/create/index.vue';
import Submenu from '@/views/database/index.vue';
import { dateFromat } from '@/utils/util';
import { onMounted, reactive, ref } from 'vue';
import { deleteMysqlDB, searchMysqlDBs } from '@/api/modules/database';
import i18n from '@/lang';
import { Cronjob } from '@/api/interface/cronjob';
import { useDeleteData } from '@/hooks/use-delete-data';

const selects = ref<any>([]);
const version = ref<string>('5.7.39');

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
        label: i18n.global.t('commons.button.edit'),
        icon: 'Delete',
        click: (row: Cronjob.CronjobInfo) => {
            onBatchDelete(row);
        },
    },
    {
        label: i18n.global.t('database.backupList') + '(1)',
        icon: 'Delete',
        click: (row: Cronjob.CronjobInfo) => {
            onBatchDelete(row);
        },
    },
    {
        label: i18n.global.t('database.loadBackup'),
        icon: 'Delete',
        click: (row: Cronjob.CronjobInfo) => {
            onBatchDelete(row);
        },
    },
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
