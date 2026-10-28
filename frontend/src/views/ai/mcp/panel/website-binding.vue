<template>
    <section class="min-w-0 border-t border-[var(--el-border-color-lighter)] pt-4">
        <div class="mb-4 flex min-w-0 flex-wrap items-center justify-between gap-2">
            <span class="font-medium">{{ $t('panelMCP.websiteBinding') }}</span>
            <el-tag
                v-if="binding"
                :type="binding.state === 'bound' ? 'success' : binding.state === 'unbound' ? 'info' : 'warning'"
                size="small"
            >
                {{ $t(`panelMCP.${stateLabels[binding.state]}`) }}
            </el-tag>
        </div>
        <el-alert v-if="errorKey" class="!mb-4" type="warning" :closable="false" :title="$t(errorKey)" show-icon />
        <el-alert
            v-else-if="binding?.state === 'pending' || uncertain"
            class="!mb-4"
            type="warning"
            :closable="false"
            :title="$t('panelMCP.websiteUncertain')"
            show-icon
        />
        <el-alert
            v-else-if="binding?.state === 'unavailable'"
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
                    type="danger"
                    plain
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
                <el-button type="danger" plain :disabled="disabled || !canManage || busy" @click="unbind">
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
            <el-button :loading="reading" :disabled="disabled || !canManage || busy" @click="load">
                {{ $t('commons.button.refresh') }}
            </el-button>
        </div>
    </section>
</template>

<script setup lang="ts">
import { computed, onActivated, onBeforeUnmount, onDeactivated, onMounted, ref, watch } from 'vue';
import { ElMessageBox } from 'element-plus';
import { v4 as uuidv4 } from 'uuid';
import type { PanelMCP } from '@/api/interface/panel-mcp';
import {
    bindPanelMCPWebsite,
    getPanelMCPWebsite,
    unbindPanelMCPWebsite,
    updatePanelMCPWebsite,
} from '@/api/modules/panel-mcp';
import { listLocalNodeSSL } from '@/api/modules/website';
import {
    getMCPWebsiteHostname,
    isMCPMutationUncertain,
    isMCPRevisionConflict,
    isMCPWebsiteCertificateMatch,
} from '@/utils/panel-mcp';
import { dateFormatSimple } from '@/utils/date';
import { listNodes } from '@/utils/node';
import { routerToNameWithParams } from '@/utils/router';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { MsgSuccess } from '@/utils/message';
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
type PendingOperation = PanelMCP.WebsitePendingRequest;
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
const errorKey = ref('');
const domainError = ref('');
const certificateError = ref('');
const localPending = ref<PendingOperation>();
const busy = computed(() => reading.value || writing.value);
const associated = computed(() => !!binding.value && binding.value.state !== 'unbound');
const pendingOperation = computed(() => localPending.value || binding.value?.pendingRequest);
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
let certificatePromise: Promise<void> | undefined;
const current = (version: number) => active && props.canManage && globalStore.isLogin && version === generation;

const acceptBinding = (value: PanelMCP.WebsiteBinding) => {
    binding.value = value;
    if (value.pendingRequest) {
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
        (item.state !== 'pending' || !!item.pendingRequest) &&
        (!item.pendingRequest ||
            (['bind', 'update', 'unbind'].includes(item.pendingRequest.operation) &&
                typeof item.pendingRequest.requestId === 'string' &&
                !!item.pendingRequest.requestId))
    );
};
const read = async (version: number) => {
    const ticket = ++readVersion;
    const response = await getPanelMCPWebsite();
    if (!current(version) || ticket !== readVersion) return;
    if (!validBinding(response.data)) throw new Error('Invalid MCP website status');
    acceptBinding(response.data);
    return response.data;
};
const load = async () => {
    if (!active || !props.canManage || busy.value || props.disabled) return;
    const version = generation;
    reading.value = true;
    try {
        await read(version);
        if (current(version)) errorKey.value = '';
    } catch {
        if (current(version)) errorKey.value = 'panelMCP.websiteStatusLoadFailed';
    } finally {
        if (current(version)) reading.value = false;
    }
};
const loadCertificates = async () => {
    if (!active || !props.canManage || certificatesLoaded.value) return;
    if (certificatePromise) return certificatePromise;
    const version = generation;
    const ticket = ++certificateVersion;
    certificatesBusy.value = true;
    const loading = (async () => {
        try {
            const response = await listLocalNodeSSL({});
            if (!current(version) || ticket !== certificateVersion) return;
            certificates.value = response.data.map((item) => ({
                id: item.id,
                primaryDomain: item.primaryDomain,
                otherDomains: item.otherDomains,
                expireDate: item.expireDate,
            }));
            certificatesLoaded.value = true;
        } catch {
            if (current(version)) certificateError.value = i18n.global.t('panelMCP.websiteCertificateRequired');
        } finally {
            if (ticket === certificateVersion) certificatesBusy.value = false;
        }
    })();
    certificatePromise = loading;
    await loading;
    if (certificatePromise === loading) certificatePromise = undefined;
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
        await loadCertificates();
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
    errorKey.value = '';
    localPending.value = { ...operation };
    try {
        const latest = binding.value || (await read(version));
        if (!current(version) || !latest) return false;
        const request = { requestId: operation.requestId, revision: latest.revision };
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
                  : await unbindPanelMCPWebsite(request);
        if (!current(version)) return false;
        if (!validBinding(response.data)) throw { code: 503 };
        acceptBinding(response.data);
        if (response.data.state !== (operation.operation === 'unbind' ? 'unbound' : 'bound')) throw { code: 503 };
        localPending.value = undefined;
        uncertain.value = false;
        editing.value = false;
        props.draft.requested = false;
        emit('changed', response.data);
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        return true;
    } catch (error) {
        if (current(version)) {
            uncertain.value = isMCPMutationUncertain(error);
            errorKey.value = isMCPRevisionConflict(error)
                ? 'panelMCP.websiteConflict'
                : uncertain.value
                  ? 'panelMCP.websiteUncertain'
                  : 'panelMCP.websiteOperationFailed';
            if (!uncertain.value) localPending.value = undefined;
            try {
                const verified = await read(version);
                if (
                    uncertain.value &&
                    verified &&
                    verified.requestId === operation.requestId &&
                    !verified.pendingRequest &&
                    verified.state === (operation.operation === 'unbind' ? 'unbound' : 'bound')
                ) {
                    localPending.value = undefined;
                    uncertain.value = editing.value = false;
                    errorKey.value = '';
                    props.draft.requested = false;
                    emit('changed', verified);
                    return true;
                }
            } catch {
                /* Keep the submitted operation until the user can check its result. */
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
        );
    } catch {
        return;
    }
    if (!current(version) || busy.value || props.disabled) return;
    await perform({ operation: 'unbind', requestId: uuidv4() });
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
    if (props.draft.enableSSL && showEditor.value) void loadCertificates();
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
    domainError.value = certificateError.value = errorKey.value = '';
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
.mcp-website-details :deep(.el-descriptions__table) {
    table-layout: fixed;
}
.mcp-website-details :deep(.el-descriptions__cell) {
    overflow-wrap: anywhere;
}
</style>
