<template>
    <!-- 平台来源 -->
    <div class="echarts" id="PlatformSourceChart"></div>
</template>

<script setup lang="ts">
// Echarts 为init（dom元素后的类型）
// EChartsOption 为 option 的类型
import { ECharts, EChartsOption, init } from 'echarts';
interface ChartProp {
    name: string;
    value: number;
    percentage: string;
}
const initChart = (data: any = {}): ECharts => {
    const charEle = document.getElementById(
        'PlatformSourceChart',
    ) as HTMLElement;
    const charEch: ECharts = init(charEle);
    const option: EChartsOption = {
        grid: {
            top: '0%',
            left: '2%',
            right: '2%',
            bottom: '0%',
            // containLabel: true
        },
        tooltip: {
            trigger: 'item',
            formatter: '{b} :  {c}人',
        },
        legend: {
            show: true,
            top: 'middle',
            left: '20px',
            icon: 'circle',
            orient: 'vertical',
            align: 'auto',
            itemWidth: 10,
            textStyle: {
                color: '#fff',
            },
            itemGap: 20,
            formatter: function (name: string) {
                let text = '';
                data.data.forEach((val: ChartProp) => {
                    if (val.name === name) {
                        text = name + ' --- ' + val.percentage;
                    }
                });
                return text;
            },
            data: data.data.map((val: ChartProp) => val.name),
        },
        series: [
            {
                type: 'pie',
                radius: ['60%', '85%'],
                center: ['68%', '45%'],
                color: [
                    '#0E7CE2',
                    '#FF8352',
                    '#E271DE',
                    '#F8456B',
                    '#00FFFF',
                    '#4AEAB0',
                ],
                itemStyle: {
                    borderColor: '#031845',
                    borderWidth: 10,
                },
                data: data.data,
                labelLine: {
                    show: false,
                },
                label: {
                    show: false,
                },
            },
            {
                type: 'pie',
                radius: ['20%', '28%'],
                center: ['68%', '45%'],
                color: ['#ffffff', 'red'],
                startAngle: 105,
                data: [
                    {
                        value: 30,
                        name: '',
                        itemStyle: {
                            color: 'transparent',
                        },
                    },
                    {
                        value: 5,
                        name: '',
                        itemStyle: {
                            color: 'transparent',
                        },
                    },
                    {
                        value: 65,
                        name: 'ddd',
                        itemStyle: {
                            color: '#ffffff',
                        },
                    },
                ],
                silent: true,
                labelLine: {
                    show: false,
                },
                label: {
                    show: false,
                },
            },
            {
                type: 'pie',
                radius: [0, '30%'],
                center: ['68%', '45%'],
                startAngle: 90,
                data: [
                    {
                        value: 25,
                        name: '1',
                        itemStyle: {
                            color: 'transparent',
                            borderWidth: 4,
                            borderColor: '#ffffff',
                        },
                    },

                    {
                        value: 75,
                        name: '2',
                        itemStyle: {
                            color: 'transparent',
                        },
                    },
                ],
                selectedOffset: 10,
                silent: true,
                labelLine: {
                    show: false,
                },
                label: {
                    show: false,
                },
            },
            {
                type: 'pie',
                radius: ['96%', '97%'],
                center: ['68%', '45%'],
                color: [
                    '#007afe',
                    'transparent',
                    '#007afe',
                    'transparent',
                    '#007afe',
                    'transparent',
                ],
                data: [
                    {
                        value: 17,
                        name: '11',
                    },
                    {
                        value: 17,
                        name: '22',
                    },
                    {
                        value: 17,
                        name: '33',
                    },
                    {
                        value: 17,
                        name: '44',
                    },
                    {
                        value: 17,
                        name: '55',
                    },
                    {
                        value: 17,
                        name: '66',
                    },
                ],
                silent: true,
                labelLine: {
                    show: false,
                },
                label: {
                    show: false,
                },
            },
            {
                type: 'pie',
                zlevel: 0,
                silent: true,
                radius: ['45%', '46%'],
                center: ['68%', '45%'],
                z: 10,
                label: {
                    show: false,
                },
                labelLine: {
                    show: false,
                },
                data: new Array(150)
                    .fill('')
                    .map((val: string, index: number) => {
                        if (index % 3 === 0) {
                            return {
                                name: (index + 1).toString(),
                                value: 10,
                                itemStyle: {
                                    color: '#fff',
                                    borderWidth: 0,
                                    borderColor: 'rgba(0,0,0,0)',
                                },
                            };
                        } else {
                            return {
                                name: (index + 1).toString(),
                                value: 25,
                                itemStyle: {
                                    color: 'rgba(0,0,0,0)',
                                    borderWidth: 0,
                                    borderColor: 'rgba(0,0,0,0)',
                                },
                            };
                        }
                    }),
            },
            {
                type: 'pie',
                zlevel: 0,
                silent: true,
                radius: ['58%', '60%'],
                center: ['68%', '45%'],
                z: 10,
                startAngle: 90,
                label: {
                    show: false,
                },
                color: ['red', 'blue', 'red', 'blue'],

                labelLine: {
                    show: false,
                },

                data: [
                    {
                        name: 'r1',
                        value: 25,
                        itemStyle: {
                            color: {
                                type: 'linear',
                                x: 0,
                                y: 0,
                                x2: 0,
                                y2: 1,
                                colorStops: [
                                    {
                                        offset: 0,
                                        color: 'rgba(51,149,191,0.5)', // 0% 处的颜色
                                    },
                                    {
                                        offset: 1,
                                        color: 'rgba(51,149,191,0)', // 100% 处的颜色
                                    },
                                ],
                                global: false, // 缺省为 false
                            },
                        },
                    },
                    {
                        name: 'r2',
                        value: 25,
                        itemStyle: {
                            color: {
                                type: 'linear',
                                x: 0,
                                y: 0,
                                x2: 0,
                                y2: 1,
                                colorStops: [
                                    {
                                        offset: 0,
                                        color: 'rgba(0,0,0,0)', // 0% 处的颜色
                                    },
                                    {
                                        offset: 1,
                                        color: 'rgba(51,149,191,0.5)', // 100% 处的颜色
                                    },
                                ],
                                global: false, // 缺省为 false
                            },
                        },
                    },
                    {
                        name: 'r3',
                        value: 25,
                        itemStyle: {
                            //线性渐变，前4个参数分别是x0,y0,x2,y2(范围0~1);相当于图形包围盒中的百分比。如果最后一个参数是‘true’，则该四个值是绝对像素位置。
                            color: {
                                type: 'linear',
                                x: 0,
                                y: 0,
                                x2: 0,
                                y2: 1,
                                colorStops: [
                                    {
                                        offset: 0,
                                        color: 'rgba(51,149,191,0)', // 0% 处的颜色
                                    },
                                    {
                                        offset: 1,
                                        color: 'rgba(51,149,191,0.5)', // 100% 处的颜色
                                    },
                                ],
                                global: false, // 缺省为 false
                            },
                        },
                    },
                    {
                        name: 'r4',
                        value: 25,
                        itemStyle: {
                            //线性渐变，前4个参数分别是x0,y0,x2,y2(范围0~1);相当于图形包围盒中的百分比。如果最后一个参数是‘true’，则该四个值是绝对像素位置。
                            color: {
                                type: 'linear',
                                x: 0,
                                y: 0,
                                x2: 0,
                                y2: 1,
                                colorStops: [
                                    {
                                        offset: 0,
                                        color: 'rgba(51,149,191,0.5)', // 0% 处的颜色
                                    },
                                    {
                                        offset: 1,
                                        color: 'rgba(0,0,0,0)', // 100% 处的颜色
                                    },
                                ],
                                global: false, // 缺省为 false
                            },
                        },
                    },
                ],
            },
        ],
    };
    charEch.setOption(option, true);
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
</style>
