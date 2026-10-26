<template>
    <DrawerPro
        v-model="visible"
        :header="item ? $t('commons.button.edit') : $t('panelMCP.create')"
        size="min(800px, 100vw)"
        :confirm-before-close="true"
        @before-close="beforeClose"
    >
        <template v-if="created">
            <el-alert
                :title="$t(secret ? 'panelMCP.saveSecret' : 'panelMCP.alreadyCreated')"
                :type="secret ? 'warning' : 'info'"
                :closable="false"
                show-icon
                class="!mb-4"
            />
            <el-form label-position="top" class="min-w-0">
                <el-form-item :label="$t('commons.table.name')">
                    <span class="break-words">{{ created.name }}</span>
                </el-form-item>
                <el-form-item v-if="secret" label="Token">
                    <el-input :model-value="secret" readonly autocomplete="off" class="[&_input]:!font-mono">
                        <template #append><CopyButton :content="secret" :is-icon="false" /></template>
                    </el-input>
                </el-form-item>
                <el-form-item v-if="secret && clientConfig" :label="$t('panelMCP.clientConfig')">
                    <div class="w-full min-w-0">
                        <div class="mb-2 flex justify-end">
                            <CopyButton :content="clientConfig" :is-icon="false" />
                        </div>
                        <pre
                            class="m-0 max-h-80 overflow-auto whitespace-pre-wrap break-all rounded border border-solid border-[var(--el-border-color-light)] bg-[var(--el-fill-color-light)] p-3 font-mono text-xs leading-5"
                            >{{ clientConfig }}</pre>
                    </div>
                </el-form-item>
                <p v-if="!endpoint" class="input-help mt-0 leading-5">{{ $t('panelMCP.endpointUnavailable') }}</p>
            </el-form>
            <el-checkbox v-if="secret" v-model="saved" class="!mr-0 !h-auto max-w-full !items-start !whitespace-normal">
                <span class="whitespace-normal break-words leading-5">{{ $t('panelMCP.tokenSaved') }}</span>
            </el-checkbox>
        </template>
        <el-form
            v-else
            ref="formRef"
            :model="form"
            :rules="rules"
            label-position="top"
            class="min-w-0"
            :disabled="busy || uncertain || conflict || !catalogAvailable || !mappingAvailable || catalogChanged"
            @submit.prevent="submit"
        >
            <el-alert
                v-if="uncertain || conflict || !catalogAvailable || !mappingAvailable || catalogChanged"
                :title="
                    $t(
                        policyConflict || catalogChanged
                            ? 'panelMCP.catalogChanged'
                            : conflict
                              ? 'panelMCP.conflict'
                              : uncertain
                                ? 'panelMCP.uncertain'
                                : 'panelMCP.catalogUnavailable',
                    )
                "
                type="warning"
                :closable="false"
                class="!mb-4"
            />
            <el-alert
                v-if="reviewRequired"
                :title="$t('panelMCP.grantsNeedReview')"
                type="warning"
                :closable="false"
                class="!mb-4"
            />
            <el-form-item :label="$t('commons.table.name')" prop="name">
                <el-input v-model.trim="form.name" :maxlength="64" />
            </el-form-item>
            <el-form-item :label="$t('commons.table.description')" prop="description">
                <el-input v-model="form.description" type="textarea" :maxlength="256" :rows="2" />
            </el-form-item>
            <el-form-item :label="$t('panelMCP.nodes')" prop="nodeIds">
                <el-select v-model="form.nodeIds" multiple filterable class="w-full">
                    <el-option
                        v-for="node in nodes"
                        :key="node.id"
                        :value="node.id"
                        :label="node.isLocal ? $t('panelMCP.primaryNode') : `${node.name} (ID: ${node.id})`"
                    />
                </el-select>
                <span class="input-help mt-2 leading-5">{{ $t('panelMCP.nodeGrantHelp') }}</span>
                <div
                    v-for="id in unavailableNodes"
                    :key="id"
                    class="mt-2 flex w-full min-w-0 flex-wrap items-center gap-2 rounded border border-solid border-[var(--el-color-warning-light-5)] bg-[var(--el-color-warning-light-9)] p-3"
                >
                    <span class="min-w-0 flex-1 break-all">{{ $t('panelMCP.nodeUnavailable', [id]) }}</span>
                    <el-button type="danger" link @click="form.nodeIds = form.nodeIds.filter((node) => node !== id)">
                        {{ $t('panelMCP.removeGrant') }}
                    </el-button>
                </div>
            </el-form-item>
            <el-form-item :label="$t('panelMCP.tools')" prop="tools">
                <div class="mb-2 flex w-full min-w-0 flex-col gap-2 sm:flex-row">
                    <el-input
                        v-model="toolSearch"
                        clearable
                        prefix-icon="Search"
                        :placeholder="$t('panelMCP.searchTools')"
                        class="min-w-0 flex-1"
                    />
                    <el-select v-model="toolFilter" :aria-label="$t('panelMCP.toolFilter')" class="w-full sm:!w-40">
                        <el-option :label="$t('panelMCP.allTools')" value="all" />
                        <el-option :label="$t('panelMCP.readOnly')" value="read" />
                        <el-option :label="$t('panelMCP.write')" value="write" />
                    </el-select>
                </div>
                <div class="mb-2 w-full min-w-0 text-xs leading-5 text-[var(--el-text-color-secondary)]">
                    <span>{{ $t('panelMCP.selectedTools', [form.tools.length, nodeTools.length]) }}</span>
                    <span>· {{ $t('panelMCP.filteredTools', [filteredTools.length]) }}</span>
                    <span v-if="hiddenSelectedCount" class="block">
                        {{ $t('panelMCP.outsideSelection', [hiddenSelectedCount]) }}
                    </span>
                </div>
                <div class="mcp-grant-toolbar mb-2 flex w-full min-w-0 flex-wrap items-center justify-between gap-2">
                    <label class="inline-flex min-h-8 cursor-pointer items-center gap-2 text-sm leading-5">
                        <input
                            v-model="selectedOnly"
                            type="checkbox"
                            class="mcp-selected-only h-3.5 w-3.5 shrink-0 accent-[var(--el-color-primary)]"
                        />
                        <span>{{ $t('panelMCP.selectedOnly') }}</span>
                    </label>
                    <el-dropdown
                        trigger="click"
                        placement="bottom-end"
                        :disabled="grantControlsDisabled || !filteredTools.length"
                        @command="changeFilteredGrants"
                    >
                        <el-button native-type="button" :disabled="grantControlsDisabled || !filteredTools.length">
                            {{ $t('panelMCP.bulkActions') }}
                            <el-icon class="el-icon--right" aria-hidden="true"><ArrowDown /></el-icon>
                        </el-button>
                        <template #dropdown>
                            <el-dropdown-menu class="mcp-tool-batch-menu max-w-[calc(100vw-2rem)]">
                                <el-dropdown-item
                                    command="readOnly"
                                    class="!max-w-[calc(100vw-2rem)] !whitespace-normal !break-words !py-2 !leading-5"
                                    :disabled="grantControlsDisabled || !filteredTools.length"
                                >
                                    {{ $t('panelMCP.readOnlyVisible', [filteredTools.length]) }}
                                </el-dropdown-item>
                                <el-dropdown-item
                                    command="select"
                                    class="!max-w-[calc(100vw-2rem)] !whitespace-normal !break-words !py-2 !leading-5"
                                    :disabled="grantControlsDisabled || !filteredTools.length"
                                >
                                    {{ $t('panelMCP.selectVisible', [filteredTools.length]) }}
                                </el-dropdown-item>
                                <el-dropdown-item
                                    command="clear"
                                    class="!max-w-[calc(100vw-2rem)] !whitespace-normal !break-words !py-2 !leading-5"
                                    :disabled="grantControlsDisabled || !filteredTools.length"
                                >
                                    {{ $t('panelMCP.clearVisible', [filteredTools.length]) }}
                                </el-dropdown-item>
                            </el-dropdown-menu>
                        </template>
                    </el-dropdown>
                </div>
                <p class="mb-3 mt-0 w-full text-xs leading-5 text-[var(--el-text-color-secondary)]">
                    {{ $t('panelMCP.bulkHelp') }}
                </p>
                <div
                    class="mcp-tool-picker grid w-full min-w-0 overflow-hidden rounded border border-solid border-[var(--el-border-color-light)] md:grid-cols-[10rem_minmax(0,1fr)]"
                >
                    <div class="border-0 border-b border-solid border-[var(--el-border-color-light)] p-3 md:hidden">
                        <label for="mcp-tool-category" class="mb-1 block text-xs text-[var(--el-text-color-secondary)]">
                            {{ $t('panelMCP.toolCategory') }}
                        </label>
                        <select
                            id="mcp-tool-category"
                            v-model="activeCategory"
                            class="w-full min-w-0 rounded border border-solid border-[var(--el-border-color)] bg-[var(--el-bg-color)] px-3 py-2 text-sm text-[var(--el-text-color-regular)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-[var(--el-color-primary)]"
                        >
                            <option value="all">{{ $t('panelMCP.allCategories') }}</option>
                            <option v-for="group in categoryGroups" :key="group.id" :value="group.id">
                                {{ categoryLabel(group.id, group.name) }} ·
                                {{ $t('panelMCP.groupSelected', [group.selected, group.total]) }}
                            </option>
                        </select>
                    </div>
                    <nav
                        :aria-label="$t('panelMCP.toolCategory')"
                        class="hidden min-w-0 flex-col gap-1 border-0 border-r border-solid border-[var(--el-border-color-light)] bg-[var(--el-fill-color-lighter)] p-2 md:flex"
                    >
                        <button
                            type="button"
                            :aria-pressed="activeCategory === 'all'"
                            class="mcp-category-button w-full cursor-pointer rounded border-0 px-3 py-2 text-left leading-5 transition-colors focus-visible:outline focus-visible:outline-2 focus-visible:outline-[var(--el-color-primary)]"
                            :class="
                                activeCategory === 'all'
                                    ? 'bg-[var(--el-color-primary-light-9)] text-[var(--el-color-primary)]'
                                    : 'bg-transparent text-[var(--el-text-color-regular)] hover:bg-[var(--el-fill-color)]'
                            "
                            @click="activeCategory = 'all'"
                        >
                            <span class="block break-words text-sm font-medium">
                                {{ $t('panelMCP.allCategories') }}
                            </span>
                            <span class="mt-1 block text-xs">
                                {{
                                    $t('panelMCP.groupSelected', [
                                        selectedToolSummary.read + selectedToolSummary.write,
                                        nodeTools.length,
                                    ])
                                }}
                            </span>
                        </button>
                        <button
                            v-for="group in categoryGroups"
                            :key="group.id"
                            type="button"
                            :aria-pressed="activeCategory === group.id"
                            class="mcp-category-button w-full cursor-pointer rounded border-0 px-3 py-2 text-left leading-5 transition-colors focus-visible:outline focus-visible:outline-2 focus-visible:outline-[var(--el-color-primary)]"
                            :class="
                                activeCategory === group.id
                                    ? 'bg-[var(--el-color-primary-light-9)] text-[var(--el-color-primary)]'
                                    : 'bg-transparent text-[var(--el-text-color-regular)] hover:bg-[var(--el-fill-color)]'
                            "
                            @click="activeCategory = group.id"
                        >
                            <span class="block break-words text-sm font-medium">
                                {{ categoryLabel(group.id, group.name) }}
                            </span>
                            <span class="mt-1 block text-xs">
                                {{ $t('panelMCP.groupSelected', [group.selected, group.total]) }}
                            </span>
                        </button>
                    </nav>
                    <div
                        class="mcp-tool-list max-h-96 min-w-0 overflow-y-auto overscroll-contain"
                        role="region"
                        :aria-label="$t('panelMCP.tools')"
                    >
                        <el-checkbox-group v-model="form.tools" class="w-full min-w-0">
                            <section v-for="group in toolGroups" :key="group.id" class="min-w-0">
                                <div
                                    class="flex flex-wrap items-center justify-between gap-2 bg-[var(--el-fill-color-lighter)] px-3 py-2 text-xs leading-5 text-[var(--el-text-color-secondary)]"
                                >
                                    <span class="min-w-0 break-words font-medium">
                                        {{ categoryLabel(group.id, group.name) }}
                                    </span>
                                    <span>{{ $t('panelMCP.groupSelected', [group.selected, group.total]) }}</span>
                                </div>
                                <div
                                    v-for="tool in group.tools"
                                    :key="tool.toolId"
                                    class="mcp-tool-row min-w-0 border-0 border-b border-solid border-[var(--el-border-color-lighter)] px-3 py-3 last:border-b-0"
                                >
                                    <div class="flex min-w-0 items-start gap-3">
                                        <el-checkbox
                                            :value="tool.toolId"
                                            class="!mr-0 !h-auto min-w-0 flex-1 !items-start !whitespace-normal"
                                        >
                                            <span class="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1">
                                                <span class="break-all">{{ tool.name }}</span>
                                                <el-tag size="small" :type="tool.readOnly ? 'info' : 'warning'">
                                                    {{ $t(tool.readOnly ? 'panelMCP.readOnly' : 'panelMCP.write') }}
                                                </el-tag>
                                            </span>
                                        </el-checkbox>
                                        <button
                                            v-if="tool.description || tool.name !== tool.toolId"
                                            type="button"
                                            :aria-expanded="expandedToolIds.includes(tool.toolId)"
                                            :aria-controls="`mcp-tool-detail-${tool.toolId}`"
                                            class="mcp-tool-details shrink-0 cursor-pointer rounded border-0 bg-transparent p-0 text-xs leading-5 text-[var(--el-color-primary)] focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-[var(--el-color-primary)]"
                                            @click="toggleToolDetails(tool.toolId)"
                                        >
                                            {{
                                                $t(
                                                    expandedToolIds.includes(tool.toolId)
                                                        ? 'panelMCP.hideToolDetails'
                                                        : 'panelMCP.toolDetails',
                                                )
                                            }}
                                        </button>
                                    </div>
                                    <p
                                        v-if="tool.description && !expandedToolIds.includes(tool.toolId)"
                                        class="mb-0 mt-1 truncate pl-6 text-xs leading-5 text-[var(--el-text-color-secondary)]"
                                    >
                                        {{ tool.description }}
                                    </p>
                                    <div
                                        v-if="expandedToolIds.includes(tool.toolId)"
                                        :id="`mcp-tool-detail-${tool.toolId}`"
                                        class="mt-2 space-y-1 pl-6 text-xs leading-5 text-[var(--el-text-color-secondary)]"
                                    >
                                        <div v-if="tool.name !== tool.toolId" class="break-all font-mono">
                                            {{ tool.toolId }}
                                        </div>
                                        <div class="whitespace-pre-wrap break-words">{{ tool.description }}</div>
                                    </div>
                                </div>
                            </section>
                        </el-checkbox-group>
                        <p
                            v-if="!displayedTools.length"
                            class="m-0 px-4 py-8 text-center text-sm leading-5 text-[var(--el-text-color-secondary)]"
                        >
                            {{ $t('panelMCP.noMatchingTools') }}
                        </p>
                    </div>
                </div>
                <div v-if="unavailableTools.length" class="mt-3 w-full min-w-0">
                    <p class="mb-2 mt-0 text-sm font-medium text-[var(--el-color-warning)]">
                        {{ $t('panelMCP.unavailableSelected', [unavailableTools.length]) }}
                    </p>
                    <div
                        v-for="id in unavailableTools"
                        :key="id"
                        class="mt-2 flex w-full min-w-0 flex-wrap items-center gap-2 rounded border border-solid border-[var(--el-color-warning-light-5)] bg-[var(--el-color-warning-light-9)] p-3"
                    >
                        <span class="min-w-0 flex-1 break-all font-mono">{{ id }}</span>
                        <el-button type="danger" link @click="removeUnavailableTool(id)">
                            {{ $t('panelMCP.removeGrant') }}
                        </el-button>
                    </div>
                </div>
            </el-form-item>
            <el-alert
                v-if="writeSelected"
                :title="$t('panelMCP.writeHelp')"
                type="warning"
                :closable="false"
                class="!mb-4"
            />
            <el-checkbox
                v-if="reviewRequired"
                v-model="policyAccepted"
                class="mb-4 !mr-0 !h-auto max-w-full !items-start !whitespace-normal"
            >
                <span class="whitespace-normal break-words leading-5">{{ $t('panelMCP.acceptToolPolicy') }}</span>
            </el-checkbox>
            <el-form-item :label="$t('apiKeyManagement.expiresAt')" prop="expiresAt">
                <el-checkbox v-model="neverExpires" class="mb-2 w-full">{{ $t('apiKeyManagement.never') }}</el-checkbox>
                <el-date-picker
                    v-if="!neverExpires"
                    v-model="form.expiresAt"
                    type="datetime"
                    :clearable="false"
                    :show-now="false"
                    class="!w-full"
                />
            </el-form-item>
        </el-form>
        <template #footer>
            <p v-if="!created" class="mb-3 mt-0 text-xs leading-5 text-[var(--el-text-color-secondary)]">
                {{
                    $t('panelMCP.selectionSummary', [
                        selectedToolSummary.read,
                        selectedToolSummary.write,
                        selectedToolSummary.unavailable,
                    ])
                }}
            </p>
            <div class="flex flex-wrap items-center justify-end gap-2 [&>.el-button]:!ml-0">
                <el-button v-if="!created" :disabled="busy" @click="beforeClose(discard)">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button v-if="created" type="primary" :disabled="!!secret && !saved" @click="discard">
                    {{ $t('commons.button.confirm') }}
                </el-button>
                <el-button v-else type="primary" :loading="busy" :disabled="!canSubmit" @click="submit">
                    {{ $t(item ? 'commons.button.save' : 'panelMCP.create') }}
                </el-button>
            </div>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onDeactivated, reactive, ref, watch } from 'vue';
