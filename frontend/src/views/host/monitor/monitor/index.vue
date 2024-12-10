<template>
    <div>
        <MonitorRouter />

        <div class="content-container__search">
            <el-card>
                <div :class="mobile ? 'flx-wrap' : 'flx-justify-between'">
                    <el-date-picker
                        @change="searchGlobal()"
                        v-model="timeRangeGlobal"
                        type="datetimerange"
                        :range-separator="$t('commons.search.timeRange')"
                        :start-placeholder="$t('commons.search.timeStart')"
                        :end-placeholder="$t('commons.search.timeEnd')"
                        :shortcuts="shortcuts"
                        style="max-width: 360px; width: 100%"
                        :size="mobile ? 'small' : 'default'"
                    ></el-date-picker>
                </div>
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
        <el-row :gutter="20" style="margin-top: 20px">
            <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
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
        <el-row :gutter="20" style="margin-top: 20px">
            <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                <el-card style="overflow: inherit">
                    <template #header>
                        <div :class="mobile ? 'flx-wrap' : 'flx-justify-between'">
                            <span class="title">{{ $t('monitor.disk') }} I/O</span>
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
                        <div :class="mobile ? 'flx-wrap' : 'flx-justify-between'">
                            <div>
                                <span class="title">{{ $t('monitor.network') }}{{ $t('commons.colon') }}</span>
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
import { loadMonitor, getNetworkOptions } from '@/api/modules/host';
import { computeSizeFromKBs, dateFormatWithoutYear } from '@/utils/util';
import i18n from '@/lang';
import MonitorRouter from '@/views/host/monitor/index.vue';
import { GlobalStore } from '@/store';
import { shortcuts } from '@/utils/shortcuts';
import { Host } from '@/api/interface/host';

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
                    chartsOption.value['loadCPUChart'] = {
                        xData: baseDate,
                        yData: [
                            {
                                name: 'CPU',
                                data: cpuData,
                            },
                        ],

                        formatStr: '%',
                    };
                }
                if (param === 'memory' || param === 'all') {
                    let memoryData = item.value.map(function (item: any) {
                        return item.memory.toFixed(2);
                    });
                    memoryData = memoryData.length === 0 ? loadEmptyData() : memoryData;
                    chartsOption.value['loadMemoryChart'] = {
                        xData: baseDate,
                        yData: [
                            {
                                name: i18n.global.t('monitor.memory'),
                                data: memoryData,
                            },
                        ],

                        formatStr: '%',
                    };
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
                let networkOut = item.value.map(function (item: any) {
                    return item.down.toFixed(2);
                });
                networkOut = networkOut.length === 0 ? loadEmptyData() : networkOut;

                chartsOption.value['loadNetworkChart'] = {
                    xData: networkDate,
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
                    grid: {
                        left: getSideWidth(true),
                        right: getSideWidth(true),
                        bottom: '20%',
                    },
                    formatStr: 'KB/s',
                };
        }
    }
};

const loadNetworkOptions = async () => {
    const res = await getNetworkOptions();
    netOptions.value = res.data;
    searchInfo.info = globalStore.defaultNetwork || (netOptions.value && netOptions.value[0]);
    networkChoose.value = searchInfo.info;
    search('all');
};

function initLoadCharts(item: Host.MonitorData) {
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
                let res = datas[0].name + '<br/>';
                for (const item of datas) {
                    if (item.seriesName === i18n.global.t('monitor.resourceUsage')) {
                        res +=
                            item.marker + ' ' + item.seriesName + i18n.global.t('commons.colon') + item.data + '%<br/>';
                    } else {
                        res +=
                            item.marker + ' ' + item.seriesName + i18n.global.t('commons.colon') + item.data + '<br/>';
                    }
                }
                return res;
            },
        },
    };
}

function initIOCharts(item: Host.MonitorData) {
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
                let res = datas[0].name + '<br/>';
                for (const item of datas) {
                    if (
                        item.seriesName === i18n.global.t('monitor.read') ||
                        item.seriesName === i18n.global.t('monitor.write')
                    ) {
                        res +=
                            item.marker +
                            ' ' +
                            item.seriesName +
                            i18n.global.t('commons.colon') +
                            computeSizeFromKBs(item.data) +
                            '<br/>';
                    }
                    if (item.seriesName === i18n.global.t('monitor.readWriteCount')) {
                        res +=
                            item.marker +
                            ' ' +
                            item.seriesName +
                            i18n.global.t('commons.colon') +
                            item.data +
                            ' ' +
                            i18n.global.t('commons.units.time') +
                            '/s' +
                            '<br/>';
                    }
                    if (item.seriesName === i18n.global.t('monitor.readWriteTime')) {
                        res +=
                            item.marker +
                            ' ' +
                            item.seriesName +
                            i18n.global.t('commons.colon') +
                            item.data +
                            ' ms' +
                            '<br/>';
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

function getSideWidth(b: boolean) {
    return !b || document.body.clientWidth > 1600 ? '7%' : '10%';
}

onMounted(() => {
    zoomStart.value = dateFormatWithoutYear(new Date(new Date().setHours(0, 0, 0, 0)));
    loadNetworkOptions();
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
