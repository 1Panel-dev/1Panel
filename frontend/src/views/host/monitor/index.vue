<template>
    <div>
        <el-row :gutter="20">
            <el-col :span="24">
                <el-card style="overflow: inherit">
                    <template #header>
                        <span style="font-size: 16px; font-weight: 500">{{ $t('monitor.avgLoad') }}</span>
                        <el-date-picker
                            @change="search('load')"
                            v-model="timeRangeLoad"
                            type="datetimerange"
                            :range-separator="$t('commons.search.timeRange')"
                            :start-placeholder="$t('commons.search.timeStart')"
                            :end-placeholder="$t('commons.search.timeEnd')"
                            :shortcuts="shortcuts"
                            style="float: right; right: 20px"
                        ></el-date-picker>
                    </template>
                    <div id="loadLoadChart" style="width: 100%; height: 400px"></div>
                </el-card>
            </el-col>
        </el-row>
        <el-row :gutter="20" style="margin-top: 20px">
            <el-col :span="12">
                <el-card style="overflow: inherit">
                    <template #header>
                        <span style="font-size: 16px; font-weight: 500">CPU</span>
                        <el-date-picker
                            @change="search('cpu')"
                            v-model="timeRangeCpu"
                            type="datetimerange"
                            :range-separator="$t('commons.search.timeRange')"
                            :start-placeholder="$t('commons.search.timeStart')"
                            :end-placeholder="$t('commons.search.timeEnd')"
                            :shortcuts="shortcuts"
                            style="float: right; right: 20px"
                        ></el-date-picker>
                    </template>
                    <div id="loadCPUChart" style="width: 100%; height: 400px"></div>
                </el-card>
            </el-col>
            <el-col :span="12">
                <el-card style="overflow: inherit">
                    <template #header>
                        <span style="font-size: 16px; font-weight: 500">{{ $t('monitor.memory') }}</span>
                        <el-date-picker
                            @change="search('memory')"
                            v-model="timeRangeMemory"
                            type="datetimerange"
                            :range-separator="$t('commons.search.timeRange')"
                            :start-placeholder="$t('commons.search.timeStart')"
                            :end-placeholder="$t('commons.search.timeEnd')"
                            :shortcuts="shortcuts"
                            style="float: right; right: 20px"
                        ></el-date-picker>
                    </template>
                    <div id="loadMemoryChart" style="width: 100%; height: 400px"></div>
                </el-card>
            </el-col>
        </el-row>
        <el-row :gutter="20" style="margin-top: 20px">
            <el-col :span="12">
                <el-card style="overflow: inherit">
                    <template #header>
                        <span style="font-size: 16px; font-weight: 500">{{ $t('monitor.disk') }} IO</span>
                        <el-date-picker
                            @change="search('io')"
                            v-model="timeRangeIO"
                            type="datetimerange"
                            :range-separator="$t('commons.search.timeRange')"
                            :start-placeholder="$t('commons.search.timeStart')"
                            :end-placeholder="$t('commons.search.timeEnd')"
                            :shortcuts="shortcuts"
                            style="float: right; right: 20px"
                        ></el-date-picker>
                    </template>
                    <div id="loadIOChart" style="width: 100%; height: 400px"></div>
                </el-card>
            </el-col>
            <el-col :span="12">
                <el-card style="overflow: inherit">
                    <template #header>
                        <span style="font-size: 16px; font-weight: 500">{{ $t('monitor.network') }} IO</span>
                        <el-select
                            v-model="networkChoose"
                            clearable
                            filterable
                            @change="search('network')"
                            style="margin-left: 20px"
                            placeholder="Select"
                        >
                            <el-option v-for="item in netOptions" :key="item" :label="item" :value="item" />
                        </el-select>
                        <el-date-picker
                            @change="search('network')"
                            v-model="timeRangeNetwork"
                            type="datetimerange"
                            :range-separator="$t('commons.search.timeRange')"
                            :start-placeholder="$t('commons.search.timeStart')"
                            :end-placeholder="$t('commons.search.timeEnd')"
                            :shortcuts="shortcuts"
                            style="float: right; right: 20px"
                        ></el-date-picker>
                    </template>
                    <div id="loadNetworkChart" style="width: 100%; height: 400px"></div>
                </el-card>
            </el-col>
        </el-row>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onBeforeUnmount } from 'vue';
