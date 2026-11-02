<template>
    <DrawerPro v-model="visible" :header="$t('xpack.vm.monitor')" :resource="title" size="large" @close="close">
        <template #content>
            <div v-loading="loading" class="vm-monitor-drawer">
                <div class="vm-monitor-summary">
                    <div>
                        <span>{{ $t('xpack.vm.cpuUsage') }}</span>
                        <strong>{{ latestPoint ? `${latestPoint.cpu}%` : '-' }}</strong>
                    </div>
                    <div>
                        <span>{{ $t('xpack.vm.memoryUsage') }}</span>
                        <strong>{{ latestPoint ? `${latestPoint.memory}%` : '-' }}</strong>
                    </div>
                    <div>
                        <span>{{ $t('xpack.vm.diskReadWrite') }}</span>
                        <strong>
                            {{
                                latestPoint
                                    ? `${formatRate(latestPoint.diskRead)} / ${formatRate(latestPoint.diskWrite)}`
                                    : '-'
                            }}
                        </strong>
                    </div>
                    <div>
                        <span>{{ $t('xpack.vm.networkRxTx') }}</span>
                        <strong>
                            {{
                                latestPoint
                                    ? `${formatRate(latestPoint.netRx)} / ${formatRate(latestPoint.netTx)}`
                                    : '-'
                            }}
                        </strong>
                    </div>
                </div>
                <el-empty v-if="!points.length && !loading" />
                <div v-else class="vm-monitor-charts">
                    <VCharts type="line" height="220px" :option="cpuOption" />
                    <VCharts type="line" height="220px" :option="memoryOption" />
                    <VCharts type="line" height="220px" :option="diskOption" />
                    <VCharts type="line" height="220px" :option="networkOption" />
                </div>
            </div>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, ref } from 'vue';
import DrawerPro from '@/components/drawer-pro/index.vue';
import VCharts from '@/components/v-charts/index.vue';
import i18n from '@/lang';
import { MsgError } from '@/utils/message';
import { getErrorMessage } from '@/utils/misc';
import { computeSize } from '@/utils/size.js';
import { getVMStats } from '@/api/modules/vm';
import { VM } from '@/api/interface/vm';

const MONITOR_INTERVAL = 5000;
const MONITOR_MAX_POINTS = 60;

type VMMonitorPoint = {
    time: string;
    timestamp: number;
    cpu: number;
    memory: number;
    diskRead: number;
    diskWrite: number;
    netRx: number;
    netTx: number;
};

const visible = ref(false);
const title = ref('');
const loading = ref(false);
const source = ref<VM.Item>();
const previous = ref<VM.Stats>();
const points = ref<VMMonitorPoint[]>([]);
let timer: ReturnType<typeof setInterval> | null = null;

const latestPoint = computed(() => points.value[points.value.length - 1]);

const cpuOption = computed(() =>
    buildMonitorOption(i18n.global.t('xpack.vm.cpuUsage'), [
        { name: i18n.global.t('xpack.vm.cpuUsage'), data: points.value.map((item) => item.cpu) },
    ]),
);

const memoryOption = computed(() =>
    buildMonitorOption(i18n.global.t('xpack.vm.memoryUsage'), [
        { name: i18n.global.t('xpack.vm.memoryUsage'), data: points.value.map((item) => item.memory) },
    ]),
);

const diskOption = computed(() =>
    buildMonitorOption(
        i18n.global.t('xpack.vm.diskIO'),
        [
            { name: i18n.global.t('monitor.read'), data: points.value.map((item) => item.diskRead) },
            { name: i18n.global.t('monitor.write'), data: points.value.map((item) => item.diskWrite) },
        ],
        'KB/s',
    ),
);

const networkOption = computed(() =>
    buildMonitorOption(
        i18n.global.t('xpack.vm.networkIO'),
        [
            { name: i18n.global.t('xpack.vm.rx'), data: points.value.map((item) => item.netRx) },
            { name: i18n.global.t('xpack.vm.tx'), data: points.value.map((item) => item.netTx) },
        ],
        'KB/s',
    ),
);

const open = async (row: VM.Item) => {
    if (row.status !== 'running') {
        MsgError(i18n.global.t('xpack.vm.monitorNotReady'));
        return;
    }
    closeTimer();
    source.value = row;
    title.value = row.name;
    previous.value = undefined;
    points.value = [];
    visible.value = true;
    const loaded = await loadStats();
    if (loaded && visible.value) {
        timer = setInterval(loadStats, MONITOR_INTERVAL);
    }
};

const close = () => {
    closeTimer();
    visible.value = false;
    title.value = '';
    source.value = undefined;
    previous.value = undefined;
    points.value = [];
};

const closeTimer = () => {
    if (timer) {
        clearInterval(timer);
        timer = null;
    }
};

