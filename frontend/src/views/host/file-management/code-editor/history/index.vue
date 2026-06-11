<template>
    <DrawerPro
        v-model="drawerVisible"
        :header="$t('file.history')"
        :size="drawerSize"
        @close="handleClose"
        destroy-on-close
    >
        <div class="flex flex-col gap-4 min-h-[60vh]">
            <el-alert type="info" :closable="false" show-icon>
                {{ $t('file.historyHint') }}
            </el-alert>
            <el-alert v-if="isRestoringCurrentFile && currentFile.dirty" type="warning" :closable="false" show-icon>
                {{ $t('file.historyDirtyHint') }}
            </el-alert>

            <el-card shadow="never" class="history-setting">
                <el-collapse v-model="activeCollapse" :accordion="true" class="!border-0">
                    <el-collapse-item :title="$t('file.historySettingTitle')" name="history" class="!border-0">
                        <el-form
                            ref="historySettingFormRef"
                            :model="historySetting"
                            :rules="historySettingRules"
                            label-position="top"
                            class="flex items-end gap-6"
                        >
                            <el-form-item :label="$t('file.historyEnable')" prop="enable">
                                <el-switch
                                    v-model="historySetting.enable"
                                    active-value="Enable"
                                    inactive-value="Disable"
                                />
                            </el-form-item>
                            <el-form-item :label="$t('file.historyMaxPerPath')" prop="maxPerPath">
                                <el-input-number
                                    v-model="historySetting.maxPerPath"
                                    :min="1"
                                    :max="1000"
                                    class="!w-52"
                                />
                            </el-form-item>
                            <el-form-item :label="$t('file.historyDiskQuota')" prop="diskQuotaMB">
                                <el-input-number
                                    v-model="historySetting.diskQuotaMB"
                                    :min="1"
                                    :max="1048576"
                                    class="!w-52"
                                />
                            </el-form-item>
                            <el-form-item class="ml-auto">
                                <el-button type="primary" :loading="settingSaving" @click="saveHistorySetting">
                                    {{ $t('commons.button.save') }}
                                </el-button>
                            </el-form-item>
                        </el-form>
                    </el-collapse-item>
                </el-collapse>
            </el-card>

            <div class="history-workspace">
                <el-card shadow="never" class="min-h-0 history-list-card">
                    <template #header>
                        <div class="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                            <el-tabs v-model="scope" type="card" class="flex-1" @tab-change="handleScopeChange">
                                <el-tab-pane
                                    :label="$t('file.historyCurrentScope')"
                                    name="current"
                                    :disabled="!canSwitchToCurrentScope"
                                />
                                <el-tab-pane :label="$t('file.historyAllScope')" name="all" />
                            </el-tabs>
                            <div class="flex flex-wrap items-center gap-2">
                                <el-select
                                    v-model="operationFilter"
                                    clearable
                                    class="w-full sm:!w-36"
                                    :placeholder="$t('commons.table.type')"
                                    @change="loadHistoryList(true)"
                                >
                                    <el-option
                                        v-for="item in operationOptions"
                                        :key="item.value"
                                        :label="item.label"
                                        :value="item.value"
                                    />
                                </el-select>
                                <el-button
                                    class="w-full sm:w-auto"
                                    :loading="historyLoading"
                                    @click="loadHistoryList(true)"
                                >
                                    {{ $t('commons.button.refresh') }}
                                </el-button>
                                <el-button
                                    class="w-full sm:w-auto !ml-0"
                                    v-permission
                                    v-node-admin
                                    :disabled="selected.length === 0"
                                    :loading="deleteLoading"
                                    @click="deleteSelected(null)"
                                >
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                            </div>
                        </div>
                    </template>

                    <div class="history-table-box">
                        <div class="history-table-wrap">
                            <ComplexTable
                                :pagination-config="paginationConfig"
                                :data="historyItems"
                                v-loading="historyLoading"
                                class="history-table"
                                @search="loadHistoryList()"
                                @row-click="openHistoryRecord"
                                @selection-change="handleSelectionChange"
                            >
                                <el-table-column type="selection" width="46" />
                                <el-table-column :label="$t('commons.table.date')" min-width="180">
                                    <template #default="{ row }">
                                        {{ dateFormatSimpleWithSecond(row.createdAt) }}
                                    </template>
                                </el-table-column>
                                <el-table-column
                                    :label="$t('file.fileName')"
                                    prop="fileName"
                                    min-width="160"
                                    show-overflow-tooltip
                                />
                                <el-table-column
                                    :label="$t('file.path')"
                                    prop="path"
                                    min-width="260"
                                    show-overflow-tooltip
                                />
                                <el-table-column :label="$t('commons.table.type')" min-width="110">
                                    <template #default="{ row }">
                                        <div class="flex flex-wrap items-center gap-2">
                                            <el-tag size="small" effect="plain">
                                                {{ getOperationLabel(row.operation) }}
                                            </el-tag>
                                        </div>
                                    </template>
                                </el-table-column>
                                <el-table-column :label="$t('file.size')" min-width="110">
                                    <template #default="{ row }">
                                        {{ computeSize(row.contentSize) }}
                                    </template>
                                </el-table-column>
                                <el-table-column :label="$t('commons.table.operate')" min-width="150" fixed="right">
                                    <template #default="{ row }">
                                        <el-button link type="primary" @click.stop="openHistoryRecord(row)">
                                            {{ $t('commons.button.view') }}
                                        </el-button>
                                        <el-button
                                            v-permission
                                            v-node-admin
                                            link
                                            type="primary"
                                            @click.stop="deleteSelected(row)"
                                        >
                                            {{ $t('commons.button.delete') }}
                                        </el-button>
                                    </template>
                                </el-table-column>
                            </ComplexTable>
                        </div>
                    </div>
                </el-card>

                <el-card v-if="selectedHistory" shadow="never" class="min-h-0 history-preview-card">
                    <template #header>
                        <div class="flex flex-col gap-1 lg:flex-row lg:items-center lg:justify-between">
                            <div class="min-w-0">
                                <div class="font-medium truncate">{{ selectedHistory.fileName }}</div>
                                <div class="text-xs text-[var(--el-text-color-secondary)] truncate">
                                    {{ selectedHistory.path }}
                                </div>
                                <div
                                    v-if="selectedHistory.sourcePath || selectedHistory.targetPath"
                                    class="text-xs text-[var(--el-text-color-secondary)] truncate"
                                >
                                    {{ selectedHistory.sourcePath || selectedHistory.path }}
                                    <span v-if="selectedHistory.targetPath">→ {{ selectedHistory.targetPath }}</span>
                                </div>
                            </div>
                            <el-space>
                                <el-tag effect="plain">{{ getOperationLabel(selectedHistory.operation) }}</el-tag>
                                <el-tag effect="plain">{{ computeSize(selectedHistory.contentSize) }}</el-tag>
                                <el-button
                                    v-if="canRestoreSelected"
                                    v-permission
                                    v-node-admin
                                    type="primary"
                                    :loading="restoreLoading"
                                    @click="confirmRestoreSelected"
                                >
                                    {{ $t('file.historyRestore') }}
                                </el-button>
                            </el-space>
                        </div>
                    </template>
                    <div class="pb-2 text-xs text-[var(--el-text-color-secondary)]">
                        {{ $t('file.historyDiffHintCurrent') }}
                    </div>
                    <div v-if="compareTargetPath" class="pb-2 text-xs text-[var(--el-text-color-secondary)]">
                        {{ $t('file.historyCompareTargetPath') }}{{ $t('commons.colon') }}{{ compareTargetPath }}
                    </div>
                    <el-alert
                        v-if="selectedHistory && !isVersionRecord(selectedHistory.operation)"
                        type="info"
                        :closable="false"
                        show-icon
                    >
                        {{ $t('file.historyOperationOnlyHint') }}
                    </el-alert>
                    <div v-else ref="diffContainer" class="history-diff"></div>
                </el-card>
            </div>
        </div>
    </DrawerPro>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, onUnmounted, reactive, ref } from 'vue';
