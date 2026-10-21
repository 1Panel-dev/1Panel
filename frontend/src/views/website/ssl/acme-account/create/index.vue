<template>
    <DrawerPro v-model="open" :header="$t('commons.button.create')" size="large" @close="handleClose">
        <el-form
            v-loading="loading"
            @submit.prevent
            ref="accountForm"
            label-position="top"
            :model="account"
            :rules="rules"
        >
            <el-form-item :label="$t('website.email')" prop="email">
                <el-input v-model.trim="account.email"></el-input>
            </el-form-item>
            <el-form-item :label="$t('website.useProxy')" prop="useProxy" v-if="isProductPro">
                <el-switch v-model="account.useProxy"></el-switch>
                <span class="input-help">
                    {{ $t('website.useProxyHelper') }}
                </span>
            </el-form-item>
            <el-form-item :label="$t('website.acmeAccountType')" prop="type">
                <el-select v-model="account.type">
                    <el-option
                        v-for="(acme, index) in AcmeAccountTypes"
                        :key="index"
                        :label="acme.label"
                        :value="acme.value"
                    ></el-option>
                </el-select>
                <span class="input-help" v-if="account.type === 'buypass'">
                    {{ $t('ssl.buypassHelper') }}
                </span>
                <span class="input-help" v-if="account.type == 'google'">
                    {{ $t('ssl.googleCloudHelper') }}
                </span>
            </el-form-item>
            <el-form-item :label="$t('website.keyType')" prop="keyType">
                <el-select v-model="account.keyType">
                    <el-option
                        v-for="(keyType, index) in KeyTypes"
                        :key="index"
                        :label="keyType.label"
                        :value="keyType.value"
                    ></el-option>
                </el-select>
            </el-form-item>
            <div v-if="account.type == 'custom'">
                <el-form-item :label="$t('ssl.customAcmeURL')" prop="caDirURL">
                    <el-input v-model.trim="account.caDirURL"></el-input>
                </el-form-item>
                <el-form-item :label="$t('ssl.useEAB')" prop="useEAB">
                    <el-switch v-model="account.useEAB"></el-switch>
                </el-form-item>
            </div>

            <div
                v-if="
                    account.type == 'google' ||
                    account.type == 'freessl' ||
                    (account.type === 'custom' && account.useEAB)
                "
            >
                <el-form-item label="EAB kid" prop="eabKid">
                    <el-input v-model.trim="account.eabKid"></el-input>
                </el-form-item>
                <el-form-item label="EAB HmacKey" prop="eabHmacKey">
                    <el-input type="textarea" :rows="3" v-model.trim="account.eabHmacKey"></el-input>
                </el-form-item>
                <el-link
                    v-if="account.type == 'google'"
                    class="ml-1.5"
                    type="primary"
                    target="_blank"
                    href="https://cloud.google.com/certificate-manager/docs/public-ca-tutorial?hl=zh-cn"
                >
                    {{ $t('ssl.googleHelper') }}
                </el-link>
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
import { FormInstance } from 'element-plus';
import { ref } from 'vue';
import { Rules } from '@/global/form-rules';
import { createAcmeAccount } from '@/api/modules/website';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { AcmeAccountTypes, KeyTypes } from '@/global/mimetype';
import { useGlobalStore } from '@/composables/useGlobalStore';

const { isProductPro } = useGlobalStore();
const open = ref();
const loading = ref(false);
const accountForm = ref<FormInstance>();
const rules = ref({
    email: [Rules.requiredInput, Rules.email],
    type: [Rules.requiredSelect],
    eabKid: [Rules.requiredInput],
    eabHmacKey: [Rules.requiredInput],
    keyType: [Rules.requiredSelect],
    caDirURL: [Rules.requiredInput],
});

const initData = () => ({
    email: '',
    type: 'letsencrypt',
    eabKid: '',
    eabHmacKey: '',
    keyType: 'EC256',
    useProxy: false,
    caDirURL: '',
    useEAB: false,
});

const account = ref(initData());
const em = defineEmits(['saved']);

const handleClose = () => {
    resetForm();
    open.value = false;
};

const resetForm = () => {
    accountForm.value?.resetFields();
    account.value = initData();
};

const acceptParams = () => {
    open.value = true;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl || loading.value) return;
    loading.value = true;
    try {
        if (!(await formEl.validate().catch(() => false))) return;
        await createAcmeAccount(account.value);
        MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
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
