<template>
    <div v-loading="loading">
        <RouterButton
            :buttons="[
                {
                    label: $t('aiTools.gpu.gpu'),
                    path: '/xpack/gpu',
                },
            ]"
        />

        <div class="content-container__search" v-if="options.length !== 0">
            <el-card>
                <div>
                    <el-date-picker
                        @change="search()"
                        v-model="timeRangeGlobal"
                        type="datetimerange"
                        range-separator="-"
                        :start-placeholder="$t('commons.search.timeStart')"
                        :end-placeholder="$t('commons.search.timeEnd')"
                        :shortcuts="shortcuts"
                        style="max-width: 360px; width: 100%"
                        :size="mobile ? 'small' : 'default'"
                    ></el-date-picker>
                    <el-select class="p-w-300 ml-2" v-model="searchInfo.productName" @change="search()">
                        <el-option v-for="item in options" :key="item" :label="item" :value="item" />
                    </el-select>
                    <TableRefresh class="float-right" @search="search()" />
                </div>
            </el-card>
        </div>
        <el-row :gutter="7" class="card-interval" v-if="options.length !== 0">
            <el-col :span="24">
                <el-card style="overflow: inherit">
                    <template #header>
                        <div :class="mobile ? 'flx-wrap' : 'flex justify-between'">
                            <span class="title">{{ $t('monitor.gpuUtil') }}</span>
                        </div>
                    </template>
                    <div class="chart">
                        <v-charts
                            height="400px"
                            id="loadGPUChart"
                            type="line"
                            :option="chartsOption['loadGPUChart']"
                            v-if="chartsOption['loadGPUChart']"
                            :dataZoom="true"
                        />
                    </div>
                </el-card>
            </el-col>
            <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                <el-card style="overflow: inherit">
                    <template #header>
                        <div :class="mobile ? 'flx-wrap' : 'flex justify-between'">
                            <span class="title">{{ $t('monitor.memoryUsage') }}</span>
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
            <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                <el-card style="overflow: inherit">
                    <template #header>
                        <div :class="mobile ? 'flx-wrap' : 'flex justify-between'">
                            <span class="title">{{ $t('monitor.powerUsage') }}</span>
                        </div>
                    </template>
                    <div class="chart">
                        <v-charts
                            height="400px"
                            id="loadPowerChart"
                            type="line"
                            :option="chartsOption['loadPowerChart']"
                            v-if="chartsOption['loadPowerChart']"
                            :dataZoom="true"
                        />
                    </div>
                </el-card>
            </el-col>
            <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                <el-card style="overflow: inherit">
                    <template #header>
                        <div>
                            {{ $t('monitor.temperature') }}
                            <el-tooltip placement="top" :content="$t('aiTools.gpu.temperatureHelper')">
                                <el-icon size="15"><InfoFilled /></el-icon>
                            </el-tooltip>
                        </div>
                    </template>
                    <div class="chart">
                        <v-charts
                            height="400px"
                            id="loadTemperatureChart"
                            type="line"
                            :option="chartsOption['loadTemperatureChart']"
                            v-if="chartsOption['loadTemperatureChart']"
                            :dataZoom="true"
                        />
                    </div>
                </el-card>
            </el-col>
            <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                <el-card style="overflow: inherit">
                    <template #header>
                        <div :class="mobile ? 'flx-wrap' : 'flex justify-between'">
                            <span class="title">{{ $t('monitor.fanSpeed') }}</span>
                        </div>
                    </template>
                    <div class="chart">
                        <v-charts
                            height="400px"
                            id="loadSpeedChart"
                            type="line"
                            :option="chartsOption['loadSpeedChart']"
                            v-if="chartsOption['loadSpeedChart']"
                            :dataZoom="true"
                        />
                    </div>
                </el-card>
            </el-col>
        </el-row>

        <LayoutContent :title="$t('aiTools.gpu.gpu')" :divider="true" v-else>
            <template #main>
                <div class="app-warn">
                    <div class="flx-center">
                        <span>{{ $t('aiTools.gpu.gpuHelper') }}</span>
                    </div>
                    <div>
                        <img src="@/assets/images/no_app.svg" />
                    </div>
                </div>
            </template>
        </LayoutContent>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue';
import { loadGPUMonitor } from '@/api/modules/host';
import { dateFormatWithoutYear } from '@/utils/util';
import { GlobalStore } from '@/store';
import { shortcuts } from '@/utils/shortcuts';
import { Host } from '@/api/interface/host';
import i18n from '@/lang';

const globalStore = GlobalStore();

const mobile = computed(() => {
    return globalStore.isMobile();
});

const loading = ref(false);
const options = ref([]);
const timeRangeGlobal = ref<[Date, Date]>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
const chartsOption = ref({
    loadPowerChart: null,
    loadGPUChart: null,
    loadMemoryChart: null,
    loadTemperatureChart: null,
    loadSpeedChart: null,
});

const searchTime = ref();
const searchInfo = reactive<Host.MonitorGPUSearch>({
    productName: '',
    startTime: new Date(new Date().setHours(0, 0, 0, 0)),
    endTime: new Date(),
});

const search = async () => {
    if (searchTime.value && searchTime.value.length === 2) {
        searchInfo.startTime = searchTime.value[0];
        searchInfo.endTime = searchTime.value[1];
    }
    loading.value = true;
    await loadGPUMonitor(searchInfo)
        .then((res) => {
            loading.value = false;
            options.value = res.data.productNames || [];
            searchInfo.productName = searchInfo.productName || (options.value.length > 0 ? options.value[0] : '');
            let baseDate = res.data.date.length === 0 ? loadEmptyDate(timeRangeGlobal.value) : res.data.date;
            let date = baseDate.map(function (item: any) {
                return dateFormatWithoutYear(item);
            });
            initCPUCharts(date, res.data.gpuValue);
            initMemoryCharts(date, res.data.memoryValue);
            initPowerCharts(date, res.data.powerValue);
            initSpeedCharts(date, res.data.speedValue);
            initTemperatureCharts(date, res.data.temperatureValue);
        })
        .catch(() => {
            loading.value = false;
        });
};

