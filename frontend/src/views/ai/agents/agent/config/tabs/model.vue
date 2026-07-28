<template>
    <el-form ref="formRef" :model="form" :rules="rules" label-position="top" v-loading="loading">
        <el-form-item :label="t('aiTools.agents.account')" prop="accountId">
            <el-select v-model="form.accountId" @change="handleAccountChange">
                <el-option v-for="item in accountOptions" :key="item.id" :label="item.name" :value="item.id" />
            </el-select>
        </el-form-item>
        <el-form-item :label="t('aiTools.agents.primaryModel')" prop="model">
            <el-select v-model="form.model" filterable @change="handleModelChange">
                <el-option v-for="item in modelOptions" :key="item.id" :label="item.name" :value="item.id" />
            </el-select>
            <span class="input-help">{{ t('aiTools.agents.accountModelsHelper') }}</span>
        </el-form-item>
        <el-form-item v-if="showFallbacks" :label="t('aiTools.agents.fallbackModels')">
            <div class="fallback-editor">
                <div class="fallback-toolbar">
                    <el-select v-model="fallbackCandidate" filterable clearable class="fallback-select">
                        <el-option
                            v-for="item in fallbackCandidateOptions"
                            :key="item.id"
                            :label="item.name"
                            :value="item.id"
                        />
                    </el-select>
                    <el-button v-permission type="primary" plain :disabled="!fallbackCandidate" @click="addFallback">
                        {{ t('aiTools.agents.addFallbackModel') }}
                    </el-button>
                </div>
                <el-table :data="fallbackRows" size="small">
                    <el-table-column :label="t('aiTools.model.model')" min-width="220">
                        <template #default="{ row }">
                            {{ row.name }}
                        </template>
                    </el-table-column>
                    <el-table-column :label="t('commons.table.operate')" width="180" fixed="right">
                        <template #default="{ $index }">
                            <el-button
                                link
                                :icon="ArrowUp"
                                :disabled="$index === 0"
                                @click="moveFallback($index, -1)"
                            />
                            <el-button
                                link
                                :icon="ArrowDown"
                                :disabled="$index === fallbackRows.length - 1"
                                @click="moveFallback($index, 1)"
                            />
                            <el-button v-permission link :icon="Delete" @click="removeFallback($index)" />
                        </template>
                    </el-table-column>
                    <template #empty>
                        <div class="fallback-empty">{{ t('aiTools.agents.fallbackModelsEmpty') }}</div>
                    </template>
                </el-table>
            </div>
        </el-form-item>
        <el-form-item>
            <el-button v-permission type="primary" :loading="saving" @click="saveModel">
                {{ t('commons.button.save') }}
            </el-button>
        </el-form-item>
    </el-form>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue';
import type { FormInstance } from 'element-plus';
import { ArrowDown, ArrowUp, Delete } from '@element-plus/icons-vue';
import { useI18n } from 'vue-i18n';
import { AI } from '@/api/interface/ai';
import { getAgentModelConfig, pageAgentAccounts, updateAgentModelConfig } from '@/api/modules/ai';
import { Rules } from '@/global/form-rules';
import { MsgSuccess } from '@/utils/message';
import { useGlobalStore } from '@/composables/useGlobalStore';

const emit = defineEmits(['updated']);
const { t } = useI18n();

const loading = ref(false);
const saving = ref(false);
const formRef = ref<FormInstance>();
const { isIntl } = useGlobalStore();
const blockedProviders = new Set(['ark-coding-plan', 'bailian-coding-plan']);

const agentId = ref(0);
const agentType = ref<AI.AgentType>('openclaw');
const accountOptions = ref<AI.AgentAccountItem[]>([]);
const modelOptions = ref<AI.AgentAccountModel[]>([]);
const showFallbacks = computed(() => agentType.value === 'openclaw');

interface AgentModelLoadParams {
    agentId: number;
    agentType: AI.AgentType;
}

const form = reactive({
    accountId: undefined as unknown as number,
    model: '',
    fallbacks: [] as string[],
});

const rules = reactive({
    accountId: [Rules.requiredSelect],
    model: [Rules.requiredSelect],
});

const fallbackCandidate = ref('');

