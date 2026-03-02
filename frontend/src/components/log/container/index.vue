<template>
    <div v-if="showControl" class="log-toolbar">
        <el-select @change="searchLogs" class="fetchClass" v-model="logSearch.mode">
            <template #prefix>{{ $t('container.fetch') }}</template>
            <el-option v-for="item in timeOptions" :key="item.label" :value="item.value" :label="item.label" />
        </el-select>
        <el-select @change="searchLogs" class="tailClass" v-model.number="logSearch.tail">
            <template #prefix>{{ $t('container.lines') }}</template>
            <el-option :value="0" :label="$t('commons.table.all')" />
            <el-option :value="100" :label="100" />
            <el-option :value="200" :label="200" />
            <el-option :value="500" :label="500" />
            <el-option :value="1000" :label="1000" />
        </el-select>
        <div class="margin-button float-left">
            <el-checkbox border @change="searchLogs" v-model="logSearch.isWatch">
                {{ $t('commons.button.watch') }}
            </el-checkbox>
        </div>
        <div class="margin-button float-left">
            <el-checkbox border @change="searchLogs" v-model="logSearch.isShowTimestamp">
                {{ $t('commons.table.date') }}
            </el-checkbox>
        </div>
        <el-button class="margin-button" @click="openDownloadDialog" icon="Download">
            {{ $t('commons.button.download') }}
        </el-button>
        <el-button class="margin-button" @click="onClean" icon="Delete">
            {{ $t('commons.button.clean') }}
        </el-button>
    </div>
    <div class="log-container" :style="styleVars">
        <div class="xterm-log-viewer" ref="terminalElement"></div>
    </div>
    <el-dialog v-model="downloadDialogVisible" :title="$t('commons.button.download')" width="420px">
        <el-form label-position="top">
            <el-form-item :label="$t('container.fetch')">
                <el-select v-model="downloadForm.mode" class="w-full">
                    <el-option v-for="item in timeOptions" :key="item.label" :value="item.value" :label="item.label" />
                </el-select>
            </el-form-item>
            <el-form-item :label="$t('container.lines')">
                <el-select
                    v-model="downloadForm.tail"
                    class="w-full"
                    filterable
                    allow-create
                    default-first-option
                    :reserve-keyword="false"
                >
                    <el-option :value="0" :label="$t('commons.table.all')" />
                    <el-option :value="100" :label="100" />
                    <el-option :value="200" :label="200" />
                    <el-option :value="500" :label="500" />
                    <el-option :value="1000" :label="1000" />
                </el-select>
                <div class="download-tail-helper">{{ $t('container.downloadLinesHelper') }}</div>
            </el-form-item>
        </el-form>
        <template #footer>
            <el-button @click="downloadDialogVisible = false">{{ $t('commons.button.cancel') }}</el-button>
            <el-button type="primary" @click="onDownload">{{ $t('commons.button.confirm') }}</el-button>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { cleanComposeLog, cleanContainerLog, DownloadFile } from '@/api/modules/container';
