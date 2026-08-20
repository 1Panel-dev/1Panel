<template>
    <div>
        <RouterMenu />
        <div>
            <LayoutContent
                v-loading="loading"
                :title="$t('aiTools.gpu.gpu')"
                :divider="true"
                v-if="hasAccelerators && !loading"
            >
                <template #rightToolBar>
                    <TableSetting title="gpu-refresh" @search="refresh()" />
                    <TableRefresh @search="refresh()" />
                </template>
                <template #main>
                    <div class="device-overview">
                        <div class="overview-item">
                            <span>{{ $t('aiTools.gpu.driverVersion') }}</span>
                            <strong>{{ gpuInfo.driverVersion }}</strong>
                        </div>
                        <div v-if="gpuInfo.cudaVersion" class="overview-item">
                            <span>{{ $t('aiTools.gpu.cudaVersion') }}</span>
                            <strong>{{ gpuInfo.cudaVersion }}</strong>
                        </div>
                        <div class="overview-count">{{ deviceCountText }}</div>
                    </div>

                    <div class="gpu-group-list">
                        <section v-for="group in gpuGroups" :key="group.key" class="gpu-group">
                            <div class="group-header">
                                <div class="group-identity">
                                    <strong>{{ group.title }}</strong>
                                </div>
                            </div>

                            <div class="device-card-grid">
                                <el-card
                                    v-for="item in group.devices"
                                    :key="deviceKey(item)"
                                    class="device-card"
                                    shadow="never"
                                >
                                    <template #header>
                                        <div class="device-card-header">
                                            <div class="device-title">
                                                <span
                                                    v-if="item.type === 'ascend'"
                                                    class="status-dot"
                                                    :class="item.health === 'OK' ? 'status-ok' : 'status-error'"
                                                ></span>
                                                <strong>{{ deviceTitle(item) }}</strong>
                                                <span class="card-product-name">· {{ item.productName }}</span>
                                            </div>
                                            <el-button
                                                type="primary"
                                                plain
                                                round
                                                size="small"
                                                class="process-count"
                                                @click.stop="openGPUProcesses(item)"
                                            >
                                                {{ $t('aiTools.gpu.processCount') }}: {{ item.processes?.length || 0 }}
                                            </el-button>
                                        </div>
                                    </template>

                                    <div class="device-metrics">
                                        <div v-if="hasUtilField(item)" class="metric-item metric-primary">
                                            <span>
                                                {{ item.type === 'ascend' ? 'AICore(%)' : $t('aiTools.gpu.gpuUtil') }}
                                            </span>
                                            <strong>{{ deviceUtil(item) }}</strong>
                                            <el-progress
                                                v-if="isAvailable(deviceUtil(item))"
                                                :percentage="percentage(deviceUtil(item))"
                                                :show-text="false"
                                                :stroke-width="5"
                                            />
                                        </div>
                                        <div class="metric-item metric-primary metric-memory">
                                            <span>{{ $t('aiTools.gpu.memoryUsed') }}</span>
                                            <strong>{{ formatMemory(item.memUsed, item.memTotal) }}</strong>
                                            <el-progress
                                                v-if="memoryPercentage(item) !== null"
                                                :percentage="memoryPercentage(item) || 0"
                                                :show-text="false"
                                                :stroke-width="5"
                                            />
                                        </div>
                                        <div class="metric-item metric-temperature">
                                            <span>{{ $t('aiTools.gpu.temperature') }}</span>
                                            <strong>{{ formatTemperature(item.temperature) }}</strong>
                                        </div>
                                    </div>
                                    <div v-if="hasDeviceDetails(item)" class="card-detail">
                                        <div v-if="item.type !== 'ascend'" class="detail-sections">
                                            <section v-if="hasGPURuntimeDetails(item)" class="detail-section">
                                                <div class="detail-section-title">
                                                    {{ $t('aiTools.gpu.runtimeInfo') }}
                                                </div>
                                                <div class="detail-grid">
                                                    <div v-if="isAvailable(item.fanSpeed)" class="detail-item">
                                                        <span class="detail-label">
                                                            {{ $t('aiTools.gpu.fanSpeed') }}
                                                        </span>
                                                        <strong>{{ item.fanSpeed }}</strong>
                                                    </div>
                                                    <div v-if="isAvailable(item.performanceState)" class="detail-item">
                                                        <span class="detail-label cell-item">
                                                            {{ $t('aiTools.gpu.performanceState') }}
                                                            <el-tooltip
                                                                placement="top"
                                                                :content="$t('aiTools.gpu.performanceStateHelper')"
                                                            >
                                                                <el-icon class="icon-item"><InfoFilled /></el-icon>
                                                            </el-tooltip>
                                                        </span>
                                                        <strong>{{ item.performanceState }}</strong>
                                                    </div>
                                                    <div v-if="hasPowerField(item)" class="detail-item">
                                                        <span class="detail-label">
                                                            {{ $t('aiTools.gpu.powerUsage') }}
                                                        </span>
                                                        <strong>{{ formatPower(item) }}</strong>
                                                    </div>
                                                </div>
                                            </section>
                                            <section v-if="hasGPUDeviceDetails(item)" class="detail-section">
                                                <div class="detail-section-title">
                                                    {{ $t('aiTools.gpu.deviceInfo') }}
                                                </div>
                                                <div class="detail-grid">
                                                    <div v-if="isAvailable(item.persistenceMode)" class="detail-item">
                                                        <span class="detail-label cell-item">
                                                            {{ $t('aiTools.gpu.persistenceMode') }}
                                                            <el-tooltip
                                                                placement="top"
                                                                :content="$t('aiTools.gpu.persistenceModeHelper')"
                                                            >
                                                                <el-icon class="icon-item"><InfoFilled /></el-icon>
                                                            </el-tooltip>
                                                        </span>
                                                        <strong>
                                                            {{
                                                                $t('aiTools.gpu.' + item.persistenceMode.toLowerCase())
                                                            }}
                                                        </strong>
                                                    </div>
                                                    <div v-if="hasField(item.busID)" class="detail-item">
                                                        <span class="detail-label">
                                                            {{ $t('aiTools.gpu.busID') }}
                                                        </span>
                                                        <strong>{{ item.busID }}</strong>
                                                    </div>
                                                    <div v-if="isAvailable(item.displayActive)" class="detail-item">
                                                        <span class="detail-label">
                                                            {{ $t('aiTools.gpu.displayActive') }}
                                                        </span>
                                                        <strong>
                                                            {{
                                                                lowerCase(item.displayActive) === 'disabled'
                                                                    ? $t('aiTools.gpu.displayActiveF')
                                                                    : $t('aiTools.gpu.displayActiveT')
                                                            }}
                                                        </strong>
                                                    </div>
                                                    <div v-if="isAvailable(item.ecc)" class="detail-item">
                                                        <span class="detail-label cell-item">
                                                            Uncorr. ECC
                                                            <el-tooltip
                                                                placement="top"
                                                                :content="$t('aiTools.gpu.ecc')"
                                                            >
                                                                <el-icon class="icon-item"><InfoFilled /></el-icon>
                                                            </el-tooltip>
                                                        </span>
                                                        <strong>{{ loadEcc(item.ecc) }}</strong>
                                                    </div>
                                                    <div v-if="isAvailable(item.computeMode)" class="detail-item">
                                                        <span class="detail-label cell-item">
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
                                                        </span>
                                                        <strong>{{ loadComputeMode(item.computeMode) }}</strong>
                                                    </div>
                                                    <div v-if="isAvailable(item.migMode)" class="detail-item">
                                                        <span class="detail-label cell-item">
                                                            MIG M.
                                                            <el-tooltip
                                                                placement="top"
                                                                :content="$t('aiTools.gpu.migModeHelper')"
                                                            >
                                                                <el-icon class="icon-item"><InfoFilled /></el-icon>
                                                            </el-tooltip>
                                                        </span>
                                                        <strong>
                                                            {{
                                                                item.migMode === 'N/A'
                                                                    ? $t('aiTools.gpu.migModeNA')
                                                                    : $t('aiTools.gpu.' + lowerCase(item.migMode))
                                                            }}
                                                        </strong>
                                                    </div>
                                                </div>
                                            </section>
                                        </div>
                                        <div v-else class="detail-sections">
                                            <section v-if="hasNPURuntimeDetails(item)" class="detail-section">
                                                <div class="detail-section-title">
                                                    {{ $t('aiTools.gpu.runtimeInfo') }}
                                                </div>
                                                <div class="detail-grid">
                                                    <div v-if="hasPowerField(item)" class="detail-item">
                                                        <span class="detail-label">
                                                            {{ $t('aiTools.gpu.powerUsage') }}
                                                        </span>
                                                        <strong>{{ formatPower(item) }}</strong>
                                                    </div>
                                                    <div
                                                        v-if="hasUsage(item.hugepagesUsed, item.hugepagesTotal)"
                                                        class="detail-item"
                                                    >
                                                        <span class="detail-label">Hugepages-Usage(page)</span>
                                                        <strong>
                                                            {{ formatUsage(item.hugepagesUsed, item.hugepagesTotal) }}
                                                        </strong>
                                                    </div>
                                                    <div
                                                        v-if="hasUsage(item.hbmUsed, item.hbmTotal)"
                                                        class="detail-item"
                                                    >
                                                        <span class="detail-label">HBM-Usage</span>
                                                        <strong>
                                                            {{ formatMemory(item.hbmUsed, item.hbmTotal) }}
                                                        </strong>
                                                    </div>
                                                </div>
                                            </section>
                                            <section v-if="hasField(item.busID)" class="detail-section">
                                                <div class="detail-section-title">
                                                    {{ $t('aiTools.gpu.deviceInfo') }}
                                                </div>
                                                <div class="detail-grid">
                                                    <div class="detail-item">
                                                        <span class="detail-label">
                                                            {{ $t('aiTools.gpu.busID') }}
                                                        </span>
                                                        <strong>{{ item.busID }}</strong>
                                                    </div>
                                                </div>
                                            </section>
                                        </div>
                                    </div>
                                </el-card>
                            </div>
                        </section>
                        <section v-if="xpuInfo.xpu.length" class="gpu-group">
                            <div class="group-header">
                                <div class="group-identity"><strong>XPU</strong></div>
                            </div>

                            <div class="device-card-grid">
                                <el-card
                                    v-for="item in xpuInfo.xpu"
                                    :key="item.basic.deviceID"
                                    class="device-card"
                                    shadow="never"
                                >
                                    <template #header>
                                        <div class="device-card-header">
                                            <div class="device-title">
                                                <strong>XPU {{ item.basic.deviceID }}</strong>
                                                <span class="card-product-name">· {{ item.basic.deviceName }}</span>
                                            </div>
                                            <el-button
                                                type="primary"
                                                size="small"
                                                plain
                                                round
                                                class="process-count"
                                                @click.stop="openXPUProcesses(item.basic.deviceID)"
                                            >
                                                {{ $t('aiTools.gpu.processCount') }}: {{ item.processes?.length || 0 }}
                                            </el-button>
                                        </div>
                                    </template>

                                    <div class="device-metrics xpu-metrics">
                                        <div class="metric-item metric-primary">
                                            <span>{{ $t('aiTools.gpu.gpuUtil') }}</span>
                                            <strong>{{ item.stats.gpuUtil || 'N/A' }}</strong>
                                            <el-progress
                                                v-if="isAvailable(item.stats.gpuUtil)"
                                                :percentage="percentage(item.stats.gpuUtil)"
                                                :show-text="false"
                                                :stroke-width="5"
                                            />
                                        </div>
                                        <div class="metric-item metric-primary metric-memory">
                                            <span>{{ $t('aiTools.gpu.memoryUsed') }}</span>
                                            <strong>
                                                {{ formatMemory(item.stats.memoryUsed, item.basic.memory) }}
                                            </strong>
                                            <el-progress
                                                v-if="
                                                    memoryPercentageValues(item.stats.memoryUsed, item.basic.memory) !==
                                                    null
                                                "
                                                :percentage="
                                                    memoryPercentageValues(item.stats.memoryUsed, item.basic.memory) ||
                                                    0
                                                "
                                                :show-text="false"
                                                :stroke-width="5"
                                            />
                                        </div>
                                        <div class="metric-item metric-temperature">
                                            <span>{{ $t('aiTools.gpu.temperature') }}</span>
                                            <strong>{{ formatTemperature(item.stats.temperature) }}</strong>
                                        </div>
                                    </div>
                                    <div v-if="hasXPUDetails(item)" class="card-detail">
                                        <div class="detail-sections">
                                            <section v-if="hasXPURuntimeDetails(item)" class="detail-section">
                                                <div class="detail-section-title">
                                                    {{ $t('aiTools.gpu.runtimeInfo') }}
                                                </div>
                                                <div class="detail-grid">
                                                    <div v-if="isAvailable(item.basic.freeMemory)" class="detail-item">
                                                        <span class="detail-label">
                                                            {{ $t('aiTools.gpu.freeMemory') }}
                                                        </span>
                                                        <strong>{{ item.basic.freeMemory }}</strong>
                                                    </div>
                                                    <div v-if="isAvailable(item.stats.power)" class="detail-item">
                                                        <span class="detail-label">
                                                            {{ $t('aiTools.gpu.powerUsage') }}
                                                        </span>
                                                        <strong>{{ item.stats.power }}</strong>
                                                    </div>
                                                    <div v-if="isAvailable(item.stats.frequency)" class="detail-item">
                                                        <span class="detail-label">
                                                            {{ $t('aiTools.gpu.frequency') }}
                                                        </span>
                                                        <strong>{{ item.stats.frequency }}</strong>
                                                    </div>
                                                    <div v-if="isAvailable(item.stats.memoryUtil)" class="detail-item">
                                                        <span class="detail-label">
                                                            {{ $t('aiTools.gpu.memoryUsage') }}
                                                        </span>
                                                        <strong>{{ item.stats.memoryUtil }}</strong>
                                                    </div>
                                                </div>
                                            </section>
                                            <section v-if="hasField(item.basic.pciBdfAddress)" class="detail-section">
                                                <div class="detail-section-title">
                                                    {{ $t('aiTools.gpu.deviceInfo') }}
                                                </div>
                                                <div class="detail-grid">
                                                    <div class="detail-item">
                                                        <span class="detail-label">
                                                            {{ $t('aiTools.gpu.busID') }}
                                                        </span>
                                                        <strong>{{ item.basic.pciBdfAddress }}</strong>
                                                    </div>
                                                </div>
                                            </section>
                                        </div>
                                    </div>
                                </el-card>
                            </div>
                        </section>
                    </div>
                </template>
            </LayoutContent>
        </div>
        <DialogPro v-model="processDrawerVisible" :title="processDrawerTitle" size="large">
            <template v-if="processGPU">
                <el-table v-if="processGPU.processes?.length" :data="processGPU.processes">
                    <el-table-column label="PID" prop="pid" />
                    <el-table-column :label="$t('aiTools.gpu.processName')" prop="processName" />
                    <el-table-column :label="$t('aiTools.gpu.processMemoryUsage')" prop="usedMemory" />
                </el-table>
                <el-empty v-else :description="$t('commons.msg.noneData')" />
            </template>
            <template v-else-if="processXPU">
                <el-table v-if="processXPU.processes?.length" :data="processXPU.processes">
                    <el-table-column label="PID" prop="pid" />
                    <el-table-column :label="$t('aiTools.gpu.processName')" prop="command" />
                    <el-table-column :label="$t('aiTools.gpu.shr')" prop="shr" />
                    <el-table-column :label="$t('aiTools.gpu.processMemoryUsage')" prop="memory" />
                </el-table>
                <el-empty v-else :description="$t('commons.msg.noneData')" />
            </template>
        </DialogPro>
        <LayoutContent :title="$t('aiTools.gpu.gpu')" :divider="true" v-if="!hasAccelerators && !loading">
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
import { computed, onMounted, ref } from 'vue';
import { loadGPUInfo } from '@/api/modules/ai';
import RouterMenu from '@/views/ai/gpu/index.vue';
import { AI } from '@/api/interface/ai';
import i18n from '@/lang';

