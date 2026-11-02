<template>
    <div>
        <LayoutContent :title="$t('apiKeyManagement.title')" v-loading="loading">
            <template #leftToolBar>
                <el-button type="primary" @click="editor?.open()">
                    {{ $t('apiKeyManagement.create') }}
                </el-button>
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
                            <div class="text-xs text-[var(--el-text-color-secondary)] break-all">{{ row.keyHint }}</div>
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
                    <el-table-column v-if="!isMobile" :label="$t('setting.ipWhiteList')" min-width="180">
                        <template #default="{ row }">
                            <span v-if="isAnyAPIKeyIP(row.ipWhiteList)">{{ $t('apiKeyManagement.anyIP') }}</span>
                            <div
                                v-else-if="whitelistEntries(row.ipWhiteList).length"
                                class="flex min-w-0 items-center gap-2"
                            >
                                <div class="min-w-0">
                                    <div
                                        v-for="(ip, index) in whitelistEntries(row.ipWhiteList).slice(0, 2)"
                                        :key="index"
                                        class="break-all"
                                    >
                                        {{ ip }}
                                    </div>
                                </div>
                                <el-popover
                                    v-if="whitelistEntries(row.ipWhiteList).length > 2"
                                    placement="top"
                                    trigger="click"
                                    :width="360"
                                    :title="$t('setting.ipWhiteList')"
                                    :popper-style="{ maxWidth: 'calc(100vw - 32px)' }"
                                >
                                    <template #reference>
                                        <el-button
                                            link
                                            type="primary"
                                            size="small"
                                            class="shrink-0"
                                            :aria-label="
                                                $t('setting.ipWhiteList') +
                                                ': ' +
                                                $t('commons.table.total', [whitelistEntries(row.ipWhiteList).length])
                                            "
                                        >
                                            +{{ whitelistEntries(row.ipWhiteList).length - 2 }}
                                        </el-button>
                                    </template>
                                    <ul class="m-0 max-h-64 list-none overflow-y-auto p-0">
                                        <li
                                            v-for="(ip, index) in whitelistEntries(row.ipWhiteList)"
                                            :key="index"
                                            class="break-all py-1"
                                        >
                                            {{ ip }}
                                        </li>
                                    </ul>
                                </el-popover>
                            </div>
                            <span v-else>—</span>
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
            @close="clearSecret"
        >
            <APIKeySummary v-if="detail" :item="detail" />
            <div v-if="detail && detail.status !== 'Revoked'" class="mt-5 flex flex-col gap-3">
                <el-alert
                    :title="
                        $t(detail.kind === 'legacy' ? 'apiKeyManagement.legacyHelp' : 'apiKeyManagement.revealHelp')
                    "
                    type="info"
                    :closable="false"
                    show-icon
                />
                <el-space v-if="detail.kind === 'legacy'" wrap>
                    <el-button :disabled="secretBusy" @click="resetLegacy">
                        {{ $t('setting.apiKeyReset') }}
                    </el-button>
                </el-space>
                <el-input
                    :model-value="secret || '********'"
                    :type="secret ? 'text' : 'password'"
                    :aria-label="$t('setting.apiKey')"
                    readonly
                    autocomplete="off"
                >
                    <template #suffix>
                        <el-button
                            link
                            :icon="secret ? Hide : View"
                            :loading="secretBusy && secretAction === 'reveal'"
                            :disabled="loading || secretBusy"
                            :aria-label="$t(secret ? 'apiKeyManagement.hideSecret' : 'apiKeyManagement.reveal')"
                            :title="$t(secret ? 'apiKeyManagement.hideSecret' : 'apiKeyManagement.reveal')"
                            :aria-pressed="!!secret"
                            @click="toggleSecret"
                        />
                    </template>
                    <template #append>
                        <el-button
                            :loading="secretBusy && secretAction === 'copy'"
                            :disabled="loading || secretBusy"
                            @click="copySecret"
                        >
                            {{ $t('commons.button.copy') }}
                        </el-button>
                    </template>
                </el-input>
            </div>
        </DrawerPro>
    </div>
</template>

<script setup lang="ts">
import { computed, onActivated, onBeforeUnmount, onDeactivated, onMounted, reactive, ref, watch } from 'vue';
import { ElMessageBox } from 'element-plus';
import { Hide, View } from '@element-plus/icons-vue';
import type { APIKey } from '@/api/interface/api-key';
import { deleteAPIKey, revealAPIKey, revokeAPIKey, searchAPIKeys, setAPIKeyStatus } from '@/api/modules/api-key';
import { generateApiKey, getUserInfo } from '@/api/modules/auth';
import APIKeyEditor from '@/components/api-key-management/editor.vue';
import APIKeySummary from '@/components/api-key-management/summary.vue';
import DrawerPro from '@/components/drawer-pro/index.vue';
import type { FuTableOperationButton } from '@/components/table/shared';
import { isAnyAPIKeyIP, normalizeAPIKeyIPs } from '@/utils/api-key';
import { MsgSuccess, MsgWarning } from '@/utils/message';
import { copyText } from '@/utils/clipboard';
import i18n from '@/lang';
import { useGlobalStore } from '@/composables/useGlobalStore';

const { isMobile, globalStore } = useGlobalStore();

const items = ref<APIKey.Item[]>([]);
const loading = ref(false);
const editor = ref<InstanceType<typeof APIKeyEditor>>();
const detail = ref<APIKey.Item>();
const detailVisible = ref(false);
const secret = ref('');
const secretBusy = ref(false);
const secretAction = ref<'reveal' | 'copy'>('reveal');
const pagination = reactive({ currentPage: 1, pageSize: 20, total: 0 });
let searchVersion = 0;
let secretRequestVersion = 0;
let active = false;
let operationPending = false;

