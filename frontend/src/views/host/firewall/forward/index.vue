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
                        {{ $t('firewall.cleanupAction') }}
                    </el-button>
                </template>
            </FireStatus>
            <div v-if="fireName !== '-'">
                <el-card v-if="!isInit || !isBind" class="mask-prompt">
                    <span v-if="!isInit">{{ $t('firewall.initHelper', [`${fireName}-forward`]) }}</span>
                    <span v-else>{{ $t('firewall.basicStatus') }}</span>
                </el-card>
                <LayoutContent :title="$t('firewall.forwardRule', 2)" :class="{ mask: !isInit || !isBind }">
                    <template #leftToolBar>
                        <el-button v-permission v-node-admin type="primary" @click="onOpenDialog('create')">
                            {{ $t('commons.button.create') }}
                        </el-button>
                        <el-button v-permission v-node-admin @click="openRuleSync">
                            {{ $t('commons.button.sync') }}
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
                            <el-button
                                v-permission
                                v-node-admin
                                :disabled="paginationConfig.total === 0"
                                @click="onExport"
                            >
                                {{ $t('commons.button.export') }}
                            </el-button>
                        </el-button-group>
                    </template>
                    <template #rightToolBar>
                        <TableSearch @search="search()" v-model:searchName="searchName" />
                        <TableRefresh @search="search()" />
                        <TableSetting title="firewall-forward-refresh" @search="search()" />
                    </template>
                    <template #main>
                        <ComplexTable
                            :pagination-config="paginationConfig"
                            v-model:selects="selects"
                            @search="search"
                            :data="data"
                            :heightDiff="370"
                        >
                            <el-table-column type="selection" fix />
                            <el-table-column label="IP" :min-width="60" prop="family">
                                <template #default="{ row }">
                                    {{ row.family === 'ipv6' ? 'IPv6' : 'IPv4' }}
                                </template>
                            </el-table-column>
                            <el-table-column :label="$t('commons.table.protocol')" :min-width="70" prop="protocol" />
                            <el-table-column :label="$t('commons.table.status')" :min-width="90" prop="syncStatus">
                                <template #default="{ row }">
                                    <el-tag :type="syncStatusType(row.syncStatus)">
                                        {{ syncStatusLabel(row.syncStatus) }}
                                    </el-tag>
                                </template>
                            </el-table-column>
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

        <OpDialog ref="opRef" @search="search" @submit="onSubmitDelete()">
            <template #content>
                <el-form class="mt-4 mb-1" ref="deleteForm" label-position="left">
                    <el-form-item>
                        <el-checkbox v-model="forceDelete" :label="$t('website.forceDelete')" />
                        <span class="input-help">
                            {{ $t('website.forceDeleteHelper') }}
                        </span>
                    </el-form-item>
                </el-form>
            </template>
        </OpDialog>
        <OperateDialog @search="search" ref="dialogRef" />
        <ImportDialog @search="search" ref="dialogImportRef" />
        <RuleSync ref="ruleSyncRef" @search="refreshAfterSync" />
        <ConfirmDialog ref="cleanupConfirmRef" @confirm="submitCleanupBackend" />
    </div>
</template>

<script lang="ts" setup>
import OperateDialog from './operate/index.vue';
import ImportDialog from './import/index.vue';
import RuleSync from '@/views/host/firewall/sync/index.vue';
import FireRouter from '@/views/host/firewall/index.vue';
import FireStatus from '@/views/host/firewall/status/index.vue';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { onMounted, reactive, ref } from 'vue';
import { operateFirewallBackend, operateForwardRule, searchForwardRule } from '@/api/modules/firewall';
import { Firewall } from '@/api/interface/firewall';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { downloadWithContent } from '@/utils/file';
import { getCurrentDateFormatted } from '@/utils/date';
import { ElMessageBox } from 'element-plus';
const loading = ref();
const selects = ref<any>([]);
const searchName = ref();
const searchStrategy = ref('');

const isInit = ref(false);
const isBind = ref(false);
const fireName = ref();
const fireStatusRef = ref();
const ruleSyncRef = ref<InstanceType<typeof RuleSync>>();
const cleanupConfirmRef = ref<InstanceType<typeof ConfirmDialog>>();

const openRuleSync = () => {
    if (fireName.value !== 'iptables' && fireName.value !== 'nftables') return;
    ruleSyncRef.value?.acceptParams(fireName.value, 'forwarding');
};

const refreshAfterSync = async () => {
    await fireStatusRef.value?.acceptParams();
};

const cleanupBackend = () => {
    if (fireName.value !== 'iptables' && fireName.value !== 'nftables') return;
    cleanupConfirmRef.value?.acceptParams({
        header: i18n.global.t('firewall.cleanupAction'),
        operationInfo: i18n.global.t('firewall.cleanupForwardingBackendHelper', [fireName.value]),
        submitInputInfo: fireName.value,
    });
};

const submitCleanupBackend = async () => {
    if (fireName.value !== 'iptables' && fireName.value !== 'nftables') return;
    loading.value = true;
    try {
        await operateFirewallBackend({ subsystem: 'forwarding', backend: fireName.value, operation: 'cleanup' });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        await fireStatusRef.value?.acceptParams();
    } finally {
        loading.value = false;
    }
};

const syncStatusLabel = (status?: Firewall.RuleForward['syncStatus']) => {
    if (status === 'converged') return i18n.global.t('firewall.effective');
    if (status === 'runtime_only') return i18n.global.t('firewall.forwardUnsynced');
    return i18n.global.t('firewall.notEffective');
};
const syncStatusType = (status?: Firewall.RuleForward['syncStatus']) => {
    if (status === 'converged') return 'success';
    return status === 'runtime_only' ? 'warning' : 'danger';
};

const opRef = ref();
const dialogImportRef = ref();
const forceDelete = ref(false);
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
        strategy: searchStrategy.value,
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
    loading.value = true;
    await operateForwardRule({ rules: operateRules.value, forceDelete: forceDelete.value })
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};

const onImport = () => {
    dialogImportRef.value.acceptParams(fireName.value);
};

const loadAllRules = async (): Promise<Firewall.RuleForward[]> => {
    if (paginationConfig.total === 0) return [];
    const response = await searchForwardRule({
        strategy: '',
        info: '',
        page: 1,
        pageSize: Math.max(1, paginationConfig.total),
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
    const exportData = rules.map((item) => ({
        family: item.family,
        protocol: item.protocol,
        port: item.port,
        targetIP: item.targetIP,
        targetPort: item.targetPort,
        interface: item.interface,
    }));
    downloadWithContent(
        JSON.stringify(exportData, null, 2),
        `1panel-firewall-forward-${getCurrentDateFormatted()}.json`,
    );
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
    forceDelete.value = false;
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
