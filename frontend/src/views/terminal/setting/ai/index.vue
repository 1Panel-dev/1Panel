<template>
    <el-form-item :label="$t('terminal.aiSettings')">
        <el-input :value="aiSummary" disabled>
            <template #append>
                <el-button @click="openDrawer" icon="Setting">
                    {{ $t('commons.button.set') }}
                </el-button>
            </template>
        </el-input>
    </el-form-item>

    <DrawerPro v-model="drawerVisible" :header="$t('terminal.aiSettings')" size="60%" @close="handleClose">
        <el-form ref="formRef" :model="formModel" label-position="top">
            <el-form-item :label="$t('terminal.aiStatus')">
                <el-switch v-model="formModel.status" active-value="Enable" inactive-value="Disable" />
            </el-form-item>
            <el-form-item
                :label="$t('aiTools.agents.account')"
                prop="accountId"
                :rules="accountRules"
                v-if="formModel.status === 'Enable'"
            >
                <el-select class="formInput" v-model="formModel.accountId" clearable filterable>
                    <el-option
                        v-for="item in agentAccountOptions"
                        :key="item.id"
                        :label="item.name"
                        :value="String(item.id)"
                    >
                        <div class="account-option">
                            <span class="account-option__name">{{ item.name }}</span>
                            <div class="account-option__tags">
                                <el-tag size="small" effect="plain">
                                    {{ item.providerName || item.provider }}
                                </el-tag>
                                <el-tag size="small" effect="plain" :type="verificationTagType(item)">
                                    {{ verificationLabel(item) }}
                                </el-tag>
                            </div>
                        </div>
                    </el-option>
                </el-select>
                <span class="input-help">{{ $t('terminal.aiAccountHelper') }}</span>
            </el-form-item>
            <el-form-item :label="$t('terminal.aiPrefix')" prop="prefix" :rules="Rules.requiredSelect">
                <el-select class="formInput" v-model="formModel.prefix">
                    <el-option v-for="item in prefixOptions" :key="item" :label="item" :value="item" />
                </el-select>
                <span class="input-help">{{ $t('terminal.aiPrefixHelper') }}</span>
            </el-form-item>
            <el-form-item :label="$t('terminal.aiRiskCommands')" prop="riskCommands" :rules="riskCommandRules">
                <div class="risk-command-list">
                    <div class="risk-command-item" v-for="(command, index) in formModel.riskCommands" :key="index">
                        <el-input :model-value="command" @update:model-value="updateRiskCommand(index, $event)" />
                        <el-button link type="danger" @click="removeRiskCommand(index)">
                            {{ $t('terminal.aiRemoveRiskCommand') }}
                        </el-button>
                    </div>
                    <div class="risk-command-actions">
                        <el-button plain @click="addRiskCommand">{{ $t('terminal.aiAddRiskCommand') }}</el-button>
                    </div>
                </div>
                <span class="input-help">{{ $t('terminal.aiRiskCommandsHelper') }}</span>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="drawerVisible = false" :disabled="saving">{{ $t('commons.button.cancel') }}</el-button>
            <el-button plain @click="resetRiskCommands" :disabled="saving">
                {{ $t('commons.button.setDefault') }}
            </el-button>
            <el-button type="primary" @click="handleConfirm" :loading="saving">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { computed, nextTick, reactive, ref, watch } from 'vue';
import type { ElForm } from 'element-plus';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { pageAgentAccounts } from '@/api/modules/ai';
import { updateAgentTerminalAIInfo } from '@/api/modules/setting';
import { MsgSuccess } from '@/utils/message';
import { AI_PREFIX_OPTIONS, DEFAULT_AI_PREFIX, normalizeRiskCommands } from '@/views/terminal/setting/ai/helper';

interface AgentAccountOption {
    id: number | string;
    name: string;
    provider?: string;
    providerName?: string;
    verified?: boolean;
}

const props = defineProps<{
    status: string;
    accountId: string;
    prefix: string;
    riskCommands: string[];
    defaultRiskCommands: string[];
}>();

const emit = defineEmits<{
    (e: 'refresh'): void;
}>();

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();
const drawerVisible = ref(false);
const saving = ref(false);
const agentAccountOptions = ref<AgentAccountOption[]>([]);
const prefixOptions = AI_PREFIX_OPTIONS;
const formModel = reactive({
    status: props.status,
    accountId: props.accountId,
    prefix: props.prefix,
    riskCommands: [...props.riskCommands],
});

