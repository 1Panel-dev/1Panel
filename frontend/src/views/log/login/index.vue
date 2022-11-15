<template>
    <div>
        <Submenu activeName="login" />
        <el-card style="margin-top: 20px">
            <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search">
                <el-table-column min-width="40" :label="$t('logs.loginIP')" prop="ip" />
                <el-table-column min-width="40" :label="$t('logs.loginAddress')" prop="address" />
                <el-table-column :label="$t('logs.loginAgent')" show-overflow-tooltip prop="agent" />
                <el-table-column min-width="40" :label="$t('logs.loginStatus')" prop="status">
                    <template #default="{ row }">
                        <div v-if="row.status === 'Success'">
                            <el-tag type="success">{{ row.status }}</el-tag>
                        </div>
                        <div v-else>
                            <el-tooltip class="box-item" effect="dark" :content="row.message" placement="top-start">
                                <el-tag type="danger">{{ row.status }}</el-tag>
                            </el-tooltip>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column
                    prop="createdAt"
                    :label="$t('commons.table.date')"
                    :formatter="dateFromat"
                    show-overflow-tooltip
                />
            </ComplexTable>
        </el-card>
    </div>
</template>

<script setup lang="ts">
import ComplexTable from '@/components/complex-table/index.vue';
import { dateFromat } from '@/utils/util';
import { getLoginLogs } from '@/api/modules/log';
import Submenu from '@/views/log/index.vue';
import { onMounted, reactive, ref } from '@vue/runtime-core';

const data = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 15,
    total: 0,
});

const search = async () => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    const res = await getLoginLogs(params);
    data.value = res.data.items;
    paginationConfig.total = res.data.total;
};

onMounted(() => {
    search();
});
</script>

<style scoped lang="scss">
.pre {
    white-space: pre-wrap;
    white-space: -moz-pre-wrap;
    white-space: -pre-wrap;
    white-space: -o-pre-wrap;
    word-wrap: break-word;
}
</style>
