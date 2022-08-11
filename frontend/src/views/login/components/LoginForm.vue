<template>
    <el-form ref="loginFormRef" :model="loginForm" :rules="loginRules" size="large">
        <el-form-item prop="username">
            <el-input v-model="loginForm.username" placeholder="用户名：admin / user">
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
    <div class="login-btn">
        <el-button :icon="CircleClose" round @click="resetForm(loginFormRef)" size="large">重置</el-button>
        <el-button :icon="UserFilled" round @click="login(loginFormRef)" size="large" type="primary" :loading="loading">
            登录
        </el-button>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { Login, getCaptcha } from '@/api/interface';
import type { ElForm } from 'element-plus';
import { ElMessage } from 'element-plus';
import { loginApi } from '@/api/modules/login';
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
    name: '',
    password: '',
});

const loading = ref<boolean>(false);
const router = useRouter();
// login
const login = (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.validate(async (valid) => {
        if (!valid) return;
        loading.value = true;
        try {
            const requestLoginForm: Login.ReqLoginForm = {
                name: loginForm.username,
                password: loginForm.password,
            };
            const res = await loginApi(requestLoginForm);
            // * 存储 token
            globalStore.setToken(res.data!.access_token);
            // * 登录成功之后清除上个账号的 menulist 和 tabs 数据
            menuStore.setMenuList([]);

            ElMessage.success('登录成功！');
            router.push({ name: 'home' });
        } finally {
            loading.value = false;
        }
    });
};

// resetForm
const resetForm = (formEl: FormInstance | undefined) => {
    if (!formEl) return;
    formEl.resetFields();
};

onMounted(() => {
    // 监听enter事件（调用登录）
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
</style>
