<template>
    <div>
        <MonitorRouter />

        <LayoutContent :title="$t('commons.button.export')" :divider="true">
            <template #main>
                <el-form @submit.prevent label-position="left" label-width="160px">
                    <el-row>
                        <el-col :span="1"><br /></el-col>
                        <el-col :xs="24" :sm="20" :md="16" :lg="12" :xl="10">
                            <el-form-item :label="$t('monitor.exportRange')">
                                <el-date-picker
                                    v-model="timeRange"
                                    type="datetimerange"
                                    range-separator="-"
                                    :start-placeholder="$t('commons.search.timeStart')"
                                    :end-placeholder="$t('commons.search.timeEnd')"
                                    :shortcuts="shortcuts"
                                    :size="isMobile ? 'small' : 'default'"
                                    :disabled="exporting"
                                    class="w-full"
                                />
                            </el-form-item>
                            <el-form-item :label="$t('monitor.exportMetrics')">
                                <div>
                                    <el-checkbox-group v-model="form.params" :disabled="exporting">
                                        <el-checkbox label="CPU" value="cpu" />
                                        <el-checkbox :label="$t('monitor.memory')" value="memory" />
                                        <el-checkbox :label="$t('monitor.avgLoad')" value="load" />
                                        <el-checkbox :label="$t('monitor.disk') + ' I/O'" value="io" />
                                        <el-checkbox :label="$t('monitor.network')" value="network" />
                                        <el-checkbox
                                            v-if="gpuOptions.length > 0 || gpuOptionsFailed"
                                            label="GPU"
                                            value="gpu"
                                        />
                                    </el-checkbox-group>
                                    <span class="input-help">{{ $t('monitor.exportMetricsHelper') }}</span>
                                </div>
                            </el-form-item>
                            <el-form-item v-if="form.params.includes('io')" :label="$t('monitor.disk')">
                                <el-select v-model="form.io" :disabled="exporting" class="w-full">
                                    <el-option :label="$t('monitor.exportAllDevices')" value="*" />
                                    <el-option
                                        v-for="item in ioOptions"
                                        :key="item"
                                        :label="item === 'all' ? $t('monitor.exportAggregate') : item"
                                        :value="item"
                                    />
                                </el-select>
                            </el-form-item>
                            <el-form-item v-if="form.params.includes('network')" :label="$t('monitor.networkCard')">
                                <el-select v-model="form.network" :disabled="exporting" class="w-full">
                                    <el-option :label="$t('monitor.exportAllInterfaces')" value="*" />
                                    <el-option
                                        v-for="item in netOptions"
                                        :key="item"
                                        :label="item === 'all' ? $t('monitor.exportAggregate') : item"
                                        :value="item"
                                    />
                                </el-select>
                            </el-form-item>
                            <el-form-item v-if="form.params.includes('gpu')" label="GPU">
                                <el-select v-model="form.gpuDevice" :disabled="exporting" class="w-full">
                                    <el-option :label="$t('monitor.exportAllDevices')" value="*" />
                                    <el-option
                                        v-for="item in gpuOptions"
                                        :key="item.value"
                                        :label="item.label"
                                        :value="item.value"
                                    />
                                </el-select>
                            </el-form-item>
                            <el-form-item :label="$t('monitor.exportFormat')">
                                <div>
                                    <el-radio-group v-model="form.format" :disabled="exporting">
                                        <el-radio value="csv">CSV</el-radio>
                                        <el-radio value="json">JSON</el-radio>
                                        <el-radio value="xlsx">Excel (xlsx)</el-radio>
                                    </el-radio-group>
                                    <span class="input-help">{{ formatHelper }}</span>
                                </div>
                            </el-form-item>
                            <el-form-item v-if="exporting">
                                <div class="w-full">
                                    <el-progress :percentage="progress" />
                                    <span v-if="progressTime" class="input-help">
                                        {{ $t('monitor.exportProgress', [progressTime]) }}
                                    </span>
                                </div>
                            </el-form-item>
                            <el-form-item>
                                <el-button
                                    v-permission:view
                                    v-node-admin
                                    type="primary"
                                    @click="onExport"
                                    :loading="exporting"
                                >
                                    {{ $t('commons.button.export') }}
                                </el-button>
                                <el-button v-if="exporting" @click="onCancel">
                                    {{ $t('commons.button.cancel') }}
                                </el-button>
                            </el-form-item>
                        </el-col>
                    </el-row>
                </el-form>
            </template>
        </LayoutContent>
    </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, reactive, ref } from 'vue';
