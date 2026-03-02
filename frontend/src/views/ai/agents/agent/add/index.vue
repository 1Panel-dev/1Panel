<template>
    <DrawerPro v-model="open" :header="$t('commons.button.create')" size="large" @close="handleClose">
        <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
            <el-card class="form-card">
                <el-form-item :label="$t('commons.table.name')" prop="name">
                    <el-input v-model="form.name" />
                </el-form-item>
                <el-form-item :label="`${$t('aiTools.agents.agents')}${$t('commons.table.type')}`" prop="agentType">
                    <el-select v-model="form.agentType" @change="handleAgentTypeChange">
                        <el-option :label="$t('aiTools.agents.openclawType')" value="openclaw" />
                        <el-option :label="$t('aiTools.agents.copawType')" value="copaw" />
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('aiTools.agents.appVersion')" prop="appVersion">
                    <el-select v-model="form.appVersion" filterable>
                        <el-option v-for="item in versions" :key="item" :label="item" :value="item" />
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('aiTools.agents.webuiPort')" prop="webUIPort">
                    <el-input-number v-model="form.webUIPort" :min="1" :max="65535" />
                </el-form-item>
                <el-form-item
                    v-if="form.agentType === 'openclaw'"
                    :label="$t('aiTools.agents.bridgePort')"
                    prop="bridgePort"
                >
                    <el-input-number v-model="form.bridgePort" :min="1" :max="65535" />
                </el-form-item>
            </el-card>
            <el-card class="form-card" v-if="form.agentType === 'openclaw'">
                <el-form-item :label="$t('aiTools.agents.provider')" prop="provider">
                    <el-select v-model="form.provider" @change="handleProviderChange">
                        <el-option
                            v-for="item in providerOptions"
                            :key="item.value"
                            :label="item.label"
                            :value="item.value"
                        >
                            <div class="option-row">
                                <span class="option-label">{{ item.label }}</span>
                                <el-tag
                                    size="small"
                                    :type="providerAccountCount[item.value] > 0 ? 'success' : 'info'"
                                    class="option-tag"
                                >
                                    {{ $t('aiTools.agents.accountCount', [providerAccountCount[item.value] || 0]) }}
                                </el-tag>
                            </div>
                        </el-option>
                    </el-select>
                </el-form-item>

                <el-form-item :label="$t('aiTools.agents.account')" prop="accountId">
                    <el-select v-model="form.accountId" @change="handleAccountChange">
                        <el-option v-for="item in accountOptions" :key="item.id" :label="item.name" :value="item.id" />
                    </el-select>
                    <span class="input-help">
                        {{ $t('aiTools.agents.noAccountHint') }}
                        <el-button type="primary" link class="inline-link" @click="openAccountCreate">
                            {{ $t('commons.button.create') }}
                        </el-button>
                    </span>
                </el-form-item>
                <el-form-item>
                    <el-checkbox v-model="manualModel">{{ $t('aiTools.agents.manualModel') }}</el-checkbox>
                </el-form-item>
                <el-form-item :label="$t('aiTools.model.model')" prop="model">
                    <el-input v-if="manualModel" v-model="form.model" />

                    <el-select v-else v-model="form.model" filterable @change="handleModelChange">
                        <el-option v-for="item in filteredModels" :key="item.id" :label="item.name" :value="item.id" />
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('aiTools.agents.baseUrl')" v-if="form.accountId" prop="baseURL">
                    <el-input v-model="form.baseURL" disabled />
                </el-form-item>
                <el-form-item :label="$t('aiTools.agents.token')">
                    <el-input v-model="form.token" disabled>
                        <template #append>
                            <CopyButton :content="form.token" />
                        </template>
                    </el-input>
                </el-form-item>
            </el-card>
            <el-card class="form-card">
                <AdvancedSetting :form="form" />
            </el-card>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="open = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="submit">{{ $t('commons.button.confirm') }}</el-button>
            </span>
        </template>
    </DrawerPro>
    <AccountAddDialog ref="accountAddRef" @search="handleAccountCreated" />
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { FormInstance } from 'element-plus';
import { checkNumberRange, Rules } from '@/global/form-rules';
import { createAgent, getAgentProviders, pageAgentAccounts } from '@/api/modules/ai';
import { AI } from '@/api/interface/ai';
import { getAppByKey, getAppDetail } from '@/api/modules/app';
import { getRandomStr, newUUID } from '@/utils/util';
import { getAgentProviderDisplayName } from '@/utils/agent';
import { App } from '@/api/interface/app';
import AdvancedSetting from '@/components/advanced-setting/index.vue';
import AccountAddDialog from '@/views/ai/agents/model/add/index.vue';

