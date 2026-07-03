<template>
    <DialogPro
        v-model="open"
        size="w-70"
        class="code-dialog !p-0"
        @opened="onOpen"
        :show-close="false"
        :fullscreen="isFullscreen"
    >
        <template #header>
            <div ref="dialogHeader" class="flex items-center justify-between code-header px-4 rounded-t">
                <div class="code-title">
                    <span class="truncate-text">{{ $t('home.dir') + ' - ' + currentEditorPath }}</span>
                    <el-tooltip v-if="currentEditorPath" :content="$t('file.copyDir')" placement="top">
                        <CopyButton class="code-title-copy" :content="currentEditorPath" />
                    </el-tooltip>
                </div>
                <el-space alignment="center" :size="1" class="dialog-header-icon">
                    <el-tooltip :content="loadTooltip()" placement="top">
                        <el-button
                            @click="toggleFullscreen"
                            v-if="!isMobile"
                            class="!border-none !bg-transparent !text-base !font-semibold !py-2 !px-1 mr-2.5"
                            icon="FullScreen"
                        ></el-button>
                    </el-tooltip>
                    <el-button
                        @click="handleClose"
                        class="!border-none !bg-transparent !text-xl !py-2 !px-1"
                        icon="Close"
                    ></el-button>
                </el-space>
            </div>
        </template>
        <template #content>
            <div ref="dialogForm" class="px-4 py-2 code-action">
                <div class="flex justify-start items-center gap-x-4 card-action">
                    <el-text class="cursor-pointer" @click="handleReset">{{ $t('commons.button.reset') }}</el-text>
                    <el-divider direction="vertical" class="!mx-0" />
                    <el-text v-permission v-node-admin class="cursor-pointer ml-0" @click="saveContent()">
                        {{ $t('commons.button.save') }}
                    </el-text>
                    <el-divider direction="vertical" class="!mx-0" />
                    <el-dropdown trigger="click" max-height="300" placement="bottom-start" @command="changeTheme">
                        <span class="el-dropdown-link cursor-pointer">{{ $t('file.theme') }}</span>
                        <template #dropdown>
                            <el-dropdown-menu>
                                <el-dropdown-item v-for="item in themes" :key="item.label" :command="item.value">
                                    <div class="flex items-center justify-between gap-4 w-full">
                                        {{ item.label }}
                                        <el-icon v-if="config.theme == item.value"><Check /></el-icon>
                                    </div>
                                </el-dropdown-item>
                            </el-dropdown-menu>
                        </template>
                    </el-dropdown>
                    <el-divider direction="vertical" class="!mx-0" />
                    <el-dropdown trigger="click" max-height="300" placement="bottom-start" @command="changeLanguage">
                        <span class="el-dropdown-link cursor-pointer">{{ $t('file.language') }}</span>
                        <template #dropdown>
                            <el-dropdown-menu>
                                <el-dropdown-item v-for="item in Languages" :key="item.label" :command="item.label">
                                    <div class="flex items-center justify-between gap-4 w-full">
                                        {{ item.label }}
                                        <el-icon v-if="config.language == item.label"><Check /></el-icon>
                                    </div>
                                </el-dropdown-item>
                            </el-dropdown-menu>
                        </template>
                    </el-dropdown>
                    <el-divider direction="vertical" class="!mx-0" />
                    <el-dropdown trigger="click" max-height="300" placement="bottom-start" @command="changeEOL">
                        <span class="el-dropdown-link cursor-pointer">{{ $t('file.eol') }}</span>
                        <template #dropdown>
                            <el-dropdown-menu>
                                <el-dropdown-item v-for="item in eols" :key="item.label" :command="item.value">
                                    <div class="flex items-center justify-between gap-4 w-full">
                                        {{ item.label }}
                                        <el-icon v-if="config.eol == item.value"><Check /></el-icon>
                                    </div>
                                </el-dropdown-item>
                            </el-dropdown-menu>
                        </template>
                    </el-dropdown>
                    <el-divider direction="vertical" class="!mx-0" />
                    <el-dropdown trigger="click" max-height="300" placement="bottom-start">
                        <span class="el-dropdown-link cursor-pointer">{{ $t('file.setting') }}</span>
                        <template #dropdown>
                            <el-dropdown-menu>
                                <el-dropdown-item @click="changeMinimap(!config.minimap)">
                                    <div class="flex items-center justify-between gap-4 w-full">
                                        {{ $t('file.minimap') }}
                                        <el-icon v-if="config.minimap"><Check /></el-icon>
                                    </div>
                                </el-dropdown-item>
                                <el-dropdown-item @click="changeWarp(config.wordWrap)">
                                    <div class="flex items-center justify-between gap-4 w-full">
                                        {{ $t('file.wordWrap') }}
                                        <el-icon v-if="config.wordWrap == 'on'"><Check /></el-icon>
                                    </div>
                                </el-dropdown-item>
                            </el-dropdown-menu>
                        </template>
                    </el-dropdown>
                </div>
            </div>
            <div v-loading="loading">
                <el-splitter
                    class="code-splitter"
                    :style="{ height: splitterHeight }"
                    layout="horizontal"
                    lazy
                    @collapse="handleSplitterCollapse"
                    @resize-end="handleSplitterResizeEnd"
                >
                    <el-splitter-panel
                        v-model:size="treePanelSize"
                        :min="isShow ? minTreePanelSize : 0"
                        :max="maxTreePanelSize"
                        :resizable="isShow"
                        collapsible
                        class="code-tree-panel"
                        :class="{ 'is-collapsed': !isShow }"
                        @update:size="handleTreePanelSizeChange"
                    >
                        <div v-show="isShow" class="monaco-editor monaco-editor-background border-0 tree-container">
                            <div class="flex items-center justify-between px-1 h-7">
                                <el-text size="small" @click="getUpData()" class="cursor-pointer">
                                    <el-icon>
                                        <Top />
                                    </el-icon>
                                    <span class="sm:inline hidden pl-1">{{ $t('file.up') }}</span>
                                </el-text>
                                <el-divider direction="vertical" class="!mx-0" />
                                <el-text size="small" @click="getRefresh(directoryPath)" class="cursor-pointer">
                                    <el-icon>
                                        <Refresh />
                                    </el-icon>
                                    <span class="sm:inline hidden pl-1">{{ $t('commons.button.refresh') }}</span>
                                </el-text>
                                <el-divider direction="vertical" v-if="!isMobile" class="!mx-0" />
                                <el-dropdown @command="handleCreate" v-if="!isMobile" trigger="click">
                                    <el-text size="small">
                                        {{ $t('commons.button.create') }}
                                        <el-icon><arrow-down /></el-icon>
                                    </el-text>
                                    <template #dropdown>
                                        <el-dropdown-menu>
                                            <fu-dropdown-item v-permission v-node-admin command="dir" class="!px-2">
                                                <svg-icon class="!w-5 !h-5" iconName="p-file-folder"></svg-icon>
                                                {{ $t('file.dir') }}
                                            </fu-dropdown-item>
                                            <fu-dropdown-item v-permission v-node-admin command="file" class="!px-2">
                                                <svg-icon class="!w-5 !h-5" iconName="p-file-normal"></svg-icon>
                                                {{ $t('menu.files') }}
                                            </fu-dropdown-item>
                                        </el-dropdown-menu>
                                    </template>
                                </el-dropdown>
                            </div>
                            <el-divider class="!my-0" />
                            <el-tree-v2
                                ref="treeRef"
                                :data="treeData"
                                :props="treeProps"
                                @node-expand="handleNodeExpand"
                                @node-collapse="handleNodeCollapse"
                                @node-click="closeTreeContextMenu"
                                @node-contextmenu="openTreeContextMenu"
                                class="monaco-editor-tree monaco-editor-background pt-2"
                                :default-expanded-keys="expandedNodeKeys"
                                :height="treeHeight"
                                :indent="6"
                                :item-size="26"
                                highlight-current
                            >
                                <template #default="{ node, data }">
                                    <span v-if="data.isDir" class="tree-node-content">
                                        <template v-if="isCreate == 'dir' && data.id == 'new-dir'">
                                            <div class="tree-node-editing">
                                                <svg-icon class="table-icon" iconName="p-file-folder"></svg-icon>
                                                <el-input
                                                    size="small"
                                                    class="!flex-1 !min-w-0"
                                                    ref="rowRefs"
                                                    v-model="newFolder"
                                                ></el-input>
                                                <el-icon
                                                    class="cursor-pointer w-4 pl-1"
                                                    size="small"
                                                    @click.stop="createFolder(true)"
                                                >
                                                    <Check />
                                                </el-icon>
                                                <el-icon
                                                    class="cursor-pointer w-4"
                                                    size="small"
                                                    @click.stop="cancelFolder()"
                                                >
                                                    <Close />
                                                </el-icon>
                                            </div>
                                        </template>
                                        <template v-else>
                                            <svg-icon class="table-icon" iconName="p-file-folder"></svg-icon>
                                            <small :title="node.label" class="tree-node-label">{{ node.label }}</small>
                                        </template>
                                    </span>
                                    <span v-else class="tree-node-content" @click="getContent(data.path)">
                                        <template v-if="isCreate == 'file' && data.id == 'new-file'">
                                            <div class="tree-node-editing">
                                                <svg-icon
                                                    class="table-icon"
                                                    :iconName="getIconName(data.extension)"
                                                ></svg-icon>
                                                <el-input
                                                    size="small"
                                                    ref="rowRefs"
                                                    class="!flex-1 !min-w-0"
                                                    v-model="newFolder"
                                                ></el-input>
                                                <el-icon
                                                    class="cursor-pointer w-4 pl-1"
                                                    size="small"
                                                    @click.stop="createFolder(false)"
                                                >
                                                    <Check />
                                                </el-icon>
                                                <el-icon
                                                    class="cursor-pointer w-4"
                                                    size="small"
                                                    @click.stop="cancelFolder()"
                                                >
                                                    <Close />
                                                </el-icon>
                                            </div>
                                        </template>
                                        <template v-else>
                                            <svg-icon
                                                class="table-icon"
                                                :iconName="getIconName(data.extension)"
                                            ></svg-icon>
                                            <small :title="node.label" class="tree-node-label">{{ node.label }}</small>
                                        </template>
                                    </span>
                                </template>
                            </el-tree-v2>
                            <div
                                v-if="treeContextMenu.visible"
                                class="tree-context-menu"
                                :style="{ left: `${treeContextMenu.x}px`, top: `${treeContextMenu.y}px` }"
                                @click.stop
                                @contextmenu.prevent
                            >
                                <div
                                    v-if="treeContextMenu.data?.isDir"
                                    v-permission
                                    v-node-admin
                                    class="tree-context-menu__item"
                                    @click="createFromContextMenu('dir')"
                                >
                                    <svg-icon class="tree-context-menu__icon" iconName="p-file-folder"></svg-icon>
                                    <span>{{ $t('file.dir') }}</span>
                                </div>
                                <div
                                    v-if="treeContextMenu.data?.isDir"
                                    v-permission
                                    v-node-admin
                                    class="tree-context-menu__item"
                                    @click="createFromContextMenu('file')"
                                >
                                    <svg-icon class="tree-context-menu__icon" iconName="p-file-normal"></svg-icon>
                                    <span>{{ $t('menu.files') }}</span>
                                </div>
                                <div class="tree-context-menu__item" @click="copyPathFromContextMenu">
                                    <el-icon class="tree-context-menu__icon"><CopyDocument /></el-icon>
                                    <span>{{ $t('file.copyDir') }}</span>
                                </div>
                                <div
                                    v-permission
                                    v-node-admin
                                    class="tree-context-menu__item"
                                    @click="renameFromContextMenu"
                                >
                                    <el-icon class="tree-context-menu__icon"><Edit /></el-icon>
                                    <span>{{ $t('file.rename') }}</span>
                                </div>
                                <div
                                    v-permission
                                    v-node-admin
                                    class="tree-context-menu__item is-danger"
                                    @click="deleteFromContextMenu"
                                >
                                    <el-icon class="tree-context-menu__icon"><Delete /></el-icon>
                                    <span>{{ $t('commons.button.delete') }}</span>
                                </div>
                            </div>
                        </div>
                    </el-splitter-panel>
                    <el-splitter-panel min="240" class="code-editor-panel">
                        <div class="code-editor-panel__inner relative">
                            <CodeTabs
                                class="monaco-editor monaco-editor-background"
                                :select-tab="selectTab"
                                :file-tabs="fileTabs"
                                :on-remove-tab="removeTab"
                                :on-change-tab="changeTab"
                                :on-remove-all-tab="removeAllTab"
                                :on-remove-other-tab="removeOtherTab"
                            ></CodeTabs>
                            <div ref="codeBox" class="code-box relative">
                                <div class="flex justify-center items-center h-full" v-if="fileTabs.length === 0">
                                    <el-empty :image="noUpdateImage" />
                                </div>
                            </div>
                        </div>
                    </el-splitter-panel>
                </el-splitter>
                <div
                    class="hidden code-footer pl-4 h-7 sm:flex justify-end items-center gap-4 rounded-b"
                    ref="dialogFooter"
                >
                    <el-divider direction="vertical" class="!h-6" v-if="config.theme" />
                    <el-dropdown trigger="click" max-height="300" placement="top" @command="changeTheme">
                        <span class="el-dropdown-link">
                            {{ themes.find((item) => item.value === config.theme)?.label || $t('file.theme') }}
                        </span>
                        <template #dropdown>
                            <el-dropdown-menu>
                                <el-dropdown-item v-for="item in themes" :key="item.label" :command="item.value">
                                    <div class="flex items-center justify-between gap-4 w-full">
                                        {{ item.label }}
                                        <el-icon v-if="config.theme == item.value"><Check /></el-icon>
                                    </div>
                                </el-dropdown-item>
                            </el-dropdown-menu>
                        </template>
                    </el-dropdown>
                    <el-divider direction="vertical" class="!h-6" />
                    <el-dropdown trigger="click" max-height="300" placement="top" @command="changeEOL">
                        <span class="el-dropdown-link">
                            {{ eols.find((item) => item.value === config.eol)?.label || $t('file.eol') }}
                        </span>
                        <template #dropdown>
                            <el-dropdown-menu>
                                <el-dropdown-item v-for="item in eols" :key="item.label" :command="item.value">
                                    <div class="flex items-center justify-between gap-4 w-full">
                                        {{ item.label }}
                                        <el-icon v-if="config.eol == item.value"><Check /></el-icon>
                                    </div>
                                </el-dropdown-item>
                            </el-dropdown-menu>
                        </template>
                    </el-dropdown>
                    <el-divider direction="vertical" class="!h-6" />
                    <el-text class="cursor-pointer inline-flex items-center gap-1" @click="openHistoryDrawer">
                        <span class="el-dropdown-link">{{ $t('file.history') }} ({{ historyVersionCount }})</span>
                    </el-text>
                    <el-divider direction="vertical" class="!h-6" />
                    <el-dropdown trigger="click" max-height="300" placement="top" @command="changeLanguage">
                        <span class="el-dropdown-link">
                            {{
                                config.language
                                    ? `${$t('file.language')}: ${
                                          Languages.find((item) => item.label === config.language)?.label ||
                                          config.language
                                      }`
                                    : $t('file.language')
                            }}
                        </span>
                        <template #dropdown>
                            <el-dropdown-menu>
                                <el-dropdown-item v-for="item in Languages" :key="item.label" :command="item.label">
                                    <div class="flex items-center justify-between gap-4 w-full">
                                        {{ item.label }}
                                        <el-icon v-if="config.language == item.label"><Check /></el-icon>
                                    </div>
                                </el-dropdown-item>
                            </el-dropdown-menu>
                        </template>
                    </el-dropdown>
                    <el-divider direction="vertical" class="!h-6" />
                    <el-dropdown trigger="click" max-height="300" placement="top">
                        <span class="el-dropdown-link">
                            {{
                                $t('file.wordWrap') +
                                ': ' +
                                $t(config.wordWrap === 'on' ? 'commons.button.enable' : 'commons.button.disable')
                            }}
                        </span>
                        <template #dropdown>
                            <el-dropdown-menu>
                                <el-dropdown-item @click="changeWarp('off')">
                                    <div class="flex items-center justify-between gap-4 w-full">
                                        {{ $t('commons.button.enable') }}
                                        <el-icon v-if="config.wordWrap == 'on'"><Check /></el-icon>
                                    </div>
                                </el-dropdown-item>
                                <el-dropdown-item @click="changeWarp('on')">
                                    <div class="flex items-center justify-between gap-4 w-full">
                                        {{ $t('commons.button.disable') }}
                                        <el-icon v-if="config.wordWrap == 'off'"><Check /></el-icon>
                                    </div>
                                </el-dropdown-item>
                            </el-dropdown-menu>
                        </template>
                    </el-dropdown>
                    <el-divider direction="vertical" class="!h-6" />
                    <el-dropdown trigger="click" max-height="300" placement="top">
                        <span class="el-dropdown-link">
                            {{
                                $t('file.minimap') +
                                ': ' +
                                $t(config.minimap ? 'commons.button.enable' : 'commons.button.disable')
                            }}
                        </span>
                        <template #dropdown>
                            <el-dropdown-menu>
                                <el-dropdown-item @click="changeMinimap(true)">
                                    <div class="flex items-center justify-between gap-4 w-full">
                                        {{ $t('commons.button.enable') }}
                                        <el-icon v-if="config.minimap"><Check /></el-icon>
                                    </div>
                                </el-dropdown-item>
                                <el-dropdown-item @click="changeMinimap(false)">
                                    <div class="flex items-center justify-between gap-4 w-full">
                                        {{ $t('commons.button.disable') }}
                                        <el-icon v-if="!config.minimap"><Check /></el-icon>
                                    </div>
                                </el-dropdown-item>
                            </el-dropdown-menu>
                        </template>
                    </el-dropdown>
                    <el-divider direction="vertical" class="!h-6 !mr-3.5" />
                </div>
            </div>
        </template>
    </DialogPro>
    <FileHistoryDrawer ref="historyDrawerRef" @restored="handleHistoryRestored" />
