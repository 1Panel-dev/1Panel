<template>
    <div>
        <el-drawer
            v-model="upVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="50%"
        >
            <template #header>
                <DrawerHeader :header="$t('commons.button.import')" :resource="title" :back="handleClose" />
            </template>
            <div v-loading="loading">
                <div class="mb-4" v-if="type === 'mysql' || type === 'mariadb'">
                    <el-alert type="error" :title="$t('database.formatHelper', [remark])" />
                </div>
                <div class="mb-4" v-if="type === 'website'">
                    <el-alert :closable="false" type="warning" :title="$t('website.websiteBackupWarn')"></el-alert>
                </div>
                <el-upload
                    :limit="1"
                    ref="uploadRef"
                    drag
                    :on-exceed="handleExceed"
                    :on-change="fileOnChange"
                    class="upload-demo"
                    :auto-upload="false"
                >
                    <el-icon class="el-icon--upload"><upload-filled /></el-icon>
                    <div class="el-upload__text">
                        {{ $t('database.dropHelper') }}
                        <em>{{ $t('database.clickHelper') }}</em>
                    </div>
                    <template #tip>
                        <el-progress
                            v-if="isUpload"
                            text-inside
                            :stroke-width="12"
                            :percentage="uploadPercent"
                        ></el-progress>
                        <div
                            v-if="type === 'mysql' || type === 'mariadb' || type === 'postgresql'"
                            style="width: 80%"
                            class="el-upload__tip"
                        >
                            <span class="input-help">{{ $t('database.supportUpType') }}</span>
                            <span class="input-help">
                                {{ $t('database.zipFormat') }}
                            </span>
                        </div>
                        <div v-else style="width: 80%" class="el-upload__tip">
                            <span class="input-help">{{ $t('website.supportUpType') }}</span>
                            <span class="input-help">
                                {{ $t('website.zipFormat', [type + '.json']) }}
                            </span>
                        </div>
                    </template>
                </el-upload>
                <el-button :disabled="isUpload || uploaderFiles.length !== 1" icon="Upload" @click="onSubmit">
                    {{ $t('commons.button.upload') }}
                </el-button>

                <el-divider />
                <ComplexTable
                    :pagination-config="paginationConfig"
                    @search="search"
                    v-model:selects="selects"
                    :data="data"
                >
                    <template #toolbar>
                        <el-button
                            style="margin-left: 10px"
                            plain
                            :disabled="selects.length === 0"
                            @click="onBatchDelete(null)"
                        >
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </template>
                    <el-table-column type="selection" fix />
                    <el-table-column :label="$t('commons.table.name')" show-overflow-tooltip prop="name" />
                    <el-table-column :label="$t('file.size')" prop="size">
                        <template #default="{ row }">
                            {{ computeSize(row.size) }}
                        </template>
                    </el-table-column>
                    <el-table-column show-overflow-tooltip :label="$t('commons.table.createdAt')" min-width="90" fix>
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
        </el-drawer>

        <el-dialog
            v-model="open"
            :title="$t('commons.button.recover') + ' - ' + name"
            width="40%"
            :close-on-click-modal="false"
            :before-close="handleBackupClose"
        >
            <el-form ref="backupForm" label-position="left" v-loading="loading">
                <el-form-item
                    :label="$t('setting.compressPassword')"
                    style="margin-top: 10px"
                    v-if="type === 'app' || type === 'website'"
                >
                    <el-input v-model="secret" :placeholder="$t('setting.backupRecoverMessage')" />
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="handleBackupClose" :disabled="loading">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button type="primary" @click="onHandleRecover" :disabled="loading">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>

        <OpDialog ref="opRef" @search="search" />
    </div>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { computeSize } from '@/utils/util';
import i18n from '@/lang';
import { UploadFile, UploadFiles, UploadInstance, UploadProps, UploadRawFile, genFileId } from 'element-plus';
import { File } from '@/api/interface/file';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { BatchDeleteFile, CheckFile, ChunkUploadFileData, GetUploadList } from '@/api/modules/files';
import { handleRecoverByUpload, loadBaseDir } from '@/api/modules/setting';
import { MsgError, MsgSuccess } from '@/utils/message';

const loading = ref();
const isUpload = ref();
const uploadPercent = ref<number>(0);
const selects = ref<any>([]);
const baseDir = ref();
const opRef = ref();

const open = ref();
const currentRow = ref();

const data = ref();
const title = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const upVisible = ref(false);
const type = ref();
const name = ref();
const detailName = ref();
const remark = ref();
const secret = ref();
interface DialogProps {
    type: string;
    name: string;
    detailName: string;
    remark: string;
}
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
    upVisible.value = true;
    search();
};

const search = async () => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        path: baseDir.value,
    };
    const res = await GetUploadList(params);
    data.value = res.data.items || [];
    paginationConfig.total = res.data.total;
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
    open.value = true;
};

const handleBackupClose = () => {
    open.value = false;
};
const onHandleRecover = async () => {
    let params = {
        source: 'LOCAL',
        type: type.value,
        name: name.value,
        detailName: detailName.value,
        file: baseDir.value + currentRow.value.name,
        secret: secret.value,
        recoverType: 'upload',
    };
    loading.value = true;
    await handleRecoverByUpload(params)
        .then(() => {
            loading.value = false;
            handleClose();
            handleBackupClose();
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};

const uploaderFiles = ref<UploadFiles>([]);
const uploadRef = ref<UploadInstance>();

const beforeAvatarUpload = (rawFile) => {
    if (type.value === 'app' || type.value === 'website') {
        if (!rawFile.name.endsWith('.tar.gz')) {
            MsgError(i18n.global.t('commons.msg.unSupportType'));
            return false;
        }
        return true;
    }
    if (!rawFile.name.endsWith('.sql') && !rawFile.name.endsWith('.tar.gz') && !rawFile.name.endsWith('.sql.gz')) {
        MsgError(i18n.global.t('commons.msg.unSupportType'));
        return false;
    }
    return true;
};

const fileOnChange = (_uploadFile: UploadFile, uploadFiles: UploadFiles) => {
    uploaderFiles.value = uploadFiles;
};

const handleClose = () => {
    uploaderFiles.value = [];
    uploadRef.value!.clearFiles();
    upVisible.value = false;
};

const handleExceed: UploadProps['onExceed'] = (files) => {
    uploadRef.value!.clearFiles();
    const file = files[0] as UploadRawFile;
    file.uid = genFileId();
    uploadRef.value!.handleStart(file);
};

const onSubmit = async () => {
    if (uploaderFiles.value.length !== 1) {
        return;
    }
    const file = uploaderFiles.value[0];
    if (!file.raw.name) {
        MsgError(i18n.global.t('commons.msg.fileNameErr'));
        return;
    }
    let reg = /^[a-zA-Z0-9\u4e00-\u9fa5]{1}[a-z:A-Z0-9_.\u4e00-\u9fa5-]{0,256}$/;
    if (!reg.test(file.raw.name)) {
        MsgError(i18n.global.t('commons.msg.fileNameErr'));
        return;
    }
    const res = await CheckFile(baseDir.value + file.raw.name);
    if (res.data) {
        MsgError(i18n.global.t('commons.msg.fileExist'));
        return;
    }
    let isOk = beforeAvatarUpload(file.raw);
    if (!isOk) {
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
            await ChunkUploadFileData(formData, {
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
        api: BatchDeleteFile,
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
