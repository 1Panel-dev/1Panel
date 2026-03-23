<template>
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <PluginInstall :installed="installed" :installing="installing" @install="installPlugin" />
        <el-form-item :label="t('commons.table.status')">
            <el-switch v-model="form.enabled" />
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.appId')" prop="appId">
            <el-input v-model="form.appId" />
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.appSecret')" prop="clientSecret">
            <el-input v-model="form.clientSecret" type="password" show-password />
        </el-form-item>
        <el-form-item>
            <el-button type="primary" :loading="saving" :disabled="!installed" @click="saveChannel">
                {{ t('commons.button.save') }}
            </el-button>
        </el-form-item>
    </el-form>
    <TaskLog ref="taskLogRef" @close="checkPluginStatus" />
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import type { FormInstance } from 'element-plus';
import { useI18n } from 'vue-i18n';
import { AI } from '@/api/interface/ai';
import { getAgentQQBotConfig, updateAgentQQBotConfig } from '@/api/modules/ai';
import { MsgSuccess } from '@/utils/message';
import { Rules } from '@/global/form-rules';
import TaskLog from '@/components/log/task/index.vue';
import PluginInstall from './components/plugin-install.vue';
import { useAgentPluginChannel } from './useAgentPluginChannel';

type QQBotForm = Omit<AI.AgentQQBotConfig, 'installed'>;

const { t } = useI18n();
const saving = ref(false);
const formRef = ref<FormInstance>();
const { agentId, installed, installing, taskLogRef, checkPluginStatus, loadPlugin, installPlugin } =
    useAgentPluginChannel('qqbot');

const form = reactive<QQBotForm>({
    enabled: true,
    appId: '',
    clientSecret: '',
});

const rules = reactive({
    appId: [Rules.requiredInput],
    clientSecret: [Rules.requiredInput],
});

const load = async (id: number) => {
    await loadPlugin(id);
    const res = await getAgentQQBotConfig({ agentId: id });
    form.enabled = res.data?.enabled ?? true;
    form.appId = res.data?.appId || '';
    form.clientSecret = res.data?.clientSecret || '';
};

const saveChannel = async () => {
    if (!agentId.value || !formRef.value) {
        return;
    }
    await formRef.value.validate();
    saving.value = true;
    try {
        await updateAgentQQBotConfig({
            agentId: agentId.value,
            enabled: form.enabled,
            appId: form.appId,
            clientSecret: form.clientSecret,
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
