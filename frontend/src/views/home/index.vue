<template>
    <div>
        <RouterButton
            :buttons="[
                {
                    label: i18n.global.t('menu.home'),
                    path: '/',
                },
            ]"
        >
            <template #route-button>
                <div class="router-button" v-if="!isProductPro">
                    <el-button link type="primary" @click="toUpload">
                        {{ $t('license.levelUpPro') }}
                    </el-button>
                </div>
            </template>
        </RouterButton>

        <el-alert
            v-if="!isSafety && globalStore.showEntranceWarn"
            style="margin-top: 20px"
            type="warning"
            @close="hideEntrance"
        >
            <template #title>
                <span class="flx-align-center">
                    <span>{{ $t('home.entranceHelper') }}</span>
                    <el-link
                        style="font-size: 12px; margin-left: 5px"
                        icon="Position"
                        @click="goRouter('/settings/safe')"
                        type="primary"
                    >
                        {{ $t('firewall.quickJump') }}
                    </el-link>
                </span>
            </template>
        </el-alert>

        <el-row :gutter="20" style="margin-top: 20px">
            <el-col :xs="24" :sm="24" :md="16" :lg="16" :xl="16">
                <CardWithHeader :header="$t('home.overview')" height="166px">
                    <template #body>
                        <div class="h-overview">
                            <el-row>
                                <el-col :span="6">
                                    <span>{{ $t('menu.website', 2) }}</span>
                                    <div class="count">
                                        <span @click="goRouter('/websites')">{{ baseInfo?.websiteNumber }}</span>
                                    </div>
                                </el-col>
                                <el-col :span="6">
                                    <span>{{ $t('menu.database', 2) }} - {{ $t('database.all') }}</span>
                                    <div class="count">
                                        <span @click="goRouter('/databases')">{{ baseInfo?.databaseNumber }}</span>
                                    </div>
                                </el-col>
                                <el-col :span="6">
                                    <span>{{ $t('menu.cronjob', 2) }}</span>
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
                                            {{ baseInfo?.appInstalledNumber }}
                                        </span>
                                    </div>
                                </el-col>
                            </el-row>
                        </div>
                    </template>
                </CardWithHeader>
                <CardWithHeader :header="$t('commons.table.status')" style="margin-top: 20px">
                    <template #body>
                        <Status ref="statusRef" style="margin-bottom: 33px" />
                    </template>
                </CardWithHeader>
                <CardWithHeader :header="$t('menu.monitor')" style="margin-top: 20px; margin-bottom: 20px">
                    <template #header-r>
                        <el-radio-group
                            style="float: right; margin-left: 5px"
                            v-model="chartOption"
                            @change="changeOption"
                        >
                            <el-radio-button value="network">{{ $t('home.network') }}</el-radio-button>
                            <el-radio-button value="io">{{ $t('home.io') }}</el-radio-button>
                        </el-radio-group>
                        <el-select
                            v-if="chartOption === 'network'"
                            @change="onLoadBaseInfo(false, 'network')"
                            v-model="searchInfo.netOption"
                            class="p-w-200 float-right"
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
                            class="p-w-200 float-right"
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
                        <div style="position: relative; margin-top: 60px">
                            <div class="monitor-tags" v-if="chartOption === 'network'">
                                <el-tag>
                                    {{ $t('monitor.up') }}: {{ computeSizeFromKBs(currentChartInfo.netBytesSent) }}
                                </el-tag>
                                <el-tag>
                                    {{ $t('monitor.down') }}: {{ computeSizeFromKBs(currentChartInfo.netBytesRecv) }}
                                </el-tag>
                                <el-tag>{{ $t('home.totalSend') }}: {{ computeSize(currentInfo.netBytesSent) }}</el-tag>
                                <el-tag>{{ $t('home.totalRecv') }}: {{ computeSize(currentInfo.netBytesRecv) }}</el-tag>
                            </div>
                            <div class="monitor-tags" v-if="chartOption === 'io'">
                                <el-tag>{{ $t('monitor.read') }}: {{ currentChartInfo.ioReadBytes }} MB</el-tag>
                                <el-tag>{{ $t('monitor.write') }}: {{ currentChartInfo.ioWriteBytes }} MB</el-tag>
                                <el-tag>
                                    {{ $t('home.rwPerSecond') }}: {{ currentChartInfo.ioCount }}
                                    {{ $t('commons.units.time') }}/s
                                </el-tag>
                                <el-tag>{{ $t('home.ioDelay') }}: {{ currentChartInfo.ioTime }} ms</el-tag>
                            </div>

                            <div v-if="chartOption === 'io'" style="margin-top: 40px" class="mobile-monitor-chart">
                                <v-charts
                                    height="383px"
                                    id="ioChart"
                                    type="line"
                                    :option="chartsOption['ioChart']"
                                    v-if="chartsOption['ioChart']"
                                    :dataZoom="true"
                                />
                            </div>
                            <div v-if="chartOption === 'network'" style="margin-top: 40px" class="mobile-monitor-chart">
                                <v-charts
                                    height="383px"
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
            <el-col :xs="24" :sm="24" :md="8" :lg="8" :xl="8">
                <CardWithHeader :header="$t('home.systemInfo')">
                    <template #body>
                        <div class="h-systemInfo">
                            <el-descriptions :column="1" border>
                                <el-descriptions-item class-name="system-content" label-class-name="system-label">
                                    <template #label>
                                        <span>{{ $t('home.hostname') }}</span>
                                    </template>
                                    <el-tooltip
                                        v-if="baseInfo.hostname.length > 30"
                                        :content="baseInfo.hostname"
                                        placement="bottom"
                                    >
                                        {{ baseInfo.hostname.substring(0, 27) + '...' }}
                                    </el-tooltip>
                                    <span v-else>{{ baseInfo.hostname }}</span>
                                </el-descriptions-item>
                                <el-descriptions-item class-name="system-content" label-class-name="system-label">
                                    <template #label>
                                        <span>{{ $t('home.platformVersion') }}</span>
                                    </template>
                                    {{
                                        baseInfo.platformVersion
                                            ? baseInfo.platform
                                            : baseInfo.platform + '-' + baseInfo.platformVersion
                                    }}
                                </el-descriptions-item>
                                <el-descriptions-item class-name="system-content" label-class-name="system-label">
                                    <template #label>
                                        <span>{{ $t('home.kernelVersion') }}</span>
                                    </template>
                                    <el-tooltip
                                        v-if="baseInfo.kernelVersion.length > 30"
                                        :content="baseInfo.kernelVersion"
                                        placement="bottom"
                                    >
                                        {{ baseInfo.kernelVersion.substring(0, 27) + '...' }}
                                    </el-tooltip>
                                    <span v-else>{{ baseInfo.kernelVersion }}</span>
                                </el-descriptions-item>
                                <el-descriptions-item class-name="system-content" label-class-name="system-label">
                                    <template #label>
                                        <span>{{ $t('home.kernelArch') }}</span>
                                    </template>
                                    {{ baseInfo.kernelArch }}
                                </el-descriptions-item>
                                <el-descriptions-item
                                    v-if="baseInfo.ipv4Addr && baseInfo.ipv4Addr !== 'IPNotFound'"
                                    class-name="system-content"
                                    label-class-name="system-label"
                                >
                                    <template #label>
                                        <span>{{ $t('home.ip') }}</span>
                                    </template>
                                    {{ baseInfo.ipv4Addr }}
                                </el-descriptions-item>
                                <el-descriptions-item
                                    v-if="baseInfo.systemProxy && baseInfo.systemProxy !== 'noProxy'"
                                    class-name="system-content"
                                    label-class-name="system-label"
                                >
                                    <template #label>
                                        <span>{{ $t('home.proxy') }}</span>
                                    </template>
                                    {{ baseInfo.systemProxy }}
                                </el-descriptions-item>
                                <el-descriptions-item class-name="system-content" label-class-name="system-label">
                                    <template #label>
                                        <span>{{ $t('home.uptime') }}</span>
                                    </template>
                                    {{ currentInfo.timeSinceUptime }}
                                </el-descriptions-item>
                                <el-descriptions-item class-name="system-content" label-class-name="system-label">
                                    <template #label>
                                        <span>{{ $t('home.runningTime') }}</span>
                                    </template>
                                    {{ loadUpTime(currentInfo.uptime) }}
                                </el-descriptions-item>
                            </el-descriptions>
                        </div>
                    </template>
                </CardWithHeader>

                <CardWithHeader :header="$t('home.app')" style="margin-top: 20px">
                    <template #body>
                        <App ref="appRef" />
                    </template>
                </CardWithHeader>
            </el-col>
        </el-row>

        <LicenseImport ref="licenseRef" />
    </div>