</template>

<script lang="ts" setup>
import {
    batchCheckFiles,
    createFile,
    deleteFile,
    getFileContent,
    getFilesTree,
    renameRile,
    saveFileContent,
    searchFileHistory,
} from '@/api/modules/files';
import i18n from '@/lang';
import { MsgError, MsgSuccess, MsgWarning } from '@/utils/message';
import { loadMonacoLanguageSupport, setupMonacoEnvironment } from '@/utils/monaco';
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue';
import { Languages } from '@/global/mimetype';
import { resolveEditorLanguage } from '@/utils/file';
import { hasManagePermissionAccess } from '@/utils/permission';

import type { TabPaneName } from 'element-plus';
import { ElMessageBox, ElTreeV2 } from 'element-plus';
import { ResultData } from '@/api/interface';
import { File } from '@/api/interface/file';
import { copyText } from '@/utils/clipboard';
import { getIcon } from '@/utils/file';
import { newUUID } from '@/utils/id';
import { TreeNodeData } from 'element-plus/es/components/tree-v2/src/types';
import { CopyDocument, Delete, Edit, Refresh, Top } from '@element-plus/icons-vue';
import { loadBaseDir } from '@/api/modules/setting';
import CodeTabs from './tabs/index.vue';
import FileHistoryDrawer from './history/index.vue';
import noUpdateImage from '@/assets/images/no_update_app.svg';
import { useGlobalStore } from '@/composables/useGlobalStore';
import {
    CodeEditorTheme,
    codeEditorThemeStorageKey,
    getDefaultCodeEditorTheme,
    resolveCodeEditorTheme,
} from '@/utils/code-editor-theme';

