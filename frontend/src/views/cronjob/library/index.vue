<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('logs.login')">
            <template #leftToolBar>
                <el-button type="primary" @click="onOpenDialog('create')">
                    {{ $t('commons.button.create') }}
                </el-button>
                <el-dropdown @command="handleSyncOp" class="mr-2.5">
                    <el-button type="primary" plain>
                        {{ $t('commons.button.sync') }}
                        <el-icon><arrow-down /></el-icon>
                    </el-button>
                    <template #dropdown>
                        <el-dropdown-menu>
                            <el-dropdown-item command="sync">
                                {{ $t('cronjob.library.syncNow') }}
                            </el-dropdown-item>
                            <el-dropdown-item v-if="scriptSync === 'Disable'" command="turnOnSync">
                                {{ $t('cronjob.library.turnOnSync') }}
                            </el-dropdown-item>
                            <el-dropdown-item v-if="scriptSync === 'Enable'" command="turnOffSync">
                                {{ $t('cronjob.library.turnOffSync') }}
                            </el-dropdown-item>
                        </el-dropdown-menu>
                    </template>
                </el-dropdown>

                <el-button type="primary" plain @click="onOpenGroupDialog()">
                    {{ $t('commons.table.group') }}
                </el-button>
                <el-button plain :disabled="selects.length === 0" @click="onDelete(null)">
                    {{ $t('commons.button.delete') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <el-select v-model="group" @change="search()" clearable class="p-w-200">
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
                <TableSearch @search="search()" v-model:searchName="searchInfo" />
                <TableRefresh @search="search()" />
                <TableSetting title="script-refresh" @search="search()" />
            </template>
            <template #main>
                <ComplexTable
                    v-model:selects="selects"
                    :pagination-config="paginationConfig"
                    :data="data"
                    @search="search"
                    :heightDiff="300"
                >
                    <el-table-column type="selection" fix />
                    <el-table-column :label="$t('commons.table.name')" show-overflow-tooltip prop="name" min-width="60">
                        <template #default="{ row }">
                            <el-text type="primary" class="cursor-pointer" @click="showScript(row)">
                                {{ row.name }}
                            </el-text>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('cronjob.library.isInteractive')" prop="isInteractive" min-width="60">
                        <template #default="{ row }">
                            <div class="-mb-1">
                                <el-icon v-if="row.isInteractive"><Check /></el-icon>
                                <el-icon v-else><Minus /></el-icon>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.group')" min-width="120" prop="group">
                        <template #default="{ row }">
                            <el-button class="mr-3" size="small" v-if="row.isSystem">{{ $t('menu.system') }}</el-button>
                            <span v-if="row.groupBelong">
                                <el-button size="small" v-for="(item, index) in row.groupBelong" :key="index">
                                    <span v-if="item === 'Default'">
                                        {{ $t('commons.table.default') }}
                                    </span>
                                    <span v-else>{{ item }}</span>
                                </el-button>
                            </span>
                        </template>
                    </el-table-column>
                    <el-table-column
                        min-width="120"
                        :label="$t('commons.table.description')"
                        show-overflow-tooltip
                        prop="description"
                    />
                    <el-table-column
                        prop="createdAt"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                        show-overflow-tooltip
                    />
                    <fu-table-operations
                        width="300px"
                        :buttons="buttons"
                        :ellipsis="10"
                        :label="$t('commons.table.operate')"
                        min-width="mobile ? 'auto' : 200"
                        :fixed="mobile ? false : 'right'"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <OpDialog ref="opRef" @search="search"></OpDialog>
        <OperateDialog @search="search" ref="dialogRef" />
        <GroupDialog @search="loadGroupOptions" ref="dialogGroupRef" />
        <CodemirrorDrawer ref="myDetail" />
        <TerminalDialog ref="runRef" />
        <TaskLog ref="taskLogRef" width="70%" @close="search" />
    </div>
</template>

<script setup lang="ts">
import { dateFormat, deepCopy, getCurrentDateFormatted, newUUID } from '@/utils/util';
import GroupDialog from '@/components/group/index.vue';
import TaskLog from '@/components/log/task/index.vue';
import OperateDialog from '@/views/cronjob/library/operate/index.vue';
import TerminalDialog from '@/views/cronjob/library/run/index.vue';
import { deleteScript, searchScript, syncScript } from '@/api/modules/cronjob';
import { onMounted, reactive, ref } from 'vue';
import { Cronjob } from '@/api/interface/cronjob';
import i18n from '@/lang';
import { GlobalStore } from '@/store';
import { getGroupList } from '@/api/modules/group';
import CodemirrorDrawer from '@/components/codemirror-pro/drawer.vue';
import { MsgSuccess } from '@/utils/message';
import { getSettingBy, updateSetting } from '@/api/modules/setting';

const globalStore = GlobalStore();
const mobile = computed(() => {
    return globalStore.isMobile();
});
const myDetail = ref();

const loading = ref();
const selects = ref<any>([]);
const opRef = ref();

const runRef = ref();
const taskLogRef = ref();
const scriptSync = ref();

const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'script-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('script-page-size')) || 20,
    total: 0,
});
const searchInfo = ref<string>('');
const group = ref<string>('');
const groupOptions = ref();

