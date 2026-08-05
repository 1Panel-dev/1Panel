<template>
    <div class="license-required-page flex items-center justify-center min-h-screen relative bg-gray-100">
        <div class="absolute inset-0 bg-cover bg-center bg-no-repeat" :style="backgroundStyle"></div>
        <div
            :style="{ width: containerWidth, height: containerHeight }"
            class="bg-white shadow-lg relative z-10 border border-gray-200 flex overflow-hidden"
            id="login-container"
        >
            <el-button
                class="community-restore-link"
                type="primary"
                link
                :loading="restoring"
                :disabled="restoring"
                @click="communityRestoreRef?.acceptParams()"
            >
                {{ $t('license.restoreCommunity') }}
            </el-button>
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

        <CommunityRestoreDialog ref="communityRestoreRef" @started="handleCommunityRestoreStarted" />

        <Transition name="restore-mask">
            <div v-if="restoring" class="restore-mask" role="status" aria-live="polite" aria-busy="true">
                <div class="restore-status-card">
                    <div class="restore-spinner" aria-hidden="true">
                        <span></span>
                    </div>
                    <div class="restore-status-title">{{ $t('license.restoreCommunity') }}</div>
                    <div class="restore-status-description">
                        {{ $t('license.restoreCommunityStarting') }}
                    </div>
                </div>
            </div>
        </Transition>
    </div>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue';
import { useRouter } from 'vue-router';
import type { UploadFile, UploadFiles, UploadInstance, UploadProps, UploadRawFile } from 'element-plus';
import { genFileId } from 'element-plus';
import { getCommunityRestoreStatus, getEnterpriseLicense, uploadEnterpriseLicense } from '@/api/modules/setting';
import { getLoginSetting } from '@/api/modules/auth';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { MsgError, MsgSuccess } from '@/utils/message';
import { preloadImage } from '@/utils/browser';
import { adjustColorToRGBA } from '@/utils/color';
import { getXpackSettingForTheme } from '@/utils/xpack';
import { copyText } from '@/utils/clipboard';
import i18n from '@/lang';
import CommunityRestoreDialog from './community-restore/index.vue';

const router = useRouter();
const {
    globalStore,
    entrance,
    isEnterprise,
    isEnterpriseLicenseLoaded,
    isEnterpriseLicensed,
    isFxplay,
    isIntl,
    isLogin,
    isMasterProductPro,
    isOffline,
    isOnRestart,
    menuAccordion,
    openMenuTabs,
    themeConfig,
} = useGlobalStore();
const loading = ref(false);
const restoring = ref(isOnRestart.value);
const communityRestoreRef = ref<InstanceType<typeof CommunityRestoreDialog>>();
const uploadRef = ref<UploadInstance>();
const uploaderFiles = ref<UploadFiles>([]);
const licenseInfo = reactive({
    deviceID: '',
});
let restoreTimer: number | undefined;
const defaultLoginImage = new URL('@/assets/images/1panel-login-enterprise.png', import.meta.url).href;
const defaultLoginBgImage = new URL('@/assets/images/1panel-login-bg.jpg', import.meta.url).href;
const loadedLoginImage = ref<string | null>(null);
const loadedBackgroundImage = ref<string | null>(null);
const backgroundStyle = ref<{ backgroundImage?: string; backgroundColor?: string }>({});
const imgLoaded = ref(false);

const FIXED_WIDTH = 1000;
const FIXED_HEIGHT = 415;
const RESTORE_POLL_INTERVAL = 10_000;
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

const loadRestoreInfo = async () => {
    const res = await getCommunityRestoreStatus();
    restoring.value = res.data.state === 'Running';
    isOnRestart.value = restoring.value;
};

const scheduleRestorePoll = () => {
    window.clearTimeout(restoreTimer);
    restoreTimer = window.setTimeout(pollRestoreStatus, RESTORE_POLL_INTERVAL);
};

const redirectToCommunityLogin = () => {
    restoring.value = false;
    isOnRestart.value = false;
    globalStore.setLogStatus(false);
    globalStore.clearAuthInfo();
    isEnterprise.value = false;
    isEnterpriseLicensed.value = false;
    isEnterpriseLicenseLoaded.value = true;
    window.location.replace(entrance.value ? `/${entrance.value}` : '/');
};

