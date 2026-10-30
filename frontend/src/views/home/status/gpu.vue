<template>
    <div class="gpu-status">
        <div class="gpu-summary">
            <div class="gpu-summary-header">
                <div class="device-counts">
                    <strong v-for="group in deviceGroups" :key="group.type">
                        {{ group.type }} × {{ group.count }}
                    </strong>
                </div>
                <div class="summary-actions">
                    <el-button v-if="devices.length > 4" link type="primary" size="small" @click="showAll = !showAll">
                        {{ $t(showAll ? 'tabs.hide' : 'tabs.more') }}
                        <el-icon>
                            <ArrowUp v-if="showAll" />
                            <ArrowDown v-else />
                        </el-icon>
                    </el-button>
                    <el-button link type="primary" size="small" @click="$emit('open')">
                        {{ $t('aiTools.gpu.viewHistory') }}
                        <el-icon><ArrowRight /></el-icon>
                    </el-button>
                </div>
            </div>
            <div v-if="dedicatedMemoryDevices.length" class="summary-memory">
                <span class="metric-label">{{ $t('aiTools.gpu.memoryTotal') }}</span>
                <strong>{{ formatMemory(totalMemory) }}</strong>
                <span class="metric-label">{{ $t('home.used') }} {{ formatMemory(usedMemory) }}</span>
            </div>
            <div v-if="sharedMemoryDevice" class="summary-memory" :title="$t('aiTools.gpu.systemMemoryHelper')">
                <span class="metric-label">{{ $t('aiTools.gpu.systemMemory') }}</span>
                <strong>{{ formatMemory(sharedMemoryDevice.memoryTotal) }}</strong>
                <span class="metric-label">
                    {{ $t('home.used') }} {{ formatMemory(sharedMemoryDevice.memoryUsed) }}
                </span>
            </div>
        </div>

        <div class="gpu-grid">
            <article v-for="device in visibleDevices" :key="device.key" class="gpu-card">
                <div class="device-header">
                    <div class="device-identity">
                        <span class="device-id">{{ device.id }}</span>
                        <h3 :title="device.name">{{ device.name }}</h3>
                    </div>
                    <el-popover :width="360" trigger="click" popper-class="dashboard-gpu-popover">
                        <el-descriptions :title="device.name" :column="1" size="small">
                            <el-descriptions-item
                                v-for="detail in device.details.filter((item) => isAvailable(item[1]))"
                                :key="detail[0]"
                                :label="detail[0]"
                            >
                                {{ detail[1] }}
                            </el-descriptions-item>
                        </el-descriptions>
                        <template #reference>
                            <el-button link type="primary" size="small">{{ $t('commons.button.view') }}</el-button>
                        </template>
                    </el-popover>
                </div>
                <div class="device-metrics">
                    <div class="device-metric metric-utilization">
                        <div class="metric-heading">
                            <span class="metric-label">
                                {{ device.type === 'NPU' ? $t('aiTools.gpu.aiCore') : $t('aiTools.gpu.gpuUtil') }}
                            </span>
                            <strong>{{ device.utilization === null ? '—' : `${device.utilization}%` }}</strong>
                        </div>
                        <el-progress :percentage="device.utilization ?? 0" :show-text="false" :stroke-width="5" />
                    </div>
                    <div class="device-metric">
                        <div class="metric-heading">
                            <span
                                class="metric-label"
                                :title="device.sharedMemory ? $t('aiTools.gpu.systemMemoryHelper') : undefined"
                            >
                                {{ $t(device.sharedMemory ? 'aiTools.gpu.systemMemory' : 'aiTools.gpu.memoryUsed') }}
                            </span>
                            <strong>
                                {{ formatMemory(device.memoryUsed) }}
                                <span class="metric-total">/ {{ formatMemory(device.memoryTotal) }}</span>
                            </strong>
                        </div>
                        <el-progress
                            :percentage="device.memoryPercent ?? 0"
                            :show-text="false"
                            :stroke-width="5"
                            color="var(--el-color-success)"
                        />
                    </div>
                </div>
                <div class="device-footer">
                    <div>
                        <span class="metric-label">{{ $t('aiTools.gpu.temperature') }}</span>
                        <strong>{{ device.temperature }}</strong>
                    </div>
                    <div>
                        <span class="metric-label">{{ $t('aiTools.gpu.powerUsage') }}</span>
                        <strong>{{ device.power }}</strong>
                    </div>
                </div>
            </article>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { ArrowDown, ArrowRight, ArrowUp } from '@element-plus/icons-vue';
