<template>
    <div v-loading="loading">
        <FireRouter />
        <LayoutContent :title="$t('commons.button.set')">
            <template #prompt>
                <el-alert type="warning" show-icon :closable="false" :title="$t('firewall.backendSwitchNotice')" />
            </template>
            <template #main>
                <el-form :label-position="isMobile ? 'top' : 'left'" label-width="150px">
                    <el-row>
                        <el-col :span="1"><br /></el-col>
                        <el-col :xs="24" :sm="20" :md="15" :lg="12" :xl="12">
                            <el-form-item v-for="item in groups" :key="item.subsystem" :label="item.label">
                                <el-select
                                    v-model="item.group.selected"
                                    v-permission
                                    v-node-admin
                                    class="sm:!w-1/2 !w-full"
                                    @change="changeBackend(item.subsystem, item.group)"
                                >
                                    <el-option
                                        v-for="option in sortedOptions(item.group)"
                                        :key="option.name"
                                        :label="option.name"
                                        :value="option.name"
                                        :disabled="!option.installed || !option.supported"
                                    >
                                        <div class="option-row">
                                            <span>{{ option.name }}</span>
                                            <div class="option-status">
                                                <el-tag v-if="!option.installed" size="small" type="info">
                                                    {{ $t('firewall.uninstalledStatus') }}
                                                </el-tag>
                                                <el-popover
                                                    v-else-if="option.message"
                                                    placement="right"
                                                    trigger="hover"
                                                    :width="220"
                                                >
                                                    <template #reference>
                                                        <el-tag size="small" type="warning">
                                                            <el-icon><WarningFilled /></el-icon>
                                                            {{ $t('commons.status.exceptional') }}
                                                        </el-tag>
                                                    </template>
                                                    <div class="backend-issue-list">
                                                        <div class="backend-issue-item">{{ option.message }}</div>
                                                    </div>
                                                </el-popover>
                                                <template v-else-if="isServiceBackend(option.name)">
                                                    <el-tag size="small" :type="option.active ? 'success' : 'info'">
                                                        {{
                                                            $t(
                                                                option.active
                                                                    ? 'commons.status.running'
                                                                    : 'commons.status.stopped',
                                                            )
                                                        }}
                                                    </el-tag>
                                                </template>
                                                <template v-else>
                                                    <el-popover
                                                        v-if="showBackendFamilyDetails(option)"
                                                        placement="right"
                                                        trigger="hover"
                                                        :width="360"
                                                    >
                                                        <template #reference>
                                                            <el-tag
                                                                size="small"
                                                                :type="backendDisplayState(option).type"
                                                            >
                                                                <el-icon v-if="backendDisplayState(option).exceptional">
                                                                    <WarningFilled />
                                                                </el-icon>
                                                                {{ $t(backendDisplayState(option).label) }}
                                                            </el-tag>
                                                        </template>
                                                        <div class="backend-issue-list">
                                                            <div
                                                                v-for="familyState in backendFamilyStates(option)"
                                                                :key="familyState.family"
                                                                class="backend-issue-item"
                                                            >
                                                                {{
                                                                    backendFamilyStatusLabel(
                                                                        familyState.family,
                                                                        familyState.status,
                                                                    )
                                                                }}
                                                            </div>
                                                        </div>
                                                    </el-popover>
                                                    <el-tag
                                                        v-else
                                                        size="small"
                                                        :type="backendDisplayState(option).type"
                                                    >
                                                        {{ $t(backendDisplayState(option).label) }}
                                                    </el-tag>
                                                </template>
                                            </div>
                                        </div>
                                    </el-option>
                                </el-select>
                                <span class="input-help">{{ item.helper }}</span>
                            </el-form-item>

                            <el-form-item :label="$t('firewall.noPing')">
                                <el-radio-group v-model="pingStatus" @change="changePing">
                                    <el-radio-button v-permission v-node-admin value="Enable">
                                        {{ $t('commons.button.enable') }}
                                    </el-radio-button>
                                    <el-radio-button v-permission v-node-admin value="Disable">
                                        {{ $t('commons.button.disable') }}
                                    </el-radio-button>
                                </el-radio-group>
                            </el-form-item>
                            <el-form-item :label="$t('firewall.portWhiteList')">
                                <div class="flex items-center gap-3">
                                    <el-button
                                        v-permission
                                        v-node-admin
                                        icon="Setting"
                                        @click="whiteListRef.acceptParams()"
                                    >
                                        {{ $t('commons.button.set') }}
                                    </el-button>
                                    <span class="input-help !mt-0">
                                        {{ $t('firewall.configuredRules', [whiteListCount]) }}
                                    </span>
                                </div>
                            </el-form-item>
                        </el-col>
                    </el-row>
                </el-form>
            </template>
        </LayoutContent>
        <WhiteList ref="whiteListRef" @search="load" />
    </div>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';
