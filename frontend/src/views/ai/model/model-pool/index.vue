<template>
    <DrawerPro
        v-model="open"
        :header="headerTitle"
        size="large"
        :auto-close="!saving && !discovering"
        @close="handleClose"
    >
        <template #content>
            <div v-loading="loading">
                <template v-if="!account.appInstallId">
                    <ModelEditor
                        ref="modelEditorRef"
                        v-model="items"
                        v-model:verify-index="draftVerifyIndex"
                        v-model:discovering="discovering"
                        :supports-discovery="supportsDiscovery"
                        :request="discoveryRequest"
                        :validate-availability="account.validateAvailability"
                        :disabled="saving"
                    />
                    <p class="input-help">{{ $t('aiTools.model.poolSaveHelp') }}</p>
                </template>
                <template v-else>
                    <div v-if="!readOnly" class="toolbar">
                        <el-button v-permission type="primary" @click="ollamaAddRef.acceptParams()">
                            {{ $t('commons.button.add') }}
                        </el-button>
                    </div>
                    <p v-if="account.validateAvailability && items.length" class="input-help verification-help">
                        {{ $t('aiTools.agents.verifyModelHelper') }}
                        {{ $t('aiTools.model.verifyModelChangeHelper') }}
                    </p>
                    <ComplexTable :data="items">
                        <el-table-column
                            v-if="account.validateAvailability"
                            :label="$t('aiTools.model.verify')"
                            width="90"
                            align="center"
                        >
                            <template #default="{ row }">
                                <el-radio
                                    v-permission
                                    class="verification-radio"
                                    :model-value="verifyModelRecordID"
                                    :value="row.recordId"
                                    :aria-label="$t('aiTools.agents.verifyModel') + ': ' + row.id"
                                    :disabled="readOnly || selectingVerifyModel || saving"
                                    @change="selectVerifyModel(row)"
                                >
                                    <span />
                                </el-radio>
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('aiTools.model.model')" prop="id" min-width="220" />
                        <el-table-column :label="$t('commons.table.name')" prop="name" min-width="180" />
                        <fu-table-operations
                            v-if="!readOnly"
                            :buttons="buttons"
                            :label="$t('commons.table.operate')"
                            :width="80"
                            fixed="right"
                        />
                    </ComplexTable>
                </template>
            </div>
        </template>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="saving || discovering" @click="open = false">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button
                    v-if="!account.appInstallId"
                    v-permission
                    type="primary"
                    :loading="saving"
                    :disabled="loading || discovering"
                    @click="saveModels"
                >
                    {{ $t('commons.button.save') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>

    <OllamaAdd ref="ollamaAddRef" :account-id="account.id" @log="onPullLog" />
    <TaskLog ref="taskLogRef" width="70%" @close="refresh" />
    <OpDialog ref="deleteRef" @search="refresh">
        <template #content>
            <el-checkbox v-if="account.provider === 'ollama'" v-model="deleteFiles" class="mt-4">
                {{ $t('aiTools.agents.deleteModelFiles') }}
            </el-checkbox>
        </template>
    </OpDialog>
</template>

<script setup lang="ts">
import ModelEditor from '../model-editor/index.vue';
import { computed, reactive, ref } from 'vue';
import OllamaAdd from '@/views/ai/ollama/add/index.vue';
import TaskLog from '@/components/log/task/index.vue';
import OpDialog from '@/components/del-dialog/index.vue';
import {
    getAgentProviders,
    updateAgentAccountModel,
    deleteAgentAccountModel,
    getAgentAccountModels,
    updateAgentAccountVerifyModel,
} from '@/api/modules/ai';
import { AI } from '@/api/interface/ai';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { getAgentProviderDisplayName } from '@/utils/agent';

const emit = defineEmits(['updated']);

const open = ref(false);
const readOnly = ref(false);
const loading = ref(false);
const saving = ref(false);
const selectingVerifyModel = ref(false);
const discovering = ref(false);
const supportsDiscovery = ref(false);
const draftVerifyIndex = ref(-1);
const modelEditorRef = ref<InstanceType<typeof ModelEditor>>();
const discoveryRequest = computed(() => ({ accountId: account.id }));
let openVersion = 0;
const items = ref<AI.AgentAccountModel[]>([]);
const verifyModelRecordID = computed(() => items.value.find((item) => item.isVerifyModel)?.recordId);
const ollamaAddRef = ref();
const taskLogRef = ref();
const deleteRef = ref();
const deleteFiles = ref(false);

const account = reactive({
    id: 0,
    name: '',
    provider: '',
    appInstallId: 0,
    validateAvailability: true,
});

const headerTitle = computed(() => {
    const accountName = account.name || getAgentProviderDisplayName(account.provider, account.provider);
    return `${accountName} · ${i18n.global.t('aiTools.agents.modelPool')}`;
});

const buttons = [
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
        disabled: (row: AI.AgentAccountModel) =>
            selectingVerifyModel.value || (account.validateAvailability && Boolean(row.isVerifyModel)),
        click: (row: AI.AgentAccountModel) => onDelete(row),
    },
];

const buildModelItem = (model?: Partial<AI.AgentAccountModel>): AI.AgentAccountModel => ({
    recordId: Number(model?.recordId || 0),
    id: String(model?.id || '').trim(),
    name: String(model?.name || '').trim(),
    isVerifyModel: Boolean(model?.isVerifyModel),
});

const selectVerifyModel = async (row: AI.AgentAccountModel) => {
    if (row.isVerifyModel || selectingVerifyModel.value || readOnly.value) return;
    selectingVerifyModel.value = true;
    try {
        await updateAgentAccountVerifyModel({ accountId: account.id, recordId: row.recordId });
        await search();
        emit('updated');
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    } finally {
        selectingVerifyModel.value = false;
    }
};

const search = async () => {
    if (!account.id) {
        items.value = [];
        return;
    }
    loading.value = true;
    const requestedAccountID = account.id;
    const version = openVersion;
    try {
        const res = await getAgentAccountModels({ accountId: requestedAccountID });
        if (!open.value || requestedAccountID !== account.id || version !== openVersion) return;
        items.value = (res.data || []).map((item) => buildModelItem(item));
        draftVerifyIndex.value = items.value.findIndex((item) => item.isVerifyModel);
    } finally {
        if (version === openVersion) loading.value = false;
    }
};

const refresh = async () => {
    await search();
    emit('updated');
};

const onPullLog = (row: { taskID: string }) => {
    taskLogRef.value.openWithTaskID(row.taskID);
};

const onDelete = (row: AI.AgentAccountModel) => {
    deleteFiles.value = false;
    deleteRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: [row.name || row.id],
        msg: i18n.global.t('commons.msg.delete', [row.name || row.id]),
        api: () =>
            deleteAgentAccountModel({
                accountId: account.id,
                recordId: row.recordId,
                deleteFiles: deleteFiles.value,
            }),
        params: {},
    });
};

