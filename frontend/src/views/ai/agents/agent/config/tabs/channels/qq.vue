<template>
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <el-alert
            v-if="!form.installed"
            type="warning"
            :closable="false"
            :title="t('aiTools.agents.pluginNotInstalled')"
            class="mb-4"
        />
        <el-form-item>
            <el-button v-if="!form.installed" type="primary" :loading="installing" @click="installPlugin">
                {{ t('commons.button.install') }}
            </el-button>
        </el-form-item>
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
            <el-button type="primary" :loading="saving" @click="saveChannel">
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
import { checkAgentPlugin, getAgentQQBotConfig, installAgentPlugin, updateAgentQQBotConfig } from '@/api/modules/ai';
import { MsgSuccess } from '@/utils/message';
import { Rules } from '@/global/form-rules';
import { newUUID } from '@/utils/util';
import TaskLog from '@/components/log/task/index.vue';

const { t } = useI18n();
const saving = ref(false);
const installing = ref(false);
const agentId = ref(0);
const formRef = ref<FormInstance>();
const taskLogRef = ref();

const form = reactive<AI.AgentQQBotConfig>({
    enabled: true,
    appId: '',
    clientSecret: '',
    installed: false,
});

const rules = reactive({
    appId: [Rules.requiredInput],
    clientSecret: [Rules.requiredInput],
});

const checkPluginStatus = async () => {
    if (!agentId.value) {
        return;
    }
    const res = await checkAgentPlugin({
        agentId: agentId.value,
        type: 'qqbot',
    });
    form.installed = Boolean(res.data?.installed);
};

const load = async (id: number) => {
    agentId.value = id;
    const res = await getAgentQQBotConfig({ agentId: id });
    Object.assign(form, res.data || {});
    await checkPluginStatus();
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

const installPlugin = async () => {
    if (!agentId.value) {
        return;
    }
    const taskID = newUUID();
    installing.value = true;
    try {
        await installAgentPlugin({
            agentId: agentId.value,
            type: 'qqbot',
            taskID,
        });
        taskLogRef.value?.openWithTaskID(taskID);
    } finally {
        installing.value = false;
    }
};

defineExpose({
    load,
});
</script>