const loadStats = async () => {
    if (!source.value || !visible.value || loading.value) {
        return false;
    }
    loading.value = !points.value.length;
    try {
        const res = await getVMStats(source.value.name);
        const current = res.data;
        if (current.status !== 'running') {
            closeTimer();
            MsgError(i18n.global.t('xpack.vm.monitorNotReady'));
            return false;
        }
        const point = buildMonitorPoint(current, previous.value, source.value);
        previous.value = current;
        points.value = [...points.value, point].slice(-MONITOR_MAX_POINTS);
        return true;
    } catch (error) {
        closeTimer();
        MsgError(loadFailedMessage(error));
        return false;
    } finally {
        loading.value = false;
    }
};

const loadFailedMessage = (error: unknown) => {
    const responseError = error as Record<string, any>;
    const reason =
        responseError?.response?.data?.message ||
        responseError?.response?.data?.msg ||
        responseError?.data?.message ||
        responseError?.data?.msg ||
        getErrorMessage(error);
    if (!reason || reason === '{}' || reason === 'undefined' || reason === 'null') {
        return i18n.global.t('xpack.vm.monitorLoadFailed');
    }
    return i18n.global.t('xpack.vm.monitorLoadFailedWithReason', [reason]);
};

const buildMonitorPoint = (current: VM.Stats, last: VM.Stats | undefined, row: VM.Item): VMMonitorPoint => {
    const intervalSeconds = last ? Math.max((current.timestamp - last.timestamp) / 1000, 1) : 1;
    const cpuBase = intervalSeconds * 1000 * 1000 * 1000 * Math.max(current.vcpu || row.cpu || 1, 1);
    const memoryTotal = current.memory.availableKiB || current.memory.actualKiB || row.memory * 1024;
    const memoryUsed =
        memoryTotal && current.memory.unusedKiB
            ? Math.max(memoryTotal - current.memory.unusedKiB, 0)
            : current.memory.rssKiB || 0;
    return {
        time: formatMonitorTime(current.timestamp),
        timestamp: current.timestamp,
        cpu: last ? toFixedNumber(clamp(((current.cpuTime - last.cpuTime) / cpuBase) * 100, 0, 100)) : 0,
        memory: memoryTotal ? toFixedNumber(clamp((memoryUsed / memoryTotal) * 100, 0, 100)) : 0,
        diskRead: last ? bytesPerSecondToKB(current.disk.readBytes, last.disk.readBytes, intervalSeconds) : 0,
        diskWrite: last ? bytesPerSecondToKB(current.disk.writeBytes, last.disk.writeBytes, intervalSeconds) : 0,
        netRx: last ? bytesPerSecondToKB(current.network.rxBytes, last.network.rxBytes, intervalSeconds) : 0,
        netTx: last ? bytesPerSecondToKB(current.network.txBytes, last.network.txBytes, intervalSeconds) : 0,
    };
};

const buildMonitorOption = (chartTitle: string, yData: Array<{ name: string; data: number[] }>, formatStr = '%') => ({
    title: chartTitle,
    xData: points.value.map((item) => item.time),
    yData,
    formatStr,
    grid: { left: '6%', right: '5%', bottom: '16%', top: '16%' },
});

const formatMonitorTime = (timestamp: number) => {
    const date = new Date(timestamp);
    return [date.getHours(), date.getMinutes(), date.getSeconds()].map((item) => `${item}`.padStart(2, '0')).join(':');
};

const bytesPerSecondToKB = (current: number, last: number, intervalSeconds: number) => {
    return toFixedNumber(Math.max(current - last, 0) / intervalSeconds / 1024);
};

const toFixedNumber = (value: number) => Number((Number.isFinite(value) ? value : 0).toFixed(2));
const clamp = (value: number, min: number, max: number) => Math.min(Math.max(value, min), max);
const formatRate = (value: number) => computeSize(value * 1024) + '/s';

onBeforeUnmount(() => {
    closeTimer();
});

defineExpose({ open });
</script>

<style scoped lang="scss">
.vm-monitor-drawer {
    min-height: 520px;
}

.vm-monitor-summary {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 10px;
    margin-bottom: 16px;
}

.vm-monitor-summary > div {
    min-width: 0;
    padding: 12px;
    border-radius: 6px;
    background: var(--el-fill-color-extra-light);
}

.vm-monitor-summary span {
    display: block;
    margin-bottom: 6px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
}

.vm-monitor-summary strong {
    display: block;
    overflow: hidden;
    color: var(--el-text-color-primary);
    font-size: 15px;
    font-weight: 600;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.vm-monitor-charts {
    display: grid;
    grid-template-columns: 1fr;
    gap: 14px;
}

@media (max-width: 768px) {
    .vm-monitor-summary {
        grid-template-columns: 1fr;
    }
}
</style>
