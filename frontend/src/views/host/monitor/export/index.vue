<template>
    <div>
        <MonitorRouter />

        <LayoutContent v-loading="loading" :title="$t('commons.button.export')" :divider="true">
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
                                    class="w-full"
                                />
                            </el-form-item>
                            <el-form-item :label="$t('monitor.exportMetrics')">
                                <div>
                                    <el-checkbox-group v-model="form.params">
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
                                <el-select v-model="form.io" class="w-full">
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
                                <el-select v-model="form.network" class="w-full">
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
                                <el-select v-model="form.productName" class="w-full">
                                    <el-option :label="$t('monitor.exportAllGPUs')" value="*" />
                                    <el-option v-for="item in gpuOptions" :key="item" :label="item" :value="item" />
                                </el-select>
                            </el-form-item>
                            <el-form-item :label="$t('monitor.exportFormat')">
                                <div>
                                    <el-radio-group v-model="form.format">
                                        <el-radio value="csv">CSV</el-radio>
                                        <el-radio value="json">JSON</el-radio>
                                        <el-radio value="xlsx">Excel (xlsx)</el-radio>
                                    </el-radio-group>
                                    <span class="input-help">{{ formatHelper }}</span>
                                </div>
                            </el-form-item>
                            <el-form-item>
                                <el-button
                                    v-permission:view
                                    v-node-admin
                                    type="primary"
                                    @click="onExport"
                                    :disabled="loading"
                                >
                                    {{ $t('commons.button.export') }}
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
import { computed, onMounted, reactive, ref } from 'vue';
import { loadMonitor, getIOOptions, getNetworkOptions } from '@/api/modules/host';
import { getGPUOptions, loadGPUMonitor } from '@/api/modules/ai';
import { Host } from '@/api/interface/host';
import { AI } from '@/api/interface/ai';
import { shortcuts } from '@/utils/shortcuts';
import { dateFormatForName } from '@/utils/date';
import { exportMonitorTables, MonitorExportTable } from '@/utils/monitor-export';
import { useGlobalStore } from '@/composables/useGlobalStore';
import MonitorRouter from '@/views/host/monitor/index.vue';
import i18n from '@/lang';
import { MsgWarning } from '@/utils/message';

const { currentNode, isMobile } = useGlobalStore();

const loading = ref(false);
const timeRange = ref<[Date, Date]>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
const ioOptions = ref<Array<string>>([]);
const netOptions = ref<Array<string>>([]);
const gpuOptions = ref<Array<string>>([]);
const ioOptionsFailed = ref(false);
const netOptionsFailed = ref(false);
const gpuOptionsFailed = ref(false);

// el-select treats '' as empty and renders the placeholder, so "all devices" uses '*' as a sentinel
const ALL_DEVICES = '*';

