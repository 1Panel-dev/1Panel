<template>
    <el-dialog
        v-model="open"
        :show-close="false"
        :before-close="handleClose"
        destroy-on-close
        width="70%"
        @opened="onOpen"
        :top="'5vh'"
        :fullscreen="isFullscreen"
    >
        <template #header>
            <div ref="dialogHeader" class="flex items-center justify-between">
                <span>{{ $t('commons.button.edit') + ' - ' + form.path }}</span>
                <el-space alignment="center" :size="1" class="dialog-header-icon">
                    <el-tooltip :content="loadTooltip()" placement="top">
                        <el-button
                            @click="toggleFullscreen"
                            v-if="!mobile"
                            class="!border-none !bg-transparent !text-base !font-semibold !py-2 !px-1"
                            icon="FullScreen"
                            style="margin-right: 10px"
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
        <div ref="dialogForm">
            <el-form :inline="true" :model="config" class="mt-1.5">
                <el-form-item :label="$t('file.theme')">
                    <el-select v-model="config.theme" @change="changeTheme()" class="p-w-200">
                        <el-option v-for="item in themes" :key="item.label" :value="item.value" :label="item.label" />
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('file.language')">
                    <el-select v-model="config.language" @change="changeLanguage()" class="!w-32">
                        <el-option
                            v-for="lang in Languages"
                            :key="lang.label"
                            :value="lang.label"
                            :label="lang.label"
                        />
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('file.eol')">
                    <el-select v-model="config.eol" @change="changeEOL()" class="p-w-150">
                        <el-option v-for="eol in eols" :key="eol.label" :value="eol.value" :label="eol.label" />
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('file.wordWrap')">
                    <el-select v-model="config.wordWrap" @change="changeWarp()" class="p-w-100">
                        <el-option :label="$t('commons.button.enable')" value="on"></el-option>
                        <el-option :label="$t('commons.button.disable')" value="off"></el-option>
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('file.minimap')">
                    <el-select v-model="config.minimap" @change="changeMinimap()" class="p-w-100">
                        <el-option :label="$t('commons.button.enable')" :value="true"></el-option>
                        <el-option :label="$t('commons.button.disable')" :value="false"></el-option>
                    </el-select>
                </el-form-item>
            </el-form>
        </div>
        <div v-loading="loading">
            <div class="flex">
                <div class="monaco-editor sm:w-48 w-1/3 monaco-editor-background border-0 tree-container" v-if="isShow">
                    <div class="flex items-center justify-between pl-1 sm:pr-4 pr-1 pt-1">
                        <el-tooltip :content="$t('file.top')" placement="top">
                            <el-text size="small" @click="getUpData()" class="cursor-pointer">
                                <el-icon>
                                    <Top />
                                </el-icon>
                                <span class="sm:inline hidden pl-1">{{ $t('file.up') }}</span>
                            </el-text>
                        </el-tooltip>

                        <el-tooltip :content="$t('file.refresh')" placement="top">
                            <el-text size="small" @click="getRefresh(directoryPath)" class="cursor-pointer">
                                <el-icon>
                                    <Refresh />
                                </el-icon>
                                <span class="sm:inline hidden pl-1">{{ $t('file.refresh') }}</span>
                            </el-text>
                        </el-tooltip>
                    </div>
                    <el-divider class="!my-1" />
                    <el-tree-v2
                        ref="treeRef"
                        :data="treeData"
                        :props="treeProps"
                        @node-expand="handleNodeExpand"
                        class="monaco-editor-tree monaco-editor-background"
                        :height="treeHeight"
                        :indent="6"
                        :item-size="24"
                        highlight-current
                    >
                        <template #default="{ node, data }">
                            <!-- 目录 -->
                            <span v-if="data.isDir" style="display: inline-flex; align-items: center">
                                <svg-icon className="table-icon" iconName="p-file-folder"></svg-icon>
                                <small :title="node.label">{{ node.label }}</small>
                            </span>

                            <!-- 文档 -->
                            <span
                                v-else
                                style="display: inline-flex; align-items: center"
                                @click="getContent(data.path, data.extension)"
                            >
                                <svg-icon className="table-icon" :iconName="getIconName(data.extension)"></svg-icon>
                                <small :title="node.label" class="min-w-32">{{ node.label }}</small>
                            </span>
                        </template>
                    </el-tree-v2>
                </div>
                <div class="relative">
                    <el-divider
                        v-if="isShow"
                        direction="vertical"
                        :style="{ height: codeHeight }"
                        class="!m-0 p-0"
                        :class="isShow ? 'opacity-100' : 'opacity-0'"
                    ></el-divider>
                    <el-icon
                        v-if="isShow"
                        class="cursor-pointer absolute bg-gray-100 py-2 rounded-l-sm block top-1/3 -left-[9px]"
                        size="9"
                        @click="toggleShow"
                    >
                        <DArrowLeft />
                    </el-icon>
                    <el-icon
                        v-else
                        class="cursor-pointer absolute bg-gray-100 py-2 rounded-r-sm block top-1/3 right-[7px]"
                        size="9"
                        @click="toggleShow"
                    >
                        <DArrowRight />
                    </el-icon>
                </div>
                <div
                    ref="codeBox"
                    id="codeBox"
                    :style="{ height: codeHeight }"
                    class="flex-1 sm:w-4/5 w-2/3 relative"
                ></div>
            </div>
        </div>
        <template #footer>
            <div class="dialog-footer" ref="dialogFooter">
                <el-button @click="handleReset">{{ $t('commons.button.reset') }}</el-button>
                <el-button type="primary" @click="saveContent()">{{ $t('commons.button.confirm') }}</el-button>
            </div>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { GetFileContent, GetFilesTree, SaveFileContent } from '@/api/modules/files';
