<template>
    <div :id="id" ref="PieChartRef" :style="{ height: height, width: width }" />
</template>
<script lang="ts" setup>
import { onMounted, nextTick, watch, onBeforeUnmount } from 'vue';
import * as echarts from 'echarts';
import { GlobalStore } from '@/store';
import { storeToRefs } from 'pinia';
const globalStore = GlobalStore();
const { isDarkTheme } = storeToRefs(globalStore);

const props = defineProps({
    id: {
        type: String,
        default: 'pieChartId',
    },
    width: {
        type: String,
        default: '100%',
    },
    height: {
        type: String,
        default: '200px',
    },
    option: {
        type: Object,
        required: true,
    },
});

function initChart() {
    let myChart = echarts?.getInstanceByDom(document.getElementById(props.id) as HTMLElement);
    if (myChart === null || myChart === undefined) {
        myChart = echarts.init(document.getElementById(props.id) as HTMLElement);
    }
    let percentText = String(props.option.data).split('.');
    const primaryLight2 = getComputedStyle(document.documentElement)
        .getPropertyValue('--panel-color-primary-light-3')
        .trim();
    const primaryLight1 = getComputedStyle(document.documentElement).getPropertyValue('--panel-color-primary').trim();
    const pieBgColor = getComputedStyle(document.documentElement).getPropertyValue('--panel-pie-bg-color').trim();

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

                    color: isDarkTheme.value ? '#ffffff' : '#0f0f0f',
                    lineHeight: 25,
                    // fontSize: 20,
                    fontWeight: 500,
                },
                left: '49%',
                top: '32%',
                subtext: props.option.title,
                subtextStyle: {
                    color: isDarkTheme.value ? '#BBBFC4' : '#646A73',
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
                    color: isDarkTheme.value ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 94, 235, 0.05)',
                },
                color: [
                    new echarts.graphic.LinearGradient(0, 1, 0, 0, [
                        {
                            offset: 0,
                            color: primaryLight2,
                        },
                        {
                            offset: 1,
                            color: primaryLight1,
                        },
                    ]),
                ],
                label: {
                    show: false,
                },
                data: [props.option.data],
            },
            {
                type: 'pie',
                radius: ['0%', '60%'],
                center: ['50%', '50%'],
                label: {
                    show: false,
                },
                color: pieBgColor,
                data: [
                    {
                        value: 0,
                        itemStyle: {
                            shadowColor: isDarkTheme.value ? '#16191D' : 'rgba(0, 94, 235, 0.1)',
                            shadowBlur: 5,
                        },
                    },
                ],
            },
        ],
    };
    // 渲染数据
    myChart.setOption(option, true);
}

function changeChartSize() {
    echarts.getInstanceByDom(document.getElementById(props.id) as HTMLElement)?.resize();
}

watch(
    () => props.option,
    (val) => {
        if (val) {
            nextTick(() => {
                initChart();
            });
        }
    },
);

onMounted(() => {
    nextTick(() => {
        initChart();
        window.addEventListener('resize', changeChartSize);
    });
});

onBeforeUnmount(() => {
    echarts.getInstanceByDom(document.getElementById(props.id) as HTMLElement).dispose();
    window.removeEventListener('resize', changeChartSize);
});
</script>
<style lang="scss" scoped></style>
