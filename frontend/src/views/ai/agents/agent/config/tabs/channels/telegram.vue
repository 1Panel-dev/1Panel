<template>
    <el-form ref="formRef" v-loading="approving" :model="form" :rules="rules" label-position="top">
        <el-form-item :label="t('commons.table.status')">
            <el-switch v-model="form.enabled" />
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.dmPolicy')" prop="dmPolicy">
            <el-select v-model="form.dmPolicy">
                <el-option :label="t('aiTools.agents.pairingCode')" value="pairing" />
                <el-option :label="t('aiTools.agents.policyOpen')" value="open" />
            </el-select>
        </el-form-item>
        <el-form-item :label="t('setting.proxy')">
            <el-input v-model="form.proxy" placeholder="http://127.0.0.1:7890" />
        </el-form-item>
        <ChannelBots
            :bots="form.bots"
            :fields="botFields"
            :create-bot="createBot"
            summary-label="Token"
            :summary-formatter="getBotSummary"
            defaultable
            approvable
            @update:bots="updateBots"
            @save="saveChannel"
            @approve="approvePairing"
        />
        <el-form-item class="mt-4">
            <el-button type="primary" :loading="saving" @click="saveChannel">
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
import ChannelBots from './components/channel-bots.vue';

const { t } = useI18n();
const saving = ref(false);
const approving = ref(false);
const agentId = ref(0);
const formRef = ref<FormInstance>();

const form = reactive<AI.AgentTelegramConfig>({
    enabled: true,
    dmPolicy: 'pairing',
    proxy: '',
    defaultAccount: '',
    bots: [],
});

const rules = reactive({
    dmPolicy: [Rules.requiredSelect],
});

const botFields = [{ prop: 'botToken', label: 'Bot Token', type: 'password', required: true }];

const createBot = (): AI.AgentTelegramBot => ({
    accountId: '',
    name: '',
    enabled: true,
    isDefault: false,
    botToken: '',
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

const getBotSummary = (bot: AI.AgentTelegramBot) => {
    return maskValue(bot.botToken);
};

const updateBots = (bots: AI.AgentTelegramBot[]) => {
    form.bots = bots;
    form.defaultAccount = bots.find((bot) => bot.isDefault)?.accountId || '';
};

const load = async (id: number) => {
    agentId.value = id;
    const res = await getAgentTelegramConfig({ agentId: id });
    form.enabled = res.data?.enabled ?? true;
    form.dmPolicy = res.data?.dmPolicy || 'pairing';
    form.proxy = res.data?.proxy || '';
    form.defaultAccount = res.data?.defaultAccount || '';
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
        await updateAgentTelegramConfig({
            agentId: agentId.value,
            enabled: form.enabled,
            dmPolicy: form.dmPolicy || 'pairing',
            proxy: form.proxy,
            defaultAccount: form.defaultAccount,
            bots: form.bots,
        });
        MsgSuccess(t('aiTools.agents.saveSuccess'));
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
