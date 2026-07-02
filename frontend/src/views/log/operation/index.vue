<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('logs.operation')">
            <template #search>
                <LogRouter current="OperationLog" />
            </template>
            <template #leftToolBar>
                <el-button v-permission type="primary" plain @click="onClean()">
                    {{ $t('logs.deleteLogs') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <el-select v-model="searchGroup" @change="search()" clearable class="p-w-200">
                    <template #prefix>{{ $t('logs.resource') }}</template>
                    <el-option :label="$t('commons.table.all')" value="" />
                    <el-option :label="$t('logs.detail.apps')" value="apps" />
                    <el-option :label="$t('logs.detail.openresty')" value="openresty" />
                    <el-option :label="$t('logs.detail.websites')" value="websites" />
                    <el-option :label="$t('logs.detail.monitor')" value="monitor" />
                    <el-option :label="$t('logs.detail.runtimes')" value="runtimes" />
                    <el-option :label="$t('logs.detail.ai')" value="ai" />
                    <el-option :label="$t('logs.detail.ai_proxy')" value="ai-proxy" />
                    <el-option :label="$t('logs.detail.ai_benchmark')" value="ai_benchmark" />
                    <el-option :label="$t('logs.detail.skills_hub')" value="skills-hub" />
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
                    <el-option :label="$t('logs.detail.alert')" value="alert" />
                    <el-option :label="$t('logs.detail.backups')" value="backups" />
                    <el-option :label="$t('logs.detail.groups')" value="groups" />
                    <el-option :label="$t('logs.detail.roles')" value="roles" />
                    <el-option :label="$t('logs.detail.commands')" value="commands" />
                </el-select>
                <el-select v-model="searchStatus" @change="search()" clearable class="p-w-200">
                    <template #prefix>{{ $t('commons.table.status') }}</template>
                    <el-option :label="$t('commons.table.all')" value="" />
                    <el-option :label="$t('commons.status.success')" value="Success" />
                    <el-option :label="$t('commons.status.failed')" value="Failed" />
                </el-select>
                <el-select v-if="isAdmin" v-model="searchNode" @change="search()" clearable class="p-w-200">
                    <template #prefix>{{ $t('xpack.node.node') }}</template>
                    <el-option :label="$t('commons.table.all')" value="" />
                    <el-option
                        v-for="(node, index) in nodes"
                        :key="index"
                        :label="loadNodeName(node.name)"
                        :value="node.name"
                    />
                </el-select>
                <TableSearch @search="search()" v-model:searchName="searchName" />
                <TableRefresh @search="search()" />
                <TableSetting title="operation-log-refresh" @search="search()" />
            </template>
            <template #main>
                <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search" :heightDiff="370">
                    <el-table-column :label="$t('logs.resource')" prop="group" fix>
                        <template #default="{ row }">
                            <span v-if="row.source && row.source.indexOf('-') === -1">
                                {{ $t('logs.detail.' + row.source) }}
                            </span>
                            <span v-else>{{ $t('logs.detail.' + row.source.replace('-', '_')) }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.user')" prop="user" show-overflow-tooltip />
                    <el-table-column :label="$t('commons.table.operate')" min-width="150px" prop="detailZH">
                        <template #default="{ row }">
                            <span v-if="language === 'zh' || language === 'zh-Hant'">
                                {{ row.detailZH }}
                            </span>
                            <span v-if="language === 'en'">{{ row.detailEN }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column v-if="isXpackOrEE" :label="$t('xpack.node.node')" prop="node">
                        <template #default="{ row }">
                            <span>{{ loadNodeName(row.node) }}</span>
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
import { dateFormat } from '@/utils/date';
import { cleanLogs, getOperationLogs } from '@/api/modules/log';
import { onMounted, reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { listNodes } from '@/utils/node';

const loading = ref();
const data = ref();
const confirmDialogRef = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'operation-log-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('operation-log-page-size')) || 20,
    total: 0,
});
const searchName = ref<string>('');
const searchGroup = ref<string>('');
const searchStatus = ref<string>('');
const searchNode = ref<string>('');
const nodes = ref();

const { globalStore, currentNode, isAdmin, isXpackOrEE, language } = useGlobalStore();

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
            for (const item of data.value) {
                item.detailZH = loadDetail(item.detailZH);
                item.detailEN = loadDetail(item.detailEN);
            }
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const loadNodeName = (node: string) => {
    if (node === 'local') {
        return globalStore.getMasterAlias();
    }
    return node;
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
    return log.replace(/\[([^\]]+)\]/g, (matched, token: string) => {
        const transKey = resolveReplacementKey(token);
        if (!transKey) {
            return matched;
        }
        return '[' + i18n.global.t(transKey) + ']';
    });
};

const loadNodes = async () => {
    await listNodes('all')
        .then((res) => {
            nodes.value = res || [];
        })
        .catch(() => {
            nodes.value = [];
        });
};

const normalizedReplacements: Record<string, string> = {
    enable: 'commons.button.enable',
    disable: 'commons.button.disable',
    start: 'commons.button.start',
    stop: 'commons.button.stop',
    restart: 'commons.button.restart',
    reload: 'commons.operate.reload',
    sync: 'commons.button.sync',
    update: 'commons.button.update',
    upgrade: 'commons.button.upgrade',
    open: 'commons.button.open',
    close: 'commons.button.close',
    up: 'commons.button.up',
    down: 'commons.button.down',
    login: 'commons.button.login',
    delete: 'commons.button.delete',
    create: 'commons.button.create',
    add: 'commons.button.add',
    edit: 'commons.button.edit',
    save: 'commons.button.save',
    clean: 'commons.button.clean',
    clear: 'commons.button.clean',
    get: 'commons.button.get',
    install: 'commons.button.install',
    uninstall: 'commons.button.uninstall',
    backup: 'commons.button.backup',
    recover: 'commons.button.recover',
    upload: 'commons.button.upload',
    download: 'commons.button.download',
    bind: 'commons.button.bind',
    unbind: 'commons.button.unbind',
    verify: 'commons.button.verify',
    rebuild: 'commons.operate.rebuild',
    remove: 'commons.msg.remove',
    kill: 'container.kill',
    pause: 'container.pause',
    unpause: 'container.unpause',
    allow: 'firewall.allow',
    deny: 'firewall.deny',
    accept: 'firewall.accept',
    drop: 'firewall.drop',
    reject: 'firewall.stop',
    running: 'commons.status.running',
    stopped: 'commons.status.stopped',
    success: 'commons.status.success',
    failed: 'commons.status.failed',
    created: 'commons.status.created',
    restarting: 'commons.status.restarting',
    paused: 'commons.status.paused',
    exited: 'commons.status.exited',
    dead: 'commons.status.dead',
    light: 'setting.light',
    dark: 'setting.dark',
    darkgold: 'setting.darkGold',
    auto: 'setting.auto',
    cn: 'setting.cn',
    intl: 'setting.intl',
    status: 'commons.table.status',
    all: 'commons.table.all',
    operate: 'commons.table.operate',
    true: 'commons.true',
    false: 'commons.false',
};

const exactReplacements: Record<string, string> = {
    disableBanPing: 'firewall.disableBanPing',
    enableBanPing: 'firewall.enableBanPing',
    UserName: 'commons.login.username',
    PanelName: 'setting.title',
    Language: 'setting.language',
    Theme: 'setting.theme',
    MenuTabs: 'setting.menuTabs',
    MenuAccordion: 'setting.menuAccordion',
    SessionTimeout: 'setting.sessionTimeout',
    SecurityEntrance: 'setting.entrance',
    ExpirationDays: 'setting.expirationTime',
    OpsReportExportFormat: 'xpack.opsReport.page.defaultFormat',
    OpsReportSchedule: 'xpack.opsReport.page.generationRule',
    OpsReportSavePath: 'xpack.opsReport.page.savePath',
    OpsReportThreshold: 'xpack.opsReport.page.threshold',
    OpsReportAutoExport: 'xpack.opsReport.page.autoExport',
    ComplexityVerification: 'setting.complexity',
    MFAStatus: 'setting.mfa',
    MonitorStatus: 'setting.enableMonitor',
    MonitorStoreDays: 'setting.monitor',
    ApiInterfaceStatus: 'setting.apiInterface',
    ComponentSize: 'setting.componentSize',
    Region: 'setting.region',
    SystemIP: 'setting.systemIP',
    ProxyType: 'setting.proxyType',
    ProxyUrl: 'setting.proxyUrl',
    ProxyPort: 'setting.proxyPort',
    ProxyPasswdKeep: 'setting.proxyPasswdKeep',
    ProxyDocker: 'setting.proxyDocker',
    SyncToNode: 'setting.syncToNode',
    IPWhiteList: 'setting.ipWhiteList',
    ApiKeyValidityTime: 'setting.apiKeyValidityTime',
    DeveloperMode: 'setting.developerMode',
};

const resolveReplacementKey = (token: string): string | undefined => {
    if (exactReplacements[token]) {
        return exactReplacements[token];
    }
    return normalizedReplacements[token.toLowerCase()];
};

const onSubmitClean = async () => {
    await cleanLogs({ logType: 'operation' });
    search();
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
};

onMounted(() => {
    if (isAdmin.value && isXpackOrEE.value) {
        loadNodes();
    }
    searchNode.value = isAdmin.value ? '' : currentNode.value;
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