import * as echarts from 'echarts';
import { loadMonitor, getNetworkOptions } from '@/api/modules/monitor';
import { Monitor } from '@/api/interface/monitor';
import { dateFromatWithoutYear } from '@/utils/util';
import i18n from '@/lang';

const zoomStart = ref();
const monitorBase = ref();
const timeRangeLoad = ref<Array<any>>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
const timeRangeCpu = ref<Array<any>>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
const timeRangeMemory = ref<Array<any>>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
const timeRangeIO = ref<Array<any>>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
const timeRangeNetwork = ref<Array<any>>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
const networkChoose = ref();
const netOptions = ref();
const shortcuts = [
    {
        text: i18n.global.t('monitor.today'),
        value: () => {
            const end = new Date();
            const start = new Date(new Date().setHours(0, 0, 0, 0));
            return [start, end];
        },
    },
    {
        text: i18n.global.t('monitor.yestoday'),
        value: () => {
            const yestoday = new Date(new Date().getTime() - 3600 * 1000 * 24 * 1);
            const end = new Date(yestoday.setHours(23, 59, 59, 999));
            const start = new Date(yestoday.setHours(0, 0, 0, 0));
            return [start, end];
        },
    },
    {
        text: i18n.global.t('monitor.lastNDay', [3]),
        value: () => {
            const start = new Date(new Date().getTime() - 3600 * 1000 * 24 * 3);
            const end = new Date();
            return [start, end];
        },
    },
    {
        text: i18n.global.t('monitor.lastNDay', [7]),
        value: () => {
            const start = new Date(new Date().getTime() - 3600 * 1000 * 24 * 7);
            const end = new Date();
            return [start, end];
        },
    },
    {
        text: i18n.global.t('monitor.lastNDay', [30]),
        value: () => {
            const start = new Date(new Date().getTime() - 3600 * 1000 * 24 * 30);
            const end = new Date();
            return [start, end];
        },
    },
];
const searchTime = ref();
const searchInfo = reactive<Monitor.MonitorSearch>({
    param: '',
    info: '',
    startTime: new Date(new Date().setHours(0, 0, 0, 0)),
    endTime: new Date(),
});

const search = async (param: string) => {
    searchInfo.param = param;
    switch (param) {
        case 'load':
            searchTime.value = timeRangeLoad.value;
            break;
        case 'cpu':
            searchTime.value = timeRangeCpu.value;
            break;
        case 'memory':
            searchTime.value = timeRangeMemory.value;
            break;
        case 'io':
            searchTime.value = timeRangeIO.value;
            break;
        case 'network':
            searchTime.value = timeRangeNetwork.value;
            searchInfo.info = networkChoose.value;
            break;
    }
    if (searchTime.value && searchTime.value.length === 2) {
        searchInfo.startTime = searchTime.value[0];
        searchInfo.endTime = searchTime.value[1];
    }
    const res = await loadMonitor(searchInfo);
    monitorBase.value = res.data;
    for (const item of monitorBase.value) {
        if (!item.value) {
            item.value = [];
            item.date = [];
        }
        switch (item.param) {
            case 'base':
                let baseDate = item.date.map(function (item: any) {
                    return dateFromatWithoutYear(item);
                });
                if (param === 'cpu' || param === 'all') {
                    let cpuData = item.value.map(function (item: any) {
                        return item.cpu.toFixed(2);
                    });
                    let yDatasOfCpu = { name: 'CPU', type: 'line', data: cpuData, showSymbol: false };
                    initCharts('loadCPUChart', baseDate, yDatasOfCpu, 'CPU', '%');
                }
                if (param === 'memory' || param === 'all') {
                    let memoryData = item.value.map(function (item: any) {
                        return item.memory.toFixed(2);
                    });
                    let yDatasOfMem = {
                        name: i18n.global.t('monitor.memory'),
                        type: 'line',
                        data: memoryData,
                        showSymbol: false,
                    };
                    initCharts('loadMemoryChart', baseDate, yDatasOfMem, i18n.global.t('monitor.memory'), '%');
                }
                if (param === 'load' || param === 'all') {
                    initLoadCharts(item);
                }
                break;
            case 'io':
                initIOCharts(item);
                break;
            case 'network':
                let networkDate = item.date.map(function (item: any) {
                    return dateFromatWithoutYear(item);
                });
                let networkUp = item.value.map(function (item: any) {
                    return item.up.toFixed(2);
                });
                let yDatasOfUp = {
                    name: i18n.global.t('monitor.up'),
                    type: 'line',
                    data: networkUp,
                    showSymbol: false,
                };
                let networkOut = item.value.map(function (item: any) {
                    return item.down.toFixed(2);
                });
                let yDatasOfDown = {
                    name: i18n.global.t('monitor.down'),
                    type: 'line',
                    data: networkOut,
                    showSymbol: false,
                };
                initCharts('loadNetworkChart', networkDate, [yDatasOfUp, yDatasOfDown], 'KB/s', 'KB/s');
        }
    }
};

