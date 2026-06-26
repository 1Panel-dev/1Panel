<template>
    <DrawerPro v-model="open" :header="$t('commons.button.create')" size="large" @close="handleClose">
        <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
            <el-card class="form-card">
                <el-form-item :label="`${$t('aiTools.agents.agent')}${$t('commons.table.type')}`" prop="agentType">
                    <el-select v-model="form.agentType" @change="handleAgentTypeChange">
                        <el-option label="OpenClaw" value="openclaw" />
                        <el-option label="Hermes Agent" value="hermes-agent" />
                        <el-option label="QwenPaw" value="copaw" />
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('commons.table.name')" prop="name">
                    <el-input v-model="form.name" />
                </el-form-item>
                <el-form-item :label="$t('website.remark')" prop="remark">
                    <el-input v-model="form.remark" />
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
                    :label="$t('aiTools.agents.allowedOrigins')"
                    prop="allowedOrigins"
                >
                    <el-input
                        v-model="form.allowedOrigins"
                        type="textarea"
                        :rows="3"
                        :placeholder="allowedOriginsPlaceholder"
                        @input="handleAllowedOriginsInput"
                    />
                </el-form-item>
            </el-card>
            <el-card class="form-card" v-if="showModelConfig">
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
                        <el-button v-permission type="primary" link class="inline-link" @click="openAccountCreate">
                            {{ $t('commons.button.create') }}
                        </el-button>
                    </span>
                </el-form-item>
                <el-form-item :label="$t('aiTools.model.model')" prop="model">
                    <el-select v-model="form.model" filterable>
                        <el-option v-for="item in filteredModels" :key="item.id" :label="item.name" :value="item.id" />
                    </el-select>
                    <span class="input-help">{{ $t('aiTools.agents.accountModelsHelper') }}</span>
                </el-form-item>
                <el-form-item label="Base URL" v-if="form.accountId" prop="baseURL">
                    <el-input v-model="form.baseURL" disabled />
                </el-form-item>
                <el-form-item v-if="form.agentType === 'openclaw'" label="Token">
                    <el-input v-model="form.token" disabled>
                        <template #append>
                            <CopyButton :content="form.token" />
                        </template>
                    </el-input>
                </el-form-item>
                <template v-if="form.agentType === 'hermes-agent'">
                    <el-form-item :label="$t('aiTools.agents.dashboardUsername')" prop="dashboardUsername">
                        <el-input v-model="form.dashboardUsername" />
                    </el-form-item>
                    <el-form-item :label="$t('aiTools.agents.dashboardPassword')" prop="dashboardPassword">
                        <el-input v-model="form.dashboardPassword" type="password" show-password>
                            <template #append>
                                <el-space>
                                    <CopyButton :content="form.dashboardPassword" />
                                    <el-button link type="primary" @click="generateDashboardPassword">
                                        {{ $t('commons.button.random') }}
                                    </el-button>
                                </el-space>
                            </template>
                        </el-input>
                    </el-form-item>
                </template>
            </el-card>
            <el-card class="form-card">
                <AdvancedSetting :form="form" />
            </el-card>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="open = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button v-permission type="primary" @click="submit">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
    <AccountAddDialog ref="accountAddRef" @search="handleAccountCreated" />
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { FormInstance } from 'element-plus';
import { checkNumberRange, Rules } from '@/global/form-rules';
import { countAgentAccountsByProvider, createAgent, getAgentProviders, pageAgentAccounts } from '@/api/modules/ai';
import { AI } from '@/api/interface/ai';
import { getAppByKey, getAppDetail } from '@/api/modules/app';
import { getAgentSettingInfo } from '@/api/modules/setting';
import { getRandomStr, newUUID } from '@/utils/id';
import {
    buildDefaultAllowedOrigin,
    getAgentProviderDisplayName,
    parseAllowedOriginsInput,
    validateAllowedOriginsInput,
} from '@/utils/agent';
import { App } from '@/api/interface/app';
import AdvancedSetting from '@/components/advanced-setting/index.vue';
import AccountAddDialog from '@/views/ai/agents/model/add/index.vue';
import { useGlobalStore } from '@/composables/useGlobalStore';

const emit = defineEmits(['search', 'task']);

const open = ref(false);
const formRef = ref<FormInstance>();
const versions = ref<string[]>([]);
const accountOptions = ref<AI.AgentAccountItem[]>([]);
const providerOptions = ref<Array<{ label: string; value: string }>>([]);
const providerModels = ref<Record<string, AI.ProviderModelInfo[]>>({});
const providerAccountCount = ref<Record<string, number>>({});
const appInfo = ref<App.AppDTO>();
const accountAddRef = ref();
const systemIP = ref('');
const lastAutoAllowedOrigins = ref('');
const allowedOriginsAutoFilled = ref(true);
const { isIntl } = useGlobalStore();

