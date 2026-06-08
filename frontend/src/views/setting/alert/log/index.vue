<template>
    <div>
        <LayoutContent :title="$t('xpack.alert.logs')" v-loading="loading">
            <template #toolbar>
                <div class="flex justify-between gap-2 flex-wrap sm:flex-row">
                    <div class="flex flex-wrap gap-3">
                        <el-button v-permission type="primary" @click="syncAll" v-if="isProductPro && !isIntl">
                            {{ $t('commons.button.sync') }}
                        </el-button>
                        <el-button v-permission type="primary" plain @click="onClean">
                            {{ $t('xpack.alert.cleanLog') }}
                        </el-button>
                    </div>
                </div>
            </template>
            <template #main>
                <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search()">
                    <el-table-column :label="$t('xpack.alert.alertMsg')" prop="message" show-overflow-tooltip>
                        <template #default="{ row }">
                            {{ formatMessage(row.alertDetail) }}
                        </template>
                    </el-table-column>

                    <el-table-column :label="$t('xpack.alert.alertMethod')" prop="method" width="200px">
                        <template #default="{ row }">
                            {{ formatMethod(row) }}
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('commons.table.status')"
                        fix
                        show-overflow-tooltip
                        prop="status"
                        width="150px"
                    >
                        <template #default="{ row }">
                            <el-tag
                                v-if="statusConfig(row)"
                                :type="statusConfig(row).type"
                                :link="statusConfig(row).link"
                            >
                                <el-tooltip v-if="row.message" :content="row.message" placement="top" trigger="click">
                                    {{ $t(statusConfig(row).text) }}
                                </el-tooltip>
                                <template v-else>
                                    {{ $t(statusConfig(row).text) }}
                                </template>
                            </el-tag>
                        </template>
                    </el-table-column>

                    <el-table-column
                        :label="$t('commons.table.createdAt')"
                        :formatter="dateFormat"
                        prop="createdAt"
                        width="180px"
                    ></el-table-column>

                    <el-table-column :label="$t('xpack.alert.sendCount')" prop="count" width="150px">
                        <template #default="{ row }">
                            {{ formatCount(row) }}
                        </template>
                    </el-table-column>
                    <fu-table-operations
                        v-if="isProductPro && !isIntl"
                        :ellipsis="2"
                        width="130px"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        :fixed="isMobile ? false : 'right'"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';
import { dateFormat } from '@/utils/date';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';
import { Alert } from '@/api/interface/alert';
import {
    SearchAlertLogs,
    SyncAlertInfo,
    CleanAlertLogs,
    SyncAlertAll,
    SyncOfflineAlert,
    ListAlertConfigs,
    PageAlertConfigs,
} from '@/api/modules/alert';
import { ElMessageBox } from 'element-plus';
import { useGlobalStore } from '@/composables/useGlobalStore';

const { isMobile, isProductPro, isIntl, isMaster } = useGlobalStore();
const { t } = i18n.global;
const loading = ref(false);
const data = ref();
const isOffline = ref('Disable');
const configMap = ref<Map<string, Alert.AlertConfigInfo>>(new Map());

const loadConfigMap = async () => {
    try {
        const res = await PageAlertConfigs({ page: 1, pageSize: 1000 });
        const map = new Map<string, Alert.AlertConfigInfo>();
        for (const c of res.data?.items || []) {
            map.set(String(c.id), c);
        }
        configMap.value = map;
    } catch {}
};
const resourceTypes = [
    'cpu',
    'memory',
    'load',
    'disk',
    'nodeException',
    'licenseException',
    'panelLogin',
    'sshLogin',
    'panelIpLogin',
    'sshIpLogin',
];
const paginationConfig = reactive({
    cacheSizeKey: 'alert-log-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('alert-log-page-size')) || 20,
    total: 0,
});
const req = reactive({
    page: 1,
    pageSize: 10,
    total: 0,
    count: null,
    message: '',
    CreatedAt: '',
    status: '',
});

