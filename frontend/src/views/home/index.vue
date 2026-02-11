<template>
    <div :key="$route.fullPath" id="dashboard">
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
                    @mouseenter="refreshOptionsOnHover"
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
                    :class="{ 'no-indicator': carouselItemCount <= 1 }"
                    :key="simpleNodes.length + carouselItemCount"
                    height="368px"
                    indicator-position=""
                    arrow="never"
                >
                    <el-carousel-item key="systemInfo">
                        <CardWithHeader :header="$t('home.systemInfo')">
                            <template #header-r>
                                <el-popover
                                    popper-class="dashboard-carousel-popover"
                                    placement="bottom"
                                    :title="$t('home.carouselSetting')"
                                    width="220"
                                    trigger="click"
                                >
                                    <div class="dashboard-carousel-setting">
                                        <div class="setting-item mt-2">
                                            <span>{{ $t('home.systemInfo') }}</span>
                                            <div class="mr-4">-</div>
                                        </div>
                                        <div class="setting-item mt-2">
                                            <span>{{ $t('home.memo') }}</span>
                                            <el-switch
                                                v-model="memoCarouselSetting"
                                                active-value="Enable"
                                                inactive-value="Disable"
                                                @change="
                                                    (val) => updateDashboardCarouselSetting('DashboardMemoVisible', val)
                                                "
                                            />
                                        </div>
                                        <div class="setting-item">
                                            <span>{{ $t('setting.panel') }}</span>
                                            <el-switch
                                                v-model="simpleNodeCarouselSetting"
                                                active-value="Enable"
                                                inactive-value="Disable"
                                                @change="
                                                    (val) =>
                                                        updateDashboardCarouselSetting(
                                                            'DashboardSimpleNodeVisible',
                                                            val,
                                                        )
                                                "
                                            />
                                        </div>
                                    </div>
                                    <template #reference>
                                        <el-button class="h-button-setting" link icon="Setting" />
                                    </template>
                                </el-popover>
                                <el-tooltip :content="$t('commons.button.refresh')" placement="top">
                                    <el-button class="h-button-setting" @click="refreshDashboard" link icon="Refresh" />
                                </el-tooltip>
                                <el-tooltip :content="$t('home.tooltipSensitiveInfo')" placement="top">
                                    <el-button
                                        class="h-button-setting"
                                        @click="toggleSensitiveInfo"
                                        link
                                        :icon="showSensitiveInfo ? 'View' : 'Hide'"
                                    />
                                </el-tooltip>
                                <el-tooltip :content="$t('commons.button.copy')" placement="top">
                                    <el-button class="h-button-setting" @click="handleCopy" link icon="CopyDocument" />
                                </el-tooltip>
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
                                            {{ showSensitiveInfo ? baseInfo.hostname : '****' }}
                                        </el-descriptions-item>
                                        <el-descriptions-item
                                            class-name="system-content"
                                            label-class-name="system-label"
                                        >
                                            <template #label>
                                                <span class="system-label">{{ $t('home.platformVersion') }}</span>
                                            </template>
                                            {{
                                                baseInfo.prettyDistro
                                                    ? baseInfo.prettyDistro
                                                    : baseInfo.platformVersion
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
                                            {{ showSensitiveInfo ? baseInfo.ipV4Addr : '****' }}
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
                                            {{ loadUpTime(currentInfo.timeSinceUptime) }}
                                        </el-descriptions-item>
                                    </el-descriptions>
                                </el-scrollbar>
                            </template>
                        </CardWithHeader>
                    </el-carousel-item>
                    <el-carousel-item key="memoInfo" v-if="showMemoCarousel">
                        <CardWithHeader :header="$t('home.memo')" class="memo-card">
                            <template #header-r>
                                <el-tooltip v-if="!memoEditing" :content="$t('commons.button.edit')" placement="top">
                                    <el-button class="h-button-setting" @click="startMemoEdit" link icon="Edit" />
                                </el-tooltip>
                                <el-tooltip v-if="memoEditing" :content="$t('commons.button.save')" placement="top">
                                    <el-button
                                        class="h-button-setting"
                                        @click="saveMemo"
                                        link
                                        icon="Check"
                                        :loading="memoSaving"
                                    />
                                </el-tooltip>
                                <el-tooltip v-if="memoEditing" :content="$t('commons.button.cancel')" placement="top">
                                    <el-button class="h-button-setting" @click="cancelMemoEdit" link icon="Close" />
                                </el-tooltip>
                            </template>
                            <template #body>
                                <el-scrollbar height="286px">
                                    <div class="memo-container ml-5 mr-5">
                                        <el-input
                                            v-if="memoEditing"
                                            v-model="memoEditContent"
                                            type="textarea"
                                            :rows="10"
                                            :maxlength="500"
                                            :placeholder="$t('home.memoPlaceholder')"
                                            show-word-limit
                                        />
                                        <div v-else class="memo-content">
                                            <MarkDownEditor v-if="memoContent" :content="memoContent" />
                                            <span v-else class="memo-placeholder">
                                                {{ $t('home.memoPlaceholder') }}
                                            </span>
                                        </div>
                                    </div>
                                </el-scrollbar>
                            </template>
                        </CardWithHeader>
                    </el-carousel-item>
                    <el-carousel-item key="simpleNode" v-if="showSimpleNode()">
                        <CardWithHeader :header="$t('setting.panel')">
                            <template #header-r>
                                <el-tooltip :content="$t('xpack.node.panelItem')" placement="top">
                                    <el-button
                                        class="h-button-setting"
                                        @click="routerToNameWithQuery('SimpleNode', { uncached: 'true' })"
                                        link
                                        icon="Setting"
                                    />
                                </el-tooltip>
                            </template>
                            <template #body>
                                <el-scrollbar height="286px">
                                    <div class="simple-node cursor-pointer" v-for="row in simpleNodes" :key="row.id">
                                        <el-row :gutter="5">
                                            <el-col :span="21">
                                                <div class="name">
                                                    {{ row.name }}
                                                    <Status :status="row.status" :msg="row.message" />
                                                </div>
                                                <div class="detail">
                                                    {{ loadSource(row) }}
                                                </div>
                                            </el-col>

                                            <el-col :span="1">
                                                <el-button
                                                    @click="jumpPanel(row)"
                                                    size="small"
                                                    :disabled="row.status !== 'Healthy'"
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
import { onMounted, onBeforeUnmount, ref, reactive, computed, nextTick } from 'vue';
import SystemStatus from '@/views/home/status/index.vue';
import AppLauncher from '@/views/home/app/index.vue';
import VCharts from '@/components/v-charts/index.vue';
import LicenseImport from '@/components/license-import/index.vue';
import QuickJump from '@/views/home/quick/index.vue';
import CardWithHeader from '@/components/card-with-header/index.vue';
import MarkDownEditor from '@/components/mkdown-editor/index.vue';
import i18n from '@/lang';
import { Dashboard } from '@/api/interface/dashboard';
import { dateFormatForSecond, computeSize, computeSizeFromKBs, loadUpTime, jumpToPath, copyText } from '@/utils/util';
import { useRouter } from 'vue-router';
import { loadBaseInfo, loadCurrentInfo } from '@/api/modules/dashboard';
import { getIOOptions, getNetworkOptions } from '@/api/modules/host';
import {
    getSettingInfo,
    getAgentSettingInfo,
    listAllSimpleNodes,
    loadUpgradeInfo,
    getMemo,
    updateMemo,
    updateSetting,
} from '@/api/modules/setting';
import { GlobalStore } from '@/store';
import { storeToRefs } from 'pinia';
import { routerToFileWithPath, routerToNameWithQuery, routerToPath } from '@/utils/router';
import { getWelcomePage } from '@/api/modules/auth';
import {
    clearDashboardCache,
    clearDashboardCacheByPrefix,
    getDashboardCache,
    setDashboardCache,
} from '@/utils/dashboardCache';
import { MsgSuccess } from '@/utils/message';
const router = useRouter();
const globalStore = GlobalStore();

const DASHBOARD_CACHE_TTL = {
    safeStatus: 10 * 60 * 1000,
    netOptions: 60 * 60 * 1000,
    ioOptions: 60 * 60 * 1000,
};

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

const showSensitiveInfo = ref(true);

const ioReadBytes = ref<Array<number>>([]);
const ioWriteBytes = ref<Array<number>>([]);
const netBytesSents = ref<Array<number>>([]);
const netBytesRecvs = ref<Array<number>>([]);
const timeIODatas = ref<Array<string>>([]);
const timeNetDatas = ref<Array<string>>([]);

const simpleNodes = ref([]);
const ioOptions = ref();
const netOptions = ref();
const netOptionsFromCache = ref(false);
const ioOptionsFromCache = ref(false);
const hasRefreshedOptionsOnHover = ref(false);

const licenseRef = ref();
const quickJumpRef = ref();
const { isProductPro, isOffLine } = storeToRefs(globalStore);

const searchInfo = reactive({
    ioOption: 'all',
    netOption: 'all',
});

const memoContent = ref('');
const memoEditContent = ref('');
const memoEditing = ref(false);
const memoSaving = ref(false);
const memoCarouselSetting = ref();
const simpleNodeCarouselSetting = ref();
const carouselSettingReady = ref(false);

const showMemoCarousel = computed(() => memoCarouselSetting.value === 'Enable');
const carouselItemCount = computed(() => {
    let count = 1;
    if (showMemoCarousel.value) count += 1;
    if (showSimpleNode()) count += 1;
    return count;
});

const baseInfo = ref<Dashboard.BaseInfo>({
    hostname: '',
    os: '',
    platform: '',
    platformFamily: '',
    platformVersion: '',
    prettyDistro: '',
    kernelArch: '',
    kernelVersion: '',
    virtualizationSystem: '',
    ipV4Addr: '',
    httpProxy: '',

    cpuCores: 0,
    cpuLogicalCores: 0,
    cpuModelName: '',
    cpuMhz: 0,
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
    cpuDetailedPercent: [] as Array<number>,

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

    topCPUItems: [],
    topMemItems: [],

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

const updateCurrentInfo = (data: Dashboard.CurrentInfo) => {
    currentInfo.value = {
        ...data,
        topCPUItems: currentInfo.value.topCPUItems || [],
        topMemItems: currentInfo.value.topMemItems || [],
    };
};

const changeOption = async () => {
    isInit.value = true;
    loadData();
};

const applyDefaultNetOption = () => {
    if (!netOptions.value || netOptions.value.length === 0) return;
    const defaultNet = globalStore.defaultNetwork || netOptions.value[0];
    if (defaultNet && searchInfo.netOption !== defaultNet) {
        searchInfo.netOption = defaultNet;
    }
};

const onLoadAgentSettingInfo = async () => {
    await getAgentSettingInfo().then((res) => {
        globalStore.defaultIO = res.data.defaultIO;
        globalStore.defaultNetwork = res.data.defaultNetwork;
    });
};

const onLoadNetworkOptions = async (force?: boolean) => {
    const cache = force ? null : getDashboardCache('netOptions');
    if (cache !== null) {
        netOptions.value = cache;
        netOptionsFromCache.value = true;
        applyDefaultNetOption();
        return;
    }
    const res = await getNetworkOptions();
    netOptions.value = res.data;
    netOptionsFromCache.value = false;
    setDashboardCache('netOptions', res.data, DASHBOARD_CACHE_TTL.netOptions);
    applyDefaultNetOption();
};

const onLoadSimpleNode = async () => {
    const res = await listAllSimpleNodes();
    simpleNodes.value = res.data || [];
};

const applyDefaultIOOption = async () => {
    if (!ioOptions.value || ioOptions.value.length === 0) return;
    const defaultIO = globalStore.defaultIO || ioOptions.value[0];
    if (defaultIO && searchInfo.ioOption !== defaultIO) {
        searchInfo.ioOption = defaultIO;
    }
};

const onLoadIOOptions = async (force?: boolean) => {
    const cache = force ? null : getDashboardCache('ioOptions');
    if (cache !== null) {
        ioOptions.value = cache;
        ioOptionsFromCache.value = true;
        applyDefaultIOOption();
        return;
    }
    const res = await getIOOptions();
    ioOptions.value = res.data;
    ioOptionsFromCache.value = false;
    setDashboardCache('ioOptions', ioOptions.value, DASHBOARD_CACHE_TTL.ioOptions);
    applyDefaultIOOption();
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
    updateCurrentInfo(baseInfo.value.currentInfo);
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
    return (
        simpleNodeCarouselSetting.value === 'Enable' &&
        globalStore.isMasterProductPro &&
        simpleNodes.value?.length !== 0
    );
};

const toggleSensitiveInfo = () => {
    showSensitiveInfo.value = !showSensitiveInfo.value;
};

const refreshDashboard = async () => {
    clearDashboardCache();
    onLoadBaseInfo(false, '');
    hasRefreshedOptionsOnHover.value = false;
    await Promise.allSettled([onLoadNetworkOptions(true), onLoadIOOptions(true), loadSettingInfo()]);
    MsgSuccess(i18n.global.t('commons.msg.refreshSuccess'));
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
    updateCurrentInfo(res.data);
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
        (baseInfo.value.prettyDistro
            ? baseInfo.value.prettyDistro
            : baseInfo.value.platformVersion
            ? baseInfo.value.platform + '-' + baseInfo.value.platformVersion
            : baseInfo.value.platform) +
        '\n' +
        i18n.global.t('home.kernelVersion') +
        ': ' +
        baseInfo.value.kernelVersion +
        '\n' +
        i18n.global.t('home.kernelArch') +
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
        loadUpTime(currentInfo.value.timeSinceUptime) +
        '\n';
    copyText(content);
};

const loadMemo = async () => {
    try {
        const res = await getMemo();
        memoContent.value = res.data || '';
    } catch (error) {
        memoContent.value = '';
    }
};

const updateDashboardCarouselSetting = async (key: string, value: 'Enable' | 'Disable') => {
    if (!carouselSettingReady.value) {
        return;
    }
    let target;
    if (key === 'DashboardMemoVisible') {
        target = memoCarouselSetting.value;
        clearDashboardCacheByPrefix(['memoCarouselSetting']);
    } else {
        target = simpleNodeCarouselSetting.value;
        clearDashboardCacheByPrefix(['simpleNodeCarouselSetting']);
    }
    const previous = value === 'Enable' ? 'Disable' : 'Enable';
    try {
        await updateSetting({ key, value });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    } catch (error) {
        target.value = previous;
    }
};

const startMemoEdit = () => {
    memoEditContent.value = memoContent.value;
    memoEditing.value = true;
};

const cancelMemoEdit = () => {
    memoEditing.value = false;
    memoEditContent.value = '';
};

const saveMemo = async () => {
    memoSaving.value = true;
    try {
        await updateMemo(memoEditContent.value);
        memoContent.value = memoEditContent.value;
        memoEditing.value = false;
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    } finally {
        memoSaving.value = false;
    }
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

const loadSettingInfo = async () => {
    const safeCache = getDashboardCache('safeStatus');
    const memoCache = getDashboardCache('memoCarouselSetting');
    const simpleNodeCache = getDashboardCache('simpleNodeCarouselSetting');
    if (safeCache === null || memoCache === null || simpleNodeCache === null) {
        const res = await getSettingInfo();
        isSafety.value = res.data.securityEntrance;
        memoCarouselSetting.value = res.data.dashboardMemoVisible;
        simpleNodeCarouselSetting.value = res.data.dashboardSimpleNodeVisible;
        setDashboardCache('safeStatus', isSafety.value, DASHBOARD_CACHE_TTL.safeStatus);
        setDashboardCache('memoCarouselSetting', memoCarouselSetting.value, DASHBOARD_CACHE_TTL.safeStatus);
        setDashboardCache('simpleNodeCarouselSetting', simpleNodeCarouselSetting.value, DASHBOARD_CACHE_TTL.safeStatus);
        if (!carouselSettingReady.value) {
            await nextTick();
            carouselSettingReady.value = true;
        }
        return;
    }
    isSafety.value = safeCache;
    memoCarouselSetting.value = memoCache;
    simpleNodeCarouselSetting.value = simpleNodeCache;
    if (!carouselSettingReady.value) {
        await nextTick();
        carouselSettingReady.value = true;
    }
};

const loadSource = (row: any) => {
    if (row.status !== 'Healthy') {
        return `- ${i18n.global.t('commons.units.core')} (-%) / - GB (-%)`;
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

const refreshOptionsOnHover = async () => {
    if (hasRefreshedOptionsOnHover.value) return;
    if (!netOptionsFromCache.value && !ioOptionsFromCache.value) return;
    hasRefreshedOptionsOnHover.value = true;
    if (netOptionsFromCache.value) {
        await onLoadNetworkOptions(true);
    }
    if (ioOptionsFromCache.value) {
        await onLoadIOOptions(true);
    }
};

const scheduleDeferredFetch = () => {
    setTimeout(() => {
        onLoadSimpleNode();
        onLoadNetworkOptions();
        onLoadIOOptions();
    }, 600);
};

const fetchData = async () => {
    window.addEventListener('focus', onFocus);
    window.addEventListener('blur', onBlur);
    hasRefreshedOptionsOnHover.value = false;
    loadSettingInfo();
    onLoadAgentSettingInfo();
    onLoadBaseInfo(true, 'all');
    scheduleDeferredFetch();
    setTimeout(() => {
        loadUpgradeStatus();
    }, 2000);
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
    loadMemo();
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
    &.no-indicator {
        .el-carousel__indicators {
            display: none;
        }
    }

    .el-carousel__button {
        margin-bottom: -4px;
        background-color: var(--el-text-color-regular);
    }

    .el-carousel__indicator.is-active .el-carousel__button {
        background-color: var(--panel-color-primary);
    }

    .el-descriptions .el-descriptions__body .el-descriptions__table {
        border-spacing: 0 5px !important;
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
        font-size: 18px !important;
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

.memo-container {
    height: 270px;
}

.memo-card {
    height: 368px;
}

.memo-content {
    min-height: 218px;
    border-radius: 4px;
    border: 1px solid var(--el-border-color);
    background-color: var(--el-fill-color-light);
    word-wrap: break-word;
    white-space: pre-wrap;

    .md-editor {
        background-color: transparent;
    }
}

.memo-placeholder {
    color: var(--el-text-color-placeholder);
    display: inline-block;
    font-size: 14px;
}

.dashboard-carousel-setting {
    display: flex;
    flex-direction: column;

    .setting-item {
        display: flex;
        align-items: center;
        justify-content: space-between;
    }
}

.dashboard-carousel-popover {
    .el-popover__title {
        padding-bottom: 10px;
        margin-bottom: 8px;
        border-bottom: 1px solid var(--el-border-color);
    }
}
</style>
