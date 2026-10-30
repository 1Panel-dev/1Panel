<template>
    <div class="gpu-status">
        <div class="gpu-summary">
            <div class="gpu-summary-header">
                <span class="metric-label">{{ $t('aiTools.gpu.deviceInfo') }}</span>
                <el-button link type="primary" size="small" @click="$emit('open')">
                    {{ $t('aiTools.gpu.gpu') }}
                    <el-icon><ArrowRight /></el-icon>
                </el-button>
            </div>
            <div class="summary-devices">
                <div class="device-counts">
                    <strong v-for="group in deviceGroups" :key="group.type">
                        {{ group.type }} × {{ group.count }}
                    </strong>
                </div>
                <div class="model-list">
                    <span v-for="model in modelGroups" :key="model.name">{{ model.name }} × {{ model.count }}</span>
                </div>
            </div>
            <div class="summary-memory">
                <span class="metric-label">{{ $t('aiTools.gpu.memoryTotal') }}</span>
                <strong>{{ formatMemory(totalMemory) }}</strong>
                <span class="metric-label">{{ $t('home.used') }} {{ formatMemory(usedMemory) }}</span>
            </div>
        </div>

        <div class="gpu-grid">
            <article v-for="device in visibleDevices" :key="device.key" class="gpu-card">
                <div class="device-header">
                    <span class="device-id">{{ device.id }}</span>
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
                <div class="device-identity">
                    <h3>{{ device.name }}</h3>
                    <span class="memory-capacity">{{ formatMemory(device.memoryTotal) }}</span>
                </div>
                <div class="device-metric">
                    <div class="metric-heading">
                        <span class="metric-label">
                            {{ device.type === 'NPU' ? $t('aiTools.gpu.aiCore') : $t('aiTools.gpu.gpuUtil') }}
                        </span>
                        <strong>{{ device.utilization === null ? '—' : `${device.utilization}%` }}</strong>
                    </div>
                    <el-progress :percentage="device.utilization ?? 0" :show-text="false" :stroke-width="6" />
                </div>
                <div class="device-metric">
                    <div class="metric-heading">
                        <span class="metric-label">{{ $t('aiTools.gpu.memoryUsed') }}</span>
                        <strong>
                            {{ formatMemory(device.memoryUsed) }}
                            <span class="metric-total">/ {{ formatMemory(device.memoryTotal) }}</span>
                        </strong>
                    </div>
                    <el-progress
                        :percentage="device.memoryPercent ?? 0"
                        :show-text="false"
                        :stroke-width="6"
                        color="var(--el-color-success)"
                    />
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
        <div v-if="devices.length > 4" class="gpu-expand">
            <el-button link type="primary" @click="showAll = !showAll">
                {{ $t(showAll ? 'tabs.hide' : 'tabs.more') }}
                <el-icon>
                    <ArrowUp v-if="showAll" />
                    <ArrowDown v-else />
                </el-icon>
            </el-button>
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
const modelGroups = computed(() => {
    const models = new Map<string, Set<string>>();
    for (const device of devices.value) {
        if (!models.has(device.name)) models.set(device.name, new Set());
        models.get(device.name)!.add(`${device.type}-${device.card}`);
    }
    return Array.from(models, ([name, cards]) => ({ name, count: cards.size }));
});
const totalMemory = computed(() =>
    devices.value.some((item) => item.memoryTotal === null)
        ? null
        : devices.value.reduce((sum, item) => sum + item.memoryTotal!, 0),
);
const usedMemory = computed(() =>
    devices.value.some((item) => item.memoryUsed === null)
        ? null
        : devices.value.reduce((sum, item) => sum + item.memoryUsed!, 0),
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
    justify-content: space-between;
    gap: 12px 24px;
    padding: 18px 20px;
    margin-bottom: 16px;
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
.summary-devices {
    flex: 1 1 200px;
    min-width: 0;
}
.device-counts {
    display: flex;
    flex-wrap: wrap;
    gap: 8px 20px;
    margin: 8px 0;
    font-size: 22px;
}
.model-list {
    display: flex;
    flex-wrap: wrap;
    gap: 4px 16px;
    font-size: 12px;
    color: var(--el-text-color-regular);
    overflow-wrap: anywhere;
}
.summary-memory {
    display: flex;
    flex-direction: column;
    gap: 8px;
    flex-shrink: 0;
    padding-left: 24px;
    border-left: 1px solid var(--el-border-color-lighter);
}
.summary-memory > strong {
    font-size: 22px;
    font-weight: 600;
}
.gpu-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 14px;
}
.gpu-card {
    padding: 16px;
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
}
.device-id {
    font-size: 12px;
    font-weight: 500;
    color: var(--el-text-color-secondary);
}
.device-identity {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 6px 10px;
    margin: 10px 0 20px;
}
.device-identity h3 {
    flex: 1 1 150px;
    margin: 0;
    font-size: 15px;
    line-height: 1.5;
    font-weight: 600;
    overflow-wrap: anywhere;
}
.memory-capacity {
    padding: 3px 7px;
    border-radius: 4px;
    font-size: 12px;
    color: var(--el-color-primary);
    background: var(--el-color-primary-light-9);
    white-space: nowrap;
}
.device-metric + .device-metric {
    margin-top: 16px;
}
.metric-heading {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    flex-wrap: wrap;
    gap: 4px 8px;
    margin-bottom: 8px;
}
.metric-heading strong {
    font-size: 13px;
    font-weight: 600;
    font-variant-numeric: tabular-nums;
}
.metric-total {
    font-weight: 400;
    color: var(--el-text-color-secondary);
}
.device-footer {
    display: flex;
    flex-wrap: wrap;
    gap: 10px 20px;
    justify-content: space-between;
    margin-top: 18px;
    padding-top: 12px;
    border-top: 1px solid var(--el-border-color-extra-light);
}
.device-footer > div {
    display: flex;
    align-items: baseline;
    flex-wrap: wrap;
    gap: 8px;
}
.device-footer strong {
    font-size: 12px;
    font-weight: 500;
}
.gpu-expand {
    margin-top: 16px;
    text-align: center;
}
@container (max-width: 560px) {
    .gpu-grid {
        grid-template-columns: minmax(0, 1fr);
    }
    .gpu-summary {
        padding: 16px;
        flex-wrap: wrap;
        gap: 16px;
    }
    .summary-memory {
        padding-left: 0;
        border-left: 0;
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
