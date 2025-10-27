<template>
    <DrawerPro v-model="drawerVisible" :header="$t('monitor.interval')" @close="handleClose" size="small">
        <el-form ref="formRef" label-position="top" :model="form" :rules="rules" @submit.prevent v-loading="loading">
            <el-form-item :label="$t('monitor.interval')" prop="monitorInterval">
                <el-input type="number" @input="loadInterval()" class="selectClass" v-model.number="form.timeItem">
                    <template #append>
                        <el-select v-model="form.timeUnit" @change="loadInterval" style="width: 80px">
                            <el-option :label="$t('commons.units.second')" value="s" />
                            <el-option :label="$t('commons.units.minute')" value="m" />
                            <el-option :label="$t('commons.units.hour')" value="h" />
                        </el-select>
                    </template>
                </el-input>
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
import { transferTimeToSecond } from '@/utils/util';

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    timeItem: number;
    timeUnit: string;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    timeItem: 5,
    timeUnit: 'h',
    monitorInterval: 300,
});

const verifyInterval = (rule: any, value: any, callback: any) => {
    if (value < 10 || value > 43200) {
        callback(new Error(i18n.global.t('monitor.intervalHelper')));
        return;
    }
    callback();
};
const rules = reactive({
    monitorInterval: [Rules.integerNumber, { validator: verifyInterval, trigger: 'blur', required: true }],
});

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.timeItem = params.timeItem;
    form.timeUnit = params.timeUnit;
    drawerVisible.value = true;
};

const loadInterval = () => {
    form.monitorInterval = transferTimeToSecond(form.timeItem + form.timeUnit);
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        await updateMonitorSetting('MonitorInterval', form.monitorInterval + '')
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
