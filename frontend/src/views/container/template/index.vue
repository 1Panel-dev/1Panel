<template>
    <div v-loading="loading">
        <el-card v-if="dockerStatus != 'Running'" class="mask-prompt">
            <span>{{ $t('container.serviceUnavailable') }}</span>
            <el-button type="primary" link class="bt" @click="goSetting">【 {{ $t('container.setting') }} 】</el-button>
            <span>{{ $t('container.startIn') }}</span>
        </el-card>

        <LayoutContent :title="$t('container.composeTemplate')" :class="{ mask: dockerStatus != 'Running' }">
            <template #toolbar>
                <el-row>
                    <el-col :span="16">
                        <el-button type="primary" @click="onOpenDialog('create')">
                            {{ $t('container.createComposeTemplate') }}
                        </el-button>
                        <el-button type="primary" plain :disabled="selects.length === 0" @click="onBatchDelete(null)">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </el-col>
                    <el-col :span="8">
                        <TableSetting @search="search()" />
                        <div class="search-button">
                            <el-input
                                v-model="searchName"
                                clearable
                                @clear="search()"
                                suffix-icon="Search"
                                @keyup.enter="search()"
                                @change="search()"
                                :placeholder="$t('commons.button.search')"
                            ></el-input>
                        </div>
                    </el-col>
                </el-row>
            </template>
            <template #main>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    :data="data"
                    @search="search"
                >
                    <el-table-column type="selection" fix />
                    <el-table-column :label="$t('commons.table.name')" min-width="100" prop="name" fix>
                        <template #default="{ row }">
                            <Tooltip @click="onOpenDetail(row)" :text="row.name" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('container.description')" prop="description" min-width="200" fix />
                    <el-table-column :label="$t('commons.table.createdAt')" min-width="80" fix>
                        <template #default="{ row }">
                            {{ dateFormatSimple(row.createdAt) }}
                        </template>
                    </el-table-column>
                    <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" />
                </ComplexTable>
            </template>
        </LayoutContent>

        <DetailDialog ref="detailRef" />
        <OperatorDialog @search="search" ref="dialogRef" />
    </div>
</template>

<script lang="ts" setup>
import Tooltip from '@/components/tooltip/index.vue';
import TableSetting from '@/components/table-setting/index.vue';
import { reactive, onMounted, ref } from 'vue';
import { dateFormatSimple } from '@/utils/util';
import { Container } from '@/api/interface/container';
import DetailDialog from '@/views/container/template/detail/index.vue';
import OperatorDialog from '@/views/container/template/operator/index.vue';
import { deleteComposeTemplate, loadDockerStatus, searchComposeTemplate } from '@/api/modules/container';
import { useDeleteData } from '@/hooks/use-delete-data';
import i18n from '@/lang';
import router from '@/routers';

const loading = ref();
const data = ref();
const selects = ref<any>([]);

const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const searchName = ref();

const detailRef = ref();

const dockerStatus = ref('Running');
const loadStatus = async () => {
    loading.value = true;
    await loadDockerStatus()
        .then((res) => {
            loading.value = false;
            dockerStatus.value = res.data;
            if (dockerStatus.value === 'Running') {
                search();
            }
        })
        .catch(() => {
            dockerStatus.value = 'Failed';
            loading.value = false;
        });
};
const goSetting = async () => {
    router.push({ name: 'ContainerSetting' });
};

const search = async () => {
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

const dialogRef = ref();
const onOpenDialog = async (
    title: string,
    rowData: Partial<Container.TemplateInfo> = {
        name: '',
        description: '',
        path: '',
    },
) => {
    let params = {
        title,
        rowData: { ...rowData },
    };
    dialogRef.value!.acceptParams(params);
};

const onBatchDelete = async (row: Container.RepoInfo | null) => {
    let ids: Array<number> = [];
    if (row) {
        ids.push(row.id);
    } else {
        selects.value.forEach((item: Container.RepoInfo) => {
            ids.push(item.id);
        });
    }
    await useDeleteData(deleteComposeTemplate, { ids: ids }, 'commons.msg.delete');
    search();
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

onMounted(() => {
    loadStatus();
});
</script>
