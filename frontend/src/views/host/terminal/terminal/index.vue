<template>
    <div :id="'terminal' + props.terminalID"></div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue';
import { Terminal } from 'xterm';
import { AttachAddon } from 'xterm-addon-attach';
import { Base64 } from 'js-base64';
import 'xterm/css/xterm.css';
import { FitAddon } from 'xterm-addon-fit';

interface WsProps {
    terminalID: string;
    wsID: number;
}
const props = withDefaults(defineProps<WsProps>(), {
    terminalID: '',
    wsID: 0,
});
const fitAddon = new FitAddon();
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
    let ifm = document.getElementById('terminal' + props.terminalID) as HTMLInputElement | null;
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
    });
    if (ifm) {
        term.open(ifm);
        if (props.wsID === 0) {
            terminalSocket = new WebSocket(
                `ws://localhost:9999/api/v1/terminals/local?cols=${term.cols}&rows=${term.rows}`,
            );
        } else {
            terminalSocket = new WebSocket(
                `ws://localhost:9999/api/v1/terminals?id=${props.wsID}&cols=${term.cols}&rows=${term.rows}`,
            );
        }
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
        term.loadAddon(fitAddon);
        setTimeout(() => {
            fitAddon.fit();
            if (isWsOpen()) {
                terminalSocket.send(
                    JSON.stringify({
                        type: 'resize',
                        cols: term.cols,
                        rows: term.rows,
                    }),
                );
            }
        }, 30);
    }
};

const fitTerm = () => {
    fitAddon.fit();
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

function onSendMsg(command: string) {
    terminalSocket.send(
        JSON.stringify({
            type: 'cmd',
            cmd: Base64.encode(command),
        }),
    );
}

function changeTerminalSize() {
    fitTerm();
    const { cols, rows } = term;
    if (isWsOpen()) {
        terminalSocket.send(
            JSON.stringify({
                type: 'resize',
                cols: cols,
                rows: rows,
            }),
        );
    }
}

defineExpose({
    onClose,
    isWsOpen,
    onSendMsg,
});

onMounted(() => {
    nextTick(() => {
        initTerm();
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
