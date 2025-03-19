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
            <el-form-item :label="$t('commons.table.name')" prop="name">
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
            <el-form-item :label="$t('terminal.command')" prop="command">
                <el-input type="textarea" clearable v-model="dialogData.rowData!.command" />
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
import { MsgSuccess } from '@/utils/message';
import { Command } from '@/api/interface/command';
import { addCommand, editCommand } from '@/api/modules/command';
import { getGroupList } from '@/api/modules/group';

const loading = ref();
const open = ref();
const groupList = ref();

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
    loadGroups();
    open.value = true;
};
const handleClose = () => {
    open.value = false;
};

const emit = defineEmits<{ (e: 'search'): void }>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
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
