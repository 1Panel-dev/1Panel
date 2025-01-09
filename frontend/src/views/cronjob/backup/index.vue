<template>
    <div>
        <el-drawer
            v-model="backupVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="50%"
        >
            <template #header>
                <DrawerHeader
                    v-if="cronjob"
                    :header="$t('commons.button.backup')"
                    :resource="cronjob"
                    :back="handleClose"
                />
                <DrawerHeader v-else :header="$t('commons.button.backup')" :resource="cronjob" :back="handleClose" />
            </template>
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
                <el-table-column :label="$t('database.source')" prop="backupType">
                    <template #default="{ row }">
                        <span v-if="row.source">
                            {{ $t('setting.' + row.source) }}
                        </span>
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
        </el-drawer>
    </div>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { computeSize, dateFormat, downloadFile } from '@/utils/util';
import i18n from '@/lang';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { downloadBackupRecord, loadCronjobBackupSize, searchBackupRecordsByCronjob } from '@/api/modules/setting';
import { Backup } from '@/api/interface/backup';

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
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
            if (paginationConfig.total !== 0) {
                loadSize(params);
            }
        })
        .catch(() => {
            loading.value = false;
        });
};

const loadSize = async (params: any) => {
    await loadCronjobBackupSize(params)
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
    let params = {
        source: row.source,
        fileDir: row.fileDir,
        fileName: row.fileName,
    };
    loading.value = true;
    await downloadBackupRecord(params)
        .then(async (res) => {
            loading.value = false;
            downloadFile(res.data);
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
