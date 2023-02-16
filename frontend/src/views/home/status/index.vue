<template>
    <el-row :gutter="10">
        <el-col :span="6" align="center">
            <el-popover placement="bottom" :width="300" trigger="hover">
                <div style="margin-bottom: 10px">
                    <el-tag>{{ baseInfo.cpuModelName }}</el-tag>
                </div>
                <el-tag>
                    {{ $t('home.core') }} *{{ baseInfo.cpuCores }}； {{ $t('home.logicCore') }} *{{
                        baseInfo.cpuLogicalCores
                    }}
                </el-tag>
                <br />
                <el-tag style="margin-top: 5px" v-for="(item, index) of currentInfo.cpuPercent" :key="index">
                    CPU-{{ index }}: {{ formatNumber(item) }}%
                </el-tag>
                <template #reference>
                    <div id="cpu" style="width: 100%; height: 160px"></div>
                </template>
            </el-popover>
            <span class="input-help" style="margin-top: -10px">
                ( {{ formatNumber(currentInfo.cpuUsed) }} / {{ currentInfo.cpuTotal }} ) Core
            </span>
        </el-col>
        <el-col :span="6" align="center">
            <div id="memory" style="width: 100%; height: 160px"></div>
            <span class="input-help">
                ( {{ formatNumber(currentInfo.memoryUsed / 1024 / 1024) }} /
                {{ formatNumber(currentInfo.memoryTotal / 1024 / 1024) }} ) MB
            </span>
        </el-col>
        <el-col :span="6" align="center">
            <el-popover placement="bottom" :width="200" trigger="hover">
                <el-tag style="margin-top: 5px">
                    {{ $t('home.loadAverage', [1]) }}: {{ formatNumber(currentInfo.load1) }}
                </el-tag>
                <el-tag style="margin-top: 5px">
                    {{ $t('home.loadAverage', [5]) }}: {{ formatNumber(currentInfo.load5) }}
                </el-tag>
                <el-tag style="margin-top: 5px">
                    {{ $t('home.loadAverage', [15]) }}: {{ formatNumber(currentInfo.load15) }}
                </el-tag>
                <template #reference>
                    <div id="load" style="width: 100%; height: 160px"></div>
                </template>
            </el-popover>
            <span class="input-help">{{ loadStatus(currentInfo.loadUsagePercent) }}</span>
        </el-col>
        <el-col :span="6" align="center">
            <el-popover placement="bottom" :width="260" trigger="hover">
                <el-row :gutter="5">
                    <el-col :span="12">
                        <el-tag>{{ $t('home.mount') }}: /</el-tag>
                        <div><el-tag style="margin-top: 10px">iNode</el-tag></div>
                        <el-tag style="margin-top: 5px">{{ $t('home.total') }}: {{ currentInfo.inodesTotal }}</el-tag>
                        <el-tag style="margin-top: 3px">{{ $t('home.used') }}: {{ currentInfo.inodesUsed }}</el-tag>
                        <el-tag style="margin-top: 3px">{{ $t('home.free') }}: {{ currentInfo.inodesFree }}</el-tag>
                        <el-tag style="margin-top: 3px">
                            {{ $t('home.percent') }}: {{ formatNumber(currentInfo.inodesUsedPercent) }}%
                        </el-tag>
                    </el-col>

                    <el-col :span="12">
                        <div>
                            <el-tag style="margin-top: 35px">{{ $t('monitor.disk') }}</el-tag>
                        </div>
                        <el-tag style="margin-top: 5px">
                            {{ $t('home.total') }}: {{ formatNumber(currentInfo.total / 1024 / 1024 / 1024) }} GB
                        </el-tag>
                        <el-tag style="margin-top: 3px">
                            {{ $t('home.used') }}: {{ formatNumber(currentInfo.used / 1024 / 1024 / 1024) }} GB
                        </el-tag>
                        <el-tag style="margin-top: 3px">
                            {{ $t('home.free') }}: {{ formatNumber(currentInfo.free / 1024 / 1024 / 1024) }} GB
                        </el-tag>
                        <el-tag style="margin-top: 3px">
                            {{ $t('home.percent') }}: {{ formatNumber(currentInfo.usedPercent) }}%
                        </el-tag>
                    </el-col>
                </el-row>
                <template #reference>
                    <div id="disk" style="width: 100%; height: 160px"></div>
                </template>
            </el-popover>
            <span class="input-help">
                ( {{ formatNumber(currentInfo.used / 1024 / 1024 / 1024) }} /
                {{ formatNumber(currentInfo.total / 1024 / 1024 / 1024) }} ) GB
            </span>
        </el-col>
    </el-row>
</template>

<script setup lang="ts">
import { Dashboard } from '@/api/interface/dashboard';
import i18n from '@/lang';
import * as echarts from 'echarts';
import { onBeforeUnmount, onMounted, ref } from 'vue';

