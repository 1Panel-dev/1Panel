<template>
    <LayoutContent>
        <el-row :gutter="20">
            <el-col :span="5">
                <el-scrollbar height="80vh">
                    <el-tree
                        :data="fileTree"
                        :props="defaultProps"
                        :load="loadNode"
                        lazy
                        v-loading="treeLoading"
                        node-key="id"
                        :default-expanded-keys="expandKeys"
                        @node-click="clickNode"
                    >
                        <template #default="{ node }">
                            <el-icon v-if="node.expanded"><FolderOpened /></el-icon>
                            <el-icon v-else><Folder /></el-icon>
                            <span class="custom-tree-node">
                                <span>{{ node.data.name }}</span>
                            </span>
                        </template>
                    </el-tree>
                </el-scrollbar>
            </el-col>

            <el-col :span="19">
                <div class="path">
                    <BreadCrumbs>
                        <BreadCrumbItem @click="jump(-1)" :right="paths.length == 0">root</BreadCrumbItem>
                        <BreadCrumbItem
                            v-for="(item, key) in paths"
                            :key="key"
                            @click="jump(key)"
                            :right="key == paths.length - 1"
                            >{{ item }}</BreadCrumbItem
                        >
                    </BreadCrumbs>
                </div>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    :data="data"
                    v-loading="loading"
                >
                    <template #toolbar>
                        <el-dropdown split-button type="primary" @command="handleCreate">
                            {{ $t('commons.button.create') }}
                            <template #dropdown>
                                <el-dropdown-menu>
                                    <el-dropdown-item command="dir">
                                        <svg-icon iconName="p-file-folder"></svg-icon>{{ $t('file.dir') }}
                                    </el-dropdown-item>
                                    <el-dropdown-item command="file">
                                        <svg-icon iconName="p-file-normal"></svg-icon>{{ $t('file.file') }}
                                    </el-dropdown-item>
                                </el-dropdown-menu>
                            </template>
                        </el-dropdown>
                        <el-button type="primary" plain @click="openUpload"> {{ $t('file.upload') }}</el-button>
                        <!-- <el-button type="primary" plain> {{ $t('file.search') }}</el-button> -->
                        <el-button type="primary" plain @click="openDownload"> {{ $t('file.remoteFile') }}</el-button>
                        <!-- <el-button type="primary" plain> {{ $t('file.sync') }}</el-button>
                        <el-button type="primary" plain> {{ $t('file.terminal') }}</el-button>
                        <el-button type="primary" plain> {{ $t('file.shareList') }}</el-button> -->
                    </template>
                    <el-table-column :label="$t('commons.table.name')" min-width="250" fix show-overflow-tooltip>
                        <template #default="{ row }">
                            <svg-icon v-if="row.isDir" className="table-icon" iconName="p-file-folder"></svg-icon>
                            <svg-icon v-else className="table-icon" iconName="p-file-normal"></svg-icon>
                            <el-link :underline="false" @click="open(row)">{{ row.name }}</el-link>
                            <span v-if="row.isSymlink"> -> {{ row.linkPath }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('file.mode')" prop="mode">
                        <template #default="{ row }">
                            <el-link :underline="false" @click="openMode(row)">{{ row.mode }}</el-link>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('file.user')" prop="user"> </el-table-column>
                    <el-table-column :label="$t('file.group')" prop="group"> </el-table-column>
                    <el-table-column :label="$t('file.size')" prop="size"> </el-table-column>
                    <el-table-column
                        :label="$t('file.updateTime')"
                        prop="modTime"
                        :formatter="dateFromat"
                        min-width="100"
                        show-overflow-tooltip
                    >
                    </el-table-column>

                    <fu-table-operations
                        :ellipsis="1"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        fixed="right"
                        fix
                    />
                </ComplexTable>
            </el-col>
            <CreateFile :open="filePage.open" :file="filePage.createForm" @close="closeCreate"></CreateFile>
            <ChangeRole :open="modePage.open" :file="modePage.modeForm" @close="closeMode"></ChangeRole>
            <Compress
                :open="compressPage.open"
                :files="compressPage.files"
                :dst="compressPage.dst"
                :name="compressPage.name"
                @close="closeCompress"
            ></Compress>
            <Decompress
                :open="deCompressPage.open"
                :dst="deCompressPage.dst"
                :path="deCompressPage.path"
                :name="deCompressPage.name"
                :mimeType="deCompressPage.mimeType"
                @close="closeDeCompress"
            ></Decompress>
            <CodeEditor
                :open="editorPage.open"
                :language="'json'"
                :content="editorPage.content"
                :loading="editorPage.loading"
                @close="closeCodeEditor"
                @qsave="quickSave"
                @save="saveContent"
            ></CodeEditor>
            <FileRename
                :open="renamePage.open"
                :path="renamePage.path"
                :oldName="renamePage.oldName"
                @close="closeRename"
            ></FileRename>
            <Upload :open="uploadPage.open" :path="uploadPage.path" @close="closeUpload"></Upload>
            <FileDown :open="downloadPage.open" :path="downloadPage.path" @close="closeDownload"></FileDown>
        </el-row>
    </LayoutContent>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from '@vue/runtime-core';
import LayoutContent from '@/layout/layout-content.vue';
import ComplexTable from '@/components/complex-table/index.vue';
import i18n from '@/lang';
import { GetFilesList, GetFilesTree, DeleteFile, GetFileContent, SaveFileContent } from '@/api/modules/files';
import { dateFromat } from '@/utils/util';
import { File } from '@/api/interface/file';
import BreadCrumbs from '@/components/bread-crumbs/index.vue';
import BreadCrumbItem from '@/components/bread-crumbs/bread-crumbs-item.vue';
import CreateFile from './create/index.vue';
import ChangeRole from './change-role/index.vue';
import Compress from './compress/index.vue';
import Decompress from './decompress/index.vue';
import Upload from './upload/index.vue';
import FileRename from './file-rename/index.vue';
import { useDeleteData } from '@/hooks/use-delete-data';
import CodeEditor from './code-editor/index.vue';
import { ElMessage } from 'element-plus';
import FileDown from './file-down/index.vue';

let data = ref();
let selects = ref<any>([]);
let req = reactive({ path: '/', expand: true });
let loading = ref(false);
let treeLoading = ref(false);
let paths = ref<string[]>([]);
let fileTree = ref<File.FileTree[]>([]);
let expandKeys = ref<string[]>([]);

const filePage = reactive({ open: false, createForm: { path: '/', isDir: false, mode: 0o755 } });
const modePage = reactive({ open: false, modeForm: { path: '/', isDir: false, mode: 0o755 } });
const compressPage = reactive({ open: false, files: [''], name: '', dst: '' });
const deCompressPage = reactive({ open: false, path: '', name: '', dst: '', mimeType: '' });
const editorPage = reactive({ open: false, content: '', loading: false });
const codeReq = reactive({ path: '', expand: false });
const uploadPage = reactive({ open: false, path: '' });
const renamePage = reactive({ open: false, path: '', oldName: '' });
const downloadPage = reactive({ open: false, path: '' });

const defaultProps = {
    children: 'children',
    label: 'name',
    id: 'id',
};

const paginationConfig = reactive({
    page: 1,
    pageSize: 5,
    total: 0,
});

const search = async (req: File.ReqFile) => {
    loading.value = true;
    await GetFilesList(req)
        .then((res) => {
            data.value = res.data.items;
            req.path = res.data.path;
            const pathArray = req.path.split('/');
            paths.value = [];
            for (const p of pathArray) {
                if (p != '') {
                    paths.value.push(p);
                }
            }
        })
        .finally(() => {
            loading.value = false;
        });
};

const open = async (row: File.File) => {
    if (row.isDir) {
        const name = row.name;
        paths.value.push(name);
        if (req.path.endsWith('/')) {
            req.path = req.path + name;
        } else {
            req.path = req.path + '/' + name;
        }
        search(req);
    } else {
        openCodeEditor(row);
    }
};

const jump = async (index: number) => {
    let path = '/';
    if (index != -1) {
        const jPaths = paths.value.slice(0, index + 1);
        for (let i in jPaths) {
            path = path + '/' + jPaths[i];
        }
    }
    req.path = path;
    search(req);
};

const getTree = async (req: File.ReqFile, node: File.FileTree | null) => {
    treeLoading.value = true;
    await GetFilesTree(req)
        .then((res) => {
            if (node) {
                if (res.data.length > 0) {
                    node.children = res.data[0].children;
                }
            } else {
                fileTree.value = res.data;
                expandKeys.value = [];
                expandKeys.value.push(fileTree.value[0].id);
            }
        })
        .finally(() => {
            treeLoading.value = false;
        });
};

const clickNode = async (node: any) => {
    if (node.path) {
        req.path = node.path;
        search(req);
    }
};

const loadNode = (node: any, resolve: (data: File.FileTree[]) => void) => {
    if (!node.hasChildNodes) {
        if (node.data.path) {
            req.path = node.data.path;
            getTree(req, node.data);
        } else {
            getTree(req, null);
        }
    }
    resolve([]);
};

const handleCreate = (commnad: string) => {
    filePage.createForm.path = req.path;
    filePage.createForm.isDir = false;
    if (commnad === 'dir') {
        filePage.createForm.isDir = true;
    }
    filePage.open = true;
};

const delFile = async (row: File.File | null) => {
    await useDeleteData(DeleteFile, row as File.FileDelete, 'commons.msg.delete', loading.value);
    search(req);
};

const closeCreate = () => {
    filePage.open = false;
    search(req);
};

const openMode = (item: File.File) => {
    modePage.modeForm = item;
    modePage.open = true;
};

const closeMode = () => {
    modePage.open = false;
    search(req);
};

const openCompress = (item: File.File) => {
    compressPage.open = true;
    compressPage.files = [item.path];
    compressPage.name = item.name;
    compressPage.dst = req.path;
};

const closeCompress = () => {
    compressPage.open = false;
    search(req);
};

const openDeCompress = (item: File.File) => {
    deCompressPage.open = true;
    deCompressPage.name = item.name;
    deCompressPage.path = item.path;
    deCompressPage.dst = req.path;
    deCompressPage.mimeType = item.mimeType;
};

const closeDeCompress = () => {
    deCompressPage.open = false;
    search(req);
};

const openCodeEditor = (row: File.File) => {
    codeReq.path = row.path;
    codeReq.expand = true;
    GetFileContent(codeReq).then((res) => {
        editorPage.content = res.data.content;
    });
    editorPage.open = true;
};

const closeCodeEditor = () => {
    editorPage.open = false;
};

const openUpload = () => {
    uploadPage.open = true;
    uploadPage.path = req.path;
};

const closeUpload = () => {
    uploadPage.open = false;
    search(req);
};

const openDownload = () => {
    downloadPage.open = true;
    downloadPage.path = req.path;
};

const closeDownload = () => {
    downloadPage.open = false;
    search(req);
};

const openRename = (item: File.File) => {
    renamePage.open = true;
    renamePage.path = req.path;
    renamePage.oldName = item.name;
};

const closeRename = () => {
    renamePage.open = false;
    search(req);
};

const saveContent = (content: string) => {
    editorPage.loading = true;
    SaveFileContent({ path: codeReq.path, content: content }).finally(() => {
        editorPage.loading = false;
        editorPage.open = false;
        ElMessage.success(i18n.global.t('commons.msg.updateSuccess'));
    });
};

const quickSave = (content: string) => {
    editorPage.loading = true;
    SaveFileContent({ path: codeReq.path, content: content }).finally(() => {
        editorPage.loading = false;
        ElMessage.success(i18n.global.t('commons.msg.updateSuccess'));
    });
};

onMounted(() => {
    search(req);
});

//TODO button增加v-if判断
//openDeCompress 增加是否可以解压判断
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
        click: openCompress,
    },
    {
        label: i18n.global.t('file.deCompress'),
        click: openDeCompress,
    },
    {
        label: i18n.global.t('file.rename'),
        click: openRename,
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: delFile,
    },
    {
        label: i18n.global.t('file.info'),
    },
];
</script>

<style>
.path {
    height: 30px;
    margin-bottom: 5px;
}
</style>
