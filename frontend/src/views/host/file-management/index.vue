<template>
    <div>
        <el-row>
            <el-col :span="2">
                <div>
                    <el-button :icon="Back" @click="back" circle :disabled="paths.length == 0" />
                    <el-button :icon="Refresh" circle @click="search" />
                </div>
            </el-col>
            <el-col :span="22">
                <div v-show="!searchableStatus" tabindex="0" @click="searchableStatus = true">
                    <div class="path" ref="pathRef">
                        <span ref="breadCrumbRef">
                            <span class="root">
                                <el-link @click.stop="jump('/')">
                                    <el-icon :size="20"><HomeFilled /></el-icon>
                                </el-link>
                            </span>
                            <span v-for="item in paths" :key="item.url" class="other">
                                <span class="split">></span>
                                <el-link @click.stop="jump(item.url)">{{ item.name }}</el-link>
                            </span>
                        </span>
                    </div>
                </div>
                <el-input
                    ref="searchableInputRef"
                    v-show="searchableStatus"
                    v-model="searchablePath"
                    @blur="searchableInputBlur"
                    @keyup.enter="
                        jump(searchablePath);
                        searchableStatus = false;
                    "
                />
            </el-col>
        </el-row>
        <LayoutContent :title="$t('file.file')" v-loading="loading">
            <template #toolbar>
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
                <el-button-group style="margin-left: 10px">
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
                    <el-button plain @click="batchDelFiles" :disabled="selects.length === 0">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </el-button-group>
                <el-button-group class="copy-button" v-if="moveOpen">
                    <el-tooltip class="box-item" effect="dark" :content="$t('file.paste')" placement="bottom">
                        <el-button plain @click="openPaste">{{ $t('file.paste') }}</el-button>
                    </el-tooltip>
                    <el-tooltip class="box-item" effect="dark" :content="$t('file.cancel')" placement="bottom">
                        <el-button plain class="close" @click="closeMove">
                            <el-icon class="close-icon"><Close /></el-icon>
                        </el-button>
                    </el-tooltip>
                </el-button-group>
                <div class="search search-button">
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
                            <el-button icon="Search" @click="search" />
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
                >
                    <el-table-column type="selection" width="30" />
                    <el-table-column :label="$t('commons.table.name')" min-width="250" fix show-overflow-tooltip>
                        <template #default="{ row }">
                            <svg-icon v-if="row.isDir" className="table-icon" iconName="p-file-folder"></svg-icon>
                            <svg-icon v-else className="table-icon" :iconName="getIconName(row.extension)"></svg-icon>
                            <span class="table-link" @click="open(row)" type="primary">{{ row.name }}</span>
                            <span v-if="row.isSymlink">-> {{ row.linkPath }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('file.mode')" prop="mode" max-width="50">
                        <template #default="{ row }">
                            <el-link :underline="false" @click="openMode(row)" type="primary">{{ row.mode }}</el-link>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.user')" prop="user" show-overflow-tooltip>
                        <template #default="{ row }">
                            <el-link :underline="false" @click="openChown(row)" type="primary">{{ row.user }}</el-link>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('file.group')" prop="group">
                        <template #default="{ row }">
                            <el-link :underline="false" @click="openChown(row)" type="primary">{{ row.group }}</el-link>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('file.size')" prop="size" max-width="50">
                        <template #default="{ row }">
                            <span v-if="row.isDir">
                                <el-button type="primary" link small @click="getDirSize(row)">
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
                        min-width="150"
                        :formatter="dateFormat"
                        show-overflow-tooltip
                    ></el-table-column>
                    <fu-table-operations
                        :ellipsis="3"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        min-width="300"
                        fixed="right"
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
        </LayoutContent>
    </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, reactive, ref } from '@vue/runtime-core';
import { GetFilesList, DeleteFile, GetFileContent, ComputeDirSize } from '@/api/modules/files';
import { computeSize, dateFormat, downloadFile, getIcon, getRandomStr } from '@/utils/util';
import { File } from '@/api/interface/file';
import { useDeleteData } from '@/hooks/use-delete-data';
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
import { Mimetypes, Languages } from '@/global/mimetype';
import Process from './process/index.vue';
import Detail from './detail/index.vue';
import { useRouter } from 'vue-router';
import { Back, Refresh } from '@element-plus/icons-vue';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { ElMessageBox } from 'element-plus';
import { useSearchable } from './hooks/searchable';
import { ResultData } from '@/api/interface';

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
});
let req = reactive(initData());
let loading = ref(false);
const paths = ref<FilePaths[]>([]);
let pathWidth = ref(0);

