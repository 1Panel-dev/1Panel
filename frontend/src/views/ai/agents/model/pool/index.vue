<template>
    <DrawerPro v-model="open" :header="headerTitle" size="large" @close="handleClose">
        <template #content>
            <div v-loading="loading">
                <div class="toolbar">
                    <el-button type="primary" :disabled="!hasManagePermission" @click="openCreate">
                        {{ $t('commons.button.add') }}
                    </el-button>
                </div>
                <ComplexTable :data="items">
                    <el-table-column :label="$t('aiTools.model.model')" prop="id" min-width="220" />
                    <el-table-column :label="$t('commons.table.name')" prop="name" min-width="180" />
                    <el-table-column label="Context Window" prop="contextWindow" min-width="140" />
                    <el-table-column label="Max Tokens" prop="maxTokens" min-width="120" />
                    <el-table-column :label="$t('aiTools.agents.modelInputTypes')" min-width="140">
                        <template #default="{ row }">
                            {{ (row.input || []).join(', ') || '-' }}
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('aiTools.agents.reasoning')" min-width="120">
                        <template #default="{ row }">
                            <el-tag :type="row.reasoning ? 'success' : 'info'">
                                {{ row.reasoning ? $t('commons.true') : $t('commons.false') }}
                            </el-tag>
                        </template>
                    </el-table-column>
                    <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fixed="right" />
                </ComplexTable>
            </div>
        </template>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="open = false">{{ $t('commons.button.cancel') }}</el-button>
            </span>
        </template>
    </DrawerPro>

    <DrawerPro v-model="editorOpen" :header="editorTitle" size="normal" @close="handleEditorClose">
        <el-form ref="formRef" :model="form" label-position="top" :rules="rules" v-loading="saving">
            <el-form-item :label="$t('aiTools.model.model')" prop="id">
                <el-input v-model="form.id" clearable />
            </el-form-item>
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input v-model="form.name" />
            </el-form-item>
            <el-form-item label="Context Window" prop="contextWindow">
                <el-input-number v-model="form.contextWindow" :min="1" :max="2000000" />
            </el-form-item>
            <el-form-item label="Max Tokens" prop="maxTokens">
                <el-input-number v-model="form.maxTokens" :min="1" :max="2000000" />
            </el-form-item>
            <el-form-item :label="$t('aiTools.agents.modelInputTypes')" prop="input">
                <el-checkbox-group v-model="form.input">
                    <el-checkbox label="text">Text</el-checkbox>
                    <el-checkbox label="image">Image</el-checkbox>
                </el-checkbox-group>
            </el-form-item>
            <el-form-item :label="$t('aiTools.agents.reasoning')" prop="reasoning">
                <el-switch v-model="form.reasoning" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="saving" @click="editorOpen = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button :disabled="saving || !hasManagePermission" type="primary" @click="submit">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { computed, reactive, ref } from 'vue';
import type { FormInstance } from 'element-plus';
import { ElMessageBox } from 'element-plus';
import { useMenuManagePermission } from '@/composables/useMenuManagePermission';
import {
    createAgentAccountModel,
    deleteAgentAccountModel,
    getAgentAccountModels,
    updateAgentAccountModel,
} from '@/api/modules/ai';
import { AI } from '@/api/interface/ai';
import i18n from '@/lang';
import { MsgError } from '@/utils/message';
import { getAgentProviderDisplayName } from '@/utils/agent';
import { Rules } from '@/global/form-rules';
const { hasManagePermission } = useMenuManagePermission();

const emit = defineEmits(['updated']);

const open = ref(false);
const editorOpen = ref(false);
const loading = ref(false);
const saving = ref(false);
const formRef = ref<FormInstance>();
const items = ref<AI.AgentAccountModel[]>([]);

const account = reactive({
    id: 0,
    name: '',
    provider: '',
});

const form = reactive<AI.AgentAccountModel>({
    recordId: 0,
    id: '',
    name: '',
    contextWindow: 128000,
    maxTokens: 8192,
    reasoning: false,
    input: ['text'],
});

