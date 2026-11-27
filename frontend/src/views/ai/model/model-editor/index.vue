<template>
    <section class="model-editor" :aria-label="$t('aiTools.agents.modelPool')">
        <div class="model-editor__header">
            <div>
                <strong>{{ $t('aiTools.agents.modelPool') }}</strong>
                <el-tag size="small" type="info" round>{{ models.filter((item) => item.id.trim()).length }}</el-tag>
                <p>{{ $t('aiTools.model.poolEditHelp') }}</p>
            </div>
            <el-button
                v-permission="'ai_model_manage'"
                :icon="Refresh"
                :disabled="!supportsDiscovery || disabled"
                :loading="discovering"
                @click="discover"
            >
                {{ $t('aiTools.agents.discoverModels') }}
            </el-button>
        </div>
        <el-alert
            v-if="!supportsDiscovery"
            class="model-editor__notice"
            type="info"
            :closable="false"
            show-icon
            :title="$t('aiTools.model.poolUnavailable')"
        />
        <el-alert
            v-if="discoveryError"
            class="model-editor__notice"
            type="warning"
            :closable="false"
            show-icon
            :title="discoveryError"
        />
        <div v-if="validateAvailability && models.length" class="model-editor__verify">
            {{ $t('aiTools.agents.verifyModelHelper') }}
        </div>
        <el-table :data="models" max-height="360" class="model-editor__table">
            <el-table-column v-if="validateAvailability" :label="$t('aiTools.model.verify')" width="72" align="center">
                <template #default="{ row, $index }">
                    <el-radio
                        v-model="verifyIndex"
                        class="model-editor__radio"
                        :value="$index"
                        :disabled="disabled || discovering || !row.id.trim()"
                        :aria-label="$t('aiTools.agents.verifyModel') + ': ' + row.id"
                    >
                        <span />
                    </el-radio>
                </template>
            </el-table-column>
            <el-table-column label="Model ID" min-width="210">
                <template #default="{ row }">
                    <el-select
                        v-if="localModels.length"
                        v-model="row.id"
                        filterable
                        allow-create
                        default-first-option
                        :disabled="disabled || discovering"
                        :aria-label="'Model ID'"
                    >
                        <el-option v-for="id in localModels" :key="id" :label="id" :value="id" />
                    </el-select>
                    <el-input v-else v-model="row.id" :disabled="disabled || discovering" aria-label="Model ID" />
                </template>
            </el-table-column>
            <el-table-column :label="$t('commons.table.name')" min-width="160">
                <template #default="{ row }">
                    <el-input
                        v-model="row.name"
                        :disabled="disabled || discovering"
                        :placeholder="row.id"
                        :aria-label="$t('commons.table.name')"
                    />
                </template>
            </el-table-column>
            <el-table-column width="60" align="center">
                <template #default="{ $index }">
                    <el-button
                        v-permission="'ai_model_manage'"
                        link
                        :icon="Delete"
                        :disabled="disabled || discovering"
                        :aria-label="$t('commons.button.delete')"
                        @click="remove($index)"
                    />
                </template>
            </el-table-column>
            <template #empty>
                <div class="model-editor__empty">{{ $t('aiTools.model.poolEmpty') }}</div>
            </template>
        </el-table>
        <el-button
            v-permission="'ai_model_manage'"
            class="model-editor__add"
            :icon="Plus"
            :disabled="disabled || discovering || (singleModelOnly && models.length > 0)"
            @click="add"
        >
            {{ $t('commons.button.add') }}
        </el-button>
    </section>
</template>

<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue';
import { ElMessageBox } from 'element-plus';
import { Delete, Plus, Refresh } from '@element-plus/icons-vue';
import type { AI } from '@/api/interface/ai';
import { discoverAgentAccountModels } from '@/api/modules/ai';
import i18n from '@/lang';
import { MsgError } from '@/utils/message';

const models = defineModel<AI.AgentAccountModel[]>({ required: true });
const verifyIndex = defineModel<number>('verifyIndex', { default: -1 });
const discovering = defineModel<boolean>('discovering', { default: false });
const props = withDefaults(
    defineProps<{
        supportsDiscovery: boolean;
        request: AI.AgentAccountModelDiscoverReq;
        validateAvailability: boolean;
        disabled?: boolean;
        singleModelOnly?: boolean;
        localModels?: string[];
        beforeDiscover?: () => Promise<boolean>;
    }>(),
    { localModels: () => [] },
);
const discoveryError = ref('');
let requestVersion = 0;
onBeforeUnmount(() => {
    requestVersion++;
});
watch(
    () => props.request,
    () => {
        requestVersion++;
        discovering.value = false;
        discoveryError.value = '';
    },
    { deep: true },
);

