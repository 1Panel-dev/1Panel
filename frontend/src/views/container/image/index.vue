<template>
    <div v-loading="loading">
        <docker-status
            v-model:isActive="isActive"
            v-model:isExist="isExist"
            v-model:loading="loading"
            @search="search"
            @mounted="loadRepos"
        />

        <LayoutContent v-if="isExist" :title="$t('container.image', 2)" :class="{ mask: !isActive }">
            <template #leftToolBar>
                <el-button v-permission type="primary" plain @click="onOpenPull">
                    {{ $t('container.imagePull') }}
                </el-button>
                <el-button v-permission type="primary" plain @click="onOpenLoad">
                    {{ $t('container.importImage') }}
                </el-button>
                <el-button v-permission type="primary" plain @click="onOpenBuild">
                    {{ $t('container.imageBuild') }}
                </el-button>
                <el-button v-permission type="primary" plain @click="onOpenBuildCache()">
                    {{ $t('container.cleanBuildCache') }}
                </el-button>
                <el-button v-permission type="primary" plain @click="onOpenPrune()">
                    {{ $t('container.imagePrune') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <TableSearch @search="search()" v-model:searchName="paginationConfig.name" />
                <TableRefresh @search="search()" />
                <TableSetting title="image-refresh" @search="search()" />
            </template>
            <template #main>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    :data="data"
                    @sort-change="search"
                    @cell-mouse-enter="showFavorite"
                    @cell-mouse-leave="hideFavorite"
                    :columns="columns"
                    @search="search"
                    :heightDiff="300"
                >
                    <el-table-column label="ID" prop="id" width="180">
                        <template #default="{ row, $index }">
                            <el-text type="primary" class="cursor-pointer" @click="onInspect(row.id)">
                                {{ row.id.replaceAll('sha256:', '').substring(0, 12) }}
                            </el-text>
                            <div class="float-right">
                                <el-tooltip
                                    :content="row.isPinned ? $t('website.cancelFavorite') : $t('website.favorite')"
                                    v-if="row.isPinned || hoveredRowIndex === $index"
                                >
                                    <el-button
                                        link
                                        size="large"
                                        :icon="row.isPinned ? 'StarFilled' : 'Star'"
                                        type="warning"
                                        v-permission
                                        @click="changePinned(row, true)"
                                    />
                                </el-tooltip>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.status')" prop="isUsed" width="100" sortable="custom">
                        <template #default="{ row }">
                            <Status :status="row.isUsed ? 'used' : 'unused'" />
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('container.tag')"
                        prop="tags"
                        sortable="custom"
                        min-width="160"
                        :width="isMobile ? 400 : 'auto'"
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
                    <el-table-column :label="$t('container.size')" prop="size" min-width="60" fix sortable="custom">
                        <template #default="{ row }">
                            {{ computeSize2(row.size) }}
                        </template>
                    </el-table-column>
                    <el-table-column
                        sortable="custom"
                        prop="createdAt"
                        min-width="80"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                    />
                    <fu-table-operations
                        width="250px"
                        :ellipsis="2"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <CodemirrorDrawer ref="myDetail" />
        <DialogPro v-model="updateDialogVisible" :title="$t('commons.button.update')" @close="handleUpdateDialogClose">
            <el-form label-position="top">
                <el-form-item :label="$t('container.tag')">
                    <el-checkbox
                        class="w-full"
                        :model-value="updateCheckAll"
                        :indeterminate="updateIndeterminate"
                        @change="onUpdateCheckAllChange"
                    >
                        {{ $t('commons.table.all') }}
                    </el-checkbox>
                    <el-checkbox-group v-model="updateSelectedTags">
                        <el-checkbox v-for="tag in updateTagOptions" :key="tag" :label="tag">
                            {{ tag }}
                        </el-checkbox>
                    </el-checkbox-group>
                    <span class="input-help">{{ $t('container.imageUpdateHelper') }}</span>
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="handleUpdateDialogClose">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button v-permission type="primary" @click="submitUpdateSelection">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </DialogPro>

        <OpDialog ref="opRef" @submit="onSubmitDelete" />
        <Pull ref="dialogPullRef" @search="search" />
        <Tag ref="dialogTagRef" @search="search" />
        <Push ref="dialogPushRef" @search="search" />
        <Save ref="dialogSaveRef" @search="search" />
        <Load ref="dialogLoadRef" @search="search" />
        <Build ref="dialogBuildRef" @search="search" />
        <Delete ref="dialogDeleteRef" @search="search" />
        <Prune ref="dialogPruneRef" @search="search" />
        <TaskLog ref="taskLogRef" width="70%" @close="search" />
    </div>
</template>

<script lang="ts" setup>
import { reactive, ref, computed } from 'vue';
import { dateFormat } from '@/utils/date';
import { newUUID } from '@/utils/id';
import { computeSize2 } from '@/utils/size';
import { Container } from '@/api/interface/container';
import Pull from '@/views/container/image/pull/index.vue';
import Tag from '@/views/container/image/tag/index.vue';
import Push from '@/views/container/image/push/index.vue';
import Save from '@/views/container/image/save/index.vue';
import Load from '@/views/container/image/load/index.vue';
import Build from '@/views/container/image/build/index.vue';
import Delete from '@/views/container/image/delete/index.vue';
import Prune from '@/views/container/image/prune/index.vue';
import DockerStatus from '@/views/container/docker-status/index.vue';
import CodemirrorDrawer from '@/components/codemirror-pro/drawer.vue';
import TaskLog from '@/components/log/task/index.vue';
import { searchImage, listImageRepo, imageRemove, inspect, containerPrune, imagePull } from '@/api/modules/container';
import i18n from '@/lang';
import { ElMessageBox } from 'element-plus';
import { updateCommonDescription } from '@/api/modules/setting';
import { MsgError, MsgSuccess } from '@/utils/message';
import { useGlobalStore } from '@/composables/useGlobalStore';

const { isMobile } = useGlobalStore();
const taskLogRef = ref();

const loading = ref(false);

const opRef = ref();

const data = ref();
const names = ref();
const repos = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'container-image-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('container-image-page-size')) || 20,
    total: 0,
    name: '',
    orderBy: 'createdAt',
    order: 'null',
});
const columns = ref([]);

