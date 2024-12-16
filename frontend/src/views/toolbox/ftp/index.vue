<template>
    <div>
        <div class="app-status" style="margin-top: 20px">
            <el-card v-if="form.isExist">
                <div class="flex w-full flex-col gap-4 md:flex-row">
                    <div class="flex flex-wrap gap-4">
                        <el-tag effect="dark" type="success">FTP</el-tag>
                        <el-tag round v-if="form.isActive" type="success">
                            {{ $t('commons.status.running') }}
                        </el-tag>
                        <el-tag round v-if="!form.isActive" type="info">
                            {{ $t('commons.status.stopped') }}
                        </el-tag>
                    </div>
                    <div class="mt-0.5">
                        <el-button v-if="form.isActive" type="primary" @click="onOperate('stop')" link>
                            {{ $t('commons.button.stop') }}
                        </el-button>
                        <el-button v-if="!form.isActive" type="primary" @click="onOperate('start')" link>
                            {{ $t('commons.button.start') }}
                        </el-button>
                        <el-divider direction="vertical" />
                        <el-button type="primary" @click="onOperate('restart')" link>
                            {{ $t('container.restart') }}
                        </el-button>
                    </div>
                </div>
            </el-card>
        </div>
        <div v-if="form.isExist">
            <LayoutContent v-loading="loading" :title="$t('toolbox.ftp.ftp', 2)">
                <template #toolbar>
                    <div class="flex justify-between gap-2 flex-wrap sm:flex-row">
                        <div class="flex flex-wrap gap-3">
                            <el-button type="primary" :disabled="!form.isActive" @click="onOpenDialog('add')">
                                {{ $t('commons.button.add') }} {{ $t('toolbox.ftp.ftp') }}
                            </el-button>
                            <el-button @click="onSync()" :disabled="!form.isActive">
                                {{ $t('commons.button.sync') }}
                            </el-button>
                            <el-button plain :disabled="selects.length === 0 || !form.isActive" @click="onDelete(null)">
                                {{ $t('commons.button.delete') }}
                            </el-button>
                        </div>
                        <div class="flex flex-wrap gap-3">
                            <TableSearch @search="search()" v-model:searchName="searchName" />
                        </div>
                    </div>
                </template>
                <template #main>
                    <ComplexTable
                        :pagination-config="paginationConfig"
                        v-model:selects="selects"
                        @sort-change="search"
                        @search="search"
                        :data="data"
                    >
                        <el-table-column type="selection" fix />
                        <el-table-column
                            :label="$t('commons.login.username')"
                            :min-width="60"
                            prop="user"
                            show-overflow-tooltip
                        />
                        <el-table-column :label="$t('commons.login.password')" prop="password">
                            <template #default="{ row }">
                                <div v-if="row.password.length === 0">-</div>
                                <div v-else class="flex items-center flex-wrap">
                                    <div class="star-center" v-if="!row.showPassword">
                                        <span>**********</span>
                                    </div>
                                    <div>
                                        <span v-if="row.showPassword">
                                            {{ row.password }}
                                        </span>
                                    </div>
                                    <el-button
                                        v-if="!row.showPassword"
                                        link
                                        @click="row.showPassword = true"
                                        icon="View"
                                        class="ml-1.5"
                                    ></el-button>
                                    <el-button
                                        v-if="row.showPassword"
                                        link
                                        @click="row.showPassword = false"
                                        icon="Hide"
                                        class="ml-1.5"
                                    ></el-button>
                                    <div>
                                        <CopyButton :content="row.password" type="icon" />
                                    </div>
                                </div>
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('commons.table.status')" :min-width="60" prop="status">
                            <template #default="{ row }">
                                <el-tag v-if="row.status === 'deleted'" type="info">
                                    {{ $t('database.isDelete') }}
                                </el-tag>
                                <el-button
                                    v-if="row.status === 'Enable'"
                                    @click="onChangeStatus(row, 'disable')"
                                    link
                                    icon="VideoPlay"
                                    type="success"
                                >
                                    {{ $t('commons.status.enabled') }}
                                </el-button>
                                <el-button
                                    v-if="row.status === 'Disable'"
                                    icon="VideoPause"
                                    @click="onChangeStatus(row, 'enable')"
                                    link
                                    type="danger"
                                >
                                    {{ $t('commons.status.disabled') }}
                                </el-button>
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('file.root')" :min-width="120" prop="path" show-overflow-tooltip>
                            <template #default="{ row }">
                                <el-button text type="primary" @click="toFolder(row.path)">{{ row.path }}</el-button>
                            </template>
                        </el-table-column>
                        <el-table-column
                            :label="$t('commons.table.description')"
                            prop="description"
                            show-overflow-tooltip
                        >
                            <template #default="{ row }">
                                <fu-input-rw-switch v-model="row.description" @blur="onChange(row)" />
                            </template>
                        </el-table-column>
                        <fu-table-operations
                            width="200px"
                            :buttons="buttons"
                            :ellipsis="10"
                            :label="$t('commons.table.operate')"
                            fix
                        />
                    </ComplexTable>
                </template>
            </LayoutContent>
        </div>
        <div v-else>
            <LayoutContent title="FTP" :divider="true">
                <template #main>
                    <div class="app-warn">
                        <div class="flex flex-col gap-2 items-center justify-center w-full sm:flex-row">
                            <span>{{ $t('toolbox.ftp.noFtp') }}</span>
                            <span @click="toDoc" class="flex items-center justify-center gap-0.5">
                                <el-icon><Position /></el-icon>
                                {{ $t('firewall.quickJump') }}
                            </span>
                        </div>
                        <div>
                            <img src="@/assets/images/no_app.svg" />
                        </div>
                    </div>
                </template>
            </LayoutContent>
        </div>

        <OpDialog ref="opRef" @search="search" @submit="onSubmitDelete()" />
        <OperateDialog @search="search" ref="dialogRef" />
        <LogDialog ref="dialogLogRef" />
    </div>
