<template>
    <div>
        <FireRouter />
        <DockerGuardStatus :base="data.base" @operate="operate" @cleanup="cleanupBackend" />
        <el-card v-if="data.base.isExist && !data.base.message && !data.base.initialized" class="mask-prompt">
            <span>{{ $t('firewall.initHelper', [data.base.name]) }}</span>
        </el-card>
        <LayoutContent
            v-if="data.base.isExist"
            :title="$t('firewall.dockerGuard')"
            :class="{ mask: !data.base.message && !data.base.initialized }"
            v-loading="loading"
        >
            <template #prompt>
                <el-alert v-if="data.base.message" type="warning" :closable="false" :title="data.base.message" />
                <el-alert v-else type="info" :closable="false" :title="$t('firewall.dockerGuardHelper')" />
            </template>
            <template #leftToolBar>
                <el-button
                    v-if="data.base.backend === 'iptables' || data.base.backend === 'nftables'"
                    v-permission
                    type="primary"
                    v-node-admin
                    @click="openRuleSync"
                >
                    {{ $t('commons.button.sync') }}
                </el-button>
                <el-button v-if="allOrphanPolicies.length" plain @click="orphanDrawerVisible = true">
                    {{ $t('firewall.orphanPolicies') }} ({{ allOrphanPolicies.length }})
                </el-button>
                <el-button
                    v-permission
                    v-node-admin
                    :disabled="selectedPolicyUUIDs.length === 0"
                    @click="removeSelectedPolicies"
                >
                    {{ $t('commons.button.delete') }}
                </el-button>
                <el-button-group>
                    <el-button v-permission v-node-admin @click="openImport">
                        {{ $t('commons.button.import') }}
                    </el-button>
                    <el-button
                        v-permission
                        v-node-admin
                        :disabled="policies.length === 0"
                        @click="exportPoliciesBySelection"
                    >
                        {{ $t('commons.button.export') }}
                    </el-button>
                </el-button-group>
            </template>
            <template #rightToolBar>
                <TableSearch v-model:searchName="searchName" />
                <TableRefresh @search="search" />
            </template>
            <template #main>
                <ComplexTable v-model:selects="selects" :data="containerRows" :heightDiff="370" row-key="key">
                    <el-table-column type="selection" :selectable="hasProtectedEndpoint" width="48" fix />
                    <el-table-column :label="$t('commons.table.name')" min-width="180">
                        <template #default="{ row }">{{ row.name }}</template>
                    </el-table-column>
                    <el-table-column :label="$t('firewall.composeOrApp')" min-width="160">
                        <template #default="{ row }">{{ row.application || row.compose || '-' }}</template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.port')" min-width="400">
                        <template #default="{ row }">
                            <div class="flex flex-wrap items-center gap-2">
                                <el-popover
                                    v-for="group in row.portGroups.slice(0, 5)"
                                    :key="group.key"
                                    placement="top"
                                    trigger="hover"
                                    :width="360"
                                    :show-after="200"
                                >
                                    <template #reference>
                                        <el-tag :type="endpointStatusType(group.endpoint)" effect="plain">
                                            <span v-if="group.endpoint.policyUUID" class="docker-guard-protected-label">
                                                <el-icon><Lock /></el-icon>
                                                <span>{{ group.label }}</span>
                                            </span>
                                            <span v-else>{{ group.label }}</span>
                                        </el-tag>
                                    </template>
                                    <el-descriptions class="docker-guard-descriptions" :column="1" border size="small">
                                        <el-descriptions-item :label="$t('firewall.protectionMode')">
                                            {{
                                                group.endpoint.policyUUID
                                                    ? protectionModeLabel(group.endpoint)
                                                    : $t('firewall.dockerGuardUnprotected')
                                            }}
                                        </el-descriptions-item>
                                        <el-descriptions-item
                                            v-if="
                                                group.endpoint.policyUUID &&
                                                group.endpoint.mode !== 'deny_all' &&
                                                group.endpoint.sources.length
                                            "
                                            :label="$t('firewall.sources')"
                                        >
                                            {{ displaySources(group.endpoint) }}
                                        </el-descriptions-item>
                                        <el-descriptions-item
                                            v-if="group.endpoint.policyUUID"
                                            :label="$t('commons.table.status')"
                                        >
                                            {{
                                                $t(
                                                    group.endpoint.effective
                                                        ? 'firewall.effective'
                                                        : 'firewall.notEffective',
                                                )
                                            }}
                                        </el-descriptions-item>
                                        <el-descriptions-item
                                            v-if="group.endpoint.description"
                                            :label="$t('commons.table.description')"
                                        >
                                            {{ group.endpoint.description }}
                                        </el-descriptions-item>
                                    </el-descriptions>
                                </el-popover>
                                <el-button v-if="row.portGroups.length > 5" plain size="small" @click="openPorts(row)">
                                    +{{ row.portGroups.length - 5 }}
                                </el-button>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.operate')" width="100" fixed="right">
                        <template #default="{ row }">
                            <el-button type="primary" link @click="openPorts(row)">
                                {{ $t('commons.button.view') }}
                            </el-button>
                        </template>
                    </el-table-column>
                </ComplexTable>
            </template>
        </LayoutContent>

        <DrawerPro v-model="orphanDrawerVisible" :header="$t('firewall.orphanPolicies')" size="large">
            <template #content>
                <el-alert
                    type="info"
                    :closable="false"
                    show-icon
                    :title="$t('firewall.orphanPoliciesHelper', [allOrphanPolicies.length])"
                />
                <div class="mt-3">
                    <el-button
                        v-permission
                        v-node-admin
                        :disabled="orphanSelects.length === 0"
                        @click="removePolicies(orphanSelects, true)"
                    >
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </div>
                <ComplexTable
                    :data="orphanRows"
                    row-key="policyUUID"
                    max-height="calc(100vh - 220px)"
                    class="mt-3"
                    @selection-change="changeOrphanSelection"
                >
                    <el-table-column type="selection" width="48" />
                    <el-table-column :label="$t('commons.table.port')" min-width="190">
                        <template #default="{ row }">
                            <span>{{ endpointLabel(row) }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('firewall.protectionMode')" min-width="150">
                        <template #default="{ row }">{{ protectionModeLabel(row) }}</template>
                    </el-table-column>
                    <el-table-column :label="$t('firewall.sources')" min-width="200" show-overflow-tooltip>
                        <template #default="{ row }">{{ displaySources(row) || '-' }}</template>
                    </el-table-column>
                    <el-table-column
                        prop="description"
                        :label="$t('commons.table.description')"
                        min-width="160"
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">{{ row.description || '-' }}</template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.operate')" width="80" fixed="right">
                        <template #default="{ row }">
                            <el-button
                                v-permission
                                v-node-admin
                                type="primary"
                                link
                                @click="removePolicies([row], false)"
                            >
                                {{ $t('commons.button.delete') }}
                            </el-button>
                        </template>
                    </el-table-column>
                </ComplexTable>
            </template>
        </DrawerPro>

        <DockerGuardDetail ref="detailRef" :containers="containerRows" @search="search" />
        <DockerGuardImport ref="importRef" @search="search" />
        <RuleSync ref="ruleSyncRef" @search="search" />
        <ConfirmDialog ref="cleanupConfirmRef" @confirm="submitCleanupBackend" />
    </div>
</template>

<script lang="ts" setup>
import { computed, reactive, ref } from 'vue';
import FireRouter from '@/views/host/firewall/index.vue';
import DockerGuardStatus from '@/views/host/firewall/docker/status/index.vue';
import DockerGuardDetail from '@/views/host/firewall/docker/detail/index.vue';
import DockerGuardImport from '@/views/host/firewall/docker/import/index.vue';
import RuleSync from '@/views/host/firewall/sync/index.vue';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { Firewall } from '@/api/interface/firewall';
import {
    deleteDockerPortGuardPolicies,
    loadDockerPortGuard,
    operateDockerPortGuard,
    operateFirewallBackend,
} from '@/api/modules/firewall';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox } from 'element-plus';
import { Lock } from '@element-plus/icons-vue';
import { downloadWithContent } from '@/utils/file';
import { getCurrentDateFormatted } from '@/utils/date';
import { dockerGuardEndpointKey } from '@/views/host/firewall/docker/model';
import { formatHostAddressList } from '@/views/host/firewall/utils/validation';

