<template>
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top" v-loading="loading">
        <el-form-item v-if="agentType === 'openclaw'" :label="t('aiTools.agents.browserEnabled')">
            <el-switch v-model="form.browserEnabled" />
        </el-form-item>
        <el-form-item v-if="agentType === 'openclaw'" :label="t('aiTools.agents.npmRegistry')" prop="npmRegistry">
            <el-select v-model="form.npmRegistry" filterable allow-create default-first-option>
                <el-option v-for="item in npmRegistryOptions" :key="item" :label="item" :value="item" />
            </el-select>
            <span class="input-help">{{ t('aiTools.agents.npmRegistryHelper') }}</span>
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.timeZone')" prop="userTimezone">
            <el-input v-model="form.userTimezone" />
        </el-form-item>
        <el-form-item>
            <el-button v-permission type="primary" :loading="saving" @click="saveConfig">
                {{ t('commons.button.save') }}
            </el-button>
        </el-form-item>
        <span class="input-help">{{ t('aiTools.agents.channelAutoRestartHelper') }}</span>
    </el-form>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import type { FormInstance } from 'element-plus';
import { useI18n } from 'vue-i18n';
import { Rules } from '@/global/form-rules';
import { AI } from '@/api/interface/ai';
import { getAgentOtherConfig, updateAgentOtherConfig } from '@/api/modules/ai';
import { MsgSuccess } from '@/utils/message';

const { t } = useI18n();
const loading = ref(false);
const saving = ref(false);
const agentId = ref(0);
const agentType = ref<AI.AgentType>('openclaw');
const formRef = ref<FormInstance>();
const defaultNPMRegistry = 'https://registry.npmjs.org/';

const form = reactive<AI.AgentOtherConfig>({
    userTimezone: '',
    browserEnabled: true,
    npmRegistry: defaultNPMRegistry,
});

const npmRegistryOptions = [
    defaultNPMRegistry,
    'https://registry.npmmirror.com',
    'https://mirrors.cloud.tencent.com/npm/',
    'https://repo.huaweicloud.com/repository/npm/',
];

const validateNPMRegistry = (_rule: any, value: string, callback: (error?: Error) => void) => {
    if (!value) {
        callback(new Error(t('commons.rule.requiredInput')));
        return;
    }
    if (!/^https?:\/\/\S+$/.test(value)) {
        callback(new Error(t('aiTools.agents.npmRegistryInvalid')));
        return;
    }
    callback();
};

const rules = reactive({
    userTimezone: [Rules.requiredInput],
    npmRegistry: [{ validator: validateNPMRegistry, trigger: 'blur' }],
});

const load = async (params: { agentId: number; agentType: AI.AgentType }) => {
    agentId.value = params.agentId;
    agentType.value = params.agentType;
    loading.value = true;
    try {
        const res = await getAgentOtherConfig({ agentId: params.agentId });
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
        await updateAgentOtherConfig({
            agentId: agentId.value,
            userTimezone: form.userTimezone,
            browserEnabled: form.browserEnabled,
            npmRegistry: agentType.value === 'openclaw' ? form.npmRegistry : defaultNPMRegistry,
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

<style lang="scss" scoped>
.input-help {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    line-height: 1.5;
}
</style>
