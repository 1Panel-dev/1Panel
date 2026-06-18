<template>
    <el-form ref="formRef" v-loading="deleting" :model="form" :rules="rules" label-position="top">
        <el-form-item v-if="configured">
            <el-button v-permission type="danger" plain :loading="deleting" @click="deleteChannel">
                {{ t('commons.button.delete') }}
            </el-button>
        </el-form-item>
        <el-form-item label="App ID" prop="appId">
            <el-input v-model="form.appId" />
        </el-form-item>
        <el-form-item label="App Secret" prop="clientSecret">
            <el-input v-model="form.clientSecret" type="password" show-password />
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.dmPolicy')" prop="dmPolicy">
            <el-select v-model="form.dmPolicy">
                <el-option :label="t('aiTools.agents.pairingCode')" value="pairing" />
                <el-option :label="t('aiTools.agents.policyOpen')" value="open" />
                <el-option :label="t('aiTools.agents.policyDisabled')" value="disabled" />
            </el-select>
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.groupPolicy')" prop="groupPolicy">
            <el-select v-model="form.groupPolicy">
                <el-option :label="t('aiTools.agents.policyOpen')" value="open" />
                <el-option :label="t('aiTools.agents.policyDisabled')" value="disabled" />
            </el-select>
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
    getAgentQQBotConfig,
    updateAgentQQBotConfig,
} from '@/api/modules/ai';
import { Rules } from '@/global/form-rules';
import { MsgSuccess, MsgWarning } from '@/utils/message';

interface QQBotForm {
    appId: string;
    clientSecret: string;
    dmPolicy: 'pairing' | 'open' | 'disabled';
    groupPolicy: 'open' | 'disabled';
}

const { t } = useI18n();
const formRef = ref<FormInstance>();
const saving = ref(false);
const approving = ref(false);
const deleting = ref(false);
const agentId = ref(0);
const pairingCode = ref('');
const configured = ref(false);
const form = reactive<QQBotForm>({
    appId: '',
    clientSecret: '',
    dmPolicy: 'open',
    groupPolicy: 'open',
});

const rules = reactive({
    appId: [Rules.requiredInput],
    clientSecret: [Rules.requiredInput],
    dmPolicy: [Rules.requiredSelect],
    groupPolicy: [Rules.requiredSelect],
});

const load = async (id: number) => {
    agentId.value = id;
    pairingCode.value = '';
    const res = await getAgentQQBotConfig({ agentId: id });
    configured.value = !!res.data?.enabled;
    if (!configured.value) {
        form.appId = '';
        form.clientSecret = '';
        form.dmPolicy = 'open';
        form.groupPolicy = 'open';
        return;
    }
    form.appId = res.data?.bots?.[0]?.appId || '';
    form.clientSecret = res.data?.bots?.[0]?.clientSecret || '';
    form.dmPolicy =
        res.data?.dmPolicy === 'disabled' ? 'disabled' : res.data?.dmPolicy === 'pairing' ? 'pairing' : 'open';
    form.groupPolicy = res.data?.groupPolicy === 'disabled' ? 'disabled' : 'open';
};

const save = async () => {
    if (!agentId.value || !formRef.value) {
        return;
    }
    await formRef.value.validate();
    saving.value = true;
    try {
        await updateAgentQQBotConfig({
            agentId: agentId.value,
            enabled: true,
            dmPolicy: form.dmPolicy,
            allowFrom: [],
            groupPolicy: form.groupPolicy,
            groupAllowFrom: [],
            bots: [
                {
                    accountId: 'default',
                    name: 'Default',
                    enabled: true,
                    isDefault: true,
                    appId: form.appId,
                    clientSecret: form.clientSecret,
                    allowFrom: [],
                    systemPrompt: '',
                },
            ],
        });
        MsgSuccess(t('aiTools.agents.successAndRestart', [t('aiTools.agents.saveSuccess')]));
        configured.value = true;
    } finally {
        saving.value = false;
    }
};

const deleteChannel = async () => {
    if (!agentId.value) {
        return;
    }
    await ElMessageBox.confirm(t('aiTools.agents.channelDeleteConfirm', ['QQ']), t('commons.msg.infoTitle'), {
        confirmButtonText: t('commons.button.confirm'),
        cancelButtonText: t('commons.button.cancel'),
        type: 'warning',
    });
    deleting.value = true;
    try {
        await deleteAgentChannelConfig({
            agentId: agentId.value,
            type: 'qqbot',
        });
        await load(agentId.value);
        MsgSuccess(t('aiTools.agents.successAndRestart', [t('commons.msg.deleteSuccess')]));
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
            type: 'qqbot',
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