const loading = ref();
const processDrawerVisible = ref(false);
const processDeviceKey = ref('');
const processXPUId = ref<number | null>(null);
const gpuInfo = ref<AI.Info>({
    cudaVersion: '',
    driverVersion: '',
    type: 'nvidia',
    gpu: [],
    npu: [],
    xpuDriverVersion: '',
    xpu: [],
});
const xpuInfo = ref<AI.XpuInfo>({
    driverVersion: '',
    type: 'xpu',
    xpu: [],
});

type AcceleratorDevice = AI.GPU | AI.NPU;

interface GPUGroup {
    key: string;
    title: string;
    devices: AcceleratorDevice[];
}

const deviceKey = (item: AcceleratorDevice) => {
    return item.type === 'ascend' ? `ascend-${item.npuIndex}-${item.chipIndex}` : `${item.type}-${item.index}`;
};

const groupKey = (item: AcceleratorDevice) => {
    return item.type === 'ascend' ? `ascend-npu-${item.npuIndex}` : `${item.type}-devices`;
};

const gpuGroups = computed<GPUGroup[]>(() => {
    const groups = new Map<string, GPUGroup>();
    for (const item of [...gpuInfo.value.gpu, ...gpuInfo.value.npu]) {
        const key = groupKey(item);
        if (!groups.has(key)) {
            groups.set(key, {
                key,
                title: item.type === 'ascend' ? `NPU ${item.npuIndex}` : item.type.toUpperCase(),
                devices: [],
            });
        }
        const group = groups.get(key)!;
        group.devices.push(item);
    }
    return Array.from(groups.values());
});

