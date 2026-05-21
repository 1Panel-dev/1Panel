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
                    >
                        <div class="provider-option">
                            <ProviderLogo :provider="item.value" :display-name="item.label" />
                            <span>{{ item.label }}</span>
                        </div>
                    </el-option>
                </el-select>
            </el-form-item>
            <el-form-item label="API Key" prop="apiKey">
                <el-input v-model="form.apiKey" type="password" show-password />
                <span class="input-help" v-if="form.provider === 'custom' || form.provider === 'vllm'">
                    {{ $t('aiTools.agents.customProviderHelper') }}
                </span>
            </el-form-item>
            <el-form-item>
                <el-checkbox v-model="form.rememberApiKey">{{ $t('terminal.rememberPassword') }}</el-checkbox>
            </el-form-item>
            <el-form-item label="Base URL" prop="baseURL">
                <el-input v-model="form.baseURL" :disabled="!editableBaseURLProviders.includes(form.provider)" />
            </el-form-item>
            <el-form-item :label="'API ' + $t('commons.table.type')" prop="apiType">
                <el-select v-model="form.apiType" :disabled="apiTypeOptions.length === 1">
                    <el-option v-for="item in apiTypeOptions" :key="item" :label="item" :value="item" />
                </el-select>
            </el-form-item>
            <template v-if="showInitialModel">
                <el-divider content-position="left">{{ $t('aiTools.agents.modelPool') }}</el-divider>
                <el-form-item :label="$t('aiTools.model.model')" prop="initialModel.id" :rules="[Rules.noSpace]">
                    <el-input v-model="form.initialModel.id" />
                </el-form-item>
                <el-form-item :label="$t('commons.table.name')" prop="initialModel.name" :rules="[Rules.requiredInput]">
                    <el-input v-model="form.initialModel.name" />
                </el-form-item>
                <el-form-item label="Context Window" prop="initialModel.contextWindow" :rules="[Rules.integerNumber]">
                    <el-input-number v-model="form.initialModel.contextWindow" :min="1" :max="2000000" />
                </el-form-item>
                <el-form-item label="Max Tokens" prop="initialModel.maxTokens" :rules="[Rules.integerNumber]">
                    <el-input-number v-model="form.initialModel.maxTokens" :min="1" :max="2000000" />
                </el-form-item>
                <el-form-item
                    :label="$t('aiTools.agents.modelInputTypes')"
                    prop="initialModel.input"
                    :rules="[Rules.requiredSelect]"
                >
                    <el-checkbox-group v-model="form.initialModel.input">
                        <el-checkbox label="text">Text</el-checkbox>
                        <el-checkbox label="image">Image</el-checkbox>
                    </el-checkbox-group>
                </el-form-item>
                <el-form-item :label="$t('aiTools.agents.reasoning')" prop="initialModel.reasoning">
                    <el-switch v-model="form.initialModel.reasoning" />
                </el-form-item>
            </template>
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
import { computed, onMounted, reactive, ref } from 'vue';
import { FormInstance } from 'element-plus';
import { AI } from '@/api/interface/ai';
import { Rules } from '@/global/form-rules';
import { createAgentAccount, getAgentProviders, updateAgentAccount } from '@/api/modules/ai';
import i18n from '@/lang';
import { getAgentProviderDisplayName } from '@/utils/agent';
import { MsgError } from '@/utils/message';
import { useGlobalStore } from '@/composables/useGlobalStore';
import ProviderLogo from '@/components/agent-provider-logo/index.vue';

const emit = defineEmits(['search']);

const open = ref(false);
const formRef = ref<FormInstance>();
const providerOptions = ref<Array<{ label: string; value: string }>>([]);
const providerBaseURL = ref<Record<string, string>>({});
const loading = ref(false);
const editableBaseURLProviders = ['ollama', 'custom', 'vllm', 'zai'];
const initialModelProviders = ['custom', 'vllm', 'ollama'];
const { isIntl } = useGlobalStore();

const form = reactive({
    id: 0,
    provider: '',
    name: '',
    baseURL: '',
    apiType: 'openai-completions',
    apiKey: '',
    rememberApiKey: false,
    initialModel: {} as AI.AgentAccountModel,
    remark: '',
    syncAgents: false,
});

