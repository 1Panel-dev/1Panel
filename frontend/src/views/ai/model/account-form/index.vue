<template>
    <component
        :is="inline ? 'div' : DrawerPro"
        v-bind="inline ? {} : { modelValue: open, header: headerTitle, size: 'large' }"
        @update:model-value="open = $event"
        @close="handleClose"
    >
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
            <el-form-item
                v-if="accountTypeOptions.length > 1"
                :label="$t('aiTools.model.accountType')"
                prop="accountType"
            >
                <el-select v-model="form.accountType" :disabled="form.id > 0" @change="handleAccountTypeChange">
                    <el-option
                        v-for="item in accountTypeOptions"
                        :key="item.key"
                        :label="$t(`aiTools.model.accountTypes.${item.key}`)"
                        :value="item.key"
                    />
                </el-select>
            </el-form-item>
            <el-form-item v-if="canSelectLocalEngine" :label="$t('ssh.createMode')">
                <el-radio-group v-model="connectionMode" @change="handleConnectionModeChange">
                    <el-radio-button value="manual">{{ $t('ssh.input') }}</el-radio-button>
                    <el-radio-button value="local">{{ $t('commons.button.select') }}</el-radio-button>
                </el-radio-group>
            </el-form-item>
            <el-form-item
                v-if="connectionMode === 'local'"
                :label="$t('aiTools.model.engines')"
                prop="localEngineId"
                :rules="[Rules.requiredSelect]"
            >
                <el-select
                    v-model="form.localEngineId"
                    fit-input-width
                    filterable
                    :loading="loadingEngines"
                    @change="handleLocalEngineChange"
                >
                    <el-option v-for="item in localEngines" :key="item.id" :label="item.name" :value="item.id">
                        <div class="local-engine-option">
                            <span class="local-engine-option__name" :title="item.name">{{ item.name }}</span>
                            <AiModelList
                                v-if="item.models.length"
                                class="local-engine-option__models"
                                :models="item.models"
                                :visible-count="1"
                                tag-type="info"
                            />
                        </div>
                    </el-option>
                </el-select>
                <el-alert
                    v-if="enginesLoadFailed"
                    class="mt-2"
                    type="error"
                    :closable="false"
                    :title="$t('aiTools.model.loadFailed')"
                />
            </el-form-item>
            <el-form-item :label="'API ' + $t('commons.table.type')" prop="apiType">
                <el-select
                    v-model="form.apiType"
                    :disabled="form.id > 0 || apiTypeOptions.length === 1"
                    @change="handleAPITypeChange"
                >
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
                <span class="input-help">{{ $t('aiTools.model.apiKeyHelper') }}</span>
            </el-form-item>
            <el-form-item>
                <el-checkbox v-model="form.rememberApiKey">{{ $t('terminal.rememberPassword') }}</el-checkbox>
            </el-form-item>
            <el-form-item>
                <el-switch v-model="form.validateAvailability" />
                <el-text class="validate-availability-label">{{ $t('aiTools.agents.validateAvailability') }}</el-text>
                <span class="input-help">
                    {{
                        isImageAPIType
                            ? $t('aiTools.model.verifyImageOnChangeHelper')
                            : $t('aiTools.model.verifyOnChangeHelper')
                    }}
                </span>
            </el-form-item>
            <ModelEditor
                v-if="showModelPool"
                ref="modelEditorRef"
                v-model="manualModels"
                v-model:verify-index="manualVerifyIndex"
                v-model:discovering="discovering"
                :supports-discovery="supportsModelDiscovery"
                :request="discoveryRequest"
                :validate-availability="form.validateAvailability"
                :single-model-only="singleModelOnly"
                :local-models="localModelOptions"
                :disabled="loading"
                :before-discover="validateDiscovery"
            />
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
            <el-form-item v-if="form.id" prop="syncAgents">
                <el-checkbox v-model="form.syncAgents" :label="$t('aiTools.agents.syncAgents')" />
                <span class="input-help">{{ $t('aiTools.agents.syncAgentsHelper') }}</span>
            </el-form-item>
            <el-form-item :label="$t('website.remark')" prop="remark">
                <el-input v-model="form.remark" />
            </el-form-item>
        </el-form>
        <template v-if="!inline" #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="open = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button v-permission :disabled="loading || disabled || discovering" type="primary" @click="submit">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </component>
