<template>
    <el-dialog width="30%" v-model="open" @open="onOpen" :before-close="handleClose">
        <template #header>
            <DrawerHeader :header="$t('file.downloadProcess')" :back="handleClose" />
        </template>
        <div v-for="(value, index) in res" :key="index">
            <span>{{ $t('file.downloading') }} {{ value['name'] }}</span>
            <el-progress :text-inside="true" :stroke-width="15" :percentage="value['percent']"></el-progress>
            <span>{{ getFileSize(value['written']) }}/{{ getFileSize(value['total']) }}</span>
        </div>
    </el-dialog>
</template>

<script lang="ts" setup>
import { FileKeys } from '@/api/modules/files';
import { computeSize } from '@/utils/util';
import { onBeforeUnmount, ref, toRefs } from 'vue';
import DrawerHeader from '@/components/drawer-header/index.vue';

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
    processSocket = new WebSocket(`${protocol}://${ipLocal}/api/v1/files/ws`);
    processSocket.onopen = onOpenProcess;
    processSocket.onmessage = onMessage;
    processSocket.onerror = onerror;
    processSocket.onclose = onClose;
    sendMsg();
};

const getKeys = () => {
    FileKeys().then((res) => {
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

const getFileSize = (size: number) => {
    return computeSize(size);
};

onBeforeUnmount(() => {
    closeSocket();
});

const onOpen = () => {
    getKeys();
};
</script>
