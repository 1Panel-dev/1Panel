<template>
    <el-row :gutter="10">
        <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6" align="center">
            <el-popover placement="bottom" :width="200" trigger="hover" v-if="chartsOption['load']">
                <el-tag class="tagClass">{{ $t('home.loadAverage', 1) }}: {{ formatNumber(currentInfo.load1) }}</el-tag>
                <el-tag class="tagClass">{{ $t('home.loadAverage', 5) }}: {{ formatNumber(currentInfo.load5) }}</el-tag>
                <el-tag class="tagClass">
                    {{ $t('home.loadAverage', 15) }}: {{ formatNumber(currentInfo.load15) }}
                </el-tag>
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
        <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6" align="center">
            <el-popover placement="bottom" :width="loadWidth()" trigger="hover" v-if="chartsOption['cpu']">
                <div>
                    <el-tooltip
                        effect="dark"
                        :content="baseInfo.cpuModelName"
                        v-if="baseInfo.cpuModelName.length > 40"
                        placement="top"
                    >
                        <el-tag class="cpuModeTag">
                            {{ baseInfo.cpuModelName.substring(0, 40) + '...' }}
                        </el-tag>
                    </el-tooltip>
                    <el-tag v-else>
                        {{ baseInfo.cpuModelName }}
                    </el-tag>
                </div>
                <el-tag class="cpuDetailTag">{{ $t('home.core') }} *{{ baseInfo.cpuCores }}</el-tag>
                <el-tag class="cpuDetailTag">{{ $t('home.logicCore') }} *{{ baseInfo.cpuLogicalCores }}</el-tag>
                <br />
                <div v-for="(item, index) of currentInfo.cpuPercent" :key="index">
                    <el-tag v-if="cpuShowAll || (!cpuShowAll && index < 32)" class="tagCPUClass">
                        CPU-{{ index }}: {{ formatNumber(item) }}%
                    </el-tag>
                </div>

                <div v-if="currentInfo.cpuPercent.length > 32" class="mt-1 float-right">
                    <el-button v-if="!cpuShowAll" @click="cpuShowAll = true" link type="primary" size="small">
                        {{ $t('commons.button.showAll') }}
                        <el-icon><DArrowRight /></el-icon>
                    </el-button>
                    <el-button v-if="cpuShowAll" @click="cpuShowAll = false" link type="primary" size="small">
                        {{ $t('commons.button.hideSome') }}
                        <el-icon><DArrowLeft /></el-icon>
                    </el-button>
                </div>
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
            <span class="input-help">
                ( {{ formatNumber(currentInfo.cpuUsed) }} / {{ currentInfo.cpuTotal }} )
                {{ $t('commons.units.core', currentInfo.cpuTotal) }}
            </span>
        </el-col>
        <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6" align="center">
            <el-popover placement="bottom" :width="160" trigger="hover" v-if="chartsOption['memory']">
                <el-tag style="font-weight: 500">{{ $t('home.mem') }}:</el-tag>
                <el-tag class="tagClass">{{ $t('home.total') }}: {{ computeSize(currentInfo.memoryTotal) }}</el-tag>
                <el-tag class="tagClass">{{ $t('home.used') }}: {{ computeSize(currentInfo.memoryUsed) }}</el-tag>
                <el-tag class="tagClass">{{ $t('home.free') }}: {{ computeSize(currentInfo.memoryAvailable) }}</el-tag>
                <el-tag class="tagClass">
                    {{ $t('home.percent') }}: {{ formatNumber(currentInfo.memoryUsedPercent) }}%
                </el-tag>
                <div v-if="currentInfo.swapMemoryTotal" class="mt-2">
                    <el-tag style="font-weight: 500">{{ $t('home.swapMem') }}:</el-tag>
                    <el-tag class="tagClass">
                        {{ $t('home.total') }}: {{ computeSize(currentInfo.swapMemoryTotal) }}
                    </el-tag>
                    <el-tag class="tagClass">
                        {{ $t('home.used') }}: {{ computeSize(currentInfo.swapMemoryUsed) }}
                    </el-tag>
                    <el-tag class="tagClass">
                        {{ $t('home.free') }}: {{ computeSize(currentInfo.swapMemoryAvailable) }}
                    </el-tag>
                    <el-tag class="tagClass">
                        {{ $t('home.percent') }}: {{ formatNumber(currentInfo.swapMemoryUsedPercent) }}%
                    </el-tag>
                </div>
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
            <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6" align="center" v-if="isShow('disk', index)">
                <el-popover placement="bottom" :width="300" trigger="hover" v-if="chartsOption[`disk${index}`]">
                    <el-row :gutter="5">
                        <el-tag style="font-weight: 500">{{ $t('home.baseInfo') }}:</el-tag>
                    </el-row>
                    <el-row :gutter="5">
                        <el-tag class="nameTag">{{ $t('home.mount') }}: {{ item.path }}</el-tag>
                    </el-row>
                    <el-row :gutter="5">
                        <el-tag class="tagClass">{{ $t('commons.table.type') }}: {{ item.type }}</el-tag>
                    </el-row>
                    <el-row :gutter="5">
                        <el-tag class="tagClass">{{ $t('home.fileSystem') }}: {{ item.device }}</el-tag>
                    </el-row>
                    <el-row :gutter="5">
                        <el-col :span="12">
                            <div><el-tag class="tagClass" style="font-weight: 500">Inode:</el-tag></div>
                            <el-tag class="tagClass">{{ $t('home.total') }}: {{ item.inodesTotal }}</el-tag>
                            <el-tag class="tagClass">{{ $t('home.used') }}: {{ item.inodesUsed }}</el-tag>
                            <el-tag class="tagClass">{{ $t('home.free') }}: {{ item.inodesFree }}</el-tag>
                            <el-tag class="tagClass">
                                {{ $t('home.percent') }}: {{ formatNumber(item.inodesUsedPercent) }}%
                            </el-tag>
                        </el-col>

                        <el-col :span="12">
                            <div>
                                <el-tag style="margin-top: 3px; font-weight: 500">{{ $t('monitor.disk') }}:</el-tag>
                            </div>
                            <el-tag class="tagClass">{{ $t('home.total') }}: {{ computeSize(item.total) }}</el-tag>
                            <el-tag class="tagClass">{{ $t('home.used') }}: {{ computeSize(item.used) }}</el-tag>
                            <el-tag class="tagClass">{{ $t('home.free') }}: {{ computeSize(item.free) }}</el-tag>
                            <el-tag class="tagClass">
                                {{ $t('home.percent') }}: {{ formatNumber(item.usedPercent) }}%
                            </el-tag>
                        </el-col>
                    </el-row>
                    <template #reference>
                        <v-charts
                            @click="toFolder(item.path)"
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
        <template v-for="(item, index) of currentInfo.gpuData" :key="index">
            <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6" align="center" v-if="isShow('gpu', index)">
                <el-popover placement="bottom" :width="250" trigger="hover" v-if="chartsOption[`gpu${index}`]">
                    <el-row :gutter="5">
                        <el-tag style="font-weight: 500">{{ $t('home.baseInfo') }}:</el-tag>
                    </el-row>
                    <el-row :gutter="5">
                        <el-tag class="tagClass">{{ $t('monitor.gpuUtil') }}: {{ item.gpuUtil }}</el-tag>
                    </el-row>
                    <el-row :gutter="5">
                        <el-tag class="tagClass">
                            {{ $t('monitor.temperature') }}: {{ item.temperature.replaceAll('C', '°C') }}
                        </el-tag>
                    </el-row>
                    <el-row :gutter="5">
                        <el-tag class="tagClass">
                            {{ $t('monitor.performanceState') }}: {{ item.performanceState }}
                        </el-tag>
                    </el-row>
                    <el-row :gutter="5">
                        <el-tag class="tagClass">{{ $t('monitor.powerUsage') }}: {{ item.powerUsage }}</el-tag>
                    </el-row>
                    <el-row :gutter="5">
                        <el-tag class="tagClass">{{ $t('monitor.memoryUsage') }}: {{ item.memoryUsage }}</el-tag>
                    </el-row>
                    <el-row :gutter="5">
                        <el-tag class="tagClass">{{ $t('monitor.fanSpeed') }}: {{ item.fanSpeed }}</el-tag>
                    </el-row>
                    <template #reference>
                        <v-charts
                            @click="goGPU()"
                            height="160px"
                            :id="`gpu${index}`"
                            type="pie"
                            :option="chartsOption[`gpu${index}`]"
                            v-if="chartsOption[`gpu${index}`]"
                        />
                    </template>
                </el-popover>
                <el-tooltip :content="item.productName" v-if="item.productName.length > 25">
                    <span class="input-help">{{ item.productName.substring(0, 22) }}...</span>
                </el-tooltip>
                <span class="input-help" v-else>{{ item.productName }}</span>
            </el-col>
        </template>
        <template v-for="(item, index) of currentInfo.xpuData" :key="index">
            <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6" align="center" v-if="isShow('xpu', index)">
                <el-popover placement="bottom" :width="250" trigger="hover" v-if="chartsOption[`xpu${index}`]">
                    <el-row :gutter="5">
                        <el-tag style="font-weight: 500">{{ $t('home.baseInfo') }}:</el-tag>
                    </el-row>
                    <el-row :gutter="5">
                        <el-tag class="tagClass">{{ $t('monitor.gpuUtil') }}: {{ item.memoryUtil }}</el-tag>
                    </el-row>
                    <el-row :gutter="5">
                        <el-tag class="tagClass">{{ $t('monitor.temperature') }}: {{ item.temperature }}</el-tag>
                    </el-row>
                    <el-row :gutter="5">
                        <el-tag class="tagClass">{{ $t('monitor.powerUsage') }}: {{ item.power }}</el-tag>
                    </el-row>
                    <el-row :gutter="5">
                        <el-tag class="tagClass">
                            {{ $t('monitor.memoryUsage') }}: {{ item.memoryUsed }}/{{ item.memory }}
                        </el-tag>
                    </el-row>
                    <template #reference>
                        <v-charts
                            @click="goGPU()"
                            height="160px"
                            :id="`xpu${index}`"
                            type="pie"
                            :option="chartsOption[`xpu${index}`]"
                            v-if="chartsOption[`xpu${index}`]"
                        />
                    </template>
                </el-popover>
                <el-tooltip :content="item.deviceName" v-if="item.deviceName.length > 25">
                    <span class="input-help">{{ item.deviceName.substring(0, 22) }}...</span>
                </el-tooltip>
                <span class="input-help" v-else>{{ item.deviceName }}</span>
            </el-col>
        </template>
        <el-col :xs="12" :sm="12" :md="6" :lg="6" :xl="6" align="center" v-if="totalCount > 5">
            <el-button v-if="!showMore" link type="primary" @click="changeShowMore(true)" class="buttonClass">
                {{ $t('tabs.more') }}
                <el-icon><Bottom /></el-icon>
            </el-button>
            <el-button v-if="showMore" type="primary" link @click="changeShowMore(false)" class="buttonClass">
                {{ $t('tabs.hide') }}
                <el-icon><Top /></el-icon>
            </el-button>
        </el-col>
    </el-row>
