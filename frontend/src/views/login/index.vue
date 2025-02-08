<template>
    <div class="flex items-center justify-center min-h-screen relative bg-gray-100">
        <div
            class="absolute inset-0 bg-cover bg-center bg-no-repeat"
            :style="{ backgroundImage: `url(${backgroundImage})` }"
        ></div>
        <div
            :style="{ opacity: backgroundOpacity }"
            class="w-[45%] min-h-[480px] bg-white rounded-lg shadow-lg relative z-10 border border-gray-200 flex overflow-hidden"
        >
            <div class="grid md:grid-cols-2 gap-4 items-stretch w-full">
                <div class="flex flex-col justify-center items-center w-full p-4">
                    <img :src="logoImage" class="max-w-full max-h-full object-contain" />
                </div>
                <div class="hidden md:block w-px bg-gray-200 absolute left-1/2 top-4 bottom-4"></div>
                <div class="hidden md:flex items-center justify-center p-4">
                    <LoginForm ref="loginRef"></LoginForm>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts" name="login">
import { checkIsSafety } from '@/api/modules/auth';
import LoginForm from './components/login-form.vue';
import { ref, onMounted } from 'vue';
import router from '@/routers';
import { GlobalStore } from '@/store';
import { getXpackSettingForTheme } from '@/utils/xpack';

const gStore = GlobalStore();
const loading = ref();
const backgroundOpacity = ref(0.8);
const backgroundImage = ref(new URL('', import.meta.url).href);
const logoImage = ref(new URL('@/assets/images/1panel-login.png', import.meta.url).href);

const mySafetyCode = defineProps({
    code: {
        type: String,
        default: '',
    },
});

const screenWidth = ref(null);

const getStatus = async () => {
    let code = mySafetyCode.code;
    if (code != '') {
        gStore.entrance = code;
    }
    loading.value = true;
    await checkIsSafety(gStore.entrance)
        .then((res) => {
            loading.value = false;
            if (res.data === 'unpass') {
                router.replace({ name: 'entrance', params: { code: gStore.entrance } });
                return;
            }
            getXpackSettingForTheme();
        })
        .catch(() => {
            loading.value = false;
        });
};

onMounted(() => {
    getStatus();
    screenWidth.value = document.body.clientWidth;
    window.onresize = () => {
        return (() => {
            screenWidth.value = document.body.clientWidth;
        })();
    };
});
</script>
