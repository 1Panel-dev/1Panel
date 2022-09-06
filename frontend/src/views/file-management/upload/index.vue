<template>
    <el-dialog v-model="open" :title="$t('file.upload')" @open="onOpen" :before-close="handleClose">
        <el-upload action="#" :auto-upload="false" ref="uploadRef" :multiple="true" :on-change="fileOnChange">
            <template #trigger>
                <el-button type="primary">{{ $t('file.selectFile') }}</el-button>
            </template>
        </el-upload>

        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit()">{{ $t('commons.button.confirm') }}</el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { ElMessage, UploadFile, UploadFiles, UploadInstance } from 'element-plus';
import { UploadFileData } from '@/api/modules/files';

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

const em = defineEmits(['close']);
const handleClose = () => {
    em('close', false);
};

const uploaderFiles = ref<UploadFiles>([]);

const fileOnChange = (_uploadFile: UploadFile, uploadFiles: UploadFiles) => {
    uploaderFiles.value = uploadFiles;
};

const submit = () => {
    const formData = new FormData();
    for (const file of uploaderFiles.value) {
        if (file.raw != undefined) {
            formData.append('file', file.raw);
        }
    }
    formData.append('path', props.path);

    UploadFileData(formData).then(() => {
        ElMessage('upload success');
        handleClose();
    });
};

const onOpen = () => {};
</script>
