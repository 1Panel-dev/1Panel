<template>
    <div ref="terminalElement" class="terminal-container"></div>
</template>

<script lang="ts" setup>
import { ref, watch, onBeforeUnmount, nextTick, computed, onMounted } from 'vue';
import { Terminal } from '@xterm/xterm';
import '@xterm/xterm/css/xterm.css';
import { FitAddon } from '@xterm/addon-fit';
import { Base64 } from 'js-base64';
import { GlobalStore, TerminalStore } from '@/store';
const globalStore = GlobalStore();

const terminalElement = ref<HTMLDivElement | null>(null);
const fitAddon = new FitAddon();
const termReady = ref(false);
const webSocketReady = ref(false);
const term = ref();
const terminalSocket = ref<WebSocket>();
const heartbeatTimer = ref<NodeJS.Timer>();
const latency = ref(0);
const initCmd = ref('');
const currentLine = ref('');
const suggestionText = ref('');
const ghostText = ref('');
let suggestTimer: ReturnType<typeof setTimeout> | null = null;
const COMPLETION_DEBOUNCE_MS = 500;
const COMPLETION_MIN_CHARS = 2;

const readyWatcher = watch(
    () => webSocketReady.value && termReady.value,
    (ready) => {
        if (ready) {
            changeTerminalSize();
            readyWatcher(); // unwatch self
        }
    },
);

const terminalStore = TerminalStore();
const lineHeight = computed(() => terminalStore.lineHeight);
const fontSize = computed(() => terminalStore.fontSize);
const fontFamily = computed(() => terminalStore.fontFamily);
const backgroundColor = computed(() => terminalStore.backgroundColor);
const foregroundColor = computed(() => terminalStore.foregroundColor);
const letterSpacing = computed(() => terminalStore.letterSpacing);
watch(
    [lineHeight, fontSize, letterSpacing, fontFamily],
    ([newLineHeight, newFontSize, newLetterSpacing, newFontFamily]) => {
        if (!term.value) return;
        term.value.options.lineHeight = newLineHeight;
        term.value.options.letterSpacing = newLetterSpacing;
        term.value.options.fontSize = newFontSize;
        term.value.options.fontFamily = newFontFamily;
        changeTerminalSize();
    },
);
watch([backgroundColor, foregroundColor], ([newBackgroundColor, newForegroundColor]) => {
    if (!term.value) return;
    term.value.options.theme = {
        ...(term.value.options.theme || {}),
        background: newBackgroundColor,
        foreground: newForegroundColor,
    };
    applyTerminalBackground(newBackgroundColor);
});
const cursorStyle = computed(() => terminalStore.cursorStyle);
watch(cursorStyle, (newCursorStyle) => {
    if (!term.value) return;
    term.value.options.cursorStyle = newCursorStyle;
});
const cursorBlink = computed(() => terminalStore.cursorBlink);
watch(cursorBlink, (newCursorBlink) => {
    if (!term.value) return;
    term.value.options.cursorBlink = String(newCursorBlink).toLowerCase() === 'enable';
});
const scrollback = computed(() => terminalStore.scrollback);
watch(scrollback, (newScrollback) => {
    if (!term.value) return;
    term.value.options.scrollback = newScrollback;
});
const scrollSensitivity = computed(() => terminalStore.scrollSensitivity);
watch(scrollSensitivity, (newScrollSensitivity) => {
    if (!term.value) return;
    term.value.options.scrollSensitivity = newScrollSensitivity;
});

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
    const bg = terminalStore.backgroundColor || '#000000';
    const fg = terminalStore.foregroundColor || '#f5f5f5';
    term.value = new Terminal({
        lineHeight: terminalStore.lineHeight || 1.2,
        fontSize: terminalStore.fontSize || 12,
        fontFamily: terminalStore.fontFamily || "Monaco, Menlo, Consolas, 'Courier New', monospace",
        theme: {
            background: bg,
            foreground: fg,
        },
        cursorBlink: terminalStore.cursorBlink ? String(terminalStore.cursorBlink).toLowerCase() === 'enable' : true,
        cursorStyle: terminalStore.cursorStyle ? getStyle() : 'underline',
        scrollback: terminalStore.scrollback || 1000,
        scrollSensitivity: terminalStore.scrollSensitivity || 15,
    });
};

