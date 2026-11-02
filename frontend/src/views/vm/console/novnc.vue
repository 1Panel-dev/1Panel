<template>
    <div class="vm-console">
        <div class="vm-console-toolbar">
            <el-dropdown
                trigger="click"
                placement="bottom-start"
                :disabled="!connected"
                @command="(shortcut) => sendKeys(shortcut.key, shortcut.modifiers)"
            >
                <el-button size="small" type="primary" plain :disabled="!connected">
                    {{ $t('xpack.vm.consoleShortcutKeys') }}
                    <el-icon class="vm-console-shortcut-arrow"><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                    <el-dropdown-menu>
                        <el-dropdown-item
                            v-for="shortcut in shortcutKeys"
                            :key="shortcut.label"
                            :command="shortcut"
                            :divided="shortcut.key === 'F1'"
                        >
                            {{ shortcut.label }}
                        </el-dropdown-item>
                    </el-dropdown-menu>
                </template>
            </el-dropdown>
            <div class="vm-console-custom-keys">
                <span class="vm-console-toolbar-label">{{ $t('xpack.vm.consoleCustomKeys') }}</span>
                <el-checkbox-group v-model="selectedModifiers" size="small" :disabled="!connected">
                    <el-checkbox-button v-for="key in modifierKeys" :key="key.code" :value="key.code">
                        {{ key.label }}
                    </el-checkbox-button>
                </el-checkbox-group>
                <el-select
                    v-model="selectedKey"
                    class="vm-console-key-select"
                    size="small"
                    filterable
                    :disabled="!connected"
                    :placeholder="$t('xpack.vm.consoleMainKey')"
                    :aria-label="$t('xpack.vm.consoleMainKey')"
                >
                    <el-option v-for="key in mainKeys" :key="key.code" :label="key.label" :value="key.code" />
                </el-select>
                <el-button size="small" type="primary" :disabled="!connected || !selectedKey" @click="sendKeys()">
                    {{ $t('xpack.vm.consoleSendKeys') }}
                </el-button>
            </div>
        </div>
        <div class="vm-console-viewport">
            <div
                ref="screenRef"
                class="vm-console-screen"
                @keydown.capture="trackPressedKey"
                @keyup.capture="pressedKeys.delete($event.code)"
                @focusout="releasePressedKeys"
            />
            <div v-if="statusText" class="vm-console-status" :class="{ 'is-error': statusError }">
                <span>{{ statusText }}</span>
                <el-button v-if="showReconnect" link type="primary" class="vm-console-reconnect" @click="reconnect">
                    {{ $t('commons.button.reconnect') }}
                </el-button>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, shallowRef } from 'vue';
import { useRoute } from 'vue-router';
import { ArrowDown } from '@element-plus/icons-vue';
import { MsgError } from '@/utils/message';
import i18n from '@/lang';

const props = defineProps({
    path: {
        type: String,
        default: '',
    },
});

interface NoVNCRFB {
    scaleViewport: boolean;
    resizeSession: boolean;
    addEventListener(type: string, listener: EventListenerOrEventListenerObject): void;
    disconnect(): void;
    sendKey(keysym: number, code: string, down?: boolean): void;
    focus(): void;
}

interface NoVNCRFBConstructor {
    new (target: HTMLElement, url: string): NoVNCRFB;
}

const route = useRoute();
const screenRef = ref<HTMLDivElement>();
const rfb = shallowRef<NoVNCRFB>();
const connected = ref(false);
const selectedModifiers = ref<string[]>([]);
const selectedKey = ref('');
const statusText = ref('');
const statusError = ref(false);
const connecting = ref(false);
const pressedKeys = new Set<string>();
let connectionGeneration = 0;
let screenReadyFrame: number | undefined;

