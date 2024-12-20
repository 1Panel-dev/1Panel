<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="30%"
        >
            <template #header>
                <DrawerHeader :header="$t('toolbox.device.hostname')" :back="handleClose" />
            </template>

            <el-row type="flex" justify="center" v-loading="loading">
                <el-col :span="22">
                    <el-alert
                        :title="$t('toolbox.device.hostnameHelper')"
                        class="common-prompt"
                        :closable="false"
                        type="warning"
                    />
                    <el-form ref="formRef" label-position="top" :model="form" @submit.prevent>
                        <el-form-item
                            :label="$t('toolbox.device.hostname')"
                            prop="hostname"
                            :rules="Rules.requiredInput"
                        >
                            <el-input clearable v-model.trim="form.hostname" />
                        </el-form-item>
                    </el-form>
                </el-col>
            </el-row>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="drawerVisible = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button :disabled="loading" type="primary" @click="onSaveHostname(formRef)">
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
import { Rules } from '@/global/form-rules';
import DrawerHeader from '@/components/drawer-header/index.vue';
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

const onSaveHostname = async (formEl: FormInstance | undefined) => {
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
