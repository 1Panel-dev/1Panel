<template>
    <DrawerPro v-model="drawerVisible" :header="$t('toolbox.device.hostname')" @close="handleClose" size="small">
        <el-alert :title="$t('toolbox.device.hostnameHelper')" class="common-prompt" :closable="false" type="warning" />
        <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
            <el-form-item :label="$t('toolbox.device.hostname')" prop="hostname" :rules="Rules.requiredInput">
                <el-input clearable v-model.trim="form.hostname" />
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button :disabled="loading" type="primary" @click="onSaveHostame(formRef)">
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
import { updateDevice } from '@/api/modules/toolbox';

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    hostname: string;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    hostname: '',
});

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.hostname = params.hostname;
    drawerVisible.value = true;
};

const onSaveHostame = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(
            i18n.global.t('ssh.sshChangeHelper', [i18n.global.t('toolbox.device.hostname'), form.hostname]),
            i18n.global.t('database.confChange'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                type: 'info',
            },
        ).then(async () => {
            loading.value = true;
            await updateDevice('Hostname', form.hostname)
                .then(async () => {
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    loading.value = false;
                    drawerVisible.value = false;
                    emit('search');
                })
                .catch(() => {
                    loading.value = false;
                });
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
