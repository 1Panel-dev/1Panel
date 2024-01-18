<template>
    <div>
        <div v-if="defaultButton">
            <el-checkbox border v-model="tailLog" class="float-left" @change="changeTail(false)">
                {{ $t('commons.button.watch') }}
            </el-checkbox>
            <el-button class="ml-2.5" @click="onDownload" icon="Download" :disabled="data.content === ''">
                {{ $t('file.download') }}
            </el-button>
            <span v-if="$slots.button" class="ml-2.5">
                <slot name="button"></slot>
            </span>
        </div>
        <div class="mt-2.5">
            <Codemirror
                ref="logContainer"
                :style="styleObject"
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
        </div>
    </div>
</template>
<script lang="ts" setup>
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, shallowRef } from 'vue';
import { downloadFile } from '@/utils/util';
import { ReadByLine } from '@/api/modules/files';
import { watch } from 'vue';

const extensions = [javascript(), oneDark];

interface LogProps {
    id?: number;
    type: string;
    name?: string;
    tail?: boolean;
}

const props = defineProps({
    config: {
        type: Object as () => LogProps | null,
        default: () => ({
            id: 0,
            type: '',
            name: '',
            tail: false,
        }),
    },
    style: {
        type: String,
        default: 'height: calc(100vh - 200px); width: 100%; min-height: 400px',
    },
    defaultButton: {
        type: Boolean,
        default: true,
    },
    loading: {
        type: Boolean,
        default: true,
    },
    hasContent: {
        type: Boolean,
        default: false,
    },
});
const data = ref({
    enable: false,
    content: '',
    path: '',
});

let timer: NodeJS.Timer | null = null;
const tailLog = ref(false);
const view = shallowRef();
const content = ref('');
const end = ref(false);
const lastContent = ref('');
const logContainer = ref();
const scrollerElement = ref<HTMLElement | null>(null);

const readReq = reactive({
    id: 0,
    type: '',
    name: '',
    page: 0,
    pageSize: 2000,
});
const emit = defineEmits(['update:loading', 'update:hasContent']);

const handleReady = (payload) => {
    view.value = payload.view;
    const editorContainer = payload.container;
    const editorElement = editorContainer.querySelector('.cm-editor');
    scrollerElement.value = editorElement.querySelector('.cm-scroller') as HTMLElement;
};

const loading = ref(props.loading);

watch(
    () => props.loading,
    (newLoading) => {
        loading.value = newLoading;
    },
);

const changeLoading = () => {
    loading.value = !loading.value;
    emit('update:loading', loading.value);
};

const styleObject = computed(() => {
    const styles = {};
    let style = 'height: calc(100vh - 200px); width: 100%; min-height: 400px';
    if (props.style != null && props.style != '') {
        style = props.style;
    }
    style.split(';').forEach((styleRule) => {
        const [property, value] = styleRule.split(':');
        if (property && value) {
            const formattedProperty = property
                .trim()
                .replace(/([a-z])([A-Z])/g, '$1-$2')
                .toLowerCase();
            styles[formattedProperty] = value.trim();
        }
    });
    return styles;
});

const stopSignals = [
    'docker-compose up failed!',
    'docker-compose up successful!',
    'image build failed!',
    'image build successful!',
    'image pull failed!',
    'image pull successful!',
    'image push failed!',
    'image push successful!',
];

const getContent = () => {
    if (!end.value) {
        readReq.page += 1;
    }
    readReq.id = props.config.id;
    readReq.type = props.config.type;
    readReq.name = props.config.name;
    ReadByLine(readReq).then((res) => {
        if (!end.value && res.data.end) {
            lastContent.value = content.value;
        }
        data.value = res.data;
        if (res.data.content != '') {
            if (stopSignals.some((singal) => res.data.content.endsWith(singal))) {
                onCloseLog();
            }
            if (end.value) {
                if (lastContent.value == '') {
                    content.value = res.data.content;
                } else {
                    content.value = lastContent.value + '\n' + res.data.content;
                }
            } else {
                if (content.value == '') {
                    content.value = res.data.content;
                } else {
                    content.value = content.value + '\n' + res.data.content;
                }
            }
        }
        end.value = res.data.end;
        emit('update:hasContent', content.value !== '');
        nextTick(() => {
            const state = view.value.state;
            view.value.dispatch({
                selection: { anchor: state.doc.length, head: state.doc.length },
            });
            view.value.focus();
            const firstLine = view.value.state.doc.line(view.value.state.doc.lines);
            const { top } = view.value.lineBlockAt(firstLine.from);
            scrollerElement.value.scrollTo({ top, behavior: 'instant' });
        });
    });
};

const changeTail = (fromOutSide: boolean) => {
    if (fromOutSide) {
        tailLog.value = !tailLog.value;
    }
    if (tailLog.value) {
        timer = setInterval(() => {
            getContent();
        }, 1000 * 2);
    } else {
        onCloseLog();
    }
};

const onDownload = async () => {
    changeLoading();
    downloadFile(data.value.path);
    changeLoading();
};

const onCloseLog = async () => {
    tailLog.value = false;
    clearInterval(Number(timer));
    timer = null;
};

function isScrolledToBottom(element: HTMLElement): boolean {
    return element.scrollTop + element.clientHeight + 1 >= element.scrollHeight;
}

const init = () => {
    if (props.config.tail) {
        tailLog.value = props.config.tail;
    } else {
        tailLog.value = false;
    }
    if (tailLog.value) {
        changeTail(false);
    }
    getContent();

    nextTick(() => {
        if (scrollerElement.value) {
            scrollerElement.value.addEventListener('scroll', function () {
                if (isScrolledToBottom(scrollerElement.value)) {
                    getContent();
                }
            });
        }
    });
};

const clearLog = (): void => {
    content.value = '';
};

onUnmounted(() => {
    onCloseLog();
});

onMounted(() => {
    init();
});

defineExpose({ changeTail, onDownload, clearLog });
</script>
