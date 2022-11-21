<template>
    <div>
        <div v-if="statusCode == 1">
            <div class="login-background">
                <div class="login-container">
                    <el-row type="flex">
                        <el-col :span="12">
                            <LoginForm ref="loginRef"></LoginForm>
                        </el-col>
                        <el-col :span="12">
                            <img class="login-image" src="@/assets/images/1Panel-login.jpg" alt="login" />
                        </el-col>
                    </el-row>
                </div>
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
// import { loginStatus, entrance } from '@/api/modules/auth';

// interface Props {
//     code: string;
// }
// const mySafetyCode = withDefaults(defineProps<Props>(), {
//     code: '',
// });

const statusCode = ref<number>(0);

const getStatus = async () => {
    statusCode.value = 1;
    // const res = await loginStatus();
    // if (res.code === 402) {
    //     statusCode.value = -1;
    // } else {
    //     statusCode.value = 1;
    //     return;
    // }
    // if (mySafetyCode.code) {
    //     const res = await entrance(mySafetyCode.code);
    //     if (res.code === 200) {
    //         statusCode.value = 1;
    //     }
    // }
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
.login-background {
    height: 100%;
    display: flex;
    @include login-center;
}

.login-container {
    margin-top: 150px;
    min-width: 640px;
    width: 960px;
    height: 480px;
    background-color: #ffffff;
    box-shadow: 0 5px 5px -3px rgba(0, 0, 0, 0.2), 0 8px 10px 1px rgba(0, 0, 0, 0.14),
        0 3px 14px 2px rgba(0, 0, 0, 0.12);
    @media only screen and (max-width: 1280px) {
        width: 900px;
        height: 380px;
    }
    .login-image {
        width: 480px;
        height: 480px;
        @media only screen and (max-width: 1280px) {
            height: 380px;
        }
    }
}
</style>