const form = reactive({
    params: ['cpu', 'memory', 'load'] as Array<string>,
    io: ALL_DEVICES,
    network: ALL_DEVICES,
    productName: ALL_DEVICES,
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

// map internal table names to the localized metric labels shown in warnings
const tableLabel = (name: string): string => {
    switch (name) {
        case 'base':
            return [
                ['cpu', 'CPU'],
                ['memory', t('monitor.memory')],
                ['load', t('monitor.avgLoad')],
            ]
                .filter(([param]) => form.params.includes(param as string))
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

const buildBaseTable = (datasets: Array<Host.MonitorData>): MonitorExportTable => {
    const base = datasets.find((item) => item.param === 'base');
    const fields = ['time'];
    const headers = [t('commons.table.date')];
    const pickers: Array<(row: any) => any> = [(row) => new Date(row.createdAt)];
    if (form.params.includes('cpu')) {
        fields.push('cpu');
        headers.push('CPU (%)');
        pickers.push((row) => row.cpu);
    }
    if (form.params.includes('memory')) {
        fields.push('memory');
        headers.push(`${t('monitor.memory')} (%)`);
        pickers.push((row) => row.memory);
    }
    if (form.params.includes('load')) {
        fields.push('loadUsage', 'load1', 'load5', 'load15');
        headers.push(`${t('monitor.avgLoad')} (%)`, 'Load 1m', 'Load 5m', 'Load 15m');
        pickers.push(
            (row) => row.loadUsage,
            (row) => row.cpuLoad1,
            (row) => row.cpuLoad5,
            (row) => row.cpuLoad15,
        );
    }
    const rows = (base?.value || []).map((row) => pickers.map((pick) => pick(row)));
    return { name: 'base', fields, headers, rows };
};

const buildIOTable = (datasets: Array<Host.MonitorData>): MonitorExportTable => {
    const rows = datasets
        .flatMap((item) => (item.param === 'io' ? item.value || [] : []))
        .filter((row) => form.io !== ALL_DEVICES || row.name !== 'all')
        .map((row) => [new Date(row.createdAt), row.name, row.read, row.write, row.count, row.time])
        .sort((a, b) => a[0].getTime() - b[0].getTime());
    return {
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
        rows,
    };
};

const buildNetworkTable = (datasets: Array<Host.MonitorData>): MonitorExportTable => {
    const rows = datasets
        .flatMap((item) => (item.param === 'network' ? item.value || [] : []))
        .filter((row) => form.network !== ALL_DEVICES || row.name !== 'all')
        .map((row) => [new Date(row.createdAt), row.name, row.up, row.down])
        .sort((a, b) => a[0].getTime() - b[0].getTime());
    return {
        name: 'network',
        fields: ['time', 'interface', 'up', 'down'],
        headers: [
            t('commons.table.date'),
            t('monitor.networkCard'),
            `${t('monitor.up')} (KB/s)`,
            `${t('monitor.down')} (KB/s)`,
        ],
        rows,
    };
};

const buildGPUTable = (datasets: Array<{ product: string; data: AI.MonitorGPUData }>): MonitorExportTable => {
    const rows = datasets
        .flatMap(({ product, data }) =>
            (data.date || []).map((date, i) => [
                new Date(date),
                product,
                data.gpuValue?.[i],
                data.temperatureValue?.[i],
                data.powerUsed?.[i],
                data.powerTotal?.[i],
                data.memoryUsed?.[i],
                data.memoryTotal?.[i],
                data.speedValue?.[i],
            ]),
        )
        .sort((a, b) => (a[0] as Date).getTime() - (b[0] as Date).getTime());
    return {
        name: 'gpu',
        fields: [
            'time',
            'productName',
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
            `${t('aiTools.gpu.gpuUtil')} (%)`,
            `${t('aiTools.gpu.temperature')} (°C)`,
            `${t('aiTools.gpu.powerUsage')} (W)`,
            `${t('aiTools.gpu.powerLimit')} (W)`,
            `${t('aiTools.gpu.memoryUsed')} (MiB)`,
            `${t('aiTools.gpu.memoryTotal')} (MiB)`,
            t('aiTools.gpu.fanSpeed'),
        ],
        rows,
    };
};

const collectTables = async (): Promise<MonitorExportTable[]> => {
    const [startTime, endTime] = timeRange.value;
    const jobs: Promise<void>[] = [];
    const datasets: Array<Host.MonitorData> = [];
    const gpuDatasets: Array<{ product: string; data: AI.MonitorGPUData }> = [];

    const search = (param: string, io = 'all', network = 'all') =>
        loadMonitor({ param, io, network, startTime, endTime }, currentNode.value).then((res) => {
            datasets.push(...(res.data || []));
        });

    if (form.params.some((item) => ['cpu', 'memory', 'load'].includes(item))) {
        jobs.push(search('load'));
    }
    if (form.params.includes('io')) {
        // empty io queries every device in one request; 'all' selects the stored aggregate rows only
        jobs.push(search('io', form.io === ALL_DEVICES ? '' : form.io));
    }
    if (form.params.includes('network')) {
        jobs.push(search('network', 'all', form.network === ALL_DEVICES ? '' : form.network));
    }
    if (form.params.includes('gpu')) {
        for (const product of form.productName === ALL_DEVICES ? gpuOptions.value : [form.productName]) {
            jobs.push(
                loadGPUMonitor({ productName: product, startTime, endTime }, currentNode.value).then((res) => {
                    gpuDatasets.push({ product, data: res.data });
                }),
            );
        }
    }
    await Promise.all(jobs);

    const tables: MonitorExportTable[] = [];
    if (form.params.some((item) => ['cpu', 'memory', 'load'].includes(item))) {
        tables.push(buildBaseTable(datasets));
    }
    if (form.params.includes('io')) {
        tables.push(buildIOTable(datasets));
    }
    if (form.params.includes('network')) {
        tables.push(buildNetworkTable(datasets));
    }
    if (form.params.includes('gpu')) {
        tables.push(buildGPUTable(gpuDatasets));
    }
    return tables;
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
    loading.value = true;
    try {
        await optionsReady.value;
        // retry only the lists that failed on mount; a stale but non-empty gpu list is still usable
        await loadOptions(true);
        if (
            form.params.includes('gpu') &&
            form.productName === ALL_DEVICES &&
            gpuOptionsFailed.value &&
            gpuOptions.value.length === 0
        ) {
            MsgWarning(t('monitor.exportOptionsFailed', ['GPU']));
            return;
        }
        const tables = await collectTables();
        if (tables.every((table) => table.rows.length === 0)) {
            MsgWarning(t('commons.msg.noneData'));
            return;
        }
        const emptyTables = tables.filter((table) => table.rows.length === 0).map((table) => tableLabel(table.name));
        const truncated = await exportMonitorTables(
            tables,
            form.format,
            `1panel-monitor-${dateFormatForName(new Date())}`,
        );
        if (emptyTables.length > 0) {
            MsgWarning(t('monitor.exportEmptyTables', [emptyTables.join(', ')]));
        }
        if (truncated.length > 0) {
            MsgWarning(t('monitor.exportTruncated', [truncated.map(tableLabel).join(', ')]));
        }
    } finally {
        loading.value = false;
    }
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
        gpuOptions.value = res.data?.options || [];
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
</script>