function initCPUCharts(baseDate: any, items: any) {
    let percents = items.map(function (item: any) {
        return Number(item.toFixed(2));
    });
    let data = percents.length === 0 ? loadEmptyData() : percents;
    chartsOption.value['loadGPUChart'] = {
        xData: baseDate,
        yData: [
            {
                name: i18n.global.t('monitor.gpuUtil'),
                data: data,
            },
        ],
        formatStr: '%',
    };
}
function initMemoryCharts(baseDate: any, items: any) {
    let lists = items.map(function (item: any) {
        return { value: Number(item.percent.toFixed(2)), data: item };
    });
    lists = lists.length === 0 ? loadEmptyData2() : lists;
    chartsOption.value['loadMemoryChart'] = {
        xData: baseDate,
        yData: [
            {
                name: i18n.global.t('monitor.memoryUsage'),
                data: lists,
            },
        ],
        tooltip: {
            trigger: 'axis',
            formatter: function (list: any) {
                return withMemoryProcess(list);
            },
        },
        formatStr: '%',
    };
}
function initPowerCharts(baseDate: any, items: any) {
    let list = items.map(function (item: any) {
        return { value: Number(item.percent.toFixed(2)), data: item };
    });
    list = list.length === 0 ? loadEmptyData2() : list;
    chartsOption.value['loadPowerChart'] = {
        xData: baseDate,
        yData: [
            {
                name: i18n.global.t('monitor.powerUsage'),
                data: list,
            },
        ],
        tooltip: {
            trigger: 'axis',
            formatter: function (list: any) {
                let res = loadDate(list[0].name);
                for (const item of list) {
                    res += loadSeries(item, item.data.value ? item.data.value : item.data, '%');
                    res += `( ${item.data?.data.used}  W / ${item.data?.data.total}  W)<br/>`;
                }
                return res;
            },
        },
        formatStr: '%',
    };
}
function initTemperatureCharts(baseDate: any, items: any) {
    let temperatures = items.map(function (item: any) {
        return Number(item);
    });
    temperatures = temperatures.length === 0 ? loadEmptyData() : temperatures;
    chartsOption.value['loadTemperatureChart'] = {
        xData: baseDate,
        yData: [
            {
                name: i18n.global.t('monitor.temperature'),
                data: temperatures,
            },
        ],
        formatStr: '°C',
    };
}
function initSpeedCharts(baseDate: any, items: any) {
    let speeds = items.map(function (item: any) {
        return Number(item);
    });
    speeds = speeds.length === 0 ? loadEmptyData() : speeds;
    chartsOption.value['loadSpeedChart'] = {
        xData: baseDate,
        yData: [
            {
                name: i18n.global.t('monitor.fanSpeed'),
                data: speeds,
            },
        ],
        formatStr: '%',
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
function loadEmptyData2() {
    return [
        { value: 0, data: {} },
        { value: 0, data: {} },
    ];
}

function withMemoryProcess(list: any) {
    let process;
    let res = loadDate(list[0].name);
    for (const item of list) {
        if (item.data?.data) {
            process = item.data?.data.gpuProcesses || [];
        }
        res += loadSeries(item, item.data.value ? item.data.value : item.data, '%');
        res += `( ${item.data?.data.used}  MiB / ${item.data?.data.total}  MiB)<br/>`;
    }
    if (!process) {
        return res;
    }
    res += `
        <div style="margin-top: 10px; border-bottom: 1px dashed black;"></div>
        <table style="border-collapse: collapse; margin-top: 20px; font-size: 12px;">
        <thead>
            <tr>
            <th style="padding: 6px 8px;">PID</th>
            <th style="padding: 6px 8px;">${i18n.global.t('aiTools.gpu.type')}</th>
            <th style="padding: 6px 8px;">${i18n.global.t('aiTools.gpu.processName')}</th>
            <th style="padding: 6px 8px;">${i18n.global.t('aiTools.gpu.processMemoryUsage')}</th>
            </tr>
        </thead>
        <tbody>
    `;
    for (const row of process) {
        res += `
            <tr>
                <td style="padding: 6px 8px; text-align: center;">
                    ${row.pid}
                </td>
                <td style="padding: 6px 8px; text-align: center;">
                    ${loadProcessType(row.type)}
                </td>
                <td style="padding: 6px 8px; text-align: center;">
                    ${row.processName}
                </td>
                <td style="padding: 6px 8px; text-align: center;">
                    ${row.usedMemory}
                </td>
            </tr>
        `;
    }
    return res;
}
function loadDate(name: any) {
    return ` <div style="display: inline-block; width: 100%; padding-bottom: 10px;">
                ${i18n.global.t('commons.search.date')}: ${name.replaceAll('\n', ' ')}
            </div>`;
}
function loadSeries(item: any, data: any, unit: any) {
    return `<div style="width: 100%;">
                ${item.marker} ${item.seriesName}: ${data} ${unit}
            </div>`;
}
const loadProcessType = (val: string) => {
    if (val === 'C' || val === 'G') {
        return i18n.global.t('aiTools.gpu.type' + val);
    }
    if (val === 'C+G') {
        return i18n.global.t('aiTools.gpu.typeCG');
    }
    return val;
};

onMounted(() => {
    search();
});
</script>

<style scoped lang="scss">
.content-container__search {
    margin-top: 7px;
    .el-card {
        --el-card-padding: 12px;
    }
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
