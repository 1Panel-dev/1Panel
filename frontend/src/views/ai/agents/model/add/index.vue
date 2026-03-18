<template>
    <DrawerPro v-model="open" :header="headerTitle" size="large" @close="handleClose">
        <el-form ref="formRef" :model="form" :rules="rules" label-position="top" v-loading="loading">
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input v-model="form.name" />
            </el-form-item>
            <el-form-item :label="$t('aiTools.agents.provider')" prop="provider">
                <el-select v-model="form.provider" :disabled="form.id > 0" @change="handleProviderChange">
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
                <span class="input-help" v-if="form.provider === 'custom' || form.provider === 'vllm'">
                    {{ $t('aiTools.agents.customProviderHelper') }}
                </span>
            </el-form-item>
            <el-form-item>
                <el-checkbox v-model="form.rememberApiKey">{{ $t('terminal.rememberPassword') }}</el-checkbox>
            </el-form-item>
            <el-form-item :label="$t('aiTools.agents.baseUrl')" prop="baseURL">
                <el-input v-model="form.baseURL" :disabled="!editableBaseURLProviders.includes(form.provider)" />
            </el-form-item>
            <el-alert
                :title="$t('aiTools.agents.accountCreateHelper')"
                type="info"
                :closable="false"
                show-icon
                class="mb-4"
            />
            <el-form-item :label="'API ' + $t('commons.table.type')" prop="apiType">
                <el-select v-model="form.apiType">
                    <el-option label="openai-completions" value="openai-completions" />
                    <el-option label="openai-responses" value="openai-responses" />
                    <el-option
                        v-if="form.provider === 'custom' || form.provider === 'vllm'"
                        label="anthropic-messages"
                        value="anthropic-messages"
                    />
                </el-select>
            </el-form-item>
            <el-form-item label="Max Tokens" prop="maxTokens">
                <el-input-number v-model="form.maxTokens" :min="1" :max="2000000" />
            </el-form-item>
            <el-form-item label="Context Window" prop="contextWindow">
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
import { MsgError } from '@/utils/message';
import { useGlobalStore } from '@/composables/useGlobalStore';

const emit = defineEmits(['search']);

const open = ref(false);
const formRef = ref<FormInstance>();
const providerOptions = ref<Array<{ label: string; value: string }>>([]);
const providerBaseURL = ref<Record<string, string>>({});
const loading = ref(false);
const editableBaseURLProviders = ['ollama', 'custom', 'vllm', 'zai'];
const { isIntl } = useGlobalStore();

const form = reactive({
    id: 0,
    provider: '',
    name: '',
    baseURL: '',
    apiType: 'openai-completions',
    maxTokens: 8192,
    contextWindow: 128000,
    apiKey: '',
    rememberApiKey: false,
    remark: '',
    syncAgents: false,
});

const headerTitle = computed(() =>
    form.id ? i18n.global.t('commons.button.edit') : i18n.global.t('commons.button.create'),
);

const rules = reactive({
    provider: [Rules.requiredSelect],
    name: [Rules.requiredInput],
    apiKey: [Rules.requiredInput],
    baseURL: [Rules.requiredInput],
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
                apiType: form.apiType,
                maxTokens: form.maxTokens,
                contextWindow: form.contextWindow,
                remark: form.remark,
            });
        }
        emit('search');
        open.value = false;
    } catch (error: any) {
        MsgError(String(error?.message || i18n.global.t('commons.res.commonError')));
    } finally {
        loading.value = false;
    }
};

const handleClose = () => {
    formRef.value?.resetFields();
    loading.value = false;
    form.id = 0;
    form.provider = '';
    form.name = '';
    form.baseURL = '';
    form.apiType = 'openai-completions';
    form.maxTokens = 8192;
    form.contextWindow = 128000;
    form.apiKey = '';
    form.rememberApiKey = false;
    form.remark = '';
    form.syncAgents = false;
};

interface OpenParams {
    id?: number;
    provider?: string;
    name?: string;
    baseURL?: string;
    apiKey?: string;
    rememberApiKey?: boolean;
    apiType?: string;
    maxTokens?: number;
    contextWindow?: number;
    remark?: string;
}

const openDrawer = async (params?: OpenParams) => {
    open.value = true;
    loading.value = false;
    if (providerOptions.value.length === 0) {
        await loadProviders();
    }
    if (params?.id) {
        form.id = params.id;
        form.provider = params.provider || '';
        form.name = params.name || '';
        form.baseURL = params.baseURL || '';
        form.apiKey = params.apiKey || '';
        form.rememberApiKey = params.rememberApiKey || false;
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
    form.apiType = 'openai-completions';
    form.maxTokens = 8192;
    form.contextWindow = 128000;
    form.remark = '';
    form.syncAgents = false;
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
    const blockedProviders = new Set(['ark-coding-plan', 'bailian-coding-plan']);
    const filteredData = isIntl.value ? data.filter((item) => !blockedProviders.has(item.provider)) : data;
    providerOptions.value = filteredData.map((item) => ({
        value: item.provider,
        label: getAgentProviderDisplayName(item.provider, item.displayName),
    }));
    providerBaseURL.value = filteredData.reduce((acc, item) => {
        acc[item.provider] = item.baseUrl || '';
        return acc;
    }, {} as Record<string, string>);
    if (!form.provider && providerOptions.value.length > 0) {
        form.provider = providerOptions.value[0].value;
        handleProviderChange();
    }
};

const handleProviderChange = () => {
    if (form.provider === 'custom' || form.provider === 'vllm') {
        form.baseURL = '';
        form.apiType = form.apiType || 'openai-completions';
    } else if (form.provider === 'ollama') {
        form.baseURL = '';
        form.apiType = 'openai-responses';
    } else {
        form.baseURL = providerBaseURL.value[form.provider] || '';
        if (!form.apiType) {
            form.apiType = 'openai-completions';
        }
    }
};

onMounted(async () => {
    await loadProviders();
});

defineExpose({
    open: openDrawer,
});
</script>
