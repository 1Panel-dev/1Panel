<template>
    <div>
        <div class="flex items-center">
            <div class="flex-shrink-0 flex items-center mr-4">
                <el-tooltip :content="$t('file.back')" placement="top">
                    <el-button :icon="Back" @click="back" circle />
                </el-tooltip>
                <el-tooltip :content="$t('file.right')" placement="top">
                    <el-button :icon="Right" @click="right" circle />
                </el-tooltip>
                <el-tooltip :content="$t('file.top')" placement="top">
                    <el-button :icon="Top" @click="top" circle :disabled="paths.length == 0" />
                </el-tooltip>
                <el-tooltip :content="$t('file.refresh')" placement="top">
                    <el-button :icon="Refresh" circle @click="search" />
                </el-tooltip>
            </div>
            <div
                v-show="!searchableStatus"
                tabindex="0"
                @click="searchableStatus = true"
                class="address-bar shadow-md rounded-md px-4 py-2 flex items-center flex-grow"
            >
                <div ref="pathRef" class="w-full">
                    <span ref="breadCrumbRef" class="w-full flex items-center">
                        <span class="root mr-2">
                            <el-link @click.stop="jump('/')">
                                <el-icon :size="20"><HomeFilled /></el-icon>
                            </el-link>
                        </span>
                        <span v-for="path in paths" :key="path.url" class="inline-flex items-center">
                            <span class="mr-2 arrow">></span>
                            <el-link class="path-segment cursor-pointer mr-2 pathname" @click.stop="jump(path.url)">
                                {{ path.name }}
                            </el-link>
                        </span>
                    </span>
                </div>
            </div>
            <el-input
                ref="searchableInputRef"
                v-show="searchableStatus"
                v-model="searchablePath"
                @blur="searchableInputBlur"
                class="px-4 py-2 border rounded-md shadow-md"
                @keyup.enter="
                    jump(searchablePath);
                    searchableStatus = false;
                "
            />
        </div>
        <LayoutContent :title="$t('file.file')" v-loading="loading">
            <template #prompt>
                <el-alert type="info" :closable="false">
                    <template #title>
                        <span class="input-help whitespace-break-spaces">
                            {{ $t('file.fileHelper') }}
                        </span>
                    </template>
                </el-alert>
            </template>
            <template #toolbar>
                <div class="hidden sm:block">
                    <div class="btn-container">
                        <div class="left-section">
                            <el-dropdown @command="handleCreate">
                                <el-button type="primary">
                                    {{ $t('commons.button.create') }}
                                    <el-icon><arrow-down /></el-icon>
                                </el-button>
                                <template #dropdown>
                                    <el-dropdown-menu>
                                        <el-dropdown-item command="dir">
                                            <svg-icon iconName="p-file-folder"></svg-icon>
                                            {{ $t('file.dir') }}
                                        </el-dropdown-item>
                                        <el-dropdown-item command="file">
                                            <svg-icon iconName="p-file-normal"></svg-icon>
                                            {{ $t('file.file') }}
                                        </el-dropdown-item>
                                    </el-dropdown-menu>
                                </template>
                            </el-dropdown>
                            <el-button-group>
                                <el-button plain @click="openUpload">{{ $t('file.upload') }}</el-button>
                                <el-button plain @click="openWget">{{ $t('file.remoteFile') }}</el-button>
                                <el-button plain @click="openMove('copy')" :disabled="selects.length === 0">
                                    {{ $t('file.copy') }}
                                </el-button>
                                <el-button plain @click="openMove('cut')" :disabled="selects.length === 0">
                                    {{ $t('file.move') }}
                                </el-button>
                                <el-button plain @click="openCompress(selects)" :disabled="selects.length === 0">
                                    {{ $t('file.compress') }}
                                </el-button>
                                <el-button plain @click="openBatchRole(selects)" :disabled="selects.length === 0">
                                    {{ $t('file.editPermissions') }}
                                </el-button>
                                <el-button plain @click="batchDelFiles" :disabled="selects.length === 0">
                                    {{ $t('commons.button.delete') }}
                                </el-button>
                            </el-button-group>

                            <el-button class="btn" @click="toTerminal">
                                {{ $t('file.terminal') }}
                            </el-button>

                            <el-button-group class="copy-button" v-if="moveOpen">
                                <el-tooltip
                                    class="box-item"
                                    effect="dark"
                                    :content="$t('file.paste')"
                                    placement="bottom"
                                >
                                    <el-button plain @click="openPaste">
                                        {{ $t('file.paste') }}({{ fileMove.count }})
                                    </el-button>
                                </el-tooltip>
                                <el-tooltip
                                    class="box-item"
                                    effect="dark"
                                    :content="$t('file.cancel')"
                                    placement="bottom"
                                >
                                    <el-button plain class="close" @click="closeMove">
                                        <el-icon class="close-icon"><Close /></el-icon>
                                    </el-button>
                                </el-tooltip>
                            </el-button-group>
                        </div>

                        <div class="right-section">
                            <el-popover placement="bottom" :width="200" trigger="hover" @before-enter="getFavoriates">
                                <template #reference>
                                    <el-button @click="openFavorite">
                                        {{ $t('file.favorite') }}
                                    </el-button>
                                </template>
                                <div class="favorite-item">
                                    <el-table :data="favorites">
                                        <el-table-column prop="name">
                                            <template #default="{ row }">
                                                <el-tooltip
                                                    class="box-item"
                                                    effect="dark"
                                                    :content="row.path"
                                                    placement="top"
                                                >
                                                    <span
                                                        class="table-link text-ellipsis"
                                                        @click="toFavorite(row)"
                                                        type="primary"
                                                    >
                                                        <svg-icon
                                                            v-if="row.isDir"
                                                            className="table-icon"
                                                            iconName="p-file-folder"
                                                        ></svg-icon>
                                                        <svg-icon
                                                            v-else
                                                            className="table-icon"
                                                            iconName="p-file-normal"
                                                        ></svg-icon>
                                                        {{ row.name }}
                                                    </span>
                                                </el-tooltip>
                                            </template>
                                        </el-table-column>
                                    </el-table>
                                </div>
                            </el-popover>

                            <el-button class="btn" @click="openRecycleBin">
                                {{ $t('file.recycleBin') }}
                            </el-button>
                            <div class="search-button">
                                <el-input
                                    v-model="req.search"
                                    clearable
                                    @clear="search()"
                                    @keydown.enter="search()"
                                    :placeholder="$t('file.search')"
                                >
                                    <template #prepend>
                                        <el-checkbox v-model="req.containSub">
                                            {{ $t('file.sub') }}
                                        </el-checkbox>
                                    </template>
                                    <template #append>
                                        <el-button icon="Search" @click="search" round />
                                    </template>
                                </el-input>
                            </div>
                        </div>
                    </div>
                </div>
                <div class="block flex flex-wrap gap-4 sm:hidden">
                    <el-dropdown @command="handleCreate">
                        <el-button type="primary">
                            {{ $t('commons.button.create') }}
                            <el-icon><arrow-down /></el-icon>
                        </el-button>
                        <template #dropdown>
                            <el-dropdown-menu>
                                <el-dropdown-item command="dir">
                                    <svg-icon iconName="p-file-folder"></svg-icon>
                                    {{ $t('file.dir') }}
                                </el-dropdown-item>
                                <el-dropdown-item command="file">
                                    <svg-icon iconName="p-file-normal"></svg-icon>
                                    {{ $t('file.file') }}
                                </el-dropdown-item>
                            </el-dropdown-menu>
                        </template>
                    </el-dropdown>
                    <el-dropdown @command="handleFilePatch">
                        <el-button type="primary">
                            {{ $t('file.batchoperation') }}
                            <el-icon><arrow-down /></el-icon>
                        </el-button>
                        <template #dropdown>
                            <el-dropdown-menu>
                                <el-dropdown-item command="openUpload">
                                    {{ $t('file.upload') }}
                                </el-dropdown-item>

                                <el-dropdown-item command="openWget">
                                    {{ $t('file.remoteFile') }}
                                </el-dropdown-item>

                                <el-dropdown-item command="openMove:copy" :disabled="selects.length === 0">
                                    {{ $t('file.copy') }}
                                </el-dropdown-item>

                                <el-dropdown-item command="openMove:cut" :disabled="selects.length === 0">
                                    {{ $t('file.move') }}
                                </el-dropdown-item>

                                <el-dropdown-item command="openCompress" :disabled="selects.length === 0">
                                    {{ $t('file.compress') }}
                                </el-dropdown-item>

                                <el-dropdown-item command="openBatchRole" :disabled="selects.length === 0">
                                    {{ $t('file.role') }}
                                </el-dropdown-item>

                                <el-dropdown-item command="batchDelFiles" :disabled="selects.length === 0">
                                    {{ $t('commons.button.delete') }}
                                </el-dropdown-item>
                            </el-dropdown-menu>
                        </template>
                    </el-dropdown>

                    <el-button-group v-if="moveOpen">
                        <el-tooltip class="box-item" effect="dark" :content="$t('file.paste')" placement="bottom">
                            <el-button plain @click="openPaste">{{ $t('file.paste') }}({{ fileMove.count }})</el-button>
                        </el-tooltip>
                        <el-tooltip class="box-item" effect="dark" :content="$t('file.cancel')" placement="bottom">
                            <el-button plain class="close" @click="closeMove">
                                <el-icon class="close-icon"><Close /></el-icon>
                            </el-button>
                        </el-tooltip>
                    </el-button-group>

                    <div class="flex flex-row gap-4">
                        <el-button @click="toTerminal">
                            {{ $t('menu.terminal') }}
                        </el-button>

                        <div>
                            <el-popover placement="bottom" trigger="hover" @before-enter="getFavoriates">
                                <template #reference>
                                    <el-button @click="openFavorite">
                                        {{ $t('file.favorite') }}
                                    </el-button>
                                </template>
                                <el-table :data="favorites">
                                    <el-table-column prop="name">
                                        <template #default="{ row }">
                                            <el-tooltip
                                                class="box-item"
                                                effect="dark"
                                                :content="row.path"
                                                placement="top"
                                            >
                                                <span
                                                    class="table-link text-ellipsis"
                                                    @click="toFavorite(row)"
                                                    type="primary"
                                                >
                                                    <svg-icon
                                                        v-if="row.isDir"
                                                        className="table-icon"
                                                        iconName="p-file-folder"
                                                    ></svg-icon>
                                                    <svg-icon
                                                        v-else
                                                        className="table-icon"
                                                        iconName="p-file-normal"
                                                    ></svg-icon>
                                                    {{ row.name }}
                                                </span>
                                            </el-tooltip>
                                        </template>
                                    </el-table-column>
                                </el-table>
                            </el-popover>
                        </div>

                        <el-button @click="openRecycleBin">
                            {{ $t('file.recycleBin') }}
                        </el-button>
                    </div>
                    <el-input
                        v-model="req.search"
                        clearable
                        @clear="search()"
                        @keydown.enter="search()"
                        :placeholder="$t('file.search')"
                    >
                        <template #prepend>
                            <el-checkbox v-model="req.containSub">
                                {{ $t('file.sub') }}
                            </el-checkbox>
                        </template>
                        <template #append>
                            <el-button icon="Search" @click="search" round />
                        </template>
                    </el-input>
                </div>
            </template>
            <template #main>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    ref="tableRef"
                    :data="data"
                    @search="search"
                    @sort-change="changeSort"
                >
                    <el-table-column type="selection" width="30" />
                    <el-table-column
                        :label="$t('commons.table.name')"
                        min-width="250"
                        fix
                        show-overflow-tooltip
                        sortable
                        prop="name"
                    >
                        <template #default="{ row, $index }">
                            <div class="file-row" @mouseenter="showFavorite($index)" @mouseleave="hideFavorite">
                                <div>
                                    <svg-icon
                                        v-if="row.isDir"
                                        className="table-icon"
                                        iconName="p-file-folder"
                                    ></svg-icon>
                                    <svg-icon
                                        v-else
                                        className="table-icon"
                                        :iconName="getIconName(row.extension)"
                                    ></svg-icon>
                                </div>
                                <div class="file-name">
                                    <span class="table-link" @click="open(row)" type="primary">{{ row.name }}</span>
                                    <span v-if="row.isSymlink">-> {{ row.linkPath }}</span>
                                </div>
                                <div>
                                    <el-button
                                        v-if="row.favoriteID > 0"
                                        link
                                        type="warning"
                                        size="large"
                                        :icon="StarFilled"
                                        @click="removeFavorite(row.favoriteID)"
                                    ></el-button>
                                    <div v-else>
                                        <el-button
                                            v-if="hoveredRowIndex === $index"
                                            link
                                            :icon="Star"
                                            @click="addFavorite(row)"
                                        ></el-button>
                                    </div>
                                </div>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('file.mode')" prop="mode" min-width="110">
                        <template #default="{ row }">
                            <el-link :underline="false" @click="openMode(row)">{{ row.mode }}</el-link>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.user')" prop="user" show-overflow-tooltip min-width="90">
                        <template #default="{ row }">
                            <el-link :underline="false" @click="openChown(row)">
                                {{ row.user ? row.user : '-' }} ({{ row.uid }})
                            </el-link>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('file.group')" prop="group" show-overflow-tooltip>
                        <template #default="{ row }">
                            <el-link :underline="false" @click="openChown(row)">
                                {{ row.group ? row.group : '-' }} ({{ row.gid }})
                            </el-link>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('file.size')" prop="size" min-width="100" sortable>
                        <template #default="{ row, $index }">
                            <span v-if="row.isDir">
                                <el-button
                                    type="primary"
                                    link
                                    small
                                    @click="getDirSize(row, $index)"
                                    :loading="btnLoading.includes($index)"
                                >
                                    <span v-if="row.dirSize == undefined">
                                        {{ $t('file.calculate') }}
                                    </span>
                                    <span v-else>{{ getFileSize(row.dirSize) }}</span>
                                </el-button>
                            </span>
                            <span v-else>{{ getFileSize(row.size) }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('file.updateTime')"
                        prop="modTime"
                        width="180"
                        :formatter="dateFormat"
                        show-overflow-tooltip
                        sortable
                    ></el-table-column>
                    <fu-table-operations
                        :ellipsis="mobile ? 0 : 3"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        :min-width="mobile ? 'auto' : 200"
                        :fixed="mobile ? false : 'right'"
                        width="300"
                        fix
                    />
                </ComplexTable>
            </template>

            <CreateFile ref="createRef" @close="search" />
            <ChangeRole ref="roleRef" @close="search" />
            <Compress ref="compressRef" @close="search" />
            <Decompress ref="deCompressRef" @close="search" />
            <CodeEditor ref="codeEditorRef" @close="search" />
            <FileRename ref="renameRef" @close="search" />
            <Upload ref="uploadRef" @close="search" />
            <Wget ref="wgetRef" @close="closeWget" />
            <Move ref="moveRef" @close="closeMovePage" />
            <Download ref="downloadRef" @close="search" />
            <Process :open="processPage.open" @close="closeProcess" />
            <Owner ref="chownRef" @close="search"></Owner>
            <Detail ref="detailRef" />
            <DeleteFile ref="deleteRef" @close="search" />
            <RecycleBin ref="recycleBinRef" @close="search" />
            <Favorite ref="favoriteRef" @close="search" />
            <BatchRole ref="batchRoleRef" @close="search" />
            <VscodeOpenDialog ref="dialogVscodeOpenRef" />
            <Preview ref="previewRef" />
        </LayoutContent>
    </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, reactive, ref, computed } from 'vue';
