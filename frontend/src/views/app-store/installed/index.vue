<template>
    <el-card>
        <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search" v-loading="loading">
            <template #toolbar>
                <el-row>
                    <el-col :span="18">
                        <el-button @click="sync" type="primary" :plain="true">{{ $t('app.sync') }}</el-button>
                    </el-col>
                    <el-col :span="6">
                        <div style="float: right">
                            <el-input
                                style="display: inline; margin-right: 5px"
                                v-model="searchName"
                                clearable
                                @clear="search()"
                            ></el-input>
                            <el-button
                                style="display: inline; margin-right: 5px"
                                v-model="searchName"
                                @click="search()"
                            >
                                {{ $t('app.search') }}
                            </el-button>
                        </div>
                    </el-col>
                </el-row>
            </template>
            <el-table-column :label="$t('app.name')" prop="name" min-width="150px" show-overflow-tooltip>
                <template #default="{ row }">
                    {{ row.name }}
                    <el-tag round effect="dark" v-if="row.canUpdate">{{ $t('app.canUpdate') }}</el-tag>
                </template>
            </el-table-column>
            <el-table-column :label="$t('app.appName')" prop="app.name"></el-table-column>
            <el-table-column :label="$t('app.version')" prop="version"></el-table-column>
            <el-table-column :label="$t('website.port')" prop="httpPort"></el-table-column>
            <el-table-column :label="$t('app.backup')">
                <template #default="{ row }">
                    <el-link :underline="false" @click="openBackups(row.id, row.name)" type="primary">
                        {{ $t('app.backup') }} ({{ row.backups.length }})
                    </el-link>
                </template>
            </el-table-column>
            <el-table-column :label="$t('app.status')">
                <template #default="{ row }">
                    <el-popover
                        v-if="row.status === 'Error'"
                        placement="bottom"
                        :width="400"
                        trigger="hover"
                        :content="row.message"
                    >
                        <template #reference>
                            <el-tag type="error">{{ row.status }}</el-tag>
                        </template>
                    </el-popover>

                    <el-tag v-else>
                        <el-icon v-if="row.status === 'Installing'" class="is-loading">
                            <Loading />
                        </el-icon>
                        {{ row.status }}
                    </el-tag>
                </template>
            </el-table-column>
            <el-table-column
                prop="createdAt"
                :label="$t('commons.table.date')"
                :formatter="dateFromat"
                show-overflow-tooltip
            />
            <fu-table-operations
                width="300px"
                :ellipsis="10"
                :buttons="buttons"
                :label="$t('commons.table.operate')"
                fixed="right"
                fix
            />
        </ComplexTable>
        <el-dialog
            v-model="open"
            :title="$t('commons.msg.operate')"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :before-close="handleClose"
            width="30%"
        >
            <div style="text-align: center">
                <p>{{ $t('app.versioneSelect') }}</p>
                <el-select v-model="operateReq.detailId">
                    <el-option
                        v-for="(version, index) in versions"
                        :key="index"
                        :value="version.detailId"
                        :label="version.version"
                    ></el-option>
                </el-select>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button
                        type="primary"
                        @click="operate"
                        :disabled="operateReq.operate == 'update' && versions == null"
                    >
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
        <Backups ref="backupRef" @close="search"></Backups>
        <AppResources ref="checkRef"></AppResources>
    </el-card>
</template>

<script lang="ts" setup>
import {
    SearchAppInstalled,
    InstalledOp,
    SyncInstalledApp,
    GetAppUpdateVersions,
    AppInstalledDeleteCheck,
} from '@/api/modules/app';
import { onMounted, onUnmounted, reactive, ref } from 'vue';
import ComplexTable from '@/components/complex-table/index.vue';
import { dateFromat } from '@/utils/util';
import i18n from '@/lang';
import { ElMessage, ElMessageBox } from 'element-plus';
import Backups from './backups.vue';
import AppResources from './check/index.vue';
import { App } from '@/api/interface/app';
import { useDeleteData } from '@/hooks/use-delete-data';

let data = ref<any>();
let loading = ref(false);
let timer: NodeJS.Timer | null = null;
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 20,
    total: 0,
});
let open = ref(false);
let operateReq = reactive({
    installId: 0,
    operate: '',
    detailId: 0,
});
let versions = ref<App.VersionDetail[]>();
const backupRef = ref();
const checkRef = ref();
let searchName = ref('');

const sync = () => {
    loading.value = true;
    SyncInstalledApp()
        .then(() => {
            ElMessage.success(i18n.global.t('app.syncSuccess'));
            search();
        })
        .finally(() => {
            loading.value = false;
        });
};

const search = () => {
    const req = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        name: searchName.value,
    };

    SearchAppInstalled(req).then((res) => {
        data.value = res.data.items;
        paginationConfig.total = res.data.total;
    });
};

const openOperate = (row: any, op: string) => {
    operateReq.installId = row.id;
    operateReq.operate = op;
    if (op == 'update') {
        GetAppUpdateVersions(row.id).then((res) => {
            versions.value = res.data;
            if (res.data != null && res.data.length > 0) {
                operateReq.detailId = res.data[0].detailId;
            }
            open.value = true;
        });
    } else if (op == 'delete') {
        AppInstalledDeleteCheck(row.id).then(async (res) => {
            const items = res.data;
            if (res.data && res.data.length > 0) {
                checkRef.value.acceptParams({ items: items });
            } else {
                await useDeleteData(InstalledOp, operateReq, 'app.deleteWarn');
                search();
            }
        });
    } else {
        onOperate(op);
    }
};

const operate = async () => {
    open.value = false;
    loading.value = true;
    await InstalledOp(operateReq)
        .then(() => {
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .finally(() => {
            loading.value = false;
        });
};

const handleClose = () => {
    open.value = false;
};

const onOperate = async (operation: string) => {
    ElMessageBox.confirm(
        i18n.global.t('app.operatorHelper', [i18n.global.t('app.' + operation)]),
        i18n.global.t('app.' + operation),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    ).then(() => {
        operate();
    });
};

const buttons = [
    {
        label: i18n.global.t('app.sync'),
        click: (row: any) => {
            openOperate(row, 'sync');
        },
    },
    {
        label: i18n.global.t('app.update'),
        click: (row: any) => {
            openOperate(row, 'update');
        },
        disabled: (row: any) => {
            return !row.canUpdate;
        },
    },
    {
        label: i18n.global.t('app.restart'),
        click: (row: any) => {
            openOperate(row, 'restart');
        },
    },
    {
        label: i18n.global.t('app.up'),
        click: (row: any) => {
            openOperate(row, 'up');
        },
        disabled: (row: any) => {
            return row.status === 'Running';
        },
    },
    {
        label: i18n.global.t('app.down'),
        click: (row: any) => {
            openOperate(row, 'down');
        },
    },
    {
        label: i18n.global.t('app.delete'),
        click: (row: any) => {
            openOperate(row, 'delete');
        },
    },
];

const openBackups = (installId: number, installName: string) => {
    let params = {
        appInstallId: installId,
        appInstallName: installName,
    };
    backupRef.value.acceptParams(params);
};

onMounted(() => {
    search();
    timer = setInterval(() => {
        search();
    }, 1000 * 8);
});

onUnmounted(() => {
    clearInterval(Number(timer));
});
</script>

<style lang="scss">
.i-card {
    height: 60px;
    cursor: pointer;
    .content {
        .image {
            width: auto;
            height: auto;
        }
    }
}
.i-card:hover {
    border: 1px solid;
    border-color: $primary-color;
    z-index: 1;
}
</style>
