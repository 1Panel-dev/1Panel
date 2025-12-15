<template>
    <el-row :gutter="20" v-loading="loading">
        <el-col :xs="24" :sm="18" :md="18" :lg="14" :xl="14">
            <el-alert :closable="false">
                <template #default>
                    <span class="whitespace-pre-line">{{ $t('website.SSLHelper') }}</span>
                </template>
            </el-alert>
            <el-form
                class="moblie-form"
                ref="httpsForm"
                label-position="right"
                label-width="150px"
                :model="form"
                :rules="rules"
            >
                <el-form-item prop="enable" :label="$t('website.enableHTTPS')">
                    <el-switch v-model="form.enable" @change="changeEnable"></el-switch>
                </el-form-item>

                <el-collapse-transition>
                    <div v-if="form.enable">
                        <el-form-item :label="'HTTPS ' + $t('commons.table.port')" prop="httpsPort">
                            <el-text>{{ form.httpsPort }}</el-text>
                        </el-form-item>

                        <el-text type="warning" class="!ml-2">{{ $t('website.ipWebsiteWarn') }}</el-text>

                        <el-divider content-position="left">{{ $t('website.SSLConfig') }}</el-divider>
                        <HttpsConfig v-model="form" :website-ssl="websiteSSL" @ssl-change="handleSSLChange" />
                        <el-form-item>
                            <el-button type="primary" @click="submit(httpsForm)">
                                {{ $t('commons.button.save') }}
                            </el-button>
                        </el-form-item>
                    </div>
                </el-collapse-transition>
            </el-form>
        </el-col>
    </el-row>
</template>
<script lang="ts" setup>
import HttpsConfig from '@/views/website/website/components/https/index.vue';

import { Website } from '@/api/interface/website';
import { getHTTPSConfig, updateHTTPSConfig } from '@/api/modules/website';
import { ElMessageBox, FormInstance } from 'element-plus';
import { computed, onMounted, reactive, ref } from 'vue';
import i18n from '@/lang';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});
const id = computed(() => {
    return props.id;
});
const httpsForm = ref<FormInstance>();
const form = reactive({
    acmeAccountID: 0,
    enable: false,
    websiteId: id.value,
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
});
const loading = ref(false);
const websiteSSL = ref();
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
const resData = ref();

const handleSSLChange = (ssl: Website.SSL) => {
    websiteSSL.value = ssl;
};

const get = () => {
    getHTTPSConfig(id.value).then((res) => {
        if (res.data) {
            form.type = 'existed';
            const data = res.data;
            resData.value = data;
            form.enable = data.enable;
            if (data.httpConfig != '') {
                form.httpConfig = data.httpConfig;
            }
            if (data.SSLProtocol && data.SSLProtocol.length > 0) {
                form.SSLProtocol = data.SSLProtocol;
            }
            if (data.algorithm != '') {
                form.algorithm = data.algorithm;
            }
            if (data.SSL && data.SSL.id > 0) {
                form.websiteSSLId = data.SSL.id;
                websiteSSL.value = data.SSL;
                form.acmeAccountID = data.SSL.acmeAccountId;
            }
            form.hsts = data.hsts;
            form.hstsIncludeSubDomains = data.hstsIncludeSubDomains || false;
            form.http3 = data.http3;
            form.httpsPort = data.httpsPort;
        }
    });
};
const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        form.websiteId = id.value;
        updateHTTPSConfig(form)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
                get();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const changeEnable = (enable: boolean) => {
    if (enable) {
        form.hsts = true;
    } else {
        form.hstsIncludeSubDomains = false;
    }
    if (resData.value.enable && !enable) {
        ElMessageBox.confirm(i18n.global.t('website.disableHTTPSHelper'), i18n.global.t('website.disableHTTPS'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'error',
            closeOnClickModal: false,
            beforeClose: async (action, instance, done) => {
                if (action !== 'confirm') {
                    form.enable = true;
                    done();
                } else {
                    instance.confirmButtonLoading = true;
                    form.enable = false;
                    form.websiteId = id.value;
                    updateHTTPSConfig(form).then(() => {
                        done();
                        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                        get();
                    });
                }
            },
        }).then(() => {});
    }
};

onMounted(() => {
    get();
});
</script>
<style lang="scss">
.el-collapse,
.el-collapse-item__wrap {
    border: none;
}
.el-collapse-item__header {
    border: none;
}
</style>
