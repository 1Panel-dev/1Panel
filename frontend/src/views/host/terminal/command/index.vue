<template>
    <div style="margin: 20px">
        <ComplexTable :pagination-config="paginationConfig" v-model:selects="selects" :data="data" @search="search">
            <template #toolbar>
                <el-button @click="onCreate()">{{ $t('commons.button.create') }}</el-button>
                <el-button type="danger" plain :disabled="selects.length === 0" @click="batchDelete(null)">
                    {{ $t('commons.button.delete') }}
                </el-button>
            </template>
            <el-table-column type="selection" fix />
            <el-table-column :label="$t('commons.table.name')" min-width="100" prop="name" fix />
            <el-table-column :label="$t('terminal.command')" min-width="300" show-overflow-tooltip prop="command" />
            <fu-table-operations type="icon" :buttons="buttons" :label="$t('commons.table.operate')" fix />
        </ComplexTable>

        <el-dialog v-model="cmdVisiable" :title="$t('terminal.addHost')" width="30%">
            <el-form ref="commandInfoRef" label-width="100px" label-position="left" :model="commandInfo" :rules="rules">
                <el-form-item :label="$t('commons.table.name')" prop="name">
                    <el-input clearable v-model="commandInfo.name" />
                </el-form-item>
                <el-form-item :label="$t('terminal.command')" prop="command">
                    <el-input type="textarea" clearable v-model="commandInfo.command" />
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="cmdVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button type="primary" @click="submitAddCommand(commandInfoRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
    </div>
</template>

<script setup lang="ts">
import ComplexTable from '@/components/complex-table/index.vue';
import { Command } from '@/api/interface/command';
import { addCommand, editCommand, deleteCommand, getCommandPage } from '@/api/modules/command';
import { reactive, ref } from '@vue/runtime-core';
import { useDeleteData } from '@/hooks/use-delete-data';
import type { ElForm } from 'element-plus';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElMessage } from 'element-plus';

const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    page: 1,
    pageSize: 5,
    total: 0,
});
const commandSearch = reactive({
    page: 1,
    pageSize: 5,
    info: '',
});
type FormInstance = InstanceType<typeof ElForm>;
const commandInfoRef = ref<FormInstance>();
const rules = reactive({
    name: [Rules.requiredInput],
    command: [Rules.requiredInput],
});
let operate = ref<string>('create');

let commandInfo = reactive<Command.CommandOperate>({
    id: 0,
    name: '',
    command: '',
});

const cmdVisiable = ref<boolean>(false);

const onCreate = async () => {
    restcommandForm();
    operate.value = 'create';
    cmdVisiable.value = true;
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
        cmdVisiable.value = false;
        search();
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
    });
};

const onEdit = async (row: Command.CommandInfo | null) => {
    if (row !== null) {
        commandInfo.id = row.id;
        commandInfo.name = row.name;
        commandInfo.command = row.command;
        operate.value = 'edit';
        cmdVisiable.value = true;
    }
};

const batchDelete = async (row: Command.CommandInfo | null) => {
    let ids: Array<number> = [];
    if (row === null) {
        selects.value.forEach((item: Command.CommandInfo) => {
            ids.push(item.id);
        });
    } else {
        ids.push(row.id);
    }
    await useDeleteData(deleteCommand, { ids: ids }, 'commons.msg.delete', true);
    search();
};

function restcommandForm() {
    if (commandInfoRef.value) {
        commandInfoRef.value.resetFields();
    }
}
const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        icon: 'Edit',
        click: onEdit,
    },
    {
        label: i18n.global.t('commons.button.delete'),
        icon: 'Delete',
        click: batchDelete,
    },
];

const search = async () => {
    commandSearch.page = paginationConfig.page;
    commandSearch.pageSize = paginationConfig.pageSize;
    const res = await getCommandPage(commandSearch);
    data.value = res.data.items;
    paginationConfig.total = res.data.total;
};

function onInit() {
    search();
}
defineExpose({
    onInit,
});
</script>
