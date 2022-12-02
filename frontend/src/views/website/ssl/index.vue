<template>
    <LayoutContent :header="$t('website.ssl')">
        <ComplexTable :data="data" :pagination-config="paginationConfig" @search="search()">
            <template #toolbar>
                <el-button type="primary" icon="Plus" @click="openSSL()">
                    {{ $t('commons.button.create') }}
                </el-button>
                <el-button type="primary" plain @click="openAcmeAccount()">
                    {{ $t('website.acmeAccountManage') }}
                </el-button>
                <el-button type="primary" plain @click="openDnsAccount()">
                    {{ $t('website.dnsAccountManage') }}
                </el-button>
            </template>
            <el-table-column
                :label="$t('website.domain')"
                fix
                show-overflow-tooltip
                prop="primaryDomain"
            ></el-table-column>
            <el-table-column
                :label="$t('website.otherDomains')"
                fix
                show-overflow-tooltip
                prop="domains"
            ></el-table-column>
            <el-table-column :label="$t('website.brand')" fix show-overflow-tooltip prop="type"></el-table-column>
            <el-table-column
                prop="expireDate"
                :label="$t('website.expireDate')"
                :formatter="dateFromat"
                show-overflow-tooltip
            />
            <fu-table-operations
                :ellipsis="1"
                :buttons="buttons"
                :label="$t('commons.table.operate')"
                fixed="right"
                fix
            />
        </ComplexTable>
        <DnsAccount ref="dnsAccountRef"></DnsAccount>
        <AcmeAccount ref="acmeAccountRef"></AcmeAccount>
        <Create ref="sslCreateRef" @close="search()"></Create>
        <Renew ref="renewRef" @close="search()"></Renew>
    </LayoutContent>
</template>

<script lang="ts" setup>
import LayoutContent from '@/layout/layout-content.vue';
import ComplexTable from '@/components/complex-table/index.vue';
import { onMounted, reactive, ref } from 'vue';
import { DeleteSSL, SearchSSL } from '@/api/modules/website';
import DnsAccount from './dns-account/index.vue';
import AcmeAccount from './acme-account/index.vue';
import Renew from './renew/index.vue';
import Create from './create/index.vue';
import { dateFromat } from '@/utils/util';
import i18n from '@/lang';
import { WebSite } from '@/api/interface/website';
import { useDeleteData } from '@/hooks/use-delete-data';

const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 20,
    total: 0,
});
const acmeAccountRef = ref();
const dnsAccountRef = ref();
const sslCreateRef = ref();
const renewRef = ref();
let data = ref();
let loading = ref(false);

const buttons = [
    {
        label: i18n.global.t('website.renewSSL'),
        click: function (row: WebSite.WebSite) {
            openRenewSSL(row.id);
        },
    },
    // {
    //     label: i18n.global.t('website.deploySSL'),
    //     click: function (row: WebSite.WebSite) {
    //         applySSL(row.id);
    //     },
    // },
    {
        label: i18n.global.t('app.delete'),
        click: function (row: WebSite.WebSite) {
            deleteSSL(row.id);
        },
    },
];

const search = () => {
    const req = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    loading.value = true;
    SearchSSL(req)
        .then((res) => {
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .finally(() => {
            loading.value = false;
        });
};

const openAcmeAccount = () => {
    acmeAccountRef.value.acceptParams();
};
const openDnsAccount = () => {
    dnsAccountRef.value.acceptParams();
};
const openSSL = () => {
    sslCreateRef.value.acceptParams();
};
const openRenewSSL = (id: number) => {
    renewRef.value.acceptParams(id);
};

const deleteSSL = async (id: number) => {
    loading.value = true;
    await useDeleteData(DeleteSSL, id, 'commons.msg.delete', false);
    loading.value = false;
    search();
};

// const renewSSL = async (id: number) => {
//     loading.value = true;
//     await useDeleteData(RenewSSL, { SSLId: id }, 'website.renewHelper', false);
//     loading.value = false;
//     search();
// };

// const applySSL = async (sslId: number) => {
//     loading.value = true;
//     const apply = {
//         websiteId: 0,
//         SSLId: sslId,
//     };
//     await useDeleteData(ApplySSL, apply, 'website.deploySSLHelper', false);
//     loading.value = false;
//     search();
// };

onMounted(() => {
    search();
});
</script>
