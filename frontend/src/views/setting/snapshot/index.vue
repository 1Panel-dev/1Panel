<template>
    <div>
        <Submenu activeName="snapshot" />
        <el-card style="margin-top: 20px">
            <ComplexTable :pagination-config="paginationConfig" v-model:selects="selects" :data="data" @search="search">
                <template #toolbar>
                    <el-button type="primary" icon="Plus" @click="onCreate()">
                        {{ $t('commons.button.create') }}
                    </el-button>
                    <!-- <el-button type="danger" plain :disabled="selects.length === 0" @click="batchDelete(null)">
                        {{ $t('commons.button.delete') }}
                    </el-button> -->
                </template>
                <el-table-column type="selection" fix />
                <el-table-column :label="$t('commons.table.name')" min-width="100" prop="name" fix />
                <el-table-column
                    :label="$t('commons.table.description')"
                    min-width="150"
                    show-overflow-tooltip
                    prop="description"
                />
                <el-table-column :label="$t('setting.backupAccount')" min-width="150" prop="backupAccount" />
                <el-table-column :label="$t('commons.table.status')" min-width="80" prop="status">
                    <template #default="{ row }">
                        <el-tag v-if="row.status === 'Success'" type="success">
                            {{ $t('commons.table.statusSuccess') }}
                        </el-tag>
                        <el-tooltip v-else class="box-item" effect="dark" :content="row.message" placement="top-start">
                            <el-tag type="danger">{{ $t('commons.table.statusFailed') }}</el-tag>
                        </el-tooltip>
                    </template>
                </el-table-column>
                <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fix />
            </ComplexTable>
        </el-card>
        <el-dialog v-model="dialogVisiable" :title="$t('commons.button.create')" width="30%">
            <el-form ref="snapRef" label-width="100px" :model="snapInfo" :rules="rules">
                <el-form-item :label="$t('cronjob.target')" prop="backupType">
                    <el-select v-model="snapInfo.backupType" clearable>
                        <el-option
                            v-for="item in backupOptions"
                            :key="item.label"
                            :value="item.value"
                            :label="item.label"
                        />
                    </el-select>
                </el-form-item>
                <el-form-item :label="$t('commons.table.description')" prop="description">
                    <el-input type="textarea" clearable v-model="snapInfo.description" />
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="dialogVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button type="primary" @click="submitAddSnapshot(snapRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
    </div>
</template>

<script setup lang="ts">
import ComplexTable from '@/components/complex-table/index.vue';
import { snapshotCreate, searchSnapshotPage } from '@/api/modules/setting';
import { onMounted, reactive, ref } from 'vue';
// import { useDeleteData } from '@/hooks/use-delete-data';
import type { ElForm } from 'element-plus';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElMessage } from 'element-plus';
import { Setting } from '@/api/interface/setting';
import Submenu from '@/views/setting/index.vue';
import { getBackupList } from '@/api/modules/backup';
import { loadBackupName } from '../helper';

const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const backupOptions = ref();

type FormInstance = InstanceType<typeof ElForm>;
const snapRef = ref<FormInstance>();
const rules = reactive({
    backupType: [Rules.requiredSelect],
});

let snapInfo = reactive<Setting.SnapshotCreate>({
    description: '',
    backupType: '',
});

const dialogVisiable = ref<boolean>(false);

const onCreate = async () => {
    restForm();
    dialogVisiable.value = true;
};

const submitAddSnapshot = (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        await snapshotCreate(snapInfo);
        dialogVisiable.value = false;
        search();
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
    });
};

const loadBackups = async () => {
    const res = await getBackupList();
    backupOptions.value = [];
    for (const item of res.data) {
        if (item.type !== 'LOCAL') {
            backupOptions.value.push({ label: loadBackupName(item.type), value: item.type });
        }
    }
};

// const batchDelete = async (row: Snapshot.SnapshotInfo | null) => {
//     let ids: Array<number> = [];
//     if (row === null) {
//         selects.value.forEach((item: Snapshot.SnapshotInfo) => {
//             ids.push(item.id);
//         });
//     } else {
//         ids.push(row.id);
//     }
//     await useDeleteData(deleteSnapshot, { ids: ids }, 'commons.msg.delete');
//     search();
// };

function restForm() {
    if (snapRef.value) {
        snapRef.value.resetFields();
    }
}
const buttons = [
    // {
    //     label: i18n.global.t('commons.button.delete'),
    //     icon: 'Delete',
    //     click: batchDelete,
    // },
];

const search = async () => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    const res = await searchSnapshotPage(params);
    data.value = res.data.items || [];
    paginationConfig.total = res.data.total;
};

onMounted(() => {
    search();
    loadBackups();
});
</script>
