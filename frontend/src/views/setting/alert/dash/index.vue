<template>
    <div v-loading="loading">
        <LayoutContent :title="$t('xpack.alert.list')" v-loading="loading">
            <template #prompt>
                <el-alert type="info" :closable="false" class="!mt-2">
                    <template #title>
                        {{ $t('xpack.alert.agentOfflineAlertHelper') }}
                    </template>
                </el-alert>
            </template>
            <template #leftToolBar>
                <el-button type="primary" @click="openView('create')">{{ $t('xpack.alert.addTask') }}</el-button>
            </template>
            <template #rightToolBar>
                <div class="dropdowns">
                    <el-select filterable clearable v-model="req.type" @change="search()" class="!w-52 dropdown">
                        <template #prefix>{{ $t('commons.table.type') }}</template>
                        <template v-if="isMaster">
                            <el-option value="panelPwdEndTime" :label="$t('xpack.alert.panelPwdEndTime')" />
                            <el-option value="panelLogin" :label="$t('xpack.alert.panelLogin')" />
                            <el-option
                                v-if="isProductPro"
                                value="licenseException"
                                :label="$t('xpack.alert.licenseException')"
                            />
                            <el-option
                                v-if="isProductPro"
                                value="nodeException"
                                :label="$t('xpack.alert.nodeException')"
                            />
                            <el-option value="panelUpdate" :label="$t('xpack.alert.panelUpdate')" />
                        </template>
                        <el-option value="sshLogin" :label="$t('xpack.alert.sshLogin')" />
                        <el-option value="ssl" :label="$t('xpack.alert.ssl')" />
                        <el-option value="siteEndTime" :label="$t('xpack.alert.siteEndTime')" />
                        <el-option value="cpu" :label="$t('xpack.alert.cpu')" />
                        <el-option value="memory" :label="$t('xpack.alert.memory')" />
                        <el-option value="disk" :label="$t('xpack.alert.disk')" />
                        <el-option value="load" :label="$t('xpack.alert.load')" />
                        <el-option value="clams" :label="$t('xpack.alert.clams')" />
                        <el-option value="shell" :label="$t('xpack.alert.cronjob', [$t('cronjob.shell')])" />
                        <el-option value="app" :label="$t('xpack.alert.cronjob', [$t('cronjob.app')])" />
                        <el-option value="website" :label="$t('xpack.alert.cronjob', [$t('cronjob.website')])" />
                        <el-option value="database" :label="$t('xpack.alert.cronjob', [$t('cronjob.database')])" />
                        <el-option value="directory" :label="$t('xpack.alert.cronjob', [$t('cronjob.directory')])" />
                        <el-option value="log" :label="$t('xpack.alert.cronjob', [$t('cronjob.log')])" />
                        <el-option value="snapshot" :label="$t('xpack.alert.cronjob', [$t('cronjob.snapshot')])" />
                        <el-option value="curl" :label="$t('xpack.alert.cronjob', [$t('cronjob.curl')])" />
                        <el-option
                            value="cutWebsiteLog"
                            :label="$t('xpack.alert.cronjob', [$t('cronjob.cutWebsiteLog')])"
                        />
                        <el-option value="clean" :label="$t('xpack.alert.cronjob', [$t('cronjob.clean')])" />
                        <el-option value="ntp" :label="$t('xpack.alert.cronjob', [$t('cronjob.ntp')])" />
                    </el-select>
                    <el-select
                        clearable
                        filterable
                        v-model="req.status"
                        @change="search()"
                        @clear="search"
                        class="!w-52 dropdown"
                    >
                        <template #prefix>{{ $t('commons.table.status') }}</template>
                        <el-option :label="$t('commons.button.enable')" value="Enable"></el-option>
                        <el-option :label="$t('commons.button.disable')" value="Disable"></el-option>
                    </el-select>
                </div>
            </template>
            <template #main>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    :data="data"
                    :height-diff="380"
                    @sort-change="changeSort"
                    @search="search()"
                >
                    <el-table-column
                        :label="$t('commons.table.title')"
                        prop="title"
                        min-width="300px"
                        show-overflow-tooltip
                    ></el-table-column>
                    <el-table-column :label="$t('commons.table.status')" prop="status" width="110px">
                        <template #default="{ row }">
                            <el-button
                                v-if="row.status === 'Enable'"
                                @click="updateAlertStatus('disable', row.id)"
                                link
                                icon="VideoPlay"
                                type="success"
                            >
                                {{ $t('commons.status.enabled') }}
                            </el-button>
                            <el-button
                                v-else
                                icon="VideoPause"
                                link
                                type="danger"
                                @click="updateAlertStatus('enable', row.id)"
                            >
                                {{ $t('commons.status.disabled') }}
                            </el-button>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('xpack.alert.alertMethod')"
                        prop="method"
                        width="200px"
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <span v-if="row.method">{{ formatMethod(row) }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('xpack.alert.alertRule')"
                        prop="rule"
                        min-width="300px"
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            {{ formatRule(row) }}
                        </template>
                    </el-table-column>
                    <fu-table-operations
                        :ellipsis="2"
                        width="130px"
                        :buttons="buttons"
                        :label="$t('commons.table.operate')"
                        :fixed="mobile ? false : 'right'"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>
        <AddTask @search="search" ref="addTaskRef" />
    </div>
</template>

<script lang="ts" setup>
import { computed, onMounted, reactive, ref } from 'vue';
import { GlobalStore } from '@/store';
import { MsgSuccess } from '@/utils/message';
import i18n from '@/lang';
import { ElMessageBox } from 'element-plus';
import AddTask from '@/views/setting/alert/dash/task/index.vue';
import { Alert } from '@/api/interface/alert';
import { UpdateAlertStatus, SearchAlerts, DeleteAlert } from '@/api/modules/alert';
import { storeToRefs } from 'pinia';

const globalStore = GlobalStore();
const { isMaster, isProductPro } = storeToRefs(globalStore);

const { t } = i18n.global;
const loading = ref(false);
const addTaskRef = ref();

const req = reactive({
    page: 1,
    pageSize: 10,
    total: 0,
    type: '',
    status: '',
    method: '',
});

const paginationConfig = reactive({
    cacheSizeKey: 'alert-list-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('alert-list-page-size')) || 20,
    total: 0,
    orderBy: 'created_at',
    order: 'null',
});
const data = ref();

const mobile = computed(() => {
    return globalStore.isMobile();
});

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: function (row: Alert.AlertInfo) {
            openView('edit', row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: function (row: Alert.AlertInfo) {
            onDelete(row);
        },
    },
];

const openView = async (
    title: string,
    rowData: Partial<Alert.AlertInfo> = {
        type: isMaster.value ? 'panelPwdEndTime' : 'sshLogin',
        cycle: 15,
        count: 0,
        sendCount: 3,
        method: '',
        project: '',
        status: 'Enable',
        title: '',
    },
) => {
    let params = {
        title,
        rowData: { ...rowData },
    };
    addTaskRef.value.acceptParams(params);
};

const changeSort = ({ prop, order }) => {
    if (order) {
        paginationConfig.orderBy = prop == 'status' ? 'status' : prop;
        paginationConfig.order = order;
    }
    search();
};

const formatRule = (row: Alert.AlertInfo) => {
    const ruleTemplates = {
        ssl: () => t('xpack.alert.timeRule', [row.cycle, row.sendCount]),
        siteEndTime: () => t('xpack.alert.timeRule', [row.cycle, row.sendCount]),
        panelPwdEndTime: () => t('xpack.alert.timeRule', [row.cycle, row.sendCount]),
        panelUpdate: () => t('xpack.alert.panelUpdateRule', [row.cycle, row.sendCount]),
        cpu: () => t('xpack.alert.avgRule', [row.cycle, t(`xpack.alert.${row.type}Name`), row.count, row.sendCount]),
        memory: () => t('xpack.alert.avgRule', [row.cycle, t(`xpack.alert.${row.type}Name`), row.count, row.sendCount]),
        load: () => t('xpack.alert.avgRule', [row.cycle, t(`xpack.alert.${row.type}Name`), row.count, row.sendCount]),
        disk: () => {
            return row.project === 'all'
                ? t('xpack.alert.allDiskRule', [row.count, row.cycle === 1 ? 'G' : '%', row.sendCount])
                : t('xpack.alert.diskRule', [row.project, row.count, row.cycle === 1 ? 'G' : '%', row.sendCount]);
        },
        clams: () => t('xpack.alert.clamsRule', [row.sendCount]),
        app: () => t('xpack.alert.cronJobAppRule', [row.sendCount]),
        website: () => t('xpack.alert.cronJobWebsiteRule', [row.sendCount]),
        database: () => t('xpack.alert.cronJobDatabaseRule', [row.sendCount]),
        directory: () => t('xpack.alert.cronJobDirectoryRule', [row.sendCount]),
        log: () => t('xpack.alert.cronJobLogRule', [row.sendCount]),
        snapshot: () => t('xpack.alert.cronJobSnapshotRule', [row.sendCount]),
        shell: () => t('xpack.alert.cronJobShellRule', [row.sendCount]),
        curl: () => t('xpack.alert.cronJobCurlRule', [row.sendCount]),
        cutWebsiteLog: () => t('xpack.alert.cronJobCutWebsiteLogRule', [row.sendCount]),
        clean: () => t('xpack.alert.cronJobCleanRule', [row.sendCount]),
        ntp: () => t('xpack.alert.cronJobNtpRule', [row.sendCount]),
        nodeException: () => t('xpack.alert.nodeExceptionRule', [row.sendCount]),
        licenseException: () => t('xpack.alert.licenseExceptionRule', [row.sendCount]),
        panelLogin: () => t('xpack.alert.panelLoginRule', [row.sendCount]),
        sshLogin: () => t('xpack.alert.sshLoginRule', [row.sendCount]),
    };

    return ruleTemplates[row.type] ? ruleTemplates[row.type]() : '';
};

const formatMethod = (row: Alert.AlertInfo) => {
    if (!row.method) return '';

    const sendMethod = row.method.split(',').filter(Boolean);
    const methodStr = sendMethod.map((item) => t('xpack.alert.' + item)).join('｜');

    return `「${methodStr}」`;
};

const search = async () => {
    if (req.status) {
        req.status = req.status.toLowerCase() === 'enable' ? 'Enable' : 'Disable';
    }
    loading.value = true;
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        type: req.type,
        status: req.status,
        method: req.method,
        orderBy: paginationConfig.orderBy,
        order: paginationConfig.order,
    };
    try {
        const res = await SearchAlerts(params);
        data.value = res.data.items || [];
        paginationConfig.total = res.data.total || 0;
    } catch (error) {
    } finally {
        loading.value = false;
    }
};

const onDelete = (row: Alert.AlertInfo) => {
    ElMessageBox.confirm(i18n.global.t('xpack.alert.deleteMsg'), i18n.global.t('xpack.alert.deleteTitle'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        await DeleteAlert({ id: row.id });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        await search();
    });
};

const updateAlertStatus = (status: string, id: number) => {
    ElMessageBox.confirm(i18n.global.t('xpack.alert.' + status + 'Msg'), i18n.global.t('xpack.alert.changeStatus'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        let itemStatus = status.toLowerCase() === 'enable' ? 'Enable' : 'Disable';
        await UpdateAlertStatus({ id: id, status: itemStatus });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        await search();
    });
};

onMounted(() => {
    search();
});
</script>

<style lang="scss" scoped>
.dropdowns {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    flex: 1 1 auto;
    justify-content: start;
}

.search-fields {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    flex: 1 1 auto;
    margin-top: 10px;
}

.el-tag {
    cursor: default;
}
</style>
