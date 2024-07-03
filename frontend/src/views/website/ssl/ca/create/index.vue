<template>
    <el-dialog
        v-model="open"
        :title="$t('commons.button.create')"
        :close-on-click-modal="false"
        width="40%"
        :before-close="handleClose"
    >
        <el-row v-loading="loading">
            <el-col :span="22" :offset="1">
                <el-form @submit.prevent ref="caForm" label-position="top" :model="ca" :rules="rules">
                    <el-form-item :label="$t('ssl.caName')" prop="name">
                        <el-input v-model.trim="ca.name"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('ssl.commonName')" prop="commonName">
                        <el-input v-model.trim="ca.commonName"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('ssl.company')" prop="organization">
                        <el-input v-model.trim="ca.organization"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('ssl.department')" prop="organizationUint">
                        <el-input v-model.trim="ca.organizationUint"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('ssl.country')" prop="country">
                        <el-input v-model.trim="ca.country"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('ssl.province')" prop="province">
                        <el-input v-model.trim="ca.province"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('ssl.city')" prop="city">
                        <el-input v-model.trim="ca.city"></el-input>
                    </el-form-item>
                    <el-form-item :label="$t('website.keyType')" prop="keyType">
                        <el-select v-model="ca.keyType">
                            <el-option
                                v-for="(keyType, index) in KeyTypes"
                                :key="index"
                                :label="keyType.label"
                                :value="keyType.value"
                            ></el-option>
                        </el-select>
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(caForm)" :disabled="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { FormInstance } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { KeyTypes } from '@/global/mimetype';
import { CreateCA } from '@/api/modules/website';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

const open = ref(false);
const loading = ref(false);
const caForm = ref<FormInstance>();
const em = defineEmits(['close']);

const rules = ref({
    keyType: [Rules.requiredSelect],
    name: [Rules.requiredInput, Rules.name],
    country: [Rules.requiredSelect],
    organization: [Rules.requiredInput, Rules.name],
    commonName: [Rules.requiredInput, Rules.name],
});

const initData = () => ({
    name: '',
    keyType: 'P256',
    commonName: '',
    country: 'CN',
    organization: '',
    organizationUint: '',
    province: '',
    city: '',
});

const ca = ref(initData());

const handleClose = () => {
    open.value = false;
    em('close', false);
    resetForm();
};

const resetForm = () => {
    caForm.value.resetFields();
    ca.value = initData();
};

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

        CreateCA(ca.value)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
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
