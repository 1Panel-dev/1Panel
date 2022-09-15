<template>
    <div>
        <div v-if="statusCode == 1" class="login-container flx-center">
            <div class="login-box">
                <div class="login-left">
                    <img src="@/assets/images/login_left0.png" alt="login" />
                </div>
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
import { loginStatus, entrance } from '@/api/modules/auth';

interface Props {
    code: string;
}
const mySafetyCode = withDefaults(defineProps<Props>(), {
    code: '',
});

const statusCode = ref<number>(0);

const getStatus = async () => {
    const res = await loginStatus();
    if (res.code === 402) {
        statusCode.value = -1;
    } else {
        statusCode.value = 1;
        return;
    }
    if (mySafetyCode.code) {
        const res = await entrance(mySafetyCode.code);
        if (res.code === 200) {
            statusCode.value = 1;
        }
    }
};

onMounted(() => {
    getStatus();
});
</script>

<style scoped lang="scss">
@import './index.scss';
</style>
