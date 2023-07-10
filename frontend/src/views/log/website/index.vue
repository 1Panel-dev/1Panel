<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('logs.websiteLog')">
            <template #toolbar>
                <el-row>
                    <el-col :span="16">
                        <el-button
                            class="tag-button"
                            :label="'access.log'"
                            :type="logReq.logType === 'access.log' ? 'primary' : ''"
                            @click="changeType('access.log')"
                        >
                            {{ $t('logs.runLog') }}
                        </el-button>
                        <el-button
                            class="tag-button"
                            :label="'error.log'"
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
                    v-model="data.content"
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
const logReq = reactive({
    id: undefined,
    operate: 'get',
    logType: 'access.log',
});
const data = ref({
    enable: false,
    content: '',
});
const confirmDialogRef = ref();
const tailLog = ref(false);
let timer: NodeJS.Timer | null = null;

const getWebsites = async () => {
    loading.value = true;
    await ListWebsites()
        .then((res) => {
            websites.value = res.data || [];
            if (websites.value.length > 0) {
                logReq.id = websites.value[0].id;
                search();
            }
        })
        .finally(() => {
            loading.value = false;
        });
};

const view = shallowRef();
const handleReady = (payload) => {
    view.value = payload.view;
};

const changeType = (type: string) => {
    logReq.logType = type;
    if (logReq.id != undefined) {
        search();
    }
};

const search = () => {
    loading.value = true;
    OpWebsiteLog(logReq)
        .then((res) => {
            data.value = res.data;
            nextTick(() => {
                const state = view.value.state;
                view.value.dispatch({
                    selection: { anchor: state.doc.length, head: state.doc.length },
                    scrollIntoView: true,
                });
            });
        })
        .finally(() => {
            loading.value = false;
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

onMounted(() => {
    getWebsites();
});
</script>

<style lang="scss" scoped>
.tag-button {
    &.el-button--primary:hover {
        background-color: var(--el-color-primary) !important;
    }
    &.el-button--primary:focus {
        background-color: var(--el-color-primary) !important;
    }
}
</style>
