<template>
    <div v-loading="loading">
        <Submenu activeName="compose" />
        <el-card width="30%" v-if="dockerStatus != 'Running'" class="mask-prompt">
            <span style="font-size: 14px">{{ $t('container.serviceUnavailable') }}</span>
            <el-button type="primary" link style="font-size: 14px; margin-bottom: 5px" @click="goSetting">
                【 {{ $t('container.setting') }} 】
            </el-button>
            <span style="font-size: 14px">{{ $t('container.startIn') }}</span>
        </el-card>
        <el-card style="margin-top: 20px" :class="{ mask: dockerStatus != 'Running' }">
            <div v-if="!isOnDetail">
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    :data="data"
                    @search="search"
                >
                    <template #toolbar>
                        <el-button icon="Plus" type="primary" @click="onOpenDialog()">
                            {{ $t('commons.button.create') }}
                        </el-button>
                    </template>
                    <el-table-column
                        :label="$t('commons.table.name')"
                        show-overflow-tooltip
                        min-width="100"
                        prop="name"
                        fix
                    >
                        <template #default="{ row }">
                            <el-link @click="loadDetail(row)" type="primary">{{ row.name }}</el-link>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('container.from')" prop="createdBy" min-width="80" fix>
                        <template #default="{ row }">
                            <span v-if="row.createdBy === ''">Local</span>
                            <span v-else>{{ row.createdBy }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('container.containerNumber')"
                        prop="containerNumber"
                        min-width="80"
                        fix
                    />
                    <el-table-column :label="$t('commons.table.createdAt')" prop="createdAt" min-width="80" fix />
                    <fu-table-operations
                        width="200px"
                        :ellipsis="10"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        fix
                    />
                </ComplexTable>
            </div>
            <div v-show="isOnDetail">
                <ComposeDetial @back="backList" ref="composeDetailRef" />
            </div>
        </el-card>

        <EditDialog ref="dialogEditRef" />
        <CreateDialog @search="search" ref="dialogRef" />
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import { reactive, onMounted, ref } from 'vue';
import EditDialog from '@/views/container/compose/edit/index.vue';
import CreateDialog from '@/views/container/compose/create/index.vue';
import ComposeDetial from '@/views/container/compose/detail/index.vue';
import Submenu from '@/views/container/index.vue';
import { composeOperator, loadDockerStatus, searchCompose } from '@/api/modules/container';
import i18n from '@/lang';
import { ElMessage } from 'element-plus';
import { Container } from '@/api/interface/container';
import { useDeleteData } from '@/hooks/use-delete-data';
import { LoadFile } from '@/api/modules/files';
import router from '@/routers';

const data = ref();
const selects = ref<any>([]);
const loading = ref(false);

const isOnDetail = ref(false);

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
    loading.value = true;
    await searchCompose(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .finally(() => {
            loading.value = false;
        });
};

const composeDetailRef = ref();
const loadDetail = async (row: Container.ComposeInfo) => {
    let params = {
        createdBy: row.createdBy,
        name: row.name,
        path: row.path,
        filters: 'com.docker.compose.project=' + row.name,
    };
    isOnDetail.value = true;
    composeDetailRef.value!.acceptParams(params);
};
const backList = async () => {
    isOnDetail.value = false;
    search();
};

const dialogRef = ref();
const onOpenDialog = async () => {
    dialogRef.value!.acceptParams();
};

const onDelete = async (row: Container.ComposeInfo) => {
    const param = {
        name: row.name,
        path: row.path,
        operation: 'down',
    };
    await useDeleteData(composeOperator, param, 'commons.msg.delete');
    search();
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
};

const dialogEditRef = ref();
const onEdit = async (row: Container.ComposeInfo) => {
    const res = await LoadFile({ path: row.path });
    let params = {
        name: row.name,
        path: row.path,
        content: res.data,
    };
    dialogEditRef.value!.acceptParams(params);
};

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: Container.ComposeInfo) => {
            onEdit(row);
        },
        disabled: (row: any) => {
            return row.createdBy !== 'Local';
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Container.ComposeInfo) => {
            onDelete(row);
        },
        disabled: (row: any) => {
            return row.createdBy === 'Apps';
        },
    },
];
onMounted(() => {
    loadStatus();
});
</script>
