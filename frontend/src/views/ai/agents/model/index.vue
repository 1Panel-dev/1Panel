<template>
    <div>
        <RouterMenu />
        <LayoutContent>
            <template #leftToolBar>
                <el-button type="primary" @click="openCreate">{{ $t('aiTools.agents.createModelAccount') }}</el-button>
            </template>
            <template #rightToolBar>
                <TableSearch v-model:searchName="searchName" @search="search" />
                <TableRefresh @search="search" />
            </template>
            <template #main>
                <ComplexTable :data="items" :pagination-config="paginationConfig" @search="search">
                    <el-table-column :label="$t('commons.table.name')" prop="name" min-width="200" />
                    <el-table-column :label="$t('aiTools.agents.provider')" prop="provider" width="120">
                        <template #default="{ row }">
                            {{ row.providerName || row.provider }}
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('aiTools.agents.baseUrl')" prop="baseUrl" min-width="200" />
                    <el-table-column :label="$t('aiTools.agents.apiKey')" prop="apiKey" min-width="160">
                        <template #default="{ row }">
                            {{ maskKey(row.apiKey) }}
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('aiTools.agents.verified')" prop="verified" width="120">
                        <template #default="{ row }">
                            <el-tag :type="row.verified ? 'success' : 'info'">
                                {{ row.verified ? 'OK' : 'N/A' }}
                            </el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('commons.table.date')"
                        prop="createdAt"
                        width="180"
                        :formatter="dateFormat"
                    />
                    <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fixed="right" />
                </ComplexTable>
            </template>
        </LayoutContent>
        <AddDialog ref="addRef" @search="search" />
    </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { deleteAgentAccount, pageAgentAccounts } from '@/api/modules/ai';
import { AI } from '@/api/interface/ai';
import RouterMenu from '@/views/ai/agents/index.vue';
import AddDialog from '@/views/ai/agents/model/add/index.vue';
import { ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import { dateFormat } from '@/utils/util';

const items = ref<AI.AgentAccountItem[]>([]);
const addRef = ref();
const searchName = ref('');

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: AI.AgentAccountItem) => onEdit(row),
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: AI.AgentAccountItem) => onDelete(row),
    },
];

const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});

const search = async () => {
    const req: AI.AgentAccountSearch = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        provider: '',
        name: searchName.value || '',
    };
    const res = await pageAgentAccounts(req);
    items.value = res.data.items || [];
    paginationConfig.total = res.data.total || 0;
};

const openCreate = () => {
    if (addRef.value?.open) {
        addRef.value.open({
            provider: 'deepseek',
        });
    }
};

const onEdit = (row: AI.AgentAccountItem) => {
    if (addRef.value?.open) {
        addRef.value.open({
            id: row.id,
            provider: row.provider,
            name: row.name,
            baseURL: row.baseUrl,
            apiKey: row.apiKey,
            remark: row.remark,
        });
    }
};

const maskKey = (value: string) => {
    if (!value) {
        return '';
    }
    if (value.length <= 6) {
        return value;
    }
    return `${value.slice(0, 3)}****${value.slice(-3)}`;
};

const onDelete = async (row: AI.AgentAccountItem) => {
    await ElMessageBox.confirm(
        i18n.global.t('commons.msg.delete', [row.name]),
        i18n.global.t('commons.button.delete'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        },
    );
    await deleteAgentAccount({ id: row.id });
    await search();
};

onMounted(async () => {
    await search();
});
</script>
