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
        title: {
            text: 'Basic Radar Chart',
            textStyle: {
                color: '#a1a1a1',
            },
        },
        legend: {
            data: ['Allocated Budget', 'Actual Spending'],
            textStyle: {
                color: '#a1a1a1',
            },
        },
        radar: {
            // shape: 'circle',
            indicator: [
                { name: 'Sales', max: 6500 },
                { name: 'Administration', max: 16000 },
                { name: 'Information Technology', max: 30000 },
                { name: 'Customer Support', max: 38000 },
                { name: 'Development', max: 52000 },
                { name: 'Marketing', max: 25000 },
            ],
        },
        series: [
            {
                name: 'Budget vs spending',
                type: 'radar',
                data: [
                    {
                        value: [4200, 3000, 20000, 35000, 50000, 18000],
                        name: 'Allocated Budget',
                    },
                    {
                        value: [5000, 14000, 28000, 26000, 42000, 21000],
                        name: 'Actual Spending',
                    },
                ],
            },
        ],
    };
    useEcharts(myChart, option);
});
</script>

<style scoped lang="scss">
@import './index.scss';
</style>
