<template>
    <el-drawer
        v-model="drawerVisible"
        :title="$t('file.aiSearchTitle')"
        :size="aiSearchDrawerSize"
        destroy-on-close
        :close-on-click-modal="!aiSearchLoading"
        :close-on-press-escape="!aiSearchLoading"
        :show-close="!aiSearchLoading"
        :before-close="handleAiSearchDrawerBeforeClose"
    >
        <div
            v-loading="drawerContentBusy"
            element-loading-lock
            class="flex flex-col gap-3 min-h-[50vh] ai-search-drawer-inner"
        >
            <el-alert type="info" :closable="false" show-icon class="shrink-0">
                {{ fileAiStatus === 'Disable' ? $t('file.aiSearchGrepOnlyHint') : $t('file.aiSearchHint') }}
            </el-alert>

            <el-card shadow="never" class="ai-search-section-card">
                <template #header>
                    <span class="text-sm font-medium">{{ $t('file.aiSearchAccountTitle') }}</span>
                </template>
                <el-form label-position="top" class="ai-search-account-form">
                    <el-form-item :label="$t('file.aiSearchAccountEnable')">
                        <el-switch v-model="fileAiStatus" active-value="Enable" inactive-value="Disable" />
                    </el-form-item>
                    <el-form-item v-if="fileAiStatus === 'Enable'" :label="$t('aiTools.agents.account')">
                        <el-select class="w-full" v-model="fileAiAccountId" clearable filterable>
                            <el-option
                                v-for="item in fileAiAgentAccountOptions"
                                :key="item.id"
                                :label="item.name"
                                :value="String(item.id)"
                            >
                                <div class="flex items-center justify-between gap-3">
                                    <span>{{ item.name }}</span>
                                    <div class="flex flex-wrap gap-1">
                                        <el-tag size="small" effect="plain">
                                            {{ item.providerName || item.provider }}
                                        </el-tag>
                                        <el-tag size="small" effect="plain" :type="fileAiVerificationTagType(item)">
                                            {{ fileAiVerificationLabel(item) }}
                                        </el-tag>
                                    </div>
                                </div>
                            </el-option>
                        </el-select>
                        <span class="input-help block">{{ $t('file.aiSearchAccountIntro') }}</span>
                    </el-form-item>
                    <el-form-item>
                        <el-button
                            type="primary"
                            plain
                            :loading="fileAiSettingsSaving"
                            @click="saveFileAiManageSettings"
                        >
                            {{ $t('file.aiSearchSaveAccount') }}
                        </el-button>
                    </el-form-item>
                </el-form>
            </el-card>

            <el-card shadow="never" class="ai-search-section-card">
                <template #header>
                    <span class="text-sm font-medium">{{ $t('file.aiSearchSectionConditions') }}</span>
                </template>
                <div class="ai-search-toolbar flex flex-col gap-4">
                    <div class="flex flex-col gap-2 sm:flex-row sm:items-stretch sm:gap-2">
                        <div class="ai-search-field-pair flex min-w-0 flex-1 items-center gap-2">
                            <span class="ai-search-field-label shrink-0">
                                {{ $t('file.aiSearchContentLabel') }}
                            </span>
                            <el-input
                                v-model="aiSearchQuery"
                                clearable
                                :placeholder="$t('file.aiSearchPlaceholder')"
                                class="ai-search-field-input min-w-0 flex-1"
                                @keyup.enter="submitAiSearch"
                            >
                                <template #suffix>
                                    <el-button-group class="shrink-0">
                                        <el-tooltip :content="$t('file.aiSearchMatchCase')" placement="top">
                                            <el-button
                                                :type="aiSearchMatchCase ? 'primary' : 'default'"
                                                class="!px-2.5 font-semibold"
                                                link
                                                @click="aiSearchMatchCase = !aiSearchMatchCase"
                                            >
                                                Aa
                                            </el-button>
                                        </el-tooltip>
                                        <el-tooltip :content="$t('file.aiSearchWholeWord')" placement="top">
                                            <el-button
                                                :type="aiSearchWholeWord ? 'primary' : 'default'"
                                                :disabled="aiSearchUseRegex"
                                                link
                                                class="!px-2.5 font-mono text-xs"
                                                @click="aiSearchWholeWord = !aiSearchWholeWord"
                                            >
                                                [ab]
                                            </el-button>
                                        </el-tooltip>
                                        <el-tooltip :content="$t('file.aiSearchUseRegex')" placement="top">
                                            <el-button
                                                link
                                                :type="aiSearchUseRegex ? 'primary' : 'default'"
                                                class="!px-2.5 font-mono text-xs"
                                                @click="aiSearchUseRegex = !aiSearchUseRegex"
                                            >
                                                .*
                                            </el-button>
                                        </el-tooltip>
                                    </el-button-group>
                                </template>
                            </el-input>
                        </div>
                        <el-button
                            type="primary"
                            class="w-full shrink-0 sm:w-auto sm:self-center"
                            :loading="aiSearchLoading"
                            @click="submitAiSearch"
                        >
                            {{ $t('file.aiSearchRun') }}
                        </el-button>
                    </div>
                    <div class="grid min-w-0 grid-cols-1 gap-3 lg:grid-cols-2 lg:items-start lg:gap-x-4 lg:gap-y-3">
                        <div class="ai-search-field-pair flex min-w-0 items-center gap-2">
                            <span class="ai-search-field-label shrink-0">
                                {{ $t('file.aiSearchDirLabel') }}
                            </span>
                            <el-input v-model="aiSearchPath" class="ai-search-field-input min-w-0 flex-1 !w-48">
                                <template #prepend>
                                    <el-checkbox v-model="aiSearchContainSub" class="shrink-0 !mr-0">
                                        {{ $t('file.sub') }}
                                    </el-checkbox>
                                </template>
                                <template #append>
                                    <el-button
                                        @click="emit('pick-directory', (aiSearchPath || props.listPath || '').trim())"
                                    >
                                        <el-icon><Folder /></el-icon>
                                    </el-button>
                                </template>
                            </el-input>
                        </div>
                        <div class="ai-search-field-pair flex min-w-0 items-center gap-2">
                            <span class="ai-search-field-label shrink-0">
                                {{ $t('file.aiSearchExtLabel') }}
                            </span>
                            <el-input
                                v-model="aiSearchExtensionsInput"
                                clearable
                                class="ai-search-field-input min-w-0 flex-1"
                                :placeholder="$t('file.aiSearchExtensions')"
                            />
                        </div>
                        <div class="flex min-w-0 flex-col gap-2 sm:flex-row sm:items-center sm:gap-2">
                            <div class="ai-search-field-pair flex min-w-0 items-center gap-2">
                                <span class="ai-search-field-label shrink-0">
                                    {{ $t('file.aiSearchModifiedPreset') }}
                                </span>
                                <el-select v-model="aiSearchTimePreset" class="!w-44 max-w-full shrink-0">
                                    <el-option :label="$t('file.aiSearchTimeAny')" value="any" />
                                    <el-option :label="$t('file.aiSearchTime3h')" value="3h" />
                                    <el-option :label="$t('file.aiSearchTime1d')" value="1d" />
                                    <el-option :label="$t('file.aiSearchTime7d')" value="7d" />
                                    <el-option :label="$t('file.aiSearchTime30d')" value="30d" />
                                    <el-option :label="$t('file.aiSearchTimeCustom')" value="custom" />
                                </el-select>
                            </div>
                            <el-date-picker
                                v-if="aiSearchTimePreset === 'custom'"
                                v-model="aiSearchModifiedRange"
                                type="datetimerange"
                                class="w-full min-w-0 sm:min-w-[16rem] sm:flex-1"
                            />
                        </div>
                        <div class="flex min-w-0 flex-col gap-2 sm:flex-row sm:flex-wrap sm:items-center sm:gap-2">
                            <div class="ai-search-field-pair flex min-w-0 items-center gap-2">
                                <span class="ai-search-field-label shrink-0">
                                    {{ $t('file.aiSearchSizePresetLabel') }}
                                </span>
                                <el-select v-model="aiSearchSizePreset" class="!w-48 max-w-full shrink-0">
                                    <el-option :label="$t('file.aiSearchSizeAny')" value="any" />
                                    <el-option :label="$t('file.aiSearchSize0_10mb')" value="lte10mb" />
                                    <el-option :label="$t('file.aiSearchSize10_100mb')" value="mb10_100" />
                                    <el-option :label="$t('file.aiSearchSize100_1gb')" value="mb100_1g" />
                                    <el-option :label="$t('file.aiSearchSizeGte1gb')" value="gte1gb" />
                                    <el-option :label="$t('file.aiSearchSizeCustom')" value="custom" />
                                </el-select>
                            </div>
                            <div
                                v-if="aiSearchSizePreset === 'custom'"
                                class="flex flex-nowrap items-center gap-2 overflow-x-auto"
                            >
                                <el-input-number
                                    v-model="aiSearchMinSize"
                                    :min="0"
                                    :controls="true"
                                    class="!w-40 shrink-0"
                                    :placeholder="$t('file.aiSearchMinSize')"
                                />
                                <span class="shrink-0 text-xs text-[var(--el-text-color-secondary)]">—</span>
                                <el-input-number
                                    v-model="aiSearchMaxSize"
                                    :min="0"
                                    :controls="true"
                                    class="!w-40 shrink-0"
                                    :placeholder="$t('file.aiSearchMaxSize')"
                                />
                            </div>
                        </div>
                    </div>
                    <el-collapse v-model="aiSearchAdvancedActiveNames">
                        <el-collapse-item name="limits">
                            <template #title>
                                <span class="text-sm font-medium">{{ $t('file.aiSearchSectionAdvanced') }}</span>
                            </template>
                            <div class="flex flex-col gap-3">
                                <div>
                                    <el-button type="primary" link @click="resetAiSearchLimitsToDefaults">
                                        {{ $t('file.aiSearchAdvancedResetDefaults') }}
                                    </el-button>
                                </div>
                                <div class="grid grid-cols-1 items-center gap-x-4 gap-y-3 min-[480px]:grid-cols-2">
                                    <div class="ai-search-field-pair flex min-w-0 items-center gap-2">
                                        <span
                                            class="w-32 shrink-0 text-sm leading-8 text-[var(--el-text-color-secondary)] sm:w-36"
                                        >
                                            {{ $t('file.aiSearchLimitMaxItems') }}
                                        </span>
                                        <el-input-number
                                            v-model="aiSearchMaxItems"
                                            :min="1"
                                            :max="2000"
                                            :controls="true"
                                            class="min-w-0 flex-1 !max-w-full sm:!w-52"
                                        >
                                            <template #suffix>
                                                <span class="pr-1 text-xs text-[var(--el-text-color-secondary)]">
                                                    {{ $t('file.aiSearchLimitUnitEntries') }}
                                                </span>
                                            </template>
                                        </el-input-number>
                                    </div>
                                    <div class="ai-search-field-pair flex min-w-0 items-center gap-2">
                                        <span
                                            class="w-32 shrink-0 text-sm leading-8 text-[var(--el-text-color-secondary)] sm:w-36"
                                        >
                                            {{ $t('file.aiSearchLimitMaxScan') }}
                                        </span>
                                        <el-input-number
                                            v-model="aiSearchMaxScanFiles"
                                            :min="1"
                                            :max="500"
                                            :controls="true"
                                            class="min-w-0 flex-1 !max-w-full sm:!w-52"
                                        >
                                            <template #suffix>
                                                <span class="pr-1 text-xs text-[var(--el-text-color-secondary)]">
                                                    {{ $t('file.aiSearchLimitUnitFiles') }}
                                                </span>
                                            </template>
                                        </el-input-number>
                                    </div>
                                    <div class="ai-search-field-pair flex min-w-0 items-center gap-2">
                                        <span
                                            class="w-32 shrink-0 text-sm leading-8 text-[var(--el-text-color-secondary)] sm:w-36"
                                        >
                                            {{ $t('file.aiSearchLimitMaxBytes') }}
                                        </span>
                                        <el-input-number
                                            v-model="aiSearchMaxFileBytes"
                                            :min="1024"
                                            :max="4194304"
                                            :step="1024"
                                            :controls="true"
                                            class="min-w-0 flex-1 !max-w-full sm:!w-52"
                                        >
                                            <template #suffix>
                                                <span class="pr-1 text-xs text-[var(--el-text-color-secondary)]">
                                                    {{ $t('file.aiSearchLimitUnitBytes') }}
                                                </span>
                                            </template>
                                        </el-input-number>
                                    </div>
                                    <div class="ai-search-field-pair flex min-w-0 items-center gap-2">
                                        <span
                                            class="w-32 shrink-0 text-sm leading-8 text-[var(--el-text-color-secondary)] sm:w-36"
                                        >
                                            {{ $t('file.aiSearchLimitHitsPerFile') }}
                                        </span>
                                        <el-input-number
                                            v-model="aiSearchMaxHitsPerFile"
                                            :min="1"
                                            :max="200"
                                            :controls="true"
                                            class="min-w-0 flex-1 !max-w-full sm:!w-52"
                                        >
                                            <template #suffix>
                                                <span class="pr-1 text-xs text-[var(--el-text-color-secondary)]">
                                                    {{ $t('file.aiSearchLimitUnitLines') }}
                                                </span>
                                            </template>
                                        </el-input-number>
                                    </div>
                                    <div class="ai-search-field-pair flex min-w-0 items-center gap-2">
                                        <span
                                            class="w-32 shrink-0 text-sm leading-8 text-[var(--el-text-color-secondary)] sm:w-36"
                                        >
                                            {{ $t('file.aiSearchLimitTotalHits') }}
                                        </span>
                                        <el-input-number
                                            v-model="aiSearchMaxTotalHits"
                                            :min="1"
                                            :max="2000"
                                            :controls="true"
                                            class="min-w-0 flex-1 !max-w-full sm:!w-52"
                                        >
                                            <template #suffix>
                                                <span class="pr-1 text-xs text-[var(--el-text-color-secondary)]">
                                                    {{ $t('file.aiSearchLimitUnitLines') }}
                                                </span>
                                            </template>
                                        </el-input-number>
                                    </div>
                                </div>
                            </div>
                        </el-collapse-item>
                    </el-collapse>
                </div>
            </el-card>

            <el-card v-if="aiSearchResult" shadow="never" class="ai-search-section-card">
                <template #header>
                    <span class="text-sm font-medium">{{ $t('file.aiSearchResultsSection') }}</span>
                </template>
                <div class="flex flex-col gap-3">
                    <el-alert v-if="aiSearchResult.truncated" type="warning" :closable="false" show-icon>
                        {{ $t('file.aiSearchTruncated') }}
                    </el-alert>
                    <el-alert v-if="aiSearchResult.preFiltered" type="info" :closable="false" show-icon>
                        {{ $t('file.aiSearchPreFiltered') }}
                    </el-alert>

                    <el-text v-if="aiSearchShowAiMeta" size="small" type="info" class="!block">
                        {{
                            $t('file.aiSearchMeta', [
                                aiSearchResult.itemCount,
                                aiSearchResult.duration,
                                aiSearchResult.totalTokens,
                            ])
                        }}
                    </el-text>
                    <el-text v-else-if="aiSearchShowGrepMeta" size="small" type="info" class="!block">
                        {{
                            $t('file.aiSearchGrepMetaContent', [
                                aiSearchResult.itemCount,
                                aiSearchResult.contentScannedFiles ?? 0,
                                aiSearchResult.duration,
                            ])
                        }}
                    </el-text>

                    <div
                        v-if="aiSearchShowSummaryPanel"
                        class="ai-search-summary-box rounded-md border border-[var(--el-border-color)] bg-[var(--el-fill-color-blank)] p-3"
                    >
                        <MkdownEditor :content="aiSearchResult.summary" />
                    </div>

                    <template v-if="aiSearchHasHitPanel">
                        <div class="text-sm font-medium text-[var(--el-text-color-primary)]">
                            {{ $t('file.aiSearchContentHits') }}
                        </div>
                        <el-text
                            v-if="(aiSearchResult.contentScannedFiles ?? 0) > 0 && aiSearchResult.mode !== 'grep'"
                            size="small"
                            type="info"
                            class="!block"
                        >
                            {{ $t('file.aiSearchContentScanMeta', [aiSearchResult.contentScannedFiles]) }}
                        </el-text>
                        <el-alert v-if="aiSearchResult.contentHitsTruncated" type="warning" :closable="false" show-icon>
                            {{ $t('file.aiSearchHitsTruncated') }}
                        </el-alert>
                        <el-alert
                            v-if="!aiSearchResult.hits?.length && (aiSearchResult.contentScannedFiles ?? 0) > 0"
                            type="info"
                            :closable="false"
                            show-icon
                        >
                            {{ $t('file.aiSearchNoLineHits') }}
                        </el-alert>
                        <el-row v-if="aiSearchResult.hits?.length" :gutter="12">
                            <el-col :xs="24" :sm="10" class="mb-2 sm:mb-0">
                                <div
                                    class="ai-search-hit-col flex flex-col rounded-md border border-[var(--el-border-color)] p-3"
                                >
                                    <div class="mb-2 shrink-0 text-sm font-medium">
                                        {{ $t('file.aiSearchHitFiles') }}
                                    </div>
                                    <el-scrollbar max-height="200" class="min-h-0 flex-1">
                                        <div
                                            v-for="p in aiSearchHitPaths"
                                            :key="p"
                                            class="cursor-pointer truncate rounded px-2 py-1.5 text-sm transition-colors"
                                            :title="p"
                                            :class="hitFileRowClass(p)"
                                            @click="aiSearchSelectedHitPath = p"
                                        >
                                            {{ p.split('/').pop() || p }}
                                        </div>
                                    </el-scrollbar>
                                    <el-button
                                        type="primary"
                                        link
                                        class="mt-2 mb-1 shrink-0"
                                        :disabled="!aiSearchSelectedHitPath"
                                        @click="
                                            emit('open-editor', {
                                                path: aiSearchSelectedHitPath,
                                                initialLine:
                                                    aiSearchHitsForSelected[0]?.line &&
                                                    aiSearchHitsForSelected[0].line > 0
                                                        ? aiSearchHitsForSelected[0].line
                                                        : undefined,
                                            })
                                        "
                                    >
                                        {{ $t('file.aiSearchOpenFile') }}
                                    </el-button>
                                </div>
                            </el-col>
                            <el-col :xs="24" :sm="14">
                                <div
                                    class="ai-search-hit-col flex flex-col rounded-md border border-[var(--el-border-color)] p-3"
                                >
                                    <div class="mb-2 shrink-0 text-sm font-medium">
                                        {{ $t('file.aiSearchHitLines') }}
                                    </div>
                                    <el-scrollbar max-height="220" class="min-h-0 flex-1">
                                        <div
                                            v-for="(h, idx) in aiSearchHitsForSelected"
                                            :key="idx"
                                            class="flex flex-col gap-1 border-0 border-b border-solid border-[var(--el-border-color-lighter)] py-2 text-xs sm:flex-row sm:items-start sm:justify-between"
                                        >
                                            <div class="min-w-0 flex-1">
                                                <span
                                                    class="mr-2 shrink-0 tabular-nums text-[var(--el-text-color-secondary)]"
                                                >
                                                    {{ h.line > 0 ? h.line : '—' }}
                                                </span>
                                                <span class="break-all font-mono">{{ h.text }}</span>
                                            </div>
                                            <el-button
                                                type="primary"
                                                link
                                                class="shrink-0 !p-0 sm:!ml-2"
                                                @click="
                                                    emit('open-editor', {
                                                        path: h.path,
                                                        initialLine: h.line > 0 ? h.line : undefined,
                                                    })
                                                "
                                            >
                                                {{
                                                    h.line > 0
                                                        ? $t('file.aiSearchOpenAtLine')
                                                        : $t('file.aiSearchOpenFile')
                                                }}
                                            </el-button>
                                        </div>
                                    </el-scrollbar>
                                </div>
                            </el-col>
                        </el-row>
                    </template>
                </div>
            </el-card>
        </div>
    </el-drawer>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue';
