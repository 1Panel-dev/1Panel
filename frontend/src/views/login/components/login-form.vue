<template>
    <div v-loading="loading">
        <div v-if="mfaShow">
            <div class="login-form">
                <el-form @submit.prevent>
                    <div class="login-title">{{ $t('commons.login.mfaTitle') }}</div>
                    <el-form-item class="no-border">
                        <el-input
                            size="default"
                            :placeholder="$t('commons.login.mfaCode')"
                            v-model.trim="mfaLoginForm.code"
                            @input="mfaLogin(true)"
                        >
                            <template #prefix>
                                <el-icon class="el-input__icon">
                                    <Finished />
                                </el-icon>
                            </template>
                        </el-input>
                        <span v-if="errMfaInfo" class="input-error" style="line-height: 14px">
                            {{ $t('commons.login.errorMfaInfo') }}
                        </span>
                    </el-form-item>
                    <el-form-item>
                        <el-button
                            @focus="mfaButtonFocused = true"
                            @blur="mfaButtonFocused = false"
                            class="login-button"
                            type="primary"
                            size="default"
                            round
                            @click="mfaLogin(false)"
                        >
                            {{ $t('commons.button.verify') }}
                        </el-button>
                    </el-form-item>
                </el-form>
            </div>
        </div>
        <div v-else>
            <div class="login-form">
                <el-form ref="loginFormRef" :model="loginForm" size="default" :rules="loginRules">
                    <div class="login-form-header">
                        <div class="title">{{ $t('commons.button.login') }}</div>
                        <div>
                            <el-dropdown @command="handleCommand">
                                <span>
                                    {{ dropdownText }}
                                    <el-icon>
                                        <arrow-down />
                                    </el-icon>
                                </span>
                                <template #dropdown>
                                    <el-dropdown-menu>
                                        <el-dropdown-item command="zh">中文(简体)</el-dropdown-item>
                                        <el-dropdown-item command="tw">中文(繁體)</el-dropdown-item>
                                        <el-dropdown-item command="en">English</el-dropdown-item>
                                    </el-dropdown-menu>
                                </template>
                            </el-dropdown>
                        </div>
                    </div>
                    <el-form-item prop="name" class="no-border">
                        <el-input
                            v-model.trim="loginForm.name"
                            :placeholder="$t('commons.login.username')"
                            class="form-input"
                        >
                            <template #prefix>
                                <el-icon class="el-input__icon">
                                    <user />
                                </el-icon>
                            </template>
                        </el-input>
                    </el-form-item>
                    <el-form-item prop="password" class="no-border">
                        <el-input
                            type="password"
                            clearable
                            v-model.trim="loginForm.password"
                            show-password
                            :placeholder="$t('commons.login.password')"
                        >
                            <template #prefix>
                                <el-icon class="el-input__icon">
                                    <lock />
                                </el-icon>
                            </template>
                        </el-input>
                        <span v-if="errAuthInfo" class="input-error" style="line-height: 14px">
                            {{ $t('commons.login.errorAuthInfo') }}
                        </span>
                    </el-form-item>
                    <el-form-item v-if="!globalStore.ignoreCaptcha" prop="captcha" class="login-captcha">
                        <el-input v-model.trim="loginForm.captcha" :placeholder="$t('commons.login.captchaHelper')">
                            <template #prefix>
                                <svg-icon style="font-size: 7px" iconName="p-yanzhengma1"></svg-icon>
                            </template>
                        </el-input>
                        <img
                            v-if="captcha.imagePath"
                            :src="captcha.imagePath"
                            :alt="$t('commons.login.captchaHelper')"
                            @click="loginVerify()"
                        />
                        <span v-if="errCaptcha" class="input-error" style="line-height: 14px">
                            {{ $t('commons.login.errorCaptcha') }}
                        </span>
                    </el-form-item>
                    <el-form-item>
                        <el-button
                            @click="login(loginFormRef)"
                            @focus="loginButtonFocused = true"
                            @blur="loginButtonFocused = false"
                            class="login-button"
                            type="primary"
                            size="default"
                            round
                        >
                            {{ $t('commons.button.login') }}
                        </el-button>
                    </el-form-item>
                    <el-form-item prop="agreeLicense">
                        <el-checkbox v-model="loginForm.agreeLicense">
                            <template #default>
                                <span class="agree" v-html="$t('commons.login.licenseHelper')"></span>
                            </template>
                        </el-checkbox>
                    </el-form-item>
                    <div class="agree-helper">
                        <span
                            v-if="!loginForm.agreeLicense && !_isMobile()"
                            class="input-error"
                            style="line-height: 14px"
                        >
                            {{ $t('commons.login.errorAgree') }}
                        </span>
                    </div>
                </el-form>
                <div class="demo">
                    <span v-if="isDemo">
                        {{ $t('commons.login.username') }}:demo {{ $t('commons.login.password') }}:1panel
                    </span>
                </div>
            </div>
        </div>
        <el-dialog v-model="agreeVisible" center :destroy-on-close="true" :close-on-click-modal="false" width="80%">
            <el-row type="flex" justify="center">
                <span class="text-base mb-4">
                    {{ $t('commons.login.agreeTitle') }}
                </span>
            </el-row>
            <div>
                <span v-html="$t('commons.login.agreeContent')"></span>
            </div>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="agreeVisible = false">
                        {{ $t('commons.button.notAgree') }}
                    </el-button>
                    <el-button type="primary" @click="agreeWithLogin()">
                        {{ $t('commons.button.agree') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import type { ElForm } from 'element-plus';
import { loginApi, getCaptcha, mfaLoginApi, checkIsDemo, getLanguage } from '@/api/modules/auth';
import { GlobalStore, MenuStore, TabsStore } from '@/store';
import { MsgSuccess } from '@/utils/message';
import { useI18n } from 'vue-i18n';
import { getSettingInfo } from '@/api/modules/setting';

const i18n = useI18n();
const themeConfig = computed(() => globalStore.themeConfig);
const globalStore = GlobalStore();
const menuStore = MenuStore();
const tabsStore = TabsStore();
const usei18n = useI18n();

const errAuthInfo = ref(false);
const errCaptcha = ref(false);
const errMfaInfo = ref(false);
const isDemo = ref(false);
const agreeVisible = ref(false);

type FormInstance = InstanceType<typeof ElForm>;
const _isMobile = () => {
    const rect = document.body.getBoundingClientRect();
    return rect.width - 1 < 600;
};

const loginButtonFocused = ref();
const loginFormRef = ref<FormInstance>();
const loginForm = reactive({
    name: '',
    password: '',
    ignoreCaptcha: true,
    captcha: '',
    captchaID: '',
    authMethod: 'session',
    agreeLicense: false,
    language: 'zh',
});

const loginRules = reactive({
    name: computed(() => [{ required: true, message: i18n.t('commons.rule.username'), trigger: 'blur' }]),
    password: computed(() => [{ required: true, message: i18n.t('commons.rule.password'), trigger: 'blur' }]),
});

let isLoggingIn = false;
const mfaButtonFocused = ref();
const mfaLoginForm = reactive({
    name: '',
    password: '',
    secret: '',
    code: '',
    authMethod: 'session',
});

const captcha = reactive({
    captchaID: '',
    imagePath: '',
    captchaLength: 0,
});

const loading = ref<boolean>(false);
const mfaShow = ref<boolean>(false);
const router = useRouter();
const dropdownText = ref('中文(简体)');

function handleCommand(command: string) {
    loginForm.language = command;
    usei18n.locale.value = command;
    globalStore.updateLanguage(command);
    if (command === 'zh') {
        dropdownText.value = '中文(简体)';
    } else if (command === 'en') {
        dropdownText.value = 'English';
    } else if (command === 'tw') {
        dropdownText.value = '中文(繁體)';
    }
}

const agreeWithLogin = () => {
    agreeVisible.value = false;
    loginForm.agreeLicense = true;
    login(loginFormRef.value);
};

const login = (formEl: FormInstance | undefined) => {
    if (!formEl || isLoggingIn) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (!loginForm.agreeLicense) {
            if (_isMobile()) {
                agreeVisible.value = true;
            }
            return;
        }
        let requestLoginForm = {
            name: loginForm.name,
            password: loginForm.password,
            ignoreCaptcha: globalStore.ignoreCaptcha,
            captcha: loginForm.captcha,
            captchaID: captcha.captchaID,
            authMethod: 'session',
            language: loginForm.language,
        };
        if (!globalStore.ignoreCaptcha && requestLoginForm.captcha == '') {
            errCaptcha.value = true;
            return;
        }
        try {
            isLoggingIn = true;
            loading.value = true;
            const res = await loginApi(requestLoginForm);
            if (res.code === 406) {
                if (res.message === 'ErrCaptchaCode') {
                    loginForm.captcha = '';
                    errCaptcha.value = true;
                    errAuthInfo.value = false;
                }
                if (res.message === 'ErrAuth') {
                    globalStore.ignoreCaptcha = false;
                    errCaptcha.value = false;
                    errAuthInfo.value = true;
                }
                loginVerify();
                return;
            }
            globalStore.ignoreCaptcha = true;
            if (res.data.mfaStatus === 'enable') {
                mfaShow.value = true;
                errMfaInfo.value = false;
                return;
            }
            globalStore.setLogStatus(true);
            globalStore.setAgreeLicense(true);
            menuStore.setMenuList([]);
            tabsStore.removeAllTabs();
            MsgSuccess(i18n.t('commons.msg.loginSuccess'));
            loadDataFromDB();
            router.push({ name: 'home' });
        } catch (error) {
            loginVerify();
        } finally {
            isLoggingIn = false;
            loading.value = false;
        }
    });
};