const { isDarkTheme, isMobile } = useGlobalStore();

type MonacoEditorApi = typeof import('monaco-editor/esm/vs/editor/editor.api');

let monacoApi: MonacoEditorApi | null = null;
let monacoThemeInitialized = false;
let editor: ReturnType<MonacoEditorApi['editor']['create']> | undefined;

const eolLf = ref(0);
const eolCrlf = ref(1);

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
    eolLf.value = monacoApi.editor.EndOfLineSequence.LF;
    eolCrlf.value = monacoApi.editor.EndOfLineSequence.CRLF;

    if (!monacoThemeInitialized) {
        monacoApi.editor.defineTheme('vs', {
            base: 'vs',
            inherit: true,
            rules: [{ token: '' }],
            colors: {
                'editor.background': '#f8f6f6',
                'minimap.background': '#f4f4f4',
                'scrollbar.shadow': '#e1e1e1',
                'scrollbarSlider.background': '#e1e1e1',
                'scrollbarSlider.hoverBackground': '#cccccc',
                'scrollbarSlider.activeBackground': '#bfbfbf',
            },
        });
        monacoThemeInitialized = true;
    }

    return monacoApi;
};

const disposeEditor = () => {
    if (!editor) {
        return;
    }
    clearPendingLineHighlight();
    editor.dispose();
    editor = undefined;
};
interface EditProps {
    language: string;
    content: string;
    path: string;
    name: string;
    extension: string;
    initialLine?: number;
}

interface EditorConfig {
    theme: CodeEditorTheme;
    language: string;
    eol: number;
    wordWrap: WordWrapOptions;
    minimap: boolean;
}

const pendingInitialLine = ref(0);
let lineHighlightDecorationIds: string[] = [];

const clearPendingLineHighlight = () => {
    if (!editor) {
        lineHighlightDecorationIds = [];
        return;
    }
    lineHighlightDecorationIds = editor.deltaDecorations(lineHighlightDecorationIds, []);
};

const revealPendingInitialLine = () => {
    const line = pendingInitialLine.value;
    if (!editor || !monacoApi || line < 1) {
        return;
    }
    const model = editor.getModel();
    if (!model) {
        return;
    }
    const targetLine = Math.min(line, model.getLineCount());
    editor.setSelection({
        startLineNumber: targetLine,
        startColumn: 1,
        endLineNumber: targetLine,
        endColumn: 1,
    });
    editor.setPosition({ lineNumber: targetLine, column: 1 });
    editor.revealLineInCenter(targetLine);
    lineHighlightDecorationIds = editor.deltaDecorations(lineHighlightDecorationIds, [
        {
            range: new monacoApi.Range(targetLine, 1, targetLine, model.getLineMaxColumn(targetLine)),
            options: {
                isWholeLine: true,
                className: 'ai-search-target-line',
                linesDecorationsClassName: 'ai-search-target-line-gutter',
            },
        },
    ]);
    editor.focus();
    pendingInitialLine.value = 0;
};

const open = ref(false);
const loading = ref(false);
const fileName = ref('');
const codeThemeKey = codeEditorThemeStorageKey;
const warpKey = 'code-warp';
const minimapKey = 'code-minimap';
const directoryPath = ref('');
const fileExtension = ref('');
const baseDir = ref();
const treeData = ref([]);
const codeBox = ref();
const defaultHeight = ref(56);
const treeHeight = ref(0);
const splitterHeight = ref('56vh');
const codeReq = reactive({ path: '', expand: false, page: 1, pageSize: 100 });
const isShow = ref(true);
const defaultTreePanelSize = 220;
const minTreePanelSize = 160;
const maxTreePanelSize = 420;
const treePanelSize = ref(defaultTreePanelSize);
const lastTreePanelSize = ref(defaultTreePanelSize);
const isEdit = ref(false);
const oldFileContent = ref('');
const dialogHeader = ref(null);
const dialogForm = ref(null);
const dialogFooter = ref(null);
const historyDrawerRef = ref();
const currentPath = ref();
const historyVersionCount = ref(0);
const rowRefs = ref();
const isCreate = ref('none');
const newFolder = ref();
const selectedParentNode = ref(null);
const expandedNodeIds = ref<Set<string>>(new Set());
const expandedNodeKeys = computed<string[]>(() => Array.from(expandedNodeIds.value));

