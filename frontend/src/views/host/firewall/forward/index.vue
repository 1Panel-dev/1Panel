<template>
    <div>
        <FireRouter />

        <div v-loading="loading">
            <FireStatus
                ref="fireStatusRef"
                @search="search"
                v-model:loading="loading"
                v-model:is-init="isInit"
                v-model:is-bind="isBind"
                v-model:name="fireName"
                current-tab="forward"
            >
                <template v-if="isInit" #actions>
                    <el-divider direction="vertical" />
                    <el-button v-permission v-node-admin type="primary" link @click="cleanupBackend">
                        {{ $t('commons.button.reset') }}
                    </el-button>
                </template>
            </FireStatus>
            <div v-if="fireName !== '-'">
                <el-card v-if="!isInit || !isBind" class="mask-prompt">
                    <span v-if="!isInit">{{ $t('firewall.initHelper', [`${fireName}-forward`]) }}</span>
                    <span v-else>{{ $t('firewall.basicStatus') }}</span>
                </el-card>
                <LayoutContent :title="$t('firewall.forwardRule', 2)" :class="{ mask: !isInit || !isBind }">
                    <template v-if="fireStatusRef?.forwardDropFamilies" #prompt>
                        <el-alert type="warning" :closable="false" :title="$t('firewall.forwardPolicyDropWarning')" />
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
                            <el-button v-permission v-node-admin :disabled="loading" @click="onExport">
                                {{ $t('commons.button.export') }}
                            </el-button>
                        </el-button-group>
                    </template>
                    <template #rightToolBar>
                        <TableSearch @search="search()" v-model:searchName="searchName" />
                        <TableRefresh @search="refreshRules" />
                        <TableSetting title="firewall-forward-refresh" @search="refreshRules" />
                    </template>
                    <template #main>
                        <ComplexTable
                            :selection-context="() => [searchName, fireName]"
                            :row-key="
                                (row) =>
                                    JSON.stringify([
                                        row.family,
                                        row.protocol,
                                        row.port,
                                        row.targetIP,
                                        row.targetPort,
                                        row.interface,
                                    ])
                            "
                            :pagination-config="paginationConfig"
                            v-model:selects="selects"
                            @search="search"
                            :data="data"
                            :heightDiff="320"
                        >
                            <el-table-column type="selection" fix />
                            <el-table-column label="IP" :min-width="60" prop="family">
                                <template #default="{ row }">
                                    {{ row.family === 'ipv6' ? 'IPv6' : 'IPv4' }}
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('commons.table.protocol')" :min-width="70" prop="protocol" />
                            <el-table-column :label="$t('firewall.sourcePort')" :min-width="70" prop="port" />
                            <el-table-column :min-width="80" :label="$t('firewall.targetIP')" prop="targetIP" />
                            <el-table-column :label="$t('firewall.targetPort')" :min-width="70" prop="targetPort" />
                            <template v-if="fireName === 'iptables' || fireName === 'nftables'">
                                <el-table-column
                                    :label="$t('firewall.forwardInboundInterface')"
                                    :min-width="70"
                                    prop="interface"
                                >
                                    <template #default="{ row }">
                                        <span>
                                            {{ row.interface === '' ? $t('commons.table.all') : row.interface }}
                                        </span>
                                    </template>
                                </el-table-column>
                            </template>
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

        <OpDialog ref="opRef" @search="search" @submit="onSubmitDelete()" />
        <OperateDialog @created="openRuleTask" ref="dialogRef" />
        <ImportDialog @created="openRuleTask" ref="dialogImportRef" />
        <TaskLog ref="taskLogRef" @close="refreshRules" />
        <RuleReset ref="cleanupConfirmRef" @confirm="submitCleanupBackend" />
    </div>
</template>

<script lang="ts" setup>
import { useSearchPersistence } from '@/composables/useSearchPersistence';
import OperateDialog from './operate/index.vue';
import ImportDialog from './import/index.vue';
import FireRouter from '@/views/host/firewall/index.vue';
import FireStatus from '@/views/host/firewall/status/index.vue';
import RuleReset from '@/views/host/firewall/components/rule-reset.vue';
import TaskLog from '@/components/log/task/index.vue';
import { onMounted, reactive, ref } from 'vue';
import { resetFirewallRules, operateForwardRule, searchForwardRule } from '@/api/modules/firewall';
import { Firewall } from '@/api/interface/firewall';
import { buildForwardRuleExport } from './transfer';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import { getErrorMessage } from '@/utils/misc';
import { isAxiosError } from 'axios';
import { downloadWithContent } from '@/utils/file';
import { getCurrentDateFormatted } from '@/utils/date';
import { ElMessageBox } from 'element-plus';
const loading = ref();
const selects = ref<any>([]);
const searchName = ref();
useSearchPersistence('host/firewall/forward/index', { search: searchName });

