<template>
    <DialogPro
        v-model="open"
        :title="$t('license.restoreCommunity')"
        size="large"
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

            <span class="input-help">
                {{
                    restoreMode === 'offline'
                        ? $t('license.restoreCommunityOfflineHelper', { path: restoreInfo.packageDirectory })
                        : $t('license.restoreCommunityOnlineHelper')
                }}
            </span>

            <el-alert
                v-if="restoreMode === 'offline'"
                v-loading="packageChecking"
                :title="
                    restoreInfo.packageExist
                        ? $t('license.restoreCommunityPackageFound')
                        : $t('license.restoreCommunityPackageMissing')
                "
                :type="restoreInfo.packageExist ? 'success' : 'warning'"
                :closable="false"
                show-icon
            />
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

.restore-mode-group {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;

    :deep(.el-radio) {
        width: 100%;
        margin-right: 0;
    }
}
</style>
