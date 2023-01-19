<template>
    <el-dialog
        v-model="open"
        :title="$t('commons.button.create')"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        width="30%"
        :before-close="handleClose"
    >
        <el-row>
            <el-col :span="22" :offset="1">
                <el-form ref="accountForm" label-position="top" :model="account" :rules="rules" v-loading="loading">
                    <el-form-item :label="$t('website.email')" prop="email">
                        <el-input v-model="account.email"></el-input>
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
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
import { ElMessage, FormInstance } from 'element-plus';
import { ref } from 'vue';
import { Rules } from '@/global/form-rules';
import { CreateAcmeAccount } from '@/api/modules/website';
import i18n from '@/lang';

let open = ref();
let loading = ref(false);
let accountForm = ref<FormInstance>();
let rules = ref({
    email: [Rules.requiredInput],
});
let account = ref({
    email: '',
});

const em = defineEmits(['close']);

const handleClose = () => {
    resetForm();
    open.value = false;
    em('close', false);
};

const resetForm = () => {};

const acceptParams = () => {
    open.value = true;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;

        CreateAcmeAccount(account.value)
            .then(() => {
                ElMessage.success(i18n.global.t('commons.msg.createSuccess'));
                handleClose();
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