import i18n from '@/lang';
import { MsgError, MsgInfo, MsgSuccess } from '@/utils/message';
import * as monaco from 'monaco-editor';
import { nextTick, onBeforeUnmount, reactive, ref, onMounted, computed } from 'vue';
import { Languages } from '@/global/mimetype';
import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker';
import cssWorker from 'monaco-editor/esm/vs/language/css/css.worker?worker';
import htmlWorker from 'monaco-editor/esm/vs/language/html/html.worker?worker';
import tsWorker from 'monaco-editor/esm/vs/language/typescript/ts.worker?worker';
import EditorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker';

import { ElMessageBox, ElTreeV2 } from 'element-plus';
import { ResultData } from '@/api/interface';
import { File } from '@/api/interface/file';
import { getIcon } from '@/utils/util';
import { TreeKey, TreeNodeData } from 'element-plus/es/components/tree-v2/src/types';
import { Top, Refresh, DArrowLeft, DArrowRight } from '@element-plus/icons-vue';
import { loadBaseDir } from '@/api/modules/setting';
import { GlobalStore } from '@/store';

let editor: monaco.editor.IStandaloneCodeEditor | undefined;

self.MonacoEnvironment = {
    getWorker(workerId, label) {
        if (label === 'json') {
            return new jsonWorker();
        }
        if (label === 'css' || label === 'scss' || label === 'less') {
            return new cssWorker();
        }
        if (label === 'html' || label === 'handlebars' || label === 'razor') {
            return new htmlWorker();
        }
        if (['typescript', 'javascript'].includes(label)) {
            return new tsWorker();
        }
        return new EditorWorker();
    },
};

interface EditProps {
    language: string;
    content: string;
    path: string;
    name: string;
    extension: string;
}

interface EditorConfig {
    theme: string;
    language: string;
    eol: number;
    wordWrap: WordWrapOptions;
    minimap: boolean;
}

interface TreeNode {
    key: TreeKey;
    level: number;
    parent?: TreeNode;
    children?: File.FileTree[];
    data: TreeNodeData;
    disabled?: boolean;
    name?: string;
    isLeaf?: boolean;
}

const open = ref(false);
const loading = ref(false);
const fileName = ref('');
const codeThemeKey = 'code-theme';
const warpKey = 'code-warp';
const minimapKey = 'code-minimap';
const directoryPath = ref('');
const fileExtension = ref('');
const baseDir = ref();
const treeData = ref([]);
const codeBox = ref();
const defaultHeight = ref(55);
const treeHeight = ref(0);
const codeHeight = ref('55vh');
const codeReq = reactive({ path: '', expand: false, page: 1, pageSize: 100 });
const isShow = ref(true);
const isEdit = ref(false);
const oldFileContent = ref('');
const dialogHeader = ref(null);
const dialogForm = ref(null);
const dialogFooter = ref(null);

