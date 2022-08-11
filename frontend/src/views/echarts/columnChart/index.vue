<template>
    <div ref="echartsRef" class="content-box"></div>
</template>

<script setup lang="ts" name="columnChart">
import { ref, onMounted } from 'vue';
import * as echarts from 'echarts';
import { useEcharts } from '@/hooks/useEcharts';
const echartsRef = ref<HTMLElement>();
onMounted(() => {
    let myChart: echarts.ECharts = echarts.init(echartsRef.value as HTMLElement);
    let option: echarts.EChartsOption = {
        tooltip: {
            trigger: 'axis',
            axisPointer: {
                type: 'shadow',
            },
        },
        legend: {
            textStyle: {
                color: '#a1a1a1',
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
                name: 'Direct',
                type: 'bar',
                emphasis: {
                    focus: 'series',
                },
                data: [320, 332, 301, 334, 390, 330, 320],
            },
            {
                name: 'Email',
                type: 'bar',
                stack: 'Ad',
                emphasis: {
                    focus: 'series',
                },
                data: [120, 132, 101, 134, 90, 230, 210],
            },
            {
                name: 'Union Ads',
                type: 'bar',
                stack: 'Ad',
                emphasis: {
                    focus: 'series',
                },
                data: [220, 182, 191, 234, 290, 330, 310],
            },
            {
                name: 'Video Ads',
                type: 'bar',
                stack: 'Ad',
                emphasis: {
                    focus: 'series',
                },
                data: [150, 232, 201, 154, 190, 330, 410],
            },
            {
                name: 'Search Engine',
                type: 'bar',
                data: [862, 1018, 964, 1026, 1679, 1600, 1570],
                emphasis: {
                    focus: 'series',
                },
                markLine: {
                    lineStyle: {
                        type: 'dashed',
                    },
                    data: [[{ type: 'min' }, { type: 'max' }]],
                },
            },
            {
                name: 'Baidu',
                type: 'bar',
                barWidth: 5,
                stack: 'Search Engine',
                emphasis: {
                    focus: 'series',
                },
                data: [620, 732, 701, 734, 1090, 1130, 1120],
            },
            {
                name: 'Google',
                type: 'bar',
                stack: 'Search Engine',
                emphasis: {
                    focus: 'series',
                },
                data: [120, 132, 101, 134, 290, 230, 220],
            },
            {
                name: 'Bing',
                type: 'bar',
                stack: 'Search Engine',
                emphasis: {
                    focus: 'series',
                },
                data: [60, 72, 71, 74, 190, 130, 110],
            },
            {
                name: 'Others',
                type: 'bar',
                stack: 'Search Engine',
                emphasis: {
                    focus: 'series',
                },
                data: [62, 82, 91, 84, 109, 110, 120],
            },
        ],
    };
    useEcharts(myChart, option);
});
</script>

<style scoped lang="scss">
@import './index.scss';
</style>
