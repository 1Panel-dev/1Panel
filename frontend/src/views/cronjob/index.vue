<template>
    <div>
        <RouterButton
            :buttons="[
                {
                    label: i18n.global.t('cronjob.cronTask', 2),
                    path: '/cronjobs',
                },
            ]"
        />
        <LayoutContent v-loading="loading" v-if="!isRecordShow" :title="$t('cronjob.cronTask', 2)">
            <template #toolbar>
                <div class="flex justify-between gap-2 flex-wrap sm:flex-row">
                    <div class="flex flex-wrap gap-3">
                        <el-button type="primary" @click="onOpenDialog('create')">
                            {{ $t('cronjob.create') }}
                        </el-button>
                        <el-button-group class="ml-4">
                            <el-button plain :disabled="selects.length === 0" @click="onBatchChangeStatus('enable')">
                                {{ $t('commons.button.enable') }}
                            </el-button>
                            <el-button plain :disabled="selects.length === 0" @click="onBatchChangeStatus('disable')">
                                {{ $t('commons.button.disable') }}
                            </el-button>
                            <el-button plain :disabled="selects.length === 0" @click="onDelete(null)">
                                {{ $t('commons.button.delete') }}
                            </el-button>
                        </el-button-group>
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
                    @sort-change="search"
                    @search="search"
                    :data="data"
                >
                    <el-table-column type="selection" fix />
                    <el-table-column
                        :label="$t('cronjob.taskName')"
                        :min-width="120"
                        prop="name"
                        sortable
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <el-text type="primary" class="cursor-pointer" @click="loadDetail(row)">
                                {{ row.name }}
                            </el-text>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.status')" :min-width="110" prop="status" sortable>
                        <template #default="{ row }">
                            <el-button
                                v-if="row.status === 'Enable'"
                                @click="onChangeStatus(row.id, 'disable')"
                                link
                                icon="VideoPlay"
                                type="success"
                            >
                                {{ $t('commons.status.enabled') }}
                            </el-button>
                            <el-button
                                v-else
                                icon="VideoPause"
                                link
                                type="danger"
                                @click="onChangeStatus(row.id, 'enable')"
                            >
                                {{ $t('commons.status.disabled') }}
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('cronjob.cronSpec')" show-overflow-tooltip :min-width="120">
                        <template #default="{ row }">
                            <div v-for="(item, index) of row.spec.split(',')" :key="index">
                                <div v-if="row.expand || (!row.expand && index < 3)">
                                    <span>
                                        {{ transSpecToStr(item) }}
                                    </span>
                                </div>
                            </div>
                            <div v-if="!row.expand && row.spec.split(',').length > 3">
                                <el-button type="primary" link @click="row.expand = true">
                                    {{ $t('commons.button.expand') }}...
                                </el-button>
                            </div>
                            <div v-if="row.expand && row.spec.split(',').length > 3">
                                <el-button type="primary" link @click="row.expand = false">
                                    {{ $t('commons.button.collapse') }}
                                </el-button>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('cronjob.retainCopies')" :min-width="120" prop="retainCopies">
                        <template #default="{ row }">
                            <el-button v-if="hasBackup(row.type)" @click="loadBackups(row)" plain size="small">
                                {{ row.retainCopies }}{{ $t('cronjob.retainCopiesUnit') }}
                            </el-button>
                            <span v-else>{{ row.retainCopies }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('cronjob.lastRecordTime')" :min-width="120" prop="lastRecordTime">
                        <template #default="{ row }">
                            {{ row.lastRecordTime }}
                        </template>
                    </el-table-column>
                    <el-table-column :min-width="120" :label="$t('setting.backupAccount')" prop="defaultDownload">
                        <template #default="{ row }">
                            <span v-if="!hasBackup(row.type)">-</span>
                            <div v-else>
                                <div v-for="(item, index) of row.backupAccounts?.split(',')" :key="index">
                                    <div v-if="row.accountExpand || (!row.accountExpand && index < 3)">
                                        <span v-if="row.backupAccounts">
                                            <span>
                                                {{ $t('setting.' + item) }}
                                            </span>
                                            <el-icon
                                                size="12"
                                                v-if="item === row.defaultDownload"
                                                class="relative top-px left-1"
                                            >
                                                <Star />
                                            </el-icon>
                                        </span>
                                        <span v-else>-</span>
                                    </div>
                                </div>
                                <div v-if="!row.accountExpand && row.backupAccounts?.split(',').length > 3">
                                    <el-button type="primary" link @click="row.accountExpand = true">
                                        {{ $t('commons.button.expand') }}...
                                    </el-button>
                                </div>
                                <div v-if="row.accountExpand && row.backupAccounts?.split(',').length > 3">
                                    <el-button type="primary" link @click="row.accountExpand = false">
                                        {{ $t('commons.button.collapse') }}
                                    </el-button>
                                </div>
                            </div>
                        </template>
                    </el-table-column>
                    <fu-table-operations
                        width="300px"
                        :buttons="buttons"
                        :ellipsis="10"
                        :label="$t('commons.table.operate')"
                        :min-width="mobile ? 'auto' : 200"
                        :fixed="mobile ? false : 'right'"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <OpDialog ref="opRef" @search="search" @submit="onSubmitDelete()">
            <template #content>
                <el-form class="mt-4 mb-1" v-if="showClean" ref="deleteForm" label-position="left">
                    <el-form-item>
                        <el-checkbox v-model="cleanData" :label="$t('cronjob.cleanData')" />
                        <span class="input-help">
                            {{ $t('cronjob.cleanDataHelper') }}
                        </span>
                    </el-form-item>
                </el-form>
            </template>
        </OpDialog>
        <OperateDialog @search="search" ref="dialogRef" />
        <Records @search="search" ref="dialogRecordRef" />
        <Backups @search="search" ref="dialogBackupRef" />
    </div>
</template>

<script lang="ts" setup>
import OperateDialog from '@/views/cronjob/operate/index.vue';
import Records from '@/views/cronjob/record/index.vue';
import Backups from '@/views/cronjob/backup/index.vue';
import { computed, onMounted, reactive, ref } from 'vue';
import { deleteCronjob, getCronjobPage, handleOnce, updateStatus } from '@/api/modules/cronjob';
import i18n from '@/lang';
import { Cronjob } from '@/api/interface/cronjob';
import { ElMessageBox } from 'element-plus';
import { MsgSuccess } from '@/utils/message';
import { transSpecToStr } from './helper';
import { GlobalStore } from '@/store';

const loading = ref();
const selects = ref<any>([]);
const isRecordShow = ref();
const operateIDs = ref();

const opRef = ref();
const showClean = ref();
const cleanData = ref();

const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'cronjob-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
    orderBy: 'created_at',
    order: 'null',
});
const searchName = ref();