import { Folder } from '@element-plus/icons-vue';
import { fileAiSearch } from '@/api/modules/files';
import { getAgentFileManageAIInfo, updateAgentFileManageAIInfo } from '@/api/modules/setting';
import { pageAgentAccounts } from '@/api/modules/ai';
import { File } from '@/api/interface/file';
import { useGlobalStore } from '@/composables/useGlobalStore';
import i18n from '@/lang';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { isAgentAccountVerificationSkipped } from '@/utils/agent';
import MkdownEditor from '@/components/mkdown-editor/index.vue';

const AI_SEARCH_DEFAULT_MAX_ITEMS = 500;
const AI_SEARCH_DEFAULT_MAX_SCAN_FILES = 150;
const AI_SEARCH_DEFAULT_MAX_FILE_BYTES = 512 * 1024;
const AI_SEARCH_DEFAULT_MAX_HITS_PER_FILE = 40;
const AI_SEARCH_DEFAULT_MAX_TOTAL_HITS = 500;

const AI_SEARCH_DRAWER_MOBILE_MQ = '(max-width: 639px)';
const aiSearchDrawerSize = ref('55%');
let aiSearchDrawerMq: MediaQueryList | null = null;

function syncAiSearchDrawerSize(mq: MediaQueryList | MediaQueryListEvent) {
    aiSearchDrawerSize.value = mq.matches ? '100%' : '55%';
}

