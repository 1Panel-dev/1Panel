<template>
    <DialogPro v-model="open" :title="`${$t('commons.button.add')}`" size="large" @close="handleClose">
        <div v-loading="loading" class="create-role-dialog">
            <el-form ref="formRef" :model="form" :rules="rules" label-position="top">
                <div class="create-role-section">
                    <el-form-item :label="$t('commons.table.name')" prop="name">
                        <el-input v-model.trim="form.name" />
                    </el-form-item>
                    <el-form-item :label="$t('aiTools.model.model')">
                        <el-select v-model="form.model" clearable filterable class="w-full">
                            <el-option
                                v-for="item in modelOptions"
                                :key="item.value"
                                :label="item.label"
                                :value="item.value"
                            />
                        </el-select>
                    </el-form-item>
                    <div class="bindings-divider"></div>
                    <el-table v-if="form.bindings.length" :data="form.bindings" class="bindings-table" size="small">
                        <el-table-column :label="$t('aiTools.agents.channelsTab')" min-width="180">
                            <template #default="{ row, $index }">
                                <el-select
                                    v-model="row.channel"
                                    clearable
                                    filterable
                                    class="w-full"
                                    @change="handleBindingChannelChange($index)"
                                >
                                    <el-option
                                        v-for="item in channelOptions"
                                        :key="item.value"
                                        :label="item.label"
                                        :value="item.value"
                                        :disabled="isChannelDisabled(item, $index)"
                                    />
                                </el-select>
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('aiTools.agents.accountIdOptional')" min-width="180">
                            <template #default="{ row }">
                                <el-select v-model="row.accountId" clearable filterable class="w-full">
                                    <el-option
                                        v-for="item in getAccountIdOptions(row.channel)"
                                        :key="item"
                                        :label="item"
                                        :value="item"
                                    />
                                </el-select>
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('commons.table.operate')" width="90" align="center">
                            <template #default="{ $index }">
                                <el-button link @click="removeBinding($index)">
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                            </template>
                        </el-table-column>
                    </el-table>
                    <el-empty v-else :description="$t('commons.msg.noneData')" :image-size="60" />
                    <div class="bindings-footer">
                        <el-button type="primary" link @click="addBinding">
                            {{ $t('commons.button.add') }}
                        </el-button>
                    </div>
                </div>
            </el-form>
        </div>
        <template #footer>
            <el-button :disabled="loading" @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" :loading="loading" @click="submit">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DialogPro>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue';
import type { FormInstance } from 'element-plus';
import { createAgentRole, getAgentRoleChannels, pageAgentAccounts } from '@/api/modules/ai';
import { AI } from '@/api/interface/ai';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { useGlobalStore } from '@/composables/useGlobalStore';

interface SelectOption {
    label: string;
    value: string;
    bound?: boolean;
    accountIds?: string[];
}

interface DialogParams {
    agentId: number;
    accountId: number;
    model: string;
}

const emit = defineEmits(['success']);
const { isIntl } = useGlobalStore();
const blockedProviders = new Set(['ark-coding-plan', 'bailian-coding-plan']);

const open = ref(false);
const loading = ref(false);
const formRef = ref<FormInstance>();
const agentId = ref(0);
const accountId = ref(0);
const modelOptions = ref<SelectOption[]>([]);
const channelOptions = ref<SelectOption[]>([]);

const form = reactive({
    name: '',
    bindings: [] as AI.AgentRoleBinding[],
    model: '',
});

const rules = reactive({
    name: [Rules.simpleName],
});

const resetForm = (currentModel?: string) => {
    form.name = '';
    form.bindings = [];
    form.model = currentModel || '';
};

const addBinding = () => {
    form.bindings.push({
        channel: '',
        accountId: '',
    });
};

const removeBinding = (index: number) => {
    form.bindings.splice(index, 1);
};

const handleBindingChannelChange = (index: number) => {
    const binding = form.bindings[index];
    if (!binding) {
        return;
    }
    binding.accountId = '';
    const options = getAccountIdOptions(binding.channel);
    if (options.length === 1) {
        binding.accountId = options[0];
    }
};

const getAccountIdOptions = (channel: string) => {
    return channelOptions.value.find((item) => item.value === channel)?.accountIds || [];
};

const isChannelDisabled = (option: SelectOption, index: number) => {
    if (option.bound) {
        return true;
    }
    if ((option.accountIds || []).length > 0) {
        return false;
    }
    return form.bindings.some((item, bindingIndex) => bindingIndex !== index && item.channel === option.value);
};

const loadAccounts = async () => {
    const res = await pageAgentAccounts({
        page: 1,
        pageSize: 200,
        provider: '',
        name: '',
    });
    const items = res.data.items || [];
    const accountOptions = isIntl.value ? items.filter((item) => !blockedProviders.has(item.provider)) : items;
    const selected = accountOptions.find((item) => item.id === accountId.value);
    modelOptions.value = (selected?.models || []).map((item) => ({
        label: item.name,
        value: item.id,
    }));
};

const loadChannels = async () => {
    const res = await getAgentRoleChannels({ agentId: agentId.value });
    channelOptions.value = (res.data || []).map((item) => ({
        label: item.name,
        value: item.name,
        bound: item.bound,
        accountIds: item.accountIds || [],
    }));
};

const acceptParams = async (params: DialogParams) => {
    agentId.value = params.agentId;
    accountId.value = params.accountId;
    resetForm(params.model);
    loading.value = true;
    try {
        await Promise.all([loadAccounts(), loadChannels()]);
        addBinding();
    } finally {
        loading.value = false;
    }
    open.value = true;
};

const submit = async () => {
    if (!formRef.value || !agentId.value) {
        return;
    }
    await formRef.value.validate();
    const bindings = form.bindings.filter((item) => item.channel);
    const hasDuplicate = bindings.some((item, index) =>
        bindings.some(
            (current, currentIndex) =>
                currentIndex !== index && current.channel === item.channel && current.accountId === item.accountId,
        ),
    );
    if (hasDuplicate) {
        MsgError(i18n.global.t('aiTools.agents.duplicateBinding'));
        return;
    }
    loading.value = true;
    try {
        await createAgentRole({
            agentId: agentId.value,
            name: form.name.trim(),
            model: form.model.trim(),
            bindings: bindings.map((item) => ({
                channel: item.channel,
                accountId: item.accountId,
            })),
        } as AI.AgentRoleCreateReq);
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        emit('success');
        handleClose();
    } finally {
        loading.value = false;
    }
};

const handleClose = () => {
    open.value = false;
};

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.create-role-dialog {
    padding-top: 4px;
}

.create-role-section {
    padding: 14px 16px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 12px;
    background: var(--el-fill-color-blank);
}

.bindings-table {
    width: 100%;
}

.bindings-table :deep(.el-select) {
    width: 100%;
}

.bindings-divider {
    margin: 4px 0 12px;
    border-top: 1px dashed var(--el-border-color);
}

.bindings-footer {
    display: flex;
    justify-content: flex-start;
    margin-top: 8px;
}

.w-full {
    width: 100%;
}
</style>
