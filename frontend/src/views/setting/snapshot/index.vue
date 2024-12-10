<template>
    <div>
        <LayoutContent v-loading="loading" v-if="!isRecordShow" :title="$t('setting.snapshot', 2)">
            <template #toolbar>
                <div class="flex justify-between gap-2 flex-wrap sm:flex-row">
                    <div class="flex flex-wrap gap-3">
                        <el-button type="primary" @click="onCreate()">
                            {{ $t('setting.createSnapshot') }}
                        </el-button>
                        <el-button type="primary" plain @click="onImport()">
                            {{ $t('setting.importSnapshot') }}
                        </el-button>
                        <el-button type="primary" plain @click="onIgnore()">
                            {{ $t('setting.editIgnoreRule') }}
                        </el-button>
                        <el-button type="primary" plain :disabled="selects.length === 0" @click="batchDelete(null)">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </div>
                    <div class="flex flex-wrap gap-3">
                        <TableSetting ref="timerRef" @search="search()" />
                        <TableSearch @search="search()" v-model:searchName="searchName" />
                    </div>
                </div>
            </template>
            <template #main>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    :data="data"
                    @sort-change="search"
                    style="margin-top: 20px"
                    @search="search"
                >
                    <el-table-column type="selection" fix />
                    <el-table-column
                        show-overflow-tooltip
                        :label="$t('commons.table.name')"
                        min-width="100"
                        prop="name"
                        sortable
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
                            <div v-if="row.hasLoad">
                                <span v-if="row.size">
                                    {{ computeSize(row.size) }}
                                </span>
                                <span v-else>-</span>
                            </div>
                            <div v-if="!row.hasLoad">
                                <el-button link loading></el-button>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.status')" min-width="80" prop="status">
                        <template #default="{ row }">
                            <el-button
                                v-if="row.status === 'Waiting' || row.status === 'OnSaveData'"
                                type="primary"
                                @click="onLoadStatus(row)"
                                link
                            >
                                {{ $t('commons.table.statusWaiting') }}
                            </el-button>
                            <el-button v-if="row.status === 'Failed'" type="danger" @click="onLoadStatus(row)" link>
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
                        prop="created_at"
                        sortable
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
        <el-drawer v-model="drawerVisible" size="50%" :close-on-click-modal="false" :close-on-press-escape="false">
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
                        <el-form-item :label="$t('setting.backupAccount')" prop="fromAccounts">
                            <el-select multiple @change="changeAccount" v-model="snapInfo.fromAccounts" clearable>
                                <el-option
                                    v-for="item in backupOptions"
                                    :key="item.label"
                                    :value="item.value"
                                    :label="item.label"
                                />
                            </el-select>
                        </el-form-item>
                        <el-form-item :label="$t('cronjob.default_download_path')" prop="defaultDownload">
                            <el-select v-model="snapInfo.defaultDownload" clearable>
                                <el-option
                                    v-for="item in accountOptions"
                                    :key="item.label"
                                    :value="item.value"
                                    :label="item.label"
                                />
                            </el-select>
                        </el-form-item>
                        <el-form-item :label="$t('setting.compressPassword')" prop="secret">
                            <el-input v-model="snapInfo.secret"></el-input>
                        </el-form-item>
                        <el-form-item :label="$t('commons.table.description')" prop="description">
                            <el-input type="textarea" clearable v-model="snapInfo.description" />
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button :disabled="loading" @click="drawerVisible = false">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button :disabled="loading" type="primary" @click="submitAddSnapshot(snapRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>

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
        <SnapStatus ref="snapStatusRef" @search="search" />
        <IgnoreRule ref="ignoreRef" />
    </div>
</template>

