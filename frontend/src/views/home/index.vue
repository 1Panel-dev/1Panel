<template>
    <div>
        <RouterButton
            :buttons="[
                {
                    label: i18n.global.t('menu.home'),
                    path: '/',
                },
            ]"
        />
        <el-row :gutter="20" style="margin-top: 20px">
            <el-col :span="16">
                <CardWithHeader :header="$t('home.overview')" height="146px">
                    <template #body>
                        <div class="h-overview">
                            <el-row>
                                <el-col :span="6">
                                    <span>{{ $t('menu.website') }}</span>
                                    <div class="count">
                                        <span @click="goRouter('/websites')">{{ baseInfo?.websiteNumber }}</span>
                                    </div>
                                </el-col>
                                <el-col :span="6">
                                    <span>{{ $t('menu.database') }}</span>
                                    <div class="count">
                                        <span @click="goRouter('/databases')">{{ baseInfo?.databaseNumber }}</span>
                                    </div>
                                </el-col>
                                <el-col :span="6">
                                    <span>{{ $t('menu.cronjob') }}</span>
                                    <div class="count">
                                        <span @click="goRouter('/cronjobs')">
                                            {{ baseInfo?.cronjobNumber }}
                                        </span>
                                    </div>
                                </el-col>
                                <el-col :span="6">
                                    <span>{{ $t('home.appInstalled') }}</span>
                                    <div class="count">
                                        <span @click="goRouter('/apps/installed')">
                                            {{ baseInfo?.appInstalldNumber }}
                                        </span>
                                    </div>
                                </el-col>
                            </el-row>
                        </div>
                    </template>
                </CardWithHeader>
                <CardWithHeader :header="$t('commons.table.status')" style="margin-top: 20px">
                    <template #body>
                        <Status ref="statuRef" style="margin-top: -7px" />
                    </template>
                </CardWithHeader>
                <CardWithHeader :header="$t('menu.monitor')" style="margin-top: 20px">
                    <template #header-r>
                        <el-radio-group
                            style="float: right; margin-left: 5px"
                            v-model="chartOption"
                            @change="changeOption"
                        >
                            <el-radio-button label="network">{{ $t('home.network') }}</el-radio-button>
                            <el-radio-button label="io">{{ $t('home.io') }}</el-radio-button>
                        </el-radio-group>
                        <el-select
                            v-if="chartOption === 'network'"
                            @change="onLoadBaseInfo(false, 'network')"
                            v-model="searchInfo.netOption"
                            style="float: right"
                        >
                            <template #prefix>{{ $t('home.networkCard') }}</template>
                            <el-option
                                v-for="item in netOptions"
                                :key="item"
                                :label="item == 'all' ? $t('commons.table.all') : item"
                                :value="item"
                            />
                        </el-select>
                        <el-select
                            v-if="chartOption === 'io'"
                            v-model="searchInfo.ioOption"
                            @change="onLoadBaseInfo(false, 'io')"
                            style="float: right"
                        >
                            <template #prefix>{{ $t('home.disk') }}</template>
                            <el-option
                                v-for="item in ioOptions"
                                :key="item"
                                :label="item == 'all' ? $t('commons.table.all') : item"
                                :value="item"
                            />
                        </el-select>
                    </template>
                    <template #body>
                        <div style="position: relative; margin-top: 20px">
                            <div class="monitor-tags" v-if="chartOption === 'network'">
                                <el-tag>{{ $t('monitor.up') }}: {{ currentChartInfo.netBytesSent }} KB/s</el-tag>
                                <el-tag>{{ $t('monitor.down') }}: {{ currentChartInfo.netBytesRecv }} KB/s</el-tag>
                                <el-tag>{{ $t('home.totalSend') }}: {{ computeSize(currentInfo.netBytesSent) }}</el-tag>
                                <el-tag>{{ $t('home.totalRecv') }}: {{ computeSize(currentInfo.netBytesRecv) }}</el-tag>
                            </div>
                            <div class="monitor-tags" v-if="chartOption === 'io'">
                                <el-tag>{{ $t('monitor.read') }}: {{ currentChartInfo.ioReadBytes }} MB</el-tag>
                                <el-tag>{{ $t('monitor.write') }}: {{ currentChartInfo.ioWriteBytes }} MB</el-tag>
                                <el-tag>
                                    {{ $t('home.rwPerSecond') }}: {{ currentChartInfo.ioCount }} {{ $t('home.time') }}
                                </el-tag>
                                <el-tag>{{ $t('home.ioDelay') }}: {{ currentChartInfo.ioTime }} ms</el-tag>
                            </div>

                            <div v-if="chartOption === 'io'" style="margin-top: 40px">
                                <v-charts
                                    height="300px"
                                    id="ioChart"
                                    type="line"
                                    :option="chartsOption['ioChart']"
                                    v-if="chartsOption['ioChart']"
                                    :dataZoom="true"
                                />
                            </div>
                            <div v-if="chartOption === 'network'" style="margin-top: 40px">
                                <v-charts
                                    height="300px"
                                    id="networkChart"
                                    type="line"
                                    :option="chartsOption['networkChart']"
                                    v-if="chartsOption['networkChart']"
                                    :dataZoom="true"
                                />
                            </div>
                        </div>
                    </template>
                </CardWithHeader>
            </el-col>
            <el-col :span="8">
                <CardWithHeader :header="$t('home.systemInfo')">
                    <template #body>
                        <el-descriptions :column="1" class="h-systemInfo">
                            <el-descriptions-item class-name="system-content">
                                <template #label>
                                    <span class="system-label">
                                        {{ $t('home.hostname') }}
                                    </span>
                                </template>
                                {{ baseInfo.hostname }}
                            </el-descriptions-item>
                            <el-descriptions-item class-name="system-content">
                                <template #label>
                                    <span class="system-label">
                                        {{ $t('home.platformVersion') }}
                                    </span>
                                </template>
                                {{ baseInfo.platform }}-{{ baseInfo.platformVersion }}
                            </el-descriptions-item>
                            <el-descriptions-item class-name="system-content">
                                <template #label>
                                    <span class="system-label">
                                        {{ $t('home.kernelVersion') }}
                                    </span>
                                </template>
                                {{ baseInfo.kernelVersion }}
                            </el-descriptions-item>
                            <el-descriptions-item class-name="system-content">
                                <template #label>
                                    <span class="system-label">
                                        {{ $t('home.kernelArch') }}
                                    </span>
                                </template>
                                {{ baseInfo.kernelArch }}
                            </el-descriptions-item>
                            <el-descriptions-item class-name="system-content">
                                <template #label>
                                    <span class="system-label">
                                        {{ $t('home.uptime') }}
                                    </span>
                                </template>
                                {{ currentInfo.timeSinceUptime }}
                            </el-descriptions-item>
                            <el-descriptions-item class-name="system-content">
                                <template #label>
                                    <span class="system-label">
                                        {{ $t('home.runningTime') }}
                                    </span>
                                </template>
                                {{ loadUpTime(currentInfo.uptime) }}
                            </el-descriptions-item>
                        </el-descriptions>
                    </template>
                </CardWithHeader>

                <CardWithHeader :header="$t('home.app')" style="margin-top: 20px">
                    <template #body>
                        <App ref="appRef" />
                    </template>
                </CardWithHeader>
            </el-col>
        </el-row>
    </div>
