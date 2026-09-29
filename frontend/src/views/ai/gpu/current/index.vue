<template>
    <div>
        <RouterMenu />
        <div>
            <LayoutContent v-loading="loading" :title="$t('aiTools.gpu.gpu')" v-if="hasAccelerators">
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
                        <div v-if="gpuInfo.collectedAt" class="overview-item">
                            <span>{{ $t('aiTools.gpu.collectedAt') }}</span>
                            <strong>{{ new Date(gpuInfo.collectedAt).toLocaleString() }}</strong>
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
                                                    :class="
                                                        item.health === 'OK'
                                                            ? 'status-ok'
                                                            : isAvailable(item.health)
                                                              ? 'status-error'
                                                              : ''
                                                    "
                                                ></span>
                                                <strong>{{ deviceTitle(item) }}</strong>
                                                <span class="card-product-name">· {{ item.productName }}</span>
                                            </div>
                                            <div class="device-actions">
                                                <el-button
                                                    type="primary"
                                                    plain
                                                    round
                                                    size="small"
                                                    class="process-count"
                                                    @click.stop="openGPUProcesses(item)"
                                                >
                                                    {{ $t('aiTools.gpu.processCount') }}:
                                                    {{
                                                        item.processStatus === 'unavailable'
                                                            ? 'N/A'
                                                            : item.processes?.length || 0
                                                    }}
                                                </el-button>
                                                <el-button
                                                    type="primary"
                                                    link
                                                    @click="detailDeviceKey = deviceKey(item)"
                                                >
                                                    {{ $t('commons.button.view') }}
                                                </el-button>
                                            </div>
                                        </div>
                                    </template>

                                    <div class="device-metrics">
                                        <div class="metric-item metric-primary">
                                            <span>
                                                {{ item.type === 'ascend' ? 'AICore(%)' : $t('aiTools.gpu.gpuUtil') }}
                                            </span>
                                            <strong>{{ deviceUtil(item) || 'N/A' }}</strong>
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
                                    <div class="device-secondary-metrics">
                                        <div class="metric-item">
                                            <span>{{ $t('aiTools.gpu.powerUsage') }}</span>
                                            <strong>{{ formatPower(item) || 'N/A' }}</strong>
                                        </div>
                                        <div class="metric-item">
                                            <span>
                                                {{
                                                    item.type === 'ascend'
                                                        ? $t('aiTools.gpu.aiCPUUtil')
                                                        : $t('aiTools.gpu.frequency')
                                                }}
                                            </span>
                                            <strong>
                                                {{
                                                    (item.type === 'ascend' ? item.aiCPUUtil : item.frequency) || 'N/A'
                                                }}
                                            </strong>
                                        </div>
                                        <div class="metric-item">
                                            <span>
                                                {{
                                                    item.type === 'ascend'
                                                        ? $t('commons.table.status')
                                                        : $t('aiTools.gpu.fanSpeed')
                                                }}
                                            </span>
                                            <strong v-if="item.type === 'ascend'">{{ item.health || 'N/A' }}</strong>
                                            <strong v-else>
                                                {{
                                                    [item.fanSpeed, item.fanRPM].filter(isAvailable).join(' / ') ||
                                                    'N/A'
                                                }}
                                            </strong>
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
                                            <div class="device-actions">
                                                <el-button
                                                    type="primary"
                                                    size="small"
                                                    plain
                                                    round
                                                    class="process-count"
                                                    @click.stop="openXPUProcesses(item.basic.deviceID)"
                                                >
                                                    {{ $t('aiTools.gpu.processCount') }}:
                                                    {{
                                                        item.processStatus === 'unavailable'
                                                            ? 'N/A'
                                                            : item.processes?.length || 0
                                                    }}
                                                </el-button>
                                                <el-button
                                                    type="primary"
                                                    link
                                                    @click="detailDeviceKey = `xpu-${item.basic.deviceID}`"
                                                >
                                                    {{ $t('commons.button.view') }}
                                                </el-button>
                                            </div>
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
                                    <div class="device-secondary-metrics">
                                        <div class="metric-item">
                                            <span>{{ $t('aiTools.gpu.powerUsage') }}</span>
                                            <strong>{{ item.stats.power || 'N/A' }}</strong>
                                        </div>
                                        <div class="metric-item">
                                            <span>{{ $t('aiTools.gpu.frequency') }}</span>
                                            <strong>{{ item.stats.frequency || 'N/A' }}</strong>
                                        </div>
                                        <div class="metric-item">
                                            <span>{{ $t('aiTools.gpu.memoryTemperature') }}</span>
                                            <strong>{{ formatTemperature(item.stats.memoryTemperature || '') }}</strong>
                                        </div>
                                    </div>
                                </el-card>
                            </div>
                        </section>
                    </div>
                </template>
            </LayoutContent>
        </div>
        <DrawerPro
            :model-value="Boolean(detailDeviceKey)"
            :header="$t('commons.button.view')"
            :resource="deviceDetail?.title || ''"
            size="large"
            @close="detailDeviceKey = ''"
        >
            <DeviceDetails
                v-if="deviceDetail?.sections.length"
                :key="detailDeviceKey"
                :sections="deviceDetail.sections"
            />
            <el-empty v-else :description="$t('commons.msg.noneData')" />
        </DrawerPro>
        <DialogPro v-model="processDrawerVisible" :title="processDrawerTitle" size="large">
            <template v-if="processGPU">
                <el-alert
                    v-if="processGPU.processStatus === 'unavailable'"
                    :title="$t('aiTools.gpu.processUnavailable')"
                    type="warning"
                    :closable="false"
                />
                <el-table v-if="processGPU.processes?.length" :data="processGPU.processes">
                    <el-table-column label="PID" prop="pid" />
                    <el-table-column :label="$t('aiTools.gpu.processName')" prop="processName" />
                    <el-table-column :label="$t('aiTools.gpu.processMemoryUsage')" prop="usedMemory" />
                </el-table>
                <el-empty
                    v-else-if="processGPU.processStatus !== 'unavailable'"
                    :description="$t('commons.msg.noneData')"
                />
            </template>
            <template v-else-if="processXPU">
                <el-alert
                    v-if="processXPU.processStatus === 'unavailable'"
                    :title="$t('aiTools.gpu.processUnavailable')"
                    type="warning"
                    :closable="false"
                />
                <el-table v-if="processXPU.processes?.length" :data="processXPU.processes">
                    <el-table-column label="PID" prop="pid" />
                    <el-table-column :label="$t('aiTools.gpu.processName')" prop="command" />
                    <el-table-column :label="$t('aiTools.gpu.shr')" prop="shr" />
                    <el-table-column :label="$t('aiTools.gpu.processMemoryUsage')" prop="memory" />
                </el-table>
                <el-empty
                    v-else-if="processXPU.processStatus !== 'unavailable'"
                    :description="$t('commons.msg.noneData')"
                />
            </template>
        </DialogPro>
        <LayoutContent :title="$t('aiTools.gpu.gpu')" :divider="true" v-if="!hasAccelerators && !loading">
            <template #rightToolBar>
                <TableRefresh @search="refresh()" />
            </template>
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
import { MsgWarning } from '@/utils/message';
import DeviceDetails, { type DetailSection } from './components/DeviceDetails.vue';

