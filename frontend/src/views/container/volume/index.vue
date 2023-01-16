<template>
    <div>
        <Submenu activeName="volume" />
        <el-card width="30%" v-if="dockerStatus != 'Running'" class="mask-prompt">
            <span style="font-size: 14px">{{ $t('container.serviceUnavailable') }}</span>
            <el-button type="primary" link style="font-size: 14px; margin-bottom: 5px" @click="goSetting">
                【 {{ $t('container.setting') }} 】
            </el-button>
            <span style="font-size: 14px">{{ $t('container.startIn') }}</span>
        </el-card>
        <el-card style="margin-top: 20px" :class="{ mask: dockerStatus != 'Running' }">
            <LayoutContent :header="$t('container.volume')">
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    :data="data"
                    @search="search"
                >
                    <template #toolbar>
                        <el-button icon="Plus" type="primary" @click="onCreate()">
                            {{ $t('commons.button.create') }}
                        </el-button>
                        <el-button type="danger" plain :disabled="selects.length === 0" @click="batchDelete(null)">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </template>
                    <el-table-column type="selection" fix />
                    <el-table-column
                        :label="$t('commons.table.name')"
                        show-overflow-tooltip
                        min-width="80"
                        prop="name"
                        fix
                    >
                        <template #default="{ row }">
                            <el-link @click="onInspect(row.name)" type="primary">{{ row.name }}</el-link>
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
            </LayoutContent>
        </el-card>

        <CodemirrorDialog ref="codemirror" />
        <CreateDialog @search="search" ref="dialogCreateRef" />
    </div>
</template>

<script lang="ts" setup>
import LayoutContent from '@/layout/layout-content.vue';
import ComplexTable from '@/components/complex-table/index.vue';
import CreateDialog from '@/views/container/volume/create/index.vue';
import Submenu from '@/views/container/index.vue';
import CodemirrorDialog from '@/components/codemirror-dialog/codemirror.vue';
import { reactive, onMounted, ref } from 'vue';
import { dateFormat } from '@/utils/util';
import { deleteVolume, searchVolume, inspect, loadDockerStatus } from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import i18n from '@/lang';
import { useDeleteData } from '@/hooks/use-delete-data';
import router from '@/routers';

const detailInfo = ref();
const codemirror = ref();

const data = ref();
const selects = ref<any>([]);
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
    }
};
const goSetting = async () => {
    router.push({ name: 'ContainerSetting' });
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
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    await searchVolume(params).then((res) => {
        data.value = res.data.items || [];
        paginationConfig.total = res.data.total;
    });
};

const onInspect = async (id: string) => {
    const res = await inspect({ id: id, type: 'volume' });
    detailInfo.value = JSON.stringify(JSON.parse(res.data), null, 2);
    let param = {
        header: i18n.global.t('commons.button.view'),
        detailInfo: detailInfo.value,
    };
    codemirror.value!.acceptParams(param);
};

const batchDelete = async (row: Container.VolumeInfo | null) => {
    let names: Array<string> = [];
    if (row === null) {
        selects.value.forEach((item: Container.VolumeInfo) => {
            names.push(item.name);
        });
    } else {
        names.push(row.name);
    }
    await useDeleteData(deleteVolume, { names: names }, 'commons.msg.delete');
    search();
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
