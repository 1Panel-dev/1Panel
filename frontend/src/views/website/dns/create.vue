<template>
    <DialogPro v-model="open" :title="$t('dns.createZone')" @close="handleClose">
        <el-row>
            <el-col :span="22" :offset="1">
                <el-form ref="formRef" label-position="top" :model="form" :rules="rules">
                    <el-form-item :label="$t('dns.zoneName')" prop="name">
                        <el-input v-model.trim="form.name" placeholder="example.com" />
                    </el-form-item>
                    <el-form-item :label="$t('dns.zoneType')" prop="type">
                        <el-select v-model="form.type" style="width: 100%">
                            <el-option :label="$t('dns.native')" value="Native" />
                            <el-option :label="$t('dns.slave')" value="Slave" />
                        </el-select>
                    </el-form-item>
                    <el-form-item v-if="form.type === 'Slave'" :label="$t('dns.masterIP')" prop="masterIP">
                        <el-input v-model.trim="form.masterIP" placeholder="192.168.1.1" />
                    </el-form-item>
                    <el-form-item :label="$t('dns.soaEmail')" prop="soaEmail">
                        <el-input v-model.trim="form.soaEmail" placeholder="admin@example.com" />
                    </el-form-item>
                    <el-form-item :label="$t('dns.soaPrimary')" prop="soaPrimary">
                        <el-input v-model.trim="form.soaPrimary" placeholder="ns1.example.com" />
                        <span class="input-help">
                            {{ $t('commons.msg.unset') || 'Leave empty to use ns1.{domain}' }}
                        </span>
                    </el-form-item>
                    <el-form-item :label="$t('dns.defaultTTL')" prop="defaultTTL">
                        <el-input-number v-model="form.defaultTTL" :min="60" :max="86400" :step="60" />
                    </el-form-item>
                </el-form>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose" :disabled="loading">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button type="primary" @click="submit(formRef)" :loading="loading">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { FormInstance } from 'element-plus';
import { createDnsZone } from '@/api/modules/dns';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

const open = ref(false);
const loading = ref(false);
const formRef = ref<FormInstance>();

const form = ref({
    name: '',
    type: 'Native',
    masterIP: '',
    soaEmail: '',
    soaPrimary: '',
    defaultTTL: 3600,
});

const rules = ref({
    name: [Rules.requiredInput],
    type: [Rules.requiredSelect],
    soaEmail: [Rules.requiredInput],
});

const emit = defineEmits(['search']);

const acceptParams = () => {
    form.value = {
        name: '',
        type: 'Native',
        masterIP: '',
        soaEmail: '',
        soaPrimary: '',
        defaultTTL: 3600,
    };
    open.value = true;
};

const handleClose = () => {
    open.value = false;
    formRef.value?.resetFields();
};

const submit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    await formEl.validate((valid) => {
        if (!valid) return;
        loading.value = true;
        createDnsZone(form.value)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
                handleClose();
                emit('search');
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
