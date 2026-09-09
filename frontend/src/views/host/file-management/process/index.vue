<template>
    <DialogPro v-model="open" :title="$t('file.downloadProcess')" size="small" @close="handleClose">
        <template #content>
            <div v-loading="loading" class="space-y-4 p-4 min-h-[160px]">
                <div
                    v-for="value in res"
                    :key="value.key"
                    class="rounded-lg p-4 shadow-sm border border-gray-100 transition-all duration-200 hover:shadow-md"
                    :class="{ completed: getStatus(value) === 'Success' }"
                >
                    <div class="flex items-center gap-3">
                        <div class="flex-1 min-w-0">
                            <el-tooltip :content="value.name" placement="top">
                                <div class="truncate text-gray-700">{{ value.name }}</div>
                            </el-tooltip>
                            <div class="text-gray-500">
                                {{ getStatusText(value) }}
                            </div>
                            <div v-if="value.error" class="text-red-500 break-all">{{ value.error }}</div>
                        </div>
                    </div>

                    <div class="space-y-2">
                        <div class="flex justify-between items-center mb-1 text-gray-500">
                            <div>
                                <span>{{ getFileSize(value.written) }}</span>
                                <span v-if="value.total > 0" class="text-gray-400">
                                    /{{ getFileSize(value.total) }}
                                </span>
                            </div>
                            <el-button
                                v-if="isActive(value) && keys.includes(value.key)"
                                link
                                type="danger"
                                size="small"
                                :loading="stoppingKeys.includes(value.key)"
                                @click="onStop(value.key)"
                            >
                                {{ $t('commons.button.stop') }}
                            </el-button>
                        </div>
                        <div class="w-full">
                            <el-progress
                                v-if="value.total === 0 && isActive(value)"
                                :percentage="100"
                                :indeterminate="true"
                                :duration="1"
                                class="progress-bar"
                                :stroke-width="8"
                                :show-text="false"
                            />
                            <el-progress
                                v-else
                                :percentage="getProgressPercent(value)"
                                :stroke-width="8"
                                class="progress-bar"
                                :status="getProgressStatus(value)"
                            />
                        </div>
                    </div>
                </div>
            </div>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { fileWgetKeys, stopWgetFile, removeWgetRecords } from '@/api/modules/files';
import { computeSize } from '@/utils/size';
import { onBeforeUnmount, ref, watch } from 'vue';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { ElMessageBox } from 'element-plus';
import { MsgError, MsgSuccess } from '@/utils/message';
import i18n from '@/lang';
import { checkStreamAuth } from '@/utils/stream-auth';
const { currentNode: globalCurrentNode } = useGlobalStore();

let processSocket: WebSocket | null = null;
let sendTimer: ReturnType<typeof setInterval> | null = null;
let initProcessToken = 0;
interface DownloadProcess {
    key: string;
    name: string;
    written: number;
    total: number;
    percent: number;
    status?: 'Downloading' | 'Retrying' | 'Success' | 'Failed' | 'Canceled';
    attempt?: number;
    error?: string;
}

const res = ref<DownloadProcess[]>([]);
const stoppingKeys = ref<string[]>([]);
const removingKeys = ref<string[]>([]);
const removedKeys = ref<string[]>([]);
const reportedFailures = new Set<string>();
const reportedSuccesses = new Set<string>();
const autoRemoveAttempts = new Map<string, number>();
const keys = ref(['']);
const open = ref(false);
const loading = ref(false);

const em = defineEmits(['close']);
const handleClose = () => {
    initProcessToken++;
    closeSocket();
    loading.value = false;
    open.value = false;
    em('close', open.value);
};

const isWsOpen = () => {
    return processSocket?.readyState === WebSocket.OPEN;
};
const clearSendTimer = () => {
    if (sendTimer) {
        clearInterval(sendTimer);
        sendTimer = null;
    }
};
const closeSocket = () => {
    clearSendTimer();
    if (processSocket) {
        processSocket.onopen = null;
        processSocket.onmessage = null;
        processSocket.onerror = null;
        processSocket.onclose = null;
        processSocket.close();
    }
    processSocket = null;
};

