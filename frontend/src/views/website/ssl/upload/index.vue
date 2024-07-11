<template>
    <el-drawer
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        v-model="open"
        size="50%"
    >
        <template #header>
            <DrawerHeader :header="$t('ssl.upload')" :back="handleClose" />
        </template>
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-form ref="sslForm" label-position="top" :model="ssl" label-width="100px" :rules="rules">
                    <el-form-item :label="$t('website.importType')" prop="type">
                        <el-select v-model="ssl.type">
                            <el-option :label="$t('website.pasteSSL')" :value="'paste'"></el-option>
                            <el-option :label="$t('website.localSSL')" :value="'local'"></el-option>
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
                                    <FileList @choose="getPrivateKeyPath" :dir="false"></FileList>
                                </template>
                            </el-input>
                        </el-form-item>
                        <el-form-item :label="$t('website.certificatePath')" prop="certificatePath">
                            <el-input v-model="ssl.certificatePath">
                                <template #prepend>
                                    <FileList @choose="getCertificatePath" :dir="false"></FileList>
                                </template>
                            </el-input>
                        </el-form-item>
                    </div>
                    <el-form-item :label="$t('website.remark')" prop="description">
                        <el-input v-model="ssl.description"></el-input>
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(sslForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import DrawerHeader from '@/components/drawer-header/index.vue';
import { UploadSSL } from '@/api/modules/website';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { FormInstance } from 'element-plus';
import { ref } from 'vue';
import { MsgSuccess } from '@/utils/message';
import { Website } from '@/api/interface/website';

const open = ref(false);
const loading = ref(false);
const sslForm = ref<FormInstance>();

const rules = ref({
    privateKey: [Rules.requiredInput],
    certificate: [Rules.requiredInput],
    privateKeyPath: [Rules.requiredInput],
    certificatePath: [Rules.requiredInput],
    type: [Rules.requiredSelect],
});
const initData = () => ({
    privateKey: '',
    certificate: '',
    privateKeyPath: '',
    certificatePath: '',
    type: 'paste',
    sslID: 0,
    description: '',
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
};

const acceptParams = (websiteSSL: Website.SSLDTO) => {
    resetForm();
    if (websiteSSL && websiteSSL.id > 0) {
        ssl.value.sslID = websiteSSL.id;
        ssl.value.description = websiteSSL.description;
    }
    open.value = true;
};

const getPrivateKeyPath = (path: string) => {
    ssl.value.privateKeyPath = path;
};

const getCertificatePath = (path: string) => {
    ssl.value.certificatePath = path;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        UploadSSL(ssl.value)
            .then(() => {
                handleClose();
                MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>
