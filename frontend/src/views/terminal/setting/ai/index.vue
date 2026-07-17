<template>
    <el-button @click="openDrawer" type="primary">
        {{ $t('terminal.aiAssistant') }}
    </el-button>

    <DrawerPro v-model="drawerVisible" :header="$t('terminal.aiAssistant')" size="60%" @close="handleClose">
        <div v-loading="loading">
            <el-form ref="formRef" :model="formModel" label-position="top">
                <el-form-item :label="$t('terminal.aiAssistant')">
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
                    <span class="input-help">
                        {{ $t('terminal.aiPrefixHelper', [formModel.prefix || DEFAULT_AI_PREFIX]) }}
                    </span>
                </el-form-item>
                <el-form-item :label="$t('terminal.aiRiskCommands')">
                    <CodemirrorPro
                        v-model="formModel.riskCommands"
                        mode="shell"
                        line-wrapping
                        :height-diff="560"
                        :min-height="180"
                    ></CodemirrorPro>
                    <span class="input-help">{{ $t('terminal.aiRiskCommandsHelper') }}</span>
                </el-form-item>
            </el-form>
        </div>
        <template #footer>
            <el-button @click="drawerVisible = false" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
            <el-button plain @click="resetRiskCommands" :disabled="loading">
                {{ $t('commons.button.setDefault') }}
            </el-button>
            <el-button type="primary" @click="handleConfirm" :loading="loading">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { nextTick, reactive, ref } from 'vue';
import type { ElForm } from 'element-plus';
import CodemirrorPro from '@/components/codemirror-pro/index.vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { pageAgentAccounts } from '@/api/modules/ai';
import { getAgentTerminalAIInfo, updateAgentTerminalAIInfo } from '@/api/modules/setting';
import { MsgSuccess } from '@/utils/message';
import { isAgentAccountVerificationSkipped } from '@/utils/agent';
import {
    AI_PREFIX_OPTIONS,
    DEFAULT_AI_PREFIX,
    formatRiskCommandsText,
    normalizeRiskCommands,
    parseRiskCommands,
} from '@/views/terminal/setting/ai/helper';

interface AgentAccountOption {
    id: number | string;
    name: string;
    provider?: string;
    providerName?: string;
    verified?: boolean;
}

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();
const drawerVisible = ref(false);
const loading = ref(false);
const agentAccountOptions = ref<AgentAccountOption[]>([]);
const defaultRiskCommands = ref<string[]>([]);
const prefixOptions = AI_PREFIX_OPTIONS;
const formModel = reactive({
    status: 'Disable',
    accountId: '',
    prefix: DEFAULT_AI_PREFIX,
    riskCommands: '',
});

const loadTerminalAISettings = async () => {
    const res = await getAgentTerminalAIInfo();
    formModel.status = res.data.aiStatus || 'Disable';
    formModel.accountId = res.data.aiAccountId || '';
    formModel.prefix = res.data.aiPrefix || DEFAULT_AI_PREFIX;
    formModel.riskCommands = formatRiskCommandsText(parseRiskCommands(res.data.aiRiskCommands || ''));
    defaultRiskCommands.value = parseRiskCommands(res.data.aiRiskCommandsDefault || '');
};

const openDrawer = () => {
    drawerVisible.value = true;
    loading.value = true;
    Promise.all([loadTerminalAISettings(), loadAgentAccounts()]).finally(() => {
        loading.value = false;
    });
};

const verificationLabel = (item: AgentAccountOption) => {
    if (isAgentAccountVerificationSkipped(item.provider)) {
        return i18n.global.t('aiTools.agents.verifySkipped');
    }
    return item.verified ? 'OK' : 'N/A';
};

const verificationTagType = (item: AgentAccountOption) => {
    if (isAgentAccountVerificationSkipped(item.provider)) {
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

const resetRiskCommands = () => {
    formModel.prefix = DEFAULT_AI_PREFIX;
    formModel.riskCommands = formatRiskCommandsText(defaultRiskCommands.value);
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
    loading.value = true;
    try {
        await updateAgentTerminalAIInfo({
            aiStatus: formModel.status,
            aiAccountId: formModel.status === 'Enable' ? formModel.accountId : '',
            aiPrefix: formModel.prefix.trim() || DEFAULT_AI_PREFIX,
            aiRiskCommands: JSON.stringify(normalizeRiskCommands(parseRiskCommands(formModel.riskCommands))),
        });
        await loadTerminalAISettings();
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        drawerVisible.value = false;
    } finally {
        loading.value = false;
    }
};

const handleClose = () => {
    formRef.value?.clearValidate();
};
</script>

<style lang="css" scoped>
.formInput {
    width: 100%;
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
