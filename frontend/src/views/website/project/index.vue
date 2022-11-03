<template>
    <LayoutContent :header="'网站'">
        <ComplexTable :data="data" @search="search()">
            <template #toolbar>
                <el-button type="primary" plain @click="openCreate">{{ $t('commons.button.create') }}</el-button>
                <el-button type="primary" plain @click="openGroup">{{ $t('website.group') }}</el-button>
                <el-button type="primary" plain>{{ '修改默认页' }}</el-button>
                <el-button type="primary" plain>{{ '默认站点' }}</el-button>
            </template>
            <el-table-column :label="$t('commons.table.name')" fix show-overflow-tooltip prop="primaryDomain">
                <template #default="{ row }">
                    <el-link @click="openConfig(row.id)">{{ row.primaryDomain }}</el-link>
                </template>
            </el-table-column>
            <el-table-column :label="$t('commons.table.status')" prop="status"></el-table-column>
            <el-table-column :label="'备份'" prop="backup"></el-table-column>
            <el-table-column :label="'备注'" prop="remark"></el-table-column>
            <el-table-column :label="'SSL证书'" prop="ssl"></el-table-column>
            <fu-table-operations
                :ellipsis="1"
                :buttons="buttons"
                :label="$t('commons.table.operate')"
                fixed="right"
                fix
            />
        </ComplexTable>
        <CreateWebSite ref="createRef" @close="search"></CreateWebSite>
        <DeleteWebsite ref="deleteRef" @close="search"></DeleteWebsite>
        <WebSiteGroup ref="groupRef"></WebSiteGroup>
    </LayoutContent>
</template>

<script lang="ts" setup>
import LayoutContent from '@/layout/layout-content.vue';
import ComplexTable from '@/components/complex-table/index.vue';
import { onMounted, reactive, ref } from '@vue/runtime-core';
import CreateWebSite from './create/index.vue';
import DeleteWebsite from './delete/index.vue';
import WebSiteGroup from './group/index.vue';
import { SearchWebSites } from '@/api/modules/website';
import { WebSite } from '@/api/interface/website';

import i18n from '@/lang';
import router from '@/routers';

const createRef = ref();
const deleteRef = ref();
const groupRef = ref();

const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 20,
    total: 0,
});

const data = ref();
const search = async () => {
    const req = {
        name: '',
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };

    SearchWebSites(req).then((res) => {
        data.value = res.data.items;
    });
};

const openConfig = (id: number) => {
    router.push({ name: 'WebsiteConfig', params: { id: id } });
};

const buttons = [
    {
        label: '设置',
        click: open,
    },
    {
        label: i18n.global.t('app.delete'),
        click: function (row: WebSite.WebSite) {
            openDelete(row.id);
        },
    },
];

const openDelete = (id: number) => {
    deleteRef.value.acceptParams(id);
};

const openCreate = () => {
    createRef.value.acceptParams();
};

const openGroup = () => {
    groupRef.value.acceptParams();
};

onMounted(() => {
    search();
});
</script>
