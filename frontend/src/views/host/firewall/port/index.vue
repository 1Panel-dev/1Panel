<template>
    <div>
        <FireRouter />

        <div v-loading="loading">
            <FireStatus
                ref="fireStatusRef"
                @search="search"
                v-model:loading="loading"
                v-model:mask-show="maskShow"
                v-model:is-active="isActive"
                v-model:is-bind="isBind"
                v-model:name="fireName"
                current-tab="base"
            />
            <div v-if="fireName !== '-'">
                <el-card v-if="!isActive && maskShow" class="mask-prompt">
                    <span>{{ $t('firewall.firewallNotStart') }}</span>
                </el-card>
                <el-card v-if="!isBind && maskShow" class="mask-prompt">
                    <span>{{ $t('firewall.basicStatus', ['1PANEL_BASIC']) }}</span>
                </el-card>

                <LayoutContent :title="$t('firewall.portRule', 2)" :class="{ mask: !isActive || !isBind }">
                    <template #prompt>
                        <div class="mb-2" v-if="fireName !== 'iptables'">
                            <el-alert :closable="false" :title="$t('firewall.iptablesHelper', [fireName])" />
                        </div>
                        <el-alert type="info" :closable="false">
                            <template #default>
                                <span class="flx-align-center">
                                    <span>{{ $t('firewall.dockerHelper', [fireName]) }}</span>
                                    <el-link
                                        style="font-size: 12px; margin-left: 5px"
                                        icon="Position"
                                        @click="quickJump()"
                                        type="primary"
                                    >
                                        {{ $t('firewall.quickJump') }}
                                    </el-link>
                                </span>
                            </template>
                        </el-alert>
                    </template>
                    <template #leftToolBar>
                        <el-button type="primary" @click="onOpenDialog('create')">
                            {{ $t('commons.button.create') }}{{ $t('firewall.portRule') }}
                        </el-button>
                        <el-button @click="onDelete(null)" plain :disabled="selects.length === 0">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                        <el-button-group>
                            <el-button @click="onImport">
                                {{ $t('commons.button.import') }}
                            </el-button>
                            <el-button :disabled="selects.length === 0" @click="onExport">
                                {{ $t('commons.button.export') }}
                            </el-button>
                        </el-button-group>
                    </template>
                    <template #rightToolBar>
                        <el-select v-model="searchStrategy" @change="search()" clearable class="p-w-200">
                            <template #prefix>{{ $t('firewall.strategy') }}</template>
                            <el-option :label="$t('commons.table.all')" value=""></el-option>
                            <el-option :label="$t('firewall.accept')" value="accept"></el-option>
                            <el-option :label="$t('firewall.drop')" value="drop"></el-option>
                        </el-select>
                        <TableSearch @search="search()" v-model:searchName="searchName" />
                        <TableRefresh @search="search()" />
                        <TableSetting title="firewall-port-refresh" @search="search()" />
                    </template>
                    <template #main>
                        <ComplexTable
                            :pagination-config="paginationConfig"
                            v-model:selects="selects"
                            @search="search"
                            :data="data"
                            :heightDiff="400"
                        >
                            <el-table-column type="selection" fix />
                            <el-table-column :label="$t('commons.table.protocol')" :min-width="70" prop="protocol" />
                            <el-table-column :label="$t('commons.table.port')" :min-width="70" prop="port" />
                            <el-table-column :label="$t('commons.table.status')" :min-width="120">
                                <template #default="{ row }">
                                    <div v-if="isSinglePort(row.port)">
                                        <el-tag type="success" v-if="row.usedStatus">
                                            {{ $t('firewall.used') + ' (' + row.usedStatus + ')' }}
                                            <el-icon
                                                v-if="row.processInfo"
                                                @click="showProcessDetail(row.processInfo.PID)"
                                                style="margin-left: 4px; cursor: pointer; vertical-align: middle"
                                            >
                                                <Expand />
                                            </el-icon>
                                        </el-tag>
                                        <el-tag type="info" v-else>{{ $t('firewall.unUsed') }}</el-tag>
                                    </div>
                                    <span v-else>-</span>
                                </template>
                            </el-table-column>
                            <el-table-column :min-width="80" :label="$t('firewall.strategy')" prop="strategy">
                                <template #default="{ row }">
                                    <el-button
                                        v-if="row.strategy === 'accept'"
                                        @click="onChangeStatus(row, 'drop')"
                                        link
                                        type="success"
                                    >
                                        {{ $t('firewall.accept') }}
                                    </el-button>
                                    <el-button v-else link type="danger" @click="onChangeStatus(row, 'accept')">
                                        {{ $t('firewall.drop') }}
                                    </el-button>
                                </template>
                            </el-table-column>
                            <el-table-column :min-width="80" :label="$t('firewall.address')" prop="address">
                                <template #default="{ row }">
                                    <span v-if="row.address && row.address !== 'Anywhere'">{{ row.address }}</span>
                                    <span v-else>{{ $t('firewall.allIP') }}</span>
                                </template>
                            </el-table-column>
                            <el-table-column
                                :min-width="150"
                                :label="$t('commons.table.description')"
                                prop="description"
                                show-overflow-tooltip
                            >
                                <template #default="{ row }">
                                    <fu-input-rw-switch
                                        v-model="row.description"
                                        @enter="onChange(row)"
                                        @blur="onChange(row)"
                                    />
                                </template>
                            </el-table-column>
                            <fu-table-operations
                                width="200px"
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
        <ImportDialog @search="search" ref="dialogImportRef" />
        <ProcessDetail ref="processDetailRef" />
    </div>
