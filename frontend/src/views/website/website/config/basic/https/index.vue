<template>
    <el-row :gutter="20" v-loading="loading">
        <el-col :xs="24" :sm="18" :md="18" :lg="18" :xl="14">
            <el-form
                class="moblie-form"
                ref="httpsForm"
                label-position="right"
                label-width="auto"
                :model="form"
                :rules="rules"
            >
                <el-form-item prop="enable" :label="$t('website.enableHTTPS')">
                    <el-switch v-model="form.enable" @change="changeEnable"></el-switch>
                </el-form-item>
                <div v-if="form.enable">
                    <el-text type="warning" class="!ml-2">{{ $t('website.ipWebsiteWarn') }}</el-text>
                    <el-divider content-position="left">{{ $t('website.SSLConfig') }}</el-divider>
                    <el-form-item :label="$t('website.HTTPConfig')" prop="httpConfig">
                        <el-select v-model="form.httpConfig" style="width: 240px">
                            <el-option :label="$t('website.HTTPToHTTPS')" :value="'HTTPToHTTPS'"></el-option>
                            <el-option :label="$t('website.HTTPAlso')" :value="'HTTPAlso'"></el-option>
                            <el-option :label="$t('website.HTTPSOnly')" :value="'HTTPSOnly'"></el-option>
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="'HSTS'" prop="hsts">
                        <el-checkbox v-model="form.hsts">{{ $t('commons.button.enable') }}</el-checkbox>
                        <span class="input-help">{{ $t('website.hstsHelper') }}</span>
                    </el-form-item>
                    <el-form-item :label="$t('website.sslConfig')" prop="type">
                        <el-select v-model="form.type" @change="changeType(form.type)">
                            <el-option :label="$t('website.oldSSL')" :value="'existed'"></el-option>
                            <el-option :label="$t('website.manualSSL')" :value="'manual'"></el-option>
                        </el-select>
                    </el-form-item>
                    <div v-if="form.type === 'existed'">
                        <el-form-item :label="$t('website.acmeAccountManage')" prop="acmeAccountID">
                            <el-select
                                v-model="form.acmeAccountID"
                                :placeholder="$t('website.selectAcme')"
                                @change="listSSL"
                            >
                                <el-option :key="0" :label="$t('website.imported')" :value="0"></el-option>
                                <el-option
                                    v-for="(acme, index) in acmeAccounts"
                                    :key="index"
                                    :label="acme.email"
                                    :value="acme.id"
                                >
                                    <span>
                                        {{ acme.email }}
                                        <el-tag class="ml-5">{{ getAccountName(acme.type) }}</el-tag>
                                    </span>
                                </el-option>
                            </el-select>
                        </el-form-item>
                        <el-form-item :label="$t('website.ssl')" prop="websiteSSLId" :hide-required-asterisk="true">
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
                    </div>
                    <div v-if="form.type === 'manual'">
                        <el-form-item :label="$t('website.importType')" prop="type">
                            <el-select v-model="form.importType">
                                <el-option :label="$t('website.pasteSSL')" :value="'paste'"></el-option>
                                <el-option :label="$t('website.localSSL')" :value="'local'"></el-option>
                            </el-select>
                        </el-form-item>
                        <div v-if="form.importType === 'paste'">
                            <el-form-item :label="$t('website.privateKey')" prop="privateKey">
                                <el-input v-model="form.privateKey" :rows="6" type="textarea" />
                            </el-form-item>
                            <el-form-item :label="$t('website.certificate')" prop="certificate">
                                <el-input v-model="form.certificate" :rows="6" type="textarea" />
                            </el-form-item>
                        </div>
                        <div v-if="form.importType === 'local'">
                            <el-form-item :label="$t('website.privateKeyPath')" prop="privateKeyPath">
                                <el-input v-model="form.privateKeyPath">
                                    <template #prepend>
                                        <FileList @choose="getPrivateKeyPath" :dir="false"></FileList>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item :label="$t('website.certificatePath')" prop="certificatePath">
                                <el-input v-model="form.certificatePath">
                                    <template #prepend>
                                        <FileList @choose="getCertificatePath" :dir="false"></FileList>
                                    </template>
                                </el-input>
                            </el-form-item>
                        </div>
                    </div>
                    <el-form-item :label="' '" v-if="websiteSSL && websiteSSL.id > 0">
                        <el-descriptions :column="6" border direction="vertical">
                            <el-descriptions-item :label="$t('website.primaryDomain')">
                                {{ websiteSSL.primaryDomain }}
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('website.otherDomains')">
                                {{ websiteSSL.domains }}
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('website.brand')">
                                {{ websiteSSL.organization }}
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
                            <el-descriptions-item :label="$t('website.remark')">
                                {{ websiteSSL.description }}
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
                        <el-input type="textarea" :rows="3" v-model.trim="form.algorithm"></el-input>
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
import { GetHTTPSConfig, ListSSL, SearchAcmeAccount, UpdateHTTPSConfig } from '@/api/modules/website';
import { ElMessageBox, FormInstance } from 'element-plus';
import { computed, onMounted, reactive, ref } from 'vue';
import i18n from '@/lang';
import { Rules } from '@/global/form-rules';
import { dateFormatSimple, getProvider, getAccountName } from '@/utils/util';
import { MsgSuccess } from '@/utils/message';
import FileList from '@/components/file-list/index.vue';

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
    algorithm:
        'ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-SHA384:ECDHE-RSA-AES128-SHA256:!aNULL:!eNULL:!EXPORT:!DSS:!DES:!RC4:!3DES:!MD5:!PSK:!KRB5:!SRP:!CAMELLIA:!SEED',
    SSLProtocol: ['TLSv1.3', 'TLSv1.2', 'TLSv1.1', 'TLSv1'],
});
const loading = ref(false);
const ssls = ref();
const acmeAccounts = ref();
const websiteSSL = ref();
const rules = ref({
    hsts: [Rules.requiredInput],
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
const sslReq = reactive({
    acmeAccountID: '',
});

const getPrivateKeyPath = (path: string) => {
    form.privateKeyPath = path;
};

const getCertificatePath = (path: string) => {
    form.certificatePath = path;
};
const listSSL = () => {
    sslReq.acmeAccountID = String(form.acmeAccountID);
    ListSSL(sslReq).then((res) => {
        ssls.value = res.data || [];
        if (ssls.value.length > 0) {
            let exist = false;
            for (const ssl of ssls.value) {
                if (ssl.id === form.websiteSSLId) {
                    exist = true;
                    break;
                }
            }
            if (!exist) {
                form.websiteSSLId = ssls.value[0].id;
            }
            changeSSl(form.websiteSSLId);
        } else {
            websiteSSL.value = {};
            form.websiteSSLId = undefined;
        }
    });
};

const listAcmeAccount = () => {
    SearchAcmeAccount({ page: 1, pageSize: 100 }).then((res) => {
        acmeAccounts.value = res.data.items || [];
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
            form.type = 'existed';
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
                form.acmeAccountID = res.data.SSL.acmeAccountId;
            }
            form.hsts = res.data.hsts;
        }
        listSSL();
        listAcmeAccount();
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
        form.hsts = true;
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
