<template>
    <el-drawer v-model="monitorVisiable" :destroy-on-close="true" :close-on-click-modal="false" size="50%">
        <template #header>
            <DrawerHeader :header="$t('container.monitor')" :back="handleClose" />
        </template>
        <el-form label-position="top">
            <el-form-item :label="$t('container.refreshTime')">
                <el-select v-model="timeInterval" @change="changeTimer">
                    <el-option label="1s" :value="1" />
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
            <!-- <div id="cpuChart" style="width: 100%; height: 200px"></div> -->
        </el-card>
        <el-card style="margin-top: 10px">
            <v-charts
                height="200px"
                id="memoryChart"
                type="line"
                :option="chartsOption['memoryChart']"
                v-if="chartsOption['memoryChart']"
            />
            <!-- <div id="memoryChart" style="width: 100%; height: 200px"></div> -->
        </el-card>
        <el-card style="margin-top: 10px">
            <v-charts
                height="200px"
                id="ioChart"
                type="line"
                :option="chartsOption['ioChart']"
                v-if="chartsOption['ioChart']"
            />
            <!-- <div id="ioChart" style="width: 100%; height: 200px"></div> -->
        </el-card>
        <el-card style="margin-top: 10px">
            <v-charts
                height="200px"
                id="networkChart"
                type="line"
                :option="chartsOption['networkChart']"
                v-if="chartsOption['networkChart']"
            />
            <!-- <div id="networkChart" style="width: 100%; height: 200px"></div> -->
        </el-card>
    </el-drawer>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { ContainerStats } from '@/api/modules/container';
import { dateFormatForSecond } from '@/utils/util';
import VCharts from '@/components/v-charts/index.vue';
// import * as echarts from 'echarts';
import i18n from '@/lang';
import DrawerHeader from '@/components/drawer-header/index.vue';

const monitorVisiable = ref(false);
const timeInterval = ref();
let timer: NodeJS.Timer | null = null;
let isInit = ref<boolean>(true);
interface DialogProps {
    containerID: string;
}
const dialogData = ref<DialogProps>({
    containerID: '',
});

