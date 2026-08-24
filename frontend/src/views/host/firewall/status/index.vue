<template>
    <div>
        <div class="app-status card-interval" v-if="baseInfo.isExist">
            <el-card>
                <div class="flex w-full flex-col gap-4 md:flex-row">
                    <div class="flex flex-wrap gap-4 ml-3">
                        <el-tag effect="dark" type="success">{{ baseInfo.name }}</el-tag>
                        <Status
                            v-if="isServiceBackend"
                            class="mt-0.5"
                            :status="baseInfo.isActive ? 'enable' : 'disable'"
                        />
                        <el-tag>{{ $t('app.version') }}: {{ baseInfo.version }}</el-tag>
                    </div>
                    <div class="mt-0.5">
                        <template v-if="isServiceBackend">
                            <el-button
                                v-permission
                                v-node-admin
                                type="primary"
                                v-if="baseInfo.isActive"
                                @click="onOperate('stop')"
                                link
                            >
                                {{ $t('commons.button.stop') }}
                            </el-button>
                            <el-tooltip
                                v-if="!baseInfo.isActive"
                                :content="$t('firewall.firewallNotStart')"
                                placement="bottom"
                            >
                                <el-button v-permission v-node-admin type="primary" @click="onOperate('start')" link>
                                    {{ $t('commons.button.start') }}
                                </el-button>
                            </el-tooltip>
                            <el-divider direction="vertical" />
                            <el-button v-permission v-node-admin type="primary" @click="onOperate('restart')" link>
                                {{ $t('commons.button.restart') }}
                            </el-button>
                        </template>
                        <template v-if="isDirectManaged">
                            <el-divider
                                v-if="isDirectBase || !anyFamilyBound || familyIssues.length"
                                direction="vertical"
                            />
                            <template v-if="isDirectBase">
                                <el-button
                                    v-if="anyFamilyBound"
                                    v-permission
                                    v-node-admin
                                    type="primary"
                                    link
                                    @click="onUnBind"
                                >
                                    {{ $t('commons.button.unbind') }}
                                </el-button>
                                <el-button
                                    v-else-if="allAvailableFamiliesInitialized"
                                    v-permission
                                    v-node-admin
                                    type="primary"
                                    link
                                    @click="onBind"
                                >
                                    {{ $t('commons.button.bind') }}
                                </el-button>
                                <el-tooltip v-else :content="initActionHelper" placement="bottom">
                                    <el-button v-permission v-node-admin type="primary" link @click="onInit">
                                        {{ $t('commons.button.init') }}
                                    </el-button>
                                </el-tooltip>
                            </template>
                            <el-tooltip
                                v-else-if="isDirectForward && !anyFamilyBound"
                                :content="initActionHelper"
                                placement="bottom"
                            >
                                <el-button v-permission v-node-admin type="primary" link @click="onInit">
                                    {{ $t('commons.button.init') }}
                                </el-button>
                            </el-tooltip>
                            <el-popover
                                v-if="familyIssues.length"
                                placement="bottom"
                                trigger="hover"
                                :width="300"
                                :show-after="120"
                                :hide-after="120"
                                popper-class="firewall-family-issue-popper"
                            >
                                <template #reference>
                                    <el-button
                                        class="firewall-family-warning-trigger"
                                        type="warning"
                                        link
                                        :aria-label="$t('commons.status.exceptional')"
                                    >
                                        <el-icon><WarningFilled /></el-icon>
                                    </el-button>
                                </template>
                                <div class="firewall-family-issue-list">
                                    <div
                                        v-for="issue in familyIssues"
                                        :key="issue.family"
                                        class="firewall-family-issue-item"
                                    >
                                        <span class="firewall-family-name">{{ issue.family }}</span>
                                        <span class="firewall-family-issue-text">{{ familyIssueLabel(issue) }}</span>
                                    </div>
                                </div>
                                <div v-if="retryableFamilyIssues.length" class="firewall-family-issue-footer">
                                    <el-button
                                        v-permission
                                        v-node-admin
                                        :loading="familyRetrying"
                                        size="small"
                                        type="primary"
                                        @click.stop="onRetryFamilyIssues"
                                    >
                                        {{ $t('commons.button.retry') }}
                                    </el-button>
                                </div>
                            </el-popover>
                        </template>
                    </div>
                </div>
            </el-card>
            <el-alert
                v-if="props.currentTab === 'forward' && baseInfo.syncError"
                class="mt-3"
                type="warning"
                show-icon
                :closable="false"
                :title="baseInfo.syncError"
            />
        </div>
        <NoSuchService
            v-else
            :name="
                props.currentTab === 'forward'
                    ? 'iptables / iptables-nft / nftables'
                    : 'Firewalld / Ufw / iptables / iptables-nft / nftables'
            "
        />

        <DockerRestart
            ref="dockerRef"
            v-model:withDockerRestart="withDockerRestart"
            @submit="onSubmit"
            :title="$t('firewall.firewallHelper', [$t('commons.button.' + operation)])"
        >
            <template #helper>
                <span>{{ $t('firewall.' + operation + 'FirewallHelper') }}</span>
            </template>
        </DockerRestart>
    </div>
