<template>
    <div v-loading="loading">
        <RouterMenu />
        <template v-if="supported">
            <el-card class="history-toolbar">
                <div class="flex flex-wrap gap-3 items-center">
                    <el-date-picker
                        v-model="timeRangeGlobal"
                        type="datetimerange"
                        style="max-width: 100%; width: 360px; flex-grow: 0"
                        range-separator="-"
                        :start-placeholder="$t('commons.search.timeStart')"
                        :end-placeholder="$t('commons.search.timeEnd')"
                        :shortcuts="shortcuts"
                        :clearable="false"
                        :size="isMobile ? 'small' : 'default'"
                        @change="changeTimeRange"
                    />
                    <el-select style="max-width: 100%" class="p-w-300" v-model="selectedDevice" @change="search">
                        <el-option v-for="item in options" :key="item.value" :label="item.label" :value="item.value" />
                    </el-select>
                    <el-radio-group v-if="history && history.bucketSeconds > 0" v-model="aggregation" @change="search">
                        <el-radio-button value="avg">{{ $t('aiTools.gpu.historyAverage') }}</el-radio-button>
                        <el-radio-button value="max">{{ $t('aiTools.gpu.historyPeak') }}</el-radio-button>
                    </el-radio-group>
                    <TableRefresh @search="refresh" />
                </div>
                <div v-if="history?.sampleCount" class="history-summary">
                    {{
                        history.bucketSeconds > 0
                            ? $t('aiTools.gpu.historyAggregated', {
                                  seconds: history.bucketSeconds,
                                  count: history.sampleCount,
                              })
                            : $t('aiTools.gpu.historyRaw', { count: history.sampleCount })
                    }}
                </div>
            </el-card>
            <el-alert
                v-if="currentDevice?.legacy"
                :title="$t('aiTools.gpu.legacyHistory')"
                type="info"
                :closable="false"
                show-icon
            />
            <el-alert
                v-if="failed"
                :title="$t('aiTools.gpu.historyLoadFailed')"
                type="error"
                :closable="false"
                show-icon
            />
            <el-empty v-else-if="!loading && !history?.sampleCount" :description="$t('commons.msg.noneData')" />
            <el-empty
                v-else-if="history?.sampleCount && !charts.length"
                :description="$t('aiTools.gpu.historyMetricEmpty')"
            />
            <el-row v-else-if="history?.sampleCount" :gutter="12">
                <el-col v-for="chart in charts" :key="chart.id" :xs="24" :md="12">
                    <el-card class="card-interval">
                        <template #header>{{ chart.title }}</template>
                        <v-charts :id="chart.id" height="320px" type="line" :option="chart.option" :dataZoom="true" />
                    </el-card>
                </el-col>
            </el-row>
        </template>
        <LayoutContent v-else-if="!loading" :title="$t('aiTools.gpu.gpu')" :divider="true">
            <template #rightToolBar>
                <TableRefresh @search="refresh" />
            </template>
            <template #main>
                <el-alert
                    v-if="failed"
                    :title="$t('aiTools.gpu.historyLoadFailed')"
                    type="error"
                    :closable="false"
                    show-icon
                />
                <div v-else class="app-warn">
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
import { ref, computed, onMounted, onBeforeUnmount } from 'vue';
import { loadGPUMonitor, getGPUOptions } from '@/api/modules/ai';
import RouterMenu from '@/views/ai/gpu/index.vue';
import { shortcuts } from '@/utils/shortcuts';
import i18n from '@/lang';
import { AI } from '@/api/interface/ai';
import { useGlobalStore } from '@/composables/useGlobalStore';

const { isMobile } = useGlobalStore();
const loading = ref(false);
const failed = ref(false);
const supported = ref(false);
const options = ref<(AI.ChartHide & { value: string; label: string })[]>([]);
const selectedDevice = ref('');
const aggregation = ref<'avg' | 'max'>('avg');
const history = ref<AI.MonitorGPUData>();
const currentDevice = computed(() => options.value.find((item) => item.value === selectedDevice.value));
const timeRangeGlobal = ref<[Date, Date]>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
let followNow = true;
let requestID = 0;