</template>

<script setup lang="ts">
import ModelEditor from '../model-editor/index.vue';
import AiModelList from '@/components/ai-model-list/index.vue';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { FormInstance } from 'element-plus';
import { AI } from '@/api/interface/ai';
import { Rules } from '@/global/form-rules';
import { createAgentAccount, getAgentProviders, searchOllamaModel, updateAgentAccount } from '@/api/modules/ai';
import { checkAppInstalled } from '@/api/modules/app';
import i18n from '@/lang';
import { getAgentProviderDisplayName } from '@/utils/agent';
import { MsgError } from '@/utils/message';
import { useGlobalStore } from '@/composables/useGlobalStore';
import ProviderLogo from '@/components/agent-provider-logo/index.vue';

import DrawerPro from '@/components/drawer-pro/index.vue';
import { readPages } from '@/utils/pagination';
import type { ReqPage, ResPage, ResultData } from '@/api/interface';

const vllmModules = import.meta.glob<{
    pageVllm: (req: ReqPage & { info: string }) => Promise<
        ResultData<
            ResPage<{
                id: number;
                name: string;
                modelAccountBaseURL: string;
                servedModelName: string;
            }>
        >
    >;
}>('@/xpack/api/modules/vllm.ts');
const vllmLoader = Object.values(vllmModules)[0];

const props = defineProps({ inline: Boolean });
const emit = defineEmits(['search', 'close']);

const open = ref(false);
const formRef = ref<FormInstance>();
const providerOptions = computed(() =>
    Object.values(providers.value)
        .filter(
            (item) =>
                !['bailian-coding-plan', 'ark-coding-plan', 'kimi-coding', 'moonshot'].includes(item.provider) ||
                (form.id > 0 && item.provider === form.provider),
        )
        .map((item) => ({
            value: item.provider,
            label: getAgentProviderDisplayName(item.provider, item.displayName),
        })),
);
const providers = ref<Record<string, AI.ProviderInfo>>({});
const loading = ref(false);
const discovering = ref(false);
const modelEditorRef = ref<InstanceType<typeof ModelEditor>>();
const manualModels = ref<AI.AgentAccountModel[]>([]);
const manualVerifyIndex = ref(-1);
const editModels = ref<AI.AgentAccountModel[]>([]);
const { isIntl, isProductPro } = useGlobalStore();
const connectionMode = ref<'manual' | 'local'>('manual');
const loadingEngines = ref(false);
const enginesLoadFailed = ref(false);
const localEngines = ref<Array<{ id: number; name: string; baseURL: string; models: string[] }>>([]);

const form = reactive({
    id: 0,
    appInstallId: 0,
    localEngineId: undefined as number | undefined,
    provider: 'custom',
    accountType: '',
    name: '',
    baseURL: '',
    apiType: 'openai-completions',
    authMode: '',
    apiKey: '',
    rememberApiKey: false,
    verifyModel: '',
    validateAvailability: true,
    remark: '',
    syncAgents: false,
});

const headerTitle = computed(() =>
    form.id ? i18n.global.t('commons.button.edit') : i18n.global.t('aiTools.model.connectExternal'),
);

