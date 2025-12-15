<template>
    <DrawerPro v-model="open" :header="$t('commons.button.set') + $t('website.ssl')" size="50%" @close="handleClose">
        <el-form ref="websiteForm" label-position="top" :model="form" :rules="rules" v-loading="loading">
            <HttpsConfig v-model="form" :website-ssl="websiteSSL" @ssl-change="handleSSLChange" />
        </el-form>
        <template #footer>
            <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" @click="submit()" :disabled="loading">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import HttpsConfig from '@/views/website/website/components/https/index.vue';

import { Website } from '@/api/interface/website';
import { batchSetHttps } from '@/api/modules/website';
import { FormInstance } from 'element-plus';
import { Rules } from '@/global/form-rules';

const open = ref(false);
const loading = ref(false);
const websiteForm = ref<FormInstance>();
const websiteSSL = ref();
const form = ref({
    ids: [] as number[],
    acmeAccountID: 0,
    enable: false,
    websiteSSLId: undefined,
    type: 'existed',
    importType: 'paste',
    privateKey: '',
    certificate: '',
    privateKeyPath: '',
    certificatePath: '',
    httpConfig: 'HTTPToHTTPS',
    hsts: true,
    hstsIncludeSubDomains: false,
    algorithm:
        'ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-SHA384:ECDHE-RSA-AES128-SHA256:!aNULL:!eNULL:!EXPORT:!DSS:!DES:!RC4:!3DES:!MD5:!PSK:!KRB5:!SRP:!CAMELLIA:!SEED',
    SSLProtocol: ['TLSv1.3', 'TLSv1.2'],
    httpsPort: '443',
    http3: false,
    taskID: '',
});
const rules = ref({
    type: [Rules.requiredSelect],
    privateKey: [Rules.requiredInput],
    certificate: [Rules.requiredInput],
    privateKeyPath: [Rules.requiredInput],
    certificatePath: [Rules.requiredInput],
    websiteSSLId: [Rules.requiredSelect],
    httpConfig: [Rules.requiredSelect],
    SSLProtocol: [Rules.requiredSelect],
    algorithm: [Rules.requiredInput],
    acmeAccountID: [Rules.requiredInput],
});
const emit = defineEmits(['openTask']);

const handleClose = () => {
    open.value = false;
};

const acceptParams = async (ids: [], taskID: string) => {
    form.value.ids = ids;
    form.value.taskID = taskID;
    open.value = true;
};

const handleSSLChange = (ssl: Website.SSL) => {
    websiteSSL.value = ssl;
};

const submit = async () => {
    loading.value = true;
    const valid = await websiteForm.value.validate();
    if (!valid) {
        loading.value = false;
        return;
    }
    batchSetHttps(form.value)
        .then(() => {
            open.value = false;
            emit('openTask', form.value.taskID);
        })
        .finally(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>
