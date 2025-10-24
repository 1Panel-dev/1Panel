<template>
    <div :key="$route.fullPath">
        <RouterButton
            :buttons="[
                {
                    label: i18n.global.t('menu.home'),
                    path: '/',
                },
            ]"
        >
            <template #route-button>
                <div class="router-button" v-if="!isOffLine">
                    <template v-if="!isProductPro">
                        <el-button link type="primary" @click="toUpload">
                            {{ $t('license.levelUpPro') }}
                        </el-button>
                    </template>
                </div>
            </template>
        </RouterButton>

        <el-alert
            v-if="!isSafety && globalStore.showEntranceWarn"
            class="card-interval"
            type="warning"
            @close="hideEntrance"
        >
            <template #title>
                <span class="flx-align-center">
                    <span>{{ $t('home.entranceHelper') }}</span>
                    <el-link
                        style="font-size: 12px; margin-left: 5px"
                        icon="Position"
                        @click="jumpToPath(router, '/settings/safe')"
                        type="primary"
                    >
                        {{ $t('firewall.quickJump') }}
                    </el-link>
                </span>
            </template>
        </el-alert>

        <el-row :gutter="7" class="card-interval">
            <el-col :xs="24" :sm="24" :md="16" :lg="16" :xl="16">
                <CardWithHeader :header="$t('menu.home')" height="166px">
                    <template #header-r>
                        <el-button class="h-button-setting" @click="quickJumpRef.acceptParams()" link icon="Setting" />
                    </template>
                    <template #body>
                        <div class="h-overview">
                            <el-row>
                                <el-col :span="6" v-for="item in baseInfo.quickJump" :key="item.name">
                                    <span>{{ $t(item.title, 2) }}</span>
                                    <div class="count">
                                        <el-tooltip
                                            v-if="item.alias || item.detail.length > 20"
                                            :content="item.detail"
                                            placement="bottom"
                                        >
                                            <span @click="quickJump(item)">
                                                {{ item.alias || item.detail.substring(0, 18) + '...' }}
                                            </span>
                                        </el-tooltip>
                                        <span @click="quickJump(item)" v-else>{{ item.detail }}</span>
                                    </div>
                                </el-col>
                            </el-row>
                        </div>
                    </template>
                </CardWithHeader>
                <CardWithHeader :header="$t('commons.table.status')" class="card-interval">
                    <template #body>
                        <SystemStatus ref="statusRef" style="margin-bottom: 33px" />
                    </template>
                </CardWithHeader>
                <CardWithHeader
                    :header="$t('menu.monitor')"
                    class="card-interval chart-card"
                    v-loading="!chartsOption['networkChart']"
                >
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
                <el-carousel
                    class="my-carousel"
                    :key="simpleNodes.length"
                    height="368px"
                    :indicator-position="showSimpleNode() ? '' : 'none'"
                    arrow="never"
                >
                    <el-carousel-item key="systemInfo">
                        <CardWithHeader :header="$t('home.systemInfo')">
                            <template #header-r>
                                <el-button class="h-button-setting" @click="handleCopy" link icon="CopyDocument" />
                            </template>
                            <template #body>
                                <el-scrollbar>
                                    <el-descriptions :column="1" class="ml-5 -mt-2 h-systemInfo" border>
                                        <el-descriptions-item
                                            class-name="system-content"
                                            label-class-name="system-label"
                                        >
                                            <template #label>
                                                <span class="system-label">{{ $t('home.hostname') }}</span>
                                            </template>
                                            {{ baseInfo.hostname }}
                                        </el-descriptions-item>
                                        <el-descriptions-item
                                            class-name="system-content"
                                            label-class-name="system-label"
                                        >
                                            <template #label>
                                                <span class="system-label">{{ $t('home.platformVersion') }}</span>
                                            </template>
                                            {{
                                                baseInfo.platformVersion
                                                    ? baseInfo.platform + '-' + baseInfo.platformVersion
                                                    : baseInfo.platform
                                            }}
                                        </el-descriptions-item>
                                        <el-descriptions-item
                                            class-name="system-content"
                                            label-class-name="system-label"
                                        >
                                            <template #label>
                                                <span class="system-label">{{ $t('home.kernelVersion') }}</span>
                                            </template>
                                            {{ baseInfo.kernelVersion }}
                                        </el-descriptions-item>
                                        <el-descriptions-item
                                            class-name="system-content"
                                            label-class-name="system-label"
                                        >
                                            <template #label>
                                                <span class="system-label">{{ $t('home.kernelArch') }}</span>
                                            </template>
                                            {{ baseInfo.kernelArch }}
                                        </el-descriptions-item>
                                        <el-descriptions-item
                                            class-name="system-content"
                                            label-class-name="system-label"
                                        >
                                            <template #label>
                                                <span class="system-label">{{ $t('home.ip') }}</span>
                                            </template>
                                            {{ baseInfo.ipV4Addr }}
                                        </el-descriptions-item>
                                        <el-descriptions-item
                                            v-if="baseInfo.httpProxy && baseInfo.httpProxy !== 'noProxy'"
                                            class-name="system-content"
                                            label-class-name="system-label"
                                        >
                                            <template #label>
                                                <span class="system-label">{{ $t('home.proxy') }}</span>
                                                {{ baseInfo.httpProxy }}
                                            </template>
                                        </el-descriptions-item>
                                        <el-descriptions-item
                                            class-name="system-content"
                                            label-class-name="system-label"
                                        >
                                            <template #label>
                                                <span class="system-label">{{ $t('home.uptime') }}</span>
                                            </template>
                                            {{ currentInfo.timeSinceUptime }}
                                        </el-descriptions-item>
                                        <el-descriptions-item
                                            class-name="system-content"
                                            label-class-name="system-label"
                                        >
                                            <template #label>
                                                <span class="system-label">{{ $t('home.runningTime') }}</span>
                                            </template>
                                            {{ loadUpTime(currentInfo.uptime) }}
                                        </el-descriptions-item>
                                    </el-descriptions>
                                </el-scrollbar>
                            </template>
                        </CardWithHeader>
                    </el-carousel-item>
                    <el-carousel-item key="simpleNode" v-if="showSimpleNode()">
                        <CardWithHeader :header="$t('setting.panel')">
                            <template #body>
                                <el-scrollbar height="286px">
                                    <div class="simple-node cursor-pointer" v-for="row in simpleNodes" :key="row.id">
                                        <el-row :gutter="5">
                                            <el-col :span="21">
                                                <div class="name">
                                                    {{ row.name }}
                                                </div>
                                                <div class="detail">
                                                    {{ loadSource(row) }}
                                                </div>
                                            </el-col>

                                            <el-col :span="1">
                                                <el-button
                                                    @click="jumpPanel(row)"
                                                    size="small"
                                                    class="visit"
                                                    round
                                                    plain
                                                    type="primary"
                                                >
                                                    {{ $t('commons.button.visit') }}
                                                </el-button>
                                            </el-col>
                                        </el-row>
                                        <div class="h-app-divider" />
                                    </div>
                                </el-scrollbar>
                            </template>
                        </CardWithHeader>
                    </el-carousel-item>
                </el-carousel>

                <AppLauncher ref="appRef" class="card-interval" />
            </el-col>
        </el-row>

        <LicenseImport ref="licenseRef" />
        <QuickJump @search="onLoadBaseInfo(false, 'all')" ref="quickJumpRef" />

        <DialogPro v-model="welcomeOpen" size="w-70" id="welcomeDialog">
            <div ref="shadowContainer" />
        </DialogPro>
    </div>
