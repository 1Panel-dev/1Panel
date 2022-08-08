<template>
    <!-- 年度使用 -->
    <div class="echarts" id="AnnualUseChart"></div>
</template>

<script setup lang="ts">
// Echarts 为init（dom元素后的类型）
// EChartsOption 为 option 的类型
import { ECharts, EChartsOption, init } from 'echarts';
interface ChartProp {
    label: string;
    value: string[];
}
const initChart = (data: any = {}): ECharts => {
    const charEle = document.getElementById('AnnualUseChart') as HTMLElement;
    const gradientColors = [
        'rgba(254, 219, 101,0.1)',
        'rgba(0, 122, 254,0.1)',
        'rgba(255, 75, 122, 0.1)',
    ];
    const charEch: ECharts = init(charEle);
    const option: EChartsOption = {
        tooltip: {
            trigger: 'axis',
            axisPointer: {
                type: 'none',
            },
            borderWidth: 0, //边框线宽
            padding: 0,
            backgroundColor: 'transparent',
            formatter: (p: any) => {
                let str = '';
                p.forEach((val: any) => {
                    str += `
          <div class="year-item">
            <span class="year-dot" style="background-color: ${
                val.color
            };"></span>
            <span class="year-name">${val.seriesName}</span>
            <span class="year-value">${
                val.data >= 10000
                    ? (val.data / 10000).toFixed(2) + 'w'
                    : val.data
            }</span>
          </div>
          `;
                });
                let dom = `
                    <div class="annual-tooTip">
                      <span class="annual-month">${p[0].dataIndex + 1}月</span>
                      <div class="annual-list">
                        ${str}
                      </div>
                    </div>
                  `;
                return dom;
            },
        },

        legend: {
            right: '2%',
            top: '0%',
            itemWidth: 15,
            itemHeight: 6,
            align: 'auto',
            icon: 'rect',
            itemGap: 15,
            textStyle: {
                color: '#ebebf0',
            },
        },
        grid: {
            top: '20%',
            left: '40',
            right: '4%',
            bottom: '15%',
            // containLabel: true
        },
        xAxis: [
            {
                name: '(月份)',
                type: 'category',
                boundaryGap: false,
                axisLine: {
                    //坐标轴轴线相关设置。数学上的x轴
                    show: true,
                    lineStyle: {
                        color: '#233653',
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
                data: data.columns,
            },
        ],
        yAxis: {
            name: '(人数)',
            nameTextStyle: {
                color: '#D6DFEA',
                fontSize: 12,
                padding: [0, 30, 0, 0],
            },
            // nameGap:18,
            minInterval: 1,
            // min: 4,
            splitNumber: 5,
            splitLine: {
                show: false,
                lineStyle: {
                    color: '#192a44',
                },
            },
            axisLine: {
                show: true,
                lineStyle: {
                    color: '#233653',
                },
            },
            axisLabel: {
                show: true,
                color: '#B9D6D6',
                padding: 0,
                formatter: function (value: any) {
                    if (value >= 10000) {
                        value = value / 10000 + 'w';
                    }
                    return value;
                },
            },
            axisTick: {
                show: false,
            },
        },
        series: data.data.map((val: ChartProp, index: number) => {
            return {
                name: val.label,
                type: 'line',
                symbol: 'circle', // 默认是空心圆（中间是白色的），改成实心圆
                showSymbol: false,
                smooth: true,
                lineStyle: {
                    width: 1,
                    color: data.colors[index], // 线条颜色
                    borderColor: data.colors[index],
                },
                itemStyle: {
                    color: data.colors[index],
                    borderColor: '#646ace',
                    borderWidth: 2,
                },
                tooltip: {
                    show: true,
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
                                color: data.colors[index], // 0% 处的颜色
                            },
                            {
                                offset: 1,
                                color: gradientColors[index], // 100% 处的颜色
                            },
                        ],
                        global: false, // 缺省为 false
                    },
                    shadowColor: 'rgba(25,163,223, 0.3)', //阴影颜色
                    shadowBlur: 20, //shadowBlur设图形阴影的模糊大小。配合shadowColor,shadowOffsetX/Y, 设置图形的阴影效果。
                },
                data: val.value,
            };
        }),
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
:deep(.annual-tooTip) {
    box-sizing: border-box;
    width: 206px;
    height: 103px;
    padding: 5px 20px;
    background: url('../images/contrast-bg.png') no-repeat;
    background-size: 100% 100%;
    .annual-month {
        display: inline-block;
        margin-bottom: 2px;
        font-size: 10px;
        color: #03b8e2;
        transform: scale(0.9);
    }
    .annual-list {
        display: flex;
        flex-direction: column;
        width: 100%;
        .year-item {
            display: flex;
            align-items: center;
            width: 100%;
            height: 22px;
            .year-dot {
                width: 5px;
                height: 5px;
                margin: 0 2px;
                border-radius: 50%;
            }
            .year-name,
            .year-value {
                font-size: 10px;
                color: #03b8e2;
                transform: scale(0.8);
            }
            .year-name {
                margin: 0 2px;
            }
            .year-value {
                display: inline-block;
                width: 25%;
            }
        }
    }
}
</style>