// function changeChartSize() {
//     echarts.getInstanceByDom(document.getElementById('cpuChart') as HTMLElement)?.resize();
//     echarts.getInstanceByDom(document.getElementById('memoryChart') as HTMLElement)?.resize();
//     echarts.getInstanceByDom(document.getElementById('ioChart') as HTMLElement)?.resize();
//     echarts.getInstanceByDom(document.getElementById('networkChart') as HTMLElement)?.resize();
// }
const acceptParams = async (params: DialogProps): Promise<void> => {
    monitorVisiable.value = true;
    dialogData.value.containerID = params.containerID;
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
    // window.addEventListener('resize', changeChartSize);
    timer = setInterval(async () => {
        if (monitorVisiable.value) {
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
        if (monitorVisiable.value) {
            loadData();
        }
    }, 1000 * timeInterval.value);
};

const loadData = async () => {
    const res = await ContainerStats(dialogData.value.containerID);
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
        xDatas: timeDatas.value,
        yDatas: [
            {
                name: 'CPU',
                data: cpuDatas.value,
            },
        ],
        formatStr: '%',
    };

    chartsOption.value['memoryChart'] = {
        title: i18n.global.t('monitor.memory'),
        xDatas: timeDatas.value,
        yDatas: [
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
        xDatas: timeDatas.value,
        yDatas: [
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
        xDatas: timeDatas.value,
        yDatas: [
            {
                name: i18n.global.t('monitor.up'),
                data: netTxDatas.value,
            },
            {
                name: i18n.global.t('monitor.down'),
                data: netRxDatas.value,
            },
        ],
        formatStr: 'KB/s',
    };
    // let cpuYDatas = {
    //     name: 'CPU',
    //     type: 'line',
    //     areaStyle: {
    //         color: '#ebdee3',
    //     },
    //     data: cpuDatas.value,
    //     showSymbol: false,
    // };
    // freshChart('cpuChart', ['CPU'], timeDatas.value, [cpuYDatas], 'CPU', '%');

    // let memoryYDatas = {
    //     name: i18n.global.t('monitor.memory'),
    //     type: 'line',
    //     areaStyle: {
    //         color: '#ebdee3',
    //     },
    //     data: memDatas.value,
    //     showSymbol: false,
    // };
    // let cacheYDatas = {
    //     name: i18n.global.t('container.cache'),
    //     type: 'line',
    //     areaStyle: {
    //         color: '#ebdee3',
    //     },
    //     data: cacheDatas.value,
    //     showSymbol: false,
    // };
    // freshChart(
    //     'memoryChart',
    //     [i18n.global.t('monitor.memory'), i18n.global.t('monitor.cache')],
    //     timeDatas.value,
    //     [memoryYDatas, cacheYDatas],
    //     i18n.global.t('monitor.memory'),
    //     ' MB',
    // );

    // let ioReadYDatas = {
    //     name: i18n.global.t('monitor.read'),
    //     type: 'line',
    //     areaStyle: {
    //         color: '#ebdee3',
    //     },
    //     data: ioReadDatas.value,
    //     showSymbol: false,
    // };
    // let ioWriteYDatas = {
    //     name: i18n.global.t('monitor.write'),
    //     type: 'line',
    //     areaStyle: {
    //         color: '#ebdee3',
    //     },
    //     data: ioWriteDatas.value,
    //     showSymbol: false,
    // };
    // freshChart(
    //     'ioChart',
    //     [i18n.global.t('monitor.read'), i18n.global.t('monitor.write')],
    //     timeDatas.value,
    //     [ioReadYDatas, ioWriteYDatas],
    //     i18n.global.t('monitor.disk') + ' IO',
    //     'MB',
    // );

    // let netTxYDatas = {
    //     name: i18n.global.t('monitor.up'),
    //     type: 'line',
    //     areaStyle: {
    //         color: '#ebdee3',
    //     },
    //     data: netTxDatas.value,
    //     showSymbol: false,
    // };
    // let netRxYDatas = {
    //     name: i18n.global.t('monitor.down'),
    //     type: 'line',
    //     areaStyle: {
    //         color: '#ebdee3',
    //     },
    //     data: netRxDatas.value,
    //     showSymbol: false,
    // };
    // freshChart(
    //     'networkChart',
    //     [i18n.global.t('monitor.up'), i18n.global.t('monitor.down')],
    //     timeDatas.value,
    //     [netTxYDatas, netRxYDatas],
    //     i18n.global.t('monitor.network'),
    //     'KB/s',
    // );
};

// function freshChart(chartName: string, legendDatas: any, xDatas: any, yDatas: any, yTitle: string, formatStr: string) {
//     if (isInit.value) {
//         echarts.init(document.getElementById(chartName) as HTMLElement);
//     }
//     let itemChart = echarts.getInstanceByDom(document.getElementById(chartName) as HTMLElement);
//     const option = {
//         title: [
//             {
//                 left: 'center',
//                 text: yTitle,
//             },
//         ],
//         zlevel: 1,
//         z: 1,
//         tooltip: {
//             trigger: 'axis',
//             formatter: function (datas: any) {
//                 let res = datas[0].name + '<br/>';
//                 for (const item of datas) {
//                     res += item.marker + ' ' + item.seriesName + '：' + item.data + formatStr + '<br/>';
//                 }
//                 return res;
//             },
//         },
//         grid: { left: '7%', right: '7%', bottom: '20%' },
//         legend: {
//             data: legendDatas,
//             right: 10,
//         },
//         xAxis: { data: xDatas, boundaryGap: false },
//         yAxis: { name: '( ' + formatStr + ' )' },
//         series: yDatas,
//     };
//     itemChart?.setOption(option, true);
// }

const handleClose = async () => {
    monitorVisiable.value = false;
    clearInterval(Number(timer));
    timer = null;
    // window.removeEventListener('resize', changeChartSize);
};

defineExpose({
    acceptParams,
});
</script>