const mfaLogin = async (auto: boolean) => {
    if (isLoggingIn) return;
    if ((!auto && mfaLoginForm.code) || (auto && mfaLoginForm.code.length === 6)) {
        isLoggingIn = true;
        mfaLoginForm.name = loginForm.name;
        mfaLoginForm.password = loginForm.password;
        const res = await mfaLoginApi(mfaLoginForm);
        if (res.code === 406) {
            errMfaInfo.value = true;
            isLoggingIn = false;
            return;
        }
        globalStore.setLogStatus(true);
        menuStore.setMenuList([]);
        tabsStore.removeAllTabs();
        MsgSuccess(i18n.t('commons.msg.loginSuccess'));
        loadDataFromDB();
        router.push({ name: 'home' });
    }
};
const loginVerify = async () => {
    const res = await getCaptcha();
    captcha.imagePath = res.data.imagePath ? res.data.imagePath : '';
    captcha.captchaID = res.data.captchaID ? res.data.captchaID : '';
    captcha.captchaLength = res.data.captchaLength ? res.data.captchaLength : 0;
};

const checkIsSystemDemo = async () => {
    const res = await checkIsDemo();
    isDemo.value = res.data;
};

const loadLanguage = async () => {
    try {
        const res = await getLanguage();
        loginForm.language = res.data;
        handleCommand(res.data);
    } catch (error) {}
};