const onOpenProcess = () => {
    sendProgressRequest();
    sendMsg();
};
const onMessage = async (message: any) => {
    const token = initProcessToken;
    let processes: DownloadProcess[];
    try {
        processes = JSON.parse(message.data) || [];
    } catch {
        return;
    }
    if (!Array.isArray(processes)) return;
    loading.value = false;
    res.value = processes
        .map((value, index) => ({
            ...value,
            key: value.key || (processes.length === keys.value.length ? keys.value[index] : `legacy:${value.name}`),
        }))
        .filter((value) => !removedKeys.value.includes(value.key));
    if (open.value) {
        const failures = res.value.filter((value) => getStatus(value) === 'Failed' && !reportedFailures.has(value.key));
        if (failures.length > 0) {
            failures.forEach((value) => reportedFailures.add(value.key));
            MsgError(failures.map((value) => `${value.name}: ${value.error || getStatusText(value)}`).join('\n'));
        }
        const successes = res.value.filter(
            (value) => getStatus(value) === 'Success' && !reportedSuccesses.has(value.key),
        );
        if (successes.length > 0) {
            successes.forEach((value) => reportedSuccesses.add(value.key));
            MsgSuccess(successes.map((value) => `${value.name}: ${getStatusText(value)}`).join('\n'));
        }
        await onRemove(getAutoRemoveKeys());
    }
    if (token !== initProcessToken) return;
    closeIdleSocket();
};
const onerror = () => {
    if (open.value && loading.value) {
        MsgError(i18n.global.t('commons.msg.operationFailed'));
        handleClose();
    }
};
const onClose = () => {
    clearSendTimer();
    onerror();
};

const getStatus = (value: DownloadProcess) => value.status || (value.percent === 100 ? 'Success' : 'Downloading');
const isActive = (value: DownloadProcess) => ['Downloading', 'Retrying'].includes(getStatus(value));
const isRemovable = (value: DownloadProcess) =>
    keys.value.includes(value.key) && ['Success', 'Failed', 'Canceled'].includes(getStatus(value));
const getFinishedKeys = () => res.value.filter(isRemovable).map((value) => value.key);
const getAutoRemoveKeys = () =>
    res.value
        .filter((value) => isRemovable(value) && (autoRemoveAttempts.get(value.key) || 0) < 3)
        .map((value) => value.key);
const closeIdleSocket = () => {
    if (open.value && res.value.length === 0 && removingKeys.value.length === 0) {
        keys.value = [];
        handleClose();
        return;
    }
    if (res.value.every((value) => !isActive(value)) && getAutoRemoveKeys().length === 0) closeSocket();
};
const onRemove = async (requestedKeys: string[]) => {
    if (removingKeys.value.length > 0) return;
    const removable = getFinishedKeys();
    const selected = [...new Set(requestedKeys.filter((key) => removable.includes(key)))];
    if (selected.length === 0) return;
    const node = globalCurrentNode.value;
    const token = initProcessToken;
    removingKeys.value = selected;
    try {
        for (let offset = 0; offset < selected.length; offset += 1000) {
            if (node !== globalCurrentNode.value || token !== initProcessToken || !open.value) return;
            const batch = selected.slice(offset, offset + 1000);
            batch.forEach((key) => autoRemoveAttempts.set(key, (autoRemoveAttempts.get(key) || 0) + 1));
            const response = await removeWgetRecords(batch, node);
            if (node !== globalCurrentNode.value || token !== initProcessToken || !open.value) return;
            const removed = (response.data?.keys || []).filter((key) => batch.includes(key));
            removedKeys.value.push(...removed);
            res.value = res.value.filter((value) => !removed.includes(value.key));
            keys.value = keys.value.filter((key) => !removed.includes(key));
            closeIdleSocket();
            if (removed.length !== batch.length) {
                if (batch.some((key) => (autoRemoveAttempts.get(key) || 0) >= 3)) {
                    MsgError(i18n.global.t('file.downloadRecordsNotRemoved'));
                }
                return;
            }
        }
    } catch (error) {
    } finally {
        if (token === initProcessToken) {
            removingKeys.value = [];
            closeIdleSocket();
        }
    }
};
const getProgressPercent = (value: DownloadProcess) => {
    if (getStatus(value) === 'Success') return 100;
    if (!Number.isFinite(value.percent)) return 0;
    const maximum = isActive(value) ? 99.99 : 100;
    return Number(Math.min(maximum, Math.max(0, value.percent)).toFixed(2));
};
const getStatusText = (value: DownloadProcess) => {
    switch (getStatus(value)) {
        case 'Success':
            return i18n.global.t('file.downloadSuccess');
        case 'Failed':
            return i18n.global.t('commons.status.failed');
        case 'Canceled':
            return i18n.global.t('commons.status.canceled');
        case 'Retrying':
            return `${i18n.global.t('commons.button.retry')} (${value.attempt}/3)`;
        default:
            return i18n.global.t('file.downloading');
    }
};
const getProgressStatus = (value: DownloadProcess) => {
    const status = getStatus(value);
    if (status === 'Success') return 'success';
    if (status === 'Failed') return 'exception';
    if (status === 'Canceled') return 'warning';
    return '';
};

