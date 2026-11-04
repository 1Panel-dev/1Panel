<template>
    <section class="mcp-website-binding min-w-0 border-t border-[var(--el-border-color-lighter)] pt-4">
        <div class="mb-4 flex min-w-0 flex-wrap items-center justify-between gap-2">
            <span class="font-medium">{{ $t('panelMCP.websiteBinding') }}</span>
            <el-tag
                v-if="binding"
                :type="
                    pending
                        ? 'warning'
                        : binding.state === 'bound'
                          ? 'success'
                          : binding.state === 'unbound'
                            ? 'info'
                            : 'warning'
                "
                size="small"
            >
                {{ $t(`panelMCP.${stateLabels[pending ? 'pending' : binding.state]}`) }}
            </el-tag>
        </div>
        <el-alert
            v-if="binding?.state === 'unavailable'"
            class="!mb-4"
            type="warning"
            :closable="false"
            :title="$t('panelMCP.websiteUnavailable')"
            show-icon
        />
        <el-checkbox
            v-if="mode === 'installation' && !associated"
            v-model="draft.requested"
            :disabled="disabled || !canManage || busy || pending"
            class="mb-4"
        >
            {{ $t('panelMCP.websiteBindRequested') }}
        </el-checkbox>
        <template v-if="associated && !editing">
            <el-descriptions :column="1" border size="small" class="mcp-website-details mb-4">
                <el-descriptions-item :label="$t('website.domain')">
                    <span class="break-all">{{ binding?.domain || draft.domain || '—' }}</span>
                </el-descriptions-item>
                <el-descriptions-item :label="$t('panelMCP.endpoint')">
                    <div class="flex min-w-0 items-start gap-2">
                        <span class="min-w-0 flex-1 break-all">
                            {{ binding?.endpoint || $t('panelMCP.endpointUnavailable') }}
                        </span>
                        <CopyButton v-if="binding?.state === 'bound' && binding.endpoint" :content="binding.endpoint" />
                    </div>
                </el-descriptions-item>
            </el-descriptions>
        </template>
        <template v-if="showEditor">
            <el-form-item :label="$t('website.domain')" :error="domainError">
                <el-input
                    v-model.trim="draft.domain"
                    :disabled="disabled || !canManage || busy || associated || pending"
                    :maxlength="253"
                    placeholder="mcp.example.com"
                    @input="domainError = ''"
                />
            </el-form-item>
            <el-form-item>
                <el-checkbox
                    v-model="draft.enableSSL"
                    :disabled="disabled || !canManage || busy || pending"
                    @change="certificateError = ''"
                >
                    {{ $t('website.enable') }} HTTPS
                </el-checkbox>
            </el-form-item>
            <el-form-item v-if="draft.enableSSL" :label="$t('website.ssl')" :error="certificateError">
                <el-select
                    v-model="selectedCertificateID"
                    :disabled="disabled || !canManage || busy || pending"
                    :loading="certificatesBusy"
                    :placeholder="$t('website.selectSSL')"
                    class="w-full"
                    filterable
                    @change="certificateError = ''"
                >
                    <el-option
                        v-for="certificate in matchingCertificates"
                        :key="certificate.id"
                        :value="certificate.id"
                        :label="certificate.primaryDomain"
                    >
                        <span>{{ certificate.primaryDomain }}</span>
                        <span v-if="certificate.expireDate" class="ml-2 text-xs text-[var(--el-text-color-secondary)]">
                            {{ dateFormatSimple(certificate.expireDate) }}
                        </span>
                    </el-option>
                </el-select>
                <span v-if="certificatesLoaded && !matchingCertificates.length" class="input-help mt-1 leading-5">
                    {{
                        $t(
                            getMCPWebsiteHostname(draft.domain)
                                ? 'panelMCP.websiteCertificateEmpty'
                                : 'panelMCP.websiteDomainInvalid',
                        )
                    }}
                </span>
            </el-form-item>
        </template>
        <div v-if="mode === 'settings'" class="flex flex-wrap items-center gap-2 [&>.el-button]:!ml-0">
            <template v-if="pending">
                <el-button type="primary" :loading="writing" :disabled="disabled || !canManage || busy" @click="retry">
                    {{ $t('panelMCP.websiteRetry') }}
                </el-button>
                <el-button
                    v-if="pendingOperation?.operation === 'bind' || pendingOperation?.operation === 'update'"
                    :disabled="disabled || !canManage || busy"
                    @click="cancelPending"
                >
                    {{
                        $t(
                            pendingOperation?.operation === 'bind'
                                ? 'panelMCP.websiteCancelPendingBind'
                                : 'panelMCP.websiteUnbind',
                        )
                    }}
                </el-button>
            </template>
            <template v-else-if="associated && !editing">
                <el-button :disabled="disabled || !canManage || busy || !binding?.websiteId" @click="openWebsite">
                    {{ $t('panelMCP.websiteSettings') }}
                </el-button>
                <el-button :disabled="disabled || !canManage || busy" @click="editCertificate">
                    {{ $t('panelMCP.websiteCertificateUpdate') }}
                </el-button>
                <el-button :disabled="disabled || !canManage || busy" @click="unbind">
                    {{ $t('panelMCP.websiteUnbind') }}
                </el-button>
            </template>
            <template v-else>
                <el-button
                    type="primary"
                    :loading="writing"
                    :disabled="disabled || !canManage || busy || !binding"
                    @click="submit"
                >
                    {{ $t(editing ? 'commons.button.save' : 'panelMCP.websiteBind') }}
                </el-button>
                <el-button v-if="editing" :disabled="busy" @click="cancelEditing">
                    {{ $t('commons.button.cancel') }}
                </el-button>
            </template>
            <el-button :loading="reading" :disabled="disabled || !canManage || busy" @click="refresh">
                {{ $t('commons.button.refresh') }}
            </el-button>
        </div>
    </section>
