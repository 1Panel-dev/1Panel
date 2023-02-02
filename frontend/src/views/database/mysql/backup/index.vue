<template>
    <div>
        <el-drawer v-model="backupVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="50%">
            <template #header>
                <DrawerHeader :header="$t('database.backup')" :resource="dbName" :back="handleClose" />
            </template>
            <ComplexTable :pagination-config="paginationConfig" v-model:selects="selects" @search="search" :data="data">
                <template #toolbar>
                    <el-button type="primary" @click="onBackup()">
                        {{ $t('database.backup') }}
                    </el-button>
                    <el-button type="primary" plain :disabled="selects.length === 0" @click="onBatchDelete(null)">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </template>
                <el-table-column type="selection" fix />
                <el-table-column :label="$t('commons.table.name')" prop="fileName" show-overflow-tooltip />
                <el-table-column :label="$t('database.source')" prop="backupType" />
                <el-table-column
                    prop="createdAt"
                    :label="$t('commons.table.date')"
                    :formatter="dateFormat"
                    show-overflow-tooltip
                />

                <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fix />
            </ComplexTable>
        </el-drawer>
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import { reactive, ref } from 'vue';
import { dateFormat } from '@/utils/util';
import { useDeleteData } from '@/hooks/use-delete-data';
import { backup, recover } from '@/api/modules/database';
import i18n from '@/lang';
import { ElMessage } from 'element-plus';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { deleteBackupRecord, downloadBackupRecord, searchBackupRecords } from '@/api/modules/backup';
import { Backup } from '@/api/interface/backup';

const selects = ref<any>([]);

const data = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const backupVisiable = ref(false);
const mysqlName = ref();
const dbName = ref();
interface DialogProps {
    mysqlName: string;
    dbName: string;
}
const acceptParams = (params: DialogProps): void => {
    mysqlName.value = params.mysqlName;
    dbName.value = params.dbName;
    backupVisiable.value = true;
    search();
};
const handleClose = () => {
    backupVisiable.value = false;
};

const search = async () => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        type: 'database-mysql',
        name: mysqlName.value,
        detailName: dbName.value,
    };
    const res = await searchBackupRecords(params);
    data.value = res.data.items || [];
    paginationConfig.total = res.data.total;
};

const onBackup = async () => {
    let params = {
        mysqlName: mysqlName.value,
        dbName: dbName.value,
    };
    await backup(params);
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
    search();
};

const onRecover = async (row: Backup.RecordInfo) => {
    let params = {
        mysqlName: mysqlName.value,
        dbName: dbName.value,
        backupName: row.fileDir + '/' + row.fileName,
    };
    await recover(params);
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
};

const onDownload = async (row: Backup.RecordInfo) => {
    let params = {
        source: row.source,
        fileDir: row.fileDir,
        fileName: row.fileName,
    };
    const res = await downloadBackupRecord(params);
    const downloadUrl = window.URL.createObjectURL(new Blob([res]));
    const a = document.createElement('a');
    a.style.display = 'none';
    a.href = downloadUrl;
    a.download = row.fileName;
    const event = new MouseEvent('click');
    a.dispatchEvent(event);
};

const onBatchDelete = async (row: Backup.RecordInfo | null) => {
    let ids: Array<number> = [];
    if (row) {
        ids.push(row.id);
    } else {
        selects.value.forEach((item: Backup.RecordInfo) => {
            ids.push(item.id);
        });
    }
    await useDeleteData(deleteBackupRecord, { ids: ids }, 'commons.msg.delete');
    search();
};

const buttons = [
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Backup.RecordInfo) => {
            onBatchDelete(row);
        },
    },
    {
        label: i18n.global.t('commons.button.recover'),
        click: (row: Backup.RecordInfo) => {
            onRecover(row);
        },
    },
    {
        label: i18n.global.t('commons.button.download'),
        click: (row: Backup.RecordInfo) => {
            onDownload(row);
        },
    },
];

defineExpose({
    acceptParams,
});
</script>
