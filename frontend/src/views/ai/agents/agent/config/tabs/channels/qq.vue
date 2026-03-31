<template>
    <el-form v-loading="loading" :model="form" label-position="top">
        <PluginInstall
            :installed="installed"
            :installing="installing"
            :upgrading="upgrading"
            :uninstalling="uninstalling"
            :current-version="currentVersion"
            :latest-version="latestVersion"
            :upgradable="upgradable"
            :install-action="installPlugin"
            :upgrade-action="upgradePlugin"
            :uninstall-action="uninstallPlugin"
            :on-task-close="reload"
        />
        <template v-if="installed">
            <el-form-item :label="t('commons.table.status')" class="mt-4">
                <el-switch v-model="form.enabled" />
            </el-form-item>
            <ChannelBots
                :bots="form.bots"
                :fields="botFields"
                :create-bot="createBot"
                summary-label="App ID"
                :summary-formatter="getBotSummary"
                :add-disabled="!installed"
                :disabled="!installed"
                :fixed-account-ids="['default']"
                :undeletable-account-ids="['default']"
                @update:bots="updateBots"
                @save="saveChannel"
            />
            <el-form-item class="mt-4">
                <el-button type="primary" :loading="saving" :disabled="!installed" @click="saveChannel">
                    {{ t('commons.button.save') }}
                </el-button>
            </el-form-item>
        </template>
    </el-form>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { AI } from '@/api/interface/ai';
import { getAgentQQBotConfig, updateAgentQQBotConfig } from '@/api/modules/ai';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import PluginInstall from './components/plugin-install.vue';
import { useAgentPluginChannel } from './useAgentPluginChannel';
import ChannelBots from './components/channel-bots.vue';

type QQBotForm = Pick<AI.AgentQQBotConfig, 'enabled' | 'bots'>;
type BotField = {
    prop: string;
    label: string;
    type?: 'text' | 'password';
    required?: boolean;
};

const { t } = useI18n();
const saving = ref(false);
const {
    agentId,
    loading,
    installed,
    installing,
    upgrading,
    uninstalling,
    currentVersion,
    latestVersion,
    upgradable,
    loadPlugin,
    installPlugin,
    upgradePlugin,
    uninstallPlugin,
} = useAgentPluginChannel('qqbot');

const form = reactive<QQBotForm>({
    enabled: true,
    bots: [],
});

const botFields: BotField[] = [
    { prop: 'appId', label: 'App ID', required: true },
    { prop: 'clientSecret', label: 'App Secret', type: 'password', required: true },
];

const createBot = (): AI.AgentQQBotBot => ({
    accountId: '',
    name: '',
    enabled: true,
    isDefault: false,
    appId: '',
    clientSecret: '',
});

const getBotSummary = (bot: AI.AgentQQBotBot) => {
    return bot.appId;
};

const updateBots = (bots: AI.AgentQQBotBot[]) => {
    form.bots = bots;
};

const load = async (id: number) => {
    await loadPlugin(id);
    const res = await getAgentQQBotConfig({ agentId: id });
    form.enabled = res.data?.enabled ?? true;
    form.bots = res.data?.bots || [];
};

const reload = async () => {
    if (!agentId.value) {
        return;
    }
    await load(agentId.value);
};

const saveChannel = async () => {
    if (!agentId.value) {
        return;
    }
    if (form.bots.length === 0) {
        MsgWarning(t('aiTools.agents.botRequired'));
        return;
    }
    saving.value = true;
    try {
        await updateAgentQQBotConfig({
            agentId: agentId.value,
            enabled: form.enabled,
            bots: form.bots,
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