import {
    GetFilesList,
    GetFileContent,
    ComputeDirSize,
    AddFavorite,
    RemoveFavorite,
    SearchFavorite,
} from '@/api/modules/files';
import { computeSize, copyText, dateFormat, downloadFile, getFileType, getIcon, getRandomStr } from '@/utils/util';
import { File } from '@/api/interface/file';
import { Mimetypes, Languages } from '@/global/mimetype';
import { useRouter } from 'vue-router';
import { Back, Refresh, StarFilled, Star, Top, Right, Close } from '@element-plus/icons-vue';
import { MsgWarning } from '@/utils/message';
import { useSearchable } from './hooks/searchable';
import { ResultData } from '@/api/interface';
import { GlobalStore } from '@/store';

import i18n from '@/lang';
import CreateFile from './create/index.vue';
import ChangeRole from './change-role/index.vue';
import Compress from './compress/index.vue';
import Decompress from './decompress/index.vue';
import Upload from './upload/index.vue';
import FileRename from './rename/index.vue';
import CodeEditor from './code-editor/index.vue';
import Wget from './wget/index.vue';
import Move from './move/index.vue';
import Download from './download/index.vue';
import Owner from './chown/index.vue';
import DeleteFile from './delete/index.vue';
import Process from './process/index.vue';
import Detail from './detail/index.vue';
import RecycleBin from './recycle-bin/index.vue';
import Favorite from './favorite/index.vue';
import BatchRole from './batch-role/index.vue';
import Preview from './preview/index.vue';
import VscodeOpenDialog from '@/components/vscode-open/index.vue';

