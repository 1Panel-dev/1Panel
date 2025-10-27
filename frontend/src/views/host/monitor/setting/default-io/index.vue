<template>
    <DrawerPro v-model="drawerVisible" :header="$t('monitor.defaultIO')" @close="handleClose" size="small">
        <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
            <el-form-item :label="$t('monitor.defaultIO')" prop="defaultIO" :rules="Rules.requiredSelect">
                <el-select v-model="form.defaultIO" filterable>
                    <el-option
                        v-for="item in ioOptions"
                        :key="item"
                        :label="item == 'all' ? $t('commons.table.all') : item"
                        :value="item"
                    />
                </el-select>
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
import { getIOOptions, updateMonitorSetting } from '@/api/modules/host';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    defaultIO: string;
}
const drawerVisible = ref();
const loading = ref();
const ioOptions = ref();

const form = reactive({
    defaultIO: '',
});

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.defaultIO = params.defaultIO;
    loadIOOptions();
    drawerVisible.value = true;
};

const loadIOOptions = async () => {
    const res = await getIOOptions();
    ioOptions.value = res.data;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        await updateMonitorSetting('DefaultIO', form.defaultIO)
            .then(async () => {
                globalStore.setDefaultIO(form.defaultIO);
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                loading.value = false;
                drawerVisible.value = false;
                emit('search');
                return;
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