import { ElMessageBox, type FormInstance } from 'element-plus';
import { onBeforeRouteLeave } from 'vue-router';
import { v4 as uuidv4 } from 'uuid';
import type { PanelMCP } from '@/api/interface/panel-mcp';
import { createPanelMCPCredential, updatePanelMCPCredential } from '@/api/modules/panel-mcp';
import { Rules } from '@/global/form-rules';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { defaultAPIKeyExpiry } from '@/utils/api-key';
import {
    buildMCPClientConfig,
    changeVisibleMCPGrants,
    filterMCPTools,
    filterMCPToolsByNodes,
    groupMCPTools,
    getMCPStaleTools,
    isMCPMutationUncertain,
    isMCPPolicyChanged,
    isMCPRevisionConflict,
    isUsableMCPToolNodes,
    type MCPToolFilter,
} from '@/utils/panel-mcp';
import { MsgSuccess } from '@/utils/message';
import { hasManagePermissionAccess } from '@/utils/permission';
import i18n from '@/lang';

const props = defineProps<{
    tools: PanelMCP.Tool[];
    toolNodeIds: Record<string, string[]>;
    nodes: PanelMCP.Node[];
    catalogAvailable: boolean;
    endpoint: string;
}>();
const emit = defineEmits<{ changed: [] }>();
const { globalStore } = useGlobalStore();
const canManageCredential = computed(() => hasManagePermissionAccess('mcp_access_manage'));
const visible = ref(false);
const busy = ref(false);
const uncertain = ref(false);
const conflict = ref(false);
const policyConflict = ref(false);
const policyAccepted = ref(false);
const editorTools = ref<PanelMCP.Tool[]>([]);
const editorToolNodeIds = ref<Record<string, string[]>>({});
const mappingAvailable = computed(
    () =>
        isUsableMCPToolNodes(props.tools, props.toolNodeIds) &&
        isUsableMCPToolNodes(editorTools.value, editorToolNodeIds.value),
);
const nodeTools = computed(() =>
    mappingAvailable.value
        ? filterMCPToolsByNodes(editorTools.value, editorToolNodeIds.value, form.nodeIds, props.toolNodeIds)
        : [],
);
const toolSearch = ref('');
const toolFilter = ref<MCPToolFilter>('all');
const activeCategory = ref('all');
const selectedOnly = ref(false);
const expandedToolIds = ref<string[]>([]);
const filteredTools = computed(() => filterMCPTools(nodeTools.value, toolSearch.value, toolFilter.value));
const categoryGroups = computed(() => groupMCPTools(nodeTools.value, nodeTools.value, form.tools));
const displayedTools = computed(() =>
    filteredTools.value.filter(
        (tool) =>
            (activeCategory.value === 'all' || tool.categoryId === activeCategory.value) &&
            (!selectedOnly.value || form.tools.includes(tool.toolId)),
    ),
);
const toolGroups = computed(() => groupMCPTools(nodeTools.value, displayedTools.value, form.tools));
const hiddenSelectedCount = computed(
    () =>
        nodeTools.value.filter(
            (tool) =>
                form.tools.includes(tool.toolId) && !displayedTools.value.some((item) => item.toolId === tool.toolId),
        ).length,
);
const selectedToolSummary = computed(() => {
    const selected = nodeTools.value.filter((tool) => form.tools.includes(tool.toolId));
    return {
        read: selected.filter((tool) => tool.readOnly).length,
        write: selected.filter((tool) => !tool.readOnly).length,
        unavailable: unavailableTools.value.length,
    };
});
const categoryLabel = (id: string, fallback: string) => {
    const key = `panelMCP.category.${id}`;
    return i18n.global.te(key) ? i18n.global.t(key) : fallback;
};
const toggleToolDetails = (id: string) => {
    expandedToolIds.value = expandedToolIds.value.includes(id)
        ? expandedToolIds.value.filter((value) => value !== id)
        : [...expandedToolIds.value, id];
};
const changeFilteredGrants = (action: 'select' | 'clear' | 'readOnly') => {
    if (
        busy.value ||
        uncertain.value ||
        conflict.value ||
        !props.catalogAvailable ||
        !mappingAvailable.value ||
        catalogChanged.value
    )
        return;
    form.tools = changeVisibleMCPGrants(form.tools, filteredTools.value, action);
};
const secret = ref('');
const saved = ref(false);
const created = ref<PanelMCP.Credential>();
const item = ref<PanelMCP.Credential>();
const neverExpires = ref(false);
const formRef = ref<FormInstance>();
const form = reactive<Omit<PanelMCP.EditableCredential, 'toolPolicies' | 'toolReadOnly'>>({
    name: '',
    description: '',
    expiresAt: null,
    tools: [],
    nodeIds: [],
});
watch(toolSearch, (value) => {
    if (value.trim()) activeCategory.value = 'all';
});
watch(categoryGroups, (groups) => {
    if (activeCategory.value !== 'all' && !groups.some((group) => group.id === activeCategory.value)) {
        activeCategory.value = 'all';
    }
});
const clientConfig = computed(() =>
    globalStore.isLogin && canManageCredential.value && secret.value
        ? buildMCPClientConfig(props.endpoint, secret.value, 'generic', created.value?.id)
        : '',
);
const writeSelected = computed(() =>
    editorTools.value.some((tool) => !tool.readOnly && form.tools.includes(tool.toolId)),
);
const unavailableTools = computed(() => form.tools.filter((id) => !nodeTools.value.some((tool) => tool.toolId === id)));
const unavailableNodes = computed(() => form.nodeIds.filter((id) => !props.nodes.some((node) => node.id === id)));
const reviewRequired = computed(() => !!item.value && getMCPStaleTools(item.value, editorTools.value).length > 0);
const catalogChanged = computed(
    () =>
        props.catalogAvailable &&
        form.tools.some((id) => {
            const snapshot = editorTools.value.find((tool) => tool.toolId === id);
            const current = props.tools.find((tool) => tool.toolId === id);
            return (
                !!snapshot && (snapshot.policyHash !== current?.policyHash || snapshot.readOnly !== current?.readOnly)
            );
        }),
);
const canSubmit = computed(
    () =>
        visible.value &&
        globalStore.isLogin &&
        canManageCredential.value &&
        props.catalogAvailable &&
        mappingAvailable.value &&
        !catalogChanged.value &&
        !conflict.value &&
        !unavailableTools.value.length &&
        form.nodeIds.length > 0 &&
        !unavailableNodes.value.length &&
        (!reviewRequired.value || policyAccepted.value) &&
        (!!item.value || editorTools.value.length > 0),
);
const grantControlsDisabled = computed(
    () =>
        busy.value ||
        uncertain.value ||
        conflict.value ||
        !props.catalogAvailable ||
        !mappingAvailable.value ||
        catalogChanged.value,
);
let version = 0;
let requestID = '';
let pendingCreate: (PanelMCP.EditableCredential & { requestID: string }) | undefined;
const removeUnavailableTool = (id: string) => {
    if (unavailableTools.value.includes(id)) form.tools = form.tools.filter((tool) => tool !== id);
};
const rules = {
    name: [Rules.requiredInput],
    nodeIds: [
        { type: 'array', required: true, min: 1, message: i18n.global.t('panelMCP.selectNodes'), trigger: 'change' },
    ],
    expiresAt: [
        {
            validator: (_rule: unknown, value: string | null, done: (error?: Error) => void) =>
                done(
                    neverExpires.value || (value && new Date(value).getTime() > Date.now())
                        ? undefined
                        : new Error(i18n.global.t('apiKeyManagement.futureExpiry')),
                ),
            trigger: 'change',
        },
    ],
};
const discard = () => {
    version++;
    secret.value = '';
    created.value = undefined;
    pendingCreate = undefined;
    visible.value = false;
    busy.value = false;
};
const open = (existing?: PanelMCP.Credential) => {
    if (
        !globalStore.isLogin ||
        !canManageCredential.value ||
        !props.catalogAvailable ||
        !isUsableMCPToolNodes(props.tools, props.toolNodeIds)
    )
        return;
    discard();
    item.value = existing;
    saved.value = uncertain.value = conflict.value = policyConflict.value = policyAccepted.value = false;
    toolSearch.value = '';
    toolFilter.value = 'all';
    activeCategory.value = 'all';
    selectedOnly.value = false;
    expandedToolIds.value = [];
    editorTools.value = props.tools.map((tool) => ({
        ...tool,
        operations: JSON.parse(JSON.stringify(tool.operations)),
    }));
    editorToolNodeIds.value = Object.fromEntries(
        Object.entries(props.toolNodeIds).map(([id, nodes]) => [id, [...nodes]]),
    );
    const initialNodeIds = props.nodes.some((node) => node.isLocal) ? ['local'] : [];
    requestID = uuidv4();
    Object.assign(
        form,
        existing
            ? {
                  name: existing.name,
                  description: existing.description,
                  expiresAt: existing.expiresAt,
                  tools: [...existing.tools],
                  nodeIds: [...existing.nodeIds],
              }
            : {
                  name: '',
                  description: '',
                  expiresAt: defaultAPIKeyExpiry(),
                  tools: filterMCPToolsByNodes(editorTools.value, editorToolNodeIds.value, initialNodeIds)
                      .filter((tool) => tool.readOnly)
                      .map((tool) => tool.toolId),
                  nodeIds: initialNodeIds,
              },
    );
    neverExpires.value = !form.expiresAt;
    visible.value = true;
    nextTick(() => formRef.value?.clearValidate());
};
const beforeClose = async (done: () => void) => {
    if (busy.value) return;
    if (secret.value && !saved.value) {
        await ElMessageBox.alert(i18n.global.t('panelMCP.saveSecret'), i18n.global.t('panelMCP.title'));
        return;
    }
    if (uncertain.value) emit('changed');
    discard();
    done();
};
const submit = async () => {
    if (busy.value || !canSubmit.value) return;
    const ticket = version;
    busy.value = true;
    try {
        if (!(await formRef.value?.validate().catch(() => false)) || ticket !== version || !canSubmit.value) return;
        const payload: PanelMCP.EditableCredential = {
            name: form.name.trim(),
            description: form.description,
            expiresAt: neverExpires.value ? null : new Date(form.expiresAt!).toISOString(),
            tools: [...form.tools],
            nodeIds: [...form.nodeIds],
            toolPolicies: Object.fromEntries(
                form.tools.map((id) => [id, editorTools.value.find((tool) => tool.toolId === id)!.policyHash]),
            ),
            toolReadOnly: Object.fromEntries(
                form.tools.map((id) => [id, editorTools.value.find((tool) => tool.toolId === id)!.readOnly]),
            ),
        };
        if (item.value) {
            await updatePanelMCPCredential({ ...payload, id: item.value.id, revision: item.value.revision });
            if (ticket !== version || !visible.value || !globalStore.isLogin || !canManageCredential.value) return;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            visible.value = false;
        } else {
            pendingCreate ??= { ...payload, requestID };
            const response = await createPanelMCPCredential(pendingCreate);
            if (ticket !== version || !visible.value || !globalStore.isLogin || !canManageCredential.value) return;
            created.value = response.data.item;
            secret.value = response.data.token || '';
            uncertain.value = false;
        }
        emit('changed');
    } catch (error) {
        if (ticket !== version) return;
        if (isMCPRevisionConflict(error)) {
            conflict.value = true;
            policyConflict.value = isMCPPolicyChanged(error);
            emit('changed');
        } else if (!item.value) {
            uncertain.value = isMCPMutationUncertain(error);
            if (!uncertain.value) pendingCreate = undefined;
        }
    } finally {
        if (ticket === version) busy.value = false;
    }
};
watch(
    [() => globalStore.isLogin, canManageCredential],
    ([loggedIn, canManage]) => {
        if (!loggedIn || !canManage) discard();
    },
    { flush: 'sync' },
);
onDeactivated(discard);
onBeforeUnmount(discard);
onBeforeRouteLeave(async (to) => {
    if (
        !globalStore.isLogin ||
        !canManageCredential.value ||
        ['entrance', 'Expired', 'EnterpriseLicenseRequired'].includes(String(to.name))
    ) {
        discard();
        return true;
    }
    if (busy.value) return false;
    if (secret.value && !saved.value) {
        await ElMessageBox.alert(i18n.global.t('panelMCP.saveSecret'), i18n.global.t('panelMCP.title'));
        return false;
    }
    discard();
    return true;
});
defineExpose({ open });
</script>

<style scoped lang="scss">
.mcp-tool-picker :deep(.el-checkbox__label) {
    min-width: 0;
    flex: 1;
    line-height: 1.5;
}
</style>
