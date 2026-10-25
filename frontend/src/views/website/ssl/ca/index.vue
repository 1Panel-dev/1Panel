<template>
    <div>
        <RouterButton :buttons="routerButton" />
        <LayoutContent backName="SSL" :title="$t('ssl.selfSigned')">
            <template #leftToolBar>
                <el-button v-permission type="primary" @click="openCreate">
                    {{ $t('commons.button.create') }}
                </el-button>
                <el-button v-permission type="primary" plain @click="uploadRef?.acceptParams()">
                    {{ $t('commons.button.upload') }}
                </el-button>
            </template>
            <template #main>
                <ComplexTable :data="data" :pagination-config="paginationConfig" @search="search()" v-loading="loading">
                    <el-table-column
                        :label="$t('commons.table.name')"
                        show-overflow-tooltip
                        prop="name"
                    ></el-table-column>
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
            </template>
        </LayoutContent>
        <Upload ref="uploadRef" @uploaded="search()" />
        <Create ref="createRef" @close="search()" />
        <Detail ref="detailRef" />
        <OpDialog ref="opRef" @search="search" />
    </div>
</template>

<script lang="ts" setup>
import { readCachedPageSize } from '@/utils/panel-cache';
import { Website } from '@/api/interface/website';
import { deleteCA, searchCAs, downloadCAFile } from '@/api/modules/website';
import i18n from '@/lang';
import { onMounted, reactive, ref, toRef } from 'vue';
import { useSearchState } from '@/utils/search-reset';
import Create from './create/index.vue';
import Upload from './upload/index.vue';
import Detail from './detail/index.vue';
import { getKeyName } from '@/utils/ssl';
import { dateFormat } from '@/utils/date';

const routerButton = [{ label: i18n.global.t('ssl.selfSigned'), path: '/websites/ssl/ca' }];
const loading = ref(false);
const data = ref();
const createRef = ref();
const uploadRef = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'ca-page-size',
    currentPage: 1,
    pageSize: readCachedPageSize('ca-page-size') || 20,
    total: 0,
});
const opRef = ref();
const detailRef = ref();

const buttons = [
    {
        label: i18n.global.t('ssl.detail'),
        click: function (row: Website.CA) {
            detailRef.value.acceptParams(row.id);
        },
    },
    {
        label: i18n.global.t('commons.button.download'),
        click: function (row: Website.CA) {
            onDownload(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
        click: function (row: Website.CA) {
            deleteca(row);
        },
    },
];

useSearchState(
    'website/ssl/ca/index',
    {
        page: toRef(paginationConfig, 'currentPage'),
        pageSize: toRef(paginationConfig, 'pageSize'),
    },
    { onRevisit: () => search() },
);

const search = async () => {
    const req = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    await searchCAs(req).then((res) => {
        data.value = res.data.items;
        paginationConfig.total = res.data.total;
    });
};

const openCreate = () => {
    createRef.value.acceptParams();
};

const deleteca = async (row: any) => {
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: [row.name],
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('ssl.ca'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: deleteCA,
        params: { id: row.id },
    });
};

const onDownload = (row: Website.CA) => {
    loading.value = true;
    downloadCAFile({ id: row.id })
        .then((res) => {
            const downloadUrl = window.URL.createObjectURL(new Blob([res]));
            const a = document.createElement('a');
            a.style.display = 'none';
            a.href = downloadUrl;
            a.download = row.name + '.zip';
            const event = new MouseEvent('click');
            a.dispatchEvent(event);
            window.URL.revokeObjectURL(downloadUrl);
        })
        .finally(() => {
            loading.value = false;
        });
};

onMounted(() => {
    search();
});
</script>