const isActive = ref(false);
const isExist = ref(false);

const hoveredRowIndex = ref(-1);

const myDetail = ref();
const dialogPullRef = ref();
const dialogTagRef = ref();
const dialogPushRef = ref();
const dialogLoadRef = ref();
const dialogSaveRef = ref();
const dialogBuildRef = ref();
const dialogDeleteRef = ref();
const dialogPruneRef = ref();
const updateDialogVisible = ref(false);
const updateTagOptions = ref<Array<string>>([]);
const updateSelectedTags = ref<Array<string>>([]);
const updateCheckAll = computed(
    () => updateTagOptions.value.length > 0 && updateSelectedTags.value.length === updateTagOptions.value.length,
);
const updateIndeterminate = computed(
    () => updateSelectedTags.value.length > 0 && updateSelectedTags.value.length < updateTagOptions.value.length,
);

const search = async (column?: any) => {
    if (!isActive.value || !isExist.value) {
        return;
    }
    paginationConfig.orderBy = column?.order ? column.prop : paginationConfig.orderBy;
    paginationConfig.order = column?.order ? column.order : paginationConfig.order;
    const params = {
        name: paginationConfig.name,
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        orderBy: paginationConfig.orderBy,
        order: paginationConfig.order,
    };
    loading.value = true;
    await searchImage(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};
const loadRepos = async () => {
    const res = await listImageRepo();
    repos.value = res.data || [];
};

const onDelete = (row: Container.ImageInfo) => {
    names.value = [row.id.replaceAll('sha256:', '').substring(0, 12)];
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names.value,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('container.image'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: null,
        params: null,
    });
};

