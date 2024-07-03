<template>
    <el-drawer :close-on-click-modal="false" :close-on-press-escape="false" v-model="open" size="60%">
        <template #header>
            <DrawerHeader :header="$t('website.acmeAccountManage')" :back="handleClose" />
        </template>
        <div class="mb-1.5">
            <el-alert :title="$t('ssl.acmeHelper')" type="info" :closable="false" />
        </div>
        <ComplexTable :data="data" :pagination-config="paginationConfig" @search="search()" v-loading="loading">
            <template #toolbar>
                <el-button type="primary" @click="openCreate">{{ $t('commons.button.create') }}</el-button>
            </template>
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
            <el-table-column :label="$t('website.keyType')" fix show-overflow-tooltip prop="keyType">
                <template #default="{ row }">
                    {{ getKeyName(row.keyType) }}
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
        <Create ref="createRef" @close="search()"></Create>
    </el-drawer>
    <OpDialog ref="opRef" @search="search" />
</template>

<script lang="ts" setup>
import DrawerHeader from '@/components/drawer-header/index.vue';
import { Website } from '@/api/interface/website';
import { DeleteAcmeAccount, SearchAcmeAccount } from '@/api/modules/website';
import i18n from '@/lang';
import { reactive, ref } from 'vue';
import Create from './create/index.vue';
import { getAccountName, getKeyName } from '@/utils/util';

const open = ref(false);
const loading = ref(false);
const data = ref();
const createRef = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'acme-account-page-size',
    currentPage: 1,
    pageSize: 20,
    total: 0,
});
const opRef = ref();

const buttons = [
    {
        label: i18n.global.t('commons.button.delete'),
        click: function (row: Website.AcmeAccount) {
            deleteAccount(row);
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

const deleteAccount = async (row: any) => {
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: [row.email],
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('website.acmeAccountManage'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: DeleteAcmeAccount,
        params: { id: row.id },
    });
};

defineExpose({
    acceptParams,
});
</script>