const props = defineProps<{
    modelValue: boolean;
    listPath: string;
}>();

const { language } = useGlobalStore();

const emit = defineEmits<{
    (e: 'update:modelValue', v: boolean): void;
    (e: 'pick-directory', path: string): void;
    (e: 'open-editor', payload: { path: string; initialLine?: number }): void;
}>();

const drawerVisible = computed({
    get: () => props.modelValue,
    set: (v: boolean) => emit('update:modelValue', v),
});

const aiSearchQuery = ref('');
const aiSearchPath = ref('');
const aiSearchContainSub = ref(true);
const aiSearchTimePreset = ref<'any' | '3h' | '1d' | '7d' | '30d' | 'custom'>('any');
const aiSearchSizePreset = ref<'any' | 'lte10mb' | 'mb10_100' | 'mb100_1g' | 'gte1gb' | 'custom'>('any');
const aiSearchMatchCase = ref(false);
const aiSearchWholeWord = ref(false);
const aiSearchUseRegex = ref(false);
const aiSearchExtensionsInput = ref('');
const aiSearchMinSize = ref<number | undefined>(undefined);
const aiSearchMaxSize = ref<number | undefined>(undefined);
const aiSearchModifiedRange = ref<[Date, Date] | null>(null);
const aiSearchMaxItems = ref(AI_SEARCH_DEFAULT_MAX_ITEMS);
const aiSearchMaxScanFiles = ref(AI_SEARCH_DEFAULT_MAX_SCAN_FILES);
const aiSearchMaxFileBytes = ref(AI_SEARCH_DEFAULT_MAX_FILE_BYTES);
const aiSearchMaxHitsPerFile = ref(AI_SEARCH_DEFAULT_MAX_HITS_PER_FILE);
const aiSearchMaxTotalHits = ref(AI_SEARCH_DEFAULT_MAX_TOTAL_HITS);
const aiSearchAdvancedActiveNames = ref<string[]>([]);
const aiSearchLoading = ref(false);
const aiSearchResult = ref<File.FileAISearchResult | null>(null);
const aiSearchSelectedHitPath = ref('');

