<template>
    <el-card>
        <LayoutContent :header="$t('website.ssl')">
            <el-alert type="info" :closable="false">
                <template #default>
                    <span><span v-html="$t('website.encryptHelper')"></span></span>
                </template>
            </el-alert>
            <br />
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
                <el-table-column :label="$t('ssl.provider')" fix show-overflow-tooltip prop="provider">
                    <template #default="{ row }">{{ getProvider(row.provider) }}</template>
                </el-table-column>
                <el-table-column
                    :label="$t('ssl.acmeAccount')"
                    fix
                    show-overflow-tooltip
                    prop="acmeAccount.email"
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
    </el-card>
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
import { Website } from '@/api/interface/website';
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
        click: function (row: Website.Website) {
            openRenewSSL(row.id);
        },
    },
    {
        label: i18n.global.t('app.delete'),
        click: function (row: Website.Website) {
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
    await useDeleteData(DeleteSSL, { id: id }, 'commons.msg.delete');
    loading.value = false;
    search();
};

const getProvider = (provider: string) => {
    switch (provider) {
        case 'dnsAccount':
            return i18n.global.t('website.dnsAccount');
        case 'dnsManual':
            return i18n.global.t('website.dnsAccount');
        case 'http':
            return 'HTTP';
        default:
            return i18n.global.t('ssl.manualCreate');
    }
};

onMounted(() => {
    search();
});
</script>
