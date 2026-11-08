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
            <template v-if="showModelPool">
                <el-divider content-position="left">{{ $t('aiTools.agents.modelPool') }}</el-divider>
                <div class="model-discovery">
                    <div class="model-discovery__header">
                        <el-radio-group v-model="modelConfigMode" size="small" @change="handleModelConfigModeChange">
                            <el-radio-button v-if="supportsModelDiscovery || defaultModels.length" value="discover">
                                {{
                                    $t(
                                        supportsModelDiscovery
                                            ? 'aiTools.agents.automaticModelDiscovery'
                                            : 'aiTools.agents.defaultModel',
                                    )
                                }}
                            </el-radio-button>
                            <el-radio-button value="manual">
                                {{ $t('aiTools.agents.manualModelConfiguration') }}
                            </el-radio-button>
                        </el-radio-group>
                        <el-button
                            v-if="supportsModelDiscovery && modelConfigMode === 'discover'"
                            v-permission="'ai_model_manage'"
                            type="primary"
                            plain
                            :loading="discovering"
                            @click="discoverModels"
                        >
                            {{ $t('aiTools.agents.discoverModels') }}
                        </el-button>
                    </div>
                    <div
                        v-if="supportsModelDiscovery && modelConfigMode === 'discover'"
                        class="input-help model-discovery__help"
                    >
                        {{ $t('aiTools.agents.discoverModelsHelper') }}
                    </div>
                    <div
                        v-if="form.validateAvailability && (modelConfigMode === 'manual' || discoveredModels.length)"
                        class="input-help model-discovery__help"
                    >
                        {{ $t('aiTools.agents.verifyModelHelper') }}
                    </div>
                    <el-alert
                        v-if="modelDiscoveryFailed"
                        type="warning"
                        :closable="false"
                        show-icon
                        :title="
                            $t(
                                defaultModels.length
                                    ? 'aiTools.agents.discoverModelsFailedUseDefaults'
                                    : 'aiTools.agents.discoverModelsFailedFallback',
                            )
                        "
                    />
                    <el-table
                        v-if="modelConfigMode === 'discover' && discoveredModels.length"
                        :data="discoveredModels"
                        max-height="320"
                    >
                        <el-table-column
                            v-if="form.validateAvailability"
                            :label="$t('aiTools.model.verify')"
                            width="90"
                            align="center"
                        >
                            <template #default="{ row }">
                                <el-radio
                                    v-model="form.verifyModel"
                                    class="verification-radio"
                                    :value="row.id"
                                    :aria-label="$t('aiTools.agents.verifyModel') + ': ' + row.id"
                                >
                                    <span />
                                </el-radio>
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('aiTools.model.model')" prop="id" min-width="220" />
                        <el-table-column :label="$t('commons.table.name')" prop="name" min-width="180" />
                    </el-table>
                    <template v-if="modelConfigMode === 'manual'">
                        <el-table :data="manualModels" max-height="320">
                            <el-table-column
                                v-if="form.validateAvailability"
                                :label="$t('aiTools.model.verify')"
                                width="90"
                                align="center"
                            >
                                <template #default="{ row, $index }">
                                    <el-radio
                                        v-model="manualVerifyIndex"
                                        class="verification-radio"
                                        :value="$index"
                                        :disabled="!row.id.trim()"
                                        :aria-label="$t('aiTools.agents.verifyModel') + ': ' + row.id"
                                    >
                                        <span />
                                    </el-radio>
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('aiTools.model.model')" min-width="220">
                                <template #default="{ row }">
                                    <el-select
                                        v-if="localModelOptions.length"
                                        v-model="row.id"
                                        filterable
                                        allow-create
                                        default-first-option
                                    >
                                        <el-option
                                            v-for="model in localModelOptions"
                                            :key="model"
                                            :label="model"
                                            :value="model"
                                        />
                                    </el-select>
                                    <el-input v-else v-model="row.id" />
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('commons.table.name')" min-width="180">
                                <template #default="{ row }">
                                    <el-input v-model="row.name" />
                                </template>
                            </el-table-column>
                            <el-table-column
                                v-if="!singleModelOnly"
                                :label="$t('commons.table.operate')"
                                width="90"
                                align="center"
                            >
                                <template #default="{ $index }">
                                    <el-button
                                        type="primary"
                                        link
                                        :disabled="
                                            form.validateAvailability &&
                                            Boolean(form.id) &&
                                            manualVerifyIndex === $index
                                        "
                                        @click="removeManualModel($index)"
                                    >
                                        {{ $t('commons.button.delete') }}
                                    </el-button>
                                </template>
                            </el-table-column>
                        </el-table>
                        <el-button
                            v-if="!singleModelOnly"
                            class="model-discovery__add"
                            type="primary"
                            plain
                            @click="addManualModel"
                        >
                            {{ $t('commons.button.add') }}
                        </el-button>
                    </template>
                </div>
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
                <el-button v-permission :disabled="loading || disabled" type="primary" @click="submit">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </component>