const globalStore = GlobalStore();

interface FilePaths {
    url: string;
    name: string;
}

const router = useRouter();
const data = ref();
const tableRef = ref();
let selects = ref<any>([]);

// origin data
const initData = () => ({
    path: '/',
    expand: true,
    showHidden: true,
    page: 1,
    pageSize: 100,
    search: '',
    containSub: false,
    sortBy: 'name',
    sortOrder: 'ascending',
});
let req = reactive(initData());
let loading = ref(false);
let btnLoading = ref([]);
const paths = ref<FilePaths[]>([]);
let pathWidth = ref(0);
const history: string[] = [];
let pointer = -1;

const fileCreate = reactive({ path: '/', isDir: false, mode: 0o755 });
const fileCompress = reactive({ files: [''], name: '', dst: '', operate: 'compress' });
const fileDeCompress = reactive({ path: '', name: '', dst: '', mimeType: '' });
const fileEdit = reactive({ content: '', path: '', name: '', language: 'plaintext', extension: '' });
const filePreview = reactive({ path: '', name: '', extension: '', fileType: '' });
const codeReq = reactive({ path: '', expand: false, page: 1, pageSize: 100 });
const fileUpload = reactive({ path: '' });
const fileRename = reactive({ path: '', oldName: '' });
const fileWget = reactive({ path: '' });
const fileMove = reactive({ oldPaths: [''], type: '', path: '', name: '', count: 0 });
const processPage = reactive({ open: false });

