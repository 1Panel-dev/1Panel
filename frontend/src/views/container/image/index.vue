<template>
    <div v-loading="loading">
        <el-card v-if="dockerStatus != 'Running'" class="mask-prompt">
            <span>{{ $t('container.serviceUnavailable') }}</span>
            <el-button type="primary" link class="bt" @click="goSetting">【 {{ $t('container.setting') }} 】</el-button>
            <span>{{ $t('container.startIn') }}</span>
        </el-card>

        <LayoutContent :title="$t('container.image', 2)" :class="{ mask: dockerStatus != 'Running' }">
            <template #toolbar>
                <div class="flex justify-between gap-2 flex-wrap sm:flex-row">
                    <div class="flex flex-wrap gap-3">
                        <el-button type="primary" plain @click="onOpenPull">
                            {{ $t('container.imagePull') }}
                        </el-button>
                        <el-button type="primary" plain @click="onOpenload">
                            {{ $t('container.importImage') }}
                        </el-button>
                        <el-button type="primary" plain @click="onOpenBuild">
                            {{ $t('container.imageBuild') }}
                        </el-button>
                        <el-button type="primary" plain @click="onOpenBuildCache()">
                            {{ $t('container.cleanBuildCache') }}
                        </el-button>
                        <el-button type="primary" plain @click="onOpenPrune()">
                            {{ $t('container.imagePrune') }}
                        </el-button>
                    </div>
                    <div class="flex flex-wrap gap-3">
                        <TableSetting @search="search()" />
                        <TableSearch @search="search()" v-model:searchName="searchName" />
                    </div>
                </div>
            </template>
            <template #main>
                <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search">
                    <el-table-column label="ID" prop="id" width="140" show-overflow-tooltip>
                        <template #default="{ row }">
                            <el-text type="primary" class="cursor-pointer" @click="onInspect(row.id)">
                                {{ row.id.replaceAll('sha256:', '').substring(0, 12) }}
                            </el-text>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.status')" prop="isUsed" width="100">
                        <template #default="{ row }">
                            <el-tag icon="Select" v-if="row.isUsed" type="success">
                                {{ $t('commons.status.used') }}
                            </el-tag>
                            <el-tag v-else type="info">
                                {{ $t('commons.status.unUsed') }}
                            </el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('container.tag')"
                        prop="tags"
                        min-width="160"
                        :width="mobile ? 400 : 'auto'"
                        fix
                    >
                        <template #default="{ row }">
                            <el-tag
                                class="ml-2.5"
                                v-for="(item, index) of row.tags"
                                :key="index"
                                :title="item"
                                type="info"
                            >
                                {{ item }}
                            </el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('container.size')" prop="size" min-width="60" fix />
                    <el-table-column
                        prop="createdAt"
                        min-width="80"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                    />
                    <fu-table-operations
                        width="250px"
                        :ellipsis="10"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <CodemirrorDialog ref="mydetail" />

        <OpDialog ref="opRef" @search="search" />
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
import { reactive, onMounted, ref, computed } from 'vue';
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
import CodemirrorDialog from '@/components/codemirror-dialog/index.vue';
import {
    searchImage,
    listImageRepo,
    loadDockerStatus,
    imageRemove,
    inspect,
    containerPrune,
} from '@/api/modules/container';
import i18n from '@/lang';
import router from '@/routers';
import { GlobalStore } from '@/store';
import { ElMessageBox } from 'element-plus';
import { MsgSuccess } from '@/utils/message';
const globalStore = GlobalStore();

const mobile = computed(() => {
    return globalStore.isMobile();
});

const loading = ref(false);

const opRef = ref();

const data = ref();
const repos = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'container-image-page-size',
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

const mydetail = ref();
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

const onDelete = (row: Container.ImageInfo) => {
    let names = [row.id.replaceAll('sha256:', '').substring(0, 12)];
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('container.image'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: imageRemove,
        params: { names: names },
    });
};

const onInspect = async (id: string) => {
    const res = await inspect({ id: id, type: 'image' });
    let detailInfo = JSON.stringify(JSON.parse(res.data), null, 2);
    let param = {
        header: i18n.global.t('commons.button.view'),
        detailInfo: detailInfo,
    };
    mydetail.value!.acceptParams(param);
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

const onOpenBuildCache = () => {
    ElMessageBox.confirm(i18n.global.t('container.delBuildCacheHelper'), i18n.global.t('container.cleanBuildCache'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        let params = {
            pruneType: 'buildcache',
            withTagAll: false,
        };
        await containerPrune(params)
            .then((res) => {
                loading.value = false;
                MsgSuccess(i18n.global.t('container.cleanSuccess', [res.data.deletedNumber]));
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const onOpenload = () => {
    dialogLoadRef.value!.acceptParams();
};

const buttons = [
    {
        label: i18n.global.t('container.tag'),
        click: (row: Container.ImageInfo) => {
            let params = {
                repos: repos.value,
                imageID: row.id,
                tags: row.tags,
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
            if (row.tags && row.tags.length > 1) {
                let params = {
                    id: row.id,
                    isUsed: row.isUsed,
                    tags: row.tags,
                };
                dialogDeleteRef.value!.acceptParams(params);
            } else {
                onDelete(row);
            }
        },
    },
];

onMounted(() => {
    loadStatus();
});
</script>
