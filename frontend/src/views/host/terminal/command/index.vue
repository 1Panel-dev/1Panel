<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('terminal.quickCommand', 2)">
            <template #prompt>
                <el-alert type="info" :title="$t('terminal.quickCommandHelper')" :closable="false" />
            </template>
            <template #toolbar>
                <div class="flex w-full flex-col gap-4 md:justify-between md:flex-row">
                    <div class="flex flex-wrap gap-4">
                        <el-button type="primary" @click="onCreate()">
                            {{ $t('commons.button.create') }}
                        </el-button>
                        <el-button type="primary" plain @click="onOpenGroupDialog()">
                            {{ $t('terminal.manageGroup') }}
                        </el-button>
                        <el-button type="primary" plain :disabled="selects.length === 0" @click="batchDelete(null)">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </div>
                </div>
            </template>
            <template #search>
                <el-row :gutter="5">
                    <el-col :xs="24" :sm="20" :md="20" :lg="20" :xl="20">
                        <el-select v-model="group" @change="search()" clearable class="p-w-200">
                            <template #prefix>{{ $t('terminal.group') }}</template>
                            <el-option :label="$t('commons.table.all')" value=""></el-option>
                            <div v-for="item in groupList" :key="item.name">
                                <el-option :value="item.id" :label="item.name" />
                            </div>
                        </el-select>
                    </el-col>
                    <el-col :xs="24" :sm="4" :md="4" :lg="4" :xl="4">
                        <TableSearch @search="search()" v-model:searchName="commandReq.name" />
                    </el-col>
                </el-row>
            </template>
            <template #main>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    :data="data"
                    @sort-change="search"
                    @search="search"
                >
                    <el-table-column type="selection" fix />
                    <el-table-column
                        :label="$t('commons.table.name')"
                        show-overflow-tooltip
                        min-width="100"
                        prop="name"
                        fix
                        sortable
                    />
                    <el-table-column
                        :label="$t('terminal.command')"
                        min-width="300"
                        show-overflow-tooltip
                        prop="command"
                        sortable
                    />
                    <el-table-column
                        :label="$t('commons.table.group')"
                        show-overflow-tooltip
                        min-width="100"
                        prop="groupBelong"
                        fix
                    />
                    <fu-table-operations width="200px" :buttons="buttons" :label="$t('commons.table.operate')" fix />
                </ComplexTable>
            </template>
        </LayoutContent>
        <el-drawer
            v-model="cmdVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="30%"
        >
            <template #header>
                <DrawerHeader
                    :header="$t('commons.button.' + operate) + $t('terminal.quickCommand').toLowerCase()"
                    :back="handleClose"
                />
            </template>
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form
                        @submit.prevent
                        ref="commandInfoRef"
                        label-width="100px"
                        label-position="top"
                        :model="commandInfo"
                        :rules="rules"
                    >
                        <el-form-item :label="$t('commons.table.name')" prop="name">
                            <el-input clearable v-model="commandInfo.name" />
                        </el-form-item>
                        <el-form-item :label="$t('commons.table.group')" prop="name">
                            <el-select filterable v-model="commandInfo.groupID" clearable style="width: 100%">
                                <div v-for="item in groupList" :key="item.id">
                                    <el-option :label="item.name" :value="item.id" />
                                </div>
                            </el-select>
                        </el-form-item>
                        <el-form-item :label="$t('terminal.command')" prop="command">
                            <el-input type="textarea" clearable v-model="commandInfo.command" />
                        </el-form-item>
                    </el-form>
                </el-col>
            </el-row>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="cmdVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button type="primary" @click="submitAddCommand(commandInfoRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>

        <OpDialog ref="opRef" @search="search" />
        <GroupDialog @search="loadGroups" ref="dialogGroupRef" />
        <GroupChangeDialog @search="search" @change="onChangeGroup" ref="dialogGroupChangeRef" />
    </div>
</template>

<script setup lang="ts">
import { Command } from '@/api/interface/command';
import GroupDialog from '@/components/group/index.vue';
import GroupChangeDialog from '@/components/group/change.vue';
import { addCommand, editCommand, deleteCommand, getCommandPage } from '@/api/modules/host';
import { reactive, ref } from 'vue';
import type { ElForm } from 'element-plus';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';
import { GetGroupList } from '@/api/modules/group';

const loading = ref();
const data = ref();
const selects = ref<any>([]);
const groupList = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'terminal-command-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
    orderBy: 'name',
    order: 'ascending',
});
const info = ref();
const group = ref<string>('');
const dialogGroupChangeRef = ref();

const opRef = ref();

type FormInstance = InstanceType<typeof ElForm>;
const commandInfoRef = ref<FormInstance>();
const rules = reactive({
    name: [Rules.requiredInput],
    command: [Rules.requiredInput],
});
let operate = ref<string>('create');

const acceptParams = () => {
    search();
    loadGroups();
};

const defaultGroupID = ref();
let commandInfo = reactive<Command.CommandOperate>({
    id: 0,
    name: '',
    groupID: 0,
    command: '',
});

const commandReq = reactive({
    name: '',
});

const cmdVisible = ref<boolean>(false);

const loadGroups = async () => {
    const res = await GetGroupList({ type: 'command' });
    groupList.value = res.data;
    for (const group of groupList.value) {
        if (group.isDefault) {
            defaultGroupID.value = group.id;
            break;
        }
    }
};

const onCreate = async () => {
    commandInfo.id = 0;
    commandInfo.name = '';
    commandInfo.command = '';
    commandInfo.groupID = defaultGroupID.value;
    operate.value = 'create';
    cmdVisible.value = true;
};

const handleClose = () => {
    cmdVisible.value = false;
};

const dialogGroupRef = ref();
const onOpenGroupDialog = () => {
    dialogGroupRef.value!.acceptParams({ type: 'command' });
};

const submitAddCommand = (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (operate.value === 'create') {
            await addCommand(commandInfo);
        } else {
            await editCommand(commandInfo);
        }
        cmdVisible.value = false;
        search();
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    });
};

const onChangeGroup = async (groupID: number) => {
    commandInfo.groupID = groupID;
    await editCommand(commandInfo);
    search();
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
};

const batchDelete = async (row: Command.CommandInfo | null) => {
    let names = [];
    let ids = [];
    if (row) {
        ids = [row.id];
        names = [row.name];
    } else {
        selects.value.forEach((item: Command.CommandInfo) => {
            ids.push(item.id);
            names.push(item.name);
        });
    }
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('terminal.quickCommand'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: deleteCommand,
        params: { ids: ids },
    });
};

const buttons = [
    {
        label: i18n.global.t('terminal.groupChange'),
        click: (row: any) => {
            commandInfo = row;
            dialogGroupChangeRef.value!.acceptParams({
                group: row.groupBelong,
                groupType: 'command',
            });
        },
    },
    {
        label: i18n.global.t('commons.button.edit'),
        icon: 'Edit',
        click: (row: any) => {
            commandInfo = row;
            operate.value = 'edit';
            cmdVisible.value = true;
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
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        groupID: Number(group.value),
        info: info.value,
        orderBy: paginationConfig.orderBy,
        order: paginationConfig.order,
        name: commandReq.name,
    };
    loading.value = true;
    await getCommandPage(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>
