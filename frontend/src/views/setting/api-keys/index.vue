<template>
    <div>
        <LayoutContent :title="$t('apiKeyManagement.title')" v-loading="loading">
            <template #leftToolBar>
                <div class="flex flex-wrap items-center gap-3">
                    <el-button type="primary" :disabled="used >= limit" @click="editor?.open()">
                        {{ $t('apiKeyManagement.create') }}
                    </el-button>
                    <el-tooltip :content="$t('apiKeyManagement.quota', [used, limit])" placement="top" trigger="click">
                        <el-button link type="info" class="whitespace-nowrap">{{ used }} / {{ limit }}</el-button>
                    </el-tooltip>
                </div>
            </template>
            <template #rightToolBar><TableRefresh @search="search" /></template>
            <template #main>
                <ComplexTable
                    :data="items"
                    :pagination-config="pagination"
                    :scrollbar-always-on="isMobile"
                    @search="search"
                >
                    <el-table-column :label="$t('commons.table.name')" :min-width="isMobile ? 140 : 200" prop="name">
                        <template #default="{ row }">
                            <div class="flex items-center gap-1">
                                <el-button
                                    link
                                    type="primary"
                                    class="min-w-0 max-w-full !whitespace-normal !text-left"
                                    @click="showDetail(row)"
                                >
                                    {{ row.kind === 'legacy' ? $t('apiKeyManagement.legacy') : row.name }}
                                </el-button>
                                <el-tooltip
                                    v-if="row.kind === 'legacy'"
                                    :content="$t('apiKeyManagement.legacyHelp')"
                                    placement="top"
                                    :trigger="isMobile ? 'click' : 'hover'"
                                    :popper-style="{ maxWidth: 'min(360px, calc(100vw - 32px))' }"
                                >
                                    <el-button
                                        link
                                        type="info"
                                        icon="InfoFilled"
                                        class="!ml-0 shrink-0"
                                        :aria-label="$t('apiKeyManagement.legacyHelp')"
                                        @click.stop
                                    />
                                </el-tooltip>
                            </div>
                            <div class="text-xs text-gray-500 break-all">{{ row.keyHint }}</div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.status')" min-width="100" prop="status">
                        <template #default="{ row }">
                            <el-tag
                                :type="
                                    row.status === 'Enable' ? 'success' : row.status === 'Disable' ? 'warning' : 'info'
                                "
                            >
                                {{ $t('apiKeyManagement.status' + row.status) }}
                            </el-tag>
                        </template>
                    </el-table-column>
                    <el-table-column
                        v-if="!isMobile"
                        :label="$t('setting.ipWhiteList')"
                        min-width="180"
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            {{ isAnyAPIKeyIP(row.ipWhiteList) ? $t('apiKeyManagement.anyIP') : row.ipWhiteList || '—' }}
                        </template>
                    </el-table-column>
                    <el-table-column v-if="!isMobile" :label="$t('apiKeyManagement.expiresAt')" min-width="170">
                        <template #default="{ row }">
                            {{
                                row.expiresAt ? new Date(row.expiresAt).toLocaleString() : $t('apiKeyManagement.never')
                            }}
                        </template>
                    </el-table-column>
                    <el-table-column v-if="!isMobile" :label="$t('apiKeyManagement.allowAppBinding')" min-width="150">
                        <template #default="{ row }">
                            {{ $t(row.allowAppBinding ? 'commons.true' : 'commons.false') }}
                        </template>
                    </el-table-column>
                    <fu-table-operations
                        :label="$t('commons.table.operate')"
                        :width="isMobile ? 80 : 200"
                        :fixed="isMobile ? false : 'right'"
                        :buttons="buttons"
                        :ellipsis="isMobile ? 0 : 2"
                        trigger="click"
                    />
                </ComplexTable>
            </template>
        </LayoutContent>
        <APIKeyEditor ref="editor" @changed="search" />
        <DrawerPro
            v-model="detailVisible"
            :header="$t('apiKeyManagement.details')"
            size="min(640px, 100vw)"
            @close="clearLegacySecret"
        >
            <APIKeySummary v-if="detail" :item="detail" />
            <div v-if="detail?.kind === 'legacy' && detail.status !== 'Revoked'" class="mt-5 flex flex-col gap-3">
                <el-alert :title="$t('apiKeyManagement.legacyHelp')" type="info" :closable="false" show-icon />
                <el-space wrap>
                    <el-button type="primary" plain :loading="legacyBusy" @click="revealLegacy">
                        {{ $t('apiKeyManagement.showLegacy') }}
                    </el-button>
                    <el-button :disabled="legacyBusy" @click="resetLegacy">{{ $t('setting.apiKeyReset') }}</el-button>
                </el-space>
                <el-input
                    v-if="legacySecret"
                    :model-value="legacySecret"
                    :aria-label="$t('setting.apiKey')"
                    readonly
                    autocomplete="off"
                >
                    <template #append>
                        <CopyButton :content="legacySecret" :is-icon="false" />
                    </template>
                </el-input>
            </div>
        </DrawerPro>
    </div>