</template>

<script setup lang="ts">
import { computed, onActivated, onBeforeUnmount, onDeactivated, onMounted, ref, watch } from 'vue';
import { ElMessageBox } from 'element-plus';
import { isAxiosError } from 'axios';
import { v4 as uuidv4 } from 'uuid';
import type { PanelMCP } from '@/api/interface/panel-mcp';
import {
    bindPanelMCPWebsite,
    getPanelMCPWebsite,
    unbindPanelMCPWebsite,
    updatePanelMCPWebsite,
} from '@/api/modules/panel-mcp';
import { listLocalNodeSSL } from '@/api/modules/website';
import { getMCPWebsiteHostname, isMCPMutationUncertain, isMCPWebsiteCertificateMatch } from '@/utils/panel-mcp';
import { dateFormatSimple } from '@/utils/date';
import { listNodes } from '@/utils/node';
import { routerToNameWithParams } from '@/utils/router';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { MsgError, MsgSuccess } from '@/utils/message';
import i18n from '@/lang';

const props = defineProps<{
    draft: PanelMCP.WebsiteDraft;
    mode: 'installation' | 'settings';
    canManage: boolean;
    disabled?: boolean;
}>();
const emit = defineEmits<{
    busy: [value: boolean];
    binding: [value: PanelMCP.WebsiteBinding | undefined];
    changed: [value: PanelMCP.WebsiteBinding];
}>();
const { globalStore } = useGlobalStore();
type Certificate = { id: number; primaryDomain: string; otherDomains: string; expireDate: string };
type PendingOperation = PanelMCP.WebsitePendingRequest & { revision?: number };
const stateLabels: Record<PanelMCP.WebsiteBinding['state'], string> = {
    unbound: 'websiteUnbound',
    pending: 'websitePending',
    bound: 'websiteBound',
    unavailable: 'websiteUnavailable',
};
const binding = ref<PanelMCP.WebsiteBinding>();
const reading = ref(false);
const writing = ref(false);
const certificatesBusy = ref(false);
const certificatesLoaded = ref(false);
const certificates = ref<Certificate[]>([]);
const editing = ref(false);
const uncertain = ref(false);
const domainError = ref('');
const certificateError = ref('');
const localPending = ref<PendingOperation>();
const busy = computed(() => reading.value || writing.value);
const associated = computed(() => !!binding.value && binding.value.state !== 'unbound');
const pendingOperation = computed<PendingOperation | undefined>(
    () => localPending.value || binding.value?.pendingRequest,
);
const pending = computed(() => binding.value?.state === 'pending' || !!pendingOperation.value || uncertain.value);
const showEditor = computed(
    () => (!associated.value && (props.mode === 'settings' || props.draft.requested)) || editing.value,
);
const matchingCertificates = computed(() =>
    certificates.value.filter((certificate) => isMCPWebsiteCertificateMatch(props.draft.domain, certificate)),
);
const selectedCertificateID = computed<number | undefined>({
    get: () =>
        props.draft.sslID > 0 && matchingCertificates.value.some((certificate) => certificate.id === props.draft.sslID)
            ? props.draft.sslID
            : undefined,
    set: (value) => {
        props.draft.sslID = value ?? 0;
    },
});
let active = false;
let generation = 0;
let readVersion = 0;
let navigationVersion = 0;
let certificateVersion = 0;
let certificatePromise: Promise<boolean> | undefined;
const current = (version: number) => active && props.canManage && globalStore.isLogin && version === generation;

