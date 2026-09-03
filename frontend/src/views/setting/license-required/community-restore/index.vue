<template>
    <DialogPro
        v-model="open"
        :title="$t('license.restoreCommunity')"
        size="large"
        align-center
        append-to-body
        destroy-on-close
        @closed="handleClosed"
    >
        <div v-loading="loading" class="restore-dialog-content">
            <el-alert
                v-if="restoreInfo.state === 'Failed'"
                :title="restoreInfo.message"
                type="error"
                :closable="false"
                show-icon
            />
            <el-alert :title="$t('license.restoreCommunityConfirm')" type="warning" :closable="false" show-icon />

            <el-radio-group v-model="restoreMode" @change="changeRestoreMode">
                <el-radio value="online" border>{{ $t('license.restoreCommunityOnline') }}</el-radio>
                <el-radio value="offline" border>{{ $t('license.restoreCommunityOffline') }}</el-radio>
            </el-radio-group>

            <span v-if="restoreMode === 'online'" class="input-help">
                {{ $t('license.restoreCommunityOnlineHelper') }}
            </span>

            <div v-else class="package-status">
                <div class="package-status-header">
                    <el-icon v-if="packageChecking" class="is-loading"><Loading /></el-icon>
                    <el-icon v-else-if="restoreInfo.packageExist"><CircleCheckFilled /></el-icon>
                    <el-icon v-else><Clock /></el-icon>
                    <span>
                        {{
                            packageChecking
                                ? $t('license.restoreCommunityPackageChecking')
                                : restoreInfo.packageExist
                                  ? $t('license.restoreCommunityPackageFound')
                                  : $t('license.restoreCommunityPackageMissing')
                        }}
                    </span>
                </div>

                <div v-if="!packageChecking" class="package-status-content">
                    <span v-if="restoreInfo.packageExist && restoreInfo.packageName" class="package-name">
                        {{ restoreInfo.packageName }}
                    </span>
                    <span class="package-status-description">
                        {{
                            restoreInfo.packageExist
                                ? $t('license.restoreCommunityPackageReadyHelper')
                                : $t('license.restoreCommunityOfflineHelper', {
                                      path: restoreInfo.packageDirectory,
                                  })
                        }}
                    </span>
                    <el-button
                        v-if="!restoreInfo.packageExist && restoreInfo.packageURL"
                        class="copy-download-link"
                        type="primary"
                        link
                        @click="copyText(restoreInfo.packageURL)"
                    >
                        {{ $t('license.restoreCommunityCopyDownloadLink') }}
                    </el-button>
                </div>
            </div>
        </div>

        <template #footer>
            <el-button :disabled="loading" @click="open = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button
                type="primary"
                :loading="loading"
                :disabled="restoreMode === 'offline' && (packageChecking || !restoreInfo.packageExist)"
                @click="submit"
            >
                {{ $t('commons.button.confirm') }}
            </el-button>
        </template>
    </DialogPro>
</template>

<script setup lang="ts">
import { onUnmounted, reactive, ref } from 'vue';
import { getCommunityRestoreStatus, restoreCommunity } from '@/api/modules/setting';
import { copyText } from '@/utils/clipboard';
import { CircleCheckFilled, Clock, Loading } from '@element-plus/icons-vue';

const emit = defineEmits<{
    started: [];
}>();

const open = ref(false);
const loading = ref(false);
const packageChecking = ref(false);
const restoreMode = ref<'online' | 'offline'>('online');
const restoreInfo = reactive({
    state: 'Ready',
    message: '',
    packageExist: false,
    packageDirectory: '',
    packageName: '',
    packageURL: '',
});
let packageTimer: number | undefined;

const loadRestoreInfo = async () => {
    const res = await getCommunityRestoreStatus();
    Object.assign(restoreInfo, res.data);
};

const refreshOfflinePackage = async (showLoading = false) => {
    packageChecking.value = showLoading;
    try {
        await loadRestoreInfo();
    } finally {
        if (showLoading) {
            packageChecking.value = false;
        }
    }
};

const schedulePackagePoll = () => {
    window.clearTimeout(packageTimer);
    packageTimer = window.setTimeout(pollOfflinePackage, 3000);
};

const pollOfflinePackage = async () => {
    try {
        await refreshOfflinePackage();
    } finally {
        if (open.value && restoreMode.value === 'offline') {
            schedulePackagePoll();
        }
    }
};

const changeRestoreMode = async () => {
    window.clearTimeout(packageTimer);
    if (restoreMode.value !== 'offline') {
        return;
    }
    await refreshOfflinePackage(true);
    schedulePackagePoll();
};

const acceptParams = async () => {
    restoreMode.value = 'online';
    open.value = true;
    await loadRestoreInfo();
};

const handleClosed = () => {
    window.clearTimeout(packageTimer);
    restoreMode.value = 'online';
    loading.value = false;
    packageChecking.value = false;
};

const submit = async () => {
    loading.value = true;
    await restoreCommunity(restoreMode.value)
        .then(() => {
            emit('started');
            open.value = false;
        })
        .finally(() => {
            loading.value = false;
        });
};

onUnmounted(() => {
    window.clearTimeout(packageTimer);
});

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.restore-dialog-content {
    display: flex;
    flex-direction: column;
    gap: 16px;
}

.package-status {
    display: flex;
    flex-direction: column;
    gap: 6px;
    padding: 12px 16px;
    background-color: var(--el-bg-color);
    border: 1px solid var(--el-border-color-light);
    border-radius: 4px;
}

.package-status-header {
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 500;
    color: var(--el-text-color-primary);
}

.package-status-content {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 4px;
    padding-left: 22px;
}

.package-status-description {
    font-size: 12px;
    line-height: 18px;
    color: var(--el-text-color-secondary);
}

.package-name {
    font-size: 12px;
    line-height: 18px;
    color: var(--el-text-color-primary);
    word-break: break-all;
}

.copy-download-link {
    margin: 4px 0 0;
    padding: 0;
    font-size: 12px;
}
</style>
