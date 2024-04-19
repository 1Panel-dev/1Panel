<template>
    <div>
        <div style="display: flex; flex-wrap: wrap">
            <el-select @change="searchLogs" v-model="logSearch.mode" class="selectWidth">
                <template #prefix>{{ $t('container.fetch') }}</template>
                <el-option v-for="item in timeOptions" :key="item.label" :value="item.value" :label="item.label" />
            </el-select>
            <el-select @change="searchLogs" class="margin-button selectWidth" v-model.number="logSearch.tail">
                <template #prefix>{{ $t('container.lines') }}</template>
                <el-option :value="0" :label="$t('commons.table.all')" />
                <el-option :value="100" :label="100" />
                <el-option :value="200" :label="200" />
                <el-option :value="500" :label="500" />
                <el-option :value="1000" :label="1000" />
            </el-select>
            <div class="margin-button">
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
            :style="{ height: `calc(100vh - ${loadHeight()})`, 'margin-top': '10px' }"
            :lineWrapping="true"
            :matchBrackets="true"
            theme="cobalt"
            :styleActiveLine="true"
            :extensions="extensions"
            @ready="handleReady"
            v-model="logInfo"
            :disabled="true"
        />
    </div>
</template>

<script lang="ts" setup>
import { cleanContainerLog } from '@/api/modules/container';
import i18n from '@/lang';
import { dateFormatForName, downloadWithContent } from '@/utils/util';
import { onBeforeUnmount, reactive, ref, shallowRef } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { MsgError, MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const extensions = [javascript(), oneDark];

const logInfo = ref();
const view = shallowRef();
const handleReady = (payload) => {
    view.value = payload.view;
};
const terminalSocket = ref<WebSocket>();

const loadHeight = () => {
    return globalStore.openMenuTabs ? '405px' : '375px';
};

const logSearch = reactive({
    isWatch: false,
    container: '',
    containerID: '',
    mode: 'all',
    tail: 100,
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
        `${protocol}://${host}/api/v1/containers/search/log?container=${logSearch.containerID}&since=${logSearch.mode}&tail=${logSearch.tail}&follow=${logSearch.isWatch}`,
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

interface DialogProps {
    container: string;
    containerID: string;
}

const acceptParams = (props: DialogProps): void => {
    logSearch.containerID = props.containerID;
    logSearch.tail = 100;
    logSearch.mode = 'all';
    logSearch.isWatch = false;
    logSearch.container = props.container;
    searchLogs();
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

onBeforeUnmount(() => {
    terminalSocket.value?.send('close conn');
});

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.margin-button {
    margin-left: 20px;
}
.selectWidth {
    width: 150px;
}
</style>
