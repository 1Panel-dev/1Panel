<template>
    <el-form ref="formRef" v-loading="approving" :model="form" :rules="rules" label-position="top">
        <el-form-item :label="t('commons.table.status')">
            <el-switch v-model="form.enabled" />
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.dmPolicy')" prop="dmPolicy">
            <el-select v-model="form.dmPolicy">
                <el-option :label="t('aiTools.agents.pairingCode')" value="pairing" />
                <el-option :label="t('aiTools.agents.policyOpen')" value="open" />
                <el-option :label="t('aiTools.agents.policyAllowlist')" value="allowlist" />
                <el-option :label="t('aiTools.agents.policyDisabled')" value="disabled" />
            </el-select>
        </el-form-item>
        <el-form-item v-if="form.dmPolicy === 'allowlist'" :label="t('aiTools.agents.allowFrom')" prop="allowFromText">
            <el-input
                v-model="form.allowFromText"
                type="textarea"
                :rows="3"
                :placeholder="t('aiTools.agents.allowFromPlaceholder')"
            />
            <span class="input-help">
                {{
                    t('aiTools.agents.allowlistPolicyHelper', [
                        t('aiTools.agents.allowFromPlaceholder'),
                        t('aiTools.agents.dmPolicy'),
                    ])
                }}
            </span>
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
            <span class="input-help">
                {{
                    t('aiTools.agents.allowlistPolicyHelper', [
                        t('aiTools.agents.groupAllowFromPlaceholder'),
                        t('aiTools.agents.groupPolicy'),
                    ])
                }}
            </span>
        </el-form-item>
        <el-form-item :label="t('setting.proxy')">
            <el-input v-model="form.proxy" placeholder="http://127.0.0.1:7890" />
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.streaming')" prop="streaming">
            <el-select v-model="form.streaming">
                <el-option :label="t('aiTools.agents.streamingOff')" value="off" />
                <el-option :label="t('aiTools.agents.streamingPartial')" value="partial" />
                <el-option :label="t('aiTools.agents.streamingBlock')" value="block" />
                <el-option :label="t('aiTools.agents.streamingProgress')" value="progress" />
            </el-select>
        </el-form-item>
        <ChannelBots
            :bots="form.bots"
            :fields="botFields"
            :create-bot="createBot"
            summary-label="Token"
            unique-field-prop="botToken"
            unique-field-label="Bot Token"
            :summary-formatter="getBotSummary"
            defaultable
            approvable
            :approve-disabled="isApproveDisabled"
            @update:bots="updateBots"
            @save="saveChannel"
            @approve="approvePairing"
        />
        <el-form-item class="mt-4">
            <el-button v-permission type="primary" :loading="saving" @click="saveChannel">
                {{ t('commons.button.save') }}
            </el-button>
        </el-form-item>
    </el-form>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import type { FormInstance } from 'element-plus';
import { ElMessageBox } from 'element-plus';
import { useI18n } from 'vue-i18n';
import { AI } from '@/api/interface/ai';
import { approveAgentChannelPairing, getAgentTelegramConfig, updateAgentTelegramConfig } from '@/api/modules/ai';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { Rules } from '@/global/form-rules';
import ChannelBots from '../components/channel-bots.vue';

const { t } = useI18n();

type BotField = {
    prop: string;
    label: string;
    type?: 'text' | 'password' | 'select';
    required?: boolean;
    options?: Array<{
        label: string;
        value: string;
    }>;
};

interface TelegramForm extends Omit<AI.AgentTelegramConfig, 'allowFrom' | 'groupAllowFrom'> {
    allowFromText: string;
    groupAllowFromText: string;
}

const saving = ref(false);
const approving = ref(false);
const agentId = ref(0);
const formRef = ref<FormInstance>();