const processGPU = computed(() => {
    return [...gpuInfo.value.gpu, ...gpuInfo.value.npu].find((item) => deviceKey(item) === processDeviceKey.value);
});

const processXPU = computed(() => {
    return xpuInfo.value.xpu.find((item) => item.basic.deviceID === processXPUId.value);
});

const processDrawerTitle = computed(() => {
    if (processGPU.value) {
        const item = processGPU.value;
        return `${deviceTitle(item)} · ${item.productName} · ${i18n.global.t('aiTools.gpu.process')}`;
    }
    if (processXPU.value) {
        const item = processXPU.value;
        return `XPU ${item.basic.deviceID} · ${item.basic.deviceName} · ${i18n.global.t('aiTools.gpu.process')}`;
    }
    return i18n.global.t('aiTools.gpu.process');
});

const deviceCountText = computed(() => {
    const parts: string[] = [];
    if (gpuInfo.value.gpu.length) {
        parts.push(`${gpuInfo.value.gpu.length} GPU`);
    }
    if (gpuInfo.value.npu.length) {
        const npuCount = new Set(gpuInfo.value.npu.map((item) => item.npuIndex)).size;
        parts.push(`${npuCount} NPU`, `${gpuInfo.value.npu.length} Chip`);
    }
    if (xpuInfo.value.xpu.length) {
        parts.push(`${xpuInfo.value.xpu.length} XPU`);
    }
    return parts.join(' · ');
});