import { loadMonitor, getIOOptions, getNetworkOptions } from '@/api/modules/host';
import { getGPUOptions, loadGPUMonitor } from '@/api/modules/ai';
import type { RequestConfig } from '@/api';
import { Host } from '@/api/interface/host';
import { AI } from '@/api/interface/ai';
import { shortcuts } from '@/utils/shortcuts';
import { dateFormatForName, dateFormatSimpleWithSecond } from '@/utils/date';
import { downloadBlob } from '@/utils/file';
import { createMonitorExportWriter, MonitorExportColumns } from '@/utils/monitor-export';
import {
    createBoundaryDeduper,
    ExportAbortedError,
    ExportWindow,
    ExportWindowError,
    isTransientRequestError,
    runExportWindows,
} from '@/utils/monitor-export-window';
import { useGlobalStore } from '@/composables/useGlobalStore';
import MonitorRouter from '@/views/host/monitor/index.vue';
import i18n from '@/lang';
import { MsgError, MsgInfo, MsgWarning } from '@/utils/message';

const { currentNode, isMobile } = useGlobalStore();

type GPUOption = AI.ChartHide & { value: string; label: string };

const exporting = ref(false);
const progress = ref(0);
const progressTime = ref('');
const timeRange = ref<[Date, Date]>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
const ioOptions = ref<Array<string>>([]);
const netOptions = ref<Array<string>>([]);
const gpuOptions = ref<GPUOption[]>([]);
const ioOptionsFailed = ref(false);
const netOptionsFailed = ref(false);
const gpuOptionsFailed = ref(false);

let exportController: AbortController | null = null;
let unmounted = false;

// el-select treats '' as empty and renders the placeholder, so "all devices" uses '*' as a sentinel
const ALL_DEVICES = '*';

// rows one request should carry; long ranges are split into windows sized towards these targets.
// base rows embed the top-process JSON, so they are far heavier than io / network rows
const BASE_ROWS_PER_REQUEST = 1000;
const DEVICE_ROWS_PER_REQUEST = 5000;
// the GPU history API averages into buckets above 1200 samples, so stay below it to export raw samples
const GPU_SAMPLES_PER_REQUEST = 1000;

const form = reactive({
    params: ['cpu', 'memory', 'load'] as Array<string>,
    io: ALL_DEVICES,
    network: ALL_DEVICES,
    gpuDevice: ALL_DEVICES,
    format: 'csv',
});

const formatHelper = computed(() => {
    switch (form.format) {
        case 'json':
            return i18n.global.t('monitor.exportJsonHelper');
        case 'xlsx':
            return i18n.global.t('monitor.exportXlsxHelper');
        default:
            return i18n.global.t('monitor.exportCsvHelper');
    }
});

const t = (key: string, args?: any) => i18n.global.t(key, args);

// settings captured when the export starts, so a long export is not affected by later form edits
interface ExportPlan {
    params: string[];
    io: string;
    network: string;
    gpuDevices: GPUOption[];
    node: string;
    format: string;
}

interface ExportRow {
    // identifies the row across overlapping window boundaries
    key: string;
    time: number;
    cells: any[];
}

interface ExportBatch {
    rows: ExportRow[];
    load: number;
    overflow?: boolean;
}

interface ExportSource {
    columns: MonitorExportColumns;
    fetch: (window: ExportWindow, config: RequestConfig) => Promise<ExportBatch>;
}

// map internal table names to the localized metric labels shown in warnings
const tableLabel = (name: string, params: string[]): string => {
    switch (name) {
        case 'base':
            return [
                ['cpu', 'CPU'],
                ['memory', t('monitor.memory')],
                ['load', t('monitor.avgLoad')],
            ]
                .filter(([param]) => params.includes(param as string))
                .map(([, label]) => label)
                .join(' / ');
        case 'io':
            return `${t('monitor.disk')} I/O`;
        case 'network':
            return t('monitor.network');
        case 'gpu':
            return 'GPU';
        default:
            return name;
    }
};

const toRow = (key: string, cells: any[]): ExportRow => ({ key, time: (cells[0] as Date).getTime(), cells });

const monitorRowKey = (row: any) => (row.id !== undefined ? String(row.id) : `${row.name ?? ''}|${row.createdAt}`);

const searchMonitor = async (
    plan: ExportPlan,
    window: ExportWindow,
    config: RequestConfig,
    param: string,
    io = 'all',
    network = 'all',
): Promise<Array<Host.MonitorData>> => {
    const res = await loadMonitor(
        { param, io, network, startTime: new Date(window.queryFrom), endTime: new Date(window.to) },
        plan.node,
        config,
    );
    if (!res) {
        // the interceptor resolves without a body while the panel is busy; never mistake that for "no data"
        throw new Error('empty monitor response');
    }
    return res.data || [];
};

