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
            <highlightjs
                ref="editorRef"
                class="editor-main"
                language="JavaScript"
                :autodetect="false"
                :code="content"
            ></highlightjs>
        </div>
    </div>
</template>
<script lang="ts" setup>
import { nextTick, onMounted, onUnmounted, reactive, ref } from 'vue';
import { downloadFile } from '@/utils/util';
import { ReadByLine } from '@/api/modules/files';
import { watch } from 'vue';

const editorRef = ref();
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
        default: 'height: calc(100vh - 200px); width: 100%; min-height: 400px; overflow: auto;',
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
const content = ref('');
const end = ref(false);
const scrollerElement = ref<HTMLElement | null>(null);
const minPage = ref(1);
const maxPage = ref(1);
const logs = ref([]);
const isLoading = ref(false);

const readReq = reactive({
    id: 0,
    type: '',
    name: '',
    page: 1,
    pageSize: 500,
    latest: false,
});
const emit = defineEmits(['update:loading', 'update:hasContent', 'update:isReading']);

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

const lastLogs = ref([]);

const getContent = (pre: boolean) => {
    if (isLoading.value) {
        return;
    }
    emit('update:isReading', true);
    readReq.id = props.config.id;
    readReq.type = props.config.type;
    readReq.name = props.config.name;
    if (readReq.page < 1) {
        readReq.page = 1;
    }
    isLoading.value = true;
    ReadByLine(readReq).then((res) => {
        if (!end.value && res.data.end) {
            lastLogs.value = [...logs.value];
        }

        data.value = res.data;
        if (res.data.lines && res.data.lines.length > 0) {
            res.data.lines = res.data.lines.map((line) =>
                line.replace(/\\u(\w{4})/g, function (match, grp) {
                    return String.fromCharCode(parseInt(grp, 16));
                }),
            );
            const newLogs = res.data.lines;
            if (newLogs.length === readReq.pageSize && readReq.page < res.data.total) {
                readReq.page++;
            }

            if (
                readReq.type == 'php' &&
                logs.value.length > 0 &&
                newLogs.length > 0 &&
                newLogs[newLogs.length - 1] === logs.value[logs.value.length - 1]
            ) {
                isLoading.value = false;
                return;
            }

            if (stopSignals.some((signal) => newLogs[newLogs.length - 1].endsWith(signal))) {
                onCloseLog();
            }
            if (end.value) {
                if ((logs.value.length = 0)) {
                    logs.value = newLogs;
                } else {
                    logs.value = pre ? [...newLogs, ...lastLogs.value] : [...lastLogs.value, ...newLogs];
                }
            } else {
                if ((logs.value.length = 0)) {
                    logs.value = newLogs;
                } else {
                    logs.value = pre ? [...newLogs, ...logs.value] : [...logs.value, ...newLogs];
                }
            }
        }

        end.value = res.data.end;
        content.value = logs.value.join('\n');

        emit('update:hasContent', content.value !== '');
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
        if (logs.value && logs.value.length > 3000) {
            logs.value.splice(0, readReq.pageSize);
            if (minPage.value > 1) {
                minPage.value--;
            }
        }

        isLoading.value = false;
    });
};

function throttle<T extends (...args: any[]) => any>(func: T, limit: number): (...args: Parameters<T>) => void {
    let inThrottle: boolean;
    let lastFunc: ReturnType<typeof setTimeout>;
    let lastRan: number;
    return function (this: any, ...args: Parameters<T>) {
        if (!inThrottle) {
            func.apply(this, args);
            lastRan = Date.now();
            inThrottle = true;
            setTimeout(() => (inThrottle = false), limit);
        } else {
            clearTimeout(lastFunc);
            lastFunc = setTimeout(() => {
                if (Date.now() - lastRan >= limit) {
                    func.apply(this, args);
                    lastRan = Date.now();
                }
            }, limit - (Date.now() - lastRan));
        }
    };
}

const throttledGetContent = throttle(getContent, 3000);

const search = () => {
    throttledGetContent(false);
};

const changeTail = (fromOutSide: boolean) => {
    if (fromOutSide) {
        tailLog.value = !tailLog.value;
    }
    if (tailLog.value) {
        timer = setInterval(() => {
            search();
        }, 1000 * 3);
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
    emit('update:isReading', false);
    tailLog.value = false;
    clearInterval(Number(timer));
    timer = null;
    isLoading.value = false;
};

function isScrolledToBottom(element: HTMLElement): boolean {
    return element.scrollTop + element.clientHeight + 1 >= element.scrollHeight;
}

function isScrolledToTop(element: HTMLElement): boolean {
    return element.scrollTop === 0;
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
    readReq.latest = true;
    search();

    nextTick(() => {});
};

const clearLog = (): void => {
    content.value = '';
};

const initCodemirror = () => {
    nextTick(() => {
        if (editorRef.value) {
            scrollerElement.value = editorRef.value.$el as HTMLElement;
            scrollerElement.value.addEventListener('scroll', function () {
                if (tailLog.value) {
                    return;
                }
                if (isScrolledToBottom(scrollerElement.value)) {
                    if (maxPage.value > 1) {
                        readReq.page = maxPage.value;
                    }
                    search();
                }
                if (isScrolledToTop(scrollerElement.value)) {
                    readReq.page = minPage.value - 1;
                    if (readReq.page < 1) {
                        return;
                    }
                    minPage.value = readReq.page;
                    throttledGetContent(true);
                }
            });
            let hljsDom = scrollerElement.value.querySelector('.hljs') as HTMLElement;
            hljsDom.style['min-height'] = '500px';
        }
    });
};

onUnmounted(() => {
    onCloseLog();
});

onMounted(() => {
    initCodemirror();
    init();
});

defineExpose({ changeTail, onDownload, clearLog });
</script>
<style lang="scss" scoped>
.editor-main {
    height: calc(100vh - 480px);
    width: 100%;
    min-height: 600px;
    overflow: auto;
}
</style>
