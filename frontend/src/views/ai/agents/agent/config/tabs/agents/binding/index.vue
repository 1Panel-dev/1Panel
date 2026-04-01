<template>
    <DialogPro v-model="open" :title="$t('commons.button.bind')" size="large" @close="handleClose">
        <div v-loading="loading" class="binding-dialog">
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
                <el-table-column :label="$t('aiTools.agents.accountIdOptional')" min-width="220">
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
                <el-table-column :label="$t('commons.table.operate')" width="100" align="right">
                    <template #default="{ $index }">
                        <el-button link type="danger" @click="removeBinding($index)">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </template>
                </el-table-column>
            </el-table>
            <div v-else class="binding-dialog__empty">
                <el-empty :description="$t('commons.msg.noneData')" :image-size="60" />
            </div>
            <div class="binding-dialog__actions">
                <el-button type="primary" link :disabled="loading || submitting" @click="addBinding">
                    {{ $t('commons.button.add') }}
                </el-button>
            </div>
        </div>
        <template #footer>
            <el-button
                type="primary"
                :loading="submitting"
                :disabled="loading || !form.bindings.length"
                @click="submitBind"
            >
                {{ $t('commons.button.bind') }}
            </el-button>
            <el-button :disabled="loading || submitting" @click="handleClose">
                {{ $t('commons.button.cancel') }}
            </el-button>
        </template>
    </DialogPro>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { bindAgentRole, getAgentRoleChannels } from '@/api/modules/ai';
import { AI } from '@/api/interface/ai';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';

interface SelectOption {
    label: string;
    value: string;
    accountIds: string[];
}

interface DialogParams {
    agentId: number;
    role: AI.AgentConfiguredAgentItem;
}

const emit = defineEmits(['success']);

const open = ref(false);
const loading = ref(false);
const submitting = ref(false);
const agentId = ref(0);
const roleId = ref('');
const channelOptions = ref<SelectOption[]>([]);

const form = ref<{ bindings: AI.AgentRoleBinding[] }>({
    bindings: [],
});

const resetForm = () => {
    form.value = {
        bindings: [],
    };
};

const addBinding = () => {
    form.value.bindings.push({
        channel: '',
        accountId: '',
    });
};

const removeBinding = (index: number) => {
    form.value.bindings.splice(index, 1);
};

const handleBindingChannelChange = (index: number) => {
    const binding = form.value.bindings[index];
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
    if (option.accountIds.length > 0) {
        return false;
    }
    return form.value.bindings.some((item, bindingIndex) => bindingIndex !== index && item.channel === option.value);
};

const loadBindOptions = async () => {
    const res = await getAgentRoleChannels({ agentId: agentId.value });
    channelOptions.value = (res.data || [])
        .filter((item) => !item.bound)
        .map((item) => ({
            label: item.name,
            value: item.name,
            accountIds: item.accountIds || [],
        }));
};

const refreshData = async () => {
    await loadBindOptions();
};

const acceptParams = async (params: DialogParams) => {
    agentId.value = params.agentId;
    roleId.value = params.role.id;
    resetForm();
    loading.value = true;
    try {
        await refreshData();
        addBinding();
    } finally {
        loading.value = false;
    }
    open.value = true;
};

const submitBind = async () => {
    const bindings = form.value.bindings.filter((item) => item.channel);
    if (!bindings.length) {
        MsgError(i18n.global.t('commons.msg.selectOne', [i18n.global.t('aiTools.agents.channelsTab')]));
        return;
    }
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
    submitting.value = true;
    loading.value = true;
    try {
        for (const item of bindings) {
            await bindAgentRole({
                agentId: agentId.value,
                id: roleId.value,
                channel: item.channel,
                accountId: item.accountId,
            });
        }
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        resetForm();
        emit('success');
        handleClose();
    } finally {
        submitting.value = false;
        loading.value = false;
    }
};

const handleClose = () => {
    open.value = false;
    channelOptions.value = [];
    resetForm();
};

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.binding-dialog {
    display: flex;
    flex-direction: column;
    gap: 20px;
}

.bindings-table {
    width: 100%;
}

.binding-dialog__actions {
    display: flex;
    justify-content: flex-start;
}

.bindings-table :deep(.el-select) {
    width: 100%;
}

.binding-dialog__empty {
    padding: 12px 0;
}
</style>