function hitFileRowClass(p: string) {
    const selected = p === aiSearchSelectedHitPath.value;
    return {
        'font-semibold text-[var(--el-color-primary)] bg-[var(--el-fill-color-light)]': selected,
        'hover:bg-[var(--el-fill-color-light)]': !selected,
    };
}

const aiSearchHitPaths = computed(() => {
    const r = aiSearchResult.value;
    if (!r?.hits?.length) {
        return [] as string[];
    }
    const set = new Set<string>();
    for (const h of r.hits) {
        if (h.path) {
            set.add(h.path);
        }
    }
    return Array.from(set).sort();
});

const aiSearchHitsForSelected = computed(() => {
    const r = aiSearchResult.value;
    const p = aiSearchSelectedHitPath.value;
    if (!r?.hits?.length || !p) {
        return [] as File.FileAIContentHit[];
    }
    return r.hits.filter((h) => h.path === p);
});

const aiSearchHasHitPanel = computed(() => {
    const r = aiSearchResult.value;
    if (!r) {
        return false;
    }
    return (r.hits?.length ?? 0) > 0 || (r.contentScannedFiles ?? 0) > 0 || !!r.contentHitsTruncated;
});

const aiSearchShowGrepMeta = computed(() => {
    const r = aiSearchResult.value;
    if (!r) {
        return false;
    }
    return r.mode === 'grep';
});

