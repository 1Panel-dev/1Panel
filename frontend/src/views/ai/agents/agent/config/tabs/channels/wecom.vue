<template>
    <el-form ref="formRef" v-loading="loading || approving" :model="form" :rules="rules" label-position="top">
        <PluginInstall
            :installed="installed"
            :installing="installing"
            :upgrading="upgrading"
            :uninstalling="uninstalling"
            :current-version="currentVersion"
            :latest-version="latestVersion"
            :upgradable="upgradable"
            :install-action="installPlugin"
            :upgrade-action="upgradePlugin"
            :uninstall-action="uninstallPlugin"
            :on-task-close="reload"
        />
        <template v-if="installed">
            <el-form-item :label="t('commons.table.status')" class="mt-4">
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
            <el-form-item
                v-if="form.dmPolicy === 'allowlist'"
                :label="t('aiTools.agents.allowFrom')"
                prop="allowFromText"
            >
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
                <el-button type="primary" :loading="saving" :disabled="!installed" @click="saveChannel">
                    {{ t('commons.button.save') }}
                </el-button>
            </el-form-item>

            <template v-if="form.dmPolicy === 'pairing'">
                <el-divider />

                <el-form-item :label="t('aiTools.agents.pairingCode')">
                    <el-input v-model="pairingCode" :placeholder="t('aiTools.agents.pairingCodePlaceholder')" />
                </el-form-item>
                <el-form-item>
                    <el-button type="primary" :loading="approving" :disabled="!installed" @click="approvePairing">
                        {{ t('aiTools.agents.approvePairing') }}
                    </el-button>
                </el-form-item>
            </template>
        </template>
    </el-form>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import type { FormInstance } from 'element-plus';
import { useI18n } from 'vue-i18n';
import { AI } from '@/api/interface/ai';
import { approveAgentChannelPairing, getAgentWecomConfig, updateAgentWecomConfig } from '@/api/modules/ai';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { Rules } from '@/global/form-rules';
import PluginInstall from './components/plugin-install.vue';
import { useAgentPluginChannel } from './useAgentPluginChannel';

interface WecomForm extends Omit<AI.AgentWecomConfig, 'installed' | 'allowFrom' | 'groupAllowFrom'> {
    allowFromText: string;
    groupAllowFromText: string;
}

const { t } = useI18n();
const saving = ref(false);
const approving = ref(false);
const pairingCode = ref('');
const formRef = ref<FormInstance>();
const {
    agentId,
    loading,
    installed,
    installing,
    upgrading,
    uninstalling,
    currentVersion,
    latestVersion,
    upgradable,
    loadPlugin,
    installPlugin,
    upgradePlugin,
    uninstallPlugin,
} = useAgentPluginChannel('wecom');

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

const rules = reactive({
    dmPolicy: [Rules.requiredSelect],
    groupPolicy: [Rules.requiredSelect],
    allowFromText: [{ validator: validateAllowFrom, trigger: 'blur' }],
    groupAllowFromText: [{ validator: validateGroupAllowFrom, trigger: 'blur' }],
    botId: [Rules.requiredInput],
    secret: [Rules.requiredInput],
});

const load = async (id: number) => {
    await loadPlugin(id);
    pairingCode.value = '';
    const res = await getAgentWecomConfig({ agentId: id });
    form.enabled = res.data?.enabled ?? true;
    form.dmPolicy = res.data?.dmPolicy || 'pairing';
    form.allowFromText = (res.data?.allowFrom || []).join('\n');
    form.groupPolicy = res.data?.groupPolicy || 'open';
    form.groupAllowFromText = (res.data?.groupAllowFrom || []).join('\n');
    form.botId = res.data?.botId || '';
    form.secret = res.data?.secret || '';
    if (!form.dmPolicy) {
        form.dmPolicy = 'pairing';
    }
};

const reload = async () => {
    if (!agentId.value) {
        return;
    }
    await load(agentId.value);
};

const saveChannel = async () => {
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

<style lang="scss" scoped>
.input-help {
    font-size: 12px;
    color: var(--el-text-color-secondary);
    line-height: 1.5;
}
</style>
