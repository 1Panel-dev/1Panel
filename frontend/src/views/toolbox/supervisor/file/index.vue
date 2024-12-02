<template>
    <el-drawer
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        v-model="open"
        :size="globalStore.isFullScreen ? '100%' : '50%'"
        :before-close="handleClose"
    >
        <template #header>
            <DrawerHeader :header="title" :back="handleClose">
                <template #extra v-if="!mobile">
                    <el-tooltip :content="loadTooltip()" placement="top">
                        <el-button @click="toggleFullscreen" class="fullScreen" icon="FullScreen" plain></el-button>
                    </el-tooltip>
                </template>
            </DrawerHeader>
        </template>
        <div v-if="req.file != 'config'">
            <el-tabs v-model="req.file" type="card" @tab-click="handleChange">
                <el-tab-pane :label="$t('logs.runLog')" name="out.log"></el-tab-pane>
                <el-tab-pane :label="$t('logs.errLog')" name="err.log"></el-tab-pane>
            </el-tabs>
            <el-checkbox border v-model="tailLog" style="float: left" @change="changeTail">
                {{ $t('commons.button.watch') }}
            </el-checkbox>
            <el-button style="margin-left: 20px" @click="cleanLog" icon="Delete">
                {{ $t('commons.button.clean') }}
            </el-button>
        </div>
        <br />
        <div v-loading="loading">
            <codemirror
                style="height: calc(100vh - 430px); min-height: 300px"
                :autofocus="true"
                :placeholder="$t('website.noLog')"
                :indent-with-tab="true"
                :tabSize="4"
                :lineWrapping="true"
                :matchBrackets="true"
                theme="cobalt"
                :styleActiveLine="true"
                :extensions="extensions"
                v-model="content"
                @ready="handleReady"
            />
        </div>

        <template #footer>
            <span>
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" :disabled="loading" @click="submit()" v-if="req.file === 'config'">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </el-drawer>

    <OpDialog ref="opRef" @search="getContent" />
</template>
<script lang="ts" setup>
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { computed, onUnmounted, reactive, ref, shallowRef, watch } from 'vue';
import { OperateSupervisorProcessFile } from '@/api/modules/host-tool';
import i18n from '@/lang';
import { TabsPaneContext } from 'element-plus';
import { MsgSuccess } from '@/utils/message';
import screenfull from 'screenfull';
import { GlobalStore } from '@/store';

const extensions = [javascript(), oneDark];
const loading = ref(false);
const content = ref('');
const tailLog = ref(false);
const open = ref(false);
const req = reactive({
    name: '',
    file: 'conf',
    operate: '',
    content: '',
});
const title = ref('');
const opRef = ref();
const globalStore = GlobalStore();

const view = shallowRef();
const handleReady = (payload) => {
    view.value = payload.view;
};
let timer: NodeJS.Timer | null = null;

const em = defineEmits(['search']);

watch(open, (val) => {
    if (screenfull.isEnabled && !val && !mobile.value) screenfull.exit();
});

const mobile = computed(() => {
    return globalStore.isMobile();
});

function toggleFullscreen() {
    globalStore.isFullScreen = !globalStore.isFullScreen;
}
const loadTooltip = () => {
    return i18n.global.t('commons.button.' + (globalStore.isFullScreen ? 'quitFullscreen' : 'fullscreen'));
};

const getContent = () => {
    loading.value = true;
    OperateSupervisorProcessFile(req)
        .then((res) => {
            content.value = res.data;
        })
        .finally(() => {
            loading.value = false;
        });
};

const handleChange = (tab: TabsPaneContext) => {
    req.file = tab.props.name.toString();
    getContent();
};

const changeTail = () => {
    if (tailLog.value) {
        timer = setInterval(() => {
            getContent();
        }, 1000 * 5);
    } else {
        onCloseLog();
    }
};

const handleClose = () => {
    content.value = '';
    open.value = false;
    globalStore.isFullScreen = false;
};

const submit = () => {
    const updateReq = {
        name: req.name,
        operate: 'update',
        file: req.file,
        content: content.value,
    };
    loading.value = true;
    OperateSupervisorProcessFile(updateReq)
        .then(() => {
            em('search');
            open.value = false;
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
        })
        .finally(() => {
            loading.value = false;
        });
};

const acceptParams = (name: string, file: string, operate: string) => {
    req.name = name;
    req.file = file;
    req.operate = operate;

    title.value = file == 'config' ? i18n.global.t('website.source') : i18n.global.t('commons.button.log');
    getContent();
    open.value = true;
    if (!mobile.value) {
        screenfull.on('change', () => {
            globalStore.isFullScreen = screenfull.isFullscreen;
        });
    }
};

const cleanLog = async () => {
    let log = req.file === 'out.log' ? i18n.global.t('logs.runLog') : i18n.global.t('logs.errLog');
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.clean'),
        names: [req.name],
        msg: i18n.global.t('commons.msg.operatorHelper', [log, i18n.global.t('commons.button.clean')]),
        api: OperateSupervisorProcessFile,
        params: { name: req.name, operate: 'clear', file: req.file },
    });
};

const onCloseLog = async () => {
    tailLog.value = false;
    clearInterval(Number(timer));
    timer = null;
};

onUnmounted(() => {
    onCloseLog();
});

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.fullScreen {
    border: none;
}
</style>
