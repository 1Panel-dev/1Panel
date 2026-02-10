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
            </el-form-item>
            <el-form-item :label="$t('aiTools.agents.baseUrl')" prop="baseURL">
                <el-input v-model="form.baseURL" :disabled="form.provider !== 'ollama'" />
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
    apiKey: '',
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
                remark: form.remark,
                syncAgents: form.syncAgents,
            });
        } else {
            await createAgentAccount({
                provider: form.provider,
                name: form.name,
                baseURL: form.baseURL,
                apiKey: form.apiKey,
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
    form.syncAgents = false;
};

interface OpenParams {
    id?: number;
    provider?: string;
    name?: string;
    baseURL?: string;
    apiKey?: string;
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
        form.remark = params.remark || '';
        form.syncAgents = false;
        return;
    }
    form.id = 0;
    form.name = '';
    form.baseURL = '';
    form.apiKey = '';
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
        label: item.displayName || item.provider,
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
