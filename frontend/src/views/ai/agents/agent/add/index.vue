<template>
    <DrawerPro v-model="open" :header="$t('aiTools.agents.createAgent')" size="large" @close="handleClose">
        <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
            <el-card class="form-card">
                <el-form-item :label="$t('commons.table.name')" prop="name">
                    <el-input v-model="form.name" />
                </el-form-item>
                <el-form-item :label="$t('aiTools.agents.appVersion')" prop="appVersion">
                    <el-select v-model="form.appVersion" filterable>
                        <el-option v-for="item in versions" :key="item" :label="item" :value="item" />
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('aiTools.agents.webuiPort')" prop="webUIPort">
                    <el-input-number v-model="form.webUIPort" :min="1" :max="65535" />
                </el-form-item>
                <el-form-item :label="$t('aiTools.agents.bridgePort')" prop="bridgePort">
                    <el-input-number v-model="form.bridgePort" :min="1" :max="65535" />
                </el-form-item>
            </el-card>
            <el-card class="form-card">
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
                <el-form-item>
                    <el-checkbox v-model="manualModel">{{ $t('aiTools.agents.manualModel') }}</el-checkbox>
                </el-form-item>
                <el-form-item :label="$t('aiTools.model.model')" prop="model">
                    <el-input v-if="manualModel" v-model="form.model" />
                    <el-select v-else v-model="form.model" filterable @change="handleModelChange">
                        <el-option v-for="item in filteredModels" :key="item.id" :label="item.name" :value="item.id" />
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('aiTools.agents.account')" prop="accountId">
                    <el-select v-model="form.accountId" @change="handleAccountChange">
                        <el-option v-for="item in accountOptions" :key="item.id" :label="item.name" :value="item.id" />
                    </el-select>
                    <span class="input-help">
                        {{ $t('aiTools.agents.noAccountHint') }}
                        <el-button type="primary" link class="inline-link" @click="openAccountCreate">
                            {{ $t('aiTools.agents.createModelAccount') }}
                        </el-button>
                    </span>
                </el-form-item>
                <el-form-item :label="$t('aiTools.agents.apiKey')" v-if="form.accountId" prop="apiKey">
                    <el-input v-model="form.apiKey" type="password" show-password readonly />
                </el-form-item>
                <el-form-item :label="$t('aiTools.agents.baseUrl')" v-if="form.accountId" prop="baseURL">
                    <el-input v-model="form.baseURL" readonly />
                </el-form-item>
                <el-form-item :label="$t('aiTools.agents.token')">
                    <el-input v-model="form.token" readonly>
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
    appVersion: '',
    webUIPort: 18789,
    bridgePort: 18790,
    provider: 'deepseek',
    accountId: undefined as unknown as number,
    model: '',
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

const rules = reactive({
    name: [Rules.requiredInput],
    appVersion: [Rules.requiredSelect],
    webUIPort: [Rules.requiredInput],
    bridgePort: [Rules.requiredInput],
    provider: [Rules.requiredSelect],
    accountId: [Rules.requiredSelect],
    model: [Rules.requiredSelect],
    containerName: [Rules.containerName],
    restartPolicy: [Rules.requiredSelect],
    cpuQuota: [checkNumberRange(0, 99999)],
    memoryLimit: [checkNumberRange(0, 9999999999)],
    specifyIP: [Rules.ipv4orV6],
});

const filteredModels = computed(() => providerModels.value[form.provider] || []);

const loadVersions = async () => {
    const res = await getAppByKey('openclaw');
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
    const res = await getAgentProviders();
    const data = res.data || [];
    providerOptions.value = data.map((item) => ({
        value: item.provider,
        label: item.displayName || item.provider,
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
    form.model = '';
    form.apiKey = '';
    form.baseURL = '';
    form.accountId = undefined as unknown as number;
    loadAccounts();
    setDefaultModel();
};

const handleModelChange = () => {
    if (manualModel.value) {
        return;
    }
};

const handleAccountChange = () => {
    const selected = accountOptions.value.find((item) => item.id === form.accountId);
    if (selected) {
        form.baseURL = selected.baseUrl || '';
        form.apiKey = selected.apiKey || '';
    }
    setDefaultModel();
};

const setDefaultModel = () => {
    if (manualModel.value) {
        return;
    }
    const models = filteredModels.value;
    if (models.length > 0 && !form.model) {
        form.model = models[0].id;
    }
};

const submit = async () => {
    if (!formRef.value) {
        return;
    }
    await formRef.value.validate();
    const taskID = newUUID();
    if (!form.token) {
        form.token = getRandomStr(32).toLowerCase();
    }
    const res = await createAgent({
        name: form.name,
        appVersion: form.appVersion,
        webUIPort: form.webUIPort,
        bridgePort: form.bridgePort,
        provider: form.provider,
        model: form.model,
        accountId: form.accountId,
        apiKey: form.apiKey,
        baseURL: form.baseURL,
        token: form.token,
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
};

const handleClose = () => {
    formRef.value?.resetFields();
    form.token = '';
    form.dockerCompose = '';
};

const openDrawer = async () => {
    form.name = 'openclaw';
    open.value = true;
    manualModel.value = false;
    form.token = getRandomStr(32).toLowerCase();
    await loadVersions();
    await loadProviders();
    await loadAccounts();
};

const openAccountCreate = () => {
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
