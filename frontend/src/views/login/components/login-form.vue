<template>
    <div>
        <div v-if="mfaShow" class="login-mfa-form">
            <div class="info-text">
                <span style="font-size: 14px">{{ $t('commons.login.codeInput') }}</span>
            </div>
            <div>
                <span>Secret:{{ mfaLoginForm.secret }}</span>
            </div>
            <el-input v-model="mfaLoginForm.code"></el-input>
            <el-button class="login-btn" type="primary" @click="mfaLogin()">
                {{ $t('commons.button.login') }}
            </el-button>
        </div>
        <div v-if="!mfaShow">
            <div class="login-form">
                <div class="login-logo">
                    <img class="login-icon" src="@/assets/images/logo.svg" alt="" />
                    <h2 class="logo-text">1Panel</h2>
                </div>
                <el-form ref="loginFormRef" :model="loginForm" :rules="loginRules" size="large">
                    <el-form-item prop="username">
                        <el-input v-model="loginForm.name">
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
                            autocomplete="new-password"
                        >
                            <template #prefix>
                                <el-icon class="el-input__icon">
                                    <lock />
                                </el-icon>
                            </template>
                        </el-input>
                    </el-form-item>
                </el-form>
                <el-form-item prop="captcha">
                    <div class="vPicBox">
                        <el-input
                            v-model="loginForm.captcha"
                            :placeholder="$t('commons.login.captchaHelper')"
                            style="width: 60%"
                        />
                        <div class="vPic">
                            <img
                                v-if="captcha.imagePath"
                                :src="captcha.imagePath"
                                :alt="$t('commons.login.captchaHelper')"
                                @click="loginVerify()"
                            />
                        </div>
                    </div>
                </el-form-item>
                <div class="login-btn">
                    <el-button round @click="resetForm(loginFormRef)" size="large">
                        {{ $t('commons.button.reset') }}
                    </el-button>
                    <el-button round @click="login(loginFormRef)" size="large" type="primary" :loading="loading">
                        {{ $t('commons.button.login') }}
                    </el-button>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { Login } from '@/api/interface';
import type { ElForm } from 'element-plus';
import { ElMessage } from 'element-plus';
import { loginApi, getCaptcha, mfaLoginApi } from '@/api/modules/auth';
import { GlobalStore } from '@/store';
import { MenuStore } from '@/store/modules/menu';
import i18n from '@/lang';

const globalStore = GlobalStore();
const menuStore = MenuStore();

type FormInstance = InstanceType<typeof ElForm>;
const loginFormRef = ref<FormInstance>();
const loginRules = reactive({
    name: [{ required: true, message: i18n.global.t('commons.rule.username'), trigger: 'blur' }],
    password: [{ required: true, message: i18n.global.t('commons.rule.password'), trigger: 'blur' }],
});

const loginForm = reactive<Login.ReqLoginForm>({
    name: 'admin',
    password: '1Panel@2022',
    captcha: '',
    captchaID: '',
    authMethod: '',
});
const mfaLoginForm = reactive<Login.MFALoginForm>({
    name: 'admin',
    password: '1Panel@2022',
    secret: '',
    code: '',
    authMethod: '',
});

const captcha = reactive<Login.ResCaptcha>({
    captchaID: '',
    imagePath: '',
    captchaLength: 0,
});

const loading = ref<boolean>(false);
const mfaShow = ref<boolean>(false);

const router = useRouter();

const login = (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        try {
            const requestLoginForm: Login.ReqLoginForm = {
                name: loginForm.name,
                password: loginForm.password,
                captcha: loginForm.captcha,
                captchaID: captcha.captchaID,
                authMethod: '',
            };
            const res = await loginApi(requestLoginForm);
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

const resetForm = (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.resetFields();
};

const loginVerify = async () => {
    const res = await getCaptcha();
    captcha.imagePath = res.data.imagePath ? res.data.imagePath : '';
    captcha.captchaID = res.data.captchaID ? res.data.captchaID : '';
    captcha.captchaLength = res.data.captchaLength ? res.data.captchaLength : 0;
};
loginVerify();

onMounted(() => {
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
@import '../index.scss';

.vPicBox {
    display: flex;
    justify-content: space-between;
    width: 100%;
}
.vPic {
    width: 33%;
    height: 38px;
    background: #ccc;
    img {
        width: 100%;
        height: 100%;
        vertical-align: middle;
    }
}
</style>