const hasAccelerators = computed(() => {
    return gpuGroups.value.length > 0 || xpuInfo.value.xpu.length > 0;
});

const normalizeAcceleratorInfo = (data: AI.Info): AI.Info => {
    const devices = (data.gpu || []) as AcceleratorDevice[];
    const legacyNPUs = devices
        .filter((item): item is AI.NPU => item.type === 'ascend')
        .map((item) => {
            const legacyItem = item as AI.NPU & { gpuUtil?: string; performanceState?: string };
            return {
                ...item,
                aiCore: item.aiCore || legacyItem.gpuUtil || '',
                health: item.health || legacyItem.performanceState || '',
            };
        });
    return {
        ...data,
        gpu: devices.filter((item): item is AI.GPU => item.type !== 'ascend'),
        npu: data.npu || legacyNPUs,
        xpuDriverVersion: data.xpuDriverVersion || (data.type === 'xpu' ? data.driverVersion : ''),
        xpu: data.xpu || [],
    };
};

const applyAcceleratorInfo = (data: AI.Info) => {
    const normalized = normalizeAcceleratorInfo(data);
    gpuInfo.value = normalized;
    xpuInfo.value = {
        type: 'xpu',
        driverVersion: normalized.xpuDriverVersion,
        xpu: normalized.xpu,
    };
};

