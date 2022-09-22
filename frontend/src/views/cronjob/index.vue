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
                <el-button type="primary" @click="onOpenDialog('create')">{{ $t('commons.button.create') }}</el-button>
                <el-button type="danger" plain :disabled="selects.length === 0" @click="onBatchDelete(null)">
                    {{ $t('commons.button.delete') }}
                </el-button>
            </template>
            <el-table-column type="expand">
                <template #default="{ row }">
                    <ul>
                        <li>{{ row.name }} {{ $t('cronjob.handle') }}记录 1</li>
                        <li>{{ row.name }} {{ $t('cronjob.handle') }}记录 2</li>
                        <li>{{ row.name }} {{ $t('cronjob.handle') }}记录 3</li>
                        <li>{{ row.name }} {{ $t('cronjob.handle') }}记录 4</li>
                    </ul>
                </template>
            </el-table-column>

            <el-table-column type="selection" fix />
            <el-table-column :label="$t('cronjob.taskName')" prop="name" />
            <el-table-column :label="$t('commons.table.status')" prop="status">
                <template #default="{ row }">
                    <el-switch
                        @change="onChangeStatus(row)"
                        :before-change="beforeChangeStatus"
                        v-model="row.status"
                        inline-prompt
                        style="--el-switch-on-color: #13ce66; --el-switch-off-color: #ff4949"
                        active-text="Y"
                        inactive-text="N"
                        active-value="running"
                        inactive-value="stoped"
                    />
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
            <el-table-column :label="$t('cronjob.target')" prop="targetDir" />
            <fu-table-operations type="icon" :buttons="buttons" :label="$t('commons.table.operate')" fix />
        </ComplexTable>

        <OperatrDialog @search="search" ref="dialogRef" />
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import OperatrDialog from '@/views/cronjob/operate/index.vue';
import { onMounted, reactive, ref } from 'vue';
import { loadBackupName } from '@/views/setting/helper';
import { deleteCronjob, editCronjob, getCronjobPage } from '@/api/modules/cronjob';
import { loadWeek } from './options';
import i18n from '@/lang';
import { Cronjob } from '@/api/interface/cronjob';
import { useDeleteData } from '@/hooks/use-delete-data';
import { ElMessage } from 'element-plus';
const selects = ref<any>([]);
const switchState = ref<boolean>(false);

const data = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 5,
    total: 0,
});

const logSearch = reactive({
    page: 1,
    pageSize: 5,
});

const search = async () => {
    logSearch.page = paginationConfig.currentPage;
    logSearch.pageSize = paginationConfig.pageSize;
    const res = await getCronjobPage(logSearch);
    data.value = res.data.items;
    for (const item of data.value) {
        if (item.targetDir !== '-') {
            item.targetDir = loadBackupName(item.targetDir);
        }
    }
    paginationConfig.total = res.data.total;
};

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
        retainCopies: 3,
    },
) => {
    let params = {
        title,
        rowData: { ...rowData },
        isView: title === 'view',
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
    await useDeleteData(deleteCronjob, { ids: ids }, 'commons.msg.delete', true);
    search();
};
const beforeChangeStatus = () => {
    switchState.value = true;
    return switchState.value;
};
const onChangeStatus = async (row: Cronjob.CronjobInfo) => {
    if (switchState.value) {
        console.log(row.status);
        await editCronjob(row);
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
        search();
    }
};

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        icon: 'Edit',
        click: (row: Cronjob.CronjobInfo) => {
            onOpenDialog('edit', row);
        },
    },
    {
        label: i18n.global.t('commons.button.view'),
        icon: 'View',
        click: (row: Cronjob.CronjobInfo) => {
            onOpenDialog('view', row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        icon: 'Delete',
        click: (row: Cronjob.CronjobInfo) => {
            onBatchDelete(row);
        },
    },
];

function loadZero(i: number) {
    return i < 10 ? '0' + i : '' + i;
}

onMounted(() => {
    search();
});
</script>