const fileCreate = reactive({ path: '/', isDir: false, mode: 0o755 });
const fileCompress = reactive({ files: [''], name: '', dst: '', operate: 'compress' });
const fileDeCompress = reactive({ path: '', name: '', dst: '', mimeType: '' });
const fileEdit = reactive({ content: '', path: '', name: '', language: 'plaintext' });
const codeReq = reactive({ path: '', expand: false, page: 1, pageSize: 100 });
const fileUpload = reactive({ path: '' });
const fileRename = reactive({ path: '', oldName: '' });
const fileWget = reactive({ path: '' });
const fileMove = reactive({ oldPaths: [''], type: '', path: '' });
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

// editablePath
const { searchableStatus, searchablePath, searchableInputRef, searchableInputBlur } = useSearchable(paths);

const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 100,
    total: 0,
});

const search = async () => {
    loading.value = true;
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
        openCodeEditor(row);
    }
};

const copyDir = (row: File.File) => {
    if (row?.path) {
        const input = document.createElement('textarea');
        input.value = row?.path;
        document.body.appendChild(input);
        input.select();
        document.execCommand('copy');
        document.body.removeChild(input);
        MsgSuccess(i18n.global.t('commons.msg.copySuccess'));
    }
};

const handlePath = () => {
    if (breadCrumbRef.value.offsetWidth > pathWidth.value) {
        paths.value.splice(0, 1);
        paths.value[0].name = '..';
        nextTick(function () {
            handlePath();
        });
    }
};

const back = () => {
    if (paths.value.length > 0) {
        let url = '/';
        if (paths.value.length >= 2) {
            url = paths.value[paths.value.length - 2].url;
        }
        jump(url);
    }
};

const jump = async (url: string) => {
    const oldUrl = req.path;
    // reset search params before exec jump
    Object.assign(req, initData());
    req.path = url;
    req.containSub = false;
    req.search = '';
    let searchResult = await searchFile();
    // check search result,the file is exists?
    if (!searchResult.data.path) {
        req.path = oldUrl;
        MsgWarning(i18n.global.t('commons.res.notFound'));
        return;
    }
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

const handleCreate = (commnad: string) => {
    fileCreate.path = req.path;
    fileCreate.isDir = false;
    if (commnad === 'dir') {
        fileCreate.isDir = true;
    }
    createRef.value.acceptParams(fileCreate);
};

const delFile = async (row: File.File | null) => {
    await useDeleteData(DeleteFile, row as File.FileDelete, 'commons.msg.delete');
    search();
};

const batchDelFiles = () => {
    ElMessageBox.confirm(i18n.global.t('commons.msg.delete'), i18n.global.t('commons.msg.deleteTitle'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(() => {
        const pros = [];
        for (const s of selects.value) {
            pros.push(DeleteFile({ path: s['path'], isDir: s['isDir'] }));
        }
        loading.value = true;
        Promise.all(pros)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
                search();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const getFileSize = (size: number) => {
    return computeSize(size);
};

const getDirSize = async (row: any) => {
    const req = {
        path: row.path,
    };
    loading.value = true;
    await ComputeDirSize(req)
        .then(async (res) => {
            row.dirSize = res.data.size;
        })
        .finally(() => {
            loading.value = false;
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

const openCodeEditor = (row: File.File) => {
    codeReq.path = row.path;
    codeReq.expand = true;

    if (row.extension != '') {
        Languages.forEach((language) => {
            const ext = row.extension.substring(1);
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
    const oldpaths = [];
    for (const s of selects.value) {
        oldpaths.push(s['path']);
    }
    fileMove.oldPaths = oldpaths;
    moveOpen.value = true;
};

const closeMove = () => {
    selects.value = [];
    tableRef.value.clearSelects();
    fileMove.oldPaths = [];
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
            return row.isDir;
        },
    },
    {
        label: i18n.global.t('file.mode'),
        click: openMode,
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
        label: i18n.global.t('commons.button.delete'),
        click: delFile,
    },
    {
        label: i18n.global.t('file.info'),
        click: openDetail,
    },
];

onMounted(() => {
    if (router.currentRoute.value.query.path) {
        req.path = String(router.currentRoute.value.query.path);
        getPaths(req.path);
    }
    pathWidth.value = pathRef.value.offsetWidth * 0.7;
    search();
});
</script>

<style scoped lang="scss">
.path {
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

.search {
    display: inline;
    width: 300px;
    float: right;
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
</style>
