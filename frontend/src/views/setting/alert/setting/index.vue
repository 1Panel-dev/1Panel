<template>
    <div>
        <LayoutContent :title="$t('commons.button.set')" v-loading="loading" :divider="true">
            <template #title>
                <div class="flex items-center justify-between">
                    <span>{{ $t('xpack.alert.commonConfig') }}</span>
                    <el-button v-permission plain round size="default" @click="onChangeCommon(commonConfig.id)">
                        {{ $t('commons.button.edit') }}
                    </el-button>
                </div>
            </template>
            <template #main>
                <el-form
                    @submit.prevent
                    ref="alertFormRef"
                    :label-position="isMobile ? 'top' : 'left'"
                    label-width="120px"
                >
                    <el-row>
                        <el-col>
                            <el-form-item :label="$t('xpack.alert.sendTimeRange')" prop="sendTimeRange">
                                {{ sendTimeRange }}
                            </el-form-item>
                            <div v-if="!isMaster">
                                <el-form-item :label="$t('xpack.alert.offline')" prop="isOffline">
                                    <el-switch
                                        v-permission
                                        @change="onChangeOffline"
                                        v-model="commonConfig.config.isOffline"
                                        active-value="Enable"
                                        inactive-value="Disable"
                                    ></el-switch>
                                    <span class="input-help">{{ $t('xpack.alert.offlineHelper') }}</span>
                                </el-form-item>
                            </div>
                        </el-col>
                    </el-row>
                </el-form>
            </template>
        </LayoutContent>
        <LayoutContent :title="$t('xpack.alert.methodConfig')" v-loading="loading" :divider="true">
            <template #leftToolBar>{{ $t('xpack.alert.methodConfig') }}</template>
            <template #rightToolBar>
                <el-button v-permission type="primary" @click="onCreate">
                    {{ $t('commons.button.create') }}
                </el-button>
            </template>
            <template #main>
                <el-alert type="info" :closable="false">
                    <template #title>
                        <div class="flex items-center justify-start">
                            {{ $t('xpack.alert.alertConfigHelper') }}
                            <span v-if="!isProductPro">
                                {{ $t('commons.units.semicolon') }}{{ $t('xpack.alert.alertConfigProHelper') }}
                            </span>
                            <el-link
                                class="ml-1 text-xs"
                                type="primary"
                                target="_blank"
                                :href="docsUrl + '/user_manual/settings/#3'"
                            >
                                {{ $t('commons.button.helpDoc') }}
                            </el-link>
                        </div>
                    </template>
                </el-alert>
                <ComplexTable
                    class="mt-3"
                    :pagination-config="paginationConfig"
                    :data="allConfigs"
                    :empty-text="$t('commons.table.noData')"
                    @search="searchConfigs"
                >
                    <el-table-column :label="$t('commons.table.type')" min-width="120">
                        <template #default="{ row }">
                            <div class="flex items-center gap-2">
                                <el-icon :size="20" :style="{ color: getTypeColor(row.type) }">
                                    <component :is="getTypeIcon(row.type)" />
                                </el-icon>
                                <el-tag :type="getTypeTagType(row.type)" effect="plain" round>
                                    {{ $t('xpack.alert.' + row.type) }}
                                </el-tag>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('xpack.alert.displayName')" min-width="120" prop="displayName">
                        <template #default="{ row }">
                            <span>{{ getDisplayName(row) || '-' }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('xpack.alert.configDetail')" min-width="240" prop="details">
                        <template #default="{ row }">
                            <div
                                class="text-sm text-gray-500 cursor-pointer select-none"
                                @click="toggleDetail(row.id!)"
                            >
                                <template v-if="expandedIds.has(row.id!)">
                                    {{ getConfigDetails(row) }}
                                    <el-icon class="ml-1 align-middle text-gray-400"><View /></el-icon>
                                </template>
                                <template v-else>
                                    {{ getConfigSummary(row) }}
                                    <el-icon class="ml-1 align-middle text-gray-400"><Hide /></el-icon>
                                </template>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.status')" width="100px" prop="status">
                        <template #default="{ row }">
                            <el-switch
                                v-model="row.status"
                                active-value="Enable"
                                inactive-value="Disable"
                                @change="onStatusChange(row)"
                            />
                        </template>
                    </el-table-column>
                    <el-table-column
                        v-if="isEE"
                        :label="$t('commons.table.creator')"
                        width="100px"
                        prop="createUser"
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <span>{{ row.createUser || '-' }}</span>
                        </template>
                    </el-table-column>
                    <fu-table-operations
                        width="130px"
                        :ellipsis="2"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <AlertDrawer ref="alertDrawerRef" @search="onDrawerSearch" />
        <SendTimeRange ref="sendTimeRangeRef" @search="search" />
    </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';
import { useGlobalStore } from '@/composables/useGlobalStore';
import { ListAlertConfigs, PageAlertConfigs, DeleteAlertConfig, UpdateAlertConfig } from '@/api/modules/alert';
import { ElMessageBox } from 'element-plus';
import { Message, ChatDotRound, Bell, View, Hide } from '@element-plus/icons-vue';
import SendTimeRange from '@/views/setting/alert/setting/time-range/index.vue';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import AlertDrawer from '@/views/setting/alert/setting/drawer/index.vue';
import { Alert } from '@/api/interface/alert';

const { docsUrl, isMaster, isMobile, isProductPro, isEE } = useGlobalStore();

const loading = ref(false);
const alertDrawerRef = ref();
const sendTimeRangeRef = ref();
const sendTimeRangeValue = ref();
const sendTimeRange = ref();
const alertFormRef = ref();

const allConfigs = ref<Alert.AlertConfigInfo[]>([]);
const expandedIds = ref<Set<number>>(new Set());

const toggleDetail = (id: number) => {
    const next = new Set(expandedIds.value);
    if (next.has(id)) {
        next.delete(id);
    } else {
        next.add(id);
    }
    expandedIds.value = next;
};

const defaultCommonConfig: Alert.CommonAlertConfig = {
    id: undefined,
    type: 'common',
    title: 'xpack.alert.commonConfig',
    status: 'Enable',
    config: {
        alertSendTimeRange:
            i18n.global.t('xpack.alert.noticeAlert') +
            ': ' +
            '08:00:00 - 23:59:59' +
            ' | ' +
            i18n.global.t('xpack.alert.resourceAlert') +
            ': ' +
            '00:00:00 - 23:59:59',
        isOffline: 'Disable',
    },
};
const commonConfig = ref<Alert.CommonAlertConfig>({ ...defaultCommonConfig });
const isInitialized = ref(false);

const config = ref<Alert.AlertConfigUpdateReq>({
    id: 0,
    type: '',
    title: '',
    status: '',
    config: '',
});

const paginationConfig = reactive({
    cacheSizeKey: 'alert-config-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('alert-config-page-size')) || 20,
    total: 0,
});

const parseConfig = <T extends object>(raw: Alert.AlertConfigInfo, fallback: T): T => {
    try {
        const parsed = JSON.parse(raw.config || '{}');
        return { ...fallback, ...parsed };
    } catch {
        return { ...fallback };
    }
};

function getDisplayName(row: Alert.AlertConfigInfo): string {
    try {
        const cfg = JSON.parse(row.config || '{}');
        if (row.type === 'email') {
            return cfg.displayName || '';
        }
        if (cfg.webhooks && cfg.webhooks.length > 0) {
            return cfg.webhooks.map((w: { displayName: string }) => w.displayName).join(', ');
        }
        return cfg.displayName || '';
    } catch {
        return '';
    }
}

function getConfigDetails(row: Alert.AlertConfigInfo): string {
    try {
        const cfg = JSON.parse(row.config || '{}');
        if (row.type === 'email') {
            const recipients =
                cfg.recipients && cfg.recipients.length > 0 ? cfg.recipients.join(', ') : cfg.recipient || '';
            return `${cfg.sender || ''} → ${cfg.host || ''}:${cfg.port || ''} | ${i18n.global.t('xpack.alert.recipient')}: ${recipients}`;
        }
        if (row.type === 'sms') {
            return `${i18n.global.t('xpack.alert.phone')}: ${cfg.phone || ''}`;
        }
        if (cfg.webhooks && cfg.webhooks.length > 0) {
            return cfg.webhooks
                .map((w: { displayName: string; url: string }) => `${w.displayName}: ${w.url}`)
                .join(' | ');
        }
        if (cfg.displayName && cfg.url) {
            return `${cfg.displayName}: ${cfg.url}`;
        }
        if (cfg.url) {
            return cfg.url;
        }
        return '';
    } catch {
        return '';
    }
}

function maskString(s: string, visibleLen = 3): string {
    if (!s || s.length <= visibleLen) return s;
    return s.slice(0, visibleLen) + '***';
}

function getConfigSummary(row: Alert.AlertConfigInfo): string {
    try {
        const cfg = JSON.parse(row.config || '{}');
        if (row.type === 'email') {
            const recipientCount =
                cfg.recipients && cfg.recipients.length > 0 ? cfg.recipients.length : cfg.recipient ? 1 : 0;
            return `${maskString(cfg.sender || '')} → ${cfg.host || ''}:${cfg.port || ''} | ${i18n.global.t('xpack.alert.recipient')}: ${recipientCount}`;
        }
        if (row.type === 'sms') {
            return `${i18n.global.t('xpack.alert.phone')}: ${maskString(cfg.phone || '', 4)}`;
        }
        if (cfg.webhooks && cfg.webhooks.length > 0) {
            return cfg.webhooks.map((w: { displayName: string }) => w.displayName).join(', ');
        }
        if (cfg.displayName) {
            return cfg.displayName;
        }
        return '***';
    } catch {
        return '***';
    }
}

const getTypeIcon = (type: string) => {
    const map: Record<string, any> = {
        email: Message,
        sms: Message,
        weCom: ChatDotRound,
        dingTalk: ChatDotRound,
        feiShu: ChatDotRound,
        bark: Bell,
    };
    return map[type] || Message;
};

const getTypeColor = (type: string) => {
    const map: Record<string, string> = {
        email: '#409eff',
        sms: '#909399',
        weCom: '#67c23a',
        dingTalk: '#409eff',
        feiShu: '#7c3aed',
        bark: '#e6a23c',
    };
    return map[type] || '#909399';
};

const getTypeTagType = (type: string) => {
    const map: Record<string, string> = {
        email: '',
        sms: 'info',
        weCom: 'success',
        dingTalk: '',
        feiShu: 'danger',
        bark: 'warning',
    };
    return map[type] || 'info';
};

const searchCommonConfig = async () => {
    try {
        const res = await ListAlertConfigs();
        const commonFound = res.data.find((s: Alert.AlertConfigInfo) => s.type === 'common');
        if (commonFound) {
            const parsedConfig = parseConfig(commonFound, defaultCommonConfig.config);
            commonConfig.value = {
                id: commonFound.id,
                type: commonFound.type,
                title: commonFound.title,
                status: commonFound.status,
                config: parsedConfig,
            };
            sendTimeRangeValue.value = commonConfig.value.config.alertSendTimeRange;
            const noticeTimeRange = sendTimeRangeValue.value.noticeAlert?.sendTimeRange || '08:00:00 - 23:59:59';
            const resourceTimeRange = sendTimeRangeValue.value.resourceAlert?.sendTimeRange || '00:00:00 - 23:59:59';
            sendTimeRange.value =
                i18n.global.t('xpack.alert.noticeAlert') +
                ': ' +
                noticeTimeRange +
                ' | ' +
                i18n.global.t('xpack.alert.resourceAlert') +
                ': ' +
                resourceTimeRange;
        }
        isInitialized.value = true;
    } catch {}
};

const searchConfigs = async () => {
    loading.value = true;
    try {
        const res = await PageAlertConfigs({
            page: paginationConfig.currentPage,
            pageSize: paginationConfig.pageSize,
        });
        allConfigs.value = res.data.items || [];
        paginationConfig.total = res.data.total || 0;
    } catch {
        try {
            const res = await ListAlertConfigs();
            allConfigs.value = res.data.filter((s: Alert.AlertConfigInfo) => s.type !== 'common');
            paginationConfig.total = allConfigs.value.length;
        } catch {}
    } finally {
        loading.value = false;
    }
};

const search = async () => {
    loading.value = true;
    try {
        await Promise.all([searchCommonConfig(), searchConfigs()]);
    } finally {
        loading.value = false;
    }
};

const onDrawerSearch = () => {
    searchConfigs();
};

const onChangeCommon = (id: any) => {
    sendTimeRangeRef.value.acceptParams({
        id: id,
        sendTimeRange: sendTimeRangeValue.value,
        isOffline: commonConfig.value.config.isOffline,
    });
};

const onChangeOffline = async () => {
    if (!isInitialized.value) return;
    if (!isMaster.value && commonConfig.value.config.isOffline != '') {
        const title =
            commonConfig.value.config.isOffline == 'Enable'
                ? i18n.global.t('xpack.alert.offlineOff')
                : i18n.global.t('xpack.alert.offlineClose');
        const content =
            commonConfig.value.config.isOffline == 'Enable'
                ? i18n.global.t('xpack.alert.offlineOffHelper')
                : i18n.global.t('xpack.alert.offlineCloseHelper');
        ElMessageBox.confirm(content, title, {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
        })
            .then(async () => {
                loading.value = true;
                try {
                    config.value.id = commonConfig.value.id!;
                    config.value.type = 'common';
                    config.value.title = 'xpack.alert.commonConfig';
                    config.value.status = 'Enable';
                    config.value.config = JSON.stringify(commonConfig.value.config);
                    await UpdateAlertConfig(config.value);
                    loading.value = false;
                    await search();
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                } catch {
                    loading.value = false;
                }
            })
            .catch(() => {
                commonConfig.value.config.isOffline =
                    commonConfig.value.config.isOffline == 'Enable' ? 'Disable' : 'Enable';
            });
    }
};

const onStatusChange = (row: Alert.AlertConfigInfo) => {
    UpdateAlertConfig({
        id: row.id!,
        type: row.type,
        title: row.title,
        status: row.status,
        config: row.config,
    })
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        })
        .catch(() => {
            row.status = row.status === 'Enable' ? 'Disable' : 'Enable';
        });
};

const onDelete = (id: number) => {
    ElMessageBox.confirm(i18n.global.t('xpack.alert.deleteConfigMsg'), i18n.global.t('xpack.alert.deleteConfigTitle'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        await DeleteAlertConfig({ id: id });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        await searchConfigs();
    });
};

const onCreate = () => {
    alertDrawerRef.value.acceptParams({ type: 'email' });
};

const openEditDrawer = (row: Alert.AlertConfigInfo) => {
    let configData: Record<string, any> = {};
    try {
        configData = JSON.parse(row.config || '{}');
    } catch {}
    alertDrawerRef.value.acceptParams({
        id: row.id,
        type: row.type,
        config: configData,
        status: row.status,
        updateUser: row.updateUser,
    });
};

const buttons = computed(() => [
    {
        label: i18n.global.t('commons.button.edit'),
        permission: true,
        click: (row: Alert.AlertConfigInfo) => {
            openEditDrawer(row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        permission: true,
        click: (row: Alert.AlertConfigInfo) => {
            onDelete(row.id!);
        },
    },
]);

onMounted(async () => {
    await search();
});
</script>

<style scoped lang="scss">
.label {
    color: var(--el-text-color-placeholder);
}
</style>
