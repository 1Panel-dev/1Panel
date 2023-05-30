<template>
    <el-row :gutter="20" v-loading="loading">
        <el-col :xs="24" :sm="18" :md="14" :lg="14" :xl="14">
            <el-form ref="httpsForm" label-position="right" label-width="auto" :model="form" :rules="rules">
                <el-form-item prop="enable" :label="$t('website.enableHTTPS')">
                    <el-switch v-model="form.enable" @change="changeEnable"></el-switch>
                </el-form-item>
                <div v-if="form.enable">
                    <el-divider content-position="left">{{ $t('website.SSLConfig') }}</el-divider>
                    <el-form-item :label="$t('website.HTTPConfig')" prop="httpConfig">
                        <el-select v-model="form.httpConfig" style="width: 240px">
                            <el-option :label="$t('website.HTTPToHTTPS')" :value="'HTTPToHTTPS'"></el-option>
                            <el-option :label="$t('website.HTTPAlso')" :value="'HTTPAlso'"></el-option>
                            <el-option :label="$t('website.HTTPSOnly')" :value="'HTTPSOnly'"></el-option>
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('website.sslConfig')" prop="type">
                        <el-select v-model="form.type" @change="changeType(form.type)">
                            <el-option :label="$t('website.oldSSL')" :value="'existed'"></el-option>
                            <el-option :label="$t('website.manualSSL')" :value="'manual'"></el-option>
                        </el-select>
                    </el-form-item>
                    <el-form-item
                        :label="$t('website.ssl')"
                        prop="websiteSSLId"
                        v-if="form.type === 'existed'"
                        :hide-required-asterisk="true"
                    >
                        <el-select
                            v-model="form.websiteSSLId"
                            :placeholder="$t('website.selectSSL')"
                            @change="changeSSl(form.websiteSSLId)"
                        >
                            <el-option
                                v-for="(ssl, index) in ssls"
                                :key="index"
                                :label="ssl.primaryDomain"
                                :value="ssl.id"
                            ></el-option>
                        </el-select>
                    </el-form-item>
                    <div v-if="form.type === 'manual'">
                        <el-form-item :label="$t('website.privateKey')" prop="privateKey">
                            <el-input v-model="form.privateKey" :rows="6" type="textarea" />
                        </el-form-item>
                        <el-form-item :label="$t('website.certificate')" prop="certificate">
                            <el-input v-model="form.certificate" :rows="6" type="textarea" />
                        </el-form-item>
                    </div>
                    <el-form-item :label="' '" v-if="websiteSSL && websiteSSL.id > 0">
                        <el-descriptions :column="5" border direction="vertical">
                            <el-descriptions-item :label="$t('website.primaryDomain')">
                                {{ websiteSSL.primaryDomain }}
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('website.otherDomains')">
                                {{ websiteSSL.domains }}
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('ssl.provider')">
                                {{ getProvider(websiteSSL.provider) }}
                            </el-descriptions-item>
                            <el-descriptions-item
                                :label="$t('ssl.acmeAccount')"
                                v-if="websiteSSL.acmeAccount && websiteSSL.provider !== 'manual'"
                            >
                                {{ websiteSSL.acmeAccount.email }}
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('website.expireDate')">
                                {{ dateFormatSimple(websiteSSL.expireDate) }}
                            </el-descriptions-item>
                        </el-descriptions>
                    </el-form-item>
                    <el-divider content-position="left">{{ $t('website.SSLProConfig') }}</el-divider>
                    <el-form-item :label="$t('website.supportProtocol')" prop="SSLProtocol">
                        <el-checkbox-group v-model="form.SSLProtocol">
                            <el-checkbox :label="'TLSv1.3'">{{ 'TLS 1.3' }}</el-checkbox>
                            <el-checkbox :label="'TLSv1.2'">{{ 'TLS 1.2' }}</el-checkbox>
                            <el-checkbox :label="'TLSv1.1'">{{ 'TLS 1.1' }}</el-checkbox>
                            <el-checkbox :label="'TLSv1'">{{ 'TLS 1.0' }}</el-checkbox>
                            <br />
                            <el-checkbox :label="'SSLv3'">
                                {{ 'SSL V3' + $t('website.notSecurity') }}
                            </el-checkbox>
                            <el-checkbox :label="'SSLv2'">
                                {{ 'SSL V2' + $t('website.notSecurity') }}
                            </el-checkbox>
                        </el-checkbox-group>
                    </el-form-item>
                    <el-form-item prop="algorithm" :label="$t('website.encryptionAlgorithm')">
                        <el-input
                            type="textarea"
                            :autosize="{ minRows: 2, maxRows: 6 }"
                            v-model.trim="form.algorithm"
                        ></el-input>
                    </el-form-item>
                    <el-form-item>
                        <el-button type="primary" @click="submit(httpsForm)">
                            {{ $t('commons.button.save') }}
                        </el-button>
                    </el-form-item>
                </div>
                <div v-else>
                    <el-alert :closable="false">
                        <template #default>
                            <span style="white-space: pre-line">{{ $t('website.SSLHelper') }}</span>
                        </template>
                    </el-alert>
                </div>
            </el-form>
        </el-col>
    </el-row>