const modifierKeys = [
    { label: 'Ctrl', code: 'ControlLeft', keysym: 0xffe3 },
    { label: 'Alt', code: 'AltLeft', keysym: 0xffe9 },
    { label: 'Shift', code: 'ShiftLeft', keysym: 0xffe1 },
    { label: 'Super / Win', code: 'MetaLeft', keysym: 0xffeb },
];
const mainKeys: { label: string; code: string; keysym: number; shiftKeysym?: number }[] = [
    { label: 'Delete', code: 'Delete', keysym: 0xffff },
    { label: 'Enter', code: 'Enter', keysym: 0xff0d },
    { label: 'Tab', code: 'Tab', keysym: 0xff09 },
    { label: 'Esc', code: 'Escape', keysym: 0xff1b },
    { label: 'Backspace', code: 'Backspace', keysym: 0xff08 },
    { label: 'Space', code: 'Space', keysym: 0x20 },
    { label: 'Insert', code: 'Insert', keysym: 0xff63 },
    { label: 'Home', code: 'Home', keysym: 0xff50 },
    { label: 'End', code: 'End', keysym: 0xff57 },
    { label: 'PageUp', code: 'PageUp', keysym: 0xff55 },
    { label: 'PageDown', code: 'PageDown', keysym: 0xff56 },
    { label: '↑', code: 'ArrowUp', keysym: 0xff52 },
    { label: '↓', code: 'ArrowDown', keysym: 0xff54 },
    { label: '←', code: 'ArrowLeft', keysym: 0xff51 },
    { label: '→', code: 'ArrowRight', keysym: 0xff53 },
    ...Array.from({ length: 12 }, (_, index) => ({
        label: `F${index + 1}`,
        code: `F${index + 1}`,
        keysym: 0xffbe + index,
    })),
    ...Array.from({ length: 26 }, (_, index) => ({
        label: String.fromCharCode(65 + index),
        code: `Key${String.fromCharCode(65 + index)}`,
        keysym: 97 + index,
        shiftKeysym: 65 + index,
    })),
    ...Array.from({ length: 10 }, (_, index) => ({
        label: String(index),
        code: `Digit${index}`,
        keysym: 48 + index,
        shiftKeysym: ')!@#$%^&*('[index].charCodeAt(0),
    })),
];

const shortcutKeys = [
    { label: 'Ctrl+Alt+Delete', key: 'Delete', modifiers: ['ControlLeft', 'AltLeft'] },
    { label: 'Alt+Tab', key: 'Tab', modifiers: ['AltLeft'] },
    { label: 'Ctrl+Shift+Esc', key: 'Escape', modifiers: ['ControlLeft', 'ShiftLeft'] },
    ...Array.from({ length: 7 }, (_, index) => ({
        label: `Ctrl+Alt+F${index + 1}`,
        key: `F${index + 1}`,
        modifiers: ['ControlLeft', 'AltLeft'],
    })),
];

const trackPressedKey = (event: KeyboardEvent) => {
    if (!connected.value || !event.code) return;
    if (event.code === 'CapsLock' && /mac|ipad|iphone|ipod/i.test(navigator.platform)) return;
    pressedKeys.add(event.code);
};

const releasePressedKeys = (event: FocusEvent) => {
    const target = event.target as HTMLElement;
    const codes = [...pressedKeys].reverse();
    pressedKeys.clear();
    for (const code of codes) {
        target.dispatchEvent(new KeyboardEvent('keyup', { code, bubbles: true }));
    }
};

const sendKeys = async (keyCode = selectedKey.value, modifierCodes = selectedModifiers.value) => {
    const client = rfb.value;
    if (!client || !connected.value) return;
    const key = mainKeys.find((item) => item.code === keyCode);
    if (!key) return;
    const modifiers = modifierKeys.filter((item) => modifierCodes.includes(item.code));
    const keysym = modifierCodes.includes('ShiftLeft') ? (key.shiftKeysym ?? key.keysym) : key.keysym;
    try {
        for (const modifier of modifiers) {
            client.sendKey(modifier.keysym, modifier.code, true);
        }
        client.sendKey(keysym, key.code);
    } finally {
        for (const modifier of modifiers.reverse()) {
            client.sendKey(modifier.keysym, modifier.code, false);
        }
    }
    await nextTick();
    if (rfb.value === client && connected.value) client.focus();
};

const websocketURL = computed(() => {
    const path = String(props.path || route.query.path || '').trim();
    if (!path) {
        return '';
    }
    if (path.startsWith('ws://') || path.startsWith('wss://')) {
        return path;
    }
    if (path.startsWith('http://') || path.startsWith('https://')) {
        const url = new URL(path);
        url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:';
        return url.toString();
    }
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    return `${protocol}//${window.location.host}/${path.replace(/^\//, '')}`;
});

const setStatus = (text: string, isError = false) => {
    if (screenReadyFrame !== undefined) {
        window.cancelAnimationFrame(screenReadyFrame);
        screenReadyFrame = undefined;
    }
    if (text) connected.value = false;
    statusText.value = text;
    statusError.value = isError;
};

const showReconnect = computed(() => Boolean(statusText.value) && !connecting.value);

const disconnect = () => {
    connected.value = false;
    pressedKeys.clear();
    connectionGeneration++;
    if (screenReadyFrame !== undefined) {
        window.cancelAnimationFrame(screenReadyFrame);
        screenReadyFrame = undefined;
    }
    const client = rfb.value;
    rfb.value = undefined;
    client?.disconnect();
};

