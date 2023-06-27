<template>
    <div>
        <el-drawer
            v-model="drawerVisiable"
            :destroy-on-close="true"
            @close="handleClose"
            :close-on-click-modal="false"
            size="30%"
        >
            <template #header>
                <DrawerHeader :header="$t('monitor.storeDays')" :back="handleClose" />
            </template>
            <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item
                            :label="$t('monitor.storeDays')"
                            :rules="[Rules.integerNumber, checkNumberRange(1, 30)]"
                            prop="monitorStoreDays"
                        >
                            <el-input clearable v-model.number="form.monitorStoreDays" />
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onSave(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { Rules, checkNumberRange } from '@/global/form-rules';
import { updateSetting } from '@/api/modules/setting';
import DrawerHeader from '@/components/drawer-header/index.vue';

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    monitorStoreDays: number;
}
const drawerVisiable = ref();
const loading = ref();

const form = reactive({
    monitorStoreDays: 30,
});

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.monitorStoreDays = params.monitorStoreDays;
    drawerVisiable.value = true;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        await updateSetting({ key: 'MonitorStoreDays', value: form.monitorStoreDays + '' })
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
    drawerVisiable.value = false;
};

defineExpose({
    acceptParams,
});
</script>
