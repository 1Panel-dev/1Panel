<template>
    <div>
        <el-card class="el-card">
            <template #header>
                <div class="card-header">
                    <span>{{ $t('commons.table.status') }}</span>
                </div>
            </template>
            <el-row :gutter="10">
                <el-col :span="12" align="center">
                    <el-popover placement="bottom" :width="300" trigger="hover">
                        <div style="margin-bottom: 10px">
                            <el-tag>{{ baseInfo.cpuModelName }}</el-tag>
                        </div>
                        <el-tag>
                            {{ $t('home.core') }} *{{ baseInfo.cpuCores }}； {{ $t('home.logicCore') }} *{{
                                baseInfo.cpuLogicalCores
                            }}
                        </el-tag>
                        <br />
                        <el-tag style="margin-top: 5px" v-for="(item, index) of currentInfo.cpuPercent" :key="index">
                            CPU-{{ index }}: {{ formatNumber(item) }}%
                        </el-tag>
                        <template #reference>
                            <el-progress
                                type="dashboard"
                                :width="80"
                                :percentage="formatNumber(currentInfo.cpuUsedPercent)"
                            >
                                <template #default="{ percentage }">
                                    <span class="percentage-value">{{ percentage }}%</span>
                                    <span class="percentage-label">CPU</span>
                                </template>
                            </el-progress>
                        </template>
                    </el-popover>
                    <br />
                    <span>( {{ formatNumber(currentInfo.cpuUsed) }} / {{ currentInfo.cpuTotal }} ) Core</span>
                </el-col>
                <el-col :span="12" align="center">
                    <el-progress type="dashboard" :width="80" :percentage="formatNumber(currentInfo.MemoryUsedPercent)">
                        <template #default="{ percentage }">
                            <span class="percentage-value">{{ percentage }}%</span>
                            <span class="percentage-label">{{ $t('monitor.memory') }}</span>
                        </template>
                    </el-progress>
                    <br />
                    <span>
                        ( {{ formatNumber(currentInfo.memoryUsed / 1024 / 1024) }} /
                        {{ currentInfo.memoryTotal / 1024 / 1024 }} ) MB
                    </span>
                </el-col>
            </el-row>
            <el-row :gutter="10" style="margin-top: 30px">
                <el-col :span="12" align="center">
                    <el-popover placement="bottom" :width="200" trigger="hover">
                        <el-tag style="margin-top: 5px">
                            {{ $t('home.loadAverage', [1]) }}: {{ formatNumber(currentInfo.load1) }}
                        </el-tag>
                        <el-tag style="margin-top: 5px">
                            {{ $t('home.loadAverage', [5]) }}: {{ formatNumber(currentInfo.load5) }}
                        </el-tag>
                        <el-tag style="margin-top: 5px">
                            {{ $t('home.loadAverage', [15]) }}: {{ formatNumber(currentInfo.load15) }}
                        </el-tag>
                        <template #reference>
                            <el-progress
                                type="dashboard"
                                :width="80"
                                :percentage="formatNumber(currentInfo.loadUsagePercent)"
                            >
                                <template #default="{ percentage }">
                                    <span class="percentage-value">{{ percentage }}%</span>
                                    <span class="percentage-label">{{ $t('home.load') }}</span>
                                </template>
                            </el-progress>
                        </template>
                    </el-popover>
                    <br />
                    <span>{{ loadStatus(currentInfo.loadUsagePercent) }}</span>
                </el-col>
                <el-col :span="12" align="center">
                    <el-popover placement="bottom" :width="160" trigger="hover">
                        <el-tag>{{ $t('home.mount') }}: /</el-tag>
                        <div><el-tag style="margin-top: 10px">iNode</el-tag></div>
                        <el-tag style="margin-top: 5px">{{ $t('home.total') }}: {{ currentInfo.inodesTotal }}</el-tag>
                        <el-tag style="margin-top: 3px">{{ $t('home.used') }}: {{ currentInfo.inodesUsed }}</el-tag>
                        <el-tag style="margin-top: 3px">{{ $t('home.free') }}: {{ currentInfo.inodesFree }}</el-tag>
                        <el-tag style="margin-top: 3px">
                            {{ $t('home.percent') }}: {{ formatNumber(currentInfo.inodesUsedPercent) }}%
                        </el-tag>

                        <div>
                            <el-tag style="margin-top: 10px">{{ $t('monitor.disk') }}</el-tag>
                        </div>
                        <el-tag style="margin-top: 5px">
                            {{ $t('home.total') }}: {{ formatNumber(currentInfo.total / 1024 / 1024 / 1024) }} GB
                        </el-tag>
                        <el-tag style="margin-top: 3px">
                            {{ $t('home.used') }}: {{ formatNumber(currentInfo.used / 1024 / 1024 / 1024) }} GB
                        </el-tag>
                        <el-tag style="margin-top: 3px">
                            {{ $t('home.free') }}: {{ formatNumber(currentInfo.free / 1024 / 1024 / 1024) }} GB
                        </el-tag>
                        <el-tag style="margin-top: 3px">
                            {{ $t('home.percent') }}: {{ formatNumber(currentInfo.usedPercent) }}%
                        </el-tag>
                        <template #reference>
                            <el-progress
                                type="dashboard"
                                :width="80"
                                :percentage="formatNumber(currentInfo.usedPercent)"
                            >
                                <template #default="{ percentage }">
                                    <span class="percentage-value">{{ percentage }}%</span>
                                    <span class="percentage-label">{{ $t('monitor.disk') }}</span>
                                </template>
                            </el-progress>
                        </template>
                    </el-popover>

                    <br />
                    <span>
                        ( {{ formatNumber(currentInfo.used / 1024 / 1024 / 1024) }} /
                        {{ formatNumber(currentInfo.total / 1024 / 1024 / 1024) }} ) GB
                    </span>
                </el-col>
            </el-row>
        </el-card>
    </div>
</template>

<script lang="ts" setup>
import { Dashboard } from '@/api/interface/dashboard';
import { ref } from 'vue';

const baseInfo = ref<Dashboard.BaseInfo>({
    haloInstatllID: 0,
    dateeaseInstatllID: 0,
    jumpserverInstatllID: 0,
    metersphereInstatllID: 0,

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
    ioTime: 0,
    ioCount: 0,

    total: 0,
    free: 0,
    used: 0,
    usedPercent: 0,

    inodesTotal: 0,
    inodesUsed: 0,
    inodesFree: 0,
    inodesUsedPercent: 0,

    netBytesSent: 0,
    netBytesRecv: 0,
    shotTime: new Date(),
});

const acceptParams = (current: Dashboard.CurrentInfo, base: Dashboard.BaseInfo): void => {
    currentInfo.value = current;
    baseInfo.value = base;
};

function formatNumber(val: number) {
    return Number(val.toFixed(2));
}

function loadStatus(val: number) {
    if (val < 30) {
        return '运行流畅';
    }
    if (val < 70) {
        return '运行正常';
    }
    if (val < 80) {
        return '运行缓慢';
    }
    return '运行堵塞';
}
defineExpose({
    acceptParams,
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
</style>
