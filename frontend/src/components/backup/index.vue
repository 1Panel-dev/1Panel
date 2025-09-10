<template>
    <DrawerPro
        v-model="backupVisible"
        :header="$t('commons.button.backup')"
        :resource="detailName ? name + ' [' + detailName + ']' : name"
        @close="handleClose"
        size="60%"
    >
        <template #content>
            <el-alert v-if="type === 'app'" :closable="false" type="warning">
                <div class="mt-2 text-xs">
                    <span>{{ $t('setting.backupJump') }}</span>
                    <span class="jump" @click="goFile()">
                        <el-icon class="ml-2"><Position /></el-icon>
                        {{ $t('firewall.quickJump') }}
                    </span>
                </div>
            </el-alert>

            <ComplexTable
                class="mt-5"
                v-loading="loading"
                :pagination-config="paginationConfig"
                v-model:selects="selects"
                @search="search"
                :data="data"
                style="width: 100%"
            >
                <template #toolbar>
                    <el-button type="primary" :disabled="status && status != 'Running'" @click="onBackup()">
                        {{ $t('commons.button.backup') }}
                    </el-button>
                    <el-button type="primary" plain :disabled="selects.length === 0" @click="onBatchDelete(null)">
                        {{ $t('commons.button.delete') }}
                    </el-button>

                    <TableRefresh class="float-right" @search="search()" />
                </template>
                <el-table-column type="selection" :selectable="selectable" fix />
                <el-table-column :label="$t('commons.table.name')" prop="fileName" show-overflow-tooltip />
                <el-table-column min-width="80px" :label="$t('commons.table.status')" prop="status">
                    <template #default="{ row }">
                        <Status
                            v-if="row.status === 'Waiting'"
                            :status="row.status"
                            @click="openTaskLog(row.taskID)"
                            :msg="row.message"
                            :operate="true"
                        />
                        <Status v-else :status="row.status" :msg="row.message" />
                    </template>
                </el-table-column>
                <el-table-column min-width="80px" :label="$t('file.size')" prop="size" show-overflow-tooltip>
                    <template #default="{ row }">
                        <div v-if="row.hasLoad && (row.status === 'Success' || row.status === 'Failed')">
                            <span v-if="row.size">
                                {{ computeSize(row.size) }}
                            </span>
                            <span v-else>-</span>
                        </div>
                        <div v-else>
                            <el-button link loading></el-button>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column min-width="100px" :label="$t('app.source')" prop="backupType">
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
                    min-width="120px"
                    :label="$t('commons.table.description')"
                    prop="description"
                    show-overflow-tooltip
                >
                    <template #default="{ row }">
                        <fu-input-rw-switch v-model="row.description" @blur="onChange(row)" />
                    </template>
                </el-table-column>
                <el-table-column
                    min-width="80px"
                    prop="createdAt"
                    :label="$t('commons.table.date')"
                    :formatter="dateFormat"
                    show-overflow-tooltip
                />

                <fu-table-operations width="200px" :buttons="buttons" :label="$t('commons.table.operate')" fix />
            </ComplexTable>
        </template>
    </DrawerPro>

    <DialogPro
        v-model="open"
        :title="isBackup ? $t('commons.button.backup') : $t('commons.button.recover') + ' - ' + name"
        size="small"
        @close="handleBackupClose"
    >
        <el-alert :closable="false">
            {{ $t('commons.msg.' + (isBackup ? 'backupHelper' : 'recoverHelper'), [name + '( ' + detailName + ' )']) }}
        </el-alert>
        <el-form class="mt-5" ref="backupForm" @submit.prevent label-position="top" v-loading="loading">
            <el-form-item :label="$t('setting.compressPassword')" v-if="type === 'app' || type === 'website'">
                <el-input v-model="secret" :placeholder="$t('setting.backupRecoverMessage')" />
            </el-form-item>
            <el-form-item v-if="isBackup" :label="$t('commons.table.description')">
                <el-input type="textarea" :autosize="{ minRows: 2, maxRows: 5 }" v-model="description" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleBackupClose" :disabled="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="onSubmit" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DialogPro>

    <OpDialog ref="opRef" @search="search" />
    <TaskLog ref="taskLogRef" @close="search" />
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { computeSize, dateFormat, downloadFile, newUUID } from '@/utils/util';
import {
    getLocalBackupDir,
    handleBackup,
    handleRecover,
    deleteBackupRecord,
    downloadBackupRecord,
    searchBackupRecords,
    loadRecordSize,
    updateRecordDescription,
} from '@/api/modules/backup';
import i18n from '@/lang';
import { Backup } from '@/api/interface/backup';
import { MsgSuccess } from '@/utils/message';
import TaskLog from '@/components/log/task/index.vue';
import { GlobalStore } from '@/store';
import { routerToFileWithPath } from '@/utils/router';
const globalStore = GlobalStore();

