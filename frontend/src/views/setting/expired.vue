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
                        <el-form-item :label="$t('setting.oldPassword')" prop="oldPass">
                            <el-input type="password" show-password clearable v-model.trim="passForm.oldPass" />
                        </el-form-item>
                        <el-form-item v-if="!isComplexity" :label="$t('setting.newPassword')" prop="newPass">
                            <el-input type="password" show-password clearable v-model.trim="passForm.newPass" />
                        </el-form-item>
                        <el-form-item v-if="isComplexity" :label="$t('setting.newPassword')" prop="newPassComplexity">
                            <el-input
                                type="password"
                                show-password
                                clearable
                                v-model.trim="passForm.newPassComplexity"
                            />
                        </el-form-item>
                        <el-form-item :label="$t('setting.retryPassword')" prop="rePass">
                            <el-input type="password" show-password clearable v-model.trim="passForm.rePass" />
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

let isComplexity = ref(false);

type FormInstance = InstanceType<typeof ElForm>;
const passFormRef = ref<FormInstance>();
const passRules = reactive({
    oldPass: [Rules.noSpace, Rules.requiredInput],
    newPass: [
        Rules.requiredInput,
        Rules.noSpace,
        { min: 6, message: i18n.global.t('commons.rule.commonPassword'), trigger: 'blur' },
    ],
    newPassComplexity: [Rules.requiredInput, Rules.noSpace, Rules.password],
    rePass: [Rules.requiredInput, Rules.noSpace, { validator: checkPasswordSame, trigger: 'blur' }],
});
const passForm = reactive({
    oldPass: '',
    newPass: '',
    newPassComplexity: '',
    rePass: '',
});

function checkPasswordSame(rule: any, value: any, callback: any) {
    let password = !isComplexity.value ? passForm.newPass : passForm.newPassComplexity;
    if (password !== passForm.rePass) {
        return callback(new Error(i18n.global.t('commons.rule.rePassword')));
    }
    callback();
}

const submitChangePassword = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let password = !isComplexity.value ? passForm.newPass : passForm.newPassComplexity;
        if (password === passForm.oldPass) {
            MsgError(i18n.global.t('setting.duplicatePassword'));
            return;
        }
        await handleExpired({ oldPassword: passForm.oldPass, newPassword: password });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        router.push({ name: 'home' });
    });
};
const search = async () => {
    const res = await getSettingInfo();
    let settingForm = res.data;
    isComplexity.value = settingForm?.complexityVerification === 'enable';
};

onMounted(() => {
    search();
});
</script>
