<template>
    <LayoutContent :header="'样例'">
        <ComplexTable :pagination-config="paginationConfig" v-model:selects="selects" :data="data" @search="search">
            <template #toolbar>
                <el-button type="primary" @click="openOperate(null)">{{ $t('commons.button.create') }}</el-button>
                <el-button type="primary" plain>{{ '其他操作' }}</el-button>
                <el-button type="danger" plain :disabled="selects.length === 0" @click="batchDelete(null)">
                    {{ $t('commons.button.delete') }}
                </el-button>
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
    pageSize: 20,
    total: 0,
});
const userSearch = reactive({
    page: 1,
    pageSize: 20,
});

const openOperate = (row: User.User | null) => {
    let params: { [key: string]: any } = {
        op: 'create',
    };
    if (row !== null) {
        params.op = 'edit';
        params.id = row.id;
    }

    router.push({ name: 'DemoOperate', params });
};

const batchDelete = async (row: User.User | null) => {
    let ids: Array<number> = [];

    if (row === null) {
        selects.value.forEach((item: User.User) => {
            ids.push(item.id);
        });
    } else {
        ids.push(row.id);
    }
    await useDeleteData(deleteUser, { ids: ids }, 'commons.msg.delete', true);
    search();
};
const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: openOperate,
    },
    {
        label: i18n.global.t('commons.button.delete'),
        type: 'danger',
        click: batchDelete,
    },
];

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