</template>

<script setup lang="ts">
import { computed, onActivated, onBeforeUnmount, onDeactivated, onMounted, reactive, ref } from 'vue';
import { ElMessageBox } from 'element-plus';
import type { APIKey } from '@/api/interface/api-key';
import { revokeAPIKey, searchAPIKeys, setAPIKeyStatus } from '@/api/modules/api-key';
import { generateApiKey, getUserInfo } from '@/api/modules/auth';
import APIKeyEditor from '@/components/api-key-management/editor.vue';
import APIKeySummary from '@/components/api-key-management/summary.vue';
import DrawerPro from '@/components/drawer-pro/index.vue';
import type { FuTableOperationButton } from '@/components/table/shared';
import { isAnyAPIKeyIP } from '@/utils/api-key';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import i18n from '@/lang';
import { useGlobalStore } from '@/composables/useGlobalStore';

const { isMobile } = useGlobalStore();

const items = ref<APIKey.Item[]>([]);
const loading = ref(false);
const used = ref(0);
const limit = ref(20);
const editor = ref<InstanceType<typeof APIKeyEditor>>();
const detail = ref<APIKey.Item>();
const detailVisible = ref(false);
const legacySecret = ref('');
const legacyBusy = ref(false);
const pagination = reactive({ currentPage: 1, pageSize: 20, total: 0 });
let searchVersion = 0;
let secretRequestVersion = 0;
let active = false;

const clearLegacySecret = () => {
    secretRequestVersion++;
    legacySecret.value = '';
    legacyBusy.value = false;
};

