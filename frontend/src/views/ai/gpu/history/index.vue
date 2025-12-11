<template>
    <div v-loading="loading">
        <RouterMenu />

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
                    <el-button icon="Setting" class="float-right mr-2" @click="quickJump()" />
                </div>
            </el-card>
        </div>
        <el-row :gutter="7" v-if="options.length !== 0">
            <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                <el-card class="card-interval" style="overflow: inherit">
                    <template #header>
                        <div :class="mobile ? 'flx-wrap' : 'flex justify-between'">
                            <span class="title">{{ $t('aiTools.gpu.memoryUsage') }}</span>
                        </div>
                    </template>
                    <div class="chart" v-if="!currentHide?.memory">
                        <v-charts
                            height="400px"
                            id="loadMemoryChart"
                            type="line"
                            :option="chartsOption['loadMemoryChart']"
                            v-if="chartsOption['loadMemoryChart']"
                            :dataZoom="true"
                        />
                    </div>
                    <el-empty class="chart" v-else :description="$t('aiTools.gpu.notSupport')" />
                </el-card>
            </el-col>
            <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                <el-card class="card-interval" style="overflow: inherit">
                    <template #header>
                        <div :class="mobile ? 'flx-wrap' : 'flex justify-between'">
                            <span class="title">{{ $t('aiTools.gpu.gpuUtil') }}</span>
                        </div>
                    </template>
                    <div class="chart" v-if="!currentHide?.gpu">
                        <v-charts
                            height="400px"
                            id="loadGPUChart"
                            type="line"
                            :option="chartsOption['loadGPUChart']"
                            v-if="chartsOption['loadGPUChart']"
                            :dataZoom="true"
                        />
                    </div>
                    <el-empty class="chart" v-else :description="$t('aiTools.gpu.notSupport')" />
                </el-card>
            </el-col>
            <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                <el-card class="card-interval" style="overflow: inherit">
                    <template #header>
                        <div :class="mobile ? 'flx-wrap' : 'flex justify-between'">
                            <span class="title">{{ $t('aiTools.gpu.process') }}</span>
                        </div>
                    </template>
                    <div class="chart" v-if="!currentHide?.process">
                        <v-charts
                            height="400px"
                            id="loadProcessChart"
                            type="line"
                            :option="chartsOption['loadProcessChart']"
                            v-if="chartsOption['loadProcessChart']"
                            :dataZoom="true"
                        />
                    </div>
                    <el-empty class="chart" v-else :description="$t('aiTools.gpu.notSupport')" />
                </el-card>
            </el-col>
            <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                <el-card class="card-interval" style="overflow: inherit">
                    <template #header>
                        <div :class="mobile ? 'flx-wrap' : 'flex justify-between'">
                            <span class="title">{{ $t('aiTools.gpu.powerUsage') }}</span>
                        </div>
                    </template>
                    <div class="chart" v-if="!currentHide?.power">
                        <v-charts
                            height="400px"
                            id="loadPowerChart"
                            type="line"
                            :option="chartsOption['loadPowerChart']"
                            v-if="chartsOption['loadPowerChart']"
                            :dataZoom="true"
                        />
                    </div>
                    <el-empty class="chart" v-else :description="$t('aiTools.gpu.notSupport')" />
                </el-card>
            </el-col>
            <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                <el-card class="card-interval" style="overflow: inherit">
                    <template #header>
                        <div>
                            {{ $t('aiTools.gpu.temperature') }}
                            <el-tooltip placement="top" :content="$t('aiTools.gpu.temperatureHelper')">
                                <el-icon size="15"><InfoFilled /></el-icon>
                            </el-tooltip>
                        </div>
                    </template>
                    <div class="chart" v-if="!currentHide?.temperature">
                        <v-charts
                            height="400px"
                            id="loadTemperatureChart"
                            type="line"
                            :option="chartsOption['loadTemperatureChart']"
                            v-if="chartsOption['loadTemperatureChart']"
                            :dataZoom="true"
                        />
                    </div>
                    <el-empty class="chart" v-else :description="$t('aiTools.gpu.notSupport')" />
                </el-card>
            </el-col>
            <el-col :xs="24" :sm="24" :md="12" :lg="12" :xl="12">
                <el-card class="card-interval" style="overflow: inherit">
                    <template #header>
                        <div :class="mobile ? 'flx-wrap' : 'flex justify-between'">
                            <span class="title">{{ $t('aiTools.gpu.fanSpeed') }}</span>
                        </div>
                    </template>
                    <div class="chart" v-if="!currentHide?.speed">
                        <v-charts
                            height="400px"
                            id="loadSpeedChart"
                            type="line"
                            :option="chartsOption['loadSpeedChart']"
                            v-if="chartsOption['loadSpeedChart']"
                            :dataZoom="true"
                        />
                    </div>
                    <el-empty class="chart" v-else :description="$t('aiTools.gpu.notSupport')" />
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
import { loadGPUMonitor, getGPUOptions } from '@/api/modules/host';
import { dateFormatWithoutYear } from '@/utils/util';
import RouterMenu from '@/views/ai/gpu/index.vue';
import { GlobalStore } from '@/store';
import { shortcuts } from '@/utils/shortcuts';
import { Host } from '@/api/interface/host';
import i18n from '@/lang';
import { routerToName } from '@/utils/router';

