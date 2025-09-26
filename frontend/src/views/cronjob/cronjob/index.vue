<template>
    <div>
        <LayoutContent v-loading="loading" v-if="!isRecordShow" :title="$t('menu.cronjob')">
            <template #leftToolBar>
                <el-button type="primary" @click="onOpenDialog('')">
                    {{ $t('commons.button.create') }}{{ $t('menu.cronjob') }}
                </el-button>
                <el-button @click="onOpenGroupDialog()">
                    {{ $t('commons.table.group') }}
                </el-button>
                <el-button-group>
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

                <el-button-group>
                    <el-button @click="onImport">
                        {{ $t('commons.button.import') }}
                    </el-button>
                    <el-button :disabled="selects.length === 0" @click="onExport">
                        {{ $t('commons.button.export') }}
                    </el-button>
                </el-button-group>
            </template>
            <template #rightToolBar>
                <el-select v-model="searchGroupID" @change="search()" clearable class="p-w-200">
                    <template #prefix>{{ $t('commons.table.group') }}</template>
                    <div v-for="item in groupOptions" :key="item.id">
                        <el-option
                            v-if="item.name === 'Default'"
                            :label="$t('commons.table.default')"
                            :value="item.id"
                        />
                        <el-option v-else :label="item.name" :value="item.id" />
                    </div>
                </el-select>
                <TableSearch @search="search()" v-model:searchName="searchName" />
                <TableRefresh @search="search()" />
                <TableSetting title="cronjob-refresh" @search="search()" />
            </template>
            <template #main>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    @sort-change="search"
                    @search="search"
                    :data="data"
                    :heightDiff="300"
                >
                    <el-table-column type="selection" :selectable="selectable" fix />
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
                    <el-table-column :label="$t('commons.table.group')" min-width="120" prop="group">
                        <template #default="{ row }">
                            <fu-select-rw-switch v-model="row.groupID" @change="updateGroup(row)">
                                <template #read>
                                    {{ row.groupBelong === 'Default' ? $t('commons.table.default') : row.groupBelong }}
                                </template>
                                <div v-for="item in groupOptions" :key="item.id">
                                    <el-option
                                        v-if="item.name === 'Default'"
                                        :label="$t('commons.table.default')"
                                        :value="item.id"
                                    />
                                    <el-option v-else :label="item.name" :value="item.id" />
                                </div>
                            </fu-select-rw-switch>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.status')" :min-width="90" prop="status" sortable>
                        <template #default="{ row }">
                            <Status
                                v-if="row.status === 'Enable'"
                                @click="onChangeStatus(row.id, 'disable')"
                                :status="row.status"
                                :operate="true"
                            />
                            <Status
                                v-if="row.status === 'Disable'"
                                @click="onChangeStatus(row.id, 'enable')"
                                :status="row.status"
                                :operate="true"
                            />
                            <Status v-if="row.status === 'Pending'" :status="row.status" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('cronjob.cronSpec')" show-overflow-tooltip :min-width="120">
                        <template #default="{ row }">
                            <div v-for="(item, index) of row.spec.split('&&')" :key="index">
                                <div v-if="row.expand || (!row.expand && index < 3)">
                                    <span>
                                        {{ row.specCustom ? item : transSpecToStr(item) }}
                                    </span>
                                </div>
                            </div>
                            <div v-if="!row.expand && row.spec.split('&&').length > 3">
                                <el-button type="primary" link @click="row.expand = true">
                                    {{ $t('commons.button.expand') }}...
                                </el-button>
                            </div>
                            <div v-if="row.expand && row.spec.split('&&').length > 3">
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
                    <el-table-column
                        :label="$t('cronjob.lastRecordTime')"
                        :min-width="120"
                        show-overflow-tooltip
                        prop="lastRecordTime"
                    >
                        <template #default="{ row }">
                            <el-button v-if="row.lastRecordStatus === 'Success'" icon="Select" link type="success" />
                            <el-button v-if="row.lastRecordStatus === 'Failed'" icon="CloseBold" link type="danger" />
                            <el-button v-if="row.lastRecordStatus === 'Waiting'" :loading="true" link type="info" />
                            {{ row.lastRecordTime }}
                        </template>
                    </el-table-column>
                    <el-table-column :min-width="120" :label="$t('setting.backupAccount')">
                        <template #default="{ row }">
                            <span v-if="!hasBackup(row.type)">-</span>
                            <div v-else>
                                <div v-for="(item, index) of row.sourceAccounts" :key="index">
                                    <div v-if="row.accountExpand || (!row.accountExpand && index < 3)">
                                        <div v-if="row.expand || (!row.expand && index < 3)">
                                            <span type="info">
                                                {{ item === 'localhost' ? $t('setting.LOCAL') : item }}
                                                <el-icon
                                                    v-if="item === row.downloadAccount"
                                                    size="12"
                                                    class="relative top-px left-1"
                                                >
                                                    <Star />
                                                </el-icon>
                                            </span>
                                        </div>
                                    </div>
                                </div>
                                <div v-if="!row.accountExpand && row.sourceAccounts?.length > 3">
                                    <el-button type="primary" link @click="row.accountExpand = true">
                                        {{ $t('commons.button.expand') }}...
                                    </el-button>
                                </div>
                                <div v-if="row.accountExpand && row.sourceAccounts?.length > 3">
                                    <el-button type="primary" link @click="row.accountExpand = false">
                                        {{ $t('commons.button.collapse') }}
                                    </el-button>
                                </div>
                            </div>
                        </template>
                    </el-table-column>
                    <fu-table-operations
                        width="200px"
                        :buttons="buttons"
                        :ellipsis="2"
                        :label="$t('commons.table.operate')"
                        min-width="mobile ? 'auto' : 200"
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
                        <el-checkbox
                            v-if="cleanData"
                            v-model="cleanRemoteData"
                            :label="$t('cronjob.cleanRemoteData')"
                        />
                        <span class="input-help">
                            {{ $t('cronjob.cleanDataHelper') }}
                        </span>
                    </el-form-item>
                </el-form>
            </template>
        </OpDialog>
        <OpDialog ref="opExportRef" @search="search" @submit="onSubmitExport()" />
        <GroupDialog @search="loadGroups" ref="dialogGroupRef" />
        <Records @search="search" ref="dialogRecordRef" />
        <Import @search="search" ref="dialogImportRef" />
        <Backups @search="search" ref="dialogBackupRef" />
    </div>
