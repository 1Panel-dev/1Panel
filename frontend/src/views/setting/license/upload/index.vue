<template>
    <el-drawer
        v-model="open"
        :before-close="handleClose"
        size="30%"
        :destroy-on-close="true"
        :close-on-click-modal="false"
    >
        <template #header>
            <DrawerHeader :header="$t('license.importLicense')" :back="handleClose" />
        </template>
        <div v-loading="loading">
            <el-upload
                action="#"
                :auto-upload="false"
                ref="uploadRef"
                class="upload-demo"
                drag
                :limit="1"
                :on-change="fileOnChange"
                :on-exceed="handleExceed"
                v-model:file-list="uploaderFiles"
            >
                <el-icon class="el-icon--upload"><upload-filled /></el-icon>
                <div class="el-upload__text">
                    {{ $t('license.importHelper') }}
                </div>
            </el-upload>
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
import { ref } from 'vue';
import { UploadFile, UploadFiles, UploadInstance, UploadProps, UploadRawFile, genFileId } from 'element-plus';
import { UploadFileData } from '@/api/modules/setting';
import i18n from '@/lang';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);
const open = ref(false);

const em = defineEmits(['search']);

const uploadRef = ref<UploadInstance>();
const uploaderFiles = ref<UploadFiles>([]);

const handleClose = () => {
    open.value = false;
    uploadRef.value!.clearFiles();
    em('search');
};

const fileOnChange = (_uploadFile: UploadFile, uploadFiles: UploadFiles) => {
    uploaderFiles.value = uploadFiles;
};

const handleExceed: UploadProps['onExceed'] = (files) => {
    uploadRef.value!.clearFiles();
    const file = files[0] as UploadRawFile;
    file.uid = genFileId();
    uploadRef.value!.handleStart(file);
};

const submit = async () => {
    if (uploaderFiles.value.length !== 1) {
        return;
    }
    const file = uploaderFiles.value[0];
    const formData = new FormData();
    formData.append('file', file.raw);
    loading.value = true;
    await UploadFileData(formData)
        .then(async () => {
            loading.value = false;
            uploadRef.value!.clearFiles();
            uploaderFiles.value = [];
            em('search');
            open.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            window.location.reload();
        })
        .catch(() => {
            loading.value = false;
            uploadRef.value!.clearFiles();
            uploaderFiles.value = [];
        });
};

const acceptParams = () => {
    uploaderFiles.value = [];
    uploadRef.value?.clearFiles();
    open.value = true;
};

defineExpose({ acceptParams });
</script>
