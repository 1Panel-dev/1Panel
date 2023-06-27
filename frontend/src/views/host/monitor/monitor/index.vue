<template>
    <div>
        <MonitorRouter />

        <div class="content-container__search">
            <el-card>
                <span style="font-size: 14px">{{ $t('monitor.globalFilter') }}</span>
                <el-date-picker
                    @change="searchGlobal()"
                    v-model="timeRangeGlobal"
                    type="datetimerange"
                    :range-separator="$t('commons.search.timeRange')"
                    :start-placeholder="$t('commons.search.timeStart')"
                    :end-placeholder="$t('commons.search.timeEnd')"
                    :shortcuts="shortcuts"
                    style="max-width: 360px; width: 100%; margin-left: 10px"
                    :size="mobile ? 'small' : 'default'"
                ></el-date-picker>
            </el-card>
        </div>
        <el-row :gutter="20" style="margin-top: 20px">
            <el-col :span="24">
                <el-card style="overflow: inherit">
                    <template #header>
                        <div :class="mobile ? 'flx-wrap' : 'flx-justify-between'">
                            <span class="title">{{ $t('monitor.avgLoad') }}</span>
                            <el-date-picker
                                @change="search('load')"
                                v-model="timeRangeLoad"
                                type="datetimerange"
                                :range-separator="$t('commons.search.timeRange')"
                                :start-placeholder="$t('commons.search.timeStart')"
                                :end-placeholder="$t('commons.search.timeEnd')"
                                :shortcuts="shortcuts"
                                style="max-width: 360px; width: 100%"
                                :size="mobile ? 'small' : 'default'"
                            ></el-date-picker>
                        </div>
                    </template>
                    <div id="loadLoadChart" class="chart"></div>
                </el-card>
            </el-col>
        </el-row>
        <el-row :gutter="20" style="margin-top: 20px">
            <el-col :xs="24" :sm="12" :md="12" :lg="12" :xl="12">
                <el-card style="overflow: inherit">
                    <template #header>
                        <div :class="mobile ? 'flx-wrap' : 'flx-justify-between'">
                            <span class="title">CPU</span>
                            <el-date-picker
                                @change="search('cpu')"
                                v-model="timeRangeCpu"
                                type="datetimerange"
                                :range-separator="$t('commons.search.timeRange')"
                                :start-placeholder="$t('commons.search.timeStart')"
                                :end-placeholder="$t('commons.search.timeEnd')"
                                :shortcuts="shortcuts"
                                style="max-width: 360px; width: 100%"
                                :size="mobile ? 'small' : 'default'"
                            ></el-date-picker>
                        </div>
                    </template>
                    <div id="loadCPUChart" class="chart"></div>
                </el-card>
            </el-col>
            <el-col :xs="24" :sm="12" :md="12" :lg="12" :xl="12">
                <el-card style="overflow: inherit">
                    <template #header>
                        <div :class="mobile ? 'flx-wrap' : 'flx-justify-between'">
                            <span class="title">{{ $t('monitor.memory') }}</span>
                            <el-date-picker
                                @change="search('memory')"
                                v-model="timeRangeMemory"
                                type="datetimerange"
                                :range-separator="$t('commons.search.timeRange')"
                                :start-placeholder="$t('commons.search.timeStart')"
                                :end-placeholder="$t('commons.search.timeEnd')"
                                :shortcuts="shortcuts"
                                style="max-width: 360px; width: 100%"
                                :size="mobile ? 'small' : 'default'"
                            ></el-date-picker>
                        </div>
                    </template>
                    <div id="loadMemoryChart" class="chart"></div>
                </el-card>
            </el-col>
        </el-row>
        <el-row :gutter="20" style="margin-top: 20px">
            <el-col :xs="24" :sm="12" :md="12" :lg="12" :xl="12">
                <el-card style="overflow: inherit">
                    <template #header>
                        <div :class="mobile ? 'flx-wrap' : 'flx-justify-between'">
                            <span class="title">{{ $t('monitor.disk') }} IO</span>
                            <el-date-picker
                                @change="search('io')"
                                v-model="timeRangeIO"
                                type="datetimerange"
                                :range-separator="$t('commons.search.timeRange')"
                                :start-placeholder="$t('commons.search.timeStart')"
                                :end-placeholder="$t('commons.search.timeEnd')"
                                :shortcuts="shortcuts"
                                style="max-width: 360px; width: 100%"
                                :size="mobile ? 'small' : 'default'"
                            ></el-date-picker>
                        </div>
                    </template>
                    <div id="loadIOChart" class="chart"></div>
                </el-card>
            </el-col>
            <el-col :xs="24" :sm="12" :md="12" :lg="12" :xl="12">
                <el-card style="overflow: inherit">
                    <template #header>
                        <div :class="mobile ? 'flx-wrap' : 'flx-justify-between'">
                            <div>
                                <span class="title">{{ $t('monitor.network') }} IO:</span>
                                <el-popover placement="bottom" :width="200" trigger="click">
                                    <el-select @change="search('network')" v-model="networkChoose">
                                        <template #prefix>{{ $t('monitor.networkCard') }}</template>
                                        <div v-for="item in netOptions" :key="item">
                                            <el-option
                                                v-if="item === 'all'"
                                                :label="$t('commons.table.all')"
                                                :value="item"
                                            />
                                            <el-option v-else :label="item" :value="item" />
                                        </div>
                                    </el-select>
                                    <template #reference>
                                        <span class="networkOption" v-if="networkChoose === 'all'">
                                            {{ $t('commons.table.all') }}
                                        </span>
                                        <span v-else class="networkOption">
                                            {{ networkChoose }}
                                        </span>
                                    </template>
                                </el-popover>
                            </div>
                            <el-date-picker
                                @change="search('network')"
                                v-model="timeRangeNetwork"
                                type="datetimerange"
                                :range-separator="$t('commons.search.timeRange')"
                                :start-placeholder="$t('commons.search.timeStart')"
                                :end-placeholder="$t('commons.search.timeEnd')"
                                :shortcuts="shortcuts"
                                style="max-width: 360px; width: 100%"
                                :size="mobile ? 'small' : 'default'"
                            ></el-date-picker>
                        </div>
                    </template>
                    <div id="loadNetworkChart" class="chart"></div>
                </el-card>
            </el-col>
        </el-row>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onBeforeUnmount, computed } from 'vue';
