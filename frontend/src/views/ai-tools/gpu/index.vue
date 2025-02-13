<template>
    <div>
        <RouterButton
            :buttons="[
                {
                    label: $t('ai_tools.gpu.gpu'),
                    path: '/xpack/gpu',
                },
            ]"
        />

        <div v-if="gpuType == 'nvidia'">
            <LayoutContent
                v-loading="loading"
                :title="$t('ai_tools.gpu.gpu')"
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
                        <el-descriptions-item :label="$t('ai_tools.gpu.driverVersion')" width="50%" :span="7">
                            {{ gpuInfo.driverVersion }}
                        </el-descriptions-item>
                        <el-descriptions-item :label="$t('ai_tools.gpu.cudaVersion')" :span="7">
                            {{ gpuInfo.cudaVersion }}
                        </el-descriptions-item>
                    </el-descriptions>
                    <el-collapse v-model="activeNames" class="mt-5">
                        <el-collapse-item v-for="item in gpuInfo.gpu" :key="item.index" :name="item.index">
                            <template #title>
                                <span class="name-class">{{ item.index + '. ' + item.productName }}</span>
                            </template>
                            <span class="title-class">{{ $t('ai_tools.gpu.base') }}</span>
                            <el-descriptions direction="vertical" :column="6" border size="small" class="mt-2">
                                <el-descriptions-item :label="$t('monitor.gpuUtil')">
                                    {{ item.gpuUtil }}
                                </el-descriptions-item>
                                <el-descriptions-item>
                                    <template #label>
                                        <div class="cell-item">
                                            {{ $t('monitor.temperature') }}
                                            <el-tooltip placement="top" :content="$t('ai_tools.gpu.temperatureHelper')">
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
                                                :content="$t('ai_tools.gpu.performanceStateHelper')"
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

                                <el-descriptions-item :label="$t('ai_tools.gpu.busID')">
                                    {{ item.busID }}
                                </el-descriptions-item>
                                <el-descriptions-item>
                                    <template #label>
                                        <div class="cell-item">
                                            {{ $t('ai_tools.gpu.persistenceMode') }}
                                            <el-tooltip
                                                placement="top"
                                                :content="$t('ai_tools.gpu.persistenceModeHelper')"
                                            >
                                                <el-icon class="icon-item"><InfoFilled /></el-icon>
                                            </el-tooltip>
                                        </div>
                                    </template>
                                    {{ $t('ai_tools.gpu.' + item.persistenceMode.toLowerCase()) }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('ai_tools.gpu.displayActive')">
                                    {{
                                        lowerCase(item.displayActive) === 'disabled'
                                            ? $t('ai_tools.gpu.displayActiveF')
                                            : $t('ai_tools.gpu.displayActiveT')
                                    }}
                                </el-descriptions-item>
                                <el-descriptions-item>
                                    <template #label>
                                        <div class="cell-item">
                                            Uncorr. ECC
                                            <el-tooltip placement="top" :content="$t('ai_tools.gpu.ecc')">
                                                <el-icon class="icon-item"><InfoFilled /></el-icon>
                                            </el-tooltip>
                                        </div>
                                    </template>
                                    {{ loadEcc(item.ecc) }}
                                </el-descriptions-item>
                                <el-descriptions-item :label="$t('ai_tools.gpu.computeMode')">
                                    <template #label>
                                        <div class="cell-item">
                                            {{ $t('ai_tools.gpu.computeMode') }}
                                            <el-tooltip placement="top">
                                                <template #content>
                                                    {{ $t('ai_tools.gpu.defaultHelper') }}
                                                    <br />
                                                    {{ $t('ai_tools.gpu.exclusiveProcessHelper') }}
                                                    <br />
                                                    {{ $t('ai_tools.gpu.exclusiveThreadHelper') }}
                                                    <br />
                                                    {{ $t('ai_tools.gpu.prohibitedHelper') }}
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
                                                    {{ $t('ai_tools.gpu.migModeHelper') }}
                                                </template>
                                                <el-icon class="icon-item"><InfoFilled /></el-icon>
                                            </el-tooltip>
                                        </div>
                                    </template>
                                    {{
                                        item.migMode === 'N/A'
                                            ? $t('ai_tools.gpu.migModeNA')
                                            : $t('ai_tools.gpu.' + lowerCase(item.migMode))
                                    }}
                                </el-descriptions-item>
                            </el-descriptions>
                            <div class="mt-5">
                                <span class="title-class">{{ $t('ai_tools.gpu.process') }}</span>
                            </div>
                            <el-table :data="item.processes" v-if="item.processes?.length !== 0">
                                <el-table-column label="PID" prop="pid" />
                                <el-table-column :label="$t('ai_tools.gpu.type')" prop="type">
                                    <template #default="{ row }">
                                        {{ loadProcessType(row.type) }}
                                    </template>
                                </el-table-column>
                                <el-table-column :label="$t('ai_tools.gpu.processName')" prop="processName" />
                                <el-table-column :label="$t('ai_tools.gpu.processMemoryUsage')" prop="usedMemory" />
                            </el-table>
                        </el-collapse-item>
                    </el-collapse>
                </template>
            </LayoutContent>
        </div>
        <div v-else>
            <LayoutContent
                v-loading="loading"
                :title="$t('ai_tools.gpu.gpu')"
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
                        <el-descriptions-item :label="$t('ai_tools.gpu.driverVersion')" width="50%" :span="7">
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
                            <span class="title-class">{{ $t('ai_tools.gpu.base') }}</span>
                            <el-descriptions direction="vertical" :column="6" border size="small" class="mt-2">
                                <el-descriptions-item :label="$t('monitor.gpuUtil')">
                                    {{ item.stats.memoryUtil }}
                                </el-descriptions-item>
                                <el-descriptions-item>
                                    <template #label>
                                        <div class="cell-item">
                                            {{ $t('monitor.temperature') }}
                                            <el-tooltip placement="top" :content="$t('ai_tools.gpu.temperatureHelper')">
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
                                <el-descriptions-item :label="$t('ai_tools.gpu.busID')">
                                    {{ item.basic.pciBdfAddress }}
                                </el-descriptions-item>
                            </el-descriptions>
                            <div class="mt-5">
                                <span class="title-class">{{ $t('ai_tools.gpu.process') }}</span>
                            </div>
                            <el-table :data="item.processes" v-if="item.processes?.length !== 0">
                                <el-table-column label="PID" prop="pid" />
                                <el-table-column :label="$t('ai_tools.gpu.processName')" prop="command" />
                                <el-table-column :label="$t('ai_tools.gpu.shr')" prop="shr" />
                                <el-table-column :label="$t('ai_tools.gpu.processMemoryUsage')" prop="memory" />
                            </el-table>
                        </el-collapse-item>
                    </el-collapse>
                </template>
            </LayoutContent>
        </div>
        <LayoutContent
            :title="$t('ai_tools.gpu.gpu')"
            :divider="true"
            v-if="gpuInfo.driverVersion.length === 0 && xpuInfo.driverVersion.length == 0 && !loading"
        >
            <template #main>
                <div class="app-warn">
                    <div class="flx-center">
                        <span>{{ $t('ai_tools.gpu.gpuHelper') }}</span>
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
import { loadGPUInfo } from '@/api/modules/ai-tool';
import { AITool } from '@/api/interface/ai-tool';
import i18n from '@/lang';

const loading = ref();
const activeNames = ref(0);
const gpuInfo = ref<AITool.Info>({
    cudaVersion: '',
    driverVersion: '',
    type: 'nvidia',
    gpu: [],
});
const xpuInfo = ref<AITool.XpuInfo>({
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
            return i18n.global.t('ai_tools.gpu.default');
        case 'Exclusive Process':
            return i18n.global.t('ai_tools.gpu.exclusiveProcess');
        case 'Exclusive Thread':
            return i18n.global.t('ai_tools.gpu.exclusiveThread');
        case 'Prohibited':
            return i18n.global.t('ai_tools.gpu.prohibited');
    }
};

const loadEcc = (val: string) => {
    if (val === 'N/A') {
        return i18n.global.t('ai_tools.gpu.migModeNA');
    }
    if (val === 'Disabled') {
        return i18n.global.t('ai_tools.gpu.disabled');
    }
    if (val === 'Enabled') {
        return i18n.global.t('ai_tools.gpu.enabled');
    }
    return val;
};

const loadProcessType = (val: string) => {
    if (val === 'C' || val === 'G') {
        return i18n.global.t('ai_tools.gpu.type' + val);
    }
    if (val === 'C+G') {
        return i18n.global.t('ai_tools.gpu.typeCG');
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
