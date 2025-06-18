<template>
    <div class="flex items-center justify-center min-h-screen relative bg-gray-100">
        <div class="absolute inset-0 bg-cover bg-center bg-no-repeat" :style="backgroundStyle"></div>
        <div
            :style="{ opacity: backgroundOpacity, width: containerWidth, height: containerHeight }"
            class="bg-white shadow-lg relative z-10 border border-gray-200 flex overflow-hidden"
        >
            <div class="grid grid-cols-1 md:grid-cols-2 items-stretch w-full">
                <div v-if="showLogo" class="flex justify-center" :style="{ height: containerHeight }">
                    <img
                        v-show="imgLoaded"
                        :src="loadImage('loginImage')"
                        class="max-w-full max-h-full object-cover bg-cover bg-center"
                        alt="1panel"
                        @load="onImgLoad"
                        @error="onImgError"
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
import { preloadImage } from '@/utils/util';

const globalStore = GlobalStore();
const backgroundOpacity = ref(1);
const defaultLoginImage = new URL('@/assets/images/1panel-login.jpg', import.meta.url).href;
const defaultLoginBgImage = new URL('@/assets/images/1panel-login-bg.jpg', import.meta.url).href;
const loadedLoginImage = ref<string | null>(null);
const loadedBackgroundImage = ref<string | null>(null);
const backgroundStyle = ref<{ backgroundImage?: string; backgroundColor?: string }>({});
const imgLoaded = ref(false);

function onImgLoad() {
    imgLoaded.value = true;
}
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
    const { loginImage, loginBackground, loginBgType } = globalStore.themeConfig;
    if (name === 'loginImage') {
        return loginImage === 'loginImage' ? loadedLoginImage.value : defaultLoginImage;
    }
    if (name === 'loginBackground') {
        if (loginBgType === 'image') {
            return loginBackground === 'loginBackground' ? loadedBackgroundImage.value : defaultLoginBgImage;
        }
        if (loginBgType === 'color') {
            return loginBackground;
        }
        return defaultLoginBgImage;
    }
    return '';
};

const onImgError = (event: any) => {
    event.target.src = defaultLoginImage;
    imgLoaded.value = true;
};

onMounted(async () => {
    await getStatus();
    const loginImageUrl = `/api/v2/images/loginImage?t=${Date.now()}`;
    const backgroundImageUrl = `/api/v2/images/loginBackground?t=${Date.now()}`;
    loadedLoginImage.value = await preloadImage(loginImageUrl);
    loadedBackgroundImage.value = await preloadImage(backgroundImageUrl);
    if (globalStore.themeConfig.loginBgType === 'color') {
        backgroundStyle.value = {
            backgroundColor: globalStore.themeConfig.loginBackground,
        };
    } else {
        const img = new Image();
        const url = loadImage('loginBackground');
        img.onload = () => {
            backgroundStyle.value = {
                backgroundImage: `url(${url})`,
            };
        };
        img.onerror = () => {
            backgroundStyle.value = {
                backgroundImage: `url(${defaultLoginBgImage})`, // 你定义的默认图
            };
        };
        img.src = url;
    }
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
