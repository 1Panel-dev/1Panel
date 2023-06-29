<template>
    <el-dialog
        v-model="open"
        :title="$t('website.addAccount')"
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
                        <el-select v-model="account.type" :disabled="accountData.mode === 'edit'">
                            <el-option
                                v-for="(type, index) in types"
                                :key="index"
                                :label="type.label"
                                :value="type.value"
                            ></el-option>
                        </el-select>
                    </el-form-item>
                    <div v-if="account.type === 'AliYun'">
                        <el-form-item label="Access Key" prop="authorization.accessKey">
                            <el-input v-model.trim="account.authorization['accessKey']"></el-input>
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
                        <el-form-item label="API Key" prop="authorization.apiKey">
                            <el-input v-model.trim="account.authorization['apiKey']"></el-input>
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

interface AccountProps {
    mode: string;
    form: any;
}
const accountData = ref<AccountProps>({
    mode: 'add',
    form: {},
});

const types = [
    {
        label: i18n.global.t('website.aliyun'),
        value: 'AliYun',
    },
    {
        label: 'DNSPod',
        value: 'DnsPod',
    },
    {
        label: 'CloudFlare',
        value: 'CloudFlare',
    },
];

let open = ref();
let loading = ref(false);
let accountForm = ref<FormInstance>();
let rules = ref<any>({
    name: [Rules.requiredInput, Rules.linuxName],
    type: [Rules.requiredSelect],
    authorization: {
        accessKey: [Rules.requiredInput],
        secretKey: [Rules.requiredInput],
        id: [Rules.requiredInput],
        token: [Rules.requiredInput],
        email: [Rules.requiredInput],
        apiKey: [Rules.requiredInput],
    },
});
let account = ref({
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
        type: 'DnsPod',
        authorization: {},
    };
    accountForm.value?.resetFields();
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
