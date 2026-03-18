<template>
    <el-form-item :label="$t('terminal.aiSettings')">
        <el-input :value="aiSummary" disabled>
            <template #append>
                <el-button @click="drawerVisible = true" icon="Setting">
                    {{ $t('commons.button.set') }}
                </el-button>
            </template>
        </el-input>
    </el-form-item>

    <DrawerPro v-model="drawerVisible" :header="$t('terminal.aiSettings')" size="60%" @close="handleClose">
        <el-form ref="formRef" :model="formModel" label-position="top">
            <el-form-item :label="$t('terminal.aiStatus')">
                <el-switch v-model="statusModel" active-value="Enable" inactive-value="Disable" />
            </el-form-item>
            <el-form-item
                :label="$t('aiTools.agents.account')"
                prop="accountId"
                :rules="accountRules"
                v-if="statusModel === 'Enable'"
            >
                <el-select class="formInput" v-model="accountIdModel" clearable filterable>
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
            <el-form-item :label="$t('terminal.aiPrefix')" prop="prefix" :rules="prefixRules">
                <el-input class="formInput" v-model="prefixModel" />
                <span class="input-help">{{ $t('terminal.aiPrefixHelper') }}</span>
            </el-form-item>
            <el-form-item :label="$t('terminal.aiRiskCommands')" prop="riskCommands" :rules="riskCommandRules">
                <div class="risk-command-list">
                    <div class="risk-command-item" v-for="(command, index) in riskCommandsModel" :key="index">
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
            <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button plain @click="resetRiskCommands">{{ $t('commons.button.setDefault') }}</el-button>
            <el-button type="primary" @click="handleConfirm">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { computed, nextTick, ref } from 'vue';
import type { ElForm } from 'element-plus';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { DEFAULT_AI_PREFIX, DEFAULT_AI_RISK_COMMANDS } from '@/views/terminal/setting/ai/helper';

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
    agentAccountOptions: AgentAccountOption[];
    submitHandler?: () => Promise<boolean>;
}>();

const emit = defineEmits<{
    (e: 'update:status', value: string): void;
    (e: 'update:accountId', value: string): void;
    (e: 'update:prefix', value: string): void;
    (e: 'update:riskCommands', value: string[]): void;
    (e: 'refresh'): void;
}>();

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();
const drawerVisible = ref(false);

const formModel = computed(() => ({
    status: props.status,
    accountId: props.accountId,
    prefix: props.prefix,
    riskCommands: props.riskCommands,
}));

const statusModel = computed({
    get: () => props.status,
    set: (value: string) => emit('update:status', value),
});

const accountIdModel = computed({
    get: () => props.accountId,
    set: (value: string) => emit('update:accountId', value),
});

const prefixModel = computed({
    get: () => props.prefix,
    set: (value: string) => emit('update:prefix', value),
});

const riskCommandsModel = computed({
    get: () => props.riskCommands,
    set: (value: string[]) => emit('update:riskCommands', value),
});

const aiSummary = computed(() => {
    if (props.status !== 'Enable') {
        return i18n.global.t('setting.unSetting');
    }
    const account = props.agentAccountOptions.find((item) => String(item.id) === props.accountId);
    if (!account) {
        return i18n.global.t('terminal.aiAccountHelper');
    }
    const prefix = String(props.prefix || '').trim();
    const accountType = account.providerName || account.provider || '-';
    return i18n.global.t('terminal.aiSummary', [prefix, account.name, accountType]);
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

const accountRules = [
    {
        ...Rules.requiredSelect,
        validator: (_rule, value, callback) => {
            if (props.status !== 'Enable') {
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

const prefixRules = [
    Rules.requiredInput,
    {
        validator: (_rule, value, callback) => {
            const normalized = String(value ?? '').trim();
            if (!normalized) {
                callback(new Error(i18n.global.t('commons.rule.requiredInput')));
                return;
            }
            if (!/^[!-~]+$/.test(normalized)) {
                callback(new Error(i18n.global.t('terminal.aiPrefixAsciiVisible')));
                return;
            }
            callback();
        },
        trigger: 'blur',
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
    emit('update:riskCommands', [...props.riskCommands, '']);
};

const updateRiskCommand = (index: number, value: string) => {
    emit(
        'update:riskCommands',
        props.riskCommands.map((item, currentIndex) => (currentIndex === index ? value : item)),
    );
};

const removeRiskCommand = (index: number) => {
    emit(
        'update:riskCommands',
        props.riskCommands.filter((_, currentIndex) => currentIndex !== index),
    );
};

const resetRiskCommands = () => {
    emit('update:prefix', DEFAULT_AI_PREFIX);
    emit('update:riskCommands', [...DEFAULT_AI_RISK_COMMANDS]);
};

const validate = async () => {
    drawerVisible.value = true;
    await nextTick();
    if (!formRef.value) {
        return false;
    }
    try {
        await formRef.value.validate();
        return true;
    } catch {
        return false;
    }
};

const handleConfirm = async () => {
    const valid = await validate();
    if (!valid) {
        return;
    }
    const saved = await props.submitHandler?.();
    if (saved === false) {
        return;
    }
    drawerVisible.value = false;
};

const handleClose = () => {
    emit('refresh');
};

defineExpose({
    validate,
});
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
