<template>
    <div>
        <Submenu activeName="repo" />
        <el-card width="30%" v-if="dockerStatus != 'Running'" class="mask-prompt">
            <span style="font-size: 14px">{{ $t('container.serviceUnavailable') }}</span>
            <el-button type="primary" link style="font-size: 14px; margin-bottom: 5px" @click="goSetting">
                【 {{ $t('container.setting') }} 】
            </el-button>
            <span style="font-size: 14px">{{ $t('container.startIn') }}</span>
        </el-card>
        <el-card style="margin-top: 20px" :class="{ mask: dockerStatus != 'Running' }">
            <ComplexTable :pagination-config="paginationConfig" v-model:selects="selects" :data="data" @search="search">
                <template #toolbar>
                    <el-button icon="Plus" type="primary" @click="onOpenDialog('create')">
                        {{ $t('commons.button.create') }}
                    </el-button>
                    <el-button type="danger" plain :disabled="selects.length === 0" @click="onBatchDelete(null)">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </template>
                <el-table-column type="selection" :selectable="selectable" fix />
                <el-table-column :label="$t('commons.table.name')" prop="name" min-width="60" />
                <el-table-column
                    :label="$t('container.downloadUrl')"
                    show-overflow-tooltip
                    prop="downloadUrl"
                    min-width="100"
                    fix
                />
                <el-table-column :label="$t('container.protocol')" prop="protocol" min-width="60" fix />
                <el-table-column :label="$t('commons.table.status')" prop="status" min-width="60" fix>
                    <template #default="{ row }">
                        <el-tag v-if="row.status === 'Success'" type="success">
                            {{ $t('commons.status.success') }}
                        </el-tag>
                        <el-tooltip v-else effect="dark" :content="row.message" placement="bottom">
                            <el-tag type="danger">{{ $t('commons.status.failed') }}</el-tag>
                        </el-tooltip>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('commons.table.createdAt')" min-width="80" fix>
                    <template #default="{ row }">
                        {{ dateFromat(0, 0, row.createdAt) }}
                    </template>
                </el-table-column>
                <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" />
            </ComplexTable>
        </el-card>
        <OperatorDialog @search="search" ref="dialogRef" />
        <DeleteDialog @search="search" ref="dialogDeleteRef" />
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import OperatorDialog from '@/views/container/repo/operator/index.vue';
import DeleteDialog from '@/views/container/repo/delete/index.vue';
import Submenu from '@/views/container/index.vue';
import { reactive, onMounted, ref } from 'vue';
import { dateFromat } from '@/utils/util';
import { Container } from '@/api/interface/container';
import { loadDockerStatus, searchImageRepo } from '@/api/modules/container';
import i18n from '@/lang';
import router from '@/routers';

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

const search = async () => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    await searchImageRepo(params).then((res) => {
        data.value = res.data.items || [];
        paginationConfig.total = res.data.total;
    });
};

function selectable(row) {
    return !(row.name === 'Docker Hub');
}

const dialogRef = ref();
const onOpenDialog = async (
    title: string,
    rowData: Partial<Container.RepoInfo> = {
        auth: true,
        protocol: 'http',
    },
) => {
    let params = {
        title,
        rowData: { ...rowData },
    };
    dialogRef.value!.acceptParams(params);
};

const dialogDeleteRef = ref();
const onBatchDelete = async (row: Container.RepoInfo | null) => {
    let ids: Array<number> = [];
    if (row) {
        ids.push(row.id);
    } else {
        selects.value.forEach((item: Container.RepoInfo) => {
            ids.push(item.id);
        });
    }
    dialogDeleteRef.value!.acceptParams({ ids: ids });
};

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        disabled: (row: Container.RepoInfo) => {
            return row.downloadUrl === 'docker.io';
        },
        click: (row: Container.RepoInfo) => {
            onOpenDialog('edit', row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        disabled: (row: Container.RepoInfo) => {
            return row.downloadUrl === 'docker.io';
        },
        click: (row: Container.RepoInfo) => {
            onBatchDelete(row);
        },
    },
];

onMounted(() => {
    loadStatus();
});
</script>
