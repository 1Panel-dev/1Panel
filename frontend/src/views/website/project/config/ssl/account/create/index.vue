<template>
    <el-dialog v-model="open" :title="$t('website.createDnsAccount')" width="40%" :before-close="handleClose">
        <el-form ref="accountForm" label-position="right" :model="account" label-width="100px" :rules="rules">
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input v-model="account.name"></el-input>
            </el-form-item>
            <el-form-item :label="$t('website.type')" prop="type">
                <el-select v-model="account.type" :disabled="accountData.mode === 'edit'">
                    <el-option
                        v-for="(type, index) in types"
                        :key="index"
                        :label="type.label"
                        :value="type.value"
                    ></el-option>
                </el-select>
            </el-form-item>
            <div v-if="account.type === 'Aliyun'">
                <el-form-item label="AccessKey" prop="authorization.accessKey">
                    <el-input v-model="account.authorization['accessKey']"></el-input>
                </el-form-item>
                <el-form-item label="SecretKey" prop="authorization.secretKey">
                    <el-input v-model="account.authorization['secretKey']"></el-input>
                </el-form-item>
            </div>
            <div v-if="account.type === 'DnsPod'">
                <el-form-item label="ID" prop="authorization.id">
                    <el-input v-model="account.authorization['id']"></el-input>
                </el-form-item>
                <el-form-item label="Token" prop="authorization.token">
                    <el-input v-model="account.authorization['token']"></el-input>
                </el-form-item>
            </div>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(accountForm)" :loading="loading">
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
import { ElMessage, FormInstance } from 'element-plus';
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
        label: i18n.global.t('website.manual'),
        value: 'Manual',
    },
    {
        label: 'DnsPod',
        value: 'DnsPod',
    },
    {
        label: i18n.global.t('website.aliyun'),
        value: 'Aliyun',
    },
    {
        label: 'CloudFlare',
        value: 'CloudFlare',
    },
];

let open = ref();
let loading = ref(false);
let accountForm = ref<FormInstance>();
let rules = ref({
    name: [Rules.requiredInput],
    type: [Rules.requiredSelect],
    authorization: {
        accessKey: [Rules.requiredInput],
        secretKey: [Rules.requiredInput],
        id: [Rules.requiredInput],
        token: [Rules.requiredInput],
    },
});
let account = ref({
    id: 0,
    name: '',
    type: '',
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
        type: '',
        authorization: {},
    };
    accountForm.value?.resetFields();
};

const acceptParams = async (props: AccountProps) => {
    resetForm();
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
                    ElMessage.success(i18n.global.t('commons.msg.updateSuccess'));
                    handleClose();
                })
                .finally(() => {
                    loading.value = false;
                });
        } else {
            CreateDnsAccount(account.value)
                .then(() => {
                    ElMessage.success(i18n.global.t('commons.msg.createSuccess'));
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
