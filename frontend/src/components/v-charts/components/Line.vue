<template>
    <div :id="id" ref="LineChartRef" :style="{ height: height, width: width }" />
</template>
<script lang="ts" setup>
import { onMounted, nextTick, watch, onBeforeUnmount } from 'vue';
import * as echarts from 'echarts';
import { GlobalStore } from '@/store';
import { computeSizeFromKBs, computeSizeFromKB, computeSizeFromMB } from '@/utils/util';
import { storeToRefs } from 'pinia';
import i18n from '@/lang';
const globalStore = GlobalStore();
const { isDarkTheme } = storeToRefs(globalStore);

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
                color: getComputedStyle(document.documentElement).getPropertyValue('--panel-color-primary').trim(),
            },
            {
                offset: 1,
                color: getComputedStyle(document.documentElement)
                    .getPropertyValue('--panel-color-primary-light-9')
                    .trim(),
            },
        ]),
    },
    {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            {
                offset: 0,
                color: 'rgba(0, 94, 235, .5)',
            },
            {
                offset: 1,
                color: 'rgba(0, 94, 235, 0)',
            },
        ]),
    },
    {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            {
                offset: 0,
                color: 'rgba(27, 143, 60, .5)',
            },
            {
                offset: 1,
                color: 'rgba(27, 143, 60, 0)',
            },
        ]),
    },
    {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            {
                offset: 0,
                color: 'rgba(249, 199, 79, .5)',
            },
            {
                offset: 1,
                color: 'rgba(249, 199, 79, 0)',
            },
        ]),
    },
    {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            {
                offset: 0,
                color: 'rgba(255, 173, 177, 0.5)',
            },
            {
                offset: 1,
                color: 'rgba(255, 173, 177, 0)',
            },
        ]),
    },
];

function initChart() {
    let itemChart = echarts?.getInstanceByDom(document.getElementById(props.id) as HTMLElement);
    const optionItem = itemChart?.getOption();
    const itemSelect = optionItem?.legend;
    if (itemChart == null) {
        itemChart = echarts.init(document.getElementById(props.id) as HTMLElement);
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
        tooltip: props.option.tooltip || {
            trigger: 'axis',
            formatter: function (dataList: any) {
                let res = dataList[0].name + '<br/>';
                switch (props.option.formatStr) {
                    case 'KB/s':
                        for (const item of dataList) {
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
                        for (const item of dataList) {
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
                        for (const item of dataList) {
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
                        for (const item of dataList) {
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
        },
        grid: props.option.grid || { left: '7%', right: '7%', bottom: '20%' },
        legend: itemSelect || {
            right: 10,
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
        dataZoom: [{ startValue: props?.option.xData[0], show: props.dataZoom }],
    };
    // 渲染数据
    itemChart.setOption(option, true);
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
