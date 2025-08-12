<template>
    <DialogPro v-model="open" :title="$t('file.downloadProcess')" size="small" @close="handleClose">
        <template #content>
            <div class="space-y-4 p-4" :loading="loading">
                <div
                    v-for="(value, index) in res"
                    :key="index"
                    class="rounded-lg p-4 shadow-sm border border-gray-100 transition-all duration-200 hover:shadow-md"
                    :class="{ completed: value.percent === 100 }"
                >
                    <div class="flex items-center gap-3">
                        <div class="flex-1">
                            <MsgInfo :info="value.name" class="text-gray-700" />
                            <div class="text-gray-500">
                                {{ value.percent === 100 ? $t('file.downloadSuccess') : $t('file.downloading') }}
                            </div>
                        </div>
                    </div>

                    <div class="space-y-2">
                        <div class="flex justify-end text-gray-500 mb-1">
                            <span>{{ getFileSize(value.written) }}</span>
                            <span v-if="value.total > 0" class="text-gray-400">/{{ getFileSize(value.total) }}</span>
                        </div>
                        <div class="w-full">
                            <el-progress
                                v-if="value.total === 0 && value.percent != 100"
                                :percentage="100"
                                :indeterminate="true"
                                :duration="1"
                                class="progress-bar"
                                :stroke-width="8"
                                :show-text="false"
                            />
                            <el-progress
                                v-else
                                :percentage="value.percent"
                                :stroke-width="8"
                                class="progress-bar"
                                :status="value.percent === 100 ? 'success' : ''"
                            />
                        </div>
                    </div>
                </div>
            </div>
        </template>
    </DialogPro>
</template>

<script lang="ts" setup>
import { fileWgetKeys } from '@/api/modules/files';
import { computeSize } from '@/utils/util';
import { onBeforeUnmount, ref } from 'vue';
import MsgInfo from '@/components/msg-info/index.vue';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

let processSocket = ref(null) as unknown as WebSocket;
const res = ref([]);
const keys = ref(['']);
const open = ref(false);
const loading = ref(false);

const em = defineEmits(['close']);
const handleClose = () => {
    closeSocket();
    open.value = false;
    em('close', open);
};

const isWsOpen = () => {
    const readyState = processSocket && processSocket.readyState;
    return readyState === 1;
};
const closeSocket = () => {
    if (isWsOpen()) {
        processSocket && processSocket.close();
    }
};

const onOpenProcess = () => {};
const onMessage = (message: any) => {
    res.value = JSON.parse(message.data);
};
const onerror = () => {};
const onClose = () => {};

const initProcess = () => {
    let href = window.location.href;
    let protocol = href.split('//')[0] === 'http:' ? 'ws' : 'wss';
    let ipLocal = href.split('//')[1].split('/')[0];
    let currentNode = globalStore.currentNode;
    processSocket = new WebSocket(`${protocol}://${ipLocal}/api/v2/files/wget/process?operateNode=${currentNode}`);
    processSocket.onopen = onOpenProcess;
    processSocket.onmessage = onMessage;
    processSocket.onerror = onerror;
    processSocket.onclose = onClose;
    sendMsg();
};

const getKeys = async () => {
    keys.value = [];
    res.value = [];
    loading.value = true;
    try {
        const res = await fileWgetKeys();
        if (res.data && res.data.keys.length > 0) {
            keys.value = res.data.keys;
            initProcess();
        }
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

const sendMsg = () => {
    setInterval(() => {
        if (isWsOpen()) {
            processSocket.send(
                JSON.stringify({
                    type: 'wget',
                    keys: keys.value,
                }),
            );
        }
    }, 1000);
};

const getFileSize = (size: number) => {
    return computeSize(size);
};

onBeforeUnmount(() => {
    closeSocket();
});

const acceptParams = () => {
    open.value = true;
    getKeys();
};

defineExpose({ acceptParams });
</script>

<style type="scss" scoped>
.download-item.completed {
    @apply bg-green-50/50;
}

.progress-bar {
    :deep(.el-progress-bar__outer) {
        @apply rounded-full bg-gray-100;
    }

    :deep(.el-progress-bar__inner) {
        @apply rounded-full transition-all duration-300;
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