const toggleShow = () => {
    isShow.value = !isShow.value;
};

const globalStore = GlobalStore();
const mobile = computed(() => {
    return globalStore.isMobile();
});

type WordWrapOptions = 'off' | 'on' | 'wordWrapColumn' | 'bounded';

const isFullscreen = ref(false);

const config = reactive<EditorConfig>({
    theme: 'vs-dark',
    language: 'plaintext',
    eol: monaco.editor.EndOfLineSequence.LF,
    wordWrap: 'on',
    minimap: false,
});

const eols = [
    {
        label: 'LF (Linux)',
        value: monaco.editor.EndOfLineSequence.LF,
    },
    {
        label: 'CRLF (Windows)',
        value: monaco.editor.EndOfLineSequence.CRLF,
    },
];

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

const em = defineEmits(['close']);

const handleClose = () => {
    const closeEditor = () => {
        open.value = false;
        if (editor) {
            editor.dispose();
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
        MsgInfo(i18n.global.t('file.noEdit'));
    }
};

const loadTooltip = () => {
    return i18n.global.t('commons.button.' + (isFullscreen.value ? 'quitFullscreen' : 'fullscreen'));
};

onMounted(() => {
    loadPath();
    updateHeights();
    window.addEventListener('resize', updateHeights);
});

const updateHeights = () => {
    const vh = window.innerHeight / 100;
    if (isFullscreen.value) {
        let paddingHeight = 75;
        const headerHeight = dialogHeader.value.offsetHeight;
        const formHeight = dialogForm.value.offsetHeight;
        const footerHeight = dialogFooter.value.offsetHeight;
        treeHeight.value = window.innerHeight - headerHeight - formHeight - footerHeight - paddingHeight - 31;
        codeHeight.value = `${
            ((window.innerHeight - headerHeight - formHeight - footerHeight - paddingHeight) / window.innerHeight) * 100
        }vh`;
    } else {
        treeHeight.value = defaultHeight.value * vh - 31;
        codeHeight.value = `${defaultHeight.value}vh`;
    }
};

const toggleFullscreen = () => {
    isFullscreen.value = !isFullscreen.value;
    updateHeights();
};

const changeLanguage = () => {
    monaco.editor.setModelLanguage(editor.getModel(), config.language);
};

const changeTheme = () => {
    monaco.editor.setTheme(config.theme);
    const themes = {
        vs: 'monaco-editor-tree-light',
        'vs-dark': 'monaco-editor-tree-dark',
        'hc-black': 'monaco-editor-tree-dark',
    };

    if (treeRef.value) {
        Object.values(themes).forEach((themeClass) => {
            treeRef.value.$el.classList.remove(themeClass);
        });
        if (themes[config.theme]) {
            treeRef.value.$el.classList.add(themes[config.theme]);
        }
    }

    localStorage.setItem(codeThemeKey, config.theme);
};

const changeEOL = () => {
    editor.getModel().pushEOL(config.eol);
};

const changeWarp = () => {
    localStorage.setItem(warpKey, config.wordWrap);
    editor.updateOptions({
        wordWrap: config.wordWrap,
    });
};

const changeMinimap = () => {
    localStorage.setItem(minimapKey, JSON.stringify(config.minimap));
    editor.updateOptions({
        minimap: {
            enabled: config.minimap,
        },
    });
};

const initEditor = () => {
    if (editor) {
        editor.dispose();
    }
    nextTick(() => {
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
        });
        if (editor.getModel().getValue() === '') {
            let defaultContent = '\n\n\n\n';
            editor.getModel().setValue(defaultContent);
        }

        editor.getModel().pushEOL(config.eol);

        editor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyS, quickSave);

        editor.onDidChangeModelContent(() => {
            if (editor) {
                form.value.content = editor.getValue();
                isEdit.value = true;
            }
        });
    });
};

const quickSave = () => {
    saveContent();
};

