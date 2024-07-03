<template>
    <el-drawer :close-on-click-modal="false" :close-on-press-escape="false" v-model="open" :size="'50%'">
        <template #header>
            <DrawerHeader :header="$t('website.dnsAccountManage')" :back="handleClose" />
        </template>
        <ComplexTable :data="data" :pagination-config="paginationConfig" @search="search()">
            <template #toolbar>
                <el-button type="primary" @click="openCreate">
                    {{ $t('commons.button.create') }}
                </el-button>
            </template>
            <el-table-column :label="$t('commons.table.name')" fix show-overflow-tooltip prop="name"></el-table-column>
            <el-table-column :label="$t('commons.table.type')" prop="type">
                <template #default="{ row }">
                    <span>{{ getDNSName(row.type) }}</span>
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
    </el-drawer>

    <OpDialog ref="opRef" @search="search" />
</template>

<script lang="ts" setup>
import DrawerHeader from '@/components/drawer-header/index.vue';
import Create from './create/index.vue';
import { Website } from '@/api/interface/website';
import { DeleteDnsAccount, SearchDnsAccount } from '@/api/modules/website';
import { onMounted, reactive, ref } from 'vue';
import i18n from '@/lang';
import { getDNSName } from '@/utils/util';

const paginationConfig = reactive({
    cacheSizeKey: 'dns-account-page-size',
    currentPage: 1,
    pageSize: 20,
    total: 0,
});
let data = ref<Website.DnsAccount[]>();
let createRef = ref();
let open = ref(false);
const opRef = ref();

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: function (row: Website.DnsAccount) {
            openEdit(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: function (row: Website.DnsAccount) {
            deleteAccount(row);
        },
    },
];

const acceptParams = () => {
    search();
    open.value = true;
};

const handleClose = () => {
    open.value = false;
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

const openEdit = (form: Website.DnsAccount) => {
    createRef.value.acceptParams({ mode: 'edit', form: form });
};

const deleteAccount = async (row: any) => {
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: [row.name],
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('website.dnsAccountManage'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: DeleteDnsAccount,
        params: { id: row.id },
    });
};

onMounted(() => {
    search();
});

defineExpose({
    acceptParams,
});
</script>