</template>

<script lang="ts" setup>
import FireRouter from '@/views/host/firewall/index.vue';
import OperateDialog from '@/views/host/firewall/port/operate/index.vue';
import ImportDialog from '@/views/host/firewall/port/import/index.vue';
import FireStatus from '@/views/host/firewall/status/index.vue';
import ProcessDetail from '@/views/host/process/process/detail/index.vue';
import { onMounted, reactive, ref } from 'vue';
import { batchOperateRule, searchFireRule, updateFirewallDescription, updatePortRule } from '@/api/modules/host';
import { getListeningProcess } from '@/api/modules/process';
import { Host } from '@/api/interface/host';
import { Process } from '@/api/interface/process';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { ElMessageBox } from 'element-plus';
import { Expand } from '@element-plus/icons-vue';
import { routerToName } from '@/utils/router';
import { downloadWithContent, getCurrentDateFormatted } from '@/utils/util';

const loading = ref();
const activeTag = ref('port');
const selects = ref<any>([]);
const searchName = ref();
const searchStrategy = ref('');

const maskShow = ref(true);
const isActive = ref(false);
const isBind = ref(false);
const fireName = ref();
const fireStatusRef = ref();

const opRef = ref();
const dialogImportRef = ref();
const processDetailRef = ref();

const listeningProcesses = ref<Process.ListeningProcess[]>([]);

const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'firewall-port-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('firewall-port-page-size')) || 20,
    total: 0,
});

const extractPortsFromObject = (portObj: { [key: string]: {} }): number[] => {
    return Object.keys(portObj)
        .map((portStr) => parseInt(portStr))
        .filter((port) => !isNaN(port));
};

const isSinglePort = (portStr: string): boolean => {
    return portStr.indexOf('-') === -1 && portStr.indexOf(':') === -1 && portStr.indexOf(',') === -1;
};

const loadListeningProcesses = async () => {
    try {
        const res = await getListeningProcess();
        listeningProcesses.value = res.data || [];

        for (const item of data.value) {
            if (!item.usedStatus && isSinglePort(item.port)) {
                const portNum = parseInt(item.port.trim());
                if (!isNaN(portNum)) {
                    const protocolNum =
                        item.protocol.toLowerCase() === 'tcp' ? 1 : item.protocol.toLowerCase() === 'udp' ? 2 : 0;

                    for (const proc of listeningProcesses.value) {
                        if (proc.Protocol === protocolNum) {
                            const procPorts = extractPortsFromObject(proc.Port);
                            if (procPorts.includes(portNum)) {
                                item.usedStatus = proc.Name;
                                item.processInfo = proc;
                                break;
                            }
                        }
                    }
                }
            }
        }
    } catch (error) {
        console.error('Failed to load listening processes:', error);
    }
};

