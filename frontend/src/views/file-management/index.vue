<template>
    <LayoutContent :header="$t('menu.files')">
        <el-row :gutter="20">
            <el-col :span="5">
                <el-scrollbar height="800px">
                    <el-tree
                        :data="fileTree"
                        :props="defaultProps"
                        :load="loadNode"
                        lazy
                        node-key="id"
                        v-loading="treeLoading"
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
                        <el-dropdown split-button type="primary">
                            {{ $t('commons.button.create') }}
                            <template #dropdown>
                                <el-dropdown-menu>
                                    <el-dropdown-item>
                                        <svg-icon iconName="p-file-folder"></svg-icon>{{ $t('file.dir') }}
                                    </el-dropdown-item>
                                    <el-dropdown-item>
                                        <svg-icon iconName="p-file-normal"></svg-icon>{{ $t('file.file') }}
                                    </el-dropdown-item>
                                    <el-dropdown-item>
                                        <svg-icon iconName="p-file-normal"></svg-icon>{{ $t('file.linkFile') }}
                                    </el-dropdown-item>
                                </el-dropdown-menu>
                            </template>
                        </el-dropdown>
                        <el-button type="primary" plain> {{ $t('file.upload') }}</el-button>
                        <el-button type="primary" plain> {{ $t('file.search') }}</el-button>
                        <el-button type="primary" plain> {{ $t('file.remoteFile') }}</el-button>
                        <el-button type="primary" plain> {{ $t('file.sync') }}</el-button>
                        <el-button type="primary" plain> {{ $t('file.terminal') }}</el-button>
                        <el-button type="primary" plain> {{ $t('file.shareList') }}</el-button>
                    </template>
                    <el-table-column :label="$t('commons.table.name')" min-width="120" fix>
                        <template #default="{ row }">
                            <svg-icon v-if="row.isDir" className="table-icon" iconName="p-file-folder"></svg-icon>
                            <svg-icon v-else className="table-icon" iconName="p-file-normal"></svg-icon>
                            <el-link :underline="false" @click="open(row.name)">{{ row.name }}</el-link>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('file.mode')" prop="mode"> </el-table-column>
                    <el-table-column :label="$t('file.user')" prop="user"> </el-table-column>
                    <el-table-column :label="$t('file.group')" prop="group"> </el-table-column>
                    <el-table-column :label="$t('file.size')" prop="size"> </el-table-column>
                    <el-table-column
                        :label="$t('file.updateTime')"
                        prop="modTime"
                        :formatter="dateFromat"
                        min-width="150"
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
        </el-row>
    </LayoutContent>
</template>

<script setup lang="ts">
import { reactive, ref } from '@vue/runtime-core';
import LayoutContent from '@/layout/layout-content.vue';
import ComplexTable from '@/components/complex-table/index.vue';
import i18n from '@/lang';
import { GetFilesList, GetFilesTree } from '@/api/modules/files';
import { dateFromat } from '@/utils/util';
import { File } from '@/api/interface/file';
import BreadCrumbs from '@/components/bread-crumbs/index.vue';
import BreadCrumbItem from '@/components/bread-crumbs/bread-crumbs-item.vue';

let data = ref();
let selects = ref<any>([]);
let req = reactive({ path: '/', expand: true });
let loading = ref<boolean>(false);
let treeLoading = ref<boolean>(false);
let paths = ref<string[]>([]);
let fileTree = ref<File.FileTree[]>([]);

const defaultProps = {
    children: 'children',
    label: 'name',
};

const paginationConfig = reactive({
    page: 1,
    pageSize: 5,
    total: 0,
});
const buttons = [
    {
        label: i18n.global.t('file.open'),
    },
    {
        label: i18n.global.t('file.mode'),
    },
    {
        label: i18n.global.t('file.zip'),
    },
    {
        label: i18n.global.t('file.rename'),
    },
    {
        label: i18n.global.t('commons.button.delete'),
    },
    {
        label: i18n.global.t('file.info'),
    },
];

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

const open = async (name: string) => {
    paths.value.push(name);
    if (req.path === '/') {
        req.path = req.path + name;
    } else {
        req.path = req.path + '/' + name;
    }
    search(req);
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
            }
            search(req);
        })
        .finally(() => {
            treeLoading.value = false;
        });
};

const loadNode = (node: any, resolve: (data: File.FileTree[]) => void) => {
    console.log(node.id);
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
</script>

<style>
.path {
    margin-top: -50px;
    height: 30px;
    margin-bottom: 5px;
}
</style>
