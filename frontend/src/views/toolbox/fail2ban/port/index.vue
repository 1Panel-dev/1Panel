<template>
    <DrawerPro v-model="drawerVisible" :header="$t('toolbox.fail2ban.sshPort')" @close="handleClose" size="small">
        <el-form ref="formRef" label-position="top" :rules="rules" :model="form" @submit.prevent v-loading="loading">
            <el-form-item :label="$t('toolbox.fail2ban.sshPort')" prop="port">
                <el-input type="number" clearable v-model.number="form.port" />
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
import { ElMessageBox, FormInstance } from 'element-plus';
import { Rules, checkNumberRange } from '@/global/form-rules';
import { updateFail2ban } from '@/api/modules/toolbox';

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    port: string;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    port: 22,
});

const rules = reactive({
    port: [Rules.integerNumber, checkNumberRange(1, 65535)],
});

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.port = Number(params.port);
    drawerVisible.value = true;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(
            i18n.global.t('ssh.sshChangeHelper', [i18n.global.t('toolbox.fail2ban.sshPort'), form.port]),
            i18n.global.t('toolbox.fail2ban.fail2banChange'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                type: 'info',
            },
        ).then(async () => {
            await updateFail2ban({ key: 'port', value: form.port + '' })
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
