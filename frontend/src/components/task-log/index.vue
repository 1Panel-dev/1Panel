<template>
    <el-dialog
        v-model="open"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        :show-close="showClose"
        :before-close="handleClose"
        :width="width"
    >
        <div>
            <highlightjs class="editor-main" ref="editorRef" :autodetect="false" :code="content"></highlightjs>
        </div>
    </el-dialog>
</template>
<script lang="ts" setup>
import { nextTick, onUnmounted, reactive, ref } from 'vue';
import { ReadByLine } from '@/api/modules/files';

const editorRef = ref();

defineProps({
    showClose: {
        type: Boolean,
        default: true,
    },
    width: {
        type: String,
        default: '30%',
    },
});

const data = ref({
    enable: false,
    content: '',
    path: '',
});

let timer: NodeJS.Timer | null = null;
const tailLog = ref(false);
const content = ref('');
const end = ref(false);
const lastContent = ref('');
const scrollerElement = ref<HTMLElement | null>(null);
const minPage = ref(1);
const maxPage = ref(1);
const open = ref(false);
const em = defineEmits(['close']);

const readReq = reactive({
    taskID: '',
    type: 'task',
    page: 1,
    pageSize: 500,
    latest: false,
    taskType: '',
    taskOperate: '',
    id: 0,
});

const stopSignals = ['[TASK-END]'];

const initData = () => {
    open.value = true;
    initCodemirror();
    init();
};

const openWithTaskID = (id: string) => {
    readReq.taskID = id;
    initData();
};

const openWithResourceID = (taskType: string, taskOperate: string, resourceID: number) => {
    readReq.taskType = taskType;
    readReq.id = resourceID;
    readReq.taskOperate = taskOperate;
    initData();
};

const getContent = (pre: boolean) => {
    if (readReq.page < 1) {
        readReq.page = 1;
    }
    ReadByLine(readReq).then((res) => {
        if (!end.value && res.data.end) {
            lastContent.value = content.value;
        }

        res.data.content = res.data.content.replace(/\\u(\w{4})/g, function (match, grp) {
            return String.fromCharCode(parseInt(grp, 16));
        });
        data.value = res.data;
        if (res.data.content != '') {
            if (stopSignals.some((signal) => res.data.content.includes(signal))) {
                onCloseLog();
            }
            if (end.value) {
                if (lastContent.value == '') {
                    content.value = res.data.content;
                } else {
                    content.value = pre
                        ? res.data.content + '\n' + lastContent.value
                        : lastContent.value + '\n' + res.data.content;
                }
            } else {
                if (content.value == '') {
                    content.value = res.data.content;
                } else {
                    content.value = pre
                        ? res.data.content + '\n' + content.value
                        : content.value + '\n' + res.data.content;
                }
            }
        }
        end.value = res.data.end;
        nextTick(() => {
            if (pre) {
                if (scrollerElement.value.scrollHeight > 2000) {
                    scrollerElement.value.scrollTop = 2000;
                }
            } else {
                scrollerElement.value.scrollTop = scrollerElement.value.scrollHeight;
            }
        });

        if (readReq.latest) {
            readReq.page = res.data.total;
            readReq.latest = false;
            maxPage.value = res.data.total;
            minPage.value = res.data.total;
        }
    });
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

const handleClose = () => {
    onCloseLog();
    open.value = false;
    em('close', open.value);
};

const onCloseLog = async () => {
    tailLog.value = false;
    clearInterval(Number(timer));
    timer = null;
};

function isScrolledToBottom(element: HTMLElement): boolean {
    return element.scrollTop + element.clientHeight + 1 >= element.scrollHeight;
}

function isScrolledToTop(element: HTMLElement): boolean {
    return element.scrollTop === 0;
}

const init = () => {
    tailLog.value = true;
    if (tailLog.value) {
        changeTail(false);
    }
    readReq.latest = true;
    getContent(false);
};

const initCodemirror = () => {
    nextTick(() => {
        if (editorRef.value) {
            scrollerElement.value = editorRef.value.$el as HTMLElement;
            scrollerElement.value.addEventListener('scroll', function () {
                if (isScrolledToBottom(scrollerElement.value)) {
                    readReq.page = maxPage.value;
                    getContent(false);
                }
                if (isScrolledToTop(scrollerElement.value)) {
                    readReq.page = minPage.value - 1;
                    if (readReq.page < 1) {
                        return;
                    }
                    minPage.value = readReq.page;
                    getContent(true);
                }
            });
            let hljsDom = scrollerElement.value.querySelector('.hljs') as HTMLElement;
            hljsDom.style['min-height'] = '400px';
        }
    });
};

onUnmounted(() => {
    onCloseLog();
});

defineExpose({ openWithResourceID, openWithTaskID });
</script>
<style lang="scss" scoped>
.task-log-dialog {
    --dialog-max-height: 80vh;
    --dialog-header-height: 50px;
    --dialog-padding: 20px;
    .el-dialog {
        max-width: 60%;
        max-height: var(--dialog-max-height);
        margin-top: 5vh !important;
        display: flex;
        flex-direction: column;
    }
    .el-dialog__body {
        flex: 1;
        overflow: hidden;
        padding: var(--dialog-padding);
    }
    .log-container {
        height: calc(var(--dialog-max-height) - var(--dialog-header-height) - var(--dialog-padding) * 2);
        overflow: hidden;
    }
}

.editor-main {
    width: 100%;
    overflow: auto;
    height: 420px;
}
</style>
