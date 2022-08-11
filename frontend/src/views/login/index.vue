<template>
    <div class="login-container flx-center">
        <SwitchDark class="dark"></SwitchDark>
        <div class="login-box">
            <div class="login-left">
                <img src="@/assets/images/login_left0.png" alt="login" />
            </div>
            <div class="login-form">
                <div class="login-logo">
                    <img class="login-icon" src="@/assets/images/logo.svg" alt="" />
                    <h2 class="logo-text">1Panel</h2>
                </div>
                <LoginForm ref="loginRef"></LoginForm>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts" name="login">
import SwitchDark from '@/components/SwitchDark/index.vue';
import LoginForm from './components/LoginForm.vue';
</script>

<style scoped lang="scss">
@import './index.scss';
</style>