const form = reactive<TelegramForm>({
    enabled: true,
    dmPolicy: 'pairing',
    allowFromText: '',
    requireMention: false,
    groupPolicy: 'open',
    groupAllowFromText: '',
    proxy: '',
    streaming: 'partial',
    defaultAccount: '',
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

const rules = reactive({
    dmPolicy: [Rules.requiredSelect],
    groupPolicy: [Rules.requiredSelect],
    allowFromText: [{ validator: validateAllowFrom, trigger: 'blur' }],
    groupAllowFromText: [{ validator: validateGroupAllowFrom, trigger: 'blur' }],
    streaming: [Rules.requiredSelect],
});

const botFields: BotField[] = [
    { prop: 'botToken', label: 'Bot Token', type: 'password', required: true },
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
        prop: 'groupPolicy',
        label: t('aiTools.agents.groupPolicy'),
        type: 'select',
        required: true,
        options: [
            { label: t('aiTools.agents.policyOpen'), value: 'open' },
            { label: t('aiTools.agents.policyAllowlist'), value: 'allowlist' },
            { label: t('aiTools.agents.policyDisabled'), value: 'disabled' },
        ],
    },
    {
        prop: 'streaming',
        label: t('aiTools.agents.streaming'),
        type: 'select',
        required: true,
        options: [
            { label: t('aiTools.agents.streamingOff'), value: 'off' },
            { label: t('aiTools.agents.streamingPartial'), value: 'partial' },
            { label: t('aiTools.agents.streamingBlock'), value: 'block' },
            { label: t('aiTools.agents.streamingProgress'), value: 'progress' },
        ],
    },
];

const createBot = (): AI.AgentTelegramBot => ({
    accountId: '',
    name: '',
    enabled: true,
    isDefault: false,
    botToken: '',
    dmPolicy: form.dmPolicy,
    groupPolicy: form.groupPolicy,
    streaming: form.streaming,
});

const maskValue = (value: string) => {
    if (!value) {
        return '';
    }
    if (value.length <= 8) {
        return value;
    }
    return `${value.slice(0, 4)}****${value.slice(-4)}`;
};

const getBotSummary = (bot: AI.AgentTelegramBot) => maskValue(bot.botToken);

const isApproveDisabled = (bot: AI.AgentTelegramBot) => bot.dmPolicy !== 'pairing';

const updateBots = (bots: AI.AgentTelegramBot[]) => {
    form.bots = bots;
    form.defaultAccount = bots.find((bot) => bot.isDefault)?.accountId || '';
};

const load = async (id: number) => {
    agentId.value = id;
    const res = await getAgentTelegramConfig({ agentId: id });
    form.enabled = res.data?.enabled ?? true;
    form.dmPolicy = res.data?.dmPolicy || 'pairing';
    form.allowFromText = (res.data?.allowFrom || []).join('\n');
    form.requireMention = res.data?.requireMention || false;
    form.groupPolicy = res.data?.groupPolicy || 'open';
    form.groupAllowFromText = (res.data?.groupAllowFrom || []).join('\n');
    form.proxy = res.data?.proxy || '';
    form.streaming = res.data?.streaming || 'partial';
    form.defaultAccount = res.data?.defaultAccount || '';
    form.bots = res.data?.bots || [];
};

const saveChannel = async (action: 'delete' | 'save' = 'save') => {
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
        await updateAgentTelegramConfig({
            agentId: agentId.value,
            enabled: form.enabled,
            dmPolicy: form.dmPolicy || 'pairing',
            allowFrom: parseTextList(form.allowFromText),
            requireMention: form.groupPolicy === 'allowlist',
            groupPolicy: form.groupPolicy || 'open',
            groupAllowFrom: parseTextList(form.groupAllowFromText),
            proxy: form.proxy,
            streaming: form.streaming || 'partial',
            defaultAccount: form.defaultAccount,
            bots: form.bots,
        });
        MsgSuccess(action === 'delete' ? t('commons.msg.deleteSuccess') : t('aiTools.agents.saveSuccess'));
    } finally {
        saving.value = false;
    }
};

const approvePairing = async (bot: AI.AgentTelegramBot) => {
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
            type: 'telegram',
            accountId: bot.accountId,
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
    font-size: 12px;
    color: var(--el-text-color-secondary);
    line-height: 1.5;
}
</style>
