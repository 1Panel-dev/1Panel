<template>
    <div>
        <MonitorRouter />

        <div class="content-container__search">
            <el-card>
                <div :class="mobile ? 'flx-wrap' : 'flex justify-between'">
                    <el-date-picker
                        @change="searchGlobal()"
                        v-model="timeRangeGlobal"
                        type="datetimerange"
                        range-separator="-"
                        :start-placeholder="$t('commons.search.timeStart')"
                        :end-placeholder="$t('commons.search.timeEnd')"
                        :shortcuts="shortcuts"
                        style="max-width: 360px; width: 100%"
                        :size="mobile ? 'small' : 'default'"
                    ></el-date-picker>
                    <TableRefresh class="float-right" @search="searchGlobal()" />
                </div>
            </el-card>
        </div>
        <el-row :gutter="7" class="card-interval">
            <el-col :span="24">
                <el-card style="overflow: inherit">
                    <template #header>
                        <div :class="mobile ? 'flx-wrap' : 'flex justify-between'">
                            <span class="title">{{ $t('monitor.avgLoad') }}</span>
                            <el-date-picker
                                @change="search('load')"
                                v-model="timeRangeLoad"
                                type="datetimerange"
                                range-separator="-"
                                :start-placeholder="$t('commons.search.timeStart')"
                                :end-placeholder="$t('commons.search.timeEnd')"
                                :shortcuts="shortcuts"
                                style="max-width: 360px; width: 100%"
                                :size="mobile ? 'small' : 'default'"
                            ></el-date-picker>
                        </div>
                    </template>
                    <div class="chart">
                        <v-charts
                            height="400px"
                            id="loadLoadChart"
                            type="line"
                            :option="chartsOption['loadLoadChart']"
                            v-if="chartsOption['loadLoadChart']"
                            :dataZoom="true"
                        />
                    </div>
                </el-card>
            </el-col>
        </el-row>
        <el-row :gutter="7" class="card-interval">
            <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                <el-card style="overflow: inherit">
                    <template #header>
                        <div :class="mobile ? 'flx-wrap' : 'flex justify-between'">
                            <span class="title">CPU</span>
                            <el-date-picker
                                @change="search('cpu')"
                                v-model="timeRangeCpu"
                                type="datetimerange"
                                range-separator="-"
                                :start-placeholder="$t('commons.search.timeStart')"
                                :end-placeholder="$t('commons.search.timeEnd')"
                                :shortcuts="shortcuts"
                                style="max-width: 360px; width: 100%"
                                :size="mobile ? 'small' : 'default'"
                            ></el-date-picker>
                        </div>
                    </template>
                    <div class="chart">
                        <v-charts
                            height="400px"
                            id="loadCPUChart"
                            type="line"
                            :option="chartsOption['loadCPUChart']"
                            v-if="chartsOption['loadCPUChart']"
                            :dataZoom="true"
                        />
                    </div>
                </el-card>
            </el-col>
            <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                <el-card style="overflow: inherit">
                    <template #header>
                        <div :class="mobile ? 'flx-wrap' : 'flex justify-between'">
                            <span class="title">{{ $t('monitor.memory') }}</span>
                            <el-date-picker
                                @change="search('memory')"
                                v-model="timeRangeMemory"
                                type="datetimerange"
                                range-separator="-"
                                :start-placeholder="$t('commons.search.timeStart')"
                                :end-placeholder="$t('commons.search.timeEnd')"
                                :shortcuts="shortcuts"
                                style="max-width: 360px; width: 100%"
                                :size="mobile ? 'small' : 'default'"
                            ></el-date-picker>
                        </div>
                    </template>
                    <div class="chart">
                        <v-charts
                            height="400px"
                            id="loadMemoryChart"
                            type="line"
                            :option="chartsOption['loadMemoryChart']"
                            v-if="chartsOption['loadMemoryChart']"
                            :dataZoom="true"
                        />
                    </div>
                </el-card>
            </el-col>
        </el-row>
        <el-row :gutter="7" class="card-interval">
            <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                <el-card style="overflow: inherit">
                    <template #header>
                        <div :class="mobile ? 'flx-wrap' : 'flex justify-between'">
                            <div>
                                <span class="title">{{ $t('monitor.disk') }} I/O{{ $t('commons.colon') }}</span>
                                <el-dropdown max-height="300px">
                                    <span class="networkOption">
                                        {{ ioChoose === 'all' ? $t('commons.table.all') : ioChoose }}
                                    </span>
                                    <template #dropdown>
                                        <el-dropdown-menu>
                                            <div v-for="item in ioOptions" :key="item">
                                                <el-dropdown-item v-if="item === 'all'" @click="changeIO('all')">
                                                    {{ $t('commons.table.all') }}
                                                </el-dropdown-item>
                                                <el-dropdown-item v-else @click="changeIO(item)">
                                                    {{ item }}
                                                </el-dropdown-item>
                                            </div>
                                        </el-dropdown-menu>
                                    </template>
                                </el-dropdown>
                            </div>
                            <el-date-picker
                                @change="search('io')"
                                v-model="timeRangeIO"
                                type="datetimerange"
                                range-separator="-"
                                :start-placeholder="$t('commons.search.timeStart')"
                                :end-placeholder="$t('commons.search.timeEnd')"
                                :shortcuts="shortcuts"
                                style="max-width: 360px; width: 100%"
                                :size="mobile ? 'small' : 'default'"
                            ></el-date-picker>
                        </div>
                    </template>
                    <div class="chart">
                        <v-charts
                            height="400px"
                            id="loadIOChart"
                            type="line"
                            :option="chartsOption['loadIOChart']"
                            v-if="chartsOption['loadIOChart']"
                            :dataZoom="true"
                        />
                    </div>
                </el-card>
            </el-col>
            <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                <el-card style="overflow: inherit">
                    <template #header>
                        <div :class="mobile ? 'flx-wrap' : 'flex justify-between'">
                            <div>
                                <span class="title">{{ $t('monitor.network') }}{{ $t('commons.colon') }}</span>
                                <el-dropdown max-height="300px">
                                    <span class="networkOption">
                                        {{ networkChoose === 'all' ? $t('commons.table.all') : networkChoose }}
                                    </span>
                                    <template #dropdown>
                                        <el-dropdown-menu>
                                            <div v-for="item in netOptions" :key="item">
                                                <el-dropdown-item v-if="item === 'all'" @click="changeNetwork('all')">
                                                    {{ $t('commons.table.all') }}
                                                </el-dropdown-item>
                                                <el-dropdown-item v-else @click="changeNetwork(item)">
                                                    {{ item }}
                                                </el-dropdown-item>
                                            </div>
                                        </el-dropdown-menu>
                                    </template>
                                </el-dropdown>
                            </div>
                            <el-date-picker
                                @change="search('network')"
                                v-model="timeRangeNetwork"
                                type="datetimerange"
                                range-separator="-"
                                :start-placeholder="$t('commons.search.timeStart')"
                                :end-placeholder="$t('commons.search.timeEnd')"
                                :shortcuts="shortcuts"
                                style="max-width: 360px; width: 100%"
                                :size="mobile ? 'small' : 'default'"
                            ></el-date-picker>
                        </div>
                    </template>
                    <div class="chart">
                        <v-charts
                            height="400px"
                            id="loadNetworkChart"
                            type="line"
                            :option="chartsOption['loadNetworkChart']"
                            v-if="chartsOption['loadNetworkChart']"
                            :dataZoom="true"
                        />
                    </div>
                </el-card>
            </el-col>
        </el-row>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import { loadMonitor, getNetworkOptions, getIOOptions } from '@/api/modules/host';
