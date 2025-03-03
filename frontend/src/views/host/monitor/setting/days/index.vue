<template>
    <DrawerPro v-model="drawerVisible" :header="$t('monitor.storeDays')" @close="handleClose" size="small">
        <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
            <el-form-item :label="$t('monitor.storeDays')" :rules="[Rules.integerNumber]" prop="monitorStoreDays">
                <el-input clearable v-model.number="form.monitorStoreDays" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button :disabled="loading" type="primary" @click="onSave(formRef)">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { Rules } from '@/global/form-rules';
import { updateMonitorSetting } from '@/api/modules/host';

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    monitorStoreDays: number;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    monitorStoreDays: 30,
});

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.monitorStoreDays = params.monitorStoreDays;
    drawerVisible.value = true;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        await updateMonitorSetting('MonitorStoreDays', form.monitorStoreDays + '')
            .then(() => {
                loading.value = false;
                handleClose();
                emit('search');
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
