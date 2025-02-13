<template>
    <div v-loading="firstLoading">
        <div v-if="defaultButton">
            <el-checkbox border v-model="tailLog" class="float-left" @change="changeTail(false)">
                {{ $t('commons.button.watch') }}
            </el-checkbox>
            <el-button class="ml-2.5" @click="onDownload" icon="Download" :disabled="logs.length === 0">
                {{ $t('file.download') }}
            </el-button>
            <span v-if="$slots.button" class="ml-2.5">
                <slot name="button"></slot>
            </span>
        </div>
        <div class="log-container" ref="logContainer" @scroll="onScroll">
            <div class="log-spacer" :style="{ height: `${totalHeight}px` }"></div>
            <div
                v-for="(log, index) in visibleLogs"
                :key="startIndex + index"
                class="log-item"
                :style="{ top: `${(startIndex + index) * logHeight}px` }"
            >
                <span>{{ log }}</span>
            </div>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { ReadByLine } from '@/api/modules/files';
import { ref, computed, onMounted, onUnmounted, watch, nextTick, reactive } from 'vue';
import { downloadFile } from '@/utils/util';

interface LogProps {
    id?: number;
    type: string;
    name?: string;
    tail?: boolean;
}
const props = defineProps({
    config: {
        type: Object as () => LogProps | null,
        default: () => ({
            id: 0,
            type: '',
            name: '',
            tail: false,
        }),
    },
    style: {
        type: String,
        default: 'height: calc(100vh - 200px); width: 100%; min-height: 400px; overflow: auto;',
    },
    defaultButton: {
        type: Boolean,
        default: true,
    },
    loading: {
        type: Boolean,
        default: true,
    },
    hasContent: {
        type: Boolean,
        default: false,
    },
});
const stopSignals = [
    'docker-compose up failed!',
    'docker-compose up successful!',
    'image build failed!',
    'image build successful!',
    'image pull failed!',
    'image pull successful!',
    'image push failed!',
    'image push successful!',
    'ollama pull failed!',
    'ollama pull successful!',
];
const emit = defineEmits(['update:loading', 'update:hasContent', 'update:isReading']);
const tailLog = ref(false);
const loading = ref(props.loading);
const readReq = reactive({
    id: 0,
    type: '',
    name: '',
    page: 1,
    pageSize: 500,
    latest: false,
});
const isLoading = ref(false);
const end = ref(false);
const lastLogs = ref([]);
const maxPage = ref(0);
const minPage = ref(0);
let timer: NodeJS.Timer | null = null;
const logPath = ref('');

const firstLoading = ref(false);
const logs = ref<string[]>([]);
const logContainer = ref<HTMLElement | null>(null);
const logHeight = 20;
const logCount = ref(0);
const totalHeight = computed(() => logHeight * logCount.value);
const containerHeight = ref(500);
const visibleCount = computed(() => Math.ceil(containerHeight.value / logHeight)); // 计算可见日志条数（容器高度 / 日志高度）
const startIndex = ref(0);

const visibleLogs = computed(() => {
    return logs.value.slice(startIndex.value, startIndex.value + visibleCount.value);
});

const onScroll = () => {
    if (logContainer.value) {
        const scrollTop = logContainer.value.scrollTop;
        if (scrollTop == 0) {
            readReq.page = minPage.value - 1;
            if (readReq.page < 1) {
                return;
            }
            minPage.value = readReq.page;
            getContent(true);
        }
        startIndex.value = Math.floor(scrollTop / logHeight);
    }
};

const changeLoading = () => {
    loading.value = !loading.value;
    emit('update:loading', loading.value);
};

const onDownload = async () => {
    changeLoading();
    downloadFile(logPath.value);
    changeLoading();
};

const changeTail = (fromOutSide: boolean) => {
    if (fromOutSide) {
        tailLog.value = !tailLog.value;
    }
    if (tailLog.value) {
        timer = setInterval(() => {
            getContent(false);
        }, 1000 * 3);
    } else {
        onCloseLog();
    }
};

const clearLog = (): void => {
    logs.value = [];
    readReq.page = 1;
    lastLogs.value = [];
};

