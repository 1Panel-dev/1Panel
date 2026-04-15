<template>
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
        <el-form-item :label="t('commons.table.status')">
            <el-switch v-model="form.enabled" />
        </el-form-item>
        <el-form-item label="Client ID" prop="clientId">
            <el-input v-model="form.clientId" />
        </el-form-item>
        <el-form-item label="Client Secret" prop="clientSecret">
            <el-input v-model="form.clientSecret" type="password" show-password />
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.allowFrom')">
            <el-input
                v-model="form.allowFromText"
                type="textarea"
                :rows="3"
                :placeholder="t('aiTools.agents.allowFromPlaceholder')"
            />
            <span class="input-help">{{ t('aiTools.agents.allowFromHelper') }}</span>
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
import { getAgentDingTalkConfig, updateAgentDingTalkConfig } from '@/api/modules/ai';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';

interface DingTalkForm {
    enabled: boolean;
    clientId: string;
    clientSecret: string;
    allowFromText: string;
}

const { t } = useI18n();
const emit = defineEmits(['saved']);
const formRef = ref<FormInstance>();
const saving = ref(false);
const agentId = ref(0);
const form = reactive<DingTalkForm>({
    enabled: true,
    clientId: '',
    clientSecret: '',
    allowFromText: '',
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
    clientId: [Rules.requiredInput],
    clientSecret: [Rules.requiredInput],
});

const load = async (id: number) => {
    agentId.value = id;
    const res = await getAgentDingTalkConfig({ agentId: id });
    form.enabled = res.data?.enabled ?? true;
    form.clientId = res.data?.bots?.[0]?.clientId || '';
    form.clientSecret = res.data?.bots?.[0]?.clientSecret || '';
    form.allowFromText = (res.data?.allowFrom || []).join('\n');
};

const save = async () => {
    if (!agentId.value || !formRef.value) {
        return;
    }
    await formRef.value.validate();
    saving.value = true;
    try {
        const allowFrom = parseTextList(form.allowFromText);
        await updateAgentDingTalkConfig({
            agentId: agentId.value,
            enabled: form.enabled,
            dmPolicy: allowFrom.length > 0 ? 'allowlist' : 'open',
            allowFrom,
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
