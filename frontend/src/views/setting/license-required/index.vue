<template>
    <div class="flex items-center justify-center min-h-screen relative bg-gray-100">
        <div class="absolute inset-0 bg-cover bg-center bg-no-repeat" :style="backgroundStyle"></div>
        <div
            :style="{ opacity: backgroundOpacity, width: containerWidth, height: containerHeight }"
            class="bg-white shadow-lg relative z-10 border border-gray-200 flex overflow-hidden"
            id="login-container"
        >
            <div v-loading="loading" class="w-full h-full flex items-center justify-center px-8">
                <div class="w-full flex-grow flex flex-col login-form">
                    <div>
                        <div class="text-2xl font-medium text-gray-900">
                            {{ $t('license.importLicense') }}
                        </div>
                        <div class="license-helper">
                            {{ $t('license.licenseRequiredTip') }}
                        </div>
                    </div>

                    <el-descriptions :column="1" border>
                        <el-descriptions-item :label="$t('license.deviceID')">
                            <span class="device-id">{{ licenseInfo.deviceID }}</span>
                            <CopyButton :content="licenseInfo.deviceID" />
                        </el-descriptions-item>
                    </el-descriptions>

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
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import type { UploadFile, UploadFiles, UploadInstance, UploadProps, UploadRawFile } from 'element-plus';
import { genFileId } from 'element-plus';
import { getEnterpriseLicense, uploadEnterpriseLicense } from '@/api/modules/setting';
import { getLoginSetting } from '@/api/modules/auth';
import { GlobalStore } from '@/store';
import { MsgSuccess } from '@/utils/message';
import { preloadImage } from '@/utils/browser';
import { adjustColorToRGBA } from '@/utils/color';
import i18n from '@/lang';

const router = useRouter();
const globalStore = GlobalStore();
const loading = ref(false);
const backgroundOpacity = ref(1);
const loginBtnLinkColor = ref('#005eeb');
const uploadRef = ref<UploadInstance>();
const uploaderFiles = ref<UploadFiles>([]);
const licenseInfo = reactive({
    deviceID: '',
});
const defaultLoginBgImage = new URL('@/assets/images/1panel-login-bg.jpg', import.meta.url).href;
const loadedBackgroundImage = ref<string | null>(null);
const backgroundStyle = ref<{ backgroundImage?: string; backgroundColor?: string }>({});

const FIXED_WIDTH = 1000;
const FIXED_HEIGHT = 415;
const containerWidth = computed(() => `${FIXED_WIDTH}px`);
const containerHeight = computed(() => `${FIXED_HEIGHT}px`);

const loadImage = (name: string) => {
    const { loginBackground, loginBgType } = globalStore.themeConfig;
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

const loadLoginTheme = async () => {
    const res = await getLoginSetting();
    globalStore.isEnterprise = res.data.isEnterprise;
    globalStore.isEnterpriseLicenseLoaded = !res.data.isEnterprise;
    globalStore.isIntl = res.data.isIntl;
    globalStore.isFxplay = res.data.isFxplay;
    globalStore.isOffLine = res.data.isOffLine;
    globalStore.openMenuTabs = res.data.menuTabs === 'Enable';
    globalStore.themeConfig = {
        ...globalStore.themeConfig,
        theme: res.data.theme,
        panelName: res.data.panelName,
    };
    document.title = res.data.panelName;
    loginBtnLinkColor.value = globalStore.themeConfig.loginBtnLinkColor || '#005eeb';
    document.documentElement.style.setProperty('--login-btn-link-color', loginBtnLinkColor.value);
    document.documentElement.style.setProperty(
        '--login-btn-link-hover-color',
        adjustColorToRGBA(loginBtnLinkColor.value, -10, 80),
    );
    document.documentElement.style.setProperty(
        '--login-loading-mask-color',
        adjustColorToRGBA(loginBtnLinkColor.value, 30, 15),
    );
    if (!globalStore.isEnterprise) {
        router.replace(
            globalStore.isLogin ? { name: 'home' } : { name: 'entrance', params: { code: globalStore.entrance } },
        );
    }
};

const loadBackground = async () => {
    const backgroundImageUrl = `/api/v2/images/loginBackground?t=${Date.now()}`;
    loadedBackgroundImage.value = await preloadImage(backgroundImageUrl);
    if (globalStore.themeConfig.loginBgType === 'color') {
        backgroundStyle.value = {
            backgroundColor: globalStore.themeConfig.loginBackground,
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
    globalStore.isEnterpriseLicenseLoaded = true;
    licenseInfo.deviceID = res.data.deviceID;
    if (res.data.status === 'Bound') {
        globalStore.isEnterpriseLicensed = true;
        router.replace({ name: 'home' });
    }
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
            globalStore.isEnterpriseLicensed = true;
            globalStore.isMasterProductPro = true;
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
    loading.value = true;
    try {
        await loadLoginTheme();
        await loadBackground();
        await loadLicenseInfo();
    } finally {
        loading.value = false;
    }
});
</script>

<style lang="scss" scoped>
#login-container {
    border-radius: 0;
}

.login-form {
    max-width: 640px;
    gap: 18px;

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

    .license-helper {
        margin-top: 8px;
        color: var(--el-text-color-secondary);
        line-height: 22px;
    }
}

.device-id {
    word-break: break-all;
}

.login-button {
    height: 40px;
}

:deep(.el-upload-dragger) {
    padding: 22px 16px;
}

:deep(.el-loading-mask) {
    background-color: var(--login-loading-mask-color) !important;

    .el-loading-spinner .path {
        stroke: var(--login-btn-link-color);
    }
}
</style>
