<template>
    <LayoutContent :title="$t('firewall.portSecurity')">
        <template #leftToolBar>
            <span style="font-weight: 500; font-size: 14px;">{{ $t('firewall.portSecurity') }}</span>
            <template v-if="!loading && overview">
                <el-tag v-if="overview.summary.dockerBypassed > 0" type="danger">
                    {{ overview.summary.dockerBypassed }} {{ $t('firewall.portSecurityRiskLabel') }}
                </el-tag>
                <el-tag v-if="overview.summary.unprotected > 0" type="warning">
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
                <el-option :label="$t('firewall.noRule')" value="noRule"></el-option>
                <el-option :label="$t('firewall.protected')" value="protected"></el-option>
                <el-option :label="$t('firewall.localOnly')" value="localOnly"></el-option>
            </el-select>
            <TableSearch @search="search()" v-model:searchName="searchName" />
            <TableRefresh @search="search()" />
            <TableSetting title="firewall-port-security-refresh" @search="search()" />
        </template>
        <template #main>
            <div v-if="allData.length === 0 && !loading && !hasError && !searchName && !searchStatus" class="all-safe">
                <el-icon color="var(--el-color-success)"><CircleCheckFilled /></el-icon>
                <span>{{ $t('firewall.portSecurityAllSafe') }}</span>
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
                        <span>{{ row.containerName || row.processName || '-' }}</span>
                        <el-tag
                            v-if="row.sourceType === 'docker' || row.sourceType === 'appStore'"
                            class="source-tag--container"
                            effect="plain"
                            style="margin-left: 4px;"
                        >
                            <svg-icon iconName="p-docker" style="width: 12px; height: 12px; margin-right: 2px; vertical-align: middle;" />
                            {{ $t('firewall.dockerLabel') }}
                        </el-tag>
                        <el-tag v-if="row.appName" type="info" effect="plain" style="margin-left: 4px;">
                            {{ row.appName }}
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('firewall.portSecurityBindAddr')" :min-width="80" prop="bindAddress" show-overflow-tooltip />
                <el-table-column :label="$t('commons.table.status')" :min-width="120">
                    <template #default="{ row }">
                        <el-tag :type="statusTagType(row.status)">
                            {{ $t('firewall.' + row.status) }}
                        </el-tag>
                        <el-tooltip
                            v-if="row.status === 'dockerBypass' && row.hasRule"
                            :content="$t('firewall.dockerBypassHasRule', [row.ruleStrategy, row.port, row.protocol])"
                            placement="top"
                        >
                            <el-icon color="var(--el-color-warning)" style="margin-left: 4px; cursor: pointer;">
                                <WarningFilled />
                            </el-icon>
                        </el-tooltip>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('commons.table.operate')" width="200px" fixed="right">
                    <template #default="{ row }">
                        <el-button
                            v-if="row.status === 'noRule' || row.status === 'firewallInactive'"
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
        case 'localOnly':
            return 'info';
        default:
            return 'info';
    }
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
.all-safe {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 16px 0;
    color: var(--el-color-success);
    font-size: 14px;
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
