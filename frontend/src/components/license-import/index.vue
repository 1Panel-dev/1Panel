<template>
    <div>
        <DialogPro v-model="open" class="level-up-pro" @close="handleClose">
            <div style="text-align: center" v-loading="loading">
                <span class="text-3xl font-medium title">
                    {{ $t('license.importLicense') }}
                </span>
                <el-row type="flex" justify="center" class="mt-4">
                    <el-col :span="22">
                        <div class="license-security-tip">
                            <div class="license-security-title">{{ $t('license.securityWarningTitle') }}</div>
                            <div>{{ $t('license.securityWarningContent') }}</div>
                        </div>
                    </el-col>
                </el-row>
                <el-row type="flex" justify="center" class="mt-4">
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

                        <span v-if="!isImport" class="input-help">{{ $t('xpack.node.syncWithMaster') }}</span>
                        <DockerProxy
                            v-if="!isImport"
                            class="w-full mt-2"
                            v-model:with-docker-restart="withDockerRestart"
                            syncList="SyncSystemProxy"
                        />
                    </el-col>
                </el-row>

                <div v-if="oldLicense">
                    <el-checkbox v-model="isForce">{{ $t('license.updateForce') }}</el-checkbox>
                </div>
                <el-button
                    type="primary"
                    class="mt-3"
                    size="large"
                    :disabled="loading || uploaderFiles.length == 0"
                    plain
                    @click="submit"
                >
                    {{ $t('commons.button.power') }}
                </el-button>
                <div class="mt-3 mb-5">
                    <el-button text type="primary" @click="toEdition">{{ $t('license.knowMorePro') }}</el-button>
                </div>
            </div>
        </DialogPro>
    </div>
</template>

<script setup lang="ts">
import i18n from '@/lang';
import { ref } from 'vue';
import { MsgSuccess } from '@/utils/message';
import { uploadLicense, uploadEnterpriseLicense } from '@/api/modules/setting';
import DockerProxy from '@/components/docker-proxy/index.vue';
import { UploadFile, UploadFiles, UploadInstance, UploadProps, UploadRawFile, genFileId } from 'element-plus';
import { getXpackSettingForTheme, loadMasterProductProFromDB, loadProductProFromDB } from '@/utils/xpack';
import { useGlobalStore } from '@/composables/useGlobalStore';
const { isIntl, isEnterprise, currentNode, isProductPro, isMasterProductPro, isEnterpriseLicensed } = useGlobalStore();

const em = defineEmits(['search']);

const loading = ref(false);
const open = ref(false);
const uploadRef = ref<UploadInstance>();
const uploaderFiles = ref<UploadFiles>([]);
const isImport = ref();
const isForce = ref();
const withoutReload = ref();

const withDockerRestart = ref();

const oldLicense = ref();
interface DialogProps {
    oldLicense: string;
    isImport: boolean;
    withoutReload: boolean;
}

const acceptParams = (params: DialogProps) => {
    oldLicense.value = params?.oldLicense || '';
    uploaderFiles.value = [];
    uploadRef.value?.clearFiles();
    isImport.value = params?.isImport;
    withoutReload.value = params?.withoutReload || false;

    open.value = true;
};

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

const toEdition = () => {
    if (!isIntl.value) {
        window.open('https://1panel.cn/versions.html' + '', '_blank', 'noopener,noreferrer');
    } else {
        window.open('https://1panel.pro/pricing' + '', '_blank', 'noopener,noreferrer');
    }
};

const submit = async () => {
    if (uploaderFiles.value.length !== 1) {
        return;
    }
    const file = uploaderFiles.value[0];
    const formData = new FormData();
    formData.append('file', file.raw);
    if (isEnterprise.value) {
        loading.value = true;
        await uploadEnterpriseLicense(formData)
            .then(async () => {
                await handleAfterSubmit();
            })
            .catch(() => {
                loading.value = false;
                uploadRef.value!.clearFiles();
                uploaderFiles.value = [];
            });
        return;
    }
    if (oldLicense.value) {
        formData.append('oldLicenseName', oldLicense.value);
    }
    if (!isImport.value) {
        formData.append('currentNode', currentNode.value);
        formData.append('withDockerRestart', withDockerRestart.value);
    }
    formData.append('isForce', isForce.value);
    loading.value = true;
    await uploadLicense(oldLicense.value, formData)
        .then(async () => {
            await handleAfterSubmit();
        })
        .catch(() => {
            loading.value = false;
            uploadRef.value!.clearFiles();
            uploaderFiles.value = [];
        });
};

const handleAfterSubmit = async () => {
    loading.value = false;
    uploadRef.value!.clearFiles();
    uploaderFiles.value = [];
    open.value = false;
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    if (!isImport.value) {
        if (!isEnterprise.value) isProductPro.value = true;
        isMasterProductPro.value = true;
    } else {
        isEnterpriseLicensed.value = true;
    }
    if (!withoutReload.value) {
        await loadMasterProductProFromDB();
        await loadProductProFromDB();
        await getXpackSettingForTheme();
        window.location.reload();
    } else {
        em('search');
    }
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.license-security-tip {
    padding: 8px 10px;
    text-align: left;
    font-size: 12px;
    line-height: 20px;
    color: var(--el-text-color-regular);
    background: var(--el-color-warning-light-9);
    border: 1px solid var(--el-color-warning-light-7);
    border-left: 3px solid var(--el-color-warning);
    border-radius: 4px;
}

.license-security-title {
    margin-bottom: 2px;
    font-weight: 600;
    color: var(--el-color-warning-dark-2);
}
</style>
