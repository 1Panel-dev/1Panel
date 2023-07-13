<template>
    <div>
        <FireRouter />
        <LayoutContent :title="$t('menu.process')" v-loading="loading">
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
                                                v-model.number="processSearch.pid"
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
                                                v-model.trim="processSearch.name"
                                                clearable
                                                @clear="search()"
                                                suffix-icon="Search"
                                                @keyup.enter="search()"
                                                @change="search()"
                                                :placeholder="$t('commons.table.name')"
                                            ></el-input>
                                        </div>
                                    </el-col>
                                    <el-col :span="8">
                                        <div class="search-button">
                                            <el-input
                                                v-model.trim="processSearch.username"
                                                clearable
                                                @clear="search()"
                                                suffix-icon="Search"
                                                @keyup.enter="search()"
                                                @change="search()"
                                                :placeholder="$t('commons.table.user')"
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
                    <el-table-column :label="'PID'" fix prop="PID" max-width="60px" sortable></el-table-column>
                    <el-table-column
                        :label="$t('commons.table.name')"
                        fix
                        prop="name"
                        min-width="120px"
                    ></el-table-column>
                    <el-table-column
                        :label="$t('process.ppid')"
                        min-width="120px"
                        fix
                        prop="PPID"
                        sortable
                    ></el-table-column>
                    <el-table-column :label="$t('process.numThreads')" fix prop="numThreads"></el-table-column>
                    <el-table-column :label="$t('commons.table.user')" fix prop="username"></el-table-column>
                    <el-table-column
                        :label="'CPU'"
                        fix
                        prop="cpuValue"
                        :formatter="cpuFormatter"
                        sortable
                    ></el-table-column>
                    <el-table-column
                        :label="$t('process.memory')"
                        fix
                        prop="rssValue"
                        :formatter="memFormatter"
                        sortable
                    ></el-table-column>
                    <el-table-column :label="$t('process.numConnections')" fix prop="numConnections"></el-table-column>
                    <el-table-column
                        :label="$t('process.status')"
                        fix
                        prop="status"
                        column-key="status"
                        :filters="[
                            { text: $t('process.running'), value: 'running' },
                            { text: $t('process.sleep'), value: 'sleep' },
                            { text: $t('process.stop'), value: 'stop' },
                            { text: $t('process.idle'), value: 'idle' },
                            { text: $t('process.wait'), value: 'wait' },
                            { text: $t('process.lock'), value: 'lock' },
                            { text: $t('process.zombie'), value: 'zombie' },
                        ]"
                        :filter-method="filterStatus"
                        :filtered-value="sortConfig.filters"
                    >
                        <template #default="{ row }">
                            <span v-if="row.status">{{ $t('process.' + row.status) }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('process.startTime')"
                        fix
                        prop="startTime"
                        min-width="140px"
                    ></el-table-column>
                    <fu-table-operations :ellipsis="10" :buttons="buttons" :label="$t('commons.table.operate')" fix />
                </ComplexTable>
            </template>
        </LayoutContent>
        <ProcessDetail ref="detailRef" />
    </div>
</template>

<script setup lang="ts">
import FireRouter from '@/views/host/process/index.vue';
import { ref, onMounted, onUnmounted, nextTick, reactive } from 'vue';
import ProcessDetail from './detail/index.vue';
import i18n from '@/lang';
import { StopProcess } from '@/api/modules/process';
import { useDeleteData } from '@/hooks/use-delete-data';

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

const processSearch = reactive({
    type: 'ps',
    pid: undefined,
    username: '',
    name: '',
});

const buttons = [
    {
        label: i18n.global.t('app.detail'),
        click: function (row: any) {
            openDetail(row);
        },
    },
    {
        label: i18n.global.t('process.stopProcess'),
        click: function (row: any) {
            stopProcess(row.PID);
        },
    },
];

let processSocket = ref(null) as unknown as WebSocket;
const data = ref([]);
const loading = ref(false);
const tableRef = ref();
const oldData = ref([]);
const detailRef = ref();
const isGetData = ref(true);

const openDetail = (row: any) => {
    detailRef.value.acceptParams({ info: row });
};

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

const filterStatus = (value: string, row: any) => {
    return row.status === value;
};

const cpuFormatter = (row: any) => {
    return row.cpuPercent;
};

const memFormatter = (row: any) => {
    return row.rss;
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
            processSearch.pid = undefined;
        }
        processSocket.send(JSON.stringify(processSearch));
    }
};

const stopProcess = async (PID: number) => {
    try {
        await useDeleteData(StopProcess, { PID: PID }, i18n.global.t('process.stopProcessWarn', [PID]));
    } catch (error) {}
};

onMounted(() => {
    initProcess();
});

onUnmounted(() => {
    closeSocket();
});
</script>
