<template>
    <DrawerPro v-model="open" :header="title" :back="handleClose" size="large" :fullScreen="true">
        <template #content>
            <div v-if="req.file != 'config'">
                <el-tabs v-model="req.file" type="card" @tab-click="handleChange">
                    <el-tab-pane :label="$t('logs.runLog')" name="out.log"></el-tab-pane>
                    <el-tab-pane :label="$t('logs.errLog')" name="err.log"></el-tab-pane>
                </el-tabs>
                <el-checkbox border v-model="tailLog" class="float-left" @change="changeTail">
                    {{ $t('commons.button.watch') }}
                </el-checkbox>
                <el-button class="ml-5" @click="cleanLog" icon="Delete">
                    {{ $t('commons.button.clean') }}
                </el-button>
            </div>
            <br />
            <div v-loading="loading">
                <CodemirrorPro class="mt-5" v-model="content" :heightDiff="400"></CodemirrorPro>
            </div>
        </template>

        <template #footer>
            <span>
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
                <el-button type="primary" :disabled="loading" @click="submit()" v-if="req.file === 'config'">
                    {{ $t('commons.button.confirm') }}
                </el-button>
            </span>
        </template>
    </DrawerPro>
    <OpDialog ref="opRef" @search="getContent" />
</template>
<script lang="ts" setup>
import { onUnmounted, reactive, ref } from 'vue';
import { operateSupervisorProcessFile } from '@/api/modules/host-tool';
import i18n from '@/lang';
import { TabsPaneContext } from 'element-plus';
import { MsgSuccess } from '@/utils/message';

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

let timer: NodeJS.Timer | null = null;

const em = defineEmits(['search']);

const getContent = () => {
    loading.value = true;
    operateSupervisorProcessFile(req)
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
    operateSupervisorProcessFile(updateReq)
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
};

const cleanLog = async () => {
    let log = req.file === 'out.log' ? i18n.global.t('logs.runLog') : i18n.global.t('logs.errLog');
    opRef.value.acceptParams({
        title: i18n.global.t('commons.msg.clean'),
        names: [req.name],
        msg: i18n.global.t('commons.msg.operatorHelper', [log, i18n.global.t('commons.msg.clean')]),
        api: operateSupervisorProcessFile,
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
