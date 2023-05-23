<template>
    <div v-loading="loading">
        <div v-show="isOnDetail">
            <ComposeDetial @back="backList" ref="composeDetailRef" />
        </div>
        <el-card v-if="dockerStatus != 'Running'" class="mask-prompt">
            <span>{{ $t('container.serviceUnavailable') }}</span>
            <el-button type="primary" link class="bt" @click="goSetting">【 {{ $t('container.setting') }} 】</el-button>
            <span>{{ $t('container.startIn') }}</span>
        </el-card>

        <LayoutContent v-if="!isOnDetail" :title="$t('container.compose')" :class="{ mask: dockerStatus != 'Running' }">
            <template #prompt>
                <el-alert type="info" :closable="false">
                    <template #default>
                        <span>
                            <span>{{ $t('container.composeHelper', [baseDir]) }}</span>
                            <el-button type="primary" link @click="toFolder">
                                <el-icon>
                                    <FolderOpened />
                                </el-icon>
                            </el-button>
                        </span>
                    </template>
                </el-alert>
            </template>
            <template #toolbar>
                <el-row>
                    <el-col :span="16">
                        <el-button type="primary" @click="onOpenDialog()">
                            {{ $t('container.createCompose') }}
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
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    :data="data"
                    @search="search"
                >
                    <el-table-column :label="$t('commons.table.name')" min-width="100" prop="name" fix>
                        <template #default="{ row }">
                            <Tooltip @click="loadDetail(row)" :text="row.name" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('container.from')" prop="createdBy" min-width="80" fix>
                        <template #default="{ row }">
                            <span v-if="row.createdBy === ''">{{ $t('container.local') }}</span>
                            <span v-if="row.createdBy === 'Apps'">{{ $t('container.apps') }}</span>
                            <span v-if="row.createdBy === '1Panel'">1Panel</span>
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
            </template>
        </LayoutContent>

        <EditDialog ref="dialogEditRef" />
        <CreateDialog @search="search" ref="dialogRef" />
        <DeleteDialog @search="search" ref="dialogDelRef" />
    </div>
</template>

<script lang="ts" setup>
import Tooltip from '@/components/tooltip/index.vue';
import TableSetting from '@/components/table-setting/index.vue';
import { reactive, onMounted, ref } from 'vue';
import EditDialog from '@/views/container/compose/edit/index.vue';
import CreateDialog from '@/views/container/compose/create/index.vue';
import DeleteDialog from '@/views/container/compose/delete/index.vue';
import ComposeDetial from '@/views/container/compose/detail/index.vue';
import { loadDockerStatus, searchCompose } from '@/api/modules/container';
import i18n from '@/lang';
import { Container } from '@/api/interface/container';
import { LoadFile } from '@/api/modules/files';
import { loadBaseDir } from '@/api/modules/setting';
import router from '@/routers';

const data = ref();
const selects = ref<any>([]);
const loading = ref(false);

const isOnDetail = ref(false);
const baseDir = ref();

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

const toFolder = async () => {
    router.push({ path: '/hosts/files', query: { path: baseDir.value + '/docker/compose' } });
};

const loadPath = async () => {
    const pathRes = await loadBaseDir();
    baseDir.value = pathRes.data;
};

const search = async () => {
    let params = {
        info: searchName.value,
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

const dialogDelRef = ref();
const onDelete = async (row: Container.ComposeInfo) => {
    const param = {
        name: row.name,
        path: row.path,
    };
    dialogDelRef.value.acceptParams(param);
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
            return row.createdBy === 'Local';
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Container.ComposeInfo) => {
            onDelete(row);
        },
        disabled: (row: any) => {
            return row.createdBy !== '1Panel';
        },
    },
];
onMounted(() => {
    loadPath();
    loadStatus();
});
</script>
