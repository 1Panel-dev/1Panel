<template>
    <el-form ref="formRef" v-loading="deleting" :model="form" :rules="rules" label-position="top">
        <el-form-item v-if="configured">
            <el-button type="danger" plain :loading="deleting" @click="deleteChannel">
                {{ t('commons.button.delete') }}
            </el-button>
        </el-form-item>
        <el-form-item label="App ID" prop="appId">
            <el-input v-model="form.appId" />
        </el-form-item>
        <el-form-item label="App Secret" prop="appSecret">
            <el-input v-model="form.appSecret" type="password" show-password />
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.dmPolicy')" prop="dmPolicy">
            <el-select v-model="form.dmPolicy">
                <el-option
                    :label="t('aiTools.agents.pairingCode')"
                    value="pairing"
                    :disabled="form.groupPolicy === 'allowlist'"
                />
                <el-option :label="t('aiTools.agents.policyOpen')" value="open" />
                <el-option :label="t('aiTools.agents.policyAllowlist')" value="allowlist" />
            </el-select>
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.groupPolicy')" prop="groupPolicy">
            <el-select v-model="form.groupPolicy">
                <el-option :label="t('aiTools.agents.policyOpen')" value="open" />
                <el-option :label="t('aiTools.agents.policyAllowlist')" value="allowlist" />
                <el-option :label="t('aiTools.agents.policyDisabled')" value="disabled" />
            </el-select>
        </el-form-item>
        <el-form-item
            v-if="form.dmPolicy === 'allowlist' || form.groupPolicy === 'allowlist'"
            :label="t('aiTools.agents.policyAllowlist')"
            prop="allowFromText"
        >
            <el-input
                v-model="form.allowFromText"
                type="textarea"
                :rows="3"
                :placeholder="t('aiTools.agents.allowFromPlaceholder')"
            />
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
import { reactive, ref, watch } from 'vue';
import { ElMessageBox, type FormInstance } from 'element-plus';
import { useI18n } from 'vue-i18n';
import {
    approveAgentChannelPairing,
    deleteAgentChannelConfig,
    getAgentFeishuConfig,
    updateAgentFeishuConfig,
} from '@/api/modules/ai';
import { Rules } from '@/global/form-rules';
import { MsgSuccess, MsgWarning } from '@/utils/message';

interface FeishuForm {
    appId: string;
    appSecret: string;
    dmPolicy: 'pairing' | 'open' | 'allowlist';
    allowFromText: string;
    groupPolicy: 'open' | 'allowlist' | 'disabled';
}

const { t } = useI18n();
const formRef = ref<FormInstance>();
const saving = ref(false);
const approving = ref(false);
const deleting = ref(false);
const agentId = ref(0);
const pairingCode = ref('');
const configured = ref(false);
const form = reactive<FeishuForm>({
    appId: '',
    appSecret: '',
    dmPolicy: 'pairing',
    allowFromText: '',
    groupPolicy: 'allowlist',
});

watch(
    () => form.groupPolicy,
    (value) => {
        if (value === 'allowlist' && form.dmPolicy === 'pairing') {
            form.dmPolicy = 'allowlist';
        }
    },
);

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
                if (
                    (form.dmPolicy === 'allowlist' || form.groupPolicy === 'allowlist') &&
                    parseTextList(String(value || '')).length === 0
                ) {
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
    configured.value = !!res.data?.enabled;
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
            enabled: true,
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
                    allowFrom: form.dmPolicy === 'allowlist' || form.groupPolicy === 'allowlist' ? allowFrom : [],
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
        t('aiTools.agents.channelDeleteConfirm', [t('aiTools.agents.feishu')]),
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
            type: 'feishu',
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

<style scoped lang="scss"></style>