import { dateFormatSimpleWithSecond } from '@/utils/date';
import { computeSize } from '@/utils/size';
import { MsgError, MsgSuccess } from '@/utils/message';
import { deleteFileHistory, getFileHistoryContent, searchFileHistory, restoreFileHistory } from '@/api/modules/files';
import { getAgentFileHistoryInfo, updateAgentFileHistoryInfo } from '@/api/modules/setting';
import { File } from '@/api/interface/file';
import { Setting } from '@/api/interface/setting';
import { loadMonacoLanguageSupport, setupMonacoEnvironment } from '@/utils/monaco';
import { ElMessageBox, type FormInstance, type FormRules } from 'element-plus';
import { Languages } from '@/global/mimetype';
import i18n from '@/lang';
import ComplexTable from '@/components/complex-table/index.vue';

type MonacoEditorApi = typeof import('monaco-editor/esm/vs/editor/editor.api');

interface HistoryDrawerProps {
    path: string;
    content: string;
    language: string;
    extension: string;
    dirty: boolean;
    scope?: 'current' | 'all';
}

const drawerVisible = ref(false);
const historyLoading = ref(false);
const deleteLoading = ref(false);
const settingSaving = ref(false);
const restoreLoading = ref(false);
const historySettingFormRef = ref<FormInstance>();
const scope = ref<'current' | 'all'>('current');
const operationFilter = ref('');
const activeCollapse = ref([]);

