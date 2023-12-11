<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('logs.operation')">
            <template #toolbar>
                <el-row>
                    <el-col :xs="24" :sm="16" :md="16" :lg="16" :xl="16">
                        <el-button type="primary" class="tag-button" @click="onChangeRoute('OperationLog')">
                            {{ $t('logs.operation') }}
                        </el-button>
                        <el-button class="tag-button no-active" @click="onChangeRoute('LoginLog')">
                            {{ $t('logs.login') }}
                        </el-button>
                        <el-button class="tag-button no-active" @click="onChangeRoute('SystemLog')">
                            {{ $t('logs.system') }}
                        </el-button>
                    </el-col>
                    <el-col :xs="24" :sm="8" :md="8" :lg="8" :xl="8">
                        <TableSetting @search="search()" />
                        <div class="search-button">
                            <el-input
                                clearable
                                v-model="searchName"
                                @clear="search()"
                                suffix-icon="Search"
                                @keyup.enter="search()"
                                @change="search()"
                                :placeholder="$t('commons.button.search')"
                            ></el-input>
                        </div>
                    </el-col>
                </el-row>
            </template>
            <template #search>
                <el-select v-model="searchGroup" @change="search()" clearable>
                    <template #prefix>{{ $t('logs.resource') }}</template>
                    <el-option :label="$t('commons.table.all')" value=""></el-option>
                    <el-option :label="$t('logs.detail.apps')" value="apps"></el-option>
                    <el-option :label="$t('logs.detail.websites')" value="websites"></el-option>
                    <el-option :label="$t('logs.detail.runtimes')" value="runtimes"></el-option>
                    <el-option :label="$t('logs.detail.databases')" value="databases"></el-option>
                    <el-option :label="$t('logs.detail.containers')" value="containers"></el-option>
                    <el-option :label="$t('logs.detail.cronjobs')" value="cronjobs"></el-option>
                    <el-option :label="$t('logs.detail.files')" value="files"></el-option>
                    <el-option :label="$t('logs.detail.hosts')" value="hosts"></el-option>
                    <el-option :label="$t('logs.detail.process')" value="process"></el-option>
                    <el-option :label="$t('logs.detail.logs')" value="logs"></el-option>
                    <el-option :label="$t('logs.detail.settings')" value="settings"></el-option>
                </el-select>
                <el-select v-model="searchStatus" @change="search()" clearable style="margin-left: 10px">
                    <template #prefix>{{ $t('commons.table.status') }}</template>
                    <el-option :label="$t('commons.table.all')" value=""></el-option>
                    <el-option :label="$t('commons.status.success')" value="Success"></el-option>
                    <el-option :label="$t('commons.status.failed')" value="Failed"></el-option>
                </el-select>
                <el-button type="primary" style="margin-left: 10px" plain @click="onClean()">
                    {{ $t('logs.deleteLogs') }}
                </el-button>
            </template>
            <template #main>
                <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search">
                    <el-table-column :label="$t('logs.resource')" prop="group" fix>
                        <template #default="{ row }">
                            <span v-if="row.source">
                                {{ $t('logs.detail.' + row.source) }}
                            </span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('logs.operate')" min-width="150px" prop="detailZH">
                        <template #default="{ row }">
                            <span v-if="globalStore.language === 'zh' || globalStore.language === 'tw'">
                                {{ row.detailZH }}
                            </span>
                            <span v-if="globalStore.language === 'en'">{{ row.detailEN }}</span>
                        </template>
                    </el-table-column>

                    <el-table-column :label="$t('commons.table.status')" prop="status">
                        <template #default="{ row }">
                            <el-tag v-if="row.status === 'Success'" class="ml-2" type="success">
                                {{ $t('commons.status.success') }}
                            </el-tag>
                            <div v-else>
                                <el-popover
                                    placement="top-start"
                                    :title="$t('commons.table.message')"
                                    :width="400"
                                    trigger="hover"
                                    :content="row.message"
                                >
                                    <template #reference>
                                        <el-tag class="ml-2" type="danger">{{ $t('commons.status.failed') }}</el-tag>
                                    </template>
                                </el-popover>
                            </div>
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
import TableSetting from '@/components/table-setting/index.vue';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { dateFormat } from '@/utils/util';
import { cleanLogs, getOperationLogs } from '@/api/modules/log';
import { onMounted, reactive, ref } from '@vue/runtime-core';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';
import { useRouter } from 'vue-router';
const router = useRouter();

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