const addExpandedNode = (id: string) => {
    expandedNodeIds.value.add(id);
};

const removeExpandedNode = (id: string) => {
    expandedNodeIds.value.delete(id);
};

const resetExpandedNodes = () => {
    expandedNodeIds.value = new Set<string>();
};

const refreshEditorLayout = () => {
    nextTick(() => {
        editor?.layout();
    });
};

const syncTreePanelState = (size: number) => {
    const nextSize = Math.max(size, 0);
    treePanelSize.value = nextSize;
    isShow.value = nextSize > 0;
    if (nextSize > 0) {
        lastTreePanelSize.value = nextSize;
    }
    closeTreeContextMenu();
    refreshEditorLayout();
};

const handleTreePanelSizeChange = (size: string | number) => {
    const nextSize = Number(size);
    if (Number.isNaN(nextSize)) {
        return;
    }
    syncTreePanelState(nextSize);
};

const handleSplitterResizeEnd = () => {
    refreshEditorLayout();
};

const handleSplitterCollapse = (index: number, type: 'start' | 'end', sizes: number[]) => {
    if (index !== 0 || !type) {
        return;
    }
    const nextSize = sizes[0] || 0;
    syncTreePanelState(nextSize > 0 ? nextSize : 0);
    if (!nextSize && lastTreePanelSize.value < minTreePanelSize) {
        lastTreePanelSize.value = defaultTreePanelSize;
    }
};

type WordWrapOptions = 'off' | 'on' | 'wordWrapColumn' | 'bounded';

const isFullscreen = ref(false);

const config = reactive<EditorConfig>({
    theme: 'vs-dark',
    language: 'plaintext',
    eol: eolLf.value,
    wordWrap: 'on',
    minimap: false,
});

const selectTab = ref();
const fileTabs = ref([]);
const maxTabs = 10;
const codeTabsStorageKey = 'code-editor-tabs';

const saveTabsToStorage = () => {
    const simpleTabs = fileTabs.value.map((tab: any) => ({
        path: tab.path,
        name: tab.name,
    }));
    localStorage.setItem(codeTabsStorageKey, JSON.stringify(simpleTabs.slice(0, maxTabs)));
};

const loadTabsFromStorage = (): { path: string; name: string }[] => {
    const raw = localStorage.getItem(codeTabsStorageKey);
    if (!raw) return [];
    const parsed = JSON.parse(raw);
    if (Array.isArray(parsed)) {
        return parsed.filter((item) => item && item.path && item.name);
    }
    return [];
};
const removeTab = (targetPath: TabPaneName) => {
    const tabs = fileTabs.value;
    let activeName = selectTab.value;

    const updateTabs = () => {
        if (activeName === targetPath) {
            const index = tabs.findIndex((tab) => tab.path === targetPath);
            const nextTab = tabs[index + 1] || tabs[index - 1];
            if (nextTab) {
                activeName = nextTab.path;
            }
        }
        selectTab.value = activeName;
        fileTabs.value = tabs.filter((tab) => tab.path !== targetPath);
        saveTabsToStorage();
        if (fileTabs.value.length === 0) {
            selectTab.value = '';
            disposeEditor();
        }
    };

    if (isEdit.value) {
        ElMessageBox.confirm(i18n.global.t('file.saveContentAndClose'), {
            confirmButtonText: i18n.global.t('commons.button.save'),
            cancelButtonText: i18n.global.t('commons.button.notSave'),
            type: 'info',
            distinguishCancelAndClose: true,
        })
            .then(() => {
                updateTabs();
                if (fileTabs.value.length > 0) {
                    saveContent();
                    getContent(selectTab.value);
                }
            })
            .catch(() => {
                updateTabs();
                isEdit.value = false;
                if (fileTabs.value.length > 0) {
                    getContent(selectTab.value);
                }
            });
    } else {
        updateTabs();
        if (fileTabs.value.length > 0) {
            getContent(selectTab.value);
        }
    }
};

const removeAllTab = (targetPath: string, type: 'left' | 'right' | 'all') => {
    const tabs = fileTabs.value;
    const targetIndex = tabs.findIndex((tab) => tab.path === targetPath);
    let activeName = selectTab.value;

    const filterTabs = (): typeof fileTabs.value => {
        if (type === 'left') return tabs.slice(targetIndex);
        if (type === 'right') return tabs.slice(0, targetIndex + 1);
        return [];
    };

    const updateTabs = () => {
        if (activeName !== targetPath && type !== 'all') {
            activeName = tabs[targetIndex]?.path || '';
        }
        const newTabs = type === 'all' ? [] : filterTabs();
        fileTabs.value = newTabs;
        selectTab.value = activeName;
        saveTabsToStorage();

        if (type === 'all') {
            selectTab.value = '';
            disposeEditor();
        } else if (newTabs.length > 0) {
            getContent(activeName);
        }
    };

    const onConfirm = () => {
        updateTabs();
        saveContent();
    };

    const onCancel = () => {
        if (type === 'left' || type === 'right') {
            editor.setValue(oldFileContent.value);
        }
        isEdit.value = false;
        isCreate.value = 'none';
        updateTabs();
    };

    if (isEdit.value) {
        ElMessageBox.confirm(i18n.global.t('file.saveContentAndClose'), {
            confirmButtonText: i18n.global.t('commons.button.save'),
            cancelButtonText: i18n.global.t('commons.button.notSave'),
            type: 'info',
            distinguishCancelAndClose: true,
        })
            .then(onConfirm)
            .catch(onCancel);
    } else {
        updateTabs();
    }
};

const removeOtherTab = (targetPath: string) => {
    const tabs = fileTabs.value;
    const targetTab = tabs.find((tab) => tab.path === targetPath);
    if (!targetTab) return;

    const updateTabs = () => {
        fileTabs.value = [targetTab];
        selectTab.value = targetTab.path;
        saveTabsToStorage();
        getContent(targetTab.path);
    };

    const onConfirm = () => {
        updateTabs();
        saveContent();
    };

    const onCancel = () => {
        editor.setValue(oldFileContent.value);
        isEdit.value = false;
        updateTabs();
    };

    if (isEdit.value) {
        ElMessageBox.confirm(i18n.global.t('file.saveContentAndClose'), {
            confirmButtonText: i18n.global.t('commons.button.save'),
            cancelButtonText: i18n.global.t('commons.button.notSave'),
            type: 'info',
            distinguishCancelAndClose: true,
        })
            .then(onConfirm)
            .catch(onCancel);
    } else {
        updateTabs();
    }
};

const changeTab = (targetPath: TabPaneName) => {
    selectTab.value = targetPath.toString();
    getContent(targetPath.toString());
};

const eols = computed(() => [
    {
        label: 'LF (Linux)',
        value: eolLf.value,
    },
    {
        label: 'CRLF (Windows)',
        value: eolCrlf.value,
    },
]);

const themes = [
    {
        label: 'Visual Studio',
        value: 'vs',
    },
    {
        label: 'Visual Studio Dark',
        value: 'vs-dark',
    },
    {
        label: 'High Contrast Dark',
        value: 'hc-black',
    },
];

let form = ref({
    content: '',
    path: '',
});
const currentEditorPath = computed(() => form.value.path || directoryPath.value || '');

const em = defineEmits(['close']);

const handleClose = () => {
    const closeEditor = () => {
        open.value = false;
        selectTab.value = '';
        fileTabs.value = [];
        isEdit.value = false;
        if (editor) {
            disposeEditor();
        }
        em('close', open.value);
    };

    if (isEdit.value) {
        ElMessageBox.confirm(i18n.global.t('file.saveContentAndClose'), {
            confirmButtonText: i18n.global.t('commons.button.save'),
            cancelButtonText: i18n.global.t('commons.button.notSave'),
            type: 'info',
            distinguishCancelAndClose: true,
        })
            .then(() => {
                saveContent();
            })
            .finally(() => {
                closeEditor();
            });
    } else {
        closeEditor();
    }
};