const aiSearchShowAiMeta = computed(() => {
    const r = aiSearchResult.value;
    if (!r) {
        return false;
    }
    const sum = (r.summary && String(r.summary).trim()) || '';
    if (!sum || r.mode === 'grep') {
        return false;
    }
    return r.mode === 'ai' || r.mode == null || r.mode === '';
});

const aiSearchShowSummaryPanel = computed(() => {
    const r = aiSearchResult.value;
    if (!r) {
        return false;
    }
    if (r.mode === 'grep') {
        return false;
    }
    return !!(r.summary && String(r.summary).trim());
});

watch(
    () => aiSearchResult.value,
    (r) => {
        if (!r?.hits?.length) {
            aiSearchSelectedHitPath.value = '';
            return;
        }
        const paths = [...new Set(r.hits.map((h) => h.path).filter(Boolean))].sort();
        aiSearchSelectedHitPath.value = paths[0] || '';
    },
);

watch(
    () => aiSearchUseRegex.value,
    (on) => {
        if (on) {
            aiSearchWholeWord.value = false;
        }
    },
);

const fileAiDrawerLoading = ref(false);
const fileAiSettingsSaving = ref(false);
const fileAiStatus = ref<'Enable' | 'Disable'>('Disable');
const fileAiAccountId = ref('');
const fileAiAgentAccountOptions = ref<
    Array<{ id: number | string; name: string; provider?: string; providerName?: string; verified?: boolean }>