const loadDataFromDB = async () => {
    const res = await getSettingInfo();
    document.title = res.data.panelName;
    i18n.locale.value = res.data.language;
    i18n.warnHtmlMessage = false;
    globalStore.entrance = res.data.securityEntrance;
    globalStore.setOpenMenuTabs(res.data.menuTabs === 'enable');
    globalStore.updateLanguage(res.data.language);
    globalStore.setThemeConfig({ ...themeConfig.value, theme: res.data.theme, panelName: res.data.panelName });
};

onMounted(() => {
    globalStore.isOnRestart = false;
    loginVerify();
    loadLanguage();
    document.title = globalStore.themeConfig.panelName;
    loginForm.agreeLicense = globalStore.agreeLicense;
    checkIsSystemDemo();
    document.onkeydown = (e: any) => {
        e = window.event || e;
        if (e.keyCode === 13) {
            if (!mfaShow.value) {
                if (!loginButtonFocused.value) {
                    login(loginFormRef.value);
                }
            }
            if (mfaShow.value && !mfaButtonFocused.value) {
                mfaLogin(false);
            }
        }
    };
});
</script>

<style scoped lang="scss">
.login-form {
    padding: 0 40px;
    .hide {
        width: 0;
        border: 0;
        position: absolute;
        visibility: hidden;
    }

    .login-title {
        font-size: 30px;
        letter-spacing: 0;
        text-align: center;
        color: #646a73;
        margin-bottom: 30px;
    }
    .no-border {
        :deep(.el-input__wrapper) {
            background: none !important;
            box-shadow: none !important;
            border-radius: 0 !important;
            border-bottom: 1px solid #dcdfe6;
        }
    }

    .el-input {
        height: 44px;
    }

    .login-captcha {
        margin-top: 10px;

        :deep(.el-input__wrapper) {
            background: none !important;
            box-shadow: none !important;
            border-radius: 0 !important;
            border-bottom: 1px solid #dcdfe6;
        }
        .el-input {
            width: 50%;
            height: 44px;
        }
        img {
            width: 45%;
            height: 44px;
            margin-left: 5%;
        }
    }

    .login-msg {
        margin-top: 10px;
        padding: 0 40px;
        text-align: center;
    }

    .login-image {
        width: 480px;
        height: 480px;
        @media only screen and (max-width: 1280px) {
            height: 380px;
        }
    }

    .submit {
        width: 100%;
        border-radius: 0;
    }

    .forget-password {
        margin-top: 40px;
        padding: 0 40px;
        float: right;
        @media only screen and (max-width: 1280px) {
            margin-top: 20px;
        }
    }

    .login-button {
        width: 100%;
        height: 45px;
        margin-top: 10px;
        background-color: #005eeb;
        color: #ffffff;
    }

    .demo {
        text-align: center;
        span {
            color: red;
        }
    }

    .login-form-header {
        display: flex;
        margin-bottom: 30px;
        justify-content: space-between;
        align-items: center;
        .title {
            color: #646a73;
            font-size: 25px;
        }
    }

    .agree {
        white-space: pre-wrap;
        line-height: 14px;
        color: #005eeb;
    }

    :deep(.el-checkbox__input .el-checkbox__inner) {
        background-color: #fff !important;
        border-color: #fff !important;
    }

    :deep(.el-checkbox__input.is-checked .el-checkbox__inner) {
        background-color: #005eeb !important;
        border-color: #005eeb !important;
    }

    :deep(.el-checkbox__input.is-checked .el-checkbox__inner::after) {
        border-color: #fff !important;
    }

    .agree-helper {
        min-height: 20px;
        margin-top: -20px;
        margin-left: 20px;
    }
}
</style>
