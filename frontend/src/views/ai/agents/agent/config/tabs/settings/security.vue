<template>
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top" v-loading="loading">
        <el-form-item :label="t('aiTools.agents.allowedOrigins')" prop="allowedOrigins">
            <el-input
                v-model="form.allowedOrigins"
                type="textarea"
                :rows="4"
                :placeholder="allowedOriginsPlaceholder"
            />
        </el-form-item>
        <el-form-item>
            <el-button v-permission type="primary" :loading="saving" @click="saveConfig">
                {{ t('commons.button.save') }}
            </el-button>
        </el-form-item>
    </el-form>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue';
import type { FormInstance } from 'element-plus';
import { useI18n } from 'vue-i18n';
import { getAgentSecurityConfig, updateAgentSecurityConfig } from '@/api/modules/ai';
import { MsgSuccess } from '@/utils/message';
import { buildDefaultAllowedOrigin, parseAllowedOriginsInput, validateAllowedOriginsInput } from '@/utils/agent';

const { t } = useI18n();
const loading = ref(false);
const saving = ref(false);
const agentId = ref(0);
const appVersion = ref('');
const formRef = ref<FormInstance>();

const form = reactive({
    allowedOrigins: '',
});

const rules = reactive({
    allowedOrigins: [
        {
            validator: (_rule: any, value: any, callback: (error?: Error) => void) => {
                const message = validateAllowedOriginsInput(String(value || ''));
                if (message) {
                    callback(new Error(message));
                    return;
                }
                callback();
            },
            trigger: 'blur',
        },
    ],
});

const allowedOriginsPlaceholder = computed(() => buildDefaultAllowedOrigin('192.168.1.2', 18789, appVersion.value));

const load = async (id: number, version: string) => {
    agentId.value = id;
    appVersion.value = version;
    loading.value = true;
    try {
        const res = await getAgentSecurityConfig({ agentId: id });
        form.allowedOrigins = (res.data?.allowedOrigins || []).join('\n');
    } finally {
        loading.value = false;
    }
};

const saveConfig = async () => {
    if (!agentId.value || !formRef.value) {
        return;
    }
    await formRef.value.validate();
    saving.value = true;
    try {
        await updateAgentSecurityConfig({
            agentId: agentId.value,
            allowedOrigins: parseAllowedOriginsInput(form.allowedOrigins),
        });
        MsgSuccess(t('aiTools.agents.saveSuccess'));
    } finally {
        saving.value = false;
    }
};

defineExpose({
    load,
});
</script>
