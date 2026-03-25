<template>
    <DrawerPro
        v-model="open"
        :header="title"
        :resource="dialogData.title === 'create' ? '' : dialogData.rowData?.name"
        @close="handleClose"
        size="small"
    >
        <el-form
            @submit.prevent
            ref="formRef"
            label-width="100px"
            label-position="top"
            :model="dialogData.rowData"
            :rules="rules"
        >
            <el-form-item v-if="dialogData.title === 'create'" :label="$t('terminal.batchInput')">
                <el-switch v-model="batchMode" />
            </el-form-item>
            <el-form-item v-if="!batchMode" :label="$t('commons.table.name')" prop="name">
                <el-input clearable v-model="dialogData.rowData!.name" />
            </el-form-item>
            <el-form-item :label="$t('commons.table.group')" prop="groupID">
                <el-select filterable v-model="dialogData.rowData!.groupID" clearable style="width: 100%">
                    <div v-for="item in groupList" :key="item.id">
                        <el-option
                            v-if="item.name === 'Default'"
                            :label="$t('commons.table.default')"
                            :value="item.id"
                        />
                        <el-option v-else :label="item.name" :value="item.id" />
                    </div>
                </el-select>
            </el-form-item>
            <el-form-item v-if="!batchMode" :label="$t('terminal.command')" prop="command">
                <el-input type="textarea" clearable v-model="dialogData.rowData!.command" />
            </el-form-item>
            <el-form-item v-else :label="$t('terminal.command')">
                <el-input
                    v-model="batchCommandText"
                    type="textarea"
                    clearable
                    :autosize="{ minRows: 8, maxRows: 16 }"
                />
                <div class="input-help mt-1 text-[var(--el-text-color-secondary)] text-xs leading-5">
                    {{ $t('terminal.quickCommandBatchHelper') }}
                </div>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="open = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { MsgError, MsgSuccess } from '@/utils/message';
import { Command } from '@/api/interface/command';
import { addCommand, editCommand, importCommands } from '@/api/modules/command';
import { getGroupList } from '@/api/modules/group';

const loading = ref();
const open = ref();
const groupList = ref();
const batchMode = ref(false);
const batchCommandText = ref('');

const rules = reactive({
    name: [Rules.requiredInput],
    command: [Rules.requiredInput],
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

interface DialogProps {
    title: string;
    rowData?: Command.CommandInfo;
    getTableList?: () => Promise<any>;
}
const title = ref<string>('');
const dialogData = ref<DialogProps>({
    title: '',
});

const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    title.value = i18n.global.t('commons.button.' + dialogData.value.title) + i18n.global.t('terminal.quickCommand');
    batchMode.value = false;
    batchCommandText.value = '';
    loadGroups();
    open.value = true;
};
const handleClose = () => {
    open.value = false;
};

const emit = defineEmits<{ (e: 'search'): void }>();

const buildBatchCommands = (): Array<Command.CommandOperate> => {
    const lines = batchCommandText.value
        .split('\n')
        .map((line) => line.trim())
        .filter(Boolean);
    return lines.map((line) => {
        const separatorIndex = line.indexOf('|');
        const hasCustomName = separatorIndex > 0;
        const name = hasCustomName ? line.slice(0, separatorIndex).trim() : line;
        const command = hasCustomName ? line.slice(separatorIndex + 1).trim() : line;
        return {
            id: 0,
            type: 'command',
            groupID: dialogData.value.rowData!.groupID,
            name,
            command,
        };
    });
};

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    if (batchMode.value) {
        if (!dialogData.value.rowData?.groupID) {
            MsgError(i18n.global.t('commons.rule.requiredSelect'));
            return;
        }
        const items = buildBatchCommands().filter((item) => item.name && item.command);
        if (items.length === 0) {
            MsgError(i18n.global.t('commons.rule.requiredInput'));
            return;
        }
        loading.value = true;
        await importCommands(items);
        open.value = false;
        emit('search');
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        return;
    }
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        if (!valid) return;
        if (dialogData.value.title === 'create') {
            await addCommand(dialogData.value.rowData);
        } else {
            await editCommand(dialogData.value.rowData);
        }
        open.value = false;
        emit('search');
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    });
};

const loadGroups = async () => {
    const res = await getGroupList('command');
    groupList.value = res.data;
    if (dialogData.value.title === 'edit') {
        return;
    }
    for (const group of groupList.value) {
        if (group.isDefault) {
            dialogData.value.rowData.groupID = group.id;
            break;
        }
    }
};

defineExpose({
    acceptParams,
});
</script>
