<template>
    <div>
        <Submenu activeName="operation" />
        <el-card style="margin-top: 20px">
            <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search">
                <el-table-column :label="$t('logs.operatoin')" fix>
                    <template #default="{ row }">
                        {{ fmtOperation(row) }}
                    </template>
                </el-table-column>
                <el-table-column :label="$t('logs.status')" prop="status">
                    <template #default="{ row }">
                        <el-tag v-if="row.status == '200'" class="ml-2" type="success">{{ row.status }}</el-tag>
                        <div v-else>
                            <el-popover
                                placement="top-start"
                                :title="$t('commons.table.message')"
                                :width="400"
                                trigger="hover"
                                :content="row.errorMessage"
                            >
                                <template #reference>
                                    <el-tag class="ml-2" type="warning">{{ row.status }}</el-tag>
                                </template>
                            </el-popover>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column label="IP" prop="ip" />
                <el-table-column :label="$t('logs.request')" prop="path">
                    <template #default="{ row }">
                        <div>
                            <el-popover :width="500" v-if="row.body" placement="left-start" trigger="click">
                                <div style="word-wrap: break-word; font-size: 12px; white-space: normal">
                                    <pre class="pre">{{ fmtBody(row.body) }}</pre>
                                </div>
                                <template #reference>
                                    <el-icon style="cursor: pointer"><warning /></el-icon>
                                </template>
                            </el-popover>
                            <span v-else>-</span>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('logs.response')" prop="path">
                    <template #default="{ row }">
                        <div>
                            <el-popover :width="500" v-if="row.resp" placement="left-start" trigger="click">
                                <div style="word-wrap: break-word; font-size: 12px; white-space: normal">
                                    <pre class="pre">{{ fmtBody(row.resp) }}</pre>
                                </div>
                                <template #reference>
                                    <el-icon style="cursor: pointer"><warning /></el-icon>
                                </template>
                            </el-popover>
                            <span v-else>-</span>
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
import { getOperationLogs } from '@/api/modules/log';
import Submenu from '@/views/log/index.vue';
import { onMounted, reactive, ref } from '@vue/runtime-core';
import { Log } from '@/api/interface/log';
import i18n from '@/lang';

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
    const res = await getOperationLogs(params);
    data.value = res.data.items;
    paginationConfig.total = res.data.total;
};

const fmtOperation = (row: Log.OperationLog) => {
    if (row.method.toLocaleLowerCase() === 'post') {
        if (row.source == '' && row.action == '') {
            return (
                i18n.global.t('logs.detail.' + row.group.toLocaleLowerCase()) +
                i18n.global.t('logs.detail.' + row.method.toLocaleLowerCase())
            );
        }
        if (row.action == '') {
            return (
                i18n.global.t('logs.detail.' + row.group.toLocaleLowerCase()) +
                i18n.global.t('logs.detail.' + row.source.toLocaleLowerCase())
            );
        }
        return;
    }
    if (row.action == '') {
        return (
            i18n.global.t('logs.detail.' + row.group.toLocaleLowerCase()) +
            i18n.global.t('logs.detail.' + row.method.toLocaleLowerCase())
        );
    } else {
        return (
            i18n.global.t('logs.detail.' + row.group.toLocaleLowerCase()) +
            i18n.global.t('logs.detail.' + row.source.toLocaleLowerCase())
        );
    }
};

const fmtBody = (value: string) => {
    try {
        return JSON.parse(value);
    } catch (err) {
        return value;
    }
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