const notifyFailure = (error: unknown, key: string) => {
    const requestError = isAxiosError(error);
    if (requestError && /timeout/i.test(error.message)) {
        MsgError(i18n.global.t('commons.msg.requestTimeout'));
        return;
    }
    const data = (requestError ? error.response?.data : error) as { code?: unknown; message?: unknown } | undefined;
    const message =
        data && (requestError || typeof data.code === 'number') && typeof data.message === 'string'
            ? data.message.trim()
            : '';
    MsgError(
        message.toLowerCase().includes('operation not permitted')
            ? i18n.global.t('license.tamperHelper')
            : message || i18n.global.t(key),
    );
};
const completed = (value: PanelMCP.WebsiteBinding, operation: PendingOperation) =>
    value.requestId === operation.requestId &&
    !value.pendingRequest &&
    value.state === (operation.operation === 'unbind' ? 'unbound' : 'bound') &&
    (operation.operation !== 'unbind' || operation.deleteWebsite !== true || value.websiteDeleted === true);
const finish = (value: PanelMCP.WebsiteBinding) => {
    localPending.value = undefined;
    uncertain.value = editing.value = false;
    props.draft.requested = false;
    emit('changed', value);
};
const acceptBinding = (value: PanelMCP.WebsiteBinding, syncDraft = true) => {
    binding.value = value;
    if (syncDraft && value.pendingRequest) {
        const request = value.pendingRequest;
        if (request.domain !== undefined) props.draft.domain = request.domain;
        if (request.enableSSL !== undefined) props.draft.enableSSL = request.enableSSL;
        if (request.sslID !== undefined) props.draft.sslID = request.sslID;
        props.draft.requested = request.operation === 'bind';
    }
    emit('binding', value);
};
const validBinding = (value: unknown): value is PanelMCP.WebsiteBinding => {
    const item = value as PanelMCP.WebsiteBinding;
    return (
        !!item &&
        Object.hasOwn(stateLabels, item.state) &&
        Number.isSafeInteger(item.revision) &&
        item.revision > 0 &&
        Number.isSafeInteger(item.websiteId) &&
        item.websiteId >= 0 &&
        typeof item.domain === 'string' &&
        typeof item.endpoint === 'string' &&
        typeof item.enableSSL === 'boolean' &&
        Number.isSafeInteger(item.sslID) &&
        item.sslID >= 0 &&
        typeof item.requestId === 'string' &&
        (item.websiteDeleted === undefined || typeof item.websiteDeleted === 'boolean') &&
        (item.state !== 'pending' || !!item.pendingRequest) &&
        (!item.pendingRequest ||
            (['bind', 'update', 'unbind'].includes(item.pendingRequest.operation) &&
                typeof item.pendingRequest.requestId === 'string' &&
                !!item.pendingRequest.requestId &&
                (item.pendingRequest.deleteWebsite === undefined ||
                    typeof item.pendingRequest.deleteWebsite === 'boolean')))
    );
};
const read = async (version: number, syncDraft = true) => {
    const ticket = ++readVersion;
    const operation = pendingOperation.value;
    const response = await getPanelMCPWebsite();
    if (!current(version) || ticket !== readVersion) return;
    if (!validBinding(response?.data)) throw new Error('Invalid MCP website status');
    acceptBinding(response.data, syncDraft);
    if (operation && completed(response.data, operation)) finish(response.data);
    return response.data;
};
const load = async (refreshCertificates = false) => {
    if (!active || !props.canManage || busy.value || props.disabled) return;
    const version = generation;
    reading.value = true;
    const request = read(version, !refreshCertificates);
    const ticket = readVersion;
    const activeRead = () => current(version) && ticket === readVersion;
    try {
        const status = request
            .catch((error) => {
                if (activeRead()) notifyFailure(error, 'panelMCP.websiteStatusLoadFailed');
                throw error;
            })
            .finally(() => {
                if (!refreshCertificates && activeRead()) reading.value = false;
            });
        const requests: Promise<unknown>[] = [status];
        if (refreshCertificates || (props.draft.enableSSL && showEditor.value)) {
            requests.push(loadCertificates(refreshCertificates, false));
        }
        const [result, certificate] = await Promise.allSettled(requests);
        if (
            activeRead() &&
            result.status === 'fulfilled' &&
            certificate?.status === 'fulfilled' &&
            certificate.value === false &&
            (refreshCertificates || (props.draft.enableSSL && showEditor.value))
        ) {
            notifyFailure(undefined, 'panelMCP.websiteCertificatesLoadFailed');
        }
    } finally {
        if (activeRead()) reading.value = false;
    }
};
const refresh = () => load(true);
const loadCertificates = async (force = false, notify = true): Promise<boolean> => {
    if (!active || !props.canManage || !globalStore.isLogin) return false;
    if (certificatePromise && !force) return certificatePromise;
    if (certificatesLoaded.value && !force) return true;
    const version = generation;
    const ticket = ++certificateVersion;
    certificatesBusy.value = true;
    certificateError.value = '';
    const loading = (async () => {
        try {
            const response = await listLocalNodeSSL({}, { errorMessageMode: 'manual' });
            if (!current(version) || ticket !== certificateVersion) return false;
            if (!Array.isArray(response?.data)) throw new Error('Invalid MCP certificate list');
            certificates.value = response.data.map((item) => ({
                id: item.id,
                primaryDomain: item.primaryDomain,
                otherDomains: item.otherDomains,
                expireDate: item.expireDate,
            }));
            certificatesLoaded.value = true;
            certificateError.value = '';
            return true;
        } catch (error) {
            if (current(version) && ticket === certificateVersion) {
                certificatesLoaded.value = false;
                if (notify && props.draft.enableSSL && showEditor.value) {
                    notifyFailure(error, 'panelMCP.websiteCertificatesLoadFailed');
                }
            }
            return false;
        } finally {
            if (ticket === certificateVersion) certificatesBusy.value = false;
        }
    })();
    certificatePromise = loading;
    const loaded = await loading;
    if (certificatePromise === loading) certificatePromise = undefined;
    return loaded;
};
const validate = async () => {
    domainError.value = certificateError.value = '';
    if (!props.canManage || !globalStore.isLogin) return false;
    if (props.mode === 'installation' && !props.draft.requested) return true;
    if (!getMCPWebsiteHostname(props.draft.domain)) {
        domainError.value = i18n.global.t('panelMCP.websiteDomainInvalid');
        return false;
    }
    if (props.draft.enableSSL) {
        const version = generation;
        if (!(await loadCertificates()) || !current(version)) return false;
        if (!matchingCertificates.value.some((item) => item.id === props.draft.sslID)) {
            certificateError.value = i18n.global.t('panelMCP.websiteCertificateRequired');
            return false;
        }
    }
    return true;
};
const perform = async (operation: PendingOperation): Promise<boolean> => {
    if (!active || !props.canManage || !globalStore.isLogin || busy.value || props.disabled) return false;
    const version = generation;
    writing.value = true;
    emit('busy', true);
    localPending.value = { ...operation };
    const deletesWebsite = operation.operation === 'unbind' && operation.deleteWebsite === true;
    try {
        const latest = binding.value || (await read(version));
        if (!current(version) || !latest) return false;
        // A retry must keep the revision that authorized the original intent.
        const revision = operation.revision ?? latest.revision;
        operation = { ...operation, revision };
        localPending.value = operation;
        const request = { requestId: operation.requestId, revision };
        const response =
            operation.operation === 'bind'
                ? await bindPanelMCPWebsite({
                      ...request,
                      domain: operation.domain || '',
                      enableSSL: operation.enableSSL === true,
                      sslID: operation.enableSSL ? operation.sslID || 0 : 0,
                  })
                : operation.operation === 'update'
                  ? await updatePanelMCPWebsite({
                        ...request,
                        enableSSL: operation.enableSSL === true,
                        sslID: operation.enableSSL ? operation.sslID || 0 : 0,
                    })
                  : await unbindPanelMCPWebsite({
                        ...request,
                        ...(operation.deleteWebsite !== undefined ? { deleteWebsite: operation.deleteWebsite } : {}),
                    });
        if (!current(version)) return false;
        if (!validBinding(response?.data) || response.data.requestId !== operation.requestId) throw { code: 503 };
        if (deletesWebsite && response.data.websiteDeleted !== true) throw { code: 503 };
        acceptBinding(response.data);
        if (!completed(response.data, operation)) throw { code: 503 };
        finish(response.data);
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        return true;
    } catch (error) {
        if (current(version)) {
            const unconfirmed = isMCPMutationUncertain(error);
            uncertain.value = unconfirmed;
            if (!uncertain.value) localPending.value = undefined;
            try {
                const verified = await read(version);
                if (unconfirmed && verified && completed(verified, operation)) return true;
            } catch {
                /* Keep the submitted operation until the user can check its result. */
            }
            if (current(version)) {
                notifyFailure(error, unconfirmed ? 'panelMCP.websiteUncertain' : 'panelMCP.websiteOperationFailed');
            }
        }
        return false;
    } finally {
        writing.value = false;
        emit('busy', false);
    }
};
const submit = async () => {
    if (props.disabled || busy.value || pending.value) return false;
    const version = generation;
    if (!(await validate()) || !current(version)) return false;
    return perform({
        operation: editing.value ? 'update' : 'bind',
        requestId: uuidv4(),
        domain: props.draft.domain,
        enableSSL: props.draft.enableSSL,
        sslID: props.draft.sslID,
    });
};
const retry = async () => {
    const operation = pendingOperation.value;
    if (!operation || props.disabled || busy.value) return;
    await perform({ ...operation });
};
const editCertificate = () => {
    if (!binding.value || busy.value || props.disabled || !props.canManage || pending.value) return;
    Object.assign(props.draft, {
        domain: binding.value.domain,
        enableSSL: binding.value.enableSSL,
        sslID: binding.value.sslID,
    });
    editing.value = true;
    certificateError.value = '';
};
const cancelEditing = () => {
    if (!busy.value) editing.value = false;
};
const requestUnbind = async (cancelPending: boolean) => {
    if (!binding.value || busy.value || props.disabled || !props.canManage) return;
    const canCancel = ['bind', 'update'].includes(pendingOperation.value?.operation || '');
    if (pending.value && (!cancelPending || !canCancel)) return;
    if (cancelPending && !canCancel) return;
    const version = generation;
    try {
        await ElMessageBox.confirm(
            i18n.global.t('panelMCP.websiteUnbindConfirm'),
            i18n.global.t('panelMCP.websiteUnbind'),
            { type: 'warning' },
        );
    } catch {
        return;
    }
    if (!current(version) || busy.value || props.disabled) return;
    await perform({ operation: 'unbind', requestId: uuidv4(), deleteWebsite: true });
};
const unbind = () => requestUnbind(false);
const cancelPending = () => requestUnbind(true);
const openWebsite = async () => {
    const id = binding.value?.websiteId;
    const version = generation;
    if (!id || !current(version) || props.disabled || busy.value) return;
    const ticket = ++navigationVersion;
    let nodes: Awaited<ReturnType<typeof listNodes>>;
    try {
        nodes = await listNodes('all');
    } catch {
        return;
    }
    if (!current(version) || ticket !== navigationVersion || props.disabled || busy.value) return;
    const primary = nodes.find((node) => node.name === 'local');
    if (!primary) return;
    globalStore.currentNode = 'local';
    globalStore.currentNodeAddr = primary.addr;
    await routerToNameWithParams('WebsiteConfig', { id, tab: 'basic' });
};
const activate = () => {
    if (active) return;
    active = true;
    if (props.canManage) void load();
};
const invalidate = () => {
    generation++;
    readVersion++;
    navigationVersion++;
    certificateVersion++;
    certificatePromise = undefined;
    binding.value = undefined;
    certificates.value = [];
    certificatesLoaded.value = certificatesBusy.value = false;
    reading.value = false;
    editing.value = uncertain.value = false;
    localPending.value = undefined;
    domainError.value = certificateError.value = '';
    emit('binding', undefined);
};
const deactivate = () => {
    active = false;
    invalidate();
};
watch(
    () => props.canManage,
    (allowed) => {
        if (!allowed) invalidate();
        else if (active) void load();
    },
    { flush: 'sync' },
);
watch(
    () => [globalStore.isLogin, props.disabled, busy.value],
    () => navigationVersion++,
    { flush: 'sync' },
);
watch(
    () => props.disabled,
    (disabled) => {
        if (!disabled && active && props.canManage && !binding.value) void load();
    },
);
watch(
    () => [props.draft.enableSSL, showEditor.value],
    () => {
        if (props.draft.enableSSL && showEditor.value) void loadCertificates();
    },
);
onMounted(activate);
onActivated(activate);
onDeactivated(deactivate);
onBeforeUnmount(deactivate);
defineExpose({ validate, load, busy, binding });
</script>

<style scoped>
.input-help {
    color: var(--el-text-color-regular);
}

html.dark .mcp-website-binding {
    border-color: var(--el-border-color-lighter);
}

.mcp-website-details :deep(.el-descriptions__table) {
    table-layout: fixed;
}
.mcp-website-details :deep(.el-descriptions__cell) {
    overflow-wrap: anywhere;
}
</style>