const headerTitle = computed(() => {
    const accountName = account.name || getAgentProviderDisplayName(account.provider, account.provider);
    return `${accountName} · ${i18n.global.t('aiTools.agents.modelPool')}`;
});

const editorTitle = computed(() =>
    form.recordId ? i18n.global.t('commons.button.edit') : i18n.global.t('commons.button.add'),
);

const rules = reactive({
    id: [Rules.requiredInput, Rules.noSpace],
    name: [Rules.requiredInput],
    contextWindow: [Rules.integerNumber],
    maxTokens: [Rules.integerNumber],
    input: [Rules.requiredSelect],
});

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        disabled: () => !hasManagePermission.value,
        click: (row: AI.AgentAccountModel) => openEdit(row),
    },
    {
        label: i18n.global.t('commons.button.delete'),
        disabled: () => !hasManagePermission.value,
        click: (row: AI.AgentAccountModel) => onDelete(row),
    },
];

const buildModelItem = (model?: Partial<AI.AgentAccountModel>): AI.AgentAccountModel => ({
    recordId: Number(model?.recordId || 0),
    id: String(model?.id || '').trim(),
    name: String(model?.name || '').trim(),
    contextWindow: Number(model?.contextWindow || 0),
    maxTokens: Number(model?.maxTokens || 0),
    reasoning: Boolean(model?.reasoning),
    input: Array.from(new Set((model?.input || []).filter((value) => value === 'text' || value === 'image'))),
});

const search = async () => {
    if (!account.id) {
        items.value = [];
        return;
    }
    loading.value = true;
    try {
        const res = await getAgentAccountModels({ accountId: account.id });
        items.value = (res.data || []).map((item) => buildModelItem(item));
    } finally {
        loading.value = false;
    }
};

const resetForm = () => {
    form.recordId = 0;
    form.id = '';
    form.name = '';
    form.contextWindow = 128000;
    form.maxTokens = 8192;
    form.reasoning = false;
    form.input = ['text'];
};

const openCreate = () => {
    resetForm();
    editorOpen.value = true;
};

const openEdit = (row: AI.AgentAccountModel) => {
    Object.assign(form, buildModelItem(row));
    editorOpen.value = true;
};

const submit = async () => {
    if (!account.id || !formRef.value) {
        return;
    }
    await formRef.value.validate();
    const duplicate = items.value.some((item) => item.id === form.id && item.recordId !== form.recordId);
    if (duplicate) {
        MsgError(i18n.global.t('aiTools.agents.accountModelsDuplicate'));
        return;
    }
    saving.value = true;
    try {
        if (form.recordId) {
            await updateAgentAccountModel({
                accountId: account.id,
                model: buildModelItem(form),
            });
        } else {
            await createAgentAccountModel({
                accountId: account.id,
                model: buildModelItem(form),
            });
        }
        editorOpen.value = false;
        await search();
        emit('updated');
    } catch (error: any) {
        MsgError(String(error?.message || i18n.global.t('commons.res.commonError')));
    } finally {
        saving.value = false;
    }
};

const onDelete = async (row: AI.AgentAccountModel) => {
    await ElMessageBox.confirm(
        i18n.global.t('commons.msg.delete', [row.name || row.id]),
        i18n.global.t('commons.button.delete'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        },
    );
    await deleteAgentAccountModel({
        accountId: account.id,
        recordId: row.recordId,
    });
    await search();
    emit('updated');
};

const handleEditorClose = () => {
    formRef.value?.resetFields();
    resetForm();
    saving.value = false;
};

const handleClose = () => {
    loading.value = false;
    account.id = 0;
    account.name = '';
    account.provider = '';
    items.value = [];
    editorOpen.value = false;
    handleEditorClose();
};

const openDrawer = async (row: AI.AgentAccountItem) => {
    account.id = row.id;
    account.name = row.name;
    account.provider = row.provider;
    open.value = true;
    await search();
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
</style>
