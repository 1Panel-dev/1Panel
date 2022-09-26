<template>
    <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search">
        <el-table-column :label="$t('app.appName')" prop="appName"></el-table-column>
        <el-table-column :label="$t('app.version')" prop="version"></el-table-column>
        <el-table-column :label="$t('app.container')">
            <template #default="{ row }">
                {{ row.ready / row.total }}
            </template>
        </el-table-column>
        <el-table-column :label="$t('app.status')">
            <template #default="{ row }">
                <el-tag>{{ row.status }}</el-tag>
            </template>
        </el-table-column>
        <el-table-column
            prop="createdAt"
            :label="$t('commons.table.date')"
            :formatter="dateFromat"
            show-overflow-tooltip
        />
        <fu-table-operations :ellipsis="10" :buttons="buttons" :label="$t('commons.table.operate')" fixed="right" fix />
    </ComplexTable>
</template>

<script lang="ts" setup>
import { GetAppInstalled } from '@/api/modules/app';
import { onMounted, reactive, ref } from 'vue';
import ComplexTable from '@/components/complex-table/index.vue';
import { dateFromat } from '@/utils/util';
import i18n from '@/lang';

let data = ref<any>();

const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 20,
    total: 0,
});

const search = () => {
    const req = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };

    GetAppInstalled(req).then((res) => {
        data.value = res.data.items;
        paginationConfig.total = res.data.total;
    });
};

const buttons = [
    {
        label: i18n.global.t('app.restart'),
    },
    {
        label: i18n.global.t('app.up'),
    },
    {
        label: i18n.global.t('app.down'),
    },
];

onMounted(() => {
    search();
});
</script>

<style lang="scss">
.i-card {
    height: 60px;
    cursor: pointer;
    .content {
        .image {
            width: auto;
            height: auto;
        }
    }
}
.i-card:hover {
    border: 1px solid;
    border-color: $primary-color;
    z-index: 1;
}
</style>
