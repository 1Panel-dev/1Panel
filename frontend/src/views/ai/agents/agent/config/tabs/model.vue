<template>
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top" v-loading="loading">
        <el-form-item :label="t('aiTools.agents.account')" prop="accountId">
            <el-select v-model="form.accountId" @change="handleAccountChange">
                <el-option v-for="item in accountOptions" :key="item.id" :label="item.name" :value="item.id" />
            </el-select>
        </el-form-item>
        <el-form-item :label="t('aiTools.model.model')" prop="model">
            <el-select v-model="form.model" filterable>
                <el-option v-for="item in modelOptions" :key="item.id" :label="item.name" :value="item.id" />
            </el-select>
            <span class="input-help">{{ t('aiTools.agents.accountModelsHelper') }}</span>
        </el-form-item>
        <el-form-item>
            <el-button type="primary" :loading="saving" @click="saveModel">
                {{ t('commons.button.save') }}
            </el-button>
        </el-form-item>
    </el-form>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import type { FormInstance } from 'element-plus';
import { useI18n } from 'vue-i18n';
import { AI } from '@/api/interface/ai';
import { pageAgentAccounts, updateAgentModelConfig } from '@/api/modules/ai';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';
import { useGlobalStore } from '@/composables/useGlobalStore';

const emit = defineEmits(['updated']);
const { t } = useI18n();

const loading = ref(false);
const saving = ref(false);
const formRef = ref<FormInstance>();
const { isIntl } = useGlobalStore();
const blockedProviders = new Set(['ark-coding-plan', 'bailian-coding-plan']);

const agentId = ref(0);
const accountOptions = ref<AI.AgentAccountItem[]>([]);
const modelOptions = ref<AI.AgentAccountModel[]>([]);

const form = reactive({
    accountId: undefined as unknown as number,
    model: '',
});

const rules = reactive({
    accountId: [Rules.requiredSelect],
    model: [Rules.requiredSelect],
});

const loadAccounts = async () => {
    const res = await pageAgentAccounts({
        page: 1,
        pageSize: 200,
        provider: '',
        name: '',
    });
    const items = res.data.items || [];
    accountOptions.value = isIntl.value ? items.filter((item) => !blockedProviders.has(item.provider)) : items;
};

const setModelOptionsByAccount = (accountId: number) => {
    const selected = accountOptions.value.find((item) => item.id === accountId);
    modelOptions.value = selected?.models || [];
    if (!modelOptions.value.some((item) => item.id === form.model)) {
        form.model = modelOptions.value[0]?.id || '';
    }
};

const handleAccountChange = () => {
    setModelOptionsByAccount(form.accountId);
};

const load = async (agent: AI.AgentItem) => {
    loading.value = true;
    try {
        agentId.value = agent.id;
        await loadAccounts();
        if (accountOptions.value.length === 0) {
            form.accountId = undefined as unknown as number;
            form.model = '';
            modelOptions.value = [];
            return;
        }
        const currentAccount =
            accountOptions.value.find((item) => item.id === agent.accountId) || accountOptions.value[0];
        form.accountId = currentAccount.id;
        form.model = agent.model || currentAccount.models?.[0]?.id || '';
        setModelOptionsByAccount(currentAccount.id);
    } finally {
        loading.value = false;
    }
};

const saveModel = async () => {
    if (!agentId.value || !formRef.value) {
        return;
    }
    await formRef.value.validate();
    saving.value = true;
    try {
        await updateAgentModelConfig({
            agentId: agentId.value,
            accountId: form.accountId,
            model: form.model,
        });
        MsgSuccess(t('aiTools.agents.switchModelSuccess'));
        emit('updated');
    } finally {
        saving.value = false;
    }
};

defineExpose({
    load,
});
</script>