import * as echarts from 'echarts';
import { loadMonitor, getNetworkOptions } from '@/api/modules/monitor';
import { Monitor } from '@/api/interface/monitor';
import { computeSizeFromKBs, dateFormatWithoutYear } from '@/utils/util';
import i18n from '@/lang';
import MonitorRouter from '@/views/host/monitor/index.vue';
import { GlobalStore } from '@/store';

const globalStore = GlobalStore();

const mobile = computed(() => {
    return globalStore.isMobile();
});

const zoomStart = ref();
const monitorBase = ref();
const timeRangeGlobal = ref<[Date, Date]>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
const timeRangeLoad = ref<[Date, Date]>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
const timeRangeCpu = ref<[Date, Date]>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
const timeRangeMemory = ref<[Date, Date]>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
const timeRangeIO = ref<[Date, Date]>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
const timeRangeNetwork = ref<[Date, Date]>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
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
        text: i18n.global.t('monitor.yesterday'),
        value: () => {
            const yesterday = new Date(new Date().getTime() - 3600 * 1000 * 24 * 1);
            const end = new Date(yesterday.setHours(23, 59, 59, 999));
            const start = new Date(yesterday.setHours(0, 0, 0, 0));
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

const searchGlobal = () => {
    timeRangeLoad.value = timeRangeGlobal.value;
    timeRangeCpu.value = timeRangeGlobal.value;
    timeRangeMemory.value = timeRangeGlobal.value;
    timeRangeIO.value = timeRangeGlobal.value;
    timeRangeNetwork.value = timeRangeGlobal.value;
    search('load');
    search('cpu');
    search('memory');
    search('io');
    search('network');
};

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
                let baseDate = item.date.length === 0 ? loadEmptyDate(timeRangeCpu.value) : item.date;
                baseDate = baseDate.map(function (item: any) {
                    return dateFormatWithoutYear(item);
                });
                if (param === 'cpu' || param === 'all') {
                    let cpuData = item.value.map(function (item: any) {
                        return item.cpu.toFixed(2);
                    });
                    cpuData = cpuData.length === 0 ? loadEmptyData() : cpuData;
                    let yDatasOfCpu = {
                        name: 'CPU',
                        type: 'line',
                        areaStyle: {
                            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                                {
                                    offset: 0,
                                    color: 'rgba(0, 94, 235, .5)',
                                },
                                {
                                    offset: 1,
                                    color: 'rgba(0, 94, 235, 0)',
                                },
                            ]),
                        },
                        data: cpuData,
                        showSymbol: false,
                    };
                    initCharts('loadCPUChart', baseDate, yDatasOfCpu, 'CPU', '%');
                }
                if (param === 'memory' || param === 'all') {
                    let memoryData = item.value.map(function (item: any) {
                        return item.memory.toFixed(2);
                    });
                    memoryData = memoryData.length === 0 ? loadEmptyData() : memoryData;
                    let yDatasOfMem = {
                        name: i18n.global.t('monitor.memory'),
                        type: 'line',
                        areaStyle: {
                            color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                                {
                                    offset: 0,
                                    color: 'rgba(0, 94, 235, .5)',
                                },
                                {
                                    offset: 1,
                                    color: 'rgba(0, 94, 235, 0)',
                                },
                            ]),
                        },
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
                let networkDate = item.date.length === 0 ? loadEmptyDate(timeRangeNetwork.value) : item.date;
                networkDate = networkDate.map(function (item: any) {
                    return dateFormatWithoutYear(item);
                });
                let networkUp = item.value.map(function (item: any) {
                    return item.up.toFixed(2);
                });
                networkUp = networkUp.length === 0 ? loadEmptyData() : networkUp;
                let yDatasOfUp = {
                    name: i18n.global.t('monitor.up'),
                    type: 'line',
                    areaStyle: {
                        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                            {
                                offset: 0,
                                color: 'rgba(0, 94, 235, .5)',
                            },
                            {
                                offset: 1,
                                color: 'rgba(0, 94, 235, 0)',
                            },
                        ]),
                    },
                    data: networkUp,
                    showSymbol: false,
                };
                let networkOut = item.value.map(function (item: any) {
                    return item.down.toFixed(2);
                });
                networkOut = networkOut.length === 0 ? loadEmptyData() : networkOut;
                let yDatasOfDown = {
                    name: i18n.global.t('monitor.down'),
                    type: 'line',
                    areaStyle: {
                        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                            {
                                offset: 0,
                                color: 'rgba(27, 143, 60, .5)',
                            },
                            {
                                offset: 1,
                                color: 'rgba(27, 143, 60, 0)',
                            },
                        ]),
                    },
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
                if (chartName !== 'loadNetworkChart') {
                    for (const item of datas) {
                        res += item.marker + ' ' + item.seriesName + '：' + item.data + formatStr + '<br/>';
                    }
                } else {
                    for (const item of datas) {
                        res += item.marker + ' ' + item.seriesName + '：' + computeSizeFromKBs(item.data) + '<br/>';
                    }
                }
                return res;
            },
        },
        legend: {
            data: chartName === 'loadNetworkChart' && [i18n.global.t('monitor.up'), i18n.global.t('monitor.down')],
        },
        grid: {
            left: getSideWidth(chartName == 'loadNetworkChart'),
            right: getSideWidth(chartName == 'loadNetworkChart'),
            bottom: '20%',
        },
        xAxis: { data: xDatas },
        yAxis: { name: '( ' + formatStr + ' )', axisLabel: { fontSize: chartName == 'loadNetworkChart' ? 10 : 12 } },
        dataZoom: [{ startValue: zoomStart.value }],
        series: yDatas,
    };
    lineChart.setOption(option, true);
}

