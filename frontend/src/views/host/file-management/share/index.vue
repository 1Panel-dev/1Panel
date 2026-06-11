<template>
    <DrawerPro v-model="open" :header="$t('file.shareFile')" size="large" @close="handleClose">
        <el-alert
            :title="$t('file.shareRiskAlert')"
            type="warning"
            :closable="false"
            show-icon
            class="share-risk-alert"
        />
        <el-form
            ref="shareFormRef"
            :model="form"
            :rules="rules"
            label-position="top"
            label-width="100px"
            v-loading="loading"
        >
            <el-form-item :label="$t('file.path')">
                <el-input :model-value="filePath" type="textarea" :rows="2" readonly />
            </el-form-item>
            <el-form-item :label="$t('file.shareExpire')">
                <el-select v-model="form.expireMinutes" class="w-full">
                    <el-option v-for="opt in expireOptions" :key="opt.value" :label="opt.label" :value="opt.value" />
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('file.sharePassword')" prop="sharePassword">
                <el-input
                    v-model="form.sharePassword"
                    type="password"
                    show-password
                    clearable
                    maxlength="256"
                    show-word-limit
                    :placeholder="$t('file.sharePasswordPlaceholder')"
                    autocomplete="new-password"
                />
            </el-form-item>

            <template v-if="shareInfo">
                <el-form-item :label="$t('file.shareLinkLabel')">
                    <div class="share-link-bar">
                        <el-link class="share-link-value" :underline="false" type="primary" @click="openShareUrl">
                            {{ shareUrl }}
                        </el-link>
                        <div class="flex items-center justify-end">
                            <el-button link class="share-link-button" @click="copyLink">
                                <CopyDocument class="share-link-icon" size="16" />
                            </el-button>
                            <el-divider direction="vertical" />
                            <el-button link class="share-link-button" @click="openQrDialog">
                                <svg-icon class="share-link-icon" iconName="p-qrcode"></svg-icon>
                            </el-button>
                        </div>
                    </div>
                </el-form-item>
                <el-form-item :label="$t('file.shareExpiresAt')">
                    <el-input :model-value="expiresAtText" readonly />
                </el-form-item>
            </template>
        </el-form>
        <template #footer>
            <span class="dialog-footer">
                <el-button :disabled="loading" @click="handleClose">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button v-permission v-node-admin v-if="shareInfo" :disabled="loading" @click="cancelShare">
                    {{ $t('file.shareClose') }}
                </el-button>
                <el-button v-if="shareInfo" :disabled="loading" @click="copyLink">
                    {{ $t('file.shareCopyLink') }}
                </el-button>
                <el-button v-permission v-node-admin type="primary" :disabled="loading" @click="generate">
                    {{ shareInfo ? $t('file.shareRegenerate') : $t('file.shareGenerate') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
    <DialogPro v-model="qrDialogOpen" :title="$t('file.shareQrDialogTitle')" size="small">
        <div class="share-qr-dialog">
            <div class="share-qr-title">{{ $t('file.shareQrDialogTitle') }}</div>
            <img v-if="qrCodeUrl" class="share-qr-image" :src="qrCodeUrl" :alt="$t('file.shareQrCode')" />
            <div class="share-qr-helper">{{ $t('file.shareQrDialogHelper') }}</div>
            <div class="share-qr-actions">
                <el-button text @click="saveQrImage">
                    <Download class="share-link-icon" />
                    {{ $t('file.shareSaveImage') }}
                </el-button>
                <el-button text @click="openQrImage">
                    <Picture class="share-link-icon" />
                    {{ $t('file.shareOpenImage') }}
                </el-button>
            </div>
        </div>
    </DialogPro>
</template>

<script lang="ts" setup>
import { createFileShare, getFileShareDetail, removeFileShare } from '@/api/modules/files';
import { File } from '@/api/interface/file';
import { CopyDocument, Download, Picture } from '@element-plus/icons-vue';
import i18n from '@/lang';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { buildFileSharePageUrl, buildFileShareQrCodeUrl } from '@/utils/file';
import { copyText } from '@/utils/clipboard';
import { dateFormat as formatDateTime } from '@/utils/date';
import type { FormInstance, FormRules } from 'element-plus';
import { computed, reactive, ref, watch } from 'vue';

interface ShareProps {
    path: string;
}

const { currentNode } = useGlobalStore();
const open = ref(false);
const loading = ref(false);
const changed = ref(false);
const filePath = ref('');
const shareInfo = ref<File.FileShareInfo | null>(null);
const shareUrl = ref('');
const expiresAtText = ref('');
const qrCodeUrl = ref('');
const qrDialogOpen = ref(false);
const shareFormRef = ref<FormInstance>();
const syncingPassword = ref(false);
const initialSharePassword = ref('');
const passwordTouched = ref(false);
const form = reactive({
    expireMinutes: 1440,
    sharePassword: '',
});

const emit = defineEmits(['close']);

const expireOptions = computed(() => [
    { label: i18n.global.t('file.shareExpire1h'), value: 60 },
    { label: i18n.global.t('file.shareExpire6h'), value: 360 },
    { label: i18n.global.t('file.shareExpire24h'), value: 1440 },
    { label: i18n.global.t('file.shareExpire3d'), value: 4320 },
    { label: i18n.global.t('file.shareExpire7d'), value: 10080 },
    { label: i18n.global.t('website.ever'), value: 0 },
]);

const validatePassword = (_rule, value, callback) => {
    const password = typeof value === 'string' ? value.trim() : '';
    if (password.length === 0) {
        callback();
        return;
    }
    if (password.length < 4 || password.length > 256) {
        callback(new Error(String(i18n.global.t('file.sharePasswordLengthHint'))));
        return;
    }
    callback();
};

const rules = reactive<FormRules>({
    sharePassword: [{ validator: validatePassword, trigger: 'blur' }],
});

watch(
    () => form.sharePassword,
    (val) => {
        if (syncingPassword.value) {
            return;
        }
        if (val === initialSharePassword.value) {
            return;
        }
        passwordTouched.value = true;
    },
);

const mapExpireMinutes = (info: File.FileShareInfo) => {
    if (info.permanent || info.expiresAt === 0) {
        return 0;
    }
    const remainMinutes = Math.max(1, Math.round((info.expiresAt * 1000 - Date.now()) / 60000));
    const matched = expireOptions.value.find((item) => item.value === remainMinutes);
    return matched ? matched.value : 1440;
};

const applyShareInfo = (info: File.FileShareInfo | null) => {
    shareInfo.value = info;
    if (!info) {
        shareUrl.value = '';
        expiresAtText.value = '';
        qrCodeUrl.value = '';
        qrDialogOpen.value = false;
        form.expireMinutes = 1440;
        form.sharePassword = '';
        syncingPassword.value = true;
        initialSharePassword.value = '';
        passwordTouched.value = false;
        syncingPassword.value = false;
        return;
    }
    shareUrl.value = buildFileSharePageUrl(info.code, currentNode.value);
    qrCodeUrl.value = buildFileShareQrCodeUrl(info.code, currentNode.value);
    expiresAtText.value = info.permanent
        ? i18n.global.t('website.ever')
        : formatDateTime(null, null, info.expiresAt * 1000);
    form.expireMinutes = mapExpireMinutes(info);
    syncingPassword.value = true;
    form.sharePassword = info.password || '';
    initialSharePassword.value = form.sharePassword;
    passwordTouched.value = false;
    syncingPassword.value = false;
};

const loadShareDetail = async () => {
    const res = await getFileShareDetail(filePath.value);
    applyShareInfo(res.data?.code ? res.data : null);
};

const resetForm = () => {
    filePath.value = '';
    changed.value = false;
    applyShareInfo(null);
    shareFormRef.value?.clearValidate();
};

const handleClose = () => {
    const needRefresh = changed.value;
    open.value = false;
    resetForm();
    emit('close', needRefresh);
};

const generate = async () => {
    const valid = await shareFormRef.value?.validate().catch(() => false);
    if (valid === false) {
        return;
    }
    loading.value = true;
    try {
        const pw = form.sharePassword.trim();
        const payload: File.FileShareCreate = {
            path: filePath.value,
            expireMinutes: form.expireMinutes,
        };

        if (passwordTouched.value) {
            payload.password = pw;
        } else if (!shareInfo.value?.hasPassword && pw.length > 0) {
            payload.password = pw;
        }

        const res = await createFileShare(payload);
        applyShareInfo(res.data as File.FileShareInfo);
        changed.value = true;
    } finally {
        loading.value = false;
    }
};

const cancelShare = async () => {
    loading.value = true;
    try {
        await removeFileShare(filePath.value);
        applyShareInfo(null);
        changed.value = true;
    } finally {
        loading.value = false;
    }
};

const copyLink = () => {
    if (shareUrl.value) {
        const password = form.sharePassword.trim();
        const content = password
            ? `${i18n.global.t('file.shareLinkLabel')}：${shareUrl.value}，${i18n.global.t(
                  'file.sharePassword',
              )}：${password}`
            : `${i18n.global.t('file.shareLinkLabel')}：${shareUrl.value}`;
        copyText(content);
    }
};

const openShareUrl = () => {
    if (shareUrl.value) {
        window.open(shareUrl.value, '_blank', 'noopener,noreferrer');
    }
};

const openQrDialog = () => {
    if (qrCodeUrl.value) {
        qrDialogOpen.value = true;
    }
};

const openQrImage = () => {
    if (qrCodeUrl.value) {
        window.open(qrCodeUrl.value, '_blank', 'noopener,noreferrer');
    }
};

const saveQrImage = () => {
    if (!qrCodeUrl.value) {
        return;
    }
    const link = document.createElement('a');
    link.href = qrCodeUrl.value;
    link.download = `share-${shareInfo.value?.code || 'qrcode'}.png`;
    link.target = '_blank';
    link.rel = 'noopener noreferrer';
    link.click();
};

const acceptParams = async (params: ShareProps) => {
    resetForm();
    filePath.value = params.path;
    open.value = true;
    loading.value = true;
    try {
        await loadShareDetail();
    } finally {
        loading.value = false;
    }
};

defineExpose({ acceptParams });
</script>

<style scoped lang="scss">
.share-link-bar {
    width: 100%;
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 8px 12px;
    line-height: 1.8;
}

.share-risk-alert {
    margin-bottom: 18px;
}

.share-link-prefix {
    color: var(--el-text-color-secondary);
}

.share-link-value {
    max-width: min(100%, 480px);
    justify-content: flex-start;
    text-align: left;
    word-break: break-all;
}

.share-link-button {
    padding: 0;
}

.share-link-icon {
    width: 14px;
    height: 14px;
    margin-right: 4px;
}

.share-qr-dialog {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
    padding: 6px 0 2px;
}

.share-qr-title {
    color: var(--el-color-primary);
    font-size: 18px;
}

.share-qr-image {
    width: 240px;
    height: 240px;
    object-fit: contain;
    background: #fff;
}

.share-qr-helper {
    color: var(--el-text-color-regular);
}

.share-qr-actions {
    width: 100%;
    display: flex;
    justify-content: center;
    gap: 12px;
    padding-top: 12px;
    border-top: 1px solid var(--el-border-color-lighter);
}
</style>