const createRef = ref();
const roleRef = ref();
const detailRef = ref();
const compressRef = ref();
const deCompressRef = ref();
const codeEditorRef = ref();
const renameRef = ref();
const uploadRef = ref();
const wgetRef = ref();
const moveRef = ref();
const downloadRef = ref();
const pathRef = ref();
const breadCrumbRef = ref();
const chownRef = ref();
const moveOpen = ref(false);
const deleteRef = ref();
const recycleBinRef = ref();
const favoriteRef = ref();
const hoveredRowIndex = ref(-1);
const favorites = ref([]);
const batchRoleRef = ref();
const dialogVscodeOpenRef = ref();
const previewRef = ref();

// editablePath
const { searchableStatus, searchablePath, searchableInputRef, searchableInputBlur } = useSearchable(paths);

const paginationConfig = reactive({
    cacheSizeKey: 'file-page-size',
    currentPage: 1,
    pageSize: 100,
    total: 0,
});

const mobile = computed(() => {
    return globalStore.isMobile();
});

const handleFilePatch = (command: string) => {
    switch (command) {
        case 'openUpload':
            openUpload();
            break;
        case 'openWget':
            openWget();
            break;
        case 'openMove:copy':
            openMove('copy');
            break;
        case 'openMove:cut':
            openMove('cut');
            break;
        case 'openCompress':
            openCompress(selects.value);
            break;
        case 'openBatchRole':
            openBatchRole(selects.value);
            break;
        case 'batchDelFiles':
            batchDelFiles();
            break;
        default:
            break;
    }
};