const selectedProvider = computed(() => providers.value[form.provider]);
const accountTypeOptions = computed(() =>
    (selectedProvider.value?.accountTypes || []).filter(
        (item) => !isIntl.value || !['bailian', 'ark'].includes(form.provider) || item.key !== 'codingPlan',
    ),
);
const selectedAccountType = computed(
    () =>
        selectedProvider.value?.accountTypes.find((item) => item.key === form.accountType) ||
        accountTypeOptions.value[0],
);
const availableAPIConfigs = computed(
    () => selectedAccountType.value?.apiTypes || selectedProvider.value?.apiTypes || [],
);
const singleModelOnly = computed(() => !form.id && ['vllm', 'llmman'].includes(form.provider));
const canSelectLocalEngine = computed(() => !form.id && ['ollama', 'vllm'].includes(form.provider));
const localModelOptions = computed(() =>
    connectionMode.value === 'local'
        ? localEngines.value.find((item) => item.id === form.localEngineId)?.models || []
        : [],
);
const disabled = computed(
    () =>
        discovering.value ||
        (connectionMode.value === 'local' &&
            (loadingEngines.value || !localEngines.value.some((item) => item.id === form.localEngineId))),
);
const selectedAPIConfig = computed(() => availableAPIConfigs.value.find((item) => item.apiType === form.apiType));
const isImageAPIType = computed(() => form.apiType.endsWith('-images'));
const supportsModelDiscovery = computed(() => Boolean(selectedAPIConfig.value?.supportsModelDiscovery));
const showModelPool = computed(() => !form.appInstallId);
const verifyModelOptions = computed(() => (form.id ? editModels.value : manualModels.value));
const showVerifyModelSelect = computed(
    () => form.validateAvailability && verifyModelOptions.value.length > 0 && !showModelPool.value,
);
const apiTypeOptions = computed(() => availableAPIConfigs.value.map((item) => item.apiType) || []);
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
    'openai-images': {
        requestPath: '/images/generations',
        example: 'http://127.0.0.1:8000/v1',
        endpointSuffixes: ['/v1/images/generations', '/images/generations'],
    },
    'dashscope-images': {
        requestPath: '/api/v1/services/aigc/multimodal-generation/generation',
        example: 'https://dashscope.aliyuncs.com',
        endpointSuffixes: ['/api/v1/services/aigc/multimodal-generation/generation'],
    },
    'minimax-images': {
        requestPath: '/v1/image_generation',
        example: 'https://api.minimaxi.com',
        endpointSuffixes: ['/v1/image_generation'],
    },
    'openrouter-images': {
        requestPath: '/api/v1/images',
        example: 'https://openrouter.ai',
        endpointSuffixes: ['/api/v1/images'],
    },
    'openai-embeddings': {
        requestPath: '/v1/embeddings',
        example: 'http://127.0.0.1:8000/v1',
        endpointSuffixes: ['/v1/embeddings', '/embeddings'],
    },
};
const showAPITypeBaseURLTips = computed(() => Boolean(selectedAPIConfig.value?.editableBaseUrl));
const apiTypeBaseURLHelper = computed(() => {
    const hint = apiTypeURLHints[form.apiType];
    if (!showAPITypeBaseURLTips.value || !hint) {
        return '';
    }
    if (form.provider === 'custom' && isImageAPIType.value) {
        return i18n.global.t('aiTools.agents.customImageURLHelper', [
            `${hint.example.replace(/\/+$/, '')}${hint.requestPath}`,
        ]);
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

const buildInitialModel = (): AI.AgentAccountModel => ({
    recordId: 0,
    id: '',
    name: '',
});

const resetManualModels = () => {
    manualModels.value = localModelOptions.value.length
        ? localModelOptions.value
              .slice(0, singleModelOnly.value ? 1 : undefined)
              .map((id) => ({ recordId: 0, id, name: id }))
        : [buildInitialModel()];
    manualVerifyIndex.value = singleModelOnly.value ? 0 : -1;
};

const resetModelDiscovery = () => {
    discovering.value = false;
    resetManualModels();
};

const discoveryRequest = computed(() => ({
    provider: form.provider,
    accountType: form.accountType,
    apiType: form.apiType,
    baseURL: form.baseURL,
    apiKey: form.apiKey,
    authMode: form.authMode,
}));
const validateDiscovery = async () => {
    try {
        await formRef.value?.validateField(['apiKey', 'baseURL']);
        return true;
    } catch {
        return false;
    }
};

const submit = async () => {
    if (!formRef.value || disabled.value || discovering.value) {
        return;
    }
    await formRef.value.validate();
    const accountModels = showModelPool.value ? modelEditorRef.value?.validate() : [];
    if (!accountModels) return;
    const verifyModel = showModelPool.value ? accountModels[manualVerifyIndex.value]?.id || '' : form.verifyModel;
    loading.value = true;
    try {
        if (form.id) {
            await updateAgentAccount({
                id: form.id,
                accountType: form.accountType,
                name: form.name,
                baseURL: form.baseURL,
                apiKey: form.apiKey,
                rememberApiKey: form.rememberApiKey,
                apiType: form.apiType,
                authMode: form.authMode,
                verifyModel,
                validateAvailability: form.validateAvailability,
                remark: form.remark,
                syncAgents: form.syncAgents,
                models: showModelPool.value ? accountModels : undefined,
            });
        } else {
            await createAgentAccount({
                provider: form.provider,
                accountType: form.accountType,
                name: form.name,
                baseURL: form.baseURL,
                apiKey: form.apiKey,
                rememberApiKey: form.rememberApiKey,
                apiType: form.apiType,
                authMode: form.authMode,
                verifyModel,
                validateAvailability: form.validateAvailability,
                models: accountModels,
                remark: form.remark,
            });
        }
        emit('search');
        open.value = false;
    } catch (error: any) {
        MsgError(
            String(
                error?.response?.data?.message ||
                    (error?.code === 'ECONNABORTED' ? i18n.global.t('commons.msg.requestTimeout') : error?.message) ||
                    i18n.global.t('commons.res.commonError'),
            ),
        );
    } finally {
        loading.value = false;
    }
};

const handleClose = () => {
    emit('close');
    formRef.value?.resetFields();
    loading.value = false;
    form.id = 0;
    form.appInstallId = 0;
    form.provider = 'custom';
    form.accountType = '';
    form.name = '';
    form.baseURL = '';
    form.apiType = 'openai-completions';
    form.authMode = '';
    form.apiKey = '';
    form.rememberApiKey = false;
    form.validateAvailability = true;
    form.verifyModel = '';
    editModels.value = [];
    form.remark = '';
    form.syncAgents = false;
    resetModelDiscovery();
};

interface OpenParams {
    id?: number;
    appInstallId?: number;
    provider?: string;
    accountType?: string;
    name?: string;
    baseURL?: string;
    apiKey?: string;
    rememberApiKey?: boolean;
    apiType?: string;
    authMode?: string;
    verifyModel?: string;
    validateAvailability?: boolean;
    models?: AI.AgentAccountModel[];
    remark?: string;
}

const openDrawer = async (params?: OpenParams) => {
    connectionMode.value = 'manual';
    form.localEngineId = undefined;
    open.value = true;
    loading.value = false;
    if (providerOptions.value.length === 0) {
        await loadProviders();
    }
    if (params?.id) {
        form.id = params.id;
        form.appInstallId = params.appInstallId || 0;
        form.provider = params.provider || '';
        form.accountType = params.accountType || '';
        form.name = params.name || '';
        form.baseURL = params.baseURL || '';
        form.apiKey = params.apiKey || '';
        form.rememberApiKey = params.rememberApiKey || false;
        form.validateAvailability = params.validateAvailability ?? true;
        editModels.value = params.models || [];
        const provider = providers.value[form.provider];
        form.apiType = availableAPIConfigs.value.some((item) => item.apiType === params.apiType)
            ? String(params.apiType)
            : provider?.defaultApiType || '';
        if (selectedAPIConfig.value && !selectedAPIConfig.value.editableBaseUrl) {
            form.baseURL = selectedAPIConfig.value.baseUrl;
        }
        form.authMode = params.authMode || selectedAPIConfig.value?.defaultAuthMode || '';
        resetModelDiscovery();
        manualModels.value = editModels.value.length
            ? editModels.value.map((item) => ({ ...item }))
            : [buildInitialModel()];
        form.verifyModel = params.verifyModel || editModels.value[0]?.id || '';
        manualVerifyIndex.value = manualModels.value.findIndex((item) => item.id === form.verifyModel);
        form.remark = params.remark || '';
        form.syncAgents = false;
        return;
    }
    form.id = 0;
    form.appInstallId = 0;
    form.name = '';
    form.baseURL = '';
    form.apiKey = '';
    form.rememberApiKey = false;
    form.validateAvailability = true;
    form.verifyModel = '';
    editModels.value = [];
    form.apiType = 'openai-completions';
    form.authMode = '';
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
    providers.value = Object.fromEntries(data.map((item) => [item.provider, item]));
    if (!form.provider && providerOptions.value.length > 0) {
        form.provider = providers.value.custom ? 'custom' : providerOptions.value[0].value;
        handleProviderChange();
    }
};

const handleProviderChange = () => {
    form.validateAvailability = true;
    if (connectionMode.value === 'local') {
        form.apiKey = '';
        form.rememberApiKey = false;
    }
    connectionMode.value = 'manual';
    form.localEngineId = undefined;
    form.accountType = accountTypeOptions.value[0]?.key || '';
    handleAccountTypeChange();
};

const handleAccountTypeChange = () => {
    const configs = availableAPIConfigs.value;
    const config = configs.find((item) => item.apiType === selectedProvider.value?.defaultApiType) || configs[0];
    form.apiType = config?.apiType || '';
    form.baseURL = config?.baseUrl || '';
    form.authMode = config?.defaultAuthMode || '';
    resetModelDiscovery();
};

const handleConnectionModeChange = () => {
    form.validateAvailability = true;
    form.localEngineId = undefined;
    form.baseURL = selectedAPIConfig.value?.baseUrl || '';
    form.apiKey = '';
    form.rememberApiKey = false;
    resetModelDiscovery();
    formRef.value?.clearValidate();
};

const handleLocalEngineChange = () => {
    const engine = localEngines.value.find((item) => item.id === form.localEngineId);
    if (!engine) return;
    if (!form.name) form.name = engine.name;
    form.baseURL = engine.baseURL;
    form.apiKey = `${form.provider}-local`;
    form.rememberApiKey = true;
    resetModelDiscovery();
    formRef.value?.clearValidate();
};

watch([open, () => form.provider, connectionMode], async ([visible, provider, mode], _, onCleanup) => {
    let cancelled = false;
    onCleanup(() => (cancelled = true));
    localEngines.value = [];
    enginesLoadFailed.value = false;
    loadingEngines.value = false;
    if (!visible || mode !== 'local') return;
    loadingEngines.value = true;
    try {
        let engines: typeof localEngines.value = [];
        if (provider === 'vllm' && isProductPro.value && vllmLoader) {
            const api = await vllmLoader();
            if (cancelled) return;
            const items = await readPages((page) => api.pageVllm({ ...page, info: '' }));
            engines = items.map((item) => ({
                id: item.id,
                name: item.name,
                baseURL: item.modelAccountBaseURL,
                models: item.servedModelName ? [item.servedModelName] : [],
            }));
        } else if (provider === 'ollama') {
            const { data: item } = await checkAppInstalled('ollama', '');
            if (cancelled) return;
            if (item.isExist) {
                const models = await readPages<AI.OllamaModelInfo>((page) => searchOllamaModel(page));
                if (cancelled) return;
                engines = [
                    {
                        id: item.appInstallId,
                        name: item.name,
                        baseURL: item.modelAccountBaseURL || `http://${item.containerName}:11434/v1`,
                        models: models.filter((model) => model.status === 'Success').map((model) => model.name),
                    },
                ];
            }
        }
        if (!cancelled) localEngines.value = engines;
    } catch {
        if (!cancelled) enginesLoadFailed.value = true;
    } finally {
        if (!cancelled) loadingEngines.value = false;
    }
});

const handleAPITypeChange = () => {
    const config = selectedAPIConfig.value;
    if (!config) {
        return;
    }
    form.authMode = config.defaultAuthMode || '';
    const providerURLs = availableAPIConfigs.value.map((item) => item.baseUrl).filter(Boolean) || [];
    if (!config.editableBaseUrl || !form.baseURL || providerURLs.includes(form.baseURL)) {
        form.baseURL = config.baseUrl || '';
    }
    resetModelDiscovery();
};

watch(open, (value) => {
    if (props.inline && !value) handleClose();
});

onMounted(async () => {
    if (!props.inline) await loadProviders();
});

defineExpose({
    open: openDrawer,
    submit,
    loading,
    disabled,
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

.local-engine-option {
    display: flex;
    align-items: center;
    gap: 8px;
    min-width: 0;

    &__name {
        flex: 1;
        min-width: 0;
        overflow: hidden;
        text-overflow: ellipsis;
    }

    &__models {
        max-width: 65%;
    }
}

.base-url-warning {
    margin-top: 8px;
}

.validate-availability-label {
    margin-left: 12px;
}
</style>
