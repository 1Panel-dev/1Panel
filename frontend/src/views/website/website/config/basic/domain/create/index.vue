<template>
    <el-dialog
        v-model="open"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :title="$t('website.create')"
        width="40%"
        :before-close="handleClose"
    >
        <el-form ref="domainForm" label-position="right" :model="domain" label-width="130px" :rules="rules">
            <el-form-item :label="$t('website.domain')" prop="domain">
                <el-input v-model="domain.domain"></el-input>
            </el-form-item>
            <el-form-item :label="$t('website.port')" prop="port">
                <el-input v-model.number="domain.port"></el-input>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit(domainForm)" :loading="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { CreateDomain } from '@/api/modules/website';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElMessage, FormInstance } from 'element-plus';
import { ref } from 'vue';

const domainForm = ref<FormInstance>();

let rules = ref({
    domain: [Rules.requiredInput],
    port: [Rules.requiredInput],
});

let open = ref(false);
let loading = ref(false);
let domain = ref({
    websiteId: 0,
    domain: '',
    port: 80,
});

const em = defineEmits(['close']);
const handleClose = () => {
    open.value = false;
    em('close', false);
};

const acceptParams = async (websiteId: number) => {
    domain.value.websiteId = Number(websiteId);
    open.value = true;
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) {
            return;
        }
        loading.value = true;
        CreateDomain(domain.value)
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