</template>

<script lang="ts" setup>
import { Firewall } from '@/api/interface/firewall';
import {
    enableForwarding,
    loadFireBaseInfo,
    loadForwardBaseInfo,
    operateFilterChain,
    operateFire,
} from '@/api/modules/firewall';
import i18n from '@/lang';
import NoSuchService from '@/components/layout-content/no-such-service.vue';
import DockerRestart from '@/components/docker-proxy/docker-restart.vue';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox } from 'element-plus';
import { computed, nextTick, ref } from 'vue';
import { loadDockerStatus } from '@/api/modules/container';
import { WarningFilled } from '@element-plus/icons-vue';

const props = defineProps({
    currentTab: String,
});

const baseInfo = ref<Firewall.FirewallBase>({
    isActive: false,
    isExist: true,
    isInit: false,
    isBind: false,
    name: '',
    backend: '',
    version: '',
    pingStatus: '',
    syncError: '',
    ipv4: { available: false, initialized: false, bound: false },
    ipv6: { available: false, initialized: false, bound: false },
});
const dockerRef = ref();
const operation = ref('restart');
const dockerStatus = ref();
const withDockerRestart = ref(false);
const familyRetrying = ref(false);
const backendName = computed(() => baseInfo.value.backend || baseInfo.value.name);
const isServiceBackend = computed(() => backendName.value === 'firewalld' || backendName.value === 'ufw');
const isDirectBase = computed(
    () => props.currentTab === 'base' && (backendName.value === 'iptables' || backendName.value === 'nftables'),
);
const isDirectForward = computed(
    () => props.currentTab === 'forward' && (backendName.value === 'iptables' || backendName.value === 'nftables'),
);
const isDirectManaged = computed(() => isDirectBase.value || isDirectForward.value);
const familyStatuses = computed(
    () =>
        [
            { family: 'IPv4', status: baseInfo.value.ipv4 },
            { family: 'IPv6', status: baseInfo.value.ipv6 },
        ] as const,
);
const availableFamilies = computed(() => familyStatuses.value.filter((item) => item.status.available));
const anyFamilyInitialized = computed(() => availableFamilies.value.some((item) => item.status.initialized));
const allAvailableFamiliesInitialized = computed(
    () => availableFamilies.value.length > 0 && availableFamilies.value.every((item) => item.status.initialized),
);
const anyFamilyBound = computed(() => availableFamilies.value.some((item) => item.status.bound));
interface FamilyIssue {
    family: 'IPv4' | 'IPv6';
    available: boolean;
    initialized: boolean;
}
const familyIssues = computed<FamilyIssue[]>(() => {
    if (!isDirectManaged.value || !anyFamilyBound.value) return [];
    return familyStatuses.value
        .filter((item) => !item.status.available || !item.status.initialized || !item.status.bound)
        .map((item) => ({
            family: item.family,
            available: item.status.available,
            initialized: item.status.initialized,
        }));
});
const retryableFamilyIssues = computed(() => familyIssues.value.filter((item) => item.available));
const familyIssueLabel = (issue: FamilyIssue) => {
    if (!issue.available) return i18n.global.t('firewall.familyUnsupported', [issue.family]);
    return i18n.global.t(issue.initialized ? 'commons.status.unbind' : 'firewall.notInitialized');
};
const initActionHelper = computed(() => {
    if (props.currentTab === 'forward' && baseInfo.value.isInit && !baseInfo.value.isBind) {
        return `${baseInfo.value.name || backendName.value}: ${i18n.global.t('commons.status.unbind')}`;
    }
    return `${baseInfo.value.name || backendName.value}: ${i18n.global.t('firewall.notInitialized')}`;
});

const acceptParams = (): void => {
    loadBaseInfo(true);
    loadDocker();
};
const emit = defineEmits([
    'search',
    'update:is-active',
    'update:is-bind',
    'update:is-init',
    'update:loading',
    'update:name',
    'update:version',
]);

