<template>
    <div>
        <BackButton name="WebsiteTemplate" :header="$t('template.outputList')" />
        <LayoutContent :title="$t('template.outputList')">
            <template #leftToolBar>
                <el-button v-permission type="primary" @click="openCreate()">
                    {{ $t('template.outputCreate') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <TableRefresh @search="search()" />
            </template>
            <template #main>
                <ComplexTable
                    :data="data"
                    :pagination-config="paginationConfig"
                    @search="search()"
                    v-loading="loading"
                >
                    <el-table-column
                        :label="$t('template.outputName')"
                        prop="name"
                        min-width="120px"
                        show-overflow-tooltip
                    />
                    <el-table-column
                        :label="$t('template.name')"
                        prop="templateName"
                        min-width="120px"
                        show-overflow-tooltip
                    />
                    <el-table-column :label="$t('template.type')" width="150px">
                        <template #default="{ row }">
                            <el-tag>
                                {{ row.templateType === 'single' ? $t('template.single') : $t('template.multi') }}
                            </el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('template.filePath')"
                        prop="outputPath"
                        min-width="200px"
                        show-overflow-tooltip
                    />
                    <el-table-column
                        prop="createdAt"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                        width="180px"
                    />
                    <fu-table-operations
                        :ellipsis="1"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        fixed="right"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>
        <OpDialog ref="opRef" @search="search" />
        <OutputCreate ref="outputCreateRef" @search="search" />
    </div>
</template>

<script lang="ts" setup>
import { Website } from '@/api/interface/website';
import { deleteTemplateOutput, searchTemplateOutputs } from '@/api/modules/website';
import OutputCreate from '@/views/website/template/output/create.vue';
import BackButton from '@/components/back-button/index.vue';
import { dateFormat } from '@/utils/date';
import i18n from '@/lang';
import { reactive, ref, onMounted } from 'vue';

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

onMounted(() => {
    search();
});
</script>