const paginationConfig = reactive({
    cacheSizeKey: 'file-history-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('file-history-page-size')) || 20,
    total: 0,
    small: false,
});
const currentFile = ref<HistoryDrawerProps>({
    path: '',
    content: '',
    language: 'plaintext',
    extension: '',
    dirty: false,
});
const historySetting = ref<Setting.FileHistoryInfo>({
    enable: 'Enable',
    maxPerPath: 20,
    diskQuotaMB: 1024,
});
const historySettingRules = reactive<FormRules<Setting.FileHistoryInfo>>({
    enable: [
        {
            required: true,
            message: i18n.global.t('commons.rule.requiredInput'),
            trigger: 'change',
        },
    ],
    maxPerPath: [
        {
            required: true,
            message: i18n.global.t('commons.rule.requiredInput'),
            trigger: 'change',
        },
        {
            validator: (_, value, callback) => {
                const parsedValue = Number(value);
                if (!Number.isInteger(parsedValue) || parsedValue <= 0) {
                    callback(new Error(i18n.global.t('commons.rule.integer')));
                    return;
                }
                callback();
            },
            trigger: 'change',
        },
    ],
    diskQuotaMB: [
        {
            required: true,
            message: i18n.global.t('commons.rule.requiredInput'),
            trigger: 'change',
        },
        {
            validator: (_, value, callback) => {
                const parsedValue = Number(value);
                if (!Number.isInteger(parsedValue) || parsedValue <= 0) {
                    callback(new Error(i18n.global.t('commons.rule.integer')));
                    return;
                }
                callback();
            },
            trigger: 'change',
        },
    ],
});
const historyItems = ref<File.FileHistoryInfo[]>([]);
const selected = ref<File.FileHistoryInfo[]>([]);
const selectedHistory = ref<File.FileHistoryInfo | null>(null);
const historyContent = ref('');
const currentFileContent = ref('');
const compareTargetPath = ref('');
const windowWidth = ref(window.innerWidth);
const operationMap = {
    '': i18n.global.t('app.all'),
    save: i18n.global.t('commons.button.save'),
    restore: i18n.global.t('commons.button.recover'),
    move: i18n.global.t('file.move'),
    rename: i18n.global.t('file.rename'),
} as const;

