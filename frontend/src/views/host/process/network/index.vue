<template>
    <div>
        <FireRouter />
        <LayoutContent :title="$t('menu.network', 2)" v-loading="loading">
            <template #toolbar>
                <div class="flex justify-between gap-2 flex-wrap sm:flex-row">
                    <div><!-- 占位 --></div>
                    <div class="flex flex-wrap gap-3">
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
                </div>
            </template>
            <template #main>
                <ComplexTable :data="data" @sort-change="changeSort" @filter-change="changeFilter" ref="tableRef">
                    <el-table-column :label="$t('commons.table.type')" fix prop="type"></el-table-column>
                    <el-table-column :label="'PID'" fix prop="PID" max-width="60px" sortable></el-table-column>
                    <el-table-column
                        :label="$t('process.processName')"
                        fix
                        prop="name"
                        min-width="120px"
                    ></el-table-column>
                    <el-table-column prop="localaddr" :label="$t('process.laddr')">
                        <template #default="{ row }">
                            <span>{{ row.localaddr.ip }}</span>
                            <span v-if="row.localaddr.port > 0">:{{ row.localaddr.port }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column prop="remoteaddr" :label="$t('process.raddr')">
                        <template #default="{ row }">
                            <span>{{ row.remoteaddr.ip }}</span>
                            <span v-if="row.remoteaddr.port > 0">:{{ row.remoteaddr.port }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column
                        prop="status"
                        column-key="status"
                        :label="$t('process.state')"
                        :filters="[
                            { text: 'LISTEN', value: 'LISTEN' },
                            { text: 'ESTABLISHED', value: 'ESTABLISHED' },
                            { text: 'TIME_WAIT', value: 'TIME_WAIT' },
                            { text: 'CLOSE_WAIT', value: 'CLOSE_WAIT' },
                            { text: 'NONE', value: 'NONE' },
                        ]"
                        :filter-method="filterStatus"
                        :filtered-value="sortConfig.filters"
                    ></el-table-column>
                </ComplexTable>
            </template>
        </LayoutContent>
    </div>
</template>

<script setup lang="ts">
import FireRouter from '@/views/host/process/index.vue';
import { ref, onMounted, onUnmounted, nextTick, reactive } from 'vue';

interface SortStatus {
    prop: '';
    order: '';
    filters: [];
}
const sortConfig: SortStatus = {
    prop: '',
    order: '',
    filters: [],
};

const netSearch = reactive({
    type: 'net',
    processID: undefined,
    processName: '',
    port: undefined,
});

let processSocket = ref(null) as unknown as WebSocket;
const data = ref([]);
const loading = ref(false);
const tableRef = ref();
const oldData = ref([]);

const changeSort = ({ prop, order }) => {
    sortConfig.prop = prop;
    sortConfig.order = order;
};

const changeFilter = (filters: any) => {
    if (filters.status && filters.status.length > 0) {
        sortConfig.filters = filters.status;
        data.value = filterByStatus();
        sortTable();
    } else {
        data.value = oldData.value;
        sortConfig.filters = [];
        sortTable();
    }
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

const filterStatus = (value: string, row: any) => {
    return row.status === value;
};

const onOpenProcess = () => {};
const onMessage = (message: any) => {
    let result: any[] = JSON.parse(message.data);
    oldData.value = result;
    data.value = filterByStatus();
    sortTable();
    loading.value = false;
};

const filterByStatus = () => {
    if (sortConfig.filters.length > 0) {
        const newData = oldData.value.filter((re: any) => {
            return (sortConfig.filters as string[]).indexOf(re.status) > -1;
        });
        return newData;
    } else {
        return oldData.value;
    }
};

const sortTable = () => {
    if (sortConfig.prop != '' && sortConfig.order != '') {
        nextTick(() => {
            tableRef.value?.sort(sortConfig.prop, sortConfig.order);
        });
    }
};

const onerror = () => {};
const onClose = () => {};

const initProcess = () => {
    let href = window.location.href;
    let protocol = href.split('//')[0] === 'http:' ? 'ws' : 'wss';
    let ipLocal = href.split('//')[1].split('/')[0];
    processSocket = new WebSocket(`${protocol}://${ipLocal}/api/v1/process/ws`);
    processSocket.onopen = onOpenProcess;
    processSocket.onmessage = onMessage;
    processSocket.onerror = onerror;
    processSocket.onclose = onClose;
    loading.value = true;
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
