<template>
    <div>
        <LayoutContent v-if="currentTab !== 'recipes'" :title="currentTitle">
            <template #search>
                <el-button
                    v-for="button in buttons"
                    :key="button.value"
                    class="tag-button"
                    :class="{ 'no-active': currentTab !== button.value }"
                    :type="currentTab === button.value ? 'primary' : undefined"
                    :aria-pressed="currentTab === button.value"
                    @click="currentTab = button.value"
                >
                    {{ button.label }}
                </el-button>
            </template>
            <template #leftToolBar>
                <el-dropdown
                    v-if="currentTab === 'featured' && !isOffline"
                    :disabled="syncingFeatured"
                    @command="handleSyncOp"
                >
                    <el-button v-permission :loading="syncingFeatured">
                        {{ $t('commons.button.sync') }}
                        <el-icon><arrow-down /></el-icon>
                    </el-button>
                    <template #dropdown>
                        <el-dropdown-menu>
                            <fu-dropdown-item v-permission command="sync">
                                {{ $t('cronjob.library.syncNow') }}
                            </fu-dropdown-item>
                            <fu-dropdown-item v-if="autoSync === 'Enable'" v-permission command="Disable">
                                {{ $t('cronjob.library.turnOffSync') }}
                            </fu-dropdown-item>
                            <fu-dropdown-item v-if="autoSync === 'Disable'" v-permission command="Enable">
                                {{ $t('cronjob.library.turnOnSync') }}
                            </fu-dropdown-item>
                        </el-dropdown-menu>
                    </template>
                </el-dropdown>
                <el-button v-if="currentTab === 'downloader'" type="primary" @click="searchOpen = true">
                    {{ $t('commons.button.search') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <div v-if="currentTab === 'featured'" class="featured-filters">
                    <TableViewSwitch v-model="featuredViewMode" storage-key="featured-models" />
                    <el-select
                        v-model="featuredFamily"
                        clearable
                        placeholder=""
                        class="model-filter"
                        :aria-label="$t('aiTools.model.modelCategory')"
                        @change="handleFeaturedSearch"
                    >
                        <template #prefix>{{ $t('aiTools.model.modelCategory') }}</template>
                        <el-option v-for="family in featuredFamilies" :key="family" :label="family" :value="family">
                            <ProviderLabel :provider="familyProviders[family] || family" :display-name="family" />
                        </el-option>
                    </el-select>
                    <el-select
                        v-model="featuredWeightProvider"
                        clearable
                        filterable
                        placeholder=""
                        class="model-filter"
                        :aria-label="$t('aiTools.model.weightProvider')"
                        @change="handleFeaturedSearch"
                    >
                        <template #prefix>{{ $t('aiTools.model.weightProvider') }}</template>
                        <el-option
                            v-for="provider in featuredWeightProviders"
                            :key="provider"
                            :label="provider"
                            :value="provider"
                        />
                    </el-select>
                    <TableSearch
                        v-model:searchName="featuredSearchName"
                        class="model-search"
                        @search="handleFeaturedSearch"
                    />
                    <TableRefresh @search="refreshCurrent" />
                    <fu-table-column-select
                        v-if="featuredViewMode === 'table'"
                        :columns="featuredColumns"
                        trigger="hover"
                        :title="$t('commons.table.selectColumn')"
                        :only-icon="true"
                    />
                </div>
                <div v-else class="flex items-center gap-3">
                    <el-select
                        v-if="currentTab === 'downloader'"
                        v-model="downloadTypeFilter"
                        class="p-w-200"
                        clearable
                        :placeholder="$t('commons.table.all')"
                        :aria-label="$t('commons.table.type')"
                        @change="changeDownloadType"
                    >
                        <template #prefix>{{ $t('commons.table.type') }}</template>
                        <el-option value="all" :label="$t('aiTools.model.model')" />
                        <el-option value="tokenizer" :label="$t('aiTools.benchmark.tokenizer')" />
                    </el-select>
                    <TableRefresh @search="refreshCurrent" />
                    <TableSetting
                        v-if="currentTab === 'downloader'"
                        title="model-download-refresh"
                        :rate="5"
                        @search="loadDownloads"
                    />
                </div>
            </template>
            <template #main>
                <ComplexTable
                    v-if="currentTab === 'featured'"
                    v-model:view-mode="featuredViewMode"
                    v-loading="featuredLoading"
                    :data="featuredModels"
                    :pagination-config="featuredPagination"
                    :height-diff="330"
                    :columns="featuredColumns"
                    local-key="featuredModelColumn"
                    scrollbar-always-on
                    @search="loadFeaturedModels"
                    row-key="repoID"
                    class="featured-models"
                >
                    <el-table-column
                        v-for="field in featuredFields"
                        :key="field.prop"
                        :prop="field.prop"
                        :label="$t('aiTools.model.' + field.label)"
                        :min-width="field.width"
                        :card-type="
                            field.prop === 'repoID' ? 'name' : field.prop === 'weightProvider' ? 'status' : 'content'
                        "
                    >
                        <template #default="{ row }">
                            <template v-if="field.prop === 'repoID'">
                                <span v-if="featuredViewMode === 'card'" class="featured-name-label">
                                    {{ $t('aiTools.model.modelName') }}
                                </span>
                                <el-button
                                    link
                                    type="primary"
                                    class="featured-name"
                                    :title="row.repoID"
                                    @click="openFeaturedModel(row)"
                                >
                                    {{ row.repoID?.slice(row.repoID.indexOf('/') + 1) || '-' }}
                                </el-button>
                                <div v-if="featuredViewMode === 'card' && row.description" class="featured-description">
                                    {{ row.description[language.toLowerCase()] || row.description.en }}
                                </div>
                            </template>
                            <el-tag
                                v-else-if="
                                    field.prop === 'weightProvider' && featuredViewMode === 'card' && row.weightProvider
                                "
                                class="featured-provider"
                                type="info"
                                effect="plain"
                                size="small"
                                :title="$t('aiTools.model.weightProvider') + ': ' + row.weightProvider"
                            >
                                {{ row.weightProvider }}
                            </el-tag>
                            <ProviderLabel
                                v-else-if="field.prop === 'family'"
                                :provider="familyProviders[row.family]"
                                :display-name="row.family"
                            />
                            <template v-else-if="field.prop === 'weightGB'">
                                {{ $t('aiTools.model.approxWeightSize', [row.weightGB]) }}
                            </template>
                            <template v-else-if="field.prop === 'activeParams' && row.activeParams === 'undisclosed'">
                                {{ $t('aiTools.model.undisclosed') }}
                            </template>
                            <template v-else>{{ row[field.prop] }}</template>
                        </template>
                    </el-table-column>
                    <el-table-column
                        v-if="featuredViewMode === 'table'"
                        prop="description"
                        :label="$t('commons.table.description')"
                        min-width="420"
                    >
                        <template #default="{ row }">
                            {{ row.description?.[language.toLowerCase()] || row.description?.en || '-' }}
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.operate')" width="160" fixed="right" card-type="button">
                        <template #default="{ row }">
                            <el-button link type="primary" @click="openFeaturedModel(row)">
                                {{ $t('commons.button.view') }}
                            </el-button>
                            <el-button
                                v-permission
                                link
                                type="primary"
                                @click="downloadRef?.acceptParams(row.repoID, 'all', row.sources, row.sourceModelIds)"
                            >
                                {{ $t('commons.button.download') }}
                            </el-button>
                        </template>
                    </el-table-column>
                </ComplexTable>

                <ComplexTable
                    v-if="currentTab === 'downloader'"
                    :data="downloads"
                    :pagination-config="downloadPagination"
                    @search="loadDownloads"
                >
                    <el-table-column prop="source" :label="$t('aiTools.model.modelSource')" width="120">
                        <template #default="{ row }">
                            {{ sourceLabel(row.source) }}
                        </template>
                    </el-table-column>
                    <el-table-column
                        prop="repoID"
                        :label="$t('aiTools.model.repoID')"
                        min-width="240"
                        show-overflow-tooltip
                    />
                    <el-table-column :label="$t('modelDownloadExtra.content')" width="140">
                        <template #default="{ row }">
                            {{
                                $t(
                                    row.downloadType === 'tokenizer'
                                        ? 'aiTools.benchmark.tokenizer'
                                        : 'aiTools.model.model',
                                )
                            }}
                        </template>
                    </el-table-column>
                    <el-table-column prop="status" :label="$t('commons.table.status')" width="120">
                        <template #default="{ row }">
                            <el-tooltip :content="row.errorMessage" :disabled="!row.errorMessage" placement="top">
                                <el-tag :type="statusTagType(row.status)" round effect="light" size="small">
                                    <span class="status-tag-content">
                                        <span>{{ statusText(row.status) }}</span>
                                        <el-icon v-if="row.status === 'Downloading'" class="is-loading">
                                            <Loading />
                                        </el-icon>
                                    </span>
                                </el-tag>
                            </el-tooltip>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('aiTools.model.progress')" min-width="220">
                        <template #default="{ row }">
                            <div class="progress-cell">
                                <el-progress
                                    class="progress-bar"
                                    :percentage="formatProgress(row.progress)"
                                    :stroke-width="6"
                                />
                                <span
                                    v-if="formatSizePair(row.downloadedSize, row.totalSize)"
                                    class="progress-size"
                                    :title="formatSizePair(row.downloadedSize, row.totalSize)"
                                >
                                    {{ formatSizePair(row.downloadedSize, row.totalSize) }}
                                </span>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column
                        prop="targetDir"
                        :label="$t('aiTools.model.modelDir')"
                        min-width="260"
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <el-button
                                v-permission:view="'host_file_view'"
                                link
                                type="primary"
                                class="model-path-link"
                                :title="row.targetDir"
                                @click="openLocalModelPath(row.targetDir)"
                            >
                                {{ row.targetDir }}
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('file.size')" width="120">
                        <template #default="{ row }">{{ row.localModel?.sizeFormatted || '-' }}</template>
                    </el-table-column>
                    <el-table-column prop="createdAt" :label="$t('commons.table.date')" width="180" />
                    <el-table-column :label="$t('commons.table.operate')" width="240" fixed="right">
                        <template #default="{ row }">
                            <el-button v-permission v-if="row.canCancel" link type="primary" @click="cancelTask(row)">
                                {{ $t('commons.button.cancel') }}
                            </el-button>
                            <el-button v-permission v-if="row.canRetry" link type="primary" @click="retryTask(row)">
                                {{ $t('commons.button.retry') }}
                            </el-button>
                            <el-button
                                v-if="row.downloadType === 'all' && row.status === 'Success' && row.localModel"
                                v-permission
                                link
                                type="primary"
                                @click="emit('deploy', row.localModel.path)"
                            >
                                {{ $t('aiTools.model.deploy') }}
                            </el-button>
                            <el-button
                                v-if="
                                    recipesView &&
                                    isProductPro &&
                                    row.downloadType === 'all' &&
                                    row.status === 'Success' &&
                                    row.localModel
                                "
                                v-permission
                                link
                                type="primary"
                                @click="
                                    recipeModelPath = row.localModel.path;
                                    currentTab = 'recipes';
                                "
                            >
                                {{ $t('aiTools.model.createRecipe') }}
                            </el-button>
                            <el-button
                                v-permission
                                v-if="row.canRemoveRecord || (row.localModel && !row.canCancel)"
                                link
                                type="primary"
                                @click="downloadDeleteRef.acceptParams(row.id ? row : row.localModel)"
                            >
                                {{ $t('commons.button.delete') }}
                            </el-button>
                        </template>
                    </el-table-column>
                </ComplexTable>

                <el-form
                    v-if="currentTab === 'settings'"
                    ref="settingFormRef"
                    :model="settings"
                    :rules="settingRules"
                    :label-position="isMobile ? 'top' : 'left'"
                    label-width="200px"
                >
                    <el-row>
                        <el-col :span="1"><br /></el-col>
                        <el-col :xs="24" :sm="20" :md="15" :lg="12" :xl="12">
                            <el-form-item :label="$t('aiTools.model.modelDir')" prop="modelDir">
                                <el-input v-model="settings.modelDir" />
                            </el-form-item>
                            <el-form-item
                                :label="$t('aiTools.model.acceleratedEndpointLabel', ['HuggingFace'])"
                                prop="hfEndpoint"
                            >
                                <el-input v-model="settings.hfEndpoint" placeholder="https://hf-mirror.com" />
                                <span class="input-help">{{ $t('aiTools.model.hfEndpointHelper') }}</span>
                            </el-form-item>
                            <el-form-item :label="$t('aiTools.model.tokenLabel', ['HuggingFace'])" prop="hfToken">
                                <el-input v-model="settings.hfToken" show-password />
                            </el-form-item>
                            <el-form-item
                                :label="$t('aiTools.model.endpointLabel', ['ModelScope'])"
                                prop="modelScopeEndpoint"
                            >
                                <el-input
                                    v-model="settings.modelScopeEndpoint"
                                    placeholder="https://www.modelscope.cn"
                                />
                            </el-form-item>
                            <el-form-item
                                :label="$t('aiTools.model.tokenLabel', ['ModelScope'])"
                                prop="modelScopeToken"
                            >
                                <el-input v-model="settings.modelScopeToken" show-password />
                            </el-form-item>
                            <el-form-item>
                                <el-button
                                    v-permission
                                    type="primary"
                                    :loading="savingSettings"
                                    :disabled="savingSettings"
                                    @click="saveSettings"
                                >
                                    {{ $t('commons.button.save') }}
                                </el-button>
                            </el-form-item>
                        </el-col>
                    </el-row>
                </el-form>
            </template>
        </LayoutContent>

        <DrawerPro v-model="searchOpen" :header="$t('commons.button.search')" size="50%" :full-screen="true">
            <div class="source-switch">
                <SourceSelect
                    v-model="currentSource"
                    :sources="['modelscope', 'huggingface']"
                    @update:model-value="handleSourceChange"
                />
            </div>
            <div v-loading="searchCardLoading">
                <div class="search-row">
                    <TableSearch
                        v-model:searchName="searchForm.query"
                        class="search-input"
                        placeholder="Qwen"
                        @search="handleSearchSubmit"
                    />
                    <el-select v-model="searchForm.sort" class="sort-select" @change="handleSortChange">
                        <el-option
                            v-for="item in sortOptions"
                            :key="item.value"
                            :label="item.label"
                            :value="item.value"
                        />
                    </el-select>
                </div>
                <ComplexTable
                    :data="models"
                    :pagination-config="searchPagination"
                    class="mt-4"
                    @search="searchModels()"
                >
                    <el-table-column
                        prop="repoID"
                        :label="$t('aiTools.model.repoID')"
                        min-width="260"
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <el-button
                                link
                                type="primary"
                                @click="
                                    toLink(
                                        currentSource === 'modelscope'
                                            ? `https://www.modelscope.cn/models/${row.repoID}`
                                            : `https://huggingface.co/${row.repoID}`,
                                    )
                                "
                            >
                                {{ row.repoID }}
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column prop="downloads" :label="$t('aiTools.model.downloadsCount')" width="130" />
                    <el-table-column prop="likes" :label="$t('aiTools.model.likes')" width="90" />
                    <el-table-column :label="$t('commons.table.operate')" width="200" fixed="right">
                        <template #default="{ row }">
                            <el-button
                                v-permission
                                link
                                type="primary"
                                icon="Download"
                                @click="downloadRef?.acceptParams(row.repoID, 'all', [currentSource])"
                            >
                                {{ $t('aiTools.model.model') }}
                            </el-button>
                            <el-button
                                v-permission
                                link
                                type="primary"
                                icon="Download"
                                @click="downloadRef?.acceptParams(row.repoID, 'tokenizer', [currentSource])"
                            >
                                {{ $t('aiTools.benchmark.tokenizer') }}
                            </el-button>
                        </template>
                    </el-table-column>
                </ComplexTable>
            </div>
        </DrawerPro>
        <component :is="recipesView" v-if="currentTab === 'recipes' && recipesView" :model-path="recipeModelPath">
            <template #search>
                <el-button
                    v-for="button in buttons"
                    :key="button.value"
                    class="tag-button"
                    :class="{ 'no-active': currentTab !== button.value }"
                    :type="currentTab === button.value ? 'primary' : undefined"
                    :aria-pressed="currentTab === button.value"
                    @click="currentTab = button.value"
                >
                    {{ button.label }}
                </el-button>
            </template>
        </component>
        <Download ref="downloadRef" @search="handleDownloadQueued" />
        <DownloadDelete ref="downloadDeleteRef" @search="loadDownloads" />
    </div>
</template>

<script setup lang="ts">
import { computed, defineAsyncComponent, onMounted, reactive, ref, watch, type Component } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { ElMessageBox, FormInstance } from 'element-plus';
import { checkFile } from '@/api/modules/files';
import i18n from '@/lang';
import { Rules } from '@/global/form-rules';
import { MsgError, MsgSuccess } from '@/utils/message';
import { routerToFileWithPath } from '@/utils/router';
import { toLink } from '@/utils/browser';
import DownloadDelete from './delete/index.vue';
import Download from './download/index.vue';
import SourceSelect from './source-select.vue';
import ProviderLabel from '@/components/agent-provider-label/index.vue';
import type { FuTableColumnConfig } from '@/components/table/shared';
import { ModelDownloader } from '@/api/interface/model-downloader';
import {
    cancelModelDownloadTask,
    getModelDownloaderSettings,
    pageFeaturedModels,
    syncFeaturedModels,
    pageModelDownloadTasks,
    retryModelDownloadTask,
    searchHFModels,
    searchModelScopeModels,
    updateModelDownloaderSettings,
    updateFeaturedModelsSync,
} from '@/api/modules/model-downloader';
import { useGlobalStore } from '@/composables/useGlobalStore';

const emit = defineEmits<{ deploy: [modelDir: string] }>();
const { currentNode, isMobile, isOffline, language, isProductPro } = useGlobalStore();
const downloadTypeFilter = ref<ModelDownloader.DownloadType | ''>('');

type DownloaderTab = 'featured' | 'downloader' | 'recipes' | 'settings';
type DownloaderSource = ModelDownloader.Source;
type SortValue = ModelDownloader.ModelSearch['sort'];

interface SortOption {
    label: string;
    value: SortValue;
}

const searchCardLoading = ref(false);
let searchRequest = 0;
const savingSettings = ref(false);
const route = useRoute();
const router = useRouter();
const pagePath = route.path;
const recipeModules = import.meta.glob<{ default: Component }>('@/xpack/views/vllm/template/index.vue');
const recipeLoader = Object.values(recipeModules)[0];
const recipesView = recipeLoader ? defineAsyncComponent(recipeLoader) : undefined;
const recipeModelPath = ref('');
const currentTab = ref<DownloaderTab>('featured');
const featuredViewMode = ref<'table' | 'card'>('table');
const featuredFamily = ref('');
const featuredWeightProvider = ref('');
const featuredSearchName = ref('');
const familyProviders: Record<string, string> = {
    Qwen: 'bailian',
    DeepSeek: 'deepseek',
    GLM: 'zai',
};
const syncingFeatured = ref(false);
const autoSync = ref<'Enable' | 'Disable'>();
const featuredModels = ref<ModelDownloader.FeaturedModel[]>([]);
const featuredColumns = ref<FuTableColumnConfig[]>([]);
const featuredFamilies = ref<string[]>([]);
const featuredWeightProviders = ref<string[]>([]);
const featuredPagination = reactive({ currentPage: 1, pageSize: 20, total: 0 });
const featuredLoading = ref(false);
let featuredRequest = 0;
const featuredFields = [
    { prop: 'repoID', label: 'modelName', width: 400 },
    { prop: 'family', label: 'modelCategory', width: 140 },
    { prop: 'weightProvider', label: 'weightProvider', width: 180 },
    { prop: 'architecture', label: 'modelArchitecture', width: 140 },
    { prop: 'backboneParams', label: 'backboneParams', width: 170 },
    { prop: 'activeParams', label: 'activeParams', width: 170 },
    { prop: 'engram', label: 'engram', width: 160 },
    { prop: 'weightGB', label: 'weightFootprint', width: 180 },
] as const;
const currentSource = ref<DownloaderSource>('modelscope');
const models = ref<ModelDownloader.ModelItem[]>([]);
const downloads = ref<ModelDownloader.TaskItem[]>([]);
const searchOpen = ref(false);
watch(
    [currentSource, currentNode, searchOpen],
    () => {
        searchRequest++;
        searchCardLoading.value = false;
    },
    { flush: 'sync' },
);
const settingFormRef = ref<FormInstance>();
const downloadDeleteRef = ref<InstanceType<typeof DownloadDelete>>();
const downloadRef = ref<InstanceType<typeof Download>>();

const openFeaturedModel = (row: ModelDownloader.FeaturedModel) => {
    const source = row.sources.includes('modelscope') ? 'modelscope' : 'huggingface';
    const repoID = row.sourceModelIds?.[source]?.trim() || row.repoID;
    toLink(
        source === 'huggingface' ? `https://huggingface.co/${repoID}` : `https://www.modelscope.cn/models/${repoID}`,
    );
};

const buttons = computed<Array<{ label: string; value: DownloaderTab }>>(() => [
    { label: i18n.global.t('aiTools.model.featuredModels'), value: 'featured' },
    { label: i18n.global.t('aiTools.model.downloader'), value: 'downloader' },
    ...(recipesView && isProductPro.value
        ? [{ label: i18n.global.t('aiTools.model.recipes'), value: 'recipes' as const }]
        : []),
    { label: i18n.global.t('commons.button.set'), value: 'settings' },
]);

const settings = reactive<ModelDownloader.Setting>({
    modelDir: '/opt/1panel/ai/models',
    hfEndpoint: '',
    hfToken: '',
    modelScopeEndpoint: 'https://www.modelscope.cn',
    modelScopeToken: '',
});

const settingRules = reactive({
    modelDir: [Rules.requiredInput],
});

const searchForm = reactive<ModelDownloader.ModelSearch>({
    query: '',
    sort: 'downloads',
});

const searchPagination = reactive({
    currentPage: 1,
    pageSize: 50,
    total: 0,
});

const downloadPagination = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const currentTitle = computed(() => {
    return (
        buttons.value.find((item) => item.value === currentTab.value)?.label ||
        i18n.global.t('aiTools.model.downloader')
    );
});

const huggingFaceSortOptions = computed<SortOption[]>(() => [
    { label: i18n.global.t('aiTools.model.sortTrending'), value: 'trending' },
    { label: i18n.global.t('aiTools.model.sortLikes'), value: 'likes' },
    { label: i18n.global.t('aiTools.model.sortDownloads'), value: 'downloads' },
    { label: i18n.global.t('aiTools.model.sortCreated'), value: 'created' },
    { label: i18n.global.t('aiTools.model.sortUpdated'), value: 'updated' },
    {
        label: i18n.global.t('aiTools.model.sortMostParameters'),
        value: 'most_params',
    },
    {
        label: i18n.global.t('aiTools.model.sortLeastParameters'),
        value: 'least_params',
    },
]);

const modelScopeSortOptions = computed<SortOption[]>(() => [
    { label: i18n.global.t('aiTools.model.sortDefault'), value: 'trending' },
    { label: i18n.global.t('aiTools.model.sortDownloads'), value: 'downloads' },
    { label: i18n.global.t('aiTools.model.sortLikes'), value: 'likes' },
    { label: i18n.global.t('aiTools.model.sortUpdated'), value: 'updated' },
]);

const sortOptions = computed(() => {
    return currentSource.value === 'modelscope' ? modelScopeSortOptions.value : huggingFaceSortOptions.value;
});

const loadSettings = async () => {
    const node = currentNode.value;
    const res = await getModelDownloaderSettings();
    if (node === currentNode.value) {
        Object.assign(settings, res.data);
        autoSync.value = res.data.featuredModelsSync;
    }
};

const saveSettings = async () => {
    await settingFormRef.value?.validate();
    savingSettings.value = true;
    try {
        await updateModelDownloaderSettings(settings);
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        await loadSettings();
    } finally {
        savingSettings.value = false;
    }
};

const handleSourceChange = () => {
    if (!sortOptions.value.some((item) => item.value === searchForm.sort)) {
        searchForm.sort = 'trending';
    }
    searchPagination.currentPage = 1;
    searchPagination.total = 0;
    models.value = [];
};

const buildSearchRequest = (sort: SortValue = searchForm.sort): ModelDownloader.ModelSearch => {
    return {
        query: searchForm.query,
        sort,
        page: searchPagination.currentPage,
        pageSize: searchPagination.pageSize,
    };
};

const handleSortChange = async (sort: SortValue) => {
    searchPagination.currentPage = 1;
    await searchModels(sort);
};

const handleSearchSubmit = async () => {
    searchPagination.currentPage = 1;
    await searchModels();
};

const searchModels = async (sort: SortValue = searchForm.sort) => {
    const request = ++searchRequest;
    searchCardLoading.value = true;
    try {
        const req = buildSearchRequest(sort);
        const res =
            currentSource.value === 'modelscope' ? await searchModelScopeModels(req) : await searchHFModels(req);
        if (request !== searchRequest) return;
        models.value = res.data.items || [];
        searchPagination.total = res.data.total || 0;
    } finally {
        if (request === searchRequest) searchCardLoading.value = false;
    }
};

const handleDownloadQueued = async () => {
    searchOpen.value = false;
    downloadTypeFilter.value = '';
    downloadPagination.currentPage = 1;
    if (currentTab.value === 'downloader') await loadDownloads();
    else currentTab.value = 'downloader';
};

const changeDownloadType = async () => {
    downloadPagination.currentPage = 1;
    await loadDownloads();
};

let downloadRequest = 0;
const loadDownloads = async () => {
    const request = ++downloadRequest;
    const node = currentNode.value;
    const res = await pageModelDownloadTasks({
        page: downloadPagination.currentPage,
        pageSize: downloadPagination.pageSize,
        info: '',
        status: '',
        downloadType: downloadTypeFilter.value,
        includeLocal: true,
    });
    if (node !== currentNode.value || request !== downloadRequest) return;
    downloads.value = res.data.items || [];
    downloadPagination.total = res.data.total || 0;
    const lastPage = Math.max(1, Math.ceil(downloadPagination.total / downloadPagination.pageSize));
    if (downloadPagination.currentPage > lastPage) {
        downloadPagination.currentPage = lastPage;
        await loadDownloads();
    }
};

const openLocalModelPath = async (path: string) => {
    const res = await checkFile(path, false);
    if (res.data) {
        await routerToFileWithPath(path);
        return;
    }
    MsgError(i18n.global.t('file.noSuchFile'));
};

const cancelTask = async (row: ModelDownloader.TaskItem) => {
    await ElMessageBox.confirm(
        i18n.global.t('aiTools.model.cancelTaskConfirm'),
        i18n.global.t('commons.button.cancel'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    );
    await cancelModelDownloadTask({ id: row.id });
    await loadDownloads();
};

const retryTask = async (row: ModelDownloader.TaskItem) => {
    await retryModelDownloadTask({ id: row.id });
    await loadDownloads();
};

const handleSyncOp = async (command: string) => {
    if (command !== 'sync' && command !== 'Enable' && command !== 'Disable') return;
    syncingFeatured.value = true;
    try {
        const node = currentNode.value;
        if (command === 'sync') {
            await syncFeaturedModels();
            if (node !== currentNode.value) return;
            await handleFeaturedSearch();
        } else {
            await updateFeaturedModelsSync(command);
            if (node !== currentNode.value) return;
            autoSync.value = command;
        }
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    } finally {
        syncingFeatured.value = false;
    }
};

const handleFeaturedSearch = async () => {
    featuredPagination.currentPage = 1;
    await loadFeaturedModels();
};

const loadFeaturedModels = async () => {
    const node = currentNode.value;
    const request = ++featuredRequest;
    featuredLoading.value = true;
    try {
        const res = await pageFeaturedModels({
            page: featuredPagination.currentPage,
            pageSize: featuredPagination.pageSize,
            family: featuredFamily.value,
            weightProvider: featuredWeightProvider.value,
            info: featuredSearchName.value,
        });
        if (node !== currentNode.value || request !== featuredRequest) return;
        featuredModels.value = res.data.items;
        featuredPagination.total = res.data.total;
        featuredFamilies.value = res.data.families;
        featuredWeightProviders.value = res.data.weightProviders;
    } finally {
        if (request === featuredRequest) featuredLoading.value = false;
    }
};

const refreshCurrent = async () => {
    if (currentTab.value === 'featured') {
        await loadFeaturedModels();
        return;
    }
    if (currentTab.value === 'downloader') {
        await loadDownloads();
        return;
    }
    if (currentTab.value === 'settings') {
        await loadSettings();
    }
};

const statusTagType = (status: string) => {
    switch (status) {
        case 'Success':
            return 'success';
        case 'Downloading':
            return 'primary';
        case 'Failed':
        case 'Canceled':
            return 'danger';
        default:
            return 'info';
    }
};

const statusText = (status: string) => {
    if (status === 'Downloading') {
        return i18n.global.t('aiTools.model.downloading');
    }
    if (status === 'Success') {
        return i18n.global.t('commons.status.success');
    }
    if (status === 'Failed') {
        return i18n.global.t('commons.status.failed');
    }
    if (status === 'Canceled') {
        return i18n.global.t('aiTools.model.canceled');
    }
    return i18n.global.t('commons.status.waiting');
};

const sourceLabel = (source: string) => {
    if (source === 'huggingface') {
        return 'HuggingFace';
    }
    if (source === 'modelscope') {
        return 'ModelScope';
    }
    return source || '-';
};

const formatBytes = (value: number) => {
    if (!value) {
        return '-';
    }
    if (value < 1024 * 1024) {
        return `${(value / 1024).toFixed(1)} KB`;
    }
    if (value < 1024 * 1024 * 1024) {
        return `${(value / 1024 / 1024).toFixed(1)} MB`;
    }
    return `${(value / 1024 / 1024 / 1024).toFixed(1)} GB`;
};

const formatProgress = (value: number) => {
    return Number(Number(value || 0).toFixed(1));
};

const formatSizePair = (downloaded: number, total: number) => {
    if (!downloaded && !total) {
        return '';
    }
    if (!total) {
        return formatBytes(downloaded);
    }
    return `${formatBytes(downloaded)} / ${formatBytes(total)}`;
};

watch(
    [() => route.query.tab, () => route.path, buttons],
    ([tab, path]) => {
        if (path !== pagePath) return;
        currentTab.value = buttons.value.some((button) => button.value === tab) ? (tab as DownloaderTab) : 'featured';
    },
    { immediate: true },
);

watch(currentTab, async (value) => {
    if (value !== 'recipes') recipeModelPath.value = '';
    if (route.path === pagePath && route.query.tab !== value)
        await router.replace({ query: { ...route.query, tab: value } });
    if (value !== 'settings') await refreshCurrent();
});

watch(currentNode, async () => {
    recipeModelPath.value = '';
    featuredRequest++;
    featuredLoading.value = false;
    featuredModels.value = [];
    featuredFamilies.value = [];
    featuredWeightProviders.value = [];
    featuredPagination.currentPage = 1;
    featuredPagination.total = 0;
    autoSync.value = undefined;
    if (currentTab.value === 'settings') {
        await loadSettings();
        return;
    }
    await Promise.all([refreshCurrent(), loadSettings()]);
});

onMounted(() => {
    loadSettings();
    refreshCurrent();
});
</script>

<style scoped lang="scss">
.featured-filters {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px;
}

.model-filter {
    width: 220px;
}

.model-search {
    width: 220px;
    :deep(.search-button) {
        width: 100%;
    }
}

@media (max-width: 768px) {
    .featured-filters,
    .model-filter,
    .model-search {
        width: 100%;
    }
}

.featured-models {
    :deep(.el-button + .el-button) {
        margin-left: 20px;
    }

    :deep(.complex-table__card) {
        box-shadow: none;
    }

    :deep(.complex-table__card > .el-card__body) {
        min-height: 0;
    }

    :deep(.complex-table__card-header) {
        align-items: flex-start;
    }

    :deep(.complex-table__card-status) {
        max-width: 40%;
    }

    :deep(.complex-table__card-value) {
        white-space: normal;
        overflow-wrap: anywhere;
    }

    :deep(.complex-table__card-buttons) {
        padding-top: 12px;
        margin-top: 16px;
    }
}

.featured-provider {
    max-width: 100%;
    height: auto;
    min-height: 20px;
    white-space: normal;
    overflow-wrap: anywhere;
}

.featured-name-label {
    display: block;
    margin-bottom: 6px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
    font-weight: 400;
}

.featured-name {
    max-width: 100%;
    font-size: inherit;
    font-weight: inherit;
    white-space: normal;
    text-align: left;
    overflow-wrap: anywhere;
    line-height: 1.6;
}

.featured-description {
    margin-top: 4px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
    font-weight: 400;
    line-height: 1.5;
    white-space: normal;
    overflow-wrap: anywhere;
}

.source-switch {
    margin-bottom: 16px;
}

.search-row {
    display: flex;
    align-items: center;
    gap: 12px;
}

.search-input {
    flex: 1;
    min-width: 0;

    :deep(.search-button) {
        width: 100%;
    }
}

.sort-select {
    flex: 0 0 160px;
    width: 160px;
}

.progress-cell {
    display: flex;
    flex-direction: column;
    align-items: stretch;
    gap: 2px;
    width: 100%;
}

.status-tag-content {
    display: inline-flex;
    align-items: center;
    gap: 4px;
}

.progress-bar {
    flex: 1;
    min-width: 110px;

    :deep(.el-progress__text) {
        min-width: 36px;
        margin-left: 6px;
        font-size: 12px !important;
    }
}

.progress-size {
    min-width: 0;
    color: var(--el-text-color-secondary);
    font-size: 12px;
    line-height: 14px;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.model-path-link {
    max-width: 100%;
    justify-content: flex-start;

    :deep(> span) {
        display: block;
        min-width: 0;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        text-align: left;
    }
}
</style>
