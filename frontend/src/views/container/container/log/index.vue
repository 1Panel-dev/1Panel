<template>
    <div>
        <el-drawer
            v-model="logVisible"
            :destroy-on-close="true"
            :close-on-click-modal="false"
            :before-close="handleClose"
            :size="globalStore.isFullScreen ? '100%' : '50%'"
        >
            <template #header>
                <DrawerHeader :header="$t('commons.button.log')" :resource="logSearch.container" :back="handleClose">
                    <template #extra v-if="!mobile">
                        <el-tooltip :content="loadTooltip()" placement="top">
                            <el-button @click="toggleFullscreen" class="fullScreen" icon="FullScreen" plain></el-button>
                        </el-tooltip>
                    </template>
                </DrawerHeader>
            </template>
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
                <div class="margin-button" style="float: left">
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

            <codemirror
                :autofocus="true"
                :placeholder="$t('commons.msg.noneData')"
                :indent-with-tab="true"
                :tabSize="4"
                style="margin-top: 20px; height: calc(100vh - 230px)"
                :lineWrapping="true"
                :matchBrackets="true"
                theme="cobalt"
                :styleActiveLine="true"
                :extensions="extensions"
                v-model="logInfo"
                @ready="handleReady"
                :disabled="true"
            />
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </el-drawer>
    </div>
</template>

<script lang="ts" setup>
import { cleanContainerLog } from '@/api/modules/container';
import i18n from '@/lang';
import { dateFormatForName, downloadWithContent } from '@/utils/util';
import { computed, onBeforeUnmount, reactive, ref, shallowRef, watch } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { ElMessageBox } from 'element-plus';
import { MsgError, MsgSuccess } from '@/utils/message';
import screenfull from 'screenfull';
import { GlobalStore } from '@/store';

const extensions = [javascript(), oneDark];
const logVisible = ref(false);
const mobile = computed(() => {
    return globalStore.isMobile();
});

const logInfo = ref<string>('');
const view = shallowRef();
const handleReady = (payload) => {
    view.value = payload.view;
};
const globalStore = GlobalStore();
const terminalSocket = ref<WebSocket>();

const logSearch = reactive({
    isWatch: true,
    container: '',
    containerID: '',
    mode: 'all',
    tail: 100,
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

function toggleFullscreen() {
    if (screenfull.isEnabled) {
        screenfull.toggle();
    }
}

const loadTooltip = () => {
    return i18n.global.t('commons.button.' + (screenfull.isFullscreen ? 'quitFullscreen' : 'fullscreen'));
};
const handleClose = async () => {
    terminalSocket.value?.send('close conn');
    logVisible.value = false;
};
watch(logVisible, (val) => {
    if (screenfull.isEnabled && !val && !mobile.value) screenfull.exit();
});
const searchLogs = async () => {
    if (Number(logSearch.tail) < 0) {
        MsgError(i18n.global.t('container.linesHelper'));
        return;
    }
    terminalSocket.value?.close();
    logInfo.value = '';
    const href = window.location.href;
    const protocol = href.split('//')[0] === 'http:' ? 'ws' : 'wss';
    const host = href.split('//')[1].split('/')[0];
    terminalSocket.value = new WebSocket(
        `${protocol}://${host}/api/v1/containers/search/log?container=${logSearch.containerID}&since=${logSearch.mode}&tail=${logSearch.tail}&follow=${logSearch.isWatch}`,
    );
    terminalSocket.value.onmessage = (event) => {
        logInfo.value += event.data.replace(/\x1B\[[0-9;]*[mG]/g, '');
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
            ? i18n.global.t('container.downLogHelper1', [logSearch.container])
            : i18n.global.t('container.downLogHelper2', [logSearch.container, logSearch.tail]);
    ElMessageBox.confirm(msg, i18n.global.t('file.download'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        downloadWithContent(logInfo.value, logSearch.container + '-' + dateFormatForName(new Date()) + '.log');
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

interface DialogProps {
    container: string;
    containerID: string;
}

const acceptParams = (props: DialogProps): void => {
    logVisible.value = true;
    logSearch.containerID = props.containerID;
    logSearch.tail = 100;
    logSearch.mode = 'all';
    logSearch.isWatch = true;
    logSearch.container = props.container;
    searchLogs();

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
</style>
