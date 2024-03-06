<template>
    <el-drawer :close-on-click-modal="false" v-model="open" size="50%">
        <template #header>
            <DrawerHeader :header="$t('ssl.create')" :back="handleClose" />
        </template>
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-form ref="sslForm" label-position="top" :model="ssl" label-width="100px" :rules="rules">
                    <el-row :gutter="20">
                        <el-col :span="12">
                            <el-form-item :label="$t('website.primaryDomain')" prop="primaryDomain">
                                <el-input v-model.trim="ssl.primaryDomain"></el-input>
                            </el-form-item>
                        </el-col>
                        <el-col :span="12">
                            <el-form-item :label="$t('ssl.fromWebsite')">
                                <el-select v-model="websiteID" @change="changeWebsite">
                                    <el-option
                                        v-for="(site, key) in websites"
                                        :key="key"
                                        :value="site.id"
                                        :label="site.primaryDomain"
                                    ></el-option>
                                </el-select>
                            </el-form-item>
                        </el-col>
                    </el-row>
                    <el-form-item :label="$t('website.otherDomains')" prop="otherDomains">
                        <el-input type="textarea" :rows="3" v-model="ssl.otherDomains"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('website.remark')" prop="description">
                        <el-input v-model="ssl.description"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('website.acmeAccount')" prop="acmeAccountId">
                        <el-select v-model="ssl.acmeAccountId">
                            <el-option
                                v-for="(acme, index) in acmeAccounts"
                                :key="index"
                                :label="acme.email + ' [' + getAccountName(acme.type) + '] '"
                                :value="acme.id"
                            >
                                <el-row>
                                    <el-col :span="6">
                                        <span>{{ acme.email }}</span>
                                    </el-col>
                                    <el-col :span="11">
                                        <span>
                                            <el-tag type="success">{{ getAccountName(acme.type) }}</el-tag>
                                        </span>
                                    </el-col>
                                </el-row>
                            </el-option>
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('website.keyType')" prop="keyType">
                        <el-select v-model="ssl.keyType">
                            <el-option
                                v-for="(keyType, index) in KeyTypes"
                                :key="index"
                                :label="keyType.label"
                                :value="keyType.value"
                            ></el-option>
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('website.provider')" prop="provider">
                        <el-radio-group v-model="ssl.provider" @change="changeProvider()">
                            <el-radio value="dnsAccount">{{ $t('website.dnsAccount') }}</el-radio>
                            <el-radio value="dnsManual">{{ $t('website.dnsManual') }}</el-radio>
                            <el-radio value="http">HTTP</el-radio>
                        </el-radio-group>
                        <span class="input-help" v-if="ssl.provider === 'dnsManual'">
                            {{ $t('ssl.dnsMauanlHelper') }}
                        </span>
                        <span class="input-help" v-if="ssl.provider === 'http'">
                            {{ $t('ssl.httpHelper') }}
                        </span>
                        <span class="input-help text-red-500" v-if="ssl.provider === 'http'">
                            {{ $t('ssl.httpHelper2') }}
                        </span>
                    </el-form-item>
                    <el-form-item
                        :label="$t('website.dnsAccount')"
                        prop="dnsAccountId"
                        v-if="ssl.provider === 'dnsAccount'"
                    >
                        <el-select v-model="ssl.dnsAccountId">
                            <el-option
                                v-for="(dns, index) in dnsAccounts"
                                :key="index"
                                :label="dns.name + ' [' + getDNSName(dns.type) + '] '"
                                :value="dns.id"
                            >
                                <el-row>
                                    <el-col :span="6">
                                        <span>{{ dns.name }}</span>
                                    </el-col>
                                    <el-col :span="11">
                                        <span>
                                            <el-tag type="success">{{ getDNSName(dns.type) }}</el-tag>
                                        </span>
                                    </el-col>
                                </el-row>
                            </el-option>
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="''" prop="autoRenew" v-if="ssl.provider !== 'dnsManual'">
                        <el-checkbox v-model="ssl.autoRenew" :label="$t('ssl.autoRenew')" />
                    </el-form-item>
                    <el-form-item :label="''" prop="pushDir">
                        <el-checkbox v-model="ssl.pushDir" :label="$t('ssl.pushDir')" />
                    </el-form-item>
                    <el-form-item :label="$t('ssl.dir')" prop="dir" v-if="ssl.pushDir">
                        <el-input v-model.trim="ssl.dir">
                            <template #prepend>
                                <FileList :path="ssl.dir" @choose="getPath" :dir="true"></FileList>
                            </template>
                        </el-input>
                        <span class="input-help">
                            {{ $t('ssl.pushDirHelper') }}
                        </span>
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
import { Website } from '@/api/interface/website';
import { CreateSSL, ListWebsites, SearchAcmeAccount, SearchDnsAccount } from '@/api/modules/website';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { FormInstance } from 'element-plus';
import { computed, reactive, ref } from 'vue';
import { MsgSuccess } from '@/utils/message';
import { KeyTypes } from '@/global/mimetype';
import { getDNSName, getAccountName } from '@/utils/util';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});

