<template>
    <div>
        <DialogPro v-model="open" class="level-up-pro" @close="handleClose">
            <div style="text-align: center" v-loading="loading">
                <span class="text-3xl font-medium title">
                    {{ isImport ? $t('license.importLicense') : $t('license.levelUpPro') }}
                </span>
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
                    {{ isImport ? $t('commons.button.confirm') : $t('commons.button.power') }}
                </el-button>
                <div class="mt-3 mb-5">
                    <el-button text type="primary" @click="toLxware">{{ $t('license.knowMorePro') }}</el-button>
                </div>
            </div>
        </DialogPro>
    </div>
</template>

<script setup lang="ts">
import i18n from '@/lang';
import { ref } from 'vue';
import { MsgSuccess } from '@/utils/message';
import { uploadLicense } from '@/api/modules/setting';
import DockerProxy from '@/components/docker-proxy/index.vue';
import { GlobalStore } from '@/store';
import { UploadFile, UploadFiles, UploadInstance, UploadProps, UploadRawFile, genFileId } from 'element-plus';
import { getXpackSettingForTheme, loadMasterProductProFromDB, loadProductProFromDB } from '@/utils/xpack';
const globalStore = GlobalStore();

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
    withoutReload.value = params.withoutReload;

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
    if (oldLicense.value) {
        formData.append('oldLicenseName', oldLicense.value);
    }
    if (!isImport.value) {
        formData.append('currentNode', globalStore.currentNode);
        formData.append('withDockerRestart', withDockerRestart.value);
    }
    formData.append('isForce', isForce.value);
    loading.value = true;
    await uploadLicense(oldLicense.value, formData)
        .then(async () => {
            loading.value = false;
            uploadRef.value!.clearFiles();
            uploaderFiles.value = [];
            open.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            if (!withoutReload.value) {
                loadMasterProductProFromDB();
                loadProductProFromDB();
                getXpackSettingForTheme();
                window.location.reload();
            } else {
                em('search');
            }
        })
        .catch(() => {
            loading.value = false;
            uploadRef.value!.clearFiles();
            uploaderFiles.value = [];
        });
};

defineExpose({
    acceptParams,
});
</script>
