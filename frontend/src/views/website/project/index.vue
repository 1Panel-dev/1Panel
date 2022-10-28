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
        <CreateWebSite ref="createRef"></CreateWebSite>
    </LayoutContent>
</template>

<script lang="ts" setup>
import LayoutContent from '@/layout/layout-content.vue';
import ComplexTable from '@/components/complex-table/index.vue';
import { onMounted, ref } from '@vue/runtime-core';
import router from '@/routers';
import CreateWebSite from './create/index.vue';

const createRef = ref();

const data = ref();
const search = async () => {
    data.value = [
        {
            primaryDomain: 'www.baicu.com',
            status: 'Running',
            backup: '1',
            remark: '主网站',
        },
    ];
};

const openConfig = () => {
    router.push({ name: 'WebsiteConfig' });
};

const buttons = [
    {
        label: '设置',
        click: open,
    },
];

const openCreate = () => {
    createRef.value.acceptParams();
};

onMounted(() => {
    search();
});
</script>