const baseSource = (plan: ExportPlan): ExportSource => {
    const fields = ['time'];
    const headers = [t('commons.table.date')];
    const pickers: Array<(row: any) => any> = [(row) => new Date(row.createdAt)];
    if (plan.params.includes('cpu')) {
        fields.push('cpu');
        headers.push('CPU (%)');
        pickers.push((row) => row.cpu);
    }
    if (plan.params.includes('memory')) {
        fields.push('memory');
        headers.push(`${t('monitor.memory')} (%)`);
        pickers.push((row) => row.memory);
    }
    if (plan.params.includes('load')) {
        fields.push('loadUsage', 'load1', 'load5', 'load15');
        headers.push(`${t('monitor.avgLoad')} (%)`, 'Load 1m', 'Load 5m', 'Load 15m');
        pickers.push(
            (row) => row.loadUsage,
            (row) => row.cpuLoad1,
            (row) => row.cpuLoad5,
            (row) => row.cpuLoad15,
        );
    }
    return {
        columns: { name: 'base', fields, headers },
        fetch: async (window, config) => {
            const datasets = await searchMonitor(plan, window, config, 'load');
            const values = datasets.find((item) => item.param === 'base')?.value || [];
            return {
                rows: values.map((row) =>
                    toRow(
                        monitorRowKey(row),
                        pickers.map((pick) => pick(row)),
                    ),
                ),
                load: values.length / BASE_ROWS_PER_REQUEST,
            };
        },
    };
};

const ioSource = (plan: ExportPlan): ExportSource => ({
    columns: {
        name: 'io',
        fields: ['time', 'device', 'read', 'write', 'count', 'ioTime'],
        headers: [
            t('commons.table.date'),
            t('monitor.disk'),
            `${t('monitor.read')} (B/s)`,
            `${t('monitor.write')} (B/s)`,
            t('monitor.readWriteCount'),
            `${t('monitor.readWriteTime')} (ms)`,
        ],
    },
    fetch: async (window, config) => {
        // empty io queries every device in one request; 'all' selects the stored aggregate rows only
        const datasets = await searchMonitor(plan, window, config, 'io', plan.io === ALL_DEVICES ? '' : plan.io);
        const values = datasets.flatMap((item) => (item.param === 'io' ? item.value || [] : []));
        return {
            rows: values
                .filter((row) => plan.io !== ALL_DEVICES || row.name !== 'all')
                .map((row) =>
                    toRow(monitorRowKey(row), [
                        new Date(row.createdAt),
                        row.name,
                        row.read,
                        row.write,
                        row.count,
                        row.time,
                    ]),
                ),
            load: values.length / DEVICE_ROWS_PER_REQUEST,
        };
    },
});

const networkSource = (plan: ExportPlan): ExportSource => ({
    columns: {
        name: 'network',
        fields: ['time', 'interface', 'up', 'down'],
        headers: [
            t('commons.table.date'),
            t('monitor.networkCard'),
            `${t('monitor.up')} (KB/s)`,
            `${t('monitor.down')} (KB/s)`,
        ],
    },
    fetch: async (window, config) => {
        const datasets = await searchMonitor(
            plan,
            window,
            config,
            'network',
            'all',
            plan.network === ALL_DEVICES ? '' : plan.network,
        );
        const values = datasets.flatMap((item) => (item.param === 'network' ? item.value || [] : []));
        return {
            rows: values
                .filter((row) => plan.network !== ALL_DEVICES || row.name !== 'all')
                .map((row) => toRow(monitorRowKey(row), [new Date(row.createdAt), row.name, row.up, row.down])),
            load: values.length / DEVICE_ROWS_PER_REQUEST,
        };
    },
});

