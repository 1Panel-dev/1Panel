<template>
    <div v-loading="loading">
        <Submenu activeName="image" />
        <el-card width="30%" v-if="dockerStatus != 'Running'" class="mask-prompt">
            <span style="font-size: 14px">{{ $t('container.serviceUnavailable') }}</span>
            <el-button type="primary" link style="font-size: 14px; margin-bottom: 5px" @click="goSetting">
                【 {{ $t('container.setting') }} 】
            </el-button>
            <span style="font-size: 14px">{{ $t('container.startIn') }}</span>
        </el-card>
        <el-card style="margin-top: 20px" :class="{ mask: dockerStatus != 'Running' }">
            <LayoutContent :header="$t('container.image')">
                <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search">
                    <template #toolbar>
                        <el-button @click="onOpenPull">
                            {{ $t('container.imagePull') }}
                        </el-button>
                        <el-button @click="onOpenload">
                            {{ $t('container.importImage') }}
                        </el-button>
                        <el-button @click="onOpenBuild">
                            {{ $t('container.build') }}
                        </el-button>
                    </template>
                    <el-table-column label="ID" show-overflow-tooltip prop="id" min-width="60" />
                    <el-table-column :label="$t('container.tag')" prop="tags" min-width="160" fix>
                        <template #default="{ row }">
                            <el-tag style="margin-left: 5px" v-for="(item, index) of row.tags" :key="index">
                                {{ item }}
                            </el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('container.size')" prop="size" min-width="70" fix />
                    <el-table-column :label="$t('commons.table.createdAt')" min-width="80" fix>
                        <template #default="{ row }">
                            {{ dateFromat(0, 0, row.createdAt) }}
                        </template>
                    </el-table-column>
                    <fu-table-operations
                        width="200px"
                        :ellipsis="10"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                    />
                </ComplexTable>
            </LayoutContent>
        </el-card>

        <Pull ref="dialogPullRef" @search="search" />
        <Tag ref="dialogTagRef" @search="search" />
        <Push ref="dialogPushRef" @search="search" />
        <Save ref="dialogSaveRef" @search="search" />
        <Load ref="dialogLoadRef" @search="search" />
        <Build ref="dialogBuildRef" @search="search" />

        <el-dialog v-model="deleteVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="30%">
            <template #header>
                <div class="card-header">
                    <span>{{ $t('container.imageDelete') }}</span>
                </div>
            </template>
            <el-form :model="deleteForm" label-width="80px">
                <el-form-item label="Tag" prop="tagName">
                    <el-checkbox-group v-model="deleteForm.deleteTags">
                        <el-checkbox v-for="item in deleteForm.tags" :key="item" :value="item" :label="item" />
                    </el-checkbox-group>
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="deleteVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button type="primary" :disabled="deleteForm.deleteTags.length === 0" @click="batchDelete()">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import { reactive, onMounted, ref } from 'vue';
import { dateFromat } from '@/utils/util';
import Submenu from '@/views/container/index.vue';
import { Container } from '@/api/interface/container';
import LayoutContent from '@/layout/layout-content.vue';
import Pull from '@/views/container/image/pull/index.vue';
import Tag from '@/views/container/image/tag/index.vue';
import Push from '@/views/container/image/push/index.vue';
import Save from '@/views/container/image/save/index.vue';
import Load from '@/views/container/image/load/index.vue';
import Build from '@/views/container/image/build/index.vue';
import { searchImage, listImageRepo, imageRemove, loadDockerStatus } from '@/api/modules/container';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { useDeleteData } from '@/hooks/use-delete-data';
import router from '@/routers';

const loading = ref(false);

const data = ref();
const repos = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const dockerStatus = ref();
const loadStatus = async () => {
    const res = await loadDockerStatus();
    dockerStatus.value = res.data;
    if (dockerStatus.value === 'Running') {
        search();
        loadRepos();
    }
};
const goSetting = async () => {
    router.push({ name: 'ContainerSetting' });
};

const dialogPullRef = ref();
const dialogTagRef = ref();
const dialogPushRef = ref();
const dialogLoadRef = ref();
const dialogSaveRef = ref();
const dialogBuildRef = ref();

const deleteVisiable = ref(false);
const deleteForm = reactive({
    deleteTags: [] as Array<string>,
    tags: [] as Array<string>,
});

const search = async () => {
    const repoSearch = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    await searchImage(repoSearch).then((res) => {
        data.value = res.data.items || [];
        paginationConfig.total = res.data.total;
    });
};
const loadRepos = async () => {
    const res = await listImageRepo();
    repos.value = res.data || [];
};

const onOpenPull = () => {
    let params = {
        repos: repos.value,
    };
    dialogPullRef.value!.acceptParams(params);
};

const onOpenBuild = () => {
    dialogBuildRef.value!.acceptParams();
};

const onOpenload = () => {
    dialogLoadRef.value!.acceptParams();
};

const batchDelete = async () => {
    let names: Array<string> = [];
    for (const item of deleteForm.deleteTags) {
        names.push(item);
    }
    await useDeleteData(imageRemove, { names: names }, 'commons.msg.delete');
    deleteVisiable.value = false;
    search();
};

const buttons = [
    {
        label: 'Tag',
        click: (row: Container.ImageInfo) => {
            let params = {
                repos: repos.value,
                sourceID: row.id,
            };
            dialogTagRef.value!.acceptParams(params);
        },
    },
    {
        label: i18n.global.t('container.push'),
        click: (row: Container.ImageInfo) => {
            let params = {
                repos: repos.value,
                tags: row.tags,
            };
            dialogPushRef.value!.acceptParams(params);
        },
    },
    {
        label: i18n.global.t('container.export'),
        click: (row: Container.ImageInfo) => {
            let params = {
                repos: repos.value,
                tags: row.tags,
            };
            dialogSaveRef.value!.acceptParams(params);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Container.ImageInfo) => {
            deleteForm.tags = row.tags;
            deleteForm.deleteTags = [];
            deleteVisiable.value = true;
        },
    },
];

onMounted(() => {
    loadStatus();
});
</script>
