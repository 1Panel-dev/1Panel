<template>
    <div class="terminal-shell">
        <div ref="terminalElement" class="terminal-container"></div>
        <transition name="ai-mask-fade">
            <div v-if="aiNotice.loading" class="ai-notice-mask"></div>
        </transition>
        <transition name="ai-notice-fade">
            <div
                v-if="aiNotice.visible"
                class="ai-notice"
                :class="[`ai-notice--${aiNotice.level}`, { 'ai-notice--loading': aiNotice.loading }]"
            >
                {{ aiNotice.message }}
            </div>
        </transition>
    </div>
</template>

<script lang="ts" setup>
import { ref, watch, onBeforeUnmount, nextTick, computed, onMounted } from 'vue';
import { Terminal } from '@xterm/xterm';
import '@xterm/xterm/css/xterm.css';
import { FitAddon } from '@xterm/addon-fit';
import { decodeBase64, encodeBase64 } from '@/utils/base64';
import { TerminalStore } from '@/store';
import { MsgError } from '@/utils/message';
import { checkStreamAuth } from '@/utils/stream-auth';
import { useGlobalStore } from '@/composables/useGlobalStore';
const { currentNode } = useGlobalStore();

const terminalElement = ref<HTMLDivElement | null>(null);
const fitAddon = new FitAddon();
const termReady = ref(false);
const webSocketReady = ref(false);
const term = ref();
const terminalSocket = ref<WebSocket>();
const heartbeatTimer = ref<NodeJS.Timer>();
let initWebSocketToken = 0;
const latency = ref(0);
const initCmd = ref('');
const hideInitCmdEcho = ref(false);
const initCmdEchoBuffer = ref('');
const waitForPrompt = ref('');
const waitForPromptBuffer = ref('');
const aiNotice = ref({
    visible: false,
    loading: false,
    level: 'info',
    message: '',
});
let aiNoticeTimer: ReturnType<typeof setTimeout> | null = null;

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
    waitForPrompt?: string;
}