const search = async () => {
    if (!isActive.value) {
        loading.value = false;
        data.value = [];
        paginationConfig.total = 0;
        return;
    }
    let params = {
        type: activeTag.value,
        strategy: searchStrategy.value,
        info: searchName.value,
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    loading.value = true;
    await searchFireRule(params)
        .then(async (res) => {
            loading.value = false;
            data.value = res.data.items || [];

            await loadListeningProcesses();

            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const dialogRef = ref();
const onOpenDialog = async (
    title: string,
    rowData: Partial<Host.RulePort> = {
        protocol: 'tcp',
        source: 'anyWhere',
        strategy: 'accept',
    },
) => {
    let params = {
        title,
        fireName: fireName.value,
        rowData: { ...rowData },
    };
    dialogRef.value!.acceptParams(params);
};

const quickJump = () => {
    routerToName('AppInstalled');
};

const onChangeStatus = async (row: Host.RuleInfo, status: string) => {
    let operation =
        status === 'accept'
            ? i18n.global.t('firewall.changeStrategyPortHelper2')
            : i18n.global.t('firewall.changeStrategyPortHelper1');
    ElMessageBox.confirm(operation, i18n.global.t('firewall.changeStrategy', [i18n.global.t('commons.table.port')]), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        let params = {
            oldRule: {
                operation: 'remove',
                address: row.address,
                port: row.port,
                source: '',
                protocol: row.protocol,
                strategy: row.strategy,
                description: row.description,
            },
            newRule: {
                operation: 'add',
                address: row.address,
                port: row.port,
                source: '',
                protocol: row.protocol,
                strategy: status,
                description: row.description,
            },
        };
        loading.value = true;
        await updatePortRule(params)
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

const onChange = async (row: any) => {
    let params = {
        type: 'port',
        chain: fireName.value === 'iptables' ? '1PANEL_BASIC' : '',
        srcIP: row.address,
        dstIP: '',
        srcPort: '',
        dstPort: row.port,
        protocol: row.protocol,
        strategy: row.strategy,

        description: row.description,
    };
    await updateFirewallDescription(params);
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
};

const onDelete = async (row: Host.RuleInfo | null) => {
    let names = [];
    let rules = [];
    if (row) {
        rules.push({
            operation: 'remove',
            chain: row.chain,
            address: row.address,
            port: row.port,
            source: '',
            protocol: row.protocol,
            strategy: row.strategy,
        });
        names = [row.port + ' (' + row.protocol + ')'];
    } else {
        for (const item of selects.value) {
            names.push(item.port + ' (' + item.protocol + ')');
            rules.push({
                operation: 'remove',
                chain: item.chain,
                address: item.address,
                port: item.port,
                source: '',
                protocol: item.protocol,
                strategy: item.strategy,
            });
        }
    }
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('firewall.portRule'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: batchOperateRule,
        params: { type: 'port', rules: rules },
    });
};

const onImport = () => {
    dialogImportRef.value.acceptParams();
};

const onExport = () => {
    ElMessageBox.confirm(
        i18n.global.t('firewall.exportHelper', [selects.value.length]),
        i18n.global.t('commons.button.export'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        },
    ).then(async () => {
        const exportData = selects.value.map((item: Host.RuleInfo) => ({
            family: item.family,
            address: item.address,
            port: item.port,
            protocol: item.protocol,
            strategy: item.strategy,
            description: item.description,
        }));
        const content = JSON.stringify(exportData, null, 2);
        const fileName = `1panel-firewall-port-${getCurrentDateFormatted()}.json`;
        downloadWithContent(content, fileName);
    });
};

const showProcessDetail = (pid: number) => {
    processDetailRef.value?.acceptParams(pid);
};

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: Host.RulePort) => {
            onOpenDialog('edit', row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Host.RuleInfo) => {
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
.svg-icon {
    font-size: 8px;
    margin-bottom: -4px;
    cursor: pointer;
}
</style>
