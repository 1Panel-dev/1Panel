<template>
    <div>
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
        <el-button class="margin-button" @click="onDownload" icon="Download">
            {{ $t('file.download') }}
        </el-button>
        <el-button class="margin-button" @click="onClean" icon="Delete">
            {{ $t('commons.button.clean') }}
        </el-button>
    </div>
    <!-- <div class="log-container" ref="logContainer">
        <DynamicScroller :items="logs" :min-item-size="32" v-if="logs.length">
            <template #default="{ item, active }">
                <DynamicScrollerItem :item="item" :active="active" :size-dependencies="[item]" :data-index="item">
                    <hightlight :log="item" type="container"></hightlight>
                </DynamicScrollerItem>
            </template>
        </DynamicScroller>
    </div> -->

    <div class="log-container" ref="logContainer">
        <div class="log-spacer" :style="{ height: `${totalHeight}px` }"></div>
        <div
            v-for="(log, index) in visibleLogs"
            :key="startIndex + index"
            class="log-item"
            :style="{ top: `${(startIndex + index) * logHeight}px` }"
        >
            <hightlight :log="log" type="container"></hightlight>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { cleanContainerLog, DownloadFile } from '@/api/modules/container';
import i18n from '@/lang';
import { dateFormatForName } from '@/utils/util';
import { onUnmounted, reactive, ref } from 'vue';
import { ElMessageBox } from 'element-plus';
import { MsgError, MsgSuccess } from '@/utils/message';
import hightlight from '@/components/hightlight/index.vue';

const props = defineProps({
    container: {
        type: String,
        default: '',
    },
    compose: {
        type: String,
        default: '',
    },
});

const logVisible = ref(false);
const logContainer = ref<HTMLElement | null>(null);
const logs = ref<string[]>([]);
let eventSource: EventSource | null = null;
const logSearch = reactive({
    isWatch: true,
    container: '',
    mode: 'all',
    tail: 100,
    compose: '',
});
const logHeight = 20;
const logCount = ref(0);
const totalHeight = computed(() => logHeight * logCount.value);
const startIndex = ref(0);
const containerHeight = ref(500);
const visibleCount = computed(() => Math.ceil(containerHeight.value / logHeight));
const visibleLogs = computed(() => {
    return logs.value.slice(startIndex.value, startIndex.value + visibleCount.value);
});

const timeOptions = ref([
    { label: i18n.global.t('container.all'), value: 'all' },
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
    }
};

const handleClose = async () => {
    stopListening();
    logVisible.value = false;
};

const searchLogs = async () => {
    if (Number(logSearch.tail) < 0) {
        MsgError(i18n.global.t('container.linesHelper'));
        return;
    }
    logs.value = [];
    let url = `/api/v2/containers/search/log?container=${logSearch.container}&since=${logSearch.mode}&tail=${logSearch.tail}&follow=${logSearch.isWatch}`;
    if (logSearch.compose !== '') {
        url = `/api/v2/containers/search/log?compose=${logSearch.compose}&since=${logSearch.mode}&tail=${logSearch.tail}&follow=${logSearch.isWatch}`;
    }
    eventSource = new EventSource(url);
    eventSource.onmessage = (event: MessageEvent) => {
        const data = event.data;
        logs.value.push(data);
        nextTick(() => {
            if (logContainer.value) {
                logContainer.value.scrollTop = logContainer.value.scrollHeight;
            }
        });
    };
    eventSource.onerror = (event: MessageEvent) => {
        stopListening();
        MsgError(event.data);
    };
};

const onDownload = async () => {
    logSearch.tail = 0;
    let msg = i18n.global.t('container.downLogHelper1', [logSearch.container]);
    ElMessageBox.confirm(msg, i18n.global.t('file.download'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        let params = {
            container: logSearch.container,
            since: logSearch.mode,
            tail: logSearch.tail,
            containerType: 'container',
        };
        let addItem = {};
        addItem['name'] = logSearch.container + '-' + dateFormatForName(new Date()) + '.log';
        DownloadFile(params).then((res) => {
            const downloadUrl = window.URL.createObjectURL(new Blob([res]));
            const a = document.createElement('a');
            a.style.display = 'none';
            a.href = downloadUrl;
            a.download = addItem['name'];
            const event = new MouseEvent('click');
            a.dispatchEvent(event);
        });
    });
};

const onClean = async () => {
    ElMessageBox.confirm(i18n.global.t('container.cleanLogHelper'), i18n.global.t('container.cleanLog'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        await cleanContainerLog(logSearch.container);
        searchLogs();
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    });
};

onUnmounted(() => {
    handleClose();
});

onMounted(() => {
    logSearch.container = props.container;
    logSearch.compose = props.compose;

    logVisible.value = true;
    logSearch.tail = 100;
    logSearch.mode = 'all';
    logSearch.isWatch = true;

    nextTick(() => {
        if (logContainer.value) {
            logContainer.value.scrollTop = totalHeight.value;
            containerHeight.value = logContainer.value.getBoundingClientRect().height;
        }
    });

    searchLogs();
});
</script>

<style scoped lang="scss">
.margin-button {
    margin-left: 20px;
}
.fullScreen {
    border: none;
}
.tailClass {
    width: 20%;
    float: left;
    margin-left: 20px;
}
.fetchClass {
    width: 30%;
    float: left;
}

.log-container {
    height: calc(100vh - 405px);
    overflow-y: auto;
    overflow-x: auto;
    position: relative;
    background-color: #1e1e1e;
    margin-top: 10px;
}

.log-item {
    position: absolute;
    width: 100%;
    padding: 2px;
    color: #f5f5f5;
    box-sizing: border-box;
    white-space: nowrap;
}
</style>
