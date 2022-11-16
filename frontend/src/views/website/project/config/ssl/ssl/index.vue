<template>
    <div>
        <ComplexTable :data="data" :pagination-config="paginationConfig" @search="search()">
            <template #toolbar>
                <el-button type="primary" plain @click="openSSL()">{{ $t('commons.button.create') }}</el-button>
                <el-button type="primary" plain @click="openAccount()">{{ $t('website.accountManage') }}</el-button>
            </template>
            <el-table-column :label="$t('website.domain')" fix show-overflow-tooltip prop="domain"></el-table-column>
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
        <Account ref="accountRef"></Account>
        <Create :id="id" ref="sslCreateRef"></Create>
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import { computed, onMounted, reactive, ref } from 'vue';
import { ApplySSL, DeleteSSL, SearchSSL } from '@/api/modules/website';
import Account from './account/index.vue';
import Create from './create/index.vue';
import { dateFromat } from '@/utils/util';
import i18n from '@/lang';
import { WebSite } from '@/api/interface/website';
import { useDeleteData } from '@/hooks/use-delete-data';

const props = defineProps({
    id: {
        type: Number,
        default: 0,
    },
});

const id = computed(() => {
    return props.id;
});

const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 20,
    total: 0,
});
const accountRef = ref();
const sslCreateRef = ref();
let data = ref();
let loading = ref(false);

const buttons = [
    {
        label: i18n.global.t('website.deploySSL'),
        click: function (row: WebSite.WebSite) {
            applySSL(row.id);
        },
    },
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

const openAccount = () => {
    accountRef.value.acceptParams();
};
const openSSL = () => {
    sslCreateRef.value.acceptParams();
};

const deleteSSL = async (id: number) => {
    loading.value = true;
    await useDeleteData(DeleteSSL, id, 'commons.msg.delete', false);
    loading.value = false;
    search();
};

const applySSL = async (sslId: number) => {
    loading.value = true;
    const apply = {
        websiteId: Number(id.value),
        SSLId: sslId,
    };
    await useDeleteData(ApplySSL, apply, 'website.deploySSLHelper', false);
    loading.value = false;
    search();
};

onMounted(() => {
    search();
});
</script>