>([]);

const loadFileAiManageSettings = async () => {
    const res = await getAgentFileManageAIInfo();
    fileAiStatus.value = (res.data.aiStatus === 'Enable' ? 'Enable' : 'Disable') as 'Enable' | 'Disable';
    fileAiAccountId.value = res.data.aiAccountId || '';
};

const loadFileAiAgentAccounts = async () => {
    const res = await pageAgentAccounts({ page: 1, pageSize: 1000, provider: '', textOnly: true, name: '' });
    fileAiAgentAccountOptions.value = res.data?.items || [];
};

const fileAiVerificationLabel = (item: { provider?: string; verified?: boolean }) => {
    if (isAgentAccountVerificationSkipped(item.provider)) {
        return i18n.global.t('aiTools.agents.verifySkipped');
    }
    return item.verified ? 'OK' : 'N/A';
};

const fileAiVerificationTagType = (item: { provider?: string; verified?: boolean }) => {
    if (isAgentAccountVerificationSkipped(item.provider)) {
        return 'info';
    }
    return item.verified ? 'success' : 'warning';
};

const saveFileAiManageSettings = async () => {
    if (fileAiStatus.value === 'Enable' && !String(fileAiAccountId.value).trim()) {
        MsgWarning(i18n.global.t('commons.rule.requiredSelect'));
        return;
    }
    fileAiSettingsSaving.value = true;
    try {
        await updateAgentFileManageAIInfo({
            aiStatus: fileAiStatus.value,
            aiAccountId: fileAiStatus.value === 'Enable' ? fileAiAccountId.value : '',
        });
        if (fileAiStatus.value === 'Disable') {
            fileAiAccountId.value = '';
        }
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    } finally {
        fileAiSettingsSaving.value = false;
    }
};

