<template>
    <DrawerPro v-model="drawerVisible" :header="$t('toolbox.fail2ban.mode')" @close="handleClose" size="small">
        <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
            <el-form-item :label="$t('toolbox.fail2ban.mode')" prop="mode" :rules="Rules.requiredSelect">
                <el-select v-model="form.mode">
                    <el-option
                        v-for="mode in modes"
                        :key="mode.value"
                        :label="mode.value"
                        :value="mode.value"
                        class="mode-option"
                    >
                        <span>{{ mode.value }}</span>
                        <span class="input-help">{{ $t(mode.description) }}</span>
                    </el-option>
                </el-select>
            </el-form-item>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                <el-button v-permission v-node-admin :disabled="loading" type="primary" @click="onSave(formRef)">
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
import { updateFail2ban } from '@/api/modules/toolbox';

const modes = [
    { value: 'normal', description: 'toolbox.fail2ban.modeNormal' },
    { value: 'ddos', description: 'toolbox.fail2ban.modeDdos' },
    { value: 'extra', description: 'toolbox.fail2ban.modeExtra' },
    { value: 'aggressive', description: 'toolbox.fail2ban.modeAggressive' },
];

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    mode: string;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    mode: '',
});

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.mode = params.mode;
    drawerVisible.value = true;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(
            i18n.global.t('ssh.sshChangeHelper', [i18n.global.t('toolbox.fail2ban.mode'), form.mode]),
            i18n.global.t('toolbox.fail2ban.fail2banChange'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                type: 'info',
            },
        ).then(async () => {
            loading.value = true;
            await updateFail2ban({ key: 'mode', value: form.mode })
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

<style lang="scss" scoped>
.mode-option {
    height: auto;
    max-width: 420px;
    padding-top: 8px;
    padding-bottom: 8px;
    line-height: 20px;
    white-space: normal;

    span {
        display: block;
    }
}
</style>
