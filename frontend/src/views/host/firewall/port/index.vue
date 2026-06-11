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
                        <el-button v-permission v-node-admin type="primary" @click="onOpenDialog('create')">
                            {{ $t('commons.button.create') }}
                        </el-button>
                        <el-button
                            v-permission
                            v-node-admin
                            @click="onDelete(null)"
                            plain
                            :disabled="selects.length === 0"
                        >
                            {{ $t('commons.button.delete') }}
                        </el-button>
                        <el-button-group>
                            <el-button v-permission v-node-admin @click="onImport">
                                {{ $t('commons.button.import') }}
                            </el-button>
                            <el-button v-permission v-node-admin :disabled="selects.length === 0" @click="onExport">
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
                            <el-table-column :label="$t('commons.table.status')" :min-width="180">
                                <template #default="{ row }">
                                    <template v-if="row.usedStatus">
                                        <template v-if="row.processInfos?.length">
                                            <span class="process-list">
                                                <span
                                                    v-for="(process, index) in row.processInfos"
                                                    v-show="row.expand || index < 3"
                                                    :key="`${process.PID || formatProcessInfo(process)}-${index}`"
                                                >
                                                    <el-button
                                                        v-if="process.PID"
                                                        size="small"
                                                        class="process-link"
                                                        :title="`${formatProcessInfo(process)} (PID: ${process.PID})`"
                                                        @click.stop="showProcessDetail(process.PID)"
                                                    >
                                                        {{ formatProcessInfo(process) }}
                                                        <el-icon class="process-detail-icon">
                                                            <Expand />
                                                        </el-icon>
                                                    </el-button>
                                                    <span v-else class="process-name">
                                                        <el-button size="small">
                                                            {{ formatProcessInfo(process) }}
                                                        </el-button>
                                                    </span>
                                                </span>
                                                <el-button
                                                    v-if="!row.expand && row.processInfos.length > 3"
                                                    type="primary"
                                                    link
                                                    class="process-toggle"
                                                    @click.stop="row.expand = true"
                                                >
                                                    {{ $t('commons.button.expand') }}...
                                                </el-button>
                                                <el-button
                                                    v-if="row.expand && row.processInfos.length > 3"
                                                    type="primary"
                                                    link
                                                    class="process-toggle"
                                                    @click.stop="row.expand = false"
                                                >
                                                    {{ $t('commons.button.collapse') }}
                                                </el-button>
                                            </span>
                                        </template>
                                        <span v-else class="process-list">
                                            <span class="process-name">{{ row.usedStatus }}</span>
                                        </span>
                                    </template>
                                    <el-tag type="info" v-else>{{ $t('firewall.unUsed') }}</el-tag>
                                </template>
                            </el-table-column>
                            <el-table-column :min-width="80" :label="$t('firewall.strategy')" prop="strategy">
                                <template #default="{ row }">
                                    <el-button
                                        v-if="row.strategy === 'accept'"
                                        v-permission
                                        v-node-admin
                                        @click="onChangeStatus(row, 'drop')"
                                        link
                                        type="success"
                                    >
                                        {{ $t('firewall.accept') }}
                                    </el-button>
                                    <el-button
                                        v-else
                                        link
                                        type="danger"
                                        v-permission
                                        v-node-admin
                                        @click="onChangeStatus(row, 'accept')"
                                    >
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
                                        v-permission
                                        v-node-admin
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
import { downloadWithContent } from '@/utils/file';
import { getCurrentDateFormatted } from '@/utils/date';
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

type RuleInfoWithProcess = Host.RuleInfo & {
    expand?: boolean;
    processInfo?: Process.ListeningProcess;
    processInfos?: ProcessInfoDisplay[];
};

type ProcessInfoDisplay = Partial<Process.ListeningProcess> & {
    Name: string;
    ports: number[];
};

const data = ref<RuleInfoWithProcess[]>([]);
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

const isPortInRule = (rulePort: string, port: number): boolean => {
    const segments = rulePort.split(',');
    for (const segment of segments) {
        const portSegment = segment.trim();
        if (!portSegment) {
            continue;
        }

        const rangeDelimiter = portSegment.includes('-') && !portSegment.startsWith('-') ? '-' : ':';
        if (portSegment.includes(rangeDelimiter) && !portSegment.startsWith(rangeDelimiter)) {
            const [startPort, endPort] = portSegment.split(rangeDelimiter).map((item) => parseInt(item.trim()));
            if (!isNaN(startPort) && !isNaN(endPort) && port >= startPort && port <= endPort) {
                return true;
            }
            continue;
        }

        if (parseInt(portSegment) === port) {
            return true;
        }
    }
    return false;
};

const formatProcessInfo = (process: ProcessInfoDisplay): string => {
    const ports = process.ports.join(', ');
    if (!process.Name) {
        return ports;
    }
    if (!ports) {
        return process.Name;
    }
    return `${process.Name} (${ports})`;
};

const parseUsedStatus = (usedStatus: string, rulePort: string): ProcessInfoDisplay[] => {
    if (!usedStatus) {
        return [];
    }

    return usedStatus
        .split(',')
        .map((item) => item.trim())
        .filter((item) => item)
        .map((item) => {
            const appMatch = item.match(/^(\d+)\s+\((.+)\)$/);
            if (appMatch) {
                return {
                    Name: appMatch[2],
                    ports: [Number(appMatch[1])],
                };
            }

            const port = Number(item);
            if (!isNaN(port)) {
                return {
                    Name: '',
                    ports: [port],
                };
            }

            const rulePorts = extractPortsFromRule(rulePort);
            return {
                Name: item,
                ports: rulePorts.length === 1 ? rulePorts : [],
            };
        });
};

