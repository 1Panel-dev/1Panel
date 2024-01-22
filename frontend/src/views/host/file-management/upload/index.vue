<template>
    <el-drawer
        v-model="open"
        :before-close="handleClose"
        size="40%"
        :destroy-on-close="true"
        :close-on-click-modal="false"
    >
        <template #header>
            <DrawerHeader :header="$t('file.upload')" :back="handleClose" />
        </template>
        <div class="button-container">
            <div>
                <el-button type="primary" @click="upload('file')">
                    {{ $t('file.upload') }}{{ $t('file.file') }}
                </el-button>
                <el-button type="primary" @click="upload('dir')">{{ $t('file.upload') }}{{ $t('file.dir') }}</el-button>
            </div>
            <el-button @click="clearFiles">{{ $t('file.clearList') }}</el-button>
        </div>

        <div>
            <div class="el-upload-dragger" @dragover="handleDragover" @drop="handleDrop" @dragleave="handleDragleave">
                <div class="flex items-center justify-center h-52">
                    <div>
                        <el-icon class="el-icon--upload"><upload-filled /></el-icon>
                        <div class="el-upload__text">
                            {{ $t('file.dropHelper') }}
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <el-upload
            action="#"
            :auto-upload="false"
            ref="uploadRef"
            :on-change="fileOnChange"
            :on-exceed="handleExceed"
            :on-success="hadleSuccess"
            :show-file-list="false"
            multiple
            v-model:file-list="uploaderFiles"
            :limit="10"
        >
            <template #tip>
                <el-text>{{ uploadHelper }}</el-text>
                <el-progress v-if="loading" text-inside :stroke-width="20" :percentage="uploadPrecent"></el-progress>
            </template>
        </el-upload>

        <div>
            <p
                v-for="(item, index) in uploaderFiles"
                :key="index"
                class="file-item"
                @mouseover="hoverIndex = index"
                @mouseout="hoverIndex = null"
            >
                <el-icon class="file-icon"><Document /></el-icon>
                <span v-if="item.raw.webkitRelativePath != ''">{{ item.raw.webkitRelativePath }}</span>
                <span v-else>{{ item.name }}</span>
                <span v-if="item.status === 'success'" class="success-icon">
                    <el-icon><Select /></el-icon>
                </span>
                <span v-else>
                    <el-button
                        class="delete-button"
                        type="primary"
                        link
                        @click="removeFile(index)"
                        :icon="Close"
                    ></el-button>
                </span>
            </p>
        </div>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit()" :disabled="loading || uploaderFiles.length == 0">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script setup lang="ts">
import { nextTick, reactive, ref } from 'vue';
import { UploadFile, UploadFiles, UploadInstance, UploadProps, UploadRawFile } from 'element-plus';
import { ChunkUploadFileData, UploadFileData } from '@/api/modules/files';
import i18n from '@/lang';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgError, MsgSuccess } from '@/utils/message';
import { Close } from '@element-plus/icons-vue';
import { TimeoutEnum } from '@/enums/http-enum';

interface UploadFileProps {
    path: string;
}

const uploadRef = ref<UploadInstance>();
const loading = ref(false);
let uploadPrecent = ref(0);
const open = ref(false);
const path = ref();
let uploadHelper = ref('');

const em = defineEmits(['close']);
const handleClose = () => {
    open.value = false;
    uploadRef.value!.clearFiles();
    em('close', false);
};
const state = reactive({
    uploadEle: null,
});
const uploaderFiles = ref<UploadFiles>([]);
const isUploadFolder = ref(false);
const hoverIndex = ref(null);
const uploadType = ref('file');

const upload = (commnad: string) => {
    uploadType.value = commnad;
    if (commnad == 'dir') {
        state.uploadEle.webkitdirectory = true;
    } else {
        state.uploadEle.webkitdirectory = false;
    }
    isUploadFolder.value = true;
    uploadRef.value.$el.querySelector('input').click();
};

const removeFile = (index: number) => {
    uploaderFiles.value.splice(index, 1);
};

const handleDragover = (event: DragEvent) => {
    event.preventDefault();
};

const handleDrop = (event: DragEvent) => {
    event.preventDefault();
    const items = event.dataTransfer.items;

    if (items) {
        for (let i = 0; i < items.length; i++) {
            const entry = items[i].webkitGetAsEntry();
            if (entry) {
                traverseFileTree(entry);
            }
        }
    }
};

const convertFileToUploadFile = (file: File, path: string): UploadFile => {
    const uid = Date.now();

    const uploadRawFile: UploadRawFile = new File([file], file.name, {
        type: file.type,
        lastModified: file.lastModified,
    }) as UploadRawFile;
    uploadRawFile.uid = uid;

    let fileName = file.name;
    if (path != '') {
        fileName = path + file.name;
    }
    return {
        name: fileName,
        size: file.size,
        status: 'ready',
        uid: uid,
        raw: uploadRawFile,
    };
};

const traverseFileTree = (item: any, path = '') => {
    path = path || '';
    if (item.isFile) {
        item.file((file: File) => {
            uploaderFiles.value.push(convertFileToUploadFile(file, path));
        });
    } else if (item.isDirectory) {
        const dirReader = item.createReader();
        dirReader.readEntries((entries) => {
            for (let i = 0; i < entries.length; i++) {
                traverseFileTree(entries[i], path + item.name + '/');
            }
        });
    }
};