const emit = defineEmits(['search', 'task']);

const open = ref(false);
const formRef = ref<FormInstance>();
const versions = ref<string[]>([]);
const accountOptions = ref<AI.AgentAccountItem[]>([]);
const providerOptions = ref<Array<{ label: string; value: string }>>([]);
const providerModels = ref<Record<string, AI.ProviderModelInfo[]>>({});
const providerAccountCount = ref<Record<string, number>>({});
const manualModel = ref(false);
const appInfo = ref<App.AppDTO>();
const accountAddRef = ref();

const form = reactive({
    name: '',
    agentType: 'openclaw' as 'openclaw' | 'copaw',
    appVersion: '',
    webUIPort: 18789,
    bridgePort: 18790,
    provider: 'deepseek',
    accountId: undefined as unknown as number,
    model: '',
    apiType: 'openai-completions',
    maxTokens: 8192,
    contextWindow: 128000,
    apiKey: '',
    baseURL: '',
    token: '',
    advanced: true,
    containerName: '',
    allowPort: true,
    specifyIP: '',
    restartPolicy: 'unless-stopped',
    cpuQuota: 0,
    memoryLimit: 0,
    memoryUnit: 'M',
    pullImage: true,
    editCompose: false,
    dockerCompose: '',
});

const validateOpenclawOnly = (field: 'provider' | 'accountId' | 'model' | 'bridgePort') => {
    return (_rule: any, value: any, callback: (error?: Error) => void) => {
        if (form.agentType !== 'openclaw') {
            callback();
            return;
        }
        if (field === 'bridgePort') {
            if (!value || Number(value) <= 0) {
                callback(new Error('bridge port is required'));
                return;
            }
            callback();
            return;
        }
        if (value === undefined || value === null || value === '') {
            callback(new Error(`${field} is required`));
            return;
        }
        callback();
    };
};

const rules = reactive({
    name: [Rules.requiredInput],
    agentType: [Rules.requiredSelect],
    appVersion: [Rules.requiredSelect],
    webUIPort: [Rules.requiredInput],
    bridgePort: [{ validator: validateOpenclawOnly('bridgePort'), trigger: 'blur' }],
    provider: [{ validator: validateOpenclawOnly('provider'), trigger: 'change' }],
    accountId: [{ validator: validateOpenclawOnly('accountId'), trigger: 'change' }],
    model: [{ validator: validateOpenclawOnly('model'), trigger: 'change' }],
    containerName: [Rules.containerName],
    restartPolicy: [Rules.requiredSelect],
    cpuQuota: [checkNumberRange(0, 99999)],
    memoryLimit: [checkNumberRange(0, 9999999999)],
    specifyIP: [Rules.ipv4orV6],
});

const filteredModels = computed(() => providerModels.value[form.provider] || []);

const loadVersions = async (appKey: 'openclaw' | 'copaw') => {
    const res = await getAppByKey(appKey);
    appInfo.value = res.data;
    versions.value = res.data.versions || [];
    if (!form.appVersion && versions.value.length > 0) {
        form.appVersion = versions.value[0];
    }
};

const loadCompose = async () => {
    if (!appInfo.value || !form.appVersion) {
        return;
    }
    const res = await getAppDetail(appInfo.value.id, form.appVersion, appInfo.value.type);
    form.dockerCompose = res.data.dockerCompose || '';
};

