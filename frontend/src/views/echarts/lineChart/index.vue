<template>
    <div ref="echartsRef" class="content-box"></div>
</template>

<script setup lang="ts" name="columnChart">
import { ref, onMounted } from 'vue';
import * as echarts from 'echarts';
import { useEcharts } from '@/hooks/useEcharts';
const echartsRef = ref<HTMLElement>();
onMounted(() => {
    let myChart: echarts.ECharts = echarts.init(
        echartsRef.value as HTMLElement,
    );
    let option: echarts.EChartsOption = {
        title: {
            text: 'Stacked Area Chart',
            textStyle: {
                color: '#a1a1a1',
            },
        },
        tooltip: {
            trigger: 'axis',
            axisPointer: {
                type: 'cross',
                label: {
                    backgroundColor: '#6a7985',
                },
            },
        },
        legend: {
            data: [
                'Email',
                'Union Ads',
                'Video Ads',
                'Direct',
                'Search Engine',
            ],
            textStyle: {
                color: '#a1a1a1',
            },
        },
        toolbox: {
            feature: {
                saveAsImage: {},
            },
        },
        grid: {
            left: '3%',
            right: '4%',
            bottom: '3%',
            containLabel: true,
        },
        xAxis: [
            {
                type: 'category',
                boundaryGap: false,
                data: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
                axisLabel: {
                    color: '#a1a1a1',
                },
            },
        ],
        yAxis: [
            {
                type: 'value',
                axisLabel: {
                    color: '#a1a1a1',
                },
            },
        ],
        series: [
            {
                name: 'Email',
                type: 'line',
                stack: 'Total',
                areaStyle: {},
                emphasis: {
                    focus: 'series',
                },
                data: [120, 132, 101, 134, 90, 230, 210],
            },
            {
                name: 'Union Ads',
                type: 'line',
                stack: 'Total',
                areaStyle: {},
                emphasis: {
                    focus: 'series',
                },
                data: [220, 182, 191, 234, 290, 330, 310],
            },
            {
                name: 'Video Ads',
                type: 'line',
                stack: 'Total',
                areaStyle: {},
                emphasis: {
                    focus: 'series',
                },
                data: [150, 232, 201, 154, 190, 330, 410],
            },
            {
                name: 'Direct',
                type: 'line',
                stack: 'Total',
                areaStyle: {},
                emphasis: {
                    focus: 'series',
                },
                data: [320, 332, 301, 334, 390, 330, 320],
            },
            {
                name: 'Search Engine',
                type: 'line',
                stack: 'Total',
                label: {
                    show: true,
                    position: 'top',
                },
                areaStyle: {},
                emphasis: {
                    focus: 'series',
                },
                data: [820, 932, 901, 934, 1290, 1330, 1320],
            },
        ],
    };
    useEcharts(myChart, option);
});
</script>

<style scoped lang="scss">
@import './index.scss';
</style>
