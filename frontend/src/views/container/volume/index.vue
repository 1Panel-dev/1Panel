<template>
    <div v-loading="loading">
        <docker-status
            v-model:isActive="isActive"
            v-model:isExist="isExist"
            v-model:loading="loading"
            @search="search"
        />

        <LayoutContent v-if="isExist" :title="$t('container.volume', 2)" :class="{ mask: !isActive }">
            <template #leftToolBar>
                <el-button type="primary" @click="onCreate()">
                    {{ $t('container.createVolume') }}
                </el-button>
                <el-button type="primary" plain @click="onClean()">
                    {{ $t('container.volumePrune') }}
                </el-button>
                <el-button :disabled="selects.length === 0" @click="batchDelete(null)">
                    {{ $t('commons.button.delete') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <TableSearch @search="search()" v-model:searchName="searchName" />
                <TableRefresh @search="search()" />
                <TableSetting title="volume-refresh" @search="search()" />
            </template>
            <template #main>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    :data="data"
                    @search="search"
                    :heightDiff="300"
                >
                    <el-table-column type="selection" fix />
                    <el-table-column
                        :label="$t('commons.table.name')"
                        min-width="100"
                        :width="mobile ? 220 : 'auto'"
                        prop="name"
                        fix
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <el-text type="primary" class="cursor-pointer" @click="onInspect(row.name)">
                                {{ row.name }}
                            </el-text>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('container.volumeDir')" min-width="100">
                        <template #default="{ row }">
                            <el-tooltip :content="row.mountpoint">
                                <el-button type="primary" link @click="routerToFileWithPath(row.mountpoint)">
                                    <el-icon>
                                        <FolderOpened />
                                    </el-icon>
                                </el-button>
                            </el-tooltip>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('container.mountpoint')"
                        show-overflow-tooltip
                        min-width="120"
                        prop="mountpoint"
                    />
                    <el-table-column
                        :label="$t('container.driver')"
                        show-overflow-tooltip
                        min-width="80"
                        prop="driver"
                    />
                    <el-table-column
                        prop="createdAt"
                        min-width="90"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                    />
                    <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fix />
                </ComplexTable>
            </template>
        </LayoutContent>

        <OpDialog ref="opRef" @search="search" />

        <CodemirrorDrawer ref="myDetail" />
        <CreateDialog @search="search" ref="dialogCreateRef" />
        <TaskLog ref="taskLogRef" width="70%" @close="search" />
    </div>
</template>

<script lang="ts" setup>
import CreateDialog from '@/views/container/volume/create/index.vue';
import CodemirrorDrawer from '@/components/codemirror-pro/drawer.vue';
import DockerStatus from '@/views/container/docker-status/index.vue';
import { reactive, ref, computed } from 'vue';
import { dateFormat, newUUID } from '@/utils/util';
import { deleteVolume, searchVolume, inspect, containerPrune } from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import TaskLog from '@/components/log/task/index.vue';
import i18n from '@/lang';
import { ElMessageBox } from 'element-plus';
import { GlobalStore } from '@/store';
import { routerToFileWithPath } from '@/utils/router';
const globalStore = GlobalStore();

const taskLogRef = ref();
const mobile = computed(() => {
    return globalStore.isMobile();
});

const loading = ref();
const myDetail = ref();

const opRef = ref();

const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    cacheSizeKey: 'container-volume-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('container-volume-page-size')) || 10,
    total: 0,
});
const searchName = ref();
const isActive = ref(false);
const isExist = ref(false);

const dialogCreateRef = ref<DialogExpose>();

interface DialogExpose {
    acceptParams: () => void;
}
const onCreate = async () => {
    dialogCreateRef.value!.acceptParams();
};

const search = async () => {
    if (!isActive.value || !isExist.value) {
        return;
    }
    const params = {
        info: searchName.value,
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    loading.value = true;
    await searchVolume(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const onInspect = async (id: string) => {
    const res = await inspect({ id: id, type: 'volume' });
    let detailInfo = JSON.stringify(JSON.parse(res.data), null, 2);
    let param = {
        header: i18n.global.t('commons.button.view'),
        detailInfo: detailInfo,
        mode: 'json',
    };
    myDetail.value!.acceptParams(param);
};

const onClean = () => {
    ElMessageBox.confirm(i18n.global.t('container.volumePruneHelper'), i18n.global.t('container.volumePrune'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        let params = {
            taskID: newUUID(),
            pruneType: 'volume',
            withTagAll: false,
        };
        await containerPrune(params)
            .then(() => {
                loading.value = false;
                openTaskLog(params.taskID);
            })
            .catch(() => {
                loading.value = false;
            });
    });
};
const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const batchDelete = async (row: Container.VolumeInfo | null) => {
    let names = [];
    if (row) {
        names.push(row.name);
    } else {
        selects.value.forEach((item: Container.VolumeInfo) => {
            names.push(item.name);
        });
    }
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('container.volume'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: deleteVolume,
        params: { names: names },
    });
};

const buttons = [
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Container.VolumeInfo) => {
            batchDelete(row);
        },
    },
];
</script>