const operationOptions = Object.entries(operationMap).map(([value, label]) => ({
    label,
    value,
}));

const getOperationLabel = (operation: string) => {
    return operationMap[operation as keyof typeof operationMap] || operation;
};
const isVersionRecord = (operation: string) => operation === 'save' || operation === 'restore';
const canRestoreSelected = computed(() => isVersionRecord(selectedHistory.value?.operation || ''));
const isRestoringCurrentFile = computed(() => selectedHistory.value?.path === currentFile.value.path);
const getCurrentTargetPath = (row: File.FileHistoryInfo) => row.currentPath || row.path;
const canSwitchToCurrentScope = computed(() => !!currentFile.value.path || historyItems.value.length > 0);

const syncCurrentFileContext = (history?: File.FileHistoryInfo | null) => {
    if (currentFile.value.path) {
        return;
    }
    const target = history || historyItems.value[0];
    if (!target) {
        return;
    }
    currentFile.value = {
        ...currentFile.value,
        path: target.currentPath || target.path || '',
        content: target.currentContent || currentFileContent.value || currentFile.value.content || '',
        extension: target.extension || currentFile.value.extension || '',
        language: currentFile.value.language || '',
        dirty: false,
        scope: 'current',
    };
};

const diffContainer = ref<HTMLElement | null>(null);
const emit = defineEmits(['restored']);
let monacoApi: MonacoEditorApi | null = null;
let diffEditor: ReturnType<MonacoEditorApi['editor']['createDiffEditor']> | null = null;
let originalModel: ReturnType<MonacoEditorApi['editor']['createModel']> | null = null;
let modifiedModel: ReturnType<MonacoEditorApi['editor']['createModel']> | null = null;

const currentLanguage = computed(() => {
    const rawLanguage = currentFile.value.language || 'plaintext';
    if (rawLanguage) {
        return rawLanguage;
    }
    const extension = currentFile.value.extension?.replace('.', '') || '';
    const matched = Languages.find((item) => item.value.includes(extension));
    return matched?.label || 'plaintext';
});

const drawerSize = computed(() => {
    if (windowWidth.value >= 1600) {
        return '90%';
    }
    if (windowWidth.value >= 1280) {
        return '80%';
    }
    if (windowWidth.value >= 960) {
        return '100%';
    }
    return '100%';
});

const updateWindowWidth = () => {
    windowWidth.value = window.innerWidth;
};

const useWindowSize = () => {
    onMounted(() => window.addEventListener('resize', updateWindowWidth));
    onUnmounted(() => window.removeEventListener('resize', updateWindowWidth));
};

const ensureMonaco = async () => {
    if (monacoApi) {
        return monacoApi;
    }
    setupMonacoEnvironment();
    const [monacoModule] = await Promise.all([
        import('monaco-editor/esm/vs/editor/editor.api'),
        loadMonacoLanguageSupport(),
    ]);
    monacoApi = monacoModule;
    return monacoApi;
};

const disposeDiffEditor = () => {
    if (diffEditor) {
        diffEditor.dispose();
        diffEditor = null;
    }
    originalModel?.dispose();
    modifiedModel?.dispose();
    originalModel = null;
    modifiedModel = null;
};

const renderDiff = async () => {
    if (!selectedHistory.value) {
        disposeDiffEditor();
        return;
    }
    const monaco = await ensureMonaco();
    await nextTick();
    if (!diffContainer.value) {
        disposeDiffEditor();
        return;
    }

    disposeDiffEditor();
    originalModel = monaco.editor.createModel(historyContent.value, currentLanguage.value);
    modifiedModel = monaco.editor.createModel(currentFileContent.value, currentLanguage.value);
    diffEditor = monaco.editor.createDiffEditor(diffContainer.value, {
        theme: 'vs-dark',
        readOnly: true,
        automaticLayout: true,
        folding: true,
        glyphMargin: true,
        roundedSelection: false,
        overviewRulerBorder: false,
        renderMarginRevertIcon: true,
        renderGutterMenu: false,
    });
    diffEditor.setModel({
        original: originalModel,
        modified: modifiedModel,
    });
};