import { computeSize, computeSizeFromKBs, dateFormat } from '@/utils/util';
import i18n from '@/lang';
import MonitorRouter from '@/views/host/monitor/index.vue';
import { GlobalStore } from '@/store';
import { shortcuts } from '@/utils/shortcuts';
import { Host } from '@/api/interface/host';

const globalStore = GlobalStore();

const mobile = computed(() => {
    return globalStore.isMobile();
});

const monitorBase = ref();
const timeRangeGlobal = ref<[Date, Date]>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
const timeRangeLoad = ref<[Date, Date]>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
const timeRangeCpu = ref<[Date, Date]>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
const timeRangeMemory = ref<[Date, Date]>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
const timeRangeIO = ref<[Date, Date]>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
const timeRangeNetwork = ref<[Date, Date]>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
const networkChoose = ref();
const netOptions = ref();
const ioChoose = ref();
const ioOptions = ref();
const chartsOption = ref({ loadLoadChart: null, loadCPUChart: null, loadMemoryChart: null, loadNetworkChart: null });

const searchTime = ref();
const searchInfo = reactive<Host.MonitorSearch>({
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
            searchInfo.info = ioChoose.value;
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
                    return dateFormat(null, null, item);
                });
                if (param === 'cpu' || param === 'all') {
                    initCPUCharts(baseDate, item);
                }
                if (param === 'memory' || param === 'all') {
                    initMemCharts(baseDate, item);
                }
                if (param === 'load' || param === 'all') {
                    initLoadCharts(item);
                }
                break;
            case 'io':
                initIOCharts(item);
                break;
            case 'network':
                initNetCharts(item);
                break;
        }
    }
};