const extractPortsFromRule = (rulePort: string): number[] => {
    const ports: number[] = [];
    const segments = rulePort.split(',');
    for (const segment of segments) {
        const portSegment = segment.trim();
        if (!portSegment) {
            continue;
        }

        const rangeDelimiter = portSegment.includes('-') && !portSegment.startsWith('-') ? '-' : ':';
        if (portSegment.includes(rangeDelimiter) && !portSegment.startsWith(rangeDelimiter)) {
            const [startPort, endPort] = portSegment.split(rangeDelimiter).map((item) => parseInt(item.trim()));
            if (!isNaN(startPort) && !isNaN(endPort)) {
                for (let port = startPort; port <= endPort; port++) {
                    ports.push(port);
                }
            }
            continue;
        }

        const port = parseInt(portSegment);
        if (!isNaN(port)) {
            ports.push(port);
        }
    }
    return ports;
};

const getProtocolNums = (protocol: string): number[] => {
    const protocolValue = protocol.toLowerCase();
    if (protocolValue === 'tcp') {
        return [1];
    }
    if (protocolValue === 'udp') {
        return [2];
    }
    if (protocolValue.includes('tcp') && protocolValue.includes('udp')) {
        return [1, 2];
    }
    return [];
};

const loadMatchedListeningProcesses = (rule: RuleInfoWithProcess): ProcessInfoDisplay[] => {
    const protocolNums = getProtocolNums(rule.protocol);
    const matchedProcesses: ProcessInfoDisplay[] = [];

    for (const proc of listeningProcesses.value) {
        if (!protocolNums.includes(proc.Protocol)) {
            continue;
        }

        const matchedPorts = extractPortsFromObject(proc.Port)
            .filter((port) => isPortInRule(rule.port, port))
            .sort((a, b) => a - b);
        if (matchedPorts.length > 0) {
            matchedProcesses.push({
                ...proc,
                ports: matchedPorts,
            });
        }
    }

    return matchedProcesses;
};

const applyProcessPID = (processInfos: ProcessInfoDisplay[], matchedProcesses: ProcessInfoDisplay[]) => {
    for (const processInfo of processInfos) {
        const matchedProcess = matchedProcesses.find((proc) =>
            processInfo.ports.some((port) => proc.ports.includes(port)),
        );
        if (!matchedProcess) {
            continue;
        }

        processInfo.PID = matchedProcess.PID;
        processInfo.Port = matchedProcess.Port;
        processInfo.Protocol = matchedProcess.Protocol;
        if (processInfo.ports.length === 0) {
            processInfo.ports = matchedProcess.ports;
        }
    }
};

const mergeListeningProcesses = (processInfos: ProcessInfoDisplay[], matchedProcesses: ProcessInfoDisplay[]) => {
    const displayedPorts = new Set<number>();
    for (const processInfo of processInfos) {
        for (const port of processInfo.ports) {
            displayedPorts.add(port);
        }
    }

    for (const matchedProcess of matchedProcesses) {
        const missingPorts = matchedProcess.ports.filter((port) => !displayedPorts.has(port));
        if (missingPorts.length === 0) {
            continue;
        }

        const sameProcess = processInfos.find(
            (processInfo) => processInfo.PID && processInfo.PID === matchedProcess.PID,
        );
        if (sameProcess) {
            sameProcess.ports = [...sameProcess.ports, ...missingPorts].sort((a, b) => a - b);
        } else {
            processInfos.push({
                ...matchedProcess,
                ports: missingPorts,
            });
        }

        for (const port of missingPorts) {
            displayedPorts.add(port);
        }
    }
};

const loadListeningProcesses = async () => {
    try {
        const res = await getListeningProcess();
        listeningProcesses.value = res.data || [];

        for (const item of data.value) {
            const matchedProcesses = loadMatchedListeningProcesses(item);

            if (item.usedStatus) {
                item.expand = false;
                item.processInfos = parseUsedStatus(item.usedStatus, item.port);
                applyProcessPID(item.processInfos, matchedProcesses);
                mergeListeningProcesses(item.processInfos, matchedProcesses);
                item.processInfo = item.processInfos.find((proc) => proc.PID) as Process.ListeningProcess;
                continue;
            }

            if (matchedProcesses.length > 0) {
                item.expand = false;
                item.usedStatus = matchedProcesses.map((proc) => proc.Name).join(', ');
                item.processInfo = matchedProcesses[0] as Process.ListeningProcess;
                item.processInfos = matchedProcesses;
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
        permission: true,
        nodeAdmin: true,
        click: (row: Host.RulePort) => {
            onOpenDialog('edit', row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
        nodeAdmin: true,
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

.process-name,
.process-link {
    display: block;
}

.process-link {
    cursor: pointer;
}

.process-port {
    margin-left: 8px;
}

.process-list {
    display: block;
    margin-top: 2px;
}

.process-detail-icon {
    margin-left: 4px;
    vertical-align: middle;
}

.process-toggle {
    height: 20px;
    padding: 0;
}
</style>
