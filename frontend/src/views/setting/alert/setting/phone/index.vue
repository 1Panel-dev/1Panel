<template>
    <DrawerPro v-model="drawerVisible" :header="$t('xpack.alert.phone')" @close="handleClose" size="736">
        <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
            <el-row type="flex" justify="center">
                <el-col :span="22">
                    <el-form-item :label="$t('xpack.alert.phone')" :rules="[Rules.phone]" prop="phone">
                        <el-input clearable v-model="form.phone" />
                        <span class="input-help">{{ $t('xpack.alert.phoneHelper') }}</span>
                    </el-form-item>
                </el-col>
            </el-row>
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
import { UpdateAlertConfig } from '@/api/modules/alert';
const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    phone: string;
    id: number;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    phone: '',
    id: undefined,
});

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.phone = params.phone;
    form.id = params.id;
    drawerVisible.value = true;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        try {
            const configInfo = { phone: form.phone };
            await UpdateAlertConfig({
                id: form.id,
                type: 'sms',
                title: 'xpack.alert.smsConfig',
                status: 'Enable',
                config: JSON.stringify(configInfo),
            });

            loading.value = false;
            handleClose();
            emit('search');
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        } catch (error) {
            loading.value = false;
        }
    });
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