const loadProviders = async () => {
    if (form.agentType !== 'openclaw') {
        providerOptions.value = [];
        providerModels.value = {};
        return;
    }
    const res = await getAgentProviders();
    const data = res.data || [];
    providerOptions.value = data.map((item) => ({
        value: item.provider,
        label: getAgentProviderDisplayName(item.provider, item.displayName),
    }));
    providerModels.value = data.reduce((acc, item) => {
        acc[item.provider] = item.models || [];
        return acc;
    }, {} as Record<string, AI.ProviderModelInfo[]>);
    if (!providerOptions.value.find((item) => item.value === form.provider) && providerOptions.value.length > 0) {
        form.provider = providerOptions.value[0].value;
    }
    await loadProviderAccountCounts(providerOptions.value.map((item) => item.value));
    setDefaultModel();
};

const loadProviderAccountCounts = async (providers: string[]) => {
    const tasks = providers.map(async (provider) => {
        const req: AI.AgentAccountSearch = {
            page: 1,
            pageSize: 1,
            provider: provider,
            name: '',
        };
        const res = await pageAgentAccounts(req);
        providerAccountCount.value[provider] = res.data.total || 0;
    });
    await Promise.all(tasks);
};

const loadAccounts = async () => {
    if (form.agentType !== 'openclaw') {
        accountOptions.value = [];
        return;
    }
    if (!form.provider) {
        accountOptions.value = [];
        return;
    }
    const req: AI.AgentAccountSearch = {
        page: 1,
        pageSize: 200,
        provider: form.provider,
        name: '',
    };
    const res = await pageAgentAccounts(req);
    accountOptions.value = res.data.items || [];
    providerAccountCount.value[form.provider] = res.data.total || accountOptions.value.length;
    if (accountOptions.value.length > 0) {
        form.accountId = accountOptions.value[0].id;
        handleAccountChange();
    } else {
        form.accountId = undefined as unknown as number;
        form.apiKey = '';
        form.baseURL = '';
    }
};

const handleProviderChange = () => {
    if (form.agentType !== 'openclaw') {
        return;
    }
    form.model = '';
    form.apiKey = '';
    form.baseURL = '';
    form.accountId = undefined as unknown as number;
    loadAccounts();
    setDefaultModel();
};

const handleAgentTypeChange = async () => {
    if (form.name === '' || form.name === 'OpenClaw' || form.name === 'CoPaw') {
        form.name = form.agentType === 'copaw' ? 'CoPaw' : 'OpenClaw';
    }
    form.appVersion = '';
    form.model = '';
    form.provider = 'deepseek';
    form.accountId = undefined as unknown as number;
    form.apiKey = '';
    form.baseURL = '';
    form.apiType = 'openai-completions';
    if (form.agentType === 'openclaw') {
        form.bridgePort = form.bridgePort || 18790;
        await loadVersions('openclaw');
        await loadProviders();
        await loadAccounts();
        return;
    }
    await loadVersions('copaw');
};

const handleModelChange = () => {
    if (manualModel.value) {
        return;
    }
};

const handleAccountChange = () => {
    if (form.agentType !== 'openclaw') {
        return;
    }
    const selected = accountOptions.value.find((item) => item.id === form.accountId);
    if (selected) {
        form.baseURL = selected.baseUrl || '';
        form.apiKey = selected.apiKey || '';
        form.apiType = selected.apiType || 'openai-completions';
        form.maxTokens = selected.maxTokens || 8192;
        form.contextWindow = selected.contextWindow || 128000;
        if (selected.provider === 'custom' && selected.model && !manualModel.value) {
            form.model = selected.model;
        }
    }
    setDefaultModel();
};

const setDefaultModel = () => {
    if (form.agentType !== 'openclaw') {
        return;
    }
    if (manualModel.value) {
        return;
    }
    const models = filteredModels.value;
    if (models.length > 0 && !form.model) {
        form.model = models[0].id;
        return;
    }
    if (form.provider === 'custom') {
        const selected = accountOptions.value.find((item) => item.id === form.accountId);
        if (selected?.model && !form.model) {
            form.model = selected.model;
        }
    }
};