const changeNetwork = (item: string) => {
    networkChoose.value = item;
    search('network');
};

const changeIO = (item: string) => {
    ioChoose.value = item;
    search('io');
};

const loadNetworkOptions = async () => {
    const res = await getNetworkOptions();
    netOptions.value = res.data;
    searchInfo.info = globalStore.defaultNetwork || (netOptions.value && netOptions.value[0]);
    networkChoose.value = searchInfo.info;
    search('all');
};

const loadIOOptions = async () => {
    const res = await getIOOptions();
    ioOptions.value = res.data;
    searchInfo.info = globalStore.defaultIO || (ioOptions.value && ioOptions.value[0]);
    ioChoose.value = searchInfo.info;
    search('all');
};

function initLoadCharts(item: Host.MonitorData) {
    let itemLoadDate = item.date.length === 0 ? loadEmptyDate(timeRangeLoad.value) : item.date;
    let loadDate = itemLoadDate.map(function (item: any) {
        return dateFormat(null, null, item);
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
        return { value: item.loadUsage.toFixed(2), top: item.topCPUItems, unit: '%' };
    });
    loadUsage = loadUsage.length === 0 ? loadTopEmptyData() : loadUsage;
    chartsOption.value['loadLoadChart'] = {
        xData: loadDate,
        yData: [
            {
                name: '1 ' + i18n.global.t('commons.units.minute', 1),
                data: load1Data,
            },
            {
                name: '5 ' + i18n.global.t('commons.units.minute', 5),
                data: load5Data,
            },
            {
                name: '15 ' + i18n.global.t('commons.units.minute', 15),
                data: load15Data,
            },
            {
                name: i18n.global.t('monitor.resourceUsage'),
                data: loadUsage,
                yAxisIndex: 1,
            },
        ],
        yAxis: [
            { type: 'value', name: i18n.global.t('monitor.loadDetail') },
            {
                type: 'value',
                name: i18n.global.t('monitor.resourceUsage') + ' ( % )',
                position: 'right',
                alignTicks: true,
            },
        ],
        grid: mobile.value ? { left: '15%', right: '15%', bottom: '20%' } : null,
        tooltip: {
            trigger: 'axis',
            formatter: function (datas: any) {
                return withCPUProcess(datas);
            },
        },
    };
}

