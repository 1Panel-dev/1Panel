<template>
    <el-drawer :close-on-click-modal="false" v-model="open" size="50%">
        <template #header>
            <DrawerHeader :header="$t('ssl.ca')" :back="handleClose" />
        </template>
        <ComplexTable :data="data" :pagination-config="paginationConfig" @search="search()" v-loading="loading">
            <template #toolbar>
                <el-button type="primary" @click="openCreate">{{ $t('ssl.createCA') }}</el-button>
            </template>
            <el-table-column :label="$t('commons.table.name')" show-overflow-tooltip prop="name"></el-table-column>
            <el-table-column :label="$t('website.keyType')" show-overflow-tooltip prop="keyType">
                <template #default="{ row }">
                    {{ getKeyName(row.keyType) }}
                </template>
            </el-table-column>
            <el-table-column
                prop="createdAt"
                :label="$t('commons.table.date')"
                :formatter="dateFormat"
                show-overflow-tooltip
            />
            <fu-table-operations
                :ellipsis="3"
                :buttons="buttons"
                :label="$t('commons.table.operate')"
                fix
                width="250px"
            />
        </ComplexTable>
        <Create ref="createRef" @close="search()" />
        <Obtain ref="obtainRef" @close="search()" />
        <Detail ref="detailRef" />
    </el-drawer>
    <OpDialog ref="opRef" @search="search" />
</template>

<script lang="ts" setup>
import DrawerHeader from '@/components/drawer-header/index.vue';
import { Website } from '@/api/interface/website';
import { DeleteCA, SearchCAs } from '@/api/modules/website';
import i18n from '@/lang';
import { reactive, ref } from 'vue';
import Create from './create/index.vue';
import Detail from './detail/index.vue';
import { getKeyName, dateFormat } from '@/utils/util';
import Obtain from './obtain/index.vue';

const open = ref(false);
const loading = ref(false);
const data = ref();
const createRef = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'ca-page-size',
    currentPage: 1,
    pageSize: 20,
    total: 0,
});
const opRef = ref();
const obtainRef = ref();
const em = defineEmits(['close']);
const detailRef = ref();

const buttons = [
    {
        label: i18n.global.t('ssl.selfSign'),
        click: function (row: Website.CA) {
            obtain(row);
        },
    },
    {
        label: i18n.global.t('ssl.detail'),
        click: function (row: Website.CA) {
            detailRef.value.acceptParams(row.id);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: function (row: Website.CA) {
            deleteCA(row);
        },
    },
];

const acceptParams = () => {
    search();
    open.value = true;
};

const obtain = (row: any) => {
    obtainRef.value.acceptParams(row.id);
};

const search = async () => {
    const req = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    await SearchCAs(req).then((res) => {
        data.value = res.data.items;
        paginationConfig.total = res.data.total;
    });
};

const openCreate = () => {
    createRef.value.acceptParams();
};

const handleClose = () => {
    em('close', false);
    open.value = false;
};

const deleteCA = async (row: any) => {
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: [row.name],
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('website.ca'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: DeleteCA,
        params: { id: row.id },
    });
};

defineExpose({
    acceptParams,
});
</script>