const search = async () => {
    loading.value = true;
    if (req.search != '') {
        req.sortBy = 'name';
        req.sortOrder = 'ascending';
        tableRef.value.clearSort();
    }

    req.page = paginationConfig.currentPage;
    req.pageSize = paginationConfig.pageSize;
    await GetFilesList(req)
        .then((res) => {
            handleSearchResult(res);
        })
        .finally(() => {
            loading.value = false;
        });
};

/** just search, no handleSearchResult */
const searchFile = async () => {
    loading.value = true;
    try {
        return await GetFilesList(req);
    } finally {
        loading.value = false;
    }
};

const handleSearchResult = (res: ResultData<File.File>) => {
    paginationConfig.total = res.data.itemTotal;
    data.value = res.data.items;
    req.path = res.data.path;
};

const open = async (row: File.File) => {
    if (row.isDir) {
        if (row.name.indexOf('.1panel_clash') > -1) {
            MsgWarning(i18n.global.t('file.clashOpenAlert'));
            return;
        }
        const name = row.name;
        if (req.path.endsWith('/')) {
            req.path = req.path + name;
        } else {
            req.path = req.path + '/' + name;
        }
        paths.value.push({
            url: req.path,
            name: name,
        });
        jump(req.path);
    } else {
        openView(row);
    }
};

