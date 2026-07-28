<template>
    <el-radio-group v-model="mode" class="view-switch">
        <el-radio-button label="installed">{{ t('aiTools.agents.pluginsInstalled') }}</el-radio-button>
        <el-radio-button label="market">{{ t('aiTools.agents.pluginsMarket') }}</el-radio-button>
    </el-radio-group>

    <div v-loading="loading" class="plugin-content">
        <div class="toolbar">
            <template v-if="mode === 'installed'">
                <el-input
                    v-model="keyword"
                    :placeholder="t('aiTools.agents.pluginSearchPlaceholder')"
                    clearable
                    class="search-input"
                >
                    <template #prefix>
                        <el-icon><Search /></el-icon>
                    </template>
                </el-input>
                <el-select v-model="origin" class="filter-select">
                    <el-option :label="t('aiTools.agents.pluginOriginAll')" value="" />
                    <el-option :label="t('aiTools.agents.pluginOriginBundled')" value="bundled" />
                    <el-option :label="t('aiTools.agents.pluginOriginExternal')" value="external" />
                </el-select>
                <el-select v-model="status" class="filter-select">
                    <el-option :label="t('aiTools.agents.pluginStatusAll')" value="" />
                    <el-option :label="t('aiTools.agents.pluginEnabled')" value="enabled" />
                    <el-option :label="t('aiTools.agents.pluginDisabled')" value="disabled" />
                </el-select>
                <el-button :loading="loading" @click="loadPlugins">
                    <el-icon><Refresh /></el-icon>
                </el-button>
            </template>
            <template v-else>
                <el-input
                    v-model="marketKeyword"
                    :placeholder="t('aiTools.agents.pluginSearchPlaceholder')"
                    clearable
                    class="search-input"
                    @keyup.enter="searchMarket"
                >
                    <template #prefix>
                        <el-icon><Search /></el-icon>
                    </template>
                </el-input>
                <el-button type="primary" :loading="searching" @click="searchMarket">
                    {{ t('commons.button.search') }}
                </el-button>
            </template>
        </div>

        <el-table v-if="mode === 'installed'" :data="pagedPlugins">
            <el-table-column :label="t('commons.table.name')" min-width="210">
                <template #default="{ row }">
                    <div class="plugin-name">{{ row.name || row.id }}</div>
                    <div class="secondary">{{ row.id }}</div>
                </template>
            </el-table-column>
            <el-table-column :label="t('app.version')" min-width="110">
                <template #default="{ row }">{{ row.version || '-' }}</template>
            </el-table-column>
            <el-table-column :label="t('aiTools.agents.pluginOrigin')" min-width="110">
                <template #default="{ row }">
                    <el-tag size="small" effect="plain">
                        {{
                            row.origin === 'bundled'
                                ? t('aiTools.agents.pluginOriginBundled')
                                : t('aiTools.agents.pluginOriginExternal')
                        }}
                    </el-tag>
                </template>
            </el-table-column>
            <el-table-column :label="t('commons.table.status')" min-width="120">
                <template #default="{ row }">
                    <el-tag :type="row.enabled ? 'success' : 'info'" size="small">
                        {{ row.enabled ? t('aiTools.agents.pluginEnabled') : t('aiTools.agents.pluginDisabled') }}
                    </el-tag>
                </template>
            </el-table-column>
            <el-table-column :label="t('commons.table.operate')" fixed="right" min-width="210">
                <template #default="{ row }">
                    <el-button
                        v-permission
                        type="primary"
                        link
                        :disabled="operating !== ''"
                        @click="operate(row, row.enabled ? 'disable' : 'enable')"
                    >
                        {{ row.enabled ? t('aiTools.agents.pluginDisable') : t('aiTools.agents.pluginEnable') }}
                    </el-button>
                    <el-button
                        v-if="row.origin !== 'bundled'"
                        v-permission
                        type="primary"
                        link
                        :disabled="operating !== ''"
                        @click="operate(row, 'update')"
                    >
                        {{ t('commons.button.upgrade') }}
                    </el-button>
                    <el-button
                        v-if="row.origin !== 'bundled'"
                        v-permission
                        type="danger"
                        link
                        :disabled="operating !== ''"
                        @click="operate(row, 'uninstall')"
                    >
                        {{ t('commons.button.uninstall') }}
                    </el-button>
                </template>
            </el-table-column>
        </el-table>

        <el-table v-else :data="marketResults">
            <el-table-column :label="t('commons.table.name')" min-width="260">
                <template #default="{ row }">
                    <div class="plugin-name">{{ row.name || row.package }}</div>
                    <div class="secondary">{{ row.package }}</div>
                </template>
            </el-table-column>
            <el-table-column :label="t('app.version')" prop="version" min-width="110" />
            <el-table-column :label="t('aiTools.agents.pluginCategory')" min-width="150">
                <template #default="{ row }">
                    <el-space wrap>
                        <el-tag v-for="item in row.categories" :key="item" size="small" type="info">
                            {{ item }}
                        </el-tag>
                    </el-space>
                </template>
            </el-table-column>
            <el-table-column :label="t('aiTools.agents.pluginVerification')" min-width="130">
                <template #default="{ row }">
                    <el-tag :type="row.official ? 'success' : 'info'" size="small" effect="plain">
                        {{ row.official ? t('aiTools.agents.pluginOfficial') : t('aiTools.agents.pluginCommunity') }}
                    </el-tag>
                    <div class="secondary">{{ row.verificationTier || '-' }}</div>
                </template>
            </el-table-column>
            <el-table-column :label="t('aiTools.agents.pluginDownloads')" prop="downloads" min-width="110" />
            <el-table-column :label="t('commons.table.operate')" fixed="right" width="100">
                <template #default="{ row }">
                    <el-button v-permission type="primary" link :disabled="operating !== ''" @click="install(row)">
                        {{ t('commons.button.install') }}
                    </el-button>
                </template>
            </el-table-column>
        </el-table>

        <el-empty
            v-if="mode === 'installed' && !loading && filteredPlugins.length === 0"
            :description="t('aiTools.agents.pluginListEmpty')"
        />
        <el-empty
            v-if="mode === 'market' && !searching && marketResults.length === 0"
            :description="
                marketSearched ? t('aiTools.agents.pluginSearchEmpty') : t('aiTools.agents.pluginsMarketHint')
            "
        />

        <div v-if="mode === 'installed' && filteredPlugins.length > pageSize" class="pagination">
            <el-pagination
                v-model:current-page="page"
                :page-size="pageSize"
                :total="filteredPlugins.length"
                layout="total, prev, pager, next"
            />
        </div>
    </div>

    <TaskLog ref="taskLogRef" @close="loadPlugins" />
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { Refresh, Search } from '@element-plus/icons-vue';
import { ElMessageBox } from 'element-plus';
import { useI18n } from 'vue-i18n';
import { AI } from '@/api/interface/ai';
import { installAgentMarketPlugin, listAgentPlugins, operateAgentPlugin, searchAgentPlugins } from '@/api/modules/ai';
import { newUUID } from '@/utils/id';
import TaskLog from '@/components/log/task/index.vue';