</template>

<script setup lang="ts">
import { Dashboard } from '@/api/interface/dashboard';
import { computeSize } from '@/utils/util';
import router from '@/routers';
import i18n from '@/lang';
import { nextTick, ref } from 'vue';
import { toFolder } from '@/global/business';
const showMore = ref(false);
const totalCount = ref();

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
    ipV4Addr: '',
    httpProxy: '',

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

const cpuShowAll = ref();

const chartsOption = ref({
    cpu: { title: 'CPU', data: 0 },
    memory: { title: i18n.global.t('monitor.memory'), data: 0 },
    load: { title: i18n.global.t('home.load'), data: 0 },
});

const acceptParams = (current: Dashboard.CurrentInfo, base: Dashboard.BaseInfo): void => {
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
        currentInfo.value.gpuData = currentInfo.value.gpuData || [];
        for (let i = 0; i < currentInfo.value.gpuData.length; i++) {
            chartsOption.value['gpu' + i] = {
                title: 'GPU-' + currentInfo.value.gpuData[i].index,
                data: formatNumber(Number(currentInfo.value.gpuData[i].gpuUtil.replaceAll(' %', ''))),
            };
        }
        currentInfo.value.xpuData = currentInfo.value.xpuData || [];
        for (let i = 0; i < currentInfo.value.xpuData.length; i++) {
            chartsOption.value['xpu' + i] = {
                title: 'XPU-' + currentInfo.value.xpuData[i].deviceID,
                data: formatNumber(Number(currentInfo.value.xpuData[i].memoryUtil.replaceAll('%', ''))),
            };
        }
        totalCount.value =
            currentInfo.value.diskData.length + currentInfo.value.gpuData.length + currentInfo.value.xpuData.length;
        showMore.value = localStorage.getItem('dashboard_show') === 'more';
    });
};

