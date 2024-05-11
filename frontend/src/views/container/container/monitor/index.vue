<template>
    <el-drawer
        v-model="monitorVisible"
        :destroy-on-close="true"
        @close="handleClose"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="50%"
    >
        <template #header>
            <DrawerHeader :header="$t('container.monitor')" :resource="title" :back="handleClose" />
        </template>
        <el-form label-position="top" @submit.prevent>
            <el-form-item :label="$t('container.refreshTime')">
                <el-select v-model="timeInterval" @change="changeTimer">
                    <el-option label="3s" :value="3" />
                    <el-option label="5s" :value="5" />
                    <el-option label="10s" :value="10" />
                    <el-option label="30s" :value="30" />
                    <el-option label="60s" :value="60" />
                </el-select>
            </el-form-item>
        </el-form>
        <el-card>
            <v-charts
                height="200px"
                id="cpuChart"
                type="line"
                :option="chartsOption['cpuChart']"
                v-if="chartsOption['cpuChart']"
            />
        </el-card>
        <el-card style="margin-top: 10px">
            <v-charts
                height="200px"
                id="memoryChart"
                type="line"
                :option="chartsOption['memoryChart']"
                v-if="chartsOption['memoryChart']"
            />
        </el-card>
        <el-card style="margin-top: 10px">
            <v-charts
                height="200px"
                id="ioChart"
                type="line"
                :option="chartsOption['ioChart']"
                v-if="chartsOption['ioChart']"
            />
        </el-card>
        <el-card style="margin-top: 10px">
            <v-charts
                height="200px"
                id="networkChart"
                type="line"
                :option="chartsOption['networkChart']"
                v-if="chartsOption['networkChart']"
            />
        </el-card>
    </el-drawer>
</template>

<script lang="ts" setup>
import { onBeforeUnmount, ref } from 'vue';
import { containerStats } from '@/api/modules/container';
import { dateFormatForSecond } from '@/utils/util';
import VCharts from '@/components/v-charts/index.vue';
import i18n from '@/lang';
import DrawerHeader from '@/components/drawer-header/index.vue';

const title = ref();
const monitorVisible = ref(false);
const timeInterval = ref();
let timer: NodeJS.Timer | null = null;
let isInit = ref<boolean>(true);
interface DialogProps {
    containerID: string;
    container: string;
}
const dialogData = ref<DialogProps>({
    containerID: '',
    container: '',
});

const acceptParams = async (params: DialogProps): Promise<void> => {
    monitorVisible.value = true;
    dialogData.value.containerID = params.containerID;
    title.value = params.container;
    cpuDatas.value = [];
    memDatas.value = [];
    cacheDatas.value = [];
    ioReadDatas.value = [];
    ioWriteDatas.value = [];
    netTxDatas.value = [];
    netRxDatas.value = [];
    timeDatas.value = [];
    timeInterval.value = 5;
    isInit.value = true;
    loadData();
    timer = setInterval(async () => {
        if (monitorVisible.value) {
            isInit.value = false;
            loadData();
        }
    }, 1000 * timeInterval.value);
};

const cpuDatas = ref<Array<string>>([]);
const memDatas = ref<Array<string>>([]);
const cacheDatas = ref<Array<string>>([]);
const ioReadDatas = ref<Array<string>>([]);
const ioWriteDatas = ref<Array<string>>([]);
const netTxDatas = ref<Array<string>>([]);
const netRxDatas = ref<Array<string>>([]);
const timeDatas = ref<Array<string>>([]);
const chartsOption = ref({ cpuChart: null, memoryChart: null, ioChart: null, networkChart: null });

const changeTimer = () => {
    clearInterval(Number(timer));
    timer = setInterval(async () => {
        if (monitorVisible.value) {
            loadData();
        }
    }, 1000 * timeInterval.value);
};

const loadData = async () => {
    const res = await containerStats(dialogData.value.containerID);
    cpuDatas.value.push(res.data.cpuPercent.toFixed(2));
    if (cpuDatas.value.length > 20) {
        cpuDatas.value.splice(0, 1);
    }
    memDatas.value.push(res.data.memory.toFixed(2));
    if (memDatas.value.length > 20) {
        memDatas.value.splice(0, 1);
    }
    cacheDatas.value.push(res.data.cache.toFixed(2));
    if (cacheDatas.value.length > 20) {
        cacheDatas.value.splice(0, 1);
    }
    ioReadDatas.value.push(res.data.ioRead.toFixed(2));
    if (ioReadDatas.value.length > 20) {
        ioReadDatas.value.splice(0, 1);
    }
    ioWriteDatas.value.push(res.data.ioWrite.toFixed(2));
    if (ioWriteDatas.value.length > 20) {
        ioWriteDatas.value.splice(0, 1);
    }
    netTxDatas.value.push(res.data.networkTX.toFixed(2));
    if (netTxDatas.value.length > 20) {
        netTxDatas.value.splice(0, 1);
    }
    netRxDatas.value.push(res.data.networkRX.toFixed(2));
    if (netRxDatas.value.length > 20) {
        netRxDatas.value.splice(0, 1);
    }
    timeDatas.value.push(dateFormatForSecond(res.data.shotTime));
    if (timeDatas.value.length > 20) {
        timeDatas.value.splice(0, 1);
    }

    chartsOption.value['cpuChart'] = {
        title: 'CPU',
        xData: timeDatas.value,
        yData: [
            {
                name: 'CPU',
                data: cpuDatas.value,
            },
        ],
        formatStr: '%',
    };

    chartsOption.value['memoryChart'] = {
        title: i18n.global.t('monitor.memory'),
        xData: timeDatas.value,
        yData: [
            {
                name: i18n.global.t('monitor.memory'),
                data: memDatas.value,
            },
            {
                name: i18n.global.t('container.cache'),
                data: cacheDatas.value,
            },
        ],
        formatStr: 'MB',
    };

    chartsOption.value['ioChart'] = {
        title: i18n.global.t('monitor.disk') + ' IO',
        xData: timeDatas.value,
        yData: [
            {
                name: i18n.global.t('monitor.read'),
                data: ioReadDatas.value,
            },
            {
                name: i18n.global.t('monitor.write'),
                data: ioWriteDatas.value,
            },
        ],
        formatStr: 'MB',
    };

    chartsOption.value['networkChart'] = {
        title: i18n.global.t('monitor.network'),
        xData: timeDatas.value,
        yData: [
            {
                name: i18n.global.t('monitor.up'),
                data: netTxDatas.value,
            },
            {
                name: i18n.global.t('monitor.down'),
                data: netRxDatas.value,
            },
        ],
        formatStr: 'KB',
    };
};
const handleClose = async () => {
    monitorVisible.value = false;
    clearInterval(Number(timer));
    timer = null;
    chartsOption.value = { cpuChart: null, memoryChart: null, ioChart: null, networkChart: null };
};

onBeforeUnmount(() => {
    handleClose;
});

defineExpose({
    acceptParams,
});
</script>
