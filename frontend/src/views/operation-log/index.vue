<template>
    <LayoutContent :header="$t('menu.operations')">
        <ComplexTable :pagination-config="paginationConfig" v-model:selects="selects" :data="data" @search="search">
            <el-table-column type="selection" fix />
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
            <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fix />
        </ComplexTable>
    </LayoutContent>
</template>

<script setup lang="ts">
import LayoutContent from '@/layout/layout-content.vue';
import ComplexTable from '@/components/complex-table/index.vue';
import { dateFromat } from '@/utils/util';
import { getOperationList, deleteOperation } from '@/api/modules/operation-log';
import { onMounted, reactive, ref } from '@vue/runtime-core';
import { ResOperationLog } from '@/api/interface/operation-log';
import { useDeleteData } from '@/hooks/use-delete-data';
import i18n from '@/lang';

const data = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 5,
    total: 0,
});

const selects = ref<any>([]);
const batchDelete = async (row: ResOperationLog | null) => {
    let ids: Array<number> = [];

    if (row === null) {
        selects.value.forEach((item: ResOperationLog) => {
            ids.push(item.id);
        });
    } else {
        ids.push(row.id);
    }
    console.log(ids);
    await useDeleteData(deleteOperation, { ids: ids }, 'commons.msg.delete');
    search();
};

const buttons = [
    {
        label: i18n.global.t('commons.button.delete'),
        type: 'danger',
        click: batchDelete,
    },
];

const search = async () => {
    const { currentPage, pageSize } = paginationConfig;
    const res = await getOperationList(currentPage, pageSize);
    data.value = res.data.items;
    paginationConfig.total = res.data.total;
};

const fmtOperation = (row: ResOperationLog) => {
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
