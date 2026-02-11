<template>
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <el-form-item :label="t('aiTools.agents.feishu')">
            <el-switch v-model="form.enabled" />
        </el-form-item>
        <el-form-item>
            <el-link type="primary" icon="Position" @click="toFeishuDoc">
                {{ t('container.mirrorsHelper2') }}
            </el-link>
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.dmPolicy')" prop="dmPolicy">
            <el-select v-model="form.dmPolicy">
                <el-option label="pairing" value="pairing" />
            </el-select>
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.botName')" prop="botName">
            <el-input v-model="form.botName" />
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.appId')" prop="appId">
            <el-input v-model="form.appId" />
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.appSecret')" prop="appSecret">
            <el-input v-model="form.appSecret" type="password" show-password />
        </el-form-item>
        <el-form-item>
            <el-button type="primary" :loading="saving" @click="saveChannel">
                {{ t('commons.button.save') }}
            </el-button>
        </el-form-item>

        <el-divider />

        <el-form-item :label="t('aiTools.agents.pairingCode')">
            <el-input v-model="pairingCode" :placeholder="t('aiTools.agents.pairingCodePlaceholder')" />
        </el-form-item>
        <el-form-item>
            <el-button type="primary" :loading="approving" @click="approvePairing">
                {{ t('aiTools.agents.approvePairing') }}
            </el-button>
        </el-form-item>
    </el-form>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import type { FormInstance } from 'element-plus';
import { useI18n } from 'vue-i18n';
import { AI } from '@/api/interface/ai';
import { approveAgentFeishuPairing, getAgentFeishuConfig, updateAgentFeishuConfig } from '@/api/modules/ai';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { Rules } from '@/global/form-rules';

const { t } = useI18n();
const saving = ref(false);
const approving = ref(false);
const agentId = ref(0);
const pairingCode = ref('');
const formRef = ref<FormInstance>();

const form = reactive<AI.AgentFeishuConfig>({
    enabled: true,
    dmPolicy: 'pairing',
    botName: '',
    appId: '',
    appSecret: '',
});

const rules = reactive({
    dmPolicy: [Rules.requiredSelect],
    botName: [Rules.requiredInput],
    appId: [Rules.requiredInput],
    appSecret: [Rules.requiredInput],
});

const toFeishuDoc = () => {
    window.open('https://openclaw.club/guides/feishu-platform', '_blank');
};

const load = async (id: number) => {
    agentId.value = id;
    pairingCode.value = '';
    const res = await getAgentFeishuConfig({ agentId: id });
    Object.assign(form, res.data || {});
    if (!form.dmPolicy) {
        form.dmPolicy = 'pairing';
    }
};

const saveChannel = async () => {
    if (!agentId.value) {
        return;
    }
    if (!formRef.value) {
        return;
    }
    await formRef.value.validate();
    saving.value = true;
    try {
        await updateAgentFeishuConfig({
            agentId: agentId.value,
            enabled: form.enabled,
            dmPolicy: form.dmPolicy || 'pairing',
            botName: form.botName,
            appId: form.appId,
            appSecret: form.appSecret,
        });
        MsgSuccess(t('aiTools.agents.feishuSaveSuccess'));
    } finally {
        saving.value = false;
    }
};

const approvePairing = async () => {
    if (!agentId.value) {
        return;
    }
    if (!pairingCode.value) {
        MsgWarning(t('aiTools.agents.pairingCodeRequired'));
        return;
    }
    approving.value = true;
    try {
        await approveAgentFeishuPairing({
            agentId: agentId.value,
            pairingCode: pairingCode.value,
        });
        MsgSuccess(t('aiTools.agents.pairingApproveSuccess'));
        pairingCode.value = '';
    } finally {
        approving.value = false;
    }
};

defineExpose({
    load,
});
</script>
