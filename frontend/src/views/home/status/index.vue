<template>
    <el-row :gutter="10">
        <el-col :span="6" align="center">
            <el-popover placement="bottom" :width="300" trigger="hover">
                <div>
                    <el-tooltip
                        effect="dark"
                        :content="baseInfo.cpuModelName"
                        v-if="baseInfo.cpuModelName.length > 40"
                        placement="top"
                    >
                        <el-tag>
                            {{ baseInfo.cpuModelName.substring(0, 40) + '...' }}
                        </el-tag>
                    </el-tooltip>
                    <el-tag v-else>
                        {{ baseInfo.cpuModelName }}
                    </el-tag>
                </div>
                <el-tag class="tagClass">
                    {{ $t('home.core') }} *{{ baseInfo.cpuCores }} {{ $t('home.logicCore') }} *{{
                        baseInfo.cpuLogicalCores
                    }}
                </el-tag>
                <br />
                <el-row :gutter="5">
                    <el-col :span="12" v-for="(item, index) of currentInfo.cpuPercent" :key="index">
                        <el-tag class="tagClass">CPU-{{ index }}: {{ formatNumber(item) }}%</el-tag>
                    </el-col>
                </el-row>
                <template #reference>
                    <div id="cpu" class="chartClass"></div>
                </template>
            </el-popover>
            <span class="input-help">
                ( {{ formatNumber(currentInfo.cpuUsed) }} / {{ currentInfo.cpuTotal }} ) Core
            </span>
        </el-col>
        <el-col :span="6" align="center">
            <div id="memory" class="chartClass"></div>
            <span class="input-help">
                ( {{ formatNumber(currentInfo.memoryUsed / 1024 / 1024) }} /
                {{ formatNumber(currentInfo.memoryTotal / 1024 / 1024) }} ) MB
            </span>
        </el-col>
        <el-col :span="6" align="center">
            <el-popover placement="bottom" :width="200" trigger="hover">
                <el-tag class="tagClass">
                    {{ $t('home.loadAverage', [1]) }}: {{ formatNumber(currentInfo.load1) }}
                </el-tag>
                <el-tag class="tagClass">
                    {{ $t('home.loadAverage', [5]) }}: {{ formatNumber(currentInfo.load5) }}
                </el-tag>
                <el-tag class="tagClass">
                    {{ $t('home.loadAverage', [15]) }}: {{ formatNumber(currentInfo.load15) }}
                </el-tag>
                <template #reference>
                    <div id="load" class="chartClass"></div>
                </template>
            </el-popover>
            <span class="input-help">{{ loadStatus(currentInfo.loadUsagePercent) }}</span>
        </el-col>
        <el-col :span="6" align="center" v-for="(item, index) of currentInfo.diskData" :key="index">
            <el-popover placement="bottom" :width="300" trigger="hover">
                <el-row :gutter="5">
                    <el-col :span="12">
                        <el-tag style="font-weight: 500">{{ $t('home.baseInfo') }}:</el-tag>
                        <el-tag class="tagClass">{{ $t('home.mount') }}: {{ item.path }}</el-tag>
                        <el-tag class="tagClass">{{ $t('commons.table.type') }}: {{ item.type }}</el-tag>
                        <el-tag class="tagClass">{{ $t('home.fileSystem') }}: {{ item.device }}</el-tag>
                        <div><el-tag class="tagClass" style="font-weight: 500">Inode:</el-tag></div>
                        <el-tag class="tagClass">{{ $t('home.total') }}: {{ item.inodesTotal }}</el-tag>
                        <el-tag class="tagClass">{{ $t('home.used') }}: {{ item.inodesUsed }}</el-tag>
                        <el-tag class="tagClass">{{ $t('home.free') }}: {{ item.inodesFree }}</el-tag>
                        <el-tag class="tagClass">
                            {{ $t('home.percent') }}: {{ formatNumber(item.inodesUsedPercent) }}%
                        </el-tag>
                    </el-col>

                    <el-col :span="12">
                        <div>
                            <el-tag style="margin-top: 108px; font-weight: 500">{{ $t('monitor.disk') }}:</el-tag>
                        </div>
                        <el-tag class="tagClass">
                            {{ $t('home.total') }}: {{ formatNumber(item.total / 1024 / 1024 / 1024) }} GB
                        </el-tag>
                        <el-tag class="tagClass">
                            {{ $t('home.used') }}: {{ formatNumber(item.used / 1024 / 1024 / 1024) }} GB
                        </el-tag>
                        <el-tag class="tagClass">
                            {{ $t('home.free') }}: {{ formatNumber(item.free / 1024 / 1024 / 1024) }} GB
                        </el-tag>
                        <el-tag class="tagClass">
                            {{ $t('home.percent') }}: {{ formatNumber(item.usedPercent) }}%
                        </el-tag>
                    </el-col>
                </el-row>
                <template #reference>
                    <div :id="`disk${index}`" class="chartClass"></div>
                </template>
            </el-popover>
            <span class="input-help">
                ( {{ formatNumber(item.used / 1024 / 1024 / 1024) }} /
                {{ formatNumber(item.total / 1024 / 1024 / 1024) }} ) GB
            </span>
        </el-col>
    </el-row>