const fallbackRows = computed(() =>
    form.fallbacks
        .map((id) => modelOptions.value.find((item) => item.id === id))
        .filter((item): item is AI.AgentAccountModel => item !== undefined),
);

const fallbackCandidateOptions = computed(() =>
    modelOptions.value.filter((item) => item.id !== form.model && !form.fallbacks.includes(item.id)),
);

const setModelOptionsByAccount = (accountId: number) => {
    const selected = accountOptions.value.find((item) => item.id === accountId);
    modelOptions.value = selected?.models || [];
};

const ensureSelectedModel = (preferred?: string) => {
    if (preferred && modelOptions.value.some((item) => item.id === preferred)) {
        form.model = preferred;
        return;
    }
    if (!modelOptions.value.some((item) => item.id === form.model)) {
        form.model = modelOptions.value[0]?.id || '';
    }
};

const syncFallbacks = (fallbacks?: string[]) => {
    const current = fallbacks || form.fallbacks;
    form.fallbacks = current.filter(
        (id, index) =>
            id !== form.model &&
            modelOptions.value.some((item) => item.id === id) &&
            current.findIndex((item) => item === id) === index,
    );
    if (!fallbackCandidateOptions.value.some((item) => item.id === fallbackCandidate.value)) {
        fallbackCandidate.value = '';
    }
};

const handleAccountChange = () => {
    setModelOptionsByAccount(form.accountId);
    ensureSelectedModel();
    form.fallbacks = [];
    fallbackCandidate.value = '';
};

const handleModelChange = () => {
    syncFallbacks();
};

const addFallback = () => {
    if (!fallbackCandidate.value) {
        return;
    }
    syncFallbacks([...form.fallbacks, fallbackCandidate.value]);
};

const moveFallback = (index: number, offset: number) => {
    const target = index + offset;
    if (target < 0 || target >= form.fallbacks.length) {
        return;
    }
    const next = [...form.fallbacks];
    [next[index], next[target]] = [next[target], next[index]];
    form.fallbacks = next;
};

const removeFallback = (index: number) => {
    form.fallbacks = form.fallbacks.filter((_, current) => current !== index);
};

const load = async (params: AgentModelLoadParams) => {
    loading.value = true;
    try {
        agentId.value = params.agentId;
        agentType.value = params.agentType;
        const [accountsRes, configRes] = await Promise.all([
            pageAgentAccounts({
                page: 1,
                pageSize: 200,
                provider: '',
                textOnly: true,
                name: '',
            }),
            getAgentModelConfig({ agentId: params.agentId }),
        ]);
        const items = accountsRes.data.items || [];
        accountOptions.value = isIntl.value ? items.filter((item) => !blockedProviders.has(item.provider)) : items;
        if (accountOptions.value.length === 0) {
            form.accountId = undefined as unknown as number;
            form.model = '';
            form.fallbacks = [];
            modelOptions.value = [];
            return;
        }
        const config = configRes.data;
        const currentAccount =
            accountOptions.value.find((item) => item.id === config.accountId) || accountOptions.value[0];
        form.accountId = currentAccount.id;
        setModelOptionsByAccount(currentAccount.id);
        ensureSelectedModel(config.model);
        syncFallbacks(showFallbacks.value ? config.fallbacks || [] : []);
    } finally {
        loading.value = false;
    }
};

const saveModel = async () => {
    if (!agentId.value || !formRef.value) {
        return;
    }
    await formRef.value.validate();
    saving.value = true;
    try {
        await updateAgentModelConfig({
            agentId: agentId.value,
            accountId: form.accountId,
            model: form.model,
            fallbacks: showFallbacks.value ? form.fallbacks : [],
        });
        MsgSuccess(t('aiTools.agents.switchModelSuccess'));
        emit('updated');
    } finally {
        saving.value = false;
    }
};

defineExpose({
    load,
});
</script>

<style scoped lang="scss">
.fallback-editor {
    width: 100%;
}

.fallback-toolbar {
    display: flex;
    gap: 12px;
    margin-bottom: 12px;
}

.fallback-select {
    width: 360px;
}

.fallback-empty {
    padding: 12px 0;
    color: var(--el-text-color-secondary);
}
</style>
