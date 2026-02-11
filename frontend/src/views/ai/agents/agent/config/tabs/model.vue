<template>
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top" v-loading="loading">
        <el-form-item :label="t('aiTools.agents.account')" prop="accountId">
            <el-select v-model="form.accountId" @change="handleAccountChange">
                <el-option v-for="item in accountOptions" :key="item.id" :label="item.name" :value="item.id" />
            </el-select>
        </el-form-item>
        <el-form-item>
            <el-checkbox v-model="form.manualModel" @change="handleManualModelChange">
                {{ t('aiTools.agents.manualModel') }}
            </el-checkbox>
        </el-form-item>
        <el-form-item :label="t('aiTools.model.model')" prop="model">
            <el-input v-if="form.manualModel" v-model="form.model" />
            <el-select v-else v-model="form.model" filterable>
                <el-option v-for="item in modelOptions" :key="item.id" :label="item.name" :value="item.id" />
            </el-select>
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
import { getAgentProviders, pageAgentAccounts, updateAgentModelConfig } from '@/api/modules/ai';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';

const emit = defineEmits(['updated']);
const { t } = useI18n();

const loading = ref(false);
const saving = ref(false);
const formRef = ref<FormInstance>();

const agentId = ref(0);
const providerModels = ref<Record<string, AI.ProviderModelInfo[]>>({});
const accountOptions = ref<AI.AgentAccountItem[]>([]);
const modelOptions = ref<AI.ProviderModelInfo[]>([]);

const form = reactive({
    accountId: undefined as unknown as number,
    manualModel: false,
    model: '',
});

const rules = reactive({
    accountId: [Rules.requiredSelect],
    model: [Rules.requiredInput],
});

const loadProviders = async () => {
    if (Object.keys(providerModels.value).length > 0) {
        return;
    }
    const res = await getAgentProviders();
    const data = res.data || [];
    providerModels.value = data.reduce((acc, item) => {
        acc[item.provider] = item.models || [];
        return acc;
    }, {} as Record<string, AI.ProviderModelInfo[]>);
};

const loadAccounts = async () => {
    const res = await pageAgentAccounts({
        page: 1,
        pageSize: 200,
        provider: '',
        name: '',
    });
    accountOptions.value = res.data.items || [];
};

const setModelsByProvider = (provider: string) => {
    modelOptions.value = providerModels.value[provider] || [];
};

const handleAccountChange = () => {
    const selected = accountOptions.value.find((item) => item.id === form.accountId);
    if (!selected) {
        modelOptions.value = [];
        form.model = '';
        return;
    }
    setModelsByProvider(selected.provider);
    if (!form.manualModel && (!form.model || !form.model.startsWith(`${selected.provider}/`))) {
        form.model = modelOptions.value.length > 0 ? modelOptions.value[0].id : '';
    }
};

const handleManualModelChange = (val: unknown) => {
    if (Boolean(val)) {
        return;
    }
    const selected = accountOptions.value.find((item) => item.id === form.accountId);
    if (!selected) {
        form.model = '';
        return;
    }
    if (!form.model || !form.model.startsWith(`${selected.provider}/`)) {
        form.model = modelOptions.value.length > 0 ? modelOptions.value[0].id : '';
    }
};

const load = async (agent: AI.AgentItem) => {
    loading.value = true;
    try {
        agentId.value = agent.id;
        await loadProviders();
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
        setModelsByProvider(currentAccount.provider);
        const inProviderModels = modelOptions.value.some((item) => item.id === agent.model);
        form.manualModel = !inProviderModels;
        if (agent.model && (form.manualModel || agent.model.startsWith(`${currentAccount.provider}/`))) {
            form.model = agent.model;
        } else {
            form.model = modelOptions.value.length > 0 ? modelOptions.value[0].id : '';
        }
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