const whitelistEntries = (value: string) =>
    normalizeAPIKeyIPs(value || '')
        .split('\n')
        .filter(Boolean);

const clearSecret = () => {
    secretRequestVersion++;
    secret.value = '';
    secretBusy.value = false;
};

const search = async (preserveSecretRequest?: number) => {
    const ticket = ++searchVersion;
    loading.value = true;
    try {
        const response = await searchAPIKeys({ page: pagination.currentPage, pageSize: pagination.pageSize });
        if (ticket !== searchVersion) return;
        const lastPage = Math.max(1, Math.ceil(response.data.total / pagination.pageSize));
        if (pagination.currentPage > lastPage) {
            pagination.currentPage = lastPage;
            await search(preserveSecretRequest);
            return;
        }
        items.value = response.data.items || [];
        pagination.total = response.data.total;
        if (detail.value) {
            const previousRevision = detail.value.revision;
            detail.value = items.value.find((entry) => entry.id === detail.value?.id);
            if (detail.value?.revision !== previousRevision && preserveSecretRequest !== secretRequestVersion)
                clearSecret();
            if (!detail.value) detailVisible.value = false;
        }
    } finally {
        if (ticket === searchVersion) loading.value = false;
    }
};
const showDetail = (entry: APIKey.Item) => {
    clearSecret();
    detail.value = entry;
    detailVisible.value = true;
};
const operate = async (entry: APIKey.Item, action: string) => {
    if (action === 'edit') {
        editor.value?.open(entry);
        return;
    }
    if (!active || operationPending || !['status', 'revoke', 'delete'].includes(action)) return;
    operationPending = true;
    const requestVersion = searchVersion;
    try {
        await ElMessageBox.confirm(
            i18n.global.t(
                action === 'delete'
                    ? 'apiKeyManagement.deleteConfirm'
                    : action === 'revoke'
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
        operationPending = false;
        return;
    }
    if (!active || requestVersion !== searchVersion) {
        operationPending = false;
        return;
    }
    clearSecret();
    loading.value = true;
    try {
        const reference = { id: entry.id, revision: entry.revision };
        const response =
            action === 'delete'
                ? await deleteAPIKey(reference)
                : action === 'revoke'
                  ? await revokeAPIKey(reference)
                  : await setAPIKeyStatus({ ...reference, status: entry.status === 'Enable' ? 'Disable' : 'Enable' });
        clearSecret();
        if (response.data?.terminalClosePending) MsgWarning(i18n.global.t('apiKeyManagement.closePending'));
        else MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    } finally {
        operationPending = false;
        if (active) await search();
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
    {
        label: i18n.global.t('commons.button.delete'),
        show: (row) => row.kind === 'apiKey',
        click: (row) => operate(row, 'delete'),
    },
]);
const toggleSecret = async () => {
    if (secretBusy.value || loading.value) return;
    if (secret.value) clearSecret();
    else await revealSecret();
};
const accessSecret = async (action: 'reveal' | 'copy') => {
    if (secretBusy.value || loading.value || !active || !detailVisible.value || !detail.value) return;
    if (detail.value.status === 'Revoked') return;
    const reference = { id: detail.value.id, revision: detail.value.revision };
    const kind = detail.value.kind;
    const ticket = ++secretRequestVersion;
    secretAction.value = action;
    secretBusy.value = true;
    try {
        if (action === 'copy' && secret.value) {
            await copyText(secret.value);
            return;
        }
        secret.value = '';
        const data =
            kind === 'legacy'
                ? { ...reference, apiKey: (await getUserInfo()).data.apiKey }
                : (await revealAPIKey(reference)).data;
        if (
            ticket === secretRequestVersion &&
            active &&
            detailVisible.value &&
            detail.value?.id === reference.id &&
            detail.value?.kind === kind &&
            detail.value?.revision === reference.revision &&
            data.id === reference.id &&
            data.revision === reference.revision &&
            data.apiKey
        ) {
            if (action === 'copy') await copyText(data.apiKey);
            else secret.value = data.apiKey;
        }
    } catch {
        if (ticket === secretRequestVersion && active) await search();
    } finally {
        if (ticket === secretRequestVersion) secretBusy.value = false;
    }
};
const revealSecret = () => accessSecret('reveal');
const copySecret = () => accessSecret('copy');
const resetLegacy = async () => {
    if (secretBusy.value || !detailVisible.value || detail.value?.kind !== 'legacy') return;
    const ticket = ++secretRequestVersion;
    try {
        await ElMessageBox.confirm(i18n.global.t('setting.apiKeyResetHelper'), i18n.global.t('setting.apiKeyReset'));
    } catch {
        return;
    }
    if (ticket !== secretRequestVersion || !detailVisible.value || !active) return;
    secretAction.value = 'reveal';
    secretBusy.value = true;
    secret.value = '';
    try {
        const response = await generateApiKey();
        await search(ticket);
        // A reset changes the revision and deliberately clears any older reveal.
        // Never restore a secret after the original details drawer was closed.
        if (ticket === secretRequestVersion && detailVisible.value && active && detail.value?.kind === 'legacy')
            secret.value = response.data;
    } finally {
        if (ticket === secretRequestVersion) secretBusy.value = false;
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
    clearSecret();
    detailVisible.value = false;
};
watch(
    () => globalStore.isLogin,
    (loggedIn) => {
        if (!loggedIn) deactivate();
    },
    { flush: 'sync' },
);
onMounted(activate);
onActivated(activate);
onDeactivated(deactivate);
onBeforeUnmount(deactivate);
</script>