const selects = ref<any>([]);
const loading = ref();
const opRef = ref();
const taskLogRef = ref();

const data = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const backupVisible = ref(false);
const type = ref();
const name = ref();
const detailName = ref();
const backupPath = ref();
const status = ref();
const secret = ref();
const description = ref();

const open = ref();
const isBackup = ref();
const recordInfo = ref();

interface DialogProps {
    type: string;
    name: string;
    detailName: string;
    status: string;
}
const acceptParams = (params: DialogProps): void => {
    type.value = params.type;
    if (type.value === 'app') {
        loadBackupDir();
    }
    name.value = params.name;
    detailName.value = params.detailName;
    backupVisible.value = true;
    status.value = params.status;
    secret.value = '';
    search();
};
const handleClose = () => {
    backupVisible.value = false;
};
const handleBackupClose = () => {
    open.value = false;
    search();
};

const loadBackupDir = async () => {
    const res = await getLocalBackupDir();
    backupPath.value = res.data;
};

const goFile = async () => {
    routerToFileWithPath(`${backupPath.value}/app/${name.value}/${detailName.value}`);
};

const onChange = async (info: any) => {
    await updateRecordDescription(info.id, info.description);
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
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
            loadSize(params);
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const loadSize = async (params: any) => {
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

const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

function selectable(row) {
    return row.status !== 'Waiting';
}

const backup = async (close: boolean) => {
    const taskID = newUUID();
    let params = {
        type: type.value,
        name: name.value,
        detailName: detailName.value,
        secret: secret.value,
        taskID: taskID,
        description: description.value,
    };
    loading.value = true;
    try {
        await handleBackup(params);
        loading.value = false;
        if (close) {
            handleClose();
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        } else {
            openTaskLog(taskID);
        }
        handleBackupClose();
    } catch (error) {
        loading.value = false;
    }
};

const recover = async (close: boolean, row?: any) => {
    const taskID = newUUID();
    let params = {
        downloadAccountID: row.downloadAccountID,
        type: type.value,
        name: name.value,
        detailName: detailName.value,
        file: row.fileDir + '/' + row.fileName,
        secret: secret.value,
        taskID: taskID,
        backupRecordID: row.id,
    };
    loading.value = true;
    await handleRecover(params)
        .then(() => {
            loading.value = false;
            handleBackupClose();
            if (close) {
                handleClose();
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                search();
            } else {
                openTaskLog(taskID);
            }
        })
        .catch(() => {
            loading.value = false;
        });
};

const onBackup = async () => {
    description.value = '';
    isBackup.value = true;
    open.value = true;
};

const onRecover = async (row: Backup.RecordInfo) => {
    secret.value = '';
    isBackup.value = false;
    if (type.value !== 'app' && type.value !== 'website') {
        ElMessageBox.confirm(
            i18n.global.t('commons.msg.recoverHelper', [name.value + '( ' + detailName.value + ' )']),
            i18n.global.t('commons.button.recover'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
            },
        ).then(async () => {
            recover(true, row);
        });
        return;
    }
    recordInfo.value = row;
    open.value = true;
};

const onSubmit = () => {
    if (isBackup.value) {
        backup(false);
    } else {
        recover(false, recordInfo.value);
    }
};

const onDownload = async (row: Backup.RecordInfo) => {
    let params = {
        downloadAccountID: row.downloadAccountID,
        fileDir: row.fileDir,
        fileName: row.fileName,
    };
    await downloadBackupRecord(params).then(async (res) => {
        downloadFile(res.data, globalStore.currentNode);
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
        disabled: (row: any) => {
            return row.status === 'Waiting';
        },
        click: (row: Backup.RecordInfo) => {
            onBatchDelete(row);
        },
    },
    {
        label: i18n.global.t('commons.button.recover'),
        disabled: (row: any) => {
            return row.size === 0 || row.status === 'Failed';
        },
        click: (row: Backup.RecordInfo) => {
            onRecover(row);
        },
    },
    {
        label: i18n.global.t('commons.button.download'),
        disabled: (row: any) => {
            return row.size === 0 || row.status === 'Failed';
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

<style lang="scss" scoped>
.jump {
    color: $primary-color;
    cursor: pointer;
    &:hover {
        color: #74a4f3;
    }
}
</style>