const globalStore = GlobalStore();

const mobile = computed(() => {
    return globalStore.isMobile();
});

const loading = ref(false);
const options = ref([]);
const gpuType = ref('gpu');
const timeRangeGlobal = ref<[Date, Date]>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
const chartsOption = ref({
    loadPowerChart: null,
    loadGPUChart: null,
    loadMemoryChart: null,
    loadTemperatureChart: null,
    loadSpeedChart: null,
});
const chartHide = ref([]);
const currentHide = ref();

const searchInfo = reactive<Host.MonitorGPUSearch>({
    productName: '',
    startTime: new Date(new Date().setHours(0, 0, 0, 0)),
    endTime: new Date(),
});

const loadOptions = async () => {
    loading.value = true;
    await getGPUOptions()
        .then((res) => {
            gpuType.value = res.data.gpuType || 'gpu';
            options.value = res.data.options || [];
            chartHide.value = res.data.chartHide || [];
            searchInfo.productName = options.value.length > 0 ? options.value[0] : '';
            search();
        })
        .catch(() => {
            loading.value = false;
            options.value = [];
        });
};

const loadCurrentHide = () => {
    for (const item of chartHide.value) {
        if (item.productName === searchInfo.productName) {
            currentHide.value = item;
            break;
        }
    }
};

const search = async () => {
    if (timeRangeGlobal.value && timeRangeGlobal.value.length === 2) {
        searchInfo.startTime = timeRangeGlobal.value[0];
        searchInfo.endTime = timeRangeGlobal.value[1];
    }
    loading.value = true;
    loadCurrentHide();
    await loadGPUMonitor(searchInfo)
        .then((res) => {
            loading.value = false;
            let baseDate = res.data.date || [];
            if (baseDate.length === 0) {
                baseDate = loadEmptyDate(timeRangeGlobal.value);
            }
            let date = baseDate.map(function (item: any) {
                return dateFormatWithoutYear(item);
            });
            if (!currentHide.value || !currentHide.value.gpu) {
                initCPUCharts(date, res.data.gpuValue || []);
            }
            if (!currentHide.value || !currentHide.value.memory) {
                initMemoryCharts(date, res.data);
            }
            if (!currentHide.value || !currentHide.value.power) {
                initPowerCharts(date, res.data);
            }
            if (!currentHide.value || !currentHide.value.speed) {
                initSpeedCharts(date, res.data.speedValue || []);
            }
            if (!currentHide.value || !currentHide.value.temperature) {
                initTemperatureCharts(date, res.data.temperatureValue || []);
            }
            if (!currentHide.value || !currentHide.value.process) {
                initProcessCharts(date, res.data);
            }
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
                name: i18n.global.t('aiTools.gpu.gpuUtil'),
                data: data,
            },
        ],
        formatStr: '%',
    };
}
function initMemoryCharts(baseDate: any, data: any) {
    chartsOption.value['loadMemoryChart'] = {
        xData: baseDate,
        yData: [
            {
                name: i18n.global.t('aiTools.gpu.memoryUsed'),
                data: data.memoryUsed,
            },
            {
                name: i18n.global.t('aiTools.gpu.memoryTotal'),
                data: data.memoryTotal,
            },
            {
                name: i18n.global.t('aiTools.gpu.percent'),
                data: data.memoryPercent,
                yAxisIndex: 1,
            },
        ],
        yAxis: [
            { type: 'value', name: i18n.global.t('aiTools.gpu.memory') },
            {
                type: 'value',
                name: i18n.global.t('aiTools.gpu.percent') + ' ( % )',
                position: 'right',
                alignTicks: true,
            },
        ],
        grid: mobile.value ? { left: '15%', right: '15%', bottom: '20%' } : null,
        tooltip: {
            trigger: 'axis',
            formatter: function (list: any) {
                return withMemoryProcess(list);
            },
        },
        formatStr: '%',
    };
}

