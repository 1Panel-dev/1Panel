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
        <el-upload
            action="#"
            drag
            :auto-upload="false"
            ref="uploadRef"
            :multiple="true"
            :on-change="fileOnChange"
            v-loading="loading"
        >
            <el-icon class="el-icon--upload"><upload-filled /></el-icon>
            <div class="el-upload__text">
                {{ $t('database.dropHelper') }}
                <em>{{ $t('database.clickHelper') }}</em>
            </div>
        </el-upload>
        <el-progress v-if="loading" :text-inside="true" :stroke-width="26" :percentage="uploadPrecent"></el-progress>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit()" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { UploadFile, UploadFiles, UploadInstance } from 'element-plus';
import { UploadFileData } from '@/api/modules/files';
import i18n from '@/lang';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';

interface UploadProps {
    path: string;
}

const uploadRef = ref<UploadInstance>();
const loading = ref(false);
let uploadPrecent = ref(0);
let open = ref(false);
let path = ref();

const em = defineEmits(['close']);
const handleClose = () => {
    open.value = false;
    uploadRef.value!.clearFiles();
    em('close', false);
};

const uploaderFiles = ref<UploadFiles>([]);

const fileOnChange = (_uploadFile: UploadFile, uploadFiles: UploadFiles) => {
    uploaderFiles.value = uploadFiles;
};

// const onProcess = (e: any) => {
//     const { loaded, total } = e;
//     uploadPrecent.value = ((loaded / total) * 100) | 0;
// };

const submit = async () => {
    loading.value = true;
    const file = uploaderFiles.value[0];

    const CHUNK_SIZE = 1024 * 1024; // 1MB
    const fileSize = file.size;
    const chunkCount = Math.ceil(fileSize / CHUNK_SIZE);
    let uploadedChunkCount = 0;

    for (let i = 0; i < chunkCount; i++) {
        const start = i * CHUNK_SIZE;
        const end = Math.min(start + CHUNK_SIZE, fileSize);
        const chunk = file.raw.slice(start, end);

        const formData = new FormData();

        formData.append('filename', file.name);
        formData.append('path', path.value);
        formData.append('chunk', chunk);
        formData.append('chunkIndex', i.toString());
        formData.append('chunkCount', chunkCount.toString());

        try {
            await UploadFileData(formData, {
                onUploadProgress: (progressEvent) => {
                    const progress = Math.round(
                        ((uploadedChunkCount + progressEvent.loaded / progressEvent.total) * 100) / chunkCount,
                    );
                    uploadPrecent.value = progress;
                },
            });
            uploadedChunkCount++;
        } catch (error) {
            loading.value = false;
        }
        if (uploadedChunkCount == chunkCount) {
            loading.value = false;
            MsgSuccess(i18n.global.t('file.uploadSuccess'));
        }
    }
};

const acceptParams = (props: UploadProps) => {
    path.value = props.path;
    open.value = true;
    uploadPrecent.value = 0;
};

defineExpose({ acceptParams });
</script>
