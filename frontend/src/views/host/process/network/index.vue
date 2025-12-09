<template>
    <div>
        <FireRouter />
        <LayoutContent :title="$t('menu.network', 2)" v-loading="processStore.netLoading">
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
                        v-model:searchName="processStore.netSearch.processID"
                    />
                    <TableSearch
                        @search="search()"
                        :placeholder="$t('process.processName')"
                        v-model:searchName="processStore.netSearch.processName"
                    />
                    <TableSearch
                        @search="search()"
                        :placeholder="$t('commons.table.port')"
                        v-model:searchName="processStore.netSearch.port"
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
import { ref, onMounted, onUnmounted, watch, h } from 'vue';
import { GlobalStore, ProcessStore } from '@/store';
import { SortBy, TableV2SortOrder, ElIcon } from 'element-plus';
import { Filter } from '@element-plus/icons-vue';
import i18n from '@/lang';

const statusOptions = [
    { text: 'LISTEN', value: 'LISTEN' },
    { text: 'ESTABLISHED', value: 'ESTABLISHED' },
    { text: 'TIME_WAIT', value: 'TIME_WAIT' },
    { text: 'CLOSE_WAIT', value: 'CLOSE_WAIT' },
    { text: 'NONE', value: 'NONE' },
];

const globalStore = GlobalStore();
const processStore = ProcessStore();

const quickSearchName = (name: string) => {
    processStore.netSearch.processID = undefined;
    processStore.netSearch.processName = name;
    processStore.netSearch.port = undefined;
    search();
};

const quickSearchPort = (port: number) => {
    processStore.netSearch.processID = undefined;
    processStore.netSearch.processName = '';
    processStore.netSearch.port = port;
    search();
};

const data = ref<any[]>([]);

const sortState = ref<SortBy>({
    key: 'PID',
    order: TableV2SortOrder.ASC,
});
const filters = ref<string[]>(['LISTEN', 'ESTABLISHED']);

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
        cellRenderer: ({ rowData }) => {
            return h('div', { class: 'flex items-center gap-1' }, [
                h('span', { class: 'truncate', title: rowData.name }, rowData.name),
                h(
                    ElIcon,
                    {
                        class: 'cursor-pointer hover:text-primary ml-1 flex-shrink-0',
                        size: 14,
                        onClick: (e: Event) => {
                            e.stopPropagation();
                            quickSearchName(rowData.name);
                        },
                    },
                    () => h(Filter),
                ),
            ]);
        },
    },
    {
        key: 'localaddr',
        title: i18n.global.t('process.laddr'),
        dataKey: 'localaddr',
        width: 350,
        cellRenderer: ({ rowData }) => {
            const addr = rowData.localaddr;
            const addrStr = addr?.ip ? `${addr.ip}${addr.port > 0 ? ':' + addr.port : ''}` : '';
            const hasPort = addr?.port > 0;
            return h('div', { class: 'flex items-center gap-1' }, [
                h('span', {}, addrStr),
                hasPort
                    ? h(
                          ElIcon,
                          {
                              class: 'cursor-pointer hover:text-primary ml-1',
                              size: 12,
                              onClick: (e: Event) => {
                                  e.stopPropagation();
                                  quickSearchPort(addr.port);
                              },
                          },
                          () => h(Filter),
                      )
                    : null,
            ]);
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
    [sortState, () => processStore.netData, filters],
    ([newState, newData, newFilters]) => {
        if (!newData?.length) {
            data.value = [];
            return;
        }

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

const search = () => {
    processStore.sendNetMessage();
};

onMounted(() => {
    processStore.connect(globalStore.currentNode);
    const initialDelay = processStore.netData.length > 0 ? 500 : 0;
    processStore.startPolling('net', 3000, initialDelay);
});

onUnmounted(() => {
    processStore.stopPolling();
    processStore.disconnect();
});
</script>
