<template>
    <div :id="id" ref="LineChartRef" :style="{ height: height, width: width }" />
</template>
<script lang="ts" setup>
import { onMounted, nextTick, watch, onBeforeUnmount, ref } from 'vue';
import echarts from '@/utils/echarts';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { computeSizeFromKBs, computeSizeFromKB, computeSizeFromMB } from '@/utils/size';
import i18n from '@/lang';
const { themeConfig } = useGlobalStore();
const isDarkTheme = ref(false);
const LineChartRef = ref<HTMLElement>();
let mediaQuery: MediaQueryList;
let resizeObserver: ResizeObserver | undefined;
const props = defineProps({
    id: {
        type: String,
        default: 'lineChartId',
    },
    width: {
        type: String,
        default: '100%',
    },
    height: {
        type: String,
        default: '200px',
    },
    dataZoom: {
        type: Boolean,
        default: false,
    },
    renderer: {
        type: String,
        default: undefined,
    },
    option: {
        type: Object,
        required: true,
    },
});

const seriesStyle = [
    {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            {
                offset: 0,
                color: getComputedStyle(document.documentElement)
                    .getPropertyValue('--panel-color-primary-light-9')
                    .trim(),
            },
            {
                offset: 1,
                color: getComputedStyle(document.documentElement).getPropertyValue('--panel-color-primary').trim(),
            },
        ]),
    },
    {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            {
                offset: 0,
                color: 'rgba(0, 94, 235, .3)',
            },
            {
                offset: 1,
                color: 'rgba(0, 94, 235, .4)',
            },
        ]),
    },
    {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            {
                offset: 0,
                color: 'rgba(27, 143, 60, .3)',
            },
            {
                offset: 1,
                color: 'rgba(27, 143, 60, .4)',
            },
        ]),
    },
    {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            {
                offset: 0,
                color: 'rgba(249, 199, 79, .3)',
            },
            {
                offset: 1,
                color: 'rgba(249, 199, 79, .4)',
            },
        ]),
    },
    {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            {
                offset: 0,
                color: 'rgba(255, 173, 177, 0.3)',
            },
            {
                offset: 1,
                color: 'rgba(255, 173, 177, .4)',
            },
        ]),
    },
];