const id = computed(() => {
    return props.id;
});

const open = ref(false);
const loading = ref(false);
const dnsReq = reactive({
    page: 1,
    pageSize: 20,
});
const acmeReq = reactive({
    page: 1,
    pageSize: 20,
});
const dnsAccounts = ref<Website.DnsAccount[]>();
const acmeAccounts = ref<Website.AcmeAccount[]>();
const sslForm = ref<FormInstance>();
const websites = ref();
const rules = ref({
    primaryDomain: [Rules.requiredInput, Rules.domain],
    acmeAccountId: [Rules.requiredSelectBusiness],
    dnsAccountId: [Rules.requiredSelectBusiness],
    provider: [Rules.requiredInput],
    autoRenew: [Rules.requiredInput],
    keyType: [Rules.requiredInput],
    dir: [Rules.requiredInput],
});
const websiteID = ref();

const initData = () => ({
    primaryDomain: '',
    otherDomains: '',
    provider: 'dnsAccount',
    websiteId: 0,
    acmeAccountId: undefined,
    dnsAccountId: undefined,
    autoRenew: true,
    keyType: 'P256',
    pushDir: false,
    dir: '',
    description: '',
});

const ssl = ref(initData());
const dnsResolve = ref<Website.DNSResolve[]>([]);

const em = defineEmits(['close', 'submit']);
const handleClose = () => {
    resetForm();
    open.value = false;
    em('close', false);
};
const resetForm = () => {
    sslForm.value?.resetFields();
    dnsResolve.value = [];
    ssl.value = initData();
    websiteID.value = undefined;
};

const acceptParams = () => {
    resetForm();
    ssl.value.websiteId = Number(id.value);
    getAcmeAccounts();
    getDnsAccounts();
    listwebsites();
    open.value = true;
};

const getPath = (dir: string) => {
    ssl.value.dir = dir;
};

const getAcmeAccounts = async () => {
    const res = await SearchAcmeAccount(acmeReq);
    acmeAccounts.value = res.data.items || [];
    if (acmeAccounts.value.length > 0) {
        ssl.value.acmeAccountId = res.data.items[0].id;
    }
};

const getDnsAccounts = async () => {
    const res = await SearchDnsAccount(dnsReq);
    dnsAccounts.value = res.data.items || [];
    if (dnsAccounts.value.length > 0) {
        ssl.value.dnsAccountId = res.data.items[0].id;
    }
};

const changeProvider = () => {
    dnsResolve.value = [];
};

const listwebsites = async () => {
    const res = await ListWebsites();
    websites.value = res.data;
};

const changeWebsite = () => {
    if (websiteID.value > 0) {
        const selectedWebsite = websites.value.find((website) => website.id == websiteID.value);

        if (selectedWebsite && selectedWebsite.domains && selectedWebsite.domains.length > 0) {
            const primaryDomain = selectedWebsite.domains[0].domain;
            const otherDomains = selectedWebsite.domains
                .slice(1)
                .map((domain) => domain.domain)
                .join('\n');

            ssl.value.primaryDomain = primaryDomain;
            ssl.value.otherDomains = otherDomains;
        }
    }
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        CreateSSL(ssl.value)
            .then((res: any) => {
                if (ssl.value.provider != 'dnsManual') {
                    em('submit', res.data.id);
                }
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