</template>

<script lang="ts" setup>
import { onMounted, onBeforeUnmount, ref, reactive } from 'vue';
import SystemStatus from '@/views/home/status/index.vue';
import AppLauncher from '@/views/home/app/index.vue';
import VCharts from '@/components/v-charts/index.vue';
import LicenseImport from '@/components/license-import/index.vue';
import QuickJump from '@/views/home/quick/index.vue';
import CardWithHeader from '@/components/card-with-header/index.vue';
import i18n from '@/lang';
import { Dashboard } from '@/api/interface/dashboard';
import { dateFormatForSecond, computeSize, computeSizeFromKBs, loadUpTime, jumpToPath, copyText } from '@/utils/util';
import { useRouter } from 'vue-router';
import { loadBaseInfo, loadCurrentInfo } from '@/api/modules/dashboard';
import { getIOOptions, getNetworkOptions } from '@/api/modules/host';
import { getSettingInfo, listAllSimpleNodes, loadUpgradeInfo } from '@/api/modules/setting';
import { GlobalStore } from '@/store';
import { storeToRefs } from 'pinia';
import { routerToFileWithPath, routerToPath } from '@/utils/router';
import { getWelcomePage } from '@/api/modules/auth';
const router = useRouter();
const globalStore = GlobalStore();