const submit = async () => {
    if (!formRef.value) {
        return;
    }
    await formRef.value.validate();
    const taskID = newUUID();
    if (form.agentType === 'openclaw' && !form.token) {
        form.token = getRandomStr(32).toLowerCase();
    }
    try {
        const res = await createAgent({
            name: form.name,
            appVersion: form.appVersion,
            webUIPort: form.webUIPort,
            bridgePort: form.agentType === 'openclaw' ? form.bridgePort : undefined,
            agentType: form.agentType,
            provider: form.agentType === 'openclaw' ? form.provider : undefined,
            model: form.agentType === 'openclaw' ? form.model : undefined,
            apiType: form.agentType === 'openclaw' ? form.apiType : undefined,
            maxTokens: form.agentType === 'openclaw' ? form.maxTokens : undefined,
            contextWindow: form.agentType === 'openclaw' ? form.contextWindow : undefined,
            accountId: form.agentType === 'openclaw' ? form.accountId : undefined,
            apiKey: form.agentType === 'openclaw' ? form.apiKey : undefined,
            baseURL: form.agentType === 'openclaw' ? form.baseURL : undefined,
            token: form.agentType === 'openclaw' ? form.token : undefined,
            taskID: taskID,
            advanced: form.advanced,
            containerName: form.containerName,
            allowPort: form.allowPort,
            specifyIP: form.specifyIP,
            restartPolicy: form.restartPolicy,
            cpuQuota: form.cpuQuota,
            memoryLimit: form.memoryLimit,
            memoryUnit: form.memoryUnit,
            pullImage: form.pullImage,
            editCompose: form.editCompose,
            dockerCompose: form.dockerCompose,
        });
        form.token = res.data.token || form.token;
        emit('search');
        emit('task', taskID);
        open.value = false;
    } catch (error: any) {
        const message = String(error?.message || '').toLowerCase();
        const isTimeout = message.includes('timeout') || error?.code === 'ECONNABORTED';
        if (isTimeout) {
            emit('task', taskID);
            open.value = false;
        }
    }
};

const handleClose = () => {
    formRef.value?.resetFields();
    form.token = '';
    form.dockerCompose = '';
};

const openDrawer = async (agentType?: 'openclaw' | 'copaw') => {
    const targetType = agentType === 'copaw' ? 'copaw' : 'openclaw';
    form.name = targetType === 'copaw' ? 'CoPaw' : 'OpenClaw';
    open.value = true;
    manualModel.value = false;
    form.agentType = targetType;
    form.token = getRandomStr(32).toLowerCase();
    if (form.agentType === 'copaw') {
        await loadVersions('copaw');
        providerOptions.value = [];
        providerModels.value = {};
        accountOptions.value = [];
        return;
    }
    await loadVersions('openclaw');
    await loadProviders();
    await loadAccounts();
};

const openAccountCreate = () => {
    if (form.agentType !== 'openclaw') {
        return;
    }
    if (accountAddRef.value?.open) {
        accountAddRef.value.open({ provider: form.provider });
    }
};

const handleAccountCreated = async () => {
    await loadAccounts();
};

watch(
    () => form.editCompose,
    async (value) => {
        if (value && !form.dockerCompose) {
            await loadCompose();
        }
    },
);

watch(
    () => form.appVersion,
    async (value, oldValue) => {
        if (!value || value === oldValue) {
            return;
        }
        if (form.editCompose) {
            await loadCompose();
        }
    },
);

defineExpose({
    open: openDrawer,
});
</script>

<style scoped>
.form-card {
    margin-bottom: 16px;
}

.inline-link {
    padding: 0;
    margin-top: -1px;
    margin-left: 5px;
    height: auto;
    line-height: inherit;
    font-size: inherit;
}

.option-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
}

.option-label {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.option-tag {
    flex-shrink: 0;
}
</style>
