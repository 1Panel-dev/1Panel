<template>
    <div>
        <RouterButton
            :buttons="[
                {
                    label: i18n.global.t('cronjob.cronTask'),
                    path: '/cronjobs',
                },
            ]"
        />
        <LayoutContent v-loading="loading" v-if="!isRecordShow" :title="$t('cronjob.cronTask')">
            <template #toolbar>
                <el-row>
                    <el-col :xs="24" :sm="16" :md="16" :lg="16" :xl="16">
                        <el-button type="primary" @click="onOpenDialog('create')">
                            {{ $t('commons.button.create') }}{{ $t('cronjob.cronTask') }}
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
                    </el-col>
                    <el-col :xs="24" :sm="8" :md="8" :lg="8" :xl="8">
                        <TableSetting @search="search()" />
                        <div class="search-button">
                            <el-input
                                v-model="searchName"
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
                    @sort-change="search"
                    @search="search"
                    :data="data"
                >
                    <el-table-column type="selection" fix />
                    <el-table-column :label="$t('cronjob.taskName')" :min-width="120" prop="name" sortable>
                        <template #default="{ row }">
                            <Tooltip @click="loadDetail(row)" :text="row.name" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.status')" :min-width="80" prop="status" sortable>
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
                            <span v-if="row.specType.indexOf('N') === -1 || row.specType === 'perWeek'">
                                {{ $t('cronjob.' + row.specType) }}&nbsp;
                            </span>
                            <span v-else>{{ $t('cronjob.per') }}</span>
                            <span v-if="row.specType === 'perMonth'">
                                {{ row.day }}{{ $t('cronjob.day') }} {{ loadZero(row.hour) }} :
                                {{ loadZero(row.minute) }}
                            </span>
                            <span v-if="row.specType === 'perWeek'">
                                {{ loadWeek(row.week) }} {{ loadZero(row.hour) }} : {{ loadZero(row.minute) }}
                            </span>
                            <span v-if="row.specType === 'perDay'">
                                &#32;{{ loadZero(row.hour) }} : {{ loadZero(row.minute) }}
                            </span>
                            <span v-if="row.specType === 'perNDay'">
                                {{ row.day }} {{ $t('commons.units.day') }}, {{ loadZero(row.hour) }} :
                                {{ loadZero(row.minute) }}
                            </span>
                            <span v-if="row.specType === 'perNHour'">
                                {{ row.hour }}{{ $t('commons.units.hour') }}, {{ loadZero(row.minute) }}
                            </span>
                            <span v-if="row.specType === 'perHour'">{{ loadZero(row.minute) }}</span>
                            <span v-if="row.specType === 'perNMinute'">
                                {{ row.minute }}{{ $t('commons.units.minute') }}
                            </span>
                            <span v-if="row.specType === 'perNSecond'">
                                {{ row.second }}{{ $t('commons.units.second') }}
                            </span>
                            {{ $t('cronjob.handle') }}
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('cronjob.retainCopies')" :min-width="90" prop="retainCopies" />

                    <el-table-column :label="$t('cronjob.lastRecordTime')" :min-width="120" prop="lastRecordTime">
                        <template #default="{ row }">
                            {{ row.lastRecordTime }}
                        </template>
                    </el-table-column>
                    <el-table-column :min-width="80" :label="$t('cronjob.target')" prop="targetDir">
                        <template #default="{ row }">
                            {{ row.targetDir }}
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
    </div>
</template>

<script lang="ts" setup>
import OpDialog from '@/components/del-dialog/index.vue';
import TableSetting from '@/components/table-setting/index.vue';
import Tooltip from '@/components/tooltip/index.vue';
import OperateDialog from '@/views/cronjob/operate/index.vue';
import Records from '@/views/cronjob/record/index.vue';
import { loadZero } from '@/utils/util';
import { onMounted, reactive, ref } from 'vue';
import { deleteCronjob, getCronjobPage, handleOnce, updateStatus } from '@/api/modules/cronjob';
import i18n from '@/lang';
import { Cronjob } from '@/api/interface/cronjob';
import { ElMessageBox } from 'element-plus';
import { MsgSuccess } from '@/utils/message';

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

const weekOptions = [
    { label: i18n.global.t('cronjob.monday'), value: 1 },
    { label: i18n.global.t('cronjob.tuesday'), value: 2 },
    { label: i18n.global.t('cronjob.wednesday'), value: 3 },
    { label: i18n.global.t('cronjob.thursday'), value: 4 },
    { label: i18n.global.t('cronjob.friday'), value: 5 },
    { label: i18n.global.t('cronjob.saturday'), value: 6 },
    { label: i18n.global.t('cronjob.sunday'), value: 0 },
];

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
                if (item.targetDir !== '-' && item.targetDir !== '') {
                    item.targetDir = i18n.global.t('setting.' + item.targetDir);
                }
            }
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const dialogRecordRef = ref();

const dialogRef = ref();
const onOpenDialog = async (
    title: string,
    rowData: Partial<Cronjob.CronjobInfo> = {
        specType: 'perMonth',
        type: 'shell',
        week: 1,
        day: 3,
        hour: 1,
        minute: 30,
        second: 30,
        keepLocal: true,
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
    await deleteCronjob({ ids: operateIDs.value, cleanData: cleanData.value });
    MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
    search();
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
        label: i18n.global.t('cronjob.record'),
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
function loadWeek(i: number) {
    for (const week of weekOptions) {
        if (week.value === i) {
            return week.label;
        }
    }
    return '';
}
onMounted(() => {
    search();
});
</script>
