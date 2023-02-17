<template>
    <el-drawer v-model="logVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="50%">
        <template #header>
            <DrawerHeader :header="$t('commons.button.log')" :back="handleClose" />
        </template>
        <div>
            <el-select @change="searchLogs" style="width: 30%; float: left" v-model="logSearch.mode">
                <el-option v-for="item in timeOptions" :key="item.label" :value="item.value" :label="item.label" />
            </el-select>
            <div style="margin-left: 20px; float: left">
                <el-checkbox border v-model="logSearch.isWatch">{{ $t('commons.button.watch') }}</el-checkbox>
            </div>
            <el-button style="margin-left: 20px" @click="onDownload" icon="Download">
                {{ $t('file.download') }}
            </el-button>
        </div>

        <codemirror
            :autofocus="true"
            placeholder="None data"
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
            :readOnly="true"
        />
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="logVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
            </span>
        </template>
    </el-drawer>
</template>

<script lang="ts" setup>
import { logContainer } from '@/api/modules/container';
import i18n from '@/lang';
import { dateFormatForName } from '@/utils/util';
import { nextTick, reactive, ref, shallowRef } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import DrawerHeader from '@/components/drawer-header/index.vue';

const extensions = [javascript(), oneDark];

const logVisiable = ref(false);

const logInfo = ref();
const view = shallowRef();
const handleReady = (payload) => {
    view.value = payload.view;
};

const logSearch = reactive({
    isWatch: false,
    container: '',
    containerID: '',
    mode: 'all',
});
let timer: NodeJS.Timer | null = null;

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

const handleClose = async () => {
    logVisiable.value = false;
    clearInterval(Number(timer));
    timer = null;
};

const searchLogs = async () => {
    const res = await logContainer(logSearch);
    logInfo.value = res.data;
    nextTick(() => {
        const state = view.value.state;
        view.value.dispatch({
            selection: { anchor: state.doc.length, head: state.doc.length },
            scrollIntoView: true,
        });
    });
};

const onDownload = async () => {
    const downloadUrl = window.URL.createObjectURL(new Blob([logInfo.value]));
    const a = document.createElement('a');
    a.style.display = 'none';
    a.href = downloadUrl;
    a.download = logSearch.container + '-' + dateFormatForName(new Date()) + '.log';
    const event = new MouseEvent('click');
    a.dispatchEvent(event);
};

interface DialogProps {
    container: string;
    containerID: string;
}

const acceptParams = (props: DialogProps): void => {
    logVisiable.value = true;
    logSearch.containerID = props.containerID;
    logSearch.mode = 'all';
    logSearch.isWatch = false;
    logSearch.container = props.container;
    searchLogs();
    timer = setInterval(() => {
        if (logSearch.isWatch) {
            searchLogs();
        }
    }, 1000 * 5);
};

defineExpose({
    acceptParams,
});
</script>
