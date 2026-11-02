<template>
    <LayoutContent v-loading="loading" :title="$t('xpack.vm.overview')">
        <template #main>
            <div class="vm-overview">
                <section class="vm-hero">
                    <div class="vm-health">
                        <div class="vm-health__top">
                            <div>
                                <div class="vm-section-label">
                                    {{ $t('xpack.vm.overview') }}
                                </div>
                                <div class="vm-health__title">{{ $t('xpack.vm.title') }}</div>
                            </div>
                            <el-button type="primary" :disabled="vmUnavailable" @click="orphanCleanRef?.open()">
                                {{ $t('xpack.vm.orphanClean') }}
                            </el-button>
                        </div>
                        <div class="vm-health__meta">
                            <span>
                                {{ $t('xpack.vm.runtimeHealth') }}
                                <strong :class="{ danger: !environmentReady }">{{ environmentStatusText }}</strong>
                            </span>
                            <span>{{ $t('xpack.vm.vm') }} {{ vms.length }}</span>
                            <span>{{ $t('xpack.vm.storagePool') }} {{ storages.length }}</span>
                            <span>{{ $t('xpack.vm.riskReminder') }} {{ riskItems.length }}</span>
                        </div>
                    </div>

                    <div class="vm-stat-grid">
                        <div v-for="item in statusCards" :key="item.key" class="vm-stat-card">
                            <div class="vm-stat-card__label">{{ item.label }}</div>
                            <div class="vm-stat-card__value">{{ item.value }}</div>
                            <div class="vm-stat-card__extra">{{ item.extra }}</div>
                        </div>
                    </div>
                </section>

                <section class="vm-overview-grid">
                    <el-card shadow="never" class="vm-panel vm-resource-panel">
                        <template #header>
                            <div class="vm-panel__header">
                                <span>{{ $t('xpack.vm.resourceOverview') }}</span>
                            </div>
                        </template>
                        <div class="vm-resource-list">
                            <div v-for="item in resourceCards" :key="item.key" class="vm-resource-row">
                                <div class="vm-resource-row__head">
                                    <div>
                                        <span>{{ item.label }}</span>
                                        <strong>{{ item.value }}</strong>
                                    </div>
                                </div>
                                <el-progress :percentage="item.percent" :status="progressStatus(item.percent)" />
                                <div class="vm-resource-row__meta">{{ item.meta }}</div>
                            </div>
                        </div>
                    </el-card>

                    <el-card shadow="never" class="vm-panel">
                        <template #header>{{ $t('xpack.vm.riskReminder') }}</template>
                        <div v-if="riskItems.length" class="vm-risk-list">
                            <div v-for="item in riskItems" :key="item.key" class="vm-risk-item" :class="item.type">
                                <el-icon><WarningFilled /></el-icon>
                                <div>
                                    <strong>{{ item.title }}</strong>
                                    <span>{{ item.message }}</span>
                                </div>
                            </div>
                        </div>
                        <el-empty v-else :description="$t('xpack.vm.noRisk')" />
                    </el-card>
                </section>

                <section class="vm-storage-section">
                    <el-card shadow="never" class="vm-panel">
                        <template #header>
                            <div class="vm-panel__header">
                                <span>{{ $t('xpack.vm.storageOverview') }}</span>
                                <el-button link type="primary" @click="go('/vm/storage-pools')">
                                    {{ $t('xpack.vm.storagePool') }}
                                </el-button>
                            </div>
                        </template>
                        <div v-if="storages.length" class="vm-storage-grid">
                            <div v-for="item in storages" :key="item.id" class="vm-storage-card">
                                <div class="vm-storage-card__head">
                                    <div>
                                        <strong>{{ item.name }}</strong>
                                        <span>{{ item.path || '-' }}</span>
                                    </div>
                                    <Status :status="resourceStatus(item)" />
                                </div>
                                <el-progress
                                    :percentage="normalizePercent(storageUsedPercent(item))"
                                    :status="progressStatus(storageUsedPercent(item))"
                                />
                                <div class="vm-storage-card__meta">
                                    <span>{{ $t('xpack.vm.available') }} {{ computeSize(item.available || 0) }}</span>
                                    <span>{{ $t('xpack.vm.vmCount') }} {{ item.vmCount || 0 }}</span>
                                </div>
                            </div>
                        </div>
                        <el-empty v-else />
                    </el-card>
                </section>
            </div>
        </template>
    </LayoutContent>
    <OrphanClean ref="orphanCleanRef" />
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRouter } from 'vue-router';
import { WarningFilled } from '@element-plus/icons-vue';
import OrphanClean from './orphan-clean.vue';
import LayoutContent from '@/components/layout-content/index.vue';
import i18n from '@/lang';
import { loadCurrentInfo } from '@/api/modules/dashboard';
import type { Dashboard } from '@/api/interface/dashboard';
import { GlobalStore } from '@/store';
import { computeSize } from '@/utils/size';
import { VM } from '@/api/interface/vm';
import { checkVMEnvironment, listVMStorages, searchVMs } from '@/api/modules/vm';
import { isVMMemoryAllocated } from '@/utils/vm';