watch(
    () => props.modelValue,
    (v) => {
        if (!v) {
            return;
        }
        aiSearchQuery.value = '';
        aiSearchResult.value = null;
        aiSearchPath.value = (props.listPath || '').trim();
        fileAiDrawerLoading.value = true;
        Promise.all([loadFileAiManageSettings(), loadFileAiAgentAccounts()]).finally(() => {
            fileAiDrawerLoading.value = false;
        });
    },
);

onMounted(() => {
    aiSearchDrawerMq = window.matchMedia(AI_SEARCH_DRAWER_MOBILE_MQ);
    syncAiSearchDrawerSize(aiSearchDrawerMq);
    aiSearchDrawerMq.addEventListener('change', syncAiSearchDrawerSize);
});

onUnmounted(() => {
    aiSearchDrawerMq?.removeEventListener('change', syncAiSearchDrawerSize);
});

const drawerContentBusy = computed(() => fileAiDrawerLoading.value || aiSearchLoading.value);

const handleAiSearchDrawerBeforeClose = (done: () => void) => {
    if (aiSearchLoading.value) {
        return;
    }
    done();
};

const applyAiSearchTimeToPayload = (payload: File.FileAISearchReq) => {
    const p = aiSearchTimePreset.value;
    if (p === 'custom') {
        const mr = aiSearchModifiedRange.value;
        if (mr?.[0] instanceof Date) {
            payload.modifiedAfter = mr[0].toISOString();
        }
        if (mr?.[1] instanceof Date) {
            payload.modifiedBefore = mr[1].toISOString();
        }
        return;
    }
    if (p === 'any') {
        return;
    }
    const now = Date.now();
    const delta: Record<string, number> = {
        '3h': 3 * 3600000,
        '1d': 86400000,
        '7d': 7 * 86400000,
        '30d': 30 * 86400000,
    };
    const ms = delta[p];
    if (ms) {
        payload.modifiedAfter = new Date(now - ms).toISOString();
        payload.modifiedBefore = new Date(now).toISOString();
    }
};

const applyAiSearchSizeToPayload = (payload: File.FileAISearchReq) => {
    const p = aiSearchSizePreset.value;
    const MB = 1024 * 1024;
    const GB = 1024 * MB;
    if (p === 'custom') {
        if (aiSearchMinSize.value != null && aiSearchMinSize.value > 0) {
            payload.minSize = aiSearchMinSize.value;
        }
        if (aiSearchMaxSize.value != null && aiSearchMaxSize.value > 0) {
            payload.maxSize = aiSearchMaxSize.value;
        }
        return;
    }
    if (p === 'any') {
        return;
    }
    if (p === 'lte10mb') {
        payload.maxSize = 10 * MB;
    } else if (p === 'mb10_100') {
        payload.minSize = 10 * MB;
        payload.maxSize = 100 * MB;
    } else if (p === 'mb100_1g') {
        payload.minSize = 100 * MB;
        payload.maxSize = 1024 * MB;
    } else if (p === 'gte1gb') {
        payload.minSize = GB;
    }
};