</template>

<script lang="ts" setup>
import { onMounted, onBeforeUnmount, ref, reactive } from 'vue';
import Status from '@/views/home/status/index.vue';
import App from '@/views/home/app/index.vue';
import VCharts from '@/components/v-charts/index.vue';
import CardWithHeader from '@/components/card-with-header/index.vue';
import i18n from '@/lang';
import { Dashboard } from '@/api/interface/dashboard';
import { dateFormatForSecond, computeSize } from '@/utils/util';
import { useRouter } from 'vue-router';
import RouterButton from '@/components/router-button/index.vue';
import { loadBaseInfo, loadCurrentInfo } from '@/api/modules/dashboard';
import { getIOOptions, getNetworkOptions } from '@/api/modules/monitor';
const router = useRouter();

const statuRef = ref();
const appRef = ref();

const chartOption = ref('network');
let timer: NodeJS.Timer | null = null;
let isInit = ref<boolean>(true);

const ioReadBytes = ref<Array<number>>([]);
const ioWriteBytes = ref<Array<number>>([]);
const netBytesSents = ref<Array<number>>([]);
const netBytesRecvs = ref<Array<number>>([]);
const timeIODatas = ref<Array<string>>([]);
const timeNetDatas = ref<Array<string>>([]);

const ioOptions = ref();
const netOptions = ref();
const searchInfo = reactive({
    ioOption: 'all',
    netOption: 'all',
});

