<template>
    <div class="w-full h-full flex items-center justify-center px-8 py-6">
        <div v-loading="loading" class="w-full flex-grow flex flex-col">
            <div v-if="mfaShow">
                <el-form @submit.prevent>
                    <div class="flex flex-col justify-center items-center mb-6">
                        <div class="text-2xl font-medium text-gray-900 text-center">
                            {{ $t('commons.login.mfaTitle') }}
                        </div>
                    </div>

                    <div class="space-y-6 flex-grow">
                        <el-form-item>
                            <el-input
                                size="large"
                                :placeholder="$t('commons.login.mfaCode')"
                                v-model.trim="mfaLoginForm.code"
                                @input="mfaLogin(true)"
                            ></el-input>
                            <div class="h-1">
                                <span v-if="errMfaInfo" class="input-error">
                                    {{ $t('commons.login.errorMfaInfo') }}
                                </span>
                            </div>
                        </el-form-item>
                        <el-form-item>
                            <el-button
                                @focus="mfaButtonFocused = true"
                                @blur="mfaButtonFocused = false"
                                class="w-full"
                                type="primary"
                                @click="mfaLogin(false)"
                            >
                                {{ $t('commons.button.verify') }}
                            </el-button>
                        </el-form-item>
                    </div>
                </el-form>
            </div>
            <div v-else>
                <el-form ref="loginFormRef" :model="loginForm" size="default" :rules="loginRules">
                    <div class="flex justify-between items-center mb-6">
                        <div class="text-2xl font-medium text-gray-900">{{ $t('commons.button.login') }}</div>
                        <div class="cursor-pointer">
                            <el-dropdown @command="handleCommand">
                                <span class="flex items-center space-x-1">
                                    {{ dropdownText }}
                                    <el-icon>
                                        <arrow-down />
                                    </el-icon>
                                </span>
                                <template #dropdown>
                                    <el-dropdown-menu>
                                        <el-dropdown-item v-if="globalStore.isIntl" command="en">
                                            English
                                        </el-dropdown-item>
                                        <el-dropdown-item command="zh">中文(简体)</el-dropdown-item>
                                        <el-dropdown-item command="tw">中文(繁體)</el-dropdown-item>
                                        <el-dropdown-item v-if="!globalStore.isIntl" command="en">
                                            English
                                        </el-dropdown-item>
                                        <el-dropdown-item command="ja">日本語</el-dropdown-item>
                                        <el-dropdown-item command="pt-BR">Português (Brasil)</el-dropdown-item>
                                        <el-dropdown-item command="ko">한국어</el-dropdown-item>
                                        <el-dropdown-item command="ru">Русский</el-dropdown-item>
                                        <el-dropdown-item command="ms">Bahasa Melayu</el-dropdown-item>
                                    </el-dropdown-menu>
                                </template>
                            </el-dropdown>
                        </div>
                    </div>
                    <div class="space-y-6 flex-grow">
                        <el-form-item prop="name" class="w-full">
                            <el-input
                                v-model.trim="loginForm.name"
                                :placeholder="$t('commons.login.username')"
                                class="w-full"
                                size="large"
                            ></el-input>
                        </el-form-item>
                        <el-form-item prop="password" class="w-full">
                            <el-input
                                type="password"
                                v-model.trim="loginForm.password"
                                class="w-full"
                                size="large"
                                :placeholder="$t('commons.login.password')"
                            ></el-input>
                        </el-form-item>
                        <el-row :gutter="10">
                            <el-col :span="12" v-if="!globalStore.ignoreCaptcha">
                                <el-form-item prop="captcha">
                                    <el-input
                                        v-model.trim="loginForm.captcha"
                                        size="large"
                                        :placeholder="$t('commons.login.captchaHelper')"
                                    ></el-input>
                                </el-form-item>
                            </el-col>
                            <el-col :span="12" v-if="!globalStore.ignoreCaptcha">
                                <img
                                    class="w-full h-10"
                                    v-if="captcha.imagePath"
                                    :src="captcha.imagePath"
                                    :alt="$t('commons.login.captchaHelper')"
                                    @click="loginVerify()"
                                />
                            </el-col>
                            <el-col :span="24" class="h-1">
                                <span v-show="errCaptcha" class="input-error">
                                    {{ $t('commons.login.errorCaptcha') }}
                                </span>
                                <span v-show="errAuthInfo" class="input-error">
                                    {{ $t('commons.login.errorAuthInfo') }}
                                </span>
                            </el-col>
                        </el-row>

                        <el-form-item>
                            <el-button
                                @click="login(loginFormRef)"
                                @focus="loginButtonFocused = true"
                                @blur="loginButtonFocused = false"
                                class="w-full"
                                type="primary"
                                size="default"
                            >
                                {{ $t('commons.button.login') }}
                            </el-button>
                        </el-form-item>

                        <template v-if="!isIntl">
                            <el-form-item prop="agreeLicense">
                                <el-checkbox v-model="loginForm.agreeLicense">
                                    <template #default>
                                        <span>
                                            {{ $t('commons.button.agree') }}
                                            <a
                                                class="agree"
                                                href="https://www.fit2cloud.com/legal/licenses.html"
                                                target="_blank"
                                            >
                                                {{ $t('commons.login.licenseHelper') }}
                                            </a>
                                        </span>
                                    </template>
                                </el-checkbox>
                            </el-form-item>
                        </template>
                    </div>
                </el-form>
                <div class="demo">
                    <span v-if="isDemo">
                        {{ $t('commons.login.username') }}:demo {{ $t('commons.login.password') }}:1panel
                    </span>
                </div>
            </div>

            <DialogPro v-model="open" center size="w-90">
                <el-row type="flex" justify="center">
                    <span class="text-base mb-4">
                        {{ $t('commons.login.agreeTitle') }}
                    </span>
                </el-row>
                <div>
                    <span v-html="$t('commons.login.agreeContent')"></span>
                </div>
                <template #footer>
                    <span class="dialog-footer login-footer-btn">
                        <el-button @click="open = false">
                            {{ $t('commons.button.notAgree') }}
                        </el-button>
                        <el-button type="primary" @click="agreeWithLogin()">
                            {{ $t('commons.button.agree') }}
                        </el-button>
                    </span>
                </template>
            </DialogPro>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import type { ElForm } from 'element-plus';
