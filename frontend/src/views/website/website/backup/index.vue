<template>
    <el-drawer v-model="backupVisiable" size="50%" :show-close="false">
        <template #header>
            <Header :header="$t('database.backup') + ' - ' + websiteName" :back="handleClose"></Header>
        </template>
        <ComplexTable
            v-loading="loading"
            :pagination-config="paginationConfig"
            v-model:selects="selects"
            @search="search"
            :data="data"
        >
            <template #toolbar>
                <el-button type="primary" @click="onBackup()">
                    {{ $t('database.backup') }}
                </el-button>
                <el-button type="danger" plain :disabled="selects.length === 0" @click="onBatchDelete(null)">
                    {{ $t('commons.button.delete') }}
                </el-button>
            </template>
            <el-table-column type="selection" fix />
            <el-table-column :label="$t('commons.table.name')" prop="fileName" show-overflow-tooltip />
            <el-table-column :label="$t('database.source')" prop="backupType" />
            <el-table-column
                prop="createdAt"
                :label="$t('commons.table.date')"
                :formatter="dateFromat"
                show-overflow-tooltip
            />
            <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fix />
        </ComplexTable>
    </el-drawer>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import { reactive, ref } from 'vue';
import { dateFromat } from '@/utils/util';
import { useDeleteData } from '@/hooks/use-delete-data';
import i18n from '@/lang';
import { ElMessage } from 'element-plus';
import { deleteBackupRecord, downloadBackupRecord, searchBackupRecords } from '@/api/modules/backup';
import { Backup } from '@/api/interface/backup';
import { BackupWebsite, RecoverWebsite } from '@/api/modules/website';
import Header from '@/components/drawer-header/index.vue';

const selects = ref<any>([]);
const loading = ref(false);

const data = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const backupVisiable = ref(false);
const websiteName = ref();
const websiteID = ref();
const websiteType = ref();

interface DialogProps {
    id: string;
    type: string;
    name: string;
}
const acceptParams = (params: DialogProps): void => {
    websiteName.value = params.name;
    websiteID.value = params.id;
    websiteType.value = params.type;
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
        type: 'website-' + websiteType.value,
        name: websiteName.value,
        detailName: '',
    };
    const res = await searchBackupRecords(params);
    data.value = res.data.items || [];
    paginationConfig.total = res.data.total;
};

const onRecover = async (row: Backup.RecordInfo) => {
    let params = {
        websiteName: websiteName.value,
        type: websiteType.value,
        backupName: row.fileDir + '/' + row.fileName,
    };
    loading.value = true;
    await RecoverWebsite(params)
        .then(() => {
            loading.value = false;
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        })
        .finally(() => {
            loading.value = false;
        });
};

const onBackup = async () => {
    loading.value = true;
    await BackupWebsite({ id: websiteID.value })
        .then(() => {
            loading.value = false;
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .finally(() => {
            loading.value = false;
        });
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