const add = () => {
    models.value = [...models.value, { recordId: 0, id: '', name: '' }];
    if (props.singleModelOnly) verifyIndex.value = 0;
};
const remove = (index: number) => {
    models.value = models.value.filter((_, i) => i !== index);
    if (verifyIndex.value === index) verifyIndex.value = -1;
    else if (verifyIndex.value > index) verifyIndex.value--;
};
const discover = async () => {
    if (!props.supportsDiscovery || discovering.value || props.disabled) return;
    if (props.beforeDiscover && !(await props.beforeDiscover())) return;
    const version = ++requestVersion;
    if (models.value.some((item) => item.id.trim() || item.name.trim())) {
        try {
            await ElMessageBox.confirm(
                i18n.global.t('aiTools.model.poolOverwrite'),
                i18n.global.t('aiTools.agents.discoverModels'),
                {
                    type: 'warning',
                    confirmButtonText: i18n.global.t('commons.button.confirm'),
                    cancelButtonText: i18n.global.t('commons.button.cancel'),
                },
            );
        } catch {
            return;
        }
    }
    if (version !== requestVersion) return;
    discovering.value = true;
    discoveryError.value = '';
    try {
        const { data } = await discoverAgentAccountModels({ ...props.request });
        if (version !== requestVersion) return;
        if (!data?.length) {
            discoveryError.value = i18n.global.t('aiTools.model.poolFetchEmpty');
            return;
        }
        if (props.singleModelOnly && data.length > 1) {
            discoveryError.value = i18n.global.t('aiTools.agents.accountModelsSingle');
            return;
        }
        const previousVerify = models.value[verifyIndex.value]?.id;
        const nextModels = data.map((item) => ({ recordId: 0, id: item.id, name: item.name || item.id }));
        models.value = nextModels;
        verifyIndex.value = nextModels.findIndex((item) => item.id === previousVerify);
    } catch {
        if (version === requestVersion) discoveryError.value = i18n.global.t('aiTools.model.poolFetchFailed');
    } finally {
        if (version === requestVersion) discovering.value = false;
    }
};

const validate = () => {
    const rows = models.value.map((item) => ({
        ...item,
        id: item.id.trim(),
        name: item.name.trim() || item.id.trim(),
    }));
    let error = '';
    if (!rows.length || rows.some((item) => !item.id)) error = 'aiTools.agents.accountModelsRequired';
    else if (rows.some((item) => /\s/.test(item.id))) error = 'setting.noSpace';
    else if (new Set(rows.map((item) => item.id)).size !== rows.length) error = 'aiTools.agents.accountModelsDuplicate';
    else if (props.singleModelOnly && rows.length > 1) error = 'aiTools.agents.accountModelsSingle';
    else if (props.validateAvailability && !rows[verifyIndex.value]) error = 'aiTools.agents.verifyModelRequired';
    if (error) {
        MsgError(i18n.global.t(error));
        return null;
    }
    return rows;
};
defineExpose({ validate });
</script>

<style scoped lang="scss">
.model-editor {
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 8px;
    overflow: hidden;
    margin: 16px 0;
    &__header {
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        gap: 16px;
        padding: 18px 18px 14px;
        background: var(--el-fill-color-light);
        strong {
            margin-right: 8px;
            font-size: 15px;
            color: var(--el-text-color-primary);
        }
        p {
            margin: 8px 0 0;
            font-size: 12px;
            line-height: 1.6;
            color: var(--el-text-color-secondary);
        }
        .el-button {
            flex-shrink: 0;
        }
    }
    &__notice {
        margin: 12px 16px;
        width: auto;
    }
    &__verify {
        padding: 12px 16px;
        font-size: 12px;
        color: var(--el-text-color-secondary);
        line-height: 1.6;
    }
    &__empty {
        padding: 24px 16px;
        line-height: 1.6;
    }
    &__add {
        margin: 12px 16px 16px;
    }
    &__radio {
        margin-right: 0;
        :deep(.el-radio__label) {
            display: none;
        }
    }
    @media (max-width: 640px) {
        &__header {
            flex-direction: column;
        }
    }
}
</style>
