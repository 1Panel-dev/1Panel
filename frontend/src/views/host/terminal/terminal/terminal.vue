<template>
    <div :id="'terminal-' + terminalID" @wheel="onTermWheel"></div>
</template>

<script setup lang="ts">
import { ref, nextTick, onBeforeUnmount, watch } from 'vue';
import { Terminal } from 'xterm';
import { AttachAddon } from 'xterm-addon-attach';
import { Base64 } from 'js-base64';
import 'xterm/css/xterm.css';
import { FitAddon } from 'xterm-addon-fit';
import { isJson } from '@/utils/util';

const terminalID = ref();
const wsID = ref();
interface WsProps {
    terminalID: string;
    wsID: number;
    error: string;
}
const acceptParams = (props: WsProps) => {
    terminalID.value = props.terminalID;
    wsID.value = props.wsID;
    nextTick(() => {
        if (props.error.length !== 0) {
            initErrorTerm(props.error);
        } else {
            initTerm();
            window.addEventListener('resize', changeTerminalSize);
        }
    });
};

const fitAddon = new FitAddon();
const webSocketReady = ref(false);
const termReady = ref(false);
const terminalSocket = ref<WebSocket>();
const term = ref<Terminal>();

const readyWatcher = watch(
    () => webSocketReady.value && termReady.value,
    (ready) => {
        if (ready) {
            changeTerminalSize();
            readyWatcher(); // unwatch self
        }
    },
);

const runRealTerminal = () => {
    webSocketReady.value = true;
};

const onWSReceive = (message: any) => {
    if (!isJson(message.data)) {
        return;
    }
    const data = JSON.parse(message.data);
    if (term.value) {
        term.value.element && term.value.focus();
        term.value.write(data.Data);
    }
};

const errorRealTerminal = (ex: any) => {
    let message = ex.message;
    if (!message) message = 'disconnected';
    if (term.value) {
        term.value.write(`\x1b[31m${message}\x1b[m\r\n`);
    }
};

const closeRealTerminal = (ev: CloseEvent) => {
    if (term.value) {
        term.value.write(ev.reason);
    }
};

const initErrorTerm = (errorInfo: string) => {
    let ifm = document.getElementById('terminal-' + terminalID.value) as HTMLInputElement | null;
    term.value = new Terminal({
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
        term.value.open(ifm);
        term.value.write(errorInfo);
        term.value.loadAddon(fitAddon);
        fitAddon.fit();
        termReady.value = true;
    }
};

const initTerm = () => {
    let ifm = document.getElementById('terminal-' + terminalID.value) as HTMLInputElement | null;
    let href = window.location.href;
    let protocol = href.split('//')[0] === 'http:' ? 'ws' : 'wss';
    let ipLocal = href.split('//')[1].split('/')[0];
    term.value = new Terminal({
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
        term.value.open(ifm);
        terminalSocket.value = new WebSocket(
            `${protocol}://${ipLocal}/api/v1/terminals?id=${wsID.value}&cols=${term.value.cols}&rows=${term.value.rows}`,
        );
        terminalSocket.value.onopen = runRealTerminal;
        terminalSocket.value.onmessage = onWSReceive;
        terminalSocket.value.onclose = closeRealTerminal;
        terminalSocket.value.onerror = errorRealTerminal;
        term.value.onData((data: any) => {
            if (isWsOpen()) {
                terminalSocket.value!.send(
                    JSON.stringify({
                        type: 'cmd',
                        cmd: Base64.encode(data),
                    }),
                );
            }
        });
        term.value.loadAddon(new AttachAddon(terminalSocket.value));
        term.value.loadAddon(fitAddon);
        termReady.value = true;
    }
};

const fitTerm = () => {
    fitAddon.fit();
};

const isWsOpen = () => {
    const readyState = terminalSocket.value && terminalSocket.value.readyState;
    return readyState === 1;
};

function onClose() {
    window.removeEventListener('resize', changeTerminalSize);
    try {
        terminalSocket.value?.close();
    } catch {}
    try {
        term.value?.dispose();
    } catch {}
}

function onSendMsg(command: string) {
    terminalSocket.value?.send(
        JSON.stringify({
            type: 'cmd',
            cmd: Base64.encode(command),
        }),
    );
}

function changeTerminalSize() {
    fitTerm();
    const { cols, rows } = term.value!;
    if (isWsOpen()) {
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
        if (term.value) {
            if (event.deltaY > 0) {
                // web font-size mini 12px
                if (term.value.options.fontSize > 12) {
                    term.value.options.fontSize = term.value.options.fontSize - 1;
                }
            } else {
                term.value.options.fontSize = term.value.options.fontSize + 1;
            }
        }
    }
};

defineExpose({
    acceptParams,
    onClose,
    isWsOpen,
    onSendMsg,
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
