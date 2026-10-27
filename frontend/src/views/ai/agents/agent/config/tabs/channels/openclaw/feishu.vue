<template>
    <el-form
        ref="formRef"
        v-loading="loading || pluginLoading || approving"
        :model="form"
        :rules="rules"
        label-position="top"
    >
        <PluginInstall
            v-if="!builtinManaged"
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
        <el-alert
            v-else
            :closable="false"
            :type="!installed || legacy ? 'warning' : 'info'"
            :title="t('aiTools.agents.feishuBuiltin')"
            :description="
                t(
                    !installed
                        ? 'aiTools.agents.feishuImageRequired'
                        : legacy
                          ? legacyConfigured
                              ? 'aiTools.agents.feishuLegacy'
                              : 'aiTools.agents.feishuRecoverConfirm'
                          : 'aiTools.agents.feishuBuiltinHint',
                )
            "
        />
        <template v-if="installed">
            <el-form-item :label="t('commons.table.status')" class="mt-4">
                <el-switch v-model="form.enabled" :disabled="legacy" />
            </el-form-item>
            <div v-if="!legacy || legacyConfigured">
                <FeishuTransition
                    v-if="builtinManaged"
                    v-model:streaming="form.streaming"
                    v-model:group-session-scope="form.groupSessionScope"
                    v-model:require-mention="form.requireMention"
                    :allow-inherit="!legacy"
                />
                <template v-else>
                    <el-form-item :label="t('aiTools.agents.threadSession')">
                        <el-switch v-model="form.threadSession" />
                    </el-form-item>
                    <el-form-item :label="t('aiTools.agents.streaming')">
                        <el-switch v-model="form.streaming" />
                    </el-form-item>
                    <el-form-item :label="t('aiTools.agents.replyMode')" prop="replyMode">
                        <el-input v-model="form.replyMode" placeholder="auto" />
                    </el-form-item>
                    <el-form-item :label="t('aiTools.agents.requireMention')" prop="requireMention">
                        <el-select v-model="form.requireMention">
                            <el-option :label="t('aiTools.agents.requireMentionTrue')" value="true" />
                            <el-option :label="t('aiTools.agents.requireMentionFalse')" value="false" />
                            <el-option :label="t('aiTools.agents.requireMentionOpen')" value="open" />
                        </el-select>
                    </el-form-item>
                </template>
                <el-form-item :label="t('aiTools.agents.groupPolicy')" prop="groupPolicy">
                    <el-select v-model="form.groupPolicy" :disabled="legacy">
                        <el-option :label="t('aiTools.agents.policyOpen')" value="open" />
                        <el-option :label="t('aiTools.agents.policyAllowlist')" value="allowlist" />
                        <el-option :label="t('aiTools.agents.policyDisabled')" value="disabled" />
                    </el-select>
                </el-form-item>
                <el-form-item
                    v-if="form.groupPolicy === 'allowlist'"
                    :label="t('aiTools.agents.groupAllowFrom')"
                    prop="groupAllowFromText"
                >
                    <el-input
                        v-model="form.groupAllowFromText"
                        :disabled="legacy"
                        type="textarea"
                        :rows="3"
                        :placeholder="t('aiTools.agents.groupAllowFromPlaceholder')"
                    />
                    <span class="input-help">
                        {{
                            t('aiTools.agents.allowlistPolicyHelper', [
                                t('aiTools.agents.groupAllowFromPlaceholder'),
                                t('aiTools.agents.groupPolicy'),
                            ])
                        }}
                    </span>
                </el-form-item>
            </div>
            <ChannelBots
                v-if="!legacy || legacyConfigured"
                :bots="form.bots"
                :fields="botFields"
                :create-bot="createBot"
                :create-preset="defaultBotCreatePreset"
                summary-label="App ID"
                unique-field-prop="appId"
                unique-field-label="App ID"
                :summary-formatter="getBotSummary"
                :add-disabled="legacy"
                :disabled="legacy"
                approvable
                :approve-disabled="isApproveDisabled"
                :fixed-account-ids="['default']"
                :undeletable-account-ids="['default']"
                @update:bots="updateBots"
                @save="saveChannel"
                @approve="approvePairing"
            />
            <el-form-item v-if="!legacy" class="mt-4">
                <el-button v-permission type="primary" :loading="saving" :disabled="legacy" @click="saveChannel">
                    {{ t('commons.button.save') }}
                </el-button>
            </el-form-item>
            <el-form-item v-else class="mt-4">
                <el-button v-permission type="primary" :loading="switching" @click="switchPlugin">
                    {{ t(legacyConfigured ? 'aiTools.agents.feishuSwitch' : 'aiTools.agents.feishuRecover') }}
                </el-button>
            </el-form-item>
        </template>
        <TaskLog ref="taskLogRef" @close="reload" />
    </el-form>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue';
