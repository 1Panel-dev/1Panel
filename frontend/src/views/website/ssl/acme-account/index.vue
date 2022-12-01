<template>
    <el-dialog
        v-model="open"
        :title="$t('website.acmeAccountManage')"
        :destroy-on-close="true"
        :close-on-click-modal="false"
    >
        <ComplexTable :data="data" :pagination-config="paginationConfig" @search="search()" v-loading="loading">
            <template #toolbar>
                <el-button type="primary" plain @click="openCreate">{{ $t('commons.button.create') }}</el-button>
            </template>
            <el-table-column :label="$t('website.email')" fix show-overflow-tooltip prop="email"></el-table-column>
            <el-table-column label="URL" show-overflow-tooltip prop="url"></el-table-column>
            <fu-table-operations
                :ellipsis="1"
                :buttons="buttons"
                :label="$t('commons.table.operate')"
                fixed="right"
                fix
            />
        </ComplexTable>
        <Create ref="createRef" @close="search()"></Create>
    </el-dialog>
</template>
<script lang="ts" setup>
import { WebSite } from '@/api/interface/website';
import { DeleteAcmeAccount, SearchAcmeAccount } from '@/api/modules/website';
import ComplexTable from '@/components/complex-table/index.vue';
import { useDeleteData } from '@/hooks/use-delete-data';
import i18n from '@/lang';
import { reactive, ref } from 'vue';
import Create from './create/index.vue';

let open = ref(false);
let loading = ref(false);
let data = ref();
let createRef = ref();
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 20,
    total: 0,
});

const buttons = [
    {
        label: i18n.global.t('app.delete'),
        click: function (row: WebSite.AcmeAccount) {
            deleteAccount(row.id);
        },
    },
];

const acceptParams = () => {
    search();
    open.value = true;
};

const search = async () => {
    const req = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    await SearchAcmeAccount(req).then((res) => {
        data.value = res.data.items;
        paginationConfig.total = res.data.total;
    });
};

const openCreate = () => {
    createRef.value.acceptParams();
};

const deleteAccount = async (id: number) => {
    await useDeleteData(DeleteAcmeAccount, id, 'commons.msg.delete', loading.value);
    search();
};

defineExpose({
    acceptParams,
});
</script>