type MetricKey = Exclude<keyof AI.MonitorGPUData, 'date' | 'gpuProcesses' | 'sampleCount' | 'bucketSeconds'>;
type Series = { key: MetricKey; label: string; unit: string; percent?: boolean };
const charts = computed(() => {
    const data = history.value;
    if (!data) return [];
    const definitions: { id: string; title: string; series: Series[] }[] = [
        {
            id: 'gpu-util',
            title: currentDevice.value?.type === 'npu' ? 'aiCore' : 'gpuUtil',
            series: [{ key: 'gpuValue', label: currentDevice.value?.type === 'npu' ? 'aiCore' : 'gpuUtil', unit: '%' }],
        },
        {
            id: 'gpu-memory',
            title: 'memory',
            series: [
                { key: 'memoryUsed', label: 'memoryUsed', unit: 'MiB' },
                { key: 'memoryTotal', label: 'memoryTotal', unit: 'MiB' },
                { key: 'memoryPercent', label: 'percent', unit: '%', percent: true },
            ],
        },
        {
            id: 'gpu-power',
            title: 'powerUsage',
            series: [
                { key: 'powerUsed', label: 'powerCurrent', unit: 'W' },
                { key: 'powerTotal', label: 'powerLimit', unit: 'W' },
                { key: 'powerPercent', label: 'percent', unit: '%', percent: true },
            ],
        },
        {
            id: 'gpu-temperature',
            title: 'temperature',
            series: [
                { key: 'temperatureValue', label: 'temperature', unit: '°C' },
                { key: 'memoryTemperatureValue', label: 'memoryTemperature', unit: '°C' },
                { key: 'hotspotTemperature', label: 'hotspotTemperature', unit: '°C' },
            ],
        },
        {
            id: 'gpu-frequency',
            title: 'frequency',
            series: [
                { key: 'frequencyValue', label: 'frequency', unit: 'MHz' },
                { key: 'memoryFrequencyValue', label: 'memoryFrequency', unit: 'MHz' },
                { key: 'mediaFrequency', label: 'mediaFrequency', unit: 'MHz' },
            ],
        },
        { id: 'gpu-fan', title: 'fanSpeed', series: [{ key: 'speedValue', label: 'fanSpeed', unit: '%' }] },
        {
            id: 'gpu-process',
            title: 'processCount',
            series: [{ key: 'processCount', label: 'processCount', unit: '' }],
        },
        {
            id: 'memory-activity',
            title: 'memoryActivity',
            series: [{ key: 'memoryActivity', label: 'memoryActivity', unit: '%' }],
        },
        {
            id: 'engine-util',
            title: 'engineUtil',
            series: [
                { key: 'encoderUtil', label: 'encoderUtil', unit: '%' },
                { key: 'decoderUtil', label: 'decoderUtil', unit: '%' },
                { key: 'jpegUtil', label: 'jpegUtil', unit: '%' },
                { key: 'ofaUtil', label: 'ofaUtil', unit: '%' },
                { key: 'mediaUtil', label: 'mediaUtil', unit: '%' },
                { key: 'computeUtil', label: 'computeUtil', unit: '%' },
                { key: 'copyUtil', label: 'copyUtil', unit: '%' },
            ],
        },
        {
            id: 'memory-bandwidth',
            title: 'memoryBandwidth',
            series: [
                { key: 'memoryBandwidth', label: 'memoryBandwidth', unit: '%' },
                { key: 'ddrBandwidth', label: 'ddrBandwidth', unit: '%' },
                { key: 'hbmBandwidth', label: 'hbmBandwidth', unit: '%' },
            ],
        },
        {
            id: 'npu-cpu',
            title: 'cpuUtil',
            series: [
                { key: 'aiCPUUtil', label: 'aiCPUUtil', unit: '%' },
                { key: 'ctrlCPUUtil', label: 'ctrlCPUUtil', unit: '%' },
            ],
        },
        {
            id: 'npu-ddr',
            title: 'ddrUsage',
            series: [
                { key: 'ddrUsed', label: 'ddrUsed', unit: 'MiB' },
                { key: 'ddrTotal', label: 'ddrTotal', unit: 'MiB' },
            ],
        },
        {
            id: 'npu-hbm',
            title: 'hbmUsage',
            series: [
                { key: 'hbmUsed', label: 'hbmUsed', unit: 'MiB' },
                { key: 'hbmTotal', label: 'hbmTotal', unit: 'MiB' },
            ],
        },
        {
            id: 'npu-hugepages',
            title: 'hugepagesUsage',
            series: [
                { key: 'hugepagesUsed', label: 'hugepagesUsed', unit: 'pages' },
                { key: 'hugepagesTotal', label: 'hugepagesTotal', unit: 'pages' },
            ],
        },
        { id: 'fan-rpm', title: 'fanRPM', series: [{ key: 'fanRPM', label: 'fanRPM', unit: 'RPM' }] },
    ];
    const dates = (data.date || []).map((date) => new Date(date).getTime());
    return definitions.flatMap((chart) => {
        const series = chart.series.filter((item) => data[item.key]?.some((value) => value != null));
        if (!series.length) return [];
        const hasPercent = series.some((item) => item.percent) && series.some((item) => !item.percent);
        const primaryUnit = (series.find((item) => !item.percent) || series[0]).unit;
        return [
            {
                id: chart.id,
                title: i18n.global.t('aiTools.gpu.' + chart.title),
                option: {
                    xData: dates,
                    xAxis: {
                        type: 'time',
                        splitNumber: isMobile.value ? 3 : 6,
                        axisLabel: { hideOverlap: true },
                        min: timeRangeGlobal.value[0].getTime(),
                        max: timeRangeGlobal.value[1].getTime(),
                    },
                    yData: series.map((item, index) => ({
                        itemStyle: {
                            color: ['#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#8b5cf6', '#0891b2', '#db2777'][
                                index % 7
                            ],
                        },
                        areaStyle: { opacity: 0 },
                        name: i18n.global.t('aiTools.gpu.' + item.label),
                        data: dates.map((date, index) => [date, data[item.key]?.[index] ?? null]),
                        yAxisIndex: item.percent && hasPercent ? 1 : 0,
                        showSymbol: true,
                        symbolSize: 3,
                    })),
                    yAxis: [
                        {
                            type: 'value',
                            name: primaryUnit,
                            ...(primaryUnit === '°C' ? {} : { min: 0 }),
                            ...(primaryUnit === '%' ? { max: 100 } : {}),
                        },
                        ...(hasPercent ? [{ type: 'value', name: '%', position: 'right', min: 0 }] : []),
                    ],
                    grid: { left: isMobile.value ? 55 : 65, right: hasPercent ? 55 : 25, bottom: 80, top: 55 },
                    legend: { type: 'scroll', top: 0, bottom: 'auto' },
                    tooltip: {
                        trigger: 'axis',
                        formatter: (items: any[]) => {
                            const tooltip = document.createElement('div');
                            const title = document.createElement('div');
                            title.textContent = items.length ? new Date(items[0].value[0]).toLocaleString() : '';
                            tooltip.appendChild(title);
                            for (const item of items) {
                                const line = document.createElement('div');
                                const value = item.value[1];
                                line.textContent = `${item.seriesName}: ${value == null ? 'N/A' : Number(value.toFixed(2))} ${series[item.seriesIndex]?.unit || ''}`;
                                tooltip.appendChild(line);
                            }
                            if (chart.id === 'gpu-process' && data.bucketSeconds === 0 && items.length) {
                                const processes = data.gpuProcesses?.[items[0].dataIndex];
                                if (processes?.length) {
                                    const typeTitle = i18n.global.t(
                                        currentDevice.value?.type === 'xpu' ? 'aiTools.gpu.shr' : 'aiTools.gpu.type',
                                    );
                                    appendProcessTable(tooltip, processes, typeTitle);
                                }
                            }
                            return tooltip;
                        },
                    },
                },
            },
        ];
    });
});

