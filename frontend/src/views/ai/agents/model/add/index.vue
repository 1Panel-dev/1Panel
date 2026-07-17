<template>
    <DrawerPro v-model="open" :header="headerTitle" size="large" @close="handleClose">
        <el-form ref="formRef" :model="form" :rules="rules" label-position="top" autocomplete="off" v-loading="loading">
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
                    >
                        <div class="provider-option">
                            <ProviderLogo :provider="item.value" :display-name="item.label" />
                            <span>{{ item.label }}</span>
                        </div>
                    </el-option>
                </el-select>
            </el-form-item>
            <el-form-item :label="'API ' + $t('commons.table.type')" prop="apiType">
                <el-select v-model="form.apiType" :disabled="apiTypeOptions.length === 1" @change="handleAPITypeChange">
                    <el-option v-for="item in apiTypeOptions" :key="item" :label="item" :value="item" />
                </el-select>
            </el-form-item>
            <el-form-item v-if="authModeOptions.length > 1" :label="$t('terminal.authMode')" prop="authMode">
                <el-select v-model="form.authMode">
                    <el-option
                        v-for="item in authModeOptions"
                        :key="item"
                        :label="item === 'bearer' ? 'Bearer Token' : 'x-api-key'"
                        :value="item"
                    />
                </el-select>
            </el-form-item>
            <el-form-item label="Base URL" prop="baseURL">
                <el-input
                    v-model="form.baseURL"
                    name="agent-account-base-url"
                    autocomplete="off"
                    :disabled="!selectedAPIConfig?.editableBaseUrl"
                />
                <span v-if="apiTypeBaseURLHelper" class="input-help">
                    {{ apiTypeBaseURLHelper }}
                </span>
                <el-alert
                    v-if="baseURLAPITypeMismatchTip"
                    class="base-url-warning"
                    type="warning"
                    :closable="false"
                    show-icon
                >
                    {{ baseURLAPITypeMismatchTip }}
                </el-alert>
            </el-form-item>
            <el-form-item label="API Key" prop="apiKey">
                <el-input
                    v-model="form.apiKey"
                    name="agent-account-api-key"
                    type="password"
                    autocomplete="new-password"
                    show-password
                />
            </el-form-item>
            <el-form-item>
                <el-checkbox v-model="form.rememberApiKey">{{ $t('terminal.rememberPassword') }}</el-checkbox>
            </el-form-item>
            <template v-if="showModelDiscovery">
                <el-divider content-position="left">{{ $t('aiTools.agents.modelPool') }}</el-divider>
                <div class="model-discovery">
                    <div class="model-discovery__header">
                        <el-radio-group v-model="modelConfigMode" size="small" @change="handleModelConfigModeChange">
                            <el-radio-button value="discover">
                                {{ $t('aiTools.agents.automaticModelDiscovery') }}
                            </el-radio-button>
                            <el-radio-button value="manual">
                                {{ $t('aiTools.agents.manualModelConfiguration') }}
                            </el-radio-button>
                        </el-radio-group>
                        <el-button
                            v-if="modelConfigMode === 'discover'"
                            type="primary"
                            plain
                            :loading="discovering"
                            @click="discoverModels"
                        >
                            {{ $t('aiTools.agents.discoverModels') }}
                        </el-button>
                    </div>
                    <div v-if="modelConfigMode === 'discover'" class="input-help model-discovery__help">
                        {{ $t('aiTools.agents.discoverModelsHelper') }}
                    </div>
                    <div v-if="discoveredModels.length" class="input-help model-discovery__help">
                        {{ $t('aiTools.agents.verifyModelHelper') }}
                    </div>
                    <el-alert
                        v-if="modelDiscoveryFailed"
                        type="warning"
                        :closable="false"
                        show-icon
                        :title="$t('aiTools.agents.discoverModelsFailedFallback')"
                    />
                    <el-table
                        v-if="modelConfigMode === 'discover' && discoveredModels.length"
                        :data="discoveredModels"
                        border
                        max-height="320"
                    >
                        <el-table-column :label="$t('aiTools.agents.verifyModel')" width="110" align="center">
                            <template #default="{ row }">
                                <el-radio v-model="form.verifyModel" :value="row.id" />
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('aiTools.model.model')" prop="id" min-width="220" />
                        <el-table-column :label="$t('commons.table.name')" prop="name" min-width="180" />
                    </el-table>
                </div>
            </template>
            <template v-if="showInitialModel">
                <el-divider v-if="!showModelDiscovery" content-position="left">
                    {{ $t('aiTools.agents.modelPool') }}
                </el-divider>
                <el-form-item :label="$t('aiTools.model.model')" prop="initialModel.id" :rules="[Rules.noSpace]">
                    <el-input v-model="form.initialModel.id" />
                </el-form-item>
                <el-form-item :label="$t('commons.table.name')" prop="initialModel.name">
                    <el-input v-model="form.initialModel.name" />
                </el-form-item>
            </template>
            <el-form-item
                v-if="showVerifyModelSelect"
                :label="$t('aiTools.agents.verifyModel')"
                prop="verifyModel"
                :rules="[Rules.requiredSelect]"
            >
                <el-select v-model="form.verifyModel" filterable>
                    <el-option
                        v-for="item in verifyModelOptions"
                        :key="item.id"
                        :label="item.name || item.id"
                        :value="item.id"
                    />
                </el-select>
                <span class="input-help">{{ $t('aiTools.agents.verifyModelHelper') }}</span>
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
                <el-button v-permission :disabled="loading" type="primary" @click="submit">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { FormInstance } from 'element-plus';