function initCPUCharts(baseDate: any, items: Host.MonitorData) {
    let data = items.value.map(function (item: any) {
        return { value: item.cpu.toFixed(2), top: item.topCPUItems, unit: '%' };
    });
    data = data.length === 0 ? loadTopEmptyData() : data;
    chartsOption.value['loadCPUChart'] = {
        xData: baseDate,
        yData: [
            {
                name: 'CPU',
                data: data,
            },
        ],
        tooltip: {
            trigger: 'axis',
            formatter: function (datas: any) {
                return withCPUProcess(datas);
            },
        },

        formatStr: '%',
    };
}

function initMemCharts(baseDate: any, items: Host.MonitorData) {
    let data = items.value.map(function (item: any) {
        return { value: item.memory.toFixed(2), top: item.topMemItems };
    });
    data = data.length === 0 ? loadTopEmptyData() : data;
    chartsOption.value['loadMemoryChart'] = {
        xData: baseDate,
        yData: [
            {
                name: i18n.global.t('monitor.memory'),
                data: data,
            },
        ],
        tooltip: {
            trigger: 'axis',
            formatter: function (datas: any) {
                return withMemProcess(datas);
            },
        },

        formatStr: '%',
    };
}

function initNetCharts(item: Host.MonitorData) {
    let networkDate = item.date.length === 0 ? loadEmptyDate(timeRangeNetwork.value) : item.date;
    let date = networkDate.map(function (item: any) {
        return dateFormat(null, null, item);
    });
    let networkUp = item.value.map(function (item: any) {
        return item.up.toFixed(2);
    });
    networkUp = networkUp.length === 0 ? loadEmptyData() : networkUp;
    let networkOut = item.value.map(function (item: any) {
        return item.down.toFixed(2);
    });
    networkOut = networkOut.length === 0 ? loadEmptyData() : networkOut;

    chartsOption.value['loadNetworkChart'] = {
        xData: date,
        yData: [
            {
                name: i18n.global.t('monitor.up'),
                data: networkUp,
            },
            {
                name: i18n.global.t('monitor.down'),
                data: networkOut,
            },
        ],
        tooltip: {
            trigger: 'axis',
            formatter: function (datas: any) {
                let res = loadDate(datas[0].name);
                for (const item of datas) {
                    res += loadSeries(item, computeSizeFromKBs(item.data), '');
                }
                return res;
            },
        },
        grid: {
            left: getSideWidth(true),
            right: getSideWidth(true),
            bottom: '20%',
        },
        formatStr: 'KB/s',
    };
}

