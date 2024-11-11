<template>
    <div ref="terminalElement" @wheel="onTermWheel"></div>
</template>

<script lang="ts" setup>
import { ref, watch, onBeforeUnmount, nextTick } from 'vue';
import { Terminal } from '@xterm/xterm';
import '@xterm/xterm/css/xterm.css';
import { FitAddon } from '@xterm/addon-fit';
import { Base64 } from 'js-base64';

const terminalElement = ref<HTMLDivElement | null>(null);
const fitAddon = new FitAddon();
const termReady = ref(false);
const webSocketReady = ref(false);
const term = ref();
const terminalSocket = ref<WebSocket>();
const heartbeatTimer = ref<NodeJS.Timer>();
const latency = ref(0);
const initCmd = ref('');

const readyWatcher = watch(
    () => webSocketReady.value && termReady.value,
    (ready) => {
        if (ready) {
            changeTerminalSize();
            readyWatcher(); // unwatch self
        }
    },
);

interface WsProps {
    endpoint: string;
    args: string;
    error: string;
    initCmd: string;
}
const acceptParams = (props: WsProps) => {
    nextTick(() => {
        if (props.error.length !== 0) {
            initError(props.error);
        } else {
            initCmd.value = props.initCmd || '';
            init(props.endpoint, props.args);
        }
    });
};

const newTerm = () => {
    const background = getComputedStyle(document.documentElement).getPropertyValue('--panel-terminal-bg-color').trim();
    term.value = new Terminal({
        lineHeight: 1.2,
        fontSize: 12,
        fontFamily: "Monaco, Menlo, Consolas, 'Courier New', monospace",
        theme: {
            background: background,
        },
        cursorBlink: true,
        cursorStyle: 'underline',
        scrollback: 1000,
        scrollSensitivity: 15,
        tabStopWidth: 4,
    });
};

const init = (endpoint: string, args: string) => {
    if (initTerminal(true)) {
        initWebSocket(endpoint, args);
    }
};

const initError = (errorInfo: string) => {
    if (initTerminal(false)) {
        term.value.write(errorInfo);
    }
};

function onClose(isKeepShow: boolean = false) {
    window.removeEventListener('resize', changeTerminalSize);
    try {
        terminalSocket.value?.close();
    } catch {}
    if (!isKeepShow) {
        try {
            term.value.dispose();
        } catch {}
    }
    terminalElement.value.innerHTML = '';
}

// terminal 相关代码 start

const initTerminal = (online: boolean = false): boolean => {
    newTerm();
    if (terminalElement.value) {
        term.value.open(terminalElement.value);
        term.value.loadAddon(fitAddon);
        window.addEventListener('resize', changeTerminalSize);
        if (online) {
            term.value.onData((data) => sendMsg(data));
        }
        termReady.value = true;
    }
    return termReady.value;
};

function changeTerminalSize() {
    fitAddon.fit();
    if (isWsOpen()) {
        const { cols, rows } = term.value;
        terminalSocket.value!.send(
            JSON.stringify({
                type: 'resize',
                cols: cols,
                rows: rows,
            }),
        );
    }
}

/**
 * Support for Ctrl+MouseWheel to scaling fonts
 * @param event WheelEvent
 */
const onTermWheel = (event: WheelEvent) => {
    if (event.ctrlKey) {
        event.preventDefault();
        if (event.deltaY > 0) {
            // web font-size mini 12px
            if (term.value.options.fontSize > 12) {
                term.value.options.fontSize = term.value.options.fontSize - 1;
            }
        } else {
            term.value.options.fontSize = term.value.options.fontSize + 1;
        }
        changeTerminalSize();
    }
};

// terminal 相关代码 end

// websocket 相关代码 start

const initWebSocket = (endpoint_: string, args: string = '') => {
    const href = window.location.href;
    const protocol = href.split('//')[0] === 'http:' ? 'ws' : 'wss';
    const host = href.split('//')[1].split('/')[0];
    const endpoint = endpoint_.replace(/^\/+/, '');
    terminalSocket.value = new WebSocket(
        `${protocol}://${host}/${endpoint}?cols=${term.value.cols}&rows=${term.value.rows}&${args}`,
    );
    terminalSocket.value.onopen = runRealTerminal;
    terminalSocket.value.onmessage = onWSReceive;
    terminalSocket.value.onclose = closeRealTerminal;
    terminalSocket.value.onerror = errorRealTerminal;
    heartbeatTimer.value = setInterval(() => {
        if (isWsOpen()) {
            terminalSocket.value!.send(
                JSON.stringify({
                    type: 'heartbeat',
                    timestamp: `${new Date().getTime()}`,
                }),
            );
        }
    }, 1000 * 10);
};

const runRealTerminal = () => {
    webSocketReady.value = true;
    if (initCmd.value !== '') {
        sendMsg(initCmd.value);
    }
};

const onWSReceive = (message: MessageEvent) => {
    const wsMsg = JSON.parse(message.data);
    switch (wsMsg.type) {
        case 'cmd': {
            term.value.element && term.value.focus();
            if (wsMsg.data) {
                let receiveMsg = Base64.decode(wsMsg.data);
                if (initCmd.value != '') {
                    receiveMsg = receiveMsg?.replace(initCmd.value.trim(), '').trim();
                    initCmd.value = '';
                }
                term.value.write(receiveMsg);
            }
            break;
        }
        case 'heartbeat': {
            latency.value = new Date().getTime() - wsMsg.timestamp;
            break;
        }
    }
};

const errorRealTerminal = (ex: any) => {
    let message = ex.message;
    if (!message) message = 'disconnected';
    term.value.write(`\x1b[31m${message}\x1b[m\r\n`);
};

const closeRealTerminal = (ev: CloseEvent) => {
    if (heartbeatTimer.value) {
        clearInterval(heartbeatTimer.value);
    }
    term.value.write('The connection has been disconnected.');
    term.value.write(ev.reason);
};

const isWsOpen = () => {
    const readyState = terminalSocket.value && terminalSocket.value.readyState;
    return readyState === 1;
};

function sendMsg(data: string) {
    if (isWsOpen()) {
        terminalSocket.value!.send(
            JSON.stringify({
                type: 'cmd',
                data: Base64.encode(data),
            }),
        );
    }
}

// websocket 相关代码 end

defineExpose({
    acceptParams,
    onClose,
    isWsOpen,
    sendMsg,
    getLatency: () => latency.value,
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
:deep(.xterm) {
    padding: 5px !important;
}
</style>