import { AI } from '@/api/interface/ai';
import { Rules } from '@/global/form-rules';
import {
    createAgentAccount,
    discoverAgentAccountModels,
    getAgentProviders,
    updateAgentAccount,
} from '@/api/modules/ai';
import i18n from '@/lang';
import { getAgentProviderDisplayName, isAgentAccountVerificationSkipped } from '@/utils/agent';
import { MsgError } from '@/utils/message';
import { useGlobalStore } from '@/composables/useGlobalStore';
import ProviderLogo from '@/components/agent-provider-logo/index.vue';

const emit = defineEmits(['search']);

const open = ref(false);
const formRef = ref<FormInstance>();
const providerOptions = ref<Array<{ label: string; value: string }>>([]);
const providers = ref<Record<string, AI.ProviderInfo>>({});
const loading = ref(false);
const discovering = ref(false);
const discoveredModels = ref<AI.AgentAccountModel[]>([]);
const editModels = ref<AI.AgentAccountModel[]>([]);
const modelConfigMode = ref<'discover' | 'manual'>('discover');
const modelDiscoveryFailed = ref(false);
const initialModelProviders = ['custom', 'vllm', 'ollama'];
const { isIntl } = useGlobalStore();

const form = reactive({
    id: 0,
    provider: 'custom',
    name: '',
    baseURL: '',
    apiType: 'openai-completions',
    authMode: '',
    apiKey: '',
    rememberApiKey: false,
    verifyModel: '',
    initialModel: {} as AI.AgentAccountModel,
    remark: '',
    syncAgents: false,
});

const headerTitle = computed(() =>
    form.id ? i18n.global.t('commons.button.edit') : i18n.global.t('commons.button.create'),
);

const selectedProvider = computed(() => providers.value[form.provider]);
const selectedAPIConfig = computed(() =>
    selectedProvider.value?.apiTypes.find((item) => item.apiType === form.apiType),
);
const showModelDiscovery = computed(
    () =>
        !form.id &&
        form.provider === 'custom' &&
        (form.apiType === 'openai-completions' || form.apiType === 'openai-responses'),
);
const showInitialModel = computed(
    () =>
        !form.id &&
        initialModelProviders.includes(form.provider) &&
        (!showModelDiscovery.value || modelConfigMode.value === 'manual'),
);
const verifyModelOptions = computed(() => (form.id ? editModels.value : selectedProvider.value?.models || []));
const showVerifyModelSelect = computed(
    () =>
        !isAgentAccountVerificationSkipped(form.provider) &&
        verifyModelOptions.value.length > 0 &&
        (Boolean(form.id) || (!showModelDiscovery.value && !showInitialModel.value)),
);
const apiTypeOptions = computed(() => selectedProvider.value?.apiTypes.map((item) => item.apiType) || []);
const authModeOptions = computed(() => selectedAPIConfig.value?.authModes || []);
const apiTypeURLHints: Record<string, { requestPath: string; example: string; endpointSuffixes: string[] }> = {
    'openai-completions': {
        requestPath: '/v1/chat/completions',
        example: 'http://127.0.0.1:8000/v1',
        endpointSuffixes: ['/v1/chat/completions', '/v1beta/chat/completions', '/chat/completions'],
    },
    'openai-responses': {
        requestPath: '/v1/responses',
        example: 'http://127.0.0.1:8000/v1',
        endpointSuffixes: ['/v1/responses', '/responses'],
    },
    'anthropic-messages': {
        requestPath: '/v1/messages',
        example: 'http://127.0.0.1:8000',
        endpointSuffixes: ['/v1/messages', '/messages'],
    },
};
const showAPITypeBaseURLTips = computed(() => Boolean(selectedAPIConfig.value?.editableBaseUrl));
const apiTypeBaseURLHelper = computed(() => {
    const hint = apiTypeURLHints[form.apiType];
    if (!showAPITypeBaseURLTips.value || !hint) {
        return '';
    }
    return i18n.global.t('aiTools.agents.apiTypeBaseURLHelper', [hint.requestPath, hint.example]);
});