const BYTE_PER_MIB = 1024 * 1024;

const props = defineProps<{
    vmUnavailable: boolean;
    vmEnvironmentLoaded: boolean;
}>();
const router = useRouter();
const globalStore = GlobalStore();
const loading = ref(false);
const orphanCleanRef = ref<InstanceType<typeof OrphanClean>>();
const vms = ref<VM.Item[]>([]);
const storages = ref<VM.StorageItem[]>([]);
const currentInfo = ref<Dashboard.CurrentInfo>();
const defaultEnvironment = (): VM.Environment => ({
    arch: '',
    archSupported: true,
    ready: false,
    helperExist: false,
    helperMessage: '',
    kvmExist: false,
    kvmMessage: '',
    libvirtExist: false,
    libvirtActive: false,
    libvirtMessage: '',
    storageExist: false,
    storageMessage: '',
});
const environment = ref<VM.Environment>(defaultEnvironment());
const vmUnavailable = computed(() => props.vmUnavailable);

const environmentReady = computed(() => environment.value.ready);
const environmentStatusText = computed(() => {
    if (!props.vmEnvironmentLoaded) {
        return i18n.global.t('commons.status.waiting');
    }
    return environmentReady.value ? i18n.global.t('commons.status.normal') : i18n.global.t('xpack.vm.runtimeAbnormal');
});
const runningCount = computed(() => vms.value.filter((item) => ['running', 'blocked'].includes(item.status)).length);
const stoppedCount = computed(() => vms.value.filter((item) => item.status === 'shut off').length);
const creatingCount = computed(() => vms.value.filter((item) => item.status?.toLowerCase() === 'creating').length);
const pausedCount = computed(() => vms.value.filter((item) => item.status === 'paused').length);
const abnormalCount = computed(
    () =>
        vms.value.filter(
            (item) =>
                !['running', 'blocked', 'shut off', 'paused', 'creating', 'shutdown', 'pmsuspended'].includes(
                    item.status,
                ),
        ).length,
);

const statusCards = computed(() => [
    {
        key: 'running',
        label: i18n.global.t('commons.status.running'),
        value: runningCount.value,
        extra: `${percent(runningCount.value, vms.value.length)}%`,
    },
    {
        key: 'stopped',
        label: i18n.global.t('commons.status.stopped'),
        value: stoppedCount.value,
        extra: `${percent(stoppedCount.value, vms.value.length)}%`,
    },
    {
        key: 'creating',
        label: i18n.global.t('commons.status.creating'),
        value: creatingCount.value,
        extra: i18n.global.t('xpack.vm.inProgress'),
    },
    {
        key: 'abnormal',
        label: i18n.global.t('xpack.vm.abnormal'),
        value: abnormalCount.value,
        extra: abnormalCount.value ? i18n.global.t('xpack.vm.needAttention') : i18n.global.t('xpack.vm.normal'),
    },
]);

