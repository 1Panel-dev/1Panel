<template>
    <div v-loading="loading">
        <docker-status
            v-model:isActive="isActive"
            v-model:isExist="isExist"
            v-model:loading="loading"
            @search="search"
        />

        <LayoutContent v-if="isExist" :title="$t('container.compose', 2)" :class="{ mask: !isActive }">
            <template #leftToolBar>
                <el-button type="primary" @click="onOpenDialog()">
                    {{ $t('container.createCompose') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <TableSearch @search="search()" v-model:searchName="searchName" />
                <TableRefresh @search="search()" />
                <TableSetting title="compose-refresh" @search="search()" />
            </template>
            <template #main>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    :data="data"
                    @search="search"
                    :heightDiff="350"
                >
                    <el-table-column
                        :label="$t('commons.table.name')"
                        width="170"
                        prop="name"
                        sortable
                        fix
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <el-text type="primary" class="cursor-pointer" @click="loadDetail(row)">
                                {{ row.name }}
                            </el-text>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('app.source')" prop="createdBy" min-width="80" fix>
                        <template #default="{ row }">
                            <span v-if="row.createdBy === ''">{{ $t('commons.table.local') }}</span>
                            <span v-if="row.createdBy === 'Apps'">{{ $t('menu.apps') }}</span>
                            <span v-if="row.createdBy === '1Panel'">1Panel</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('container.composeDirectory')" min-width="80" fix>
                        <template #default="{ row }">
                            <el-tooltip :content="row.workdir">
                                <el-button type="primary" link @click="toComposeFolder(row)">
                                    <el-icon>
                                        <FolderOpened />
                                    </el-icon>
                                </el-button>
                            </el-tooltip>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('container.containerStatus')" min-width="80" fix>
                        <template #default="{ row }">
                            <el-text class="mx-1" v-if="row.containerCount == 0" type="danger">
                                {{ $t('container.exited') }}
                            </el-text>
                            <el-popover width="300px" v-else>
                                <template #reference>
                                    <el-text
                                        class="cursor-pointer"
                                        size="small"
                                        :type="row.containerCount === row.runningCount ? 'success' : 'warning'"
                                    >
                                        {{ $t('container.running', [row.runningCount, row.containerCount]) }}
                                    </el-text>
                                </template>
                                <div v-for="(item, index) in row.containers" :key="index" class="mt-2">
                                    <span>{{ item.name }}</span>
                                    <Status class="float-right" :key="item.state" :status="item.state" />
                                </div>
                            </el-popover>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.createdAt')" prop="createdAt" min-width="80" fix />
                    <fu-table-operations
                        width="200px"
                        :ellipsis="2"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <ComposeLogs ref="composeLogRef" />
        <EditDialog ref="dialogEditRef" @search="search" />
        <CreateDialog @search="search" ref="dialogRef" />
        <DeleteDialog @search="search" ref="dialogDelRef" />
    </div>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import EditDialog from '@/views/container/compose/edit/index.vue';
import CreateDialog from '@/views/container/compose/create/index.vue';
import DeleteDialog from '@/views/container/compose/delete/index.vue';
import ComposeLogs from '@/components/log/compose/index.vue';
import { composeOperator, inspect, searchCompose } from '@/api/modules/container';
import DockerStatus from '@/views/container/docker-status/index.vue';
import i18n from '@/lang';
import { Container } from '@/api/interface/container';
import { routerToFileWithPath, routerToNameWithQuery } from '@/utils/router';
import { MsgSuccess } from '@/utils/message';

const data = ref();
const selects = ref<any>([]);
const loading = ref(false);

const paginationConfig = reactive({
    cacheSizeKey: 'container-compose-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('container-compose-page-size')) || 10,
    total: 0,
});
const searchName = ref();

const composeLogRef = ref();

const isActive = ref(false);
const isExist = ref(false);

const toComposeFolder = async (row: Container.ComposeInfo) => {
    routerToFileWithPath(row.workdir);
};

const search = async () => {
    if (!isActive.value || !isExist.value) {
        return;
    }
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

const loadDetail = async (row: Container.ComposeInfo) => {
    routerToNameWithQuery('ContainerItem', { filters: 'com.docker.compose.project=' + row.name });
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
    const res = await inspect({ id: row.name, type: 'compose' });
    let params = {
        name: row.name,
        path: row.path,
        content: res.data,
        env: row.env,
        createdBy: row.createdBy,
    };
    dialogEditRef.value!.acceptParams(params);
};

const onComposeOperate = async (operation: string, row: any) => {
    let mes =
        operation === 'down'
            ? i18n.global.t('container.composeDownHelper', [row.name])
            : i18n.global.t('container.composeOperatorHelper', [
                  row.name,
                  i18n.global.t('commons.operate.' + operation),
              ]);
    ElMessageBox.confirm(mes, i18n.global.t('commons.operate.' + operation), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        let params = {
            name: row.name,
            path: row.path,
            operation: operation,
            withFile: false,
        };
        loading.value = true;
        await composeOperator(params)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const openLog = (row: any) => {
    composeLogRef.value.acceptParams({
        compose: row.path,
        resource: row.name,
        container: row.container,
    });
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
        label: i18n.global.t('commons.button.log'),
        click: (row: Container.ComposeInfo) => {
            openLog(row);
        },
    },
    {
        label: i18n.global.t('commons.operate.start'),
        click: (row: Container.ComposeInfo) => {
            onComposeOperate('up', row);
        },
    },
    {
        label: i18n.global.t('commons.operate.stop'),
        click: (row: Container.ComposeInfo) => {
            onComposeOperate('stop', row);
        },
    },
    {
        label: i18n.global.t('commons.operate.restart'),
        click: (row: Container.ComposeInfo) => {
            onComposeOperate('restart', row);
        },
    },
    {
        label: i18n.global.t('commons.operate.delete'),
        click: (row: Container.ComposeInfo) => {
            onDelete(row);
        },
        disabled: (row: any) => {
            return row.createdBy !== '1Panel';
        },
    },
];
</script>
