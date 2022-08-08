<template>
    <!-- 未来30天访问量趋势预测图 -->
    <div class="echarts" id="OverNext30Chart"></div>
</template>

<script setup lang="ts">
// Echarts 为init（dom元素后的类型）
// EChartsOption 为 option 的类型
import { ECharts, EChartsOption, init } from 'echarts';
interface ChartProp {
    unit: string[];
    data: number[];
}
const initChart = (data: any = {}): ECharts => {
    const charEle = document.getElementById('OverNext30Chart') as HTMLElement;
    const charEch: ECharts = init(charEle);
    const option: EChartsOption = {
        tooltip: {
            trigger: 'axis',
            confine: true,
            formatter: (params: any) => {
                let tipData = params[0];
                let html = `<div class="lineChart-bg">
                        <span style="">${tipData.name} <i >${tipData.value}</i> 人次访问</span>
                    </div>`;
                return html;
            },
            backgroundColor: 'transparent', //提示标签背景颜色
            borderColor: 'transparent',
            axisPointer: {
                lineStyle: {
                    type: 'dashed',
                },
                snap: true,
            },
            extraCssText: 'box-shadow: none;padding:0',
        },
        grid: {
            top: '15%',
            left: '5%',
            right: '5%',
            bottom: '15%',
            // containLabel: true
        },
        xAxis: [
            {
                type: 'category',
                boundaryGap: false,
                axisLine: {
                    //坐标轴轴线相关设置。数学上的x轴
                    show: true,
                    symbol: ['none', 'arrow'],
                    symbolOffset: [0, 30],
                    lineStyle: {
                        color: '#233653',
                        shadowOffsetX: 20,
                        shadowColor: '#233653',
                    },
                },
                axisLabel: {
                    //坐标轴刻度标签的相关设置
                    color: '#7ec7ff',
                    padding: 0,
                    fontSize: 12,
                    formatter: function (data) {
                        return data;
                    },
                },
                splitLine: {
                    show: false,
                    lineStyle: {
                        color: '#192a44',
                    },
                },
                axisTick: {
                    show: false,
                },
                data: initDate(),
            },
        ],
        yAxis: data.unit.map((val: ChartProp, index: number) => {
            return {
                name: '(访问量)',
                nameTextStyle: {
                    color: '#7ec7ff',
                    fontSize: 12,
                    padding: [0, 30, -4, 0],
                },
                // nameGap:18,
                minInterval: 1,
                splitLine: {
                    show: false,
                    lineStyle: {
                        color: '#192a44',
                    },
                },
                axisLine: {
                    show: index === 0 ? true : false,
                    lineStyle: {
                        color: '#233653',
                    },
                },
                axisLabel: {
                    show: true,
                    color: '#7ec7ff',
                    padding: 0,
                    formatter: function (value: string) {
                        if (Number(value) >= 10000) {
                            value = Number(value) / 10000 + 'w';
                        }
                        return value;
                    },
                },
                axisTick: {
                    show: false,
                },
            };
        }),
        series: data.data.map(() => {
            // console.log("30天预计", val);
            return {
                name: '',
                type: 'line',
                symbol: 'circle', // 默认是空心圆（中间是白色的），改成实心圆
                showSymbol: false,
                smooth: true,
                lineStyle: {
                    width: 1,
                    color: '#707070', // 线条颜色
                    borderColor: '#707070',
                },
                itemStyle: {
                    color: '#F5B348',
                    shadowColor: 'rgba(245, 179, 72, 0.3)',
                    shadowBlur: 3,
                },
                emphasis: {
                    scale: true,
                },
                areaStyle: {
                    //区域填充样式
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
                                color: '#846B38', // 0% 处的颜色
                            },
                            {
                                offset: 0.5,
                                color: '#403E47', // 0% 处的颜色
                            },
                            {
                                offset: 1,
                                color: '#11144E', // 100% 处的颜色
                            },
                        ],
                        global: false, // 缺省为 false
                    },

                    shadowColor: 'rgba(255, 199, 37, 0)', //阴影颜色
                    shadowBlur: 20, //shadowBlur设图形阴影的模糊大小。配合shadowColor,shadowOffsetX/Y, 设置图形的阴影效果。
                },
                data: data.data,
            };
        }),
    };
    charEch.setOption(option);
    return charEch;
};
// js 获取当前日期到之后一个月30天的日期区间
const initDate = (): string[] => {
    let dateList = [];
    let startDate = new Date();
    let endDate = new Date();
    endDate.setDate(startDate.getDate() + 30);
    startDate.setDate(startDate.getDate() + 1);
    while (endDate.getTime() - startDate.getTime() >= 0) {
        let month =
            (startDate.getMonth() + 1).toString().length === 1
                ? '0' + (startDate.getMonth() + 1).toString()
                : startDate.getMonth() + 1;
        let day =
            startDate.getDate().toString().length === 1
                ? '0' + startDate.getDate()
                : startDate.getDate();
        dateList.push(month + '/' + day);
        startDate.setDate(startDate.getDate() + 1);
    }
    return dateList;
};
defineExpose({
    initChart,
});
</script>
<style lang="scss" scoped>
.echarts {
    width: 100%;
    height: 100%;
    :deep(.lineChart-bg) {
        box-sizing: border-box;
        display: flex;
        align-items: center;
        width: 180px;
        height: 60px;
        padding-left: 20px;
        background: url('../images/line-bg.png') no-repeat;
        background-size: 100% 100%;
        span {
            font-size: 12px;
            color: rgb(255 255 255 / 80%);
            i {
                font-style: normal;
                color: #f5b348;
            }
        }
    }
}
</style>
