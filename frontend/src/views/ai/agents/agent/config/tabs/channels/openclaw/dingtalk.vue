<template>
    <VersionSupport v-if="!supported" :min-version="openclawMinSupportedVersion" />
    <el-form v-else ref="formRef" v-loading="loading" :model="form" :rules="rules" label-position="top">
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
            <div>
                <el-form-item :label="t('aiTools.agents.dmPolicy')" prop="dmPolicy">
                    <el-select v-model="form.dmPolicy">
                        <el-option :label="t('aiTools.agents.policyAllowlist')" value="allowlist" />
                        <el-option :label="t('aiTools.agents.policyOpen')" value="open" />
                        <el-option :label="t('aiTools.agents.policyDisabled')" value="disabled" />
                    </el-select>
                </el-form-item>
                <el-form-item
                    v-if="form.dmPolicy === 'allowlist'"
                    :label="t('aiTools.agents.allowFrom')"
                    prop="allowFromText"
                >
                    <el-input
                        v-model="form.allowFromText"
                        type="textarea"
                        :rows="3"
                        :placeholder="t('aiTools.agents.allowFromPlaceholder')"
                    />
                    <span class="input-help">{{ t('aiTools.agents.allowFromHelper') }}</span>
                </el-form-item>
                <el-form-item :label="t('aiTools.agents.groupPolicy')" prop="groupPolicy">
                    <el-select v-model="form.groupPolicy">
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
                        type="textarea"
                        :rows="3"
                        :placeholder="t('aiTools.agents.groupAllowFromPlaceholder')"
                    />
                    <span class="input-help">{{ t('aiTools.agents.groupAllowFromHelper') }}</span>
                </el-form-item>
                <el-form-item :label="t('aiTools.agents.separateSessionByConversation')">
                    <el-switch v-model="form.separateSessionByConversation" />
                </el-form-item>
                <el-form-item :label="t('aiTools.agents.groupSessionScope')" prop="groupSessionScope">
                    <el-select v-model="form.groupSessionScope">
                        <el-option :label="t('aiTools.agents.groupSessionScopeGroup')" value="group" />
                        <el-option :label="t('aiTools.agents.groupSessionScopeGroupSender')" value="group_sender" />
                    </el-select>
                </el-form-item>
                <el-form-item :label="t('aiTools.agents.sharedMemoryAcrossConversations')">
                    <el-switch v-model="form.sharedMemoryAcrossConversations" />
                </el-form-item>
                <el-form-item :label="t('aiTools.agents.asyncMode')">
                    <el-switch v-model="form.asyncMode" />
                </el-form-item>
                <el-form-item v-if="form.asyncMode" :label="t('aiTools.agents.ackText')">
                    <el-input v-model="form.ackText" />
                </el-form-item>
            </div>
            <ChannelBots
                :bots="form.bots"
                :fields="botFields"
                :create-bot="createBot"
                :show-name-field="false"
                summary-label="Client ID"
                unique-field-prop="clientId"
                unique-field-label="Client ID"
                :summary-formatter="getBotSummary"
                :add-disabled="!installed"
                :disabled="!installed"
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
import { computed, reactive, ref } from 'vue';
import type { FormInstance, FormRules } from 'element-plus';
import { useI18n } from 'vue-i18n';
import { AI } from '@/api/interface/ai';
import { getAgentDingTalkConfig, updateAgentDingTalkConfig } from '@/api/modules/ai';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { Rules } from '@/global/form-rules';
import { isOpenclawCurrentHTTPVersion } from '@/utils/agent';
import PluginInstall from '../components/plugin-install.vue';
import VersionSupport from '../../components/version-support.vue';
import { useAgentPluginChannel } from './useAgentPluginChannel';
import ChannelBots from '../components/channel-bots.vue';

interface DingTalkForm extends Omit<AI.AgentDingTalkConfig, 'installed' | 'allowFrom' | 'groupAllowFrom'> {
    allowFromText: string;
    groupAllowFromText: string;
}
type BotField = {
    prop: string;
    label: string;
    type?: 'text' | 'password';
    required?: boolean;
};

const openclawMinSupportedVersion = '2026.3.23';
const props = defineProps<{
    appVersion: string;
}>();

const { t } = useI18n();
const saving = ref(false);
const agentId = ref(0);
const formRef = ref<FormInstance>();
const {
    installed,
    loading,
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
} = useAgentPluginChannel('dingtalk');
const supported = computed(() => isOpenclawCurrentHTTPVersion(props.appVersion));
const defaultAckText = t('aiTools.agents.ackTextDefault');