const gpuSource = (plan: ExportPlan): ExportSource => ({
    columns: {
        name: 'gpu',
        fields: [
            'time',
            'productName',
            'deviceID',
            'legacy',
            'gpuUtil',
            'temperature',
            'powerDraw',
            'powerLimit',
            'memUsed',
            'memTotal',
            'fanSpeed',
        ],
        headers: [
            t('commons.table.date'),
            'GPU',
            'GPU ID',
            t('aiTools.gpu.legacyDevice'),
            `${t('aiTools.gpu.gpuUtil')} (%)`,
            `${t('aiTools.gpu.temperature')} (°C)`,
            `${t('aiTools.gpu.powerUsage')} (W)`,
            `${t('aiTools.gpu.powerLimit')} (W)`,
            `${t('aiTools.gpu.memoryUsed')} (MiB)`,
            `${t('aiTools.gpu.memoryTotal')} (MiB)`,
            t('aiTools.gpu.fanSpeed'),
        ],
    },
    fetch: async (window, config) => {
        const results = await Promise.all(
            plan.gpuDevices.map(async (device) => {
                const res = await loadGPUMonitor(
                    {
                        deviceID: device.deviceID,
                        productName: device.productName,
                        legacy: device.legacy,
                        startTime: new Date(window.queryFrom),
                        endTime: new Date(window.to),
                    },
                    plan.node,
                    config,
                );
                if (!res) {
                    throw new Error('empty GPU monitor response');
                }
                return { device, data: res.data };
            }),
        );
        const batch: ExportBatch = { rows: [], load: 0 };
        for (const { device, data } of results) {
            if (!data) {
                continue;
            }
            batch.load = Math.max(batch.load, (data.sampleCount || 0) / GPU_SAMPLES_PER_REQUEST);
            if (data.bucketSeconds > 0) {
                // averaged buckets instead of raw samples: ask for a smaller window
                batch.overflow = true;
            }
            (data.date || []).forEach((date, i) => {
                const values = [
                    data.gpuValue?.[i],
                    data.temperatureValue?.[i],
                    data.powerUsed?.[i],
                    data.powerTotal?.[i],
                    data.memoryUsed?.[i],
                    data.memoryTotal?.[i],
                    data.speedValue?.[i],
                ];
                // the API marks collection gaps with empty points; they carry no sample, so leave them out
                if (values.every((value) => value === null || value === undefined)) {
                    return;
                }
                batch.rows.push(
                    toRow(`${device.value}|${date}`, [
                        new Date(date),
                        device.productName,
                        device.deviceID || '',
                        !!device.legacy,
                        ...values,
                    ]),
                );
            });
        }
        return batch;
    },
});

const buildSources = (plan: ExportPlan): ExportSource[] => {
    const sources: ExportSource[] = [];
    if (plan.params.some((item) => ['cpu', 'memory', 'load'].includes(item))) {
        sources.push(baseSource(plan));
    }
    if (plan.params.includes('io')) {
        sources.push(ioSource(plan));
    }
    if (plan.params.includes('network')) {
        sources.push(networkSource(plan));
    }
    if (plan.params.includes('gpu')) {
        sources.push(gpuSource(plan));
    }
    return sources;
};

// fetches the range window by window and streams every committed window straight into the file writer
const exportSources = async (
    plan: ExportPlan,
    sources: ExportSource[],
    start: Date,
    end: Date,
    controller: AbortController,
) => {
    const writer = createMonitorExportWriter(
        sources.map((source) => source.columns),
        plan.format,
        `1panel-monitor-${dateFormatForName(new Date())}`,
    );
    const dedupers = sources.map(() => createBoundaryDeduper());

    await runExportWindows({
        start: start.getTime(),
        end: end.getTime(),
        signal: controller.signal,
        isRetryable: isTransientRequestError,
        fetch: async (window) => {
            // one attempt per window: when any request fails, cancel its siblings before retrying smaller
            const attempt = new AbortController();
            const onAbort = () => attempt.abort();
            controller.signal.addEventListener('abort', onAbort);
            try {
                const config: RequestConfig = { signal: attempt.signal, errorMessageMode: 'manual' };
                const batches = await Promise.all(sources.map((source) => source.fetch(window, config))).catch(
                    (error) => {
                        attempt.abort();
                        throw error;
                    },
                );
                return {
                    data: batches,
                    load: Math.max(0, ...batches.map((batch) => batch.load)),
                    overflow: batches.some((batch) => batch.overflow),
                };
            } finally {
                controller.signal.removeEventListener('abort', onAbort);
            }
        },
        commit: async (batches) => {
            for (let i = 0; i < sources.length; i++) {
                const rows = dedupers[i](batches[i].rows, (row) => row.key).sort((a, b) => a.time - b.time);
                await writer.append(
                    sources[i].columns.name,
                    rows.map((row) => row.cells),
                );
            }
        },
        onProgress: (ratio, window) => {
            progress.value = Math.min(100, Math.floor(ratio * 100));
            progressTime.value = window.to > window.from ? dateFormatSimpleWithSecond(window.to) : '';
        },
    });
    return writer;
};