import { Dashboard } from '@/api/interface/dashboard';
import { computeSizeForDocker } from '@/utils/size';
import i18n from '@/lang';

const props = defineProps<{ current: Dashboard.CurrentInfo }>();
defineEmits<{ open: [] }>();
const showAll = ref(false);
const t = i18n.global.t;

const isAvailable = (value?: string) =>
    Boolean(value?.trim() && !/^(N\/A|NA|NOT SUPPORTED|\[NOT SUPPORTED\])$/i.test(value.trim()));

const memoryBytes = (value?: string): number | null => {
    const match = value?.trim().match(/^(\d+(?:\.\d+)?)\s*(B|bytes?|[KMGT]i?B)?$/i);
    if (!match) return null;
    const units: Record<string, number> = {
        '': 1024 ** 2,
        b: 1,
        byte: 1,
        bytes: 1,
        kb: 1000,
        kib: 1024,
        mb: 1000 ** 2,
        mib: 1024 ** 2,
        gb: 1000 ** 3,
        gib: 1024 ** 3,
        tb: 1000 ** 4,
        tib: 1024 ** 4,
    };
    const bytes = Number(match[1]) * units[(match[2] || '').toLowerCase()];
    return Number.isFinite(bytes) ? bytes : null;
};

const formatMemory = (value: number | null) => (value === null ? '—' : computeSizeForDocker(value));

const devices = computed(() =>
    [
        ...(props.current.gpuData || []).map((item) => ({
            key: `gpu-${item.type}-${item.busID || item.index}`,
            type: 'GPU',
            sharedMemory: !!item.sharedMemory,
            card: `${item.type}-${item.busID || item.index}`,
            id: `GPU ${item.index}`,
            name: item.productName || 'GPU',
            util: item.gpuUtil,
            used: item.memUsed || item.memoryUsage?.split('/')[0],
            total: item.memTotal || item.memoryUsage?.split('/')[1],
            temperature: item.temperature,
            power: item.powerUsage || item.powerDraw,
            details: [
                [t('aiTools.gpu.busID'), item.busID],
                [t('aiTools.gpu.fanSpeed'), item.fanSpeed],
                [t('aiTools.gpu.performanceState'), item.performanceState],
                [t('aiTools.gpu.powerUsage'), item.powerUsage || item.powerDraw],
            ],
        })),
        ...(props.current.npuData || []).map((item) => ({
            key: `npu-${item.npuIndex}-${item.chipIndex}`,
            type: 'NPU',
            sharedMemory: false,
            card: item.npuIndex,
            id: `NPU ${item.npuIndex} · Chip ${item.chipIndex}`,
            name: item.productName || 'NPU',
            util: item.aiCore,
            used: item.memUsed,
            total: item.memTotal,
            temperature: item.temperature,
            power: item.powerDraw,
            details: [
                [t('aiTools.gpu.busID'), item.busID],
                [t('commons.table.status'), item.health],
                [t('aiTools.gpu.ddrUsage'), `${item.memoryUsed || '—'} / ${item.memoryTotal || '—'}`],
                [t('aiTools.gpu.hbmUsage'), `${item.hbmUsed || '—'} / ${item.hbmTotal || '—'}`],
                [t('aiTools.gpu.hugepagesUsage'), `${item.hugepagesUsed || '—'} / ${item.hugepagesTotal || '—'}`],
            ],
        })),
        ...(props.current.xpuData || []).map((item) => ({
            key: `xpu-${item.deviceID}`,
            type: 'XPU',
            sharedMemory: false,
            card: item.deviceID,
            id: `XPU ${item.deviceID}`,
            name: item.deviceName || 'XPU',
            util: item.gpuUtil,
            used: item.memoryUsed,
            total: item.memory,
            temperature: item.temperature,
            power: item.power,
            details: [[t('aiTools.gpu.busID'), item.pciBdfAddress]],
        })),
    ].map((item) => {
        const memoryUsed = memoryBytes(item.used);
        const memoryTotal = memoryBytes(item.total);
        const utilization = item.util?.trim().match(/^(\d+(?:\.\d+)?)\s*%?$/);
        return {
            ...item,
            memoryUsed,
            memoryTotal,
            utilization: utilization ? Math.min(100, Number(utilization[1])) : null,
            memoryPercent:
                memoryUsed !== null && memoryTotal !== null && memoryTotal > 0
                    ? Math.min(100, Number(((memoryUsed / memoryTotal) * 100).toFixed(1)))
                    : null,
            temperature: isAvailable(item.temperature) ? item.temperature.replace(/\s*°?C\b/, ' °C') : '—',
            power: isAvailable(item.power) ? item.power : '—',
        };
    }),
);