</template>

<script lang="ts" setup>
import { onMounted, onBeforeUnmount, ref, reactive } from 'vue';
import Status from '@/views/home/status/index.vue';
import App from '@/views/home/app/index.vue';
import VCharts from '@/components/v-charts/index.vue';
import LicenseImport from '@/components/license-import/index.vue';
import CardWithHeader from '@/components/card-with-header/index.vue';
import i18n from '@/lang';
import { Dashboard } from '@/api/interface/dashboard';
import { dateFormatForSecond, computeSize, computeSizeFromKBs } from '@/utils/util';
import { useRouter } from 'vue-router';
import { loadBaseInfo, loadCurrentInfo } from '@/api/modules/dashboard';
import { getIOOptions, getNetworkOptions } from '@/api/modules/host';
import { getSettingInfo, loadUpgradeInfo } from '@/api/modules/setting';
import { GlobalStore } from '@/store';
const router = useRouter();
const globalStore = GlobalStore();

const statusRef = ref();
const appRef = ref();

const isSafety = ref();

const chartOption = ref('network');
let timer: NodeJS.Timer | null = null;
let isInit = ref<boolean>(true);
let isActive = ref(true);

const ioReadBytes = ref<Array<number>>([]);
const ioWriteBytes = ref<Array<number>>([]);
const netBytesSents = ref<Array<number>>([]);
const netBytesRecvs = ref<Array<number>>([]);
const timeIODatas = ref<Array<string>>([]);
const timeNetDatas = ref<Array<string>>([]);