const copyDir = (row: File.File) => {
    if (row?.path) {
        copyText(row?.path);
    }
};

const handlePath = () => {
    const breadcrumbElement = breadCrumbRef.value as any;
    const pathNodes = breadcrumbElement.querySelectorAll('.pathname');

    let totalpathWidth = 0;
    pathNodes.forEach((node) => {
        totalpathWidth += node.offsetWidth;
    });
    if (totalpathWidth > pathWidth.value) {
        paths.value.splice(0, 1);
        paths.value[0].name = '..';
        nextTick(function () {
            handlePath();
        });
    }
};

const right = () => {
    if (pointer < history.length - 1) {
        pointer++;
        let url = history[pointer];
        backForwardJump(url);
    }
};

const back = () => {
    if (pointer > 0) {
        pointer--;
        let url = history[pointer];
        backForwardJump(url);
    }
};

const top = () => {
    if (paths.value.length > 0) {
        let url = '/';
        if (paths.value.length >= 2) {
            url = paths.value[paths.value.length - 2].url;
        }
        jump(url);
    }
};

const jump = async (url: string) => {
    history.splice(pointer + 1);
    history.push(url);
    pointer = history.length - 1;

    const { path: oldUrl, pageSize: oldPageSize } = req;
    Object.assign(req, initData(), { path: url, containSub: false, search: '', pageSize: oldPageSize });
    let searchResult = await searchFile();
    // check search result,the file is exists?
    if (!searchResult.data.path) {
        req.path = oldUrl;
        globalStore.setLastFilePath(req.path);
        MsgWarning(i18n.global.t('commons.res.notFound'));
        return;
    }
    req.path = searchResult.data.path;
    globalStore.setLastFilePath(req.path);
    handleSearchResult(searchResult);
    getPaths(req.path);
    nextTick(function () {
        handlePath();
    });
};

const backForwardJump = async (url: string) => {
    const oldPageSize = req.pageSize;
    // reset search params before exec jump
    Object.assign(req, initData());
    req.path = url;
    req.containSub = false;
    req.search = '';
    req.pageSize = oldPageSize;
    let searchResult = await searchFile();
    handleSearchResult(searchResult);
    getPaths(req.path);
    nextTick(function () {
        handlePath();
    });
};

const getPaths = (reqPath: string) => {
    const pathArray = reqPath.split('/');
    paths.value = [];
    let base = '/';
    for (const p of pathArray) {
        if (p != '') {
            if (base.endsWith('/')) {
                base = base + p;
            } else {
                base = base + '/' + p;
            }
            paths.value.push({
                url: base,
                name: p,
            });
        }
    }
};

const handleCreate = (command: string) => {
    fileCreate.path = req.path;
    fileCreate.isDir = false;
    if (command === 'dir') {
        fileCreate.isDir = true;
    }
    createRef.value.acceptParams(fileCreate);
};

const delFile = async (row: File.File | null) => {
    deleteRef.value.acceptParams([row]);
};

const batchDelFiles = () => {
    deleteRef.value.acceptParams(selects.value);
};

const getFileSize = (size: number) => {
    return computeSize(size);
};

const getDirSize = async (row: any, index: number) => {
    const req = {
        path: row.path,
    };
    btnLoading.value.push(index);
    await ComputeDirSize(req)
        .then(async (res) => {
            let newData = [...data.value];
            newData[index].dirSize = res.data.size;
            data.value = newData;
        })
        .finally(() => {
            btnLoading.value = btnLoading.value.filter((item) => item !== index);
        });
};

const getIconName = (extension: string) => {
    return getIcon(extension);
};

const openMode = (item: File.File) => {
    roleRef.value.acceptParams(item);
};

const openChown = (item: File.File) => {
    chownRef.value.acceptParams(item);
};

const openCompress = (items: File.File[]) => {
    const paths = [];
    for (const item of items) {
        paths.push(item.path);
    }
    fileCompress.files = paths;
    if (paths.length === 1) {
        fileCompress.name = items[0].name;
    } else {
        fileCompress.name = getRandomStr(6);
    }
    fileCompress.dst = req.path;

    compressRef.value.acceptParams(fileCompress);
};

