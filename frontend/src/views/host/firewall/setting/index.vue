<template>
    <div v-loading="loading">
        <FireRouter />
        <LayoutContent :title="$t('commons.button.set')" :divider="true">
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
                                                <template v-else>
                                                    <el-tag v-if="option.active" size="small" type="success">
                                                        {{ $t('commons.status.running') }}
                                                    </el-tag>
                                                    <el-tag
                                                        v-for="family in backendFamilies"
                                                        :key="family.key"
                                                        size="small"
                                                        :type="option[family.key].initialized ? 'success' : 'info'"
                                                        effect="plain"
                                                    >
                                                        {{ family.label }}
                                                        {{
                                                            $t(
                                                                option[family.key].initialized
                                                                    ? 'firewall.initializedStatus'
                                                                    : 'firewall.notInitialized',
                                                            )
                                                        }}
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
    const previousBackend = savedBackends.value[subsystem];
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
        await ElMessageBox.alert(
            i18n.global.t('firewall.switchBackendSuccessHelper', [backend, previousBackend]),
            i18n.global.t('commons.msg.operationSuccess'),
            {
                confirmButtonText: i18n.global.t('commons.button.confirm'),
                type: 'info',
            },
        );
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
</style>
