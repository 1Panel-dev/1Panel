<template>
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <el-form-item :label="t('commons.table.status')">
            <el-switch v-model="form.enabled" />
        </el-form-item>
        <el-form-item label="App ID" prop="appId">
            <el-input v-model="form.appId" />
        </el-form-item>
        <el-form-item label="App Secret" prop="appSecret">
            <el-input v-model="form.appSecret" type="password" show-password />
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
        <el-form-item :label="t('aiTools.agents.groupPolicy')" prop="groupPolicy">
            <el-select v-model="form.groupPolicy">
                <el-option :label="t('aiTools.agents.policyOpen')" value="open" />
                <el-option :label="t('aiTools.agents.policyAllowlist')" value="allowlist" />
                <el-option :label="t('aiTools.agents.policyDisabled')" value="disabled" />
            </el-select>
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
import { approveAgentChannelPairing, getAgentFeishuConfig, updateAgentFeishuConfig } from '@/api/modules/ai';
import { Rules } from '@/global/form-rules';
import { MsgSuccess, MsgWarning } from '@/utils/message';

interface FeishuForm {
    enabled: boolean;
    appId: string;
    appSecret: string;
    dmPolicy: 'pairing' | 'open' | 'allowlist';
    allowFromText: string;
    groupPolicy: 'open' | 'allowlist' | 'disabled';
}

const { t } = useI18n();
const emit = defineEmits(['saved']);
const formRef = ref<FormInstance>();
const saving = ref(false);
const approving = ref(false);
const agentId = ref(0);
const pairingCode = ref('');
const form = reactive<FeishuForm>({
    enabled: true,
    appId: '',
    appSecret: '',
    dmPolicy: 'pairing',
    allowFromText: '',
    groupPolicy: 'allowlist',
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
    appId: [Rules.requiredInput],
    appSecret: [Rules.requiredInput],
    dmPolicy: [Rules.requiredSelect],
    groupPolicy: [Rules.requiredSelect],
    allowFromText: [
        {
            validator: (_rule, value, callback) => {
                if (form.dmPolicy === 'allowlist' && parseTextList(String(value || '')).length === 0) {
                    callback(new Error(t('aiTools.agents.allowFromRequired')));
                    return;
                }
                callback();
            },
            trigger: 'blur',
        },
    ],
});

const load = async (id: number) => {
    agentId.value = id;
    pairingCode.value = '';
    const res = await getAgentFeishuConfig({ agentId: id });
    form.enabled = res.data?.enabled ?? true;
    form.appId = res.data?.bots?.[0]?.appId || '';
    form.appSecret = res.data?.bots?.[0]?.appSecret || '';
    form.dmPolicy = (res.data?.bots?.[0]?.dmPolicy as FeishuForm['dmPolicy']) || 'pairing';
    form.allowFromText = (res.data?.bots?.[0]?.allowFrom || []).join('\n');
    form.groupPolicy = (res.data?.groupPolicy as FeishuForm['groupPolicy']) || 'allowlist';
};

const save = async () => {
    if (!agentId.value || !formRef.value) {
        return;
    }
    await formRef.value.validate();
    saving.value = true;
    try {
        const allowFrom = parseTextList(form.allowFromText);
        await updateAgentFeishuConfig({
            agentId: agentId.value,
            enabled: form.enabled,
            threadSession: true,
            replyMode: 'auto',
            streaming: false,
            requireMention: 'true',
            groupPolicy: form.groupPolicy,
            groupAllowFrom: [],
            bots: [
                {
                    accountId: 'default',
                    name: 'Default',
                    enabled: true,
                    isDefault: true,
                    appId: form.appId,
                    appSecret: form.appSecret,
                    dmPolicy: form.dmPolicy,
                    allowFrom: form.dmPolicy === 'allowlist' ? allowFrom : [],
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
            type: 'feishu',
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