const handleReset = () => {
    if (isEdit.value) {
        loading.value = true;
        form.value.content = oldFileContent.value;
        editor.setValue(oldFileContent.value);
        isEdit.value = false;
        MsgSuccess(i18n.global.t('commons.msg.resetSuccess'));
        loading.value = false;
    } else {
        MsgWarning(i18n.global.t('file.noEdit'));
    }
};

const loadTooltip = () => {
    return i18n.global.t('commons.button.' + (isFullscreen.value ? 'quitFullscreen' : 'fullscreen'));
};

onMounted(() => {
    isCreate.value = 'none';
    loadPath();
    updateHeights();
    window.addEventListener('resize', updateHeights);
    document.addEventListener('click', closeTreeContextMenu);
    window.addEventListener('scroll', closeTreeContextMenu, true);
});

const updateHeights = () => {
    const vh = window.innerHeight / 100;
    if (isFullscreen.value) {
        let paddingHeight = 30;
        const headerHeight = dialogHeader.value.offsetHeight;
        const formHeight = dialogForm.value.offsetHeight;
        const footerHeight = dialogFooter.value.offsetHeight;
        const contentHeight = window.innerHeight - headerHeight - formHeight - footerHeight - paddingHeight;
        treeHeight.value = contentHeight - 31;
        splitterHeight.value = `${contentHeight}px`;
    } else {
        splitterHeight.value = `${defaultHeight.value}vh`;
        treeHeight.value = defaultHeight.value * vh - 31;
    }
    refreshEditorLayout();
};

const toggleFullscreen = () => {
    isFullscreen.value = !isFullscreen.value;
    updateHeights();
};

const changeLanguage = (command: string) => {
    if (!editor || !monacoApi) {
        return;
    }
    config.language = command;
    const model = editor.getModel();
    if (!model) {
        return;
    }
    monacoApi.editor.setModelLanguage(model, config.language);
};

const applyCodeEditorTheme = (theme: CodeEditorTheme, persist = false) => {
    config.theme = theme;
    if (!monacoApi) {
        return;
    }
    monacoApi.editor.setTheme(config.theme);
    applyTreeThemeClass();
    if (persist) {
        localStorage.setItem(codeThemeKey, config.theme);
    }
};

const changeTheme = (command: string) => {
    applyCodeEditorTheme(resolveCodeEditorTheme(command, isDarkTheme.value), true);
};

const applyTreeThemeClass = () => {
    const themes: Record<CodeEditorTheme, string> = {
        vs: 'monaco-editor-tree-light',
        'vs-dark': 'monaco-editor-tree-dark',
        'hc-black': 'monaco-editor-tree-dark',
    };

    if (!treeRef.value) {
        return;
    }
    Object.values(themes).forEach((themeClass) => {
        treeRef.value.$el.classList.remove(themeClass);
    });
    treeRef.value.$el.classList.add(themes[config.theme]);
};

const syncDefaultThemeWithPanelTheme = () => {
    const nextTheme = getDefaultCodeEditorTheme(isDarkTheme.value);
    localStorage.removeItem(codeThemeKey);
    applyCodeEditorTheme(nextTheme);
};

watch(isDarkTheme, syncDefaultThemeWithPanelTheme);

const changeEOL = (command: number) => {
    if (!editor) {
        return;
    }
    config.eol = command;
    editor.getModel()?.pushEOL(config.eol);
};

const changeWarp = (command: string) => {
    config.wordWrap = command === 'on' ? 'off' : 'on';
    localStorage.setItem(warpKey, config.wordWrap);
    editor.updateOptions({
        wordWrap: config.wordWrap,
    });
};

const changeMinimap = (command: boolean) => {
    config.minimap = command;
    localStorage.setItem(minimapKey, JSON.stringify(config.minimap));
    editor.updateOptions({
        minimap: {
            enabled: config.minimap,
        },
    });
};

const initEditor = async () => {
    const monaco = await ensureMonaco();

    disposeEditor();
    await nextTick();

    editor = monaco.editor.create(codeBox.value as HTMLElement, {
        theme: config.theme,
        value: form.value.content,
        readOnly: false,
        automaticLayout: true,
        language: config.language,
        folding: true,
        roundedSelection: false,
        overviewRulerBorder: false,
        wordWrap: config.wordWrap,
        minimap: {
            enabled: config.minimap,
        },
        lineNumbersMinChars: 6,
    });
    if (editor.getModel()?.getValue() === '') {
        editor.getModel()?.setValue('');
    }

    editor.getModel()?.pushEOL(config.eol);
    applyTreeThemeClass();

    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, quickSave);
    editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Slash, quickToggleComment);
    editor.focus();

    editor.onDidChangeModelContent(() => {
        if (editor) {
            form.value.content = editor.getValue();
            isEdit.value = true;
        }
    });

    revealPendingInitialLine();
};

const quickSave = () => {
    saveContent();
};

const quickToggleComment = () => {
    void editor?.getAction('editor.action.commentLine')?.run();
};

const openHistoryDrawer = () => {
    if (!form.value.path) {
        MsgWarning(i18n.global.t('file.historyNeedFile'));
        return;
    }
    if (!historyDrawerRef.value) {
        return;
    }
    historyDrawerRef.value.acceptParams({
        path: form.value.path,
        content: form.value.content,
        language: config.language,
        extension: fileExtension.value,
        dirty: isEdit.value,
    });
};

const loadHistoryVersionCount = async (path: string) => {
    if (!path) {
        historyVersionCount.value = 0;
        return;
    }

    try {
        const res = await searchFileHistory({
            page: 1,
            pageSize: 1,
            path,
            scope: 'current',
        });
        if (form.value.path === path) {
            historyVersionCount.value = res.data?.total || 0;
        }
    } catch {
        if (form.value.path === path) {
            historyVersionCount.value = 0;
        }
    }
};

const saveContent = async () => {
    if (!hasManagePermissionAccess(undefined, { nodeAdmin: true })) {
        MsgError(i18n.global.t('commons.res.forbidden'));
        return;
    }
    if (isEdit.value) {
        loading.value = true;
        try {
            const res = await saveFileContent(form.value);
            if (res) {
                isEdit.value = false;
                oldFileContent.value = form.value.content;
                await loadHistoryVersionCount(form.value.path);
                MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            } else {
                MsgError(i18n.global.t('commons.status.failed'));
                isEdit.value = false;
            }
        } finally {
            loading.value = false;
        }
    } else {
        MsgWarning(i18n.global.t('file.noEdit'));
    }
};

const acceptParams = async (props: EditProps) => {
    const monaco = await ensureMonaco();

    pendingInitialLine.value = props.initialLine && props.initialLine > 0 ? Math.floor(props.initialLine) : 0;
    form.value.content = props.content;
    oldFileContent.value = props.content;
    form.value.path = props.path;
    historyVersionCount.value = 0;
    currentPath.value = getDirectoryPath(props.path);
    directoryPath.value = getDirectoryPath(props.path);
    fileExtension.value = props.extension;
    fileName.value = props.name;
    config.language = resolveEditorLanguage(props.path, props.extension, props.name);

    let savedTabs = loadTabsFromStorage();
    const withoutCurrent = savedTabs.filter((tab) => tab.path !== props.path);
    if (withoutCurrent.length > 0) {
        try {
            const existRes = await batchCheckFiles(withoutCurrent.map((t) => t.path));
            const existList = existRes?.data ?? [];
            const existingPaths = new Set(existList.map((r) => r.path));
            savedTabs = withoutCurrent.filter((t) => existingPaths.has(t.path));
        } catch {
            savedTabs = withoutCurrent;
        }
    } else {
        savedTabs = [];
    }
    const merged = [...savedTabs, { path: props.path, name: props.name }];
    fileTabs.value = merged.slice(-maxTabs);
    selectTab.value = props.path;

    if (props.language) {
        config.language = props.language;
    }
    config.eol = monaco.editor.EndOfLineSequence.LF;
    config.theme = resolveCodeEditorTheme(localStorage.getItem(codeThemeKey), isDarkTheme.value);
    config.wordWrap = (localStorage.getItem(warpKey) as WordWrapOptions) || 'on';
    config.minimap = localStorage.getItem(minimapKey) !== null ? localStorage.getItem(minimapKey) === 'true' : true;
    open.value = true;
    saveTabsToStorage();
    loadHistoryVersionCount(props.path);
    nextTick(() => {
        if (editor) {
            editor.setValue(form.value.content);
            const model = editor.getModel();
            if (model) {
                monacoApi?.editor.setModelLanguage(model, config.language);
            }
            isEdit.value = false;
            revealPendingInitialLine();
        }
    });
};