const loading = ref(false);
const detailRef = ref<InstanceType<typeof DockerGuardDetail>>();
const importRef = ref<InstanceType<typeof DockerGuardImport>>();
const ruleSyncRef = ref<InstanceType<typeof RuleSync>>();
const cleanupConfirmRef = ref<InstanceType<typeof ConfirmDialog>>();
const orphanDrawerVisible = ref(false);
const searchName = ref('');
const selects = ref<Firewall.DockerGuardContainer[]>([]);
const orphanSelects = ref<Firewall.DockerGuardEndpoint[]>([]);
const data = reactive<Firewall.DockerGuardList>({
    base: {
        name: 'iptables-docker',
        version: '-',
        isExist: true,
        initialized: false,
        bound: false,
        ipv4: { state: 'disabled', initialized: false, bound: false, effective: false },
        ipv6: { state: 'disabled', initialized: false, bound: false, effective: false },
        backend: '',
    },
    containers: [],
    orphanPolicies: [],
});

const allOrphanPolicies = computed(() => {
    const result = new Map<string, Firewall.DockerGuardEndpoint>();
    for (const endpoint of data.orphanPolicies || []) {
        result.set(dockerGuardEndpointKey(endpoint), endpoint);
    }
    for (const container of data.containers || []) {
        if (container.key !== '__orphan__' && container.endpoints.some((endpoint) => endpoint.containerID)) continue;
        for (const endpoint of container.endpoints) {
            if (!endpoint.policyUUID) continue;
            result.set(dockerGuardEndpointKey(endpoint), endpoint);
        }
    }
    return [...result.values()];
});

