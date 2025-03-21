<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('logs.login')">
            <template #leftToolBar>
                <el-button type="primary" @click="onOpenDialog('create')">
                    {{ $t('commons.button.add') }}
                </el-button>
                <el-button type="primary" plain @click="onOpenGroupDialog()">
                    {{ $t('commons.table.group') }}
                </el-button>
                <el-button plain :disabled="selects.length === 0" @click="onDelete(null)">
                    {{ $t('commons.button.delete') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <el-select v-model="group" @change="search()" clearable class="p-w-200 mr-2.5">
                    <template #prefix>{{ $t('commons.table.group') }}</template>
                    <div v-for="item in groupOptions" :key="item.id">
                        <el-option :label="item.name" :value="item.id" />
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
                            <el-text type="primary" class="cursor-pointer" @click="showScript(row.script)">
                                {{ row.name }}
                            </el-text>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.group')" min-width="120" prop="group">
                        <template #default="{ row }">
                            <el-button class="mr-3" size="small" v-if="row.isSystem">system</el-button>
                            <span v-if="row.groupBelong">
                                <el-button size="small" v-for="(item, index) in row.groupBelong" :key="index">
                                    {{ item }}
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
    </div>
</template>

<script setup lang="ts">
import { dateFormat } from '@/utils/util';
import GroupDialog from '@/components/group/index.vue';
import OperateDialog from '@/views/cronjob/library/operate/index.vue';
import TerminalDialog from '@/views/cronjob/library/run/index.vue';
import { deleteScript, searchScript } from '@/api/modules/cronjob';
import { onMounted, reactive, ref } from '@vue/runtime-core';
import { Cronjob } from '@/api/interface/cronjob';
import i18n from '@/lang';
import { GlobalStore } from '@/store';
import { getGroupList } from '@/api/modules/group';
import CodemirrorDrawer from '@/components/codemirror-pro/drawer.vue';

const globalStore = GlobalStore();
const mobile = computed(() => {
    return globalStore.isMobile();
});
const myDetail = ref();

const loading = ref();
const selects = ref<any>([]);
const opRef = ref();

const runRef = ref();

const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'script-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const searchInfo = ref<string>('');
const group = ref<string>('');
const groupOptions = ref();

const dialogGroupRef = ref();
const onOpenGroupDialog = () => {
    dialogGroupRef.value!.acceptParams({ type: 'script', hideDefaultButton: true });
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

const showScript = async (script: string) => {
    let param = {
        header: i18n.global.t('commons.button.view'),
        detailInfo: script,
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
            runRef.value!.acceptParams({ scriptID: row.id, scriptName: row.name });
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
    loadGroupOptions();
});
</script>
