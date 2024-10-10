<template>
    <el-drawer
        v-model="open"
        :size="globalStore.isFullScreen ? '100%' : '50%'"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        :before-close="handleClose"
    >
        <template #header>
            <DrawerHeader :header="$t('commons.button.log')" :resource="resource" :back="handleClose">
                <template #extra v-if="!mobile">
                    <el-tooltip :content="loadTooltip()" placement="top">
                        <el-button @click="toggleFullscreen" class="fullScreen" icon="FullScreen" plain></el-button>
                    </el-tooltip>
                </template>
            </DrawerHeader>
        </template>
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

        <codemirror
            :autofocus="true"
            :placeholder="$t('commons.msg.noneData')"
            :indent-with-tab="true"
            :tabSize="4"
            style="margin-top: 10px; height: calc(100vh - 200px)"
            :lineWrapping="true"
            :matchBrackets="true"
            theme="cobalt"
            :styleActiveLine="true"
            :extensions="extensions"
            @ready="handleReady"
            v-model="logInfo"
            :disabled="true"
        />
    </el-drawer>
</template>

<script lang="ts" setup>
import i18n from '@/lang';
import { dateFormatForName } from '@/utils/util';
import { computed, onBeforeUnmount, reactive, ref, shallowRef, watch } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { MsgError } from '@/utils/message';
import { GlobalStore } from '@/store';
import screenfull from 'screenfull';
import { DownloadFile } from '@/api/modules/container';

const extensions = [javascript(), oneDark];

const logInfo = ref();
const view = shallowRef();
const handleReady = (payload) => {
    view.value = payload.view;
};
const terminalSocket = ref<WebSocket>();
const open = ref(false);
const resource = ref('');
const globalStore = GlobalStore();
const logVisible = ref(false);

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
    terminalSocket.value?.send('close conn');
    open.value = false;
    globalStore.isFullScreen = false;
};

function toggleFullscreen() {
    globalStore.isFullScreen = !globalStore.isFullScreen;
}
const loadTooltip = () => {
    return i18n.global.t('commons.button.' + (globalStore.isFullScreen ? 'quitFullscreen' : 'fullscreen'));
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
    terminalSocket.value?.send('close conn');
    terminalSocket.value?.close();
    logInfo.value = '';
    const href = window.location.href;
    const protocol = href.split('//')[0] === 'http:' ? 'ws' : 'wss';
    const host = href.split('//')[1].split('/')[0];
    terminalSocket.value = new WebSocket(
        `${protocol}://${host}/api/v1/containers/compose/search/log?compose=${logSearch.compose}&since=${logSearch.mode}&tail=${logSearch.tail}&follow=${logSearch.isWatch}`,
    );
    terminalSocket.value.onmessage = (event) => {
        logInfo.value += event.data;
        const state = view.value.state;
        view.value.dispatch({
            selection: { anchor: state.doc.length, head: state.doc.length },
            scrollIntoView: true,
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
    searchLogs();
    open.value = true;
    if (!mobile.value) {
        screenfull.on('change', () => {
            globalStore.isFullScreen = screenfull.isFullscreen;
        });
    }
};

onBeforeUnmount(() => {
    handleClose();
});

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.fullScreen {
    border: none;
}
.selectWidth {
    width: 200px;
}
</style>
