<template>
    <LayoutContent :title="$t('firewall.portSecurity')">
        <template #leftToolBar>
            <span style="font-weight: 500; font-size: 14px;">{{ $t('firewall.portSecurity') }}</span>
            <template v-if="!loading && overview">
                <el-tag
                    v-if="overview.summary.dockerBypassed > 0"
                    type="danger"
                    class="summary-chip"
                    @click="onChipFilter('dockerBypass')"
                >
                    {{ overview.summary.dockerBypassed }} {{ $t('firewall.portSecurityRiskLabel') }}
                </el-tag>
                <el-tag
                    v-if="overview.summary.unprotected > 0"
                    type="warning"
                    class="summary-chip"
                    @click="onChipFilter(overview.fireActive ? 'noRule' : 'firewallInactive')"
                >
                    {{ overview.summary.unprotected }} {{ $t('firewall.portSecurityPendingLabel') }}
                </el-tag>
                <el-tag type="success">
                    {{ overview.summary.protected + overview.summary.localOnly }}
                    {{ $t('firewall.portSecuritySafeLabel') }}
                </el-tag>
            </template>
        </template>
        <template #rightToolBar>
            <el-select v-model="searchStatus" @change="search" clearable class="p-w-200">
                <template #prefix>{{ $t('commons.table.status') }}</template>
                <el-option :label="$t('commons.table.all')" value=""></el-option>
                <el-option :label="$t('firewall.dockerBypass')" value="dockerBypass"></el-option>
                <el-option :label="$t('firewall.firewallInactive')" value="firewallInactive"></el-option>
                <el-option :label="$t('firewall.noRule')" value="noRule"></el-option>
                <el-option :label="$t('firewall.protected')" value="protected"></el-option>
                <el-option :label="$t('firewall.blocked')" value="blocked"></el-option>
                <el-option :label="$t('firewall.localOnly')" value="localOnly"></el-option>
            </el-select>
            <TableSearch @search="search()" v-model:searchName="searchName" />
            <TableRefresh @search="search()" />
            <TableSetting title="firewall-port-security-refresh" @search="search()" />
        </template>
        <template #main>
            <el-alert
                v-if="hasError && !loading"
                :title="$t('firewall.portSecurityScanFailed')"
                type="warning"
                show-icon
                :closable="false"
                style="margin-bottom: 12px;"
            />
            <div
                v-else-if="allData.length === 0 && !loading && !searchName && !searchStatus"
                class="scan-clear"
            >
                <el-icon color="var(--el-color-success)"><CircleCheckFilled /></el-icon>
                <span>{{ $t('firewall.portSecurityAllSafe') }}</span>
                <span class="scan-clear__note">{{ $t('firewall.portSecurityScanLimit') }}</span>
            </div>
            <ComplexTable
                v-else
                :pagination-config="paginationConfig"
                :data="tableData"
                @search="() => {}"
            >
                <el-table-column :label="$t('commons.table.port')" :min-width="70" prop="port" />
                <el-table-column :label="$t('commons.table.protocol')" :min-width="70" prop="protocol" />
                <el-table-column :label="$t('firewall.portSecuritySource')" :min-width="120">
                    <template #default="{ row }">
                        <div class="cell-inline">
                            <span>{{ row.containerName || row.processName || '-' }}</span>
                            <el-tag
                                v-if="row.sourceType === 'docker' || row.sourceType === 'appStore'"
                                class="source-tag--container"
                                effect="plain"
                            >
                                <svg-icon iconName="p-docker" style="width: 12px; height: 12px; margin-right: 2px; vertical-align: middle;" />
                                {{ $t('firewall.dockerLabel') }}
                            </el-tag>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('firewall.portSecurityBindAddr')" :min-width="80" prop="bindAddress" show-overflow-tooltip />
                <el-table-column :label="$t('commons.table.status')" :min-width="120">
                    <template #default="{ row }">
                        <div class="cell-inline">
                            <el-tooltip
                                v-if="row.hasRule && hasSourceScope(row.ruleAddress)"
                                :content="`${row.ruleStrategy} ${row.port}/${row.protocol} from ${row.ruleAddress}`"
                                placement="top"
                            >
                                <el-tag :type="statusTagType(row.status)">
                                    {{ $t('firewall.' + row.status) }}
                                </el-tag>
                            </el-tooltip>
                            <el-tag v-else :type="statusTagType(row.status)">
                                {{ $t('firewall.' + row.status) }}
                            </el-tag>
                            <el-tooltip
                                v-if="row.status === 'dockerBypass' && row.hasRule"
                                :content="$t('firewall.dockerBypassHasRule', [row.ruleStrategy, row.port, row.protocol])"
                                placement="top"
                            >
                                <el-icon color="var(--el-color-warning)" style="cursor: pointer;">
                                    <WarningFilled />
                                </el-icon>
                            </el-tooltip>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('commons.table.operate')" width="200px" fixed="right" align="center" header-align="center">
                    <template #default="{ row }">
                        <el-button
                            v-if="row.status === 'noRule'"
                            link
                            type="primary"
                            @click="onCreateRule(row)"
                        >
                            {{ $t('firewall.createRuleQuick') }}
                        </el-button>
                        <el-popover
                            v-else-if="row.status === 'dockerBypass'"
                            placement="left"
                            :width="320"
                            trigger="click"
                        >
                            <template #reference>
                                <el-button link type="primary">
                                    {{ $t('firewall.dockerBypassDetail') }}
                                </el-button>
                            </template>
                            <div class="docker-tip">
                                <p>{{ $t('firewall.dockerBypassTip') }}</p>
                                <p class="suggest">{{ $t('firewall.dockerBypassSuggest', [row.port]) }}</p>
                                <el-button
                                    v-if="row.sourceType === 'appStore'"
                                    type="primary"
                                    size="small"
                                    @click="goToApp"
                                    style="margin-top: 8px;"
                                >
                                    {{ $t('firewall.goToApp') }}
                                </el-button>
                            </div>
                        </el-popover>
                    </template>
                </el-table-column>
            </ComplexTable>
        </template>
    </LayoutContent>
