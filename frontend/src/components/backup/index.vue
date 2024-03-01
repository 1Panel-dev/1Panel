<template>
    <div>
        <el-drawer v-model="backupVisible" :destroy-on-close="true" :close-on-click-modal="false" size="50%">
            <template #header>
                <DrawerHeader
                    v-if="detailName"
                    :header="$t('commons.button.backup')"
                    :resource="name + '(' + detailName + ')'"
                    :back="handleClose"
                />
                <DrawerHeader v-else :header="$t('commons.button.backup')" :resource="name" :back="handleClose" />
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
                        {{ $t('commons.button.backup') }}
                    </el-button>
                    <el-button type="primary" plain :disabled="selects.length === 0" @click="onBatchDelete(null)">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </template>
                <el-table-column type="selection" fix />
                <el-table-column :label="$t('commons.table.name')" prop="fileName" show-overflow-tooltip />
                <el-table-column :label="$t('file.size')" prop="size" show-overflow-tooltip>
                    <template #default="{ row }">
                        <span v-if="row.size">
                            {{ computeSize(row.size) }}
                        </span>
                        <span v-else>-</span>
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

                <fu-table-operations width="230px" :buttons="buttons" :label="$t('commons.table.operate')" fix />
            </ComplexTable>
        </el-drawer>

        <OpDialog ref="opRef" @search="search" />
    </div>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { computeSize, dateFormat, downloadFile } from '@/utils/util';
import { handleBackup, handleRecover } from '@/api/modules/setting';
import i18n from '@/lang';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { deleteBackupRecord, downloadBackupRecord, searchBackupRecords } from '@/api/modules/setting';
import { Backup } from '@/api/interface/backup';
import { MsgSuccess } from '@/utils/message';

const selects = ref<any>([]);
const loading = ref();
const opRef = ref();

const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'backup-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const backupVisible = ref(false);
const type = ref();
const name = ref();
const detailName = ref();

interface DialogProps {
    type: string;
    name: string;
    detailName: string;
}
const acceptParams = (params: DialogProps): void => {
    type.value = params.type;
    name.value = params.name;
    detailName.value = params.detailName;
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
        type: type.value,
        name: name.value,
        detailName: detailName.value,
    };
    loading.value = true;
    await searchBackupRecords(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const onBackup = async () => {
    ElMessageBox.confirm(
        i18n.global.t('commons.msg.backupHelper', [name.value + '( ' + detailName.value + ' )']),
        i18n.global.t('commons.button.backup'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        },
    ).then(async () => {
        let params = {
            type: type.value,
            name: name.value,
            detailName: detailName.value,
        };
        loading.value = true;
        await handleBackup(params)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const onRecover = async (row: Backup.RecordInfo) => {
    ElMessageBox.confirm(
        i18n.global.t('commons.msg.recoverHelper', [row.fileName]),
        i18n.global.t('commons.button.recover'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        },
    ).then(async () => {
        let params = {
            source: row.source,
            type: type.value,
            name: name.value,
            detailName: detailName.value,
            file: row.fileDir + '/' + row.fileName,
        };
        loading.value = true;
        await handleRecover(params)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const onDownload = async (row: Backup.RecordInfo) => {
    let params = {
        source: row.source,
        fileDir: row.fileDir,
        fileName: row.fileName,
    };
    await downloadBackupRecord(params).then(async (res) => {
        downloadFile(res.data);
    });
};

const onBatchDelete = async (row: Backup.RecordInfo | null) => {
    let ids: Array<number> = [];
    let names = [];
    if (row) {
        ids.push(row.id);
        names.push(row.fileName);
    } else {
        selects.value.forEach((item: Backup.RecordInfo) => {
            ids.push(item.id);
            names.push(item.fileName);
        });
    }
    opRef.value.acceptParams({
        names: names,
        title: i18n.global.t('commons.button.delete'),
        api: deleteBackupRecord,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('commons.button.backup'),
            i18n.global.t('commons.button.delete'),
        ]),
        params: { ids: ids },
    });
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
        disabled: (row: any) => {
            return row.size === 0;
        },
        click: (row: Backup.RecordInfo) => {
            onRecover(row);
        },
    },
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
