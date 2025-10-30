<template>
    <div>
        <div v-loading="loading">
            <el-card>
                <div class="flex w-full flex-col gap-4 md:flex-row">
                    <div class="flex flex-wrap gap-4 ml-3">
                        <el-tag effect="dark" type="success">{{ 'iptables v' + iptablesVersion }}</el-tag>
                    </div>
                </div>
            </el-card>
            <div>
                <LayoutContent :title="$t('firewall.filterRule')">
                    <template #prompt>
                        <el-alert type="info" :closable="false">
                            <span>{{ $t('firewall.filterHelper') }}</span>
                        </el-alert>
                    </template>

                    <template #leftToolBar>
                        <el-button type="primary" @click="onOpenDialog('create')" :disabled="!isCustomChain">
                            {{ $t('commons.button.create') }}{{ $t('firewall.filterRule') }}
                        </el-button>
                        <el-button @click="onDelete(null)" plain :disabled="selects.length === 0 || !isCustomChain">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                        <el-button @click="onInitChains" plain>
                            {{ $t('firewall.initChains') }}
                        </el-button>
                        <el-button @click="onApplyFirewall('apply')" plain type="success" :disabled="!canApply">
                            {{ $t('firewall.applyFirewall') }}
                        </el-button>
                        <el-button @click="onApplyFirewall('unload')" plain type="warning" :disabled="!isApplied">
                            {{ $t('firewall.unloadFirewall') }}
                        </el-button>
                    </template>

                    <template #rightToolBar>
                        <el-select v-model="selectedChain" @change="search()" clearable class="p-w-200">
                            <template #prefix>{{ $t('firewall.chain') }}</template>
                            <el-option label="INPUT" value="INPUT"></el-option>
                            <el-option label="OUTPUT" value="OUTPUT"></el-option>
                            <el-option label="1PANEL_INPUT" value="1PANEL_INPUT"></el-option>
                            <el-option label="1PANEL_OUTPUT" value="1PANEL_OUTPUT"></el-option>
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
                                        <div class="chain-title">
                                            INPUT {{ $t('firewall.chain') }}
                                            <el-tag :type="inputChainInfo?.isApplied ? 'success' : 'info'" size="small">
                                                {{
                                                    inputChainInfo?.isApplied
                                                        ? $t('firewall.applied')
                                                        : $t('firewall.notApplied')
                                                }}
                                            </el-tag>
                                        </div>
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
                                        <div class="chain-title">
                                            OUTPUT {{ $t('firewall.chain') }}
                                            <el-tag
                                                :type="outputChainInfo?.isApplied ? 'success' : 'info'"
                                                size="small"
                                            >
                                                {{
                                                    outputChainInfo?.isApplied
                                                        ? $t('firewall.applied')
                                                        : $t('firewall.notApplied')
                                                }}
                                            </el-tag>
                                        </div>
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
                            <el-table-column v-if="isCustomChain" type="selection" fix />
                            <el-table-column :label="$t('firewall.ruleOrder')" :min-width="80" prop="ruleOrder" />
                            <el-table-column :label="$t('commons.table.protocol')" :min-width="80" prop="protocol" />
                            <el-table-column :label="$t('firewall.sourceIP')" :min-width="120" prop="sourceIP">
                                <template #default="{ row }">
                                    <span>{{ row.sourceIP || '-' }}</span>
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('firewall.sourcePort')" :min-width="100" prop="sourcePort">
                                <template #default="{ row }">
                                    <span>{{ row.sourcePort || '-' }}</span>
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('firewall.destIP')" :min-width="120" prop="destIP">
                                <template #default="{ row }">
                                    <span>{{ row.destIP || '-' }}</span>
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('firewall.destPort')" :min-width="100" prop="destPort">
                                <template #default="{ row }">
                                    <span>{{ row.destPort || '-' }}</span>
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
import OperateDialog from '@/views/host/filter/operate/index.vue';
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

const opRef = ref();

const chainInfoMap = ref<Map<string, Host.IptablesChainInfo>>(new Map());
const data = computed(() => {
    const chainInfo = chainInfoMap.value.get(selectedChain.value);
    return chainInfo?.rules || [];
});

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

const onInitChains = async () => {
    ElMessageBox.confirm(i18n.global.t('firewall.initChainsConfirm'), i18n.global.t('firewall.initChains'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        loading.value = true;
        await applyFilterFirewall({ operation: 'init' })
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

const onApplyFirewall = async (operation: string) => {
    const confirmMsg =
        operation === 'apply' ? i18n.global.t('firewall.applyConfirm') : i18n.global.t('firewall.unloadConfirm');
    const title =
        operation === 'apply' ? i18n.global.t('firewall.applyFirewall') : i18n.global.t('firewall.unloadFirewall');

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
    search();
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
