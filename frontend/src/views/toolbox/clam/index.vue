<template>
    <div>
        <LayoutContent v-loading="loading" v-if="!isRecordShow && !isSettingShow" :title="$t('toolbox.clam.clam')">
            <template #prompt>
                <el-alert type="info" :closable="false">
                    <template #title>
                        {{ $t('toolbox.clam.clamHelper') }}
                        <el-link class="ml-1 text-xs" @click="toDoc()" type="primary">
                            {{ $t('commons.button.helpDoc') }}
                        </el-link>
                    </template>
                </el-alert>
            </template>
            <template #app>
                <ClamStatus
                    @setting="setting"
                    v-model:loading="loading"
                    @get-status="getStatus"
                    v-model:mask-show="maskShow"
                />
            </template>
            <template #leftToolBar v-if="clamStatus.isExist">
                <el-button type="primary" :disabled="!clamStatus.isRunning" @click="onOpenDialog('add')">
                    {{ $t('toolbox.clam.clamCreate') }}
                </el-button>
                <el-button plain :disabled="selects.length === 0 || !clamStatus.isRunning" @click="onDelete(null)">
                    {{ $t('commons.button.delete') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <TableSearch @search="search()" v-model:searchName="searchName" />
                <TableRefresh @search="search()" />
                <TableSetting title="clam-refresh" @search="search()" />
            </template>
            <el-card v-if="clamStatus.isExist && !clamStatus.isRunning && maskShow" class="mask-prompt">
                <span>{{ $t('toolbox.clam.notStart') }}</span>
            </el-card>
            <template #main v-if="clamStatus.isExist">
                <ComplexTable
                    :class="{ mask: !clamStatus.isRunning }"
                    v-if="!isSettingShow"
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    @sort-change="search"
                    @search="search"
                    :data="data"
                >
                    <el-table-column type="selection" fix />
                    <el-table-column
                        :label="$t('commons.table.name')"
                        :min-width="60"
                        prop="name"
                        sortable
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <el-text type="primary" class="cursor-pointer" @click="onOpenRecord(row)">
                                {{ row.name }}
                            </el-text>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('toolbox.clam.scanDir')"
                        :min-width="120"
                        prop="path"
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <el-button link type="primary" @click="routerToFileWithPath(row.path)">
                                {{ row.path }}
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column
                        v-if="isProductPro"
                        :label="$t('commons.table.status')"
                        :min-width="70"
                        prop="status"
                        sortable
                    >
                        <template #default="{ row }">
                            <Status
                                v-if="row.status === 'Enable'"
                                :status="row.status"
                                @click="onChangeStatus(row.id, 'disable')"
                            />
                            <Status
                                v-if="row.status === 'Disable'"
                                :status="row.status"
                                @click="onChangeStatus(row.id, 'enable')"
                            />
                            <span v-if="row.status === ''">-</span>
                        </template>
                    </el-table-column>
                    <el-table-column
                        v-if="isProductPro"
                        :label="$t('cronjob.cronSpec')"
                        show-overflow-tooltip
                        :min-width="120"
                    >
                        <template #default="{ row }">
                            <span>
                                {{ row.spec !== '' ? transSpecToStr(row.spec) : '-' }}
                            </span>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('toolbox.clam.infectedDir')"
                        :min-width="120"
                        prop="path"
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <el-button
                                v-if="row.infectedStrategy === 'copy' || row.infectedStrategy === 'move'"
                                link
                                type="primary"
                                @click="routerToFileWithPath(row.infectedDir + '/1panel-infected/' + row.name)"
                            >
                                {{ row.infectedDir + '/1panel-infected/' + row.name }}
                            </el-button>
                            <span v-else>-</span>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('cronjob.lastRecordTime')"
                        :min-width="120"
                        show-overflow-tooltip
                        prop="lastRecordTime"
                    >
                        <template #default="{ row }">
                            <el-button v-if="row.lastRecordStatus === 'Done'" icon="Select" link type="success" />
                            <el-button v-if="row.lastRecordStatus === 'Failed'" icon="CloseBold" link type="danger" />
                            <el-button v-if="row.lastRecordStatus === 'Waiting'" :loading="true" link type="info" />
                            {{ row.lastRecordTime }}
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.description')" prop="description" show-overflow-tooltip>
                        <template #default="{ row }">
                            <fu-input-rw-switch v-model="row.description" @blur="onChange(row)" />
                        </template>
                    </el-table-column>
                    <fu-table-operations
                        width="300px"
                        :buttons="buttons"
                        :ellipsis="10"
                        :label="$t('commons.table.operate')"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <OpDialog ref="opRef" @search="search" @submit="onSubmitDelete()">
            <template #content>
                <el-form class="mt-4 mb-1" ref="deleteForm" label-position="left">
                    <el-form-item>
                        <el-checkbox v-model="removeInfected" :label="$t('toolbox.clam.removeInfected')" />
                        <span class="input-help">{{ $t('toolbox.clam.removeInfectedHelper') }}</span>
                    </el-form-item>
                </el-form>
            </template>
        </OpDialog>
        <OperateDialog @search="search" ref="dialogRef" />
        <LogDialog ref="dialogLogRef" />
        <SettingDialog v-if="isSettingShow" />
    </div>
</template>

<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { deleteClam, handleClamScan, searchClam, updateClam, updateClamStatus } from '@/api/modules/toolbox';
import OperateDialog from '@/views/toolbox/clam/operate/index.vue';
import LogDialog from '@/views/toolbox/clam/record/index.vue';
import ClamStatus from '@/views/toolbox/clam/status/index.vue';
import SettingDialog from '@/views/toolbox/clam/setting/index.vue';
import { Toolbox } from '@/api/interface/toolbox';
import { transSpecToStr } from '@/views/cronjob/cronjob/helper';
import { GlobalStore } from '@/store';
import { storeToRefs } from 'pinia';
import { routerToFileWithPath, routerToName } from '@/utils/router';

const loading = ref();
const selects = ref<any>([]);

const globalStore = GlobalStore();
const { isProductPro } = storeToRefs(globalStore);
const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'clam-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('clam-page-size')) || 20,
    total: 0,
    orderBy: 'createdAt',
    order: 'null',
});
const searchName = ref();

