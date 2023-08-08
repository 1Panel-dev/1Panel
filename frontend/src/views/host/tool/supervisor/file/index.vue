<template>
    <el-drawer :close-on-click-modal="false" v-model="open" size="50%">
        <template #header>
            <DrawerHeader :header="title" :back="handleClose"></DrawerHeader>
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
                style="height: calc(100vh - 430px)"
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
</template>
<script lang="ts" setup>
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { onMounted, onUnmounted, reactive, ref, shallowRef } from 'vue';
import { useDeleteData } from '@/hooks/use-delete-data';
import { OperateSupervisorProcessFile } from '@/api/modules/host-tool';
import i18n from '@/lang';
import { TabsPaneContext } from 'element-plus';
import { MsgSuccess } from '@/utils/message';

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

const view = shallowRef();
const handleReady = (payload) => {
    view.value = payload.view;
};
let timer: NodeJS.Timer | null = null;

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
            MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
            getContent();
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
};

const cleanLog = async () => {
    const clearReq = {
        name: req.name,
        operate: 'clear',
        file: req.file,
    };
    try {
        await useDeleteData(OperateSupervisorProcessFile, clearReq, 'commons.msg.delete');
        getContent();
    } catch (error) {
    } finally {
    }
};

const onCloseLog = async () => {
    tailLog.value = false;
    clearInterval(Number(timer));
    timer = null;
};

onMounted(() => {
    getContent();
});

onUnmounted(() => {
    onCloseLog();
});

defineExpose({
    acceptParams,
});
</script>