const globalStore = GlobalStore();

const mobile = computed(() => {
    return globalStore.isMobile();
});

const search = async (column?: any) => {
    paginationConfig.orderBy = column?.order ? column.prop : paginationConfig.orderBy;
    paginationConfig.order = column?.order ? column.order : paginationConfig.order;
    let params = {
        info: searchName.value,
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        orderBy: paginationConfig.orderBy,
        order: paginationConfig.order,
    };
    loading.value = true;
    await getCronjobPage(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            for (const item of data.value) {
                let itemAccounts = item.backupAccounts.split(',') || [];
                let accounts = [];
                for (const account of itemAccounts) {
                    if (account == item.defaultDownload) {
                        accounts.unshift(account);
                    } else {
                        accounts.push(account);
                    }
                }
                item.itemAccounts = accounts.join(',');
            }
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const dialogRecordRef = ref();
const dialogBackupRef = ref();

const dialogRef = ref();
const onOpenDialog = async (
    title: string,
    rowData: Partial<Cronjob.CronjobInfo> = {
        specObjs: [
            {
                specType: 'perMonth',
                week: 1,
                day: 3,
                hour: 1,
                minute: 30,
                second: 30,
            },
        ],
        type: 'shell',
        retainCopies: 7,
    },
) => {
    let params = {
        title,
        rowData: { ...rowData },
    };
    dialogRef.value!.acceptParams(params);
};

const onDelete = async (row: Cronjob.CronjobInfo | null) => {
    let names = [];
    let ids = [];
    showClean.value = false;
    cleanData.value = false;
    if (row) {
        ids = [row.id];
        names = [row.name];
        if (hasBackup(row.type)) {
            showClean.value = true;
        }
    } else {
        for (const item of selects.value) {
            names.push(item.name);
            ids.push(item.id);
            if (hasBackup(item.type)) {
                showClean.value = true;
            }
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
    await deleteCronjob({ ids: operateIDs.value, cleanData: cleanData.value })
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};

const onChangeStatus = async (id: number, status: string) => {
    ElMessageBox.confirm(i18n.global.t('cronjob.' + status + 'Msg'), i18n.global.t('cronjob.changeStatus'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        let itemStatus = status === 'enable' ? 'Enable' : 'Disable';
        await updateStatus({ id: id, status: itemStatus });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        search();
    });
};

const onBatchChangeStatus = async (status: string) => {
    ElMessageBox.confirm(i18n.global.t('cronjob.' + status + 'Msg'), i18n.global.t('cronjob.changeStatus'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        let itemStatus = status === 'enable' ? 'Enable' : 'Disable';
        for (const item of selects.value) {
            await updateStatus({ id: item.id, status: itemStatus });
        }
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        search();
    });
};

const loadBackups = async (row: any) => {
    dialogBackupRef.value!.acceptParams({ cronjobID: row.id, cronjob: row.name });
};

const onHandle = async (row: Cronjob.CronjobInfo) => {
    loading.value = true;
    await handleOnce(row.id)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};

const hasBackup = (type: string) => {
    return (
        type === 'app' ||
        type === 'website' ||
        type === 'database' ||
        type === 'directory' ||
        type === 'snapshot' ||
        type === 'log'
    );
};

const loadDetail = (row: any) => {
    isRecordShow.value = true;
    let params = {
        rowData: { ...row },
    };
    dialogRecordRef.value!.acceptParams(params);
};

const buttons = [
    {
        label: i18n.global.t('commons.button.handle'),
        click: (row: Cronjob.CronjobInfo) => {
            onHandle(row);
        },
    },
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: Cronjob.CronjobInfo) => {
            onOpenDialog('edit', row);
        },
    },
    {
        label: i18n.global.t('cronjob.viewRecords'),
        click: (row: Cronjob.CronjobInfo) => {
            loadDetail(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Cronjob.CronjobInfo) => {
            onDelete(row);
        },
    },
];

onMounted(() => {
    search();
});
</script>
