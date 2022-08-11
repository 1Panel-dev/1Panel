<template>
    <LayoutContent :header="'样例'">
        <ComplexTable :pagination-config="paginationConfig" v-model:selects="selects" :data="data" @search="search">
            <template #toolbar>
                <el-button type="primary" @click="openOperate({})">{{ $t('commons.button.create') }}</el-button>
                <el-button type="primary" plain>{{ '其他操作' }}</el-button>
                <el-button type="danger" plain :disabled="selects.length === 0" @click="batchDelete">{{
                    $t('commons.button.delete')
                }}</el-button>
            </template>
            <el-table-column type="selection" fix />
            <el-table-column label="ID" min-width="100" prop="id" fix />
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
import LayoutContent from '@/layout/layout-content.vue';
import ComplexTable from '@/components/complex-table/index.vue';
import { dateFromat } from '@/utils/util';
import { User } from '@/api/interface/user';
import { deleteUser, getUserList } from '@/api/modules/user';
import { onMounted, reactive, ref } from '@vue/runtime-core';
import { useDeleteData } from '@/hooks/use-delete-data';
import i18n from '@/lang';
import { useRouter } from 'vue-router';
const router = useRouter();
const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    page: 1,
    pageSize: 5,
    total: 0,
});
const userSearch = reactive({
    page: 1,
    pageSize: 5,
});
const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
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

// interface OperateOpen {
//     acceptParams: (params: any) => void;
// }
// const operateRef = ref<OperateOpen>();

const openOperate = (row: User.User) => {
    console.log(row);
    // let title = 'commons.button.create';
    // if (row != null) {
    //     title = 'commons.button.edit';
    // }
    // let params = {
    //     titke: title,
    //     row: row,
    // };
    // operateRef.value!.acceptParams(params);
    router.push({ name: 'DemoCreate', params: { op: 'create' } });
    // router.push({ name: 'operate', params: { operate: 'create' } });
};

const batchDelete = async () => {
    let ids: Array<number> = [];
    selects.value.forEach((item: User.User) => {
        ids.push(item.id);
    });
    await useDeleteData(deleteUser, { ids: ids }, 'commons.msg.delete');
};

const search = async () => {
    userSearch.page = paginationConfig.page;
    userSearch.pageSize = paginationConfig.pageSize;
    const res = await getUserList(userSearch);
    data.value = res.data.items;
    paginationConfig.total = res.data.total;
};

onMounted(() => {
    search();
});
</script>