</template>

<script lang="ts" setup>
import Records from '@/views/cronjob/cronjob/record/index.vue';
import Backups from '@/views/cronjob/cronjob/backup/index.vue';
import Import from '@/views/cronjob/cronjob/import/index.vue';
import { computed, onMounted, reactive, ref } from 'vue';
import {
    deleteCronjob,
    editCronjobGroup,
    exportCronjob,
    searchCronjobPage,
    handleOnce,
    updateStatus,
} from '@/api/modules/cronjob';
import i18n from '@/lang';
import { Cronjob } from '@/api/interface/cronjob';
import GroupDialog from '@/components/group/index.vue';
import { ElMessageBox } from 'element-plus';
import { MsgSuccess } from '@/utils/message';
import { hasBackup, transSpecToStr } from './helper';
import { GlobalStore } from '@/store';
import { getCurrentDateFormatted } from '@/utils/util';
import { getGroupList } from '@/api/modules/group';
import { routerToNameWithQuery } from '@/utils/router';

const globalStore = GlobalStore();
const mobile = computed(() => {
    return globalStore.isMobile();
});

const loading = ref();
const selects = ref<any>([]);
const isRecordShow = ref();
const operateIDs = ref();

const opRef = ref();
const showClean = ref();
const cleanData = ref();
const cleanRemoteData = ref();
const opExportRef = ref();
const dialogImportRef = ref();

