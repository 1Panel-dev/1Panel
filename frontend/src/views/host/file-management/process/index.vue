<template>
    <el-dialog
        width="30%"
        :title="$t('file.downloadProcess')"
        v-model="open"
        @open="onOpen"
        :before-close="handleClose"
    >
        <div v-for="(value, index) in res" :key="index">
            <span>{{ value['name'] }}</span>
            <el-progress :text-inside="true" :stroke-width="15" :percentage="value['percent']"></el-progress>
            <span>{{ value['written'] }}/{{ value['total'] }}</span>
        </div>
    </el-dialog>
</template>

<script lang="ts" setup>
import { FileKeys } from '@/api/modules/files';
import { onBeforeUnmount, ref, toRefs } from 'vue';

const props = defineProps({
    open: {
        type: Boolean,
        default: false,
    },
});

const { open } = toRefs(props);
let processSocket = ref(null) as unknown as WebSocket;
const res = ref([]);
const keys = ref(['']);

const em = defineEmits(['close']);
const handleClose = () => {
    closeSocket();
    em('close', open);
};

const closeSocket = () => {
    processSocket && processSocket.close();
};

const isWsOpen = () => {
    const readyState = processSocket && processSocket.readyState;
    return readyState === 1;
};

const onOpenProcess = () => {};
const onMessage = (message: any) => {
    res.value = JSON.parse(message.data);
};
const onerror = () => {};
const onClose = () => {};

const initProcess = () => {
    processSocket = new WebSocket(`ws://localhost:9999/api/v1/files/ws`);
    processSocket.onopen = onOpenProcess;
    processSocket.onmessage = onMessage;
    processSocket.onerror = onerror;
    processSocket.onclose = onClose;
    sendMsg();
};

const getKeys = () => {
    FileKeys().then((res) => {
        console.log(res);
        if (res.data.keys.length > 0) {
            keys.value = res.data.keys;
            initProcess();
        }
    });
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

onBeforeUnmount(() => {
    closeSocket();
});

const onOpen = () => {
    getKeys();
};
</script>
