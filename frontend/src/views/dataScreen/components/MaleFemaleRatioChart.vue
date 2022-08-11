<template>
    <!-- 男女比例 -->
    <div class="malefemaleRatio-main">
        <div class="malefemaleRatio-header">
            <div class="man">
                <span>男士</span>
                <img src="../images/man.png" alt="" />
            </div>
            <div class="woman">
                <span>女士</span>
                <img src="../images/woman.png" alt="" />
            </div>
        </div>
        <!-- echarts -->
        <div class="echarts" id="MaleFemaleRatioChart"></div>
    </div>
</template>

<script setup lang="ts">
// Echarts 为init（dom元素后的类型）
// EChartsOption 为 option 的类型
import { ECharts, EChartsOption, init } from 'echarts';
interface ChartProp {
    man: number;
    woman: number;
}
const initChart = (data: ChartProp): ECharts => {
    /* 获取容器 */
    const charEle = document.getElementById('MaleFemaleRatioChart') as HTMLElement;
    /* 生成echarts */
    const charEch: ECharts = init(charEle);
    const option: EChartsOption = {
        xAxis: {
            type: 'value',
            show: false,
        },
        grid: {
            left: 0,
            top: '30px',
            bottom: 0,
            right: 0,
        },
        yAxis: [
            {
                type: 'category',
                position: 'left',
                data: ['男生'],
                axisTick: {
                    show: false,
                },
                axisLine: {
                    show: false,
                },
                axisLabel: {
                    show: false,
                },
            },
            {
                type: 'category',
                position: 'right',
                data: ['女士'],
                axisTick: {
                    show: false,
                },
                axisLine: {
                    show: false,
                },
                axisLabel: {
                    show: false,
                    padding: [0, 0, 40, -60],
                    fontSize: 12,
                    lineHeight: 60,
                    color: 'rgba(255, 255, 255, 0.9)',
                    formatter: '{value}' + data.woman * 100 + '%',
                    rich: {
                        a: {
                            color: 'transparent',
                            lineHeight: 30,
                            fontFamily: 'digital',
                            fontSize: 12,
                        },
                    },
                },
            },
        ],
        series: [
            {
                type: 'bar',
                barWidth: 20,
                data: [data.man],
                z: 20,
                itemStyle: {
                    borderRadius: 10,
                    color: '#007AFE',
                },
                label: {
                    show: true,
                    color: '#E7E8ED',
                    position: 'insideLeft',
                    offset: [0, -20],
                    fontSize: 12,
                    formatter: () => {
                        return `男士 ${data.man * 100}%`;
                    },
                },
            },
            {
                type: 'bar',
                barWidth: 20,
                data: [1],
                barGap: '-100%',
                itemStyle: {
                    borderRadius: 10,
                    color: '#FF4B7A',
                },
                label: {
                    show: true,
                    color: '#E7E8ED',
                    position: 'insideRight',
                    offset: [0, -20],
                    fontSize: 12,
                    formatter: () => {
                        return `女士 ${data.woman * 100}%`;
                    },
                },
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
.malefemaleRatio-main {
    box-sizing: border-box;
    width: 100%;
    height: 100%;
    padding: 40px 65px;
    .malefemaleRatio-header {
        display: flex;
        justify-content: space-between;
        width: 100%;
        height: 115px;
        .man,
        .woman {
            display: flex;
            flex-direction: column;
            align-items: center;
            width: 110px;
            height: 115px;
            background: url('../images/man-bg.png') no-repeat;
            background-size: 100% 100%;
            img {
                width: 60px;
                height: 60px;
                margin-top: 20px;
            }
            span {
                margin-top: 2px;
                font-size: 13px;
                color: #ffffff;
            }
        }
        .woman {
            background: url('../images/woman-bg.png') no-repeat;
        }
    }
    .echarts {
        width: 100%;
        height: calc(100% - 115px);
    }
}
</style>
