<template>
    <el-dialog v-model="open" :title="$t('file.upload')" :before-close="handleClose" width="30%" :file-list="files">
        <el-upload
            action="#"
            :auto-upload="false"
            ref="uploadRef"
            :multiple="true"
            :on-change="fileOnChange"
            v-loading="loading"
        >
            <template #trigger>
                <el-button type="primary">{{ $t('file.selectFile') }}</el-button>
            </template>
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
    </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { ElMessage, UploadFile, UploadFiles, UploadInstance } from 'element-plus';
import { UploadFileData } from '@/api/modules/files';
import i18n from '@/lang';

const props = defineProps({
    open: {
        type: Boolean,
        default: false,
    },
    path: {
        type: String,
        default: '',
    },
});

const uploadRef = ref<UploadInstance>();
const files = ref();
const loading = ref(false);
let uploadPrecent = ref(0);

const em = defineEmits(['close']);
const handleClose = () => {
    uploadRef.value!.clearFiles();
    em('close', false);
};

const uploaderFiles = ref<UploadFiles>([]);

const fileOnChange = (_uploadFile: UploadFile, uploadFiles: UploadFiles) => {
    uploaderFiles.value = uploadFiles;
};

const onProcess = (e: any) => {
    const { loaded, total } = e;
    uploadPrecent.value = ((loaded / total) * 100) | 0;
};

const submit = () => {
    const formData = new FormData();
    for (const file of uploaderFiles.value) {
        if (file.raw != undefined) {
            formData.append('file', file.raw);
        }
    }
    formData.append('path', props.path);
    loading.value = true;
    UploadFileData(formData, { onUploadProgress: onProcess })
        .then(() => {
            ElMessage.success(i18n.global.t('file.uploadSuccess'));
            handleClose();
        })
        .finally(() => {
            loading.value = false;
        });
};
</script>
