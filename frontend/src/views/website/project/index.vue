<template>
    <LayoutContent :header="'网站'">
        <ComplexTable :data="data" @search="search()">
            <template #toolbar>
                <el-button type="primary" plain @click="openCreate">{{ '新建网站' }}</el-button>
                <el-button type="primary" plain>{{ '修改默认页' }}</el-button>
                <el-button type="primary" plain>{{ '默认站点' }}</el-button>
            </template>
            <el-table-column :label="$t('commons.table.name')" fix show-overflow-tooltip prop="primaryDomain">
                <template #default="{ row }">
                    <el-link @click="openConfig">{{ row.primaryDomain }}</el-link>
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
    </LayoutContent>
</template>

<script lang="ts" setup>
import LayoutContent from '@/layout/layout-content.vue';
import ComplexTable from '@/components/complex-table/index.vue';
import { onMounted, reactive, ref } from '@vue/runtime-core';
import router from '@/routers';
import CreateWebSite from './create/index.vue';
import DeleteWebsite from './delete/index.vue';
import { SearchWebSites } from '@/api/modules/website';
import i18n from '@/lang';
import { WebSite } from '@/api/interface/website';

const createRef = ref();
const deleteRef = ref();

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

const openConfig = () => {
    router.push({ name: 'WebsiteConfig' });
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

onMounted(() => {
    search();
});
</script>