const waitForCommunityService = async () => {
    try {
        const response = await fetch(`${import.meta.env.VITE_API_URL}/core/auth/setting`, {
            credentials: 'include',
            cache: 'no-store',
        });
        if (response.ok) {
            const result = await response.json();
            if (result?.data?.isEnterprise === false) {
                redirectToCommunityLogin();
                return;
            }
        }
    } catch {
        // The core service is expected to be temporarily unavailable while binaries are switched.
    }
    scheduleRestorePoll();
};

const pollRestoreStatus = async () => {
    try {
        const res = await getCommunityRestoreStatus();
        if (res.data.state === 'Failed') {
            restoring.value = false;
            isOnRestart.value = false;
            MsgError(res.data.message);
            return;
        }
        restoring.value = res.data.state === 'Running';
        isOnRestart.value = restoring.value;
        if (restoring.value) {
            scheduleRestorePoll();
        }
    } catch {
        await waitForCommunityService();
    }
};

const handleCommunityRestoreStarted = () => {
    restoring.value = true;
    isOnRestart.value = true;
    scheduleRestorePoll();
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
        try {
            await loadRestoreInfo();
        } catch {
            if (restoring.value) {
                await waitForCommunityService();
                return;
            }
        }
        if (restoring.value) {
            scheduleRestorePoll();
            return;
        }

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
    window.clearTimeout(restoreTimer);
});
</script>

<style lang="scss" scoped>
#login-container {
    border-radius: 0;
}

.license-required-page {
    isolation: isolate;
}

.community-restore-link {
    position: absolute;
    top: 16px;
    right: 20px;
    z-index: 1;
    color: var(--login-btn-link-color, #005eeb);

    &:hover {
        color: var(--login-btn-link-hover-color, #0054d3);
    }
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

.restore-mask {
    position: fixed;
    z-index: 2000;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
    background: rgb(15 23 42 / 48%);
    backdrop-filter: blur(8px) saturate(90%);
}

.restore-status-card {
    display: flex;
    width: min(400px, 100%);
    flex-direction: column;
    align-items: center;
    padding: 32px 36px;
    border: 1px solid rgb(255 255 255 / 70%);
    border-radius: 16px;
    background: rgb(255 255 255 / 96%);
    box-shadow:
        0 24px 64px rgb(15 23 42 / 24%),
        0 4px 16px rgb(15 23 42 / 10%);
    text-align: center;
}

.restore-spinner {
    position: relative;
    width: 54px;
    height: 54px;
    margin-bottom: 22px;

    &::before,
    &::after {
        position: absolute;
        inset: 0;
        border: 3px solid rgb(148 163 184 / 22%);
        border-radius: 50%;
        content: '';
    }

    &::after {
        border-color: var(--login-btn-link-color, #005eeb) transparent transparent;
        animation: restore-spin 0.9s linear infinite;
    }

    span {
        position: absolute;
        inset: 19px;
        border-radius: 50%;
        background: var(--login-btn-link-color, #005eeb);
        box-shadow: 0 0 0 6px color-mix(in srgb, var(--login-btn-link-color, #005eeb) 10%, transparent);
        animation: restore-pulse 1.5s ease-in-out infinite;
    }
}

.restore-status-title {
    color: var(--el-text-color-primary, #1f2937);
    font-size: 18px;
    font-weight: 600;
    line-height: 26px;
}

.restore-status-description {
    max-width: 320px;
    margin-top: 8px;
    color: var(--el-text-color-secondary, #64748b);
    font-size: 14px;
    line-height: 22px;
}

.restore-mask-enter-active,
.restore-mask-leave-active {
    transition: opacity 0.22s ease;

    .restore-status-card {
        transition:
            opacity 0.22s ease,
            transform 0.22s ease;
    }
}

.restore-mask-enter-from,
.restore-mask-leave-to {
    opacity: 0;

    .restore-status-card {
        opacity: 0;
        transform: translateY(8px) scale(0.98);
    }
}

@keyframes restore-spin {
    to {
        transform: rotate(360deg);
    }
}

@keyframes restore-pulse {
    0%,
    100% {
        opacity: 0.75;
        transform: scale(0.9);
    }

    50% {
        opacity: 1;
        transform: scale(1);
    }
}

@media (prefers-reduced-motion: reduce) {
    .restore-spinner::after {
        animation-duration: 1.8s;
    }

    .restore-spinner span {
        animation: none;
    }
}
</style>