const loadNetworkOptions = async () => {
    const res = await getNetworkOptions();
    netOptions.value = res.data;
    searchInfo.info = netOptions.value && netOptions.value[0];
    networkChoose.value = searchInfo.info;
    search('all');
};

function initCharts(chartName: string, xDatas: any, yDatas: any, yTitle: string, formatStr: string) {
    const lineChart = echarts.init(document.getElementById(chartName) as HTMLElement);
    const option = {
        zlevel: 1,
        z: 1,
        tooltip: {
            trigger: 'axis',
            formatter: function (datas: any) {
                let res = datas[0].name + '<br/>';
                for (const item of datas) {
                    res += item.marker + ' ' + item.seriesName + '：' + item.data + formatStr + '<br/>';
                }
                return res;
            },
        },
        legend: {
            data: chartName === 'loadNetworkChart' && [i18n.global.t('monitor.up'), i18n.global.t('monitor.down')],
        },
        grid: { left: '7%', right: '7%', bottom: '20%' },
        xAxis: { data: xDatas },
        yAxis: { name: '( ' + formatStr + ' )' },
        dataZoom: [{ startValue: zoomStart.value }],
        series: yDatas,
    };
    lineChart.setOption(option, true);
}

function initLoadCharts(item: Monitor.MonitorData) {
    const lineChart = echarts.init(document.getElementById('loadLoadChart') as HTMLElement);
    const option = {
        zlevel: 1,
        z: 1,
        tooltip: {
            trigger: 'axis',
            formatter: function (datas: any) {
                let res = datas[0].name + '<br/>';
                for (const item of datas) {
                    res += item.marker + ' ' + item.seriesName + '：' + item.data + '%' + '<br/>';
                }
                return res;
            },
        },
        legend: {
            data: [
                '1 ' + i18n.global.t('monitor.min'),
                '5 ' + i18n.global.t('monitor.min'),
                '15 ' + i18n.global.t('monitor.min'),
                i18n.global.t('monitor.resourceUsage'),
            ],
        },
        grid: { left: '7%', right: '7%', bottom: '20%' },
        xAxis: {
            data: item.date.map(function (item: any) {
                return dateFromatWithoutYear(item);
            }),
        },
        yAxis: [
            { type: 'value', name: i18n.global.t('monitor.loadDetail') + ' ( % )' },
            {
                type: 'value',
                name: i18n.global.t('monitor.resourceUsage') + ' ( % )',
                position: 'right',
                alignTicks: true,
            },
        ],
        dataZoom: [{ startValue: zoomStart.value }],
        series: [
            {
                name: '1 ' + i18n.global.t('monitor.min'),
                type: 'line',
                showSymbol: false,
                data: item.value.map(function (item: any) {
                    return item.cpuLoad1.toFixed(2);
                }),
            },
            {
                name: '5 ' + i18n.global.t('monitor.min'),
                type: 'line',
                showSymbol: false,
                data: item.value.map(function (item: any) {
                    return item.cpuLoad5.toFixed(2);
                }),
            },
            {
                name: '15 ' + i18n.global.t('monitor.min'),
                type: 'line',
                showSymbol: false,
                data: item.value.map(function (item: any) {
                    return item.cpuLoad15.toFixed(2);
                }),
            },
            {
                name: i18n.global.t('monitor.resourceUsage'),
                type: 'line',
                showSymbol: false,
                data: item.value.map(function (item: any) {
                    return item.loadUsage.toFixed(2);
                }),
                yAxisIndex: 1,
            },
        ],
    };
    lineChart.setOption(option, true);
}

