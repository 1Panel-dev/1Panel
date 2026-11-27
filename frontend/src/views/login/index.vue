<template>
    <div class="flex items-center justify-center min-h-screen relative bg-gray-100">
        <div class="absolute inset-0 bg-cover bg-center bg-no-repeat" :style="backgroundStyle"></div>
        <div
            v-if="externalLoginPending"
            v-loading="true"
            class="absolute inset-0 z-20 bg-gray-100"
            aria-busy="true"
        ></div>
        <div
            v-show="!externalLoginPending"
            :style="{ opacity: backgroundOpacity, width: containerWidth, height: containerHeight }"
            class="bg-white shadow-lg relative z-10 border border-gray-200 flex overflow-hidden"
            id="login-container"
        >
            <div class="grid items-stretch w-full" :style="loginGridStyle">
                <div
                    v-if="showLogo"
                    class="relative flex justify-center overflow-hidden"
                    :style="{ height: containerHeight }"
                    :aria-busy="imageStatus === 'loading'"
                >
                    <Transition name="login-image-placeholder">
                        <div
                            v-if="imageStatus !== 'loaded'"
                            class="absolute inset-0 flex flex-col items-center justify-center gap-2 px-8 text-center"
                            role="status"
                        >
                            <el-icon
                                class="login-image-loading"
                                :class="{ 'is-loading': imageStatus === 'loading' }"
                                :size="18"
                                color="var(--el-text-color-placeholder)"
                                aria-hidden="true"
                            >
                                <Loading v-if="imageStatus === 'loading'" />
                                <Picture v-else />
                            </el-icon>
                            <template v-if="imageStatus === 'error'">
                                <span class="text-sm text-[var(--el-text-color-secondary)]">
                                    {{ $t('commons.login.imageLoadError') }}
                                </span>
                                <span class="text-xs text-[var(--el-text-color-placeholder)] max-w-80">
                                    {{ $t('commons.login.imageLoadErrorHint') }}
                                </span>
                            </template>
                        </div>
                    </Transition>
                    <img
                        v-if="loginImageSrc"
                        :src="loginImageSrc"
                        :style="{ opacity: imageStatus === 'loaded' ? 1 : 0 }"
                        class="login-image relative max-w-full max-h-full object-cover bg-cover bg-center"
                        alt="1panel"
                        @load="onImgLoad"
                        @error="imageStatus = 'error'"
                    />
                </div>
                <div :class="loginFormClass">
                    <LoginForm
                        ref="loginRef"
                        @external-login-ready="externalLoginPending = false"
                        @theme-ready="loadImages"
                    ></LoginForm>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import LoginForm from './components/login-form.vue';
import { ref, onMounted } from 'vue';
import { Loading, Picture } from '@element-plus/icons-vue';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { hasExternalLoginTicket } from '@/utils/external-login';
defineOptions({ name: 'Login' });
const { entrance, isEnterprise, themeConfig } = useGlobalStore();
const backgroundOpacity = ref(1);
const defaultLoginImage = new URL('@/assets/images/1panel-login.jpg', import.meta.url).href;
const defaultEnterpriseLoginImage = new URL('@/assets/images/1panel-login-enterprise.png', import.meta.url).href;
const defaultLoginBgImage = new URL('@/assets/images/1panel-login-bg.jpg', import.meta.url).href;
const loginImageSrc = ref('');
const backgroundStyle = ref<{ backgroundImage?: string; backgroundColor?: string }>({});
const imageStatus = ref<'loading' | 'loaded' | 'error'>('loading');
const currentDefaultLoginImage = computed(() => (isEnterprise.value ? defaultEnterpriseLoginImage : defaultLoginImage));

const externalLoginPending = ref(hasExternalLoginTicket());

function onImgLoad() {
    imageStatus.value = 'loaded';
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
        entrance.value = code;
    }
};

const loadImages = () => {
    const { loginImage, loginBackground, loginBgType } = themeConfig.value;
    imageStatus.value = 'loading';
    loginImageSrc.value =
        loginImage === 'loginImage'
            ? `/api/v2/images/loginImage?t=${Date.now()}`
            : loginImage || currentDefaultLoginImage.value;
    if (loginBgType === 'color') {
        backgroundStyle.value = {
            backgroundColor: loginBackground,
        };
        return;
    }
    let url = defaultLoginBgImage;
    if (loginBgType === 'image' && loginBackground) {
        url =
            loginBackground === 'loginBackground' ? `/api/v2/images/loginBackground?t=${Date.now()}` : loginBackground;
    }
    const img = new Image();
    img.onload = () => {
        backgroundStyle.value = {
            backgroundImage: `url(${url})`,
        };
    };
    img.onerror = () => {
        backgroundStyle.value = {
            backgroundImage: `url(${defaultLoginBgImage})`,
        };
    };
    img.src = url;
};

onMounted(getStatus);

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
const loginGridStyle = computed(() => ({
    gridTemplateColumns: showLogo.value ? 'repeat(2, minmax(0, 1fr))' : 'minmax(0, 1fr)',
}));
const loginFormClass = computed(() => {
    return showLogo.value
        ? 'flex items-center justify-center p-4 min-w-0'
        : 'flex items-center justify-center p-4 w-full min-w-0';
});
</script>

<style scoped lang="scss">
.login-image {
    transition: opacity 200ms ease;
}

.login-image-loading {
    opacity: 0.5;
}

.login-image-placeholder-leave-active {
    transition: opacity 200ms ease;
}

.login-image-placeholder-leave-to {
    opacity: 0;
}

@media (prefers-reduced-motion: reduce) {
    .login-image,
    .login-image-placeholder-leave-active {
        transition: none;
    }

    .login-image-loading {
        animation: none;
    }
}
</style>
