<template>
    <div>
        <MonitorRouter />

        <LayoutContent
            v-loading="loading"
            title="GPU"
            :divider="true"
            v-if="gpuInfo.driverVersion.length !== 0 && !loading"
        >
            <template #toolbar>
                <el-row>
                    <el-col :xs="24" :sm="16" :md="16" :lg="16" :xl="16" />
                    <el-col :xs="24" :sm="8" :md="8" :lg="8" :xl="8">
                        <TableSetting @search="refresh()" />
                    </el-col>
                </el-row>
            </template>
            <template #main>
                <el-descriptions direction="vertical" :column="14" border>
                    <el-descriptions-item :label="$t('monitor.driverVersion')" width="50%" :span="7">
                        {{ gpuInfo.driverVersion }}
                    </el-descriptions-item>
                    <el-descriptions-item :label="$t('monitor.cudaVersion')" :span="7">
                        {{ gpuInfo.cudaVersion }}
                    </el-descriptions-item>
                </el-descriptions>
                <el-collapse v-model="activeNames" class="mt-5">
                    <el-collapse-item v-for="item in gpuInfo.gpu" :key="item.index" :name="item.index">
                        <template #title>
                            <span class="name-class">{{ item.index + '. ' + item.productName }}</span>
                        </template>
                        <span class="title-class">{{ $t('monitor.base') }}</span>
                        <el-descriptions direction="vertical" :column="6" border size="small" class="mt-2">
                            <el-descriptions-item :label="$t('monitor.gpuUtil')">
                                {{ item.gpuUtil }}
                            </el-descriptions-item>
                            <el-descriptions-item>
                                <template #label>
                                    <div class="cell-item">
                                        {{ $t('monitor.temperature') }}
                                        <el-tooltip placement="top" :content="$t('monitor.temperatureHelper')">
                                            <el-icon class="icon-item"><InfoFilled /></el-icon>
                                        </el-tooltip>
                                    </div>
                                </template>
                                {{ item.temperature.replaceAll('C', '°C') }}
                            </el-descriptions-item>
                            <el-descriptions-item>
                                <template #label>
                                    <div class="cell-item">
                                        {{ $t('monitor.performanceState') }}
                                        <el-tooltip placement="top" :content="$t('monitor.performanceStateHelper')">
                                            <el-icon class="icon-item"><InfoFilled /></el-icon>
                                        </el-tooltip>
                                    </div>
                                </template>
                                {{ item.performanceState }}
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('monitor.powerUsage')">
                                {{ item.powerDraw }} / {{ item.maxPowerLimit }}
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('monitor.memoryUsage')">
                                {{ item.memUsed }} / {{ item.memTotal }}
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('monitor.fanSpeed')">
                                {{ item.fanSpeed }}
                            </el-descriptions-item>

                            <el-descriptions-item :label="$t('monitor.busID')">
                                {{ item.busID }}
                            </el-descriptions-item>
                            <el-descriptions-item>
                                <template #label>
                                    <div class="cell-item">
                                        {{ $t('monitor.persistenceMode') }}
                                        <el-tooltip placement="top" :content="$t('monitor.persistenceModeHelper')">
                                            <el-icon class="icon-item"><InfoFilled /></el-icon>
                                        </el-tooltip>
                                    </div>
                                </template>
                                {{ $t('monitor.' + item.persistenceMode.toLowerCase()) }}
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('monitor.displayActive')">
                                {{
                                    lowerCase(item.displayActive) === 'disabled'
                                        ? $t('monitor.displayActiveF')
                                        : $t('monitor.displayActiveT')
                                }}
                            </el-descriptions-item>
                            <el-descriptions-item>
                                <template #label>
                                    <div class="cell-item">
                                        Uncorr. ECC
                                        <el-tooltip placement="top" :content="$t('monitor.ecc')">
                                            <el-icon class="icon-item"><InfoFilled /></el-icon>
                                        </el-tooltip>
                                    </div>
                                </template>
                                {{ loadEcc(item.ecc) }}
                            </el-descriptions-item>
                            <el-descriptions-item :label="$t('monitor.computeMode')">
                                <template #label>
                                    <div class="cell-item">
                                        {{ $t('monitor.computeMode') }}
                                        <el-tooltip placement="top">
                                            <template #content>
                                                {{ $t('monitor.defaultHelper') }}
                                                <br />
                                                {{ $t('monitor.exclusiveProcessHelper') }}
                                                <br />
                                                {{ $t('monitor.exclusiveThreadHelper') }}
                                                <br />
                                                {{ $t('monitor.prohibitedHelper') }}
                                            </template>
                                            <el-icon class="icon-item"><InfoFilled /></el-icon>
                                        </el-tooltip>
                                    </div>
                                </template>
                                {{ loadComputeMode(item.computeMode) }}
                            </el-descriptions-item>
                            <el-descriptions-item label="MIG.M">
                                <template #label>
                                    <div class="cell-item">
                                        MIG M.
                                        <el-tooltip placement="top">
                                            <template #content>
                                                {{ $t('monitor.migModeHelper') }}
                                            </template>
                                            <el-icon class="icon-item"><InfoFilled /></el-icon>
                                        </el-tooltip>
                                    </div>
                                </template>
                                {{
                                    item.migMode === 'N/A'
                                        ? $t('monitor.migModeNA')
                                        : $t('monitor.' + lowerCase(item.migMode))
                                }}
                            </el-descriptions-item>
                        </el-descriptions>
                        <div class="mt-5">
                            <span class="title-class">{{ $t('monitor.process') }}</span>
                        </div>
                        <el-table :data="item.processes" v-if="item.processes?.length !== 0">
                            <el-table-column label="PID" prop="pid" />
                            <el-table-column :label="$t('monitor.type')" prop="type">
                                <template #default="{ row }">
                                    {{ loadProcessType(row.type) }}
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('monitor.processName')" prop="processName" />
                            <el-table-column :label="$t('monitor.processMemoryUsage')" prop="usedMemory" />
                        </el-table>
                    </el-collapse-item>
                </el-collapse>
            </template>
        </LayoutContent>

        <LayoutContent title="GPU" :divider="true" v-if="gpuInfo.driverVersion.length === 0 && !loading">
            <template #main>
                <div class="app-warn">
                    <div class="flx-center">
                        <span>{{ $t('monitor.gpuHelper') }}</span>
                    </div>
                    <div>
                        <img src="@/assets/images/no_app.svg" />
                    </div>
                </div>
            </template>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { loadGPUInfo } from '@/api/modules/host';