const search = async () => {
    const id = ++requestID;
    history.value = undefined;
    failed.value = false;
    const device = currentDevice.value;
    if (!device || !timeRangeGlobal.value) {
        loading.value = false;
        return;
    }
    loading.value = true;
    try {
        const response = await loadGPUMonitor({
            deviceID: device.deviceID,
            productName: device.productName,
            legacy: device.legacy,
            startTime: timeRangeGlobal.value[0],
            endTime: timeRangeGlobal.value[1],
            aggregation: aggregation.value,
        });
        if (id === requestID) history.value = response.data;
    } catch {
        if (id === requestID) failed.value = true;
    } finally {
        if (id === requestID) loading.value = false;
    }
};
const changeTimeRange = () => {
    followNow = false;
    void search();
};
const loadOptions = async () => {
    const id = ++requestID;
    history.value = undefined;
    loading.value = true;
    failed.value = false;
    try {
        const response = await getGPUOptions();
        if (id !== requestID) return;
        supported.value = response.data.supported;
        options.value = (supported.value ? response.data.chartHide || [] : []).map((item) => ({
            ...item,
            value: item.deviceID || `legacy:${item.productName}`,
            label: `${item.productName} · ${item.deviceID || i18n.global.t('aiTools.gpu.legacyDevice')}`,
        }));
        if (!currentDevice.value) selectedDevice.value = options.value[0]?.value || '';
        await search();
    } catch {
        if (id === requestID) failed.value = true;
    } finally {
        if (id === requestID) loading.value = false;
    }
};
const refresh = () => {
    if (followNow) timeRangeGlobal.value = [timeRangeGlobal.value[0], new Date()];
    void loadOptions();
};

function appendProcessTable(tooltip: HTMLElement, process: AI.GPUProcess[], typeTitle: string) {
    const separator = document.createElement('div');
    separator.style.marginTop = '10px';
    separator.style.borderBottom = '1px dashed black';
    tooltip.appendChild(separator);

    const table = document.createElement('table');
    table.style.borderCollapse = 'collapse';
    table.style.marginTop = '20px';
    table.style.fontSize = '12px';
    const header = table.createTHead().insertRow();
    for (const title of [
        'PID',
        i18n.global.t('aiTools.gpu.processName'),
        typeTitle,
        i18n.global.t('aiTools.gpu.memoryUsed'),
    ]) {
        const cell = document.createElement('th');
        cell.style.padding = '6px 8px';
        cell.textContent = title;
        header.appendChild(cell);
    }

    const body = table.createTBody();
    for (const row of process) {
        const tableRow = body.insertRow();
        for (const value of [row.pid, row.processName, loadProcessType(row.type), row.usedMemory]) {
            const cell = tableRow.insertCell();
            cell.style.padding = '6px 8px';
            cell.style.textAlign = 'center';
            cell.textContent = value || '';
        }
    }
    tooltip.appendChild(table);
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

onMounted(loadOptions);
onBeforeUnmount(() => {
    requestID++;
});
</script>

<style scoped lang="scss">
.history-toolbar {
    margin: 7px 0 12px;
    --el-card-padding: 12px;
}
.history-summary {
    margin-top: 12px;
    color: var(--el-text-color-secondary);
    font-size: 13px;
}
</style>
