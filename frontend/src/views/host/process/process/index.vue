<template>
    <div>
        <FireRouter />
        <LayoutContent :title="$t('menu.process', 2)" v-loading="loading">
            <template #rightToolBar>
                <div class="w-full flex justify-end items-center gap-5">
                    <el-select
                        v-model="filters"
                        :placeholder="$t('commons.table.status')"
                        clearable
                        @change="search()"
                        class="p-w-400"
                        multiple
                        collapse-tags
                        collapse-tags-tooltip
                        :max-collapse-tags="4"
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
                        v-model:searchName="processSearch.pid"
                    />
                    <TableSearch
                        @search="search()"
                        :placeholder="$t('commons.table.name')"
                        v-model:searchName="processSearch.name"
                    />
                    <TableSearch
                        @search="search()"
                        :placeholder="$t('commons.table.user')"
                        v-model:searchName="processSearch.username"
                    />
                </div>
            </template>
            <template #main>
                <div class="!h-[900px]">
                    <el-auto-resizer>
                        <template #default="{ height, width }">
                            <el-table-v2
                                @column-sort="changeSort"
                                :columns="columns"
                                :data="data"
                                :width="width"
                                :height="height"
                                :sort-by="sortState"
                            ></el-table-v2>
                        </template>
                    </el-auto-resizer>
                </div>
            </template>
        </LayoutContent>

        <OpDialog ref="opRef" @search="search" />
        <ProcessDetail ref="detailRef" />
    </div>
</template>

<script setup lang="ts">
import FireRouter from '@/views/host/process/index.vue';
import { ref, onMounted, onUnmounted, reactive } from 'vue';
import ProcessDetail from './detail/index.vue';
import i18n from '@/lang';
import { stopProcess } from '@/api/modules/process';
import { GlobalStore } from '@/store';
import { SortBy, TableV2SortOrder, ElButton } from 'element-plus';
const globalStore = GlobalStore();

const statusOptions = computed(() => [
    { text: i18n.global.t('process.running'), value: 'running' },
    { text: i18n.global.t('process.sleep'), value: 'sleep' },
    { text: i18n.global.t('process.stop'), value: 'stop' },
    { text: i18n.global.t('process.idle'), value: 'idle' },
    { text: i18n.global.t('process.wait'), value: 'wait' },
    { text: i18n.global.t('process.lock'), value: 'lock' },
    { text: i18n.global.t('process.zombie'), value: 'zombie' },
]);

const processSearch = reactive({
    type: 'ps',
    pid: undefined,
    username: '',
    name: '',
});
const opRef = ref();
const sortState = ref<SortBy>({
    key: 'PID',
    order: TableV2SortOrder.ASC,
});

let processSocket = ref(null) as unknown as WebSocket;
const data = ref([]);
const loading = ref(false);
const oldData = ref([]);
const detailRef = ref();
const isGetData = ref(true);
const filters = ref([]);

const sortByNum = (a: any, b: any, prop: string): number => {
    const aVal = parseFloat(a[prop]) || 0;
    const bVal = parseFloat(b[prop]) || 0;
    return aVal - bVal;
};