const openDeCompress = (item: File.File) => {
    if (Mimetypes.get(item.mimeType) == undefined) {
        MsgWarning(i18n.global.t('file.canNotDeCompress'));
        return;
    }

    fileDeCompress.name = item.name;
    fileDeCompress.path = item.path;
    fileDeCompress.dst = req.path;
    fileDeCompress.mimeType = item.mimeType;

    deCompressRef.value.acceptParams(fileDeCompress);
};

const openView = (item: File.File) => {
    const fileType = getFileType(item.extension);

    const previewTypes = ['image', 'video', 'audio', 'word', 'excel'];
    if (previewTypes.includes(fileType)) {
        return openPreview(item, fileType);
    }

    const actionMap = {
        compress: openDeCompress,
        text: () => openCodeEditor(item.path, item.extension),
    };
    const path = item.isSymlink ? item.linkPath : item.path;
    return actionMap[fileType] ? actionMap[fileType](item) : openCodeEditor(path, item.extension);
};

const openPreview = (item: File.File, fileType: string) => {
    if (item.mode.toString() == '-' && item.user == '-' && item.group == '-') {
        MsgWarning(i18n.global.t('file.fileCanNotRead'));
        return;
    }
    filePreview.path = item.isSymlink ? item.linkPath : item.path;
    filePreview.name = item.name;
    filePreview.extension = item.extension;
    filePreview.fileType = fileType;

    previewRef.value.acceptParams(filePreview);
};

const openCodeEditor = (path: string, extension: string) => {
    codeReq.path = path;
    codeReq.expand = true;

    if (extension != '') {
        Languages.forEach((language) => {
            const ext = extension.substring(1);
            if (language.value.indexOf(ext) > -1) {
                fileEdit.language = language.label;
            }
        });
    }

    GetFileContent(codeReq)
        .then((res) => {
            fileEdit.content = res.data.content;
            fileEdit.path = res.data.path;
            fileEdit.name = res.data.name;
            fileEdit.extension = res.data.extension;

            codeEditorRef.value.acceptParams(fileEdit);
        })
        .catch(() => {});
};

const openUpload = () => {
    fileUpload.path = req.path;
    uploadRef.value.acceptParams(fileUpload);
};

const openWget = () => {
    fileWget.path = req.path;
    wgetRef.value.acceptParams(fileWget);
};

const openBatchRole = (items: File.File[]) => {
    batchRoleRef.value.acceptParams({ files: items });
};

const closeWget = (submit: Boolean) => {
    search();
    if (submit) {
        openProcess();
    }
};

const closeMovePage = (submit: Boolean) => {
    if (submit) {
        search();
        closeMove();
    }
};

const openProcess = () => {
    processPage.open = true;
};

const closeProcess = () => {
    processPage.open = false;
};

const openRename = (item: File.File) => {
    fileRename.path = req.path;
    fileRename.oldName = item.name;
    renameRef.value.acceptParams(fileRename);
};

const openMove = (type: string) => {
    fileMove.type = type;
    fileMove.name = '';
    const oldpaths = [];
    for (const s of selects.value) {
        oldpaths.push(s['path']);
    }
    fileMove.count = selects.value.length;
    fileMove.oldPaths = oldpaths;
    if (selects.value.length == 1) {
        fileMove.name = selects.value[0].name;
    }
    moveOpen.value = true;
};

const closeMove = () => {
    selects.value = [];
    tableRef.value.clearSelects();
    fileMove.oldPaths = [];
    fileMove.name = '';
    fileMove.count = 0;
    moveOpen.value = false;
};

const openPaste = () => {
    fileMove.path = req.path;
    moveRef.value.acceptParams(fileMove);
};

const openDownload = (file: File.File) => {
    downloadFile(file.path);
};

const openDetail = (row: File.File) => {
    detailRef.value.acceptParams({ path: row.path });
};

const openRecycleBin = () => {
    recycleBinRef.value.acceptParams();
};

const openFavorite = () => {
    favoriteRef.value.acceptParams();
};

const changeSort = ({ prop, order }) => {
    req.sortBy = prop;
    req.sortOrder = order;
    req.search = '';
    req.page = 1;
    req.pageSize = paginationConfig.pageSize;
    req.containSub = false;
    search();
};

const showFavorite = (index: any) => {
    hoveredRowIndex.value = index;
};

const hideFavorite = () => {
    hoveredRowIndex.value = -1;
};

const addFavorite = async (row: File.File) => {
    try {
        await AddFavorite(row.path);
        search();
    } catch (error) {}
};

