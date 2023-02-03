<template>
    <div>
        <div class="login-backgroud" v-if="statusCode == 1">
            <div class="login-title">
                <span>1Panel 现代化的 Linux 面板</span>
            </div>
            <div class="login-container">
                <LoginForm ref="loginRef"></LoginForm>
            </div>
        </div>
        <div style="margin-left: 50px" v-if="statusCode == -1">
            <h1>{{ $t('commons.login.safeEntrance') }}</h1>
            <div style="line-height: 30px">
                <span style="font-weight: 500">{{ $t('commons.login.reason') }}</span>
                <span>
                    {{ $t('commons.login.reasonHelper') }}
                </span>
            </div>
            <div style="line-height: 30px">
                <span style="font-weight: 500">{{ $t('commons.login.solution') }}</span>
                <span>{{ $t('commons.login.solutionHelper') }}</span>
            </div>
            <div style="line-height: 30px">
                <span style="color: red">
                    {{ $t('commons.login.warnning') }}
                </span>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts" name="login">
import LoginForm from './components/login-form.vue';
import { ref, onMounted } from 'vue';

const statusCode = ref<number>(0);

const getStatus = async () => {
    statusCode.value = 1;
};

onMounted(() => {
    getStatus();
});
</script>

<style scoped lang="scss">
@mixin login-center {
    display: flex;
    justify-content: center;
    align-items: center;
}
.login-backgroud {
    height: 100vh;
    display: flex;
    // @include login-center;
    background-image: url(@/assets/images/1Panel-login.png);
    background-color: $menu-backgroup-color;
    background-position: 0% 40%;
    background-size: 70%;
    background-repeat: no-repeat;

    .login-title {
        margin-top: 10%;
        text-align: center;
        margin-left: 15%;
        span:first-child {
            color: $primary-color;
            font-size: 48px;
            font-family: pingFangSC-Regular;
        }
    }
}

.login-container {
    margin-top: 19%;
    margin-left: 15%;
    width: 394px;
    height: 422px;
    background-color: rgba(255, 255, 255, 0.55);
    border-radius: 4px;

    box-shadow: 2px 4px 22px rgba(0, 94, 235, 0.2);
    @media only screen and (max-width: 1280px) {
        width: 900px;
        height: 380px;
    }
}
</style>