</template>

<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { deleteFtp, searchFtp, updateFtp, syncFtp, operateFtp, getFtpBase } from '@/api/modules/toolbox';
import OperateDialog from '@/views/toolbox/ftp/operate/index.vue';
import LogDialog from '@/views/toolbox/ftp/log/index.vue';
import { Toolbox } from '@/api/interface/toolbox';
import router from '@/routers';
import { GlobalStore } from '@/store';

const globalStore = GlobalStore();

const loading = ref();
const selects = ref<any>([]);

const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'ftp-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('ftp-page-size')) || 10,
    total: 0,
    orderBy: 'created_at',
    order: 'null',
});
const searchName = ref();

const form = reactive({
    isActive: false,
    isExist: false,
});

const opRef = ref();
const dialogRef = ref();
const operateIDs = ref();
const dialogLogRef = ref();

const search = async (column?: any) => {
    loading.value = true;
    await getFtpBase()
        .then(async (res) => {
            form.isActive = res.data.isActive;
            form.isExist = res.data.isExist;
            paginationConfig.orderBy = column?.order ? column.prop : paginationConfig.orderBy;
            paginationConfig.order = column?.order ? column.order : paginationConfig.order;
            let params = {
                info: searchName.value,
                page: paginationConfig.currentPage,
                pageSize: paginationConfig.pageSize,
            };
            await searchFtp(params)
                .then((res) => {
                    loading.value = false;
                    data.value = res.data.items || [];
                    paginationConfig.total = res.data.total;
                })
                .catch(() => {
                    loading.value = false;
                });
        })
        .catch(() => {
            loading.value = false;
        });
};

const toDoc = () => {
    window.open(globalStore.docsUrl + '/user_manual/toolbox/ftp/', '_blank', 'noopener,noreferrer');
};

const toFolder = (folder: string) => {
    router.push({ path: '/hosts/files', query: { path: folder } });
};

const onOperate = async (operation: string) => {
    let msg = operation === 'enable' || operation === 'disable' ? 'ssh.' : 'commons.button.';
    ElMessageBox.confirm(i18n.global.t('toolbox.ftp.operation', [i18n.global.t(msg + operation)]), 'FTP', {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    })
        .then(async () => {
            loading.value = true;
            await operateFtp(operation)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    search();
                })
                .catch(() => {
                    loading.value = false;
                });
        })
        .catch(() => {
            search();
        });
};

const onChangeStatus = async (row: Toolbox.FtpInfo, status: string) => {
    if (row.password.length === 0) {
        MsgError(i18n.global.t('toolbox.ftp.noPasswdMsg'));
        return;
    }
    ElMessageBox.confirm(i18n.global.t('toolbox.ftp.' + status + 'Helper'), i18n.global.t('cronjob.changeStatus'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        row.status = status === 'enable' ? 'Enable' : 'Disable';
        await updateFtp(row);
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        search();
    });
};

const onChange = async (row: any) => {
    await await updateFtp(row);
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
};

const onOpenDialog = async (title: string, rowData: Partial<Toolbox.FtpInfo> = {}) => {
    let params = {
        title,
        rowData: { ...rowData },
    };
    dialogRef.value!.acceptParams(params);
};

const onSync = async () => {
    ElMessageBox.confirm(i18n.global.t('toolbox.ftp.syncHelper'), i18n.global.t('commons.button.sync'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        loading.value = true;
        await syncFtp()
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

const onDelete = async (row: Toolbox.FtpInfo | null) => {
    let names = [];
    let ids = [];
    if (row) {
        ids = [row.id];
        names = [row.user];
    } else {
        for (const item of selects.value) {
            names.push(item.user);
            ids.push(item.id);
        }
    }
    operateIDs.value = ids;
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('toolbox.ftp.ftp'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: null,
        params: null,
    });
};

const onSubmitDelete = async () => {
    loading.value = true;
    await deleteFtp({ ids: operateIDs.value })
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        disabled: (row: Toolbox.FtpInfo) => {
            return row.status === 'deleted';
        },
        click: (row: Toolbox.FtpInfo) => {
            onOpenDialog('edit', row);
        },
    },
    {
        label: i18n.global.t('commons.button.log'),
        disabled: (row: Toolbox.FtpInfo) => {
            return row.status === 'deleted';
        },
        click: (row: Toolbox.FtpInfo) => {
            dialogLogRef.value!.acceptParams({ user: row.user, path: row.path });
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Toolbox.FtpInfo) => {
            onDelete(row);
        },
    },
];

onMounted(() => {
    search();
});
</script>
