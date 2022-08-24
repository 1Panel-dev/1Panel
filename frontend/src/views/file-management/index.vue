<template>
    <LayoutContent :header="$t('menu.files')">
        <el-row :gutter="20">
            <el-col :span="6">
                <el-tree :data="dataSource" node-key="id">
                    <template #default="{ node }">
                        <el-icon v-if="node.data.isDir && node.expanded"><FolderOpened /></el-icon>
                        <el-icon v-if="node.data.isDir && !node.expanded"><Folder /></el-icon>
                        <el-icon v-if="!node.data.isDir"><Document /></el-icon>
                        <span class="custom-tree-node">
                            <span>{{ node.data.label }}</span>
                        </span>
                    </template>
                </el-tree>
            </el-col>
            <el-col :span="18">
                <div class="path">
                    <el-breadcrumb :separator-icon="ArrowRight">
                        <el-breadcrumb-item @click="jump(-1)">root</el-breadcrumb-item>
                        <el-breadcrumb-item v-for="(item, key) in paths" :key="key" @click="jump(key)">{{
                            item
                        }}</el-breadcrumb-item>
                    </el-breadcrumb>
                </div>

                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    :data="data"
                    :loading="loading"
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
import { onMounted, reactive, ref } from '@vue/runtime-core';
import LayoutContent from '@/layout/layout-content.vue';
import ComplexTable from '@/components/complex-table/index.vue';
import i18n from '@/lang';
import { GetFilesList } from '@/api/modules/files';
import { dateFromat } from '@/utils/util';
import { ArrowRight } from '@element-plus/icons-vue';
import { File } from '@/api/interface/file';
interface Tree {
    id: number;
    label: string;
    isDir: Boolean;
    children?: Tree[];
}
let data = ref();
let selects = ref<any>([]);
let req = reactive({ path: '/', expand: true });
let loading = ref<boolean>(false);
let paths = ref<string[]>([]);
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

const search = (req: File.ReqFile) => {
    loading.value = true;
    GetFilesList(req)
        .then((res) => {
            data.value = res.data.items;
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

const dataSource = ref<Tree[]>([
    {
        id: 1,
        label: 'var',
        isDir: true,
        children: [
            {
                id: 4,
                label: 'log',
                isDir: true,
                children: [
                    {
                        id: 9,
                        isDir: false,
                        label: 'ko.log',
                    },
                    {
                        id: 10,
                        isDir: false,
                        label: 'kubepi.log',
                    },
                ],
            },
        ],
    },
    {
        id: 2,
        label: 'opt',
        isDir: true,
        children: [
            {
                id: 5,
                isDir: false,
                label: 'app.conf',
            },
            {
                id: 6,
                isDir: false,
                label: 'test.txt',
            },
        ],
    },
]);

onMounted(() => {
    search(req);
});
</script>

<style>
.path {
    margin-top: -50px;
    height: 30px;
    margin-bottom: 5px;
}
</style>
