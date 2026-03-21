<template>
    <DialogPro v-model="open" :title="$t('mail.createDomain')" @close="handleClose">
        <el-row>
            <el-col :span="22" :offset="1">
                <el-form ref="formRef" label-position="top" :model="form" :rules="rules">
                    <el-form-item :label="$t('mail.domainName')" prop="name">
                        <el-input v-model.trim="form.name" placeholder="example.com" />
                    </el-form-item>
                    <el-form-item>
                        <el-checkbox v-model="form.dnsAutoGen">
                            {{ $t('mail.dnsAutoGen') }}
                        </el-checkbox>
                        <span class="input-help">{{ $t('mail.dnsAutoGenHelper') }}</span>
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" @click="submit(formRef)" :loading="loading">{{ $t('commons.button.confirm') }}</el-button>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { FormInstance } from 'element-plus';
import { createMailDomain } from '@/api/modules/mail';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

const open = ref(false);
const loading = ref(false);
const formRef = ref<FormInstance>();
const form = ref({ name: '', dnsAutoGen: true });
const rules = ref({ name: [Rules.requiredInput] });
const emit = defineEmits(['search']);

const acceptParams = () => {
    form.value = { name: '', dnsAutoGen: true };
    open.value = true;
};

const handleClose = () => { open.value = false; formRef.value?.resetFields(); };

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) return;
        loading.value = true;
        createMailDomain(form.value)
            .then(() => { MsgSuccess(i18n.global.t('commons.msg.createSuccess')); handleClose(); emit('search'); })
            .finally(() => { loading.value = false; });
    });
};

defineExpose({ acceptParams });
</script>