const loadHistorySetting = async () => {
    const res = await getAgentFileHistoryInfo();
    historySetting.value = res.data;
};

const normalizeHistorySettingNumber = (value: number, defaultValue: number) => {
    const parsedValue = Number(value);
    return parsedValue > 0 ? parsedValue : defaultValue;
};

const saveHistorySetting = async () => {
    settingSaving.value = true;
    try {
        const valid = await historySettingFormRef.value?.validate().catch(() => false);
        if (!valid) {
            return;
        }
        await updateAgentFileHistoryInfo({
            ...historySetting.value,
            maxPerPath: normalizeHistorySettingNumber(historySetting.value.maxPerPath, 20),
            diskQuotaMB: normalizeHistorySettingNumber(historySetting.value.diskQuotaMB, 1024),
        });
        MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
    } catch (error) {
        MsgError(String(error));
    } finally {
        settingSaving.value = false;
    }
};

const loadHistoryList = async (resetPage = false) => {
    if (resetPage) {
        paginationConfig.currentPage = 1;
    }
    historyLoading.value = true;
    try {
        const currentScopePath =
            currentFile.value.path ||
            selectedHistory.value?.currentPath ||
            selectedHistory.value?.path ||
            historyItems.value[0]?.currentPath ||
            historyItems.value[0]?.path ||
            '';
        const res = await searchFileHistory({
            page: paginationConfig.currentPage,
            pageSize: paginationConfig.pageSize,
            path: scope.value === 'current' ? currentScopePath : '',
            scope: scope.value,
            operation: operationFilter.value,
        });
        historyItems.value = res.data.items || [];
        paginationConfig.total = res.data.total || 0;
        if (historyItems.value.length > 0) {
            const existSelected = selectedHistory.value
                ? historyItems.value.find((item) => item.id === selectedHistory.value?.id)
                : null;
            try {
                if (existSelected) {
                    await openHistoryRecord(existSelected);
                } else {
                    await openHistoryRecord(historyItems.value[0]);
                }
            } catch (error) {
                MsgError(String(error));
            }
        } else {
            selectedHistory.value = null;
            historyContent.value = '';
            currentFileContent.value = '';
            compareTargetPath.value = '';
            disposeDiffEditor();
        }
    } finally {
        historyLoading.value = false;
    }
};

const loadHistoryContent = async (historyId: number) => {
    const res = await getFileHistoryContent(historyId);
    return res.data;
};

const openHistoryRecord = async (row: File.FileHistoryInfo) => {
    selectedHistory.value = row;
    try {
        const data = await loadHistoryContent(row.id);
        historyContent.value = data.content ?? '';
        currentFileContent.value = data.currentContent ?? data.content ?? currentFile.value.content ?? '';
        compareTargetPath.value = getCurrentTargetPath(row);
        await renderDiff();
    } catch (error) {
        MsgError(String(error));
        throw error;
    }
};

const handleSelectionChange = (rows: File.FileHistoryInfo[]) => {
    selected.value = rows;
};

