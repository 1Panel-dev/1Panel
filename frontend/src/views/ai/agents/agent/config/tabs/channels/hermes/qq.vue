<template>
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <el-form-item :label="t('commons.table.status')">
            <el-switch v-model="form.enabled" />
        </el-form-item>
        <el-form-item label="App ID" prop="appId">
            <el-input v-model="form.appId" />
        </el-form-item>
        <el-form-item label="App Secret" prop="clientSecret">
            <el-input v-model="form.clientSecret" type="password" show-password />
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.dmPolicy')" prop="dmPolicy">
            <el-select v-model="form.dmPolicy">
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
        <el-form-item>
            <el-button type="primary" :loading="saving" @click="save">
                {{ t('commons.button.save') }}
            </el-button>
        </el-form-item>
    </el-form>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import type { FormInstance } from 'element-plus';
import { useI18n } from 'vue-i18n';
import { getAgentQQBotConfig, updateAgentQQBotConfig } from '@/api/modules/ai';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';

interface QQBotForm {
    enabled: boolean;
    appId: string;
    clientSecret: string;
    dmPolicy: 'open' | 'allowlist' | 'disabled';
    allowFromText: string;
    groupPolicy: 'open' | 'allowlist' | 'disabled';
    groupAllowFromText: string;
}

const { t } = useI18n();
const emit = defineEmits(['saved']);
const formRef = ref<FormInstance>();
const saving = ref(false);
const agentId = ref(0);
const form = reactive<QQBotForm>({
    enabled: true,
    appId: '',
    clientSecret: '',
    dmPolicy: 'open',
    allowFromText: '',
    groupPolicy: 'open',
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

const rules = reactive({
    appId: [Rules.requiredInput],
    clientSecret: [Rules.requiredInput],
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
    groupAllowFromText: [
        {
            validator: (_rule, value, callback) => {
                if (form.groupPolicy === 'allowlist' && parseTextList(String(value || '')).length === 0) {
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
    const res = await getAgentQQBotConfig({ agentId: id });
    form.enabled = res.data?.enabled ?? true;
    form.appId = res.data?.bots?.[0]?.appId || '';
    form.clientSecret = res.data?.bots?.[0]?.clientSecret || '';
    form.dmPolicy = (res.data?.dmPolicy as QQBotForm['dmPolicy']) || 'open';
    form.allowFromText = (res.data?.allowFrom || []).join('\n');
    form.groupPolicy = (res.data?.groupPolicy as QQBotForm['groupPolicy']) || 'open';
    form.groupAllowFromText = (res.data?.groupAllowFrom || []).join('\n');
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
            enabled: form.enabled,
            dmPolicy: form.dmPolicy,
            allowFrom: parseTextList(form.allowFromText),
            groupPolicy: form.groupPolicy,
            groupAllowFrom: parseTextList(form.groupAllowFromText),
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
        MsgSuccess(t('aiTools.agents.saveSuccess'));
        emit('saved');
    } finally {
        saving.value = false;
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
