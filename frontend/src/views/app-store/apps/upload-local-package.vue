<template>
    <el-dialog v-model="open" width="520px" :title="$t('app.uploadLocalAppPackage')" @close="handleClose">
        <el-alert class="mb-4" type="info" :closable="false" :title="$t('app.uploadLocalAppPackageHelper')" />
        <el-upload
            ref="uploadRef"
            v-model:file-list="fileList"
            drag
            :auto-upload="false"
            :limit="1"
            accept=".tar.gz"
            :on-exceed="handleExceed"
            :before-upload="beforeUpload"
        >
            <el-icon class="el-icon--upload"><UploadFilled /></el-icon>
            <div class="el-upload__text">
                {{ $t('file.dropHelper') }}
                <em>{{ $t('file.clickHelper') }}</em>
            </div>
            <template #tip>
                <div class="el-upload__tip">{{ $t('xpack.customApp.appStoreUrlHelper') }}</div>
            </template>
        </el-upload>
        <el-checkbox v-model="overwrite" class="mt-4">
            {{ $t('app.overwriteLocalAppPackage') }}
        </el-checkbox>
        <template #footer>
            <el-button :disabled="loading" @click="open = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" :loading="loading" @click="submit">
                {{ $t('commons.button.upload') }}
            </el-button>
        </template>
    </el-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { genFileId, type UploadInstance, type UploadProps, type UploadRawFile, type UploadUserFile } from 'element-plus';
import { UploadFilled } from '@element-plus/icons-vue';
import { uploadLocalAppPackage } from '@/api/modules/app';
import { MsgError } from '@/utils/message';
import { newUUID } from '@/utils/id';
import i18n from '@/lang';

const open = ref(false);
const loading = ref(false);
const overwrite = ref(true);
const fileList = ref<UploadUserFile[]>([]);
const uploadRef = ref<UploadInstance>();

const emit = defineEmits<{
    uploaded: [payload: { taskID: string; apps: string[] }];
}>();

const beforeUpload: UploadProps['beforeUpload'] = (rawFile) => {
    const isTarGz = rawFile.name.toLowerCase().endsWith('.tar.gz');
    if (!isTarGz) {
        MsgError(i18n.global.t('app.uploadLocalAppPackageFormatError'));
        return false;
    }
    return true;
};

const handleExceed: UploadProps['onExceed'] = (files) => {
    uploadRef.value?.clearFiles();
    const file = files[0] as UploadRawFile;
    file.uid = genFileId();
    uploadRef.value?.handleStart(file);
};

const submit = async () => {
    const targetFile = fileList.value[0]?.raw;
    if (!targetFile) {
        MsgError(i18n.global.t('app.uploadLocalAppPackageSelect'));
        return;
    }
    loading.value = true;
    try {
        const taskID = newUUID();
        const formData = new FormData();
        formData.append('file', targetFile);
        formData.append('taskID', taskID);
        formData.append('overwrite', String(overwrite.value));
        const res = await uploadLocalAppPackage(formData);
        open.value = false;
        fileList.value = [];
        emit('uploaded', { taskID: res.data.taskID || taskID, apps: res.data.apps || [] });
    } finally {
        loading.value = false;
    }
};

const handleClose = () => {
    fileList.value = [];
    overwrite.value = true;
};

const acceptParams = () => {
    fileList.value = [];
    overwrite.value = true;
    open.value = true;
};

defineExpose({
    acceptParams,
});
</script>
