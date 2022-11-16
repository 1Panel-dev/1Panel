<template>
    <div>
        <el-row :gutter="20" class="row-box">
            <el-col :span="8">
                <el-card class="el-card">
                    <template #header>
                        <div class="card-header">
                            <span>概 览</span>
                        </div>
                    </template>
                    <el-row :gutter="20">
                        <el-col :span="12">
                            <el-card style="font-size: 12px; height: 80px; border-radius: 10px">
                                <svg-icon style="float: left; margin-left: 5px" iconName="p-website"></svg-icon>
                                <span style="float: left; margin-left: 5px; margin-top: 10px">网站</span>
                                <el-link
                                    style="float: right; font-size: 24px; margin-right: 5px"
                                    @click="goRouter('/websites')"
                                    type="primary"
                                >
                                    2
                                </el-link>
                            </el-card>
                        </el-col>
                        <el-col :span="12">
                            <el-card style="font-size: 12px; height: 80px; border-radius: 10px">
                                <svg-icon style="float: left; margin-left: 5px" iconName="p-database"></svg-icon>
                                <span style="float: left; margin-left: 5px; margin-top: 10px">数据库</span>
                                <el-link
                                    style="float: right; font-size: 24px; margin-right: 5px"
                                    @click="goRouter('/databases')"
                                    type="primary"
                                >
                                    5
                                </el-link>
                            </el-card>
                        </el-col>
                    </el-row>
                    <el-row :gutter="20" style="margin-top: 20px; margin-top: 30px">
                        <el-col :span="12">
                            <el-card style="font-size: 12px; height: 80px; border-radius: 10px">
                                <svg-icon style="float: left; margin-left: 5px" iconName="p-plan"></svg-icon>
                                <span style="float: left; margin-left: 5px; margin-top: 10px">定时任务</span>
                                <el-link
                                    style="float: right; font-size: 24px; margin-right: 5px"
                                    @click="goRouter('/cronjobs')"
                                    type="primary"
                                >
                                    7
                                </el-link>
                            </el-card>
                        </el-col>
                        <el-col :span="12">
                            <el-card style="font-size: 12px; height: 80px; border-radius: 10px">
                                <svg-icon style="float: left; margin-left: 5px" iconName="p-appstore"></svg-icon>
                                <span style="float: left; margin-left: 5px; margin-top: 10px">已安装应用</span>
                                <el-link
                                    style="float: right; font-size: 24px; margin-right: 5px"
                                    @click="goRouter('/apps')"
                                    type="primary"
                                >
                                    3
                                </el-link>
                            </el-card>
                        </el-col>
                    </el-row>
                </el-card>
            </el-col>
            <el-col :span="8">
                <el-card class="el-card">
                    <template #header>
                        <div class="card-header">
                            <span>状态</span>
                        </div>
                    </template>
                    <el-row :gutter="10">
                        <el-col :span="12" align="center">
                            <el-progress type="dashboard" :width="80" :percentage="80">
                                <template #default="{ percentage }">
                                    <span class="percentage-value">{{ percentage }}%</span>
                                    <span class="percentage-label">CPU</span>
                                </template>
                            </el-progress>
                            <br />
                            <span>(0.56 / 8.00) Core</span>
                        </el-col>
                        <el-col :span="12" align="center">
                            <el-progress type="dashboard" :width="80" :percentage="30">
                                <template #default="{ percentage }">
                                    <span class="percentage-value">{{ percentage }}%</span>
                                    <span class="percentage-label">内存</span>
                                </template>
                            </el-progress>
                            <br />
                            <span>(851 / 7812) MB</span>
                        </el-col>
                    </el-row>
                    <el-row :gutter="10" style="margin-top: 30px">
                        <el-col :span="12" align="center">
                            <el-progress type="dashboard" :width="80" :percentage="50">
                                <template #default="{ percentage }">
                                    <span class="percentage-value">{{ percentage }}%</span>
                                    <span class="percentage-label">负载</span>
                                </template>
                            </el-progress>
                            <br />
                            <span>(0.36 / 8.00) Core</span>
                        </el-col>
                        <el-col :span="12" align="center">
                            <el-progress type="dashboard" :width="80" :percentage="40">
                                <template #default="{ percentage }">
                                    <span class="percentage-value">{{ percentage }}%</span>
                                    <span class="percentage-label">磁盘</span>
                                </template>
                            </el-progress>
                            <br />
                            <span>(6.23 / 46.97) GB</span>
                        </el-col>
                    </el-row>
                </el-card>
            </el-col>
            <el-col :span="8">
                <el-card class="el-card">
                    <template #header>
                        <div class="card-header">
                            <span>系统信息</span>
                        </div>
                    </template>
                    <el-form>
                        <el-form-item label="主机名称">ko-deploy</el-form-item>
                        <el-form-item label="发行版本">centos-7.6.1810</el-form-item>
                        <el-form-item label="内核版本">
                            Linux version 3.10.0-957.el7.x86_64 (mockbuild@kbuilder.bsys.centos.org) (gcc version 4.8.5
                            20150623 (Red Hat 4.8.5-36) (GCC) ) #1 SMP Thu Nov 8 23:39:32 UTC 2018
                        </el-form-item>
                        <el-form-item label="系统类型">x86_64</el-form-item>
                    </el-form>
                </el-card>
            </el-col>
        </el-row>
        <el-row :gutter="20" style="margin-top: 20px" class="row-box">
            <el-col :span="12">
                <el-card class="el-card">
                    <template #header>
                        <div class="card-header">
                            <span>应用</span>
                        </div>
                    </template>
                    <el-row :gutter="20">
                        <el-col :span="12">
                            <el-card @mouseover="hoverApp = 'halo'" @mouseleave="hoverApp = ''" style="height: 110px">
                                <el-row>
                                    <el-col :span="4">
                                        <img style="width: 40px; height: 40px" src="./images/halo.jpg" alt="" />
                                    </el-col>
                                    <el-col :span="2"><br /></el-col>
                                    <el-col :span="18">
                                        <span style="font-size: 20px; color: #214bc8">Halo</span>
                                        <div><span class="input-help">现代化开源建站 / CMS系统</span></div>
                                    </el-col>
                                </el-row>
                                <div v-show="hoverApp === 'halo'" style="float: right; margin-top: 10px">
                                    <el-button>停止</el-button>
                                    <el-button>重启</el-button>
                                </div>
                            </el-card>
                        </el-col>
                        <el-col :span="12">
                            <el-card @mouseover="hoverApp = 'de'" @mouseleave="hoverApp = ''" style="height: 110px">
                                <el-row>
                                    <el-col :span="4">
                                        <img style="width: 40px; height: 40px" src="./images/de.jpg" alt="" />
                                    </el-col>
                                    <el-col :span="2"><br /></el-col>
                                    <el-col :span="18">
                                        <span style="font-size: 20px; color: #0070d6">Dataease</span>
                                        <div><span class="input-help">开源数据可视化分析工具</span></div>
                                    </el-col>
                                </el-row>
                                <div v-show="hoverApp === 'de'" style="float: right; margin-top: 10px">
                                    <el-button>启动</el-button>
                                </div>
                            </el-card>
                        </el-col>
                    </el-row>
                    <el-row :gutter="20" style="margin-top: 20px">
                        <el-col :span="12">
                            <el-card @mouseover="hoverApp = 'js'" @mouseleave="hoverApp = ''" style="height: 110px">
                                <el-row>
                                    <el-col :span="4">
                                        <img style="width: 40px; height: 40px" src="./images/js.jpg" alt="" />
                                    </el-col>
                                    <el-col :span="2"><br /></el-col>
                                    <el-col :span="18">
                                        <span style="font-size: 16px; color: #008d75">JumpServer</span>
                                        <div><span class="input-help">广受欢迎的开源堡垒机</span></div>
                                    </el-col>
                                </el-row>
                                <div v-show="hoverApp === 'js'" style="float: right; margin-top: 10px">
                                    <el-button>停止</el-button>
                                    <el-button>重启</el-button>
                                </div>
                            </el-card>
                        </el-col>
                        <el-col :span="12">
                            <el-card @mouseover="hoverApp = 'ms'" @mouseleave="hoverApp = ''" style="height: 110px">
                                <el-row>
                                    <el-col :span="4">
                                        <img style="width: 40px; height: 40px" src="./images/ms.jpg" alt="" />
                                    </el-col>
                                    <el-col :span="2"><br /></el-col>
                                    <el-col :span="18">
                                        <span style="font-size: 16px; color: #723279">MeterSphere</span>
                                        <div><span class="input-help">一站式开源持续测试平台</span></div>
                                    </el-col>
                                </el-row>
                                <div v-show="hoverApp === 'ms'" style="float: right; margin-top: 10px">
                                    <el-button>启动</el-button>
                                </div>
                            </el-card>
                        </el-col>
                    </el-row>
                </el-card>
            </el-col>
            <el-col :span="12">
                <el-card class="el-card">
                    <el-radio-group v-model="chartOption" @change="changeOption">
                        <el-radio-button label="io">流量</el-radio-button>
                        <el-radio-button label="network">网络</el-radio-button>
                    </el-radio-group>
                    <div v-if="chartOption === 'io'" id="ioChart" style="width: 100%; height: 320px"></div>
                    <div v-if="chartOption === 'network'" id="networkChart" style="width: 100%; height: 320px"></div>
                </el-card>
            </el-col>
        </el-row>
    </div>
