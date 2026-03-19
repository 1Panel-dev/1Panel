<template>
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <el-alert
            v-if="!form.installed"
            type="warning"
            :closable="false"
            :title="t('aiTools.agents.pluginNotInstalled')"
            class="mb-4"
        />
        <el-form-item>
            <el-button v-if="!form.installed" type="primary" :loading="installing" @click="installPlugin">
                {{ t('commons.button.install') }}
            </el-button>
        </el-form-item>
        <el-form-item :label="t('commons.table.status')">
            <el-switch v-model="form.enabled" />
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.clientId')" prop="clientId">
            <el-input v-model="form.clientId" />
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.clientSecret')" prop="clientSecret">
            <el-input v-model="form.clientSecret" type="password" show-password />
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.dmPolicy')" prop="dmPolicy">
            <el-select v-model="form.dmPolicy">
                <el-option :label="t('aiTools.agents.policyPairing')" value="pairing" />
                <el-option :label="t('aiTools.agents.policyAllowlist')" value="allowlist" />
                <el-option :label="t('aiTools.agents.policyOpen')" value="open" />
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
        <el-form-item>
            <el-button type="primary" :loading="saving" :disabled="!form.installed" @click="saveChannel">
                {{ t('commons.button.save') }}
            </el-button>
        </el-form-item>

        <el-divider />

        <el-form-item :label="t('aiTools.agents.pairingCode')">
            <el-input v-model="pairingCode" :placeholder="t('aiTools.agents.pairingCodePlaceholder')" />
        </el-form-item>
        <el-form-item>
            <el-button type="primary" :loading="approving" :disabled="!form.installed" @click="approvePairing">
                {{ t('aiTools.agents.approvePairing') }}
            </el-button>
        </el-form-item>
    </el-form>
    <TaskLog ref="taskLogRef" @close="checkPluginStatus" />
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import type { FormInstance, FormRules } from 'element-plus';
import { useI18n } from 'vue-i18n';
import { AI } from '@/api/interface/ai';
import {
    approveAgentChannelPairing,
    checkAgentPlugin,
    getAgentDingTalkConfig,
    installAgentPlugin,
    updateAgentDingTalkConfig,
} from '@/api/modules/ai';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { Rules } from '@/global/form-rules';
import { newUUID } from '@/utils/util';
import TaskLog from '@/components/log/task/index.vue';

interface DingTalkForm extends AI.AgentDingTalkConfig {
    allowFromText: string;
    groupAllowFromText: string;
}

const { t } = useI18n();
const saving = ref(false);
const approving = ref(false);
const installing = ref(false);
const agentId = ref(0);
const pairingCode = ref('');
const formRef = ref<FormInstance>();
const taskLogRef = ref();

const form = reactive<DingTalkForm>({
    enabled: true,
    clientId: '',
    clientSecret: '',
    dmPolicy: 'pairing',
    allowFrom: [],
    groupPolicy: 'disabled',
    groupAllowFrom: [],
    installed: false,
    allowFromText: '',
    groupAllowFromText: '',
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

const rules = reactive<FormRules>({
    clientId: [Rules.requiredInput],
    clientSecret: [Rules.requiredInput],
    dmPolicy: [Rules.requiredSelect],
    groupPolicy: [Rules.requiredSelect],
    allowFromText: [{ validator: validateAllowFrom, trigger: 'blur' }],
    groupAllowFromText: [{ validator: validateGroupAllowFrom, trigger: 'blur' }],
});

const checkPluginStatus = async () => {
    if (!agentId.value) {
        return;
    }
    const res = await checkAgentPlugin({
        agentId: agentId.value,
        type: 'dingtalk',
    });
    form.installed = Boolean(res.data?.installed);
};

const load = async (id: number) => {
    agentId.value = id;
    pairingCode.value = '';
    const res = await getAgentDingTalkConfig({ agentId: id });
    Object.assign(form, res.data || {});
    form.dmPolicy = form.dmPolicy || 'pairing';
    form.groupPolicy = form.groupPolicy || 'disabled';
    form.allowFromText = (res.data?.allowFrom || []).join('\n');
    form.groupAllowFromText = (res.data?.groupAllowFrom || []).join('\n');
    await checkPluginStatus();
};

const saveChannel = async () => {
    if (!agentId.value || !formRef.value) {
        return;
    }
    await formRef.value.validate();
    saving.value = true;
    try {
        await updateAgentDingTalkConfig({
            agentId: agentId.value,
            enabled: form.enabled,
            clientId: form.clientId,
            clientSecret: form.clientSecret,
            dmPolicy: form.dmPolicy,
            allowFrom: parseTextList(form.allowFromText),
            groupPolicy: form.groupPolicy,
            groupAllowFrom: parseTextList(form.groupAllowFromText),
        });
        MsgSuccess(t('aiTools.agents.saveSuccess'));
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
        await approveAgentChannelPairing({
            agentId: agentId.value,
            type: 'dingtalk-connector',
            pairingCode: pairingCode.value,
        });
        MsgSuccess(t('aiTools.agents.pairingApproveSuccess'));
        pairingCode.value = '';
    } finally {
        approving.value = false;
    }
};

const installPlugin = async () => {
    if (!agentId.value) {
        return;
    }
    const taskID = newUUID();
    installing.value = true;
    try {
        await installAgentPlugin({
            agentId: agentId.value,
            type: 'dingtalk',
            taskID,
        });
        taskLogRef.value?.openWithTaskID(taskID);
    } finally {
        installing.value = false;
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
