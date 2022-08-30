<template>
    <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search">
        <el-table-column :label="$t('operations.operatoin')" fix>
            <template #default="{ row }">
                {{ fmtOperation(row) }}
            </template>
        </el-table-column>
        <el-table-column :label="$t('operations.status')" prop="status">
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
        <el-table-column align="left" :label="$t('operations.request')" prop="path">
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
                    <span v-else>无</span>
                </div>
            </template>
        </el-table-column>
        <el-table-column align="left" :label="$t('operations.response')" prop="path">
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
                    <span v-else>无</span>
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
</template>

<script setup lang="ts">
import ComplexTable from '@/components/complex-table/index.vue';
import { dateFromat } from '@/utils/util';
import { getOperationList } from '@/api/modules/operation-log';
import { onMounted, reactive, ref } from '@vue/runtime-core';
import { ResOperationLog } from '@/api/interface/operation-log';
import i18n from '@/lang';

const data = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 5,
    total: 0,
});

const logSearch = reactive({
    page: 1,
    pageSize: 5,
});

const search = async () => {
    logSearch.page = paginationConfig.currentPage;
    logSearch.pageSize = paginationConfig.pageSize;
    const res = await getOperationList(logSearch);
    data.value = res.data.items;
    paginationConfig.total = res.data.total;
};

const fmtOperation = (row: ResOperationLog) => {
    if (row.method.toLocaleLowerCase() !== 'put') {
        if (row.source == '' && row.action == '') {
            return (
                i18n.global.t('operations.detail.' + row.group.toLocaleLowerCase()) +
                i18n.global.t('operations.detail.' + row.method.toLocaleLowerCase())
            );
        }
        if (row.action == '') {
            return (
                i18n.global.t('operations.detail.' + row.group.toLocaleLowerCase()) +
                i18n.global.t('operations.detail.' + row.source.toLocaleLowerCase())
            );
        }
        return;
    }
    if (row.action == '') {
        return (
            i18n.global.t('operations.detail.' + row.group.toLocaleLowerCase()) +
            i18n.global.t('operations.detail.' + row.method.toLocaleLowerCase())
        );
    } else {
        return (
            i18n.global.t('operations.detail.' + row.group.toLocaleLowerCase()) +
            i18n.global.t('operations.detail.' + row.source.toLocaleLowerCase())
        );
    }
    return '';
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
