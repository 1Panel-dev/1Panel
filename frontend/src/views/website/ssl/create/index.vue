<template>
    <el-dialog v-model="open" :title="$t('commons.button.create')" width="30%" :before-close="handleClose">
        <el-form
            ref="sslForm"
            label-position="right"
            :model="ssl"
            label-width="125px"
            :rules="rules"
            v-loading="loading"
        >
            <el-form-item :label="$t('website.primaryDomain')" prop="primaryDomain">
                <el-input v-model="ssl.primaryDomain"></el-input>
            </el-form-item>
            <el-form-item :label="$t('website.otherDomains')" prop="otherDomains">
                <el-input type="textarea" :autosize="{ minRows: 2, maxRows: 6 }" v-model="ssl.otherDomains"></el-input>
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
                <el-radio-group v-model="ssl.provider">
                    <el-radio label="dnsAccount">{{ $t('website.dnsAccount') }}</el-radio>
                    <el-radio label="dnsManual">{{ $t('website.dnsCommon') }}</el-radio>
                    <el-radio label="http">HTTP</el-radio>
                </el-radio-group>
            </el-form-item>
            <el-form-item :label="$t('website.dnsAccount')" prop="dnsAccountId" v-if="ssl.provider === 'dnsAccount'">
                <el-select v-model="ssl.dnsAccountId">
                    <el-option
                        v-for="(dns, index) in dnsAccounts"
                        :key="index"
                        :label="dns.name + ' ( ' + dns.type + ' )'"
                        :value="dns.id"
                    ></el-option>
                </el-select>
            </el-form-item>
            <!-- <el-form-item :label="$t('website.domain')" prop="domains">
                <el-checkbox-group v-model="ssl.domains">
                    <el-checkbox v-for="domain in domains" :key="domain.domain" :label="domain.domain"></el-checkbox>
                </el-checkbox-group>
            </el-form-item> -->
            <!-- <el-form-item>
                <div>
                    <span>解析域名: {{ dnsResolve.key }}</span>
                    <span>记录值: {{ dnsResolve.value }}</span>
                    <span>类型: {{ dnsResolve.type }}</span>
                </div>
            </el-form-item> -->
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(sslForm)" :loading="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { WebSite } from '@/api/interface/website';
import { CreateSSL, SearchAcmeAccount, SearchDnsAccount } from '@/api/modules/website';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElMessage, FormInstance } from 'element-plus';
import { computed, reactive, ref } from 'vue';

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
let dnsAccounts = ref<WebSite.DnsAccount[]>();
let acmeAccounts = ref<WebSite.AcmeAccount[]>();
// let domains = ref<WebSite.Domain[]>([]);
let sslForm = ref<FormInstance>();
let rules = ref({
    primaryDomain: [Rules.requiredInput],
    acmeAccountId: [Rules.requiredSelectBusiness],
    dnsAccountId: [Rules.requiredSelectBusiness],
    provider: [Rules.requiredInput],
});
let ssl = ref({
    primaryDomain: '',
    otherDomains: '',
    provider: 'dnsAccount',
    websiteId: 0,
    acmeAccountId: 0,
    dnsAccountId: 0,
});
// let dnsResolve = ref<WebSite.DNSResolve>({
//     key: '',
//     value: '',
//     type: '',
// });
let hasResolve = ref(false);

const em = defineEmits(['close']);
const handleClose = () => {
    resetForm();
    open.value = false;
    em('close', false);
};
const resetForm = () => {
    sslForm.value?.resetFields();
    ssl.value = {
        primaryDomain: '',
        otherDomains: '',
        provider: 'dnsAccount',
        websiteId: 0,
        acmeAccountId: 0,
        dnsAccountId: 0,
    };
};

const acceptParams = () => {
    resetForm();
    ssl.value.websiteId = Number(id.value);
    // getWebsite(id.value);
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

// const getWebsite = async (id: number) => {
//     domains.value = (await GetWebsite(id)).data.domains || [];
// };

// const getDnsResolve = async (acmeAccountId: number, domains: string[]) => {
//     hasResolve.value = false;
//     const res = await GetDnsResolve({ acmeAccountId: acmeAccountId, domains: domains });
//     if (res.data) {
//         dnsResolve.value = res.data;
//         hasResolve.value = true;
//     }
// };

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
                    ElMessage.success(i18n.global.t('commons.msg.createSuccess'));
                })
                .finally(() => {
                    loading.value = false;
                });
        } else {
            // getDnsResolve(ssl.value.acmeAccountId, ssl.value.domains);
        }
    });
};

defineExpose({
    acceptParams,
});
</script>