const cpuAllocated = computed(() => vms.value.reduce((total, item) => total + (Number(item.cpu) || 0), 0));
const cpuTotal = computed(() => currentInfo.value?.cpuTotal || 0);
const memoryAllocated = computed(() =>
    vms.value.reduce((total, item) => total + (Number(item.memory) || 0) * BYTE_PER_MIB, 0),
);
const memoryTotal = computed(() => currentInfo.value?.memoryTotal || 0);
const runningMemoryAllocated = computed(() =>
    vms.value
        .filter((item) => isVMMemoryAllocated(item.status))
        .reduce((total, item) => total + (Number(item.memory) || 0) * BYTE_PER_MIB, 0),
);
const normalizePath = (path?: string) => {
    const value = (path || '').trim();
    if (!value) {
        return '';
    }
    return value.replace(/\/+/g, '/').replace(/\/$/, '') || '/';
};
const isPathInDir = (path: string, dir: string) => dir === '/' || path === dir || path.startsWith(`${dir}/`);
const storageRowsForSummary = computed(() => {
    const diskData = currentInfo.value?.diskData || [];
    const rowsByPath = new Map<string, VM.StorageItem>();
    for (const storage of storages.value) {
        const storagePath = normalizePath(storage.path);
        if (!storagePath) {
            continue;
        }
        const disk = diskData
            .filter((item) => {
                const diskPath = normalizePath(item.path);
                return diskPath && isPathInDir(storagePath, diskPath);
            })
            .sort((left, right) => normalizePath(right.path).length - normalizePath(left.path).length)[0];
        const key = normalizePath(disk?.path || storage.path);
        const existing = rowsByPath.get(key);
        if (!existing || (storage.capacity || 0) > (existing.capacity || 0)) {
            rowsByPath.set(key, storage);
        }
    }
    return Array.from(rowsByPath.values());
});
const storageTotal = computed(() => {
    if (storageRowsForSummary.value.length) {
        return storageRowsForSummary.value.reduce((total, item) => total + (item.capacity || 0), 0);
    }
    return storages.value.reduce((total, item) => total + (item.capacity || 0), 0);
});
const storageAvailable = computed(() => {
    if (storageRowsForSummary.value.length) {
        return storageRowsForSummary.value.reduce((total, item) => total + (item.available || 0), 0);
    }
    return storages.value.reduce((total, item) => total + (item.available || 0), 0);
});
const storageUsed = computed(() => Math.max(storageTotal.value - storageAvailable.value, 0));

const normalizePercent = (value: number) => Number(Math.min(Math.max(value || 0, 0), 100).toFixed(2));
const percent = (used: number, total: number) => {
    if (!total) {
        return 0;
    }
    return normalizePercent((used / total) * 100);
};

const cpuPercent = computed(() => percent(cpuAllocated.value, cpuTotal.value));
const memoryPercent = computed(() => percent(memoryAllocated.value, memoryTotal.value));
const storagePercent = computed(() => percent(storageUsed.value, storageTotal.value));

const resourceCards = computed(() => [
    {
        key: 'cpu',
        label: i18n.global.t('xpack.vm.cpuAllocated'),
        value: `${cpuAllocated.value} / ${cpuTotal.value || '-'} ${i18n.global.t('commons.units.core')}`,
        percent: cpuPercent.value,
        meta: i18n.global.t('xpack.vm.totalAllocated'),
    },
    {
        key: 'memory',
        label: i18n.global.t('xpack.vm.memoryAllocated'),
        value: `${computeSize(memoryAllocated.value)} / ${computeSize(memoryTotal.value)}`,
        percent: memoryPercent.value,
        meta: i18n.global.t('xpack.vm.runningAllocated', [computeSize(runningMemoryAllocated.value)]),
    },
    {
        key: 'storage',
        label: i18n.global.t('xpack.vm.storageUsed'),
        value: `${computeSize(storageUsed.value)} / ${computeSize(storageTotal.value)}`,
        percent: storagePercent.value,
        meta: i18n.global.t('xpack.vm.storageAvailable', [computeSize(storageAvailable.value)]),
    },
]);

