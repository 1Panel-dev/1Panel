<template>
    <el-form ref="formRef" v-loading="deleting" :model="form" :rules="rules" label-position="top">
        <el-form-item v-if="configured">
            <el-button type="danger" plain :loading="deleting" :disabled="!hasManagePermission" @click="deleteChannel">
                {{ t('commons.button.delete') }}
            </el-button>
        </el-form-item>
        <el-form-item label="Client ID" prop="clientId">
            <el-input v-model="form.clientId" />
        </el-form-item>
        <el-form-item label="Client Secret" prop="clientSecret">
            <el-input v-model="form.clientSecret" type="password" show-password />
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.dmPolicy')" prop="dmPolicy">
            <el-select v-model="form.dmPolicy">
                <el-option :label="t('aiTools.agents.pairingCode')" value="pairing" />
                <el-option :label="t('aiTools.agents.policyOpen')" value="open" />
                <el-option :label="t('aiTools.agents.policyDisabled')" value="disabled" />
            </el-select>
        </el-form-item>
        <el-form-item>
            <el-button type="primary" :loading="saving" :disabled="!hasManagePermission" @click="save">
                {{ t('commons.button.save') }}
            </el-button>
        </el-form-item>
        <el-alert type="info" :closable="false" :title="t('aiTools.agents.channelAutoRestartHelper')" />
        <template v-if="form.dmPolicy === 'pairing'">
            <el-form-item :label="t('aiTools.agents.pairingCode')">
                <el-input v-model="pairingCode" :placeholder="t('aiTools.agents.pairingCodePlaceholder')" />
            </el-form-item>
            <el-form-item>
                <el-button
                    type="primary"
                    plain
                    :loading="approving"
                    :disabled="!hasManagePermission"
                    @click="approvePairing"
                >
                    {{ t('aiTools.agents.approvePairing') }}
                </el-button>
            </el-form-item>
        </template>
    </el-form>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import { ElMessageBox, type FormInstance } from 'element-plus';
import { useMenuManagePermission } from '@/composables/useMenuManagePermission';
import { useI18n } from 'vue-i18n';
import {
    approveAgentChannelPairing,
    deleteAgentChannelConfig,
    getAgentDingTalkConfig,
    updateAgentDingTalkConfig,
} from '@/api/modules/ai';
import { Rules } from '@/global/form-rules';
import { MsgSuccess, MsgWarning } from '@/utils/message';

interface DingTalkForm {
    clientId: string;
    clientSecret: string;
    dmPolicy: 'pairing' | 'open' | 'disabled';
}

const { t } = useI18n();
const { hasManagePermission } = useMenuManagePermission();
const formRef = ref<FormInstance>();
const saving = ref(false);
const approving = ref(false);
const deleting = ref(false);
const agentId = ref(0);
const pairingCode = ref('');
const configured = ref(false);
const form = reactive<DingTalkForm>({
    clientId: '',
    clientSecret: '',
    dmPolicy: 'pairing',
});

const rules = reactive({
    clientId: [Rules.requiredInput],
    clientSecret: [Rules.requiredInput],
    dmPolicy: [Rules.requiredSelect],
});

const load = async (id: number) => {
    agentId.value = id;
    pairingCode.value = '';
    const res = await getAgentDingTalkConfig({ agentId: id });
    configured.value = !!res.data?.enabled;
    if (!configured.value) {
        form.clientId = '';
        form.clientSecret = '';
        form.dmPolicy = 'pairing';
        return;
    }
    form.clientId = res.data?.bots?.[0]?.clientId || '';
    form.clientSecret = res.data?.bots?.[0]?.clientSecret || '';
    form.dmPolicy = res.data?.dmPolicy === 'disabled' ? 'disabled' : res.data?.dmPolicy === 'open' ? 'open' : 'pairing';
};

const save = async () => {
    if (!agentId.value || !formRef.value) {
        return;
    }
    await formRef.value.validate();
    saving.value = true;
    try {
        await updateAgentDingTalkConfig({
            agentId: agentId.value,
            enabled: true,
            dmPolicy: form.dmPolicy,
            allowFrom: [],
            groupPolicy: 'open',
            groupAllowFrom: [],
            separateSessionByConversation: true,
            groupSessionScope: 'group_sender',
            sharedMemoryAcrossConversations: false,
            asyncMode: false,
            ackText: t('aiTools.agents.ackTextDefault'),
            bots: [
                {
                    accountId: 'default',
                    name: 'Default',
                    enabled: true,
                    isDefault: true,
                    clientId: form.clientId,
                    clientSecret: form.clientSecret,
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
    await ElMessageBox.confirm(
        t('aiTools.agents.channelDeleteConfirm', [t('aiTools.agents.dingtalk')]),
        t('commons.msg.infoTitle'),
        {
            confirmButtonText: t('commons.button.confirm'),
            cancelButtonText: t('commons.button.cancel'),
            type: 'warning',
        },
    );
    deleting.value = true;
    try {
        await deleteAgentChannelConfig({
            agentId: agentId.value,
            type: 'dingtalk',
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
            type: 'dingtalk',
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
