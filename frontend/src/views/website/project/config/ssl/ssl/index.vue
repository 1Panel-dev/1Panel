<template>
    <div>
        <ComplexTable :data="data" :pagination-config="paginationConfig" @search="search()">
            <template #toolbar>
                <el-button type="primary" plain>{{ $t('commons.button.create') }}</el-button>
                <el-button type="primary" plain @click="openAccount()">{{ $t('website.accountManage') }}</el-button>
            </template>
            <el-table-column :label="$t('commons.table.name')" fix show-overflow-tooltip prop="name"></el-table-column>
        </ComplexTable>
        <Account ref="accountRef"></Account>
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import { onMounted, reactive, ref } from 'vue';
import { SearchSSL } from '@/api/modules/website';
import Account from './account/index.vue';

const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 20,
    total: 0,
});
const accountRef = ref();
let data = ref();

const search = () => {
    const req = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    SearchSSL(req).then((res) => {
        data.value = res.data.items;
        paginationConfig.total = res.data.total;
    });
};

const openAccount = () => {
    accountRef.value.acceptParams();
};

onMounted(() => {
    search();
});
</script>