const normalizeBaseURLPath = (baseURL: string) => {
    const value = baseURL.trim().replace(/\/+$/, '');
    if (!value) {
        return '';
    }
    try {
        return (new URL(value).pathname || '/').replace(/\/+$/, '').toLowerCase();
    } catch {
        return value.split(/[?#]/)[0].replace(/\/+$/, '').toLowerCase();
    }
};

const detectAPITypeFromBaseURL = (baseURL: string) => {
    const path = normalizeBaseURLPath(baseURL);
    if (!path) {
        return '';
    }
    for (const [apiType, hint] of Object.entries(apiTypeURLHints)) {
        if (hint.endpointSuffixes.some((suffix) => path.endsWith(suffix))) {
            return apiType;
        }
    }
    return '';
};
const baseURLAPITypeMismatchTip = computed(() => {
    if (!showAPITypeBaseURLTips.value || !form.baseURL) {
        return '';
    }
    const detectedAPIType = detectAPITypeFromBaseURL(form.baseURL);
    const expected = apiTypeURLHints[form.apiType];
    if (!detectedAPIType || detectedAPIType === form.apiType || !expected) {
        return '';
    }
    return i18n.global.t('aiTools.agents.apiTypeBaseURLMismatch', [detectedAPIType, form.apiType, expected.example]);
});

const rules = reactive({
    provider: [Rules.requiredSelect],
    name: [Rules.requiredInput],
    apiKey: [Rules.requiredInput],
    baseURL: [Rules.requiredInput],
    apiType: [Rules.requiredSelect],
});

const isInitialModelProvider = (provider: string) => initialModelProviders.includes(provider);

const buildInitialModel = (): AI.AgentAccountModel => ({
    recordId: 0,
    id: '',
    name: '',
});

const normalizeInitialModel = () => {
    const item = {
        recordId: 0,
        id: String(form.initialModel.id || '').trim(),
        name: String(form.initialModel.name || '').trim(),
    };
    if (item.id === '') {
        return null;
    }
    return item;
};

const resetInitialModel = () => {
    form.initialModel = isInitialModelProvider(form.provider) ? buildInitialModel() : ({} as AI.AgentAccountModel);
};

const resetModelDiscovery = () => {
    discovering.value = false;
    discoveredModels.value = [];
    modelConfigMode.value = 'discover';
    modelDiscoveryFailed.value = false;
    form.verifyModel = '';
};

const handleModelConfigModeChange = () => {
    modelDiscoveryFailed.value = false;
    form.verifyModel = '';
    if (modelConfigMode.value === 'manual') {
        discoveredModels.value = [];
    }
};

const discoverModels = async () => {
    if (!formRef.value) {
        return;
    }
    try {
        await formRef.value.validateField(['apiKey', 'baseURL']);
    } catch {
        return;
    }
    discovering.value = true;
    try {
        const res = await discoverAgentAccountModels({
            provider: form.provider,
            baseURL: form.baseURL,
            apiKey: form.apiKey,
            apiType: form.apiType,
        });
        discoveredModels.value = res.data || [];
        form.verifyModel = '';
        if (discoveredModels.value.length === 0) {
            modelConfigMode.value = 'manual';
            modelDiscoveryFailed.value = true;
        } else {
            modelDiscoveryFailed.value = false;
        }
    } catch (error: any) {
        discoveredModels.value = [];
        modelConfigMode.value = 'manual';
        modelDiscoveryFailed.value = true;
        MsgError(String(error?.message || i18n.global.t('commons.res.commonError')));
    } finally {
        discovering.value = false;
    }
};

const submit = async () => {
    if (!formRef.value) {
        return;
    }
    await formRef.value.validate();
    const initialModel = showInitialModel.value ? normalizeInitialModel() : null;
    if (showInitialModel.value && !initialModel) {
        MsgError(i18n.global.t('aiTools.agents.accountModelsRequired'));
        return;
    }
    if (showModelDiscovery.value && modelConfigMode.value === 'discover' && discoveredModels.value.length === 0) {
        MsgError(i18n.global.t('aiTools.agents.accountModelsRequired'));
        return;
    }
    const verifyModel = showInitialModel.value ? initialModel?.id || '' : form.verifyModel;
    if (!isAgentAccountVerificationSkipped(form.provider) && !verifyModel) {
        MsgError(i18n.global.t('aiTools.agents.verifyModelRequired'));
        return;
    }
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
                authMode: form.authMode,
                verifyModel,
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
                authMode: form.authMode,
                verifyModel,
                models: discoveredModels.value.length ? discoveredModels.value : initialModel ? [initialModel] : [],
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
    form.provider = 'custom';
    form.name = '';
    form.baseURL = '';
    form.apiType = 'openai-completions';
    form.authMode = '';
    form.apiKey = '';
    form.rememberApiKey = false;
    form.verifyModel = '';
    form.initialModel = {} as AI.AgentAccountModel;
    editModels.value = [];
    form.remark = '';
    form.syncAgents = false;
    resetModelDiscovery();
};

interface OpenParams {
    id?: number;
    provider?: string;
    name?: string;
    baseURL?: string;
    apiKey?: string;
    rememberApiKey?: boolean;
    apiType?: string;
    authMode?: string;
    verifyModel?: string;
    models?: AI.AgentAccountModel[];
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
        editModels.value = params.models || [];
        const provider = providers.value[form.provider];
        form.apiType = provider?.apiTypes.some((item) => item.apiType === params.apiType)
            ? String(params.apiType)
            : provider?.defaultApiType || '';
        form.authMode = params.authMode || selectedAPIConfig.value?.defaultAuthMode || '';
        form.verifyModel = params.verifyModel || editModels.value[0]?.id || '';
        form.remark = params.remark || '';
        form.syncAgents = false;
        return;
    }
    form.id = 0;
    form.name = '';
    form.baseURL = '';
    form.apiKey = '';
    form.rememberApiKey = false;
    form.verifyModel = '';
    editModels.value = [];
    form.apiType = 'openai-completions';
    form.authMode = '';
    form.initialModel = {} as AI.AgentAccountModel;
    form.remark = '';
    form.syncAgents = false;
    resetModelDiscovery();
    if (params?.provider) {
        form.provider = params.provider;
    } else {
        form.provider = providers.value.custom ? 'custom' : providerOptions.value[0]?.value || '';
    }
    handleProviderChange();
};

const loadProviders = async () => {
    const res = await getAgentProviders();
    const data = res.data || [];
    const blockedProviders = new Set(['ark-coding-plan', 'bailian-coding-plan']);
    const filteredData = isIntl.value ? data.filter((item) => !blockedProviders.has(item.provider)) : data;
    providers.value = Object.fromEntries(filteredData.map((item) => [item.provider, item]));
    providerOptions.value = filteredData.map((item) => ({
        value: item.provider,
        label: getAgentProviderDisplayName(item.provider, item.displayName),
    }));
    if (!form.provider && providerOptions.value.length > 0) {
        form.provider = providers.value.custom ? 'custom' : providerOptions.value[0].value;
        handleProviderChange();
    }
};

const handleProviderChange = () => {
    resetModelDiscovery();
    const provider = selectedProvider.value;
    form.apiType = provider?.defaultApiType || '';
    const config = provider?.apiTypes.find((item) => item.apiType === form.apiType);
    form.baseURL = config?.baseUrl || '';
    form.authMode = config?.defaultAuthMode || '';
    if (!form.id) {
        resetInitialModel();
    }
};

const handleAPITypeChange = () => {
    resetModelDiscovery();
    const config = selectedAPIConfig.value;
    if (!config) {
        return;
    }
    form.authMode = config.defaultAuthMode || '';
    const providerURLs = selectedProvider.value?.apiTypes.map((item) => item.baseUrl).filter(Boolean) || [];
    if (!config.editableBaseUrl || !form.baseURL || providerURLs.includes(form.baseURL)) {
        form.baseURL = config.baseUrl || '';
    }
    resetInitialModel();
};

watch([() => form.apiKey, () => form.baseURL], () => {
    if (showModelDiscovery.value) {
        discoveredModels.value = [];
        modelDiscoveryFailed.value = false;
    }
});

onMounted(async () => {
    await loadProviders();
});

defineExpose({
    open: openDrawer,
});
</script>

<style scoped lang="scss">
.provider-option {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    min-width: 0;
    white-space: nowrap;

    span {
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }
}

.base-url-warning {
    margin-top: 8px;
}

.model-discovery {
    margin-bottom: 18px;
}

.model-discovery__header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 12px;
}

.model-discovery__help {
    margin-bottom: 12px;
}

.model-discovery .el-alert {
    margin-bottom: 12px;
}
</style>
