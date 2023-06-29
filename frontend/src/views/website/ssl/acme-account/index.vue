<template>
    <el-drawer :close-on-click-modal="false" v-model="open" :size="'50%'">
        <template #header>
            <DrawerHeader :header="$t('website.acmeAccountManage')" :back="handleClose" />
        </template>
        <el-alert :title="$t('ssl.acmeHelper')" type="info" :closable="false" style="margin-bottom: 5px" />
        <ComplexTable :data="data" :pagination-config="paginationConfig" @search="search()" v-loading="loading">
            <template #toolbar>
                <el-button type="primary" @click="openCreate">{{ $t('website.addAccount') }}</el-button>
            </template>
            <el-table-column :label="$t('website.email')" fix show-overflow-tooltip prop="email"></el-table-column>
            <el-table-column label="URL" show-overflow-tooltip prop="url" min-width="300px"></el-table-column>
            <fu-table-operations
                :ellipsis="1"
                :buttons="buttons"
                :label="$t('commons.table.operate')"
                fixed="right"
                fix
            />
        </ComplexTable>
        <Create ref="createRef" @close="search()"></Create>
    </el-drawer>
</template>
<script lang="ts" setup>
import DrawerHeader from '@/components/drawer-header/index.vue';
import { Website } from '@/api/interface/website';
import { DeleteAcmeAccount, SearchAcmeAccount } from '@/api/modules/website';
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
        label: i18n.global.t('commons.button.delete'),
        click: function (row: Website.AcmeAccount) {
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

const handleClose = () => {
    open.value = false;
};

const deleteAccount = async (id: number) => {
    await useDeleteData(DeleteAcmeAccount, { id: id }, 'commons.msg.delete');
    search();
};

defineExpose({
    acceptParams,
});
</script>