const loading = ref();
const processDrawerVisible = ref(false);
const processDeviceKey = ref('');
const detailDeviceKey = ref('');
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
    return item.type === 'ascend'
        ? `ascend-${item.busID || item.npuIndex}-${item.chipIndex}`
        : `${item.type}-${item.uuid || item.busID || item.index}`;
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

const deviceDetail = computed(() => {
    const device = [...gpuInfo.value.gpu, ...gpuInfo.value.npu].find(
        (item) => deviceKey(item) === detailDeviceKey.value,
    );
    if (device) {
        const title =
            device.type === 'ascend' ? `NPU ${device.npuIndex} · ${deviceTitle(device)}` : deviceTitle(device);
        return { title: `${title} · ${device.productName}`, sections: deviceDetailSections(device) };
    }
    const xpu = xpuInfo.value.xpu.find((item) => `xpu-${item.basic.deviceID}` === detailDeviceKey.value);
    if (xpu) {
        return {
            title: `XPU ${xpu.basic.deviceID} · ${xpu.basic.deviceName}`,
            sections: xpuDetailSections(xpu),
        };
    }
    return null;
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
    if (loading.value) return;
    loading.value = true;
    try {
        const res = await loadGPUInfo();
        applyAcceleratorInfo(res.data);
        if (res.data.warnings?.length) {
            MsgWarning(res.data.warnings.join('; '));
        }
    } catch {
    } finally {
        loading.value = false;
    }
};

const refresh = search;

const visibleSections = (sections: DetailSection[]) => {
    return sections
        .map((section) => ({
            ...section,
            items: section.items.filter((item) => isAvailable(item.value)),
        }))
        .filter((section) => section.items.length);
};

const deviceDetailSections = (item: AcceleratorDevice): DetailSection[] => {
    const t = i18n.global.t;
    if (item.type === 'ascend') {
        return visibleSections([
            {
                id: 'runtime',
                title: t('aiTools.gpu.runtimeInfo'),
                items: [
                    { label: t('aiTools.gpu.powerUsage'), value: item.powerDraw },
                    { label: t('commons.table.status'), value: item.health },
                ],
            },
            {
                id: 'memory',
                title: t('aiTools.gpu.memory'),
                items: [
                    {
                        label: t('aiTools.gpu.ddrUsage'),
                        value: hasUsage(item.memoryUsed, item.memoryTotal)
                            ? formatMemory(item.memoryUsed || 'N/A', item.memoryTotal || 'N/A')
                            : '',
                    },
                    {
                        label: t('aiTools.gpu.hbmUsage'),
                        value: hasUsage(item.hbmUsed, item.hbmTotal)
                            ? formatMemory(item.hbmUsed || 'N/A', item.hbmTotal || 'N/A')
                            : '',
                    },
                    {
                        label: t('aiTools.gpu.hugepagesUsage'),
                        value: hasUsage(item.hugepagesUsed, item.hugepagesTotal)
                            ? formatUsage(item.hugepagesUsed, item.hugepagesTotal)
                            : '',
                    },
                    { label: t('aiTools.gpu.ddrBandwidth'), value: item.ddrBandwidth },
                    { label: t('aiTools.gpu.hbmBandwidth'), value: item.hbmBandwidth },
                ],
            },
            {
                id: 'engines',
                title: t('aiTools.gpu.cpuUtil'),
                items: [
                    { label: t('aiTools.gpu.aiCPUUtil'), value: item.aiCPUUtil },
                    { label: t('aiTools.gpu.ctrlCPUUtil'), value: item.ctrlCPUUtil },
                ],
            },
            {
                id: 'device',
                title: t('aiTools.gpu.deviceInfo'),
                items: [{ label: t('aiTools.gpu.busID'), value: item.busID }],
            },
        ]);
    }
    return visibleSections([
        {
            id: 'runtime',
            title: t('aiTools.gpu.runtimeInfo'),
            items: [
                { label: t('aiTools.gpu.powerUsage'), value: formatPower(item) },
                { label: t('aiTools.gpu.frequency'), value: item.frequency },
                { label: t('aiTools.gpu.memoryFrequency'), value: item.memoryFrequency },
                { label: t('aiTools.gpu.mediaFrequency'), value: item.mediaFrequency },
                { label: t('aiTools.gpu.fanSpeed') + ' (%)', value: item.fanSpeed },
                { label: t('aiTools.gpu.fanRPM'), value: item.fanRPM },
                { label: t('aiTools.gpu.hotspotTemperature'), value: item.hotspotTemperature },
                { label: t('aiTools.gpu.memoryTemperature'), value: item.memoryTemperature },
                {
                    label: t('aiTools.gpu.performanceState'),
                    value: item.performanceState,
                    help: t('aiTools.gpu.performanceStateHelper'),
                },
                { label: t('aiTools.gpu.clockEvents'), value: item.clockEvents?.join(', ') },
            ],
        },
        {
            id: 'memory',
            title: t('aiTools.gpu.memory'),
            items: [
                { label: t('aiTools.gpu.memoryActivity'), value: item.memoryActivity },
                { label: t('aiTools.gpu.freeMemory'), value: item.memoryFree },
                { label: t('aiTools.gpu.memoryReserved'), value: item.memoryReserved },
            ],
        },
        {
            id: 'engines',
            title: t('aiTools.gpu.engineUtil'),
            items: [
                { label: t('aiTools.gpu.encoderUtil'), value: item.encoderUtil },
                { label: t('aiTools.gpu.decoderUtil'), value: item.decoderUtil },
                { label: t('aiTools.gpu.jpegUtil'), value: item.jpegUtil },
                { label: t('aiTools.gpu.ofaUtil'), value: item.ofaUtil },
                { label: t('aiTools.gpu.mediaUtil'), value: item.mediaUtil },
            ],
        },
        {
            id: 'ecc',
            title: 'ECC',
            items: [
                { label: 'ECC', value: isAvailable(item.ecc) ? loadEcc(item.ecc) : '', help: t('aiTools.gpu.ecc') },
                {
                    label: t('aiTools.gpu.eccPending'),
                    value: isAvailable(item.eccPending) ? loadEcc(item.eccPending!) : '',
                },
                ...(item.eccErrors || []).flatMap((error) => [
                    { label: `${error.scope} · ${t('aiTools.gpu.eccCorrectable')}`, value: error.correctable },
                    { label: `${error.scope} · ${t('aiTools.gpu.eccUncorrectable')}`, value: error.uncorrectable },
                ]),
            ],
        },
        {
            id: 'device',
            title: t('aiTools.gpu.deviceInfo'),
            items: [
                { label: 'UUID', value: item.uuid, ellipsis: true },
                { label: t('aiTools.gpu.busID'), value: item.busID },
                { label: t('aiTools.gpu.driverVersion'), value: item.driverVersion },
                { label: t('aiTools.gpu.architecture'), value: item.architecture },
                { label: t('aiTools.gpu.powerDefaultLimit'), value: item.defaultPowerLimit },
                { label: t('aiTools.gpu.powerMaxLimit'), value: item.maxPowerLimit },
                {
                    label: t('aiTools.gpu.pcieGeneration'),
                    value: isAvailable(item.pcieGeneration)
                        ? `${item.pcieGeneration} / ${item.pcieMaxGeneration || 'N/A'}`
                        : '',
                },
                {
                    label: t('aiTools.gpu.pcieWidth'),
                    value: isAvailable(item.pcieWidth) ? `${item.pcieWidth} / ${item.pcieMaxWidth || 'N/A'}` : '',
                },
                {
                    label: t('aiTools.gpu.persistenceMode'),
                    value: isAvailable(item.persistenceMode) ? loadEcc(item.persistenceMode) : '',
                    help: t('aiTools.gpu.persistenceModeHelper'),
                },
                {
                    label: t('aiTools.gpu.displayActive'),
                    value: isAvailable(item.displayActive)
                        ? t(
                              'aiTools.gpu.' +
                                  (item.displayActive.toLowerCase() === 'disabled'
                                      ? 'displayActiveF'
                                      : 'displayActiveT'),
                          )
                        : '',
                },
                {
                    label: t('aiTools.gpu.computeMode'),
                    value: isAvailable(item.computeMode) ? loadComputeMode(item.computeMode) : '',
                    help: ['defaultHelper', 'exclusiveProcessHelper', 'exclusiveThreadHelper', 'prohibitedHelper']
                        .map((key) => t('aiTools.gpu.' + key))
                        .join('\n'),
                },
                {
                    label: 'MIG',
                    value: isAvailable(item.migMode) ? loadEcc(item.migMode) : '',
                    help: t('aiTools.gpu.migModeHelper'),
                },
            ],
        },
    ]);
};

const xpuDetailSections = (item: AI.XpuInfo['xpu'][number]): DetailSection[] => {
    const t = i18n.global.t;
    return visibleSections([
        {
            id: 'runtime',
            title: t('aiTools.gpu.runtimeInfo'),
            items: [
                { label: t('aiTools.gpu.powerUsage'), value: item.stats.power },
                { label: t('aiTools.gpu.frequency'), value: item.stats.frequency },
                { label: t('aiTools.gpu.mediaFrequency'), value: item.stats.mediaFrequency },
                { label: t('aiTools.gpu.memoryTemperature'), value: item.stats.memoryTemperature },
            ],
        },
        {
            id: 'memory',
            title: t('aiTools.gpu.memory'),
            items: [
                { label: t('aiTools.gpu.freeMemory'), value: item.basic.freeMemory },
                { label: t('aiTools.gpu.memoryUsage'), value: item.stats.memoryUtil },
                { label: t('aiTools.gpu.memoryBandwidth'), value: item.stats.memoryBandwidthUtil },
            ],
        },
        {
            id: 'engines',
            title: t('aiTools.gpu.engineUtil'),
            items: [
                { label: t('aiTools.gpu.computeUtil'), value: item.stats.computeUtil },
                { label: t('aiTools.gpu.mediaUtil'), value: item.stats.mediaUtil },
                { label: t('aiTools.gpu.copyUtil'), value: item.stats.copyUtil },
            ],
        },
        ...(item.tiles || []).map((tile) => ({
            id: `tile-${tile.tileID}`,
            title: `Tile ${tile.tileID}`,
            items: [
                { label: t('aiTools.gpu.gpuUtil'), value: tile.stats.gpuUtil },
                { label: t('aiTools.gpu.powerCurrent'), value: tile.stats.power },
                { label: t('aiTools.gpu.frequency'), value: tile.stats.frequency },
                { label: t('aiTools.gpu.mediaFrequency'), value: tile.stats.mediaFrequency },
                { label: t('aiTools.gpu.temperature'), value: tile.stats.temperature },
                { label: t('aiTools.gpu.memoryTemperature'), value: tile.stats.memoryTemperature },
                { label: t('aiTools.gpu.memoryUsed'), value: tile.stats.memoryUsed },
                { label: t('aiTools.gpu.memoryUsage'), value: tile.stats.memoryUtil },
                { label: t('aiTools.gpu.memoryBandwidth'), value: tile.stats.memoryBandwidthUtil },
                { label: t('aiTools.gpu.computeUtil'), value: tile.stats.computeUtil },
                { label: t('aiTools.gpu.mediaUtil'), value: tile.stats.mediaUtil },
                { label: t('aiTools.gpu.copyUtil'), value: tile.stats.copyUtil },
            ],
        })),
        {
            id: 'device',
            title: t('aiTools.gpu.deviceInfo'),
            items: [
                { label: 'UUID', value: item.basic.uuid, ellipsis: true },
                { label: t('aiTools.gpu.busID'), value: item.basic.pciBdfAddress },
                { label: t('aiTools.gpu.driverVersion'), value: item.basic.driverVersion },
            ],
        },
    ]);
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
    return Boolean(
        value &&
        value.trim() !== '' &&
        !['N/A', 'NA', 'NOT SUPPORTED', '[NOT SUPPORTED]', 'NAN', 'INF'].includes(value.trim().toUpperCase()),
    );
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

const deviceUtil = (item: AcceleratorDevice) => {
    return item.type === 'ascend' ? item.aiCore : item.gpuUtil;
};

const formatPower = (item: AcceleratorDevice) => {
    if (item.type === 'ascend') {
        return item.powerDraw;
    }
    return isAvailable(item.powerLimit) ? `${item.powerDraw || 'N/A'} / ${item.powerLimit}` : item.powerDraw;
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
    return val || 'N/A';
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
    flex-wrap: wrap;
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
    align-items: stretch;
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
.device-actions {
    display: flex;
    align-items: center;
    flex-shrink: 0;
    gap: 12px;
    margin-left: 12px;
    .el-button + .el-button {
        margin-left: 0;
    }
}
.process-count {
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
.device-secondary-metrics {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 16px 20px;
    margin-top: 16px;
    padding-top: 16px;
    border-top: 1px solid var(--el-border-color-lighter);
}
.metric-primary {
    min-height: 67px;
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
    .device-secondary-metrics {
        grid-template-columns: repeat(2, minmax(0, 1fr));
        .metric-item:last-child {
            grid-column: 1 / -1;
        }
    }
    .metric-temperature {
        grid-column: 1 / -1;
        padding-top: 12px;
        border-top: 1px solid var(--el-border-color-lighter);
        &::before {
            display: none;
        }
    }
}
</style>
