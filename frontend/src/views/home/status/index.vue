<template>
    <el-row :gutter="10">
        <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6" align="center">
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
                ( {{ formatNumber(currentInfo.cpuUsed) }} / {{ currentInfo.cpuTotal }} ) {{ $t('commons.units.core') }}
            </span>
        </el-col>
        <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6" align="center">
            <div id="memory" class="chartClass"></div>
            <span class="input-help">
                ( {{ formatNumber(currentInfo.memoryUsed / 1024 / 1024) }} /
                {{ formatNumber(currentInfo.memoryTotal / 1024 / 1024) }} ) MB
            </span>
        </el-col>
        <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6" align="center">
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
        <el-col
            :xs="12"
            :sm="12"
            :md="6"
            :lg="6"
            :xl="6"
            align="center"
            v-for="(item, index) of currentInfo.diskData"
            :key="index"
            v-show="showMore || index < 4"
        >
            <el-popover placement="bottom" :width="300" trigger="hover">
                <el-row :gutter="5">
                    <el-tag style="font-weight: 500">{{ $t('home.baseInfo') }}:</el-tag>
                </el-row>
                <el-row :gutter="5">
                    <el-tag class="nameTag">{{ $t('home.mount') }}: {{ item.path }}</el-tag>
                </el-row>
                <el-row :gutter="5">
                    <el-tag class="tagClass">{{ $t('commons.table.type') }}: {{ item.type }}</el-tag>
                </el-row>
                <el-row :gutter="5">
                    <el-tag class="tagClass">{{ $t('home.fileSystem') }}: {{ item.device }}</el-tag>
                </el-row>
                <el-row :gutter="5">
                    <el-col :span="12">
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
                            <el-tag style="margin-top: 3px; font-weight: 500">{{ $t('monitor.disk') }}:</el-tag>
                        </div>
                        <el-tag class="tagClass">{{ $t('home.total') }}: {{ computeSize(item.total) }}</el-tag>
                        <el-tag class="tagClass">{{ $t('home.used') }}: {{ computeSize(item.used) }}</el-tag>
                        <el-tag class="tagClass">{{ $t('home.free') }}: {{ computeSize(item.free) }}</el-tag>
                        <el-tag class="tagClass">
                            {{ $t('home.percent') }}: {{ formatNumber(item.usedPercent) }}%
                        </el-tag>
                    </el-col>
                </el-row>
                <template #reference>
                    <div :id="`disk${index}`" class="chartClass"></div>
                </template>
            </el-popover>
            <span class="input-help">{{ computeSize(item.used) }} / {{ computeSize(item.total) }}</span>
        </el-col>
        <el-col v-if="!showMore" :xs="12" :sm="12" :md="6" :lg="6" :xl="6" align="center">
            <el-button link type="primary" @click="showMore = true" class="buttonClass">
                {{ $t('tabs.more') }}
                <el-icon><Bottom /></el-icon>
            </el-button>
        </el-col>
        <el-col
            v-if="showMore && currentInfo.diskData.length > 5"
            :xs="12"
            :sm="12"
            :md="6"
            :lg="6"
            :xl="6"
            align="center"
            style="float: right"
        >
            <el-button type="primary" link @click="showMore = false" class="buttonClass">
                {{ $t('tabs.hide') }}
                <el-icon><Top /></el-icon>
            </el-button>
        </el-col>
    </el-row>
</template>

<script setup lang="ts">
import { Dashboard } from '@/api/interface/dashboard';
import { computeSize } from '@/utils/util';
import i18n from '@/lang';
import * as echarts from 'echarts';
import { nextTick, onBeforeUnmount, onMounted, ref } from 'vue';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const showMore = ref(true);

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

const acceptParams = (current: Dashboard.CurrentInfo, base: Dashboard.BaseInfo, isInit: boolean): void => {
    currentInfo.value = current;
    baseInfo.value = base;
    freshChart('cpu', 'CPU', formatNumber(currentInfo.value.cpuUsedPercent));
    freshChart('memory', i18n.global.t('monitor.memory'), formatNumber(currentInfo.value.MemoryUsedPercent));
    freshChart('load', i18n.global.t('home.load'), formatNumber(currentInfo.value.loadUsagePercent));
    currentInfo.value.diskData = currentInfo.value.diskData || [];
    nextTick(() => {
        for (let i = 0; i < currentInfo.value.diskData.length; i++) {
            let itemPath = currentInfo.value.diskData[i].path;
            itemPath = itemPath.length > 12 ? itemPath.substring(0, 9) + '...' : itemPath;
            freshChart('disk' + i, itemPath, formatNumber(currentInfo.value.diskData[i].usedPercent));
        }
        if (currentInfo.value.diskData.length > 5) {
            showMore.value = isInit ? false : showMore.value || false;
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
    nextTick(function () {
        myChart.setOption(option, true);
    });
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
    for (let i = 0; i < currentInfo.value.diskData.length; i++) {
        echarts.getInstanceByDom(document.getElementById('disk' + i) as HTMLElement)?.resize();
    }
}

function disposeChart() {
    echarts.getInstanceByDom(document.getElementById('cpu') as HTMLElement)?.dispose();
    echarts.getInstanceByDom(document.getElementById('memory') as HTMLElement)?.dispose();
    echarts.getInstanceByDom(document.getElementById('load') as HTMLElement)?.dispose();
    for (let i = 0; i < currentInfo.value.diskData.length; i++) {
        echarts.getInstanceByDom(document.getElementById('disk' + i) as HTMLElement)?.dispose();
    }
}

onMounted(() => {
    window.addEventListener('resize', changeChartSize);
});

onBeforeUnmount(() => {
    disposeChart();
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
.buttonClass {
    margin-top: 28%;
    font-size: 14px;
}
.nameTag {
    margin-top: 3px;
    height: auto;
    display: inline-block;
    white-space: normal;
    line-height: 1.8;
}
</style>
