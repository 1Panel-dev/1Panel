<template>
    <div>
        <DrawerPro
            v-model="uploadOpen"
            :header="$t('commons.button.import')"
            :resource="title"
            @close="handleUploadClose"
            size="large"
        >
            <template #content>
                <div v-loading="loading">
                    <div>
                        <el-alert :closable="false" type="warning">
                            <template #default>
                                <ul>
                                    <li v-if="type === 'mysql' || type === 'mariadb'">
                                        {{ $t('database.formatHelper', [remark]) }}
                                    </li>
                                    <li v-if="type === 'website'">{{ $t('website.websiteBackupWarn') }}</li>
                                    <span v-if="isDb()">
                                        <li>{{ $t('database.supportUpType') }}</li>
                                        <li>{{ $t('database.zipFormat') }}</li>
                                    </span>
                                    <span v-else>
                                        <li>{{ $t('website.supportUpType') }}</li>
                                        <li>{{ $t('website.zipFormat', [type + '.json']) }}</li>
                                    </span>
                                </ul>
                            </template>
                        </el-alert>
                    </div>

                    <ComplexTable
                        :pagination-config="paginationConfig"
                        class="mt-5"
                        @search="search"
                        v-model:selects="selects"
                        :data="data"
                    >
                        <template #toolbar>
                            <el-upload
                                :limit="1"
                                class="float-left"
                                ref="uploadRef"
                                accept=".tar.gz,.sql,.sql.gz"
                                :show-file-list="false"
                                :on-exceed="handleExceed"
                                :on-change="fileOnChange"
                                :auto-upload="false"
                            >
                                <el-button class="float-left">
                                    {{ $t('database.localUpload') }}
                                </el-button>
                            </el-upload>
                            <el-button class="float-left ml-3" @click="fileRef.acceptParams({ dir: false })">
                                {{ $t('database.hostSelect') }}
                            </el-button>
                            <el-button :disabled="selects.length === 0" @click="onBatchDelete(null)">
                                {{ $t('commons.button.delete') }}
                            </el-button>

                            <el-progress v-if="isUpload" text-inside :stroke-width="12" :percentage="uploadPercent" />
                        </template>
                        <el-table-column type="selection" fix />
                        <el-table-column :label="$t('commons.table.name')" show-overflow-tooltip prop="name" />
                        <el-table-column :label="$t('file.size')" prop="size">
                            <template #default="{ row }">
                                {{ computeSize(row.size) }}
                            </template>
                        </el-table-column>
                        <el-table-column
                            show-overflow-tooltip
                            :label="$t('commons.table.createdAt')"
                            min-width="90"
                            fix
                        >
                            <template #default="{ row }">
                                {{ row.createdAt }}
                            </template>
                        </el-table-column>
                        <fu-table-operations
                            width="150px"
                            :buttons="buttons"
                            :ellipsis="10"
                            :label="$t('commons.table.operate')"
                            fix
                        />
                    </ComplexTable>
                </div>
            </template>
        </DrawerPro>

        <DialogPro
            v-model="recoverDialog"
            :title="$t('commons.button.recover') + ' - ' + name"
            @close="handleRecoverClose"
        >
            <el-form ref="backupForm" label-position="left" v-loading="loading">
                <el-form-item
                    :label="$t('setting.compressPassword')"
                    class="mt-5"
                    v-if="type === 'app' || type === 'website'"
                >
                    <el-input v-model="secret" :placeholder="$t('setting.backupRecoverMessage')" />
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="handleRecoverClose" :disabled="loading">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button type="primary" @click="onHandleRecover" :disabled="loading">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </DialogPro>

        <OpDialog ref="opRef" @search="search" />
        <FileList ref="fileRef" @choose="loadFile" />
        <TaskLog ref="taskLogRef" @close="search" />
    </div>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { computeSize, newUUID } from '@/utils/util';
