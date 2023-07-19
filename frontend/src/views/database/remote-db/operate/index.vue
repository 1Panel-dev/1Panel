<template>
    <el-drawer v-model="drawerVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="50%">
        <template #header>
            <DrawerHeader :header="title" :resource="dialogData.rowData?.name" :back="handleClose" />
        </template>
        <el-form ref="formRef" label-position="top" :model="dialogData.rowData" :rules="rules">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('cronjob.taskName')" prop="name">
                        <el-input
                            :disabled="dialogData.title === 'edit'"
                            clearable
                            v-model.trim="dialogData.rowData!.name"
                        />
                    </el-form-item>
                    <el-form-item :label="$t('cronjob.taskType')" prop="type">
                        <el-select v-model="dialogData.rowData!.type">
                            <el-option value="Mysql" label="Mysql" />
                            <el-option value="Redis" label="Redis" />
                        </el-select>
                    </el-form-item>

                    <el-form-item :label="$t('cronjob.taskName')" prop="address">
                        <el-input clearable v-model.trim="dialogData.rowData!.address" />
                    </el-form-item>
                    <el-form-item :label="$t('cronjob.taskName')" prop="port">
                        <el-input clearable v-model.number="dialogData.rowData!.port" />
                    </el-form-item>
                    <el-form-item :label="$t('cronjob.taskName')" prop="username">
                        <el-input clearable v-model.trim="dialogData.rowData!.username" />
                    </el-form-item>
                    <el-form-item :label="$t('cronjob.taskName')" prop="password">
                        <el-input clearable v-model.trim="dialogData.rowData!.password" />
                    </el-form-item>
                    <el-form-item :label="$t('cronjob.description')" prop="description">
                        <el-input clearable v-model.trim="dialogData.rowData!.description" />
                    </el-form-item>
                </el-col>
            </el-row>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" @click="onSubmit(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { ElForm } from 'element-plus';
import { Cronjob } from '@/api/interface/cronjob';
import { addCronjob, editCronjob } from '@/api/modules/cronjob';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';

interface DialogProps {
    title: string;
    rowData?: Cronjob.CronjobInfo;
    getTableList?: () => Promise<any>;
}
const title = ref<string>('');
const drawerVisiable = ref(false);
const dialogData = ref<DialogProps>({
    title: '',
});
const acceptParams = (params: DialogProps): void => {
    dialogData.value = params;
    title.value = i18n.global.t('cronjob.' + dialogData.value.title);
    drawerVisiable.value = true;
};
const emit = defineEmits<{ (e: 'search'): void }>();

const handleClose = () => {
    drawerVisiable.value = false;
};

const rules = reactive({
    name: [Rules.requiredInput],
    type: [Rules.requiredSelect],
    address: [Rules.requiredInput],
    port: [Rules.port],
});

type FormInstance = InstanceType<typeof ElForm>;
const formRef = ref<FormInstance>();

const onSubmit = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (dialogData.value.title === 'create') {
            await addCronjob(dialogData.value.rowData);
        }
        if (dialogData.value.title === 'edit') {
            await editCronjob(dialogData.value.rowData);
        }

        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        emit('search');
        drawerVisiable.value = false;
    });
};

defineExpose({
    acceptParams,
});
</script>
