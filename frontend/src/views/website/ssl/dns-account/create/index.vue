<template>
    <DrawerPro v-model="open" :header="$t('commons.button.' + accountData.mode)" size="large" @close="handleClose">
        <el-form v-loading="loading" ref="accountForm" label-position="top" :model="account" :rules="rules">
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input v-model.trim="account.name"></el-input>
            </el-form-item>
            <el-form-item :label="$t('commons.table.type')" prop="type">
                <el-select v-model="account.type" :disabled="accountData.mode === 'edit'">
                    <el-option
                        v-for="(type, index) in DNSTypes"
                        :key="index"
                        :label="type.deprecated ? `${type.label} (${$t('ssl.dnsPodRemoved')})` : type.label"
                        :value="type.value"
                        :disabled="type.deprecated && accountData.mode === 'create'"
                    ></el-option>
                </el-select>
                <span class="input-help text-red-500" v-if="account.type === 'DnsPod'">
                    {{ $t('ssl.dnsPodRemovedTip') }}
                </span>
            </el-form-item>
            <div
                v-if="
                    account.type === 'AliYun' ||
                    account.type === 'AliESA' ||
                    account.type === 'HuaweiCloud' ||
                    account.type === 'BaiduCloud' ||
                    account.type === 'AWSRoute53'
                "
            >
                <el-form-item label="Access key" prop="authorization.accessKey">
                    <el-input v-model.trim="account.authorization['accessKey']"></el-input>
                </el-form-item>
                <el-form-item label="Secret key" prop="authorization.secretKey">
                    <el-input v-model.trim="account.authorization['secretKey']"></el-input>
                </el-form-item>
            </div>
            <el-form-item label="Region" prop="authorization.region" v-if="account.type === 'HuaweiCloud'">
                <el-input v-model.trim="account.authorization['region']" :placeholder="'cn-north-1'"></el-input>
            </el-form-item>
            <div v-if="account.type === 'TencentCloud'">
                <el-form-item label="Secret ID" prop="authorization.secretID">
                    <el-input v-model.trim="account.authorization['secretID']"></el-input>
                </el-form-item>
                <el-form-item label="Secret Key" prop="authorization.secretKey">
                    <el-input v-model.trim="account.authorization['secretKey']"></el-input>
                </el-form-item>
            </div>
            <div v-if="account.type === 'Volcengine'">
                <el-form-item label="Access key" prop="authorization.accessKey">
                    <el-input v-model.trim="account.authorization['accessKey']"></el-input>
                </el-form-item>
                <el-form-item label="Secret key" prop="authorization.secretKey">
                    <el-input v-model.trim="account.authorization['secretKey']"></el-input>
                </el-form-item>
            </div>
            <div v-if="account.type === 'CloudDns'">
                <el-form-item label="Client ID" prop="authorization.clientID">
                    <el-input v-model.trim="account.authorization['clientID']"></el-input>
                </el-form-item>
                <el-form-item label="Email" prop="authorization.email">
                    <el-input v-model.trim="account.authorization['email']"></el-input>
                </el-form-item>
                <el-form-item label="Password" prop="authorization.password">
                    <el-input v-model.trim="account.authorization['password']"></el-input>
                </el-form-item>
            </div>
            <div v-if="account.type === 'CloudFlare'">
                <el-form-item label="EMAIL" prop="authorization.email">
                    <el-input v-model.trim="account.authorization['email']"></el-input>
                </el-form-item>
                <el-form-item label="API Token" prop="authorization.apiKey">
                    <el-input v-model.trim="account.authorization['apiKey']"></el-input>
                    <span class="input-help text-red-500">
                        {{ $t('ssl.cfHelper') }}
                    </span>
                </el-form-item>
            </div>
            <el-form-item
                label="API Key"
                prop="authorization.apiKey"
                v-if="
                    account.type === 'NameCheap' ||
                    account.type === 'NameSilo' ||
                    account.type === 'Godaddy' ||
                    account.type === 'RainYun' ||
                    account.type === 'Spaceship' ||
                    account.type === 'Dynadot' ||
                    account.type === 'Dynu'
                "
            >
                <el-input v-model.trim="account.authorization['apiKey']"></el-input>
            </el-form-item>
            <el-form-item label="API User" prop="authorization.apiUser" v-if="account.type === 'NameCheap'">
                <el-input v-model.trim="account.authorization['apiUser']"></el-input>
            </el-form-item>
            <el-form-item
                label="API Secret"
                prop="authorization.apiSecret"
                v-if="account.type === 'Godaddy' || account.type === 'Spaceship' || account.type === 'Dynadot'"
            >
                <el-input v-model.trim="account.authorization['apiSecret']"></el-input>
            </el-form-item>
            <div v-if="account.type === 'NameCom'">
                <el-form-item label="Username" prop="authorization.apiUser">
                    <el-input v-model.trim="account.authorization['apiUser']"></el-input>
                </el-form-item>
                <el-form-item label="Token" prop="authorization.token">
                    <el-input v-model.trim="account.authorization['token']"></el-input>
                </el-form-item>
            </div>
            <div v-if="account.type === 'Netcup'">
                <el-form-item label="Customer Number" prop="authorization.customerNumber">
                    <el-input v-model.trim="account.authorization['customerNumber']"></el-input>
                </el-form-item>
                <el-form-item label="API Key" prop="authorization.apiKey">
                    <el-input v-model.trim="account.authorization['apiKey']"></el-input>
                </el-form-item>
                <el-form-item label="API Password" prop="authorization.password">
                    <el-input v-model.trim="account.authorization['password']" type="password" show-password></el-input>
                </el-form-item>
            </div>
            <div v-if="account.type === 'FreeMyIP' || account.type === 'Vercel' || account.type === 'IonosCloud'">
                <el-form-item label="Token" prop="authorization.token">
                    <el-input v-model.trim="account.authorization['token']"></el-input>
                </el-form-item>
            </div>
            <div v-if="account.type === 'Ionos'">
                <el-form-item label="API Prefix" prop="authorization.apiPrefix">
                    <el-input v-model.trim="account.authorization['apiPrefix']"></el-input>
                </el-form-item>
                <el-form-item label="API Secret" prop="authorization.apiSecret">
                    <el-input v-model.trim="account.authorization['apiSecret']"></el-input>
                </el-form-item>
            </div>
            <div v-if="account.type === 'ClouDNS'">
                <el-form-item label="Auth ID" prop="authorization.authID">
                    <el-input v-model.trim="account.authorization['authID']"></el-input>
                </el-form-item>
                <el-form-item label="Sub Auth ID" prop="authorization.subAuthID">
                    <el-input v-model.trim="account.authorization['subAuthID']"></el-input>
                </el-form-item>
                <el-form-item label="Auth Password" prop="authorization.authPassword">
                    <el-input v-model.trim="account.authorization['authPassword']"></el-input>
                </el-form-item>
            </div>
            <div v-if="account.type === 'WestCN' || account.type === 'RegRu'">
                <el-form-item label="Username" prop="authorization.username">
                    <el-input v-model.trim="account.authorization['username']"></el-input>
                </el-form-item>
                <el-form-item label="Password" prop="authorization.password">
                    <el-input v-model.trim="account.authorization['password']"></el-input>
                </el-form-item>
            </div>
            <div v-if="account.type === 'Ovh'">
                <el-form-item label="Endpoint" prop="authorization.endpoint" :rules="[Rules.requiredInput]">
                    <el-input v-model.trim="account.authorization['endpoint']"></el-input>
                </el-form-item>
                <el-form-item label="AccessToken" prop="authorization.accessToken" :rules="[Rules.requiredInput]">
                    <el-input v-model.trim="account.authorization['accessToken']"></el-input>
                </el-form-item>
            </div>
            <div v-if="account.type === 'AcmeDNS'">
                <el-form-item label="API BASE" prop="authorization.endpoint" :rules="[Rules.requiredInput]">
                    <el-input v-model.trim="account.authorization['endpoint']"></el-input>
                </el-form-item>
                <el-form-item label="BASE URL" prop="authorization.baseURL" :rules="[Rules.requiredInput]">
                    <el-input v-model.trim="account.authorization['baseURL']"></el-input>
                </el-form-item>
            </div>
            <div v-if="account.type === 'AWSRoute53'">
                <el-form-item label="Region" prop="authorization.region">
                    <el-input v-model.trim="account.authorization['region']" :placeholder="'us-east-1'"></el-input>
                </el-form-item>
                <el-form-item
                    :label="$t('ssl.hostedZoneID')"
                    prop="authorization.endpoint"
                    :rules="[Rules.requiredInput]"
                >
                    <el-input v-model.trim="account.authorization['endpoint']"></el-input>
                </el-form-item>
            </div>
            <div v-if="account.type === 'PorkBun'">
                <el-form-item label="API Key" prop="authorization.apiKey">
                    <el-input v-model.trim="account.authorization['apiKey']"></el-input>
                </el-form-item>
                <el-form-item label="Secret Key" prop="authorization.secretKey">
                    <el-input v-model.trim="account.authorization['secretKey']"></el-input>
                </el-form-item>
            </div>
            <div v-if="account.type === 'Technitium'">
                <el-form-item label="BASE URL" prop="authorization.baseURL" :rules="[Rules.requiredInput]">
                    <el-input v-model.trim="account.authorization['baseURL']"></el-input>
                </el-form-item>
                <el-form-item label="Token" prop="authorization.token">
                    <el-input v-model.trim="account.authorization['token']"></el-input>
                </el-form-item>
            </div>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button v-permission type="primary" @click="submit(accountForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { createDnsAccount, updateDnsAccount } from '@/api/modules/website';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { ref } from 'vue';
