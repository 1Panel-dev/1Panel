<template>
    <div>
        <FireRouter />
        <DockerGuardStatus :base="data.base" @operate="operate" />
        <LayoutContent :title="$t('firewall.dockerGuard')" v-loading="loading">
            <template #prompt>
                <el-alert v-if="data.base.message" type="warning" :closable="false" :title="data.base.message" />
                <el-alert v-else type="info" :closable="false" :title="$t('firewall.dockerGuardHelper')" />
            </template>
            <template #leftToolBar>
                <el-button v-if="data.base.initialized" v-permission v-node-admin type="primary" @click="sync">
                    {{ $t('commons.button.sync') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <div class="firewall-filter-bar">
                    <el-select
                        v-model="selectedFilters"
                        class="firewall-rule-filter"
                        :placeholder="$t('menu.filter')"
                        multiple
                        clearable
                        collapse-tags
                        collapse-tags-tooltip
                        :max-collapse-tags="4"
                        popper-class="firewall-rule-filter-popper"
                    >
                        <el-option-group label="IP">
                            <el-option label="IPv4" value="family:ipv4" />
                            <el-option label="IPv6" value="family:ipv6" />
                        </el-option-group>
                        <el-option-group :label="$t('firewall.exposure')">
                            <el-option :label="$t('firewall.externallyExposed')" value="exposure:external" />
                            <el-option :label="$t('firewall.restrictedBinding')" value="exposure:restricted" />
                        </el-option-group>
                    </el-select>
                </div>
                <TableSearch v-model:searchName="searchName" />
                <TableRefresh @search="search" />
            </template>
            <template #main>
                <ComplexTable :data="containerRows" :heightDiff="360">
                    <el-table-column :label="$t('commons.table.name')" min-width="180">
                        <template #default="{ row }">{{ row.name }}</template>
                    </el-table-column>
                    <el-table-column :label="$t('firewall.composeOrApp')" min-width="160">
                        <template #default="{ row }">{{ row.application || row.compose || '-' }}</template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.port')" min-width="400">
                        <template #default="{ row }">
                            <div class="flex flex-wrap items-center gap-2">
                                <el-tooltip
                                    v-for="group in row.portGroups.slice(0, 5)"
                                    :key="group.key"
                                    :content="protectionDetail(group.endpoint)"
                                    placement="top"
                                >
                                    <el-tag :type="endpointStatusType(group.endpoint)" effect="plain">
                                        {{ group.label }}
                                    </el-tag>
                                </el-tooltip>
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

        <DockerGuardDetail ref="detailRef" :containers="containerRows" @search="search" />
    </div>
</template>

<script lang="ts" setup>
import { computed, reactive, ref } from 'vue';
import FireRouter from '@/views/host/firewall/index.vue';
import DockerGuardStatus from '@/views/host/firewall/docker/status/index.vue';
import DockerGuardDetail from '@/views/host/firewall/docker/detail/index.vue';
import { Firewall } from '@/api/interface/firewall';
import { loadDockerPortGuard, operateDockerPortGuard, syncDockerPortGuard } from '@/api/modules/firewall';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox } from 'element-plus';

const loading = ref(false);
const detailRef = ref<InstanceType<typeof DockerGuardDetail>>();
const searchName = ref('');
const selectedFilters = ref<string[]>([]);
const data = reactive<Firewall.DockerGuardList>({
    base: {
        initialized: false,
        bound: false,
        ipv4: { state: 'disabled', initialized: false, bound: false, effective: false },
        ipv6: { state: 'disabled', initialized: false, bound: false, effective: false },
        backend: '',
    },
    containers: [],
});

const containerRows = computed(() => {
    const keyword = searchName.value.trim().toLowerCase();
    const families = selectedFilters.value
        .filter((item) => item.startsWith('family:'))
        .map((item) => item.slice('family:'.length));
    const exposures = selectedFilters.value
        .filter((item) => item.startsWith('exposure:'))
        .map((item) => item.slice('exposure:'.length));

    return data.containers
        .map((container) => {
            const name = container.name || i18n.global.t('firewall.orphanEndpoints');
            const containerMatches = [name, container.application, container.compose]
                .filter(Boolean)
                .some((item) => item!.toLowerCase().includes(keyword));
            const endpointMatches = (endpoint: Firewall.DockerGuardEndpoint) => {
                if (families.length && !families.includes(endpoint.family)) return false;
                const exposure = isExternallyExposed(endpoint) ? 'external' : 'restricted';
                if (exposures.length && !exposures.includes(exposure)) return false;
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

const isExternallyExposed = (endpoint: Firewall.DockerGuardEndpoint) =>
    (endpoint.family === 'ipv4' && (!endpoint.hostIP || endpoint.hostIP === '0.0.0.0')) ||
    (endpoint.family === 'ipv6' && (!endpoint.hostIP || endpoint.hostIP === '::'));
const endpointLabel = (endpoint: Firewall.DockerGuardEndpoint) => {
    const address = endpoint.hostIP.includes(':') ? `[${endpoint.hostIP}]` : endpoint.hostIP;
    return `${address}:${endpoint.hostPort}/${endpoint.protocol}`;
};

const search = async () => {
    loading.value = true;
    try {
        Object.assign(data, (await loadDockerPortGuard()).data);
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
    await operateDockerPortGuard(operation);
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    await search();
};
const sync = async () => {
    try {
        await syncDockerPortGuard();
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    } finally {
        await search();
    }
};
const openPorts = (row: Firewall.DockerGuardContainer) => {
    detailRef.value?.acceptParams(row);
};
const endpointStatusType = (row: Firewall.DockerGuardEndpoint) => {
    if (row.effective) return 'success';
    return row.policyUUID ? 'warning' : 'info';
};
const protectionDetail = (row: Firewall.DockerGuardEndpoint) => {
    if (!row.policyUUID) return i18n.global.t('firewall.dockerGuardUnprotected');
    let detail = i18n.global.t('firewall.denyAll');
    if (row.mode === 'deny_sources') {
        detail = `${i18n.global.t('firewall.deniedSources')}: ${row.sources.join(', ')}`;
    }
    if (row.mode === 'allow_sources' && row.sources.length) {
        detail = `${i18n.global.t('firewall.allowedSources')}: ${row.sources.join(', ')}`;
    }
    if (!row.effective) {
        detail = `${detail} · ${i18n.global.t('firewall.notEffective')}`;
    }
    return row.description ? `${detail}\n${row.description}` : detail;
};
search();
</script>

<style lang="scss" scoped>
.firewall-filter-bar {
    display: inline-flex;
    flex: none;
    flex-wrap: nowrap;
    gap: 8px;
}

.firewall-rule-filter {
    width: 480px;
}

:global(.firewall-rule-filter-popper .el-select-group__wrap) {
    padding: 6px 8px 8px;
}

:global(.firewall-rule-filter-popper .el-select-group__wrap:not(:last-of-type)) {
    padding-bottom: 10px;
}

:global(.firewall-rule-filter-popper .el-select-group__wrap:not(:last-of-type)::after) {
    display: none;
}

:global(.firewall-rule-filter-popper .el-select-group__title) {
    height: 30px;
    margin-bottom: 4px;
    padding: 0 10px;
    border-left: 3px solid var(--el-color-primary);
    border-radius: 4px;
    background: var(--el-fill-color-light);
    color: var(--el-text-color-primary);
    font-size: 13px;
    font-weight: 600;
    line-height: 30px;
}
</style>