import i18n from '@/lang';
import { UploadFile, UploadFiles, UploadInstance, UploadProps, UploadRawFile, genFileId } from 'element-plus';
import { File } from '@/api/interface/file';
import { batchDeleteFile, checkFile, chunkUploadFileData, getUploadList } from '@/api/modules/files';
import { loadBaseDir } from '@/api/modules/setting';
import { MsgError, MsgSuccess } from '@/utils/message';
import { handleRecoverByUpload, uploadByRecover } from '@/api/modules/backup';
import TaskLog from '@/components/log/task/index.vue';

interface DialogProps {
    type: string;
    name: string;
    detailName: string;
    remark: string;
}
const loading = ref();
const fileRef = ref();
const isUpload = ref();
const uploadPercent = ref<number>(0);
const selects = ref<any>([]);
const baseDir = ref();
const opRef = ref();
const currentRow = ref();
const data = ref();
const title = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const uploadOpen = ref(false);
const type = ref();
const name = ref();
const detailName = ref();
const remark = ref();
const secret = ref();
const taskLogRef = ref();

const recoverDialog = ref();

const acceptParams = async (params: DialogProps): Promise<void> => {
    type.value = params.type;
    name.value = params.name;
    detailName.value = params.detailName;
    remark.value = params.remark;

    const pathRes = await loadBaseDir();
    switch (type.value) {
        case 'mysql':
        case 'mariadb':
        case 'postgresql':
        case 'mysql-cluster':
        case 'postgresql-cluster':
            title.value = name.value + ' [ ' + detailName.value + ' ]';
            if (detailName.value) {
                baseDir.value = `${pathRes.data}/uploads/database/${type.value}/${name.value}/${detailName.value}/`;
            } else {
                baseDir.value = `${pathRes.data}/uploads/database/${type.value}/${name.value}/`;
            }
            break;
        case 'website':
            title.value = name.value;
            baseDir.value = `${pathRes.data}/uploads/website/${type.value}/${detailName.value}/`;
            break;
        case 'app':
            title.value = name.value;
            baseDir.value = `${pathRes.data}/uploads/app/${type.value}/${name.value}/`;
    }
    uploadOpen.value = true;
    search();
};

const search = async () => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        path: baseDir.value,
    };
    const res = await getUploadList(params);
    data.value = res.data.items || [];
    paginationConfig.total = res.data.total;
};