const ioOptions = ref();
const netOptions = ref();

const licenseRef = ref();
const isProductPro = ref();

const searchInfo = reactive({
    ioOption: 'all',
    netOption: 'all',
    scope: 'all',
});

const baseInfo = ref<Dashboard.BaseInfo>({
    websiteNumber: 0,
    databaseNumber: 0,
    cronjobNumber: 0,
    appInstalledNumber: 0,

    hostname: '',
    os: '',
    platform: '',
    platformFamily: '',
    platformVersion: '',
    kernelArch: '',
    kernelVersion: '',
    virtualizationSystem: '',
    ipv4Addr: '',
    systemProxy: '',
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
    memoryUsedPercent: 0,
    swapMemoryTotal: 0,
    swapMemoryAvailable: 0,
    swapMemoryUsed: 0,
    swapMemoryUsedPercent: 0,

    ioReadBytes: 0,
    ioWriteBytes: 0,
    ioCount: 0,
    ioReadTime: 0,
    ioWriteTime: 0,

    diskData: [],
    gpuData: [],
    xpuData: [],

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
    searchInfo.netOption = globalStore.defaultNetwork || (netOptions.value && netOptions.value[0]);
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

    const resData = res.data.currentInfo;
    currentInfo.value.ioReadBytes = resData.ioReadBytes;
    currentInfo.value.ioWriteBytes = resData.ioWriteBytes;
    currentInfo.value.ioCount = resData.ioCount;
    currentInfo.value.ioReadTime = resData.ioReadTime;
    currentInfo.value.ioWriteTime = resData.ioWriteTime;
    currentInfo.value.netBytesSent = resData.netBytesSent;
    currentInfo.value.netBytesRecv = resData.netBytesRecv;
    currentInfo.value.uptime = resData.uptime;

    loadAppCurrentInfo();
    statusRef.value.acceptParams(currentInfo.value, baseInfo.value);
    appRef.value.acceptParams();
    if (isInit) {
        timer = setInterval(async () => {
            if (isActive.value && !globalStore.isOnRestart) {
                loadAppCurrentInfo();
            }
        }, 3000);
    }
};

const loadAppCurrentInfo = async () => {
    await Promise.all([onLoadCurrentInfo('gpu'), onLoadCurrentInfo('basic'), onLoadCurrentInfo('ioNet')]);
    statusRef.value.acceptParams(currentInfo.value, baseInfo.value);
};

const onLoadCurrentInfo = async (scope: string) => {
    const req = {
        scope: scope,
        ioOption: searchInfo.ioOption,
        netOption: searchInfo.netOption,
    };
    const res = await loadCurrentInfo(req);
    const resData = res.data;

    if (scope === 'ioNet') {
        let timeInterval = Number(res.data.uptime - currentInfo.value.uptime) || 3;
        currentChartInfo.netBytesSent =
            res.data.netBytesSent - currentInfo.value.netBytesSent > 0
                ? Number(((res.data.netBytesSent - currentInfo.value.netBytesSent) / 1024 / timeInterval).toFixed(2))
                : 0;
        netBytesSents.value.push(currentChartInfo.netBytesSent);

        if (netBytesSents.value.length > 20) {
            netBytesSents.value.splice(0, 1);
        }

        currentChartInfo.netBytesRecv =
            res.data.netBytesRecv - currentInfo.value.netBytesRecv > 0
                ? Number(((res.data.netBytesRecv - currentInfo.value.netBytesRecv) / 1024 / timeInterval).toFixed(2))
                : 0;
        netBytesRecvs.value.push(currentChartInfo.netBytesRecv);
        if (netBytesRecvs.value.length > 20) {
            netBytesRecvs.value.splice(0, 1);
        }

        currentChartInfo.ioReadBytes =
            res.data.ioReadBytes - currentInfo.value.ioReadBytes > 0
                ? Number(
                      ((res.data.ioReadBytes - currentInfo.value.ioReadBytes) / 1024 / 1024 / timeInterval).toFixed(2),
                  )
                : 0;
        ioReadBytes.value.push(currentChartInfo.ioReadBytes);
        if (ioReadBytes.value.length > 20) {
            ioReadBytes.value.splice(0, 1);
        }

        currentChartInfo.ioWriteBytes =
            res.data.ioWriteBytes - currentInfo.value.ioWriteBytes > 0
                ? Number(
                      ((res.data.ioWriteBytes - currentInfo.value.ioWriteBytes) / 1024 / 1024 / timeInterval).toFixed(
                          2,
                      ),
                  )
                : 0;
        ioWriteBytes.value.push(currentChartInfo.ioWriteBytes);
        if (ioWriteBytes.value.length > 20) {
            ioWriteBytes.value.splice(0, 1);
        }
        currentChartInfo.ioCount = Math.round(Number((res.data.ioCount - currentInfo.value.ioCount) / timeInterval));
        let ioReadTime = res.data.ioReadTime - currentInfo.value.ioReadTime;
        let ioWriteTime = res.data.ioWriteTime - currentInfo.value.ioWriteTime;
        let ioChoose = ioReadTime > ioWriteTime ? ioReadTime : ioWriteTime;
        currentChartInfo.ioTime = Math.round(Number(ioChoose / timeInterval));

        timeIODatas.value.push(dateFormatForSecond(res.data.shotTime));
        if (timeIODatas.value.length > 20) {
            timeIODatas.value.splice(0, 1);
        }
        timeNetDatas.value.push(dateFormatForSecond(res.data.shotTime));
        if (timeNetDatas.value.length > 20) {
            timeNetDatas.value.splice(0, 1);
        }
        loadData();

        currentInfo.value.ioReadBytes = resData.ioReadBytes;
        currentInfo.value.ioWriteBytes = resData.ioWriteBytes;
        currentInfo.value.ioCount = resData.ioCount;
        currentInfo.value.ioReadTime = resData.ioReadTime;
        currentInfo.value.ioWriteTime = resData.ioWriteTime;

        currentInfo.value.netBytesSent = resData.netBytesSent;
        currentInfo.value.netBytesRecv = resData.netBytesRecv;
    }
    if (scope === 'gpu') {
        currentInfo.value.gpuData = resData.gpuData;
        currentInfo.value.xpuData = resData.xpuData;
    }
    if (scope === 'basic') {
        currentInfo.value.uptime = resData.uptime;
        currentInfo.value.timeSinceUptime = resData.timeSinceUptime;
        currentInfo.value.procs = resData.procs;

        currentInfo.value.load1 = resData.load1;
        currentInfo.value.load5 = resData.load5;
        currentInfo.value.load15 = resData.load15;
        currentInfo.value.loadUsagePercent = resData.loadUsagePercent;

        currentInfo.value.cpuPercent = resData.cpuPercent;
        currentInfo.value.cpuUsedPercent = resData.cpuUsedPercent;
        currentInfo.value.cpuUsed = resData.cpuUsed;
        currentInfo.value.cpuTotal = resData.cpuTotal;

        currentInfo.value.memoryTotal = resData.memoryTotal;
        currentInfo.value.memoryAvailable = resData.memoryAvailable;
        currentInfo.value.memoryUsed = resData.memoryUsed;
        currentInfo.value.memoryUsedPercent = resData.memoryUsedPercent;

        currentInfo.value.swapMemoryTotal = resData.swapMemoryTotal;
        currentInfo.value.swapMemoryAvailable = resData.swapMemoryAvailable;
        currentInfo.value.swapMemoryUsed = resData.swapMemoryUsed;
        currentInfo.value.swapMemoryUsedPercent = resData.swapMemoryUsedPercent;

        currentInfo.value.timeSinceUptime = res.data.timeSinceUptime;
        currentInfo.value.shotTime = resData.shotTime;
        currentInfo.value.diskData = resData.diskData;
    }
};

function loadUpTime(uptime: number) {
    if (uptime <= 0) {
        return '-';
    }
    let days = Math.floor(uptime / 86400);
    let hours = Math.floor((uptime % 86400) / 3600);
    let minutes = Math.floor((uptime % 3600) / 60);
    let seconds = uptime % 60;
    let uptimeParts = [];
    let lead = false;
    if (days !== 0) {
        uptimeParts.push(days + i18n.global.t('commons.units.dayUnit', days));
        lead = true;
    }
    if (lead || hours !== 0) {
        uptimeParts.push(hours + i18n.global.t('commons.units.hourUnit', hours));
        lead = true;
    }
    if (lead || minutes !== 0) {
        uptimeParts.push(minutes + i18n.global.t('commons.units.minuteUnit', minutes));
        lead = true;
    }
    if (lead || seconds !== 0) {
        uptimeParts.push(seconds + i18n.global.t('commons.units.secondUnit', seconds));
        lead = true;
    }
    return lead ? uptimeParts.join(' ') : '-';
}

const loadData = async () => {
    if (chartOption.value === 'io') {
        chartsOption.value['ioChart'] = {
            xData: timeIODatas.value,
            yData: [
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
            xData: timeNetDatas.value,
            yData: [
                {
                    name: i18n.global.t('monitor.up'),
                    data: netBytesSents.value,
                },
                {
                    name: i18n.global.t('monitor.down'),
                    data: netBytesRecvs.value,
                },
            ],
            formatStr: 'KB/s',
        };
    }
};

const hideEntrance = () => {
    globalStore.setShowEntranceWarn(false);
};

const loadUpgradeStatus = async () => {
    const res = await loadUpgradeInfo();
    if (res.data.testVersion || res.data.newVersion || res.data.latestVersion) {
        globalStore.hasNewVersion = true;
    } else {
        globalStore.hasNewVersion = false;
    }
};

const loadSafeStatus = async () => {
    const res = await getSettingInfo();
    isSafety.value = res.data.securityEntrance;
};

const onFocus = () => {
    isActive.value = true;
};
const onBlur = () => {
    isActive.value = false;
};

const toUpload = () => {
    licenseRef.value.acceptParams();
};

onMounted(() => {
    isProductPro.value = globalStore.isProductPro;
    window.addEventListener('focus', onFocus);
    window.addEventListener('blur', onBlur);
    loadSafeStatus();
    loadUpgradeStatus();
    onLoadNetworkOptions();
    onLoadIOOptions();
    onLoadBaseInfo(true, 'all');
});

onBeforeUnmount(() => {
    window.removeEventListener('focus', onFocus);
    window.removeEventListener('blur', onBlur);
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
    @media only screen and (max-width: 1300px) {
        span:first-child {
            font-size: 12px;
            color: var(--el-text-color-regular);
        }
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
    height: 296px;
    overflow: auto;
}
@-moz-document url-prefix() {
    .h-systemInfo {
        height: auto;
    }
}

.system-label {
    font-weight: 400 !important;
    font-size: 14px !important;
    color: var(--panel-text-color);
    border: none !important;
    background: none !important;
    max-width: 150px !important;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
}

.system-content {
    font-size: 13px !important;
    border: none !important;
    width: 100% !important;
    line-height: normal !important;
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

.version {
    font-size: 14px;
    color: #858585;
    text-decoration: none;
    letter-spacing: 0.5px;
}

.system-link {
    margin-left: 15px;

    .svg-icon {
        font-size: 7px;
    }
    span {
        line-height: 20px;
    }
}
</style>