const handleDragleave = (event) => {
    event.preventDefault();
};

const fileOnChange = (_uploadFile: UploadFile, uploadFiles: UploadFiles) => {
    if (_uploadFile.size == 64 || _uploadFile.size == 0) {
        uploaderFiles.value = uploadFiles;
        const reader = new FileReader();
        reader.readAsDataURL(_uploadFile.raw);
        reader.onload = async () => {};
        reader.onerror = () => {
            uploaderFiles.value = uploaderFiles.value.filter((file) => file.uid !== _uploadFile.uid);
            MsgError(i18n.global.t('file.typeErrOrEmpty', [_uploadFile.name]));
        };
    } else {
        uploaderFiles.value = uploadFiles;
    }
};

const clearFiles = () => {
    uploadRef.value!.clearFiles();
};

const handleExceed: UploadProps['onExceed'] = (files) => {
    uploadRef.value!.clearFiles();
    for (let i = 0; i < files.length; i++) {
        const file = files[i] as UploadRawFile;
        uploadRef.value!.handleStart(file);
    }
};

const hadleSuccess: UploadProps['onSuccess'] = (res, file) => {
    file.status = 'success';
};

const submit = async () => {
    loading.value = true;
    let success = 0;
    const files = uploaderFiles.value.slice();
    for (let i = 0; i < files.length; i++) {
        const file = files[i];
        const fileSize = file.size;

        uploadHelper.value = i18n.global.t('file.fileUploadStart', [file.name]);
        if (fileSize <= 1024 * 1024 * 5) {
            const formData = new FormData();
            formData.append('file', file.raw);
            if (file.raw.webkitRelativePath != '') {
                formData.append('path', path.value + '/' + getPathWithoutFilename(file.raw.webkitRelativePath));
            } else {
                formData.append('path', path.value + '/' + getPathWithoutFilename(file.name));
            }
            uploadPrecent.value = 0;
            await UploadFileData(formData, {
                onUploadProgress: (progressEvent) => {
                    const progress = Math.round((progressEvent.loaded / progressEvent.total) * 100);
                    uploadPrecent.value = progress;
                },
                timeout: 40000,
            });
            success++;
            uploaderFiles.value[i].status = 'success';
        } else {
            const CHUNK_SIZE = 1024 * 1024 * 5;
            const chunkCount = Math.ceil(fileSize / CHUNK_SIZE);
            let uploadedChunkCount = 0;
            for (let c = 0; c < chunkCount; c++) {
                const start = c * CHUNK_SIZE;
                const end = Math.min(start + CHUNK_SIZE, fileSize);
                const chunk = file.raw.slice(start, end);
                const formData = new FormData();

                formData.append('filename', file.name);
                if (file.raw.webkitRelativePath != '') {
                    formData.append('path', path.value + '/' + getPathWithoutFilename(file.raw.webkitRelativePath));
                } else {
                    formData.append('path', path.value + '/' + getPathWithoutFilename(file.name));
                }
                formData.append('chunk', chunk);
                formData.append('chunkIndex', c.toString());
                formData.append('chunkCount', chunkCount.toString());

                try {
                    await ChunkUploadFileData(formData, {
                        onUploadProgress: (progressEvent) => {
                            const progress = Math.round(
                                ((uploadedChunkCount + progressEvent.loaded / progressEvent.total) * 100) / chunkCount,
                            );
                            uploadPrecent.value = progress;
                        },
                        timeout: TimeoutEnum.T_60S,
                    });
                    uploadedChunkCount++;
                } catch (error) {
                    uploaderFiles.value[i].status = 'fail';
                    break;
                }
                if (uploadedChunkCount == chunkCount) {
                    success++;
                    uploaderFiles.value[i].status = 'success';
                    break;
                }
            }
        }

        if (i == files.length - 1) {
            loading.value = false;
            uploadHelper.value = '';
            if (success == files.length) {
                uploadRef.value!.clearFiles();
                uploaderFiles.value = [];
                MsgSuccess(i18n.global.t('file.uploadSuccess'));
            }
        }
    }
};

const getPathWithoutFilename = (path: string) => {
    return path ? path.split('/').slice(0, -1).join('/') : path;
};

const acceptParams = (props: UploadFileProps) => {
    path.value = props.path;
    open.value = true;
    uploadPrecent.value = 0;
    uploadHelper.value = '';

    nextTick(() => {
        const uploadEle = document.querySelector('.el-upload__input');
        state.uploadEle = uploadEle;
    });
};

defineExpose({ acceptParams });
</script>

<style lang="scss" scoped>
.button-container {
    display: flex;
    justify-content: space-between;
    margin-bottom: 10px;
}

.file-item {
    font-size: 14px;
    color: #888;
    position: relative;
    display: flex;
    align-items: center;
}

.file-item:hover {
    background-color: #f5f5f5;
}

.file-icon {
    margin-right: 8px;
}

.delete-button {
    position: absolute;
    right: 0;
    top: 50%;
    transform: translateY(-50%);
}

.success-icon {
    color: green;
    position: absolute;
    right: 0;
}
</style>
