<template>
    <div>
        <el-drawer v-model="upVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="50%">
            <template #header>
                <DrawerHeader :header="$t('commons.button.import')" :resource="title" :back="handleClose" />
            </template>
            <div v-loading="loading">
                <el-upload ref="uploadRef" drag :on-change="fileOnChange" class="upload-demo" :auto-upload="false">
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
                            :percentage="uploadPrecent"
                        ></el-progress>
                        <div v-if="type === 'mysql'" style="width: 80%" class="el-upload__tip">
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
                <el-button :disabled="isUpload" v-if="uploaderFiles.length === 1" icon="Upload" @click="onSubmit">
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
                    <el-table-column :label="$t('commons.table.createdAt')" min-width="80" fix>
                        <template #default="{ row }">
                            {{ row.createdAt }}
                        </template>
                    </el-table-column>
                    <fu-table-operations
                        width="300px"
                        :buttons="buttons"
                        :ellipsis="10"
                        :label="$t('commons.table.operate')"
                        fix
                    />
                </ComplexTable>
            </div>
        </el-drawer>
    </div>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { computeSize } from '@/utils/util';
import { useDeleteData } from '@/hooks/use-delete-data';
import { handleRecoverByUpload } from '@/api/modules/setting';
import i18n from '@/lang';
import { UploadFile, UploadFiles, UploadInstance } from 'element-plus';
import { File } from '@/api/interface/file';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { BatchDeleteFile, CheckFile, ChunkUploadFileData, GetUploadList } from '@/api/modules/files';
import { loadBaseDir } from '@/api/modules/setting';
import { MsgError, MsgSuccess } from '@/utils/message';

const loading = ref();
const isUpload = ref();
const uploadPrecent = ref<number>(0);
const selects = ref<any>([]);
const baseDir = ref();

const data = ref();
const title = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const upVisiable = ref(false);
const type = ref();
const name = ref();
const detailName = ref();
interface DialogProps {
    type: string;
    name: string;
    detailName: string;
}
const acceptParams = async (params: DialogProps): Promise<void> => {
    type.value = params.type;
    name.value = params.name;
    detailName.value = params.detailName;

    const pathRes = await loadBaseDir();
    if (type.value === 'mysql') {
        title.value = name.value + ' [ ' + detailName.value + ' ]';
    }
    if (type.value === 'website' || type.value === 'app') {
        title.value = name.value;
    }
    if (detailName.value) {
        baseDir.value = `${pathRes.data}/uploads/${type.value}/${name.value}/${detailName.value}/`;
    } else {
        baseDir.value = `${pathRes.data}/uploads/${type.value}/${name.value}/`;
    }
    upVisiable.value = true;
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
    let params = {
        source: 'LOCAL',
        type: type.value,
        name: name.value,
        detailName: detailName.value,
        file: baseDir.value + row.name,
    };
    loading.value = true;
    await handleRecoverByUpload(params)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
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
    upVisiable.value = false;
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
    if (!res.data) {
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
                    uploadPrecent.value = progress;
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
    if (row) {
        files.push(baseDir.value + row.name);
    } else {
        selects.value.forEach((item: File.File) => {
            files.push(baseDir.value + item.name);
        });
    }
    await useDeleteData(BatchDeleteFile, { paths: files, isDir: false }, 'commons.msg.delete');
    search();
};

const buttons = [
    {
        label: i18n.global.t('commons.button.recover'),
        click: (row: File.File) => {
            if (type.value !== 'app') {
                onRecover(row);
            } else {
                ElMessageBox.confirm(i18n.global.t('app.restoreWarn'), i18n.global.t('commons.button.recover'), {
                    confirmButtonText: i18n.global.t('commons.button.confirm'),
                    cancelButtonText: i18n.global.t('commons.button.cancel'),
                }).then(async () => {
                    onRecover(row);
                });
            }
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
