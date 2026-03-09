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
            <el-input v-if="form.manualModel && isOllamaProvider" v-model="ollamaModelSuffix">
                <template #prepend>ollama/</template>
            </el-input>
            <el-input v-else-if="form.manualModel" v-model="form.model" />
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
import { computed, reactive, ref } from 'vue';
import type { FormInstance } from 'element-plus';
import { useI18n } from 'vue-i18n';
import { AI } from '@/api/interface/ai';
import { getAgentProviders, pageAgentAccounts, updateAgentModelConfig } from '@/api/modules/ai';
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
const providerModels = ref<Record<string, AI.ProviderModelInfo[]>>({});
const accountOptions = ref<AI.AgentAccountItem[]>([]);
const modelOptions = ref<AI.ProviderModelInfo[]>([]);
const provdier = ref('');

const form = reactive({
    accountId: undefined as unknown as number,
    manualModel: false,
    model: '',
});

const isOllamaProvider = computed(() => provdier.value === 'ollama');

const ollamaModelSuffix = computed({
    get: () => {
        const model = form.model || '';
        if (model.startsWith('ollama/')) {
            return model.slice('ollama/'.length);
        }
        return model;
    },
    set: (value: string) => {
        form.model = `ollama/${value || ''}`;
    },
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
    const filteredData = isIntl.value ? data.filter((item) => !blockedProviders.has(item.provider)) : data;
    providerModels.value = filteredData.reduce((acc, item) => {
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
    const items = res.data.items || [];
    accountOptions.value = isIntl.value ? items.filter((item) => !blockedProviders.has(item.provider)) : items;
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
    provdier.value = selected.provider;
    if (selected.provider === 'custom' || selected.provider === 'vllm') {
        form.manualModel = true;
        form.model = selected.model || form.model;
        modelOptions.value = [];
        return;
    }
    if (selected.provider === 'ollama') {
        form.manualModel = true;
        if (!form.model.startsWith('ollama/')) {
            form.model = 'ollama/';
        }
        setModelsByProvider(selected.provider);
        return;
    }
    setModelsByProvider(selected.provider);
    if (!form.manualModel && (!form.model || !form.model.startsWith(`${selected.provider}/`))) {
        form.model = modelOptions.value.length > 0 ? modelOptions.value[0].id : '';
    }
};

const handleManualModelChange = (val: unknown) => {
    const selected = accountOptions.value.find((item) => item.id === form.accountId);
    if ((selected?.provider === 'custom' || selected?.provider === 'vllm') && !Boolean(val)) {
        form.manualModel = true;
        return;
    }
    if (selected?.provider === 'ollama' && Boolean(val) && !form.model.startsWith('ollama/')) {
        form.model = 'ollama/';
    }
    if (Boolean(val)) {
        return;
    }
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
        provdier.value = currentAccount.provider;
        setModelsByProvider(currentAccount.provider);
        const inProviderModels = modelOptions.value.some((item) => item.id === agent.model);
        form.manualModel =
            currentAccount.provider === 'custom' ||
            currentAccount.provider === 'vllm' ||
            currentAccount.provider === 'ollama' ||
            !inProviderModels;
        if (agent.model && (form.manualModel || agent.model.startsWith(`${currentAccount.provider}/`))) {
            form.model = agent.model;
        } else {
            form.model = modelOptions.value.length > 0 ? modelOptions.value[0].id : '';
        }
        if (
            (currentAccount.provider === 'custom' || currentAccount.provider === 'vllm') &&
            currentAccount.model &&
            !form.model
        ) {
            form.model = currentAccount.model;
        }
        if (currentAccount.provider === 'ollama' && !form.model.startsWith('ollama/')) {
            form.model = 'ollama/';
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
