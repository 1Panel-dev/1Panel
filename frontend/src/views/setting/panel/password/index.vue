<template>
    <div v-loading="loading">
        <el-drawer v-model="passwordVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="30%">
            <template #header>
                <DrawerHeader :header="$t('setting.changePassword')" :back="handleClose" />
            </template>
            <el-form ref="passFormRef" label-position="top" :model="passForm" :rules="passRules">
                <el-row type="flex" justify="center">
                    <el-col :span="22">
                        <el-form-item :label="$t('setting.oldPassword')" prop="oldPassword">
                            <el-input type="password" show-password clearable v-model="passForm.oldPassword" />
                        </el-form-item>
                        <el-form-item
                            v-if="complexityVerification === 'disable'"
                            :label="$t('setting.newPassword')"
                            prop="newPassword"
                        >
                            <el-input type="password" show-password clearable v-model="passForm.newPassword" />
                        </el-form-item>
                        <el-form-item
                            v-if="complexityVerification === 'enable'"
                            :label="$t('setting.newPassword')"
                            prop="newPasswordComplexity"
                        >
                            <el-input
                                type="password"
                                show-password
                                clearable
                                v-model="passForm.newPasswordComplexity"
                            />
                        </el-form-item>
                        <el-form-item :label="$t('setting.retryPassword')" prop="retryPassword">
                            <el-input type="password" show-password clearable v-model="passForm.retryPassword" />
                        </el-form-item>
                    </el-col>
                </el-row>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button :disabled="loading" @click="passwordVisiable = false">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button :disabled="loading" type="primary" @click="submitChangePassword(passFormRef)">
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
import router from '@/routers';
import { MsgError, MsgSuccess } from '@/utils/message';
import { FormInstance } from 'element-plus';
import { GlobalStore } from '@/store';
import { reactive, ref } from 'vue';
import { updatePassword } from '@/api/modules/setting';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { logOutApi } from '@/api/modules/auth';

const globalStore = GlobalStore();
const passFormRef = ref<FormInstance>();
const passRules = reactive({
    oldPassword: [Rules.requiredInput],
    newPassword: [
        Rules.requiredInput,
        { min: 6, message: i18n.global.t('commons.rule.commonPassword'), trigger: 'blur' },
    ],
    newPasswordComplexity: [Rules.requiredInput, Rules.password],
    retryPassword: [Rules.requiredInput, { validator: checkPassword, trigger: 'blur' }],
});
const loading = ref(false);
const passwordVisiable = ref<boolean>(false);
const passForm = reactive({
    oldPassword: '',
    newPassword: '',
    newPasswordComplexity: '',
    retryPassword: '',
});
const complexityVerification = ref();

interface DialogProps {
    complexityVerification: string;
}
const acceptParams = (params: DialogProps): void => {
    complexityVerification.value = params.complexityVerification;
    passForm.oldPassword = '';
    passForm.newPassword = '';
    passForm.newPasswordComplexity = '';
    passForm.retryPassword = '';
    passwordVisiable.value = true;
};

function checkPassword(rule: any, value: any, callback: any) {
    let password = complexityVerification.value === 'disable' ? passForm.newPassword : passForm.newPasswordComplexity;
    if (password !== passForm.retryPassword) {
        return callback(new Error(i18n.global.t('commons.rule.rePassword')));
    }
    callback();
}

const submitChangePassword = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let password =
            complexityVerification.value === 'disable' ? passForm.newPassword : passForm.newPasswordComplexity;
        if (password === passForm.oldPassword) {
            MsgError(i18n.global.t('setting.duplicatePassword'));
            return;
        }
        loading.value = true;
        await updatePassword({ oldPassword: passForm.oldPassword, newPassword: password })
            .then(async () => {
                loading.value = false;
                passwordVisiable.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                await logOutApi();
                router.push({ name: 'entrance', params: { code: globalStore.entrance } });
                globalStore.setLogStatus(false);
            })
            .catch(() => {
                loading.value = false;
            });
    });
};
const handleClose = () => {
    passwordVisiable.value = false;
};

defineExpose({
    acceptParams,
});
</script>
