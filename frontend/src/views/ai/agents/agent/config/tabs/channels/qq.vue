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

type BotField = {
    prop: string;
    label: string;
    type?: 'text' | 'password' | 'textarea';
    required?: boolean;
    placeholder?: string;
};

interface QQBotFormItem extends AI.AgentQQBotBot {
    allowFromText: string;
}

interface QQBotForm {
    enabled: boolean;
    bots: QQBotFormItem[];
}

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

const parseTextList = (value: string): string[] => {
    return Array.from(
        new Set(
            String(value || '')
                .split(/\r?\n/)
                .map((item) => item.trim())
                .filter(Boolean),
        ),
    );
};

const botFields: BotField[] = [
    { prop: 'appId', label: 'App ID', required: true },
    { prop: 'clientSecret', label: 'App Secret', type: 'password', required: true },
    {
        prop: 'allowFromText',
        label: t('aiTools.agents.allowFrom'),
        type: 'textarea',
        placeholder: t('aiTools.agents.allowFromPlaceholder'),
    },
    {
        prop: 'systemPrompt',
        label: t('aiTools.agents.systemPrompt'),
        type: 'textarea',
    },
];

const createBot = (): QQBotFormItem => ({
    accountId: '',
    name: '',
    enabled: true,
    isDefault: false,
    appId: '',
    clientSecret: '',
    allowFrom: [],
    allowFromText: '',
    systemPrompt: '',
});

const normalizeBot = (bot?: Partial<AI.AgentQQBotBot>): QQBotFormItem => ({
    accountId: bot?.accountId || '',
    name: bot?.name || bot?.accountId || '',
    enabled: bot?.enabled ?? true,
    isDefault: bot?.isDefault ?? false,
    appId: bot?.appId || '',
    clientSecret: bot?.clientSecret || '',
    allowFrom: bot?.allowFrom || [],
    allowFromText: (bot?.allowFrom || []).join('\n'),
    systemPrompt: bot?.systemPrompt || '',
});

const getBotSummary = (bot: QQBotFormItem) => {
    return bot.appId;
};

const updateBots = (bots: QQBotFormItem[]) => {
    form.bots = bots;
};

const load = async (id: number) => {
    await loadPlugin(id);
    const res = await getAgentQQBotConfig({ agentId: id });
    form.enabled = res.data?.enabled ?? true;
    form.bots = (res.data?.bots || []).map((bot) => normalizeBot(bot));
};

const reload = async () => {
    if (!agentId.value) {
        return;
    }
    await load(agentId.value);
};

const saveChannel = async (action: 'delete' | 'save' = 'save') => {
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
            bots: form.bots.map((bot) => ({
                accountId: bot.accountId,
                name: bot.name,
                enabled: bot.enabled,
                isDefault: bot.isDefault,
                appId: bot.appId,
                clientSecret: bot.clientSecret,
                allowFrom: parseTextList(bot.allowFromText),
                systemPrompt: bot.systemPrompt,
            })),
        });
        MsgSuccess(action === 'delete' ? t('commons.msg.deleteSuccess') : t('aiTools.agents.saveSuccess'));
    } finally {
        saving.value = false;
    }
};

defineExpose({
    load,
});
</script>