const syncFormFromProps = () => {
    formModel.status = props.status;
    formModel.accountId = props.accountId;
    formModel.prefix = props.prefix;
    formModel.riskCommands = [...props.riskCommands];
};

const openDrawer = () => {
    syncFormFromProps();
    loadAgentAccounts();
    drawerVisible.value = true;
};

watch(
    () => [props.status, props.accountId, props.prefix, props.riskCommands],
    () => {
        if (drawerVisible.value) {
            return;
        }
        syncFormFromProps();
    },
    { deep: true },
);

const aiSummary = computed(() => {
    if (props.status !== 'Enable') {
        return i18n.global.t('setting.unSetting');
    }
    const prefix = String(props.prefix || '').trim();
    return i18n.global.t('terminal.aiSummary', [prefix]);
});

const isVerificationSkipped = (provider?: string) => {
    const key = (provider || '').toLowerCase();
    return key === 'custom' || key === 'vllm' || key === 'ollama' || key === 'kimi-coding';
};

const verificationLabel = (item: AgentAccountOption) => {
    if (isVerificationSkipped(item.provider)) {
        return i18n.global.t('aiTools.agents.verifySkipped');
    }
    return item.verified ? 'OK' : 'N/A';
};

const verificationTagType = (item: AgentAccountOption) => {
    if (isVerificationSkipped(item.provider)) {
        return 'info';
    }
    return item.verified ? 'success' : 'warning';
};

const loadAgentAccounts = async () => {
    await pageAgentAccounts({
        page: 1,
        pageSize: 1000,
        provider: '',
        name: '',
    }).then((res) => {
        agentAccountOptions.value = res.data?.items || [];
    });
};

const accountRules = [
    {
        ...Rules.requiredSelect,
        validator: (_rule, value, callback) => {
            if (formModel.status !== 'Enable') {
                callback();
                return;
            }
            if (!value) {
                callback(new Error(i18n.global.t('commons.rule.requiredSelect')));
                return;
            }
            callback();
        },
    },
];

const riskCommandRules = [
    {
        validator: (_rule, value, callback) => {
            const commands = Array.isArray(value) ? value : [];
            if (commands.some((item) => String(item ?? '').trim().length === 0)) {
                callback(new Error(i18n.global.t('commons.rule.requiredInput')));
                return;
            }
            callback();
        },
        trigger: 'blur',
    },
];

const addRiskCommand = () => {
    formModel.riskCommands = [...formModel.riskCommands, ''];
};

const updateRiskCommand = (index: number, value: string) => {
    formModel.riskCommands = formModel.riskCommands.map((item, currentIndex) =>
        currentIndex === index ? value : item,
    );
};

const removeRiskCommand = (index: number) => {
    formModel.riskCommands = formModel.riskCommands.filter((_, currentIndex) => currentIndex !== index);
};

const resetRiskCommands = () => {
    formModel.prefix = DEFAULT_AI_PREFIX;
    formModel.riskCommands = [...props.defaultRiskCommands];
};

const handleConfirm = async () => {
    await nextTick();
    if (!formRef.value) {
        return;
    }
    try {
        await formRef.value.validate();
    } catch {
        return;
    }
    saving.value = true;
    try {
        await updateAgentTerminalAIInfo({
            aiStatus: formModel.status,
            aiAccountId: formModel.status === 'Enable' ? formModel.accountId : '',
            aiPrefix: formModel.prefix.trim() || DEFAULT_AI_PREFIX,
            aiRiskCommands: JSON.stringify(normalizeRiskCommands(formModel.riskCommands)),
        });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        drawerVisible.value = false;
    } finally {
        saving.value = false;
    }
};

const handleClose = () => {
    syncFormFromProps();
    emit('refresh');
};
</script>

<style lang="css" scoped>
.formInput {
    width: 100%;
}

.risk-command-list {
    width: 100%;
    display: flex;
    flex-direction: column;
    gap: 8px;
}

.risk-command-actions {
    display: flex;
    gap: 8px;
}

.risk-command-item {
    display: flex;
    gap: 8px;
    align-items: center;
}

.account-option {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    width: 100%;
}

.account-option__name {
    min-width: 0;
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.account-option__tags {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    flex-shrink: 0;
}
</style>
