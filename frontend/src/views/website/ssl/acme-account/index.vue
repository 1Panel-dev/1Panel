<template>
    <DrawerPro v-model="open" :header="$t('website.acmeAccountManage')" size="large" @close="handleClose">
        <template #content>
            <div class="mb-1.5">
                <el-alert :title="$t('ssl.acmeHelper')" type="info" :closable="false" />
            </div>
            <ComplexTable :data="data" :pagination-config="paginationConfig" v-loading="loading">
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
                <el-table-column :label="$t('website.keyType')" prop="keyType">
                    <template #default="{ row }">
                        {{ getKeyName(row.keyType) }}
                    </template>
                </el-table-column>
                <el-table-column :label="$t('website.useProxy')" min-width="100px" v-if="globalStore.isProductPro">
                    <template #default="{ row }">
                        <el-switch v-model="row.useProxy" @change="update(row)"></el-switch>
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
        </template>
    </DrawerPro>
    <OpDialog ref="opRef" @search="search" />
</template>

<script lang="ts" setup>
import { Website } from '@/api/interface/website';
import { deleteAcmeAccount, searchAcmeAccount, updateAcmeAccount } from '@/api/modules/website';
import i18n from '@/lang';
import { reactive, ref } from 'vue';
import Create from './create/index.vue';
import { getAccountName, getKeyName } from '@/utils/util';
import { MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';
const globalStore = GlobalStore();

const open = ref(false);
const loading = ref(false);
const data = ref();
const createRef = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'acme-account-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('acme-account-page-size')) || 20,
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
        api: deleteAcmeAccount,
        params: { id: row.id },
    });
};

defineExpose({
    acceptParams,
});
</script>