const isShow = (val: string, index: number) => {
    let showCount = totalCount.value < 6 ? 5 : 4;
    switch (val) {
        case 'disk':
            return showMore.value || index < showCount;
        case 'gpu':
            let gpuCount = showCount - currentInfo.value.diskData.length;
            return showMore.value || index < gpuCount;
        case 'xpu':
            let xpuCount = showCount - currentInfo.value.diskData.length - currentInfo.value.gpuData.length;
            return showMore.value || index < xpuCount;
    }
};

const changeShowMore = (show: boolean) => {
    showMore.value = show;
    localStorage.setItem('dashboard_show', show ? 'more' : 'hide');
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

const goGPU = () => {
    router.push({ name: 'GPU' });
};

const loadWidth = () => {
    if (!cpuShowAll.value || currentInfo.value.cpuPercent.length < 32) {
        return 310;
    }
    let line = Math.floor(currentInfo.value.cpuPercent.length / 16);
    return line * 141 + 28;
};

function formatNumber(val: number) {
    return Number(val.toFixed(2));
}

defineExpose({
    acceptParams,
});
</script>

<style scoped lang="scss">
.cpuModeTag {
    justify-content: flex-start !important;
    text-align: left !important;
    width: 280px;
}
.cpuDetailTag {
    justify-content: flex-start !important;
    text-align: left !important;
    width: 140px;
    margin-top: 3px;
    margin-left: 1px;
}
.tagClass {
    margin-top: 3px;
}

.tagCPUClass {
    justify-content: flex-start !important;
    text-align: left !important;
    float: left;
    margin-top: 3px;
    margin-left: 1px;
    width: 140px;
}

.buttonClass {
    margin-top: 28%;
    font-size: 14px;
}
.nameTag {
    margin-top: 3px;
    height: auto;
    display: inline-block;
    white-space: normal;
    line-height: 1.8;
}
</style>