</template>
<script lang="ts" setup>
import { Website } from '@/api/interface/website';
import { GetHTTPSConfig, ListSSL, UpdateHTTPSConfig } from '@/api/modules/website';
import { ElMessageBox, FormInstance } from 'element-plus';
import { computed, onMounted, reactive, ref } from 'vue';
import i18n from '@/lang';
import { Rules } from '@/global/form-rules';
import { dateFormatSimple, getProvider } from '@/utils/util';
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
    enable: false,
    websiteId: id.value,
    websiteSSLId: undefined,
    type: 'existed',
    privateKey: '',
    certificate: '',
    httpConfig: 'HTTPToHTTPS',
    algorithm:
        'EECDH+CHACHA20:EECDH+CHACHA20-draft:EECDH+AES128:RSA+AES128:EECDH+AES256:RSA+AES256:EECDH+3DES:RSA+3DES:!MD5',
    SSLProtocol: ['TLSv1.3', 'TLSv1.2', 'TLSv1.1', 'TLSv1'],
});
const loading = ref(false);
const ssls = ref();
const websiteSSL = ref();
const rules = ref({
    type: [Rules.requiredSelect],
    privateKey: [Rules.requiredInput],
    certificate: [Rules.requiredInput],
    websiteSSLId: [Rules.requiredSelect],
    httpConfig: [Rules.requiredSelect],
    SSLProtocol: [Rules.requiredSelect],
    algorithm: [Rules.requiredInput],
});
const resData = ref();

const listSSL = () => {
    ListSSL({}).then((res) => {
        ssls.value = res.data;
        changeSSl(form.websiteSSLId);
    });
};

const changeSSl = (sslid: number) => {
    const res = ssls.value.filter((element: Website.SSL) => {
        return element.id == sslid;
    });
    websiteSSL.value = res[0];
};

const changeType = (type: string) => {
    if (type != 'existed') {
        websiteSSL.value = {};
        form.websiteSSLId = undefined;
    }
};

const get = () => {
    GetHTTPSConfig(id.value).then((res) => {
        if (res.data) {
            resData.value = res.data;
            form.enable = res.data.enable;
            if (res.data.httpConfig != '') {
                form.httpConfig = res.data.httpConfig;
            }
            if (res.data.SSLProtocol && res.data.SSLProtocol.length > 0) {
                form.SSLProtocol = res.data.SSLProtocol;
            }
            if (res.data.algorithm != '') {
                form.algorithm = res.data.algorithm;
            }
            if (res.data.SSL && res.data.SSL.id > 0) {
                form.websiteSSLId = res.data.SSL.id;
                websiteSSL.value = res.data.SSL;
            }
        }
        listSSL();
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
        UpdateHTTPSConfig(form)
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
        listSSL();
    }
    if (resData.value.enable && !enable) {
        ElMessageBox.confirm(i18n.global.t('website.disbaleHTTTPSHelper'), i18n.global.t('website.disbaleHTTTPS'), {
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
                    UpdateHTTPSConfig(form).then(() => {
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
