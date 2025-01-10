<template>
    <div>
        <el-dialog class="level-up-pro" v-model="open" :close-on-click-modal="false" @close="handleClose">
            <div style="text-align: center" v-loading="loading">
                <span class="text-3xl font-medium title">{{ $t('license.levelUpPro') }}</span>
                <el-row type="flex" justify="center" class="mt-6">
                    <el-col :span="22">
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
                    </el-col>
                </el-row>
                <el-button
                    type="primary"
                    class="mt-3 w-52"
                    :disabled="loading || uploaderFiles.length == 0"
                    plain
                    @click="submit"
                >
                    {{ $t('license.power') }}
                </el-button>
                <div class="mt-3 mb-5">
                    <el-button text type="primary" @click="toLxware">{{ $t('license.knowMorePro') }}</el-button>
                </div>
            </div>
        </el-dialog>
    </div>
</template>

<script setup lang="ts">
import i18n from '@/lang';
import { ref } from 'vue';
import { MsgSuccess } from '@/utils/message';
import { UploadFileData } from '@/api/modules/setting';
import { GlobalStore } from '@/store';
import { UploadFile, UploadFiles, UploadInstance, UploadProps, UploadRawFile, genFileId } from 'element-plus';
import { useTheme } from '@/hooks/use-theme';
import { getXpackSetting, initFavicon } from '@/utils/xpack';
const globalStore = GlobalStore();

const { switchTheme } = useTheme();
const loading = ref(false);
const open = ref(false);
const uploadRef = ref<UploadInstance>();
const uploaderFiles = ref<UploadFiles>([]);

const handleClose = () => {
    open.value = false;
    uploadRef.value!.clearFiles();
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

const toLxware = () => {
    if (!globalStore.isIntl) {
        window.open('https://www.lxware.cn/1panel' + '', '_blank', 'noopener,noreferrer');
    } else {
        window.open('https://1panel.hk/pricing' + '', '_blank', 'noopener,noreferrer');
    }
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
            globalStore.isProductPro = true;
            const xpackRes = await getXpackSetting();
            if (xpackRes) {
                globalStore.themeConfig.theme = xpackRes.data.theme;
                globalStore.themeConfig.themeColor = xpackRes.data.themeColor;
            }
            loading.value = false;
            switchTheme();
            initFavicon();
            uploadRef.value!.clearFiles();
            uploaderFiles.value = [];
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

defineExpose({
    acceptParams,
});
</script>
