<template>
    <div v-loading="loading" class="w-full h-full flex items-center justify-center px-8">
        <div class="w-full flex-grow flex flex-col login-form">
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
                                ref="mfaLoginRef"
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
                                class="w-full login-button"
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
                                        <el-dropdown-item command="zh-Hant">中文(繁體)</el-dropdown-item>
                                        <el-dropdown-item v-if="!globalStore.isIntl" command="en">
                                            English
                                        </el-dropdown-item>
                                        <el-dropdown-item command="ja">日本語</el-dropdown-item>
                                        <el-dropdown-item command="pt-BR">Português (Brasil)</el-dropdown-item>
                                        <el-dropdown-item command="ko">한국어</el-dropdown-item>
                                        <el-dropdown-item command="ru">Русский</el-dropdown-item>
                                        <el-dropdown-item command="ms">Bahasa Melayu</el-dropdown-item>
                                        <el-dropdown-item command="Tr">Turkish</el-dropdown-item>
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
                                name="username"
                                autocomplete="username"
                                ref="userNameRef"
                            ></el-input>
                        </el-form-item>
                        <el-form-item prop="password" class="w-full">
                            <el-input
                                type="password"
                                show-password
                                v-model.trim="loginForm.password"
                                class="w-full"
                                size="large"
                                :placeholder="$t('commons.login.password')"
                                name="password"
                                autocomplete="current-password"
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
                            <el-col :span="24" class="h-0.5">
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
                                class="w-full login-button"
                                type="primary"
                                size="default"
                            >
                                {{ $t('commons.button.login') }}
                            </el-button>
                        </el-form-item>
                        <el-text v-if="isDemo" type="danger" class="demo">
                            {{ $t('commons.login.username') }}:demo {{ $t('commons.login.password') }}:1panel
                        </el-text>
                        <el-form-item prop="agreeLicense" v-if="!isIntl && !isFxplay">
                            <el-checkbox v-model="loginForm.agreeLicense">
                                <template #default>
                                    <span class="agree-title">
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
                    </div>
                </el-form>
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
import { ref, reactive, onMounted, computed, nextTick } from 'vue';
import type { ElForm } from 'element-plus';
import { loginApi, getCaptcha, mfaLoginApi, getLoginSetting } from '@/api/modules/auth';
import { GlobalStore, MenuStore, TabsStore } from '@/store';
import { MsgError, MsgSuccess } from '@/utils/message';
import { useI18n } from 'vue-i18n';
import { encryptPassword } from '@/utils/util';
import { getXpackSettingForTheme } from '@/utils/xpack';
import { routerToName } from '@/utils/router';
import { changeToLocal } from '@/utils/node';

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
const isFxplay = ref(false);
const open = ref(false);
const loginBtnLinkColor = ref<string | null>(null);

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
const userNameRef = ref();
const mfaLoginRef = ref();
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
const dropdownText = ref('中文(简体)');

function handleCommand(command: string) {
    loginForm.language = command;
    usei18n.locale.value = command;
    globalStore.updateLanguage(command);
    if (command === 'zh') {
        dropdownText.value = '中文(简体)';
    } else if (command === 'en') {
        dropdownText.value = 'English';
    } else if (command === 'pt-BR') {
        dropdownText.value = 'Português (Brasil)';
    } else if (command === 'zh-Hant') {
        dropdownText.value = '中文(繁體)';
    } else if (command === 'ko') {
        dropdownText.value = '한국어';
    } else if (command === 'ja') {
        dropdownText.value = '日本語';
    } else if (command === 'ru') {
        dropdownText.value = 'Русский';
    } else if (command === 'ms') {
        dropdownText.value = 'Bahasa Melayu';
    } else if (command === 'tr') {
        dropdownText.value = 'Turkish';
    } else if (command === 'es-ES') {
        dropdownText.value = 'España - Español';
    }
}

const agreeWithLogin = () => {
    open.value = false;
    loginForm.agreeLicense = true;
    login(loginFormRef.value);
};

