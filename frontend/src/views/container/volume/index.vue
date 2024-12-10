<template>
    <div v-loading="loading">
        <el-card v-if="dockerStatus != 'Running'" class="mask-prompt">
            <span>{{ $t('container.serviceUnavailable') }}</span>
            <el-button type="primary" link class="bt" @click="goSetting">【 {{ $t('container.setting') }} 】</el-button>
            <span>{{ $t('container.startIn') }}</span>
        </el-card>

        <LayoutContent :title="$t('container.volume', 2)" :class="{ mask: dockerStatus != 'Running' }">
            <template #toolbar>
                <div class="flex justify-between gap-2 flex-wrap sm:flex-row">
                    <div class="flex flex-wrap gap-3">
                        <el-button type="primary" @click="onCreate()">
                            {{ $t('container.createVolume') }}
                        </el-button>
                        <el-button type="primary" plain @click="onClean()">
                            {{ $t('container.volumePrune') }}
                        </el-button>
                        <el-button :disabled="selects.length === 0" @click="batchDelete(null)">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </div>
                    <div class="flex flex-wrap gap-3">
                        <TableSetting @search="search()" />
                        <TableSearch @search="search()" v-model:searchName="searchName" />
                    </div>
                </div>
            </template>
            <template #main>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    :data="data"
                    @search="search"
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
                            <el-button type="primary" link @click="toFolder(row.mountpoint)">
                                <el-icon>
                                    <FolderOpened />
                                </el-icon>
                            </el-button>
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

        <CodemirrorDialog ref="codemirror" />
        <CreateDialog @search="search" ref="dialogCreateRef" />
    </div>
</template>

<script lang="ts" setup>
import CreateDialog from '@/views/container/volume/create/index.vue';
import CodemirrorDialog from '@/components/codemirror-dialog/index.vue';
import { reactive, onMounted, ref, computed } from 'vue';
import { computeSize, dateFormat } from '@/utils/util';
import { deleteVolume, searchVolume, inspect, loadDockerStatus, containerPrune } from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import i18n from '@/lang';
import router from '@/routers';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox } from 'element-plus';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const mobile = computed(() => {
    return globalStore.isMobile();
});

const loading = ref();
const codemirror = ref();

const opRef = ref();

const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    cacheSizeKey: 'container-volume-page-size',
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
const toFolder = (folder: string) => {
    router.push({ path: '/hosts/files', query: { path: folder } });
};
const dialogCreateRef = ref<DialogExpose>();

interface DialogExpose {
    acceptParams: () => void;
}
const onCreate = async () => {
    dialogCreateRef.value!.acceptParams();
};

const search = async () => {
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
    };
    codemirror.value!.acceptParams(param);
};

const onClean = () => {
    ElMessageBox.confirm(i18n.global.t('container.volumePruneHelper'), i18n.global.t('container.volumePrune'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        let params = {
            pruneType: 'volume',
            withTagAll: false,
        };
        await containerPrune(params)
            .then((res) => {
                loading.value = false;
                MsgSuccess(
                    i18n.global.t('container.cleanSuccessWithSpace', [
                        res.data.deletedNumber,
                        computeSize(res.data.spaceReclaimed),
                    ]),
                );
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    });
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

onMounted(() => {
    loadStatus();
});
</script>