import type { FormInstance } from 'element-plus';
import { ElMessageBox } from 'element-plus';
import { useI18n } from 'vue-i18n';
import { AI } from '@/api/interface/ai';
import {
    approveAgentChannelPairing,
    getAgentFeishuConfig,
    updateAgentFeishuConfig,
    switchAgentFeishuPlugin,
} from '@/api/modules/ai';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { Rules } from '@/global/form-rules';
import TaskLog from '@/components/log/task/index.vue';
import { newUUID } from '@/utils/id';
import PluginInstall from '../components/plugin-install.vue';
import { useAgentPluginChannel } from './useAgentPluginChannel';
import ChannelBots from '../components/channel-bots.vue';
import FeishuTransition from '../components/feishu-transition.vue';

const { t } = useI18n();

type BotField = {
    prop: string;
    label: string;
    type?: 'text' | 'password' | 'select' | 'textarea';
    required?: boolean;
    placeholder?: string;
    options?: Array<{
        label: string;
        value: string;
    }>;
};

interface FeishuBotForm extends AI.AgentFeishuBot {
    allowFromText: string;
}

interface FeishuForm extends Omit<AI.AgentFeishuConfig, 'bots'> {
    groupAllowFromText: string;
    bots: FeishuBotForm[];
}

const saving = ref(false);
const approving = ref(false);
const formRef = ref<FormInstance>();
const loading = ref(false);
const {
    agentId,
    loading: pluginLoading,
    installed: pluginInstalled,
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
} = useAgentPluginChannel('feishu');
const builtinManaged = ref(true);
const builtinAvailable = ref(false);
const installed = computed(() => (builtinManaged.value ? builtinAvailable.value : pluginInstalled.value));
const legacy = ref(false);
const legacyConfigured = ref(false);
const switching = ref(false);
const taskLogRef = ref();

