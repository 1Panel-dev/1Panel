<template>
    <DrawerPro v-model="open" size="large" :header="$t('menu.msgCenter')" @close="handleClose">
        <template #content>
            <LayoutContent v-loading="loading" :title="$t('logs.task')">
                <template #leftToolBar>
                    <NodeSelect v-model="targeNode" @change="search()" />
                </template>
                <template #rightToolBar>
                    <el-select
                        v-model="req.status"
                        :placeholder="$t('commons.table.all')"
                        @change="search()"
                        clearable
                        class="p-w-200"
                    >
                        <template #prefix>{{ $t('commons.table.status') }}</template>
                        <el-option :label="$t('commons.table.all')" value=""></el-option>
                        <el-option :label="$t('commons.status.success')" value="Success"></el-option>
                        <el-option :label="$t('commons.status.failed')" value="Failed"></el-option>
                        <el-option :label="$t('logs.taskRunning')" value="Executing"></el-option>
                    </el-select>
                    <TableRefresh @search="search()" />
                    <TableSetting title="task-refresh" @search="!loading && search()" :rate="5" />
                </template>
                <template #main>
                    <ComplexTable
                        :cache-pagination="false"
                        :pagination-config="paginationConfig"
                        :data="data"
                        @search="search()"
                        :heightDiff="320"
                    >
                        <el-table-column :label="$t('logs.taskName')" prop="name" min-width="180px">
                            <template #default="{ row }">
                                {{ translateTaskText(row.name) }}
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('commons.table.status')" prop="status" max-width="80px">
                            <template #default="{ row }">
                                <Status :status="row.status" :msg="row.errorMsg" />
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('commons.button.log')" prop="log" max-width="80px">
                            <template #default="{ row }">
                                <el-button @click="openTaskLog(row)" link type="primary">
                                    {{ $t('website.check') }}
                                </el-button>
                            </template>
                        </el-table-column>
                        <el-table-column
                            width="180px"
                            prop="createdAt"
                            :label="$t('commons.table.date')"
                            :formatter="dateFormat"
                            show-overflow-tooltip
                        />
                    </ComplexTable>
                </template>
            </LayoutContent>

            <TaskLog ref="taskLogRef" width="70%" />
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { readCachedPageSize } from '@/utils/panel-cache';
import TaskLog from '@/components/log/task/index.vue';
import NodeSelect from '@/components/node-select/index.vue';

import { dateFormat } from '@/utils/date';
import { searchTasks } from '@/api/modules/log';
import { reactive, ref } from 'vue';
import { Log } from '@/api/interface/log';
import bus from '@/global/bus';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { translateTaskText } from '@/utils/task';
const { currentNode } = useGlobalStore();

const open = ref(false);
const handleClose = () => {
    open.value = false;
};
const loading = ref();
const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'task-list-page-size',
    currentPage: 1,
    pageSize: readCachedPageSize('task-list-page-size') || 20,
    total: 0,
    small: true,
});
const taskLogRef = ref();
const req = reactive({
    type: '',
    status: 'Executing',
    page: 1,
    pageSize: 10,
});
const targeNode = ref('local');

let latestSearch = 0;
const search = async (fallbackToAll = false) => {
    const searchID = ++latestSearch;
    bus.emit('refreshTask', true);
    req.page = paginationConfig.currentPage;
    req.pageSize = paginationConfig.pageSize;
    loading.value = true;
    try {
        let res = await searchTasks({ ...req }, targeNode.value);
        if (fallbackToAll && res.data.total === 0 && searchID === latestSearch) {
            req.status = '';
            res = await searchTasks({ ...req }, targeNode.value);
        }
        if (searchID !== latestSearch) return;
        data.value = res.data.items;
        paginationConfig.total = res.data.total;
    } catch (error) {
    } finally {
        if (searchID === latestSearch) loading.value = false;
    }
};

const openTaskLog = (row: Log.Task) => {
    taskLogRef.value.openWithTaskID(row.id, row.status == 'Executing', targeNode.value);
};

const acceptParams = () => {
    targeNode.value = currentNode.value;
    req.status = 'Executing';
    paginationConfig.currentPage = 1;
    search(true);
    open.value = true;
};

defineExpose({
    acceptParams,
});
</script>
