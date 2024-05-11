<template>
    <el-drawer
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        v-model="open"
        size="50%"
        :before-close="handleClose"
    >
        <template #header>
            <DrawerHeader :header="$t('php.extensions')" :back="handleClose" />
        </template>
        <ComplexTable :data="data" @search="search()" :pagination-config="paginationConfig">
            <template #toolbar>
                <el-button type="primary" @click="openCreate">{{ $t('commons.button.create') }}</el-button>
            </template>
            <el-table-column :label="$t('commons.table.name')" width="150px" prop="name"></el-table-column>
            <el-table-column :label="$t('php.extension')" fix prop="extensions"></el-table-column>
            <fu-table-operations
                :ellipsis="10"
                width="120px"
                :buttons="buttons"
                :label="$t('commons.table.operate')"
                fixed="right"
                fix
            />
        </ComplexTable>
        <Create ref="createRef" @close="search()" />
        <OpDialog ref="opRef" @search="search" />
    </el-drawer>
</template>
<script lang="ts" setup>
import { DeletePHPExtensions, SearchPHPExtensions } from '@/api/modules/runtime';
import { reactive, ref } from 'vue';
import Create from './operate/index.vue';
import { Runtime } from '@/api/interface/runtime';
import i18n from '@/lang';

const open = ref(false);
const data = ref();
const createRef = ref();
const opRef = ref();

const paginationConfig = reactive({
    cacheSizeKey: 'website-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('website-page-size')) || 10,
    total: 0,
});

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: function (row: Runtime.PHPExtensions) {
            openUpdate(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: function (row: Runtime.PHPExtensions) {
            openDelete(row);
        },
    },
];

const handleClose = () => {
    open.value = false;
};

const acceptParams = (): void => {
    open.value = true;
    search();
};

const openCreate = () => {
    createRef.value.acceptParams('create');
};

const search = async () => {
    try {
        const res = await SearchPHPExtensions({
            page: paginationConfig.currentPage,
            pageSize: paginationConfig.pageSize,
        });
        data.value = res.data.items;
        paginationConfig.total = res.data.total;
    } catch (error) {}
};

const openDelete = async (row: Runtime.PHPExtensions) => {
    opRef.value.acceptParams({
        title: i18n.global.t('commons.msg.deleteTitle'),
        names: [row.name],
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('php.extensions'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: DeletePHPExtensions,
        params: { id: row.id },
    });
};

const openUpdate = async (row: Runtime.PHPExtensions) => {
    createRef.value.acceptParams('edit', row);
};

defineExpose({ acceptParams });
</script>