interface TerminalBufferLine {
    isWrapped?: boolean;
    translateToString(trimRight?: boolean, startColumn?: number, endColumn?: number): string;
}
const acceptParams = (props: WsProps) => {
    nextTick(() => {
        if (props.error.length !== 0) {
            initError(props.error);
        } else {
            initCmd.value = props.initCmd || '';
            waitForPrompt.value = props.waitForPrompt || '';
            waitForPromptBuffer.value = '';
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
        scrollSensitivity: terminalStore.scrollSensitivity || 6,
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
    initWebSocketToken++;
    window.removeEventListener('resize', changeTerminalSize);
    clearAINotice();
    webSocketReady.value = false;
    try {
        terminalSocket.value?.close();
    } catch {}
    if (heartbeatTimer.value) {
        clearInterval(Number(heartbeatTimer.value));
        heartbeatTimer.value = undefined;
    }
    terminalSocket.value = undefined;
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

const initWebSocket = async (endpoint_: string, args: string = '') => {
    const token = ++initWebSocketToken;
    const href = window.location.href;
    const protocol = href.split('//')[0] === 'http:' ? 'ws' : 'wss';
    const host = href.split('//')[1].split('/')[0];
    const endpoint = endpoint_.replace(/^\/+/, '');
    let node = args.indexOf('id=') !== -1 ? 'local' : currentNode.value;
    let conn = `${protocol}://${host}/${endpoint}?cols=${term.value.cols}&rows=${term.value.rows}&${args}&operateNode=${node}`;
    if (args.indexOf('operateNode=') !== -1) {
        conn = `${protocol}://${host}/${endpoint}?cols=${term.value.cols}&rows=${term.value.rows}&${args}`;
    }
    const authError = await checkStreamAuth(conn);
    if (token !== initWebSocketToken || !termReady.value) {
        return;
    }
    if (authError) {
        showWebSocketAuthError(authError);
        return;
    }
    if (heartbeatTimer.value) {
        clearInterval(Number(heartbeatTimer.value));
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

const showWebSocketAuthError = (message: string) => {
    clearAINotice();
    MsgError(message);
    term.value?.write(`\x1b[31m${message}\x1b[m\r\n`);
};

const runRealTerminal = () => {
    webSocketReady.value = true;
    if (initCmd.value !== '') {
        hideInitCmdEcho.value = true;
        initCmdEchoBuffer.value = '';
        sendMsg(initCmd.value);
    }
};

const stripInitCmdEchoLine = (message: string) => {
    if (!hideInitCmdEcho.value) {
        return message;
    }
    initCmdEchoBuffer.value += message;
    const lineBreakIndex = initCmdEchoBuffer.value.search(/\r?\n/);
    if (lineBreakIndex === -1) {
        return '';
    }

    const lineBreakLength = initCmdEchoBuffer.value[lineBreakIndex] === '\r' ? 2 : 1;
    const remaining = initCmdEchoBuffer.value.slice(lineBreakIndex + lineBreakLength);
    hideInitCmdEcho.value = false;
    initCmdEchoBuffer.value = '';
    initCmd.value = '';
    return remaining;
};

const flushPromptBuffer = (message: string) => {
    if (!waitForPrompt.value) {
        return message;
    }
    waitForPromptBuffer.value += message;
    const promptIndex = waitForPromptBuffer.value.indexOf(waitForPrompt.value);
    if (promptIndex === -1) {
        return '';
    }

    const visible = waitForPromptBuffer.value.slice(promptIndex);
    waitForPrompt.value = '';
    waitForPromptBuffer.value = '';
    return visible;
};

const onWSReceive = (message: MessageEvent) => {
    const wsMsg = JSON.parse(message.data);
    switch (wsMsg.type) {
        case 'cmd': {
            term.value.element && term.value.focus();
            if (wsMsg.data) {
                let receiveMsg = decodeBase64(wsMsg.data);
                if (hideInitCmdEcho.value) {
                    receiveMsg = stripInitCmdEchoLine(receiveMsg);
                }
                if (receiveMsg && waitForPrompt.value) {
                    receiveMsg = flushPromptBuffer(receiveMsg);
                }
                if (!receiveMsg) {
                    break;
                }
                term.value.write(receiveMsg);
            }
            break;
        }
        case 'heartbeat': {
            latency.value = new Date().getTime() - wsMsg.timestamp;
            break;
        }
        case 'ai_notice': {
            const message = wsMsg.message?.trim();
            if (!message) {
                break;
            }
            showAINotice(wsMsg.level || 'info', message);
            break;
        }
    }
};

const errorRealTerminal = (ex: any) => {
    clearAINotice();
    let message = ex.message;
    if (!message) message = 'disconnected';
    term.value.write(`\x1b[31m${message}\x1b[m\r\n`);
};

const closeRealTerminal = (ev: CloseEvent) => {
    clearAINotice();
    webSocketReady.value = false;
    if (heartbeatTimer.value) {
        clearInterval(Number(heartbeatTimer.value));
        heartbeatTimer.value = undefined;
    }
    terminalSocket.value = undefined;
    term.value?.write('The connection has been disconnected.');
    term.value?.write(ev.reason);
};

const isWsOpen = () => {
    const readyState = terminalSocket.value && terminalSocket.value.readyState;
    return readyState === 1;
};

function isEnterInputData(data: string): boolean {
    return data === '\r' || data === '\n' || data === '\r\n';
}

function getCurrentTerminalLine(): string {
    const xterm = term.value;
    if (!xterm?.buffer?.active) return '';
    const buffer = xterm.buffer.active;
    const cursorRow = buffer.baseY + buffer.cursorY;
    let startRow = cursorRow;
    let endRow = cursorRow;

    for (let row = cursorRow; row > 0; row--) {
        const line = buffer.getLine(row) as TerminalBufferLine | undefined;
        if (!line?.isWrapped) {
            startRow = row;
            break;
        }
        startRow = row - 1;
    }

    for (let row = cursorRow + 1; row < buffer.length; row++) {
        const line = buffer.getLine(row) as TerminalBufferLine | undefined;
        if (!line?.isWrapped) {
            break;
        }
        endRow = row;
    }

    let content = '';
    for (let row = startRow; row <= endRow; row++) {
        const line = buffer.getLine(row) as TerminalBufferLine | undefined;
        if (!line) continue;
        content += line.translateToString(false);
    }
    return content.trimEnd();
}

function sendMsg(data: string, line: string = '') {
    if (isWsOpen()) {
        terminalSocket.value!.send(
            JSON.stringify({
                type: 'cmd',
                data: encodeBase64(data),
                line,
            }),
        );
    }
}

function onTermData(data: string) {
    if (!data) return;
    if (aiNotice.value.loading) return;
    sendMsg(data, isEnterInputData(data) ? getCurrentTerminalLine() : '');
}

function clearAINotice() {
    if (aiNoticeTimer) {
        clearTimeout(aiNoticeTimer);
        aiNoticeTimer = null;
    }
    aiNotice.value = {
        ...aiNotice.value,
        visible: false,
        loading: false,
    };
}

function showAINotice(level: string, message: string) {
    if (aiNoticeTimer) {
        clearTimeout(aiNoticeTimer);
        aiNoticeTimer = null;
    }
    const resolvedLevel = ['success', 'error', 'info'].includes(level) ? level : 'info';
    aiNotice.value = {
        visible: true,
        loading: resolvedLevel === 'info',
        level: resolvedLevel,
        message,
    };
    if (resolvedLevel === 'info') {
        return;
    }
    aiNoticeTimer = setTimeout(() => {
        aiNotice.value = {
            ...aiNotice.value,
            visible: false,
            loading: false,
        };
        aiNoticeTimer = null;
    }, 2600);
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

.terminal-shell {
    position: relative;
    width: 100%;
    height: 100%;
}

.ai-notice-mask {
    position: absolute;
    inset: 0;
    z-index: 10;
    background: rgba(8, 10, 14, 0.12);
    backdrop-filter: blur(1.5px);
    pointer-events: auto;
    cursor: progress;
}

.ai-notice {
    position: absolute;
    left: 50%;
    top: 24px;
    transform: translateX(-50%);
    z-index: 12;
    width: fit-content;
    min-width: 240px;
    max-width: min(72%, 560px);
    padding: 9px 14px;
    border-radius: 999px;
    border: 1px solid rgba(255, 255, 255, 0.16);
    background: rgba(16, 18, 24, 0.78);
    color: #f3f4f6;
    font-size: 12px;
    line-height: 1.4;
    text-align: center;
    box-shadow: 0 10px 28px rgba(0, 0, 0, 0.2);
    backdrop-filter: blur(8px);
    pointer-events: none;
    white-space: pre-wrap;
}

.ai-notice--loading {
    top: 50%;
    width: min(72%, 560px);
    padding: 12px 16px;
    border-radius: 12px;
    font-size: 13px;
    line-height: 1.5;
    transform: translate(-50%, -50%);
    background: rgba(16, 18, 24, 0.92);
    box-shadow: 0 16px 40px rgba(0, 0, 0, 0.3);
    backdrop-filter: blur(10px);
}

.ai-notice--success {
    border-color: rgba(34, 197, 94, 0.45);
    background: rgba(10, 28, 18, 0.78);
}

.ai-notice--error {
    border-color: rgba(248, 113, 113, 0.45);
    background: rgba(40, 16, 16, 0.8);
}

.ai-notice-fade-enter-active,
.ai-notice-fade-leave-active {
    transition:
        opacity 180ms ease,
        transform 180ms ease;
}

.ai-mask-fade-enter-active,
.ai-mask-fade-leave-active {
    transition: opacity 180ms ease;
}

.ai-notice-fade-enter-from,
.ai-notice-fade-leave-to {
    opacity: 0;
    transform: translateX(-50%) translateY(-6px);
}

.ai-notice--loading.ai-notice-fade-enter-from,
.ai-notice--loading.ai-notice-fade-leave-to {
    transform: translate(-50%, calc(-50% + 8px));
}

.ai-mask-fade-enter-from,
.ai-mask-fade-leave-to {
    opacity: 0;
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