const baseInfo = ref<Dashboard.BaseInfo>({
    haloID: 0,
    dateeaseID: 0,
    jumpserverID: 0,
    metersphereID: 0,
    kubeoperatorID: 0,
    kubepiID: 0,

    websiteNumber: 0,
    databaseNumber: 0,
    cronjobNumber: 0,
    appInstalldNumber: 0,

    hostname: '',
    os: '',
    platform: '',
    platformFamily: '',
    platformVersion: '',
    kernelArch: '',
    kernelVersion: '',
    virtualizationSystem: '',

    cpuCores: 0,
    cpuLogicalCores: 0,
    cpuModelName: '',
    currentInfo: null,
});
const currentInfo = ref<Dashboard.CurrentInfo>({
    uptime: 0,
    timeSinceUptime: '',
    procs: 0,

    load1: 0,
    load5: 0,
    load15: 0,
    loadUsagePercent: 0,

    cpuPercent: [] as Array<number>,
    cpuUsedPercent: 0,
    cpuUsed: 0,
    cpuTotal: 0,

    memoryTotal: 0,
    memoryAvailable: 0,
    memoryUsed: 0,
    MemoryUsedPercent: 0,

    ioReadBytes: 0,
    ioWriteBytes: 0,
    ioCount: 0,
    ioReadTime: 0,
    ioWriteTime: 0,

    diskData: [],

    netBytesSent: 0,
    netBytesRecv: 0,

    shotTime: new Date(),
});
const currentChartInfo = reactive({
    ioReadBytes: 0,
    ioWriteBytes: 0,
    ioCount: 0,
    ioTime: 0,

    netBytesSent: 0,
    netBytesRecv: 0,
});

const chartsOption = ref({ ioChart1: null, networkChart: null });

const changeOption = async () => {
    isInit.value = true;
    loadData();
};

const goRouter = async (path: string) => {
    router.push({ path: path });
};

const onLoadNetworkOptions = async () => {
    const res = await getNetworkOptions();
    netOptions.value = res.data;
    searchInfo.netOption = netOptions.value && netOptions.value[0];
};

const onLoadIOOptions = async () => {
    const res = await getIOOptions();
    ioOptions.value = res.data;
    searchInfo.ioOption = ioOptions.value && ioOptions.value[0];
};

const onLoadBaseInfo = async (isInit: boolean, range: string) => {
    if (range === 'all' || range === 'io') {
        ioReadBytes.value = [];
        ioWriteBytes.value = [];
        timeIODatas.value = [];
    } else if (range === 'all' || range === 'network') {
        netBytesSents.value = [];
        netBytesRecvs.value = [];
        timeNetDatas.value = [];
    }
    const res = await loadBaseInfo(searchInfo.ioOption, searchInfo.netOption);
    baseInfo.value = res.data;
    currentInfo.value = baseInfo.value.currentInfo;
    onLoadCurrentInfo();
    statuRef.value.acceptParams(currentInfo.value, baseInfo.value);
    appRef.value.acceptParams(baseInfo.value);
    if (isInit) {
        // window.addEventListener('resize', changeChartSize);
        timer = setInterval(async () => {
            onLoadCurrentInfo();
        }, 3000);
    }
};

