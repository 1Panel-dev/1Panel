<template>
    <div>
        <ComplexTable
            :pagination-config="paginationConfig"
            v-model:selects="selects"
            @search="search"
            style="margin-top: 20px"
            :data="data"
        >
            <template #toolbar>
                <el-button type="primary" icon="Plus" @click="onOpenDialog('create')">
                    {{ $t('commons.button.create') }}
                </el-button>
                <el-button type="danger" plain :disabled="selects.length === 0" @click="onBatchDelete(null)">
                    {{ $t('commons.button.delete') }}
                </el-button>
            </template>
            <el-table-column type="selection" fix />
            <el-table-column :label="$t('cronjob.taskName')" prop="name" />
            <el-table-column :label="$t('commons.table.status')" prop="status">
                <template #default="{ row }">
                    <el-button
                        v-if="row.status === 'Enable'"
                        @click="onChangeStatus(row.id, 'disable')"
                        link
                        type="success"
                        icon="VideoPlay"
                    >
                        {{ $t('commons.status.enabled') }}
                    </el-button>
                    <el-button v-else link type="danger" @click="onChangeStatus(row.id, 'enable')" icon="VideoPause">
                        {{ $t('commons.status.disabled') }}
                    </el-button>
                </template>
            </el-table-column>
            <el-table-column :label="$t('cronjob.cronSpec')">
                <template #default="{ row }">
                    <span v-if="row.specType.indexOf('N') === -1 || row.specType === 'perWeek'">
                        {{ $t('cronjob.' + row.specType) }}
                    </span>
                    <span v-else>{{ $t('cronjob.per') }}</span>
                    <span v-if="row.specType === 'perMonth'">
                        {{ row.day }}{{ $t('cronjob.day') }} {{ loadZero(row.hour) }} :
                        {{ loadZero(row.minute) }}
                    </span>
                    <span v-if="row.specType === 'perWeek'">
                        {{ loadWeek(row.week) }} {{ loadZero(row.hour) }} : {{ loadZero(row.minute) }}
                    </span>
                    <span v-if="row.specType === 'perNDay'">
                        {{ row.day }}{{ $t('cronjob.day1') }}, {{ loadZero(row.hour) }} : {{ loadZero(row.minute) }}
                    </span>
                    <span v-if="row.specType === 'perNHour'">
                        {{ row.hour }}{{ $t('cronjob.hour') }}, {{ loadZero(row.minute) }}
                    </span>
                    <span v-if="row.specType === 'perHour'">{{ loadZero(row.minute) }}</span>
                    <span v-if="row.specType === 'perNMinute'">{{ row.minute }}{{ $t('cronjob.minute') }}</span>
                    {{ $t('cronjob.handle') }}
                </template>
            </el-table-column>
            <el-table-column :label="$t('cronjob.retainCopies')" prop="retainCopies" />
            <el-table-column :label="$t('cronjob.lastRecrodTime')" prop="lastRecrodTime">
                <template #default="{ row }">
                    {{ row.lastRecrodTime }}
                </template>
            </el-table-column>
            <el-table-column :label="$t('cronjob.target')" prop="targetDir">
                <template #default="{ row }">
                    {{ loadBackupName(row.targetDir) }}
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

        <OperatrDialog @search="search" ref="dialogRef" />
        <RecordDialog ref="dialogRecordRef" />
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import OperatrDialog from '@/views/cronjob/operate/index.vue';
import RecordDialog from '@/views/cronjob/record/index.vue';
import { loadZero } from '@/utils/util';
import { onMounted, reactive, ref } from 'vue';
import { deleteCronjob, getCronjobPage, handleOnce, updateStatus } from '@/api/modules/cronjob';
import { loadBackupName } from '@/views/setting/helper';
import i18n from '@/lang';
import { Cronjob } from '@/api/interface/cronjob';
import { useDeleteData } from '@/hooks/use-delete-data';
import { ElMessage, ElMessageBox } from 'element-plus';

const selects = ref<any>([]);

const data = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const weekOptions = [
    { label: i18n.global.t('cronjob.monday'), value: 1 },
    { label: i18n.global.t('cronjob.tuesday'), value: 2 },
    { label: i18n.global.t('cronjob.wednesday'), value: 3 },
    { label: i18n.global.t('cronjob.thursday'), value: 4 },
    { label: i18n.global.t('cronjob.friday'), value: 5 },
    { label: i18n.global.t('cronjob.saturday'), value: 6 },
    { label: i18n.global.t('cronjob.sunday'), value: 7 },
];

const search = async () => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    const res = await getCronjobPage(params);
    data.value = res.data.items || [];
    for (const item of data.value) {
        if (item.targetDir !== '-') {
            item.targetDir = loadBackupName(item.targetDir);
        }
    }
    paginationConfig.total = res.data.total;
};

const dialogRecordRef = ref<DialogExpose>();

interface DialogExpose {
    acceptParams: (params: any) => void;
}
const dialogRef = ref<DialogExpose>();
const onOpenDialog = async (
    title: string,
    rowData: Partial<Cronjob.CronjobInfo> = {
        specType: 'perMonth',
        week: 1,
        day: 1,
        hour: 2,
        minute: 3,
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

const onBatchDelete = async (row: Cronjob.CronjobInfo | null) => {
    let ids: Array<number> = [];
    if (row) {
        ids.push(row.id);
    } else {
        selects.value.forEach((item: Cronjob.CronjobInfo) => {
            ids.push(item.id);
        });
    }
    await useDeleteData(deleteCronjob, { ids: ids }, 'commons.msg.delete');
    search();
};

const onChangeStatus = async (id: number, status: string) => {
    ElMessageBox.confirm(i18n.global.t('cronjob.' + status + 'Msg'), i18n.global.t('cronjob.changeStatus'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        let itemStatus = status === 'enable' ? 'Enable' : 'Disable';
        await updateStatus({ id: id, status: itemStatus });
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        search();
    });
};

const onHandle = async (row: Cronjob.CronjobInfo) => {
    await handleOnce(row.id);
    ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
    search();
};

const buttons = [
    {
        label: i18n.global.t('commons.button.handle'),
        icon: 'Pointer',
        click: (row: Cronjob.CronjobInfo) => {
            onHandle(row);
        },
    },
    {
        label: i18n.global.t('commons.button.edit'),
        icon: 'Edit',
        click: (row: Cronjob.CronjobInfo) => {
            onOpenDialog('edit', row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        icon: 'Delete',
        click: (row: Cronjob.CronjobInfo) => {
            onBatchDelete(row);
        },
    },
    {
        label: i18n.global.t('commons.button.view'),
        icon: 'Clock',
        click: (row: Cronjob.CronjobInfo) => {
            onOpenRecordDialog(row);
        },
    },
];
const onOpenRecordDialog = async (rowData: Partial<Cronjob.CronjobInfo> = {}) => {
    let params = {
        rowData: { ...rowData },
    };
    dialogRecordRef.value!.acceptParams(params);
};

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