const initProcess = async () => {
    const token = initProcessToken;
    let href = window.location.href;
    let protocol = href.split('//')[0] === 'http:' ? 'ws' : 'wss';
    let ipLocal = href.split('//')[1].split('/')[0];
    let currentNode = globalCurrentNode.value;
    const url = `${protocol}://${ipLocal}/api/v2/files/wget/process?operateNode=${currentNode}`;
    const authError = await checkStreamAuth(url, currentNode);
    if (token !== initProcessToken || !open.value) {
        return;
    }
    if (authError) {
        MsgError(authError);
        handleClose();
        return;
    }
    closeSocket();
    processSocket = new WebSocket(url);
    processSocket.onopen = onOpenProcess;
    processSocket.onmessage = onMessage;
    processSocket.onerror = onerror;
    processSocket.onclose = onClose;
};

const getKeys = async () => {
    const token = ++initProcessToken;
    keys.value = [];
    res.value = [];
    removingKeys.value = [];
    removedKeys.value = [];
    reportedFailures.clear();
    reportedSuccesses.clear();
    autoRemoveAttempts.clear();
    loading.value = true;
    try {
        const res = await fileWgetKeys();
        if (token !== initProcessToken || !open.value) return;
        if (res.data?.keys?.length > 0) {
            keys.value = res.data.keys;
            await initProcess();
        } else {
            handleClose();
        }
    } catch (error) {
        if (token === initProcessToken && open.value) handleClose();
    }
};

const sendProgressRequest = () => {
    if (isWsOpen()) {
        processSocket?.send(JSON.stringify({ type: 'wget', keys: keys.value }));
    }
};
const sendMsg = () => {
    clearSendTimer();
    sendTimer = setInterval(sendProgressRequest, 1000);
};

const getFileSize = (size: number) => {
    return computeSize(size);
};

const onStop = async (key: string) => {
    if (!keys.value.includes(key) || stoppingKeys.value.includes(key)) return;
    const node = globalCurrentNode.value;
    const token = initProcessToken;
    try {
        await ElMessageBox.confirm(i18n.global.t('file.stopWgetConfirm'), i18n.global.t('commons.button.tip'), {
            type: 'warning',
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        });
    } catch {
        return;
    }
    if (node !== globalCurrentNode.value || token !== initProcessToken || !open.value) return;
    stoppingKeys.value.push(key);
    try {
        await stopWgetFile(key, node);
        if (token === initProcessToken && open.value && node === globalCurrentNode.value) {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        }
    } catch (e) {
    } finally {
        stoppingKeys.value = stoppingKeys.value.filter((item) => item !== key);
    }
};

watch(globalCurrentNode, () => {
    handleClose();
    keys.value = [];
    res.value = [];
    removingKeys.value = [];
    removedKeys.value = [];
    reportedFailures.clear();
    reportedSuccesses.clear();
    autoRemoveAttempts.clear();
});

onBeforeUnmount(() => {
    initProcessToken++;
    closeSocket();
});

const acceptParams = () => {
    open.value = true;
    getKeys();
};

defineExpose({ acceptParams });
</script>

<style lang="scss" scoped>
.download-item.completed {
    background-color: rgb(240 253 244 / 0.5);
}

.progress-bar {
    :deep(.el-progress-bar__outer) {
        border-radius: 9999px;
        background-color: rgb(243 244 246);
    }

    :deep(.el-progress-bar__inner) {
        border-radius: 9999px;
        transition-property: all;
        transition-duration: 300ms;
        transition-timing-function: cubic-bezier(0.4, 0, 0.2, 1);
    }
}

@keyframes bounce {
    0%,
    100% {
        transform: translateY(-10%);
        animation-timing-function: cubic-bezier(0.8, 0, 1, 1);
    }
    50% {
        transform: translateY(0);
        animation-timing-function: cubic-bezier(0, 0, 0.2, 1);
    }
}
</style>
