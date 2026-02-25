<template>
    <DrawerPro v-model="open" :header="headerTitle" size="large" @close="handleClose">
        <el-form ref="formRef" :model="form" :rules="rules" label-position="top" v-loading="loading">
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input v-model="form.name" />
            </el-form-item>
            <el-form-item :label="$t('aiTools.agents.provider')" prop="provider">
                <el-select v-model="form.provider" @change="handleProviderChange" :disabled="form.id > 0">
                    <el-option
                        v-for="item in providerOptions"
                        :key="item.value"
                        :label="item.label"
                        :value="item.value"
                    />
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('aiTools.agents.apiKey')" prop="apiKey">
                <el-input v-model="form.apiKey" type="password" show-password />
                <span class="input-help">{{ $t('aiTools.agents.customProviderHelper') }}</span>
            </el-form-item>
            <el-form-item>
                <el-checkbox v-model="form.rememberApiKey">{{ $t('terminal.rememberPassword') }}</el-checkbox>
            </el-form-item>
            <el-form-item :label="$t('aiTools.agents.baseUrl')" prop="baseURL">
                <el-input v-model="form.baseURL" :disabled="form.provider !== 'ollama' && form.provider !== 'custom'" />
            </el-form-item>
            <el-form-item :label="$t('aiTools.model.model')" prop="model" v-if="form.provider === 'custom'">
                <el-input v-model="form.model" placeholder="gpt-4o-mini" />
            </el-form-item>
            <el-form-item :label="'API ' + $t('commons.table.type')" prop="apiType" v-if="form.provider === 'custom'">
                <el-select v-model="form.apiType">
                    <el-option label="openai-completions" value="openai-completions" />
                    <el-option label="openai-responses" value="openai-responses" />
                </el-select>
            </el-form-item>
            <el-form-item label="Max Tokens" prop="maxTokens" v-if="form.provider === 'custom'">
                <el-input-number v-model="form.maxTokens" :min="1" :max="2000000" />
            </el-form-item>
            <el-form-item label="Context Window" prop="contextWindow" v-if="form.provider === 'custom'">
                <el-input-number v-model="form.contextWindow" :min="1" :max="2000000" />
            </el-form-item>
            <el-form-item :label="$t('website.remark')" prop="remark">
                <el-input v-model="form.remark" />
            </el-form-item>
            <el-form-item v-if="form.id" prop="syncAgents">
                <el-checkbox v-model="form.syncAgents" :label="$t('aiTools.agents.syncAgents')" />
                <span class="input-help">{{ $t('aiTools.agents.syncAgentsHelper') }}</span>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="open = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button :disabled="loading" type="primary" @click="submit">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { FormInstance } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { createAgentAccount, getAgentProviders, updateAgentAccount } from '@/api/modules/ai';
import i18n from '@/lang';
import { getAgentProviderDisplayName } from '@/utils/agent';

const emit = defineEmits(['search']);

const open = ref(false);
const formRef = ref<FormInstance>();
const providerOptions = ref<Array<{ label: string; value: string }>>([]);
const providerBaseURL = ref<Record<string, string>>({});
const loading = ref(false);

const form = reactive({
    id: 0,
    provider: '',
    name: '',
    baseURL: '',
    model: '',
    apiType: 'openai-completions',
    maxTokens: 8192,
    contextWindow: 128000,
    apiKey: '',
    rememberApiKey: false,
    remark: '',
    syncAgents: false,
});

const headerTitle = computed(() =>
    form.id ? i18n.global.t('commons.button.edit') : i18n.global.t('aiTools.agents.createModelAccount'),
);

const rules = reactive({
    provider: [Rules.requiredSelect],
    name: [Rules.requiredInput],
    apiKey: [Rules.requiredInput],
    baseURL: [Rules.requiredInput],
    model: [Rules.requiredInput],
    apiType: [Rules.requiredSelect],
});

