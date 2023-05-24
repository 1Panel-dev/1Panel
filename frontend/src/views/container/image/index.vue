<template>
    <div v-loading="loading">
        <el-card v-if="dockerStatus != 'Running'" class="mask-prompt">
            <span>{{ $t('container.serviceUnavailable') }}</span>
            <el-button type="primary" link class="bt" @click="goSetting">【 {{ $t('container.setting') }} 】</el-button>
            <span>{{ $t('container.startIn') }}</span>
        </el-card>

        <LayoutContent :title="$t('container.image')" :class="{ mask: dockerStatus != 'Running' }">
            <template #toolbar>
                <el-row>
                    <el-col :span="16">
                        <el-button type="primary" plain @click="onOpenPull">
                            {{ $t('container.imagePull') }}
                        </el-button>
                        <el-button type="primary" plain @click="onOpenload">
                            {{ $t('container.importImage') }}
                        </el-button>
                        <el-button type="primary" plain @click="onOpenBuild">
                            {{ $t('container.imageBuild') }}
                        </el-button>
                        <el-button type="primary" plain @click="onOpenPrune()">
                            {{ $t('container.imagePrune') }}
                        </el-button>
                    </el-col>
                    <el-col :span="8">
                        <TableSetting @search="search()" />
                        <div class="search-button">
                            <el-input
                                v-model="searchName"
                                clearable
                                @clear="search()"
                                suffix-icon="Search"
                                @keyup.enter="search()"
                                @change="search()"
                                :placeholder="$t('commons.button.search')"
                            ></el-input>
                        </div>
                    </el-col>
                </el-row>
            </template>
            <template #main>
                <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search">
                    <el-table-column label="ID" prop="id" min-width="60">
                        <template #default="{ row }">
                            <Tooltip :islink="false" :text="row.id" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('container.tag')" prop="tags" min-width="160" fix>
                        <template #default="{ row }">
                            <el-tag style="margin-left: 5px" v-for="(item, index) of row.tags" :key="index">
                                {{ item }}
                            </el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('container.size')" prop="size" min-width="70" fix />
                    <el-table-column
                        prop="createdAt"
                        min-width="90"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                    />
                    <fu-table-operations
                        width="200px"
                        :ellipsis="10"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <Pull ref="dialogPullRef" @search="search" />
        <Tag ref="dialogTagRef" @search="search" />
        <Push ref="dialogPushRef" @search="search" />
        <Save ref="dialogSaveRef" @search="search" />
        <Load ref="dialogLoadRef" @search="search" />
        <Build ref="dialogBuildRef" @search="search" />
        <Delete ref="dialogDeleteRef" @search="search" />
        <Prune ref="dialogPruneRef" @search="search" />
    </div>
</template>

<script lang="ts" setup>
import Tooltip from '@/components/tooltip/index.vue';
import TableSetting from '@/components/table-setting/index.vue';
import { reactive, onMounted, ref } from 'vue';
import { dateFormat } from '@/utils/util';
import { Container } from '@/api/interface/container';
import Pull from '@/views/container/image/pull/index.vue';
import Tag from '@/views/container/image/tag/index.vue';
import Push from '@/views/container/image/push/index.vue';
import Save from '@/views/container/image/save/index.vue';
import Load from '@/views/container/image/load/index.vue';
import Build from '@/views/container/image/build/index.vue';
import Delete from '@/views/container/image/delete/index.vue';
import Prune from '@/views/container/image/prune/index.vue';
import { searchImage, listImageRepo, loadDockerStatus, imageRemove } from '@/api/modules/container';
import i18n from '@/lang';
import router from '@/routers';
import { useDeleteData } from '@/hooks/use-delete-data';

const loading = ref(false);

const data = ref();
const repos = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const searchName = ref();

const dockerStatus = ref('Running');
const loadStatus = async () => {
    loading.value = true;
    await loadDockerStatus()
        .then((res) => {
            loading.value = false;
            dockerStatus.value = res.data;
            if (dockerStatus.value === 'Running') {
                search();
                loadRepos();
            }
        })
        .catch(() => {
            dockerStatus.value = 'Failed';
            loading.value = false;
        });
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
const dialogDeleteRef = ref();
const dialogPruneRef = ref();

const search = async () => {
    const repoSearch = {
        info: searchName.value,
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

const onOpenPrune = () => {
    dialogPruneRef.value!.acceptParams();
};

const onOpenload = () => {
    dialogLoadRef.value!.acceptParams();
};

const buttons = [
    {
        label: i18n.global.t('container.tag'),
        click: (row: Container.ImageInfo) => {
            let params = {
                itemName: row.tags.length !== 0 ? row.tags[0].split(':')[0] : '',
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
        click: async (row: Container.ImageInfo) => {
            if (!row.tags?.length || row.tags.length <= 1) {
                await useDeleteData(imageRemove, { names: [row.id] }, 'commons.msg.delete');
                search();
                return;
            }
            let params = {
                id: row.id,
                tags: row.tags,
            };
            dialogDeleteRef.value!.acceptParams(params);
        },
    },
];

onMounted(() => {
    loadStatus();
});
</script>
