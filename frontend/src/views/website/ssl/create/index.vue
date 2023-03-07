<template>
    <el-drawer :close-on-click-modal="false" v-model="open" size="50%">
        <template #header>
            <DrawerHeader :header="$t('ssl.create')" :back="handleClose" />
        </template>
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-form ref="sslForm" label-position="top" :model="ssl" label-width="100px" :rules="rules">
                    <el-form-item :label="$t('website.primaryDomain')" prop="primaryDomain">
                        <el-input v-model="ssl.primaryDomain"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('website.otherDomains')" prop="otherDomains">
                        <el-input
                            type="textarea"
                            :autosize="{ minRows: 2, maxRows: 6 }"
                            v-model="ssl.otherDomains"
                        ></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('website.acmeAccount')" prop="acmeAccountId">
                        <el-select v-model="ssl.acmeAccountId">
                            <el-option
                                v-for="(acme, index) in acmeAccounts"
                                :key="index"
                                :label="acme.email"
                                :value="acme.id"
                            ></el-option>
                        </el-select>
                    </el-form-item>
                    <el-form-item :label="$t('website.provider')" prop="provider">
                        <el-radio-group v-model="ssl.provider" @change="changeProvider()">
                            <el-radio label="dnsAccount">{{ $t('website.dnsAccount') }}</el-radio>
                            <el-radio label="dnsManual">{{ $t('website.dnsCommon') }}</el-radio>
                            <el-radio label="http">HTTP</el-radio>
                        </el-radio-group>
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
                                :label="dns.name + ' ( ' + dns.type + ' )'"
                                :value="dns.id"
                            ></el-option>
                        </el-select>
                    </el-form-item>
                    <el-form-item v-if="ssl.provider === 'dnsManual' && dnsResolve.length > 0">
                        <span>{{ $t('ssl.dnsResolveHelper') }}</span>
                        <div v-for="(re, index) in dnsResolve" :key="index">
                            <el-descriptions direction="vertical" :column="4" border>
                                <el-descriptions-item :label="$t('website.domain')">
                                    {{ re.domain }}
                                </el-descriptions-item>
                                <div v-if="re.err != ''">
                                    <el-descriptions-item :label="$t('ssl.err')">{{ re.err }}</el-descriptions-item>
                                </div>
                                <div v-else>
                                    <el-descriptions-item :label="$t('ssl.resolveDomain')">
                                        {{ re.resolve }}
                                    </el-descriptions-item>
                                    <el-descriptions-item :label="$t('ssl.value')">{{ re.value }}</el-descriptions-item>
                                    <el-descriptions-item :label="$t('ssl.type')">TXT</el-descriptions-item>
                                </div>
                            </el-descriptions>
                        </div>
                    </el-form-item>
                    <el-form-item :label="''" prop="autoRenew" v-if="ssl.provider !== 'dnsManual'">
                        <el-checkbox v-model="ssl.autoRenew" :label="$t('ssl.autoRenew')" />
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
import { CreateSSL, GetDnsResolve, SearchAcmeAccount, SearchDnsAccount } from '@/api/modules/website';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { FormInstance } from 'element-plus';
import { computed, reactive, ref } from 'vue';
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

let open = ref(false);
let loading = ref(false);
let dnsReq = reactive({
    page: 1,
    pageSize: 20,
});
let acmeReq = reactive({
    page: 1,
    pageSize: 20,
});
let dnsAccounts = ref<Website.DnsAccount[]>();
let acmeAccounts = ref<Website.AcmeAccount[]>();
let sslForm = ref<FormInstance>();
let rules = ref({
    primaryDomain: [Rules.requiredInput, Rules.domain],
    acmeAccountId: [Rules.requiredSelectBusiness],
    dnsAccountId: [Rules.requiredSelectBusiness],
    provider: [Rules.requiredInput],
    autoRenew: [Rules.requiredInput],
});
let ssl = ref({
    primaryDomain: '',
    otherDomains: '',
    provider: 'dnsAccount',
    websiteId: 0,
    acmeAccountId: undefined,
    dnsAccountId: undefined,
    autoRenew: true,
});
let dnsResolve = ref<Website.DNSResolve[]>([]);
let hasResolve = ref(false);

const em = defineEmits(['close']);
const handleClose = () => {
    resetForm();
    open.value = false;
    em('close', false);
};
const resetForm = () => {
    sslForm.value?.resetFields();
    dnsResolve.value = [];
    ssl.value = {
        primaryDomain: '',
        otherDomains: '',
        provider: 'dnsAccount',
        websiteId: 0,
        acmeAccountId: undefined,
        dnsAccountId: undefined,
        autoRenew: true,
    };
};

const acceptParams = () => {
    resetForm();
    ssl.value.websiteId = Number(id.value);
    getAcmeAccounts();
    getDnsAccounts();
    open.value = true;
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

const getDnsResolve = async (acmeAccountId: number, domains: string[]) => {
    hasResolve.value = false;
    loading.value = true;
    const res = await GetDnsResolve({ acmeAccountId: acmeAccountId, domains: domains });
    if (res.data) {
        dnsResolve.value = res.data;
        hasResolve.value = true;
    }
    loading.value = false;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        if (ssl.value.provider != 'dnsManual' || hasResolve.value) {
            loading.value = true;
            CreateSSL(ssl.value)
                .then(() => {
                    handleClose();
                    MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                })
                .finally(() => {
                    loading.value = false;
                });
        } else {
            let domains = [ssl.value.primaryDomain];
            if (ssl.value.otherDomains != '') {
                let otherDomains = ssl.value.otherDomains.split('\n');
                domains = domains.concat(otherDomains);
            }
            getDnsResolve(ssl.value.acmeAccountId, domains);
        }
    });
};

defineExpose({
    acceptParams,
});
</script>