const login = (formEl: FormInstance | undefined) => {
    if (!formEl || isLoggingIn) return;
    errAuthInfo.value = false;
    errCaptcha.value = false;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (isIntl.value || isFxplay.value) {
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
            password: encryptPassword(loginForm.password),
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
                nextTick(() => {
                    mfaLoginRef.value?.focus();
                });
                return;
            }
            globalStore.setLogStatus(true);
            globalStore.setAgreeLicense(true);
            menuStore.setMenuList([]);
            tabsStore.removeAllTabs();
            changeToLocal();
            MsgSuccess(i18n.t('commons.msg.loginSuccess'));
            routerToName('home');
            document.onkeydown = null;
        } catch (res) {
            if (res.code === 401) {
                if (res.message === 'ErrCaptchaCode') {
                    loginForm.captcha = '';
                    errCaptcha.value = true;
                    errAuthInfo.value = false;
                    loginVerify();
                    return;
                }
                if (res.message === 'ErrAuth') {
                    globalStore.ignoreCaptcha = false;
                    errCaptcha.value = false;
                    errAuthInfo.value = true;
                    loginVerify();
                    return;
                }
                MsgError(res.message);
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
        mfaLoginForm.password = encryptPassword(loginForm.password);
        try {
            await mfaLoginApi(mfaLoginForm);
            globalStore.setLogStatus(true);
            menuStore.setMenuList([]);
            tabsStore.removeAllTabs();
            MsgSuccess(i18n.t('commons.msg.loginSuccess'));
            changeToLocal();
            routerToName('home');
            document.onkeydown = null;
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

const getSetting = async () => {
    try {
        const res = await getLoginSetting();
        isDemo.value = res.data.isDemo;
        loginForm.language = res.data.language;
        handleCommand(loginForm.language);
        isIntl.value = res.data.isIntl;
        isFxplay.value = res.data.isFxplay;
        globalStore.isFxplay = isFxplay.value;
        globalStore.isOffLine = res.data.isOffLine;

        document.title = res.data.panelName;
        i18n.locale.value = res.data.language;
        i18n.warnHtmlMessage = false;
        globalStore.setOpenMenuTabs(res.data.menuTabs === 'Enable');
        globalStore.updateLanguage(res.data.language);
        globalStore.setThemeConfig({ ...themeConfig.value, theme: res.data.theme, panelName: res.data.panelName });
    } catch (error) {}
};

function adjustColorToRGBA(color: string, percent: number, opacity: number): string {
    let r = 0,
        g = 0,
        b = 0,
        a = opacity;

    color = color.trim();

    if (color.startsWith('#')) {
        if (color.length === 4) {
            r = parseInt(color[1] + color[1], 16);
            g = parseInt(color[2] + color[2], 16);
            b = parseInt(color[3] + color[3], 16);
        } else if (color.length === 7) {
            r = parseInt(color.slice(1, 3), 16);
            g = parseInt(color.slice(3, 5), 16);
            b = parseInt(color.slice(5, 7), 16);
        } else {
            return color;
        }
    } else if (color.startsWith('rgb')) {
        const result = color.match(/rgba?\((\d+),\s*(\d+),\s*(\d+)(?:,\s*([0-9.]+))?\)/);
        if (!result) return color;
        r = parseInt(result[1], 10);
        g = parseInt(result[2], 10);
        b = parseInt(result[3], 10);
        if (result[4] !== undefined) {
            a = parseFloat(result[4]);
        }
    } else {
        return color;
    }

    r = Math.min(255, Math.max(0, Math.round(r * (1 + percent / 100))));
    g = Math.min(255, Math.max(0, Math.round(g * (1 + percent / 100))));
    b = Math.min(255, Math.max(0, Math.round(b * (1 + percent / 100))));
    a = Math.min(1, Math.max(0, opacity / 100));

    return `rgba(${r}, ${g}, ${b}, ${a})`;
}

onMounted(() => {
    globalStore.isOnRestart = false;
    getSetting();
    getXpackSettingForTheme();
    if (!globalStore.ignoreCaptcha) {
        loginVerify();
    }
    document.title = globalStore.themeConfig.panelName;
    loginBtnLinkColor.value = globalStore.themeConfig.loginBtnLinkColor || '#005eeb';
    document.documentElement.style.setProperty('--login-btn-link-color', loginBtnLinkColor.value);
    document.documentElement.style.setProperty(
        '--login-btn-link-hover-color',
        adjustColorToRGBA(loginBtnLinkColor.value, -10, 80),
    );
    document.documentElement.style.setProperty(
        '--login-loading-mask-color',
        adjustColorToRGBA(loginBtnLinkColor.value, 30, 15),
    );
    nextTick(() => {
        userNameRef.value?.focus();
    });
    loginForm.agreeLicense = globalStore.agreeLicense;
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

.login-form {
    .login-button {
        background-color: var(--login-btn-link-color);
        border-color: var(--login-btn-link-color);
        color: #ffffff;
        &:hover {
            background-color: var(--login-btn-link-hover-color) !important;
            border-color: var(--login-btn-link-hover-color) !important;
            outline: none !important;
        }
    }

    :deep(.el-input) {
        --el-input-border-color: #dcdfe6;
        background: none !important;
    }

    :deep(.el-input__wrapper) {
        background: none !important;
    }
    :deep(.el-input__wrapper.is-focus) {
        box-shadow: 0 0 0 1px var(--login-btn-link-color) inset !important;
    }

    .demo {
        text-align: center;
        span {
            color: red;
        }
    }

    .agree-title {
        color: var(--login-btn-link-color);
    }

    .agree {
        white-space: pre-wrap;
        color: var(--login-btn-link-color);
    }

    :deep(a) {
        color: var(--login-btn-link-color);
        &:hover {
            opacity: 75%;
        }
    }

    :deep(.el-checkbox__input .el-checkbox__inner) {
        background-color: #fff !important;
        border-color: var(--login-btn-link-color) !important;
    }

    :deep(.el-checkbox__input.is-checked .el-checkbox__inner) {
        background-color: var(--login-btn-link-color) !important;
        border-color: var(--login-btn-link-color) !important;
    }

    :deep(.el-checkbox__input.is-checked .el-checkbox__inner::after) {
        border-color: #ffffff !important;
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
    background-color: var(--login-btn-link-color) !important;
    color: #fff !important;
}
:deep(.el-dropdown-menu__item:not(.is-disabled):focus) {
    background-color: var(--login-btn-link-color) !important;
    color: #fff !important;
}

:deep(.el-loading-mask) {
    background-color: var(--login-loading-mask-color) !important;
    .el-loading-spinner .path {
        stroke: var(--login-btn-link-color);
    }
}

.login-footer-btn {
    .el-button--primary {
        border-color: var(--login-btn-link-color) !important;
        background-color: var(--login-btn-link-color) !important;
    }
}
</style>