const handleClose = () => {
    openVersion++;
    saving.value = false;
    discovering.value = false;
    loading.value = false;
    account.id = 0;
    account.name = '';
    account.provider = '';
    account.appInstallId = 0;
    items.value = [];
};

const openDrawer = async (row: AI.AgentAccountItem) => {
    readOnly.value = Boolean(row.appInstallId) && row.provider !== 'ollama';
    account.appInstallId = row.appInstallId || 0;
    account.id = row.id;
    account.name = row.name;
    account.provider = row.provider;
    account.validateAvailability = row.validateAvailability;
    open.value = true;
    supportsDiscovery.value = false;
    items.value = [];
    draftVerifyIndex.value = -1;
    const version = ++openVersion;
    await search();
    if (row.appInstallId || version !== openVersion) return;
    loading.value = true;
    try {
        const { data } = await getAgentProviders();
        if (version !== openVersion) return;
        const provider = data.find((item) => item.provider === row.provider);
        const preset = provider?.accountTypes.find((item) => item.key === row.accountType);
        supportsDiscovery.value = Boolean(
            (preset?.apiTypes || provider?.apiTypes)?.find((item) => item.apiType === row.apiType)
                ?.supportsModelDiscovery,
        );
    } finally {
        if (version === openVersion) loading.value = false;
    }
};

const saveModels = async () => {
    if (saving.value || discovering.value) return;
    const models = modelEditorRef.value?.validate();
    if (!models) return;
    const version = openVersion;
    saving.value = true;
    try {
        await updateAgentAccountModel({
            accountId: account.id,
            models,
            verifyModel: models[draftVerifyIndex.value]?.id || '',
        });
        emit('updated');
        if (version !== openVersion) return;
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        open.value = false;
    } finally {
        if (version === openVersion) saving.value = false;
    }
};

defineExpose({
    open: openDrawer,
});
</script>

<style lang="scss" scoped>
.toolbar {
    display: flex;
    justify-content: flex-start;
    margin-bottom: 16px;
    margin-top: 5px;
}

.verification-help {
    margin: 0 0 12px;
}

.verification-radio {
    margin-right: 0;
    height: 32px;

    :deep(.el-radio__label) {
        display: none;
    }
}
</style>