const getIconName = (extension: string) => getIcon(extension);

const loadPath = async () => {
    const pathRes = await loadBaseDir();
    baseDir.value = pathRes.data;
};

const getDirectoryPath = (filePath: string) => {
    if (!filePath) {
        return baseDir.value;
    }

    const lastSlashIndex = filePath.lastIndexOf('/');

    if (lastSlashIndex === -1) {
        return baseDir.value;
    }

    const directoryPath = filePath.substring(0, lastSlashIndex);
    if (directoryPath === '' || directoryPath === '.' || directoryPath === '/') {
        return baseDir.value;
    }
    return directoryPath;
};

const onOpen = async () => {
    await initEditor();
    applyCodeEditorTheme(config.theme);
    search(directoryPath.value).then((res) => {
        handleSearchResult(res);
    });
};

const handleSearchResult = (res: ResultData<File.FileTree[]>) => {
    resetExpandedNodes();
    if (res.data.length > 0 && res.data[0].children) {
        treeData.value = res.data[0].children.map((item) => ({
            ...item,
            children: item.isDir ? item.children || [] : undefined,
        }));
    } else {
        treeData.value = [];
    }
};

const getRefresh = (path: string) => {
    loading.value = true;
    try {
        search(path).then((res) => {
            treeData.value = res.data[0].children;
            loadedNodes.value = new Set();
            resetExpandedNodes();
            isCreate.value = 'none';
            currentPath.value = path;
            selectedParentNode.value = null;
        });
    } finally {
        loading.value = false;
        MsgSuccess(i18n.global.t('commons.msg.refreshSuccess'));
    }
};

