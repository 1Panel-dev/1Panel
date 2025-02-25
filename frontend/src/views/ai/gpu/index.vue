<template>
    <div>
        <RouterButton
            :buttons="[
                {
                    label: $t('aiTools.gpu.gpu'),
                    path: '/xpack/gpu',
                },
            ]"
        />

        <div v-if="gpuType == 'nvidia'">
            <LayoutContent
                v-loading="loading"
                :title="$t('aiTools.gpu.gpu')"
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
                        <el-descriptions-item :label="$t('aiTools.gpu.driverVersion')" width="50%" :span="7">
                            {{ gpuInfo.driverVersion }}
                        </el-descriptions-item>
                        <el-descriptions-item :label="$t('aiTools.gpu.cudaVersion')" :span="7">
                            {{ gpuInfo.cudaVersion }}
                        </el-descriptions-item>
                    </el-descriptions>
                    <el-collapse v-model="activeNames" class="mt-5">
                        <el-collapse-item v-for="item in gpuInfo.gpu" :key="item.index" :name="item.index">
                            <template #title>
                                <span class="name-class">{{ item.index + '. ' + item.productName }}</span>
                            </template>
                            <span class="title-class">{{ $t('aiTools.gpu.base') }}</span>
                            <el-descriptions direction="vertical" :column="6" border size="small" class="mt-2">
                                <el-descriptions-item :label="$t('monitor.gpuUtil')">
                                    {{ item.gpuUtil }}
                                </el-descriptions-item>
                                <el-descriptions-item>
                                    <template #label>
                                        <div class="cell-item">
                                            {{ $t('monitor.temperature') }}
                                            <el-tooltip placement="top" :content="$t('aiTools.gpu.temperatureHelper')">
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
                                            <el-tooltip
                                                placement="top"
                                                :content="$t('aiTools.gpu.performanceStateHelper')"
                                            >
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

                                <el-descriptions-item :label="$t('aiTools.gpu.busID')">
                                    {{ item.busID }}
                                </el-descriptions-item>
                                <el-descriptions-item>
                                    <template #label>
                                        <div class="cell-item">
                                            {{ $t('aiTools.gpu.persistenceMode') }}
                                            <el-tooltip
                                                placement="top"
                                                :content="$t('aiTools.gpu.persistenceModeHelper')"
                                            >
                                                <el-icon class="icon-item"><InfoFilled /></el-icon>
                                            </el-tooltip>
                                        </div>
                                    </template>
                                    {{ $t('aiTools.gpu.' + item.persistenceMode.toLowerCase()) }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('aiTools.gpu.displayActive')">
                                    {{
                                        lowerCase(item.displayActive) === 'disabled'
                                            ? $t('aiTools.gpu.displayActiveF')
                                            : $t('aiTools.gpu.displayActiveT')
                                    }}
                                </el-descriptions-item>
                                <el-descriptions-item>
                                    <template #label>
                                        <div class="cell-item">
                                            Uncorr. ECC
                                            <el-tooltip placement="top" :content="$t('aiTools.gpu.ecc')">
                                                <el-icon class="icon-item"><InfoFilled /></el-icon>
                                            </el-tooltip>
                                        </div>
                                    </template>
                                    {{ loadEcc(item.ecc) }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('aiTools.gpu.computeMode')">
                                    <template #label>
                                        <div class="cell-item">
                                            {{ $t('aiTools.gpu.computeMode') }}
                                            <el-tooltip placement="top">
                                                <template #content>
                                                    {{ $t('aiTools.gpu.defaultHelper') }}
                                                    <br />
                                                    {{ $t('aiTools.gpu.exclusiveProcessHelper') }}
                                                    <br />
                                                    {{ $t('aiTools.gpu.exclusiveThreadHelper') }}
                                                    <br />
                                                    {{ $t('aiTools.gpu.prohibitedHelper') }}
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
                                                    {{ $t('aiTools.gpu.migModeHelper') }}
                                                </template>
                                                <el-icon class="icon-item"><InfoFilled /></el-icon>
                                            </el-tooltip>
                                        </div>
                                    </template>
                                    {{
                                        item.migMode === 'N/A'
                                            ? $t('aiTools.gpu.migModeNA')
                                            : $t('aiTools.gpu.' + lowerCase(item.migMode))
                                    }}
                                </el-descriptions-item>
                            </el-descriptions>
                            <div class="mt-5">
                                <span class="title-class">{{ $t('aiTools.gpu.process') }}</span>
                            </div>
                            <el-table :data="item.processes" v-if="item.processes?.length !== 0">
                                <el-table-column label="PID" prop="pid" />
                                <el-table-column :label="$t('aiTools.gpu.type')" prop="type">
                                    <template #default="{ row }">
                                        {{ loadProcessType(row.type) }}
                                    </template>
                                </el-table-column>
                                <el-table-column :label="$t('aiTools.gpu.processName')" prop="processName" />
                                <el-table-column :label="$t('aiTools.gpu.processMemoryUsage')" prop="usedMemory" />
                            </el-table>
                        </el-collapse-item>
                    </el-collapse>
                </template>
            </LayoutContent>
        </div>
        <div v-else>
            <LayoutContent
                v-loading="loading"
                :title="$t('aiTools.gpu.gpu')"
                :divider="true"
                v-if="xpuInfo.driverVersion.length !== 0 && !loading"
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
                        <el-descriptions-item :label="$t('aiTools.gpu.driverVersion')" width="50%" :span="7">
                            {{ xpuInfo.driverVersion }}
                        </el-descriptions-item>
                    </el-descriptions>
                    <el-collapse v-model="activeNames" class="mt-5">
                        <el-collapse-item
                            v-for="item in xpuInfo.xpu"
                            :key="item.basic.deviceID"
                            :name="item.basic.deviceID"
                        >
                            <template #title>
                                <span class="name-class">{{ item.basic.deviceID + '. ' + item.basic.deviceName }}</span>
                            </template>
                            <span class="title-class">{{ $t('aiTools.gpu.base') }}</span>
                            <el-descriptions direction="vertical" :column="6" border size="small" class="mt-2">
                                <el-descriptions-item :label="$t('monitor.gpuUtil')">
                                    {{ item.stats.memoryUtil }}
                                </el-descriptions-item>
                                <el-descriptions-item>
                                    <template #label>
                                        <div class="cell-item">
                                            {{ $t('monitor.temperature') }}
                                            <el-tooltip placement="top" :content="$t('aiTools.gpu.temperatureHelper')">
                                                <el-icon class="icon-item"><InfoFilled /></el-icon>
                                            </el-tooltip>
                                        </div>
                                    </template>
                                    {{ item.stats.temperature }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('monitor.powerUsage')">
                                    {{ item.stats.power }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('monitor.memoryUsage')">
                                    {{ item.stats.memoryUsed }} / {{ item.basic.memory }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('aiTools.gpu.busID')">
                                    {{ item.basic.pciBdfAddress }}
                                </el-descriptions-item>
                            </el-descriptions>
                            <div class="mt-5">
                                <span class="title-class">{{ $t('aiTools.gpu.process') }}</span>
                            </div>
                            <el-table :data="item.processes" v-if="item.processes?.length !== 0">
                                <el-table-column label="PID" prop="pid" />
                                <el-table-column :label="$t('aiTools.gpu.processName')" prop="command" />
                                <el-table-column :label="$t('aiTools.gpu.shr')" prop="shr" />
                                <el-table-column :label="$t('aiTools.gpu.processMemoryUsage')" prop="memory" />
                            </el-table>
                        </el-collapse-item>
                    </el-collapse>
                </template>
            </LayoutContent>
        </div>
        <LayoutContent
            :title="$t('aiTools.gpu.gpu')"
            :divider="true"
            v-if="gpuInfo.driverVersion.length === 0 && xpuInfo.driverVersion.length == 0 && !loading"
        >
            <template #main>
                <div class="app-warn">
                    <div class="flx-center">
                        <span>{{ $t('aiTools.gpu.gpuHelper') }}</span>
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
import { loadGPUInfo } from '@/api/modules/ai';
import { AI } from '@/api/interface/ai';
import i18n from '@/lang';

const loading = ref();
const activeNames = ref(0);
const gpuInfo = ref<AI.Info>({
    cudaVersion: '',
    driverVersion: '',
    type: 'nvidia',
    gpu: [],
});
const xpuInfo = ref<AI.XpuInfo>({
    driverVersion: '',
    type: 'xpu',
    xpu: [],
});
const gpuType = ref('nvidia');

const search = async () => {
    loading.value = true;
    await loadGPUInfo()
        .then((res) => {
            loading.value = false;
            gpuType.value = res.data.type;
            if (res.data.type == 'nvidia') {
                gpuInfo.value = res.data;
            } else {
                xpuInfo.value = res.data;
            }
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
            return i18n.global.t('aiTools.gpu.default');
        case 'Exclusive Process':
            return i18n.global.t('aiTools.gpu.exclusiveProcess');
        case 'Exclusive Thread':
            return i18n.global.t('aiTools.gpu.exclusiveThread');
        case 'Prohibited':
            return i18n.global.t('aiTools.gpu.prohibited');
    }
};

const loadEcc = (val: string) => {
    if (val === 'N/A') {
        return i18n.global.t('aiTools.gpu.migModeNA');
    }
    if (val === 'Disabled') {
        return i18n.global.t('aiTools.gpu.disabled');
    }
    if (val === 'Enabled') {
        return i18n.global.t('aiTools.gpu.enabled');
    }
    return val || 0;
};

const loadProcessType = (val: string) => {
    if (val === 'C' || val === 'G') {
        return i18n.global.t('aiTools.gpu.type' + val);
    }
    if (val === 'C+G') {
        return i18n.global.t('aiTools.gpu.typeCG');
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
