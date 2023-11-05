<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('logs.websiteLog')">
            <template #toolbar>
                <el-row>
                    <el-col :span="16">
                        <el-button
                            class="tag-button"
                            :class="logReq.logType === 'access.log' ? '' : 'no-active'"
                            :type="logReq.logType === 'access.log' ? 'primary' : ''"
                            @click="changeType('access.log')"
                        >
                            {{ $t('logs.runLog') }}
                        </el-button>
                        <el-button
                            class="tag-button"
                            :class="logReq.logType === 'error.log' ? '' : 'no-active'"
                            :type="logReq.logType === 'error.log' ? 'primary' : ''"
                            @click="changeType('error.log')"
                        >
                            {{ $t('logs.errLog') }}
                        </el-button>
                    </el-col>
                </el-row>
            </template>
            <template #search>
                <div>
                    <el-select v-model="logReq.id" @change="search()">
                        <template #prefix>{{ $t('website.website') }}</template>
                        <el-option
                            v-for="(website, index) in websites"
                            :key="index"
                            :label="website.primaryDomain"
                            :value="website.id"
                        ></el-option>
                    </el-select>
                    <el-button class="left-button">
                        <el-checkbox v-model="tailLog" @change="changeTail">
                            {{ $t('commons.button.watch') }}
                        </el-checkbox>
                    </el-button>
                    <el-button class="left-button" @click="onDownload" icon="Download" :disabled="data.content === ''">
                        {{ $t('file.download') }}
                    </el-button>
                    <el-button
                        type="primary"
                        plain
                        @click="onClean()"
                        class="left-button"
                        :disabled="data.content.length === 0"
                    >
                        {{ $t('logs.deleteLogs') }}
                    </el-button>
                </div>
            </template>
            <template #main>
                <Codemirror
                    style="height: calc(100vh - 368px); width: 100%"
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
            </template>
        </LayoutContent>
        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmitClean"></ConfirmDialog>
    </div>
</template>
<script setup lang="ts">
import { ListWebsites, OpWebsiteLog } from '@/api/modules/website';
import { nextTick, reactive, shallowRef } from 'vue';
import { onMounted } from 'vue';
import { ref } from 'vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { dateFormatForName, downloadWithContent } from '@/utils/util';

const extensions = [javascript(), oneDark];

const loading = ref(false);
const websites = ref();
const data = ref({
    enable: false,
    content: '',
});
const confirmDialogRef = ref();
const tailLog = ref(false);
let timer: NodeJS.Timer | null = null;

const content = ref('');
const end = ref(false);
const lastContent = ref('');
const editorContainer = ref<HTMLDivElement | null>(null);

const logReq = reactive({
    id: undefined,
    operate: 'get',
    logType: 'access.log',
    page: 0,
    pageSize: 500,
});

const getWebsites = async () => {
    loading.value = true;
    await ListWebsites()
        .then((res) => {
            websites.value = res.data || [];
            if (websites.value.length > 0) {
                logReq.id = websites.value[0].id;
                search();
                nextTick(() => {
                    let editorElement = editorContainer.value.querySelector('.cm-editor');
                    let scrollerElement = editorElement.querySelector('.cm-scroller') as HTMLElement;
                    if (scrollerElement) {
                        scrollerElement.addEventListener('scroll', function () {
                            if (isScrolledToBottom(scrollerElement)) {
                                search();
                            }
                        });
                    }
                });
            }
        })
        .finally(() => {
            loading.value = false;
        });
};

const view = shallowRef();
const handleReady = (payload) => {
    view.value = payload.view;
    editorContainer.value = payload.container;
};

const changeType = (type: string) => {
    logReq.logType = type;
    if (logReq.id != undefined) {
        logReq.page = 0;
        logReq.pageSize = 500;
        search();
    }
};

const search = () => {
    if (!end.value) {
        logReq.page += 1;
    }
    OpWebsiteLog(logReq).then((res) => {
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
        } else {
            content.value = '';
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

const onClean = async () => {
    const params = {
        header: i18n.global.t('logs.deleteLogs'),
        operationInfo: i18n.global.t('commons.msg.delete'),
        submitInputInfo: i18n.global.t('logs.deleteLogs'),
    };
    confirmDialogRef.value!.acceptParams(params);
};

const onDownload = async () => {
    downloadWithContent(data.value.content, logReq.logType + '-' + dateFormatForName(new Date()) + '.log');
};

const changeTail = () => {
    if (tailLog.value) {
        timer = setInterval(() => {
            search();
        }, 1000 * 5);
    } else {
        onCloseLog();
    }
};

const onCloseLog = async () => {
    tailLog.value = false;
    clearInterval(Number(timer));
    timer = null;
};

const onSubmitClean = async () => {
    search();
    const req = {
        id: logReq.id,
        operate: 'delete',
        logType: logReq.logType,
    };
    loading.value = true;
    OpWebsiteLog(req)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .finally(() => {
            loading.value = false;
        });
};

function isScrolledToBottom(element: HTMLElement): boolean {
    return element.scrollTop + element.clientHeight === element.scrollHeight;
}

onMounted(() => {
    logReq.logType = 'access.log';
    getWebsites();
});
</script>