</template>

<script setup lang="ts" name="dataVisualize">
import { onMounted, onBeforeUnmount, ref } from 'vue';
import * as echarts from 'echarts';
import i18n from '@/lang';
import { ContainerStats } from '@/api/modules/container';
import { dateFromatForSecond } from '@/utils/util';
import { useRouter } from 'vue-router';
const router = useRouter();

const hoverApp = ref();
const chartOption = ref('io');
let timer: NodeJS.Timer | null = null;
let isInit = ref<boolean>(true);

const ioReadDatas = ref<Array<string>>([]);
const ioWriteDatas = ref<Array<string>>([]);
const netTxDatas = ref<Array<string>>([]);
const netRxDatas = ref<Array<string>>([]);
const timeDatas = ref<Array<string>>([]);

const changeOption = async () => {
    isInit.value = true;
    loadData();
};

const goRouter = async (path: string) => {
    router.push({ path: path });
};

const loadData = async () => {
    try {
        const res = await ContainerStats('kubeoperator_nginx');
        ioReadDatas.value.push(res.data.ioRead.toFixed(2));
        if (ioReadDatas.value.length > 20) {
            ioReadDatas.value.splice(0, 1);
        }
        ioWriteDatas.value.push(res.data.ioWrite.toFixed(2));
        if (ioWriteDatas.value.length > 20) {
            ioWriteDatas.value.splice(0, 1);
        }
        netTxDatas.value.push(res.data.networkTX.toFixed(2));
        if (netTxDatas.value.length > 20) {
            netTxDatas.value.splice(0, 1);
        }
        netRxDatas.value.push(res.data.networkRX.toFixed(2));
        if (netRxDatas.value.length > 20) {
            netRxDatas.value.splice(0, 1);
        }
        timeDatas.value.push(dateFromatForSecond(res.data.shotTime));
        if (timeDatas.value.length > 20) {
            timeDatas.value.splice(0, 1);
        }
        if (chartOption.value === 'io') {
            let ioReadYDatas = {
                name: i18n.global.t('monitor.read'),
                type: 'line',
                areaStyle: {
                    color: '#ebdee3',
                },
                data: ioReadDatas.value,
                showSymbol: false,
            };
            let ioWriteYDatas = {
                name: i18n.global.t('monitor.write'),
                type: 'line',
                areaStyle: {
                    color: '#ebdee3',
                },
                data: ioWriteDatas.value,
                showSymbol: false,
            };
            freshChart(
                'ioChart',
                [i18n.global.t('monitor.read'), i18n.global.t('monitor.write')],
                timeDatas.value,
                [ioReadYDatas, ioWriteYDatas],
                '流量',
                'MB',
            );
        } else {
            let netTxYDatas = {
                name: i18n.global.t('monitor.up'),
                type: 'line',
                areaStyle: {
                    color: '#ebdee3',
                },
                data: netTxDatas.value,
                showSymbol: false,
            };
            let netRxYDatas = {
                name: i18n.global.t('monitor.down'),
                type: 'line',
                areaStyle: {
                    color: '#ebdee3',
                },
                data: netRxDatas.value,
                showSymbol: false,
            };
            freshChart(
                'networkChart',
                [i18n.global.t('monitor.up'), i18n.global.t('monitor.down')],
                timeDatas.value,
                [netTxYDatas, netRxYDatas],
                i18n.global.t('monitor.network'),
                'KB/s',
            );
        }
    } catch {
        clearInterval(Number(timer));
        timer = null;
    }
};