const submit = async () => {
    if (!formRef.value) {
        return;
    }
    await formRef.value.validate();
    loading.value = true;
    try {
        if (form.id) {
            await updateAgentAccount({
                id: form.id,
                name: form.name,
                baseURL: form.baseURL,
                apiKey: form.apiKey,
                rememberApiKey: form.rememberApiKey,
                model: form.model,
                apiType: form.apiType,
                maxTokens: form.maxTokens,
                contextWindow: form.contextWindow,
                remark: form.remark,
                syncAgents: form.syncAgents,
            });
        } else {
            await createAgentAccount({
                provider: form.provider,
                name: form.name,
                baseURL: form.baseURL,
                apiKey: form.apiKey,
                rememberApiKey: form.rememberApiKey,
                model: form.model,
                apiType: form.apiType,
                maxTokens: form.maxTokens,
                contextWindow: form.contextWindow,
                remark: form.remark,
            });
        }
        emit('search');
        open.value = false;
    } finally {
        loading.value = false;
    }
};

const handleClose = () => {
    formRef.value?.resetFields();
    loading.value = false;
    form.id = 0;
    form.model = '';
    form.apiType = 'openai-completions';
    form.maxTokens = 8192;
    form.contextWindow = 128000;
    form.rememberApiKey = false;
    form.syncAgents = false;
};

interface OpenParams {
    id?: number;
    provider?: string;
    name?: string;
    baseURL?: string;
    apiKey?: string;
    rememberApiKey?: boolean;
    model?: string;
    apiType?: string;
    maxTokens?: number;
    contextWindow?: number;
    remark?: string;
}

const openDrawer = async (params?: OpenParams) => {
    open.value = true;
    loading.value = false;
    if (params?.id) {
        form.id = params.id;
        form.provider = params.provider || '';
        form.name = params.name || '';
        form.baseURL = params.baseURL || '';
        form.apiKey = params.apiKey || '';
        form.rememberApiKey = params.rememberApiKey || false;
        form.model = params.model || '';
        form.apiType = params.apiType || 'openai-completions';
        form.maxTokens = params.maxTokens || 8192;
        form.contextWindow = params.contextWindow || 128000;
        form.remark = params.remark || '';
        form.syncAgents = false;
        return;
    }
    form.id = 0;
    form.name = '';
    form.baseURL = '';
    form.apiKey = '';
    form.rememberApiKey = false;
    form.model = '';
    form.apiType = 'openai-completions';
    form.maxTokens = 8192;
    form.contextWindow = 128000;
    form.remark = '';
    form.syncAgents = false;
    if (providerOptions.value.length === 0) {
        await loadProviders();
    }
    if (params?.provider) {
        form.provider = params.provider;
    } else if (providerOptions.value.length > 0) {
        form.provider = providerOptions.value[0].value;
    }
    handleProviderChange();
};

const loadProviders = async () => {
    const res = await getAgentProviders();
    const data = res.data || [];
    providerOptions.value = data.map((item) => ({
        value: item.provider,
        label: getAgentProviderDisplayName(item.provider, item.displayName),
    }));
    providerBaseURL.value = data.reduce((acc, item) => {
        acc[item.provider] = item.baseUrl || '';
        return acc;
    }, {} as Record<string, string>);
    if (!form.provider && providerOptions.value.length > 0) {
        form.provider = providerOptions.value[0].value;
        handleProviderChange();
    }
};

const handleProviderChange = () => {
    if (form.provider === 'custom') {
        form.baseURL = '';
        form.apiType = form.apiType || 'openai-completions';
        form.maxTokens = form.maxTokens || 8192;
        form.contextWindow = form.contextWindow || 128000;
        form.model = form.model || '';
        return;
    }
    if (form.provider !== 'ollama') {
        form.baseURL = providerBaseURL.value[form.provider] || '';
    } else {
        form.baseURL = '';
    }
};

onMounted(async () => {
    await loadProviders();
});

defineExpose({
    open: openDrawer,
});
</script>