import { FitAddon } from '@xterm/addon-fit';
import { Terminal } from '@xterm/xterm';
import '@xterm/xterm/css/xterm.css';
import i18n from '@/lang';
import { dateFormatForName } from '@/utils/util';
import { computed, nextTick, onMounted, onUnmounted, reactive, ref } from 'vue';
import { ElMessageBox } from 'element-plus';
import { MsgError, MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const em = defineEmits(['update:loading']);

const props = defineProps({
    container: {
        type: String,
        default: '',
    },
    compose: {
        type: String,
        default: '',
    },
    resource: {
        type: String,
        default: '',
    },
    highlightDiff: {
        type: Number,
        default: 320,
    },
    node: {
        type: String,
        default: '',
    },
    showControl: {
        type: Boolean,
        default: true,
    },
    defaultFollow: {
        type: Boolean,
        default: false,
    },
    defaultIsShowTimestamp: {
        type: Boolean,
        default: false,
    },
});

const styleVars = computed(() => ({
    '--custom-height': `${props.highlightDiff || 320}px`,
}));

const terminalElement = ref<HTMLDivElement | null>(null);
let eventSource: EventSource | null = null;
let term: Terminal | null = null;
const fitAddon = new FitAddon();
let onScrollDisposable: { dispose: () => void } | null = null;
const MAX_VIEW_LINES = 20000;
const followBottom = ref(true);

const logSearch = reactive({
    isWatch: props.defaultFollow ? true : true,
    isShowTimestamp: props.defaultIsShowTimestamp,
    container: '',
    mode: 'all',
    tail: props.defaultFollow ? 0 : 100,
    compose: '',
    resource: '',
});
const downloadDialogVisible = ref(false);
const downloadForm = reactive<{ mode: string; tail: number | string }>({
    mode: 'all',
    tail: 0,
});

const timeOptions = ref([
    { label: i18n.global.t('commons.table.all'), value: 'all' },
    {
        label: i18n.global.t('container.lastDay'),
        value: '24h',
    },
    {
        label: i18n.global.t('container.last4Hour'),
        value: '4h',
    },
    {
        label: i18n.global.t('container.lastHour'),
        value: '1h',
    },
    {
        label: i18n.global.t('container.last10Min'),
        value: '10m',
    },
]);

const stopListening = () => {
    if (eventSource) {
        eventSource.close();
        eventSource = null;
    }
};

const clearTerminal = () => {
    term?.reset();
    followBottom.value = true;
};

const writeLogLine = (data: string) => {
    if (!term) return;
    term.writeln(data);
    if (followBottom.value) {
        term.scrollToBottom();
    }
};

const bindXTermEvents = () => {
    if (!term) return;
    onScrollDisposable?.dispose();
    onScrollDisposable = term.onScroll(() => {
        if (!term) return;
        const active = term.buffer.active;
        followBottom.value = active.baseY + active.cursorY >= active.length - 2;
    });
};

const initTerminal = () => {
    if (!terminalElement.value || term) return;
    term = new Terminal({
        cursorBlink: false,
        cursorStyle: 'block',
        disableStdin: true,
        convertEol: true,
        scrollback: MAX_VIEW_LINES,
        fontSize: 14,
        fontFamily: "'JetBrains Mono', Monaco, Menlo, Consolas, 'Courier New', monospace",
        fontWeight: '500',
        lineHeight: 1.2,
        theme: {
            background: '#1e1e1e',
            foreground: '#666666',
            selectionBackground: 'rgba(102, 178, 255, 0.30)',
            selectionInactiveBackground: 'rgba(102, 178, 255, 0.20)',
        },
    });
    term.open(terminalElement.value);
    term.loadAddon(fitAddon);
    fitAddon.fit();
    bindXTermEvents();
};

const handleClose = async () => {
    stopListening();
};

const searchLogs = async () => {
    if (Number(logSearch.tail) < 0) {
        MsgError(i18n.global.t('container.linesHelper'));
        return;
    }
    stopListening();
    clearTerminal();

    let currentNode = globalStore.currentNode;
    if (props.node && props.node !== '') {
        currentNode = props.node;
    }

    let url = `/api/v2/containers/search/log?container=${logSearch.container}&since=${logSearch.mode}&tail=${logSearch.tail}&follow=${logSearch.isWatch}&timestamp=${logSearch.isShowTimestamp}&operateNode=${currentNode}`;
    if (logSearch.compose !== '') {
        url = `/api/v2/containers/search/log?compose=${logSearch.compose}&since=${logSearch.mode}&tail=${logSearch.tail}&follow=${logSearch.isWatch}&timestamp=${logSearch.isShowTimestamp}&operateNode=${currentNode}`;
    }

    eventSource = new EventSource(url);
    eventSource.onmessage = (event: MessageEvent) => {
        writeLogLine(event.data);
    };
    eventSource.onerror = (event: MessageEvent) => {
        stopListening();
        if (event.data && event.data != '') {
            MsgError(event.data);
        }
    };
};

const openDownloadDialog = () => {
    downloadForm.mode = logSearch.mode;
    downloadForm.tail = logSearch.tail;
    downloadDialogVisible.value = true;
};

const onDownload = async () => {
    const customTail = Number(downloadForm.tail);
    if (Number.isNaN(customTail) || customTail < 0) {
        MsgError(i18n.global.t('container.linesHelper'));
        return;
    }
    const container = logSearch.compose === '' ? logSearch.container : logSearch.compose;
    let resource = container;
    if (props.resource) {
        resource = props.resource;
    }
    const containerType = logSearch.compose === '' ? 'container' : 'compose';
    const params = {
        container: container,
        since: downloadForm.mode,
        tail: customTail,
        timestamp: logSearch.isShowTimestamp,
        containerType: containerType,
    };
    const addItem = {};
    addItem['name'] = resource + '-' + dateFormatForName(new Date()) + '.log';
    DownloadFile(params).then((res) => {
        const downloadUrl = window.URL.createObjectURL(new Blob([res]));
        const a = document.createElement('a');
        a.style.display = 'none';
        a.href = downloadUrl;
        a.download = addItem['name'];
        const event = new MouseEvent('click');
        a.dispatchEvent(event);
    });
    downloadDialogVisible.value = false;
};

const onClean = async () => {
    ElMessageBox.confirm(i18n.global.t('container.cleanLogHelper'), i18n.global.t('container.cleanLog'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        let currentNode = globalStore.currentNode;
        if (props.node && props.node !== '') {
            currentNode = props.node;
        }
        if (logSearch.compose !== '') {
            em('update:loading', true);
            await cleanComposeLog(logSearch.resource, logSearch.compose, currentNode)
                .then(() => {
                    em('update:loading', false);
                    searchLogs();
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                })
                .finally(() => {
                    em('update:loading', false);
                });
            return;
        }
        await cleanContainerLog(logSearch.container, currentNode);
        searchLogs();
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    });
};

const resizeObserver = ref<ResizeObserver | null>(null);

onMounted(() => {
    logSearch.container = props.container;
    logSearch.compose = props.compose;
    logSearch.resource = props.resource;

    logSearch.tail = 100;
    logSearch.mode = 'all';
    logSearch.isWatch = true;

    nextTick(() => {
        initTerminal();
        if (terminalElement.value) {
            resizeObserver.value = new ResizeObserver(() => {
                fitAddon.fit();
            });
            resizeObserver.value.observe(terminalElement.value);
        }
        searchLogs();
    });
});

onUnmounted(() => {
    handleClose();
    onScrollDisposable?.dispose();
    if (term) {
        term.dispose();
        term = null;
    }
    resizeObserver.value?.disconnect();
});
</script>

<style scoped lang="scss">
.margin-button {
    margin-left: 0;
}
.fullScreen {
    border: none;
}
.tailClass {
    width: 160px;
}
.fetchClass {
    width: 220px;
}

.log-toolbar {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 10px;
}

.log-toolbar :deep(.el-button),
.log-toolbar :deep(.el-checkbox) {
    white-space: nowrap;
    flex-shrink: 0;
}

.download-tail-helper {
    margin-top: 6px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
}

.log-container {
    height: calc(100vh - var(--custom-height, 320px));
    overflow: hidden;
    position: relative;
    background-color: #1e1e1e;
    margin-top: 10px;
}

.xterm-log-viewer {
    width: 100%;
    height: 100%;
}

:deep(.xterm) {
    padding: 2px !important;
}
</style>