const form = reactive({
    name: '',
    remark: '',
    agentType: 'openclaw' as AI.AgentType,
    appVersion: '',
    webUIPort: 18789,
    allowedOrigins: '',
    provider: '',
    accountId: undefined as unknown as number,
    model: '',
    baseURL: '',
    token: '',
    dashboardUsername: 'admin',
    dashboardPassword: '',
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

const showModelConfig = computed(() => form.agentType === 'openclaw' || form.agentType === 'hermes-agent');

const generateDashboardPassword = () => {
    form.dashboardPassword = getRandomStr(8);
};

const ensureHermesDashboardAuth = () => {
    if (form.agentType !== 'hermes-agent') {
        return;
    }
    if (!form.dashboardUsername) {
        form.dashboardUsername = 'admin';
    }
    if (!form.dashboardPassword) {
        generateDashboardPassword();
    }
};

const setDefaultWebUIPort = () => {
    if (form.agentType === 'copaw') {
        form.webUIPort = 8088;
        return;
    }
    if (form.agentType === 'hermes-agent') {
        form.webUIPort = 9119;
        return;
    }
    form.webUIPort = 18789;
};

const rules = reactive({
    name: [Rules.appName],
    agentType: [Rules.requiredSelect],
    appVersion: [Rules.requiredSelect],
    webUIPort: [Rules.requiredInput],
    allowedOrigins: [
        Rules.requiredInput,
        {
            validator: (_rule: any, value: any, callback: (error?: Error) => void) => {
                const message = validateAllowedOriginsInput(String(value || ''));
                if (message) {
                    callback(new Error(message));
                    return;
                }
                callback();
            },
            trigger: 'blur',
        },
    ],
    provider: [Rules.requiredSelect],
    accountId: [Rules.requiredSelect],
    model: [Rules.requiredInput],
    dashboardUsername: [Rules.requiredInput],
    dashboardPassword: [Rules.requiredInput],
    containerName: [Rules.containerName],
    restartPolicy: [Rules.requiredSelect],
    cpuQuota: [checkNumberRange(0, 99999)],
    memoryLimit: [checkNumberRange(0, 9999999999)],
    specifyIP: [Rules.ipv4orV6],
});

const filteredModels = computed(() => {
    const selected = accountOptions.value.find((item) => item.id === form.accountId);
    return selected?.models || [];
});

const allowedOriginsPlaceholder = computed(() => buildDefaultAllowedOrigin('192.168.1.2', 18789, form.appVersion));

const syncAllowedOriginsWithDefault = (force = false) => {
    if (form.agentType !== 'openclaw') {
        return;
    }
    const defaultOrigin = buildDefaultAllowedOrigin(systemIP.value, form.webUIPort, form.appVersion);
    if (!force && !allowedOriginsAutoFilled.value && form.allowedOrigins !== lastAutoAllowedOrigins.value) {
        return;
    }
    form.allowedOrigins = defaultOrigin;
    lastAutoAllowedOrigins.value = defaultOrigin;
    allowedOriginsAutoFilled.value = true;
};

const loadSystemIP = async () => {
    try {
        const res = await getAgentSettingInfo();
        systemIP.value = String(res.data?.systemIP || '').trim();
    } catch (error) {
        systemIP.value = '';
    }
};

const loadVersions = async (appKey: AI.AgentType) => {
    const res = await getAppByKey(appKey);
    appInfo.value = res.data;
    versions.value = res.data.versions || [];
    if (!form.appVersion && versions.value.length > 0) {
        form.appVersion = versions.value[0];
    }
};

const getDefaultAgentName = (agentType: AI.AgentType) => {
    switch (agentType) {
        case 'copaw':
            return 'QwenPaw';
        case 'hermes-agent':
            return 'Hermes-Agent';
        default:
            return 'OpenClaw';
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
    if (!showModelConfig.value) {
        providerOptions.value = [];
        providerModels.value = {};
        return;
    }
    const res = await getAgentProviders();
    const data = res.data || [];
    const blockedProviders = new Set(['ark-coding-plan', 'bailian-coding-plan']);
    const filteredData = isIntl.value ? data.filter((item) => !blockedProviders.has(item.provider)) : data;
    providerOptions.value = filteredData.map((item) => ({
        value: item.provider,
        label: getAgentProviderDisplayName(item.provider, item.displayName),
    }));
    providerModels.value = filteredData.reduce(
        (acc, item) => {
            acc[item.provider] = item.models || [];
            return acc;
        },
        {} as Record<string, AI.ProviderModelInfo[]>,
    );
    await loadProviderAccountCounts(providerOptions.value.map((item) => item.value));
    providerOptions.value.sort((a, b) => {
        const aCount = providerAccountCount.value[a.value] || 0;
        const bCount = providerAccountCount.value[b.value] || 0;
        if (aCount > 0 && bCount === 0) {
            return -1;
        }
        if (aCount === 0 && bCount > 0) {
            return 1;
        }
        return 0;
    });
    if (!providerOptions.value.find((item) => item.value === form.provider) && providerOptions.value.length > 0) {
        form.provider = providerOptions.value[0].value;
    }
    setDefaultModel();
};

const loadProviderAccountCounts = async (providers: string[]) => {
    const providerList = Array.from(new Set(providers.filter(Boolean)));
    providerAccountCount.value = providerList.reduce(
        (acc, provider) => {
            acc[provider] = 0;
            return acc;
        },
        {} as Record<string, number>,
    );
    if (providerList.length === 0) {
        return;
    }
    const res = await countAgentAccountsByProvider({ providers: providerList });
    const counts = res.data || {};
    providerList.forEach((provider) => {
        providerAccountCount.value[provider] = counts[provider] || 0;
    });
};

const loadAccounts = async () => {
    if (!showModelConfig.value) {
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
        form.baseURL = '';
    }
};

const handleProviderChange = () => {
    if (!showModelConfig.value) {
        return;
    }
    form.model = '';
    form.baseURL = '';
    form.accountId = undefined as unknown as number;
    loadAccounts();
};

const handleAgentTypeChange = async () => {
    if (
        form.name === '' ||
        form.name === 'OpenClaw' ||
        form.name === 'CoPaw' ||
        form.name === 'QwenPaw' ||
        form.name === 'Hermes-Agent'
    ) {
        form.name = getDefaultAgentName(form.agentType);
    }
    setDefaultWebUIPort();
    form.appVersion = '';
    form.model = '';
    form.provider = '';
    form.accountId = undefined as unknown as number;
    form.baseURL = '';
    ensureHermesDashboardAuth();
    if (form.agentType === 'openclaw') {
        await loadSystemIP();
        allowedOriginsAutoFilled.value = true;
        syncAllowedOriginsWithDefault(true);
        await loadVersions('openclaw');
        await loadProviders();
        await loadAccounts();
        return;
    }
    form.allowedOrigins = '';
    lastAutoAllowedOrigins.value = '';
    allowedOriginsAutoFilled.value = true;
    await loadVersions(form.agentType);
    if (showModelConfig.value) {
        await loadProviders();
        await loadAccounts();
        return;
    }
    providerOptions.value = [];
    providerModels.value = {};
    accountOptions.value = [];
};

const handleAccountChange = () => {
    if (!showModelConfig.value) {
        return;
    }
    const selected = accountOptions.value.find((item) => item.id === form.accountId);
    if (selected) {
        form.baseURL = selected.baseUrl || '';
        if (!selected.models?.some((item) => item.id === form.model)) {
            form.model = selected.models?.[0]?.id || '';
        }
    }
    setDefaultModel();
};

const setDefaultModel = () => {
    if (!showModelConfig.value) {
        return;
    }
    const models = filteredModels.value;
    if (models.length > 0 && !form.model) {
        form.model = models[0].id;
        return;
    }
    if (models.length === 0) {
        form.model = '';
    }
};

const handleAllowedOriginsInput = () => {
    allowedOriginsAutoFilled.value = form.allowedOrigins === lastAutoAllowedOrigins.value;
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
            remark: form.remark,
            appVersion: form.appVersion,
            webUIPort: form.webUIPort,
            allowedOrigins: form.agentType === 'openclaw' ? parseAllowedOriginsInput(form.allowedOrigins) : undefined,
            agentType: form.agentType,
            model: showModelConfig.value ? form.model : undefined,
            accountId: showModelConfig.value ? form.accountId : undefined,
            token: form.agentType === 'openclaw' ? form.token : undefined,
            dashboardUsername: form.agentType === 'hermes-agent' ? form.dashboardUsername : undefined,
            dashboardPassword: form.agentType === 'hermes-agent' ? form.dashboardPassword : undefined,
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
    form.dashboardUsername = 'admin';
    form.dashboardPassword = '';
    form.remark = '';
    form.allowedOrigins = '';
    form.dockerCompose = '';
    lastAutoAllowedOrigins.value = '';
    allowedOriginsAutoFilled.value = true;
};

const openDrawer = async (agentType?: AI.AgentType) => {
    const targetType =
        agentType === 'copaw' || agentType === 'hermes-agent' || agentType === 'openclaw' ? agentType : 'openclaw';
    form.name = getDefaultAgentName(targetType);
    open.value = true;
    form.agentType = targetType;
    setDefaultWebUIPort();
    form.token = getRandomStr(32).toLowerCase();
    ensureHermesDashboardAuth();
    if (form.agentType === 'copaw') {
        form.allowedOrigins = '';
        lastAutoAllowedOrigins.value = '';
        allowedOriginsAutoFilled.value = true;
        await loadVersions('copaw');
        providerOptions.value = [];
        providerModels.value = {};
        accountOptions.value = [];
        return;
    }
    if (form.agentType === 'hermes-agent') {
        form.allowedOrigins = '';
        lastAutoAllowedOrigins.value = '';
        allowedOriginsAutoFilled.value = true;
        await loadVersions('hermes-agent');
        await loadProviders();
        await loadAccounts();
        return;
    }
    await loadSystemIP();
    allowedOriginsAutoFilled.value = true;
    syncAllowedOriginsWithDefault(true);
    await loadVersions('openclaw');
    await loadProviders();
    await loadAccounts();
};

const openAccountCreate = () => {
    if (!showModelConfig.value) {
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
        syncAllowedOriginsWithDefault();
        if (form.editCompose) {
            await loadCompose();
        }
    },
);

watch(
    () => form.webUIPort,
    () => {
        syncAllowedOriginsWithDefault();
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
