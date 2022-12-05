<template>
    <div>
        <div v-if="isFirst">
            <div class="login-container">
                <el-form ref="registerFormRef" :model="registerForm" size="default" :rules="registerRules">
                    <div class="login-title">1Panel</div>
                    <div class="login-border"></div>
                    <div class="login-welcome">{{ $t('commons.login.firstLogin') }}</div>
                    <div class="login-form">
                        <el-form-item prop="name">
                            <el-input v-model="registerForm.name" :placeholder="$t('commons.login.username')">
                                <template #prefix>
                                    <el-icon class="el-input__icon">
                                        <user />
                                    </el-icon>
                                </template>
                            </el-input>
                        </el-form-item>
                        <el-form-item prop="password">
                            <el-input
                                type="password"
                                v-model="registerForm.password"
                                show-password
                                :placeholder="$t('commons.login.password')"
                            >
                                <template #prefix>
                                    <el-icon class="el-input__icon">
                                        <lock />
                                    </el-icon>
                                </template>
                            </el-input>
                        </el-form-item>
                        <el-form-item prop="rePassword">
                            <el-input
                                type="password"
                                v-model="registerForm.rePassword"
                                show-password
                                :placeholder="$t('commons.login.rePassword')"
                            >
                                <template #prefix>
                                    <el-icon class="el-input__icon">
                                        <lock />
                                    </el-icon>
                                </template>
                            </el-input>
                        </el-form-item>
                        <el-form-item>
                            <el-button
                                @click="register(registerFormRef)"
                                style="width: 100%"
                                type="primary"
                                size="default"
                            >
                                {{ $t('commons.button.init') }}
                            </el-button>
                        </el-form-item>
                    </div>
                </el-form>
            </div>
        </div>
        <div v-else>
            <div v-if="mfaShow">
                <div class="login-container">
                    <el-form>
                        <div class="login-title">1Panel</div>
                        <div class="login-border"></div>
                        <div class="login-welcome">{{ $t('commons.login.codeInput') }}</div>
                        <div class="login-form">
                            <el-form-item>
                                <el-input
                                    size="default"
                                    :placeholder="$t('commons.login.captchaHelper')"
                                    v-model="mfaLoginForm.code"
                                >
                                    <template #prefix>
                                        <el-icon class="el-input__icon">
                                            <Finished />
                                        </el-icon>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item>
                                <el-button
                                    class="login-btn"
                                    style="width: 100%"
                                    size="default"
                                    type="primary"
                                    @click="mfaLogin()"
                                >
                                    {{ $t('commons.button.verify') }}
                                </el-button>
                            </el-form-item>
                        </div>
                    </el-form>
                </div>
            </div>
            <div v-if="!mfaShow">
                <div class="login-container">
                    <el-form ref="loginFormRef" :model="loginForm" size="default" :rules="loginRules">
                        <div class="login-title">1Panel</div>
                        <div class="login-border"></div>
                        <div class="login-welcome">{{ $t('commons.login.welcome') }}</div>
                        <div class="login-form">
                            <el-form-item prop="username">
                                <el-input v-model="loginForm.name" :placeholder="$t('commons.login.username')">
                                    <template #prefix>
                                        <el-icon class="el-input__icon">
                                            <user />
                                        </el-icon>
                                    </template>
                                </el-input>
                            </el-form-item>
                            <el-form-item prop="password">
                                <el-input
                                    type="password"
                                    v-model="loginForm.password"
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
                            <el-form-item prop="captcha">
                                <el-input
                                    v-model="loginForm.captcha"
                                    :placeholder="$t('commons.login.captchaHelper')"
                                    style="width: 70%"
                                />
                                <img
                                    v-if="captcha.imagePath"
                                    :src="captcha.imagePath"
                                    style="width: 30%; height: 32px"
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
                                    style="width: 100%"
                                    type="primary"
                                    size="default"
                                >
                                    {{ $t('commons.button.login') }}
                                </el-button>
                            </el-form-item>
                        </div>
                    </el-form>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import type { ElForm } from 'element-plus';
import { ElMessage } from 'element-plus';
import { loginApi, getCaptcha, mfaLoginApi, checkIsFirst, initUser } from '@/api/modules/auth';
import { GlobalStore } from '@/store';
import { MenuStore } from '@/store/modules/menu';
import i18n from '@/lang';
import { Rules } from '@/global/form-rules';

const globalStore = GlobalStore();
const menuStore = MenuStore();

const errAuthInfo = ref(false);
const errCaptcha = ref(false);

const isFirst = ref();

type FormInstance = InstanceType<typeof ElForm>;

