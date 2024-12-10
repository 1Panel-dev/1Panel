<template>
    <el-drawer
        v-model="open"
        :before-close="handleClose"
        size="40%"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
    >
        <template #header>
            <DrawerHeader :header="$t('file.upload')" :back="handleClose" />
        </template>
        <div class="button-container">
            <div>
                <el-button type="primary" @click="upload('file')">
                    {{ $t('file.uploadFile') }}
                </el-button>
                <el-button type="primary" @click="upload('dir')">{{ $t('file.uploadDirectory') }}</el-button>
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
            :on-success="handleSuccess"
            :show-file-list="false"
            multiple
            v-model:file-list="uploaderFiles"
            :limit="1000"
        >
            <template #tip>
                <el-text>{{ uploadHelper }}</el-text>
                <el-progress v-if="loading" text-inside :stroke-width="20" :percentage="uploadPercent"></el-progress>
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
                        :disabled="loading"
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
import { MsgError, MsgSuccess, MsgWarning } from '@/utils/message';
import { Close, Document, UploadFilled } from '@element-plus/icons-vue';
import { TimeoutEnum } from '@/enums/http-enum';

interface UploadFileProps {
    path: string;
}

const uploadRef = ref<UploadInstance>();
const loading = ref(false);
let uploadPercent = ref(0);
const open = ref(false);
const path = ref();
let uploadHelper = ref('');

const em = defineEmits(['close']);
const handleClose = () => {
    open.value = false;
    clearFiles();
    em('close', false);
};
const state = reactive({
    uploadEle: null,
});
const uploaderFiles = ref<UploadFiles>([]);
const hoverIndex = ref<number | null>(null);
const tmpFiles = ref<UploadFiles>([]);
const breakFlag = ref(false);

const upload = (command: string) => {
    state.uploadEle.webkitdirectory = command == 'dir';
    uploadRef.value.$el.querySelector('input').value = '';
    uploadRef.value.$el.querySelector('input').click();
};

const removeFile = (index: number) => {
    uploaderFiles.value.splice(index, 1);
};

const handleDragover = (event: DragEvent) => {
    event.preventDefault();
};

const initTempFiles = () => {
    tmpFiles.value = [];
    breakFlag.value = false;
};

const handleDrop = async (event: DragEvent) => {
    initTempFiles();
    event.preventDefault();
    const items = event.dataTransfer?.items;
    if (items) {
        const entries = Array.from(items).map((item) => item.webkitGetAsEntry());
        await Promise.all(entries.map((entry) => traverseFileTree(entry)));
        if (!breakFlag.value) {
            uploaderFiles.value = uploaderFiles.value.concat(tmpFiles.value);
        } else {
            MsgWarning(i18n.global.t('file.uploadOverLimit'));
        }
        initTempFiles();
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

const traverseFileTree = async (item: any, path = '') => {
    path = path || '';
    if (!item) {
        return;
    }
    if (item.isFile) {
        if (tmpFiles.value.length > 1000) {
            breakFlag.value = true;
            return;
        }
        await new Promise<void>((resolve) => {
            item.file((file: File) => {
                if (!breakFlag.value) {
                    tmpFiles.value.push(convertFileToUploadFile(file, path));
                }
                resolve();
            });
        });
    } else if (item.isDirectory) {
        const dirReader = item.createReader();
        const readEntries = async () => {
            const entries = await new Promise<any[]>((resolve) => {
                dirReader.readEntries((entries: any[] | PromiseLike<any[]>) => {
                    resolve(entries);
                });
            });

            if (entries.length === 0) {
                return;
            }

            for (const element of entries) {
                await traverseFileTree(element, path + item.name + '/');
                if (breakFlag.value) {
                    return;
                }
            }
            await readEntries();
        };
        await readEntries();
    }
};

const handleDragleave = (event: { preventDefault: () => void }) => {
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
    uploaderFiles.value = [];
};

const handleExceed: UploadProps['onExceed'] = () => {
    clearFiles();
    MsgWarning(i18n.global.t('file.uploadOverLimit'));
};

const handleSuccess: UploadProps['onSuccess'] = (res, file) => {
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
        if (fileSize <= 1024 * 1024 * 10) {
            const formData = new FormData();
            formData.append('file', file.raw);
            if (file.raw.webkitRelativePath != '') {
                formData.append('path', path.value + '/' + getPathWithoutFilename(file.raw.webkitRelativePath));
            } else {
                formData.append('path', path.value + '/' + getPathWithoutFilename(file.name));
            }
            formData.append('overwrite', 'True');
            uploadPercent.value = 0;
            await UploadFileData(formData, {
                onUploadProgress: (progressEvent) => {
                    const progress = Math.round((progressEvent.loaded / progressEvent.total) * 100);
                    uploadPercent.value = progress;
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

                formData.append('filename', getFilenameFromPath(file.name));
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
                            uploadPercent.value = progress;
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
                clearFiles();
                MsgSuccess(i18n.global.t('file.uploadSuccess'));
            }
        }
    }
};

const getPathWithoutFilename = (path: string) => {
    return path ? path.split('/').slice(0, -1).join('/') : path;
};

const getFilenameFromPath = (path: string) => {
    return path ? path.split('/').pop() : path;
};

const acceptParams = (props: UploadFileProps) => {
    path.value = props.path;
    open.value = true;
    uploadPercent.value = 0;
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