function initLoadCharts(item: Monitor.MonitorData) {
    let itemLoadDate = item.date.length === 0 ? loadEmptyDate(timeRangeLoad.value) : item.date;
    let loadDate = itemLoadDate.map(function (item: any) {
        return dateFormatWithoutYear(item);
    });
    let load1Data = item.value.map(function (item: any) {
        return item.cpuLoad1.toFixed(2);
    });
    load1Data = load1Data.length === 0 ? loadEmptyData() : load1Data;
    let load5Data = item.value.map(function (item: any) {
        return item.cpuLoad5.toFixed(2);
    });
    load5Data = load5Data.length === 0 ? loadEmptyData() : load5Data;
    let load15Data = item.value.map(function (item: any) {
        return item.cpuLoad15.toFixed(2);
    });
    load15Data = load15Data.length === 0 ? loadEmptyData() : load15Data;
    let loadUsage = item.value.map(function (item: any) {
        return item.loadUsage.toFixed(2);
    });
    loadUsage = loadUsage.length === 0 ? loadEmptyData() : loadUsage;

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
            data: loadDate,
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
                areaStyle: {
                    color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                        {
                            offset: 0,
                            color: 'rgba(0, 94, 235, .5)',
                        },
                        {
                            offset: 1,
                            color: 'rgba(0, 94, 235, 0)',
                        },
                    ]),
                },
                showSymbol: false,
                data: load1Data,
            },
            {
                name: '5 ' + i18n.global.t('monitor.min'),
                type: 'line',
                areaStyle: {
                    color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                        {
                            offset: 0,
                            color: 'rgba(27, 143, 60, .5)',
                        },
                        {
                            offset: 1,
                            color: 'rgba(27, 143, 60, 0)',
                        },
                    ]),
                },
                showSymbol: false,
                data: load5Data,
            },
            {
                name: '15 ' + i18n.global.t('monitor.min'),
                type: 'line',
                areaStyle: {
                    color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                        {
                            offset: 0,
                            color: 'rgba(249, 199, 79, .5)',
                        },
                        {
                            offset: 1,
                            color: 'rgba(249, 199, 79, 0)',
                        },
                    ]),
                },
                showSymbol: false,
                data: load15Data,
            },
            {
                name: i18n.global.t('monitor.resourceUsage'),
                type: 'line',
                areaStyle: {
                    color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                        {
                            offset: 0,
                            color: 'rgba(255, 173, 177, 0.5)',
                        },
                        {
                            offset: 1,
                            color: 'rgba(255, 173, 177, 0)',
                        },
                    ]),
                },
                showSymbol: false,
                data: loadUsage,
                yAxisIndex: 1,
            },
        ],
    };
    lineChart.setOption(option, true);
}

