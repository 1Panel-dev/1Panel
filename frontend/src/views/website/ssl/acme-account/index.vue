<template>
    <div>
        <RouterButton :buttons="routerButton" />
        <LayoutContent backName="SSL" :title="$t('website.acmeAccountManage')">
            <template #leftToolBar>
                <el-button v-permission type="primary" @click="openCreate">
                    {{ $t('commons.button.create') }}
                </el-button>
            </template>
            <template #prompt>
                <el-alert :title="$t('ssl.acmeHelper')" type="info" :closable="false" />
            </template>
            <template #main>
                <ComplexTable :data="data" :pagination-config="paginationConfig" @search="search()" v-loading="loading">
                    <el-table-column
                        :label="$t('website.email')"
                        fix
                        show-overflow-tooltip
                        prop="email"
                        min-width="100px"
                    ></el-table-column>
                    <el-table-column :label="$t('website.acmeAccountType')" fix show-overflow-tooltip prop="type">
                        <template #default="{ row }">
                            {{ getAccountName(row.type) }}
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('website.keyType')" prop="keyType">
                        <template #default="{ row }">
                            {{ getKeyName(row.keyType) }}
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('website.useProxy')" min-width="100px" v-if="isProductPro">
                        <template #default="{ row }">
                            <el-switch v-permission v-model="row.useProxy" @change="update(row)"></el-switch>
                        </template>
                    </el-table-column>
                    <el-table-column label="URL" show-overflow-tooltip prop="url" min-width="300px"></el-table-column>
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
import { Website } from '@/api/interface/website';
import { deleteAcmeAccount, searchAcmeAccount, updateAcmeAccount } from '@/api/modules/website';
import i18n from '@/lang';
import { useSearchState } from '@/utils/search-reset';
import { onMounted, reactive, ref, toRef } from 'vue';
import Create from './create/index.vue';
import { getAccountName, getKeyName } from '@/utils/ssl';
import { MsgSuccess } from '@/utils/message';
import { useGlobalStore } from '@/composables/useGlobalStore';

const { isProductPro } = useGlobalStore();
const loading = ref(false);
const data = ref();
const createRef = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'acme-account-page-size',
    currentPage: 1,
    pageSize: readCachedPageSize('acme-account-page-size') || 20,
    total: 0,
});
const opRef = ref();

const buttons = [
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
        click: function (row: Website.AcmeAccount) {
            deleteAccount(row);
        },
    },
];

const search = async () => {
    const req = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    await searchAcmeAccount(req).then((res) => {
        data.value = res.data.items;
        paginationConfig.total = res.data.total;
    });
};

const update = (row: Website.AcmeAccount) => {
    updateAcmeAccount(row).then(() => {
        search();
        MsgSuccess(i18n.global.t('commons.msg.updateSuccess'));
    });
};

const openCreate = () => {
    createRef.value.acceptParams();
};

const deleteAccount = async (row: any) => {
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: [row.email],
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('website.acmeAccountManage'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: deleteAcmeAccount,
        params: { id: row.id },
    });
};

const routerButton = [{ label: i18n.global.t('website.acmeAccountManage'), path: '/websites/ssl/acme-account' }];
useSearchState(
    'website/ssl/acme-account/index',
    {
        page: toRef(paginationConfig, 'currentPage'),
        pageSize: toRef(paginationConfig, 'pageSize'),
    },
    { onRevisit: () => search() },
);
onMounted(() => {
    search();
});
</script>
