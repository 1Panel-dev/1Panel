<template>
    <div class="flex items-center justify-center min-h-screen relative bg-gray-100">
        <div class="absolute inset-0 bg-cover bg-center bg-no-repeat" :style="backgroundStyle"></div>
        <div
            :style="{ width: containerWidth, height: containerHeight }"
            class="bg-white shadow-lg relative z-10 border border-gray-200 flex overflow-hidden"
            id="login-container"
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
                <div :class="licenseFormClass">
                    <div v-loading="loading" class="w-full h-full flex items-center justify-center px-8">
                        <div class="w-full flex-grow flex flex-col login-form">
                            <div class="license-upload-section">
                                <div class="license-section-title">{{ $t('license.importLicense') }}</div>

                                <el-upload
                                    action="#"
                                    :auto-upload="false"
                                    ref="uploadRef"
                                    drag
                                    :limit="1"
                                    :on-change="fileOnChange"
                                    :on-exceed="handleExceed"
                                    v-model:file-list="uploaderFiles"
                                >
                                    <el-icon class="el-icon--upload"><upload-filled /></el-icon>
                                    <div class="el-upload__text">
                                        {{ $t('license.importHelper') }}
                                    </div>
                                </el-upload>
                            </div>

                            <div class="license-device-section">
                                <div class="license-device-header">
                                    <div class="license-helper-title">
                                        {{ $t('license.deviceID') }}
                                    </div>
                                    <span class="input-help license-helper-desc">
                                        {{ $t('license.licenseRequiredShortTip') }}
                                    </span>
                                </div>

                                <el-input
                                    v-model="licenseInfo.deviceID"
                                    readonly
                                    @focus="blurDeviceIDInput"
                                    @mousedown.prevent
                                >
                                    {{ licenseInfo.deviceID }}
                                    <template #append>
                                        <el-button
                                            class="copy-tip"
                                            @click="copyDeviceID"
                                            icon="DocumentCopy"
                                            link
                                            type="primary"
                                        >
                                            {{ $t('commons.button.copy') }}
                                        </el-button>
                                    </template>
                                </el-input>
                            </div>

                            <el-button
                                class="w-full login-button"
                                type="primary"
                                size="default"
                                :disabled="uploaderFiles.length === 0"
                                @click="submit"
                            >
                                {{ $t('commons.button.import') }}
                            </el-button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import type { UploadFile, UploadFiles, UploadInstance, UploadProps, UploadRawFile } from 'element-plus';
import { genFileId } from 'element-plus';
import { getEnterpriseLicense, uploadEnterpriseLicense } from '@/api/modules/setting';
import { getLoginSetting } from '@/api/modules/auth';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { MsgSuccess } from '@/utils/message';
import { preloadImage } from '@/utils/browser';
import { adjustColorToRGBA } from '@/utils/color';
import { getXpackSettingForTheme } from '@/utils/xpack';
import { copyText } from '@/utils/clipboard';
import i18n from '@/lang';

const router = useRouter();
const {
    entrance,
    isEnterprise,
    isEnterpriseLicenseLoaded,
    isEnterpriseLicensed,
    isFxplay,
    isIntl,
    isLogin,
    isMasterProductPro,
    isOffline,
    menuAccordion,
    openMenuTabs,
    themeConfig,
} = useGlobalStore();
const loading = ref(false);
const uploadRef = ref<UploadInstance>();
const uploaderFiles = ref<UploadFiles>([]);
const licenseInfo = reactive({
    deviceID: '',
});
const defaultLoginImage = new URL('@/assets/images/1panel-login-enterprise.png', import.meta.url).href;
const defaultLoginBgImage = new URL('@/assets/images/1panel-login-bg.jpg', import.meta.url).href;
const loadedLoginImage = ref<string | null>(null);
const loadedBackgroundImage = ref<string | null>(null);
const backgroundStyle = ref<{ backgroundImage?: string; backgroundColor?: string }>({});
const imgLoaded = ref(false);

