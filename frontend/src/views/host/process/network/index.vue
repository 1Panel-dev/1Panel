<template>
    <div>
        <FireRouter />
        <LayoutContent :title="$t('menu.network', 2)" v-loading="loading">
            <template #rightToolBar>
                <div class="w-full flex justify-end items-center gap-5">
                    <el-select
                        v-model="filters"
                        :placeholder="$t('commons.table.status')"
                        clearable
                        multiple
                        collapse-tags
                        collapse-tags-tooltip
                        :max-collapse-tags="2"
                        @change="search()"
                        class="p-w-300"
                    >
                        <el-option
                            v-for="item in statusOptions"
                            :key="item.value"
                            :label="item.text"
                            :value="item.value"
                        />
                    </el-select>
                    <TableSearch
                        @search="search()"
                        :placeholder="$t('process.pid')"
                        v-model:searchName="netSearch.processID"
                    />
                    <TableSearch
                        @search="search()"
                        :placeholder="$t('process.processName')"
                        v-model:searchName="netSearch.processName"
                    />
                    <TableSearch
                        @search="search()"
                        :placeholder="$t('commons.table.port')"
                        v-model:searchName="netSearch.port"
                    />
                </div>
            </template>

            <template #main>
                <div class="!h-[900px]">
                    <el-auto-resizer>
                        <template #default="{ height, width }">
                            <el-table-v2
                                :columns="columns"
                                :data="data"
                                :width="width"
                                :height="height"
                                :sort-by="sortState"
                                @column-sort="changeSort"
                            />
                        </template>
                    </el-auto-resizer>
                </div>
            </template>
        </LayoutContent>
    </div>
</template>

<script setup lang="ts">
import FireRouter from '@/views/host/process/index.vue';
import { ref, onMounted, onUnmounted, reactive, watch } from 'vue';
import { GlobalStore } from '@/store';
import { SortBy, TableV2SortOrder } from 'element-plus';
import i18n from '@/lang';

const statusOptions = [
    { text: 'LISTEN', value: 'LISTEN' },
    { text: 'ESTABLISHED', value: 'ESTABLISHED' },
    { text: 'TIME_WAIT', value: 'TIME_WAIT' },
    { text: 'CLOSE_WAIT', value: 'CLOSE_WAIT' },
    { text: 'NONE', value: 'NONE' },
];

const globalStore = GlobalStore();

const netSearch = reactive({
    type: 'net',
    processID: undefined,
    processName: '',
    port: undefined,
});

let processSocket = ref(null) as unknown as WebSocket;
const data = ref<any[]>([]);
const oldData = ref<any[]>([]);
const loading = ref(false);

const sortState = ref<SortBy>({
    key: 'PID',
    order: TableV2SortOrder.ASC,
});
const filters = ref<string[]>([]);

const sortByNum = (a: any, b: any, prop: string): number => {
    const aVal = parseFloat(a[prop]) || 0;
    const bVal = parseFloat(b[prop]) || 0;
    return aVal - bVal;
};

const columns = ref([
    {
        key: 'type',
        title: i18n.global.t('commons.table.type'),
        dataKey: 'type',
        width: 220,
    },
    {
        key: 'PID',
        title: 'PID',
        dataKey: 'PID',
        width: 220,
        sortable: true,
        sortMethod: sortByNum,
    },
    {
        key: 'name',
        title: i18n.global.t('process.processName'),
        dataKey: 'name',
        width: 300,
    },
    {
        key: 'localaddr',
        title: i18n.global.t('process.laddr'),
        dataKey: 'localaddr',
        width: 350,
        cellRenderer: ({ rowData }) => {
            const addr = rowData.localaddr;
            return addr?.ip ? `${addr.ip}${addr.port > 0 ? ':' + addr.port : ''}` : '';
        },
    },
    {
        key: 'remoteaddr',
        title: i18n.global.t('process.raddr'),
        dataKey: 'remoteaddr',
        width: 350,
        cellRenderer: ({ rowData }) => {
            const addr = rowData.remoteaddr;
            return addr?.ip ? `${addr.ip}${addr.port > 0 ? ':' + addr.port : ''}` : '';
        },
    },
    {
        key: 'status',
        title: i18n.global.t('commons.table.status'),
        dataKey: 'status',
        width: 380,
        cellRenderer: ({ rowData }) => rowData.status,
    },
]);

watch(
    [sortState, oldData, filters],
    ([newState, newData, newFilters]) => {
        if (!newData?.length) return;

        let filtered = newData;
        if (newFilters.length > 0) {
            filtered = filtered.filter((row) => newFilters.includes(row.status));
        }

        const { key, order } = newState ?? {};
        if (!key || !order) {
            data.value = filtered;
            return;
        }

        const currCol = columns.value.find((c) => c.key === key);
        if (!currCol) {
            data.value = filtered;
            return;
        }

        const sortMethod = currCol.sortMethod ?? sortByNum;
        data.value = filtered.slice().sort((a, b) => {
            const res = (sortMethod as any)(a, b, currCol.dataKey);
            return order === TableV2SortOrder.ASC ? res : -res;
        });
    },
    { immediate: true },
);

const changeSort = ({ key, order }) => {
    if (!order) order = TableV2SortOrder.ASC;
    sortState.value = { key, order };
};

const filterByStatus = () => {
    if (filters.value.length > 0) {
        return oldData.value.filter((row) => filters.value.includes(row.status));
    }
    return oldData.value;
};

const isWsOpen = () => processSocket && processSocket.readyState === 1;
const closeSocket = () => {
    if (isWsOpen()) processSocket.close();
};

const onOpenProcess = () => {
    loading.value = true;
    processSocket.send(JSON.stringify(netSearch));
};
const onMessage = (message: any) => {
    oldData.value = JSON.parse(message.data);
    data.value = filterByStatus();
    loading.value = false;
};
const onerror = () => {};
const onClose = () => {};

const initProcess = () => {
    let href = window.location.href;
    let protocol = href.split('//')[0] === 'http:' ? 'ws' : 'wss';
    let ipLocal = href.split('//')[1].split('/')[0];
    let currentNode = globalStore.currentNode;
    processSocket = new WebSocket(`${protocol}://${ipLocal}/api/v2/process/ws?operateNode=${currentNode}`);
    processSocket.onopen = onOpenProcess;
    processSocket.onmessage = onMessage;
    processSocket.onerror = onerror;
    processSocket.onclose = onClose;

    search();
    sendMsg();
};

const sendMsg = () => {
    setInterval(() => {
        search();
    }, 3000);
};

const search = () => {
    if (isWsOpen()) {
        if (typeof netSearch.processID === 'string') {
            netSearch.processID = Number(netSearch.processID);
        }
        if (typeof netSearch.port === 'string') {
            netSearch.port = Number(netSearch.port);
        }
        processSocket.send(JSON.stringify(netSearch));
    }
};

onMounted(() => {
    initProcess();
});
onUnmounted(() => {
    closeSocket();
});
</script>