const deleteSelected = async (row: File.FileHistoryInfo | null) => {
    const ids = selected.value.map((item) => item.id);
    const targetIds = row ? [row.id] : ids;
    if (!targetIds.length) return;
    try {
        await ElMessageBox.confirm(i18n.global.t('file.historyDeleteConfirm'), {
            confirmButtonText: i18n.global.t('commons.button.delete'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'warning',
        });
    } catch {
        return;
    }
    deleteLoading.value = true;
    try {
        await deleteFileHistory(targetIds);
        MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
        selected.value = [];
        await loadHistoryList(true);
    } finally {
        deleteLoading.value = false;
    }
};

const handleScopeChange = async () => {
    if (scope.value === 'current' && !canSwitchToCurrentScope.value) {
        scope.value = 'all';
        return;
    }
    if (scope.value === 'current') {
        syncCurrentFileContext(historyItems.value[0]);
    }
    await loadHistoryList(true);
};

const handleClose = () => {
    drawerVisible.value = false;
    disposeDiffEditor();
};

const confirmRestoreSelected = async () => {
    if (!selectedHistory.value) {
        return;
    }
    try {
        const confirmMessage = [
            currentFile.value.dirty
                ? i18n.global.t('file.historyRestoreDirtyConfirm')
                : i18n.global.t('file.historyRestoreConfirm'),
            '',
            `${i18n.global.t('file.historyCompareTarget')}${i18n.global.t('commons.colon')}${selectedHistory.value.fileName}`,
            `${i18n.global.t('file.historyCompareTargetPath')}${i18n.global.t('commons.colon')}${compareTargetPath.value || selectedHistory.value.path}`,
            '',
            currentFile.value.dirty ? i18n.global.t('file.historyDirtyHint') : '',
        ]
            .filter(Boolean)
            .join('\n');
        await ElMessageBox.confirm(confirmMessage, {
            confirmButtonText: i18n.global.t('commons.button.recover'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'warning',
            distinguishCancelAndClose: true,
        });
    } catch {
        return;
    }
    restoreLoading.value = true;
    try {
        const res = await restoreFileHistory(selectedHistory.value.id);
        MsgSuccess(i18n.global.t('file.historyRestoreSuccess'));
        currentFile.value.content = res.data.content || currentFile.value.content;
        currentFile.value.dirty = false;
        emit('restored', res.data.path);
        await loadHistoryList(true);
    } catch (error) {
        MsgError(String(error));
    } finally {
        restoreLoading.value = false;
    }
};

const acceptParams = async (params: HistoryDrawerProps): Promise<void> => {
    currentFile.value = params;
    operationFilter.value = '';
    scope.value = params.scope || 'current';
    if (scope.value === 'current' && !params.path) {
        scope.value = 'all';
    }
    drawerVisible.value = true;
    if (scope.value === 'current') {
        syncCurrentFileContext();
    }
    await loadHistorySetting();
    await loadHistoryList(true);
};

defineExpose({
    acceptParams,
});

useWindowSize();

onBeforeUnmount(() => {
    disposeDiffEditor();
});
</script>

<style lang="scss" scoped>
.history-workspace {
    display: flex;
    flex-direction: column;
    gap: 16px;
}

.history-list-card,
.history-preview-card {
    min-width: 0;
}

.history-table-box {
    min-height: 0;
}

.history-table-wrap {
    flex: 1;
    min-width: 0;
    overflow-x: auto;
    max-height: clamp(240px, 36vh, 420px);
    overflow-y: auto;
}

.history-diff {
    width: 100%;
    height: clamp(360px, 56vh, 760px);
    min-height: 360px;
}

:deep(.history-table .el-table__body tr) {
    cursor: pointer;
}

:deep(.el-collapse-item__header) {
    border: none;
}

:deep(.el-collapse-item__wrap) {
    border: none;
}

:deep(.el-collapse-item__content) {
    padding-bottom: 0;
}

:deep(.el-collapse-item__title) {
    font-size: 15px;
    font-weight: normal;
}
.history-setting {
    :deep(.el-card__body) {
        padding-top: 2px;
        padding-bottom: 2px;
    }
}

.history-list-card {
    :deep(.el-card__body) {
        display: flex;
        flex-direction: column;
        min-height: 0;
        padding-top: 2px;
    }
}
</style>
