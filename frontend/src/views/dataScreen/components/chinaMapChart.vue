<template>
    <!-- 中国地图 -->
    <div class="map-ball"></div>
    <div class="echarts" id="mapChart"></div>
</template>

<script setup lang="ts">
// Echarts 为init（dom元素后的类型）
// EChartsOption 为 option 的类型
import { ECharts, EChartsOption, init } from 'echarts';
import echarts from '@/utils/echarts';
import mapJson from '../assets/china.json';
/* echarts sysmbol */
let planePath: string =
    'path://M1705.06,1318.313v-89.254l-319.9-221.799l0.073-208.063c0.521-84.662-26.629-121.796-63.961-121.491c-37.332-0.305-64.482,36.829-63.961,121.491l0.073,208.063l-319.9,221.799v89.254l330.343-157.288l12.238,241.308l-134.449,92.931l0.531,42.034l175.125-42.917l175.125,42.917l0.531-42.034l-134.449-92.931l12.238-241.308L1705.06,1318.313z';
const initChart = (data: any = []): ECharts => {
    const charEle = document.getElementById('mapChart') as HTMLElement;
    const charEch: ECharts = init(charEle);
    echarts.registerMap('china', mapJson as any);
    const option: EChartsOption = {
        // 悬浮窗
        tooltip: {
            trigger: 'item',
            formatter: function (params: any) {
                return `${params.name}: ${params.value || '-'}`;
            },
        },
        // echarts大小位置
        grid: {
            left: '0px',
            right: '80px',
            top: '10px',
            bottom: '10px',
        },
        geo: {
            map: 'china',
            zoom: 1.5,
            center: [102.848234, 32.82333],
            scaleLimit: {
                min: 0.8,
            },
            label: {
                color: '#fff',
                show: true,
            },
            emphasis: {
                label: {
                    color: '#fff',
                    show: true,
                },
                itemStyle: {
                    areaColor: {
                        x: 0,
                        y: 0,
                        x2: 0,
                        y2: 1,
                        colorStops: [
                            {
                                offset: 0,
                                color: '#073684', // 0% 处的颜色
                            },
                            {
                                offset: 1,
                                color: '#2B91B7', // 100% 处的颜色
                            },
                        ],
                    },
                },
            },
            roam: false,
            itemStyle: {
                areaColor: {
                    x: 0,
                    y: 0,
                    x2: 0,
                    y2: 1,
                    colorStops: [
                        {
                            offset: 0,
                            color: '#073684', // 0% 处的颜色
                        },
                        {
                            offset: 1,
                            color: '#061E3D', // 100% 处的颜色
                        },
                    ],
                },
                borderColor: new echarts.graphic.LinearGradient(
                    0,
                    0,
                    0,
                    1,
                    [
                        {
                            offset: 0,
                            color: '#00F6FF',
                        },
                        {
                            offset: 1,
                            color: '#87ADCB',
                        },
                    ],
                    false,
                ),
                shadowColor: 'rgba(10,76,139,1)',
                shadowOffsetY: 0,
                shadowBlur: 60,
                borderWidth: 1,
            },
            tooltip: {
                show: false,
            },
        },
        // 要显示的散点数据
        series: [
            {
                name: '',
                type: 'lines',
                coordinateSystem: 'geo',
                zlevel: 1,
                effect: {
                    show: true,
                    period: 6,
                    trailLength: 0.7,
                    color: 'red', //arrow箭头的颜色
                    symbolSize: 3,
                },
                lineStyle: {
                    color: '#fff',
                    width: 0,
                    curveness: 0.2,
                },
                data,
            },
            {
                name: '',
                type: 'lines',
                coordinateSystem: 'geo',
                zlevel: 2,
                symbol: ['none', 'arrow'],
                symbolSize: 10,
                effect: {
                    show: true,
                    period: 6,
                    trailLength: 0,
                    symbol: planePath,
                    symbolSize: 15,
                },
                lineStyle: {
                    color: '#fff',
                    width: 1,
                    opacity: 0.6,
                    curveness: 0.2,
                },
                data,
            },
        ],
    };
    charEch.setOption(option);
    return charEch;
};
defineExpose({
    initChart,
});
</script>
<style lang="scss" scoped>
.echarts {
    width: 100%;
    height: 100%;
}
.map-ball {
    position: absolute;
    top: 50%;
    left: 50%;
    width: 900px;
    height: 900px;
    transform: translate(-50%, -50%);
    img {
        width: 500px;
        height: 500px;
    }
}
</style>
