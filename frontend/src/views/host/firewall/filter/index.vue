<template>
    <div>
        <FireRouter />

        <div v-loading="loading">
            <FireStatus
                ref="fireStatusRef"
                @search="search"
                @advanced-operate="onApplyFirewall"
                v-model:loading="loading"
                v-model:mask-show="maskShow"
                v-model:is-active="isActive"
                v-model:name="fireName"
                :show-advanced-controls="true"
                :can-apply="canApply"
                :is-applied="isApplied"
            />
            <div v-if="fireName === '-'">
                <LayoutContent :divider="true">
                    <template #main>
                        <div class="app-warn">
                            <div class="flex flex-col gap-2 items-center justify-center w-full sm:flex-row">
                                <span>{{ $t('firewall.advancedControlNotAvailable', [firewallName]) }}</span>
                            </div>
                            <div>
                                <img src="@/assets/images/no_app.svg" />
                            </div>
                        </div>
                    </template>
                </LayoutContent>
            </div>

            <div v-else-if="!isIptablesComputed">
                <LayoutContent :divider="true">
                    <template #main>
                        <div class="app-warn">
                            <div class="flex flex-col gap-2 items-center justify-center w-full sm:flex-row">
                                <span>{{ $t('firewall.advancedControlNotAvailable', [fireName]) }}</span>
                            </div>
                            <div>
                                <img src="@/assets/images/no_app.svg" />
                            </div>
                        </div>
                    </template>
                </LayoutContent>
            </div>

            <div v-else>
                <el-card v-if="!isActive && maskShow" class="mask-prompt">
                    <span>{{ $t('firewall.firewallNotStart') }}</span>
                </el-card>
                <LayoutContent :title="$t('firewall.filterRule')" :class="{ mask: !isActive }">
                    <template #leftToolBar>
                        <el-button type="primary" @click="onOpenDialog('create')" :disabled="!isCustomChain">
                            {{ $t('firewall.create') }}
                        </el-button>
                        <el-button @click="onDelete(null)" plain :disabled="selects.length === 0 || !isCustomChain">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </template>

                    <template #rightToolBar>
                        <el-select v-model="selectedChain" @change="search()" clearable class="p-w-200">
                            <template #prefix>{{ $t('firewall.chain') }}</template>
                            <el-option :label="$t('firewall.inboundDirection')" value="1PANEL_INPUT"></el-option>
                            <el-option :label="$t('firewall.outboundDirection')" value="1PANEL_OUTPUT"></el-option>
                            <el-option label="INPUT" value="INPUT"></el-option>
                            <el-option label="OUTPUT" value="OUTPUT"></el-option>
                        </el-select>
                        <TableRefresh @search="search()" />
                        <TableSetting title="firewall-filter-refresh" @search="search()" />
                    </template>

                    <template #main>
                        <!-- Chain Status Cards -->
                        <el-row :gutter="20" class="mb-4">
                            <el-col :span="12">
                                <el-card>
                                    <div class="chain-card">
                                        <div class="chain-title">{{ $t('firewall.inboundDirection') }}</div>
                                        <div class="chain-policy">
                                            {{ $t('firewall.defaultPolicy') }}:
                                            {{ inputChainInfo?.defaultPolicy || 'ACCEPT' }}
                                        </div>
                                    </div>
                                </el-card>
                            </el-col>
                            <el-col :span="12">
                                <el-card>
                                    <div class="chain-card">
                                        <div class="chain-title">{{ $t('firewall.outboundDirection') }}</div>
                                        <div class="chain-policy">
                                            {{ $t('firewall.defaultPolicy') }}:
                                            {{ outputChainInfo?.defaultPolicy || 'ACCEPT' }}
                                        </div>
                                    </div>
                                </el-card>
                            </el-col>
                        </el-row>

                        <ComplexTable
                            :pagination-config="paginationConfig"
                            v-model:selects="selects"
                            @search="search"
                            :data="data"
                            :heightDiff="520"
                        >
                            <el-table-column type="selection" fix />
                            <el-table-column :label="$t('firewall.ruleOrder')" :min-width="80" prop="ruleOrder" />
                            <el-table-column :label="$t('commons.table.protocol')" :min-width="80" prop="protocol">
                                <template #default="{ row }">
                                    <span>{{ row.protocol === '' ? 'ALL' : row.protocol.toUpperCase() }}</span>
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('firewall.sourceIP')" :min-width="120" prop="sourceIP">
                                <template #default="{ row }">
                                    <span>{{ row.sourceIP || $t('firewall.anyWhere') }}</span>
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('firewall.sourcePort')" :min-width="100" prop="sourcePort">
                                <template #default="{ row }">
                                    <span>{{ formatPort(row.sourcePort) }}</span>
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('firewall.destIP')" :min-width="120" prop="destIP">
                                <template #default="{ row }">
                                    <span>{{ row.destIP || $t('firewall.anyWhere') }}</span>
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('firewall.destPort')" :min-width="100" prop="destPort">
                                <template #default="{ row }">
                                    <span>{{ formatPort(row.destPort) }}</span>
                                </template>
                            </el-table-column>
                            <el-table-column :min-width="100" :label="$t('firewall.action')" prop="action">
                                <template #default="{ row }">
                                    <el-tag v-if="row.action === 'ACCEPT'" type="success">{{ row.action }}</el-tag>
                                    <el-tag v-else-if="row.action === 'DROP'" type="danger">{{ row.action }}</el-tag>
                                    <el-tag v-else-if="row.action === 'REJECT'" type="warning">{{ row.action }}</el-tag>
                                    <el-tag v-else type="info">{{ row.action }}</el-tag>
                                </template>
                            </el-table-column>
                            <el-table-column
                                :min-width="150"
                                :label="$t('commons.table.description')"
                                prop="description"
                                show-overflow-tooltip
                            />
                            <fu-table-operations
                                v-if="isCustomChain"
                                width="120px"
                                :buttons="buttons"
                                :ellipsis="10"
                                :label="$t('commons.table.operate')"
                                fix
                            />
                        </ComplexTable>
                    </template>
                </LayoutContent>
            </div>
        </div>

        <OpDialog ref="opRef" @search="search" />
        <OperateDialog @search="search" ref="dialogRef" />
    </div>