const connect = async () => {
    disconnect();
    const generation = connectionGeneration;
    connecting.value = true;
    await nextTick();
    await new Promise((resolve) => window.requestAnimationFrame(resolve));

    if (generation !== connectionGeneration) {
        return;
    }
    const target = screenRef.value;
    const url = websocketURL.value;
    if (!target || !url) {
        connecting.value = false;
        const message = i18n.global.t('xpack.vm.consoleNotReady');
        setStatus(message, true);
        MsgError(message);
        return;
    }
    let RFB: NoVNCRFBConstructor;
    try {
        ({ default: RFB } = (await import('@novnc/novnc')) as { default: NoVNCRFBConstructor });
    } catch {
        if (generation === connectionGeneration) {
            connecting.value = false;
            setStatus(i18n.global.t('xpack.vm.consoleConnectFailed'), true);
        }
        return;
    }
    if (generation !== connectionGeneration || !screenRef.value) {
        return;
    }
    const client = new RFB(target, url);
    rfb.value = client;
    client.scaleViewport = true;
    client.resizeSession = true;
    const handleConnected = () => {
        if (rfb.value !== client) return;
        connecting.value = false;
        setStatus('');
    };
    client.addEventListener('connect', () => {
        if (rfb.value !== client) return;
        connected.value = true;
        handleConnected();
    });
    client.addEventListener('desktopname', handleConnected);
    const waitForScreenReady = () => {
        if (generation !== connectionGeneration || rfb.value !== client) return;
        const canvas = target.querySelector('canvas');
        if (canvas && canvas.width > 0 && canvas.height > 0) {
            handleConnected();
            return;
        }
        screenReadyFrame = window.requestAnimationFrame(waitForScreenReady);
    };
    screenReadyFrame = window.requestAnimationFrame(waitForScreenReady);
    client.addEventListener('disconnect', (event: CustomEvent<{ clean: boolean }>) => {
        if (rfb.value !== client) return;
        connecting.value = false;
        if (event.detail?.clean) {
            setStatus(i18n.global.t('xpack.vm.consoleDisconnected'));
            return;
        }
        setStatus(i18n.global.t('xpack.vm.consoleConnectFailed'), true);
    });
    client.addEventListener('securityfailure', (event: CustomEvent<{ reason?: string }>) => {
        if (rfb.value !== client) return;
        connecting.value = false;
        setStatus(event.detail?.reason || i18n.global.t('xpack.vm.consoleConnectFailed'), true);
    });
    client.addEventListener('credentialsrequired', () => {
        if (rfb.value !== client) return;
        connecting.value = false;
        setStatus(i18n.global.t('xpack.vm.consoleCredentialsRequired'), true);
    });
};

const reconnect = () => {
    connect();
};

onMounted(() => {
    connect();
});

onBeforeUnmount(() => {
    disconnect();
});
</script>

<style scoped lang="scss">
.vm-console {
    display: flex;
    flex-direction: column;
    position: relative;
    width: 100%;
    height: 100%;
    overflow: hidden;
    background: #111827;
}

.vm-console-toolbar {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 12px 16px;
    padding: 10px 12px;
    color: var(--el-text-color-primary);
    background: var(--el-bg-color);
    border-bottom: 1px solid var(--el-border-color);
}

.vm-console-shortcut-arrow {
    margin-left: 8px;
}

.vm-console-custom-keys {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 8px;
    padding-inline-start: 16px;
    border-inline-start: 1px solid var(--el-border-color);
}

.vm-console-toolbar-label {
    margin-right: 4px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
    white-space: nowrap;
}

.vm-console-key-select {
    width: 150px;
}

@media (max-width: 768px) {
    .vm-console-custom-keys {
        flex-basis: 100%;
        padding-inline-start: 0;
        border-inline-start: 0;
    }
}

.vm-console-viewport {
    position: relative;
    flex: 1;
    min-height: 0;
    overflow: hidden;
}

.vm-console-screen {
    width: 100%;
    height: 100%;
}

.vm-console-status {
    position: absolute;
    top: 16px;
    left: 50%;
    z-index: 1;
    max-width: calc(100% - 32px);
    padding: 8px 12px;
    color: #d1d5db;
    text-align: center;
    background: rgb(17 24 39 / 84%);
    border: 1px solid rgb(255 255 255 / 12%);
    border-radius: 4px;
    transform: translateX(-50%);
}

.vm-console-reconnect {
    margin-left: 8px;
}

.vm-console-status.is-error {
    color: #fecaca;
    border-color: rgb(248 113 113 / 45%);
}
</style>