const buttons = [
    {
        label: i18n.global.t('commons.button.sync'),
        permission: true,
        click: function (row: Alert.AlertLog) {
            syncAlert(row);
        },
        disabled: (row: Alert.AlertLog) => {
            const isSms = row.method
                .split(',')
                .filter(Boolean)
                .some((item) => {
                    const config = configMap.value.get(item.trim());
                    return config ? config.type === 'sms' : item.trim() === 'sms';
                });
            return (!isSms && row.status != 'PushSuccess' && row.status != 'SyncError') || row.status == 'Success';
        },
    },
];

const statusMap = {
    PushSuccess: { type: 'success', text: 'xpack.alert.pushSuccess' },
    Pushing: { type: 'warning', text: 'xpack.alert.pushing' },
    Success: { type: 'success', text: 'xpack.alert.success' },
    Error: { type: 'danger', text: 'xpack.alert.error' },
    SyncError: { type: 'danger', text: 'xpack.alert.syncError', link: true },
    default: { type: 'danger', text: 'xpack.alert.pushError', link: true },
};

const statusConfig = (row) => {
    return statusMap[row.status] || statusMap['default'];
};

const syncAlert = (row: Alert.AlertLog) => {
    ElMessageBox.confirm(t('xpack.alert.syncAlertInfoMsg'), t('xpack.alert.syncAlertInfo'), {
        confirmButtonText: t('commons.button.confirm'),
        cancelButtonText: t('commons.button.cancel'),
    }).then(async () => {
        if (!isMaster.value && isOffline.value == 'Enable') {
            await SyncOfflineAlert();
        } else {
            await SyncAlertInfo({ id: row.id });
        }
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        await search();
    });
};

const formatMessage = (row: Alert.AlertInfo) => {
    const messageTemplates = {
        ssl: () => {
            return row.project === 'all' ? t('xpack.alert.allSslTitle') : t('xpack.alert.sslTitle', [row.project]);
        },
        siteEndTime: () => {
            return row.project === 'all'
                ? t('xpack.alert.allSiteEndTimeTitle')
                : t('xpack.alert.siteEndTimeTitle', [row.project]);
        },
        panelPwdEndTime: () => t('xpack.alert.panelPwdEndTimeTitle'),
        panelUpdate: () => t('xpack.alert.panelUpdateTitle'),
        cpu: () => t('xpack.alert.cpuTitle'),
        memory: () => t('xpack.alert.memoryTitle'),
        load: () => t('xpack.alert.loadTitle'),
        disk: () => {
            return row.project === 'all' ? t('xpack.alert.allDiskTitle') : t('xpack.alert.diskTitle', [row.project]);
        },
        clams: () => t('xpack.alert.clamsTitle', [row.project]),
        app: () => t('xpack.alert.cronJobAppTitle', [row.project]),
        website: () => t('xpack.alert.cronJobWebsiteTitle', [row.project]),
        database: () => t('xpack.alert.cronJobDatabaseTitle', [row.project]),
        directory: () => t('xpack.alert.cronJobDirectoryTitle', [row.project]),
        log: () => t('xpack.alert.cronJobLogTitle', [row.project]),
        snapshot: () => t('xpack.alert.cronJobSnapshotTitle', [row.project]),
        shell: () => t('xpack.alert.cronJobShellTitle', [row.project]),
        curl: () => t('xpack.alert.cronJobCurlTitle', [row.project]),
        cutWebsiteLog: () => t('xpack.alert.cronJobCutWebsiteLogTitle', [row.project]),
        clean: () => t('xpack.alert.cronJobCleanTitle', [row.project]),
        ntp: () => t('xpack.alert.cronJobNtpTitle', [row.project]),
        nodeException: () => t('xpack.alert.nodeException'),
        licenseException: () => t('xpack.alert.licenseException'),
        panelLogin: () => t('xpack.alert.panelLogin'),
        sshLogin: () => t('xpack.alert.sshLogin'),
        panelIpLogin: () => t('xpack.alert.panelIpLogin'),
        sshIpLogin: () => t('xpack.alert.sshIpLogin'),
    };
    let type = row.type === 'cronJob' ? row.subType : row.type;
    return messageTemplates[type] ? messageTemplates[type]() : '';
};