const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'cronjob-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('cronjob-page-size')) || 20,
    total: 0,
    orderBy: 'createdAt',
    order: 'null',
});
const searchName = ref();

const defaultGroupID = ref<number>();
const searchGroupID = ref<number>();
const groupOptions = ref();
const dialogGroupRef = ref();

const search = async (column?: any) => {
    paginationConfig.orderBy = column?.order ? column.prop : paginationConfig.orderBy;
    paginationConfig.order = column?.order ? column.order : paginationConfig.order;
    let groupIDs;
    if (searchGroupID.value) {
        groupIDs = searchGroupID.value === defaultGroupID.value ? [searchGroupID.value, 0] : [searchGroupID.value];
    }
    let params = {
        info: searchName.value,
        groupIDs: groupIDs,
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        orderBy: paginationConfig.orderBy,
        order: paginationConfig.order,
    };
    loading.value = true;
    await searchCronjobPage(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            loadGroups();
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const dialogRecordRef = ref();
const dialogBackupRef = ref();

const onOpenDialog = async (id: string) => {
    routerToNameWithQuery('CronjobOperate', { id: id });
};

function selectable(row) {
    return row.status !== 'Pending';
}

const onDelete = async (row: Cronjob.CronjobInfo | null) => {
    let names = [];
    let ids = [];
    showClean.value = false;
    cleanData.value = false;
    cleanRemoteData.value = true;
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
            i18n.global.t('menu.cronjob'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: null,
        params: null,
    });
};

const onSubmitDelete = async () => {
    loading.value = true;
    await deleteCronjob({ ids: operateIDs.value, cleanData: cleanData.value, cleanRemoteData: cleanRemoteData.value })
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};

const onImport = () => {
    dialogImportRef.value.acceptParams();
};

const onExport = async () => {
    let names = [];
    let ids = [];
    for (const item of selects.value) {
        names.push(item.name);
        ids.push(item.id);
    }
    operateIDs.value = ids;
    opExportRef.value.acceptParams({
        title: i18n.global.t('commons.button.export'),
        names: names,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('menu.cronjob'),
            i18n.global.t('commons.button.export'),
        ]),
        api: null,
        params: null,
    });
};

const onSubmitExport = async () => {
    loading.value = true;
    await exportCronjob({ ids: operateIDs.value })
        .then((res) => {
            const downloadUrl = window.URL.createObjectURL(new Blob([res]));
            const a = document.createElement('a');
            a.style.display = 'none';
            a.href = downloadUrl;
            a.download = '1panel-cronjob-' + getCurrentDateFormatted() + '.json';
            const event = new MouseEvent('click');
            a.dispatchEvent(event);
        })
        .finally(() => {
            loading.value = false;
        });
};

const loadGroups = async () => {
    const res = await getGroupList('cronjob');
    groupOptions.value = res.data || [];
    for (const group of groupOptions.value) {
        if (group.name === 'Default') {
            defaultGroupID.value = group.id;
            break;
        }
    }
    for (const item of data.value) {
        if (item.groupID === 0) {
            item.groupBelong = 'Default';
            item.groupID = defaultGroupID.value;
            continue;
        }
        let hasGroup = false;
        for (const group of groupOptions.value) {
            if (item.groupID === group.id) {
                hasGroup = true;
                item.groupBelong = group.name;
            }
        }
        if (!hasGroup) {
            item.groupID = null;
            item.groupBelong = '-';
        }
    }
};

const updateGroup = async (row: any) => {
    await editCronjobGroup(row.id, row.groupID);
    search();
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
};

const onOpenGroupDialog = () => {
    dialogGroupRef.value!.acceptParams({ type: 'cronjob', hideDefaultButton: true });
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
        disabled: (row: any) => {
            return row.status === 'Pending';
        },
    },
    {
        label: i18n.global.t('cronjob.record'),
        click: (row: Cronjob.CronjobInfo) => {
            loadDetail(row);
        },
    },
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: Cronjob.CronjobInfo) => {
            onOpenDialog(row.id + '');
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
