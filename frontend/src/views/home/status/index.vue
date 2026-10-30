<template>
    <div class="custom-row">
        <el-col v-if="view === 'system'" :xs="6" :sm="6" :md="3" :lg="3" :xl="3" align="center">
            <el-popover
                :hide-after="20"
                :teleported="false"
                :width="320"
                popper-class="dashboard-status-popover"
                :trigger="hasFinePointer ? 'hover' : 'click'"
                v-if="chartsOption['load']"
                @hide="onCpuPopoverHide"
            >
                <el-descriptions :column="1" size="small">
                    <el-descriptions-item :label="$t('home.loadAverage', [1])">
                        {{ formatNumber(currentInfo.load1) }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('home.loadAverage', [5])">
                        {{ formatNumber(currentInfo.load5) }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('home.loadAverage', [15])">
                        {{ formatNumber(currentInfo.load15) }}
                    </el-descriptions-item>
                </el-descriptions>

                <el-button link size="small" type="primary" class="float-left mb-2" @click="toggleCpuTop">
                    {{ $t('home.cpuTop') }}
                    <el-icon v-if="!showCpuTop"><ArrowRight /></el-icon>
                    <el-icon v-if="showCpuTop"><ArrowDown /></el-icon>
                </el-button>
                <ComplexTable v-if="showCpuTop" :data="currentInfo.topCPUItems">
                    <el-table-column :min-width="120" show-overflow-tooltip :label="$t('menu.process')" prop="name" />
                    <el-table-column :min-width="60" :label="$t('monitor.percent')" prop="percent">
                        <template #default="{ row }">{{ row.percent.toFixed(2) }}%</template>
                    </el-table-column>
                    <el-table-column :width="80" :label="$t('commons.table.operate')">
                        <template #default="{ row }">
                            <el-button type="primary" link @click="onKill(row)">
                                {{ $t('process.stopProcess') }}
                            </el-button>
                        </template>
                    </el-table-column>
                </ComplexTable>
                <template #reference>
                    <v-charts
                        height="160px"
                        id="load"
                        type="pie"
                        :option="chartsOption['load']"
                        v-if="chartsOption['load']"
                    />
                </template>
            </el-popover>
            <span class="input-help">{{ loadStatus(currentInfo.loadUsagePercent) }}</span>
        </el-col>
        <el-col v-if="view === 'system'" :xs="6" :sm="6" :md="3" :lg="3" :xl="3">
            <el-popover
                :hide-after="20"
                :teleported="false"
                :width="430"
                popper-class="dashboard-status-popover"
                :trigger="hasFinePointer ? 'hover' : 'click'"
                v-if="chartsOption['cpu']"
                @hide="onCpuPopoverHide"
            >
                <el-descriptions :title="baseInfo.cpuModelName" class="ml-1" :column="3" size="small">
                    <el-descriptions-item :label="$t('home.core')">
                        {{ baseInfo.cpuCores }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('home.logicCore')">
                        {{ baseInfo.cpuLogicalCores }}
                    </el-descriptions-item>
                    <el-descriptions-item v-if="baseInfo.cpuMhz > 0" :label="$t('home.cpuFrequency')">
                        {{ formatNumber(baseInfo.cpuMhz) }} MHz
                    </el-descriptions-item>
                </el-descriptions>

                <el-button size="small" link type="primary" class="mb-2">
                    {{ $t('home.corePercent') }}
                </el-button>
                <el-space wrap :size="5" class="ml-1">
                    <template v-for="(item, index) of currentInfo.cpuPercent" :key="index">
                        <div class="cpu-detail" v-if="cpuShowAll || (!cpuShowAll && index < 8)">
                            CPU-{{ index }}: {{ formatNumber(item) }}%
                        </div>
                    </template>
                </el-space>
                <div v-if="currentInfo.cpuPercent.length > 8">
                    <el-button v-if="!cpuShowAll" @click="cpuShowAll = true" icon="More" link size="small" />
                    <el-button v-if="cpuShowAll" @click="cpuShowAll = false" icon="ArrowUp" link size="small" />
                </div>

                <div v-if="currentInfo.cpuDetailedPercent?.length >= 8" class="mt-2">
                    <el-button
                        link
                        size="small"
                        type="primary"
                        class="mb-2"
                        @click="showCpuDetailedPercent = !showCpuDetailedPercent"
                    >
                        {{ $t('home.cpuDetailedPercent') }}
                        <el-icon v-if="!showCpuDetailedPercent"><ArrowRight /></el-icon>
                        <el-icon v-if="showCpuDetailedPercent"><ArrowDown /></el-icon>
                    </el-button>
                    <el-space wrap :size="5" class="ml-1 mb-2" v-if="showCpuDetailedPercent">
                        <div class="cpu-detail">
                            {{ $t('home.cpuUser') }}: {{ formatNumber(currentInfo.cpuDetailedPercent[0]) }}%
                        </div>
                        <div class="cpu-detail">
                            {{ $t('home.cpuSystem') }}: {{ formatNumber(currentInfo.cpuDetailedPercent[1]) }}%
                        </div>
                        <div class="cpu-detail">Nice: {{ formatNumber(currentInfo.cpuDetailedPercent[2]) }}%</div>
                        <div class="cpu-detail">
                            {{ $t('home.cpuIdle') }}: {{ formatNumber(currentInfo.cpuDetailedPercent[3]) }}%
                        </div>
                        <div class="cpu-detail">I/O: {{ formatNumber(currentInfo.cpuDetailedPercent[4]) }}%</div>
                        <div class="cpu-detail">
                            {{ $t('home.cpuIrq') }}: {{ formatNumber(currentInfo.cpuDetailedPercent[5]) }}%
                        </div>
                        <div class="cpu-detail">
                            {{ $t('home.cpuSoftirq') }}: {{ formatNumber(currentInfo.cpuDetailedPercent[6]) }}%
                        </div>
                        <div class="cpu-detail">
                            {{ $t('home.cpuSteal') }}: {{ formatNumber(currentInfo.cpuDetailedPercent[7]) }}%
                        </div>
                    </el-space>
                </div>

                <el-button link size="small" type="primary" class="mt-2 mb-2" @click="toggleCpuTop">
                    {{ $t('home.cpuTop') }}
                    <el-icon v-if="!showCpuTop"><ArrowRight /></el-icon>
                    <el-icon v-if="showCpuTop"><ArrowDown /></el-icon>
                </el-button>
                <ComplexTable v-if="showCpuTop" :data="currentInfo.topCPUItems">
                    <el-table-column :min-width="120" show-overflow-tooltip :label="$t('menu.process')" prop="name" />
                    <el-table-column :min-width="60" :label="$t('monitor.percent')" prop="percent">
                        <template #default="{ row }">{{ row.percent.toFixed(2) }}%</template>
                    </el-table-column>
                    <el-table-column :width="80" :label="$t('commons.table.operate')">
                        <template #default="{ row }">
                            <el-button type="primary" link @click="onKill(row)">
                                {{ $t('process.stopProcess') }}
                            </el-button>
                        </template>
                    </el-table-column>
                </ComplexTable>
                <template #reference>
                    <v-charts
                        height="160px"
                        id="cpu"
                        type="pie"
                        :option="chartsOption['cpu']"
                        v-if="chartsOption['cpu']"
                    />
                </template>
            </el-popover>
            <div class="text-center">
                <span class="input-help">
                    ( {{ formatNumber(currentInfo.cpuUsed) }} / {{ currentInfo.cpuTotal }} )
                    {{ $t('commons.units.core', currentInfo.cpuTotal) }}
                </span>
            </div>
        </el-col>
        <el-col v-if="view === 'system'" :xs="6" :sm="6" :md="3" :lg="3" :xl="3" align="center">
            <el-popover
                :hide-after="20"
                :teleported="false"
                :width="480"
                popper-class="dashboard-status-popover"
                :trigger="hasFinePointer ? 'hover' : 'click'"
                v-if="chartsOption['memory']"
                @hide="onMemPopoverHide"
            >
                <el-descriptions direction="vertical" :title="$t('home.mem')" class="ml-1" :column="4" size="small">
                    <el-descriptions-item :label-width="60" :label="$t('home.total')">
                        {{ computeSize(currentInfo.memoryTotal) }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('home.used')">
                        {{ computeSize(currentInfo.memoryUsed) }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('home.free')">
                        {{ computeSize(currentInfo.memoryFree) }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('home.available')">
                        {{ computeSize(currentInfo.memoryAvailable) }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('home.shard')">
                        {{ computeSize(currentInfo.memoryShard) }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('home.cache')">
                        {{ computeSize(currentInfo.memoryCache) }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('home.percent')">
                        {{ formatNumber(currentInfo.memoryUsedPercent) }}%
                    </el-descriptions-item>
                </el-descriptions>

                <el-descriptions
                    v-if="currentInfo.swapMemoryTotal"
                    direction="vertical"
                    :title="$t('home.swapMem')"
                    :column="4"
                    size="small"
                    class="ml-1"
                >
                    <el-descriptions-item :label-width="60" :label="$t('home.total')">
                        {{ computeSize(currentInfo.swapMemoryTotal) }}
                    </el-descriptions-item>
                    <el-descriptions-item :label-width="60" :label="$t('home.used')">
                        {{ computeSize(currentInfo.swapMemoryUsed) }}
                    </el-descriptions-item>
                    <el-descriptions-item :label-width="60" :label="$t('home.free')">
                        {{ computeSize(currentInfo.swapMemoryAvailable) }}
                    </el-descriptions-item>
                    <el-descriptions-item :label-width="60" :label="$t('home.percent')">
                        {{ formatNumber(currentInfo.swapMemoryUsedPercent) }}%
                    </el-descriptions-item>
                </el-descriptions>

                <el-button link size="small" type="primary" class="float-left mb-2" @click="toggleMemTop">
                    {{ $t('home.memTop') }}
                    <el-icon v-if="!showMemTop"><ArrowRight /></el-icon>
                    <el-icon v-if="showMemTop"><ArrowDown /></el-icon>
                </el-button>
                <ComplexTable v-if="showMemTop" :data="currentInfo.topMemItems">
                    <el-table-column :min-width="120" show-overflow-tooltip :label="$t('menu.process')" prop="name" />
                    <el-table-column :min-width="100" :label="$t('monitor.memory')" prop="memory">
                        <template #default="{ row }">
                            {{ computeSize(row.memory) }}
                        </template>
                    </el-table-column>
                    <el-table-column :min-width="80" :label="$t('monitor.percent')" prop="percent">
                        <template #default="{ row }">{{ row.percent.toFixed(2) }}%</template>
                    </el-table-column>
                    <el-table-column :width="80" :label="$t('commons.table.operate')">
                        <template #default="{ row }">
                            <el-button type="primary" link @click="onKill(row)">
                                {{ $t('process.stopProcess') }}
                            </el-button>
                        </template>
                    </el-table-column>
                </ComplexTable>
                <template #reference>
                    <v-charts
                        height="160px"
                        id="memory"
                        type="pie"
                        :option="chartsOption['memory']"
                        v-if="chartsOption['memory']"
                    />
                </template>
            </el-popover>
            <span class="input-help">
                {{ computeSize(currentInfo.memoryUsed) }} / {{ computeSize(currentInfo.memoryTotal) }}
            </span>
        </el-col>
        <template v-for="(item, index) of currentInfo.diskData" :key="index">
            <el-col
                :xs="6"
                :sm="6"
                :md="3"
                :lg="3"
                :xl="3"
                align="center"
                v-if="view === 'system' && (showMore || index < (currentInfo.diskData.length > 5 ? 4 : 5))"
            >
                <el-popover
                    :hide-after="20"
                    :teleported="false"
                    :width="450"
                    popper-class="dashboard-status-popover"
                    v-if="chartsOption[`disk${index}`]"
                >
                    <el-descriptions :column="1" size="small">
                        <el-descriptions-item :label="$t('home.mount')">
                            {{ item.path }}
                        </el-descriptions-item>
                        <el-descriptions-item :label="$t('commons.table.type')">
                            {{ item.type }}
                        </el-descriptions-item>
                        <el-descriptions-item :label="$t('home.fileSystem')">
                            {{ item.device }}
                        </el-descriptions-item>
                    </el-descriptions>
                    <el-descriptions title="Inode" direction="vertical" :column="4" size="small">
                        <el-descriptions-item :label="$t('home.total')">{{ item.inodesTotal }}</el-descriptions-item>
                        <el-descriptions-item :label="$t('home.used')">{{ item.inodesUsed }}</el-descriptions-item>
                        <el-descriptions-item :label="$t('home.free')">{{ item.inodesFree }}</el-descriptions-item>
                        <el-descriptions-item :label="$t('home.percent')">
                            {{ formatNumber(item.inodesUsedPercent) }}%
                        </el-descriptions-item>
                    </el-descriptions>

                    <el-descriptions :title="$t('monitor.disk')" direction="vertical" :column="4" size="small">
                        <el-descriptions-item :label="$t('home.total')">
                            {{ computeSize(item.total) }}
                        </el-descriptions-item>
                        <el-descriptions-item :label="$t('home.used')">
                            {{ computeSize(item.used) }}
                        </el-descriptions-item>
                        <el-descriptions-item :label="$t('home.free')">
                            {{ computeSize(item.free) }}
                        </el-descriptions-item>
                        <el-descriptions-item :label="$t('home.percent')">
                            {{ formatNumber(item.usedPercent) }}%
                        </el-descriptions-item>
                    </el-descriptions>
                    <template #reference>
                        <v-charts
                            @click="routerToFileWithPath(item.path)"
                            height="160px"
                            :id="`disk${index}`"
                            type="pie"
                            :option="chartsOption[`disk${index}`]"
                            v-if="chartsOption[`disk${index}`]"
                        />
                    </template>
                </el-popover>
                <span class="input-help">{{ computeSize(item.used) }} / {{ computeSize(item.total) }}</span>
            </el-col>
        </template>
        <GPUStatus v-if="view === 'gpu'" :current="currentInfo" @open="routerToName('GPU')" />
        <el-col
            :xs="6"
            :sm="6"
            :md="3"
            :lg="3"
            :xl="3"
            align="center"
            v-if="view === 'system' && currentInfo.diskData.length > 5"
        >
            <el-button v-if="!showMore" link type="primary" @click="changeShowMore(true)" class="buttonClass">
                {{ $t('tabs.more') }}
                <el-icon><Bottom /></el-icon>
            </el-button>
            <el-button v-if="showMore" type="primary" link @click="changeShowMore(false)" class="buttonClass">
                {{ $t('tabs.hide') }}
                <el-icon><Top /></el-icon>
            </el-button>
        </el-col>
        <ConfirmDialog ref="confirmConfRef" @confirm="submitKill" />
    </div>
</template>

<script setup lang="ts">
import { Dashboard } from '@/api/interface/dashboard';
import { computeSize } from '@/utils/size';
import i18n from '@/lang';
import { nextTick, onBeforeUnmount, ref, watch } from 'vue';
import GPUStatus from './gpu.vue';
import { useMediaQuery } from '@vueuse/core';
import { routerToFileWithPath, routerToName } from '@/utils/router';
import { stopProcess } from '@/api/modules/process';
import { loadTopCPU, loadTopMem } from '@/api/modules/dashboard';
import { MsgSuccess } from '@/utils/message';

const hasFinePointer = useMediaQuery('(hover: hover) and (pointer: fine)');
const props = defineProps<{ view: 'system' | 'gpu' }>();
const showMore = ref(false);
let cpuPopoverTimer: ReturnType<typeof setTimeout> | null = null;
let memPopoverTimer: ReturnType<typeof setTimeout> | null = null;
let cpuLoading = false;
let memLoading = false;

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
    cpuMhz: 0,
    currentInfo: null,
    prettyDistro: '',
    quickJump: [],
});
const currentInfo = ref<Dashboard.CurrentInfo>({
    uptime: 0,
    timeSinceUptime: '',
    runningTime: {
        days: 0,
        hours: 0,
        minutes: 0,
        seconds: 0,
    },
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
    memoryCache: 0,
    memoryFree: 0,
    memoryShard: 0,
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
    npuData: [],
    xpuData: [],

    topCPUItems: [],
    topMemItems: [],

    netBytesSent: 0,
    netBytesRecv: 0,
    shotTime: new Date(),
});

const cpuShowAll = ref();
const showCpuDetailedPercent = ref(false);
const showCpuTop = ref(false);
const showMemTop = ref(false);
const killProcessID = ref();
const confirmConfRef = ref();

const chartsOption = ref({
    cpu: { title: 'CPU', data: 0 },
    memory: { title: i18n.global.t('monitor.memory'), data: 0 },
    load: { title: i18n.global.t('home.load'), data: 0 },
});

const acceptParams = (current: Dashboard.CurrentInfo, base: Dashboard.BaseInfo): void => {
    normalizeDashboardAccelerators(current);
    currentInfo.value = current;
    baseInfo.value = base;
    chartsOption.value['cpu'] = {
        title: 'CPU',
        data: formatNumber(currentInfo.value.cpuUsedPercent),
    };
    chartsOption.value['memory'] = {
        title: i18n.global.t('monitor.memory'),
        data: formatNumber(currentInfo.value.memoryUsedPercent),
    };
    chartsOption.value['load'] = {
        title: i18n.global.t('home.load'),
        data: formatNumber(currentInfo.value.loadUsagePercent),
    };
    currentInfo.value.diskData = currentInfo.value.diskData || [];
    nextTick(() => {
        for (let i = 0; i < currentInfo.value.diskData.length; i++) {
            let itemPath = currentInfo.value.diskData[i].path;
            itemPath = itemPath.length > 12 ? itemPath.substring(0, 9) + '...' : itemPath;
            chartsOption.value['disk' + i] = {
                title: itemPath,
                data: formatNumber(currentInfo.value.diskData[i].usedPercent),
            };
        }
        showMore.value = localStorage.getItem('dashboard_show') === 'more';
    });
};

const changeShowMore = (show: boolean) => {
    showMore.value = show;
    localStorage.setItem('dashboard_show', show ? 'more' : 'hide');
};

const onKill = async (row: any) => {
    let params = {
        header: i18n.global.t('process.kill'),
        operationInfo: i18n.global.t('process.killHelper'),
        submitInputInfo: i18n.global.t('process.killNow'),
    };
    killProcessID.value = row.pid;
    confirmConfRef.value!.acceptParams(params);
};
const submitKill = async () => {
    await stopProcess({ PID: killProcessID.value }).then(() => {
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    });
};

function loadStatus(val: number) {
    if (val < 30) {
        return i18n.global.t('home.runSmoothly');
    }
    if (val < 70) {
        return i18n.global.t('home.runNormal');
    }
    if (val < 80) {
        return i18n.global.t('home.runSlowly');
    }
    return i18n.global.t('home.runJam');
}

function formatNumber(val: number) {
    return Number(val.toFixed(2));
}

const normalizeDashboardAccelerators = (current: Dashboard.CurrentInfo) => {
    const legacyNPUs = (current.gpuData || [])
        .filter((item) => item.type === 'ascend')
        .map<Dashboard.NPUInfo>((item) => ({
            type: 'ascend',
            index: item.index,
            npuIndex: item.npuIndex,
            chipIndex: item.chipIndex,
            productName: item.productName,
            busID: item.busID || '',
            health: item.performanceState || '',
            temperature: item.temperature || '',
            powerDraw: item.powerDraw || item.powerUsage || '',
            aiCore: item.gpuUtil || '',
            memUsed: item.memUsed || '',
            memTotal: item.memTotal || '',
            memoryUsed: '',
            memoryTotal: '',
            hbmUsed: '',
            hbmTotal: '',
            hugepagesUsed: '',
            hugepagesTotal: '',
        }));
    current.gpuData = (current.gpuData || []).filter((item) => item.type !== 'ascend');
    current.npuData = current.npuData?.length ? current.npuData : legacyNPUs;
};

const toggleCpuTop = async () => {
    showCpuTop.value = !showCpuTop.value;
    if (showCpuTop.value) {
        await loadTopCPUData();
        if (cpuPopoverTimer) {
            clearInterval(Number(cpuPopoverTimer));
        }
        cpuPopoverTimer = setInterval(loadTopCPUData, 5000);
    } else {
        if (cpuPopoverTimer) {
            clearInterval(Number(cpuPopoverTimer));
            cpuPopoverTimer = null;
        }
    }
};

const onCpuPopoverHide = () => {
    showCpuTop.value = false;
    if (cpuPopoverTimer) {
        clearInterval(Number(cpuPopoverTimer));
        cpuPopoverTimer = null;
    }
};

const toggleMemTop = async () => {
    showMemTop.value = !showMemTop.value;
    if (showMemTop.value) {
        await loadTopMemData();
        if (memPopoverTimer) {
            clearInterval(Number(memPopoverTimer));
        }
        memPopoverTimer = setInterval(loadTopMemData, 5000);
    } else {
        if (memPopoverTimer) {
            clearInterval(Number(memPopoverTimer));
            memPopoverTimer = null;
        }
    }
};

const onMemPopoverHide = () => {
    showMemTop.value = false;
    if (memPopoverTimer) {
        clearInterval(Number(memPopoverTimer));
        memPopoverTimer = null;
    }
};

const loadTopCPUData = async () => {
    if (cpuLoading) return;
    cpuLoading = true;
    try {
        const res = await loadTopCPU();
        currentInfo.value.topCPUItems = res.data || [];
    } catch (_error) {
        // ignore load errors
    } finally {
        cpuLoading = false;
    }
};

const loadTopMemData = async () => {
    if (memLoading) return;
    memLoading = true;
    try {
        const res = await loadTopMem();
        currentInfo.value.topMemItems = res.data || [];
    } catch (_error) {
        // ignore load errors
    } finally {
        memLoading = false;
    }
};

watch(
    () => props.view,
    () => {
        onCpuPopoverHide();
        onMemPopoverHide();
    },
);

onBeforeUnmount(() => {
    if (cpuPopoverTimer) {
        clearInterval(Number(cpuPopoverTimer));
        cpuPopoverTimer = null;
    }
    if (memPopoverTimer) {
        clearInterval(Number(memPopoverTimer));
        memPopoverTimer = null;
    }
});

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.buttonClass {
    margin-top: 28%;
    font-size: 14px;
}
.cpu-detail {
    font-size: 12px;
    width: 95px;
}
:deep(.el-descriptions__label) {
    width: 80px;
    background-color: transparent !important;
}
.custom-row {
    display: grid;
    grid-template-columns: repeat(12, 1fr);
    gap: 10px;
    width: 100%;
}
.custom-row .el-col {
    width: 100% !important;
    max-width: 100% !important;
    flex: none !important;
    float: none !important;
    display: block !important;
}
.custom-row .el-col.el-col-xs-6 {
    grid-column: span 6;
}
@media (min-width: 768px) {
    .custom-row .el-col.el-col-sm-6 {
        grid-column: span 6;
    }
}
@media (min-width: 992px) {
    .custom-row .el-col.el-col-md-3 {
        grid-column: span 3;
    }
}
@media (min-width: 1200px) {
    .custom-row .el-col.el-col-lg-3 {
        grid-column: span 3;
    }
}
@media (min-width: 1920px) {
    .custom-row .el-col.el-col-xl-3 {
        grid-column: span 3;
    }
}
@media (max-width: 767px) {
    :deep(.dashboard-status-popover) {
        max-width: calc(100vw - 24px);
        max-height: calc(100vh - 24px);
        max-height: calc(100dvh - 24px);
        overflow: auto;
    }
}
</style>
