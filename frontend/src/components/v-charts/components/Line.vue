<template>
    <div :id="id" class="v-charts" ref="chartRef" :style="{ height: height, width: width }" />
</template>
<script lang="ts" setup>
import { onMounted, onUnmounted, nextTick, watch } from 'vue';
import * as echarts from 'echarts';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();
const props = defineProps({
    id: {
        type: String,
        default: 'VChartsId',
    },
    width: {
        type: String,
        default: '100%',
    },
    height: {
        type: String,
        default: '200px',
    },
    type: {
        type: String,
        default: 'line',
    },
    option: {
        type: Object,
        required: true,
    }, // option: { title , xDatas, yDatas, formatStr }
});

function initChart() {
    let itemChart = echarts?.getInstanceByDom(document.getElementById(props.id) as HTMLElement);
    // 如果不存在，就进行初始化
    if (itemChart == null) {
        itemChart = echarts.init(document.getElementById(props.id) as HTMLElement);
    }

    const theme = globalStore.$state.themeConfig.theme || 'light';
    // 把配置和数据放这里
    const option = {
        title: [
            {
                left: 'center',
                text: props.option.title,
                show: false,
            },
        ],
        zlevel: 1,
        z: 1,
        tooltip: {
            trigger: 'axis',
            formatter: function (datas: any) {
                let res = datas[0].name + '<br/>';
                for (const item of datas) {
                    res += item.marker + ' ' + item.seriesName + '：' + item.data + props.option.formatStr + '<br/>';
                }
                return res;
            },
        },
        grid: { left: '7%', right: '7%', bottom: '20%' },
        legend: {
            data: [props.option?.yDatas[0]?.name, props.option?.yDatas[1]?.name],
            right: 10,
            itemWidth: 8,
            textStyle: {
                color: '#646A73',
            },
            icon: 'circle',
        },
        xAxis: { data: props.option.xDatas, boundaryGap: false },
        yAxis: {
            name: '( ' + props.option.formatStr + ' )',
            splitLine: {
                //分隔辅助线
                lineStyle: {
                    type: 'dashed', //线的类型 虚线0
                    opacity: theme === 'dark' ? 0.1 : 1, //透明度
                },
            },
        },
        series: [
            {
                name: props.option?.yDatas[0]?.name,
                type: 'line',
                areaStyle: {
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
                data: props.option?.yDatas[0]?.data,
                showSymbol: false,
            },
            {
                name: props.option?.yDatas[1]?.name,
                type: 'line',
                areaStyle: {
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
                data: props.option?.yDatas[1]?.data,
                showSymbol: false,
            },
        ],
        dataZoom: [{ startValue: props?.option.xDatas[0] }],
    };
    // 渲染数据
    itemChart.setOption(option, true);

    window.onresize = function () {
        //自适应大小
        itemChart.resize();
    };
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
    });
});

onUnmounted(() => {
    echarts.dispose;
});
</script>
<style lang="scss" scoped></style>