const getContent = (path: string, forceReload = false) => {
    if (!forceReload && (form.value.path === path || isCreate.value == 'file')) {
        return;
    }
    const existsInTabs = fileTabs.value.some((tab) => tab.path === path);
    if (!existsInTabs && fileTabs.value.length >= maxTabs) {
        fileTabs.value.shift();
    }
    const fetchFileContent = () => {
        codeReq.path = path;
        codeReq.expand = true;

        getFileContent(codeReq)
            .then((res) => {
                form.value.content = res.data.content;
                oldFileContent.value = res.data.content;
                form.value.path = res.data.path;
                fileExtension.value = res.data.extension;
                fileName.value = res.data.name;
                config.language = resolveEditorLanguage(res.data.path, res.data.extension, res.data.name);
                initEditor();
                const exists = fileTabs.value.some((tab) => tab.path === path);
                if (exists) {
                    const tab = fileTabs.value.find((t) => t.path === path);
                    if (tab) tab.name = res.data.name;
                } else {
                    fileTabs.value.push({
                        name: res.data.name,
                        path: res.data.path,
                    });
                }
                saveTabsToStorage();
                selectTab.value = res.data.path;
                loadHistoryVersionCount(res.data.path);
            })
            .catch(() => {});
    };

    if (isEdit.value) {
        ElMessageBox.confirm(i18n.global.t('file.saveAndOpenNewFile'), {
            confirmButtonText: i18n.global.t('commons.button.sure'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        })
            .then(() => {
                saveContent();
                fetchFileContent();
            })
            .catch(() => {
                selectTab.value = form.value.path;
            })
            .finally(() => {});
    } else {
        fetchFileContent();
    }
};

const handleHistoryRestored = (path: string) => {
    if (path === form.value.path) {
        getContent(path, true);
        loadHistoryVersionCount(path);
    }
};

const initTreeData = () => ({
    path: '/',
    expand: true,
    showHidden: true,
    page: 1,
    pageSize: 1000,
    search: '',
    containSub: true,
    dir: false,
    sortBy: 'name',
    sortOrder: 'ascending',
});

let req = reactive(initTreeData());

const loadedNodes = ref(new Set());

const search = async (path: string) => {
    req.path = path;
    if (req.search != '') {
        req.sortBy = 'name';
        req.sortOrder = 'ascending';
    }
    return await getFilesTree(req);
};

const getUpData = async () => {
    if ('/' === directoryPath.value) {
        MsgWarning(i18n.global.t('commons.msg.rootInfoErr'));
        return;
    }
    let pathParts = directoryPath.value.split('/');
    pathParts.pop();
    let newPath = pathParts.join('/') || '/';

    try {
        const response = await search(newPath);
        treeData.value = response.data[0]?.children || [];
        loadedNodes.value = new Set();
        resetExpandedNodes();
        isCreate.value = 'none';
        currentPath.value = newPath;
        selectedParentNode.value = null;
    } catch (error) {
    } finally {
        directoryPath.value = newPath;
    }
};

const treeRef = ref<InstanceType<typeof ElTreeV2>>();

const treeContextMenu = reactive<{
    visible: boolean;
    x: number;
    y: number;
    data: any | null;
    node: any | null;
}>({
    visible: false,
    x: 0,
    y: 0,
    data: null,
    node: null,
});

const treeProps = {
    value: 'id',
    label: 'name',
    children: 'children',
};

const closeTreeContextMenu = () => {
    treeContextMenu.visible = false;
    treeContextMenu.data = null;
    treeContextMenu.node = null;
};

const openTreeContextMenu = (event: MouseEvent, data: any, node: any) => {
    if (isMobile.value || data.id === 'new-dir' || data.id === 'new-file') {
        return;
    }
    event.preventDefault();
    event.stopPropagation();
    treeContextMenu.visible = true;
    treeContextMenu.x = event.clientX;
    treeContextMenu.y = event.clientY;
    treeContextMenu.data = data;
    treeContextMenu.node = node;
};

const handleNodeCollapse = (data: TreeNodeData, node: any) => {
    closeTreeContextMenu();
    isCreate.value = 'none';
    removeExpandedNode(data.id);

    const parentNode = node.parent;
    if (!parentNode) {
        selectedParentNode.value = null;
        return;
    }

    const hasExpandedChildren = parentNode.data.children?.some((child) => expandedNodeIds.value.has(child.id));

    if (hasExpandedChildren) {
        selectedParentNode.value = parentNode;
    } else {
        selectedParentNode.value = null;
    }
};

const handleNodeExpand = (data: TreeNodeData, node: any) => {
    closeTreeContextMenu();
    if (node.data.id == 'new-dir' || node.data.id == 'new-file') {
        return;
    }
    if (!node.data.isDir || loadedNodes.value.has(node.data.path)) {
        return;
    }
    if (node.data.isDir && isCreate.value == 'none') {
        currentPath.value = node.data.path;
        selectedParentNode.value = node;
        addExpandedNode(node.data.id);
    }
    search(data.path)
        .then((response) => {
            const newTreeData = JSON.parse(JSON.stringify(treeData.value));
            if (response.data.length > 0 && response.data[0].children) {
                node.children = response.data[0].children;
                loadedNodes.value.add(node.data.path);
                updateNodeChildren(newTreeData, node.data.path, response.data[0].children);
            } else {
                node.children = [];
            }
            treeData.value = newTreeData;
        })
        .catch(() => {});
};

const getChildItems = (res: ResultData<File.FileTree[]>) => {
    return res.data.length > 0 && res.data[0].children ? res.data[0].children : [];
};

const loadDirectoryChildren = async (data: any, node: any) => {
    currentPath.value = data.path;
    selectedParentNode.value = node;
    addExpandedNode(data.id);
    if (loadedNodes.value.has(data.path)) {
        return;
    }
    const response = await search(data.path);
    const children = getChildItems(response);
    node.children = children;
    node.data.children = children;
    updateNodeChildren(treeData.value, data.path, children);
    treeData.value = [...treeData.value];
    loadedNodes.value.add(data.path);
};

const createFromContextMenu = async (command: string) => {
    const data = treeContextMenu.data;
    const node = treeContextMenu.node;
    closeTreeContextMenu();
    if (!data?.isDir || !node) {
        return;
    }
    try {
        await loadDirectoryChildren(data, node);
        handleCreate(command);
    } catch {
        MsgError(i18n.global.t('commons.status.failed'));
    }
};

const copyPathFromContextMenu = () => {
    const path = treeContextMenu.data?.path;
    closeTreeContextMenu();
    if (!path) {
        return;
    }
    copyText(path);
};

const updateNodeChildren = (nodes: any[], path: any, newChildren: File.FileTree[]) => {
    const updateNode = (nodes: string | any[]) => {
        for (const element of nodes) {
            if (element.path === path) {
                element.children = newChildren;
                break;
            }
            if (element.children && element.children.length) {
                updateNode(element.children);
            }
        }
    };
    updateNode(nodes);
};

const joinPath = (dir: string, name: string) => {
    if (dir === '/') {
        return `/${name}`;
    }
    return `${dir}/${name}`;
};

const isPathAffected = (path: string, targetPath: string, targetIsDir: boolean) => {
    return path === targetPath || (targetIsDir && path.startsWith(`${targetPath}/`));
};

const replaceAffectedPath = (path: string, oldPath: string, newPath: string, isDir: boolean) => {
    if (path === oldPath) {
        return newPath;
    }
    if (isDir && path.startsWith(`${oldPath}/`)) {
        return `${newPath}${path.slice(oldPath.length)}`;
    }
    return path;
};

const renameTreeNodePaths = (nodes: any[], oldPath: string, newPath: string, newName: string, isDir: boolean) => {
    for (const node of nodes) {
        if (node.path === oldPath) {
            node.name = newName;
        }
        node.path = replaceAffectedPath(node.path, oldPath, newPath, isDir);
        if (node.children?.length) {
            renameTreeNodePaths(node.children, oldPath, newPath, newName, isDir);
        }
    }
};

const removeTreeNodeByPath = (nodes: any[], targetPath: string, targetIsDir: boolean) => {
    return nodes
        .filter((node) => !isPathAffected(node.path, targetPath, targetIsDir))
        .map((node) => {
            if (node.children?.length) {
                node.children = removeTreeNodeByPath(node.children, targetPath, targetIsDir);
            }
            return node;
        });
};

const syncTabsAfterRename = (oldPath: string, newPath: string, newName: string, isDir: boolean) => {
    const isCurrentFileRenamed = form.value.path === oldPath;
    if (currentPath.value && isPathAffected(currentPath.value, oldPath, isDir)) {
        currentPath.value = replaceAffectedPath(currentPath.value, oldPath, newPath, isDir);
    }
    if (directoryPath.value && isPathAffected(directoryPath.value, oldPath, isDir)) {
        directoryPath.value = replaceAffectedPath(directoryPath.value, oldPath, newPath, isDir);
    }
    loadedNodes.value = new Set(
        Array.from(loadedNodes.value).map((path) => replaceAffectedPath(String(path), oldPath, newPath, isDir)),
    );
    fileTabs.value = fileTabs.value.map((tab) => {
        if (!isPathAffected(tab.path, oldPath, isDir)) {
            return tab;
        }
        const nextPath = replaceAffectedPath(tab.path, oldPath, newPath, isDir);
        return {
            ...tab,
            path: nextPath,
            name: tab.path === oldPath ? newName : tab.name,
        };
    });
    if (selectTab.value && isPathAffected(selectTab.value, oldPath, isDir)) {
        selectTab.value = replaceAffectedPath(selectTab.value, oldPath, newPath, isDir);
    }
    if (form.value.path && isPathAffected(form.value.path, oldPath, isDir)) {
        form.value.path = replaceAffectedPath(form.value.path, oldPath, newPath, isDir);
        if (isCurrentFileRenamed) {
            fileName.value = newName;
        }
    }
    saveTabsToStorage();
};

const closeTabsAfterDelete = (targetPath: string, targetIsDir: boolean) => {
    const removedCurrent = selectTab.value && isPathAffected(selectTab.value, targetPath, targetIsDir);
    fileTabs.value = fileTabs.value.filter((tab) => !isPathAffected(tab.path, targetPath, targetIsDir));
    if (!removedCurrent) {
        saveTabsToStorage();
        return;
    }
    isEdit.value = false;
    const nextTab = fileTabs.value[fileTabs.value.length - 1];
    if (nextTab) {
        selectTab.value = nextTab.path;
        getContent(nextTab.path, true);
    } else {
        selectTab.value = '';
        form.value.content = '';
        form.value.path = '';
        oldFileContent.value = '';
        fileName.value = '';
        fileExtension.value = '';
        historyVersionCount.value = 0;
        disposeEditor();
    }
    saveTabsToStorage();
};

const renameFromContextMenu = async () => {
    const data = treeContextMenu.data;
    closeTreeContextMenu();
    if (!data) {
        return;
    }
    try {
        const res = await ElMessageBox.prompt(i18n.global.t('file.rename'), i18n.global.t('file.rename'), {
            inputValue: data.name,
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            inputValidator: (value) => !!value?.trim(),
        });
        const newName = String(res.value || '').trim();
        if (!newName || newName === data.name) {
            return;
        }
        const oldPath = data.path;
        const parentPath = getDirectoryPath(oldPath);
        const newPath = joinPath(parentPath, newName);
        loading.value = true;
        await renameRile({ oldName: oldPath, newName: newPath });
        renameTreeNodePaths(treeData.value, oldPath, newPath, newName, data.isDir);
        treeData.value = [...treeData.value];
        syncTabsAfterRename(oldPath, newPath, newName, data.isDir);
        loadedNodes.value.delete(oldPath);
        loadedNodes.value.delete(newPath);
        MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
    } finally {
        loading.value = false;
    }
};

const deleteFromContextMenu = async () => {
    const data = treeContextMenu.data;
    closeTreeContextMenu();
    if (!data) {
        return;
    }
    try {
        await ElMessageBox.confirm(
            i18n.global.t(data.isDir ? 'file.deleteHelper' : 'file.deleteHelper2'),
            i18n.global.t('commons.button.delete'),
            {
                confirmButtonText: i18n.global.t('commons.button.delete'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                type: 'warning',
            },
        );
        loading.value = true;
        await deleteFile({ path: data.path, isDir: data.isDir, forceDelete: false });
        closeTabsAfterDelete(data.path, data.isDir);
        treeData.value = removeTreeNodeByPath(treeData.value, data.path, data.isDir);
        loadedNodes.value.delete(data.path);
        loadedNodes.value.delete(getDirectoryPath(data.path));
        if (
            selectedParentNode.value &&
            isPathAffected(selectedParentNode.value.data?.path || '', data.path, data.isDir)
        ) {
            selectedParentNode.value = null;
        }
        if (currentPath.value && isPathAffected(currentPath.value, data.path, data.isDir)) {
            currentPath.value = getDirectoryPath(data.path);
        }
        MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
    } finally {
        loading.value = false;
    }
};

const currentEditingNode = ref<any>(null);

const createNewNode = (command: string) => {
    const isDir = command === 'dir';
    const fileName = isDir ? 'dir' : 'file';
    return {
        id: isDir ? 'new-dir' : 'new-file',
        name: fileName,
        path: `${currentPath.value}/${fileName}`,
        isDir,
        extension: '',
        children: [],
    };
};

const removeExistingNode = (data: any[], command: string) => {
    const targetId = command == 'dir' ? 'new-dir' : 'new-file';
    data = filterNodes(data, targetId);
    treeData.value = [...data];
    return data;
};

const getRawNodeChildren = (node: any) => {
    if (!node?.data) {
        return [];
    }
    if (!Array.isArray(node.data.children)) {
        node.data.children = [];
    }
    node.data.children = node.data.children.map((child) => (child?.data?.path ? child.data : child));
    return node.data.children;
};

const handleCreate = (command: string) => {
    removeExistingNode(treeData.value, command);
    if ((command === 'dir' && isCreate.value === 'file') || (command === 'file' && isCreate.value === 'dir')) {
        cancelFolder();
        isCreate.value = 'none';
    }

    if (isCreate.value !== 'none') return;

    const newFileNode = createNewNode(command);
    newFolder.value = newFileNode.name;
    currentEditingNode.value = newFileNode;
    if (selectedParentNode.value) {
        const children = getRawNodeChildren(selectedParentNode.value);
        children.unshift(newFileNode);
        updateNodeChildren(treeData.value, selectedParentNode.value.data.path, children);
        treeData.value = [...treeData.value];
    } else {
        treeData.value = [newFileNode, ...treeData.value];
    }

    nextTick(() => {
        rowRefs.value?.focus();
    });

    isCreate.value = command;
};

const filterNodes = (nodes, targetId) => {
    const filtered = nodes.filter((node) => node.id !== targetId);

    filtered.forEach((node) => {
        if (node.children && node.children.length > 0) {
            node.children = filterNodes(node.children, targetId);
        }
        if (node.data && node.data.children && node.data.children.length > 0) {
            node.data.children = filterNodes(node.data.children, targetId);
        }
    });

    return filtered;
};

const cancelFolder = () => {
    const targetId = isCreate.value == 'dir' ? 'new-dir' : 'new-file';
    isCreate.value = 'none';
    newFolder.value = '';
    if (selectedParentNode.value) {
        selectedParentNode.value.data.children = getRawNodeChildren(selectedParentNode.value).filter(
            (node) => node.id !== targetId,
        );
    }
    treeData.value = filterNodes(treeData.value, targetId);
    loadedNodes.value.delete(currentPath.value);
};

let addForm = reactive({ path: '', name: '', isDir: true, mode: 0o755, isLink: false, isSymlink: true, linkPath: '' });

const createFolder = async (isDir: boolean) => {
    addForm.path = `${currentPath.value}/${newFolder.value}`;
    const editingNode = currentEditingNode.value;
    if (!editingNode) return;
    if (addForm.path.indexOf('.1panel_clash') > -1) {
        MsgWarning(i18n.global.t('file.clashDidNotSupport'));
        return;
    }
    addForm.isDir = isDir;
    addForm.name = newFolder.value;
    let addItem = {};
    Object.assign(addItem, addForm);
    loading.value = true;
    createFile(addItem as File.FileCreate)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.createSuccess'));
            editingNode.id = newUUID();
            editingNode.name = addForm.name;
            editingNode.path = addForm.path;
            treeData.value = [...treeData.value];
            isCreate.value = 'none';
            currentEditingNode.value = null;
            newFolder.value = '';
        })
        .finally(() => {
            loading.value = false;
        });
};

