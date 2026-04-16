<template>
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
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
        <el-form-item :label="t('aiTools.agents.botId')" prop="botId">
            <el-input v-model="form.botId" />
        </el-form-item>
        <el-form-item :label="t('setting.secret')" prop="secret">
            <el-input v-model="form.secret" type="password" show-password />
        </el-form-item>
        <el-form-item>
            <el-button type="primary" :loading="saving" @click="save">
                {{ t('commons.button.save') }}
            </el-button>
        </el-form-item>
        <el-alert type="info" :closable="false" :title="t('aiTools.agents.channelAutoRestartHelper')" />
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
import { approveAgentChannelPairing, getAgentWecomConfig, updateAgentWecomConfig } from '@/api/modules/ai';
import { Rules } from '@/global/form-rules';
import { MsgSuccess, MsgWarning } from '@/utils/message';

interface WecomForm {
    enabled: boolean;
    dmPolicy: 'pairing' | 'open' | 'allowlist' | 'disabled';
    allowFromText: string;
    groupPolicy: 'open' | 'allowlist' | 'disabled';
    groupAllowFromText: string;
    botId: string;
    secret: string;
}

const { t } = useI18n();
const formRef = ref<FormInstance>();
const saving = ref(false);
const approving = ref(false);
const agentId = ref(0);
const pairingCode = ref('');
const form = reactive<WecomForm>({
    enabled: true,
    dmPolicy: 'pairing',
    allowFromText: '',
    groupPolicy: 'open',
    groupAllowFromText: '',
    botId: '',
    secret: '',
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
    if (form.dmPolicy !== 'allowlist' || parseTextList(value).length > 0) {
        callback();
        return;
    }
    callback(new Error(t('aiTools.agents.allowFromRequired')));
};

const validateGroupAllowFrom = (_rule: any, value: string, callback: (error?: Error) => void) => {
    if (form.groupPolicy !== 'allowlist' || parseTextList(value).length > 0) {
        callback();
        return;
    }
    callback(new Error(t('aiTools.agents.allowFromRequired')));
};

const rules = reactive({
    dmPolicy: [Rules.requiredSelect],
    allowFromText: [{ validator: validateAllowFrom, trigger: 'blur' }],
    groupPolicy: [Rules.requiredSelect],
    groupAllowFromText: [{ validator: validateGroupAllowFrom, trigger: 'blur' }],
    botId: [Rules.requiredInput],
    secret: [Rules.requiredInput],
});

const load = async (id: number) => {
    agentId.value = id;
    pairingCode.value = '';
    const res = await getAgentWecomConfig({ agentId: id });
    form.enabled = res.data?.enabled ?? true;
    form.dmPolicy = (res.data?.dmPolicy as WecomForm['dmPolicy']) || 'pairing';
    form.allowFromText = (res.data?.allowFrom || []).join('\n');
    form.groupPolicy = (res.data?.groupPolicy as WecomForm['groupPolicy']) || 'open';
    form.groupAllowFromText = (res.data?.groupAllowFrom || []).join('\n');
    form.botId = res.data?.botId || '';
    form.secret = res.data?.secret || '';
};

const save = async () => {
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
            allowFrom: parseTextList(form.allowFromText),
            groupPolicy: form.groupPolicy,
            groupAllowFrom: parseTextList(form.groupAllowFromText),
            botId: form.botId,
            secret: form.secret,
        });
        MsgSuccess(t('aiTools.agents.saveAndRestartSuccess'));
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
            type: 'wecom',
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