const loadBaseInfo = async (search: boolean) => {
    const loader = props.currentTab === 'forward' ? loadForwardBaseInfo() : loadFireBaseInfo(props.currentTab);
    await loader
        .then(async (res) => {
            baseInfo.value = {
                ...res.data,
                ipv4: res.data.ipv4 || { available: true, initialized: res.data.isInit, bound: res.data.isBind },
                ipv6: res.data.ipv6 || { available: false, initialized: false, bound: false },
            };
            emit('update:name', backendName.value);
            emit('update:is-active', baseInfo.value.isActive);
            emit('update:is-init', isDirectManaged.value ? anyFamilyInitialized.value : baseInfo.value.isInit);
            emit('update:is-bind', isDirectManaged.value ? anyFamilyBound.value : baseInfo.value.isBind);
            emit('update:version', baseInfo.value.version);

            if (search) {
                await nextTick();
                emit('search');
            } else {
                emit('update:loading', false);
            }
        })
        .catch(() => {
            emit('update:loading', false);
            emit('update:is-init', false);
            emit('update:name', '-');
            emit('update:version', '');
        });
};

const loadDocker = async () => {
    const res = await loadDockerStatus();
    dockerStatus.value = res.data.isExist;
};

const onInit = async () => {
    let chainName = '';
    let msg = '';
    switch (props.currentTab) {
        case 'base':
            chainName = '1PANEL_BASIC';
            msg = i18n.global.t('firewall.initMsg', [baseInfo.value.name || backendName.value]);
            break;
        case 'forward':
            chainName = '1PANEL_FORWARD';
            msg = i18n.global.t('firewall.initMsg', [baseInfo.value.name || backendName.value]);
            break;
        default:
            return;
    }
    try {
        await ElMessageBox.confirm(msg, i18n.global.t('commons.button.init'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        });
    } catch {
        return;
    }
    const initializer =
        props.currentTab === 'forward' ? enableForwarding() : operateFilterChain(chainName, 'init-' + props.currentTab);
    await initializer;
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    await loadBaseInfo(true);
};

const onBind = async () => {
    try {
        await ElMessageBox.confirm(i18n.global.t('firewall.bindHelper'), i18n.global.t('commons.button.bind'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        });
    } catch {
        return;
    }
    await operateFilterChain('1PANEL_BASIC', 'bind-base');
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    await loadBaseInfo(true);
};

const onRetryFamilyIssues = async () => {
    if (familyRetrying.value || retryableFamilyIssues.value.length === 0) return;
    familyRetrying.value = true;
    try {
        if (isDirectForward.value) {
            await enableForwarding();
        } else {
            await operateFilterChain('1PANEL_BASIC', 'bind-base');
        }
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        await loadBaseInfo(true);
    } finally {
        familyRetrying.value = false;
    }
};

const onUnBind = async () => {
    try {
        await ElMessageBox.confirm(i18n.global.t('firewall.unbindHelper'), i18n.global.t('commons.button.unbind'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        });
    } catch {
        return;
    }
    await operateFilterChain('1PANEL_BASIC', 'unbind-base');
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    await loadBaseInfo(true);
};

const onOperate = async (op: string) => {
    operation.value = op;
    if (backendName.value === 'iptables' || backendName.value === 'nftables' || !dockerStatus.value) {
        emit('update:loading', true);
        await operateFire(operation.value, false)
            .then(() => {
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                loadBaseInfo(true);
            })
            .catch(() => {
                loadBaseInfo(true);
            });
    } else {
        dockerRef.value.acceptParams({ title: i18n.global.t('firewall.dockerRestart') });
    }
};

const onSubmit = async () => {
    emit('update:loading', true);
    await operateFire(operation.value, withDockerRestart.value)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            loadBaseInfo(true);
        })
        .catch(() => {
            loadBaseInfo(true);
        });
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss">
.firewall-family-warning-trigger {
    width: 26px;
    height: 26px;
    margin-left: 4px;
    border-radius: 50%;
    background: var(--el-color-warning-light-9);
    font-size: 16px;

    &:hover,
    &:focus-visible {
        background: var(--el-color-warning-light-8);
    }
}

.firewall-family-issue-popper.el-popover {
    padding: 12px;
    border-color: var(--el-color-warning-light-7);
    border-radius: 8px;
    box-shadow: var(--el-box-shadow-light);
}

.firewall-family-issue-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
}

.firewall-family-issue-item {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: center;
    gap: 8px;
    color: var(--el-text-color-regular);
    font-size: 13px;
    line-height: 20px;
}

.firewall-family-issue-footer {
    display: flex;
    justify-content: flex-end;
    margin-top: 12px;
    padding-top: 10px;
    border-top: 1px solid var(--el-border-color-lighter);
}

.firewall-family-name {
    color: var(--el-text-color-primary);
    font-weight: 600;
}

.firewall-family-issue-text {
    color: var(--el-color-warning-dark-2);
    white-space: nowrap;
}
</style>
