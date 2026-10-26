<template>
    <div v-if="status?.installed" class="app-status" v-loading="busy">
        <el-card>
            <div class="flex w-full min-w-0 flex-col gap-4 md:flex-row">
                <div class="ml-3 flex min-w-0 flex-wrap items-start gap-4">
                    <el-tag effect="dark" type="success">{{ $t('panelMCP.title') }}</el-tag>
                    <el-tag class="mt-0.5" size="small" round effect="light" :type="serviceStatusType">
                        {{ statusLabel }}
                    </el-tag>
                    <el-tag v-if="status.adapterVersion" class="!h-auto min-h-6 max-w-full">
                        <span class="whitespace-normal break-all">
                            {{ $t('app.version') }}: {{ status.adapterVersion }}
                        </span>
                    </el-tag>
                </div>
                <div
                    class="mt-0.5 flex min-w-0 flex-1 flex-wrap items-center gap-x-3 gap-y-2 md:gap-x-0 [&>.el-button]:!ml-0"
                >
                    <el-button type="primary" link @click="openStatus">{{ $t('commons.table.status') }}</el-button>
                    <template v-if="canManageService">
                        <template v-if="status.installed">
                            <el-divider direction="vertical" class="!hidden md:!inline-block" />
                            <el-button
                                type="primary"
                                link
                                :disabled="!canOperateService"
                                @click="operate(serviceOperation)"
                            >
                                {{ $t(`commons.operate.${serviceOperation}`) }}
                            </el-button>
                            <el-divider direction="vertical" class="!hidden md:!inline-block" />
                            <el-button type="primary" link :disabled="!canOperateService" @click="operate('restart')">
                                {{ $t('commons.operate.restart') }}
                            </el-button>
                            <el-divider direction="vertical" class="!hidden md:!inline-block" />
                            <el-button type="primary" link @click="openSettings">
                                {{ $t('commons.button.set') }}
                            </el-button>
                            <el-divider direction="vertical" class="!hidden md:!inline-block" />
                            <el-button type="primary" link @click="openLogs">{{ $t('commons.button.log') }}</el-button>
                        </template>
                    </template>
                    <el-divider v-if="status?.installed" direction="vertical" class="!hidden md:!inline-block" />
                    <el-button v-if="status?.installed" type="primary" link @click="$emit('connection')">
                        {{ $t('panelMCP.connection') }}
                    </el-button>
                    <el-divider direction="vertical" class="!hidden md:!inline-block" />
                    <el-button type="primary" link @click="refreshService">
                        {{ $t('commons.button.refresh') }}
                    </el-button>
                </div>
            </div>
        </el-card>
    </div>

    <DrawerPro
        v-model="statusVisible"
        :header="$t('panelMCP.serviceStatus')"
        :resource="$t('panelMCP.title')"
        size="min(640px, 100vw)"
    >
        <el-descriptions :column="1" :label-width="130" border size="small" class="mcp-details">
            <el-descriptions-item :label="$t('commons.table.status')">
                <el-tag size="small" round effect="light" :type="serviceStatusType">{{ statusLabel }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item :label="$t('panelMCP.ready')">
                {{ $t(status?.ready ? 'commons.true' : 'commons.false') }}
            </el-descriptions-item>
            <el-descriptions-item :label="$t('app.version')">
                <span class="break-all font-mono text-xs">{{ status?.adapterVersion || '—' }}</span>
            </el-descriptions-item>
            <template v-if="status && canManageService">
                <el-descriptions-item :label="$t('panelMCP.dockerImage')">
                    <div class="flex min-w-0 items-start gap-2">
                        <span class="min-w-0 flex-1 break-all font-mono text-xs">{{ status.dockerImage || '—' }}</span>
                        <CopyButton v-if="status.dockerImage" :content="status.dockerImage" class="shrink-0" />
                    </div>
                </el-descriptions-item>
                <el-descriptions-item :label="$t('panelMCP.containerName')">
                    <div class="flex min-w-0 items-start gap-2">
                        <span class="min-w-0 flex-1 break-all font-mono text-xs">
                            {{ status.containerName || '—' }}
                        </span>
                        <CopyButton v-if="status.containerName" :content="status.containerName" class="shrink-0" />
                    </div>
                </el-descriptions-item>
                <el-descriptions-item v-if="status.imageId" :label="$t('panelMCP.imageId')">
                    <div class="flex min-w-0 items-start gap-2">
                        <span class="min-w-0 flex-1 break-all font-mono text-xs">{{ status.imageId }}</span>
                        <CopyButton :content="status.imageId" class="shrink-0" />
                    </div>
                </el-descriptions-item>
            </template>
            <el-descriptions-item :label="$t('panelMCP.endpoint')">
                <div class="flex min-w-0 items-start gap-2">
                    <span class="min-w-0 flex-1 break-all">{{ endpoint || $t('panelMCP.endpointUnavailable') }}</span>
                    <CopyButton v-if="endpoint" :content="endpoint" class="shrink-0" />
                </div>
            </el-descriptions-item>
            <el-descriptions-item v-if="status?.internalEndpoint" :label="$t('panelMCP.internalEndpoint')">
                <span class="break-all">{{ status.internalEndpoint }}</span>
                <p class="input-help mb-0 mt-1 leading-5">{{ $t('panelMCP.internalEndpointHelp') }}</p>
            </el-descriptions-item>
            <el-descriptions-item v-if="status?.reason" :label="$t('commons.table.description')">
                <span class="break-words">{{ status.reason }}</span>
            </el-descriptions-item>
        </el-descriptions>
        <p v-if="!canManageService" class="input-help mt-2 leading-5">
            {{ $t('panelMCP.serviceAdminOnly') }}
        </p>
        <template #footer>
            <div class="flex flex-wrap items-center justify-end gap-2 [&>.el-button]:!ml-0">
                <el-button @click="statusVisible = false">{{ $t('commons.button.close') }}</el-button>
            </div>
        </template>
    </DrawerPro>

    <DrawerPro
        v-model="configurationVisible"
        :header="drawerMode === 'installation' ? installTitle : $t('commons.button.set')"
        :resource="drawerMode === 'installation' ? '' : $t('panelMCP.title')"
        size="min(640px, 100vw)"
        :confirm-before-close="true"
        @before-close="closeConfigurationWhenIdle"
    >
        <el-tabs
            v-if="drawerMode === 'settings'"
            :model-value="settingsTab"
            :before-leave="changeSettingsTab"
            class="min-w-0"
        >
            <el-tab-pane :label="$t('panelMCP.serviceConfiguration')" name="configuration" :disabled="busy">
                <el-alert
                    v-if="conflict"
                    type="warning"
                    :closable="false"
                    :title="$t('panelMCP.conflict')"
                    show-icon
                    class="!mb-5"
                />
                <el-form
                    v-if="settingsTab === 'configuration'"
                    ref="settingsRef"
                    :model="settings"
                    :rules="settingsRules"
                    label-position="top"
                    :disabled="busy || conflict || !canManageService"
                    @validate="onSettingsValidate"
                >
                    <el-form-item :label="$t('panelMCP.listenPort')" prop="listenPort">
                        <el-input-number
                            v-model="settings.listenPort"
                            :min="1024"
                            :max="65535"
                            :precision="0"
                            class="!w-64 max-w-full"
                        />
                    </el-form-item>
                    <el-form-item :label="$t('panelMCP.endpoint')" prop="externalUrl">
                        <el-input v-model.trim="settings.externalUrl" :placeholder="endpoint" :maxlength="2048" />
                        <span class="input-help mt-2 leading-5">{{ $t('panelMCP.externalUrlHelp') }}</span>
                    </el-form-item>
                    <el-collapse v-model="advancedSettings" class="mt-4">
                        <el-collapse-item :title="$t('app.advanced')" name="source">
                            <el-form-item :label="$t('panelMCP.accessPolicy')" prop="accessPolicy">
                                <el-select v-model="settings.accessPolicy" class="w-full">
                                    <el-option :label="$t('panelMCP.accessInherit')" value="inherit" />
                                    <el-option :label="$t('panelMCP.accessRestricted')" value="restricted" />
                                    <el-option :label="$t('panelMCP.accessAll')" value="all" />
                                </el-select>
                                <span class="input-help mt-2 leading-5">{{ $t('panelMCP.accessPolicyHelp') }}</span>
                            </el-form-item>
                            <el-form-item
                                v-if="settings.accessPolicy === 'restricted'"
                                :label="$t('panelMCP.allowedIPs')"
                                prop="allowedIPs"
                            >
                                <el-input
                                    v-model="allowedIPsText"
                                    type="textarea"
                                    :rows="3"
                                    :maxlength="4160"
                                    :placeholder="$t('panelMCP.allowedIPsPlaceholder')"
                                />
                            </el-form-item>
                        </el-collapse-item>
                    </el-collapse>
                </el-form>
            </el-tab-pane>
            <el-tab-pane :label="$t('panelMCP.imageUpgrade')" name="upgrade" :disabled="busy">
                <el-alert
                    v-if="installConflict"
                    type="warning"
                    :closable="false"
                    :title="$t('panelMCP.conflict')"
                    show-icon
                    class="!mb-5"
                >
                    <el-button link type="primary" :disabled="busy" @click="reloadInstallation">
                        {{ $t('commons.button.refresh') }}
                    </el-button>
                </el-alert>
                <el-form
                    v-if="settingsTab === 'upgrade'"
                    ref="installRef"
                    :model="installation"
                    :rules="installRules"
                    label-position="top"
                    :disabled="busy || installConflict || !canManageService"
                >
                    <el-form-item :label="$t('panelMCP.dockerImage')" prop="dockerImage">
                        <el-input
                            v-model.trim="installation.dockerImage"
                            :placeholder="DEFAULT_MCP_DOCKER_IMAGE"
                            :maxlength="180"
                        />
                        <span class="input-help mt-2 leading-5">{{ $t('panelMCP.dockerImageHelp') }}</span>
                    </el-form-item>
                </el-form>
                <el-collapse class="mt-4">
                    <el-collapse-item :title="$t('panelMCP.offlineInstall')" name="offline">
                        <p class="input-help !my-0 leading-5">{{ $t('panelMCP.dockerOfflineHelp') }}</p>
                    </el-collapse-item>
                </el-collapse>
            </el-tab-pane>
        </el-tabs>
        <template v-else>
            <el-alert
                v-if="installConflict"
                type="warning"
                :closable="false"
                :title="$t('panelMCP.conflict')"
                show-icon
                class="!mb-5"
            >
                <el-button link type="primary" :disabled="busy" @click="reloadInstallation">
                    {{ $t('commons.button.refresh') }}
                </el-button>
            </el-alert>
            <el-form
                ref="installRef"
                :model="installation"
                :rules="installRules"
                label-position="top"
                :disabled="busy || installConflict || !canManageService"
            >
                <el-form-item :label="$t('panelMCP.dockerImage')" prop="dockerImage">
                    <el-input
                        v-model.trim="installation.dockerImage"
                        :placeholder="DEFAULT_MCP_DOCKER_IMAGE"
                        :maxlength="180"
                    />
                    <span class="input-help mt-2 leading-5">{{ $t('panelMCP.dockerImageHelp') }}</span>
                </el-form-item>
                <el-form-item v-if="!installIsUpgrade" :label="$t('panelMCP.listenPort')" prop="listenPort">
                    <el-input-number
                        v-model="installation.listenPort"
                        :min="1024"
                        :max="65535"
                        :precision="0"
                        class="!w-64 max-w-full"
                    />
                </el-form-item>
                <el-form-item v-if="!installIsUpgrade" :label="$t('panelMCP.endpoint')" prop="externalUrl">
                    <el-input v-model.trim="installation.externalUrl" :placeholder="endpoint" :maxlength="2048" />
                    <span class="input-help mt-2 leading-5">{{ $t('panelMCP.externalUrlHelp') }}</span>
                </el-form-item>
            </el-form>
            <el-collapse class="mt-4">
                <el-collapse-item :title="$t('panelMCP.offlineInstall')" name="offline">
                    <p class="input-help !my-0 leading-5">{{ $t('panelMCP.dockerOfflineHelp') }}</p>
                </el-collapse-item>
            </el-collapse>
        </template>
        <template #footer>
            <div class="flex flex-wrap items-center justify-end gap-2 [&>.el-button]:!ml-0">
                <el-button :disabled="busy" @click="cancelConfiguration">
                    {{ $t('commons.button.cancel') }}
                </el-button>
                <el-button
                    v-if="drawerMode === 'settings' && settingsTab === 'configuration'"
                    type="primary"
                    :loading="busy"
                    :disabled="conflict || !canManageService"
                    @click="saveSettings"
                >
                    {{ $t('commons.button.save') }}
                </el-button>
                <el-button
                    v-else
                    type="primary"
                    :loading="busy"
                    :disabled="installConflict || !canManageService"
                    @click="install"
                >
                    {{ $t(installIsUpgrade ? 'panelMCP.upgradeTitle' : 'panelMCP.installEnable') }}
                </el-button>
            </div>
        </template>
    </DrawerPro>

    <DrawerPro
        v-model="logsVisible"
        :header="$t('commons.button.log')"
        :resource="$t('panelMCP.title')"
        size="large"
        :full-screen="true"
    >
        <template #content>
            <div class="flex flex-wrap items-center gap-2 [&>.el-button]:!ml-0">
                <el-checkbox v-model="followLogs" border @change="changeLogFollowing">
                    {{ $t('commons.button.watch') }}
                </el-checkbox>
                <el-button icon="Download" :disabled="!logs || !canManageService" @click="downloadLogs">
                    {{ $t('commons.button.download') }}
                </el-button>
                <el-button icon="Refresh" :loading="logsBusy" @click="loadLogs">
                    {{ $t('commons.button.refresh') }}
                </el-button>
            </div>
            <pre
                ref="logContainer"
                tabindex="0"
                :aria-label="$t('commons.button.log')"
                class="m-0 mt-2.5 box-border h-[calc(100dvh-300px)] min-h-48 w-full overflow-auto whitespace-pre bg-[var(--panel-logs-bg-color)] p-1.5 font-mono text-sm leading-6 text-[#f5f5f5]"
                >{{ logs || $t('panelMCP.logEmpty') }}</pre>
        </template>
        <template #footer>
            <el-button @click="logsVisible = false">{{ $t('commons.button.cancel') }}</el-button>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { computed, nextTick, onActivated, onBeforeUnmount, onDeactivated, onMounted, reactive, ref, watch } from 'vue';
import { ElMessageBox, type FormInstance } from 'element-plus';
import type { PanelMCP } from '@/api/interface/panel-mcp';
import {
    getPanelMCPLogs,
    getPanelMCPSettings,
    getPanelMCPStatus,
    installPanelMCPService,
    operatePanelMCPService,
    savePanelMCPSettings,
} from '@/api/modules/panel-mcp';
import { Rules } from '@/global/form-rules';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { MsgSuccess } from '@/utils/message';
import { downloadBlob } from '@/utils/file';
import {
    DEFAULT_MCP_DOCKER_IMAGE,
    hasMCPSettingsChanges,
    isMCPDockerImage,
    isMCPRevisionConflict,
    copyMCPSettings,
    parseMCPAllowedIPs,
    isMCPAllowedIPList,
} from '@/utils/panel-mcp';
import i18n from '@/lang';

const emit = defineEmits<{
    endpoint: [value: string];
    status: [value: PanelMCP.ServiceStatus | undefined];
    connection: [];
    changed: [];
}>();
const { globalStore } = useGlobalStore();
const status = ref<PanelMCP.ServiceStatus>();
const canManageService = computed(
    () => globalStore.isLogin && globalStore.isAdmin && status.value?.canManageService === true,
);
const serviceStatusType = computed(() => {
    if (status.value?.ready) return 'success';
    if (status.value?.status === 'stopped' || status.value?.status === 'unhealthy') return 'danger';
    return 'warning';
});
const busy = ref(false);
const statusVisible = ref(false);
const settingsVisible = ref(false);
const installVisible = ref(false);
const drawerMode = ref<'installation' | 'settings'>('settings');
const settingsTab = ref<'configuration' | 'upgrade'>('configuration');
const advancedSettings = ref<string[]>([]);
const configurationVisible = computed({
    get: () => settingsVisible.value || installVisible.value,
    set: (value: boolean) => {
        if (!value) cancelConfiguration();
    },
});
const logsVisible = ref(false);
const logsBusy = ref(false);
const logs = ref('');
const followLogs = ref(false);
const logContainer = ref<HTMLElement>();
const conflict = ref(false);
const installConflict = ref(false);
const settingsRef = ref<FormInstance>();
const installRef = ref<FormInstance>();
const installation = reactive<PanelMCP.Installation>({
    dockerImage: DEFAULT_MCP_DOCKER_IMAGE,
    listenPort: 18081,
    externalUrl: '',
    revision: 0,
});
const installDraft = ref(false);
const installSnapshot = ref<PanelMCP.Settings>();
const installIsUpgrade = ref(false);
const savedSettings = ref<PanelMCP.Settings>();
const defaultSettings = (): PanelMCP.Settings => ({
    runtimeType: 'docker',
    dockerImage: DEFAULT_MCP_DOCKER_IMAGE,
    enabled: false,
    listenPort: 18081,
    externalUrl: '',
    revision: 0,
    accessPolicy: 'inherit',
    allowedIPs: [],
});
const settings = reactive(defaultSettings());
const settingsDirty = computed(() => savedSettings.value && hasMCPSettingsChanges(settings, savedSettings.value));
const allowedIPsText = ref('');
watch(
    allowedIPsText,
    (value) => {
        settings.allowedIPs = parseMCPAllowedIPs(value);
    },
    { flush: 'sync' },
);
const installTitle = computed(() =>
    i18n.global.t(
        (installDraft.value ? installIsUpgrade.value : status.value?.installed)
            ? 'panelMCP.upgradeTitle'
            : 'panelMCP.installTitle',
    ),
);
const endpoint = computed(() => status.value?.endpoint || '');
const serviceOperation = computed<PanelMCP.Operation>(() =>
    status.value?.status === 'stopped' || !status.value?.enabled ? 'start' : 'stop',
);
const canOperateService = computed(
    () =>
        canManageService.value &&
        status.value?.installed &&
        ['stopped', 'running', 'unhealthy', 'incompatible'].includes(status.value.status),
);
const statusKeys: Record<PanelMCP.ServiceStatus['status'], string> = {
    not_installed: 'statusNotInstalled',
    unsupported: 'statusUnsupported',
    stopped: 'statusStopped',
    running: 'statusRunning',
    unhealthy: 'statusUnhealthy',
    incompatible: 'statusIncompatible',
};
const statusLabel = computed(() => i18n.global.t(`panelMCP.${statusKeys[status.value?.status] || 'statusUnknown'}`));
const settingsRules = {
    allowedIPs: [
        {
            validator: (_rule: unknown, value: string[], done: (error?: Error) => void) =>
                done(
                    settings.accessPolicy !== 'restricted' || (value.length > 0 && isMCPAllowedIPList(value))
                        ? undefined
                        : new Error(i18n.global.t('panelMCP.allowedIPsInvalid')),
                ),
            trigger: 'blur',
        },
    ],
    listenPort: [
        Rules.requiredInput,
        {
            validator: (_rule: unknown, value: number, done: (error?: Error) => void) =>
                done(
                    Number.isInteger(value) && value >= 1024 && value <= 65535
                        ? undefined
                        : new Error(i18n.global.t('panelMCP.listenPortInvalid')),
                ),
            trigger: 'blur',
        },
    ],
    externalUrl: [
        {
            validator: (_rule: unknown, value: string, done: (error?: Error) => void) => {
                if (!value) return done();
                try {
                    const url = new URL(value);
                    if (
                        ['http:', 'https:'].includes(url.protocol) &&
                        new TextEncoder().encode(value).length <= 2048 &&
                        !/[\r\n]/.test(value) &&
                        url.pathname === '/mcp' &&
                        !url.username &&
                        !url.password &&
                        !url.search &&
                        !url.hash
                    )
                        return done();
                } catch {
                    /* Report the same validation error for malformed URLs. */
                }
                done(new Error(i18n.global.t('panelMCP.invalidUrl')));
            },
            trigger: 'blur',
        },
    ],
};
const installRules = {
    ...settingsRules,
    dockerImage: [
        {
            validator: (_rule: unknown, value: string, done: (error?: Error) => void) =>
                done(!isMCPDockerImage(value) ? new Error(i18n.global.t('panelMCP.dockerImageInvalid')) : undefined),
            trigger: 'blur',
        },
    ],
};
const onSettingsValidate = async (prop: string | string[], valid: boolean) => {
    if (valid || prop !== 'allowedIPs') return;
    advancedSettings.value = ['source'];
    await nextTick();
    settingsRef.value?.scrollToField(prop);
};
let active = false;
let pageVersion = 0;
let requestVersion = 0;
let logVersion = 0;
let timer: ReturnType<typeof setInterval> | undefined;
let logTimer: ReturnType<typeof setInterval> | undefined;
const isCurrentPage = (version: number) => active && globalStore.isLogin && version === pageVersion;

const refresh = async () => {
    if (!active || !globalStore.isLogin) return;
    const ticket = ++requestVersion;
    try {
        const response = await getPanelMCPStatus();
        if (!active || !globalStore.isLogin || ticket !== requestVersion) return;
        status.value = response.data;
        emit('endpoint', endpoint.value);
        emit('status', status.value);
    } catch {
        if (active && globalStore.isLogin && ticket === requestVersion) {
            status.value = undefined;
            emit('endpoint', '');
            emit('status', undefined);
        }
    }
};
const openStatus = () => {
    if (!active || !globalStore.isLogin) return;
    statusVisible.value = true;
    void refresh();
};
const refreshService = async () => {
    const version = pageVersion;
    await refresh();
    if (isCurrentPage(version)) emit('changed');
};
const closeWhenIdle = (done: () => void) => {
    if (!busy.value) done();
};
const cancelSettings = () => {
    if (busy.value) return;
    if (savedSettings.value) Object.assign(settings, savedSettings.value);
    allowedIPsText.value = settings.allowedIPs.join('\n');
    settingsVisible.value = false;
};
const closeSettingsWhenIdle = (done: () => void) => {
    if (busy.value) return;
    cancelSettings();
    done();
};
const cancelConfiguration = () => {
    if (busy.value) return;
    if (settingsVisible.value) cancelSettings();
    else installVisible.value = false;
};
const closeConfigurationWhenIdle = (done: () => void) => {
    if (settingsVisible.value) closeSettingsWhenIdle(done);
    else closeWhenIdle(done);
};
const changeSettingsTab = async (name: string | number) => {
    if (name === settingsTab.value) return true;
    if (busy.value) return false;
    const version = pageVersion;
    try {
        if (name === 'upgrade') await openInstall();
        else if (name === 'configuration') await openSettings();
        else return false;
        return isCurrentPage(version) && settingsTab.value === name;
    } catch {
        return false;
    }
};
const reloadSettings = async (version = pageVersion) => {
    const response = await getPanelMCPSettings();
    if (!isCurrentPage(version)) return;
    const loaded = copyMCPSettings(response.data);
    savedSettings.value = { ...loaded };
    Object.assign(settings, loaded);
    allowedIPsText.value = settings.allowedIPs.join('\n');
    conflict.value = false;
};
const copyInstallation = (source: PanelMCP.Settings, upgrading: boolean) => {
    installSnapshot.value = { ...source };
    installIsUpgrade.value = upgrading;
    const { dockerImage, listenPort, externalUrl, revision } = source;
    Object.assign(installation, { dockerImage, listenPort, externalUrl, revision });
    installDraft.value = true;
    installConflict.value = false;
};
const openInstall = async () => {
    if (busy.value || !canManageService.value) return;
    const version = pageVersion;
    busy.value = true;
    try {
        if (!installDraft.value) {
            const upgrading = !!status.value?.installed;
            const response = await getPanelMCPSettings();
            if (!isCurrentPage(version)) return;
            copyInstallation(copyMCPSettings(response.data), upgrading);
        }
        if (!isCurrentPage(version) || !canManageService.value) return;
        drawerMode.value = installIsUpgrade.value ? 'settings' : 'installation';
        settingsTab.value = 'upgrade';
        settingsVisible.value = false;
        installVisible.value = true;
    } finally {
        busy.value = false;
    }
};
const reloadInstallation = async () => {
    if (busy.value || !canManageService.value) return;
    const version = pageVersion;
    try {
        await ElMessageBox.confirm(i18n.global.t('panelMCP.reloadInstallConfirm'), installTitle.value);
    } catch {
        return;
    }
    if (!isCurrentPage(version) || busy.value || !canManageService.value) return;
    busy.value = true;
    try {
        const upgrading = !!status.value?.installed;
        const response = await getPanelMCPSettings();
        if (isCurrentPage(version)) {
            copyInstallation(copyMCPSettings(response.data), upgrading);
            drawerMode.value = upgrading ? 'settings' : 'installation';
            settingsTab.value = 'upgrade';
        }
    } finally {
        busy.value = false;
    }
};
const openSettings = async () => {
    if (busy.value || !canManageService.value || !status.value?.installed) return;
    const version = pageVersion;
    busy.value = true;
    try {
        if (!settingsDirty.value) await reloadSettings(version);
        if (!isCurrentPage(version)) return;
        drawerMode.value = 'settings';
        settingsTab.value = 'configuration';
        advancedSettings.value = [];
        installVisible.value = false;
        settingsVisible.value = true;
    } finally {
        busy.value = false;
    }
};
const saveSettings = async () => {
    if (busy.value || conflict.value || !canManageService.value || !savedSettings.value) return;
    const version = pageVersion;
    busy.value = true;
    try {
        if (!(await settingsRef.value?.validate().catch(() => false))) return;
        if (!isCurrentPage(version)) return;
        if (settings.accessPolicy === 'all' && savedSettings.value.accessPolicy !== 'all') {
            await ElMessageBox.confirm(
                i18n.global.t('panelMCP.accessAllConfirm'),
                i18n.global.t('panelMCP.accessPolicy'),
            );
        }
        if (!isCurrentPage(version) || !canManageService.value) return;
        const { runtimeType, dockerImage, enabled, revision } = savedSettings.value;
        const { listenPort, externalUrl, accessPolicy } = settings;
        const allowedIPs = accessPolicy === 'restricted' ? [...settings.allowedIPs] : [];
        await savePanelMCPSettings({
            runtimeType,
            dockerImage,
            enabled,
            listenPort,
            externalUrl,
            revision,
            accessPolicy,
            allowedIPs,
        });
        if (!isCurrentPage(version)) return;
        await reloadSettings(version);
        if (!isCurrentPage(version)) return;
        emit('changed');
        settingsVisible.value = false;
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        await refresh();
    } catch (error) {
        if (isCurrentPage(version)) conflict.value = isMCPRevisionConflict(error);
    } finally {
        busy.value = false;
    }
};
const operate = async (operation: PanelMCP.Operation) => {
    if (busy.value || !canOperateService.value) return;
    const version = pageVersion;
    try {
        await ElMessageBox.confirm(
            i18n.global.t('panelMCP.serviceOperateConfirm', [i18n.global.t(`commons.operate.${operation}`)]),
            i18n.global.t('panelMCP.title'),
        );
    } catch {
        return;
    }
    if (!isCurrentPage(version) || busy.value || !canOperateService.value) return;
    busy.value = true;
    try {
        await operatePanelMCPService(operation);
        if (!isCurrentPage(version)) return;
        emit('changed');
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    } finally {
        busy.value = false;
        if (isCurrentPage(version)) await refresh();
    }
};
const install = async () => {
    if (busy.value || installConflict.value || !canManageService.value || !installSnapshot.value) return;
    const version = pageVersion;
    busy.value = true;
    try {
        if (!(await installRef.value?.validate().catch(() => false))) return;
        if (!isCurrentPage(version) || !canManageService.value) return;
        const { dockerImage } = installation;
        const { listenPort, externalUrl, revision } = installIsUpgrade.value ? installSnapshot.value : installation;
        await installPanelMCPService({ dockerImage, listenPort, externalUrl, revision });
        if (!isCurrentPage(version)) return;
        emit('changed');
        installVisible.value = false;
        installDraft.value = false;
        installSnapshot.value = undefined;
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        if (!settingsDirty.value) await reloadSettings(version);
    } catch (error) {
        if (isCurrentPage(version)) installConflict.value = isMCPRevisionConflict(error);
    } finally {
        busy.value = false;
        if (isCurrentPage(version)) await refresh();
    }
};
const loadLogs = async () => {
    if (!active || !canManageService.value || !logsVisible.value || logsBusy.value) return;
    const ticket = ++logVersion;
    logsBusy.value = true;
    try {
        const response = await getPanelMCPLogs();
        if (active && canManageService.value && logsVisible.value && ticket === logVersion) {
            logs.value = response.data.content;
            await nextTick();
            if (ticket === logVersion && logsVisible.value && followLogs.value && logContainer.value) {
                logContainer.value.scrollTop = logContainer.value.scrollHeight;
            }
        }
    } catch {
        if (ticket === logVersion) stopLogFollowing();
    } finally {
        if (ticket === logVersion) logsBusy.value = false;
    }
};
const stopLogFollowing = () => {
    followLogs.value = false;
    if (logTimer !== undefined) clearInterval(logTimer);
    logTimer = undefined;
};
const clearLogState = () => {
    logVersion++;
    stopLogFollowing();
    logs.value = '';
    logsBusy.value = false;
};
const changeLogFollowing = () => {
    if (logTimer !== undefined) clearInterval(logTimer);
    logTimer = undefined;
    if (!active || !canManageService.value || !logsVisible.value || !followLogs.value) {
        stopLogFollowing();
        return;
    }
    logTimer = setInterval(() => void loadLogs(), 3000);
    void loadLogs();
};
const downloadLogs = () => {
    if (!active || !canManageService.value || !logsVisible.value || !logs.value) return;
    downloadBlob(logs.value, '1panel-mcp.log', 'text/plain;charset=utf-8');
};
const openLogs = () => {
    if (!canManageService.value || !status.value?.installed) return;
    clearLogState();
    logsVisible.value = true;
    void loadLogs();
};
const activate = () => {
    if (active) return;
    active = true;
    void refresh();
    timer = setInterval(() => {
        if (!busy.value) void refresh();
    }, 15000);
};
const clearManagementState = () => {
    pageVersion++;
    clearLogState();
    settingsVisible.value = installVisible.value = logsVisible.value = false;
    savedSettings.value = installSnapshot.value = undefined;
    installDraft.value = installIsUpgrade.value = conflict.value = installConflict.value = false;
    drawerMode.value = 'settings';
    settingsTab.value = 'configuration';
    advancedSettings.value = [];
    Object.assign(settings, defaultSettings());
    allowedIPsText.value = '';
    const { dockerImage, listenPort, externalUrl, revision } = settings;
    Object.assign(installation, { dockerImage, listenPort, externalUrl, revision });
};
watch(
    logsVisible,
    (visible) => {
        if (!visible) clearLogState();
    },
    { flush: 'sync' },
);
const clearStatus = () => {
    requestVersion++;
    status.value = undefined;
    statusVisible.value = false;
};
watch(
    () => [globalStore.isLogin, globalStore.isAdmin],
    ([loggedIn, admin]) => {
        if (!loggedIn || !admin) {
            clearManagementState();
            clearStatus();
            emit('endpoint', '');
            emit('status', undefined);
        }
        if (active && loggedIn) void refresh();
    },
    { flush: 'sync' },
);
watch(
    () => status.value?.canManageService,
    (allowed) => {
        if (allowed === false) clearManagementState();
    },
    { flush: 'sync' },
);
const deactivate = () => {
    active = false;
    clearManagementState();
    clearStatus();
    if (timer) clearInterval(timer);
    timer = undefined;
};
onMounted(activate);
onActivated(activate);
onDeactivated(deactivate);
onBeforeUnmount(deactivate);

defineExpose({ openInstall, openStatus, refreshService, operate, busy });
</script>

<style scoped lang="scss">
.mcp-details {
    :deep(.el-descriptions__table) {
        table-layout: fixed;
    }
    :deep(.el-descriptions__label) {
        white-space: normal;
    }
    :deep(.el-descriptions__cell) {
        overflow-wrap: anywhere;
    }
}
</style>