const applyTerminalBackground = (color: string) => {
    if (!terminalElement.value) return;
    terminalElement.value.style.backgroundColor = color || '#000000';
    terminalElement.value.style.backgroundImage = '';
    terminalElement.value.style.backgroundSize = '';
    terminalElement.value.style.backgroundPosition = '';
    terminalElement.value.style.backgroundRepeat = '';
    terminalElement.value.style.imageRendering = '';
};

const getStyle = (): 'underline' | 'block' | 'bar' => {
    switch (terminalStore.cursorStyle) {
        case 'bar':
            return 'bar';
        case 'block':
            return 'block';
        default:
            return 'underline';
    }
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
    if (terminalElement.value) {
        terminalElement.value.innerHTML = '';
    }
}

// terminal 相关代码 start

const initTerminal = (online: boolean = false): boolean => {
    newTerm();
    if (terminalElement.value) {
        term.value.open(terminalElement.value);
        applyTerminalBackground(terminalStore.backgroundColor);
        term.value.loadAddon(fitAddon);
        window.addEventListener('resize', changeTerminalSize);
        if (online) {
            term.value.onData((data) => onTermData(data));
        }
        termReady.value = true;
    }
    return termReady.value;
};

function changeTerminalSize() {
    if (!terminalElement.value || !term.value) return;
    if (terminalElement.value.clientWidth <= 0 || terminalElement.value.clientHeight <= 0) {
        return;
    }

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

// terminal 相关代码 end

// websocket 相关代码 start

const initWebSocket = (endpoint_: string, args: string = '') => {
    const href = window.location.href;
    const protocol = href.split('//')[0] === 'http:' ? 'ws' : 'wss';
    const host = href.split('//')[1].split('/')[0];
    const endpoint = endpoint_.replace(/^\/+/, '');
    let conn = `${protocol}://${host}/${endpoint}?cols=${term.value.cols}&rows=${term.value.rows}&${args}&operateNode=${globalStore.currentNode}`;
    if (args.indexOf('&operateNode=') !== -1) {
        conn = `${protocol}://${host}/${endpoint}?cols=${term.value.cols}&rows=${term.value.rows}&${args}`;
    }
    terminalSocket.value = new WebSocket(conn);
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
            clearGhost();
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
        case 'complete': {
            if (!currentLine.value || currentLine.value.trim().length === 0) {
                clearGhost();
                break;
            }
            if (wsMsg.data) {
                const raw = Base64.decode(wsMsg.data);
                const items = raw
                    .split('\n')
                    .map((item) => item.trim())
                    .filter((item) => item.length > 0);
                if (items.length >= 1) {
                    applySuggestion(items[0]);
                } else {
                    clearGhost();
                }
            } else {
                clearGhost();
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
        clearInterval(Number(heartbeatTimer.value));
    }
    term.value?.write('The connection has been disconnected.');
    term.value?.write(ev.reason);
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

function sendSuggestRequest(line: string) {
    if (!line || line.trim().length === 0) {
        return;
    }
    if (isWsOpen()) {
        terminalSocket.value!.send(
            JSON.stringify({
                type: 'complete',
                data: Base64.encode(line),
            }),
        );
    }
}

function scheduleSuggest() {
    if (suggestTimer) {
        clearTimeout(suggestTimer);
    }
    if (!currentLine.value || currentLine.value.trim().length === 0) {
        clearGhost();
        return;
    }
    const token = currentLine.value.trim().split(/\s+/).pop() || '';
    if (token.length < COMPLETION_MIN_CHARS) {
        clearGhost();
        return;
    }
    suggestTimer = setTimeout(() => {
        sendSuggestRequest(currentLine.value);
    }, COMPLETION_DEBOUNCE_MS);
}

function applySuggestion(raw: string) {
    if (!raw) {
        clearGhost();
        return;
    }
    const lastTokenMatch = currentLine.value.match(/(\S+)$/);
    const lastToken = lastTokenMatch ? lastTokenMatch[1] : '';
    let suffix = raw;
    if (lastToken && raw.startsWith(lastToken)) {
        suffix = raw.slice(lastToken.length);
    }
    if (!suffix) {
        clearGhost();
        return;
    }
    suggestionText.value = suffix;
    renderGhost(suffix);
}

function renderGhost(suffix: string) {
    if (!term.value) return;
    term.value.write('\x1b7');
    term.value.write('\x1b[0K');
    term.value.write(`\x1b[90m${suffix}\x1b[0m`);
    term.value.write('\x1b8');
    ghostText.value = suffix;
}

function clearGhost() {
    if (!ghostText.value || !term.value) return;
    term.value.write('\x1b7');
    term.value.write('\x1b[0K');
    term.value.write('\x1b8');
    ghostText.value = '';
    suggestionText.value = '';
}

function onTermData(data: string) {
    if (!data) return;
    if (data === '\t') {
        if (ghostText.value) {
            sendMsg(ghostText.value);
            currentLine.value += ghostText.value;
            clearGhost();
            scheduleSuggest();
            return;
        }
        sendMsg(data);
        return;
    }
    if (data === '\r' || data === '\n') {
        currentLine.value = '';
        clearGhost();
        sendMsg(data);
        return;
    }
    if (data === '\x7f') {
        if (currentLine.value.length > 0) {
            currentLine.value = currentLine.value.slice(0, -1);
        }
        clearGhost();
        sendMsg(data);
        scheduleSuggest();
        return;
    }
    if (data === '\x15') {
        currentLine.value = '';
        clearGhost();
        sendMsg(data);
        return;
    }
    if (data === '\x17') {
        currentLine.value = currentLine.value.replace(/\s+\S*$/, '');
        clearGhost();
        sendMsg(data);
        scheduleSuggest();
        return;
    }
    if (data.startsWith('\x1b')) {
        clearGhost();
        sendMsg(data);
        return;
    }
    currentLine.value += data;
    clearGhost();
    sendMsg(data);
    scheduleSuggest();
}

// websocket 相关代码 end

const resizeObserver = ref<ResizeObserver>();

onMounted(() => {
    // 使用 ResizeObserver 监听容器大小变化
    resizeObserver.value = new ResizeObserver(() => {
        if (termReady.value && webSocketReady.value) {
            changeTerminalSize();
        }
    });

    if (terminalElement.value) {
        resizeObserver.value.observe(terminalElement.value);
    }
});

defineExpose({
    acceptParams,
    onClose,
    isWsOpen,
    sendMsg,
    getLatency: () => latency.value,
});

onBeforeUnmount(() => {
    onClose();
    resizeObserver.value?.disconnect();
});
</script>

<style lang="scss" scoped>
.terminal-container {
    width: 100%;
    height: 100%;
}
:deep(.xterm) {
    padding: 5px !important;
    background-color: transparent !important;
}

:deep(.xterm .xterm-viewport) {
    background-color: transparent !important;
    scrollbar-width: thin;
    scrollbar-color: rgba(255, 255, 255, 0.3) rgba(255, 255, 255, 0.1);
}

:deep(.xterm .xterm-viewport::-webkit-scrollbar) {
    width: 10px;
    height: 10px;
    background: rgba(255, 255, 255, 0.1);
}

:deep(.xterm .xterm-viewport::-webkit-scrollbar-thumb) {
    border-radius: 6px;
    border: 2px solid transparent;
    background-clip: content-box;
    background-color: rgba(255, 255, 255, 0.3);
}

:deep(.xterm .xterm-viewport::-webkit-scrollbar-thumb:hover) {
    background-color: rgba(255, 255, 255, 0.45);
}

:deep(.xterm .xterm-viewport::-webkit-scrollbar-corner) {
    background: transparent;
}
</style>