import { Firewall } from '@/api/interface/firewall';
import { loadFirewallSettings, operateFire, operateFirewallBackend } from '@/api/modules/firewall';
import FireRouter from '@/views/host/firewall/index.vue';
import WhiteList from '@/views/host/firewall/setting/white-list/index.vue';
import { whiteListRuleCount } from '@/views/host/firewall/setting/white-list/model';
import { useGlobalStore } from '@/composables/useGlobalStore';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox } from 'element-plus';
import { WarningFilled } from '@element-plus/icons-vue';

const { isMobile } = useGlobalStore();
const loading = ref(false);
const settings = ref<Firewall.Settings>();
const savedBackends = ref<Record<Firewall.BackendSubsystem, string>>({
    system: '',
    forwarding: '',
    docker: '',
});
const pingStatus = ref('Disable');
const oldPingStatus = ref('Disable');
const whiteListRef = ref();

const backendFamilies = [
    { key: 'ipv4', label: 'IPv4' },
    { key: 'ipv6', label: 'IPv6' },
] as const;

const providerOrder: Record<Firewall.Provider, number> = {
    iptables: 0,
    nftables: 1,
    firewalld: 2,
    ufw: 3,
};

const isServiceBackend = (provider: Firewall.Provider) => {
    return provider === 'firewalld' || provider === 'ufw';
};

type BackendTagType = 'success' | 'info' | 'warning';
interface BackendDisplayState {
    label: string;
    type: BackendTagType;
    exceptional: boolean;
}

const backendFamilyStates = (option: Firewall.BackendOption) =>
    backendFamilies.map((family) => ({ family: family.label, status: option[family.key] }));

const backendDisplayState = (option: Firewall.BackendOption): BackendDisplayState => {
    if (option.message) {
        return { label: 'commons.status.exceptional', type: 'warning', exceptional: true };
    }
    const families = backendFamilyStates(option);
    if (families.every(({ status }) => !status.initialized && !status.bound)) {
        return { label: 'firewall.notInitialized', type: 'info', exceptional: false };
    }
    if (families.every(({ status }) => status.available && status.initialized && status.bound && !status.reason)) {
        return { label: 'commons.status.bound', type: 'success', exceptional: false };
    }
    if (families.every(({ status }) => status.available && status.initialized && !status.bound && !status.reason)) {
        return { label: 'commons.status.unbind', type: 'info', exceptional: false };
    }
    return { label: 'commons.status.exceptional', type: 'warning', exceptional: true };
};

const showBackendFamilyDetails = (option: Firewall.BackendOption) =>
    backendDisplayState(option).label !== 'commons.status.bound' ||
    backendFamilyStates(option).some(({ status }) => !status.available || Boolean(status.reason));

const dockerGuardReasons = new Set([
    'docker_chain_missing',
    'guard_chain_missing',
    'jump_missing',
    'jump_not_first',
    'jump_duplicate',
    'inspect_failed',
]);