function initPowerCharts(baseDate: any, data: any) {
    chartsOption.value['loadPowerChart'] = {
        xData: baseDate,
        yData: [
            {
                name: i18n.global.t('aiTools.gpu.powerCurrent'),
                data: data.powerUsed,
            },
            {
                name: i18n.global.t('aiTools.gpu.powerLimit'),
                data: data.powerTotal,
            },
            {
                name: i18n.global.t('aiTools.gpu.percent'),
                data: data.powerPercent,
                yAxisIndex: 1,
            },
        ],
        yAxis: [
            { type: 'value', name: i18n.global.t('aiTools.gpu.power') },
            {
                type: 'value',
                name: i18n.global.t('aiTools.gpu.percent') + ' ( % )',
                position: 'right',
                alignTicks: true,
            },
        ],
        grid: mobile.value ? { left: '15%', right: '15%', bottom: '20%' } : null,
        tooltip: {
            trigger: 'axis',
            formatter: function (list: any) {
                let res = loadDate(list[0].name);
                for (const item of list) {
                    if (
                        item.seriesName === i18n.global.t('aiTools.gpu.powerCurrent') ||
                        item.seriesName === i18n.global.t('aiTools.gpu.powerLimit')
                    ) {
                        res += loadSeries(item, item.data, 'W');
                    } else {
                        res += loadSeries(item, Number(item.data.toFixed(2)), '%');
                    }
                }
                return res;
            },
        },
        formatStr: '%',
    };
}

const quickJump = () => {
    routerToName('HostMonitorSetting');
};

function initTemperatureCharts(baseDate: any, items: any) {
    let temperatures = items.map(function (item: any) {
        return Number(item);
    });
    temperatures = temperatures.length === 0 ? loadEmptyData() : temperatures;
    chartsOption.value['loadTemperatureChart'] = {
        xData: baseDate,
        yData: [
            {
                name: i18n.global.t('aiTools.gpu.temperature'),
                data: temperatures,
            },
        ],
        formatStr: '°C',
    };
}
function initProcessCharts(baseDate: any, data: any) {
    let process = data.processCount.map(function (item: any) {
        return Number(item);
    });
    process = process.length === 0 ? loadEmptyData() : process;
    chartsOption.value['loadProcessChart'] = {
        xData: baseDate,
        yData: [
            {
                name: i18n.global.t('aiTools.gpu.processCount'),
                data: process,
            },
        ],
        tooltip: {
            trigger: 'axis',
            formatter: function (list: any) {
                const param = list[0];
                const index = param.dataIndex;
                let process = data.gpuProcesses?.length > index ? data.gpuProcesses[index] : [];
                return withProcess(list, process);
            },
        },
        formatStr: '-',
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
                name: i18n.global.t('aiTools.gpu.fanSpeed'),
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

function withMemoryProcess(list: any) {
    let res = loadDate(list[0].name);
    for (const item of list) {
        if (
            item.seriesName === i18n.global.t('aiTools.gpu.memoryUsed') ||
            item.seriesName === i18n.global.t('aiTools.gpu.memoryTotal')
        ) {
            res += loadSeries(item, item.data, 'MiB');
        } else {
            res += loadSeries(item, Number(item.data.toFixed(2)), '%');
        }
    }
    return res;
}

function withProcess(list: any, process: any) {
    let res = loadDate(list[0].name);
    for (const item of list) {
        res += loadSeries(item, item.data, '');
    }
    let title = gpuType.value === 'gpu' ? i18n.global.t('aiTools.gpu.type') : i18n.global.t('aiTools.gpu.shr');
    res += `
        <div style="margin-top: 10px; border-bottom: 1px dashed black;"></div>
        <table style="border-collapse: collapse; margin-top: 20px; font-size: 12px;">
        <thead>
            <tr>
            <th style="padding: 6px 8px;">PID</th>
            <th style="padding: 6px 8px;">${i18n.global.t('aiTools.gpu.processName')}</th>
            <th style="padding: 6px 8px;">${title}</th>
            <th style="padding: 6px 8px;">${i18n.global.t('aiTools.gpu.memoryUsed')}</th>
            </tr>
        </thead>
        <tbody>
    `;
    if (!process) {
        return res;
    }
    for (const row of process) {
        res += `
            <tr>
                <td style="padding: 6px 8px; text-align: center;">
                    ${row.pid}
                </td>
                <td style="padding: 6px 8px; text-align: center;">
                    ${row.processName}
                </td>
                <td style="padding: 6px 8px; text-align: center;">
                    ${loadProcessType(row.type)}
                </td>
                <td style="padding: 6px 8px; text-align: center;">
                    ${row.usedMemory.replaceAll('MB', 'MiB')}
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
    loadOptions();
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