const search = async () => {
    loading.value = true;
    await loadGPUInfo()
        .then((res) => {
            loading.value = false;
            applyAcceleratorInfo(res.data);
        })
        .catch(() => {
            loading.value = false;
        });
};

const refresh = async () => {
    const res = await loadGPUInfo();
    applyAcceleratorInfo(res.data);
};

const openGPUProcesses = (item: AcceleratorDevice) => {
    processDeviceKey.value = deviceKey(item);
    processXPUId.value = null;
    processDrawerVisible.value = true;
};

const openXPUProcesses = (deviceID: number) => {
    processDeviceKey.value = '';
    processXPUId.value = deviceID;
    processDrawerVisible.value = true;
};

const deviceTitle = (item: AcceleratorDevice) => {
    return item.type === 'ascend' ? `Chip ${item.chipIndex}` : `GPU ${item.index}`;
};

const isAvailable = (value?: string) => {
    return Boolean(value && value.trim() !== '' && value.trim().toUpperCase() !== 'N/A');
};

const hasField = (value?: string) => {
    return typeof value === 'string' && value.trim() !== '';
};

interface Quantity {
    value: number;
    unit: string;
}

const parseQuantity = (value: string): Quantity | null => {
    const matched = value.trim().match(/^([+-]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+))\s*(.*?)$/);
    if (!matched) {
        return null;
    }
    const parsed = Number.parseFloat(matched[1]);
    return Number.isFinite(parsed) ? { value: parsed, unit: matched[2].replace(/\s+/g, '').toLowerCase() } : null;
};

