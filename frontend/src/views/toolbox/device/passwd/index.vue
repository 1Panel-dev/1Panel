<template>
    <div>
        <el-drawer
            v-model="passwordVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="30%"
        >
            <template #header>
                <DrawerHeader :header="$t('setting.changePassword')" :back="handleClose" />
            </template>

            <el-row type="flex" justify="center" v-loading="loading">
                <el-col :span="22">
                    <el-form ref="formRef" label-position="top" :model="form" :rules="passRules">
                        <el-alert
                            :title="$t('toolbox.device.passwordHelper')"
                            class="common-prompt"
                            :closable="false"
                            type="warning"
                        />
                        <el-form-item :label="$t('setting.user')" prop="user">
                            <el-input clearable v-model.trim="form.user" />
                            <span class="input-help">{{ $t('toolbox.device.userHelper') }}</span>
                        </el-form-item>
                        <el-form-item :label="$t('setting.newPassword')" prop="newPassword">
                            <el-input type="password" show-password clearable v-model.trim="form.newPassword" />
                        </el-form-item>
                        <el-form-item :label="$t('setting.retryPassword')" prop="retryPassword">
                            <el-input type="password" show-password clearable v-model.trim="form.retryPassword" />
                        </el-form-item>
                    </el-form>
                </el-col>
            </el-row>
            <template #footer>
                <span class="dialog-footer">
                    <el-button :disabled="loading" @click="passwordVisible = false">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button :disabled="loading" type="primary" @click="submitChangePassword(formRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>

<script lang="ts" setup>
import { Rules } from '@/global/form-rules';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { reactive, ref } from 'vue';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { updateDevicePasswd } from '@/api/modules/toolbox';

const formRef = ref<FormInstance>();
const passRules = reactive({
    user: Rules.requiredInput,
    newPassword: [Rules.requiredInput, Rules.noSpace, { validator: checkPassword, trigger: 'blur' }],
    retryPassword: [Rules.requiredInput, Rules.noSpace, { validator: checkRePassword, trigger: 'blur' }],
});

const loading = ref(false);
const passwordVisible = ref<boolean>(false);
const form = reactive({
    user: '',
    newPassword: '',
    retryPassword: '',
});

interface DialogProps {
    user: string;
}
const acceptParams = (params: DialogProps): void => {
    form.user = params.user;
    form.newPassword = '';
    form.retryPassword = '';
    passwordVisible.value = true;
};

function checkPassword(rule: any, value: any, callback: any) {
    if (form.newPassword.indexOf('&') !== -1 || form.newPassword.indexOf('$') !== -1) {
        return callback(new Error(i18n.global.t('toolbox.device.passwdHelper')));
    }
    callback();
}

function checkRePassword(rule: any, value: any, callback: any) {
    if (form.newPassword !== form.retryPassword) {
        return callback(new Error(i18n.global.t('commons.rule.rePassword')));
    }
    callback();
}

const submitChangePassword = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        await updateDevicePasswd(form.user, form.newPassword)
            .then(async () => {
                loading.value = false;
                passwordVisible.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    });
};
const handleClose = () => {
    passwordVisible.value = false;
};

defineExpose({
    acceptParams,
});
</script>
