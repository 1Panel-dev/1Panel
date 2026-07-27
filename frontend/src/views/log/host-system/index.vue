<template>
    <LayoutContent v-loading="loading" :title="$t('logs.hostSystem')">
        <template #rightToolBar>
            <el-input v-model="keyword" class="p-w-200" clearable :placeholder="$t('logs.filter')" />
            <el-date-picker
                v-model="timeRange"
                class="p-w-360"
                type="datetimerange"
                range-separator="-"
                :start-placeholder="$t('commons.search.timeStart')"
                :end-placeholder="$t('commons.search.timeEnd')"
                :shortcuts="shortcuts"
                @change="changeTimeRange"
            />
            <el-select v-model="priority" class="p-w-200" clearable>
                <template #prefix>{{ $t('logs.priority') }}</template>
                <el-option
                    v-for="item in priorities"
                    :key="item.value"
                    :label="priorityLabel(item.value)"
                    :value="item.value"
                />
            </el-select>
            <el-select v-model="service" class="p-w-200" clearable filterable allow-create>
                <template #prefix>{{ $t('logs.service') }}</template>
                <el-option v-for="item in services" :key="item" :label="item" :value="item" />
            </el-select>
            <el-checkbox border v-model="watching" @change="changeWatch">
                {{ $t('commons.button.watch') }}
            </el-checkbox>
            <TableRefresh @search="resetAndLoadLogs" />
        </template>
        <template #main>
            <ComplexTable :data="logs" :pagination-config="paginationConfig" @row-click="openDetail">
                <el-table-column prop="time" :label="$t('commons.table.date')" width="220" show-overflow-tooltip />
                <el-table-column :label="$t('logs.priority')" width="120">
                    <template #default="{ row }">
                        <el-tag size="small" :type="priorityType(row.priority)">
                            {{ priorityLabel(row.priority) }}
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('logs.service')" width="220" show-overflow-tooltip>
                    <template #default="{ row }">{{ row.service || '-' }}</template>
                </el-table-column>
                <el-table-column prop="message" :label="$t('logs.message')" min-width="360" show-overflow-tooltip />
                <template #pagination>
                    <div class="flex items-center gap-2">
                        <el-select v-model="paginationConfig.pageSize" class="p-w-100" @change="changePageSize">
                            <el-option v-for="size in pageSizes" :key="size" :label="String(size)" :value="size" />
                        </el-select>
                        <el-button :disabled="cursorIndex === 0 || loading" @click="loadPreviousPage">
                            {{ $t('logs.previousPage') }}
                        </el-button>
                        <el-button :disabled="!hasMore || loading" @click="loadNextPage">
                            {{ $t('logs.nextPage') }}
                        </el-button>
                    </div>
                </template>
            </ComplexTable>
        </template>
    </LayoutContent>

    <el-drawer v-model="detailOpen" :title="$t('logs.hostSystem')" size="50%">
        <el-descriptions v-if="selectedLog" :column="1" border>
            <el-descriptions-item :label="$t('commons.table.date')">{{ selectedLog.time || '-' }}</el-descriptions-item>
            <el-descriptions-item :label="$t('logs.priority')">
                {{ priorityLabel(selectedLog.priority) }}
            </el-descriptions-item>
            <el-descriptions-item :label="$t('logs.service')">{{ selectedLog.service || '-' }}</el-descriptions-item>
            <el-descriptions-item :label="$t('logs.message')">{{ selectedLog.message }}</el-descriptions-item>
        </el-descriptions>
        <pre v-if="selectedLog" class="log-raw">{{ selectedLog.raw }}</pre>
    </el-drawer>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, reactive, ref, watch } from 'vue';
import { Log } from '@/api/interface/log';
import { listRunningServices, readSystemLogs } from '@/api/modules/log';
import { useGlobalStore } from '@/composables/useGlobalStore';
import i18n from '@/lang';
import { shortcuts } from '@/utils/shortcuts';

const { currentNode } = useGlobalStore();
const logs = ref<Log.SystemLogItem[]>([]);
const loading = ref(false);
const watching = ref(false);
const keyword = ref('');
const priority = ref('');
const service = ref('');
const services = ref<string[]>([]);
const detailOpen = ref(false);
const selectedLog = ref<Log.SystemLogItem>();
const timeRange = ref<[Date, Date]>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
const paginationConfig = reactive({
    pageSize: 100,
    cacheSizeKey: 'host-system-logs-page-size',
});
const pageSizes = [5, 10, 20, 50, 100, 200, 500];
const cursorHistory = ref<string[]>(['']);
const cursorIndex = ref(0);
const hasMore = ref(false);
let timer: ReturnType<typeof setTimeout> | undefined;
let filterTimer: ReturnType<typeof setTimeout> | undefined;
let requestID = 0;