const removeFavorite = async (id: number) => {
    ElMessageBox.confirm(i18n.global.t('file.removeFavorite'), i18n.global.t('commons.msg.remove'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        try {
            await RemoveFavorite(id);
            search();
        } catch (error) {}
    });
};

const getFavoriates = async () => {
    try {
        const res = await SearchFavorite(req);
        favorites.value = res.data.items;
    } catch (error) {}
};

const toFavorite = (row: File.Favorite) => {
    if (row.isDir) {
        jump(row.path);
    } else {
        let file = {} as File.File;
        file.path = row.path;
        file.extension = '.' + row.name.split('.').pop();
        openView(file);
    }
};

const toTerminal = () => {
    router.push({ path: '/hosts/terminal', query: { path: req.path } });
};

const openWithVSCode = (row: File.File) => {
    dialogVscodeOpenRef.value.acceptParams({ path: row.path + (row.isDir ? '' : ':1:1') });
};

const buttons = [
    {
        label: i18n.global.t('file.open'),
        click: open,
    },
    {
        label: i18n.global.t('file.download'),
        click: (row: File.File) => {
            openDownload(row);
        },
        disabled: (row: File.File) => {
            return row.isDir;
        },
    },
    {
        label: i18n.global.t('file.deCompress'),
        click: openDeCompress,
        disabled: (row: File.File) => {
            return !isDecompressFile(row);
        },
    },
    {
        label: i18n.global.t('file.editPermissions'),
        click: (row: File.File) => {
            openBatchRole([row]);
        },
    },
    {
        label: i18n.global.t('file.compress'),
        click: (row: File.File) => {
            openCompress([row]);
        },
    },
    {
        label: i18n.global.t('file.rename'),
        click: openRename,
    },
    {
        label: i18n.global.t('file.copyDir'),
        click: copyDir,
    },
    {
        label: i18n.global.t('file.openWithVscode'),
        click: openWithVSCode,
    },
    {
        label: i18n.global.t('commons.button.delete'),
        disabled: (row: File.File) => {
            return row.name == '.1panel_clash';
        },
        click: delFile,
    },
    {
        label: i18n.global.t('file.info'),
        click: openDetail,
    },
];

const isDecompressFile = (row: File.File) => {
    if (row.isDir) {
        return false;
    }
    if (getFileType(row.extension) === 'compress') {
        return true;
    }
    if (row.mimeType == 'application/octet-stream') {
        return false;
    } else {
        return Mimetypes.get(row.mimeType) != undefined;
    }
};

onMounted(() => {
    if (router.currentRoute.value.query.path) {
        req.path = String(router.currentRoute.value.query.path);
        getPaths(req.path);
        globalStore.setLastFilePath(req.path);
    } else {
        if (globalStore.lastFilePath && globalStore.lastFilePath != '') {
            req.path = globalStore.lastFilePath;
            getPaths(req.path);
        }
    }
    pathWidth.value = pathRef.value.offsetWidth;
    search();
    history.push(req.path);
    pointer = history.length - 1;
    nextTick(function () {
        handlePath();
    });
});
</script>

<style scoped lang="scss">
.path {
    display: flex;
    align-items: center;
    border: 1px solid #ebeef5;
    background-color: var(--panel-path-bg);
    height: 30px;
    border-radius: 2px !important;
    &:hover {
        cursor: text;
        box-shadow: var(--el-box-shadow);
    }

    .root {
        vertical-align: middle;
        margin-left: 10px;
    }
    .other {
        vertical-align: middle;
    }
    .split {
        margin-left: 5px;
        margin-right: 5px;
    }
}

.copy-button {
    margin-left: 10px;
    .close {
        width: 10px;
        .close-icon {
            color: red;
        }
    }
}

.btn-container {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
}

.left-section,
.right-section {
    display: flex;
    align-items: center;
    gap: 1rem;
}

.left-section > *:not(:first-child) {
    margin-left: 5px;
}

.right-section {
    .btn {
        margin-right: 10px;
    }
}
.favorite-item {
    max-height: 650px;
    overflow: auto;
}

.file-row {
    display: flex;
    align-items: center;
    width: 100%;
}

.file-name {
    flex-grow: 1;
    margin-left: 1px;
    width: 95%;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
}
.address-bar {
    border: var(--el-border);
    .arrow {
        color: #726e6e;
    }
}
.search-button {
    width: 20vw;
}
</style>