const percentage = (value: string) => {
    const parsed = parseQuantity(value);
    if (!parsed || !['', '%', 'percent', 'pct'].includes(parsed.unit)) {
        return 0;
    }
    return Math.min(100, Math.max(0, parsed.value));
};

const formatTemperature = (value: string) => {
    if (!isAvailable(value)) {
        return value || 'N/A';
    }
    return value.replace(/\s*°?C\b/, ' °C');
};

const memoryPercentage = (item: AcceleratorDevice): number | null => {
    return memoryPercentageValues(item.memUsed, item.memTotal);
};

const memoryPercentageValues = (usedValue: string, totalValue: string): number | null => {
    const used = memoryMiB(usedValue);
    const total = memoryMiB(totalValue);
    if (used === null || total === null || total <= 0) {
        return null;
    }
    return Math.min(100, Math.max(0, Number(((used / total) * 100).toFixed(1))));
};

const formatMemory = (usedValue: string, totalValue: string) => {
    const used = memoryMiB(usedValue);
    const total = memoryMiB(totalValue);
    if (used === null || total === null) {
        return `${usedValue} / ${totalValue}`;
    }
    if (total >= 1024) {
        return `${formatQuantity(used / 1024)} / ${formatQuantity(total / 1024)} GiB`;
    }
    return `${formatQuantity(used)} / ${formatQuantity(total)} MiB`;
};

const memoryMiB = (value: string): number | null => {
    const parsed = parseQuantity(value);
    if (!parsed) {
        return null;
    }
    const factors: Record<string, number> = {
        '': 1,
        b: 1 / (1024 * 1024),
        byte: 1 / (1024 * 1024),
        bytes: 1 / (1024 * 1024),
        kb: 1000 / (1024 * 1024),
        kib: 1 / 1024,
        mb: 1_000_000 / (1024 * 1024),
        mib: 1,
        gb: 1_000_000_000 / (1024 * 1024),
        gib: 1024,
        tb: 1_000_000_000_000 / (1024 * 1024),
        tib: 1024 * 1024,
    };
    const factor = factors[parsed.unit];
    return factor === undefined ? null : parsed.value * factor;
};

const formatQuantity = (value: number) => Number(value.toFixed(1)).toString();

const hasUsage = (usedValue?: string, totalValue?: string) => {
    return isAvailable(usedValue) || isAvailable(totalValue);
};

const formatUsage = (usedValue?: string, totalValue?: string) => {
    return `${usedValue || 'N/A'} / ${totalValue || 'N/A'}`;
};

const hasDeviceDetails = (item: AcceleratorDevice) => {
    if (item.type !== 'ascend') {
        return hasGPURuntimeDetails(item) || hasGPUDeviceDetails(item);
    }
    return (
        hasField(item.powerDraw) ||
        hasField(item.busID) ||
        hasUsage(item.hugepagesUsed, item.hugepagesTotal) ||
        hasUsage(item.hbmUsed, item.hbmTotal)
    );
};

const hasGPURuntimeDetails = (item: AI.GPU) => {
    return isAvailable(item.fanSpeed) || isAvailable(item.performanceState) || hasPowerField(item);
};