const isInit = ref(false);
const isBind = ref(false);
const fireName = ref();
const fireStatusRef = ref();
const cleanupConfirmRef = ref<InstanceType<typeof RuleReset>>();
const taskLogRef = ref<InstanceType<typeof TaskLog>>();
const openRuleTask = (taskID: string) => taskLogRef.value?.openWithTaskID(taskID, true);

const refreshRules = async () => {
    await fireStatusRef.value?.acceptParams();
};

const cleanupBackend = () => {
    if (fireName.value !== 'iptables' && fireName.value !== 'nftables') return;
    cleanupConfirmRef.value?.acceptParams({
        message: i18n.global.t('firewall.cleanupForwardingBackendHelper', [fireName.value]),
        provider: fireName.value,
    });
};

const submitCleanupBackend = async (backup: boolean) => {
    if (fireName.value !== 'iptables' && fireName.value !== 'nftables') return;
    loading.value = true;
    try {
        const result = await resetFirewallRules({
            subsystem: 'forwarding',
            provider: fireName.value,
            backup,
        });
        if (result.data.backupPath)
            await ElMessageBox.alert(result.data.backupPath, i18n.global.t('commons.button.export'));
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        await fireStatusRef.value?.acceptParams();
    } finally {
        loading.value = false;
    }
};

const opRef = ref();
const dialogImportRef = ref();
const operateRules = ref();

const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'firewall-forward-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('firewall-forward-page-size')) || 20,
    total: 0,
});

const search = async () => {
    if (!isInit.value || !isBind.value || fireName.value === '-') {
        loading.value = false;
        data.value = [];
        paginationConfig.total = 0;
        return;
    }
    let params = {
        strategy: '',
        info: searchName.value,
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };
    loading.value = true;
    await searchForwardRule(params)
        .then((res) => {
            loading.value = false;
            data.value =
                res.data.items?.map((item) => {
                    return {
                        ...item,
                        interface: item.interface === '*' ? '' : item.interface,
                    };
                }) || [];
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const dialogRef = ref();
const onOpenDialog = async (
    title: string,
    rowData: Partial<Firewall.RuleForward> = {
        family: 'ipv4',
        protocol: 'tcp',
        port: '8080',
        targetIP: '',
        targetPort: '',
        interface: '',
    },
) => {
    let params = {
        title,
        rowData: { ...rowData },
    };
    dialogRef.value!.acceptParams(params);
};
const onDelete = async (row: Firewall.RuleForward | null) => {
    let names = [];
    let rules = [];
    if (row) {
        rules.push({
            ...row,
            operation: 'remove',
        });
        names = [row.port + ' (' + row.protocol + ')'];
    } else {
        for (const item of selects.value) {
            names.push(item.port + ' (' + item.protocol + ')');
            rules.push({
                ...item,
                operation: 'remove',
            });
        }
    }
    operateRules.value = rules;
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('firewall.forwardRule'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: null,
        params: null,
    });
};
const onSubmitDelete = async () => {
    if (loading.value) return;
    loading.value = true;
    try {
        const result = (await operateForwardRule({ rules: operateRules.value })).data;
        if (!result.taskID || !result.queued) {
            MsgError(i18n.global.t('commons.msg.operationFailed'));
            return;
        }
        openRuleTask(result.taskID);
    } catch (error) {
        MsgError(
            (isAxiosError(error) && error.response?.data?.message) ||
                (error && getErrorMessage(error)) ||
                i18n.global.t('commons.res.commonError'),
        );
    } finally {
        loading.value = false;
    }
};

const onImport = () => {
    dialogImportRef.value.acceptParams(fireName.value);
};

const loadAllRules = async (): Promise<Firewall.RuleForward[]> => {
    const response = await searchForwardRule({
        all: true,
        strategy: '',
        info: '',
        page: 1,
        pageSize: paginationConfig.pageSize,
    });
    return (response.data.items || []).map((item) => ({
        operation: '',
        family: item.family === 'ipv6' ? 'ipv6' : 'ipv4',
        protocol: item.protocol,
        port: item.port,
        targetIP: item.targetIP,
        targetPort: item.targetPort,
        interface: item.interface || '',
    }));
};

const exportRules = async (rules: Firewall.RuleForward[]) => {
    if (rules.length === 0) return;
    await ElMessageBox.confirm(
        i18n.global.t('firewall.exportHelper', [rules.length]),
        i18n.global.t('commons.button.export'),
        {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        },
    );
    const exported = buildForwardRuleExport(rules);
    downloadWithContent(JSON.stringify(exported, null, 2), `1panel-firewall-forward-${getCurrentDateFormatted()}.json`);
};

const onExport = async () => exportRules(selects.value.length > 0 ? selects.value : await loadAllRules());

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        permission: true,
        nodeAdmin: true,
        click: (row: Firewall.RuleForward) => {
            onOpenDialog('edit', row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
        nodeAdmin: true,
        click: (row: Firewall.RuleForward) => {
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