type PluginMode = 'installed' | 'market';
type PluginOperate = AI.AgentPluginOperateReq['operate'];

const { t } = useI18n();
const mode = ref<PluginMode>('installed');
const agentId = ref(0);
const loading = ref(false);
const searching = ref(false);
const operating = ref('');
const keyword = ref('');
const marketKeyword = ref('');
const origin = ref('');
const status = ref('');
const page = ref(1);
const pageSize = 10;
const plugins = ref<AI.AgentPluginItem[]>([]);
const marketResults = ref<AI.AgentPluginSearchItem[]>([]);
const marketSearched = ref(false);
const taskLogRef = ref<InstanceType<typeof TaskLog>>();

const filteredPlugins = computed(() => {
    const search = keyword.value.trim().toLowerCase();
    return plugins.value.filter((plugin) => {
        const matchesKeyword =
            !search || plugin.name.toLowerCase().includes(search) || plugin.id.toLowerCase().includes(search);
        const matchesOrigin =
            !origin.value ||
            (origin.value === 'external' ? plugin.origin !== 'bundled' : plugin.origin === origin.value);
        const matchesStatus = !status.value || (status.value === 'enabled' ? plugin.enabled : !plugin.enabled);
        return matchesKeyword && matchesOrigin && matchesStatus;
    });
});

const pagedPlugins = computed(() => {
    const start = (page.value - 1) * pageSize;
    return filteredPlugins.value.slice(start, start + pageSize);
});

watch([keyword, origin, status], () => {
    page.value = 1;
});

async function loadPlugins() {
    if (!agentId.value) {
        return;
    }
    loading.value = true;
    try {
        const res = await listAgentPlugins({ agentId: agentId.value });
        plugins.value = res.data || [];
    } finally {
        loading.value = false;
        operating.value = '';
    }
}

async function searchMarket() {
    const search = marketKeyword.value.trim();
    if (!search) {
        return;
    }
    searching.value = true;
    try {
        const res = await searchAgentPlugins({ agentId: agentId.value, keyword: search, limit: 20 });
        marketResults.value = res.data || [];
        marketSearched.value = true;
    } finally {
        searching.value = false;
    }
}

async function install(plugin: AI.AgentPluginSearchItem) {
    await ElMessageBox.confirm(
        t('aiTools.agents.pluginInstallConfirm', [plugin.name || plugin.package]),
        t('commons.button.install'),
        { type: 'warning' },
    );
    const taskID = newUUID();
    operating.value = plugin.package;
    try {
        await installAgentMarketPlugin({
            agentId: agentId.value,
            package: plugin.package,
            version: plugin.version,
            taskID,
        });
        taskLogRef.value?.openWithTaskID(taskID);
    } finally {
        operating.value = '';
    }
}

async function operate(plugin: AI.AgentPluginItem, action: PluginOperate) {
    if (action === 'disable' || action === 'uninstall') {
        const key =
            action === 'disable' ? 'aiTools.agents.pluginDisableConfirm' : 'aiTools.agents.pluginUninstallConfirm';
        await ElMessageBox.confirm(t(key, [plugin.name || plugin.id]), t('commons.button.confirm'), {
            type: 'warning',
        });
    }
    const taskID = newUUID();
    operating.value = plugin.id;
    try {
        await operateAgentPlugin({
            agentId: agentId.value,
            pluginId: plugin.id,
            operate: action,
            taskID,
        });
        taskLogRef.value?.openWithTaskID(taskID);
    } finally {
        operating.value = '';
    }
}

async function load(id: number) {
    agentId.value = id;
    mode.value = 'installed';
    keyword.value = '';
    marketKeyword.value = '';
    origin.value = '';
    status.value = '';
    page.value = 1;
    marketResults.value = [];
    marketSearched.value = false;
    await loadPlugins();
}

defineExpose({ load });
</script>

<style scoped lang="scss">
.view-switch {
    padding-top: 1px;
    margin-bottom: 16px;
}

.plugin-content {
    min-height: 390px;
}

.toolbar {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 16px;
}

.search-input {
    width: 280px;
    max-width: 100%;
}

.filter-select {
    width: 150px;
}

.plugin-name {
    color: var(--el-text-color-primary);
    font-weight: 500;
}

.secondary {
    margin-top: 2px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
}

.pagination {
    display: flex;
    justify-content: flex-end;
    margin-top: 16px;
}
</style>