const baseInfo = ref<Dashboard.BaseInfo>({
    haloID: 0,
    dateeaseID: 0,
    jumpserverID: 0,
    metersphereID: 0,
    kubeoperatorID: 0,
    kubepiID: 0,

    websiteNumber: 0,
    databaseNumber: 0,
    cronjobNumber: 0,
    appInstalldNumber: 0,

    hostname: '',
    os: '',
    platform: '',
    platformFamily: '',
    platformVersion: '',
    kernelArch: '',
    kernelVersion: '',
    virtualizationSystem: '',

    cpuCores: 0,
    cpuLogicalCores: 0,
    cpuModelName: '',
    currentInfo: null,
});
const currentInfo = ref<Dashboard.CurrentInfo>({
    uptime: 0,
    timeSinceUptime: '',
    procs: 0,

    load1: 0,
    load5: 0,
    load15: 0,
    loadUsagePercent: 0,

    cpuPercent: [] as Array<number>,
    cpuUsedPercent: 0,
    cpuUsed: 0,
    cpuTotal: 0,

    memoryTotal: 0,
    memoryAvailable: 0,
    memoryUsed: 0,
    MemoryUsedPercent: 0,

    ioReadBytes: 0,
    ioWriteBytes: 0,
    ioTime: 0,
    ioCount: 0,

    total: 0,
    free: 0,
    used: 0,
    usedPercent: 0,

    inodesTotal: 0,
    inodesUsed: 0,
    inodesFree: 0,
    inodesUsedPercent: 0,

    netBytesSent: 0,
    netBytesRecv: 0,
    shotTime: new Date(),
});

const acceptParams = (current: Dashboard.CurrentInfo, base: Dashboard.BaseInfo): void => {
    currentInfo.value = current;
    baseInfo.value = base;
    freshChart('cpu', 'CPU', formatNumber(currentInfo.value.cpuUsedPercent));
    freshChart('memory', i18n.global.t('monitor.memory'), formatNumber(currentInfo.value.MemoryUsedPercent));
    freshChart('load', i18n.global.t('home.load'), formatNumber(currentInfo.value.loadUsagePercent));
    freshChart('disk', i18n.global.t('monitor.disk'), formatNumber(currentInfo.value.usedPercent));
};

const freshChart = (chartName: string, Title: string, Data: number) => {
    let myChart = echarts.getInstanceByDom(document.getElementById(chartName) as HTMLElement);
    if (myChart === null || myChart === undefined) {
        myChart = echarts.init(document.getElementById(chartName) as HTMLElement);
    }
    const option = {
        title: [
            {
                text: Data + '%',
                textStyle: {
                    color: '#0f0f0f',
                    lineHeight: 25,
                    fontSize: 20,
                    fontWeight: '400',
                },
                left: '49%',
                top: '32%',
                subtext: Title,
                subtextStyle: {
                    color: '#646A73',
                    fontSize: 13,
                },
                textAlign: 'center',
            },
        ],
        polar: {
            radius: ['71%', '82%'],
            center: ['50%', '50%'],
        },
        angleAxis: {
            max: 100,
            show: false,
        },
        radiusAxis: {
            type: 'category',
            show: true,
            axisLabel: {
                show: false,
            },
            axisLine: {
                show: false,
            },
            axisTick: {
                show: false,
            },
        },
        series: [
            {
                type: 'bar',
                roundCap: true,
                barWidth: 30,
                showBackground: true,
                coordinateSystem: 'polar',
                color: [
                    new echarts.graphic.LinearGradient(0, 1, 0, 0, [
                        {
                            offset: 0,
                            color: 'rgba(81, 192, 255, .1)',
                        },
                        {
                            offset: 1,
                            color: '#4261F6',
                        },
                    ]),
                    '#f2f2f2',
                ],
                label: {
                    show: false,
                },
                data: [Data],
            },
            {
                type: 'pie',
                radius: ['0%', '61%'],
                center: ['50%', '50%'],
                label: {
                    show: false,
                },
                color: '#fff',
                data: [
                    {
                        value: 0,
                        itemStyle: {
                            shadowColor: '#e3e3e3',
                            shadowBlur: 5,
                        },
                    },
                ],
            },
        ],
    };
    myChart.setOption(option, true);
};

function loadStatus(val: number) {
    if (val < 30) {
        return i18n.global.t('home.runSmoothly');
    }
    if (val < 70) {
        return i18n.global.t('home.runNormal');
    }
    if (val < 80) {
        return i18n.global.t('home.runSlowly');
    }
    return i18n.global.t('home.runJam');
}

function formatNumber(val: number) {
    return Number(val.toFixed(2));
}

function changeChartSize() {
    echarts.getInstanceByDom(document.getElementById('cpu') as HTMLElement)?.resize();
    echarts.getInstanceByDom(document.getElementById('memory') as HTMLElement)?.resize();
    echarts.getInstanceByDom(document.getElementById('load') as HTMLElement)?.resize();
    echarts.getInstanceByDom(document.getElementById('disk') as HTMLElement)?.resize();
}

onMounted(() => {
    window.addEventListener('resize', changeChartSize);
});

onBeforeUnmount(() => {
    window.removeEventListener('resize', changeChartSize);
});

defineExpose({
    acceptParams,
});
</script>