import { DNSTypes } from '@/global/mimetype';

interface AccountProps {
    mode: string;
    form: any;
}
const accountData = ref<AccountProps>({
    mode: 'create',
    form: {},
});

const open = ref();
const loading = ref(false);
const accountForm = ref<FormInstance>();
const rules = ref<any>({
    name: [Rules.requiredInput, Rules.linuxName],
    type: [Rules.requiredSelect],
    authorization: {
        accessKey: [Rules.requiredInput],
        secretKey: [Rules.requiredInput],
        id: [Rules.requiredInput],
        token: [Rules.requiredInput],
        apiKey: [Rules.requiredInput],
        apiUser: [Rules.requiredInput],
        secretID: [Rules.requiredInput],
        apiSecret: [Rules.requiredInput],
        apiPrefix: [Rules.requiredInput],
        username: [Rules.requiredInput],
        password: [Rules.requiredInput],
        customerNumber: [Rules.requiredInput],
    },
});
const account = ref({
    id: 0,
    name: '',
    type: 'AliYun',
    authorization: {},
});
const em = defineEmits(['saved']);

const handleClose = () => {
    resetForm();
    open.value = false;
};

const resetForm = () => {
    accountForm.value?.resetFields();
    account.value = { id: 0, name: '', type: 'AliYun', authorization: {} };
};

const acceptParams = (props: AccountProps) => {
    resetForm();
    accountData.value.mode = props.mode;
    if (props.mode === 'edit') {
        account.value = { ...props.form, authorization: { ...props.form.authorization } };
    }
    open.value = true;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl || loading.value) return;
    loading.value = true;
    try {
        if (!(await formEl.validate().catch(() => false))) return;
        const save = accountData.value.mode === 'edit' ? updateDnsAccount : createDnsAccount;
        await save(account.value);
        MsgSuccess(
            i18n.global.t(
                accountData.value.mode === 'edit' ? 'commons.msg.updateSuccess' : 'commons.msg.createSuccess',
            ),
        );
        handleClose();
        em('saved');
    } finally {
        loading.value = false;
    }
};

defineExpose({
    acceptParams,
});
</script>
