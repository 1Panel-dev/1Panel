<template>
    <div>
        <el-row :gutter="20">
            <el-col :span="24">
                <el-card style="overflow: inherit">
                    <template #header>
                        <span style="font-size: '30px'; font-weight: 500">平均负载</span>
                        <el-radio-group style="float: right" size="small" v-model="loadRadio">
                            <el-radio-button label="昨天" />
                            <el-radio-button label="今天" />
                            <el-radio-button label="最近 7 天" />
                            <el-radio-button label="最近 30 天" />
                            <el-radio-button label="自定义时间" />
                        </el-radio-group>
                    </template>
                    <div id="loadChart1" style="width: 100%; height: 300%"></div>
                </el-card>
            </el-col>
        </el-row>
        <el-row :gutter="20">
            <el-col :span="12">
                <el-card style="overflow: inherit">
                    <template #header>
                        <span style="font-size: '30px'; font-weight: 500">CPU</span>
                        <el-radio-group style="float: right" size="small" v-model="loadRadio">
                            <el-radio-button label="昨天" />
                            <el-radio-button label="今天" />
                            <el-radio-button label="最近 7 天" />
                            <el-radio-button label="最近 30 天" />
                            <el-radio-button label="自定义时间" />
                        </el-radio-group>
                    </template>
                    <div id="loadChart2" style="width: 100%; height: 300%"></div>
                </el-card>
            </el-col>
            <el-col :span="12">
                <el-card style="overflow: inherit">
                    <template #header>
                        <span style="font-size: '30px'; font-weight: 500">内存</span>
                        <el-radio-group style="float: right" size="small" v-model="loadRadio">
                            <el-radio-button label="昨天" />
                            <el-radio-button label="今天" />
                            <el-radio-button label="最近 7 天" />
                            <el-radio-button label="最近 30 天" />
                            <el-radio-button label="自定义时间" />
                        </el-radio-group>
                    </template>
                    <div id="loadChart3" style="width: 100%; height: 300%"></div>
                </el-card>
            </el-col>
        </el-row>
        <el-row :gutter="20">
            <el-col :span="12">
                <el-card style="overflow: inherit">
                    <template #header>
                        <span style="font-size: '30px'; font-weight: 500">磁盘IO</span>
                        <el-radio-group style="float: right" size="small" v-model="loadRadio">
                            <el-radio-button label="昨天" />
                            <el-radio-button label="今天" />
                            <el-radio-button label="最近 7 天" />
                            <el-radio-button label="最近 30 天" />
                            <el-radio-button label="自定义时间" />
                        </el-radio-group>
                    </template>
                    <div id="loadChart4" style="width: 100%; height: 300%"></div>
                </el-card>
            </el-col>
            <el-col :span="12">
                <el-card style="overflow: inherit">
                    <template #header>
                        <span style="font-size: '30px'; font-weight: 500">网络IO:</span>
                        <el-radio-group style="float: right" size="small" v-model="loadRadio">
                            <el-radio-button label="昨天" />
                            <el-radio-button label="今天" />
                            <el-radio-button label="最近 7 天" />
                            <el-radio-button label="最近 30 天" />
                            <el-radio-button label="自定义时间" />
                        </el-radio-group>
                    </template>
                    <div id="loadChart5" style="width: 100%; height: 300%"></div>
                </el-card>
            </el-col>
        </el-row>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import * as echarts from 'echarts';

const loadRadio = ref();

function initCharts(key: string) {
    console.log(key);
    const lineChart = echarts.init(document.getElementById(key) as HTMLElement);
    const option = {
        xAxis: {
            type: 'category',
            data: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
        },
        yAxis: {
            type: 'value',
        },
        series: [
            {
                data: [820, 932, 901, 934, 1290, 1330, 1320],
                type: 'line',
                smooth: true,
            },
        ],
    };
    lineChart.setOption(option, true);
}
onMounted(() => {
    for (let i = 1; i < 6; i++) {
        initCharts('loadChart' + i);
    }
});
</script>
