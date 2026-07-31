<template>
    <div>
        <RouterButton :buttons="routerButton" />
        <LayoutContent :title="$t('menu.template')">
            <template #leftToolBar>
                <el-button v-permission type="primary" @click="openCreate()">
                    {{ $t('template.create') }}
                </el-button>
                <el-button v-permission type="primary" plain @click="openOutputList()">
                    {{ $t('template.outputList') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <TableSearch @search="search()" v-model:searchName="searchName" />
                <TableRefresh @search="search()" />
            </template>
            <template #main>
                <ComplexTable :data="data" :pagination-config="paginationConfig" @search="search()" v-loading="loading">
                    <el-table-column :label="$t('template.name')" prop="name" min-width="120px" show-overflow-tooltip />
                    <el-table-column :label="$t('template.type')" prop="type" width="150px">
                        <template #default="{ row }">
                            <el-tag>
                                {{ row.type === 'single' ? $t('template.single') : $t('template.multi') }}
                            </el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('template.variables')" width="100px">
                        <template #default="{ row }">
                            {{ getVariableCount(row.variables) }}
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('template.remark')"
                        prop="remark"
                        min-width="150px"
                        show-overflow-tooltip
                    />
                    <el-table-column
                        prop="createdAt"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                        width="180px"
                    />
                    <fu-table-operations
                        :ellipsis="3"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        fixed="right"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>
        <OpDialog ref="opRef" @search="search" />
        <TemplateOperate ref="operateRef" @search="search" />
        <OutputCreate ref="outputCreateRef" @search="search" />
        <OutputList ref="outputListRef" />
    </div>
</template>

<script lang="ts" setup>
import { Website } from '@/api/interface/website';
import { deleteTemplate, searchTemplates } from '@/api/modules/website';
import TemplateOperate from '@/views/website/template/operate/index.vue';
import OutputCreate from '@/views/website/template/output/create.vue';
import OutputList from '@/views/website/template/output/index.vue';
import { dateFormat } from '@/utils/date';
import i18n from '@/lang';
import { reactive, ref, onMounted } from 'vue';

const loading = ref(false);
const data = ref<Website.Template[]>([]);
const searchName = ref('');
const opRef = ref();
const operateRef = ref();
const outputCreateRef = ref();
const outputListRef = ref();

const routerButton = [
    {
        label: i18n.global.t('menu.template'),
        path: '/websites/templates',
    },
];

const paginationConfig = reactive({
    cacheSizeKey: 'website-template-page-size',
    currentPage: 1,
    pageSize: 20,
    total: 0,
});

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: Website.Template) => {
            operateRef.value.acceptParams(row.id);
        },
    },
    {
        label: i18n.global.t('template.generate'),
        click: (row: Website.Template) => {
            outputCreateRef.value.acceptParams(row.id);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Website.Template) => {
            deleteTemplateRow(row);
        },
    },
];

const getVariableCount = (variables: string) => {
    try {
        const arr = JSON.parse(variables || '[]');
        return Array.isArray(arr) ? arr.length : 0;
    } catch {
        return 0;
    }
};

const search = async () => {
    loading.value = true;
    try {
        const res = await searchTemplates({
            page: paginationConfig.currentPage,
            pageSize: paginationConfig.pageSize,
            name: searchName.value,
            type: '',
        });
        data.value = res.data.items || [];
        paginationConfig.total = res.data.total;
    } finally {
        loading.value = false;
    }
};

const openCreate = () => {
    operateRef.value.acceptParams();
};

const openOutputList = () => {
    outputListRef.value.acceptParams();
};

const deleteTemplateRow = (row: Website.Template) => {
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: [row.name],
        msg: i18n.global.t('template.confirmDelete'),
        api: deleteTemplate,
        params: { id: row.id },
    });
};

onMounted(() => {
    search();
});
</script>
