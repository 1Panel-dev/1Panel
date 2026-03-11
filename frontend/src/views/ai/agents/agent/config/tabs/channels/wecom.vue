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
        <el-form-item :label="t('aiTools.agents.dmPolicy')" prop="dmPolicy">
            <el-select v-model="form.dmPolicy">
                <el-option :label="t('aiTools.agents.policyPairing')" value="pairing" />
                <el-option :label="t('aiTools.agents.policyOpen')" value="open" />
            </el-select>
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.botId')" prop="botId">
            <el-input v-model="form.botId" />
        </el-form-item>
        <el-form-item :label="t('setting.secret')" prop="secret">
            <el-input v-model="form.secret" type="password" show-password />
        </el-form-item>
        <el-form-item>
            <el-button type="primary" :loading="saving" @click="saveChannel">
                {{ t('commons.button.save') }}
            </el-button>
        </el-form-item>

        <el-divider />

        <el-form-item :label="t('aiTools.agents.pairingCode')">
            <el-input v-model="pairingCode" :placeholder="t('aiTools.agents.pairingCodePlaceholder')" />
        </el-form-item>
        <el-form-item>
            <el-button type="primary" :loading="approving" @click="approvePairing">
                {{ t('aiTools.agents.approvePairing') }}
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
import {
    approveAgentChannelPairing,
    checkAgentPlugin,
    getAgentWecomConfig,
    installAgentPlugin,
    updateAgentWecomConfig,
} from '@/api/modules/ai';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { Rules } from '@/global/form-rules';
import { newUUID } from '@/utils/util';
import TaskLog from '@/components/log/task/index.vue';

const { t } = useI18n();
const saving = ref(false);
const approving = ref(false);
const installing = ref(false);
const agentId = ref(0);
const pairingCode = ref('');
const formRef = ref<FormInstance>();
const taskLogRef = ref();

const form = reactive<AI.AgentWecomConfig>({
    enabled: true,
    dmPolicy: 'pairing',
    botId: '',
    secret: '',
    installed: false,
});

const rules = reactive({
    dmPolicy: [Rules.requiredSelect],
    botId: [Rules.requiredInput],
    secret: [Rules.requiredInput],
});

const checkPluginStatus = async () => {
    if (!agentId.value) {
        return;
    }
    const res = await checkAgentPlugin({
        agentId: agentId.value,
        type: 'wecom',
    });
    form.installed = Boolean(res.data?.installed);
};

const load = async (id: number) => {
    agentId.value = id;
    pairingCode.value = '';
    const res = await getAgentWecomConfig({ agentId: id });
    Object.assign(form, res.data || {});
    if (!form.dmPolicy) {
        form.dmPolicy = 'pairing';
    }
    await checkPluginStatus();
};

const saveChannel = async () => {
    if (!agentId.value || !formRef.value) {
        return;
    }
    await formRef.value.validate();
    saving.value = true;
    try {
        await updateAgentWecomConfig({
            agentId: agentId.value,
            enabled: form.enabled,
            dmPolicy: form.dmPolicy,
            botId: form.botId,
            secret: form.secret,
        });
        MsgSuccess(t('aiTools.agents.saveSuccess'));
    } finally {
        saving.value = false;
    }
};

const approvePairing = async () => {
    if (!agentId.value) {
        return;
    }
    if (!pairingCode.value) {
        MsgWarning(t('aiTools.agents.pairingCodeRequired'));
        return;
    }
    approving.value = true;
    try {
        await approveAgentChannelPairing({
            agentId: agentId.value,
            type: 'wecom',
            pairingCode: pairingCode.value,
        });
        MsgSuccess(t('aiTools.agents.pairingApproveSuccess'));
        pairingCode.value = '';
    } finally {
        approving.value = false;
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
            type: 'wecom',
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
