<template>
    <div :id="'terminal' + props.id"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue';
import { Terminal } from 'xterm';
import { AttachAddon } from 'xterm-addon-attach';
import { Base64 } from 'js-base64';
import 'xterm/css/xterm.css';

interface WsProps {
    id: number;
}
const props = withDefaults(defineProps<WsProps>(), {
    id: 0,
});
const loading = ref(true);
let terminalSocket = ref(null) as unknown as WebSocket;
let term = ref(null) as unknown as Terminal;

const runRealTerminal = () => {
    loading.value = false;
};

const onWSReceive = (message: any) => {
    if (!isJson(message.data)) {
        return;
    }
    const data = JSON.parse(message.data);
    term.element && term.focus();
    term.write(data.Data);
};

function isJson(str: string) {
    try {
        if (typeof JSON.parse(str) === 'object') {
            return true;
        }
    } catch {
        return false;
    }
}

const errorRealTerminal = (ex: any) => {
    let message = ex.message;
    if (!message) message = 'disconnected';
    term.write(`\x1b[31m${message}\x1b[m\r\n`);
    console.log('err');
};

const closeRealTerminal = (ev: CloseEvent) => {
    term.write(ev.reason);
};

const initTerm = () => {
    let ifm = document.getElementById('terminal' + props.id) as HTMLInputElement | null;
    term = new Terminal({
        lineHeight: 1.2,
        fontSize: 12,
        fontFamily: "Monaco, Menlo, Consolas, 'Courier New', monospace",
        theme: {
            background: '#000000',
        },
        cursorBlink: true,
        cursorStyle: 'underline',
        scrollback: 100,
        tabStopWidth: 4,
        cols: ifm ? Math.floor(document.documentElement.clientWidth / 7) : 200,
        rows: ifm ? Math.floor(document.documentElement.clientHeight / 20) : 25,
    });
    if (ifm) {
        term.open(ifm);
        term.write('\n');
        terminalSocket = new WebSocket(
            `ws://localhost:9999/api/v1/terminals?id=${props.id}&cols=${term.cols}&rows=${term.rows}`,
        );
        terminalSocket.onopen = runRealTerminal;
        terminalSocket.onmessage = onWSReceive;
        terminalSocket.onclose = closeRealTerminal;
        terminalSocket.onerror = errorRealTerminal;
        term.onData((data: any) => {
            if (isWsOpen()) {
                terminalSocket.send(
                    JSON.stringify({
                        type: 'cmd',
                        cmd: Base64.encode(data),
                    }),
                );
            }
        });
        term.loadAddon(new AttachAddon(terminalSocket));
    }
};

const isWsOpen = () => {
    const readyState = terminalSocket && terminalSocket.readyState;
    return readyState === 1;
};

function onClose() {
    window.removeEventListener('resize', changeTerminalSize);
    terminalSocket && terminalSocket.close();
    term && term.dispose();
}

function changeTerminalSize() {
    let ifm = document.getElementById('terminal' + props.id) as HTMLInputElement | null;
    if (ifm) {
        ifm.style.height = document.documentElement.clientHeight - 300 + 'px';
        if (isWsOpen()) {
            terminalSocket.send(
                JSON.stringify({
                    type: 'resize',
                    cols: Math.floor(document.documentElement.clientWidth / 7),
                    rows: Math.floor(document.documentElement.clientHeight / 20),
                }),
            );
        }
    }
}

defineExpose({
    onClose,
    isWsOpen,
});

onMounted(() => {
    nextTick(() => {
        initTerm();
        changeTerminalSize();
        window.addEventListener('resize', changeTerminalSize);
    });
});

onBeforeUnmount(() => {
    onClose();
});
</script>
<style lang="scss" scoped>
#terminal {
    width: 100%;
    height: 100%;
}
</style>
