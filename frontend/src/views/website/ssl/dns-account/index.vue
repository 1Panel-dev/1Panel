<template>
    <div>
        <RouterButton :buttons="routerButton" />
        <LayoutContent backName="SSL" :title="$t('website.dnsAccountManage')">
            <template #leftToolBar>
                <el-button v-permission type="primary" @click="openCreate">
                    {{ $t('commons.button.create') }}
                </el-button>
            </template>
            <template #main>
                <ComplexTable :data="data" :pagination-config="paginationConfig" @search="search()">
                    <el-table-column
                        :label="$t('commons.table.name')"
                        fix
                        show-overflow-tooltip
                        prop="name"
                    ></el-table-column>
                    <el-table-column :label="$t('commons.table.type')" prop="type">
                        <template #default="{ row }">
                            <span>{{ getDNSName(row.type) }}</span>
                            <el-tooltip
                                v-if="row.type === 'DnsPod'"
                                :content="$t('ssl.dnsPodRemovedTip')"
                                placement="top"
                            >
                                <el-tag type="danger" size="small" class="ml-2">
                                    {{ $t('ssl.dnsPodRemoved') }}
                                </el-tag>
                            </el-tooltip>
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
            </template>
        </LayoutContent>
        <Create ref="createRef" @saved="search()" />
        <OpDialog drawer ref="opRef" @search="search" />
    </div>
</template>

<script lang="ts" setup>
import { readCachedPageSize } from '@/utils/panel-cache';
import Create from './create/index.vue';
import { Website } from '@/api/interface/website';
import { deleteDnsAccount, searchDnsAccount } from '@/api/modules/website';
import { onMounted, reactive, ref, toRef } from 'vue';
import i18n from '@/lang';
import { useSearchState } from '@/utils/search-reset';
import { getDNSName } from '@/utils/ssl';
const paginationConfig = reactive({
    cacheSizeKey: 'dns-account-page-size',
    currentPage: 1,
    pageSize: readCachedPageSize('dns-account-page-size') || 20,
    total: 0,
});
let data = ref<Website.DnsAccount[]>();
let createRef = ref();
const opRef = ref();

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        permission: true,
        click: function (row: Website.DnsAccount) {
            openEdit(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
        click: function (row: Website.DnsAccount) {
            deleteAccount(row);
        },
    },
];

const search = () => {
    const req = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    searchDnsAccount(req).then((res) => {
        data.value = res.data.items;
        paginationConfig.total = res.data.total;
    });
};

const openCreate = () => {
    createRef.value.acceptParams({ mode: 'create' });
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
        api: deleteDnsAccount,
        params: { id: row.id },
    });
};

onMounted(() => {
    search();
});

const routerButton = [{ label: i18n.global.t('website.dnsAccountManage'), path: '/websites/ssl/dns-account' }];
useSearchState(
    'website/ssl/dns-account/index',
    {
        page: toRef(paginationConfig, 'currentPage'),
        pageSize: toRef(paginationConfig, 'pageSize'),
    },
    { onRevisit: () => search() },
);
</script>