</template>

<script setup lang="ts">
import { Dashboard } from '@/api/interface/dashboard';
import i18n from '@/lang';
import * as echarts from 'echarts';
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const baseInfo = ref<Dashboard.BaseInfo>({
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
    ioCount: 0,
    ioReadTime: 0,
    ioWriteTime: 0,

    diskData: [],

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
    nextTick(() => {
        for (let i = 0; i < currentInfo.value.diskData.length; i++) {
            freshChart(
                'disk' + i,
                currentInfo.value.diskData[i].path,
                formatNumber(currentInfo.value.diskData[i].usedPercent),
            );
        }
    });
};

const freshChart = (chartName: string, Title: string, Data: number) => {
    let myChart = echarts.getInstanceByDom(document.getElementById(chartName) as HTMLElement);
    if (myChart === null || myChart === undefined) {
        myChart = echarts.init(document.getElementById(chartName) as HTMLElement);
    }
    const theme = globalStore.$state.themeConfig.theme || 'light';
    let percentText = String(Data).split('.');
    const option = {
        title: [
            {
                text: `{a|${percentText[0]}.}{b|${percentText[1] || 0} %}`,
                textStyle: {
                    rich: {
                        a: {
                            fontSize: '22',
                        },
                        b: {
                            fontSize: '14',
                            padding: [5, 0, 0, 0],
                        },
                    },

                    color: theme === 'dark' ? '#ffffff' : '#0f0f0f',
                    lineHeight: 25,
                    // fontSize: 20,
                    fontWeight: 500,
                },
                left: '49%',
                top: '32%',
                subtext: Title,
                subtextStyle: {
                    color: theme === 'dark' ? '#BBBFC4' : '#646A73',
                    fontSize: 13,
                },
                textAlign: 'center',
            },
        ],
        polar: {
            radius: ['71%', '80%'],
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
                backgroundStyle: {
                    color: theme === 'dark' ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 94, 235, 0.05)',
                },
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
                ],
                label: {
                    show: false,
                },
                data: [Data],
            },
            {
                type: 'pie',
                radius: ['0%', '60%'],
                center: ['50%', '50%'],
                label: {
                    show: false,
                },
                color: theme === 'dark' ? '#16191D' : '#fff',
                data: [
                    {
                        value: 0,
                        itemStyle: {
                            shadowColor: theme === 'dark' ? '#16191D' : 'rgba(0, 94, 235, 0.1)',
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

<style scoped lang="scss">
.tagClass {
    margin-top: 3px;
}
.chartClass {
    width: 100%;
    height: 160px;
}
</style>
