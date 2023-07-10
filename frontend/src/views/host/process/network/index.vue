<template>
    <div>
        <FireRouter />
        <LayoutContent :title="$t('menu.network')" v-loading="loading">
            <template #toolbar>
                <el-row>
                    <el-col :span="24">
                        <div style="width: 100%">
                            <el-form-item style="float: right">
                                <el-row :gutter="20">
                                    <el-col :span="8">
                                        <div class="search-button">
                                            <el-input
                                                typpe="number"
                                                v-model.number="netSearch.processID"
                                                clearable
                                                @clear="search()"
                                                suffix-icon="Search"
                                                @keyup.enter="search()"
                                                @change="search()"
                                                :placeholder="$t('process.pid')"
                                            ></el-input>
                                        </div>
                                    </el-col>
                                    <el-col :span="8">
                                        <div class="search-button">
                                            <el-input
                                                v-model.trim="netSearch.processName"
                                                clearable
                                                @clear="search()"
                                                suffix-icon="Search"
                                                @keyup.enter="search()"
                                                @change="search()"
                                                :placeholder="$t('process.processName')"
                                            ></el-input>
                                        </div>
                                    </el-col>
                                    <el-col :span="8">
                                        <div class="search-button">
                                            <el-input
                                                type="number"
                                                v-model.number="netSearch.port"
                                                clearable
                                                @clear="search()"
                                                suffix-icon="Search"
                                                @keyup.enter="search()"
                                                @change="search()"
                                                :placeholder="$t('commons.table.port')"
                                            ></el-input>
                                        </div>
                                    </el-col>
                                </el-row>
                            </el-form-item>
                        </div>
                    </el-col>
                </el-row>
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
                        :label="$t('app.status')"
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
            netSearch.processID = undefined;
        }
        if (typeof netSearch.port === 'string') {
            netSearch.port = undefined;
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
