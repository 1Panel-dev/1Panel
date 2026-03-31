<template>
    <el-form ref="formRef" v-loading="approving" :model="form" :rules="rules" label-position="top">
        <PluginInstall :installed="installed" :installing="installing" @install="installPlugin" />
        <el-form-item :label="t('commons.table.status')">
            <el-switch v-model="form.enabled" />
        </el-form-item>
        <el-form-item>
            <el-link type="primary" icon="Position" @click="toFeishuDoc">
                {{ t('container.mirrorsHelper2') }}
            </el-link>
        </el-form-item>
        <ChannelBots
            :bots="form.bots"
            :fields="botFields"
            :create-bot="createBot"
            summary-label="App ID"
            :summary-formatter="getBotSummary"
            :add-disabled="!installed"
            approvable
            :fixed-account-ids="['default']"
            :undeletable-account-ids="['default']"
            @update:bots="updateBots"
            @save="saveChannel"
            @approve="approvePairing"
        />
        <el-form-item class="mt-4">
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
import { ElMessageBox } from 'element-plus';
import { useI18n } from 'vue-i18n';
import { AI } from '@/api/interface/ai';
import { approveAgentChannelPairing, getAgentFeishuConfig, updateAgentFeishuConfig } from '@/api/modules/ai';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import TaskLog from '@/components/log/task/index.vue';
import PluginInstall from './components/plugin-install.vue';
import { useAgentPluginChannel } from './useAgentPluginChannel';
import ChannelBots from './components/channel-bots.vue';

const { t } = useI18n();
const saving = ref(false);
const approving = ref(false);
const formRef = ref<FormInstance>();
const { agentId, installed, installing, taskLogRef, checkPluginStatus, loadPlugin, installPlugin } =
    useAgentPluginChannel('feishu');

const form = reactive<AI.AgentFeishuConfig>({
    enabled: true,
    bots: [],
    installed: false,
});

const rules = reactive({});

const botFields = [
    { prop: 'appId', label: 'App ID', required: true },
    { prop: 'appSecret', label: 'App Secret', type: 'password', required: true },
];

const createBot = (): AI.AgentFeishuBot => ({
    accountId: '',
    name: '',
    enabled: true,
    isDefault: false,
    appId: '',
    appSecret: '',
});

const getBotSummary = (bot: AI.AgentFeishuBot) => {
    return bot.appId;
};

const updateBots = (bots: AI.AgentFeishuBot[]) => {
    form.bots = bots;
};

const toFeishuDoc = () => {
    window.open('https://bytedance.larkoffice.com/docx/MFK7dDFLFoVlOGxWCv5cTXKmnMh', '_blank');
};

const load = async (id: number) => {
    await loadPlugin(id);
    const res = await getAgentFeishuConfig({ agentId: id });
    form.enabled = res.data?.enabled ?? true;
    form.bots = res.data?.bots || [];
};

const saveChannel = async () => {
    if (!agentId.value || !formRef.value) {
        return;
    }
    if (form.bots.length === 0) {
        MsgWarning(t('aiTools.agents.botRequired'));
        return;
    }
    await formRef.value.validate();
    saving.value = true;
    try {
        await updateAgentFeishuConfig({
            agentId: agentId.value,
            enabled: form.enabled,
            bots: form.bots,
        });
        MsgSuccess(t('aiTools.agents.saveSuccess'));
    } finally {
        saving.value = false;
    }
};

const approvePairing = async (bot: AI.AgentFeishuBot) => {
    if (!agentId.value) {
        return;
    }
    try {
        const res = await ElMessageBox.prompt(
            t('aiTools.agents.pairingCodePlaceholder'),
            t('aiTools.agents.pairingCode'),
            {
                confirmButtonText: t('commons.button.confirm'),
                cancelButtonText: t('commons.button.cancel'),
                inputPlaceholder: t('aiTools.agents.pairingCodePlaceholder'),
            },
        );
        approving.value = true;
        await approveAgentChannelPairing({
            agentId: agentId.value,
            type: 'feishu',
            accountId: bot.accountId === 'default' ? '' : bot.accountId,
            pairingCode: res.value,
        });
        MsgSuccess(t('aiTools.agents.pairingApproveSuccess'));
    } catch (error) {
        return;
    } finally {
        approving.value = false;
    }
};

defineExpose({
    load,
});
</script>