const hasGPUDeviceDetails = (item: AI.GPU) => {
    return (
        isAvailable(item.persistenceMode) ||
        hasField(item.busID) ||
        isAvailable(item.displayActive) ||
        isAvailable(item.ecc) ||
        isAvailable(item.computeMode) ||
        isAvailable(item.migMode)
    );
};

const hasXPUDetails = (item: AI.XpuInfo['xpu'][number]) => {
    return (
        hasField(item.basic.pciBdfAddress) ||
        isAvailable(item.basic.freeMemory) ||
        isAvailable(item.stats.power) ||
        isAvailable(item.stats.frequency) ||
        isAvailable(item.stats.memoryUtil)
    );
};

const hasNPURuntimeDetails = (item: AI.NPU) => {
    return (
        hasPowerField(item) ||
        hasUsage(item.hugepagesUsed, item.hugepagesTotal) ||
        hasUsage(item.hbmUsed, item.hbmTotal)
    );
};

const hasXPURuntimeDetails = (item: AI.XpuInfo['xpu'][number]) => {
    return (
        isAvailable(item.basic.freeMemory) ||
        isAvailable(item.stats.power) ||
        isAvailable(item.stats.frequency) ||
        isAvailable(item.stats.memoryUtil)
    );
};

const deviceUtil = (item: AcceleratorDevice) => {
    return item.type === 'ascend' ? item.aiCore : item.gpuUtil;
};

const hasUtilField = (item: AcceleratorDevice) => {
    return item.type === 'ascend' ? hasField(item.aiCore) : true;
};

const hasPowerField = (item: AcceleratorDevice) => {
    return item.type === 'ascend' ? hasField(item.powerDraw) : isAvailable(item.powerDraw);
};