const columns = ref([
    {
        key: 'PID',
        title: 'PID',
        dataKey: 'PID',
        width: 120,
    },
    {
        key: 'name',
        title: i18n.global.t('commons.table.name'),
        dataKey: 'name',
        width: 400,
    },
    {
        key: 'ppid',
        title: i18n.global.t('process.ppid'),
        dataKey: 'PPID',
        width: 120,
        sortable: true,
    },
    {
        key: 'numThreads',
        title: i18n.global.t('process.numThreads'),
        dataKey: 'numThreads',
        width: 120,
    },
    {
        key: 'username',
        title: i18n.global.t('commons.table.user'),
        dataKey: 'username',
        width: 200,
    },
    {
        key: 'cpuValue',
        title: 'CPU',
        dataKey: 'cpuValue',
        width: 200,
        sortable: true,
        sortMethod: sortByNum,
        cellRenderer: ({ rowData }) => {
            return rowData.cpuPercent;
        },
    },
    {
        key: 'rssValue',
        title: i18n.global.t('process.memory'),
        dataKey: 'rssValue',
        width: 200,
        sortable: true,
        sortMethod: sortByNum,
        cellRenderer: ({ rowData }) => {
            return rowData.rss;
        },
    },
    {
        key: 'numConnections',
        title: i18n.global.t('process.numConnections'),
        dataKey: 'numConnections',
        width: 100,
    },
    {
        key: 'status',
        title: i18n.global.t('commons.table.status'),
        dataKey: 'status',
        width: 100,
        cellRenderer: ({ rowData }) => {
            if (rowData.status) {
                return i18n.global.t('process.' + rowData.status);
            }
            return '';
        },
    },
    {
        key: 'startTime',
        title: i18n.global.t('process.startTime'),
        dataKey: 'startTime',
        width: 300,
    },
    {
        key: 'actions',
        title: i18n.global.t('commons.table.operate'),
        dataKey: 'actions',
        width: 200,
        cellRenderer: ({ rowData }) => {
            return h('div', { class: 'action-buttons' }, [
                h(
                    ElButton,
                    {
                        type: 'text',
                        onClick: () => openDetail(rowData),
                    },
                    () => i18n.global.t('process.viewDetails'),
                ),
                h(
                    ElButton,
                    {
                        type: 'text',
                        onClick: () => stop(rowData),
                    },
                    () => i18n.global.t('process.stopProcess'),
                ),
            ]);
        },
    },
]);

watch(
    [sortState, oldData],
    ([newState, newData]) => {
        if (!newData?.length) return;

        const { key, order } = newState ?? {};
        if (!key || !order) {
            data.value = filterByStatus();
            return;
        }

        const currCol = columns.value.find((c) => c.key === key);
        if (!currCol) return;

        const currSortMethod = currCol.sortMethod ?? sortByNum;
        const filteredData = filterByStatus();

        data.value = filteredData.slice(0).sort((a, b) => {
            const res = (currSortMethod as any)(a, b, currCol.dataKey);
            return order === TableV2SortOrder.ASC ? res : 0 - res;
        });
    },
    { immediate: true },
);

const openDetail = (row: any) => {
    detailRef.value.acceptParams(row.PID);
};

const changeSort = ({ key, order }) => {
    if (!order) order = TableV2SortOrder.ASC;
    sortState.value = { key, order };
};

const isWsOpen = () => {
    const readyState = processSocket && processSocket.readyState;
    return readyState === 1;
};
const closeSocket = () => {
    if (isWsOpen()) {
        processSocket && processSocket.close();
    }
};

const onOpenProcess = () => {
    loading.value = true;
    isGetData.value = true;
    processSocket.send(JSON.stringify(processSearch));
};
const onMessage = (message: any) => {
    isGetData.value = false;
    oldData.value = JSON.parse(message.data);
    data.value = filterByStatus();
    loading.value = false;
};

const filterByStatus = () => {
    if (filters.value.length > 0) {
        const newData = oldData.value.filter((re: any) => {
            return (filters.value as string[]).indexOf(re.status) > -1;
        });
        return newData;
    } else {
        return oldData.value;
    }
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
    sendMsg();
};

const sendMsg = () => {
    setInterval(() => {
        search();
    }, 3000);
};

const search = () => {
    if (isWsOpen() && !isGetData.value) {
        isGetData.value = true;
        if (typeof processSearch.pid === 'string') {
            processSearch.pid = Number(processSearch.pid);
        }
        processSocket.send(JSON.stringify(processSearch));
    }
};

const stop = async (row: any) => {
    opRef.value.acceptParams({
        title: i18n.global.t('process.stopProcess'),
        names: [row.name],
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('menu.process'),
            i18n.global.t('process.stopProcess'),
        ]),
        api: stopProcess,
        params: { PID: row.PID },
        successMsg: i18n.global.t('commons.msg.operationSuccess'),
    });
};

onMounted(() => {
    initProcess();
});

onUnmounted(() => {
    closeSocket();
});
</script>
