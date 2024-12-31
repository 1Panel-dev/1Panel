<template>
    <el-dialog
        v-model="open"
        :title="$t('commons.button.create')"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        width="40%"
        :before-close="handleClose"
    >
        <el-row>
            <el-col :span="22" :offset="1">
                <el-form ref="accountForm" label-position="top" :model="account" :rules="rules">
                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input v-model.trim="account.name"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('commons.table.type')" prop="type">
                        <el-select
                            v-model="account.type"
                            :disabled="accountData.mode === 'edit'"
                            @change="changeType(account.type)"
                        >
                            <el-option
                                v-for="(type, index) in DNSTypes"
                                :key="index"
                                :label="type.label"
                                :value="type.value"
                            ></el-option>
                        </el-select>
                        <span class="input-help text-red-500" v-if="account.type === 'DnsPod'">
                            {{ $t('ssl.deprecatedHelper') }}
                        </span>
                        <span class="input-help text-red-500" v-if="account.type === 'CloudFlare'">
                            {{ $t('ssl.cfHelper') }}
                        </span>
                    </el-form-item>
                    <div v-if="account.type === 'AliYun' || account.type === 'HuaweiCloud'">
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
                    <div v-if="account.type === 'Volcengine'">
                        <el-form-item label="Access key" prop="authorization.accessKey">
                            <el-input v-model.trim="account.authorization['accessKey']"></el-input>
                        </el-form-item>
                        <el-form-item label="Secret key" prop="authorization.secretKey">
                            <el-input v-model.trim="account.authorization['secretKey']"></el-input>
                        </el-form-item>
                    </div>
                    <div v-if="account.type === 'TencentCloud'">
                        <el-form-item label="Secret ID" prop="authorization.secretID">
                            <el-input v-model.trim="account.authorization['secretID']"></el-input>
                        </el-form-item>
                        <el-form-item label="Secret Key" prop="authorization.secretKey">
                            <el-input v-model.trim="account.authorization['secretKey']"></el-input>
                        </el-form-item>
                    </div>
                    <div v-if="account.type === 'DnsPod'">
                        <el-form-item label="ID" prop="authorization.id">
                            <el-input v-model.trim="account.authorization['id']"></el-input>
                        </el-form-item>
                        <el-form-item label="Token" prop="authorization.token">
                            <el-input v-model.trim="account.authorization['token']"></el-input>
                        </el-form-item>
                    </div>

                    <div v-if="account.type === 'CloudFlare'">
                        <el-form-item label="EMAIL" prop="authorization.email">
                            <el-input v-model.trim="account.authorization['email']"></el-input>
                        </el-form-item>
                        <el-form-item label="API Token" prop="authorization.apiKey">
                            <el-input v-model.trim="account.authorization['apiKey']"></el-input>
                        </el-form-item>
                    </div>
                    <div v-if="account.type === 'RainYun'">
                        <el-form-item label="API Key" prop="authorization.apiKey">
                            <el-input v-model.trim="account.authorization['apiKey']"></el-input>
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
                    <el-form-item
                        label="API Key"
                        prop="authorization.apiKey"
                        v-if="account.type === 'NameCheap' || account.type === 'NameSilo' || account.type === 'Godaddy'"
                    >
                        <el-input v-model.trim="account.authorization['apiKey']"></el-input>
                    </el-form-item>

                    <el-form-item label="API User" prop="authorization.apiUser" v-if="account.type === 'NameCheap'">
                        <el-input v-model.trim="account.authorization['apiUser']"></el-input>
                    </el-form-item>
                    <el-form-item label="API Secret" prop="authorization.apiSecret" v-if="account.type === 'Godaddy'">
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
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(accountForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { CreateDnsAccount, UpdateDnsAccount } from '@/api/modules/website';
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
    mode: 'add',
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
        region: [Rules.requiredInput],
        clientID: [Rules.requiredInput],
        email: [Rules.email],
        password: [Rules.requiredInput],
    },
});
const account = ref({
    id: 0,
    name: '',
    type: 'AliYun',
    authorization: {},
});
const em = defineEmits(['close']);

const handleClose = () => {
    resetForm();
    open.value = false;
    em('close', false);
};

const resetForm = () => {
    account.value = {
        id: 0,
        name: '',
        type: 'AliYun',
        authorization: {},
    };
    accountForm.value?.resetFields();
};

const changeType = (type: string) => {
    account.value.type = type;
    account.value.authorization = {};
    if (account.value.type == 'HuaweiCloud') {
        account.value.authorization['region'] = 'cn-north-1';
    }
};

const acceptParams = async (props: AccountProps) => {
    accountData.value.mode = props.mode;
    if (props.mode === 'edit') {
        account.value = props.form;
    }
    open.value = true;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;

        if (accountData.value.mode === 'edit') {
            UpdateDnsAccount(account.value)
                .then(() => {
                    MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
                    handleClose();
                })
                .finally(() => {
                    loading.value = false;
                });
        } else {
            CreateDnsAccount(account.value)
                .then(() => {
                    MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                    handleClose();
                })
                .finally(() => {
                    loading.value = false;
                });
        }
    });
};

defineExpose({
    acceptParams,
});
</script>