const headerTitle = computed(() =>
    form.id ? i18n.global.t('commons.button.edit') : i18n.global.t('commons.button.create'),
);

const showInitialModel = computed(() => !form.id && initialModelProviders.includes(form.provider));
const apiTypeOptions = computed(() => {
    if (form.provider === 'minimax' || form.provider === 'anthropic' || form.provider === 'kimi-coding') {
        return ['anthropic-messages'];
    }
    if (form.provider === 'custom' || form.provider === 'vllm') {
        return ['openai-completions', 'openai-responses', 'anthropic-messages'];
    }
    return ['openai-completions', 'openai-responses'];
});

const rules = reactive({
    provider: [Rules.requiredSelect],
    name: [Rules.requiredInput],
    apiKey: [Rules.requiredInput],
    baseURL: [Rules.requiredInput],
    apiType: [Rules.requiredSelect],
});

const isInitialModelProvider = (provider: string) => initialModelProviders.includes(provider);

const defaultContextWindowForProvider = (provider: string) => {
    if (provider === 'custom' || provider === 'vllm') {
        return 128000;
    }
    return 256000;
};

const buildInitialModel = (provider: string): AI.AgentAccountModel => ({
    recordId: 0,
    id: '',
    name: '',
    contextWindow: defaultContextWindowForProvider(provider),
    maxTokens: 8192,
    reasoning: false,
    input: ['text'],
});

const normalizeInitialModel = () => {
    const item = {
        recordId: 0,
        id: String(form.initialModel.id || '').trim(),
        name: String(form.initialModel.name || '').trim(),
        contextWindow: Number(form.initialModel.contextWindow || 0),
        maxTokens: Number(form.initialModel.maxTokens || 0),
        reasoning: Boolean(form.initialModel.reasoning),
        input: Array.from(
            new Set((form.initialModel.input || []).filter((value) => value === 'text' || value === 'image')),
        ),
    };
    if (
        item.id === '' ||
        item.name === '' ||
        item.contextWindow <= 0 ||
        item.maxTokens <= 0 ||
        item.input.length === 0
    ) {
        return null;
    }
    return item;
};

const resetInitialModel = () => {
    form.initialModel = isInitialModelProvider(form.provider)
        ? buildInitialModel(form.provider)
        : ({} as AI.AgentAccountModel);
};

const normalizeProviderAPIType = (provider: string, apiType?: string) => {
    if (provider === 'minimax' || provider === 'anthropic' || provider === 'kimi-coding') {
        return 'anthropic-messages';
    }
    if (provider === 'ollama') {
        return 'openai-responses';
    }
    if (provider === 'openai') {
        if (apiType === 'openai-completions' || apiType === 'openai-responses') {
            return apiType;
        }
        return 'openai-responses';
    }
    if (apiType) {
        return apiType;
    }
    return 'openai-completions';
};

const submit = async () => {
    if (!formRef.value) {
        return;
    }
    await formRef.value.validate();
    const initialModel = normalizeInitialModel();
    if (showInitialModel.value && !initialModel) {
        MsgError(i18n.global.t('aiTools.agents.accountModelsRequired'));
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
                models: initialModel ? [initialModel] : [],
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
    form.apiKey = '';
    form.rememberApiKey = false;
    form.initialModel = {} as AI.AgentAccountModel;
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
        form.apiType = normalizeProviderAPIType(form.provider, params.apiType);
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
    form.initialModel = {} as AI.AgentAccountModel;
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
    providerBaseURL.value = filteredData.reduce(
        (acc, item) => {
            acc[item.provider] = item.baseUrl || '';
            return acc;
        },
        {} as Record<string, string>,
    );
    if (!form.provider && providerOptions.value.length > 0) {
        form.provider = providerOptions.value[0].value;
        handleProviderChange();
    }
};

const handleProviderChange = () => {
    if (form.provider === 'custom' || form.provider === 'vllm') {
        form.baseURL = '';
    } else if (form.provider === 'ollama') {
        form.baseURL = '';
    } else {
        form.baseURL = providerBaseURL.value[form.provider] || '';
    }
    form.apiType = form.id
        ? normalizeProviderAPIType(form.provider, form.apiType)
        : normalizeProviderAPIType(form.provider);
    if (!form.id) {
        resetInitialModel();
    }
};

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
</style>