const progressStatus = (value: number) => {
    if (value >= 90) {
        return 'exception';
    }
    if (value >= 75) {
        return 'warning';
    }
    return undefined;
};

const storageUsedPercent = (item: VM.StorageItem) => {
    if (item.usedPercent !== undefined) {
        return Number(item.usedPercent) || 0;
    }
    if (!item.capacity) {
        return 0;
    }
    return ((item.capacity - (item.available || 0)) / item.capacity) * 100;
};

const resourceStatus = (item: VM.StorageItem) => {
    if (!resourceExists(item.status)) {
        return item.status || 'deleted';
    }
    if (item.status?.toLowerCase() === 'running') {
        return 'enable';
    }
    if (item.status?.toLowerCase() === 'inactive') {
        return 'stopped';
    }
    return item.status || 'unknown';
};

const resourceExists = (status: string) => status?.toLowerCase() !== 'deleted';

const riskItems = computed(() => {
    const risks: Array<{
        key: string;
        title: string;
        message: string;
        type: 'warning' | 'danger';
    }> = [];
    if (!environmentReady.value) {
        risks.push({
            key: 'environment',
            title: i18n.global.t('xpack.vm.runtimeAbnormal'),
            message: i18n.global.t('xpack.vm.environmentRisk'),
            type: 'danger',
        });
    }
    storages.value
        .filter((item) => storageUsedPercent(item) >= 80)
        .forEach((item) => {
            risks.push({
                key: `storage-${item.id}`,
                title: i18n.global.t('xpack.vm.storagePool'),
                message: i18n.global.t('xpack.vm.storageRisk', [item.name, storageUsedPercent(item).toFixed(2)]),
                type: storageUsedPercent(item) >= 90 ? 'danger' : 'warning',
            });
        });
    if (abnormalCount.value > 0) {
        risks.push({
            key: 'abnormal-vm',
            title: i18n.global.t('xpack.vm.abnormal'),
            message: i18n.global.t('xpack.vm.abnormalRisk', [abnormalCount.value]),
            type: 'danger',
        });
    }
    if (pausedCount.value > 0) {
        risks.push({
            key: 'paused-vm',
            title: i18n.global.t('xpack.vm.suspend'),
            message: i18n.global.t('xpack.vm.pausedRisk', [pausedCount.value]),
            type: 'warning',
        });
    }
    if (memoryPercent.value >= 85) {
        risks.push({
            key: 'memory',
            title: i18n.global.t('xpack.vm.memory'),
            message: i18n.global.t('xpack.vm.memoryRisk', [memoryPercent.value]),
            type: memoryPercent.value >= 95 ? 'danger' : 'warning',
        });
    }
    return risks;
});

const loadOverview = async () => {
    if (!props.vmEnvironmentLoaded || vmUnavailable.value) {
        vms.value = [];
        storages.value = [];
        currentInfo.value = undefined;
        environment.value = defaultEnvironment();
        return;
    }
    loading.value = true;
    try {
        const [environmentRes, vmRes, storageRes, currentRes] = await Promise.all([
            checkVMEnvironment(),
            searchVMs({ page: 1, pageSize: 10000, info: '' }),
            listVMStorages(),
            loadCurrentInfo('all', 'all', globalStore.currentNode || 'local'),
        ]);
        environment.value = environmentRes.data || environment.value;
        vms.value = vmRes.data.items || [];
        storages.value = storageRes.data || [];
        currentInfo.value = currentRes.data;
    } finally {
        loading.value = false;
    }
};

const go = (path: string) => {
    router.push(path);
};

watch(() => [props.vmEnvironmentLoaded, props.vmUnavailable], loadOverview, {
    immediate: true,
});
</script>

<style scoped lang="scss">
.vm-overview {
    display: flex;
    flex-direction: column;
    gap: 16px;
}

.vm-hero {
    display: grid;
    grid-template-columns: minmax(320px, 0.9fr) minmax(0, 1.1fr);
    gap: 16px;
}

.vm-health,
.vm-stat-card,
.vm-panel {
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 6px;
    background: var(--el-bg-color);
}

