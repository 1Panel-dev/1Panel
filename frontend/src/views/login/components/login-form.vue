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
                                autocomplete="one-time-code"
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
            <div v-else-if="showPasskeyOnly">
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
                                    <el-dropdown-item v-if="isIntl" command="en">English</el-dropdown-item>
                                    <el-dropdown-item command="zh">中文(简体)</el-dropdown-item>
                                    <el-dropdown-item command="zh-Hant">中文(繁體)</el-dropdown-item>
                                    <el-dropdown-item v-if="!isIntl" command="en">English</el-dropdown-item>
                                    <el-dropdown-item command="ja">日本語</el-dropdown-item>
                                    <el-dropdown-item command="pt-BR">Português (Brasil)</el-dropdown-item>
                                    <el-dropdown-item command="ko">한국어</el-dropdown-item>
                                    <el-dropdown-item command="ru">Русский</el-dropdown-item>
                                    <el-dropdown-item command="ms">Bahasa Melayu</el-dropdown-item>
                                    <el-dropdown-item command="tr">Turkish</el-dropdown-item>
                                    <el-dropdown-item command="fa">فارسی</el-dropdown-item>
                                    <el-dropdown-item command="lo">ພາສາລາວ</el-dropdown-item>
                                </el-dropdown-menu>
                            </template>
                        </el-dropdown>
                    </div>
                </div>
                <div class="space-y-6">
                    <el-form-item>
                        <el-button class="w-full login-button" type="primary" size="default" @click="passkeyLogin">
                            <el-icon class="mr-2"><Key /></el-icon>
                            {{ $t('commons.login.passkey') }}
                        </el-button>
                    </el-form-item>
                    <div v-if="hasExternalLoginMethods" class="external-login-section">
                        <div class="external-login-divider">
                            <span>{{ $t('commons.login.otherLoginMethods') }}</span>
                        </div>
                        <div class="external-login-methods">
                            <el-button
                                v-if="ldapEnabled"
                                class="external-login-button ldap-login-button"
                                link
                                native-type="button"
                                :aria-label="$t('xpack.user.auth.ldap.loginWith')"
                                @click="switchToLDAPLogin"
                            >
                                <span>LDAP</span>
                            </el-button>
                            <span
                                v-if="ldapEnabled && (oidcEnabled || saml2Enabled)"
                                class="external-login-separator"
                                aria-hidden="true"
                            ></span>
                            <el-button
                                v-if="oidcEnabled"
                                class="external-login-button oidc-login-button"
                                link
                                native-type="button"
                                :aria-label="$t('xpack.user.auth.oidc.loginWith', { provider: oidcDisplayName })"
                                :loading="oidcStarting"
                                @click="beginOIDCLogin"
                            >
                                <span>{{ oidcDisplayName }}</span>
                            </el-button>
                            <span
                                v-if="oidcEnabled && saml2Enabled"
                                class="external-login-separator"
                                aria-hidden="true"
                            ></span>
                            <el-button
                                v-if="saml2Enabled"
                                class="external-login-button saml2-login-button"
                                link
                                native-type="button"
                                :aria-label="$t('xpack.user.auth.saml2.loginWith', { provider: saml2DisplayName })"
                                :loading="saml2Starting"
                                @click="beginSAML2Login"
                            >
                                <span>{{ saml2DisplayName }}</span>
                            </el-button>
                        </div>
                    </div>
                    <el-form-item>
                        <el-link type="primary" :underline="false" @click="switchToPasswordLogin">
                            {{ $t('commons.login.passkeyToPassword') }}
                        </el-link>
                    </el-form-item>
                    <el-form-item v-if="!isIntl && !isEnterprise && !isFxplay">
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
            </div>
            <div v-else>
                <div class="flex justify-between items-center mb-6">
                    <div>
                        <div class="text-2xl font-medium text-gray-900">
                            {{
                                loginSource === 'ldap'
                                    ? $t('xpack.user.auth.ldap.loginTitle')
                                    : $t('commons.button.login')
                            }}
                        </div>
                        <el-link
                            v-if="loginSource === 'ldap'"
                            class="local-login-link"
                            type="primary"
                            :underline="false"
                            @click="switchToLocalLogin"
                        >
                            {{ $t('xpack.user.auth.ldap.backToLocalLogin') }}
                        </el-link>
                    </div>
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
                                    <el-dropdown-item v-if="isIntl" command="en">English</el-dropdown-item>
                                    <el-dropdown-item command="zh">中文(简体)</el-dropdown-item>
                                    <el-dropdown-item command="zh-Hant">中文(繁體)</el-dropdown-item>
                                    <el-dropdown-item v-if="!isIntl" command="en">English</el-dropdown-item>
                                    <el-dropdown-item command="ja">日本語</el-dropdown-item>
                                    <el-dropdown-item command="pt-BR">Português (Brasil)</el-dropdown-item>
                                    <el-dropdown-item command="ko">한국어</el-dropdown-item>
                                    <el-dropdown-item command="ru">Русский</el-dropdown-item>
                                    <el-dropdown-item command="ms">Bahasa Melayu</el-dropdown-item>
                                    <el-dropdown-item command="tr">Turkish</el-dropdown-item>
                                    <el-dropdown-item command="fa">فارسی</el-dropdown-item>
                                    <el-dropdown-item command="lo">ພາສາລາວ</el-dropdown-item>
                                </el-dropdown-menu>
                            </template>
                        </el-dropdown>
                    </div>
                </div>
                <el-form ref="loginFormRef" :model="loginForm" size="default" :rules="loginRules">
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
                            <el-col :span="12" v-if="!ignoreCaptcha">
                                <el-form-item prop="captcha">
                                    <el-input
                                        v-model.trim="loginForm.captcha"
                                        size="large"
                                        :placeholder="$t('commons.login.captchaHelper')"
                                    ></el-input>
                                </el-form-item>
                            </el-col>
                            <el-col :span="12" v-if="!ignoreCaptcha">
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
                        <div v-if="loginSource === 'local' && hasExternalLoginMethods" class="external-login-section">
                            <div class="external-login-divider">
                                <span>{{ $t('commons.login.otherLoginMethods') }}</span>
                            </div>
                            <div class="external-login-methods">
                                <el-button
                                    v-if="ldapEnabled"
                                    class="external-login-button ldap-login-button"
                                    link
                                    native-type="button"
                                    :aria-label="$t('xpack.user.auth.ldap.loginWith')"
                                    @click="switchToLDAPLogin"
                                >
                                    <span>LDAP</span>
                                </el-button>
                                <span
                                    v-if="ldapEnabled && (oidcEnabled || saml2Enabled)"
                                    class="external-login-separator"
                                    aria-hidden="true"
                                ></span>
                                <el-button
                                    v-if="oidcEnabled"
                                    class="external-login-button oidc-login-button"
                                    link
                                    native-type="button"
                                    :aria-label="$t('xpack.user.auth.oidc.loginWith', { provider: oidcDisplayName })"
                                    :loading="oidcStarting"
                                    @click="beginOIDCLogin"
                                >
                                    <span>{{ oidcDisplayName }}</span>
                                </el-button>
                                <span
                                    v-if="oidcEnabled && saml2Enabled"
                                    class="external-login-separator"
                                    aria-hidden="true"
                                ></span>
                                <el-button
                                    v-if="saml2Enabled"
                                    class="external-login-button saml2-login-button"
                                    link
                                    native-type="button"
                                    :aria-label="$t('xpack.user.auth.saml2.loginWith', { provider: saml2DisplayName })"
                                    :loading="saml2Starting"
                                    @click="beginSAML2Login"
                                >
                                    <span>{{ saml2DisplayName }}</span>
                                </el-button>
                            </div>
                        </div>
                        <el-text v-if="isDemo" type="danger" class="demo">
                            {{ $t('commons.login.username') }}:demo {{ $t('commons.login.password') }}:1panel
                        </el-text>
                        <el-form-item prop="agreeLicense" v-if="!isIntl && !isEnterprise && !isFxplay">
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
import { ref, reactive, onMounted, onBeforeUnmount, computed, nextTick } from 'vue';
import type { ElForm } from 'element-plus';
import {
    loginApi,
    getCaptcha,
    mfaLoginApi,
    getLoginSetting,
    passkeyBeginApi,
    passkeyFinishApi,
    ldapStatusApi,
    oidcStatusApi,
    oidcBeginApi,
    oidcFinishApi,
    saml2StatusApi,
    saml2BeginApi,
    saml2FinishApi,
} from '@/api/modules/auth';
import type { Login as LoginModel } from '@/api/interface/auth';
import { MenuStore, TabsStore } from '@/store';
import { MsgError, MsgSuccess } from '@/utils/message';
import { useI18n } from 'vue-i18n';
import { encryptPassword, base64UrlToBuffer, bufferToBase64Url } from '@/utils/auth';
import { takeExternalTicketsFromURL } from '@/utils/external-login';
import { getXpackSettingForTheme } from '@/utils/xpack';
import { routerToName } from '@/utils/router';
import { Key } from '@element-plus/icons-vue';
import { changeToLocal } from '@/utils/node';
import { syncAuthInfo } from '@/utils/rbac';
import { adjustColorToRGBA } from '@/utils/color';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { submitSAML2Navigation } from '@/utils/saml2';

