<template>
    <DrawerPro
        v-model="open"
        :header="$t('commons.button.log') + ' - ' + supervisorName"
        @close="handleClose"
        :size="globalStore.isFullScreen ? 'full' : 'large'"
        :fullScreen="true"
    >
        <template #content>
            <el-tabs v-model="tab" type="card" @tab-click="handleChange">
                <el-tab-pane :label="$t('logs.runLog')" name="out.log"></el-tab-pane>
                <el-tab-pane :label="$t('logs.errLog')" name="err.log"></el-tab-pane>
            </el-tabs>

            <LogFile
                :key="logKey"
                ref="logRef"
                v-if="openLog"
                :config="logConfig"
                v-model:loading="loading"
                v-model:has-content="hasContent"
                :height-diff="300"
            >
                <template #button>
                    <el-button @click="cleanLog" icon="Delete" :disabled="hasContent === false">
                        {{ $t('commons.button.clean') }}
                    </el-button>
                </template>
            </LogFile>
        </template>

        <template #footer>
            <span>
                <el-button @click="handleClose" :disabled="loading">{{ $t('commons.button.cancel') }}</el-button>
            </span>
        </template>
    </DrawerPro>
    <OpDialog ref="opRef" @search="refreshLog" />
</template>
<script lang="ts" setup>
import LogFile from '@/components/log/file/index.vue';

import { onUnmounted, reactive, ref } from 'vue';
import { GlobalStore } from '@/store';
import { operateSupervisorProcessFile } from '@/api/modules/host-tool';
import i18n from '@/lang';
import { TabsPaneContext } from 'element-plus';
const globalStore = GlobalStore();

const logConfig = reactive({
    type: 'supervisor',
    id: undefined,
    name: '',
    colorMode: 'nginx',
});
const loading = ref(false);
const open = ref(false);
const tab = ref('out.log');
const openLog = ref();
const supervisorName = ref('');
const opRef = ref();
const logKey = ref(0);
const hasContent = ref(false);

const em = defineEmits(['close']);

const handleChange = (tab: TabsPaneContext) => {
    openLog.value = false;
    logConfig.name = supervisorName.value + '.' + tab.props.name.toString();
    logKey.value++;
    openLog.value = true;
};

const handleClose = () => {
    open.value = false;
};

const acceptParams = (name: string) => {
    tab.value = 'out.log';
    supervisorName.value = name;
    logConfig.name = name + '.' + tab.value;
    open.value = true;
    openLog.value = true;
};

const onCloseLog = async () => {
    em('close');
};

const refreshLog = () => {
    logKey.value++;
};

const cleanLog = async () => {
    let log = tab.value === 'out.log' ? i18n.global.t('logs.runLog') : i18n.global.t('logs.errLog');
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.clean'),
        names: [supervisorName.value],
        msg: i18n.global.t('commons.msg.operatorHelper', [log, i18n.global.t('commons.button.clean')]),
        api: operateSupervisorProcessFile,
        params: { name: supervisorName.value, operate: 'clear', file: tab.value },
    });
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
