<template>
    <el-form ref="formRef" v-loading="deleting" :model="form" :rules="rules" label-position="top">
        <el-form-item v-if="configured">
            <el-button v-permission type="danger" plain :loading="deleting" @click="deleteChannel">
                {{ t('commons.button.delete') }}
            </el-button>
        </el-form-item>
        <el-form-item label="Token" prop="token">
            <el-input v-model="form.token" show-password />
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.dmPolicy')" prop="dmPolicy">
            <el-select v-model="form.dmPolicy">
                <el-option :label="t('aiTools.agents.pairingCode')" value="pairing" />
                <el-option :label="t('aiTools.agents.policyOpen')" value="open" />
            </el-select>
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.requireMention')">
            <el-switch v-model="form.requireMention" />
        </el-form-item>
        <el-form-item>
            <el-button v-permission type="primary" :loading="saving" @click="save">
                {{ t('commons.button.save') }}
            </el-button>
        </el-form-item>
        <el-alert type="info" :closable="false" :title="t('aiTools.agents.channelAutoRestartHelper')" />
        <template v-if="form.dmPolicy === 'pairing'">
            <el-form-item :label="t('aiTools.agents.pairingCode')">
                <el-input v-model="pairingCode" :placeholder="t('aiTools.agents.pairingCodePlaceholder')" />
            </el-form-item>
            <el-form-item>
                <el-button v-permission type="primary" plain :loading="approving" @click="approvePairing">
                    {{ t('aiTools.agents.approvePairing') }}
                </el-button>
            </el-form-item>
        </template>
    </el-form>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { ElMessageBox, type FormInstance } from 'element-plus';
import { useI18n } from 'vue-i18n';
import {
    approveAgentChannelPairing,
    deleteAgentChannelConfig,
    getAgentDiscordConfig,
    updateAgentDiscordConfig,
} from '@/api/modules/ai';
import { Rules } from '@/global/form-rules';
import { MsgSuccess, MsgWarning } from '@/utils/message';

interface DiscordForm {
    token: string;
    dmPolicy: 'pairing' | 'open';
    requireMention: boolean;
}

const { t } = useI18n();
const formRef = ref<FormInstance>();
const saving = ref(false);
const approving = ref(false);
const deleting = ref(false);
const agentId = ref(0);
const pairingCode = ref('');
const configured = ref(false);
const form = reactive<DiscordForm>({
    token: '',
    dmPolicy: 'open',
    requireMention: true,
});

const rules = reactive({
    token: [Rules.requiredInput],
    dmPolicy: [Rules.requiredSelect],
});

const load = async (id: number) => {
    agentId.value = id;
    pairingCode.value = '';
    const res = await getAgentDiscordConfig({ agentId: id });
    configured.value = !!res.data?.enabled;
    if (!configured.value) {
        form.token = '';
        form.dmPolicy = 'open';
        form.requireMention = true;
        return;
    }
    form.token = res.data?.bots?.[0]?.token || '';
    form.dmPolicy = res.data?.dmPolicy === 'pairing' ? 'pairing' : 'open';
    form.requireMention = res.data?.requireMention ?? true;
};

const save = async () => {
    if (!agentId.value || !formRef.value) {
        return;
    }
    await formRef.value.validate();
    saving.value = true;
    try {
        const groupPolicy = form.requireMention ? 'allowlist' : 'open';
        await updateAgentDiscordConfig({
            agentId: agentId.value,
            enabled: true,
            dmPolicy: form.dmPolicy,
            allowFrom: [],
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
        MsgSuccess(t('aiTools.agents.saveAndRestartSuccess'));
        configured.value = true;
    } finally {
        saving.value = false;
    }
};

const deleteChannel = async () => {
    if (!agentId.value) {
        return;
    }
    await ElMessageBox.confirm(t('aiTools.agents.channelDeleteConfirm', ['Discord']), t('commons.msg.infoTitle'), {
        confirmButtonText: t('commons.button.confirm'),
        cancelButtonText: t('commons.button.cancel'),
        type: 'warning',
    });
    deleting.value = true;
    try {
        await deleteAgentChannelConfig({
            agentId: agentId.value,
            type: 'discord',
        });
        await load(agentId.value);
        MsgSuccess(t('aiTools.agents.deleteAndRestartSuccess'));
    } finally {
        deleting.value = false;
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