const emit = defineEmits<{
    (e: 'external-login-ready'): void;
}>();
const i18n = useI18n();
const {
    globalStore,
    agreeLicense,
    currentNode,
    ignoreCaptcha,
    isAdmin,
    isEnterprise,
    isEnterpriseLicenseLoaded,
    isFxplay,
    isIntl,
    isLogin,
    isOffline,
    isOnRestart,
    openMenuTabs,
    menuAccordion,
    themeConfig,
} = useGlobalStore();
const menuStore = MenuStore();
const tabsStore = TabsStore();

const errAuthInfo = ref(false);
const errCaptcha = ref(false);
const errMfaInfo = ref(false);
const passkeySetting = ref(false);
const passkeySupported = ref(false);
const ldapEnabled = ref(false);
const oidcEnabled = ref(false);
const oidcDisplayName = ref('OIDC');
const oidcStarting = ref(false);
const saml2Enabled = ref(false);
const saml2DisplayName = ref('SAML2');
const saml2Starting = ref(false);
const autoPasskeyEnabledKey = '1panel-passkey-auto-enabled';
const showPasswordLogin = ref(false);
const loginSource = ref<'local' | 'ldap'>('local');
const isDemo = ref(false);
const open = ref(false);
const loginBtnLinkColor = ref<string | null>(null);
let loginViewActive = true;

