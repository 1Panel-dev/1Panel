<template>
    <div class="share-page" v-loading="initializing">
        <div class="share-card">
            <h1>{{ $t('file.shareLinkLabel') }}</h1>
            <div v-if="shareInfo" class="share-meta">
                <div class="meta-item">
                    <span class="meta-label">{{ $t('file.fileName') }}</span>
                    <span class="meta-value">{{ shareInfo.fileName }}</span>
                </div>
                <div class="meta-item">
                    <span class="meta-label">{{ $t('file.shareExpiresAt') }}</span>
                    <span class="meta-value">{{ expiresAtText }}</span>
                </div>
            </div>
            <p v-if="shareInfo?.hasPassword" class="share-description">{{ $t('file.shareDownloadPasswordTip') }}</p>
            <p v-else-if="shareInfo && !errorMessage" class="share-description">
                {{ $t('file.shareDownloadingHint') }}
            </p>
            <el-form v-if="shareInfo" @submit.prevent>
                <el-form-item v-if="shareInfo.hasPassword" :label="$t('file.sharePasswordRequired')">
                    <el-input
                        v-model="password"
                        type="password"
                        show-password
                        clearable
                        maxlength="256"
                        :placeholder="$t('file.sharePasswordPlaceholder')"
                        autocomplete="current-password"
                        @keyup.enter="downloadWithPassword"
                    />
                </el-form-item>
                <div v-if="errorMessage" class="share-error">{{ errorMessage }}</div>
                <el-button type="primary" :loading="downloading" class="share-button" @click="downloadWithPassword">
                    {{ $t('file.shareExtractFile') }}
                </el-button>
            </el-form>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { checkFileShare, getPublicFileShareInfo } from '@/api/modules/files';
import { File } from '@/api/interface/file';
import i18n, { loadLocaleMessages } from '@/lang';
import { buildFileShareDownloadUrl } from '@/utils/file';
import { dateFormat } from '@/utils/date';
import { computed, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';

const route = useRoute();
const password = ref('');
const downloading = ref(false);
const initializing = ref(true);
const errorMessage = ref('');
const shareInfo = ref<File.FileSharePublicInfo | null>(null);
const shareLocale = ref('en');
const supportedLocales = ['zh', 'zh-Hant', 'en', 'pt-BR', 'ja', 'ru', 'ms', 'ko', 'tr', 'es-ES', 'lo'];

const code = computed(() => String(route.params.code || '').trim());
const currentNode = computed(() => String(route.query.operateNode || 'local'));
const expiresAtText = computed(() => {
    if (!shareInfo.value) {
        return '--';
    }
    if (shareInfo.value.permanent || shareInfo.value.expiresAt === 0) {
        return String(i18n.global.t('website.ever'));
    }
    return dateFormat(null, null, shareInfo.value.expiresAt * 1000);
});

const getFallbackText = (type: 'invalid' | 'failed') => {
    return String(i18n.global.t(type === 'invalid' ? 'file.shareInvalid' : 'file.shareDownloadFailed'));
};

const getPasswordRequiredText = () => {
    return String(i18n.global.t('file.sharePasswordRequiredInput'));
};

const triggerDownload = (pwd = '') => {
    window.location.href = buildFileShareDownloadUrl(code.value, currentNode.value, pwd);
};

const resolveBrowserLocale = () => {
    if (typeof navigator === 'undefined') {
        return 'en';
    }
    const browserLocale = String(navigator.language || '').trim();
    const normalized = browserLocale.toLowerCase();
    if (normalized.startsWith('zh-hant') || normalized.startsWith('zh-tw') || normalized.startsWith('zh-hk')) {
        return 'zh-Hant';
    }
    if (normalized.startsWith('zh')) {
        return 'zh';
    }
    const exactMatch = supportedLocales.find((locale) => locale.toLowerCase() === normalized);
    if (exactMatch) {
        return exactMatch;
    }
    const prefix = normalized.split('-')[0];
    const prefixMatch = supportedLocales.find((locale) => locale.toLowerCase().split('-')[0] === prefix);
    return prefixMatch || 'en';
};

const applyPublicLocale = async () => {
    const locale = resolveBrowserLocale();
    const loaded = await loadLocaleMessages(locale);
    i18n.global.locale.value = loaded;
    shareLocale.value = loaded;
};

const loadShareInfo = async () => {
    if (!code.value) {
        errorMessage.value = getFallbackText('invalid');
        return;
    }
    const res = await getPublicFileShareInfo(code.value, currentNode.value, {
        'Accept-Language': shareLocale.value,
    });
    shareInfo.value = res.data;
};

const downloadWithPassword = async () => {
    if (!code.value) {
        errorMessage.value = getFallbackText('invalid');
        return;
    }
    if (shareInfo.value?.hasPassword && !password.value.trim()) {
        errorMessage.value = getPasswordRequiredText();
        return;
    }
    downloading.value = true;
    errorMessage.value = '';
    try {
        await checkFileShare(
            {
                code: code.value,
                password: password.value.trim(),
                operateNode: currentNode.value,
            },
            {
                'Accept-Language': shareLocale.value,
            },
        );
        triggerDownload(password.value.trim());
    } catch (error) {
        errorMessage.value = (error as { message?: string })?.message || getFallbackText('failed');
    } finally {
        downloading.value = false;
    }
};

onMounted(async () => {
    try {
        await applyPublicLocale();
        await loadShareInfo();
        if (shareInfo.value && !shareInfo.value.hasPassword) {
            await downloadWithPassword();
        }
    } catch (error) {
        errorMessage.value = (error as { message?: string })?.message || getFallbackText('invalid');
    } finally {
        initializing.value = false;
    }
});
</script>

<style scoped lang="scss">
.share-page {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 24px;
    background:
        radial-gradient(circle at top, rgba(0, 94, 235, 0.14), transparent 32%),
        linear-gradient(160deg, #f5f8ff 0%, #eef4ff 42%, #ffffff 100%);
}

.share-card {
    width: min(100%, 460px);
    padding: 32px;
    border-radius: 24px;
    background: rgba(255, 255, 255, 0.92);
    border: 1px solid rgba(0, 94, 235, 0.12);
    box-shadow: 0 20px 60px rgba(15, 44, 92, 0.12);

    h1 {
        margin: 14px 0 10px;
        font-size: 28px;
        line-height: 1.2;
        color: #16325c;
    }
}

.share-description {
    margin: 0 0 20px;
    color: #5f6f8a;
    line-height: 1.6;
}

.share-meta {
    display: grid;
    gap: 12px;
    margin: 0 0 20px;
    padding: 14px 16px;
    border-radius: 14px;
    background: #f7f9fd;
    border: 1px solid rgba(0, 94, 235, 0.1);
}

.meta-item {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
}

.meta-label {
    flex: none;
    color: #7a879d;
    font-size: 13px;
}

.meta-value {
    color: #16325c;
    font-size: 14px;
    text-align: right;
    word-break: break-all;
}

.share-error {
    margin: -6px 0 16px;
    color: #d03050;
    font-size: 13px;
    line-height: 1.5;
}

.share-button {
    width: 100%;
}
</style>
