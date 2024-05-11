<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            @close="handleClose"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="30%"
        >
            <template #header>
                <DrawerHeader :header="$t('ssh.permitRootLogin')" :back="handleClose" />
            </template>
            <el-form ref="formRef" label-position="top" :model="form" @submit.prevent v-loading="loading">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('ssh.permitRootLogin')" prop="permitRootLogin">
                            <el-select v-model="form.permitRootLogin" style="width: 100%">
                                <el-option :label="$t('ssh.rootHelper1')" value="yes" />
                                <el-option :label="$t('ssh.rootHelper2')" value="no" />
                                <el-option :label="$t('ssh.rootHelper3')" value="without-password" />
                                <el-option :label="$t('ssh.rootHelper4')" value="forced-commands-only" />
                            </el-select>
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
        </el-drawer>
    </div>
</template>
<script lang="ts" setup>
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox, FormInstance } from 'element-plus';
import { updateSSH } from '@/api/modules/host';
import DrawerHeader from '@/components/drawer-header/index.vue';

const emit = defineEmits<{ (e: 'search'): void }>();

interface DialogProps {
    permitRootLogin: string;
}
const drawerVisible = ref();
const loading = ref();

const form = reactive({
    permitRootLogin: 'yes',
});

const formRef = ref<FormInstance>();

const acceptParams = (params: DialogProps): void => {
    form.permitRootLogin = params.permitRootLogin;
    drawerVisible.value = true;
};

const onSave = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        ElMessageBox.confirm(
            i18n.global.t('ssh.sshChangeHelper', [
                i18n.global.t('ssh.permitRootLogin'),
                loadPermitLabel(form.permitRootLogin),
            ]),
            i18n.global.t('ssh.sshChange'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                type: 'info',
            },
        )
            .then(async () => {
                let params = {
                    key: 'PermitRootLogin',
                    oldValue: '',
                    newValue: form.permitRootLogin,
                };
                loading.value = true;
                await updateSSH(params)
                    .then(() => {
                        loading.value = false;
                        handleClose();
                        emit('search');
                        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    })
                    .catch(() => {
                        loading.value = false;
                    });
            })
            .catch(() => {
                emit('search');
            });
    });
};

const loadPermitLabel = (value: string) => {
    switch (value) {
        case 'yes':
            return i18n.global.t('ssh.rootHelper1');
        case 'no':
            return i18n.global.t('ssh.rootHelper2');
        case 'without-password':
            return i18n.global.t('ssh.rootHelper3');
        case 'forced-commands-only':
            return i18n.global.t('ssh.rootHelper4');
    }
};

const handleClose = () => {
    drawerVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