function initChart() {
    const chartDom = LineChartRef.value;
    if (!chartDom) {
        return;
    }
    if (themeConfig.value.theme === 'auto') {
        isDarkTheme.value = window.matchMedia('(prefers-color-scheme: dark)').matches;
    } else {
        isDarkTheme.value = themeConfig.value.theme === 'dark';
    }
    let itemChart = echarts?.getInstanceByDom(chartDom);
    const optionItem = itemChart?.getOption();
    const itemSelect = Array.isArray(optionItem?.legend) ? optionItem?.legend?.[0] : optionItem?.legend;
    if (itemChart == null) {
        const initOptions = props.renderer ? { renderer: props.renderer as 'canvas' | 'svg' } : undefined;
        itemChart = echarts.init(chartDom, undefined, initOptions);
    }
    const style = getComputedStyle(document.documentElement);
    const primaryTextColor = style.getPropertyValue('--el-text-color-primary').trim() || '#303133';
    const regularTextColor = style.getPropertyValue('--el-text-color-regular').trim() || '#606266';
    const secondaryTextColor = style.getPropertyValue('--el-text-color-secondary').trim() || '#909399';
    const borderColor = style.getPropertyValue('--el-border-color-light').trim() || '#e4e7ed';
    const tooltipBackgroundColor = style.getPropertyValue('--el-bg-color-overlay').trim() || '#ffffff';

    const series = [];
    if (props.option?.yData?.length) {
        props.option?.yData.forEach((item: any, index: number) => {
            series.push({
                name: item?.name,
                type: 'line',
                itemStyle: seriesStyle[index + 2],
                areaStyle: seriesStyle[index],
                data: item?.data,
                showSymbol: false,
                yAxisIndex: item.yAxisIndex ? 1 : null,
            });
        });
    }
    const yAxis = [];
    if (props.option.yAxis && props.option.yAxis.length > 0) {
        props.option.yAxis.forEach((item: any) => {
            yAxis.push({
                ...item,
                splitLine: {
                    ...item.splitLine,
                    show: true,
                    lineStyle: {
                        ...item.splitLine?.lineStyle,
                        type: 'dashed',
                        opacity: isDarkTheme.value ? 0.1 : 1,
                        color: borderColor,
                    },
                },
                axisLabel: {
                    color: secondaryTextColor,
                    ...item.axisLabel,
                },
                nameTextStyle: {
                    color: secondaryTextColor,
                    ...item.nameTextStyle,
                },
            });
        });
    }

    const grid = props.option.grid || { left: '7%', right: '7%', bottom: '20%' };
    const defaultTooltip = {
        trigger: 'axis',
        formatter: function (datas: any) {
            let res = datas[0].name + '<br/>';
            switch (props.option.formatStr) {
                case 'KB/s':
                    for (const item of datas) {
                        res +=
                            item.marker +
                            ' ' +
                            item.seriesName +
                            i18n.global.t('commons.colon') +
                            computeSizeFromKBs(item.data) +
                            '<br/>';
                    }
                    break;
                case 'KB':
                    for (const item of datas) {
                        res +=
                            item.marker +
                            ' ' +
                            item.seriesName +
                            i18n.global.t('commons.colon') +
                            computeSizeFromKB(item.data) +
                            '<br/>';
                    }
                    break;
                case 'MB':
                    for (const item of datas) {
                        res +=
                            item.marker +
                            ' ' +
                            item.seriesName +
                            i18n.global.t('commons.colon') +
                            computeSizeFromMB(item.data) +
                            '<br/>';
                    }
                    break;
                default:
                    for (const item of datas) {
                        res +=
                            item.marker +
                            ' ' +
                            item.seriesName +
                            i18n.global.t('commons.colon') +
                            item.data +
                            props.option.formatStr +
                            '<br/>';
                    }
                    break;
            }
            return res;
        },
    };
    const tooltip = props.option.tooltip || defaultTooltip;
    const option = {
        title: [
            {
                left: 'center',
                text: props.option.title,
                show: props.option.title,
                textStyle: {
                    color: primaryTextColor,
                    fontWeight: 500,
                },
            },
        ],
        zlevel: 1,
        z: 1,
        tooltip: {
            appendToBody: true,
            backgroundColor: tooltipBackgroundColor,
            borderColor,
            textStyle: {
                color: regularTextColor,
            },
            ...tooltip,
            extraCssText: `${tooltip.extraCssText || ''}; z-index: 3000;`,
        },
        grid,
        legend: {
            ...itemSelect,
            right: grid.right || 10,
            itemWidth: 8,
            textStyle: {
                ...itemSelect?.textStyle,
                color: regularTextColor,
            },
            icon: 'circle',
        },
        xAxis: {
            ...props.option.xAxis,
            data: props.option.xData,
            boundaryGap: false,
            axisLabel: {
                ...props.option.xAxis?.axisLabel,
                color: secondaryTextColor,
            },
            axisLine: {
                ...props.option.xAxis?.axisLine,
                lineStyle: {
                    ...props.option.xAxis?.axisLine?.lineStyle,
                    color: borderColor,
                },
            },
            axisTick: {
                ...props.option.xAxis?.axisTick,
                lineStyle: {
                    ...props.option.xAxis?.axisTick?.lineStyle,
                    color: borderColor,
                },
            },
        },
        yAxis: props.option.yAxis
            ? yAxis
            : {
                  name: '( ' + props.option.formatStr + ' )',
                  nameTextStyle: {
                      color: secondaryTextColor,
                  },
                  axisLabel: {
                      color: secondaryTextColor,
                  },
                  splitLine: {
                      //分隔辅助线
                      lineStyle: {
                          type: 'dashed', //线的类型 虚线0
                          opacity: isDarkTheme.value ? 0.1 : 1, //透明度
                          color: borderColor,
                      },
                  },
              },
        series: series,
        dataZoom: [
            {
                ...(props.option.xData?.[0] ? { startValue: props.option.xData[0] } : {}),
                left: grid.left,
                right: grid.right,
                show: props.dataZoom,
            },
        ],
    };
    // 渲染数据
    itemChart.setOption(option, true);
}

function changeChartSize() {
    const chartDom = LineChartRef.value;
    if (!chartDom) {
        return;
    }
    echarts.getInstanceByDom(chartDom)?.resize();
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
        if (LineChartRef.value) {
            resizeObserver = new ResizeObserver(changeChartSize);
            resizeObserver.observe(LineChartRef.value);
        }
    });
});

onBeforeUnmount(() => {
    if (LineChartRef.value) {
        echarts.getInstanceByDom(LineChartRef.value)?.dispose();
    }
    window.removeEventListener('resize', changeChartSize);
    mediaQuery?.removeEventListener('change', handleThemeChange);
    resizeObserver?.disconnect();
});
</script>
<style lang="scss" scoped></style>
