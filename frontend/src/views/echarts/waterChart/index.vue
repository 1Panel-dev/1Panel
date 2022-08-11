<template>
    <div ref="echartsRef" class="content-box"></div>
</template>

<script setup lang="ts" name="columnChart">
import { ref, onMounted } from 'vue';
import { useEcharts } from '@/hooks/useEcharts';
import * as echarts from 'echarts';
import 'echarts-liquidfill';

const echartsRef = ref<HTMLElement>();
onMounted(() => {
    let myChart: echarts.ECharts = echarts.init(echartsRef.value as HTMLElement);
    let value = 0.5;
    let data = [value, value, value];
    let option: echarts.EChartsCoreOption = {
        title: [
            {
                text: '预约量',
                x: '25%',
                y: 30,
                textAlign: 'center',
                textStyle: {
                    color: '#a1a1a1',
                    fontSize: 16,
                    fontFamily: 'Microsoft Yahei',
                    fontWeight: '100',
                    textAlign: 'center',
                },
            },
            {
                text: '实时客流量',
                x: '75%',
                y: 30,
                textAlign: 'center',
                textStyle: {
                    color: '#a1a1a1',
                    fontSize: 16,
                    fontFamily: 'Microsoft Yahei',
                    fontWeight: '100',
                    textAlign: 'center',
                },
            },
            {
                text: (value * 100).toFixed(0) + '%',
                left: '25%',
                top: '38%',
                textAlign: 'center',
                textStyle: {
                    fontSize: '50',
                    fontWeight: '300',
                    color: '#a06a0a',
                    textAlign: 'center',
                    textBorderColor: 'rgba(0, 0, 0, 0)',
                    textShadowColor: '#fff',
                    textShadowBlur: '0',
                    textShadowOffsetX: 0,
                    textShadowOffsetY: 1,
                },
            },
            {
                text: (value * 100).toFixed(0) + '%',
                left: '75%',
                top: '38%',
                textAlign: 'center',
                textStyle: {
                    fontSize: '50',
                    fontWeight: '300',
                    color: '#02456d',
                    textAlign: 'center',
                    textBorderColor: 'rgba(0, 0, 0, 0)',
                    textShadowColor: '#fff',
                    textShadowBlur: '0',
                    textShadowOffsetX: 0,
                    textShadowOffsetY: 1,
                },
            },
        ],
        series: [
            {
                type: 'liquidFill',
                radius: '50%',
                z: 6,
                center: ['25%', '50%'],
                color: [
                    {
                        type: 'linear',
                        x: 0,
                        y: 0,
                        x2: 0,
                        y2: 1,
                        colorStops: [
                            {
                                offset: 1,
                                color: 'rgba(251, 173, 23, 0)',
                            },
                            {
                                offset: 0.5,
                                color: 'rgba(251, 173, 23, .2)',
                            },
                            {
                                offset: 0,
                                color: 'rgba(251, 173, 23, 1)',
                            },
                        ],
                        globalCoord: false,
                    },
                ],
                data: data,
                backgroundStyle: {
                    borderWidth: 1,
                    color: 'transparent',
                },
                label: {
                    normal: {
                        formatter: '',
                    },
                },
                outline: {
                    show: true,
                    itemStyle: {
                        borderWidth: 0,
                    },
                    borderDistance: 0,
                },
            },
            {
                name: '第二层白边',
                type: 'pie',
                z: 3,
                radius: ['0%', '55%'],
                center: ['25%', '50%'],
                hoverAnimation: false,
                itemStyle: {
                    normal: {
                        label: {
                            show: false,
                        },
                    },
                },
                data: [
                    {
                        value: 100,
                        itemStyle: {
                            normal: {
                                color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                                    {
                                        offset: 0,
                                        color: '#fefefe',
                                    },
                                    {
                                        offset: 1,
                                        color: '#e7e8ea',
                                    },
                                ]),
                            },
                        },
                    },
                    {
                        value: 0,
                        itemStyle: {
                            normal: {
                                color: 'transparent',
                            },
                        },
                    },
                ],
            },
            {
                name: '最外绿边',
                type: 'pie',
                z: 1,
                radius: ['0%', '58%'],
                center: ['25%', '50%'],
                hoverAnimation: false,
                itemStyle: {
                    normal: {
                        label: {
                            show: false,
                        },
                    },
                },
                data: [
                    {
                        value: 100,
                        itemStyle: {
                            color: '#fdc56e',
                        },
                    },
                    {
                        value: 0,
                        itemStyle: {
                            normal: {
                                color: 'transparent',
                            },
                        },
                    },
                ],
            },
            {
                type: 'liquidFill',
                radius: '50%',
                z: 6,
                center: ['75%', '50%'],
                color: ['#c1dce7', '#90d3f0', '#009bdb'],
                data: [0.6, { value: 0.5, direction: 'left' }, 0.4, 0.3],
                backgroundStyle: {
                    borderWidth: 1,
                    color: 'transparent',
                },
                label: {
                    normal: {
                        formatter: '',
                    },
                },
                outline: {
                    show: true,
                    itemStyle: {
                        borderWidth: 0,
                    },
                    borderDistance: 0,
                },
            },
            {
                name: '第二层白边',
                type: 'pie',
                z: 3,
                radius: ['0%', '55%'],
                center: ['75%', '50%'],
                hoverAnimation: false,
                itemStyle: {
                    normal: {
                        label: {
                            show: false,
                        },
                    },
                },
                data: [
                    {
                        value: 100,
                        itemStyle: {
                            normal: {
                                color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                                    {
                                        offset: 0,
                                        color: '#fefefe',
                                    },
                                    {
                                        offset: 1,
                                        color: '#e7e8ea',
                                    },
                                ]),
                            },
                        },
                    },
                    {
                        value: 0,
                        itemStyle: {
                            normal: {
                                color: 'transparent',
                            },
                        },
                    },
                ],
            },
            {
                name: '最外蓝边',
                type: 'pie',
                z: 1,
                radius: ['0%', '58%'],
                center: ['75%', '50%'],
                hoverAnimation: false,
                itemStyle: {
                    normal: {
                        label: {
                            show: false,
                        },
                    },
                },
                data: [
                    {
                        value: 100,
                        itemStyle: {
                            color: '#07a2e3',
                        },
                    },
                    {
                        value: 0,
                        itemStyle: {
                            normal: {
                                color: 'transparent',
                            },
                        },
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