function initIOCharts(item: Host.MonitorData) {
    let itemIODate = item.date?.length === 0 ? loadEmptyDate(timeRangeIO.value) : item.date;
    let ioDate = itemIODate.map(function (item: any) {
        return dateFormat(null, null, item);
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
    chartsOption.value['loadIOChart'] = {
        xData: ioDate,
        yData: [
            {
                name: i18n.global.t('monitor.read'),
                data: ioRead,
            },
            {
                name: i18n.global.t('monitor.write'),
                data: ioWrite,
            },
            {
                name: i18n.global.t('monitor.readWriteCount'),
                data: ioCount,
                yAxisIndex: 1,
            },
            {
                name: i18n.global.t('monitor.readWriteTime'),
                data: ioTime,
                yAxisIndex: 1,
            },
        ],
        tooltip: {
            trigger: 'axis',
            formatter: function (datas: any) {
                let res = loadDate(datas[0].name);
                for (const item of datas) {
                    if (
                        item.seriesName === i18n.global.t('monitor.read') ||
                        item.seriesName === i18n.global.t('monitor.write')
                    ) {
                        res += loadSeries(item, computeSizeFromKBs(item.data), '');
                    }
                    if (item.seriesName === i18n.global.t('monitor.readWriteCount')) {
                        res += loadSeries(item, item.data, i18n.global.t('commons.units.time') + '/s');
                    }
                    if (item.seriesName === i18n.global.t('monitor.readWriteTime')) {
                        res += loadSeries(item, item.data, ' ms');
                    }
                }
                return res;
            },
        },
        grid: { left: getSideWidth(true), right: getSideWidth(true), bottom: '20%' },
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
    };
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
function loadTopEmptyData() {
    return [{ value: 0, top: 0, unit: '' }];
}

function withCPUProcess(datas: any) {
    let tops;
    let res = loadDate(datas[0].name);
    for (const item of datas) {
        if (item.data?.top) {
            tops = item.data?.top;
        }
        res += loadSeries(item, item.data.value ? item.data.value : item.data, item.data.unit || '');
    }
    if (!tops) {
        return '';
    }
    res += `
        <div style="margin-top: 10px; border-bottom: 1px dashed black;"></div>
        <table style="border-collapse: collapse; margin-top: 20px; font-size: 12px;">
        <thead>
            <tr>
            <th style="padding: 6px 8px;">PID</th>
            <th style="padding: 6px 8px;">${i18n.global.t('commons.table.user')}</th>
            <th style="padding: 6px 8px;">${i18n.global.t('menu.process')}</th>
            <th style="padding: 6px 8px;">${i18n.global.t('monitor.percent')}</th>
            </tr>
        </thead>
        <tbody>
    `;
    for (const row of tops) {
        res += `
            <tr>
                <td style="padding: 6px 8px; text-align: center;">
                    ${row.pid}
                </td>
                <td style="padding: 6px 8px; text-align: center;">
                    ${row.user}
                </td>
                <td style="padding: 6px 8px; text-align: center;">
                    ${row.name}
                </td>
                <td style="padding: 6px 8px; text-align: center;">
                    ${row.percent.toFixed(2)}%
                </td>
            </tr>
        `;
    }
    return res;
}

function withMemProcess(datas: any) {
    let res = loadDate(datas[0].name);
    for (const item of datas) {
        res += loadSeries(item, item.data.value ? item.data.value : item.data, ' %');
    }
    if (!datas[0].data.top) {
        return res;
    }
    res += `
        <div style="margin-top: 10px; border-bottom: 1px dashed black;"></div>
        <table style="border-collapse: collapse; margin-top: 20px; font-size: 12px;">
        <thead>
            <tr>
                <th style="padding: 6px 8px;">PID</th>
                <th style="padding: 6px 8px;">${i18n.global.t('commons.table.user')}</th>
                <th style="padding: 6px 8px;">${i18n.global.t('menu.process')}</th>
                <th style="padding: 6px 8px;">${i18n.global.t('monitor.memory')}</th>
                <th style="padding: 6px 8px;">${i18n.global.t('monitor.percent')}</th>
            </tr>
        </thead>
        <tbody>
    `;
    for (const item of datas) {
        for (const row of item.data.top) {
            res += `
                  <tr>
                    <td style="padding: 6px 8px; text-align: center;">
                      <span style="display: inline-block;"></span>
                      ${row.pid}
                    </td>
                    <td style="padding: 6px 8px; text-align: center;">
                      ${row.user}
                    </td>
                    <td style="padding: 6px 8px; text-align: center;">
                      ${row.name}
                    </td>
                    <td style="padding: 6px 8px; text-align: center;">
                      ${computeSize(row.memory)}
                    </td>
                    <td style="padding: 6px 8px; text-align: center;">
                      ${row.percent.toFixed(2)}%
                    </td>
                  </tr>
                `;
        }
    }
    return res;
}

function loadDate(name: any) {
    return ` <div style="display: inline-block; width: 100%; padding-bottom: 10px;">
                ${i18n.global.t('commons.search.date')}: ${name}
            </div>`;
}

function loadSeries(item: any, data: any, unit: any) {
    return `<div style="width: 100%;">
                ${item.marker} ${item.seriesName}: ${data} ${unit}
            </div>`;
}

function getSideWidth(b: boolean) {
    return !b || document.body.clientWidth > 1600 ? '7%' : '10%';
}

onMounted(() => {
    loadNetworkOptions();
    loadIOOptions();
});
</script>

<style scoped lang="scss">
.content-container__search {
    margin-top: 7px;
    .el-card {
        --el-card-padding: 12px;
    }
}
.networkOption {
    font-size: 16px;
    font-weight: 500;
    margin-top: 3px;
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
