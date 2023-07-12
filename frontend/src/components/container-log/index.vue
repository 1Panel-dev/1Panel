<template>
    <div>
        <div>
            <el-select @change="searchLogs" style="width: 10%; float: left" v-model="logSearch.mode">
                <template #prefix>{{ $t('container.fetch') }}</template>
                <el-option v-for="item in timeOptions" :key="item.label" :value="item.value" :label="item.label" />
            </el-select>
            <el-input
                @change="searchLogs"
                class="margin-button"
                style="width: 10%; float: left"
                v-model.number="logSearch.tail"
            >
                <template #prefix>
                    <div style="margin-left: 2px">{{ $t('container.lines') }}</div>
                </template>
            </el-input>
            <div class="margin-button" style="float: left">
                <el-checkbox border v-model="logSearch.isWatch">{{ $t('commons.button.watch') }}</el-checkbox>
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
            style="margin-top: 10px; height: calc(100vh - 375px)"
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

const extensions = [javascript(), oneDark];

const logInfo = ref();
const view = shallowRef();
const handleReady = (payload) => {
    view.value = payload.view;
};
const terminalSocket = ref<WebSocket>();

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
    if (!Number(logSearch.tail) || Number(logSearch.tail) < 0) {
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
        logInfo.value += event.data;
        const state = view.value.state;
        view.value.dispatch({
            selection: { anchor: state.doc.length, head: state.doc.length },
            scrollIntoView: true,
        });
    };
};

const onDownload = async () => {
    downloadWithContent(logInfo.value, logSearch.container + '-' + dateFormatForName(new Date()) + '.log');
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
    ElMessageBox.confirm(i18n.global.t('commons.msg.clean'), i18n.global.t('container.cleanLog'), {
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
    terminalSocket.value?.close();
});

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.margin-button {
    margin-left: 20px;
}
</style>