const formatPower = (item: AcceleratorDevice) => {
    if (item.type === 'ascend') {
        return item.powerDraw;
    }
    return isAvailable(item.maxPowerLimit) ? `${item.powerDraw} / ${item.maxPowerLimit}` : item.powerDraw;
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

onMounted(() => {
    search();
});
</script>

<style lang="scss" scoped>
.device-overview {
    display: flex;
    align-items: center;
    gap: 0;
    min-height: 44px;
    padding: 9px 14px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 8px;
}
.overview-item {
    display: flex;
    align-items: center;
    gap: 8px;
    white-space: nowrap;
    & + .overview-item {
        margin-left: 16px;
        padding-left: 16px;
        border-left: 1px solid var(--el-border-color-lighter);
    }
    span {
        color: var(--el-text-color-secondary);
        font-size: 12px;
        line-height: 18px;
    }
    strong {
        color: var(--el-text-color-primary);
        font-size: 14px;
        font-weight: 600;
        line-height: 18px;
    }
}
.overview-count {
    margin-left: 16px;
    padding-left: 16px;
    border-left: 1px solid var(--el-border-color-lighter);
    color: var(--el-color-primary);
    font-size: 12px;
    font-weight: 600;
    line-height: 18px;
    white-space: nowrap;
}
.gpu-group-list {
    display: flex;
    flex-direction: column;
    gap: 16px;
    margin-top: 16px;
}
.gpu-group {
    padding-top: 2px;
}
.gpu-group + .gpu-group {
    padding-top: 18px;
    border-top: 1px solid var(--el-border-color-lighter);
}
.group-header,
.device-card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
}
.group-header {
    justify-content: flex-start;
    gap: 8px;
    margin-bottom: 14px;
}
.device-card-header {
    width: 100%;
    min-width: 0;
}
.group-identity,
.device-title {
    display: flex;
    align-items: center;
    gap: 8px;
}
.group-identity strong {
    color: var(--el-text-color-regular);
    font-size: 14px;
    font-weight: 600;
    line-height: 20px;
}
.card-product-name {
    color: var(--el-text-color-secondary);
    font-size: 12px;
}
.status-dot {
    width: 8px;
    height: 8px;
    flex: 0 0 8px;
    border-radius: 50%;
}
.status-ok {
    background: var(--el-color-success);
    box-shadow: 0 0 0 3px var(--el-color-success-light-9);
}
.status-error {
    background: var(--el-color-danger);
    box-shadow: 0 0 0 3px var(--el-color-danger-light-9);
}
.device-card-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
}
.device-card {
    box-sizing: border-box;
    width: 100%;
    min-width: 0;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 10px;
    outline: none;
    background: var(--el-bg-color);
    transition: border-color 0.2s;
    &:hover {
        border-color: var(--el-color-primary-light-5);
    }
    :deep(.el-card__header) {
        display: flex;
        align-items: center;
        box-sizing: border-box;
        min-height: 51px;
        padding: 13px 16px;
    }
    :deep(.el-card__body) {
        padding: 16px;
    }
}
.device-card:only-child {
    grid-column: 1 / -1;
}
.device-title {
    flex: 1;
    align-items: center;
    min-width: 0;
    strong {
        color: var(--el-text-color-primary);
        font-size: 16px;
        font-weight: 600;
        line-height: 20px;
        white-space: nowrap;
    }
}
.card-product-name {
    overflow: hidden;
    line-height: 20px;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.process-count {
    flex: 0 0 auto;
    align-self: center;
    margin-left: 12px;
    font-size: 12px;
    white-space: nowrap;
}
.device-metrics {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 16px 20px;
}
.metric-item {
    min-width: 0;
    span {
        display: block;
        margin-bottom: 6px;
        color: var(--el-text-color-secondary);
        font-size: 12px;
        line-height: 18px;
    }
    strong {
        display: block;
        overflow: hidden;
        color: var(--el-text-color-primary);
        font-size: 14px;
        font-weight: 500;
        font-variant-numeric: tabular-nums;
        line-height: 26px;
        text-overflow: ellipsis;
        white-space: nowrap;
    }
    :deep(.el-progress) {
        margin-top: 8px;
    }
}
.metric-primary {
    strong {
        font-size: 20px;
        font-weight: 600;
        line-height: 26px;
    }
    :deep(.el-progress) {
        margin-top: 10px;
    }
}
.metric-memory {
    min-width: 0;
}
.metric-temperature {
    position: relative;
    &::before {
        position: absolute;
        top: 0;
        bottom: 0;
        left: -10px;
        border-left: 1px solid var(--el-border-color-lighter);
        content: '';
    }
    strong {
        font-size: 18px;
        font-weight: 600;
        line-height: 26px;
    }
}
.xpu-metrics {
    grid-template-columns: repeat(3, minmax(0, 1fr));
}
.card-detail {
    margin-top: 16px;
    padding-top: 14px;
    border-top: 1px solid var(--el-border-color-lighter);
}
.detail-section + .detail-section {
    margin-top: 14px;
    padding-top: 14px;
    border-top: 1px solid var(--el-border-color-lighter);
}
.detail-section-title {
    display: flex;
    align-items: center;
    gap: 7px;
    margin-bottom: 10px;
    color: var(--el-text-color-regular);
    font-size: 12px;
    font-weight: 600;
    line-height: 18px;
    &::before {
        width: 3px;
        height: 12px;
        border-radius: 2px;
        background: var(--el-color-primary);
        content: '';
    }
}
.detail-grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 14px 20px;
}
.detail-item {
    min-width: 0;
    .detail-label {
        display: block;
        margin-bottom: 5px;
        overflow: hidden;
        color: var(--el-text-color-secondary);
        font-size: 12px;
        line-height: 18px;
        text-overflow: ellipsis;
        white-space: nowrap;
    }
    strong {
        display: block;
        overflow: hidden;
        color: var(--el-text-color-primary);
        font-size: 13px;
        font-weight: 500;
        line-height: 20px;
        text-overflow: ellipsis;
        white-space: nowrap;
    }
}
.cell-item {
    display: flex;
    align-items: center;
    .icon-item {
        margin-left: 4px;
        margin-top: -1px;
    }
}
.detail-item .detail-label.cell-item {
    display: flex;
}

@media (max-width: 1200px) {
    .device-card-grid {
        grid-template-columns: 1fr;
    }
}

@media (max-width: 768px) {
    .device-overview {
        flex-wrap: wrap;
        gap: 8px 0;
        padding: 10px 12px;
    }
    .overview-count {
        width: auto;
    }
    .device-metrics {
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }
    .metric-temperature {
        grid-column: 1 / -1;
        padding-top: 12px;
        border-top: 1px solid var(--el-border-color-lighter);
        &::before {
            display: none;
        }
    }
    .detail-grid {
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }
}

@media (max-width: 480px) {
    .detail-grid {
        grid-template-columns: 1fr;
    }
}
</style>
