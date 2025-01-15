<template>
    <DrawerPro
        v-model="backupVisible"
        :header="$t('commons.button.backup')"
        :resource="cronjob"
        :back="handleClose"
        size="large"
    >
        <template #content>
            <ComplexTable
                v-loading="loading"
                :pagination-config="paginationConfig"
                v-model:selects="selects"
                @search="search"
                :data="data"
            >
                <el-table-column :label="$t('commons.table.name')" prop="fileName" show-overflow-tooltip />
                <el-table-column :label="$t('file.size')" prop="size" show-overflow-tooltip>
                    <template #default="{ row }">
                        <div v-if="row.hasLoad">
                            <span v-if="row.size">
                                {{ computeSize(row.size) }}
                            </span>
                            <span v-else>-</span>
                        </div>
                        <div v-if="!row.hasLoad">
                            <el-button link loading></el-button>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('database.source')" prop="accountType" show-overflow-tooltip>
                    <template #default="{ row }">
                        <span v-if="row.accountType === 'LOCAL'">
                            {{ $t('setting.LOCAL') }}
                        </span>
                        <span v-if="row.accountType && row.accountType !== 'LOCAL'">
                            {{ $t('setting.' + row.accountType) + ' - ' + row.accountName }}
                        </span>
                        <span v-if="!row.accountType">-</span>
                    </template>
                </el-table-column>
                <el-table-column
                    prop="createdAt"
                    :label="$t('commons.table.date')"
                    :formatter="dateFormat"
                    show-overflow-tooltip
                />

                <fu-table-operations width="130px" :buttons="buttons" :label="$t('commons.table.operate')" fix />
            </ComplexTable>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { computeSize, dateFormat, downloadFile } from '@/utils/util';
import i18n from '@/lang';
import { downloadBackupRecord, loadRecordSize, searchBackupRecordsByCronjob } from '@/api/modules/backup';
import { Backup } from '@/api/interface/backup';
import { MsgError } from '@/utils/message';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const selects = ref<any>([]);
const loading = ref();

const data = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const backupVisible = ref(false);
const cronjob = ref();
const cronjobID = ref();

interface DialogProps {
    cronjob: string;
    cronjobID: number;
}
const acceptParams = (params: DialogProps): void => {
    cronjob.value = params.cronjob;
    cronjobID.value = params.cronjobID;
    backupVisible.value = true;
    search();
};
const handleClose = () => {
    backupVisible.value = false;
};

const search = async () => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        cronjobID: cronjobID.value,
    };
    loading.value = true;
    await searchBackupRecordsByCronjob(params)
        .then((res) => {
            loading.value = false;
            loadSize(params);
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const loadSize = async (params: any) => {
    params.type = 'cronjob';
    await loadRecordSize(params)
        .then((res) => {
            let stats = res.data || [];
            if (stats.length === 0) {
                return;
            }
            for (const backup of data.value) {
                for (const item of stats) {
                    if (backup.id === item.id) {
                        backup.hasLoad = true;
                        backup.size = item.size;
                        break;
                    }
                }
            }
        })
        .catch(() => {
            loading.value = false;
        });
};

const onDownload = async (row: Backup.RecordInfo) => {
    if (row.accountType === 'ALIYUN' && row.size < 100 * 1024 * 1024) {
        MsgError(i18n.global.t('setting.ALIYUNHelper'));
        return;
    }
    let params = {
        downloadAccountID: row.downloadAccountID,
        fileDir: row.fileDir,
        fileName: row.fileName,
    };
    loading.value = true;
    await downloadBackupRecord(params)
        .then(async (res) => {
            loading.value = false;
            downloadFile(res.data, globalStore.currentNode);
        })
        .catch(() => {
            loading.value = false;
        });
};

const buttons = [
    {
        label: i18n.global.t('commons.button.download'),
        disabled: (row: any) => {
            return row.size === 0;
        },
        click: (row: Backup.RecordInfo) => {
            onDownload(row);
        },
    },
];

defineExpose({
    acceptParams,
});
</script>