import { loginApi, getCaptcha, mfaLoginApi, checkIsDemo, getLanguage, checkIsIntl } from '@/api/modules/auth';
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
const isIntl = ref(true);
const open = ref(false);

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
    name: [{ required: true, validator: checkUsername, trigger: 'blur' }],
    password: [{ required: true, validator: checkPassword, trigger: 'blur' }],
    agreeLicense: [{ required: true, validator: checkAgreeLicense, trigger: 'blur' }],
});

function checkUsername(rule: any, value: any, callback: any) {
    if (value === '') {
        return callback(new Error(i18n.t('commons.rule.username')));
    }
    callback();
}
function checkPassword(rule: any, value: any, callback: any) {
    if (value === '') {
        return callback(new Error(i18n.t('commons.rule.password')));
    }
    callback();
}
function checkAgreeLicense(rule: any, value: any, callback: any) {
    if (!value && !_isMobile()) {
        return callback(new Error(i18n.t('commons.login.errorAgree')));
    }
    callback();
}

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

const checkIsSystemIntl = async () => {
    const res = await checkIsIntl();
    isIntl.value = res.data;
    globalStore.isIntl = isIntl.value;
};

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
    } else if (command === 'ru') {
        dropdownText.value = 'Русский';
    } else if (command === 'ms') {
        dropdownText.value = 'Bahasa Melayu';
    }
}

const agreeWithLogin = () => {
    open.value = false;
    loginForm.agreeLicense = true;
    login(loginFormRef.value);
};

const login = (formEl: FormInstance | undefined) => {
    if (!formEl || isLoggingIn) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (isIntl.value) {
            loginForm.agreeLicense = true;
        }
        if (!loginForm.agreeLicense) {
            if (_isMobile()) {
                open.value = true;
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
            globalStore.ignoreCaptcha = true;
            if (res.data.mfaStatus === 'Enable') {
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
        } catch (res) {
            if (res.code === 401) {
                if (res.message === 'ErrCaptchaCode') {
                    loginForm.captcha = '';
                    errCaptcha.value = true;
                    errAuthInfo.value = false;
                }
                if (res.message === 'ErrAuth') {
                    globalStore.ignoreCaptcha = false;
                    errCaptcha.value = false;
                    errAuthInfo.value = true;
                    console.log('11111');
                }
            }
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
        try {
            await mfaLoginApi(mfaLoginForm);
            globalStore.setLogStatus(true);
            menuStore.setMenuList([]);
            tabsStore.removeAllTabs();
            MsgSuccess(i18n.t('commons.msg.loginSuccess'));
            loadDataFromDB();
            router.push({ name: 'home' });
        } catch (res) {
            if (res.code === 401) {
                errMfaInfo.value = true;
                isLoggingIn = false;
                return;
            }
        } finally {
            isLoggingIn = false;
        }
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
    let theme = globalStore.themeConfig.theme === res.data.theme ? res.data.theme : globalStore.themeConfig.theme;
    globalStore.setThemeConfig({ ...themeConfig.value, theme: theme, panelName: res.data.panelName });
};

onMounted(() => {
    globalStore.isOnRestart = false;
    checkIsSystemIntl();
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
.agree {
    text-decoration: none;
}
.agree:hover {
    text-decoration: underline;
}

:deep(.el-button) {
    height: 2.5rem;
}

:deep(.el-input__inner) {
    -webkit-box-shadow: 0 0 0px 1000px transparent inset !important;
    transition: background-color 50000s ease-in-out 0s;
}

:deep(.el-row) {
    padding: 0 !important;
}
</style>
<!-- <style scoped lang='scss' > .login-form {
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
        border-color: #005eeb;
        color: #ffffff;
        &:hover {
            --el-button-hover-border-color: #005eeb;
        }
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

    :deep(a) {
        color: #005eeb;
        &:hover {
            color: #005eeb95;
        }
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
    :deep(.el-input__inner) {
        color: #000 !important;
    }
}

.cursor-pointer {
    outline: none;
}
.el-dropdown:focus-visible {
    outline: none;
}
.el-tooltip__trigger:focus-visible {
    outline: none;
}
:deep(.el-dropdown-menu__item:not(.is-disabled):hover) {
    color: #005eeb !important;
    background-color: #e5eefd !important;
}
:deep(.el-dropdown-menu__item:not(.is-disabled):focus) {
    color: #005eeb !important;
    background-color: #e5eefd !important;
}

.login-footer-btn {
    .el-button--primary {
        border-color: #005eeb !important;
        background-color: #005eeb !important;
    }
}
</style>
-->