const opRef = ref();
const dialogRef = ref();
const operateIDs = ref();
const dialogLogRef = ref();
const isRecordShow = ref();

const removeInfected = ref();

const isSettingShow = ref();
const maskShow = ref(true);
const clamStatus = ref({
    isExist: false,
    isRunning: true,
});

const search = async (column?: any) => {
    paginationConfig.orderBy = column?.order ? column.prop : paginationConfig.orderBy;
    paginationConfig.order = column?.order ? column.order : paginationConfig.order;
    loading.value = true;
    let params = {
        info: searchName.value,
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        orderBy: paginationConfig.orderBy,
        order: paginationConfig.order,
    };
    await searchClam(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const setting = () => {
    routerToName('Clam-Setting');
};
const getStatus = (status: any) => {
    clamStatus.value = status;
    search();
};

const toDoc = () => {
    window.open(globalStore.docsUrl + '/user_manual/toolbox/clam/', '_blank', 'noopener,noreferrer');
};

const onChange = async (row: any) => {
    await updateClam(row);
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
};

const onOpenDialog = async (
    title: string,
    rowData: Partial<Toolbox.ClamInfo> = {
        infectedStrategy: 'none',
        specObj: {
            specType: 'perDay',
            week: 1,
            day: 3,
            hour: 1,
            minute: 30,
            second: 30,
        },
    },
) => {
    let params = {
        title,
        rowData: { ...rowData },
    };
    dialogRef.value!.acceptParams(params);
};
const onOpenRecord = (row: Toolbox.ClamInfo) => {
    isRecordShow.value = true;
    let params = {
        rowData: { ...row },
    };
    dialogLogRef.value!.acceptParams(params);
};

const onDelete = async (row: Toolbox.ClamInfo | null) => {
    let names = [];
    let ids = [];
    if (row) {
        ids = [row.id];
        names = [row.name];
    } else {
        for (const item of selects.value) {
            names.push(item.name);
            ids.push(item.id);
        }
    }
    operateIDs.value = ids;
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('toolbox.clam.clam'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: null,
        params: null,
    });
};

const onSubmitDelete = async () => {
    loading.value = true;
    await deleteClam({ ids: operateIDs.value, removeInfected: removeInfected.value })
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
    ElMessageBox.confirm(i18n.global.t('toolbox.clam.' + status + 'Msg'), i18n.global.t('cronjob.changeStatus'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        let itemStatus = status === 'enable' ? 'Enable' : 'Disable';
        await updateClamStatus(id, itemStatus);
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        search();
    });
};

const buttons = [
    {
        label: i18n.global.t('commons.button.handle'),
        click: async (row: Toolbox.ClamInfo) => {
            loading.value = true;
            await handleClamScan(row.id)
                .then(() => {
                    loading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    search();
                })
                .catch(() => {
                    loading.value = false;
                });
        },
    },
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: Toolbox.ClamInfo) => {
            onOpenDialog('edit', row);
        },
    },
    {
        label: i18n.global.t('cronjob.record'),
        click: (row: Toolbox.ClamInfo) => {
            onOpenRecord(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Toolbox.ClamInfo) => {
            onDelete(row);
        },
    },
];

onMounted(() => {
    search();
});
</script>