const search = async (preserveLegacyRequest?: number) => {
    const ticket = ++searchVersion;
    loading.value = true;
    try {
        const response = await searchAPIKeys({ page: pagination.currentPage, pageSize: pagination.pageSize });
        if (ticket !== searchVersion) return;
        const lastPage = Math.max(1, Math.ceil(response.data.total / pagination.pageSize));
        if (pagination.currentPage > lastPage) {
            pagination.currentPage = lastPage;
            await search(preserveLegacyRequest);
            return;
        }
        items.value = response.data.items || [];
        pagination.total = response.data.total;
        used.value = response.data.used;
        limit.value = response.data.limit;
        if (detail.value) {
            const previousRevision = detail.value.revision;
            detail.value = items.value.find((entry) => entry.id === detail.value?.id);
            if (detail.value?.revision !== previousRevision && preserveLegacyRequest !== secretRequestVersion)
                clearLegacySecret();
            if (!detail.value) detailVisible.value = false;
        }
    } finally {
        if (ticket === searchVersion) loading.value = false;
    }
};
const showDetail = (entry: APIKey.Item) => {
    clearLegacySecret();
    detail.value = entry;
    detailVisible.value = true;
};
const operate = async (entry: APIKey.Item, action: string) => {
    if (action === 'edit') {
        editor.value?.open(entry);
        return;
    }
    if (action !== 'status' && action !== 'revoke') return;
    try {
        await ElMessageBox.confirm(
            i18n.global.t(
                action === 'revoke'
                    ? 'apiKeyManagement.revokeConfirm'
                    : entry.status === 'Enable'
                      ? 'apiKeyManagement.disableConfirm'
                      : 'apiKeyManagement.enableConfirm',
                [entry.name],
            ),
            i18n.global.t('apiKeyManagement.title'),
            {
                type: 'warning',
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
            },
        );
    } catch {
        return;
    }
    loading.value = true;
    try {
        const reference = { id: entry.id, revision: entry.revision };
        const response =
            action === 'revoke'
                ? await revokeAPIKey(reference)
                : await setAPIKeyStatus({ ...reference, status: entry.status === 'Enable' ? 'Disable' : 'Enable' });
        clearLegacySecret();
        if (response.data?.terminalClosePending) MsgWarning(i18n.global.t('apiKeyManagement.closePending'));
        else MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    } finally {
        await search();
    }
};
const buttons = computed<FuTableOperationButton<APIKey.Item>[]>(() => [
    {
        label: i18n.global.t('commons.button.view'),
        show: () => !isMobile.value,
        click: showDetail,
    },
    {
        label: i18n.global.t('commons.button.edit'),
        show: (row) => row.status !== 'Revoked',
        click: (row) => editor.value?.open(row),
    },
    {
        label: i18n.global.t('commons.button.disable'),
        show: (row) => row.status === 'Enable',
        click: (row) => operate(row, 'status'),
    },
    {
        label: i18n.global.t('commons.button.enable'),
        show: (row) => row.status === 'Disable',
        click: (row) => operate(row, 'status'),
    },
    {
        label: i18n.global.t('apiKeyManagement.revoke'),
        show: (row) => row.status !== 'Revoked',
        click: (row) => operate(row, 'revoke'),
    },
]);
const revealLegacy = async () => {
    if (legacyBusy.value || !detailVisible.value || detail.value?.kind !== 'legacy') return;
    const ticket = ++secretRequestVersion;
    legacyBusy.value = true;
    try {
        const response = await getUserInfo();
        if (ticket === secretRequestVersion && detailVisible.value && active && detail.value?.kind === 'legacy')
            legacySecret.value = response.data.apiKey;
    } finally {
        if (ticket === secretRequestVersion) legacyBusy.value = false;
    }
};
const resetLegacy = async () => {
    if (legacyBusy.value || !detailVisible.value || detail.value?.kind !== 'legacy') return;
    const ticket = ++secretRequestVersion;
    try {
        await ElMessageBox.confirm(i18n.global.t('setting.apiKeyResetHelper'), i18n.global.t('setting.apiKeyReset'));
    } catch {
        return;
    }
    if (ticket !== secretRequestVersion || !detailVisible.value || !active) return;
    legacyBusy.value = true;
    legacySecret.value = '';
    try {
        const response = await generateApiKey();
        await search(ticket);
        // A reset changes the revision and deliberately clears any older reveal.
        // Never restore a secret after the original details drawer was closed.
        if (ticket === secretRequestVersion && detailVisible.value && active && detail.value?.kind === 'legacy')
            legacySecret.value = response.data;
    } finally {
        if (ticket === secretRequestVersion) legacyBusy.value = false;
    }
};
const activate = () => {
    if (active) return;
    active = true;
    void search();
};
const deactivate = () => {
    active = false;
    searchVersion++;
    clearLegacySecret();
    detailVisible.value = false;
};
onMounted(activate);
onActivated(activate);
onDeactivated(deactivate);
onBeforeUnmount(deactivate);
</script>
