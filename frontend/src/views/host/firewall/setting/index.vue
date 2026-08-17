<template>
    <div v-loading="loading">
        <FireRouter />
        <LayoutContent :title="$t('commons.button.set')" :divider="true">
            <template #prompt>
                <el-alert class="mb-4" type="info" :closable="false" :title="$t('firewall.firewallSettingHelper')" />
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
                                        :label="optionLabel(option)"
                                        :value="option.name"
                                        :disabled="!option.installed || !option.supported"
                                    >
                                        <div class="option-row" :title="option.message">
                                            <div class="option-name">
                                                <span>{{ optionLabel(option) }}</span>
                                                <el-tooltip
                                                    v-if="option.name === 'iptables' || option.name === 'nftables'"
                                                    placement="right"
                                                    :content="optionSuggestion(option.name)"
                                                >
                                                    <span class="option-suggestion">
                                                        {{ optionSuggestionTag(option.name) }}
                                                        <el-icon><InfoFilled /></el-icon>
                                                    </span>
                                                </el-tooltip>
                                            </div>
                                            <el-tag v-if="!option.installed" size="small" type="info">
                                                {{ $t('firewall.uninstalledStatus') }}
                                            </el-tag>
                                            <el-tag
                                                v-else-if="
                                                    item.subsystem === 'docker' && item.group.current === option.name
                                                "
                                                size="small"
                                                type="success"
                                            >
                                                {{ $t('firewall.currentUse') }}
                                            </el-tag>
                                            <el-tooltip
                                                v-else-if="
                                                    item.subsystem !== 'docker' && initState(option) === 'partial'
                                                "
                                                placement="right"
                                                :content="familyStatusText(option)"
                                            >
                                                <el-tag size="small" type="warning">
                                                    {{ $t('firewall.partiallyInitialized') }}
                                                </el-tag>
                                            </el-tooltip>
                                            <el-tag
                                                v-else-if="item.subsystem !== 'docker'"
                                                size="small"
                                                :type="initState(option) === 'initialized' ? 'success' : 'info'"
                                            >
                                                {{ initStateText(option) }}
                                            </el-tag>
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
import { loadFirewallSettings, operateFirewallBackend } from '@/api/modules/firewall';
import { operateFire } from '@/api/modules/host';
import FireRouter from '@/views/host/firewall/index.vue';
import WhiteList from '@/views/host/firewall/status/white-list/index.vue';
import { whiteListRuleCount } from '@/views/host/firewall/status/white-list/model';
import { useGlobalStore } from '@/composables/useGlobalStore';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox } from 'element-plus';
import { InfoFilled } from '@element-plus/icons-vue';

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

const groups = computed(() => {
    if (!settings.value) return [];
    return [
        {
            subsystem: 'system' as const,
            label: i18n.global.t('firewall.systemFirewall'),
            helper: i18n.global.t('firewall.systemFirewallHelper'),
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

const optionLabel = (option: Firewall.BackendOption) => {
    return option.name;
};

const optionSuggestionTag = (provider: Firewall.Provider) => {
    return i18n.global.t(provider === 'iptables' ? 'firewall.iptablesSuggestionTag' : 'firewall.nftablesSuggestionTag');
};

const optionSuggestion = (provider: Firewall.Provider) => {
    return i18n.global.t(provider === 'iptables' ? 'firewall.iptablesSuggestion' : 'firewall.nftablesSuggestion');
};

type InitState = 'initialized' | 'partial' | 'uninitialized';

const initState = (option: Firewall.BackendOption): InitState => {
    const ipv4 = option.ipv4.initialized;
    const ipv6 = option.ipv6.initialized;
    if (ipv4 && ipv6) return 'initialized';
    if (ipv4 || ipv6) return 'partial';
    return option.initialized ? 'initialized' : 'uninitialized';
};

const initStateText = (option: Firewall.BackendOption) => {
    return i18n.global.t(
        initState(option) === 'initialized' ? 'firewall.initializedStatus' : 'firewall.notInitialized',
    );
};

const familyStatusText = (option: Firewall.BackendOption) => {
    const status = (initialized: boolean) =>
        i18n.global.t(initialized ? 'firewall.initializedStatus' : 'firewall.notInitialized');
    return `IPv4: ${status(option.ipv4.initialized)} / IPv6: ${status(option.ipv6.initialized)}`;
};

const sortedOptions = (group: Firewall.BackendGroup) => {
    const order: Record<Firewall.Provider, number> = {
        iptables: 0,
        nftables: 1,
        firewalld: 2,
        ufw: 3,
    };
    return [...group.options].sort((left, right) => order[left.name] - order[right.name]);
};

const changeBackend = async (subsystem: Firewall.BackendSubsystem, group: Firewall.BackendGroup) => {
    const backend = group.selected;
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
            restartDocker: subsystem === 'docker',
        });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        await load();
    } catch {
        group.selected = savedBackends.value[subsystem];
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

.option-name,
.option-suggestion {
    display: inline-flex;
    align-items: center;
}

.option-name {
    gap: 8px;
}

.option-suggestion {
    gap: 3px;
    color: var(--el-text-color-secondary);
    font-size: 12px;
}
</style>