function freshChart(chartName: string, legendDatas: any, xDatas: any, yDatas: any, yTitle: string, formatStr: string) {
    console.log(chartName, echarts.getInstanceByDom(document.getElementById(chartName) as HTMLElement));
    if (isInit.value) {
        echarts.init(document.getElementById(chartName) as HTMLElement);
        isInit.value = false;
    }
    let itemChart = echarts.getInstanceByDom(document.getElementById(chartName) as HTMLElement);
    const option = {
        title: [
            {
                left: 'center',
                text: yTitle,
            },
        ],
        zlevel: 1,
        z: 1,
        tooltip: {
            trigger: 'axis',
            formatter: function (datas: any) {
                let res = datas[0].name + '<br/>';
                for (const item of datas) {
                    res += item.marker + ' ' + item.seriesName + '：' + item.data + formatStr + '<br/>';
                }
                return res;
            },
        },
        grid: { left: '7%', right: '7%', bottom: '20%' },
        legend: {
            data: legendDatas,
            right: 10,
        },
        xAxis: { data: xDatas, boundaryGap: false },
        yAxis: { name: '( ' + formatStr + ' )' },
        series: yDatas,
    };
    itemChart?.setOption(option, true);
}

function changeChartSize() {
    echarts.getInstanceByDom(document.getElementById('ioChart') as HTMLElement)?.resize();
    echarts.getInstanceByDom(document.getElementById('networkChart') as HTMLElement)?.resize();
}

onMounted(() => {
    loadData();
    window.addEventListener('resize', changeChartSize);
    timer = setInterval(async () => {
        loadData();
    }, 3000);
});

onBeforeUnmount(() => {
    clearInterval(Number(timer));
    timer = null;
    window.removeEventListener('resize', changeChartSize);
});
</script>

<style scoped>
.percentage-value {
    display: block;
    font-size: 16px;
}
.percentage-label {
    display: block;
    margin-top: 10px;
    font-size: 12px;
}
.card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
}
</style>
