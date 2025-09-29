<template>
    <DrawerPro v-model="open" :header="$t('ssl.upload')" size="large" @close="handleClose">
        <el-form ref="sslForm" label-position="top" :model="ssl" label-width="100px" :rules="rules" v-loading="loading">
            <el-form-item :label="$t('website.importType')" prop="type">
                <el-select v-model="ssl.type">
                    <el-option :label="$t('website.pasteSSL')" :value="'paste'"></el-option>
                    <el-option :label="$t('website.localSSL')" :value="'local'"></el-option>
                    <el-option :label="$t('commons.button.upload') + $t('menu.files')" :value="'upload'"></el-option>
                </el-select>
            </el-form-item>
            <div v-if="ssl.type === 'paste'">
                <el-form-item :label="$t('website.privateKey')" prop="privateKey">
                    <el-input v-model="ssl.privateKey" :rows="6" type="textarea" />
                </el-form-item>
                <el-form-item :label="$t('website.certificate')" prop="certificate">
                    <el-input v-model="ssl.certificate" :rows="6" type="textarea" />
                </el-form-item>
            </div>
            <div v-if="ssl.type === 'local'">
                <el-form-item :label="$t('website.privateKeyPath')" prop="privateKeyPath">
                    <el-input v-model="ssl.privateKeyPath">
                        <template #prepend>
                            <el-button icon="Folder" @click="keyFileRef.acceptParams({ dir: false })" />
                        </template>
                    </el-input>
                </el-form-item>
                <el-form-item :label="$t('website.certificatePath')" prop="certificatePath">
                    <el-input v-model="ssl.certificatePath">
                        <template #prepend>
                            <el-button icon="Folder" @click="certFileRef.acceptParams({ dir: false })" />
                        </template>
                    </el-input>
                </el-form-item>
            </div>
            <div v-if="ssl.type === 'upload'">
                <el-form-item :label="$t('website.privateKey')" prop="privateKeyFile">
                    <el-upload
                        ref="privateKeyUpload"
                        :auto-upload="false"
                        :limit="1"
                        :on-change="handlePrivateKeyChange"
                        :file-list="privateKeyFileList"
                        class="p-w-200"
                    >
                        <template #trigger>
                            <el-button type="primary" icon="Upload">
                                {{ $t('file.selectFile') }}
                            </el-button>
                        </template>
                    </el-upload>
                </el-form-item>
                <el-form-item :label="$t('website.certificate')" prop="certificateFile">
                    <el-upload
                        ref="certificateUpload"
                        :auto-upload="false"
                        :limit="1"
                        :on-change="handleCertificateChange"
                        :file-list="certificateFileList"
                        class="p-w-200"
                    >
                        <template #trigger>
                            <el-button type="primary" icon="Upload">
                                {{ $t('file.selectFile') }}
                            </el-button>
                        </template>
                    </el-upload>
                </el-form-item>
            </div>
            <el-form-item :label="$t('website.remark')" prop="description">
                <el-input v-model="ssl.description"></el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit()" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
    <FileList ref="keyFileRef" @choose="getPrivateKeyPath" />
    <FileList ref="certFileRef" @choose="getCertificatePath" />
</template>

<script lang="ts" setup>
import { uploadSSL, uploadSSLFile } from '@/api/modules/website';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { FormInstance } from 'element-plus';
import FileList from '@/components/file-list/index.vue';
import { ref } from 'vue';
import { MsgSuccess } from '@/utils/message';
import { Website } from '@/api/interface/website';

const open = ref(false);
const keyFileRef = ref();
const certFileRef = ref();
const loading = ref(false);
const sslForm = ref<FormInstance>();
const privateKeyFileList = ref([]);
const certificateFileList = ref([]);
const privateKeyUpload = ref();
const certificateUpload = ref();

const rules = ref({
    privateKey: [Rules.requiredInput],
    certificate: [Rules.requiredInput],
    privateKeyPath: [Rules.requiredInput],
    certificatePath: [Rules.requiredInput],
    type: [Rules.requiredSelect],
    certificateFile: [Rules.requiredInput],
    privateKeyFile: [Rules.requiredInput],
});
const initData = () => ({
    privateKey: '',
    certificate: '',
    privateKeyPath: '',
    certificatePath: '',
    type: 'paste',
    sslID: 0,
    description: '',
    privateKeyFile: null as File | null,
    certificateFile: null as File | null,
});
const ssl = ref(initData());

const em = defineEmits(['close']);
const handleClose = () => {
    resetForm();
    open.value = false;
    em('close', false);
};
const resetForm = () => {
    sslForm.value?.resetFields();
    ssl.value = initData();
    privateKeyFileList.value = [];
    certificateFileList.value = [];
};

const handlePrivateKeyChange = (file: any) => {
    ssl.value.privateKeyFile = file.raw;
    privateKeyFileList.value = [file];
};

const handleCertificateChange = (file: any) => {
    ssl.value.certificateFile = file.raw;
    certificateFileList.value = [file];
};

const acceptParams = (websiteSSL: Website.SSLDTO) => {
    resetForm();
    if (websiteSSL && websiteSSL.id > 0) {
        ssl.value.sslID = websiteSSL.id;
        ssl.value.description = websiteSSL.description;
        ssl.value.privateKeyPath = websiteSSL.privateKeyPath;
        ssl.value.certificatePath = websiteSSL.certPath;
        if (ssl.value.certificatePath != '' && ssl.value.privateKeyPath != '') {
            ssl.value.type = 'local';
        }
    }
    open.value = true;
};

const getPrivateKeyPath = (path: string) => {
    ssl.value.privateKeyPath = path;
};

const getCertificatePath = (path: string) => {
    ssl.value.certificatePath = path;
};

const submit = async () => {
    try {
        await sslForm.value?.validate();
        loading.value = true;
        if (ssl.value.type === 'upload') {
            const formData = new FormData();
            formData.append('type', ssl.value.type);
            formData.append('description', ssl.value.description);
            formData.append('sslID', ssl.value.sslID.toString());

            if (ssl.value.privateKeyFile) {
                formData.append('privateKeyFile', ssl.value.privateKeyFile);
            }
            if (ssl.value.certificateFile) {
                formData.append('certificateFile', ssl.value.certificateFile);
            }
            await uploadSSLFile(formData);
        } else {
            await uploadSSL(ssl.value);
        }
        handleClose();
        MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
    } catch (err) {
    } finally {
        loading.value = false;
    }
};

defineExpose({
    acceptParams,
});
</script>