const form = reactive<DingTalkForm>({
    enabled: true,
    dmPolicy: 'open',
    groupPolicy: 'disabled',
    separateSessionByConversation: true,
    groupSessionScope: 'group',
    sharedMemoryAcrossConversations: false,
    asyncMode: false,
    ackText: defaultAckText,
    bots: [],
    allowFromText: '',
    groupAllowFromText: '',
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

const validateAllowFrom = (_rule: any, value: string, callback: (error?: Error) => void) => {
    if (form.dmPolicy !== 'allowlist') {
        callback();
        return;
    }
    if (parseTextList(value).length === 0) {
        callback(new Error(t('aiTools.agents.allowFromRequired')));
        return;
    }
    callback();
};

const validateGroupAllowFrom = (_rule: any, value: string, callback: (error?: Error) => void) => {
    if (form.groupPolicy !== 'allowlist') {
        callback();
        return;
    }
    if (parseTextList(value).length === 0) {
        callback(new Error(t('aiTools.agents.allowFromRequired')));
        return;
    }
    callback();
};

const rules = reactive<FormRules>({
    dmPolicy: [Rules.requiredSelect],
    groupPolicy: [Rules.requiredSelect],
    groupSessionScope: [Rules.requiredSelect],
    allowFromText: [{ validator: validateAllowFrom, trigger: 'blur' }],
    groupAllowFromText: [{ validator: validateGroupAllowFrom, trigger: 'blur' }],
});

const botFields: BotField[] = [
    { prop: 'clientId', label: 'Client ID', required: true },
    { prop: 'clientSecret', label: 'Client Secret', type: 'password', required: true },
];

const createBot = (): AI.AgentDingTalkBot => ({
    accountId: '',
    name: '',
    enabled: true,
    isDefault: false,
    clientId: '',
    clientSecret: '',
});

const getBotSummary = (bot: AI.AgentDingTalkBot) => {
    return bot.clientId;
};

const updateBots = (bots: AI.AgentDingTalkBot[]) => {
    form.bots = bots;
};

const load = async (id: number) => {
    if (!supported.value) {
        return;
    }
    agentId.value = id;
    await loadPlugin(id);
    const res = await getAgentDingTalkConfig({ agentId: id });
    form.enabled = res.data?.enabled ?? true;
    form.dmPolicy = res.data?.dmPolicy || 'open';
    form.groupPolicy = res.data?.groupPolicy || 'disabled';
    form.separateSessionByConversation = res.data?.separateSessionByConversation ?? true;
    form.groupSessionScope = res.data?.groupSessionScope || 'group';
    form.sharedMemoryAcrossConversations = res.data?.sharedMemoryAcrossConversations ?? false;
    form.asyncMode = res.data?.asyncMode ?? false;
    form.ackText = res.data?.ackText || defaultAckText;
    form.bots = res.data?.bots || [];
    form.allowFromText = (res.data?.allowFrom || []).join('\n');
    form.groupAllowFromText = (res.data?.groupAllowFrom || []).join('\n');
};

const reload = async () => {
    if (!agentId.value) {
        return;
    }
    await load(agentId.value);
};

const saveChannel = async (action: 'delete' | 'save' = 'save') => {
    if (!supported.value || !agentId.value || !formRef.value) {
        return;
    }
    if (form.bots.length === 0) {
        MsgWarning(t('aiTools.agents.botRequired'));
        return;
    }
    await formRef.value.validate();
    saving.value = true;
    try {
        await updateAgentDingTalkConfig({
            agentId: agentId.value,
            enabled: form.enabled,
            dmPolicy: form.dmPolicy,
            allowFrom: parseTextList(form.allowFromText),
            groupPolicy: form.groupPolicy,
            groupAllowFrom: parseTextList(form.groupAllowFromText),
            separateSessionByConversation: form.separateSessionByConversation,
            groupSessionScope: form.groupSessionScope,
            sharedMemoryAcrossConversations: form.sharedMemoryAcrossConversations,
            asyncMode: form.asyncMode,
            ackText: form.ackText,
            bots: form.bots,
        });
        MsgSuccess(action === 'delete' ? t('commons.msg.deleteSuccess') : t('aiTools.agents.saveSuccess'));
    } catch (error) {
        await reload();
        throw error;
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
