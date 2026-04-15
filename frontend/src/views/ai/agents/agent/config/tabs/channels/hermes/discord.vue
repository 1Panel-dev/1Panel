<template>
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <el-form-item label="Token" prop="token">
            <el-input v-model="form.token" show-password />
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.dmPolicy')" prop="dmPolicy">
            <el-select v-model="form.dmPolicy">
                <el-option :label="t('aiTools.agents.pairingCode')" value="pairing" />
                <el-option :label="t('aiTools.agents.policyOpen')" value="open" />
                <el-option :label="t('aiTools.agents.policyAllowlist')" value="allowlist" />
            </el-select>
        </el-form-item>
        <el-form-item v-if="form.dmPolicy === 'allowlist'" :label="t('aiTools.agents.allowFrom')" prop="allowFromText">
            <el-input
                v-model="form.allowFromText"
                type="textarea"
                :rows="3"
                :placeholder="t('aiTools.agents.allowFromPlaceholder')"
            />
            <span class="input-help">{{ t('aiTools.agents.allowFromHelper') }}</span>
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.requireMention')">
            <el-switch v-model="form.requireMention" />
        </el-form-item>
        <el-form-item>
            <el-button type="primary" :loading="saving" @click="save">
                {{ t('commons.button.save') }}
            </el-button>
        </el-form-item>
        <template v-if="form.dmPolicy === 'pairing'">
            <el-form-item :label="t('aiTools.agents.pairingCode')">
                <el-input v-model="pairingCode" :placeholder="t('aiTools.agents.pairingCodePlaceholder')" />
            </el-form-item>
            <el-form-item>
                <el-button type="primary" plain :loading="approving" @click="approvePairing">
                    {{ t('aiTools.agents.approvePairing') }}
                </el-button>
            </el-form-item>
        </template>
    </el-form>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import type { FormInstance } from 'element-plus';
import { useI18n } from 'vue-i18n';
import { approveAgentChannelPairing, getAgentDiscordConfig, updateAgentDiscordConfig } from '@/api/modules/ai';
import { Rules } from '@/global/form-rules';
import { MsgSuccess, MsgWarning } from '@/utils/message';

interface DiscordForm {
    token: string;
    dmPolicy: 'pairing' | 'open' | 'allowlist';
    allowFromText: string;
    requireMention: boolean;
}

const { t } = useI18n();
const emit = defineEmits(['saved']);
const formRef = ref<FormInstance>();
const saving = ref(false);
const approving = ref(false);
const agentId = ref(0);
const pairingCode = ref('');
const form = reactive<DiscordForm>({
    token: '',
    dmPolicy: 'pairing',
    allowFromText: '',
    requireMention: true,
});

const parseTextList = (value: string): string[] => {
    return Array.from(
        new Set(
            value
                .split(/\r?\n/)
                .map((item) => item.trim())
                .filter(Boolean),
        ),
    );
};

const validateAllowFrom = (_rule: any, value: string, callback: (error?: Error) => void) => {
    if (form.dmPolicy !== 'allowlist' || parseTextList(value).length > 0) {
        callback();
        return;
    }
    callback(new Error(t('aiTools.agents.allowFromRequired')));
};

const rules = reactive({
    token: [Rules.requiredInput],
    dmPolicy: [Rules.requiredSelect],
    allowFromText: [{ validator: validateAllowFrom, trigger: 'blur' }],
});

const load = async (id: number) => {
    agentId.value = id;
    const res = await getAgentDiscordConfig({ agentId: id });
    form.token = res.data?.bots?.[0]?.token || '';
    form.dmPolicy = (res.data?.dmPolicy as DiscordForm['dmPolicy']) || 'pairing';
    form.allowFromText = (res.data?.allowFrom || []).join('\n');
    form.requireMention = res.data?.requireMention ?? true;
};

const save = async () => {
    if (!agentId.value || !formRef.value) {
        return;
    }
    await formRef.value.validate();
    saving.value = true;
    try {
        const allowFrom = parseTextList(form.allowFromText);
        const groupPolicy = form.requireMention ? 'allowlist' : 'open';
        await updateAgentDiscordConfig({
            agentId: agentId.value,
            enabled: true,
            dmPolicy: form.dmPolicy,
            allowFrom,
            requireMention: form.requireMention,
            groupPolicy,
            proxy: '',
            defaultAccount: 'default',
            bots: [
                {
                    accountId: 'default',
                    name: 'Default',
                    enabled: true,
                    isDefault: true,
                    token: form.token,
                },
            ],
        });
        MsgSuccess(t('aiTools.agents.saveSuccess'));
        emit('saved');
    } finally {
        saving.value = false;
    }
};

const approvePairing = async () => {
    if (!agentId.value) {
        return;
    }
    if (!pairingCode.value) {
        MsgWarning(t('aiTools.agents.pairingCodePlaceholder'));
        return;
    }
    approving.value = true;
    try {
        await approveAgentChannelPairing({
            agentId: agentId.value,
            type: 'discord',
            pairingCode: pairingCode.value,
        });
        pairingCode.value = '';
        MsgSuccess(t('aiTools.agents.pairingApproveSuccess'));
    } finally {
        approving.value = false;
    }
};

defineExpose({
    load,
});
</script>

<style scoped lang="scss">
.input-help {
    display: block;
    margin-top: 8px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
}
</style>
