<template>
    <DrawerPro v-model="drawerVisible" :header="$t('setting.sessionTimeout')" @close="handleClose" size="small">
        <el-form ref="formRef" label-position="top" :rules="rules" :model="form" @submit.prevent v-loading="loading">
            <el-form-item :label="$t('setting.sessionTimeout')" prop="sessionTimeout">
                <el-input clearable v-model.number="form.sessionTimeout" />
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button :disabled="loading" type="primary" @click="onSaveTimeout(formRef)">
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DrawerPro>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { Rules, checkNumberRange } from '@/global/form-rules';
import { updateSetting } from '@/api/modules/setting';

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    sessionTimeout: number;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    sessionTimeout: 86400,
});

const rules = reactive({
    sessionTimeout: [Rules.integerNumber, checkNumberRange(300, 864000)],
});

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.sessionTimeout = params.sessionTimeout;
    drawerVisible.value = true;
};

const onSaveTimeout = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        await updateSetting({ key: 'SessionTimeout', value: form.sessionTimeout + '' })
            .then(async () => {
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
