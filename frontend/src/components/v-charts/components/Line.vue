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
    const itemSelect = optionItem?.legend;
    if (itemChart == null) {
        itemChart = echarts.init(chartDom);
    }

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
                splitLine: {
                    show: true,
                    lineStyle: {
                        type: 'dashed',
                        opacity: isDarkTheme.value ? 0.1 : 1,
                    },
                },
                ...item,
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
            },
        ],
        zlevel: 1,
        z: 1,
        tooltip: {
            appendToBody: true,
            ...tooltip,
            extraCssText: `${tooltip.extraCssText || ''}; z-index: 3000;`,
        },
        grid,
        legend: itemSelect || {
            right: grid.right || 10,
            itemWidth: 8,
            textStyle: {
                color: '#646A73',
            },
            icon: 'circle',
        },
        xAxis: { data: props.option.xData, boundaryGap: false },
        yAxis: props.option.yAxis
            ? yAxis
            : {
                  name: '( ' + props.option.formatStr + ' )',
                  splitLine: {
                      //分隔辅助线
                      lineStyle: {
                          type: 'dashed', //线的类型 虚线0
                          opacity: isDarkTheme.value ? 0.1 : 1, //透明度
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