</template>

<script lang="ts" setup>
import FireRouter from '@/views/host/firewall/index.vue';
import FireStatus from '@/views/host/firewall/status/index.vue';
import OperateDialog from '@/views/host/firewall/filter/operate/index.vue';
import { computed, onMounted, reactive, ref } from 'vue';
import { getFilterRules, applyFilterFirewall, batchOperateFilterRule } from '@/api/modules/host';
import { Host } from '@/api/interface/host';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox } from 'element-plus';

const loading = ref();
const selects = ref<any>([]);
const selectedChain = ref('1PANEL_INPUT');
const iptablesVersion = ref('');
const firewallName = ref('');

const maskShow = ref(true);
const isActive = ref(false);
const fireName = ref();
const fireStatusRef = ref();

const opRef = ref();

const chainInfoMap = ref<Map<string, Host.IptablesChainInfo>>(new Map());
const data = computed(() => {
    const chainInfo = chainInfoMap.value.get(selectedChain.value);
    return chainInfo?.rules || [];
});

const formatPort = (port?: number | null | string) => {
    if (port === 0 || port === '0') {
        return i18n.global.t('firewall.allPorts');
    }
    if (port === undefined || port === null || port === '') {
        return '-';
    }
    return port;
};

const isIptablesComputed = computed(() => fireName.value === 'iptables');
const inputChainInfo = computed(() => chainInfoMap.value.get('INPUT'));
const outputChainInfo = computed(() => chainInfoMap.value.get('OUTPUT'));

const isCustomChain = computed(() => {
    return selectedChain.value === '1PANEL_INPUT' || selectedChain.value === '1PANEL_OUTPUT';
});

const canApply = computed(() => {
    const inputInfo = chainInfoMap.value.get('1PANEL_INPUT');
    const outputInfo = chainInfoMap.value.get('1PANEL_OUTPUT');
    return (inputInfo && !inputInfo.isApplied) || (outputInfo && !outputInfo.isApplied);
});