function initIOCharts(item: Monitor.MonitorData) {
    let itemIODate = item.date?.length === 0 ? loadEmptyDate(timeRangeIO.value) : item.date;
    let ioDate = itemIODate.map(function (item: any) {
        return dateFormatWithoutYear(item);
    });
    let ioRead = item.value.map(function (item: any) {
        return Number((item.read / 1024).toFixed(2));
    });
    ioRead = ioRead.length === 0 ? loadEmptyData() : ioRead;
    let ioWrite = item.value.map(function (item: any) {
        return Number((item.write / 1024).toFixed(2));
    });
    ioWrite = ioWrite.length === 0 ? loadEmptyData() : ioWrite;
    let ioCount = item.value.map(function (item: any) {
        return item.count;
    });
    ioCount = ioCount.length === 0 ? loadEmptyData() : ioCount;
    let ioTime = item.value.map(function (item: any) {
        return item.time;
    });
    ioTime = ioTime.length === 0 ? loadEmptyData() : ioTime;

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
                        res += item.marker + ' ' + item.seriesName + '：' + computeSizeFromKBs(item.data) + '<br/>';
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
        grid: { left: getSideWidth(true), right: getSideWidth(true), bottom: '20%' },
        xAxis: {
            data: ioDate,
        },
        yAxis: [
            { type: 'value', name: '( KB/s )', axisLabel: { fontSize: 10 } },
            {
                type: 'value',
                position: 'right',
                alignTicks: true,
                axisLabel: {
                    fontSize: 10,
                },
            },
        ],
        dataZoom: [{ startValue: zoomStart.value }],
        series: [
            {
                name: i18n.global.t('monitor.read'),
                type: 'line',
                areaStyle: {
                    color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                        {
                            offset: 0,
                            color: 'rgba(0, 94, 235, .5)',
                        },
                        {
                            offset: 1,
                            color: 'rgba(0, 94, 235, 0)',
                        },
                    ]),
                },
                showSymbol: false,
                data: ioRead,
            },
            {
                name: i18n.global.t('monitor.write'),
                type: 'line',
                areaStyle: {
                    color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                        {
                            offset: 0,
                            color: 'rgba(27, 143, 60, .5)',
                        },
                        {
                            offset: 1,
                            color: 'rgba(27, 143, 60, 0)',
                        },
                    ]),
                },
                showSymbol: false,
                data: ioWrite,
            },
            {
                name: i18n.global.t('monitor.readWriteCount'),
                type: 'line',
                areaStyle: {
                    color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                        {
                            offset: 0,
                            color: 'rgba(249, 199, 79, .5)',
                        },
                        {
                            offset: 1,
                            color: 'rgba(249, 199, 79, 0)',
                        },
                    ]),
                },
                showSymbol: false,
                data: ioCount,
                yAxisIndex: 1,
            },
            {
                name: i18n.global.t('monitor.readWriteTime'),
                type: 'line',
                areaStyle: {
                    color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                        {
                            offset: 0,
                            color: 'rgba(255, 173, 177, 0.5)',
                        },
                        {
                            offset: 1,
                            color: 'rgba(255, 173, 177, 0)',
                        },
                    ]),
                },
                showSymbol: false,
                data: ioTime,
                yAxisIndex: 1,
            },
        ],
    };
    lineChart.setOption(option, true);
}