const onExport = async () => {
    if (form.params.length === 0) {
        MsgWarning(t('monitor.selectMetrics'));
        return;
    }
    if (!timeRange.value || timeRange.value.length !== 2) {
        MsgWarning(t('monitor.selectTimeRange'));
        return;
    }
    const controller = new AbortController();
    exportController = controller;
    exporting.value = true;
    progress.value = 0;
    progressTime.value = '';
    try {
        await optionsReady.value;
        // retry only the lists that failed on mount; a stale but non-empty gpu list is still usable
        await loadOptions(true);
        if (controller.signal.aborted) {
            throw new ExportAbortedError();
        }
        if (
            form.params.includes('gpu') &&
            form.gpuDevice === ALL_DEVICES &&
            gpuOptionsFailed.value &&
            gpuOptions.value.length === 0
        ) {
            MsgWarning(t('monitor.exportOptionsFailed', ['GPU']));
            return;
        }
        const plan: ExportPlan = {
            params: [...form.params],
            io: form.io,
            network: form.network,
            gpuDevices: gpuOptions.value.filter(
                (device) => form.gpuDevice === ALL_DEVICES || device.value === form.gpuDevice,
            ),
            node: currentNode.value,
            format: form.format,
        };
        const [startTime, endTime] = timeRange.value;
        const sources = buildSources(plan);
        const writer = await exportSources(plan, sources, startTime, endTime, controller);

        const names = sources.map((source) => source.columns.name);
        const emptyTables = names.filter((name) => writer.rowCount(name) === 0);
        if (emptyTables.length === names.length) {
            MsgWarning(t('commons.msg.noneData'));
            return;
        }
        const file = await writer.finish();
        downloadBlob(file.content, file.fileName, file.type);
        const labels = (list: string[]) => list.map((name) => tableLabel(name, plan.params)).join(', ');
        if (emptyTables.length > 0) {
            MsgWarning(t('monitor.exportEmptyTables', [labels(emptyTables)]));
        }
        if (file.truncated.length > 0) {
            MsgWarning(t('monitor.exportTruncated', [labels(file.truncated)]));
        }
    } catch (error) {
        if (unmounted) {
            return;
        }
        if (error instanceof ExportAbortedError) {
            MsgInfo(t('monitor.exportCancelled'));
            return;
        }
        if (error instanceof ExportWindowError) {
            const cause = error.cause as any;
            const message = cause?.response?.data?.message || (typeof cause?.code === 'number' ? cause.message : '');
            // a rejected request carries the backend's own explanation; transient failures get a retry hint
            if (!isTransientRequestError(cause) && message) {
                MsgError(message);
            } else {
                MsgError(t('monitor.exportFailed', [dateFormatSimpleWithSecond(error.window.from)]));
            }
            return;
        }
        throw error;
    } finally {
        exporting.value = false;
        exportController = null;
    }
};

const onCancel = () => {
    exportController?.abort();
};

const optionsReady = ref<Promise<void>>(Promise.resolve());

const loadIOOptions = async () => {
    try {
        const res = await getIOOptions(currentNode.value);
        ioOptions.value = res.data || [];
        ioOptionsFailed.value = false;
    } catch {
        ioOptionsFailed.value = true;
    }
};

const loadNetOptions = async () => {
    try {
        const res = await getNetworkOptions(currentNode.value);
        netOptions.value = res.data || [];
        netOptionsFailed.value = false;
    } catch {
        netOptionsFailed.value = true;
    }
};

const loadGPUOptions = async () => {
    try {
        const res = await getGPUOptions(currentNode.value);
        gpuOptions.value = (res.data?.supported ? res.data.chartHide || [] : []).map((device) => ({
            ...device,
            value: device.deviceID || `legacy:${device.productName}`,
            label: `${device.productName} · ${device.deviceID || t('aiTools.gpu.legacyDevice')}`,
        }));
        gpuOptionsFailed.value = false;
    } catch {
        gpuOptionsFailed.value = true;
    }
};

const loadOptions = async (onlyFailed = false) => {
    const jobs: Promise<void>[] = [];
    if (!onlyFailed || ioOptionsFailed.value) {
        jobs.push(loadIOOptions());
    }
    if (!onlyFailed || netOptionsFailed.value) {
        jobs.push(loadNetOptions());
    }
    if (!onlyFailed || gpuOptionsFailed.value) {
        jobs.push(loadGPUOptions());
    }
    await Promise.all(jobs);
};

onMounted(() => {
    optionsReady.value = loadOptions();
});

onBeforeUnmount(() => {
    // leaving the page stops the export instead of letting it keep fetching in the background
    unmounted = true;
    exportController?.abort();
});
</script>
