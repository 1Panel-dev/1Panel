<template>
    <div :id="id" ref="PieChartRef" :style="{ height: height, width: width }" />
</template>
<script lang="ts" setup>
import { onMounted, nextTick, watch, onBeforeUnmount, ref } from 'vue';
import echarts from '@/utils/echarts';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();
const isDarkTheme = ref(false);
let mediaQuery: MediaQueryList;

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
function changeChartSize() {
    echarts.getInstanceByDom(document.getElementById(props.id) as HTMLElement)?.resize();
}
function getThemeColors() {
    return {
        primaryLight2: getComputedStyle(document.documentElement)
            .getPropertyValue('--panel-color-primary-light-3')
            .trim(),
        primaryLight1: getComputedStyle(document.documentElement).getPropertyValue('--panel-color-primary').trim(),
        pieBgColor: isDarkTheme.value ? '#434552' : '#ffffff',
        textColor: isDarkTheme.value ? '#ffffff' : '#0f0f0f',
        subtextColor: isDarkTheme.value ? '#BBBFC4' : '#646A73',
        shadowColor: isDarkTheme.value ? '#16191D' : 'rgba(0, 94, 235, 0.1)',
        backgroundStyleColor: isDarkTheme.value ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 94, 235, 0.05)',
    };
}

function initChart() {
    if (globalStore.themeConfig.theme === 'auto') {
        isDarkTheme.value = window.matchMedia('(prefers-color-scheme: dark)').matches;
    } else {
        isDarkTheme.value = globalStore.themeConfig.theme === 'dark';
    }
    let myChart = echarts?.getInstanceByDom(document.getElementById(props.id) as HTMLElement);
    if (myChart === null || myChart === undefined) {
        myChart = echarts.init(document.getElementById(props.id) as HTMLElement);
    }
    let percentText = String(props.option.data).split('.');
    const { primaryLight2, primaryLight1, pieBgColor, textColor, subtextColor, shadowColor, backgroundStyleColor } =
        getThemeColors();
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

                    color: textColor,
                    lineHeight: 25,
                    // fontSize: 20,
                    fontWeight: 500,
                },
                left: '49%',
                top: '32%',
                subtext: props.option.title,
                subtextStyle: {
                    color: subtextColor,
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
                    color: backgroundStyleColor,
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
                            shadowColor: shadowColor,
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

function handleThemeChange() {
    nextTick(() => initChart());
}

onMounted(() => {
    nextTick(() => {
        mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
        mediaQuery.addEventListener('change', handleThemeChange);
        initChart();
        window.addEventListener('resize', changeChartSize);
    });
});

onBeforeUnmount(() => {
    echarts.getInstanceByDom(document.getElementById(props.id) as HTMLElement).dispose();
    window.removeEventListener('resize', changeChartSize);
    mediaQuery.removeEventListener('change', handleThemeChange);
});
</script>
<style lang="scss" scoped></style>
