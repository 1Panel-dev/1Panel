<template>
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top" v-loading="loading">
        <el-form-item :label="t('aiTools.agents.browserEnabled')">
            <el-switch v-model="form.enabled" />
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.headless')">
            <el-switch v-model="form.headless" />
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.noSandbox')">
            <el-switch v-model="form.noSandbox" />
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.defaultProfile')" prop="defaultProfile">
            <el-input v-model="form.defaultProfile" />
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.executablePath')">
            <el-input v-model="form.executablePath" disabled />
        </el-form-item>
        <el-form-item>
            <el-button type="primary" :loading="saving" @click="saveConfig">
                {{ t('commons.button.save') }}
            </el-button>
        </el-form-item>
    </el-form>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import type { FormInstance } from 'element-plus';
import { useI18n } from 'vue-i18n';
import { Rules } from '@/global/form-rules';
import { AI } from '@/api/interface/ai';
import { getAgentBrowserConfig, updateAgentBrowserConfig } from '@/api/modules/ai';
import { MsgSuccess } from '@/utils/message';

const { t } = useI18n();
const loading = ref(false);
const saving = ref(false);
const agentId = ref(0);
const formRef = ref<FormInstance>();

const form = reactive<AI.AgentBrowserConfig>({
    enabled: true,
    executablePath: '',
    headless: true,
    noSandbox: true,
    defaultProfile: 'openclaw',
});

const rules = reactive({
    defaultProfile: [Rules.requiredInput],
});

const load = async (id: number) => {
    agentId.value = id;
    loading.value = true;
    try {
        const res = await getAgentBrowserConfig({ agentId: id });
        Object.assign(form, res.data || {});
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
        await updateAgentBrowserConfig({
            agentId: agentId.value,
            enabled: form.enabled,
            headless: form.headless,
            noSandbox: form.noSandbox,
            defaultProfile: form.defaultProfile,
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