const dialogGroupRef = ref();
const onOpenGroupDialog = () => {
    dialogGroupRef.value!.acceptParams({ type: 'script' });
};

const dialogRef = ref();
const onOpenDialog = async (
    title: string,
    rowData: Partial<Cronjob.ScriptOperate> = {
        name: '',
    },
) => {
    let params = {
        title,
        rowData: { ...rowData },
    };
    dialogRef.value!.acceptParams(params);
};

const showScript = async (row: any) => {
    let param = {
        header: i18n.global.t('commons.button.view') + ' - ' + row.name,
        detailInfo: row.script,
        mode: 'shell',
    };
    myDetail.value!.acceptParams(param);
};

const onDelete = async (row: Cronjob.ScriptInfo | null) => {
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
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('cronjob.library.script'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: deleteScript,
        params: ids,
    });
};

const onSync = async () => {
    ElMessageBox.confirm(i18n.global.t('cronjob.library.syncHelper'), i18n.global.t('cronjob.library.syncNow'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        let taskID = newUUID();
        await syncScript(taskID)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                openTaskLog(taskID);
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const openTaskLog = (taskID: string) => {
    taskLogRef.value.openWithTaskID(taskID, true, 'local');
};

const loadSyncStatus = async () => {
    const res = await getSettingBy('ScriptSync');
    scriptSync.value = res.data;
};
const handleSyncOp = async (command: string) => {
    let val = 'Enable';
    switch (command) {
        case 'sync':
            onSync();
            return;
        case 'turnOnSync':
            val = 'Enable';
            break;
        case 'turnOffSync':
            val = 'Disable';
            break;
        default:
            return;
    }
    ElMessageBox.confirm(
        i18n.global.t('cronjob.library.' + command + 'Helper'),
        i18n.global.t('cronjob.library.' + command),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'info',
        },
    ).then(async () => {
        loading.value = true;
        await updateSetting({ key: 'ScriptSync', value: val })
            .then(() => {
                loadSyncStatus();
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const search = async () => {
    let params = {
        info: searchInfo.value,
        groupID: Number(group.value),
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    loading.value = true;
    await searchScript(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items;
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const loadGroupOptions = async () => {
    const res = await getGroupList('script');
    groupOptions.value = res.data || [];
};

const buttons = [
    {
        label: i18n.global.t('commons.button.handle'),
        click: (row: Cronjob.ScriptInfo) => {
            ElMessageBox.confirm(
                i18n.global.t('cronjob.library.handleHelper', [
                    globalStore.currentNode === 'local' ? globalStore.getMasterAlias() : globalStore.currentNode,
                    row.name,
                ]),
                i18n.global.t('commons.button.handle'),
                {
                    confirmButtonText: i18n.global.t('commons.button.confirm'),
                    cancelButtonText: i18n.global.t('commons.button.cancel'),
                    type: 'info',
                },
            ).then(() => {
                runRef.value!.acceptParams({ scriptID: row.id, scriptName: row.name });
            });
        },
    },
    {
        label: i18n.global.t('commons.button.clone'),
        disabled: (row: any) => {
            return !row.isSystem;
        },
        click: (row: Cronjob.ScriptInfo) => {
            let item = deepCopy(row) as Cronjob.ScriptInfo;
            item.id = 0;
            item.name += '-' + getCurrentDateFormatted();
            item.groupList = row.groupList || [];
            item.groupBelong = row.groupBelong || [];
            onOpenDialog('clone', item);
        },
    },
    {
        label: i18n.global.t('commons.button.edit'),
        disabled: (row: any) => {
            return row.isSystem;
        },
        click: (row: Cronjob.ScriptInfo) => {
            onOpenDialog('edit', row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        disabled: (row: any) => {
            return row.isSystem;
        },
        click: (row: Cronjob.ScriptInfo) => {
            onDelete(row);
        },
    },
];

onMounted(() => {
    search();
    loadSyncStatus();
    loadGroupOptions();
});
</script>