const pendingExternalTickets = takeExternalTicketsFromURL();
const pendingOIDCTicket = pendingExternalTickets.oidcTicket;
const pendingSAML2Ticket = pendingOIDCTicket ? '' : pendingExternalTickets.saml2Ticket;
const hasPendingExternalTicket = Boolean(pendingOIDCTicket || pendingSAML2Ticket);

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
const pendingLoginMethod = ref<'password' | 'passkey'>('password');
const mfaLoginForm = reactive({
    sessionId: '',
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

const isAutoPasskeyEnabled = () => {
    try {
        return localStorage.getItem(autoPasskeyEnabledKey) === '1';
    } catch (error) {
        return false;
    }
};
const enableAutoPasskey = () => {
    try {
        localStorage.setItem(autoPasskeyEnabledKey, '1');
    } catch (error) {}
};
const disableAutoPasskey = () => {
    try {
        localStorage.removeItem(autoPasskeyEnabledKey);
    } catch (error) {}
};

const languageLabelMap: Record<string, string> = {
    zh: '中文(简体)',
    en: 'English',
    'pt-BR': 'Português (Brasil)',
    'zh-Hant': '中文(繁體)',
    ko: '한국어',
    ja: '日本語',
    ru: 'Русский',
    ms: 'Bahasa Melayu',
    tr: 'Turkish',
    'es-ES': 'España - Español',
    fa: 'فارسی',
    lo: 'ພາສາລາວ',
};

const handleCommand = async (command: string) => {
    const activeLocale = await globalStore.updateLanguage(command);
    loginForm.language = activeLocale;
    dropdownText.value = languageLabelMap[activeLocale] || languageLabelMap.zh;
};

const agreeWithLogin = () => {
    open.value = false;
    loginForm.agreeLicense = true;
    if (pendingLoginMethod.value === 'passkey') {
        passkeyLogin();
        return;
    }
    login(loginFormRef.value);
};

const showPasskeyOnly = computed(() => {
    return passkeySetting.value && passkeySupported.value && !showPasswordLogin.value;
});
const hasExternalLoginMethods = computed(() => ldapEnabled.value || oidcEnabled.value || saml2Enabled.value);

const switchToPasswordLogin = () => {
    loginSource.value = 'local';
    showPasswordLogin.value = true;
    nextTick(() => {
        userNameRef.value?.focus();
    });
};

const switchToLDAPLogin = () => {
    loginSource.value = 'ldap';
    showPasswordLogin.value = true;
    errAuthInfo.value = false;
    errCaptcha.value = false;
    nextTick(() => {
        userNameRef.value?.focus();
    });
};

const switchToLocalLogin = () => {
    loginSource.value = 'local';
    errAuthInfo.value = false;
    errCaptcha.value = false;
    nextTick(() => {
        userNameRef.value?.focus();
    });
};

const navigateAfterLogin = async () => {
    try {
        await routerToName('home');
    } catch {
        if (hasPendingExternalTicket) {
            emit('external-login-ready');
        }
    }
};

const completeLogin = async (result: LoginModel.ResLogin) => {
    isLogin.value = true;
    agreeLicense.value = true;
    menuStore.setMenuList([]);
    tabsStore.removeAllTabs();
    isAdmin.value = result.role === 'ADMIN';
    await changeToLocal();
    await syncAuthInfo(currentNode.value);
    MsgSuccess(i18n.t('commons.msg.loginSuccess'));
    localStorage.removeItem('dashboardCache');
    localStorage.removeItem('upgradeChecked');
    clearLoginKeydownHandler();
    await navigateAfterLogin();
};

const handleLoginResult = async (result: LoginModel.ResLogin) => {
    if (result.mfaStatus === 'Enable') {
        mfaLoginForm.sessionId = result.mfaSession || '';
        mfaLoginForm.code = '';
        mfaShow.value = true;
        errMfaInfo.value = false;
        errCaptcha.value = false;
        nextTick(() => {
            mfaLoginRef.value?.focus();
        });
        return;
    }
    await completeLogin(result);
};

const login = (formEl: FormInstance | undefined) => {
    if (!formEl || isLoggingIn) return;
    errAuthInfo.value = false;
    errCaptcha.value = false;
    formEl.validate(async (valid) => {
        if (!valid) return;
        if (isIntl.value || isFxplay.value || isEnterprise.value) {
            loginForm.agreeLicense = true;
        }
        if (!loginForm.agreeLicense) {
            if (_isMobile()) {
                pendingLoginMethod.value = 'password';
                open.value = true;
            }
            return;
        }
        let requestLoginForm = {
            name: loginForm.name,
            password: encryptPassword(loginForm.password),
            captcha: loginForm.captcha,
            captchaID: captcha.captchaID,
            authMethod: 'session',
            authSource: loginSource.value,
            language: loginForm.language,
        };
        if (!ignoreCaptcha.value && requestLoginForm.captcha == '') {
            errCaptcha.value = true;
            return;
        }
        try {
            isLoggingIn = true;
            loading.value = true;
            const res = await loginApi(requestLoginForm);
            ignoreCaptcha.value = true;
            await handleLoginResult(res.data);
        } catch (res) {
            if (res.code === 401) {
                if (res.message === 'ErrCaptchaCode') {
                    ignoreCaptcha.value = false;
                    loginForm.captcha = '';
                    errCaptcha.value = true;
                    errAuthInfo.value = false;
                    loginVerify();
                    return;
                }
                if (res.message === 'ErrAuth') {
                    ignoreCaptcha.value = false;
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
        try {
            errMfaInfo.value = false;
            const res = await mfaLoginApi(mfaLoginForm);
            await completeLogin(res.data);
        } catch (res) {
            if (res.code === 401) {
                if (res.message === 'ErrCaptchaCode') {
                    ignoreCaptcha.value = false;
                    mfaLoginForm.code = '';
                    mfaShow.value = false;
                    loginVerify();
                    nextTick(() => {
                        userNameRef.value?.focus();
                    });
                } else if (res.message === 'ErrMFA') {
                    errMfaInfo.value = true;
                } else if (res.message) {
                    MsgError(res.message);
                }
                isLoggingIn = false;
                return;
            }
            loginVerify();
        } finally {
            isLoggingIn = false;
        }
    }
};

const passkeyLogin = async () => {
    if (isLoggingIn || !passkeySetting.value) return;
    if (!passkeySupported.value) {
        disableAutoPasskey();
        MsgError(i18n.t('commons.login.passkeyNotSupported'));
        return;
    }
    if (!isIntl.value && !isEnterprise.value && !isFxplay.value && !loginForm.agreeLicense) {
        if (_isMobile() || showPasskeyOnly.value) {
            pendingLoginMethod.value = 'passkey';
            open.value = true;
        } else {
            MsgError(i18n.t('commons.login.errorAgree'));
        }
        return;
    }
    try {
        isLoggingIn = true;
        loading.value = true;
        const res = await passkeyBeginApi();
        const publicKey = normalizePasskeyRequest(res.data.publicKey);
        const credential = (await navigator.credentials.get({ publicKey })) as PublicKeyCredential | null;
        if (!credential) {
            disableAutoPasskey();
            MsgError(i18n.t('commons.login.passkeyFailed'));
            return;
        }
        const payload = buildPasskeyAssertion(credential);
        const loginRes = await passkeyFinishApi(payload, res.data.sessionId);
        enableAutoPasskey();
        ignoreCaptcha.value = true;
        await handleLoginResult(loginRes.data);
    } catch (res: any) {
        disableAutoPasskey();
        if (res?.message) {
            MsgError(i18n.t('commons.login.passkeyFailed'));
        }
    } finally {
        isLoggingIn = false;
        loading.value = false;
    }
};

const loadLDAPStatus = async () => {
    ldapEnabled.value = false;
    if (!isEnterprise.value) return;
    try {
        const res = await ldapStatusApi();
        ldapEnabled.value = Boolean(res.data.enabled);
    } catch {
        // LDAP is optional. A status failure must not affect local or any other login method.
        ldapEnabled.value = false;
    }
};

const loadOIDCStatus = async () => {
    oidcEnabled.value = false;
    if (!isEnterprise.value) return;
    try {
        const res = await oidcStatusApi();
        oidcEnabled.value = Boolean(res.data.enabled && res.data.authorizationCode);
        oidcDisplayName.value = res.data.displayName?.trim() || 'OIDC';
    } catch {
        // OIDC is optional. A status failure must not affect local password or Passkey login.
        oidcEnabled.value = false;
    }
};

const beginOIDCLogin = async () => {
    if (isLoggingIn || !oidcEnabled.value) return;
    try {
        isLoggingIn = true;
        oidcStarting.value = true;
        const res = await oidcBeginApi();
        if (!res.data.authorizationURL) return;
        window.location.assign(res.data.authorizationURL);
    } catch {
        // The request layer displays the backend-localized error.
    } finally {
        isLoggingIn = false;
        oidcStarting.value = false;
    }
};

const finishOIDCLogin = async (ticket: string) => {
    if (!ticket) return;
    try {
        isLoggingIn = true;
        loading.value = true;
        const res = await oidcFinishApi({ ticket });
        ignoreCaptcha.value = true;
        await handleLoginResult(res.data);
    } catch {
        // The ticket was already removed from the URL; the request layer displays the localized failure.
    } finally {
        isLoggingIn = false;
        loading.value = false;
    }
};

const loadSAML2Status = async () => {
    saml2Enabled.value = false;
    if (!isEnterprise.value) return;
    try {
        const res = await saml2StatusApi();
        saml2Enabled.value = Boolean(res.data.enabled);
        saml2DisplayName.value = res.data.displayName?.trim() || 'SAML2';
    } catch {
        // SAML2 is optional. A status failure must not affect any other login method.
        saml2Enabled.value = false;
    }
};

const beginSAML2Login = async () => {
    if (isLoggingIn || !saml2Enabled.value) return;
    try {
        isLoggingIn = true;
        saml2Starting.value = true;
        const res = await saml2BeginApi();
        try {
            submitSAML2Navigation(res.data.navigation);
        } catch {
            MsgError(i18n.t('commons.msg.operationFailed'));
        }
    } catch {
        // The request layer displays the backend-localized error.
    } finally {
        isLoggingIn = false;
        saml2Starting.value = false;
    }
};

const finishSAML2Login = async (ticket: string) => {
    if (!ticket) return;
    try {
        isLoggingIn = true;
        loading.value = true;
        const res = await saml2FinishApi({ ticket });
        ignoreCaptcha.value = true;
        await handleLoginResult(res.data);
    } catch {
        // The ticket was already removed from the URL; the request layer displays the localized failure.
    } finally {
        isLoggingIn = false;
        loading.value = false;
    }
};

const normalizePasskeyRequest = (publicKey: Record<string, any>): PublicKeyCredentialRequestOptions => {
    const request = { ...publicKey };
    request.challenge = base64UrlToBuffer(request.challenge);
    if (request.allowCredentials && Array.isArray(request.allowCredentials)) {
        request.allowCredentials = request.allowCredentials.map((item) => {
            return { ...item, id: base64UrlToBuffer(item.id) };
        });
    }
    return request as PublicKeyCredentialRequestOptions;
};

const buildPasskeyAssertion = (credential: PublicKeyCredential) => {
    const response = credential.response as AuthenticatorAssertionResponse;
    const payload: Record<string, any> = {
        id: credential.id,
        rawId: bufferToBase64Url(credential.rawId),
        type: credential.type,
        response: {
            clientDataJSON: bufferToBase64Url(response.clientDataJSON),
            authenticatorData: bufferToBase64Url(response.authenticatorData),
            signature: bufferToBase64Url(response.signature),
        },
        clientExtensionResults: credential.getClientExtensionResults(),
        authenticatorAttachment: credential.authenticatorAttachment,
    };
    if (response.userHandle) {
        payload.response.userHandle = bufferToBase64Url(response.userHandle);
    }
    return payload;
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
        const language = res.data.language || loginForm.language;
        await handleCommand(language);
        isIntl.value = res.data.isIntl;
        isFxplay.value = res.data.isFxplay;
        isOffline.value = res.data.isOffline;
        isEnterprise.value = res.data.isEnterprise;
        isEnterpriseLicenseLoaded.value = !res.data.isEnterprise;
        ignoreCaptcha.value = !res.data.needCaptcha;
        passkeySetting.value = res.data.passkeySetting;
        if (!ignoreCaptcha.value) {
            loginVerify();
        }

        document.title = res.data.panelName;
        i18n.warnHtmlMessage = false;
        openMenuTabs.value = res.data.menuTabs === 'Enable';
        menuAccordion.value = res.data.menuAccordion === 'Enable';
        themeConfig.value = { ...themeConfig.value, theme: res.data.theme, panelName: res.data.panelName };

        if (res.data.passkeySetting && !isIntl.value && !isFxplay.value) {
            loginForm.agreeLicense = true;
        }
        if (passkeySetting.value && passkeySupported.value && isAutoPasskeyEnabled() && !hasPendingExternalTicket) {
            passkeyLogin();
        }
    } catch (error) {}
};

const applyLoginButtonTheme = () => {
    loginBtnLinkColor.value = themeConfig.value.loginBtnLinkColor || '#005eeb';
    document.documentElement.style.setProperty('--login-btn-link-color', loginBtnLinkColor.value);
    document.documentElement.style.setProperty(
        '--login-btn-link-hover-color',
        adjustColorToRGBA(loginBtnLinkColor.value, -10, 80),
    );
    document.documentElement.style.setProperty(
        '--login-loading-mask-color',
        adjustColorToRGBA(loginBtnLinkColor.value, 30, 15),
    );
};

function loginKeydownHandler(event: KeyboardEvent) {
    const target = event.target;
    if (event.defaultPrevented || (target instanceof Element && target.closest('.external-login-button'))) return;
    if (event.key === 'Enter' || event.keyCode === 13) {
        if (!mfaShow.value) {
            if (!loginButtonFocused.value) {
                login(loginFormRef.value);
            }
        }
        if (mfaShow.value && !mfaButtonFocused.value) {
            mfaLogin(false);
        }
    }
}

function clearLoginKeydownHandler() {
    if (document.onkeydown === loginKeydownHandler) {
        document.onkeydown = null;
    }
}

onMounted(async () => {
    isOnRestart.value = false;
    passkeySupported.value = !!window.PublicKeyCredential && window.isSecureContext;
    applyLoginButtonTheme();
    await getSetting();
    if (!loginViewActive) return;
    if (pendingOIDCTicket) {
        await finishOIDCLogin(pendingOIDCTicket);
    } else if (pendingSAML2Ticket) {
        await finishSAML2Login(pendingSAML2Ticket);
    }
    if (hasPendingExternalTicket && !isLogin.value) {
        emit('external-login-ready');
    }
    if (!loginViewActive || isLogin.value) return;
    if (!isLogin.value && !mfaShow.value) {
        await Promise.all([loadLDAPStatus(), loadOIDCStatus(), loadSAML2Status()]);
    }
    try {
        await getXpackSettingForTheme();
    } catch (error) {
        // 即使获取失败也不影响登录，默认为之前的主题配置
    }
    if (!loginViewActive) return;
    applyLoginButtonTheme();
    if (!ignoreCaptcha.value) {
        loginVerify();
    }
    document.title = themeConfig.value.panelName;
    nextTick(() => {
        userNameRef.value?.focus();
    });
    loginForm.agreeLicense = agreeLicense.value;
    document.onkeydown = loginKeydownHandler;
});

onBeforeUnmount(() => {
    loginViewActive = false;
    clearLoginKeydownHandler();
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

    .external-login-section {
        width: 100%;
        padding-top: 2px;
    }

    .local-login-link {
        height: auto;
        margin-top: 6px;
        font-size: 13px;
    }

    .external-login-divider {
        display: flex;
        align-items: center;
        gap: 14px;
        color: var(--el-text-color-secondary);
        font-size: 13px;
        white-space: nowrap;

        &::before,
        &::after {
            width: 100%;
            border-top: 1px dashed var(--el-border-color-light);
            content: '';
        }
    }

    .external-login-methods {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        justify-content: center;
        gap: 10px 12px;
        margin-top: 12px;
    }

    .external-login-separator {
        width: 1px;
        height: 14px;
        background-color: var(--el-border-color);
    }

    .external-login-button {
        height: auto;
        margin: 0 !important;
        padding: 5px 2px;
        border: 0;
        background: transparent !important;
        color: var(--login-btn-link-color);
        font-size: 14px;

        &:hover {
            color: var(--login-btn-link-hover-color) !important;
        }

        &:focus-visible {
            border-radius: 3px;
            outline: 2px solid var(--login-btn-link-color) !important;
            outline-offset: 2px;
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