</template>

<script setup lang="ts">
import AiModelList from '@/components/ai-model-list/index.vue';
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { FormInstance } from 'element-plus';
import { AI } from '@/api/interface/ai';
import { Rules } from '@/global/form-rules';
import {
    createAgentAccount,
    discoverAgentAccountModels,
    getAgentProviders,
    searchOllamaModel,
    updateAgentAccount,
} from '@/api/modules/ai';
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
const providerOptions = ref<Array<{ label: string; value: string }>>([]);
const providers = ref<Record<string, AI.ProviderInfo>>({});
const loading = ref(false);
const discovering = ref(false);
const discoveredModels = ref<AI.AgentAccountModel[]>([]);
const manualModels = ref<AI.AgentAccountModel[]>([]);
const manualVerifyIndex = ref(-1);
const editModels = ref<AI.AgentAccountModel[]>([]);
const modelConfigMode = ref<'discover' | 'manual'>('discover');
const modelDiscoveryFailed = ref(false);
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
const singleModelOnly = computed(() => !form.id && ['vllm', 'llmman'].includes(form.provider));
const canSelectLocalEngine = computed(() => !form.id && ['ollama', 'vllm'].includes(form.provider));
const localModelOptions = computed(() =>
    connectionMode.value === 'local'
        ? localEngines.value.find((item) => item.id === form.localEngineId)?.models || []
        : [],
);
const disabled = computed(
    () =>
        connectionMode.value === 'local' &&
        (loadingEngines.value || !localEngines.value.some((item) => item.id === form.localEngineId)),
);
const selectedAPIConfig = computed(() =>
    selectedProvider.value?.apiTypes.find((item) => item.apiType === form.apiType),
);
const isImageAPIType = computed(() => form.apiType.endsWith('-images'));
const supportsModelDiscovery = computed(() => Boolean(selectedAPIConfig.value?.supportsModelDiscovery));
const defaultModels = computed(() => {
    if (selectedAPIConfig.value?.models.length) {
        return selectedAPIConfig.value.models;
    }
    return isImageAPIType.value ? [] : selectedProvider.value?.models || [];
});
const showModelPool = computed(() => !form.appInstallId);
const verifyModelOptions = computed(() => (form.id ? editModels.value : defaultModels.value));
const showVerifyModelSelect = computed(
    () => form.validateAvailability && verifyModelOptions.value.length > 0 && !showModelPool.value,
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

const loadDefaultModels = () => {
    discoveredModels.value = defaultModels.value.map((item) => ({
        recordId: 0,
        id: item.id,
        name: item.name,
    }));
    modelDiscoveryFailed.value = false;
    form.verifyModel = '';
};

const resetModelDiscovery = () => {
    discovering.value = false;
    modelConfigMode.value = supportsModelDiscovery.value || defaultModels.value.length ? 'discover' : 'manual';
    resetManualModels();
    loadDefaultModels();
};

const handleModelConfigModeChange = () => {
    modelDiscoveryFailed.value = false;
    form.verifyModel = '';
};

const addManualModel = () => {
    if (singleModelOnly.value && manualModels.value.length > 0) {
        return;
    }
    manualModels.value.push(buildInitialModel());
    if (singleModelOnly.value) {
        manualVerifyIndex.value = 0;
    }
};

const removeManualModel = (index: number) => {
    manualModels.value.splice(index, 1);
    if (manualVerifyIndex.value === index) {
        manualVerifyIndex.value = -1;
    } else if (manualVerifyIndex.value > index) {
        manualVerifyIndex.value -= 1;
    }
};

const normalizeManualModels = () => {
    const models = manualModels.value.map((item) => ({
        recordId: 0,
        id: String(item.id || '').trim(),
        name: String(item.name || '').trim(),
    }));
    if (models.some((item) => !item.id)) {
        MsgError(i18n.global.t('aiTools.agents.accountModelsRequired'));
        return null;
    }
    if (models.some((item) => item.id.includes(' '))) {
        MsgError(i18n.global.t('setting.noSpace'));
        return null;
    }
    if (new Set(models.map((item) => item.id)).size !== models.length) {
        MsgError(i18n.global.t('aiTools.agents.accountModelsDuplicate'));
        return null;
    }
    return models;
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
            if (defaultModels.value.length > 0) {
                loadDefaultModels();
            } else {
                modelConfigMode.value = 'manual';
            }
            modelDiscoveryFailed.value = true;
        } else {
            modelDiscoveryFailed.value = false;
        }
    } catch (error: any) {
        if (defaultModels.value.length > 0) {
            loadDefaultModels();
            modelDiscoveryFailed.value = true;
        } else {
            discoveredModels.value = [];
            modelConfigMode.value = 'manual';
        }
        modelDiscoveryFailed.value = true;
        MsgError(String(error?.message || i18n.global.t('commons.res.commonError')));
    } finally {
        discovering.value = false;
    }
};