const formatMethod = (row: Alert.AlertLog) => {
    if (!row.method) return '-';

    const formatMethodPart = (method: string) => {
        const config = configMap.value.get(method);
        if (config) {
            const typeKey = config.type === 'email' ? 'mail' : config.type;
            const typeLabel = i18n.global.t('xpack.alert.' + typeKey);
            try {
                const cfg = JSON.parse(config.config || '{}');
                const name = cfg.displayName || cfg.sender || cfg.phone || '';
                return name ? `${name}(${typeLabel})` : typeLabel;
            } catch {
                return typeLabel;
            }
        }

        switch (method) {
            case 'mail':
            case 'email':
                return t('xpack.alert.mail');
            case 'dingTalk':
                return t('xpack.alert.dingTalk');
            case 'weCom':
                return t('xpack.alert.weCom');
            case 'feiShu':
                return t('xpack.alert.feiShu');
            case 'wechat':
                return t('xpack.alert.wechat');
            case 'sms':
                return t('xpack.alert.sms');
            case 'webhook':
                return t('xpack.alert.webhook');
            case 'bark':
                return t('xpack.alert.bark');
            default:
                return /^\d+$/.test(method) ? `#${method}` : method;
        }
    };

    return `「${row.method
        .split(',')
        .filter(Boolean)
        .map((item) => formatMethodPart(item.trim()))
        .join('｜')}」`;
};

const formatCount = (row: Alert.AlertInfo) => {
    return resourceTypes.includes(row.type) || row.type === 'cronJob' || row.type === 'clams'
        ? t('xpack.alert.daily', [row.count])
        : t('xpack.alert.cumulative', [row.count]);
};

const search = async () => {
    loading.value = true;
    if (req.status) {
        req.status = req.status.toLowerCase() === 'enable' ? 'Enable' : 'Disable';
    }
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        count: req.count,
        status: req.status,
    };
    try {
        const res = await SearchAlertLogs(params);
        data.value = res.data.items || [];
        paginationConfig.total = res.data.total || 0;
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

const syncAll = async () => {
    ElMessageBox.confirm(t('xpack.alert.syncAlertInfoMsg'), t('xpack.alert.syncAlertInfo'), {
        confirmButtonText: t('commons.button.confirm'),
        cancelButtonText: t('commons.button.cancel'),
    }).then(async () => {
        await syncAllAlert();
        await search();
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
    });
};

const syncAllAlert = async () => {
    if (!isMaster.value && isOffline.value == 'Enable') {
        await SyncOfflineAlert();
    } else {
        await SyncAlertAll();
    }
};

const onClean = async () => {
    ElMessageBox.confirm(i18n.global.t('commons.msg.clean'), i18n.global.t('xpack.alert.cleanAlertLogs'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'info',
    }).then(async () => {
        loading.value = true;
        await CleanAlertLogs()
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            })
            .catch(() => {
                loading.value = false;
            });
        await search();
    });
};

const searchAlertInfo = async () => {
    if (!isMaster.value) {
        loading.value = true;
        try {
            const res = await ListAlertConfigs();
            const commonFound = res.data.find((s: any) => s.type === 'common');
            const config: Alert.CommonConfig = JSON.parse(commonFound.config);
            isOffline.value = config.isOffline;
        } finally {
            loading.value = false;
        }
    }
};

onMounted(async () => {
    await loadConfigMap();
    await searchAlertInfo();
    if (isProductPro.value && !isIntl.value) {
        await syncAllAlert();
    }
    await search();
});
</script>
