<template>
    <div :id="id" ref="PieChartRef" :style="{ height: height, width: width }" />
</template>
<script lang="ts" setup>
import { onMounted, nextTick, watch, onBeforeUnmount, ref } from 'vue';
import echarts from '@/utils/echarts';
const PieChartRef = ref<HTMLElement>();
let themeObserver: MutationObserver | undefined;
let resizeObserver: ResizeObserver | undefined;
let resizeFrame: number | undefined;

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
    if (resizeFrame !== undefined) {
        return;
    }
    resizeFrame = requestAnimationFrame(() => {
        resizeFrame = undefined;
        resizeChart();
    });
}

function resizeChart() {
    if (!PieChartRef.value) {
        return;
    }
    echarts.getInstanceByDom(PieChartRef.value)?.resize();
}
function getThemeColors() {
    // Match the applied page theme, including while settings are still loading.
    const isDarkTheme = document.documentElement.classList.contains('dark');
    const style = getComputedStyle(document.documentElement);
    return {
        primaryLight2: style.getPropertyValue('--panel-color-primary-light-3').trim() || '#4c8ef1',
        primaryLight1: style.getPropertyValue('--panel-color-primary').trim() || '#005eeb',
        pieBgColor: isDarkTheme ? '#434552' : '#ffffff',
        textColor: isDarkTheme ? '#ffffff' : '#0f0f0f',
        subtextColor: isDarkTheme ? '#BBBFC4' : '#646A73',
        shadowColor: isDarkTheme ? '#16191D' : 'rgba(0, 94, 235, 0.1)',
        backgroundStyleColor: isDarkTheme ? 'rgba(255, 255, 255, 0.05)' : 'rgba(0, 94, 235, 0.05)',
    };
}

function initChart() {
    if (!PieChartRef.value) {
        return;
    }
    let myChart = echarts?.getInstanceByDom(PieChartRef.value);
    if (myChart === null || myChart === undefined) {
        myChart = echarts.init(PieChartRef.value);
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
        themeObserver = new MutationObserver(handleThemeChange);
        themeObserver.observe(document.documentElement, {
            attributes: true,
            attributeFilter: ['class', 'style'],
        });
        initChart();
        window.addEventListener('resize', changeChartSize);
        if (PieChartRef.value) {
            resizeObserver = new ResizeObserver(changeChartSize);
            resizeObserver.observe(PieChartRef.value);
        }
    });
});

onBeforeUnmount(() => {
    if (resizeFrame !== undefined) {
        cancelAnimationFrame(resizeFrame);
        resizeFrame = undefined;
    }
    if (PieChartRef.value) {
        echarts.getInstanceByDom(PieChartRef.value)?.dispose();
    }
    window.removeEventListener('resize', changeChartSize);
    themeObserver?.disconnect();
    resizeObserver?.disconnect();
});
</script>
<style lang="scss" scoped></style>