<script setup lang="ts">
import DrawerHeader from '@/components/drawer-header/index.vue';
import {
    snapshotCreate,
    searchSnapshotPage,
    snapshotDelete,
    updateSnapshotDescription,
    loadSnapshotSize,
} from '@/api/modules/setting';
import { onMounted, reactive, ref } from 'vue';
import { computeSize, dateFormat } from '@/utils/util';
import { ElForm } from 'element-plus';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { Setting } from '@/api/interface/setting';
import IgnoreRule from '@/views/setting/snapshot/ignore-rule/index.vue';
import SnapStatus from '@/views/setting/snapshot/snap_status/index.vue';
import RecoverStatus from '@/views/setting/snapshot/status/index.vue';
import SnapshotImport from '@/views/setting/snapshot/import/index.vue';
import { getBackupList } from '@/api/modules/setting';
import { MsgSuccess } from '@/utils/message';

const loading = ref(false);
const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    cacheSizeKey: 'snapshot-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
    orderBy: 'created_at',
    order: 'null',
});
const searchName = ref();

const opRef = ref();
const ignoreRef = ref();

const snapStatusRef = ref();
const recoverStatusRef = ref();
const importRef = ref();
const isRecordShow = ref();
const backupOptions = ref();
const accountOptions = ref();

const operateIDs = ref();

type FormInstance = InstanceType<typeof ElForm>;
const snapRef = ref<FormInstance>();
const rules = reactive({
    fromAccounts: [Rules.requiredSelect],
    defaultDownload: [Rules.requiredSelect],
});

let snapInfo = reactive<Setting.SnapshotCreate>({
    id: 0,
    from: '',
    defaultDownload: '',
    fromAccounts: [],
    description: '',
    secret: '',
});
const cleanData = ref();

const drawerVisible = ref<boolean>(false);

const onCreate = async () => {
    restForm();
    drawerVisible.value = true;
};

const onImport = () => {
    let names = [];
    for (const item of data.value) {
        names.push(item.name);
    }
    importRef.value.acceptParams({ names: names });
};

const onIgnore = () => {
    ignoreRef.value.acceptParams();
};

const handleClose = () => {
    drawerVisible.value = false;
};

const onChange = async (info: any) => {
    await updateSnapshotDescription({ id: info.id, description: info.description });
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
};

const submitAddSnapshot = (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        snapInfo.from = snapInfo.fromAccounts.join(',');
        await snapshotCreate(snapInfo)
            .then(() => {
                loading.value = false;
                drawerVisible.value = false;
                search();
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const onLoadStatus = (row: Setting.SnapshotInfo) => {
    snapStatusRef.value.acceptParams({
        id: row.id,
        from: row.from,
        defaultDownload: row.defaultDownload,
        description: row.description,
    });
};

const loadBackups = async () => {
    const res = await getBackupList();
    backupOptions.value = [];
    for (const item of res.data) {
        if (item.id !== 0) {
            backupOptions.value.push({ label: i18n.global.t('setting.' + item.type), value: item.type });
        }
    }
    changeAccount();
};

const changeAccount = async () => {
    accountOptions.value = [];
    let isInAccounts = false;
    for (const item of backupOptions.value) {
        let exist = false;
        for (const ac of snapInfo.fromAccounts) {
            if (item.value == ac) {
                exist = true;
                break;
            }
        }
        if (exist) {
            if (item.value === snapInfo.defaultDownload) {
                isInAccounts = true;
            }
            accountOptions.value.push(item);
        }
    }
    if (!isInAccounts) {
        snapInfo.defaultDownload = '';
    }
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
    await searchSnapshotPage(params)
        .then((res) => {
            loading.value = false;
            loadSize(params);
            cleanData.value = false;
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const loadSize = async (params: any) => {
    await loadSnapshotSize(params)
        .then((res) => {
            let stats = res.data || [];
            if (stats.length === 0) {
                return;
            }
            for (const snap of data.value) {
                for (const item of stats) {
                    if (snap.id === item.id) {
                        snap.hasLoad = true;
                        snap.size = item.size;
                        break;
                    }
                }
            }
        })
        .catch(() => {
            loading.value = false;
        });
};

onMounted(() => {
    search();
    loadBackups();
});
</script>