function initIOCharts(item: Monitor.MonitorData) {
    const lineChart = echarts.init(document.getElementById('loadIOChart') as HTMLElement);
    const option = {
        zlevel: 1,
        z: 1,
        tooltip: {
            trigger: 'axis',
            formatter: function (datas: any) {
                let res = datas[0].name + '<br/>';
                for (const item of datas) {
                    if (
                        item.seriesName === i18n.global.t('monitor.read') ||
                        item.seriesName === i18n.global.t('monitor.write')
                    ) {
                        res += item.marker + ' ' + item.seriesName + '：' + item.data + ' KB/s' + '<br/>';
                    }
                    if (item.seriesName === i18n.global.t('monitor.readWriteCount')) {
                        res +=
                            item.marker +
                            ' ' +
                            item.seriesName +
                            '：' +
                            item.data +
                            ' ' +
                            i18n.global.t('monitor.count') +
                            '/s' +
                            '<br/>';
                    }
                    if (item.seriesName === i18n.global.t('monitor.readWriteTime')) {
                        res += item.marker + ' ' + item.seriesName + '：' + item.data + ' ms' + '<br/>';
                    }
                }
                return res;
            },
        },
        legend: {
            data: [
                i18n.global.t('monitor.read'),
                i18n.global.t('monitor.write'),
                i18n.global.t('monitor.readWriteCount'),
                i18n.global.t('monitor.readWriteTime'),
            ],
        },
        grid: { left: '7%', right: '7%', bottom: '20%' },
        xAxis: {
            data: item.date.map(function (item: any) {
                return dateFromatWithoutYear(item);
            }),
        },
        yAxis: [
            { type: 'value', name: '( KB/s )' },
            { type: 'value', position: 'right', alignTicks: true },
        ],
        dataZoom: [{ startValue: zoomStart.value }],
        series: [
            {
                name: i18n.global.t('monitor.read'),
                type: 'line',
                showSymbol: false,
                data: item.value.map(function (item: any) {
                    return (item.read / 1024).toFixed(2);
                }),
            },
            {
                name: i18n.global.t('monitor.write'),
                type: 'line',
                showSymbol: false,
                data: item.value.map(function (item: any) {
                    return (item.write / 1024).toFixed(2);
                }),
            },
            {
                name: i18n.global.t('monitor.readWriteCount'),
                type: 'line',
                showSymbol: false,
                data: item.value.map(function (item: any) {
                    return item.count;
                }),
                yAxisIndex: 1,
            },
            {
                name: i18n.global.t('monitor.readWriteTime'),
                type: 'line',
                showSymbol: false,
                data: item.value.map(function (item: any) {
                    return item.time;
                }),
                yAxisIndex: 1,
            },
        ],
    };
    lineChart.setOption(option, true);
}

function changeChartSize() {
    echarts.getInstanceByDom(document.getElementById('loadLoadChart') as HTMLElement)?.resize();
    echarts.getInstanceByDom(document.getElementById('loadCPUChart') as HTMLElement)?.resize();
    echarts.getInstanceByDom(document.getElementById('loadMemoryChart') as HTMLElement)?.resize();
    echarts.getInstanceByDom(document.getElementById('loadIOChart') as HTMLElement)?.resize();
    echarts.getInstanceByDom(document.getElementById('loadNetworkChart') as HTMLElement)?.resize();
}

onMounted(() => {
    zoomStart.value = dateFromatWithoutYear(new Date(new Date().setHours(0, 0, 0, 0)));
    loadNetworkOptions();
    window.addEventListener('resize', changeChartSize);
});
onBeforeUnmount(() => {
    window.removeEventListener('resize', changeChartSize);
});
</script>