const submit = async () => {
    if (!formRef.value || disabled.value) {
        return;
    }
    await formRef.value.validate();
    let accountModels: AI.AgentAccountModel[] = [];
    if (showModelPool.value && modelConfigMode.value === 'manual') {
        const models = normalizeManualModels();
        if (!models) {
            return;
        }
        accountModels = models;
    } else if (showModelPool.value) {
        accountModels = discoveredModels.value;
    }
    if ((!form.id || showModelPool.value) && accountModels.length === 0) {
        MsgError(i18n.global.t('aiTools.agents.accountModelsRequired'));
        return;
    }
    if (singleModelOnly.value && accountModels.length > 1) {
        MsgError(i18n.global.t('aiTools.agents.accountModelsSingle'));
        return;
    }
    let verifyModel = form.verifyModel;
    if (modelConfigMode.value === 'manual' && showModelPool.value) {
        verifyModel = accountModels[manualVerifyIndex.value]?.id || '';
    }
    if (form.validateAvailability && !verifyModel && (!form.id || editModels.value.length > 0)) {
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
                validateAvailability: form.validateAvailability,
                remark: form.remark,
                syncAgents: form.syncAgents,
                models: showModelPool.value ? accountModels : undefined,
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
        form.name = params.name || '';
        form.baseURL = params.baseURL || '';
        form.apiKey = params.apiKey || '';
        form.rememberApiKey = params.rememberApiKey || false;
        form.validateAvailability = params.validateAvailability ?? true;
        editModels.value = params.models || [];
        const provider = providers.value[form.provider];
        form.apiType = provider?.apiTypes.some((item) => item.apiType === params.apiType)
            ? String(params.apiType)
            : provider?.defaultApiType || '';
        form.authMode = params.authMode || selectedAPIConfig.value?.defaultAuthMode || '';
        resetModelDiscovery();
        modelConfigMode.value = 'manual';
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
    form.validateAvailability = true;
    if (connectionMode.value === 'local') {
        form.apiKey = '';
        form.rememberApiKey = false;
    }
    connectionMode.value = 'manual';
    form.localEngineId = undefined;
    const provider = selectedProvider.value;
    form.apiType = provider?.defaultApiType || '';
    const config = provider?.apiTypes.find((item) => item.apiType === form.apiType);
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
    const providerURLs = selectedProvider.value?.apiTypes.map((item) => item.baseUrl).filter(Boolean) || [];
    if (!config.editableBaseUrl || !form.baseURL || providerURLs.includes(form.baseURL)) {
        form.baseURL = config.baseUrl || '';
    }
    resetModelDiscovery();
};

watch([() => form.apiKey, () => form.baseURL], () => {
    if (!form.id && showModelPool.value) {
        loadDefaultModels();
    }
});

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

.model-discovery__add {
    display: inline-block;
    margin-top: 12px;
}

.verification-radio {
    margin-right: 0;
    height: 32px;

    :deep(.el-radio__label) {
        display: none;
    }
}
</style>
