<template>
    <div>
        <LayoutContent>
            <template #leftToolBar>
                <el-button v-permission type="primary" @click="openCreate">
                    {{ $t('commons.button.create') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <el-select
                    v-model="apiType"
                    clearable
                    filterable
                    class="api-type-filter"
                    :placeholder="'API ' + $t('commons.table.type')"
                    @change="handleAPITypeChange"
                >
                    <el-option v-for="item in apiTypeOptions" :key="item" :label="item" :value="item">
                        <ApiTypeTag :api-type="item" />
                    </el-option>
                </el-select>
                <TableSearch v-model:searchName="searchName" @search="search" />
                <TableRefresh @search="search" />
            </template>
            <template #main>
                <ComplexTable :data="items" :pagination-config="paginationConfig" @search="search">
                    <el-table-column :label="$t('commons.table.name')" prop="name" min-width="200" />
                    <el-table-column :label="$t('aiTools.agents.provider')" prop="provider" width="200">
                        <template #default="{ row }">
                            <ProviderLabel :provider="row.provider" :display-name="row.providerName" />
                        </template>
                    </el-table-column>
                    <el-table-column :label="'API ' + $t('commons.table.type')" prop="apiType" min-width="170">
                        <template #default="{ row }">
                            <ApiTypeTag :api-type="row.apiType" />
                        </template>
                    </el-table-column>
                    <el-table-column label="Base URL" prop="baseUrl" min-width="200" />
                    <el-table-column label="API Key" prop="apiKey" min-width="160">
                        <template #default="{ row }">
                            {{ maskKey(row.apiKey) }}
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('aiTools.agents.verified')" prop="verified" width="120">
                        <template #default="{ row }">
                            <el-tag :type="verificationTagType(row)">
                                {{ verificationLabel(row) }}
                            </el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('commons.table.date')"
                        prop="createdAt"
                        width="180"
                        :formatter="dateFormat"
                    />
                    <fu-table-operations
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        fixed="right"
                        width="180"
                    />
                </ComplexTable>
            </template>
        </LayoutContent>
        <AddDialog ref="addRef" @search="search" />
        <ModelPoolDialog ref="modelPoolRef" @updated="search" />
    </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue';
import { deleteAgentAccount, getAgentProviders, pageAgentAccounts } from '@/api/modules/ai';
import { AI } from '@/api/interface/ai';
import AddDialog from '@/views/ai/agents/model/add/index.vue';
import ModelPoolDialog from '@/views/ai/agents/model/pool/index.vue';
import ProviderLabel from '@/components/agent-provider-label/index.vue';
import ApiTypeTag from '@/components/api-type-tag/index.vue';
import { ElMessageBox } from 'element-plus';
import i18n from '@/lang';
import { dateFormat } from '@/utils/date';
import { isAgentAccountVerificationSkipped } from '@/utils/agent';

const items = ref<AI.AgentAccountItem[]>([]);
const addRef = ref();
const modelPoolRef = ref();
const searchName = ref('');
const apiType = ref('');
const apiTypeOptions = ref<string[]>([]);

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        permission: true,
        click: (row: AI.AgentAccountItem) => onEdit(row),
    },
    {
        label: i18n.global.t('aiTools.agents.modelPool'),
        click: (row: AI.AgentAccountItem) => onManageModelPool(row),
    },
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
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
        apiType: apiType.value,
        name: searchName.value || '',
    };
    const res = await pageAgentAccounts(req);
    items.value = res.data.items || [];
    paginationConfig.total = res.data.total || 0;
};

const handleAPITypeChange = () => {
    paginationConfig.currentPage = 1;
    search();
};

const loadAPITypes = async () => {
    const res = await getAgentProviders();
    apiTypeOptions.value = Array.from(
        new Set((res.data || []).flatMap((provider) => provider.apiTypes.map((item) => item.apiType))),
    ).sort();
};

const openCreate = () => {
    if (addRef.value?.open) {
        addRef.value.open();
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
            rememberApiKey: row.rememberApiKey,
            apiType: row.apiType,
            authMode: row.authMode,
            verifyModel: row.verifyModel,
            models: row.models,
            remark: row.remark,
        });
    }
};

const onManageModelPool = (row: AI.AgentAccountItem) => {
    if (modelPoolRef.value?.open) {
        modelPoolRef.value.open(row);
    }
};

const verificationLabel = (row: AI.AgentAccountItem) => {
    if (isAgentAccountVerificationSkipped(row.provider)) {
        return i18n.global.t('aiTools.agents.verifySkipped');
    }
    return row.verified ? 'OK' : 'N/A';
};

const verificationTagType = (row: AI.AgentAccountItem) => {
    if (isAgentAccountVerificationSkipped(row.provider)) {
        return 'info';
    }
    return row.verified ? 'success' : 'info';
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
    await Promise.all([search(), loadAPITypes()]);
});
</script>

<style scoped>
.api-type-filter {
    width: 220px;
}
</style>
