<template>
    <div>
        <LayoutContent v-loading="loading" title="FTP">
            <template #toolbar>
                <el-row>
                    <el-col :xs="24" :sm="16" :md="16" :lg="16" :xl="16">
                        <el-button type="primary" @click="onOpenDialog('add')">
                            {{ $t('commons.button.add') }} FTP
                        </el-button>
                        <el-button @click="onSync()">
                            {{ $t('commons.button.sync') }}
                        </el-button>
                        <el-button plain :disabled="selects.length === 0" @click="onDelete(null)">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </el-col>
                    <el-col :xs="24" :sm="8" :md="8" :lg="8" :xl="8">
                        <TableSearch @search="search()" v-model:searchName="searchName" />
                    </el-col>
                </el-row>
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
                            <div v-else class="flex items-center">
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
                            <el-tag v-if="row.status === 'deleted'" type="info">{{ $t('database.isDelete') }}</el-tag>
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
                    <el-table-column :label="$t('file.root')" :min-width="120" prop="path" show-overflow-tooltip />
                    <el-table-column
                        :label="$t('commons.table.description')"
                        :min-width="80"
                        prop="description"
                        show-overflow-tooltip
                    />
                    <el-table-column
                        :label="$t('commons.table.createdAt')"
                        :formatter="dateFormat"
                        :min-width="80"
                        prop="createdAt"
                    />
                    <fu-table-operations
                        width="240px"
                        :buttons="buttons"
                        :ellipsis="10"
                        :label="$t('commons.table.operate')"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <OpDialog ref="opRef" @search="search" @submit="onSubmitDelete()" />
        <OperateDialog @search="search" ref="dialogRef" />
    </div>
</template>

<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';
import i18n from '@/lang';
import { dateFormat } from '@/utils/util';
import { MsgSuccess } from '@/utils/message';
import { deleteFtp, searchFtp, updateFtp, syncFtp } from '@/api/modules/toolbox';
import OperateDialog from '@/views/toolbox/ftp/operate/index.vue';
import { Toolbox } from '@/api/interface/toolbox';

const loading = ref();
const selects = ref<any>([]);

const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'ftp-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
    orderBy: 'created_at',
    order: 'null',
});
const searchName = ref();

const opRef = ref();
const dialogRef = ref();
const operateIDs = ref();

const search = async (column?: any) => {
    paginationConfig.orderBy = column?.order ? column.prop : paginationConfig.orderBy;
    paginationConfig.order = column?.order ? column.order : paginationConfig.order;
    let params = {
        info: searchName.value,
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    loading.value = true;
    await searchFtp(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const onChangeStatus = async (row: Toolbox.FtpInfo, status: string) => {
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
                MsgSuccess(i18n.global.t('toolbox.ftp.operationSuccess'));
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
            i18n.global.t('cronjob.cronTask'),
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
        label: i18n.global.t('commons.button.delete'),
        disabled: (row: Toolbox.FtpInfo) => {
            return row.status === 'deleted';
        },
        click: (row: Toolbox.FtpInfo) => {
            onDelete(row);
        },
    },
];

onMounted(() => {
    search();
});
</script>