const statusRef = ref();
const appRef = ref();

const isSafety = ref();

const welcomeOpen = ref();
const shadowContainer = ref();

const chartOption = ref('network');
let timer: NodeJS.Timer | null = null;
let isInit = ref<boolean>(true);
let isStatusInit = ref<boolean>(true);
let isActive = ref(true);
let isCurrentActive = ref(true);

const ioReadBytes = ref<Array<number>>([]);
const ioWriteBytes = ref<Array<number>>([]);
const netBytesSents = ref<Array<number>>([]);
const netBytesRecvs = ref<Array<number>>([]);
const timeIODatas = ref<Array<string>>([]);
const timeNetDatas = ref<Array<string>>([]);

const simpleNodes = ref([]);
const ioOptions = ref();
const netOptions = ref();

const licenseRef = ref();
const quickJumpRef = ref();
const { isProductPro, isOffLine } = storeToRefs(globalStore);

const searchInfo = reactive({
    ioOption: 'all',
    netOption: 'all',
});

const baseInfo = ref<Dashboard.BaseInfo>({
    hostname: '',
    os: '',
    platform: '',
    platformFamily: '',
    platformVersion: '',
    kernelArch: '',
    kernelVersion: '',
    virtualizationSystem: '',
    ipV4Addr: '',
    httpProxy: '',

    cpuCores: 0,
    cpuLogicalCores: 0,
    cpuModelName: '',
    currentInfo: null,

    quickJump: [],
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
    memoryFree: 0,
    memoryShard: 0,
    memoryCache: 0,
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

const onLoadNetworkOptions = async () => {
    const res = await getNetworkOptions();
    netOptions.value = res.data;
    searchInfo.netOption = globalStore.defaultNetwork || (netOptions.value && netOptions.value[0]);
};

const onLoadSimpleNode = async () => {
    const res = await listAllSimpleNodes();
    simpleNodes.value = res.data || [];
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
    isStatusInit.value = false;
    statusRef.value?.acceptParams(currentInfo.value, baseInfo.value);
    appRef.value?.acceptParams();
    if (isInit) {
        clearTimer();
        timer = setInterval(async () => {
            try {
                if (!isCurrentActive.value) {
                    throw new Error('jump out');
                }
                if (isActive.value && !globalStore.isOnRestart) {
                    await onLoadCurrentInfo();
                    await onLoadSimpleNode();
                }
            } catch {
                clearTimer();
            }
        }, 3000);
    }
};

const quickJump = (item: any) => {
    if (item.name === 'File') {
        return routerToFileWithPath(item.detail);
    }
    return routerToPath(item.router);
};

const showSimpleNode = () => {
    return globalStore.isMasterProductPro && simpleNodes.value?.length !== 0;
};

const jumpPanel = (row: any) => {
    let entrance = row.securityEntrance.startsWith('/') ? row.securityEntrance.slice(1) : row.securityEntrance;
    entrance = entrance ? '/' + entrance : '';
    let addr = row.addr.endsWith('/') ? row.addr.slice(0, -1) : row.addr;
    window.open(addr + entrance, '_blank', 'noopener,noreferrer');
};

const onLoadCurrentInfo = async () => {
    const res = await loadCurrentInfo(searchInfo.ioOption, searchInfo.netOption);
    currentInfo.value.timeSinceUptime = res.data.timeSinceUptime;

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
            ? Number(((res.data.ioReadBytes - currentInfo.value.ioReadBytes) / 1024 / 1024 / timeInterval).toFixed(2))
            : 0;
    ioReadBytes.value.push(currentChartInfo.ioReadBytes);
    if (ioReadBytes.value.length > 20) {
        ioReadBytes.value.splice(0, 1);
    }

    currentChartInfo.ioWriteBytes =
        res.data.ioWriteBytes - currentInfo.value.ioWriteBytes > 0
            ? Number(((res.data.ioWriteBytes - currentInfo.value.ioWriteBytes) / 1024 / 1024 / timeInterval).toFixed(2))
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
    currentInfo.value = res.data;
    statusRef.value?.acceptParams(currentInfo.value, baseInfo.value);
};

const handleCopy = () => {
    let content =
        i18n.global.t('home.hostname') +
        ': ' +
        baseInfo.value.hostname +
        '\n' +
        i18n.global.t('home.platformVersion') +
        ': ' +
        (baseInfo.value.platformVersion
            ? baseInfo.value.platform + '-' + baseInfo.value.platformVersion
            : baseInfo.value.platform) +
        '\n' +
        i18n.global.t('home.kernelVersion') +
        ': ' +
        baseInfo.value.kernelVersion +
        '\n' +
        i18n.global.t('home.kernelVersion') +
        ': ' +
        baseInfo.value.kernelArch +
        '\n' +
        i18n.global.t('home.ip') +
        ': ' +
        baseInfo.value.ipV4Addr +
        '\n' +
        i18n.global.t('home.uptime') +
        ': ' +
        currentInfo.value.timeSinceUptime +
        '\n' +
        i18n.global.t('home.runningTime') +
        ': ' +
        loadUpTime(currentInfo.value.uptime) +
        '\n';
    copyText(content);
};

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
    if (res && (res.data.testVersion || res.data.newVersion || res.data.latestVersion)) {
        globalStore.hasNewVersion = true;
    } else {
        globalStore.hasNewVersion = false;
    }
};

const loadSafeStatus = async () => {
    const res = await getSettingInfo();
    isSafety.value = res.data.securityEntrance;
};

const loadSource = (row: any) => {
    if (row.status !== 'Healthy') {
        return '-';
    }
    return (
        row.cpuTotal +
        ' ' +
        i18n.global.t('commons.units.core') +
        ' (' +
        row.cpuUsedPercent?.toFixed(2) +
        '%) / ' +
        computeSize(row.memoryTotal) +
        ' (' +
        row.memoryUsedPercent?.toFixed(2) +
        '%)'
    );
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

const fetchData = () => {
    window.addEventListener('focus', onFocus);
    window.addEventListener('blur', onBlur);
    loadSafeStatus();
    loadUpgradeStatus();
    onLoadNetworkOptions();
    onLoadIOOptions();
    onLoadBaseInfo(true, 'all');
    onLoadSimpleNode();
};

const loadWelcome = async () => {
    await getWelcomePage().then((res) => {
        if (res.data) {
            welcomeOpen.value = true;
            nextTick(() => {
                const shadowRoot = shadowContainer.value.attachShadow({ mode: 'open' });
                shadowRoot.innerHTML = res.data;
            });
            localStorage.setItem('welcomeShow', 'false');
        } else {
            localStorage.setItem('welcomeShow', 'false');
        }
    });
};

onBeforeRouteUpdate((to, from, next) => {
    if (to.name === 'home') {
        clearTimer();
        fetchData();
    }
    next();
});

const clearTimer = () => {
    clearInterval(Number(timer));
    timer = null;
};

onMounted(() => {
    fetchData();
    if (localStorage.getItem('welcomeShow') !== 'false') {
        loadWelcome();
    }
});

onBeforeUnmount(() => {
    window.removeEventListener('focus', onFocus);
    window.removeEventListener('blur', onBlur);
    isCurrentActive.value = false;
    clearTimer();
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
            font-size: 18px;
            color: $primary-color;
            line-height: 32px;
            cursor: pointer;
        }
    }
}

.h-systemInfo {
    margin-left: 18px;
    height: 306px;
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
    width: fit-content !important;
    white-space: nowrap !important;
}

.system-content {
    font-size: 13px !important;
    border: none !important;
    width: 100% !important;
}

.my-carousel {
    .el-carousel__button {
        margin-bottom: -4px;
        background-color: var(--el-text-color-regular);
    }
    .el-carousel__indicator.is-active .el-carousel__button {
        background-color: var(--panel-color-primary);
    }
    .el-descriptions .el-descriptions__body .el-descriptions__table {
        border-spacing: 0 5px !important; /* 垂直间距15px */
    }
}

.simple-node {
    padding: 10px 15px 10px 0px;
    margin: -8px 10px 3px 20px;
    &:hover {
        background-color: rgba(0, 94, 235, 0.03);
    }
    .name {
        font-weight: 500 !important;
        font-size: 16px !important;
        line-height: 30px;
        color: var(--panel-text-color);
    }
    .detail {
        font-size: 12px !important;
    }
    .visit {
        margin-bottom: -25px;
    }
}

.h-app-divider {
    margin-top: 3px;
    border: 0;
    border-top: var(--panel-border);
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

.chart-card {
    min-height: 383px;
}
</style>