const containerRows = computed(() => {
    const keyword = searchName.value.trim().toLowerCase();

    return data.containers
        .filter((container) => container.key !== '__orphan__')
        .map((container) => {
            const name = container.name || i18n.global.t('firewall.orphanEndpoints');
            const containerMatches = [name, container.application, container.compose]
                .filter(Boolean)
                .some((item) => item!.toLowerCase().includes(keyword));
            const endpointMatches = (endpoint: Firewall.DockerGuardEndpoint) => {
                if (!keyword || containerMatches) return true;
                return [endpoint.family, endpoint.hostIP, endpoint.hostPort, endpoint.containerPort, endpoint.protocol]
                    .filter((item) => item !== undefined)
                    .some((item) => String(item).toLowerCase().includes(keyword));
            };
            const portGroups = container.portGroups.flatMap((group) => {
                const endpoints = group.endpoints.filter(endpointMatches);
                if (!endpoints.length) return [];
                if (endpoints.length === group.endpoints.length) return [group];
                return endpoints.map((endpoint) => ({
                    key: `${group.key}|${endpoint.hostPort}|${endpoint.containerPort || 0}`,
                    label: endpointLabel(endpoint),
                    endpoint,
                    endpoints: [endpoint],
                }));
            });
            return {
                ...container,
                name,
                endpoints: container.endpoints.filter(endpointMatches),
                portGroups,
            };
        })
        .filter((container) => container.portGroups.length > 0);
});

const orphanRows = computed(() => {
    const keyword = searchName.value.trim().toLowerCase();
    return allOrphanPolicies.value.filter((endpoint) => {
        if (!keyword) return true;
        return [
            endpoint.family,
            endpoint.hostIP,
            endpoint.hostPort,
            endpoint.protocol,
            endpoint.mode,
            endpoint.description,
            ...(endpoint.sources || []),
        ]
            .filter((item) => item !== undefined)
            .some((item) => String(item).toLowerCase().includes(keyword));
    });
});

const policies = computed<Firewall.DockerGuardPolicy[]>(() => {
    const result = new Map<string, Firewall.DockerGuardPolicy>();
    for (const container of data.containers) {
        for (const endpoint of container.endpoints) {
            if (!endpoint.policyUUID || !endpoint.mode) continue;
            const key = dockerGuardEndpointKey(endpoint);
            result.set(key, {
                family: endpoint.family,
                hostIP: endpoint.hostIP,
                hostPort: endpoint.hostPort,
                protocol: endpoint.protocol,
                mode: endpoint.mode,
                sources: endpoint.sources || [],
                description: endpoint.description || '',
            });
        }
    }
    for (const endpoint of allOrphanPolicies.value) {
        if (!endpoint.policyUUID || !endpoint.mode) continue;
        result.set(dockerGuardEndpointKey(endpoint), {
            family: endpoint.family,
            hostIP: endpoint.hostIP,
            hostPort: endpoint.hostPort,
            protocol: endpoint.protocol,
            mode: endpoint.mode,
            sources: endpoint.sources || [],
            description: endpoint.description || '',
        });
    }
    return [...result.values()];
});

const policyUUIDs = (endpoints: Firewall.DockerGuardEndpoint[]) => [
    ...new Set(endpoints.map((endpoint) => endpoint.policyUUID).filter(Boolean) as string[]),
];
const selectedEndpoints = computed(() => selects.value.flatMap((container) => container.endpoints));
const selectedPolicyUUIDs = computed(() => policyUUIDs(selectedEndpoints.value));
const hasProtectedEndpoint = (container: Firewall.DockerGuardContainer) =>
    container.endpoints.some((endpoint) => Boolean(endpoint.policyUUID));
const policiesFromEndpoints = (endpoints: Firewall.DockerGuardEndpoint[]) => {
    const result = new Map<string, Firewall.DockerGuardPolicy>();
    for (const endpoint of endpoints) {
        if (!endpoint.policyUUID || !endpoint.mode) continue;
        result.set(dockerGuardEndpointKey(endpoint), {
            family: endpoint.family,
            hostIP: endpoint.hostIP,
            hostPort: endpoint.hostPort,
            protocol: endpoint.protocol,
            mode: endpoint.mode,
            sources: endpoint.sources || [],
            description: endpoint.description || '',
        });
    }
    return [...result.values()];
};