function loadEmptyDate(timeRange: any) {
    if (timeRange.length != 2) {
        return;
    }
    let date1 = new Date(timeRange[0]);
    let date2 = new Date(timeRange[1]);
    return [date1, date2];
}
function loadEmptyData() {
    return [0, 0];
}

function getSideWidth(b: boolean) {
    return !b || document.body.clientWidth > 1600 ? '7%' : '10%';
}

function changeChartSize() {
    echarts.getInstanceByDom(document.getElementById('loadLoadChart') as HTMLElement)?.resize();
    echarts.getInstanceByDom(document.getElementById('loadCPUChart') as HTMLElement)?.resize();
    echarts.getInstanceByDom(document.getElementById('loadMemoryChart') as HTMLElement)?.resize();
    echarts.getInstanceByDom(document.getElementById('loadIOChart') as HTMLElement)?.resize();
    echarts.getInstanceByDom(document.getElementById('loadNetworkChart') as HTMLElement)?.resize();
}

onMounted(() => {
    zoomStart.value = dateFormatWithoutYear(new Date(new Date().setHours(0, 0, 0, 0)));
    loadNetworkOptions();
    window.addEventListener('resize', changeChartSize);
});
onBeforeUnmount(() => {
    window.removeEventListener('resize', changeChartSize);
});
</script>

<style scoped lang="scss">
.content-container__search {
    margin-top: 20px;
    .el-card {
        --el-card-padding: 12px;
    }
}
.networkOption {
    font-size: 16px;
    font-weight: 500;
    margin-left: 5px;
    cursor: pointer;
    color: var(--el-color-primary);
}
.title {
    font-size: 16px;
    font-weight: 500;
}
.chart {
    width: 100%;
    height: 400px;
}
</style>
