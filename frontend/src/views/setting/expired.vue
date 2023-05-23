<template>
    <div>
        <el-card style="margin-top: 20px">
            <template #header>
                <div class="card-header">
                    <span style="font-size: 14px; font-weight: 500">{{ $t('setting.expiredHelper') }}</span>
                </div>
            </template>
            <el-row>
                <el-col :span="1"><br /></el-col>
                <el-col :span="10">
                    <el-form
                        :model="passForm"
                        ref="passFormRef"
                        :rules="passRules"
                        label-position="left"
                        label-width="160px"
                    >
                        <el-form-item :label="$t('setting.oldPassword')" prop="oldPassword">
                            <el-input type="password" show-password clearable v-model="passForm.oldPassword" />
                        </el-form-item>
                        <el-form-item
                            v-if="settingForm?.complexityVerification === 'disable'"
                            :label="$t('setting.newPassword')"
                            prop="newPassword"
                        >
                            <el-input type="password" show-password clearable v-model="passForm.newPassword" />
                        </el-form-item>
                        <el-form-item
                            v-if="settingForm?.complexityVerification === 'enable'"
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
                        <el-form-item>
                            <el-button type="primary" @click="submitChangePassword(passFormRef)">
                                {{ $t('commons.button.confirm') }}
                            </el-button>
                        </el-form-item>
                    </el-form>
                </el-col>
            </el-row>
        </el-card>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue';
import { getSettingInfo, handleExpired } from '@/api/modules/setting';
import { ElForm } from 'element-plus';
import i18n from '@/lang';
import { Rules } from '@/global/form-rules';
import router from '@/routers';
import { MsgError, MsgSuccess } from '@/utils/message';
let settingForm = ref();

type FormInstance = InstanceType<typeof ElForm>;
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
const passForm = reactive({
    oldPassword: '',
    newPassword: '',
    newPasswordComplexity: '',
    retryPassword: '',
});

function checkPassword(rule: any, value: any, callback: any) {
    let password =
        settingForm.value.complexityVerification === 'disable' ? passForm.newPassword : passForm.newPasswordComplexity;
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
            settingForm.value.complexityVerification === 'disable'
                ? passForm.newPassword
                : passForm.newPasswordComplexity;
        if (password === passForm.oldPassword) {
            MsgError(i18n.global.t('setting.duplicatePassword'));
            return;
        }
        await handleExpired({ oldPassword: passForm.oldPassword, newPassword: password });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        router.push({ name: 'home' });
    });
};
const search = async () => {
    const res = await getSettingInfo();
    settingForm.value = res.data;
};

onMounted(() => {
    search();
});
</script>
