<template>
    <el-drawer
        v-model="open"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="50%"
    >
        <template #header>
            <DrawerHeader :header="$t('aitool.proxy')" :back="handleClose" />
        </template>
        <div v-loading="loading">
            <el-form ref="formRef" label-position="top" @submit.prevent :model="req" :rules="rules">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-alert class="common-prompt" :closable="false" type="warning">
                            <template #default>
                                <ul>
                                    <li>{{ $t('aitool.proxyHelper1') }}</li>
                                    <li>{{ $t('aitool.proxyHelper2') }}</li>
                                    <li>{{ $t('aitool.proxyHelper3') }}</li>
                                </ul>
                            </template>
                        </el-alert>
                        <el-form-item :label="$t('website.domain')" prop="domain">
                            <el-input v-model.trim="req.domain" :disabled="operate === 'update'" />
                            <span class="input-help">
                                {{ $t('aitool.proxyHelper4') }}
                            </span>
                            <span class="input-help">
                                {{ $t('aitool.proxyHelper6') }}
                                <el-link
                                    class="pageRoute"
                                    icon="Position"
                                    @click="toWebsite(req.websiteID)"
                                    type="primary"
                                >
                                    {{ $t('firewall.quickJump') }}
                                </el-link>
                            </span>
                        </el-form-item>
                        <el-form-item :label="$t('xpack.waf.whiteList') + ' IP'" prop="ipList">
                            <el-input
                                :rows="3"
                                type="textarea"
                                clearable
                                v-model="req.ipList"
                                :placeholder="$t('xpack.waf.ipGroupHelper')"
                            />
                            <span class="input-help">
                                {{ $t('aitool.whiteListHelper') }}
                            </span>
                        </el-form-item>
                        <el-form-item>
                            <el-checkbox v-model="req.enableSSL" @change="changeSSL">
                                {{ $t('website.enable') + ' ' + 'HTTPS' }}
                            </el-checkbox>
                        </el-form-item>
                        <el-form-item
                            :label="$t('website.acmeAccountManage')"
                            prop="acmeAccountID"
                            v-if="req.enableSSL"
                        >
                            <el-select
                                v-model="req.acmeAccountID"
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
                        <el-form-item :label="$t('website.ssl')" prop="sslID" v-if="req.enableSSL">
                            <el-select
                                v-model="req.sslID"
                                :placeholder="$t('website.selectSSL')"
                                @change="changeSSl(req.sslID)"
                            >
                                <el-option
                                    v-for="(ssl, index) in ssls"
                                    :key="index"
                                    :label="ssl.primaryDomain"
                                    :value="ssl.id"
                                ></el-option>
                            </el-select>
                        </el-form-item>
                        <el-alert :closable="false">
                            {{ $t('aitool.proxyHelper5') }}
                            <el-link class="pageRoute" icon="Position" @click="toInstalled()" type="primary">
                                {{ $t('firewall.quickJump') }}
                            </el-link>
                        </el-alert>
                    </el-col>
                </el-row>
            </el-form>
        </div>
        <template #footer>
            <el-button @click="handleClose">
                {{ $t('commons.button.cancel') }}
            </el-button>
            <el-button type="primary" @click="onSubmit(formRef)">
                {{ $t('commons.button.add') }}
            </el-button>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { Website } from '@/api/interface/website';
import { ListSSL, SearchAcmeAccount } from '@/api/modules/website';
import { Rules } from '@/global/form-rules';
import { FormInstance, FormRules } from 'element-plus';
import { reactive, ref } from 'vue';
import { getAccountName } from '@/utils/util';
import { bindDomain, getBindDomain, updateBindDomain } from '@/api/modules/ai';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

const open = ref(false);
const operate = ref('create');
const loading = ref(false);
const ssls = ref([]);
const websiteSSL = ref<Website.SSL>();
const acmeAccounts = ref();
const formRef = ref();
const req = ref({
    domain: '',
    sslID: undefined,
    ipList: '',
    acmeAccountID: 0,
    enableSSL: false,
    allowIPs: [],
    appInstallID: 0,
    websiteID: 0,
});
const rules = reactive<FormRules>({
    domain: [Rules.domainWithPort],
    sslID: [Rules.requiredSelectBusiness],
});
const emit = defineEmits(['search']);

const handleClose = () => {
    emit('search');
    open.value = false;
};

const acceptParams = (installID: number) => {
    req.value.appInstallID = installID;
    search(installID);
    open.value = true;
};

const changeSSl = (sslid: number) => {
    const res = ssls.value.filter((element: Website.SSL) => {
        return element.id == sslid;
    });
    websiteSSL.value = res[0];
};

const changeSSL = () => {
    if (!req.value.enableSSL) {
        req.value.sslID = undefined;
    } else {
        listAcmeAccount();
    }
};

const listSSL = () => {
    const sslReq = {
        acmeAccountID: String(req.value.acmeAccountID),
    };
    ListSSL(sslReq).then((res) => {
        ssls.value = res.data || [];
        if (ssls.value.length > 0) {
            let exist = false;
            for (const ssl of ssls.value) {
                if (ssl.id === req.value.sslID) {
                    exist = true;
                    break;
                }
            }
            if (!exist) {
                req.value.sslID = ssls.value[0].id;
            }
            changeSSl(req.value.sslID);
        } else {
            req.value.sslID = undefined;
        }
    });
};

const listAcmeAccount = () => {
    SearchAcmeAccount({ page: 1, pageSize: 100 }).then((res) => {
        acmeAccounts.value = res.data.items || [];
        listSSL();
    });
};

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (operate.value === 'update') {
            await updateBindDomain(req.value);
        } else {
            await bindDomain(req.value);
        }
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        handleClose();
    });
};

const search = async (appInstallID: number) => {
    try {
        const res = await getBindDomain({
            appInstallID: appInstallID,
        });
        if (res.data.websiteID > 0) {
            operate.value = 'update';
            req.value.domain = res.data.domain;
            req.value.websiteID = res.data.websiteID;
            if (res.data.allowIPs && res.data.allowIPs.length > 0) {
                req.value.ipList = res.data.allowIPs.join('\n');
            }
            if (res.data.sslID > 0) {
                req.value.enableSSL = true;
                req.value.sslID = res.data.sslID;
                req.value.acmeAccountID = res.data.acmeAccountID;
                listAcmeAccount();
            }
        }
    } catch (e) {}
};

const toWebsite = (websiteID: number) => {
    if (websiteID != undefined && websiteID > 0) {
        window.location.href = `/websites/${websiteID}/config/basic`;
    } else {
        window.location.href = '/websites';
    }
};

const toInstalled = () => {
    window.location.href = '/apps/installed';
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.pageRoute {
    font-size: 12px;
    margin-left: 5px;
}
</style>