onBeforeUnmount(() => {
    if (editor) {
        editor.dispose();
    }
    isCreate.value = 'none';
    currentPath.value = '';
    selectedParentNode.value = null;
    window.removeEventListener('resize', updateHeights);
    document.removeEventListener('click', closeTreeContextMenu);
    window.removeEventListener('scroll', closeTreeContextMenu, true);
});

defineExpose({ acceptParams });
</script>

<style scoped lang="scss">
.dialog-top {
    top: 0;
}

.dialog-header-icon {
    color: var(--el-color-info);
    flex-shrink: 0;
}

.monaco-editor-tree {
    color: var(--el-color-primary) !important;
}

.monaco-editor-background {
    outline-style: none;
    background-color: var(--vscode-editor-background) !important;
}

.code-splitter {
    width: 100%;
    min-width: 0;
    background-color: var(--vscode-editor-background);
}

.code-splitter :deep(.el-splitter-bar__horizontal-collapse-icon-start),
.code-splitter :deep(.el-splitter-bar__horizontal-collapse-icon-end) {
    top: 33%;
}

.code-tree-panel,
.code-editor-panel {
    min-width: 0;
    background-color: var(--vscode-editor-background);
}

.code-tree-panel.is-collapsed {
    overflow: hidden;
}

.code-editor-panel__inner,
.tree-container {
    width: 100%;
    height: 100%;
    overflow: hidden;
}

.code-editor-panel__inner {
    display: flex;
    flex-direction: column;
}

.code-box {
    min-height: 0;
    flex: 1;
}

:deep(.code-splitter .el-splitter-panel) {
    min-width: 0;
}

:deep(.code-splitter .el-splitter-bar) {
    width: 1px;
    background-color: var(--el-border-color-light);
}

:deep(.code-splitter .el-splitter-bar__dragger-horizontal::before) {
    width: 1px;
    background-color: var(--el-border-color-light);
}

.tree-widget {
    background-color: var(--el-button--primary);
}

.tree-context-menu {
    position: fixed;
    z-index: 3000;
    min-width: 148px;
    padding: 4px;
    border: 1px solid var(--el-border-color-light);
    border-radius: 4px;
    background: var(--el-bg-color-overlay);
    box-shadow: var(--el-box-shadow-light);
}

.tree-context-menu__item {
    display: flex;
    align-items: center;
    gap: 8px;
    height: 30px;
    padding: 0 8px;
    border-radius: 3px;
    color: var(--el-text-color-primary);
    cursor: pointer;
    font-size: 13px;
}

.tree-context-menu__item:hover {
    background: var(--el-fill-color-light);
}

.tree-context-menu__item.is-danger {
    color: var(--el-color-danger);
}

.tree-context-menu__icon {
    width: 16px;
    height: 16px;
    flex-shrink: 0;
}

.truncate-text {
    display: block;
    min-width: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}
.code-title {
    display: inline-flex;
    align-items: center;
    flex: 1;
    min-width: 0;
    margin-right: 8px;
}
.code-title-copy {
    flex-shrink: 0;
}
.code-header {
    background-color: var(--panel-code-header-footer-color);
}
.code-footer {
    background-color: var(--panel-code-header-footer-color);
}
.card-action {
    .el-button + .el-button {
        margin-left: 0 !important;
    }
    .el-button.is-link {
        padding: 0;
    }
}
.code-dialog {
    .el-dialog__footer {
        padding-top: 0 !important;
    }
}
.code-action {
    border-bottom: 1px solid var(--el-border-color-light) !important;
}

.table-icon {
    width: 1.35em;
    height: 1.35em;
    position: relative;
    flex-shrink: 0;
    fill: currentColor;
    vertical-align: middle;
}

.tree-node-content {
    display: inline-flex;
    align-items: center;
    width: 100%;
    min-width: 0;
}

.tree-node-editing {
    display: flex;
    align-items: center;
    gap: 2px;
    width: 100%;
    min-width: 0;
    padding-right: 8px;
}

.tree-node-label {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

:deep(.el-tabs) {
    --el-tabs-header-height: 29px;
    .el-tabs__header {
        height: 29px;
        margin: 0;
    }
    .el-tabs__nav-wrap {
        height: 28px;
        line-height: 28px;
    }
    .el-tabs__nav {
        border-right: 1px solid var(--el-border-color-light) !important;
        border-top: none !important;
        border-left: none !important;
        border-bottom: none !important;
        border-radius: 0 !important;
        box-sizing: border-box !important;
    }
    .el-tabs__nav,
    .el-tabs__nav-next,
    .el-tabs__nav-prev {
        height: 28px;
        line-height: 28px;
    }
    .el-tabs__item:hover {
        color: var(--el-color-primary) !important;
        .el-dropdown {
            color: var(--el-color-primary) !important;
        }
    }
    .el-tabs__item.is-active {
        color: var(--el-color-primary) !important;
        .el-dropdown {
            color: var(--el-color-primary) !important;
        }
    }
}

:deep(.el-dropdown .el-text:focus) {
    outline: none !important;
}

:deep(.el-input__inner:focus) {
    outline: none !important;
}

:deep(.monaco-editor .ai-search-target-line) {
    background-color: rgba(64, 158, 255, 0.14);
}

:deep(.monaco-editor .ai-search-target-line-gutter) {
    border-left: 3px solid var(--el-color-primary);
}
</style>
