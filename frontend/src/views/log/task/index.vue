<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('logs.task')">
            <template #search>
                <LogRouter current="Task" />
            </template>
            <template #rightToolBar>
                <el-select v-model="req.status" @change="search()" clearable class="p-w-200 mr-2.5">
                    <template #prefix>{{ $t('commons.table.status') }}</template>
                    <el-option :label="$t('commons.table.all')" value=""></el-option>
                    <el-option :label="$t('commons.status.success')" value="Success"></el-option>
                    <el-option :label="$t('commons.status.failed')" value="Failed"></el-option>
                    <el-option :label="$t('logs.taskRunning')" value="Executing"></el-option>
                </el-select>
                <TableRefresh @search="search()" />
                <TableSetting title="task-log-refresh" @search="search()" />
            </template>
            <template #main>
                <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search" :heightDiff="370">
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
    </div>
</template>

<script setup lang="ts">
import LogRouter from '@/views/log/router/index.vue';
import { dateFormat } from '@/utils/util';
import { searchTasks } from '@/api/modules/log';
import { onMounted, reactive, ref } from '@vue/runtime-core';
import { Log } from '@/api/interface/log';
import TaskLog from '@/components/log/task/index.vue';

const loading = ref();
const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'task-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const taskLogRef = ref();
const req = reactive({
    type: '',
    status: '',
    page: 1,
    pageSize: 10,
});

const search = async () => {
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

onMounted(() => {
    search();
});
</script>