const FIXED_WIDTH = 1000;
const FIXED_HEIGHT = 415;
const containerWidth = computed(() => `${FIXED_WIDTH}px`);
const containerHeight = computed(() => `${FIXED_HEIGHT}px`);
const width = ref(window.innerWidth);
const showLogo = computed(() => width.value >= FIXED_WIDTH);
const licenseFormClass = computed(() => {
    return showLogo.value
        ? 'hidden md:flex items-center justify-center p-4'
        : 'flex items-center justify-center p-4 w-full';
});

const updateSize = () => {
    width.value = window.innerWidth;
};

function onImgLoad() {
    imgLoaded.value = true;
}

const onImgError = (event: Event) => {
    const target = event.target as HTMLImageElement;
    target.src = defaultLoginImage;
    imgLoaded.value = true;
};

const loadImage = (name: string) => {
    const { loginImage, loginBackground, loginBgType } = themeConfig.value;
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

const applyLoginThemeColor = () => {
    const loginBtnLinkColor = themeConfig.value.loginBtnLinkColor || '#005eeb';
    document.documentElement.style.setProperty('--login-btn-link-color', loginBtnLinkColor);
    document.documentElement.style.setProperty(
        '--login-btn-link-hover-color',
        adjustColorToRGBA(loginBtnLinkColor, -10, 80),
    );
    document.documentElement.style.setProperty(
        '--login-loading-mask-color',
        adjustColorToRGBA(loginBtnLinkColor, 30, 15),
    );
};

const loadLoginTheme = async () => {
    const res = await getLoginSetting();
    isEnterprise.value = res.data.isEnterprise;
    isEnterpriseLicenseLoaded.value = !res.data.isEnterprise;
    isIntl.value = res.data.isIntl;
    isFxplay.value = res.data.isFxplay;
    isOffline.value = res.data.isOffline;
    openMenuTabs.value = res.data.menuTabs === 'Enable';
    menuAccordion.value = res.data.menuAccordion === 'Enable';
    themeConfig.value = {
        ...themeConfig.value,
        theme: res.data.theme,
        panelName: res.data.panelName,
    };
    document.title = res.data.panelName;
    applyLoginThemeColor();
    if (!isEnterprise.value) {
        router.replace(isLogin.value ? { name: 'home' } : { name: 'entrance', params: { code: entrance.value } });
        return false;
    }
    return true;
};

const loadXpackLoginTheme = async () => {
    try {
        await getXpackSettingForTheme();
        applyLoginThemeColor();
    } catch {
        applyLoginThemeColor();
    }
};

const loadBackground = async () => {
    const loginImageUrl = `/api/v2/images/loginImage?t=${Date.now()}`;
    const backgroundImageUrl = `/api/v2/images/loginBackground?t=${Date.now()}`;
    [loadedLoginImage.value, loadedBackgroundImage.value] = await Promise.all([
        preloadImage(loginImageUrl),
        preloadImage(backgroundImageUrl),
    ]);
    if (themeConfig.value.loginBgType === 'color') {
        backgroundStyle.value = {
            backgroundColor: themeConfig.value.loginBackground,
        };
        return;
    }
    const img = new Image();
    const url = loadImage('loginBackground');
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

const loadLicenseInfo = async () => {
    const res = await getEnterpriseLicense();
    isEnterpriseLicenseLoaded.value = true;
    licenseInfo.deviceID = res.data.deviceID;
    if (res.data.status === 'Bound') {
        isEnterpriseLicensed.value = true;
        router.replace({ name: 'home' });
        return;
    }
    isEnterpriseLicensed.value = false;
};

const fileOnChange = (_uploadFile: UploadFile, uploadFiles: UploadFiles) => {
    uploaderFiles.value = uploadFiles;
};

const handleExceed: UploadProps['onExceed'] = (files) => {
    uploadRef.value?.clearFiles();
    const file = files[0] as UploadRawFile;
    file.uid = genFileId();
    uploadRef.value?.handleStart(file);
};

const copyDeviceID = () => {
    if (!licenseInfo.deviceID) {
        return;
    }
    copyText(licenseInfo.deviceID);
};

const blurDeviceIDInput = (event: FocusEvent) => {
    (event.target as HTMLInputElement).blur();
};

const submit = async () => {
    if (uploaderFiles.value.length !== 1) {
        return;
    }
    const file = uploaderFiles.value[0];
    if (!file.raw) {
        return;
    }
    const formData = new FormData();
    formData.append('file', file.raw);
    loading.value = true;
    await uploadEnterpriseLicense(formData)
        .then(() => {
            isEnterpriseLicensed.value = true;
            isMasterProductPro.value = true;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            router.replace({ name: 'home' });
        })
        .finally(() => {
            loading.value = false;
            uploadRef.value?.clearFiles();
            uploaderFiles.value = [];
        });
};

onMounted(async () => {
    window.addEventListener('resize', updateSize);
    loading.value = true;
    try {
        const shouldLoadLicense = await loadLoginTheme();
        if (!shouldLoadLicense) {
            return;
        }
        await loadXpackLoginTheme();
        await loadBackground();
        await loadLicenseInfo();
    } finally {
        loading.value = false;
    }
});

onUnmounted(() => {
    window.removeEventListener('resize', updateSize);
});
</script>

<style lang="scss" scoped>
#login-container {
    border-radius: 0;
}

.login-form {
    max-width: 560px;
    gap: 16px;
    transform: translateY(-8px);
    font-size: 13px;

    .license-upload-section,
    .license-device-section {
        display: flex;
        min-width: 0;
        flex-direction: column;
    }

    .license-upload-section {
        gap: 10px;
    }

    .license-section-title {
        color: var(--el-text-color-primary);
        font-size: 24px;
        font-weight: 500;
        line-height: 32px;
    }

    .login-button {
        background-color: var(--login-btn-link-color, #005eeb);
        border-color: var(--login-btn-link-color, #005eeb);
        color: #ffffff;

        &:hover {
            background-color: var(--login-btn-link-hover-color, #0054d3) !important;
            border-color: var(--login-btn-link-hover-color, #0054d3) !important;
            outline: none !important;
        }
    }

    .license-device-section {
        gap: 8px;
        min-width: 0;
        font-size: 13px;
        line-height: 20px;

        .license-device-header {
            display: flex;
            flex-direction: column;
            gap: 2px;
        }

        .license-helper-title {
            color: var(--el-text-color-primary);
            font-size: 14px;
            font-weight: 500;
            line-height: 22px;
        }

        .license-helper-desc {
            display: block;
            margin: 0;
        }

        :deep(.el-input) {
            min-width: 0;
        }

        :deep(.el-input__wrapper) {
            min-width: 0;
            border: 1px dashed var(--el-border-color);
            box-shadow: none;
        }

        :deep(.el-input__inner) {
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
        }
    }

    .copy-tip {
        flex-shrink: 0;
        width: 80px;
        height: 20px;
        padding: 0;
        color: var(--login-btn-link-color, #005eeb);

        &:hover {
            color: var(--login-btn-link-hover-color, #0054d3);
        }
    }
}

.login-button {
    height: 36px;
    font-size: 13px;
}

:deep(.el-upload-dragger) {
    padding: 18px 12px;
}

:deep(.el-upload) {
    width: 100%;
}

:deep(.el-icon--upload) {
    margin-bottom: 4px;
    font-size: 30px;
}

:deep(.el-upload__text) {
    font-size: 13px;
    line-height: 20px;
}

:deep(.el-loading-mask) {
    background-color: var(--login-loading-mask-color) !important;

    .el-loading-spinner .path {
        stroke: var(--login-btn-link-color);
    }
}
</style>
