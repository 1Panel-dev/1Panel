<template>
    <div v-loading="loading">
        <docker-status
            v-model:isActive="isActive"
            v-model:isExist="isExist"
            v-model:loading="loading"
            @search="search"
        />

        <LayoutContent v-if="isExist" :title="$t('container.composeTemplate', 2)" :class="{ mask: !isActive }">
            <template #leftToolBar>
                <el-button type="primary" @click="onOpenDialog('create')">
                    {{ $t('container.createComposeTemplate') }}
                </el-button>
                <el-button type="primary" plain :disabled="selects.length === 0" @click="onBatchDelete(null)">
                    {{ $t('commons.button.delete') }}
                </el-button>

                <el-button-group>
                    <el-button @click="onImport">
                        {{ $t('commons.button.import') }}
                    </el-button>
                    <el-button :disabled="selects.length === 0" @click="onExport">
                        {{ $t('commons.button.export') }}
                    </el-button>
                </el-button-group>
            </template>
            <template #rightToolBar>
                <TableSearch @search="search()" v-model:searchName="searchName" />
                <TableRefresh @search="search()" />
                <TableSetting title="template-refresh" @search="search()" />
            </template>
            <template #main>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    :data="data"
                    @search="search"
                    :heightDiff="300"
                >
                    <el-table-column type="selection" fix />
                    <el-table-column
                        :label="$t('commons.table.name')"
                        min-width="100"
                        prop="name"
                        sortable
                        fix
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <el-text type="primary" class="cursor-pointer" @click="onOpenDetail(row)">
                                {{ row.name }}
                            </el-text>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.description')" prop="description" min-width="200" fix />
                    <el-table-column :label="$t('commons.table.createdAt')" min-width="80" fix>
                        <template #default="{ row }">
                            {{ dateFormat(0, 0, row.createdAt) }}
                        </template>
                    </el-table-column>
                    <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" />
                </ComplexTable>
            </template>
        </LayoutContent>

        <OpDialog ref="opRef" @search="search" />
        <OpDialog ref="opRef" @search="search" />
        <ImportDialog @search="search" ref="dialogImportRef" />
        <DetailDialog ref="detailRef" />
        <OperatorDialog @search="search" ref="dialogRef" />
    </div>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { dateFormat, downloadWithContent, getCurrentDateFormatted } from '@/utils/util';
import { Container } from '@/api/interface/container';
import DetailDialog from '@/views/container/template/detail/index.vue';
import ImportDialog from '@/views/container/template/import/index.vue';
import OperatorDialog from '@/views/container/template/operator/index.vue';
import { deleteComposeTemplate, searchComposeTemplate } from '@/api/modules/container';
import DockerStatus from '@/views/container/docker-status/index.vue';
import i18n from '@/lang';

const loading = ref();
const data = ref();
const selects = ref<any>([]);

const dialogImportRef = ref();
const detailRef = ref();
const dialogRef = ref();
const opRef = ref();
const isActive = ref(false);
const isExist = ref(false);

const paginationConfig = reactive({
    cacheSizeKey: 'compose-template-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('compose-template-page-size')) || 20,
    total: 0,
});
const searchName = ref();

const search = async () => {
    if (!isActive.value || !isExist.value) {
        return;
    }
    let params = {
        info: searchName.value,
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    loading.value = true;
    await searchComposeTemplate(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const onOpenDetail = async (row: Container.TemplateInfo) => {
    detailRef.value.acceptParams({ content: row.content });
};

const onImport = () => {
    dialogImportRef.value.acceptParams();
};

const onExport = () => {
    ElMessageBox.confirm(
        i18n.global.t('container.exportHelper', [selects.value.length]),
        i18n.global.t('commons.button.export'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        },
    ).then(async () => {
        const exportData = data.value.map((item: Container.TemplateInfo) => ({
            name: item.name,
            description: item.description,
            content: item.content,
        }));
        const content = JSON.stringify(exportData, null, 2);
        const fileName = `1panel-docker-compose-template-${getCurrentDateFormatted()}.json`;
        downloadWithContent(content, fileName);
    });
};

const onOpenDialog = async (
    title: string,
    rowData: Partial<Container.TemplateInfo> = {
        name: '',
        content: '',
        description: '',
    },
) => {
    let params = {
        title,
        rowData: { ...rowData },
    };
    dialogRef.value!.acceptParams(params);
};

const onBatchDelete = async (row: Container.RepoInfo | null) => {
    let ids = [];
    let names = [];
    if (row) {
        names.push(row.name);
        ids.push(row.id);
    } else {
        selects.value.forEach((item: Container.RepoInfo) => {
            names.push(item.name);
            ids.push(item.id);
        });
    }
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('container.composeTemplate'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: deleteComposeTemplate,
        params: { ids: ids },
    });
};

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        disabled: (row: Container.RepoInfo) => {
            return row.downloadUrl === 'docker.io';
        },
        click: (row: Container.RepoInfo) => {
            onOpenDialog('edit', row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        disabled: (row: Container.RepoInfo) => {
            return row.downloadUrl === 'docker.io';
        },
        click: (row: Container.RepoInfo) => {
            onBatchDelete(row);
        },
    },
];
</script>
