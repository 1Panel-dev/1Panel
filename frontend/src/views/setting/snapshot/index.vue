<template>
    <div>
        <LayoutContent v-loading="loading" v-if="!isRecordShow" :title="$t('setting.snapshot')">
            <template #toolbar>
                <el-row>
                    <el-col :xs="24" :sm="16" :md="16" :lg="16" :xl="16">
                        <el-button type="primary" @click="onCreate()">
                            {{ $t('setting.createSnapshot') }}
                        </el-button>
                        <el-button type="primary" plain @click="onImport()">
                            {{ $t('setting.importSnapshot') }}
                        </el-button>
                        <el-button type="primary" plain :disabled="selects.length === 0" @click="batchDelete(null)">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </el-col>
                    <el-col :xs="24" :sm="8" :md="8" :lg="8" :xl="8">
                        <TableSetting ref="timerRef" @search="search()" />
                        <div class="search-button">
                            <el-input
                                v-model="searchName"
                                clearable
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
                    :data="data"
                    style="margin-top: 20px"
                    @search="search"
                >
                    <el-table-column type="selection" fix />
                    <el-table-column
                        show-overflow-tooltip
                        :label="$t('commons.table.name')"
                        min-width="100"
                        prop="name"
                        fix
                    />
                    <el-table-column prop="version" :label="$t('app.version')" />
                    <el-table-column :label="$t('setting.backupAccount')" min-width="80" prop="from">
                        <template #default="{ row }">
                            <span v-if="row.from">
                                {{ $t('setting.' + row.from) }}
                            </span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.status')" min-width="80" prop="status">
                        <template #default="{ row }">
                            <el-tag v-if="row.status === 'Success'" type="success">
                                {{ $t('commons.table.statusSuccess') }}
                            </el-tag>
                            <el-tag v-if="row.status === 'Waiting'" type="info">
                                {{ $t('commons.table.statusWaiting') }}
                            </el-tag>
                            <el-tag v-if="row.status === 'Uploading'" type="info">
                                {{ $t('commons.status.uploading') }}...
                            </el-tag>
                            <el-tooltip v-if="row.status === 'Failed'" effect="dark" placement="top">
                                <template #content>
                                    <div style="width: 300px; word-break: break-all">{{ row.message }}</div>
                                </template>
                                <el-tag type="danger">{{ $t('commons.table.statusFailed') }}</el-tag>
                            </el-tooltip>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.description')" prop="description">
                        <template #default="{ row }">
                            <fu-read-write-switch :data="row.description" v-model="row.edit" @change="onChange(row)">
                                <el-input v-model="row.description" @blur="row.edit = false" />
                            </fu-read-write-switch>
                        </template>
                    </el-table-column>
                    <el-table-column
                        prop="createdAt"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
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
            </template>
        </LayoutContent>
        <RecoverStatus ref="recoverStatusRef" @search="search()"></RecoverStatus>
        <SnapshotImport ref="importRef" @search="search()" />
        <el-drawer v-model="drawerVisiable" size="50%">
            <template #header>
                <DrawerHeader :header="$t('setting.createSnapshot')" :back="handleClose" />
            </template>
            <el-form
                v-loading="loading"
                label-position="top"
                ref="snapRef"
                label-width="100px"
                :model="snapInfo"
                :rules="rules"
            >
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item
                            :label="$t('cronjob.target') + ' ( ' + $t('setting.thirdPartySupport') + ' )'"
                            prop="from"
                        >
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
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button :disabled="loading" @click="drawerVisiable = false">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button :disabled="loading" type="primary" @click="submitAddSnapshot(snapRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>

<script setup lang="ts">
import TableSetting from '@/components/table-setting/index.vue';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { snapshotCreate, searchSnapshotPage, snapshotDelete, updateSnapshotDescription } from '@/api/modules/setting';
import { onMounted, reactive, ref } from 'vue';
import { dateFormat } from '@/utils/util';
import { useDeleteData } from '@/hooks/use-delete-data';
import { ElForm } from 'element-plus';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { Setting } from '@/api/interface/setting';
import RecoverStatus from '@/views/setting/snapshot/status/index.vue';
import SnapshotImport from '@/views/setting/snapshot/import/index.vue';
import { getBackupList } from '@/api/modules/setting';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);
const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const searchName = ref();

const recoverStatusRef = ref();
const importRef = ref();
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

const drawerVisiable = ref<boolean>(false);

const onCreate = async () => {
    restForm();
    drawerVisiable.value = true;
};

const onImport = () => {
    importRef.value.acceptParams();
};

const handleClose = () => {
    drawerVisiable.value = false;
};

const onChange = async (info: any) => {
    if (!info.edit) {
        await updateSnapshotDescription({ id: info.id, description: info.description });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    }
};

const submitAddSnapshot = (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        await snapshotCreate(snapInfo)
            .then(() => {
                loading.value = false;
                drawerVisiable.value = false;
                search();
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
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
        if (item.type !== 'LOCAL' && item.id !== 0) {
            backupOptions.value.push({ label: i18n.global.t('setting.' + item.type), value: item.type });
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
        info: searchName.value,
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