const priorities = [
    { value: '0', labelKey: 'logs.priorityEmergency' },
    { value: '1', labelKey: 'logs.priorityAlert' },
    { value: '2', labelKey: 'logs.priorityCritical' },
    { value: '3', labelKey: 'logs.priorityError' },
    { value: '4', labelKey: 'logs.priorityWarning' },
    { value: '5', labelKey: 'logs.priorityNotice' },
    { value: '6', labelKey: 'logs.priorityInfo' },
    { value: '7', labelKey: 'logs.priorityDebug' },
];

const loadLogs = async (cursor = cursorHistory.value[cursorIndex.value]) => {
    const currentRequestID = ++requestID;
    loading.value = true;
    try {
        const res = await readSystemLogs(
            {
                pageSize: paginationConfig.pageSize,
                cursor,
                startTime: timeRange.value?.[0],
                endTime: timeRange.value?.[1],
                keyword: keyword.value,
                priority: priority.value,
                service: service.value,
            },
            currentNode.value,
        );
        if (currentRequestID !== requestID) return;
        logs.value = res.data.items || [];
        hasMore.value = res.data.hasMore;
        if (res.data.nextCursor) {
            cursorHistory.value[cursorIndex.value + 1] = res.data.nextCursor;
        }
    } finally {
        if (currentRequestID === requestID) loading.value = false;
    }
};

const resetAndLoadLogs = () => {
    requestID++;
    cursorHistory.value = [''];
    cursorIndex.value = 0;
    hasMore.value = false;
    return loadLogs('');
};

const loadNextPage = () => {
    if (!hasMore.value) return;
    cursorIndex.value++;
    loadLogs();
};

const loadPreviousPage = () => {
    if (cursorIndex.value === 0) return;
    cursorIndex.value--;
    loadLogs();
};

const changePageSize = () => {
    localStorage.setItem(paginationConfig.cacheSizeKey, String(paginationConfig.pageSize));
    resetAndLoadLogs();
};

const loadServices = async () => {
    try {
        const res = await listRunningServices(currentNode.value);
        services.value = res.data || [];
    } catch {
        services.value = [];
    }
};

const scheduleWatch = () => {
    timer = setTimeout(async () => {
        if (!watching.value) return;
        try {
            if (!loading.value) {
                await resetAndLoadLogs();
            }
        } finally {
            if (watching.value) {
                scheduleWatch();
            }
        }
    }, 3000);
};

const changeWatch = () => {
    if (timer) {
        clearTimeout(timer);
        timer = undefined;
    }
    if (watching.value) {
        scheduleWatch();
    }
};

const changeTimeRange = () => {
    resetAndLoadLogs();
};

const openDetail = (row: Log.SystemLogItem) => {
    selectedLog.value = row;
    detailOpen.value = true;
};

const priorityLabel = (value: string) => {
    const priority = priorities.find((item) => item.value === value);
    return priority ? i18n.global.t(priority.labelKey) : value || '-';
};
const priorityType = (value: string) => {
    if (['0', '1', '2', '3'].includes(value)) return 'danger';
    if (value === '4') return 'warning';
    return 'info';
};

onMounted(() => {
    const cachedPageSize = Number(localStorage.getItem(paginationConfig.cacheSizeKey));
    if (pageSizes.includes(cachedPageSize)) {
        paginationConfig.pageSize = cachedPageSize;
    }
    resetAndLoadLogs();
    loadServices();
});
onUnmounted(() => {
    if (timer) clearTimeout(timer);
    if (filterTimer) clearTimeout(filterTimer);
});

watch([keyword, priority, service], () => {
    requestID++;
    if (filterTimer) clearTimeout(filterTimer);
    filterTimer = setTimeout(resetAndLoadLogs, 300);
});
</script>

<style scoped lang="scss">
.log-raw {
    overflow: auto;
    padding: 12px;
    color: #f5f5f5;
    font-family: 'SFMono-Regular', Consolas, 'Liberation Mono', monospace;
    font-size: 13px;
    line-height: 1.6;
    white-space: pre-wrap;
    word-break: break-all;
    background: var(--panel-logs-bg-color);
    border: 1px solid var(--el-border-color-darker);
    border-radius: 4px;
}
</style>
