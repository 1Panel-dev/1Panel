<template>
    <DrawerPro
        v-model="visible"
        :header="$t('panelMCP.invocations')"
        :resource="credential?.name || ''"
        size="min(1100px, 100vw)"
        @close="close"
    >
        <div class="mb-4 flex flex-wrap items-center justify-between gap-2">
            <el-select
                v-model="result"
                class="!w-full sm:!w-80"
                :placeholder="$t('panelMCP.allResults')"
                clearable
                @change="resetSearch"
            >
                <template #prefix>{{ $t('commons.table.status') }}</template>
                <el-option v-for="value in results" :key="value" :value="value" :label="resultLabel(value)" />
            </el-select>
            <el-button class="ml-auto" icon="Refresh" :loading="loading" @click="search">
                {{ $t('commons.button.refresh') }}
            </el-button>
        </div>
        <div class="w-full min-w-0 overflow-x-auto" v-loading="loading">
            <ComplexTable
                :data="items"
                :pagination-config="pagination"
                :page-sizes="[10, 20, 50, 100]"
                :height="500"
                @search="search"
            >
                <el-table-column
                    :label="$t('commons.table.date')"
                    prop="startedAt"
                    :formatter="dateFormat"
                    min-width="170"
                />
                <el-table-column
                    :label="$t('panelMCP.credentials')"
                    prop="credentialName"
                    min-width="140"
                    show-overflow-tooltip
                />
                <el-table-column :label="$t('panelMCP.tools')" prop="toolId" min-width="170" show-overflow-tooltip>
                    <template #default="{ row }">
                        <span class="font-mono text-xs">{{ row.toolId }}</span>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('panelMCP.targetNode')" prop="nodeId" min-width="120" />
                <el-table-column :label="$t('commons.table.status')" min-width="95">
                    <template #default="{ row }">
                        <el-tag
                            size="small"
                            round
                            effect="light"
                            class="!h-auto min-h-5 max-w-full"
                            :type="
                                row.status === 'success'
                                    ? 'success'
                                    : row.status === 'denied' || row.status === 'running'
                                      ? 'warning'
                                      : 'danger'
                            "
                        >
                            <span class="whitespace-normal break-words">{{ resultLabel(row.status) }}</span>
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('panelMCP.resultCode')" prop="code" min-width="160" show-overflow-tooltip>
                    <template #default="{ row }">
                        <span class="font-mono text-xs">{{ row.code }}</span>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('panelMCP.duration')" prop="durationMs" min-width="120" align="right" />
                <el-table-column
                    :label="$t('panelMCP.requestId')"
                    prop="requestId"
                    min-width="220"
                    show-overflow-tooltip
                >
                    <template #default="{ row }">
                        <span class="font-mono text-xs">{{ row.requestId }}</span>
                    </template>
                </el-table-column>
            </ComplexTable>
        </div>
        <template #footer>
            <div class="flex flex-wrap items-center justify-end gap-2 [&>.el-button]:!ml-0">
                <el-button @click="close">{{ $t('commons.button.close') }}</el-button>
            </div>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onDeactivated, reactive, ref } from 'vue';
import type { PanelMCP } from '@/api/interface/panel-mcp';
import { searchPanelMCPInvocations } from '@/api/modules/panel-mcp';
import { dateFormat } from '@/utils/date';
import i18n from '@/lang';

const visible = ref(false);
const loading = ref(false);
const credential = ref<PanelMCP.Credential>();
const items = ref<PanelMCP.Invocation[]>([]);
const result = ref('');
const results = ['running', 'success', 'denied', 'failed'];
const pagination = reactive({ currentPage: 1, pageSize: 20, total: 0, cacheSizeKey: 'panel-mcp-invocations-page' });
let version = 0;
const resultLabel = (value: string) =>
    i18n.global.t(
        `panelMCP.${{ running: 'resultRunning', success: 'resultSuccess', denied: 'resultDenied', failed: 'resultFailed' }[value] || 'resultFailed'}`,
    );
const search = async () => {
    if (!visible.value) return;
    const ticket = ++version;
    loading.value = true;
    try {
        const response = await searchPanelMCPInvocations({
            page: pagination.currentPage,
            pageSize: pagination.pageSize,
            ...(credential.value ? { credentialId: credential.value.id } : {}),
            ...(result.value ? { status: result.value } : {}),
        });
        if (ticket !== version || !visible.value) return;
        items.value = response.data.items || [];
        pagination.total = response.data.total;
    } catch {
        if (ticket === version) {
            items.value = [];
            pagination.total = 0;
        }
    } finally {
        if (ticket === version) loading.value = false;
    }
};
const resetSearch = () => {
    pagination.currentPage = 1;
    void search();
};
const open = (item?: PanelMCP.Credential) => {
    version++;
    credential.value = item;
    result.value = '';
    items.value = [];
    pagination.currentPage = 1;
    visible.value = true;
    void search();
};
const close = () => {
    version++;
    visible.value = false;
    loading.value = false;
    items.value = [];
};
onDeactivated(close);
onBeforeUnmount(close);
defineExpose({ open });
</script>