const visibleDevices = computed(() => (showAll.value ? devices.value : devices.value.slice(0, 4)));
const deviceGroups = computed(() =>
    ['GPU', 'NPU', 'XPU']
        .map((type) => ({
            type,
            count: new Set(devices.value.filter((item) => item.type === type).map((item) => item.card)).size,
        }))
        .filter((item) => item.count > 0),
);
const sharedMemoryDevice = computed(() => devices.value.find((item) => item.sharedMemory));
const dedicatedMemoryDevices = computed(() => devices.value.filter((item) => !item.sharedMemory));
const totalMemory = computed(() =>
    dedicatedMemoryDevices.value.some((item) => item.memoryTotal === null)
        ? null
        : dedicatedMemoryDevices.value.reduce((sum, item) => sum + item.memoryTotal!, 0),
);
const usedMemory = computed(() =>
    dedicatedMemoryDevices.value.some((item) => item.memoryUsed === null)
        ? null
        : dedicatedMemoryDevices.value.reduce((sum, item) => sum + item.memoryUsed!, 0),
);
</script>

<style scoped lang="scss">
.gpu-status {
    container-type: inline-size;
    grid-column: 1 / -1;
    min-width: 0;
    color: var(--el-text-color-primary);
}
.gpu-summary {
    display: flex;
    flex-wrap: wrap;
    gap: 4px 24px;
    padding: 8px 16px;
    margin-bottom: 8px;
    border-radius: 8px;
    background: var(--el-fill-color-light);
}
.gpu-summary-header {
    display: flex;
    flex: 0 0 100%;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
}
.metric-label {
    color: var(--el-text-color-secondary);
    font-size: 12px;
}
.summary-actions {
    display: flex;
    align-items: center;
    flex-shrink: 0;
}
.device-counts {
    display: flex;
    flex-wrap: wrap;
    gap: 4px 16px;
    font-size: 14px;
}
.summary-memory {
    display: flex;
    align-items: baseline;
    flex-wrap: wrap;
    gap: 4px 8px;
    line-height: 20px;
}
.summary-memory > strong {
    font-size: 13px;
    font-weight: 600;
}
.gpu-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 10px;
}
.gpu-card {
    padding: 10px 16px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 8px;
    min-width: 0;
    background: var(--el-bg-color);
}
.device-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 10px;
}
.device-identity {
    flex: 1;
    min-width: 0;
}
.device-id {
    display: block;
    margin-bottom: 2px;
    font-size: 12px;
    color: var(--el-text-color-secondary);
}
.device-header h3 {
    flex: 1;
    min-width: 0;
    margin: 0;
    font-size: 14px;
    line-height: 20px;
    font-weight: 600;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
.device-header .el-button {
    flex-shrink: 0;
}
.device-metrics {
    display: grid;
    grid-template-columns: minmax(0, 0.8fr) minmax(0, 1.2fr);
    gap: 20px;
}
.device-metric {
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    min-width: 0;
}
.metric-heading {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin-bottom: 6px;
}
.metric-heading strong {
    font-size: 15px;
    line-height: 24px;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
}
.metric-utilization strong {
    font-size: 22px;
}
.metric-total {
    font-size: 12px;
    font-weight: 400;
    color: var(--el-text-color-secondary);
}
.device-footer {
    display: flex;
    flex-wrap: wrap;
    gap: 4px 16px;
    justify-content: space-between;
    margin-top: 8px;
    padding-top: 8px;
    border-top: 1px solid var(--el-border-color-extra-light);
}
.device-footer > div {
    display: flex;
    align-items: baseline;
    flex-wrap: wrap;
    gap: 6px;
}
.device-footer strong {
    font-size: 12px;
    font-weight: 500;
}
@container (max-width: 560px) {
    .gpu-grid {
        grid-template-columns: minmax(0, 1fr);
    }
}
</style>
<style lang="scss">
.dashboard-gpu-popover.el-popover {
    max-width: calc(100vw - 24px);
    max-height: calc(100dvh - 24px);
    overflow: auto;
    box-sizing: border-box;
}
</style>
