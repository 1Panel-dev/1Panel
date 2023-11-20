<template>
    <el-drawer v-model="open" :destroy-on-close="true" :close-on-click-modal="false" size="40%">
        <template #header>
            <DrawerHeader :header="$t('website.log')" :back="handleClose" />
        </template>
        <div>
            <div class="mt-2.5">
                <el-checkbox border v-model="tailLog" class="float-left" @change="changeTail">
                    {{ $t('commons.button.watch') }}
                </el-checkbox>
                <el-button class="ml-5" @click="onDownload" icon="Download" :disabled="data.content === ''">
                    {{ $t('file.download') }}
                </el-button>
            </div>
        </div>
        <br />
        <Codemirror
            ref="logContainer"
            style="height: calc(100vh - 200px); width: 100%; min-height: 400px"
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
            :disabled="true"
            @ready="handleReady"
        />
    </el-drawer>
</template>
<script lang="ts" setup>
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { nextTick, onUnmounted, reactive, ref, shallowRef } from 'vue';
import { downloadFile } from '@/utils/util';
import { ReadByLine } from '@/api/modules/files';

const extensions = [javascript(), oneDark];

interface LogProps {
    path: string;
}

const data = ref({
    enable: false,
    content: '',
    path: '',
});
const tailLog = ref(false);
let timer: NodeJS.Timer | null = null;

const view = shallowRef();
const editorContainer = ref<HTMLDivElement | null>(null);
const handleReady = (payload) => {
    view.value = payload.view;
    editorContainer.value = payload.container;
};
const content = ref('');
const end = ref(false);
const lastContent = ref('');
const open = ref(false);
const logContainer = ref();

const readReq = reactive({
    path: '',
    page: 0,
    pageSize: 100,
});
const em = defineEmits(['close']);

const handleClose = (search: boolean) => {
    open.value = false;
    em('close', search);
};

const getContent = () => {
    if (!end.value) {
        readReq.page += 1;
    }
    ReadByLine(readReq).then((res) => {
        if (!end.value && res.data.end) {
            lastContent.value = content.value;
        }
        data.value = res.data;
        if (res.data.content != '') {
            if (end.value) {
                content.value = lastContent.value + '\n' + res.data.content;
            } else {
                if (content.value == '') {
                    content.value = res.data.content;
                } else {
                    content.value = content.value + '\n' + res.data.content;
                }
            }
        }
        end.value = res.data.end;
        nextTick(() => {
            const state = view.value.state;
            view.value.dispatch({
                selection: { anchor: state.doc.length, head: state.doc.length },
            });
            view.value.focus();
        });
    });
};

const changeTail = () => {
    if (tailLog.value) {
        timer = setInterval(() => {
            getContent();
        }, 1000 * 1);
    } else {
        onCloseLog();
    }
};

const onDownload = async () => {
    downloadFile(data.value.path);
};

const onCloseLog = async () => {
    tailLog.value = false;
    clearInterval(Number(timer));
    timer = null;
};

function isScrolledToBottom(element: HTMLElement): boolean {
    return element.scrollTop + element.clientHeight === element.scrollHeight;
}

const acceptParams = (props: LogProps) => {
    readReq.path = props.path;
    open.value = true;
    tailLog.value = false;

    getContent();
    nextTick(() => {
        let editorElement = editorContainer.value.querySelector('.cm-editor');
        let scrollerElement = editorElement.querySelector('.cm-scroller') as HTMLElement;
        if (scrollerElement) {
            scrollerElement.addEventListener('scroll', function () {
                if (isScrolledToBottom(scrollerElement)) {
                    getContent();
                }
            });
        }
    });
};

onUnmounted(() => {
    onCloseLog();
});

defineExpose({ acceptParams });
</script>