const getContent = async (pre: boolean) => {
    if (isLoading.value) {
        return;
    }
    readReq.id = props.config.id;
    readReq.type = props.config.type;
    readReq.name = props.config.name;
    if (readReq.page < 1) {
        readReq.page = 1;
    }
    isLoading.value = true;
    emit('update:isReading', true);

    const res = await ReadByLine(readReq);
    logPath.value = res.data.path;
    firstLoading.value = false;

    if (!end.value && res.data.end) {
        lastLogs.value = [...logs.value];
    }
    if (res.data.lines && res.data.lines.length > 0) {
        res.data.lines = res.data.lines.map((line) =>
            line
                .replace(/\\u(\w{4})/g, function (match, grp) {
                    return String.fromCharCode(parseInt(grp, 16));
                })
                .replace(/\x1b\[[0-9;]*[A-Za-z?](?!\d)/g, ''),
        );
        const newLogs = res.data.lines;
        if (newLogs.length === readReq.pageSize && readReq.page < res.data.total) {
            readReq.page++;
        }
        if (
            readReq.type == 'php' &&
            logs.value.length > 0 &&
            newLogs.length > 0 &&
            newLogs[newLogs.length - 1] === logs.value[logs.value.length - 1]
        ) {
            isLoading.value = false;
            return;
        }

        if (stopSignals.some((signal) => newLogs[newLogs.length - 1].endsWith(signal))) {
            onCloseLog();
        }
        if (end.value) {
            if ((logs.value.length = 0)) {
                logs.value = newLogs;
            } else {
                logs.value = pre ? [...newLogs, ...lastLogs.value] : [...lastLogs.value, ...newLogs];
            }
        } else {
            if ((logs.value.length = 0)) {
                logs.value = newLogs;
            } else {
                logs.value = pre ? [...newLogs, ...logs.value] : [...logs.value, ...newLogs];
            }
        }

        nextTick(() => {
            if (pre) {
                logContainer.value.scrollTop = 2000;
            } else {
                logContainer.value.scrollTop = totalHeight.value;
                containerHeight.value = logContainer.value.getBoundingClientRect().height;
            }
        });
    }

    logCount.value = logs.value.length;
    end.value = res.data.end;
    emit('update:hasContent', logs.value.length > 0);
    if (readReq.latest) {
        readReq.page = res.data.total;
        readReq.latest = false;
        maxPage.value = res.data.total;
        minPage.value = res.data.total;
    }
    if (logs.value && logs.value.length > 3000) {
        if (pre) {
            logs.value.splice(logs.value.length - readReq.pageSize, readReq.pageSize);
            if (maxPage.value > 1) {
                maxPage.value--;
            }
        } else {
            logs.value.splice(0, readReq.pageSize);
            if (minPage.value > 1) {
                minPage.value++;
            }
        }
    }
    isLoading.value = false;
};

const onCloseLog = async () => {
    tailLog.value = false;
    clearInterval(Number(timer));
    timer = null;
    isLoading.value = false;
    emit('update:isReading', false);
};

watch(
    () => props.loading,
    (newLoading) => {
        loading.value = newLoading;
    },
);

const init = async () => {
    if (props.config.tail) {
        tailLog.value = props.config.tail;
    } else {
        tailLog.value = false;
    }
    if (tailLog.value) {
        changeTail(false);
    }
    readReq.latest = true;
    await getContent(false);
};

onMounted(async () => {
    firstLoading.value = true;
    await init();
    nextTick(() => {
        if (logContainer.value) {
            logContainer.value.scrollTop = totalHeight.value;
            containerHeight.value = logContainer.value.getBoundingClientRect().height;
        }
    });
});

onUnmounted(() => {
    onCloseLog();
});

defineExpose({ changeTail, onDownload, clearLog });
</script>

<style scoped>
.log-container {
    height: calc(100vh - 405px);
    overflow-y: auto;
    overflow-x: auto;
    position: relative;
    background-color: var(--panel-logs-bg-color);
    margin-top: 10px;
}

.log-spacer {
    position: relative;
    width: 100%;
}

.log-item {
    position: absolute;
    width: 100%;
    padding: 5px;
    color: #f5f5f5;
    box-sizing: border-box;
    white-space: nowrap;
}

.log-item span {
    font-size: 14px;
    font-weight: 300;
    color: #cccccc;
}
</style>