</template>

<script setup lang="ts">
import { ref, computed, reactive, onMounted } from 'vue';
import { CircleCheckFilled, WarningFilled } from '@element-plus/icons-vue';
import { getPortSecurity } from '@/api/modules/host';
import { Host } from '@/api/interface/host';
import { routerToName } from '@/utils/router';

const emit = defineEmits(['createRule']);

const loading = ref(false);
const hasError = ref(false);
const searchName = ref('');
const searchStatus = ref('');
const allData = ref<Host.PortSecurityItem[]>([]);
const overview = ref<Host.PortSecurityOverview>();

const paginationConfig = reactive({
    cacheSizeKey: 'firewall-port-security-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('firewall-port-security-page-size')) || 5,
    total: 0,
});

const tableData = computed(() => {
    const start = (paginationConfig.currentPage - 1) * paginationConfig.pageSize;
    return allData.value.slice(start, start + paginationConfig.pageSize);
});

const statusTagType = (status: string) => {
    switch (status) {
        case 'dockerBypass':
        case 'firewallInactive':
            return 'danger';
        case 'noRule':
            return 'warning';
        case 'protected':
            return 'success';
        case 'blocked':
        case 'localOnly':
            return 'info';
        default:
            return 'info';
    }
};

const hasSourceScope = (addr: string) => {
    return !!addr && addr !== 'Anywhere' && addr !== '0.0.0.0/0' && addr !== '::/0';
};

const onChipFilter = (status: string) => {
    if (searchStatus.value === status) {
        // toggle off when clicking the active chip
        searchStatus.value = '';
    } else {
        searchStatus.value = status;
    }
    search();
};

const search = async () => {
    if (loading.value) return;
    loading.value = true;
    hasError.value = false;
    await getPortSecurity({ info: searchName.value, status: searchStatus.value })
        .then((res) => {
            overview.value = res.data;
            allData.value = res.data.items || [];
            paginationConfig.total = allData.value.length;
            paginationConfig.currentPage = 1;
        })
        .catch(() => {
            hasError.value = true;
            overview.value = undefined;
            allData.value = [];
            paginationConfig.total = 0;
        })
        .finally(() => {
            loading.value = false;
        });
};

const onCreateRule = (row: Host.PortSecurityItem) => {
    emit('createRule', { port: String(row.port), protocol: row.protocol });
};

const goToApp = () => {
    routerToName('AppInstalled');
};

onMounted(() => {
    search();
});
</script>

<style scoped lang="scss">
.scan-clear {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 6px;
    padding: 16px 0;
    color: var(--el-text-color-regular);
    font-size: 14px;
}

.scan-clear__note {
    color: var(--el-text-color-secondary);
    font-size: 12px;
    margin-left: 8px;
}

.summary-chip {
    cursor: pointer;
    transition: filter 0.15s ease-in-out, transform 0.05s ease-in-out;

    &:hover {
        filter: brightness(0.95);
    }

    &:active {
        transform: translateY(1px);
    }
}

.cell-inline {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    vertical-align: middle;
}

.source-tag--container {
    --el-tag-bg-color: var(--el-color-primary-light-9);
    --el-tag-border-color: var(--el-color-primary-light-7);
    --el-tag-text-color: var(--el-color-primary);
}

.docker-tip {
    font-size: 13px;
    line-height: 1.6;
    p {
        margin: 0 0 6px;
    }
    .suggest {
        color: var(--el-color-warning);
        font-size: 12px;
    }
}
</style>
