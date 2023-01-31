<template>
    <div>
        <Submenu activeName="safe" />
        <el-form :model="form" ref="panelFormRef" v-loading="loading" label-position="left" label-width="160px">
            <el-card style="margin-top: 20px">
                <LayoutContent :header="$t('setting.safe')">
                    <el-row>
                        <el-col :span="1"><br /></el-col>
                        <el-col :span="10">
                            <el-form-item :label="$t('setting.passwd')" :rules="Rules.requiredInput" prop="password">
                                <el-input type="password" clearable disabled v-model="form.password">
                                    <template #append>
                                        <el-button icon="Setting" @click="onChangePassword">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item
                                :label="$t('setting.expirationTime')"
                                prop="expirationTime"
                                :rules="Rules.requiredInput"
                            >
                                <el-input disabled v-model="form.expirationTime">
                                    <template #append>
                                        <el-button @click="onChangeExpirationTime" icon="Setting">
                                            {{ $t('commons.button.set') }}
                                        </el-button>
                                    </template>
                                </el-input>
                                <div>
                                    <span class="input-help" v-if="form.expirationTime !== $t('setting.unSetting')">
                                        {{ $t('setting.timeoutHelper', [loadTimeOut()]) }}
                                    </span>
                                    <span class="input-help" v-else>
                                        {{ $t('setting.noneSetting') }}
                                    </span>
                                </div>
                            </el-form-item>
                            <el-form-item
                                :label="$t('setting.complexity')"
                                prop="complexityVerification"
                                :rules="Rules.requiredSelect"
                            >
                                <el-switch
                                    @change="
                                        onSave(panelFormRef, 'ComplexityVerification', form.complexityVerification)
                                    "
                                    v-model="form.complexityVerification"
                                    active-value="enable"
                                    inactive-value="disable"
                                />
                                <span class="input-help">
                                    {{ $t('setting.complexityHelper') }}
                                </span>
                            </el-form-item>
                            <el-form-item
                                :label="$t('setting.mfa')"
                                prop="securityEntrance"
                                :rules="Rules.requiredSelect"
                            >
                                <el-switch
                                    @change="handleMFA"
                                    v-model="form.mfaStatus"
                                    active-value="enable"
                                    inactive-value="disable"
                                />
                                <span class="input-help">
                                    {{ $t('setting.mfaHelper') }}
                                </span>
                            </el-form-item>
                            <el-form-item v-if="isMFAShow">
                                <el-card style="width: 100%">
                                    <ul style="line-height: 24px">
                                        <li>
                                            {{ $t('setting.mfaHelper1') }}
                                            <ul>
                                                <li>Google Authenticator</li>
                                                <li>Microsoft Authenticator</li>
                                                <li>1Password</li>
                                                <li>LastPass</li>
                                                <li>Authenticator</li>
                                            </ul>
                                        </li>
                                        <li>{{ $t('setting.mfaHelper2') }}</li>
                                        <el-image
                                            style="margin-left: 15px; width: 100px; height: 100px"
                                            :src="otp.qrImage"
                                        />
                                        <li>{{ $t('setting.mfaHelper3') }}</li>
                                        <el-input v-model="mfaCode"></el-input>
                                        <div style="margin-top: 10px; margin-bottom: 10px; float: right">
                                            <el-button @click="onCancelMfaBind">
                                                {{ $t('commons.button.cancel') }}
                                            </el-button>
                                            <el-button type="primary" @click="onBind">
                                                {{ $t('commons.button.saveAndEnable') }}
                                            </el-button>
                                        </div>
                                    </ul>
                                </el-card>
                            </el-form-item>
                        </el-col>
                    </el-row>
                </LayoutContent>
            </el-card>
        </el-form>
        <el-dialog
            v-model="timeoutVisiable"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :title="$t('setting.expirationTime')"
            width="30%"
        >
            <el-form ref="timeoutFormRef" label-width="80px" label-position="left" :model="timeoutForm">
                <el-form-item :label="$t('setting.days')" prop="days" :rules="Rules.number">
                    <el-input clearable v-model.number="timeoutForm.days" />
                    <span class="input-help">{{ $t('setting.expirationHelper') }}</span>
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="timeoutVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                    <el-button type="primary" @click="submitTimeout(timeoutFormRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
        <el-dialog
            v-model="passwordVisiable"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :title="$t('setting.changePassword')"
            width="30%"
        >
            <el-form
                v-loading="dialogLoading"
                ref="passFormRef"
                label-width="80px"
                label-position="left"
                :model="passForm"
                :rules="passRules"
            >
                <el-form-item :label="$t('setting.oldPassword')" prop="oldPassword">
                    <el-input type="password" show-password clearable v-model="passForm.oldPassword" />
                </el-form-item>
                <el-form-item
                    v-if="form.complexityVerification === 'disable'"
                    :label="$t('setting.newPassword')"
                    prop="newPassword"
                >
                    <el-input type="password" show-password clearable v-model="passForm.newPassword" />
                </el-form-item>
                <el-form-item
                    v-if="form.complexityVerification === 'enable'"
                    :label="$t('setting.newPassword')"
                    prop="newPasswordComplexity"
                >
                    <el-input type="password" show-password clearable v-model="passForm.newPasswordComplexity" />
                </el-form-item>
                <el-form-item :label="$t('setting.retryPassword')" prop="retryPassword">
                    <el-input type="password" show-password clearable v-model="passForm.retryPassword" />
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button :disabled="dialogLoading" @click="passwordVisiable = false">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button :disabled="dialogLoading" type="primary" @click="submitChangePassword(passFormRef)">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
    </div>
</template>

<script lang="ts" setup>
import { ref, reactive, onMounted } from 'vue';
import { ElMessage, ElForm } from 'element-plus';
import Submenu from '@/views/setting/index.vue';
import { Setting } from '@/api/interface/setting';
import LayoutContent from '@/layout/layout-content.vue';
import { updatePassword, updateSetting, getMFA, bindMFA, getSettingInfo } from '@/api/modules/setting';
import i18n from '@/lang';
import { Rules } from '@/global/form-rules';
import { dateFormatSimple } from '@/utils/util';
import { GlobalStore } from '@/store';
import router from '@/routers';

const loading = ref(false);
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
const dialogLoading = ref(false);
const passwordVisiable = ref<boolean>(false);
const passForm = reactive({
    oldPassword: '',
    newPassword: '',
    newPasswordComplexity: '',
    retryPassword: '',
});

const form = reactive({
    password: '',
    serverPort: '',
    securityEntrance: '',
    expirationDays: 0,
    expirationTime: '',
    complexityVerification: '',
    mfaStatus: 'disable',
    mfaSecret: 'disable',
});
type FormInstance = InstanceType<typeof ElForm>;
const timeoutFormRef = ref<FormInstance>();
const timeoutVisiable = ref<boolean>(false);
const timeoutForm = reactive({
    days: 0,
});

const search = async () => {
    const res = await getSettingInfo();
    form.password = '******';
    form.securityEntrance = res.data.securityEntrance;
    form.expirationDays = Number(res.data.expirationDays);
    form.expirationTime = res.data.expirationTime;
    form.complexityVerification = res.data.complexityVerification;
    form.mfaStatus = res.data.mfaStatus;
    form.mfaSecret = res.data.mfaSecret;
};

const isMFAShow = ref<boolean>(false);
const otp = reactive<Setting.MFAInfo>({
    secret: '',
    qrImage: '',
});
const mfaCode = ref();
const panelFormRef = ref<FormInstance>();

const onSave = async (formEl: FormInstance | undefined, key: string, val: any) => {
    if (!formEl) return;
    const result = await formEl.validateField(key.replace(key[0], key[0].toLowerCase()), callback);
    if (!result) {
        return;
    }
    if (val === '') {
        return;
    }
    let param = {
        key: key,
        value: val + '',
    };
    loading.value = true;
    await updateSetting(param)
        .then(() => {
            loading.value = false;
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};
function callback(error: any) {
    if (error) {
        return error.message;
    } else {
        return;
    }
}

function checkPassword(rule: any, value: any, callback: any) {
    let password = form.complexityVerification === 'disable' ? passForm.newPassword : passForm.newPasswordComplexity;
    if (password !== passForm.retryPassword) {
        return callback(new Error(i18n.global.t('commons.rule.rePassword')));
    }
    callback();
}

const onChangePassword = async () => {
    passForm.oldPassword = '';
    passForm.newPassword = '';
    passForm.newPasswordComplexity = '';
    passForm.retryPassword = '';
    passwordVisiable.value = true;
};

const submitChangePassword = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let password =
            form.complexityVerification === 'disable' ? passForm.newPassword : passForm.newPasswordComplexity;
        if (password === passForm.oldPassword) {
            ElMessage.error(i18n.global.t('setting.duplicatePassword'));
            return;
        }
        dialogLoading.value = true;
        await updatePassword({ oldPassword: passForm.oldPassword, newPassword: password })
            .then(() => {
                dialogLoading.value = false;
                passwordVisiable.value = false;
                ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
                router.push({ name: 'login', params: { code: '' } });
                globalStore.setLogStatus(false);
            })
            .catch(() => {
                dialogLoading.value = false;
            });
    });
};