import MonitorRouter from '@/views/host/monitor/index.vue';
import { Host } from '@/api/interface/host';
import i18n from '@/lang';

const loading = ref();
const activeNames = ref(0);
const gpuInfo = ref<Host.GPUInfo>({
    cudaVersion: '',
    driverVersion: '',
    gpu: [],
});

const search = async () => {
    loading.value = true;
    await loadGPUInfo()
        .then((res) => {
            loading.value = false;
            gpuInfo.value = res.data;
        })
        .catch(() => {
            loading.value = false;
        });
};

const refresh = async () => {
    const res = await loadGPUInfo();
    gpuInfo.value = res.data;
};

const lowerCase = (val: string) => {
    return val.toLowerCase();
};

const loadComputeMode = (val: string) => {
    switch (val) {
        case 'Default':
            return i18n.global.t('monitor.default');
        case 'Exclusive Process':
            return i18n.global.t('monitor.exclusiveProcess');
        case 'Exclusive Thread':
            return i18n.global.t('monitor.exclusiveThread');
        case 'Prohibited':
            return i18n.global.t('monitor.prohibited');
    }
};

const loadEcc = (val: string) => {
    if (val === '0') {
        return i18n.global.t('monitor.disabled');
    }
    if (val === '1') {
        return i18n.global.t('monitor.enabled');
    }
    if (val === 'N/A') {
        return i18n.global.t('monitor.migModeNA');
    }
    return val;
};

const loadProcessType = (val: string) => {
    if (val === 'C' || val === 'G') {
        return i18n.global.t('monitor.type' + val);
    }
    if (val === 'C+G') {
        return i18n.global.t('monitor.typeCG');
    }
    return val;
};

onMounted(() => {
    search();
});
</script>

<style lang="scss" scoped>
.name-class {
    font-size: 18px;
    font-weight: 500;
}
.title-class {
    font-size: 14px;
    font-weight: 500;
}
.cell-item {
    display: flex;
    align-items: center;
    .icon-item {
        margin-left: 4px;
        margin-top: -1px;
    }
}
</style>