const endpointLabel = (endpoint: Firewall.DockerGuardEndpoint) => {
    const address = endpoint.hostIP.includes(':') ? `[${endpoint.hostIP}]` : endpoint.hostIP;
    return `${address}:${endpoint.hostPort}/${endpoint.protocol}`;
};
const displaySources = (endpoint: Firewall.DockerGuardEndpoint) =>
    formatHostAddressList(endpoint.sources, endpoint.family);

const search = async () => {
    loading.value = true;
    try {
        Object.assign(data, (await loadDockerPortGuard()).data);
        selects.value = [];
        orphanSelects.value = [];
    } finally {
        loading.value = false;
    }
};
const operate = async (operation: 'initialize' | 'bind' | 'unbind') => {
    if (operation === 'unbind') {
        try {
            await ElMessageBox.confirm(
                i18n.global.t('firewall.dockerGuardUnbindConfirm'),
                i18n.global.t('commons.button.unbind'),
                { type: 'warning' },
            );
        } catch {
            return;
        }
    }
    try {
        await operateDockerPortGuard(operation);
    } catch {
        await search();
        return;
    }
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    await search();
};
const cleanupBackend = () => {
    if (data.base.backend !== 'iptables' && data.base.backend !== 'nftables') return;
    cleanupConfirmRef.value?.acceptParams({
        header: i18n.global.t('firewall.cleanupAction'),
        operationInfo: i18n.global.t('firewall.cleanupDockerBackendHelper', [data.base.backend]),
        submitInputInfo: data.base.backend,
    });
};

const submitCleanupBackend = async () => {
    if (data.base.backend !== 'iptables' && data.base.backend !== 'nftables') return;
    loading.value = true;
    try {
        await operateFirewallBackend({ subsystem: 'docker', backend: data.base.backend, operation: 'cleanup' });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        await search();
    } finally {
        loading.value = false;
    }
};
const openRuleSync = () => {
    if (data.base.backend !== 'iptables' && data.base.backend !== 'nftables') return;
    ruleSyncRef.value?.acceptParams(data.base.backend, 'docker');
};
const openImport = () => {
    importRef.value?.acceptParams();
};
const exportPolicies = async (items: Firewall.DockerGuardPolicy[]) => {
    if (!items.length) return;
    try {
        await ElMessageBox.confirm(
            i18n.global.t('firewall.exportHelper', [items.length]),
            i18n.global.t('commons.button.export'),
        );
    } catch {
        return;
    }
    downloadWithContent(JSON.stringify(items, null, 2), `1panel-docker-port-guard-${getCurrentDateFormatted()}.json`);
};
const exportPoliciesBySelection = () => {
    const selected = policiesFromEndpoints(selectedEndpoints.value);
    return exportPolicies(selected.length > 0 ? selected : policies.value);
};
const openPorts = (row: Firewall.DockerGuardContainer) => {
    detailRef.value?.acceptParams(row);
};
const changeOrphanSelection = (rows: Firewall.DockerGuardEndpoint[]) => {
    orphanSelects.value = rows;
};
const removePolicies = async (endpoints: Firewall.DockerGuardEndpoint[], batch: boolean) => {
    const uuids = policyUUIDs(endpoints);
    if (!uuids.length) return;
    try {
        await ElMessageBox.confirm(
            i18n.global.t(
                batch ? 'firewall.clearDockerGuardPoliciesConfirm' : 'firewall.deleteDockerGuardPolicyConfirm',
            ),
            i18n.global.t('commons.button.delete'),
            { type: 'warning' },
        );
    } catch {
        return;
    }
    loading.value = true;
    try {
        for (let offset = 0; offset < uuids.length; offset += 256) {
            await deleteDockerPortGuardPolicies({ uuids: uuids.slice(offset, offset + 256) });
        }
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        await search();
    } catch {
        await search();
    } finally {
        loading.value = false;
    }
};
const removeSelectedPolicies = () => removePolicies(selectedEndpoints.value, true);
const protectionModeLabel = (row: Firewall.DockerGuardEndpoint) => {
    if (row.mode === 'deny_sources') return i18n.global.t('firewall.denySources');
    if (row.mode === 'allow_sources') return i18n.global.t('firewall.allowSources');
    return i18n.global.t('firewall.denyAll');
};
const endpointStatusType = (row: Firewall.DockerGuardEndpoint) => {
    if (row.effective) return 'success';
    return row.policyUUID ? 'warning' : 'info';
};
search();
</script>

<style lang="scss" scoped>
.docker-guard-protected-label {
    display: inline-flex;
    align-items: flex-end;
    gap: 1px;
}

.docker-guard-descriptions :deep(.el-descriptions__label) {
    width: 96px;
    min-width: 96px;
    max-width: 96px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}
</style>