const registerFormRef = ref<FormInstance>();
const registerForm = reactive({
    name: '',
    password: '',
    rePassword: '',
});
const registerRules = reactive({
    name: [Rules.requiredInput],
    password: [Rules.requiredInput],
    rePassword: [Rules.requiredInput, { validator: checkPassword, trigger: 'blur' }],
});

const loginFormRef = ref<FormInstance>();
const loginForm = reactive({
    name: '',
    password: '',
    captcha: '',
    captchaID: '',
    authMethod: '',
});
const loginRules = reactive({
    name: [{ required: true, message: i18n.global.t('commons.rule.username'), trigger: 'blur' }],
    password: [{ required: true, message: i18n.global.t('commons.rule.password'), trigger: 'blur' }],
});
const mfaLoginForm = reactive({
    name: '',
    password: '',
    secret: '',
    code: '',
    authMethod: '',
});

const captcha = reactive({
    captchaID: '',
    imagePath: '',
    captchaLength: 0,
});

const loading = ref<boolean>(false);
const mfaShow = ref<boolean>(false);

const router = useRouter();

const register = (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        await initUser(registerForm);
        checkStatus();
        ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
    });
};

const login = (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        try {
            let requestLoginForm = {
                name: loginForm.name,
                password: loginForm.password,
                captcha: loginForm.captcha,
                captchaID: captcha.captchaID,
                authMethod: '',
            };
            const res = await loginApi(requestLoginForm);
            if (res.code === 406) {
                if (res.message === 'ErrCaptchaCode') {
                    errCaptcha.value = true;
                    errAuthInfo.value = false;
                    loginVerify();
                }
                if (res.message === 'ErrAuth') {
                    errCaptcha.value = false;
                    errAuthInfo.value = true;
                    loginVerify();
                }
                return;
            }
            if (res.data.mfaStatus === 'enable') {
                mfaShow.value = true;
                mfaLoginForm.secret = res.data.mfaSecret;
                return;
            }
            globalStore.setLogStatus(true);
            menuStore.setMenuList([]);
            ElMessage.success(i18n.global.t('commons.msg.loginSuccess'));
            router.push({ name: 'home' });
        } catch (error) {
            loginVerify();
        } finally {
            loading.value = false;
        }
    });
};

const mfaLogin = async () => {
    if (mfaLoginForm.code) {
        mfaLoginForm.name = loginForm.name;
        mfaLoginForm.password = loginForm.password;
        await mfaLoginApi(mfaLoginForm);
        globalStore.setLogStatus(true);
        menuStore.setMenuList([]);
        ElMessage.success(i18n.global.t('commons.msg.loginSuccess'));
        router.push({ name: 'home' });
    }
};
const loginVerify = async () => {
    const res = await getCaptcha();
    captcha.imagePath = res.data.imagePath ? res.data.imagePath : '';
    captcha.captchaID = res.data.captchaID ? res.data.captchaID : '';
    captcha.captchaLength = res.data.captchaLength ? res.data.captchaLength : 0;
};

const checkStatus = async () => {
    const res = await checkIsFirst();
    isFirst.value = res.data;
    if (!isFirst.value) {
        loginVerify();
    }
};

function checkPassword(rule: any, value: any, callback: any) {
    if (registerForm.password !== registerForm.rePassword) {
        return callback(new Error(i18n.global.t('commons.rule.rePassword')));
    }
    callback();
}

onMounted(() => {
    checkStatus();
    document.onkeydown = (e: any) => {
        e = window.event || e;
        if (e.code === 'Enter' || e.code === 'enter' || e.code === 'NumpadEnter') {
            if (loading.value) return;
            login(loginFormRef.value);
        }
    };
});
</script>

<style scoped lang="scss">
.login-container {
    .login-logo {
        margin-top: 30px;
        margin-left: 30px;
        @media only screen and (max-width: 1280px) {
            margin-top: 20px;
        }

        img {
            height: 45px;
        }
    }

    .login-title {
        margin-top: 50px;
        font-size: 32px;
        letter-spacing: 0;
        text-align: center;
        color: #999999;

        @media only screen and (max-width: 1280px) {
            margin-top: 20px;
        }
    }

    .login-border {
        height: 2px;
        margin: 20px auto 20px;
        position: relative;
        width: 80px;
        @media only screen and (max-width: 1280px) {
            margin: 10px auto 10px;
        }
    }

    .login-welcome {
        margin-top: 10px;
        font-size: 14px;
        color: #999999;
        letter-spacing: 0;
        line-height: 18px;
        text-align: center;
        @media only screen and (max-width: 1280px) {
            margin-top: 20px;
        }
    }

    .login-form {
        margin-top: 30px;
        padding: 0 40px;

        @media only screen and (max-width: 1280px) {
            margin-top: 10px;
        }

        & ::v-deep .el-input__inner {
            border-radius: 0;
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
}
</style>