const globalStore = GlobalStore();

const search = async () => {
    let params = {
        operation: searchName.value,
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        status: searchStatus.value,
        source: searchGroup.value,
    };
    loading.value = true;
    await getOperationLogs(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            if (globalStore.language === 'zh' || globalStore.language === 'tw') {
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

const onChangeRoute = async (addr: string) => {
    router.push({ name: addr });
};

const loadDetail = (log: string) => {
    if (log.indexOf('[enable]') !== -1) {
        log = log.replace('[enable]', '[' + i18n.global.t('commons.button.enable') + ']');
    }
    if (log.indexOf('[disable]') !== -1) {
        log = log.replace('[disable]', '[' + i18n.global.t('commons.button.disable') + ']');
    }
    if (log.indexOf('[light]') !== -1) {
        log = log.replace('[light]', '[' + i18n.global.t('setting.light') + ']');
    }
    if (log.indexOf('[dark]') !== -1) {
        log = log.replace('[dark]', '[' + i18n.global.t('setting.dark') + ']');
    }
    if (log.indexOf('[delete]') !== -1) {
        log = log.replace('[delete]', '[' + i18n.global.t('commons.button.delete') + ']');
    }
    if (log.indexOf('[get]') !== -1) {
        log = log.replace('[get]', '[' + i18n.global.t('commons.button.get') + ']');
    }
    if (log.indexOf('[operate]') !== -1) {
        log = log.replace('[operate]', '[' + i18n.global.t('commons.table.operate') + ']');
    }
    if (log.indexOf('[UserName]') !== -1) {
        return log.replace('[UserName]', '[' + i18n.global.t('commons.login.username') + ']');
    }
    if (log.indexOf('[PanelName]') !== -1) {
        return log.replace('[PanelName]', '[' + i18n.global.t('setting.title') + ']');
    }
    if (log.indexOf('[Language]') !== -1) {
        return log.replace('[Language]', '[' + i18n.global.t('setting.language') + ']');
    }
    if (log.indexOf('[Theme]') !== -1) {
        return log.replace('[Theme]', '[' + i18n.global.t('setting.theme') + ']');
    }
    if (log.indexOf('[SessionTimeout]') !== -1) {
        return log.replace('[SessionTimeout]', '[' + i18n.global.t('setting.sessionTimeout') + ']');
    }
    if (log.indexOf('SecurityEntrance') !== -1) {
        return log.replace('[SecurityEntrance]', '[' + i18n.global.t('setting.entrance') + ']');
    }
    if (log.indexOf('[ExpirationDays]') !== -1) {
        return log.replace('[ExpirationDays]', '[' + i18n.global.t('setting.expirationTime') + ']');
    }
    if (log.indexOf('[ComplexityVerification]') !== -1) {
        return log.replace('[ComplexityVerification]', '[' + i18n.global.t('setting.complexity') + ']');
    }
    if (log.indexOf('[MFAStatus]') !== -1) {
        return log.replace('[MFAStatus]', '[' + i18n.global.t('setting.mfa') + ']');
    }
    if (log.indexOf('[MonitorStatus]') !== -1) {
        return log.replace('[MonitorStatus]', '[' + i18n.global.t('monitor.enableMonitor') + ']');
    }
    if (log.indexOf('[MonitorStoreDays]') !== -1) {
        return log.replace('[MonitorStoreDays]', '[' + i18n.global.t('setting.monitor') + ']');
    }
    if (log.indexOf('[MonitorStoreDays]') !== -1) {
        return log.replace('[MonitorStoreDays]', '[' + i18n.global.t('setting.monitor') + ']');
    }
    if (log.indexOf('[MonitorStoreDays]') !== -1) {
        return log.replace('[MonitorStoreDays]', '[' + i18n.global.t('setting.monitor') + ']');
    }
    return log;
};

const onSubmitClean = async () => {
    await cleanLogs({ logType: 'operation' });
    search();
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
};

onMounted(() => {
    search();
});
</script>
