<template>
    <el-dialog
        v-model="open"
        :destroy-on-close="true"
        :close-on-click-modal="false"
        :title="$t('website.dnsAccountManage')"
    >
        <ComplexTable :data="data" :pagination-config="paginationConfig" @search="search()">
            <template #toolbar>
                <el-button type="primary" plain @click="openCreate">{{ $t('commons.button.create') }}</el-button>
            </template>
            <el-table-column :label="$t('commons.table.name')" fix show-overflow-tooltip prop="name"></el-table-column>
            <el-table-column :label="$t('commons.table.type')" prop="type">
                <template #default="{ row }">
                    <span v-if="row.type == 'AliYun'">{{ $t('website.aliyun') }}</span>
                    <span v-else>{{ row.type }}</span>
                </template>
            </el-table-column>
            <el-table-column :label="$t('website.key')">
                <template #default="{ row }">
                    <el-link @click="openEdit(row)">{{ $t('website.check') }}</el-link>
                </template>
            </el-table-column>
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
import ComplexTable from '@/components/complex-table/index.vue';
import Create from './create/index.vue';
import { WebSite } from '@/api/interface/website';
import { DeleteDnsAccount, SearchDnsAccount } from '@/api/modules/website';
import { onMounted, reactive, ref } from 'vue';
import i18n from '@/lang';
import { useDeleteData } from '@/hooks/use-delete-data';

const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 20,
    total: 0,
});
let data = ref<WebSite.DnsAccount[]>();
let createRef = ref();
let loading = ref(false);
let open = ref(false);

const buttons = [
    {
        label: i18n.global.t('app.delete'),
        click: function (row: WebSite.WebSite) {
            deleteAccount(row.id);
        },
    },
];

const acceptParams = () => {
    search();
    open.value = true;
};

const search = () => {
    const req = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    SearchDnsAccount(req).then((res) => {
        data.value = res.data.items;
        paginationConfig.total = res.data.total;
    });
};

const openCreate = () => {
    createRef.value.acceptParams({ mode: 'add' });
};

const openEdit = (form: WebSite.DnsAccount) => {
    createRef.value.acceptParams({ mode: 'edit', form: form });
};

const deleteAccount = async (id: number) => {
    loading.value = true;
    await useDeleteData(DeleteDnsAccount, id, 'commons.msg.delete', false);
    loading.value = false;
    search();
};

onMounted(() => {
    search();
});

defineExpose({
    acceptParams,
});
</script>
