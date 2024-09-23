<template>
    <div>
        <LayoutContent v-loading="loading" v-if="!isRecordShow" :title="$t('setting.snapshot')">
            <template #leftToolBar>
                <el-button type="primary" @click="onCreate()">
                    {{ $t('setting.createSnapshot') }}
                </el-button>
                <el-button type="primary" plain @click="onImport()">
                    {{ $t('setting.importSnapshot') }}
                </el-button>
                <el-button type="primary" plain @click="onIgnore()">
                    {{ $t('setting.ignoreRule') }}
                </el-button>
                <el-button type="primary" plain :disabled="selects.length === 0" @click="batchDelete(null)">
                    {{ $t('commons.button.delete') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <TableSearch @search="search()" v-model:searchName="searchName" class="mr-2.5" />
                <TableSetting ref="timerRef" @search="search()" />
            </template>
            <template #main>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    :data="data"
                    class="mt-5"
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
                            <div v-for="(item, index) of row.from.split(',')" :key="index" class="mt-1">
                                <div v-if="row.expand || (!row.expand && index < 3)">
                                    <span v-if="row.from" type="info">
                                        <span>
                                            {{ $t('setting.' + item) }}
                                        </span>
                                        <el-icon
                                            v-if="item === row.defaultDownload"
                                            size="12"
                                            class="relative top-px left-1"
                                        >
                                            <Star />
                                        </el-icon>
                                    </span>
                                    <span v-else>-</span>
                                </div>
                            </div>
                            <div v-if="!row.expand && row.from.split(',').length > 3">
                                <el-button type="primary" link @click="row.expand = true">
                                    {{ $t('commons.button.expand') }}...
                                </el-button>
                            </div>
                            <div v-if="row.expand && row.from.split(',').length > 3">
                                <el-button type="primary" link @click="row.expand = false">
                                    {{ $t('commons.button.collapse') }}
                                </el-button>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('file.size')" prop="size" min-width="60" show-overflow-tooltip>
                        <template #default="{ row }">
                            <span v-if="row.size">
                                {{ computeSize(row.size) }}
                            </span>
                            <span v-else>-</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.status')" min-width="80" prop="status">
                        <template #default="{ row }">
                            <el-button
                                v-if="row.status === 'Waiting' || row.status === 'OnSaveData'"
                                type="primary"
                                link
                            >
                                {{ $t('commons.table.statusWaiting') }}
                            </el-button>
                            <el-button v-if="row.status === 'Failed'" @click="reCreate(row)" type="danger" link>
                                {{ $t('commons.status.error') }}
                            </el-button>
                            <el-tag v-if="row.status === 'Success'" type="success">
                                {{ $t('commons.status.success') }}
                            </el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.description')" prop="description" show-overflow-tooltip>
                        <template #default="{ row }">
                            <fu-input-rw-switch v-model="row.description" @blur="onChange(row)" />
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
        <SnapshotCreate ref="createRef" @search="search()" />
        <SnapshotImport ref="importRef" @search="search()" />
        <IgnoreRule ref="ignoreRef" @search="search()" />

        <OpDialog ref="opRef" @search="search" @submit="onSubmitDelete()">
            <template #content>
                <el-form class="mt-4 mb-1" ref="deleteForm" label-position="left">
                    <el-form-item>
                        <el-checkbox v-model="cleanData" :label="$t('cronjob.cleanData')" />
                        <span class="input-help">
                            {{ $t('setting.deleteHelper') }}
                        </span>
                    </el-form-item>
                </el-form>
            </template>
        </OpDialog>
        <TaskLog ref="taskLogRef" width="70%" />
    </div>
</template>

<script setup lang="ts">
import { searchSnapshotPage, snapshotDelete, snapshotRecreate, updateSnapshotDescription } from '@/api/modules/setting';
import { onMounted, reactive, ref } from 'vue';
import { computeSize, dateFormat } from '@/utils/util';
import { ElForm } from 'element-plus';
import IgnoreRule from '@/views/setting/snapshot/ignore-rule/index.vue';
import i18n from '@/lang';
import { Setting } from '@/api/interface/setting';
import TaskLog from '@/components/task-log/index.vue';
import RecoverStatus from '@/views/setting/snapshot/status/index.vue';
import SnapshotImport from '@/views/setting/snapshot/import/index.vue';
import SnapshotCreate from '@/views/setting/snapshot/create/index.vue';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);
const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    cacheSizeKey: 'snapshot-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const searchName = ref();

const opRef = ref();

const createRef = ref();
const ignoreRef = ref();
const recoverStatusRef = ref();
const importRef = ref();
const isRecordShow = ref();
const taskLogRef = ref();

const operateIDs = ref();
const cleanData = ref();

const onImport = () => {
    let names = [];
    for (const item of data.value) {
        names.push(item.name);
    }
    importRef.value.acceptParams({ names: names });
};

const onCreate = () => {
    createRef.value.acceptParams();
};

const reCreate = (row: any) => {
    ElMessageBox.confirm(row.message, i18n.global.t('setting.reCreate'), {
        confirmButtonText: i18n.global.t('commons.button.retry'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'error',
    }).then(async () => {
        await snapshotRecreate(row.id)
            .then(() => {
                loading.value = false;
                openTaskLog(row.taskID);
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};
const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID);
};

const onIgnore = () => {
    ignoreRef.value.acceptParams();
};

const onChange = async (info: any) => {
    await updateSnapshotDescription({ id: info.id, description: info.description });
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
};

const batchDelete = async (row: Setting.SnapshotInfo | null) => {
    let names = [];
    let ids = [];
    if (row) {
        ids.push(row.id);
        names.push(row.name);
    } else {
        selects.value.forEach((item: Setting.SnapshotInfo) => {
            ids.push(item.id);
            names.push(item.name);
        });
    }
    operateIDs.value = ids;
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('setting.snapshot'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: null,
        params: null,
    });
};

const onSubmitDelete = async () => {
    loading.value = true;
    await snapshotDelete({ ids: operateIDs.value, deleteWithFile: cleanData.value })
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
    loading.value = true;
    await searchSnapshotPage(params)
        .then((res) => {
            loading.value = false;
            cleanData.value = false;
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

onMounted(() => {
    search();
});
</script>
