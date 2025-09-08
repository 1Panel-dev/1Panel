<template>
    <DrawerPro
        v-model="open"
        size="large"
        :header="$t('menu.msgCenter')"
        :resource="globalStore.currentNode"
        @close="handleClose"
    >
        <template #content>
            <LayoutContent v-loading="loading" :title="$t('logs.task')">
                <template #rightToolBar>
                    <el-select v-model="req.status" @change="search()" clearable class="p-w-200">
                        <template #prefix>{{ $t('commons.table.status') }}</template>
                        <el-option :label="$t('commons.table.all')" value=""></el-option>
                        <el-option :label="$t('commons.status.success')" value="Success"></el-option>
                        <el-option :label="$t('commons.status.failed')" value="Failed"></el-option>
                        <el-option :label="$t('logs.taskRunning')" value="Executing"></el-option>
                    </el-select>
                    <TableRefresh @search="search()" />
                    <TableSetting title="task-refresh" @search="search()" :rate="5" />
                </template>
                <template #main>
                    <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search" :heightDiff="320">
                        <el-table-column :label="$t('logs.taskName')" prop="name" min-width="180px"></el-table-column>
                        <el-table-column :label="$t('commons.table.status')" prop="status" max-width="100px">
                            <template #default="{ row }">
                                <Status :status="row.status" :msg="row.errorMsg" />
                            </template>
                        </el-table-column>
                        <el-table-column :label="$t('commons.button.log')" prop="log" max-width="100px">
                            <template #default="{ row }">
                                <el-button @click="openTaskLog(row)" link type="primary">
                                    {{ $t('website.check') }}
                                </el-button>
                            </template>
                        </el-table-column>
                        <el-table-column
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
import { dateFormat } from '@/utils/util';
import { searchTasks } from '@/api/modules/log';
import { reactive, ref } from 'vue';
import { Log } from '@/api/interface/log';
import TaskLog from '@/components/log/task/index.vue';
import bus from '@/global/bus';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const open = ref(false);
const handleClose = () => {
    open.value = false;
};
const loading = ref();
const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'task-list-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('task-list-page-size')) || 20,
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

const search = async () => {
    bus.emit('refreshTask', true);
    req.page = paginationConfig.currentPage;
    req.pageSize = paginationConfig.pageSize;
    loading.value = true;
    try {
        const res = await searchTasks(req);
        loading.value = false;
        data.value = res.data.items;
        paginationConfig.total = res.data.total;
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

const openTaskLog = (row: Log.Task) => {
    taskLogRef.value.openWithTaskID(row.id, row.status == 'Executing');
};

const acceptParams = () => {
    search();
    open.value = true;
};

defineExpose({
    acceptParams,
});
</script>
