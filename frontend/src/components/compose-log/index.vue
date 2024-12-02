<template>
    <DrawerPro
        v-model="open"
        :header="$t('commons.button.log')"
        :back="handleClose"
        :size="globalStore.isFullScreen ? 'full' : 'large'"
        :resource="resource"
    >
        <template #extra v-if="!mobile">
            <el-tooltip :content="loadTooltip()" placement="top">
                <el-button @click="toggleFullscreen" class="fullScreen" icon="FullScreen" plain></el-button>
            </el-tooltip>
        </template>
        <template #content>
            <div class="flex flex-wrap">
                <el-select @change="searchLogs" v-model="logSearch.mode" class="selectWidth">
                    <template #prefix>{{ $t('container.fetch') }}</template>
                    <el-option v-for="item in timeOptions" :key="item.label" :value="item.value" :label="item.label" />
                </el-select>
                <el-select @change="searchLogs" class="ml-5 selectWidth" v-model.number="logSearch.tail">
                    <template #prefix>{{ $t('container.lines') }}</template>
                    <el-option :value="0" :label="$t('commons.table.all')" />
                    <el-option :value="100" :label="100" />
                    <el-option :value="200" :label="200" />
                    <el-option :value="500" :label="500" />
                    <el-option :value="1000" :label="1000" />
                </el-select>
                <div class="ml-5">
                    <el-checkbox border @change="searchLogs" v-model="logSearch.isWatch">
                        {{ $t('commons.button.watch') }}
                    </el-checkbox>
                </div>
                <el-button class="ml-5" @click="onDownload" icon="Download">
                    {{ $t('file.download') }}
                </el-button>
            </div>
            <div class="mt-2.5">
                <highlightjs
                    ref="editorRef"
                    class="editor-main"
                    language="JavaScript"
                    :autodetect="false"
                    :code="logInfo"
                ></highlightjs>
            </div>
        </template>
    </DrawerPro>
</template>

<script lang="ts" setup>
import i18n from '@/lang';
import { dateFormatForName } from '@/utils/util';
import { computed, onBeforeUnmount, reactive, ref, watch } from 'vue';
import { MsgError } from '@/utils/message';
import { GlobalStore } from '@/store';
import screenfull from 'screenfull';
import { DownloadFile } from '@/api/modules/container';

const logInfo = ref('');
const logSocket = ref<WebSocket>();
const open = ref(false);
const resource = ref('');
const globalStore = GlobalStore();
const logVisible = ref(false);
const editorRef = ref();
const scrollerElement = ref<HTMLElement | null>(null);

const mobile = computed(() => {
    return globalStore.isMobile();
});

const logSearch = reactive({
    isWatch: true,
    compose: '',
    mode: 'all',
    tail: 500,
});

const handleClose = () => {
    logSocket.value?.send('close conn');
    open.value = false;
};

function toggleFullscreen() {
    if (screenfull.isEnabled) {
        screenfull.toggle();
    }
}
const loadTooltip = () => {
    return i18n.global.t('commons.button.' + (screenfull.isFullscreen ? 'quitFullscreen' : 'fullscreen'));
};

watch(logVisible, (val) => {
    if (screenfull.isEnabled && !val && !mobile.value) screenfull.exit();
});

const timeOptions = ref([
    { label: i18n.global.t('container.all'), value: 'all' },
    {
        label: i18n.global.t('container.lastDay'),
        value: new Date(new Date().getTime() - 3600 * 1000 * 24 * 1).getTime() / 1000 + '',
    },
    {
        label: i18n.global.t('container.last4Hour'),
        value: new Date(new Date().getTime() - 3600 * 1000 * 4).getTime() / 1000 + '',
    },
    {
        label: i18n.global.t('container.lastHour'),
        value: new Date(new Date().getTime() - 3600 * 1000).getTime() / 1000 + '',
    },
    {
        label: i18n.global.t('container.last10Min'),
        value: new Date(new Date().getTime() - 600 * 1000).getTime() / 1000 + '',
    },
]);

const searchLogs = async () => {
    if (Number(logSearch.tail) < 0) {
        MsgError(i18n.global.t('container.linesHelper'));
        return;
    }
    logSocket.value?.send('close conn');
    logSocket.value?.close();
    logInfo.value = '';
    const href = window.location.href;
    const protocol = href.split('//')[0] === 'http:' ? 'ws' : 'wss';
    const host = href.split('//')[1].split('/')[0];
    logSocket.value = new WebSocket(
        `${protocol}://${host}/api/v2/containers/compose/search/log?compose=${logSearch.compose}&since=${logSearch.mode}&tail=${logSearch.tail}&follow=${logSearch.isWatch}`,
    );
    logSocket.value.onmessage = (event) => {
        logInfo.value += event.data;
        nextTick(() => {
            scrollerElement.value.scrollTop = scrollerElement.value.scrollHeight;
        });
    };
};

const onDownload = async () => {
    let msg =
        logSearch.tail === 0
            ? i18n.global.t('app.downloadLogHelper1', [resource.value])
            : i18n.global.t('app.downloadLogHelper2', [resource.value, logSearch.tail]);
    ElMessageBox.confirm(msg, i18n.global.t('file.download'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        let params = {
            container: logSearch.compose,
            since: logSearch.mode,
            tail: logSearch.tail,
            containerType: 'compose',
        };
        let addItem = {};
        addItem['name'] = logSearch.compose + '-' + dateFormatForName(new Date()) + '.log';
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

interface DialogProps {
    compose: string;
    resource: string;
}

const acceptParams = (props: DialogProps): void => {
    logSearch.compose = props.compose;
    logSearch.tail = 200;
    logSearch.mode = timeOptions.value[3].value;
    logSearch.isWatch = true;
    resource.value = props.resource;
    open.value = true;
    if (!mobile.value) {
        screenfull.on('change', () => {
            globalStore.isFullScreen = screenfull.isFullscreen;
        });
    }
    nextTick(() => {
        if (editorRef.value) {
            searchLogs();
            scrollerElement.value = editorRef.value.$el as HTMLElement;
            let hljsDom = scrollerElement.value.querySelector('.hljs') as HTMLElement;
            hljsDom.style['min-height'] = '300px';
            hljsDom.style['flex-grow'] = 1;
        }
    });
};

onBeforeUnmount(() => {
    handleClose();
});

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.selectWidth {
    width: 200px;
}

.editor-main {
    display: flex;
    flex-direction: column;
    height: 80vh;
}
</style>