const form = reactive<FeishuForm>({
    enabled: true,
    threadSession: true,
    streaming: false,
    replyMode: 'auto',
    groupSessionScope: 'group',
    requireMention: 'true',
    groupPolicy: 'allowlist',
    groupAllowFrom: [],
    groupAllowFromText: '',
    bots: [],
    installed: false,
});
const defaultBotCreatePreset = computed(() => {
    if (form.bots.length > 0) {
        return undefined;
    }
    return {
        accountId: 'default',
        name: 'Default',
        isDefault: true,
    };
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

const rules = reactive({
    replyMode: [Rules.requiredInput],
    groupSessionScope: [Rules.requiredSelect],
    requireMention: [Rules.requiredSelect],
    groupPolicy: [Rules.requiredSelect],
    groupAllowFromText: [
        {
            validator: (_rule, value, callback) => {
                if (
                    !legacy.value &&
                    form.groupPolicy === 'allowlist' &&
                    parseTextList(String(value || '')).length === 0
                ) {
                    callback(new Error(t('aiTools.agents.allowFromRequired')));
                    return;
                }
                callback();
            },
            trigger: 'blur',
        },
    ],
});

const botFields: BotField[] = [
    { prop: 'appId', label: 'App ID', required: true },
    { prop: 'appSecret', label: 'App Secret', type: 'password', required: true },
    {
        prop: 'dmPolicy',
        label: t('aiTools.agents.dmPolicy'),
        type: 'select',
        required: true,
        options: [
            { label: t('aiTools.agents.pairingCode'), value: 'pairing' },
            { label: t('aiTools.agents.policyOpen'), value: 'open' },
            { label: t('aiTools.agents.policyAllowlist'), value: 'allowlist' },
            { label: t('aiTools.agents.policyDisabled'), value: 'disabled' },
        ],
    },
    {
        prop: 'allowFromText',
        label: t('aiTools.agents.allowFrom'),
        type: 'textarea',
        placeholder: t('aiTools.agents.allowFromPlaceholder'),
    },
];

const createBot = (): FeishuBotForm => ({
    accountId: '',
    name: '',
    enabled: true,
    isDefault: false,
    appId: '',
    appSecret: '',
    dmPolicy: 'pairing',
    allowFrom: [],
    allowFromText: '',
});

const normalizeBot = (bot?: Partial<AI.AgentFeishuBot>): FeishuBotForm => ({
    accountId: bot?.accountId || '',
    name: bot?.name || bot?.accountId || '',
    enabled: bot?.enabled ?? true,
    isDefault: bot?.isDefault ?? false,
    appId: bot?.appId || '',
    appSecret: bot?.appSecret || '',
    dmPolicy: bot?.dmPolicy || 'pairing',
    allowFrom: bot?.allowFrom || [],
    allowFromText: (bot?.allowFrom || []).join('\n'),
});

const getBotSummary = (bot: FeishuBotForm) => bot.appId;

const updateBots = (bots: FeishuBotForm[]) => {
    form.bots = bots;
};

const isApproveDisabled = (bot: FeishuBotForm) => legacy.value || bot.dmPolicy !== 'pairing';

const load = async (id: number) => {
    agentId.value = id;
    loading.value = true;
    try {
        const res = await getAgentFeishuConfig({ agentId: id });
        builtinManaged.value = Boolean(res.data?.builtinManaged);
        builtinAvailable.value = Boolean(res.data?.builtinAvailable);
        legacy.value = builtinManaged.value && Boolean(res.data?.legacy);
        legacyConfigured.value = Boolean(res.data?.legacyConfigured);
        if (!builtinManaged.value) {
            await loadPlugin(id);
        }
        form.groupSessionScope = res.data?.groupSessionScope || 'group';
        form.enabled = res.data?.enabled ?? true;
        form.threadSession = res.data?.threadSession ?? true;
        form.streaming = res.data?.streaming ?? false;
        form.replyMode = res.data?.replyMode || 'auto';
        form.requireMention = res.data?.requireMention || 'true';
        form.groupPolicy = res.data?.groupPolicy || (builtinManaged.value ? 'allowlist' : 'open');
        form.groupAllowFrom = res.data?.groupAllowFrom || [];
        form.groupAllowFromText = (res.data?.groupAllowFrom || []).join('\n');
        form.bots = (res.data?.bots || []).map((bot) => normalizeBot(bot));
    } finally {
        loading.value = false;
    }
};

const reload = async () => {
    if (!agentId.value) {
        return;
    }
    await load(agentId.value);
};

const buildPayloadBots = (): AI.AgentFeishuBot[] => {
    return form.bots.map((bot) => ({
        accountId: bot.accountId,
        name: bot.name,
        enabled: bot.enabled,
        isDefault: bot.isDefault,
        appId: bot.appId,
        appSecret: bot.appSecret,
        dmPolicy: bot.dmPolicy,
        allowFrom: parseTextList(bot.allowFromText),
    }));
};

const validateBots = () => {
    for (const bot of form.bots) {
        if (bot.dmPolicy === 'allowlist' && parseTextList(bot.allowFromText).length === 0) {
            MsgWarning(t('aiTools.agents.allowFromRequired'));
            return false;
        }
    }
    return true;
};

const saveChannel = async (action: 'delete' | 'save' = 'save') => {
    if (!agentId.value || !formRef.value) {
        return;
    }
    if (form.bots.length === 0) {
        MsgWarning(t('aiTools.agents.botRequired'));
        return;
    }
    if (!validateBots()) {
        return;
    }
    await formRef.value.validate();
    saving.value = true;
    try {
        await updateAgentFeishuConfig({
            agentId: agentId.value,
            enabled: form.enabled,
            threadSession: form.threadSession,
            ...(builtinManaged.value ? { groupSessionScope: form.groupSessionScope } : {}),
            streaming: form.streaming,
            replyMode: form.replyMode,
            requireMention: form.requireMention,
            groupPolicy: form.groupPolicy,
            groupAllowFrom: parseTextList(form.groupAllowFromText),
            bots: buildPayloadBots(),
        });
        MsgSuccess(action === 'delete' ? t('commons.msg.deleteSuccess') : t('aiTools.agents.saveSuccess'));
    } catch (error) {
        await reload();
        throw error;
    } finally {
        saving.value = false;
    }
};

const switchPlugin = async () => {
    if (!agentId.value || switching.value || !formRef.value) return;
    if (legacyConfigured.value) await formRef.value.validateField(['groupSessionScope', 'requireMention']);
    if (legacyConfigured.value && form.requireMention === 'open') {
        MsgWarning(t('aiTools.agents.feishuMentionChoice'));
        return;
    }
    await ElMessageBox.confirm(
        legacyConfigured.value
            ? t('aiTools.agents.feishuSwitchConfirm') + ' ' + t('aiTools.agents.feishuRecoveryNotice')
            : t('aiTools.agents.feishuRecoverConfirm'),
        t('aiTools.agents.feishuSwitch'),
        {
            type: 'warning',
            confirmButtonText: t('commons.button.confirm'),
            cancelButtonText: t('commons.button.cancel'),
        },
    );
    switching.value = true;
    try {
        const taskID = newUUID();
        await switchAgentFeishuPlugin({
            agentId: agentId.value,
            taskID,
            streaming: form.streaming,
            groupSessionScope: form.groupSessionScope || 'group',
            requireMention: legacyConfigured.value && form.requireMention === 'false' ? 'false' : 'true',
        });
        taskLogRef.value?.openWithTaskID(taskID);
    } finally {
        switching.value = false;
    }
};

const approvePairing = async (bot: FeishuBotForm) => {
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

<style lang="scss" scoped>
.input-help {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 1.5;
    margin-top: 4px;
}
</style>