const onLoadCurrentInfo = async () => {
    const res = await loadCurrentInfo(searchInfo.ioOption, searchInfo.netOption);
    currentInfo.value.timeSinceUptime = res.data.timeSinceUptime;

    currentChartInfo.netBytesSent = Number(
        ((res.data.netBytesSent - currentInfo.value.netBytesSent) / 1024 / 3).toFixed(2),
    );
    netBytesSents.value.push(currentChartInfo.netBytesSent);
    if (netBytesSents.value.length > 20) {
        netBytesSents.value.splice(0, 1);
    }
    currentChartInfo.netBytesRecv = Number(
        ((res.data.netBytesRecv - currentInfo.value.netBytesRecv) / 1024 / 3).toFixed(2),
    );
    netBytesRecvs.value.push(currentChartInfo.netBytesRecv);
    if (netBytesRecvs.value.length > 20) {
        netBytesRecvs.value.splice(0, 1);
    }

    currentChartInfo.ioReadBytes = Number(
        ((res.data.ioReadBytes - currentInfo.value.ioReadBytes) / 1024 / 1024 / 3).toFixed(2),
    );
    ioReadBytes.value.push(currentChartInfo.ioReadBytes);
    if (ioReadBytes.value.length > 20) {
        ioReadBytes.value.splice(0, 1);
    }
    currentChartInfo.ioWriteBytes = Number(
        ((res.data.ioWriteBytes - currentInfo.value.ioWriteBytes) / 1024 / 1024 / 3).toFixed(2),
    );
    ioWriteBytes.value.push(currentChartInfo.ioWriteBytes);
    if (ioWriteBytes.value.length > 20) {
        ioWriteBytes.value.splice(0, 1);
    }
    currentChartInfo.ioCount = Math.round(Number((res.data.ioCount - currentInfo.value.ioCount) / 3));
    let ioReadTime = res.data.ioReadTime - currentInfo.value.ioReadTime;
    let ioWriteTime = res.data.ioWriteTime - currentInfo.value.ioWriteTime;
    let ioChoose = ioReadTime > ioWriteTime ? ioReadTime : ioWriteTime;
    currentChartInfo.ioTime = Math.round(Number(ioChoose / 3));

    timeIODatas.value.push(dateFormatForSecond(res.data.shotTime));
    if (timeIODatas.value.length > 20) {
        timeIODatas.value.splice(0, 1);
    }
    timeNetDatas.value.push(dateFormatForSecond(res.data.shotTime));
    if (timeNetDatas.value.length > 20) {
        timeNetDatas.value.splice(0, 1);
    }
    loadData();
    currentInfo.value = res.data;
    statuRef.value.acceptParams(currentInfo.value, baseInfo.value);
};

function loadUpTime(uptime: number) {
    if (uptime <= 0) {
        return '-';
    }
    let days = Math.floor(uptime / 86400);
    let hours = Math.floor((uptime % 86400) / 3600);
    let minutes = Math.floor((uptime % 3600) / 60);
    let seconds = uptime % 60;
    if (days !== 0) {
        return (
            days +
            i18n.global.t('home.Day') +
            ' ' +
            hours +
            i18n.global.t('home.Hour') +
            ' ' +
            minutes +
            i18n.global.t('home.Minute') +
            ' ' +
            seconds +
            i18n.global.t('home.Second')
        );
    }
    if (hours !== 0) {
        return (
            hours +
            i18n.global.t('home.Hour') +
            ' ' +
            minutes +
            i18n.global.t('home.Minute') +
            ' ' +
            seconds +
            i18n.global.t('home.Second')
        );
    }
    if (minutes !== 0) {
        return minutes + i18n.global.t('home.Minute') + ' ' + seconds + i18n.global.t('home.Second');
    }
    return seconds + i18n.global.t('home.Second');
}

const loadData = async () => {
    if (chartOption.value === 'io') {
        chartsOption.value['ioChart'] = {
            xDatas: timeIODatas.value,
            yDatas: [
                {
                    name: i18n.global.t('monitor.read'),
                    data: ioReadBytes.value,
                },
                {
                    name: i18n.global.t('monitor.write'),
                    data: ioWriteBytes.value,
                },
            ],
            formatStr: 'MB',
        };
    } else {
        chartsOption.value['networkChart'] = {
            xDatas: timeNetDatas.value,
            yDatas: [
                {
                    name: i18n.global.t('monitor.up'),
                    data: netBytesRecvs.value,
                },
                {
                    name: i18n.global.t('monitor.down'),
                    data: netBytesSents.value,
                },
            ],
            formatStr: 'KB/s',
        };
    }
};

onMounted(() => {
    onLoadNetworkOptions();
    onLoadIOOptions();
    onLoadBaseInfo(true, 'all');
});

onBeforeUnmount(() => {
    clearInterval(Number(timer));
    timer = null;
});
</script>

<style lang="scss">
.h-overview {
    text-align: center;

    span:first-child {
        font-size: 14px;
        color: var(--el-text-color-regular);
    }

    .count {
        margin-top: 10px;
        span {
            font-size: 25px;
            color: $primary-color;
            font-weight: 500;
            line-height: 32px;
            cursor: pointer;
        }
    }
}

.h-systemInfo {
    margin-left: 18px;
}

.system-label {
    font-weight: 400 !important;
    font-size: 14px !important;
    color: #1f2329;
}

.system-content {
    font-size: 13px !important;
}

.monitor-tags {
    position: absolute;
    top: -10px;
    left: 20px;

    .el-tag {
        margin-right: 10px;
        margin-bottom: 10px;
    }
}
</style>