const handleMFA = async () => {
    console.log('dawdwda');
    if (form.mfaStatus === 'enable') {
        const res = await getMFA();
        otp.secret = res.data.secret;
        otp.qrImage = res.data.qrImage;
        isMFAShow.value = true;
    } else {
        isMFAShow.value = false;
        loading.value = true;
        await updateSetting({ key: 'MFAStatus', value: 'disable' })
            .then(() => {
                loading.value = false;
                search();
                ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
    }
};

const onBind = async () => {
    loading.value = true;
    await bindMFA({ code: mfaCode.value, secret: otp.secret })
        .then(() => {
            loading.value = false;
            search();
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
            isMFAShow.value = false;
        })
        .catch(() => {
            loading.value = false;
        });
};

const onCancelMfaBind = async () => {
    form.mfaStatus = 'disable';
    isMFAShow.value = false;
};

const onChangeExpirationTime = async () => {
    timeoutForm.days = form.expirationDays;
    timeoutVisiable.value = true;
};

const submitTimeout = async (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        let time = new Date(new Date().getTime() + 3600 * 1000 * 24 * timeoutForm.days);
        loading.value = true;
        await updateSetting({ key: 'ExpirationDays', value: timeoutForm.days + '' })
            .then(() => {
                loading.value = false;
                search();
                loadTimeOut();
                form.expirationTime = dateFormatSimple(time);
                timeoutVisiable.value = false;
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

function loadTimeOut() {
    if (form.expirationDays === 0) {
        form.expirationTime = i18n.global.t('setting.unSetting');
        return i18n.global.t('setting.unSetting');
    }
    let staytimeGap = new Date(form.expirationTime).getTime() - new Date().getTime();
    if (staytimeGap < 0) {
        form.expirationTime = i18n.global.t('setting.unSetting');
        return i18n.global.t('setting.unSetting');
    }
    return Math.floor(staytimeGap / (3600 * 1000 * 24));
}

onMounted(() => {
    search();
});
</script>