const saveContent = () => {
    if (isEdit.value) {
        loading.value = true;
        SaveFileContent(form.value)
            .then(() => {
                loading.value = false;
                isEdit.value = false;
                oldFileContent.value = form.value.content;
                MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    } else {
        MsgInfo(i18n.global.t('file.noEdit'));
    }
};

const acceptParams = (props: EditProps) => {
    form.value.content = props.content;
    oldFileContent.value = props.content;
    form.value.path = props.path;
    directoryPath.value = getDirectoryPath(props.path);
    fileExtension.value = props.extension;
    fileName.value = props.name;
    config.language = props.language;
    config.eol = monaco.editor.EndOfLineSequence.LF;
    config.theme = localStorage.getItem(codeThemeKey) || 'vs-dark';
    config.wordWrap = (localStorage.getItem(warpKey) as WordWrapOptions) || 'on';
    config.minimap = localStorage.getItem(minimapKey) !== null ? localStorage.getItem(minimapKey) === 'true' : true;
    open.value = true;
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

const onOpen = () => {
    initEditor();
    changeTheme();
    search(directoryPath.value).then((res) => {
        handleSearchResult(res);
    });
};

const handleSearchResult = (res: ResultData<File.FileTree[]>) => {
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
        });
    } finally {
        loading.value = false;
        MsgSuccess(i18n.global.t('commons.msg.refreshSuccess'));
    }
};

const getContent = (path: string, extension: string) => {
    if (form.value.path === path) {
        return;
    }

    const fetchFileContent = () => {
        codeReq.path = path;
        codeReq.expand = true;

        if (extension !== '') {
            Languages.forEach((language) => {
                const ext = extension.substring(1);
                if (language.value.indexOf(ext) > -1) {
                    config.language = language.label;
                }
            });
        }

        GetFileContent(codeReq)
            .then((res) => {
                form.value.content = res.data.content;
                oldFileContent.value = res.data.content;
                form.value.path = res.data.path;
                fileExtension.value = res.data.extension;
                fileName.value = res.data.name;
                initEditor();
            })
            .catch(() => {});
    };

    if (isEdit.value) {
        ElMessageBox.confirm(i18n.global.t('file.saveAndOpenNewFile'), {
            confirmButtonText: i18n.global.t('commons.button.open'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        })
            .then(() => {
                saveContent();
                fetchFileContent();
            })
            .finally(() => {});
    } else {
        fetchFileContent();
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
    return await GetFilesTree(req);
};

const getUpData = async () => {
    if ('/' === directoryPath.value) {
        MsgInfo(i18n.global.t('commons.msg.rootInfoErr'));
        return;
    }
    let pathParts = directoryPath.value.split('/');
    pathParts.pop();
    let newPath = pathParts.join('/') || '/';

    try {
        const response = await search(newPath);
        treeData.value = response.data[0]?.children || [];
        loadedNodes.value = new Set();
    } catch (error) {
        MsgError(i18n.global.t('commons.msg.notRecords'));
    } finally {
        directoryPath.value = newPath;
    }
};

const treeRef = ref<InstanceType<typeof ElTreeV2>>();

const treeProps = {
    value: 'id',
    label: 'name',
    children: 'children',
};

const handleNodeExpand = async (node: any, data: TreeNode) => {
    if (!data.data.isDir || loadedNodes.value.has(data.data.path)) {
        return;
    }
    try {
        const response = await search(node.path);
        const newTreeData = JSON.parse(JSON.stringify(treeData.value));
        if (response.data.length > 0 && response.data[0].children) {
            data.children = response.data[0].children;
            loadedNodes.value.add(data.data.path);
            updateNodeChildren(newTreeData, data.data.path, response.data[0].children);
        } else {
            data.children = [];
        }
        treeData.value = newTreeData;
    } catch (error) {
        MsgError(i18n.global.t('commons.msg.notRecords'));
    }
};

// 更新指定节点的 children 数据
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

onBeforeUnmount(() => {
    if (editor) {
        editor.dispose();
    }
    window.removeEventListener('resize', updateHeights);
});

defineExpose({ acceptParams });
</script>

<style scoped lang="scss">
.dialog-top {
    top: 0;
}

.dialog-header-icon {
    color: var(--el-color-info);
}

.monaco-editor-tree {
    color: var(--el-color-primary) !important;
}

.monaco-editor-background {
    outline-style: none;
    background-color: var(--vscode-editor-background) !important;
}

.tree-widget {
    background-color: var(--el-button--primary);
}
</style>
