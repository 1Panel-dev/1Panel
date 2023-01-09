<template>
    <div v-loading="loading">
        <Submenu activeName="snapshot" />
        <el-card v-if="!isRecordShow" style="margin-top: 20px">
            <ComplexTable :pagination-config="paginationConfig" v-model:selects="selects" :data="data" @search="search">
                <template #toolbar>
                    <el-button type="primary" icon="Plus" @click="onCreate()">
                        {{ $t('setting.createSnapshot') }}
                    </el-button>
                    <el-button type="danger" plain :disabled="selects.length === 0" @click="batchDelete(null)">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </template>
                <el-table-column type="selection" fix />
                <el-table-column
                    show-overflow-tooltip
                    :label="$t('commons.table.name')"
                    min-width="100"
                    prop="name"
                    fix
                />
                <el-table-column
                    :label="$t('commons.table.description')"
                    min-width="150"
                    show-overflow-tooltip
                    prop="description"
                />
                <el-table-column :label="$t('setting.backupAccount')" min-width="150" prop="from" />
                <el-table-column :label="$t('setting.backup')" min-width="80" prop="status">
                    <template #default="{ row }">
                        <el-tag v-if="row.status === 'Success'" type="success">
                            {{ $t('commons.table.statusSuccess') }}
                        </el-tag>
                        <el-tag v-if="row.status === 'Waiting'" type="info">
                            {{ $t('commons.table.statusWaiting') }}
                        </el-tag>
                        <el-tooltip
                            v-if="row.status === 'Failed'"
                            class="box-item"
                            effect="dark"
                            :content="row.message"
                            placement="top-start"
                        >
                            <el-tag type="danger">{{ $t('commons.table.statusFailed') }}</el-tag>
                        </el-tooltip>
                    </template>
                </el-table-column>
                <el-table-column
                    prop="createdAt"
                    :label="$t('commons.table.date')"
                    :formatter="dateFromat"
                    show-overflow-tooltip
                />
                <fu-table-operations
                    width="200px"
                    :ellipsis="10"
                    :buttons="buttons"
                    :label="$t('commons.table.operate')"
                    fix
                />
            </ComplexTable>
        </el-card>
        <RecoverStatus ref="recoverStatusRef"></RecoverStatus>
        <el-dialog v-model="dialogVisiable" :title="$t('setting.createSnapshot')" width="30%">
            <el-form v-loading="loading" ref="snapRef" label-width="100px" :model="snapInfo" :rules="rules">
                <el-form-item :label="$t('cronjob.target')" prop="from">
                    <el-select v-model="snapInfo.from" clearable>
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
                    <el-button :disabled="loading" @click="dialogVisiable = false">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button :disabled="loading" type="primary" @click="submitAddSnapshot(snapRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
    </div>
</template>

<script setup lang="ts">
import ComplexTable from '@/components/complex-table/index.vue';
import { snapshotCreate, searchSnapshotPage, snapshotDelete } from '@/api/modules/setting';
import { onMounted, reactive, ref } from 'vue';
import { dateFromat } from '@/utils/util';
import { useDeleteData } from '@/hooks/use-delete-data';
import { ElForm } from 'element-plus';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElMessage } from 'element-plus';
import { Setting } from '@/api/interface/setting';
import Submenu from '@/views/setting/index.vue';
import RecoverStatus from '@/views/setting/snapshot/status/index.vue';
import { getBackupList } from '@/api/modules/backup';
import { loadBackupName } from '../helper';

const loading = ref(false);
const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const recoverStatusRef = ref();
const isRecordShow = ref();
const backupOptions = ref();
type FormInstance = InstanceType<typeof ElForm>;
const snapRef = ref<FormInstance>();
const rules = reactive({
    from: [Rules.requiredSelect],
});

let snapInfo = reactive<Setting.SnapshotCreate>({
    from: '',
    description: '',
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
        loading.value = true;
        await snapshotCreate(snapInfo)
            .then(() => {
                loading.value = false;
                dialogVisiable.value = false;
                search();
                ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
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

const batchDelete = async (row: Setting.SnapshotInfo | null) => {
    let ids: Array<number> = [];
    if (row === null) {
        selects.value.forEach((item: Setting.SnapshotInfo) => {
            ids.push(item.id);
        });
    } else {
        ids.push(row.id);
    }
    await useDeleteData(snapshotDelete, { ids: ids }, 'commons.msg.delete');
    search();
};

function restForm() {
    if (snapRef.value) {
        snapRef.value.resetFields();
    }
}
const buttons = [
    {
        label: i18n.global.t('commons.button.recover'),
        icon: 'RefreshLeft',
        click: (row: any) => {
            recoverStatusRef.value.acceptParams({ snapInfo: row });
        },
        disabled: (row: any) => {
            return !(row.status === 'Success');
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        icon: 'Delete',
        click: batchDelete,
    },
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
