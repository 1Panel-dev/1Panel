<template>
    <el-form
        ref="loginFormRef"
        :model="loginForm"
        :rules="loginRules"
        size="large"
    >
        <el-form-item prop="username">
            <el-input
                v-model="loginForm.name"
                placeholder="用户名：admin / user"
            >
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
                placeholder="密码：123456"
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
                placeholder="请输入验证码"
                style="width: 60%"
            />
            <div class="vPic">
                <img
                    v-if="captcha.imagePath"
                    :src="captcha.imagePath"
                    alt="请输入验证码"
                    @click="loginVerify()"
                />
            </div>
        </div>
    </el-form-item>
    <div class="login-btn">
        <el-button round @click="resetForm(loginFormRef)" size="large"
            >重置</el-button
        >
        <el-button
            round
            @click="login(loginFormRef)"
            size="large"
            type="primary"
            :loading="loading"
        >
            登录
        </el-button>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { Login } from '@/api/interface';
import type { ElForm } from 'element-plus';
import { ElMessage } from 'element-plus';
import { loginApi, getCaptcha } from '@/api/modules/login';
import { GlobalStore } from '@/store';
import { MenuStore } from '@/store/modules/menu';

const globalStore = GlobalStore();
const menuStore = MenuStore();

// 定义 formRef（校验规则）
type FormInstance = InstanceType<typeof ElForm>;
const loginFormRef = ref<FormInstance>();
const loginRules = reactive({
    name: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
    password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
});

// 登录表单数据
const loginForm = reactive<Login.ReqLoginForm>({
    name: 'admin',
    password: 'Songliu123++',
    captcha: '',
    captchaID: '',
    authMethod: '',
});

const captcha = reactive<Login.ResCaptcha>({
    captchaID: '',
    imagePath: '',
    captchaLength: 0,
});

const loading = ref<boolean>(false);
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
            globalStore.setUserInfo(res.data.name);
            globalStore.setLogStatus(true);
            menuStore.setMenuList([]);
            ElMessage.success('登录成功！');
            router.push({ name: 'home' });
        } catch (error) {
            loginVerify();
        } finally {
            loading.value = false;
        }
    });
};

const resetForm = (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.resetFields();
};

const loginVerify = () => {
    getCaptcha().then(async (ele) => {
        captcha.imagePath = ele.data?.imagePath ? ele.data.imagePath : '';
        captcha.captchaID = ele.data?.captchaID ? ele.data.captchaID : '';
        captcha.captchaLength = ele.data?.captchaLength
            ? ele.data.captchaLength
            : 0;
    });
};
loginVerify();

onMounted(() => {
    // 监听enter事件（调用登录）
    document.onkeydown = (e: any) => {
        e = window.event || e;
        if (
            e.code === 'Enter' ||
            e.code === 'enter' ||
            e.code === 'NumpadEnter'
        ) {
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
