<template>
    <div class="flex items-center justify-center min-h-screen relative bg-gray-100">
        <div
            class="absolute inset-0 bg-cover bg-center bg-no-repeat"
            :style="
                globalStore.themeConfig.loginBgType === 'color'
                    ? { backgroundColor: globalStore.themeConfig.loginBackground }
                    : { backgroundImage: `url(${loadImage('loginBackground')})` }
            "
        ></div>
        <div
            :style="{ opacity: backgroundOpacity, width: containerWidth, height: containerHeight }"
            class="bg-white shadow-lg relative z-10 border border-gray-200 flex overflow-hidden"
        >
            <div class="grid grid-cols-1 md:grid-cols-2 items-stretch w-full">
                <div v-if="showLogo" class="flex justify-center" :style="{ height: containerHeight }">
                    <img
                        :src="loadImage('loginImage')"
                        class="max-w-full max-h-full object-cover bg-cover bg-center"
                        alt="1panel"
                    />
                </div>
                <div :class="loginFormClass">
                    <LoginForm ref="loginRef"></LoginForm>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts" name="login">
import LoginForm from './components/login-form.vue';
import { ref, onMounted } from 'vue';
import { GlobalStore } from '@/store';

const globalStore = GlobalStore();
const backgroundOpacity = ref(1);

const mySafetyCode = defineProps({
    code: {
        type: String,
        default: '',
    },
});

const getStatus = async () => {
    let code = mySafetyCode.code;
    if (code != '') {
        globalStore.entrance = code;
    }
};

const loadImage = (name: string) => {
    switch (name) {
        case 'loginImage':
            if (globalStore.themeConfig.loginImage === 'loginImage') {
                return `/api/v2/images/loginImage?t=${Date.now()}`;
            }
            return new URL('@/assets/images/1panel-login.jpg', import.meta.url).href;
        case 'loginBackground':
            if (globalStore.themeConfig.loginBgType === 'image') {
                if (globalStore.themeConfig.loginBackground === 'loginBackground') {
                    return `/api/v2/images/loginBackground?t=${Date.now()}`;
                }
                return new URL('@/assets/images/1panel-login-bg.jpg', import.meta.url).href;
            } else if (globalStore.themeConfig.loginBgType === 'color') {
                return globalStore.themeConfig.loginBackground;
            } else {
                return new URL('@/assets/images/1panel-login-bg.jpg', import.meta.url).href;
            }
    }
    return '';
};

onMounted(() => {
    getStatus();
});

const FIXED_WIDTH = 1000;
const FIXED_HEIGHT = 415;
const useWindowSize = () => {
    const width = ref(window.innerWidth);
    const height = ref(window.innerHeight);

    const updateSize = () => {
        width.value = window.innerWidth;
        height.value = window.innerHeight;
    };

    onMounted(() => window.addEventListener('resize', updateSize));
    onUnmounted(() => window.removeEventListener('resize', updateSize));

    return { width, height };
};
const { width } = useWindowSize();
const showLogo = computed(() => width.value >= FIXED_WIDTH);
const containerWidth = computed(() => `${FIXED_WIDTH}px`);
const containerHeight = computed(() => `${FIXED_HEIGHT}px`);
const loginFormClass = computed(() => {
    return showLogo.value
        ? 'hidden md:flex items-center justify-center p-4'
        : 'flex items-center justify-center p-4 w-full';
});
</script>