const isApplied = computed(() => {
    const inputInfo = chainInfoMap.value.get('INPUT');
    const outputInfo = chainInfoMap.value.get('OUTPUT');
    return inputInfo?.isApplied || outputInfo?.isApplied;
});

const paginationConfig = reactive({
    cacheSizeKey: 'firewall-filter-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('firewall-filter-page-size')) || 20,
    total: 0,
});

const search = async () => {
    if (!isActive.value) {
        loading.value = false;
        chainInfoMap.value.clear();
        paginationConfig.total = 0;
        return;
    }
    const params: Host.IptablesFilterRuleSearch = {
        chains: ['INPUT', 'OUTPUT', '1PANEL_INPUT', '1PANEL_OUTPUT'],
    };
    loading.value = true;
    await getFilterRules(params)
        .then((res) => {
            loading.value = false;
            chainInfoMap.value.clear();
            res.data.forEach((chainInfo: Host.IptablesChainInfo) => {
                chainInfoMap.value.set(chainInfo.name, chainInfo);
            });
            const currentChainData = data.value || [];
            iptablesVersion.value = chainInfoMap.value.get('INPUT')?.version || '';
            paginationConfig.total = currentChainData.length;
        })
        .catch(() => {
            loading.value = false;
        });
};

const dialogRef = ref();
const onOpenDialog = async (title: string, rowData?: Host.IptablesFilterRuleOperate) => {
    if (!isCustomChain.value) {
        return;
    }
    const params = {
        title,
        rowData: rowData || {
            chain: selectedChain.value,
            protocol: 'all',
            action: 'ACCEPT',
            sourcePort: 0,
            destPort: 0,
        },
    };
    dialogRef.value!.acceptParams(params);
};

const onApplyFirewall = async (operation: string) => {
    let confirmMsg = '';
    let title = '';

    if (operation === 'init') {
        confirmMsg = i18n.global.t('firewall.initChainsConfirm');
        title = i18n.global.t('firewall.initChains');
    } else if (operation === 'apply') {
        confirmMsg = i18n.global.t('firewall.applyConfirm');
        title = i18n.global.t('firewall.applyFirewall');
    } else {
        confirmMsg = i18n.global.t('firewall.unloadConfirm');
        title = i18n.global.t('firewall.unloadFirewall');
    }

    ElMessageBox.confirm(confirmMsg, title, {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        loading.value = true;
        await applyFilterFirewall({ operation })
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

const onDelete = async (row: Host.IptablesFilterRuleInfo | null) => {
    let names = [];
    let rules = [];
    if (row) {
        rules.push({
            operation: 'remove',
            id: row.id,
            chain: selectedChain.value,
            protocol: row.protocol,
            action: row.action,
        });
        names = [
            `${row.protocol} ${row.sourceIP || '*'}:${row.sourcePort || '*'} -> ${row.destIP || '*'}:${
                row.destPort || '*'
            }`,
        ];
    } else {
        for (const item of selects.value) {
            names.push(
                `${item.protocol} ${item.sourceIP || '*'}:${item.sourcePort || '*'} -> ${item.destIP || '*'}:${
                    item.destPort || '*'
                }`,
            );
            rules.push({
                operation: 'remove',
                id: item.id,
                chain: selectedChain.value,
                protocol: item.protocol,
                action: item.action,
            });
        }
    }
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names,
        msg: i18n.global.t('firewall.deleteRuleConfirm', [rules.length]),
        api: batchOperateFilterRule,
        params: { rules: rules },
    });
};

const buttons = [
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Host.IptablesFilterRuleInfo) => {
            onDelete(row);
        },
    },
];

onMounted(() => {
    if (fireName.value !== '-') {
        loading.value = true;
        fireStatusRef.value.acceptParams();
    }
});
</script>

<style lang="scss" scoped>
.chain-card {
    .chain-title {
        font-size: 16px;
        font-weight: 500;
        margin-bottom: 8px;
        display: flex;
        align-items: center;
        gap: 8px;
    }
    .chain-policy {
        font-size: 14px;
        color: var(--el-text-color-secondary);
    }
}
</style>