.vm-health {
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    gap: 18px;
    padding: 20px;
}

.vm-health__top,
.vm-panel__header,
.vm-resource-row__head,
.vm-storage-card__head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
}

.vm-section-label,
.vm-stat-card__label,
.vm-resource-row__meta,
.vm-storage-card__meta,
.vm-health__meta,
.vm-storage-card__head span,
.vm-risk-item span {
    color: var(--el-text-color-secondary);
    font-size: 12px;
}

.vm-health__title {
    margin-top: 6px;
    font-size: 24px;
    font-weight: 600;
    color: var(--el-text-color-primary);
}

.vm-health__meta {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;

    span {
        display: inline-flex;
        align-items: center;
        gap: 6px;
        min-height: 26px;
        padding: 0 10px;
        border: 1px solid var(--el-border-color-lighter);
        border-radius: 6px;
        background: var(--el-fill-color-extra-light);
    }

    strong {
        color: var(--el-text-color-primary);
        font-weight: 500;
    }

    strong.danger {
        color: var(--el-color-danger);
    }
}

.vm-stat-grid {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 12px;
}

.vm-stat-card {
    min-width: 0;
    padding: 16px;
    border-top: 3px solid var(--el-border-color);
    border-top-color: var(--el-color-primary);
}

.vm-stat-card__value {
    margin-top: 10px;
    font-size: 30px;
    font-weight: 600;
    color: var(--el-text-color-primary);
}

.vm-stat-card__extra {
    margin-top: 6px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
}

.vm-overview-grid {
    display: grid;
    grid-template-columns: minmax(0, 1.35fr) minmax(320px, 0.65fr);
    gap: 16px;
}

.vm-overview-grid--bottom {
    grid-template-columns: minmax(0, 1.35fr) minmax(320px, 0.65fr);
}

.vm-panel {
    min-width: 0;

    :deep(.el-card__header) {
        padding: 14px 16px;
        font-weight: 600;
    }

    :deep(.el-card__body) {
        padding: 16px;
    }
}

.vm-resource-list,
.vm-risk-list {
    display: flex;
    flex-direction: column;
    gap: 14px;
}

.vm-resource-row,
.vm-risk-item {
    min-width: 0;
    padding: 12px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 6px;
    background: var(--el-fill-color-extra-light);
}

.vm-resource-row__head {
    align-items: center;
    margin-bottom: 8px;

    div {
        min-width: 0;
    }

    span,
    strong {
        display: block;
    }

    strong {
        margin-top: 4px;
        color: var(--el-text-color-primary);
        font-weight: 600;
    }
}

.vm-resource-row__meta {
    margin-top: 8px;
}

.vm-storage-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px;
}

.vm-storage-card {
    min-width: 0;
    padding: 12px;
    border: 1px solid var(--el-border-color-lighter);
    border-radius: 6px;
}

.vm-storage-card__head {
    margin-bottom: 10px;

    div {
        min-width: 0;
    }

    strong,
    span {
        display: block;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
    }
}

.vm-storage-card__meta {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    margin-top: 8px;
}

.vm-risk-item {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    background: var(--el-color-warning-light-9);
    border-color: var(--el-color-warning-light-5);

    .el-icon {
        margin-top: 2px;
        color: var(--el-color-warning);
    }

    strong,
    span {
        display: block;
    }

    strong {
        margin-bottom: 4px;
        color: var(--el-text-color-primary);
    }
}

.vm-risk-item.danger {
    background: var(--el-color-danger-light-9);
    border-color: var(--el-color-danger-light-5);

    .el-icon {
        color: var(--el-color-danger);
    }
}

@media (max-width: 1200px) {
    .vm-hero,
    .vm-overview-grid,
    .vm-overview-grid--bottom {
        grid-template-columns: 1fr;
    }

    .vm-stat-grid,
    .vm-storage-grid {
        grid-template-columns: repeat(2, minmax(0, 1fr));
    }
}

@media (max-width: 640px) {
    .vm-stat-grid,
    .vm-storage-grid {
        grid-template-columns: 1fr;
    }
}
</style>