const backendFamilyStatusLabel = (family: 'IPv4' | 'IPv6', status: Firewall.BackendFamilyStatus) => {
    if (!status.available || status.reason === 'command_missing') {
        return i18n.global.t('firewall.familyUnsupported', [family]);
    }
    if (status.reason && dockerGuardReasons.has(status.reason)) {
        return i18n.global.t(`firewall.dockerGuardStatusReason.${status.reason}`, [family]);
    }
    return `${family} ${i18n.global.t(status.initialized ? (status.bound ? 'commons.status.bound' : 'commons.status.unbind') : 'firewall.notInitialized')}`;
};

const groups = computed(() => {
    if (!settings.value) return [];
    return [
        {
            subsystem: 'system' as const,
            label: i18n.global.t('firewall.systemFirewall'),
            helper: `${i18n.global.t('firewall.systemFirewallHelper')} ${i18n.global.t(
                'firewall.backendRecommendation',
            )}`,
            group: settings.value.system,
        },
        {
            subsystem: 'docker' as const,
            label: i18n.global.t('firewall.dockerGuard'),
            helper: i18n.global.t('firewall.dockerFirewallHelper'),
            group: settings.value.docker,
        },
        {
            subsystem: 'forwarding' as const,
            label: i18n.global.t('firewall.forwardRule', 2),
            helper: i18n.global.t('firewall.forwardingHelper'),
            group: settings.value.forwarding,
        },
    ];
});

const whiteListCount = computed(() => whiteListRuleCount(settings.value?.portWhiteList || ''));

const sortedOptions = (group: Firewall.BackendGroup) => {
    return [...group.options].sort((left, right) => providerOrder[left.name] - providerOrder[right.name]);
};

const load = async () => {
    loading.value = true;
    try {
        const res = await loadFirewallSettings();
        settings.value = res.data;
        savedBackends.value = {
            system: res.data.system.selected,
            forwarding: res.data.forwarding.selected,
            docker: res.data.docker.selected,
        };
        pingStatus.value = res.data.pingStatus;
        oldPingStatus.value = pingStatus.value;
    } finally {
        loading.value = false;
    }
};

const changeBackend = async (subsystem: Firewall.BackendSubsystem, group: Firewall.BackendGroup) => {
    const backend = group.selected;
    const previous = savedBackends.value[subsystem];
    const previousOption = group.options.find((option) => option.name === previous);
    const previousInitialized =
        previousOption?.initialized || previousOption?.ipv4.initialized || previousOption?.ipv6.initialized;
    if (subsystem !== 'system' && previous && previous !== backend && previousInitialized) {
        group.selected = previous;
        await ElMessageBox.alert(
            i18n.global.t('firewall.cleanupBeforeBackendSwitch', [previous, backend]),
            i18n.global.t('firewall.cleanupAction'),
            { type: 'warning' },
        );
        return;
    }
    try {
        await ElMessageBox.confirm(
            i18n.global.t('firewall.switchBackendHelper', [backend]),
            i18n.global.t('commons.button.confirm'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                cancelButtonText: i18n.global.t('commons.button.cancel'),
                type: 'info',
            },
        );
        loading.value = true;
        await operateFirewallBackend({
            subsystem,
            backend: backend as Firewall.Provider,
            operation: 'select',
        });
        await load();
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    } catch {
        if (savedBackends.value[subsystem] !== backend) {
            group.selected = savedBackends.value[subsystem];
        }
    } finally {
        loading.value = false;
    }
};

const changePing = async (value: string | number | boolean) => {
    const next = String(value);
    try {
        await operateFire(next === 'Enable' ? 'enableBanPing' : 'disableBanPing', false);
        oldPingStatus.value = next;
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    } catch {
        pingStatus.value = oldPingStatus.value;
    }
};

load();
</script>

<style scoped>
.option-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
}

.option-status {
    display: inline-flex;
    align-items: center;
    justify-content: flex-end;
    gap: 5px;
    white-space: nowrap;
    font-size: 12px;
}

.backend-issue-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
}

.backend-issue-item {
    color: var(--el-text-color-regular);
    font-size: 13px;
    line-height: 20px;
}
</style>
