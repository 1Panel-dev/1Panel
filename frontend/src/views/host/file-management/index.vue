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
                <div class="path" ref="pathRef">
                    <span ref="breadCrumbRef">
                        <span class="root">
                            <el-link @click="jump('/')">
                                <el-icon :size="20"><HomeFilled /></el-icon>
                            </el-link>
                        </span>
                        <span v-for="item in paths" :key="item.url" class="other">
                            <span class="split">></span>
                            <el-link @click="jump(item.url)">{{ item.name }}</el-link>
                        </span>
                    </span>
                </div>
            </el-col>
        </el-row>
        <LayoutContent :title="$t('file.file')" v-loading="loading">
            <template #toolbar>
                <el-dropdown @command="handleCreate">
                    <el-button type="primary">
                        {{ $t('commons.button.create') }}
                        <el-icon class="el-icon--right"><arrow-down /></el-icon>
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
                    <el-button plain @click="openDownload" :disabled="selects.length === 0">
                        {{ $t('file.download') }}
                    </el-button>
                    <el-button plain @click="batchDelFiles" :disabled="selects.length === 0">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </el-button-group>
                <div class="search search-button">
                    <el-input
                        v-model="req.search"
                        clearable
                        @clear="search()"
                        suffix-icon="Search"
                        @blur="search()"
                        :placeholder="$t('file.search')"
                    >
                        <!-- <template #prepend>
                            <el-checkbox :disabled="req.path == '/'" v-model="req.containSub">
                                {{ $t('file.sub') }}
                            </el-checkbox>
                        </template> -->
                    </el-input>
                </div>
            </template>
            <template #main>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    :data="data"
                    @search="search"
                >
                    <el-table-column type="selection" width="55" />
                    <el-table-column :label="$t('commons.table.name')" min-width="250" fix show-overflow-tooltip>
                        <template #default="{ row }">
                            <svg-icon v-if="row.isDir" className="table-icon" iconName="p-file-folder"></svg-icon>
                            <svg-icon v-else className="table-icon" :iconName="getIconName(row.extension)"></svg-icon>
                            <span class="table-link" @click="open(row)" type="primary">{{ row.name }}</span>
                            <span v-if="row.isSymlink">-> {{ row.linkPath }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('file.mode')" prop="mode">
                        <template #default="{ row }">
                            <el-link :underline="false" @click="openMode(row)" type="primary">{{ row.mode }}</el-link>
                        </template>
                    </el-table-column>
                    <!-- <el-table-column :label="$t('file.user')" prop="user" show-overflow-tooltip></el-table-column>
                    <el-table-column :label="$t('file.group')" prop="group"></el-table-column> -->
                    <el-table-column :label="$t('file.size')" prop="size">
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
                        :formatter="dateFormat"
                        min-width="150"
                        show-overflow-tooltip
                    ></el-table-column>

                    <fu-table-operations
                        :ellipsis="1"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
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
            <Move ref="moveRef" @close="search" />
            <Download ref="downloadRef" @close="search" />
            <Process :open="processPage.open" @close="closeProcess" />
            <!-- <Detail ref="detailRef" /> -->
        </LayoutContent>
    </div>
</template>

<script setup lang="ts">
import { nextTick, onMounted, reactive, ref } from '@vue/runtime-core';
import { GetFilesList, DeleteFile, GetFileContent, ComputeDirSize, DownloadFile } from '@/api/modules/files';
import { computeSize, dateFormat, getIcon, getRandomStr } from '@/utils/util';
import { File } from '@/api/interface/file';
import { useDeleteData } from '@/hooks/use-delete-data';
import LayoutContent from '@/layout/layout-content.vue';
import ComplexTable from '@/components/complex-table/index.vue';
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
import { Mimetypes, Languages } from '@/global/mimetype';
import Process from './process/index.vue';
// import Detail from './detail/index.vue';
import { useRouter } from 'vue-router';
import { Back, Refresh } from '@element-plus/icons-vue';
import { MsgSuccess, MsgWarning } from '@/utils/message';

interface FilePaths {
    url: string;
    name: string;
}

const router = useRouter();
const data = ref();
let selects = ref<any>([]);
let req = reactive({
    path: '/',
    expand: true,
    showHidden: true,
    page: 1,
    pageSize: 100,
    search: '',
    containSub: false,
});
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
const fileMove = reactive({ oldPaths: [''], type: '' });
const fileDownload = reactive({ paths: [''], name: '' });
const processPage = reactive({ open: false });

const createRef = ref();
const roleRef = ref();
// const detailRef = ref();
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
            paginationConfig.total = res.data.itemTotal;
            data.value = res.data.items;
            req.path = res.data.path;
        })
        .finally(() => {
            loading.value = false;
        });
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
    req.path = url;
    req.containSub = false;
    req.search = '';
    await search();
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
            if (language.value == ext) {
                fileEdit.language = language.value;
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
    moveRef.value.acceptParams(fileMove);
};

const openDownload = () => {
    const paths = [];
    for (const s of selects.value) {
        paths.push(s['path']);
    }
    fileDownload.paths = paths;
    if (selects.value.length > 1 || selects.value[0].isDir) {
        fileDownload.name = selects.value.length > 1 ? getRandomStr(6) : selects.value[0].name;
        downloadRef.value.acceptParams(fileDownload);
    } else {
        loading.value = true;
        fileDownload.name = selects.value[0].name;
        DownloadFile(fileDownload as File.FileDownload)
            .then((res) => {
                const downloadUrl = window.URL.createObjectURL(new Blob([res], { type: 'application/octet-stream' }));
                const a = document.createElement('a');
                a.style.display = 'none';
                a.href = downloadUrl;
                a.download = fileDownload.name;
                const event = new MouseEvent('click');
                a.dispatchEvent(event);
            })
            .finally(() => {
                loading.value = false;
            });
    }
};

// const openDetail = (row: File.File) => {
//     detailRef.value.acceptParams({ path: row.path });
// };

const buttons = [
    {
        label: i18n.global.t('file.open'),
        click: open,
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
        label: i18n.global.t('file.deCompress'),
        click: openDeCompress,
        disabled: (row: File.File) => {
            return row.isDir;
        },
    },
    {
        label: i18n.global.t('file.rename'),
        click: openRename,
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: delFile,
    },
    // {
    //     label: i18n.global.t('file.info'),
    //     click: openDetail,
    // },
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
    float: right;
    display: inline;
    width: 20%;
}
</style>
