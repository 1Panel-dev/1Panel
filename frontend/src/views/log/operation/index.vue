<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('logs.operation')">
            <template #search>
                <LogRouter current="OperationLog" />
            </template>
            <template #leftToolBar>
                <el-button type="primary" plain @click="onClean()">
                    {{ $t('logs.deleteLogs') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <el-select v-model="searchGroup" @change="search()" clearable class="p-w-200">
                    <template #prefix>{{ $t('logs.resource') }}</template>
                    <el-option :label="$t('commons.table.all')" value="" />
                    <el-option :label="$t('logs.detail.apps')" value="apps" />
                    <el-option :label="$t('logs.detail.websites')" value="websites" />
                    <el-option :label="$t('logs.detail.runtimes')" value="runtimes" />
                    <el-option :label="$t('logs.detail.ai')" value="ai" />
                    <el-option :label="$t('logs.detail.databases')" value="databases" />
                    <el-option :label="$t('logs.detail.containers')" value="containers" />
                    <el-option :label="$t('menu.system')" value="hosts" />
                    <el-option :label="$t('logs.detail.files')" value="files" />
                    <el-option :label="$t('logs.detail.cronjobs')" value="cronjobs" />
                    <el-option :label="$t('logs.detail.toolbox')" value="toolbox" />
                    <el-option :label="$t('logs.detail.process')" value="process" />
                    <el-option label="WAF" value="waf" />
                    <el-option :label="$t('logs.detail.nodes')" value="nodes" />
                    <el-option :label="$t('logs.detail.tampers')" value="tampers" />
                    <el-option :label="$t('logs.detail.xsetting')" value="xsetting" />
                    <el-option :label="$t('logs.detail.licenses')" value="licenses" />
                    <el-option :label="$t('logs.detail.logs')" value="logs" />
                    <el-option :label="$t('logs.detail.settings')" value="settings" />
                    <el-option :label="$t('logs.detail.backups')" value="backups" />
                    <el-option :label="$t('logs.detail.groups')" value="groups" />
                    <el-option :label="$t('logs.detail.commands')" value="commands" />
                </el-select>
                <el-select v-model="searchStatus" @change="search()" clearable class="p-w-200">
                    <template #prefix>{{ $t('commons.table.status') }}</template>
                    <el-option :label="$t('commons.table.all')" value="" />
                    <el-option :label="$t('commons.status.success')" value="Success" />
                    <el-option :label="$t('commons.status.failed')" value="Failed" />
                </el-select>
                <el-select v-model="searchNode" @change="search()" clearable class="p-w-200">
                    <template #prefix>{{ $t('xpack.node.node') }}</template>
                    <el-option :label="$t('commons.table.all')" value="" />
                    <el-option :label="$t('xpack.node.master')" value="local" />
                    <el-option v-for="(node, index) in nodes" :key="index" :label="node.name" :value="node.name" />
                </el-select>
                <TableSearch @search="search()" v-model:searchName="searchName" />
                <TableRefresh @search="search()" />
                <TableSetting title="operation-log-refresh" @search="search()" />
            </template>
            <template #main>
                <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search" :heightDiff="370">
                    <el-table-column :label="$t('logs.resource')" prop="group" fix>
                        <template #default="{ row }">
                            <span v-if="row.source">
                                {{ $t('logs.detail.' + row.source) }}
                            </span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.operate')" min-width="150px" prop="detailZH">
                        <template #default="{ row }">
                            <span v-if="globalStore.language === 'zh' || globalStore.language === 'zh-Hant'">
                                {{ row.detailZH }}
                            </span>
                            <span v-if="globalStore.language === 'en'">{{ row.detailEN }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column v-if="globalStore.isMasterProductPro" :label="$t('xpack.node.node')" prop="node">
                        <template #default="{ row }">
                            <span>{{ row.node === 'local' ? $t('xpack.node.master') : row.node }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.status')" prop="status">
                        <template #default="{ row }">
                            <Status :status="row.status" :msg="row.message" />
                        </template>
                    </el-table-column>
                    <el-table-column
                        prop="createdAt"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                        show-overflow-tooltip
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <ConfirmDialog ref="confirmDialogRef" @confirm="onSubmitClean"></ConfirmDialog>
    </div>
</template>

<script setup lang="ts">
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import LogRouter from '@/views/log/router/index.vue';
import { dateFormat } from '@/utils/util';
import { cleanLogs, getOperationLogs } from '@/api/modules/log';
import { onMounted, reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';
import { listNodeOptions } from '@/api/modules/setting';

const loading = ref();
const data = ref();
const confirmDialogRef = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'operation-log-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const searchName = ref<string>('');
const searchGroup = ref<string>('');
const searchStatus = ref<string>('');
const searchNode = ref<string>('');
const nodes = ref();

const globalStore = GlobalStore();

const search = async () => {
    let params = {
        operation: searchName.value,
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        status: searchStatus.value,
        source: searchGroup.value,
        node: searchNode.value,
    };
    loading.value = true;
    await getOperationLogs(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            if (globalStore.language === 'zh' || globalStore.language === 'zh-Hant') {
                for (const item of data.value) {
                    item.detailZH = loadDetail(item.detailZH);
                }
            }
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const onClean = async () => {
    let params = {
        header: i18n.global.t('logs.deleteLogs'),
        operationInfo: i18n.global.t('commons.msg.delete'),
        submitInputInfo: i18n.global.t('logs.deleteLogs'),
    };
    confirmDialogRef.value!.acceptParams(params);
};

const loadDetail = (log: string) => {
    for (const [key, value] of Object.entries(replacements)) {
        if (log.indexOf(key) !== -1) {
            log = log.replace(key, '[' + i18n.global.t(value) + ']');
        }
    }
    return log;
};

const loadNodes = async () => {
    await listNodeOptions('')
        .then((res) => {
            if (!res) {
                nodes.value = [];
                return;
            }
            nodes.value = res.data || [];
        })
        .catch(() => {
            nodes.value = [];
        });
};

const replacements = {
    '[enable]': 'commons.button.enable',
    '[Enable]': 'commons.button.enable',
    '[disable]': 'commons.button.disable',
    '[Disable]': 'commons.button.disable',
    '[light]': 'setting.light',
    '[dark]': 'setting.dark',
    '[delete]': 'commons.button.delete',
    '[get]': 'commons.button.get',
    '[operate]': 'commons.table.operate',
    '[UserName]': 'commons.login.username',
    '[PanelName]': 'setting.title',
    '[Language]': 'setting.language',
    '[Theme]': 'setting.theme',
    '[MenuTabs]': 'setting.menuTabs',
    '[SessionTimeout]': 'setting.sessionTimeout',
    '[SecurityEntrance]': 'setting.entrance',
    '[ExpirationDays]': 'setting.expirationTime',
    '[ComplexityVerification]': 'setting.complexity',
    '[MFAStatus]': 'setting.mfa',
    '[MonitorStatus]': 'setting.enableMonitor',
    '[MonitorStoreDays]': 'setting.monitor',
    '[ApiInterfaceStatus]': 'setting.apiInterface',
};

const onSubmitClean = async () => {
    await cleanLogs({ logType: 'operation' });
    search();
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
};

onMounted(() => {
    if (globalStore.isMasterProductPro) {
        loadNodes();
    }
    search();
});
</script>

<style scoped lang="scss">
.tag-button {
    &.no-active {
        background: none;
        border: none;
    }
}
</style>