const submitAiSearch = async () => {
    const q = aiSearchQuery.value.trim();
    if (!q) {
        MsgWarning(i18n.global.t('file.aiSearchQueryRequired'));
        return;
    }
    aiSearchLoading.value = true;
    aiSearchResult.value = null;
    try {
        const ext = aiSearchExtensionsInput.value
            .split(/[,;\s]+/)
            .map((s) => s.trim())
            .filter(Boolean);
        const rootPath = (aiSearchPath.value || props.listPath || '').trim() || props.listPath;
        const payload: File.FileAISearchReq = {
            path: rootPath,
            query: q,
            responseLanguage: language.value,
            containSub: aiSearchContainSub.value,
            matchCase: aiSearchMatchCase.value,
            wholeWord: aiSearchWholeWord.value,
            useRegex: aiSearchUseRegex.value,
        };
        if (ext.length) {
            payload.extensions = ext;
        }
        applyAiSearchSizeToPayload(payload);
        applyAiSearchTimeToPayload(payload);
        const limItems = Number(aiSearchMaxItems.value) || AI_SEARCH_DEFAULT_MAX_ITEMS;
        const limTotal = Number(aiSearchMaxTotalHits.value) || AI_SEARCH_DEFAULT_MAX_TOTAL_HITS;
        if (limItems > 0) {
            payload.maxItems = limItems;
        }
        const limScan = Number(aiSearchMaxScanFiles.value) || AI_SEARCH_DEFAULT_MAX_SCAN_FILES;
        const limBytes = Number(aiSearchMaxFileBytes.value) || AI_SEARCH_DEFAULT_MAX_FILE_BYTES;
        const limPer = Number(aiSearchMaxHitsPerFile.value) || AI_SEARCH_DEFAULT_MAX_HITS_PER_FILE;
        if (limScan > 0) {
            payload.maxScanFiles = limScan;
        }
        if (limBytes > 0) {
            payload.maxFileBytes = limBytes;
        }
        if (limPer > 0) {
            payload.maxHitsPerFile = limPer;
        }
        if (limTotal > 0) {
            payload.maxTotalHits = limTotal;
        }
        const res = await fileAiSearch(payload);
        aiSearchResult.value = res.data;
    } finally {
        aiSearchLoading.value = false;
    }
};

function applyPathFromPicker(path: string | string[]) {
    if (Array.isArray(path)) {
        aiSearchPath.value = (path[0] || '').trim();
        return;
    }
    aiSearchPath.value = (path || '').trim();
}

function resetAiSearchLimitsToDefaults() {
    aiSearchMaxItems.value = AI_SEARCH_DEFAULT_MAX_ITEMS;
    aiSearchMaxScanFiles.value = AI_SEARCH_DEFAULT_MAX_SCAN_FILES;
    aiSearchMaxFileBytes.value = AI_SEARCH_DEFAULT_MAX_FILE_BYTES;
    aiSearchMaxHitsPerFile.value = AI_SEARCH_DEFAULT_MAX_HITS_PER_FILE;
    aiSearchMaxTotalHits.value = AI_SEARCH_DEFAULT_MAX_TOTAL_HITS;
}

defineExpose({ applyPathFromPicker });
</script>

<style scoped>
.ai-search-drawer-inner :deep(.el-card__header) {
    padding: 12px 16px;
    border-bottom: 1px solid var(--el-border-color-lighter);
    background-color: var(--el-fill-color-light);
}

.ai-search-section-card {
    border-color: var(--el-border-color-lighter);
}

.ai-search-section-card :deep(.el-card__body) {
    padding: 16px;
}

.ai-search-summary-box {
    max-height: 40vh;
    overflow: auto;
}

.ai-search-hit-col {
    min-height: 240px;
    max-height: min(40vh, 20rem);
}

.ai-search-field-pair {
    flex-wrap: nowrap;
    overflow-x: auto;
}

.ai-search-field-label {
    width: 4.5rem;
    flex-shrink: 0;
    font-size: 0.875rem;
    line-height: 2rem;
    color: var(--el-text-color-secondary);
}

@media (min-width: 640px) {
    .ai-search-field-label {
        width: 6rem;
    }
}

.ai-search-field-input {
    min-width: 0;
}

.ai-search-advanced-collapse :deep(.el-collapse-item__header) {
    padding: 10px 12px;
    border-radius: var(--el-border-radius-base);
    background-color: var(--el-fill-color-light);
    border: 1px solid var(--el-border-color-lighter);
    font-weight: 500;
}

.ai-search-advanced-collapse :deep(.el-collapse-item__wrap) {
    border: none;
}

.ai-search-advanced-collapse :deep(.el-collapse-item__content) {
    padding-bottom: 0;
}
</style>
