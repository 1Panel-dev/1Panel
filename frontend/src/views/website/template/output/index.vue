<template>
    <DrawerPro v-model="open" :header="$t('template.outputList')" size="60%" @close="handleClose">
        <template #buttons>
            <el-button type="primary" plain @click="openCreate()">
                {{ $t('template.outputCreate') }}
            </el-button>
        </template>
        <ComplexTable :data="data" :pagination-config="paginationConfig" @search="search()" v-loading="loading">
            <el-table-column :label="$t('template.outputName')" prop="name" min-width="100px" show-overflow-tooltip />
            <el-table-column :label="$t('template.name')" prop="templateName" min-width="100px" show-overflow-tooltip />
            <el-table-column :label="$t('template.type')" width="110px">
                <template #default="{ row }">
                    <el-tag>
                        {{ row.templateType === 'single' ? $t('template.single') : $t('template.multi') }}
                    </el-tag>
                </template>
            </el-table-column>
            <el-table-column
                :label="$t('template.filePath')"
                prop="outputPath"
                min-width="160px"
                show-overflow-tooltip
            />
            <el-table-column prop="createdAt" :label="$t('commons.table.date')" :formatter="dateFormat" width="165px" />
            <fu-table-operations :ellipsis="1" :buttons="buttons" :label="$t('commons.table.operate')" fixed="right" fix />
        </ComplexTable>
        <OpDialog ref="opRef" @search="search" />
        <OutputCreate ref="outputCreateRef" @search="search" />
    </DrawerPro>
</template>

<script lang="ts" setup>
import { Website } from '@/api/interface/website';
import { deleteTemplateOutput, searchTemplateOutputs } from '@/api/modules/website';
import OutputCreate from '@/views/website/template/output/create.vue';
import { dateFormat } from '@/utils/date';
import i18n from '@/lang';
import { reactive, ref } from 'vue';

const open = ref(false);
const loading = ref(false);
const data = ref<Website.TemplateOutputDTO[]>([]);
const opRef = ref();
const outputCreateRef = ref();

const paginationConfig = reactive({
    cacheSizeKey: 'website-template-output-page-size',
    currentPage: 1,
    pageSize: 20,
    total: 0,
});

const buttons = [
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Website.TemplateOutputDTO) => {
            opRef.value.acceptParams({
                title: i18n.global.t('commons.button.delete'),
                names: [row.name],
                msg: i18n.global.t('template.confirmDeleteOutput'),
                api: deleteTemplateOutput,
                params: { id: row.id },
            });
        },
    },
];

const search = async () => {
    loading.value = true;
    try {
        const res = await searchTemplateOutputs({
            page: paginationConfig.currentPage,
            pageSize: paginationConfig.pageSize,
            templateID: 0,
        });
        data.value = res.data.items || [];
        paginationConfig.total = res.data.total;
    } finally {
        loading.value = false;
    }
};

const openCreate = () => {
    outputCreateRef.value.acceptParams();
};

const handleClose = () => {
    open.value = false;
};

const acceptParams = () => {
    open.value = true;
    search();
};

defineExpose({
    acceptParams,
});
</script>