const onSubmitDelete = async () => {
    loading.value = true;
    let taskID = newUUID();
    await imageRemove({ names: names.value, taskID: taskID })
        .then(() => {
            loading.value = false;
            openTaskLog(taskID);
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            loading.value = false;
        });
};

const showFavorite = (row: any) => {
    hoveredRowIndex.value = data.value.findIndex((item) => item === row);
};
const hideFavorite = () => {
    hoveredRowIndex.value = -1;
};
const changePinned = (row: any, isPinned: boolean) => {
    let params = {
        id: row.id.replaceAll('sha256:', ''),
        type: 'image',
        detailType: '',
        isPinned: !row.isPinned,
        description: row.description || '',
    };
    if (isPinned) {
        params.isPinned = !row.isPinned;
    }
    updateCommonDescription(params).then(() => {
        search();
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    });
};

const onInspect = async (id: string) => {
    const res = await inspect({ id: id, type: 'image', detail: '' });
    let detailInfo = JSON.stringify(JSON.parse(res.data), null, 2);
    let param = {
        header: i18n.global.t('commons.button.view'),
        detailInfo: detailInfo,
        mode: 'json',
    };
    myDetail.value!.acceptParams(param);
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
            taskID: newUUID(),
            pruneType: 'buildcache',
            withTagAll: false,
        };
        await containerPrune(params)
            .then(() => {
                loading.value = false;
                openTaskLog(params.taskID);
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    });
};
const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const onOpenLoad = () => {
    dialogLoadRef.value!.acceptParams();
};

const normalizeImageTags = (tags: string[]) => {
    return (tags || []).filter((tag) => tag && !tag.includes('<none>'));
};

const pullImageTags = async (tags: string[]) => {
    const taskID = newUUID();
    await imagePull({
        taskID: taskID,
        repoID: 0,
        imageName: tags,
    });
    openTaskLog(taskID);
};

const runUpdate = async (tags: string[]) => {
    const validTags = normalizeImageTags(tags);
    if (validTags.length === 0) {
        MsgError(i18n.global.t('container.imageUpdateTagEmpty'));
        return;
    }
    try {
        await ElMessageBox.confirm(
            i18n.global.t('container.imageUpdateHelper'),
            i18n.global.t('commons.button.update'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                type: 'info',
            },
        );
    } catch {
        return;
    }
    await pullImageTags(validTags);
};

const onUpdate = async (row: Container.ImageInfo) => {
    const tags = normalizeImageTags(row.tags || []);
    if (tags.length === 0) {
        MsgError(i18n.global.t('container.imageUpdateTagEmpty'));
        return;
    }
    if (tags.length === 1) {
        await runUpdate(tags);
        return;
    }
    updateTagOptions.value = tags;
    updateSelectedTags.value = [...tags];
    updateDialogVisible.value = true;
};

const onUpdateCheckAllChange = (checked: boolean) => {
    updateSelectedTags.value = checked ? [...updateTagOptions.value] : [];
};

const handleUpdateDialogClose = () => {
    updateDialogVisible.value = false;
    updateTagOptions.value = [];
    updateSelectedTags.value = [];
};

const submitUpdateSelection = async () => {
    if (updateSelectedTags.value.length === 0) {
        MsgError(i18n.global.t('commons.msg.confirmNoNull', [i18n.global.t('container.tag')]));
        return;
    }
    const selected = [...updateSelectedTags.value];
    handleUpdateDialogClose();
    await runUpdate(selected);
};

const buttons = [
    {
        label: i18n.global.t('container.push'),
        permission: true,
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
        permission: true,
        click: (row: Container.ImageInfo) => {
            let params = {
                repos: repos.value,
                tags: row.tags,
            };
            dialogSaveRef.value!.acceptParams(params);
        },
    },
    {
        label: i18n.global.t('commons.button.update'),
        permission: true,
        click: (row: Container.ImageInfo) => {
            onUpdate(row);
        },
    },
    {
        label: i18n.global.t('container.tag'),
        permission: true,
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
        label: i18n.global.t('commons.button.delete'),
        permission: true,
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
</script>
