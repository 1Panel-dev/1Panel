<template>
    <LayoutContent :header="'样例'">
        <ComplexTable :pagination-config="paginationConfig" v-model:selects="selects" :data="data" @search="search">
            <template #toolbar>
                <el-button type="primary">{{ $t('commons.button.create') }}</el-button>
                <el-button type="primary" plain>{{ '其他操作' }}</el-button>
                <el-button type="danger" plain :disabled="selects.length === 0" @click="batchDelete">{{
                    $t('commons.button.delete')
                }}</el-button>
            </template>
            <el-table-column type="selection" fix />
            <el-table-column label="ID" min-width="100" prop="ID" fix />
            <el-table-column :label="$t('commons.table.name')" min-width="100" prop="name" fix>
                <template #default="{ row }">
                    <fu-input-rw-switch v-model="row.name" size="mini" />
                </template>
            </el-table-column>
            <el-table-column label="Email" min-width="100" prop="email" />
            <el-table-column
                prop="createdAt"
                :label="$t('commons.table.createdAt')"
                :formatter="dateFromat"
                show-overflow-tooltip
                width="200"
            />
            <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fix />
        </ComplexTable>
    </LayoutContent>
</template>
<script setup lang="ts">
import LayoutContent from '@/layout/LayoutContent.vue';
import ComplexTable from '@/components/complex-table/index.vue';
import { dateFromat } from '@/utils/util';
import { User } from '@/api/interface/user';
import { deleteUser, getUserList } from '@/api/modules/user';
import { onMounted, reactive, ref } from '@vue/runtime-core';
import { useDeleteData } from '@/hooks/useDeleteData';
import i18n from '@/lang';
const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 5,
    total: 0,
});
const buttons = [
    {
        label: '编辑',
        click: edit,
    },
    // {
    //     label: '执行',
    //     click: buttonClick,
    // },
    // {
    //     label: '删除',
    //     type: 'danger',
    //     click: buttonClick,
    // },
    // {
    //     label: '复制',
    //     click: buttonClick,
    // },
    // {
    //     label: '定时任务',
    //     click: buttonClick,
    // },
];

// function select(row: any) {
//     console.log(row);
//     selects.value.push(row.ID);
//     console.log(selects);
// }
function edit(row: User.User) {
    console.log(row);
}

const batchDelete = async () => {
    let ids: Array<number> = [];
    selects.value.forEach((item: User.User) => {
        ids.push(item.ID);
    });
    await useDeleteData(deleteUser, { ids: ids }, i18n.global.t('commons.msg.delete'));
};

const search = async () => {
    const { currentPage, pageSize } = paginationConfig;
    const res = getUserList({ currentPage, pageSize });
    data.value = res.items;
    paginationConfig.total = res.total;
};

onMounted(() => {
    search();
});
</script>