const loadFile = async (path: string) => {
    let filaName = path.split('/').pop();
    if (!filaName) {
        MsgError(i18n.global.t('commons.msg.fileNameErr'));
        return;
    }
    let reg = /^[a-zA-Z0-9\u4e00-\u9fa5]{1}[a-z:A-Z0-9_.\u4e00-\u9fa5-]{0,256}$/;
    if (!reg.test(filaName)) {
        MsgError(i18n.global.t('commons.msg.fileNameErr'));
        return;
    }
    ElMessageBox.confirm(i18n.global.t('database.selectHelper', [path]), i18n.global.t('database.loadBackup'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        uploadByRecover(path, baseDir.value)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                search();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const onHandleRecover = async () => {
    let params = {
        downloadAccountID: 1,
        type: type.value,
        name: name.value,
        detailName: detailName.value,
        file: baseDir.value + currentRow.value.name,
        secret: secret.value,
        taskID: newUUID(),
    };
    loading.value = true;
    await handleRecoverByUpload(params)
        .then(() => {
            loading.value = false;
            handleUploadClose();
            handleRecoverClose();
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
            openTaskLog(params.taskID);
        })
        .catch(() => {
            loading.value = false;
        });
};

const onRecover = async (row: File.File) => {
    currentRow.value = row;
    if (type.value !== 'app' && type.value !== 'website') {
        ElMessageBox.confirm(
            i18n.global.t('commons.msg.recoverHelper', [row.name]),
            i18n.global.t('commons.button.recover'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
            },
        ).then(async () => {
            onHandleRecover();
        });
        return;
    }
    recoverDialog.value = true;
};

const isDb = () => {
    return type.value === 'mysql' || type.value === 'mariadb' || type.value === 'postgresql';
};
const uploaderFiles = ref<UploadFiles>([]);
const uploadRef = ref<UploadInstance>();

const fileOnChange = (_uploadFile: UploadFile, uploadFiles: UploadFiles) => {
    uploaderFiles.value = uploadFiles;
    if (uploaderFiles.value.length !== 1) {
        return;
    }
    const file = uploaderFiles.value[0];
    if (!file.raw.name) {
        MsgError(i18n.global.t('commons.msg.fileNameErr'));
        return;
    }
    ElMessageBox.confirm(
        i18n.global.t('database.selectHelper', [file.raw.name]),
        i18n.global.t('database.loadBackup'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        },
    ).then(async () => {
        onSubmit();
    });
};

const handleUploadClose = () => {
    uploaderFiles.value = [];
    uploadRef.value!.clearFiles();
    uploadOpen.value = false;
};

const handleRecoverClose = () => {
    recoverDialog.value = false;
};

const handleExceed: UploadProps['onExceed'] = (files) => {
    uploadRef.value!.clearFiles();
    const file = files[0] as UploadRawFile;
    file.uid = genFileId();
    uploadRef.value!.handleStart(file);
};

const onSubmit = async () => {
    const file = uploaderFiles.value[0];
    let reg = /^[a-zA-Z0-9\u4e00-\u9fa5]{1}[a-z:A-Z0-9_.\u4e00-\u9fa5-]{0,256}$/;
    if (!reg.test(file.raw.name)) {
        MsgError(i18n.global.t('commons.msg.fileNameErr'));
        return;
    }
    const res = await checkFile(baseDir.value + file.raw.name, false);
    if (res.data) {
        MsgError(i18n.global.t('commons.msg.fileExist'));
        return;
    }
    submitUpload(file);
};

const submitUpload = async (file: any) => {
    isUpload.value = true;
    const CHUNK_SIZE = 1024 * 1024;
    const fileSize = file.size;
    const chunkCount = Math.ceil(fileSize / CHUNK_SIZE);
    let uploadedChunkCount = 0;

    for (let i = 0; i < chunkCount; i++) {
        const start = i * CHUNK_SIZE;
        const end = Math.min(start + CHUNK_SIZE, fileSize);
        const chunk = file.raw.slice(start, end);

        const formData = new FormData();

        formData.append('filename', file.name);
        formData.append('path', baseDir.value);
        formData.append('chunk', chunk);
        formData.append('chunkIndex', i.toString());
        formData.append('chunkCount', chunkCount.toString());

        try {
            await chunkUploadFileData(formData, {
                onUploadProgress: (progressEvent) => {
                    const progress = Math.round(
                        ((uploadedChunkCount + progressEvent.loaded / progressEvent.total) * 100) / chunkCount,
                    );
                    uploadPercent.value = progress;
                },
            });
            uploadedChunkCount++;
        } catch (error) {
            isUpload.value = false;
            break;
        }
        if (uploadedChunkCount == chunkCount) {
            isUpload.value = false;
            uploadRef.value?.clearFiles();
            uploaderFiles.value = [];
            MsgSuccess(i18n.global.t('file.uploadSuccess'));
            search();
        }
    }
};

const onBatchDelete = async (row: File.File | null) => {
    let files: Array<string> = [];
    let names: Array<string> = [];
    if (row) {
        files.push(baseDir.value + row.name);
        names.push(row.name);
    } else {
        selects.value.forEach((item: File.File) => {
            files.push(baseDir.value + item.name);
            names.push(item.name);
        });
    }
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('commons.button.import'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: batchDeleteFile,
        params: { paths: files, isDir: false },
    });
};

const buttons = [
    {
        label: i18n.global.t('commons.button.recover'),
        click: (row: File.File) => {
            onRecover(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: File.File) => {
            onBatchDelete(row);
        },
    },
];

defineExpose({
    acceptParams,
});
</script>
