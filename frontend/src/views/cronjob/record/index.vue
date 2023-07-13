<template>
    <div v-if="recordShow" v-loading="loading">
        <div class="a-card" style="margin-top: 20px">
            <el-card>
                <div>
                    <el-popover
                        v-if="dialogData.rowData.name.length >= 15"
                        placement="top-start"
                        trigger="hover"
                        width="250"
                        :content="$t('cronjob.' + dialogData.rowData.type) + ' - ' + dialogData.rowData.name"
                    >
                        <template #reference>
                            <el-tag style="float: left" effect="dark" type="success">
                                {{ $t('cronjob.' + dialogData.rowData.type) }} -
                                {{ dialogData.rowData.name.substring(0, 12) }}...
                            </el-tag>
                        </template>
                    </el-popover>
                    <el-tag v-if="dialogData.rowData.name.length < 15" style="float: left" effect="dark" type="success">
                        {{ $t('cronjob.' + dialogData.rowData.type) }} - {{ dialogData.rowData.name }}
                    </el-tag>

                    <el-tag v-if="dialogData.rowData.status === 'Enable'" round class="status-content" type="success">
                        {{ $t('commons.status.running') }}
                    </el-tag>
                    <el-tag v-if="dialogData.rowData.status === 'Disable'" round class="status-content" type="info">
                        {{ $t('commons.status.stopped') }}
                    </el-tag>
                    <el-tag class="status-content">
                        <span
                            v-if="
                                dialogData.rowData?.specType.indexOf('N') === -1 ||
                                dialogData.rowData?.specType === 'perWeek'
                            "
                        >
                            {{ $t('cronjob.' + dialogData.rowData?.specType) }}&nbsp;
                        </span>
                        <span v-else>{{ $t('cronjob.per') }}</span>
                        <span v-if="dialogData.rowData?.specType === 'perMonth'">
                            {{ dialogData.rowData?.day }}{{ $t('cronjob.day') }}&nbsp;
                            {{ loadZero(dialogData.rowData?.hour) }} :
                            {{ loadZero(dialogData.rowData?.minute) }}
                        </span>
                        <span v-if="dialogData.rowData?.specType === 'perWeek'">
                            {{ loadWeek(dialogData.rowData?.week) }}&nbsp; {{ loadZero(dialogData.rowData?.hour) }} :
                            {{ loadZero(dialogData.rowData?.minute) }}
                        </span>
                        <span v-if="dialogData.rowData?.specType === 'perNDay'">
                            {{ dialogData.rowData?.day }}{{ $t('commons.units.day') }},&nbsp;
                            {{ loadZero(dialogData.rowData?.hour) }} :
                            {{ loadZero(dialogData.rowData?.minute) }}
                        </span>
                        <span v-if="dialogData.rowData?.specType === 'perNHour'">
                            {{ dialogData.rowData?.hour }}{{ $t('commons.units.hour') }},&nbsp;
                            {{ loadZero(dialogData.rowData?.minute) }}
                        </span>
                        <span v-if="dialogData.rowData?.specType === 'perHour'">
                            &nbsp;{{ loadZero(dialogData.rowData?.minute) }}
                        </span>
                        <span v-if="dialogData.rowData?.specType === 'perNMinute'">
                            &nbsp;{{ dialogData.rowData?.minute }}{{ $t('commons.units.minute') }}
                        </span>
                        <span v-if="dialogData.rowData?.specType === 'perNSecond'">
                            &nbsp;{{ dialogData.rowData?.second }}{{ $t('commons.units.second') }}
                        </span>
                        &nbsp;{{ $t('cronjob.handle') }}
                    </el-tag>
                    <span class="buttons">
                        <el-button type="primary" @click="onHandle(dialogData.rowData)" link>
                            {{ $t('commons.button.handle') }}
                        </el-button>
                        <el-divider direction="vertical" />
                        <el-button
                            type="primary"
                            v-if="dialogData.rowData.status === 'Enable'"
                            @click="onChangeStatus(dialogData.rowData.id, 'disable')"
                            link
                        >
                            {{ $t('commons.button.disable') }}
                        </el-button>
                        <el-button
                            type="primary"
                            v-if="dialogData.rowData.status === 'Disable'"
                            @click="onChangeStatus(dialogData.rowData.id, 'enable')"
                            link
                        >
                            {{ $t('commons.button.enable') }}
                        </el-button>
                        <el-divider direction="vertical" />
                        <el-button :disabled="!hasRecords" type="primary" @click="onClean" link>
                            {{ $t('commons.button.clean') }}
                        </el-button>
                    </span>
                </div>
            </el-card>
        </div>

        <LayoutContent :title="$t('cronjob.record')" :reload="true">
            <template #search>
                <el-row :gutter="20">
                    <el-col :span="8">
                        <el-date-picker
                            style="width: calc(100% - 20px)"
                            @change="search()"
                            v-model="timeRangeLoad"
                            type="datetimerange"
                            :range-separator="$t('commons.search.timeRange')"
                            :start-placeholder="$t('commons.search.timeStart')"
                            :end-placeholder="$t('commons.search.timeEnd')"
                            :shortcuts="shortcuts"
                        ></el-date-picker>
                    </el-col>
                    <el-col :span="16">
                        <el-select @change="search()" v-model="searchInfo.status">
                            <template #prefix>{{ $t('commons.table.status') }}</template>
                            <el-option :label="$t('commons.table.all')" value="" />
                            <el-option :label="$t('commons.status.success')" value="Success" />
                            <el-option :label="$t('commons.status.waiting')" value="Waiting" />
                            <el-option :label="$t('commons.status.failed')" value="Failed" />
                        </el-select>
                    </el-col>
                </el-row>
            </template>
            <template #main>
                <div class="mainClass">
                    <el-row :gutter="20" v-show="hasRecords" class="mainRowClass">
                        <el-col :span="8">
                            <div>
                                <ul class="infinite-list" style="overflow: auto">
                                    <li
                                        v-for="(item, index) in records"
                                        :key="index"
                                        @click="forDetail(item)"
                                        class="infinite-list-item"
                                    >
                                        <el-icon v-if="item.status === 'Success'"><Select /></el-icon>
                                        <el-icon v-if="item.status === 'Waiting'"><Loading /></el-icon>
                                        <el-icon v-if="item.status === 'Failed'"><CloseBold /></el-icon>
                                        <span v-if="item.id === currentRecord.id" style="color: red">
                                            {{ dateFormat(0, 0, item.startTime) }}
                                        </span>
                                        <span v-else>{{ dateFormat(0, 0, item.startTime) }}</span>
                                    </li>
                                </ul>
                                <div style="margin-top: 10px; font-size: 12px; float: right">
                                    <el-pagination
                                        :page-size="searchInfo.pageSize"
                                        :current-page="searchInfo.page"
                                        @current-change="handleCurrentChange"
                                        @size-change="handleSizeChange"
                                        :pager-count="5"
                                        :page-sizes="[6, 8, 10, 12, 14]"
                                        small
                                        layout="total, sizes, prev, pager, next"
                                        :total="searchInfo.recordTotal"
                                    />
                                </div>
                            </div>
                        </el-col>
                        <el-col :span="16">
                            <el-form label-position="top" :v-key="refresh">
                                <el-row type="flex" justify="center">
                                    <el-form-item class="descriptionWide" v-if="isBackup()">
                                        <template #label>
                                            <span class="status-label">{{ $t('cronjob.target') }}</span>
                                        </template>
                                        <span class="status-count">{{ dialogData.rowData!.targetDir }}</span>
                                        <el-button
                                            v-if="currentRecord?.status === 'Success'"
                                            type="primary"
                                            style="margin-left: 10px"
                                            link
                                            icon="Download"
                                            @click="onDownload(currentRecord, dialogData.rowData!.targetDirID)"
                                        >
                                            {{ $t('file.download') }}
                                        </el-button>
                                    </el-form-item>
                                    <el-form-item class="description" v-if="dialogData.rowData!.type === 'website'">
                                        <template #label>
                                            <span class="status-label">{{ $t('cronjob.website') }}</span>
                                        </template>
                                        <span v-if="dialogData.rowData!.website !== 'all'" class="status-count">
                                            {{ dialogData.rowData!.website }}
                                        </span>
                                        <span v-else class="status-count">
                                            {{ $t('commons.table.all') }}
                                        </span>
                                    </el-form-item>
                                    <el-form-item class="description" v-if="dialogData.rowData!.type === 'database'">
                                        <template #label>
                                            <span class="status-label">{{ $t('cronjob.database') }}</span>
                                        </template>
                                        <span v-if="dialogData.rowData!.dbName !== 'all'" class="status-count">
                                            {{ dialogData.rowData!.dbName }}
                                        </span>
                                        <span v-else class="status-count">
                                            {{ $t('commons.table.all') }}
                                        </span>
                                    </el-form-item>
                                    <el-form-item class="description" v-if="dialogData.rowData!.type === 'directory'">
                                        <template #label>
                                            <span class="status-label">{{ $t('cronjob.directory') }}</span>
                                        </template>
                                        <span v-if="dialogData.rowData!.sourceDir.length <= 12" class="status-count">
                                            {{ dialogData.rowData!.sourceDir }}
                                        </span>
                                        <div v-else>
                                            <el-popover
                                                placement="top-start"
                                                trigger="hover"
                                                width="250"
                                                :content="dialogData.rowData!.sourceDir"
                                            >
                                                <template #reference>
                                                    <span class="status-count">
                                                        {{ dialogData.rowData!.sourceDir.substring(0, 12) }}...
                                                    </span>
                                                </template>
                                            </el-popover>
                                        </div>
                                    </el-form-item>
                                    <el-form-item class="description" v-if="isBackup()">
                                        <template #label>
                                            <span class="status-label">{{ $t('cronjob.retainCopies') }}</span>
                                        </template>
                                        <span class="status-count">{{ dialogData.rowData!.retainCopies }}</span>
                                    </el-form-item>
                                </el-row>
                                <el-form-item
                                    class="description"
                                    v-if="dialogData.rowData!.type === 'website' || dialogData.rowData!.type === 'directory'"
                                >
                                    <template #label>
                                        <span class="status-label">{{ $t('cronjob.exclusionRules') }}</span>
                                    </template>
                                    <span v-if="dialogData.rowData!.exclusionRules.length <= 12" class="status-count">
                                        {{ dialogData.rowData!.exclusionRules }}
                                    </span>
                                    <div v-else>
                                        <el-popover
                                            placement="top-start"
                                            trigger="hover"
                                            width="250"
                                            :content="dialogData.rowData!.exclusionRules"
                                        >
                                            <template #reference>
                                                <span class="status-count">
                                                    {{ dialogData.rowData!.exclusionRules.substring(0, 12) }}...
                                                </span>
                                            </template>
                                        </el-popover>
                                    </div>
                                </el-form-item>
                                <el-row type="flex" justify="center">
                                    <el-form-item class="descriptionWide">
                                        <template #label>
                                            <span class="status-label">{{ $t('commons.search.timeStart') }}</span>
                                        </template>
                                        <span class="status-count">
                                            {{ dateFormat(0, 0, currentRecord?.startTime) }}
                                        </span>
                                    </el-form-item>
                                    <el-form-item class="description">
                                        <template #label>
                                            <span class="status-label">{{ $t('commons.table.interval') }}</span>
                                        </template>
                                        <span class="status-count" v-if="currentRecord?.interval! <= 1000">
                                            {{ currentRecord?.interval }} ms
                                        </span>
                                        <span class="status-count" v-if="currentRecord?.interval! > 1000">
                                            {{ currentRecord?.interval! / 1000 }} s
                                        </span>
                                    </el-form-item>
                                    <el-form-item class="description">
                                        <template #label>
                                            <span class="status-label">{{ $t('commons.table.status') }}</span>
                                        </template>
                                        <el-tooltip v-if="currentRecord?.status === 'Failed'" placement="top">
                                            <template #content>
                                                <div style="width: 300px; word-break: break-all">
                                                    {{ currentRecord?.message }}
                                                </div>
                                            </template>
                                            <el-tag type="danger">{{ $t('commons.table.statusFailed') }}</el-tag>
                                        </el-tooltip>
                                        <el-tag type="success" v-if="currentRecord?.status === 'Success'">
                                            {{ $t('commons.table.statusSuccess') }}
                                        </el-tag>
                                        <el-tag type="info" v-if="currentRecord?.status === 'Waiting'">
                                            {{ $t('commons.table.statusWaiting') }}
                                        </el-tag>
                                    </el-form-item>
                                </el-row>
                                <el-row v-if="currentRecord?.records">
                                    <span>{{ $t('commons.table.records') }}</span>
                                    <codemirror
                                        ref="mymirror"
                                        :autofocus="true"
                                        :placeholder="$t('cronjob.noLogs')"
                                        :indent-with-tab="true"
                                        :tabSize="4"
                                        style="height: calc(100vh - 484px); width: 100%; margin-top: 5px"
                                        :lineWrapping="true"
                                        :matchBrackets="true"
                                        theme="cobalt"
                                        :styleActiveLine="true"
                                        :extensions="extensions"
                                        v-model="currentRecordDetail"
                                        :disabled="true"
                                    />
                                </el-row>
                            </el-form>
                        </el-col>
                    </el-row>
                </div>
                <div class="app-warn" v-show="!hasRecords">
                    <div>
                        <span>{{ $t('cronjob.noRecord') }}</span>
                        <div>
                            <img src="@/assets/images/no_app.svg" />
                        </div>
                    </div>
                </div>
            </template>
        </LayoutContent>

        <el-dialog
            v-model="deleteVisiable"
            :title="$t('commons.button.clean')"
            width="30%"
            :close-on-click-modal="false"
        >
            <el-form ref="deleteForm" label-position="left" v-loading="delLoading">
                <el-form-item>
                    <el-checkbox v-model="cleanData" :label="$t('cronjob.cleanData')" />
                    <span class="input-help">
                        {{ $t('cronjob.cleanDataHelper') }}
                    </span>
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="deleteVisiable = false" :disabled="delLoading">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button type="primary" @click="cleanRecord">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </el-dialog>
    </div>
</template>

<script lang="ts" setup>
import { onBeforeUnmount, reactive, ref } from 'vue';
import { Cronjob } from '@/api/interface/cronjob';
import { loadZero } from '@/utils/util';
import { searchRecords, download, handleOnce, updateStatus, cleanRecords } from '@/api/modules/cronjob';
import { dateFormat } from '@/utils/util';
import i18n from '@/lang';
import { ElMessageBox } from 'element-plus';
import { DownloadByPath, LoadFile } from '@/api/modules/files';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { MsgError, MsgInfo, MsgSuccess } from '@/utils/message';

const loading = ref();
const refresh = ref(false);
const hasRecords = ref();

let timer: NodeJS.Timer | null = null;

const mymirror = ref();
const extensions = [javascript(), oneDark];

interface DialogProps {
    rowData: Cronjob.CronjobInfo;
}
const recordShow = ref(false);
const dialogData = ref();
const records = ref<Array<Cronjob.Record>>([]);
const currentRecord = ref<Cronjob.Record>();
const currentRecordDetail = ref<string>('');

const deleteVisiable = ref();
const delLoading = ref();
const cleanData = ref();

const acceptParams = async (params: DialogProps): Promise<void> => {
    recordShow.value = true;
    dialogData.value = params;
    search();
    timer = setInterval(() => {
        search();
    }, 1000 * 5);
};

const handleSizeChange = (val: number) => {
    searchInfo.pageSize = val;
    search();
};
const handleCurrentChange = (val: number) => {
    searchInfo.page = val;
    search();
};

const shortcuts = [
    {
        text: i18n.global.t('monitor.today'),
        value: () => {
            const end = new Date(new Date().setHours(23, 59, 59, 999));
            const start = new Date(new Date().setHours(0, 0, 0, 0));
            return [start, end];
        },
    },
    {
        text: i18n.global.t('monitor.yesterday'),
        value: () => {
            const itemDate = new Date(new Date().getTime() - 3600 * 1000 * 24 * 1);
            const end = new Date(itemDate.setHours(23, 59, 59, 999));
            const start = new Date(itemDate.setHours(0, 0, 0, 0));
            return [start, end];
        },
    },
    {
        text: i18n.global.t('monitor.lastNDay', [3]),
        value: () => {
            const itemDate = new Date(new Date().getTime() - 3600 * 1000 * 24 * 3);
            const end = new Date(new Date().setHours(23, 59, 59, 999));
            const start = new Date(itemDate.setHours(0, 0, 0, 0));
            return [start, end];
        },
    },
    {
        text: i18n.global.t('monitor.lastNDay', [7]),
        value: () => {
            const itemDate = new Date(new Date().getTime() - 3600 * 1000 * 24 * 7);
            const end = new Date(new Date().setHours(23, 59, 59, 999));
            const start = new Date(itemDate.setHours(0, 0, 0, 0));
            return [start, end];
        },
    },
    {
        text: i18n.global.t('monitor.lastNDay', [30]),
        value: () => {
            const itemDate = new Date(new Date().getTime() - 3600 * 1000 * 24 * 30);
            const end = new Date(new Date().setHours(23, 59, 59, 999));
            const start = new Date(itemDate.setHours(0, 0, 0, 0));
            return [start, end];
        },
    },
];
const weekOptions = [
    { label: i18n.global.t('cronjob.monday'), value: 1 },
    { label: i18n.global.t('cronjob.tuesday'), value: 2 },
    { label: i18n.global.t('cronjob.wednesday'), value: 3 },
    { label: i18n.global.t('cronjob.thursday'), value: 4 },
    { label: i18n.global.t('cronjob.friday'), value: 5 },
    { label: i18n.global.t('cronjob.saturday'), value: 6 },
    { label: i18n.global.t('cronjob.sunday'), value: 0 },
];
const timeRangeLoad = ref<[Date, Date]>([
    new Date(new Date(new Date().getTime() - 3600 * 1000 * 24 * 7).setHours(0, 0, 0, 0)),
    new Date(new Date().setHours(23, 59, 59, 999)),
]);
const searchInfo = reactive({
    page: 1,
    pageSize: 8,
    recordTotal: 0,
    cronjobID: 0,
    startTime: new Date(),
    endTime: new Date(),
    status: '',
});

const onHandle = async (row: Cronjob.CronjobInfo) => {
    loading.value = true;
    await handleOnce(row.id)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};

const onChangeStatus = async (id: number, status: string) => {
    ElMessageBox.confirm(i18n.global.t('cronjob.' + status + 'Msg'), i18n.global.t('cronjob.changeStatus'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        let itemStatus = status === 'enable' ? 'Enable' : 'Disable';
        await updateStatus({ id: id, status: itemStatus });
        MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
        dialogData.value.rowData.status = itemStatus;
    });
};

const search = async () => {
    if (timeRangeLoad.value && timeRangeLoad.value.length === 2) {
        searchInfo.startTime = timeRangeLoad.value[0];
        searchInfo.endTime = timeRangeLoad.value[1];
    } else {
        searchInfo.startTime = new Date(new Date().setHours(0, 0, 0, 0));
        searchInfo.endTime = new Date();
    }
    let params = {
        page: searchInfo.page,
        pageSize: searchInfo.pageSize,
        cronjobID: dialogData.value.rowData!.id,
        startTime: searchInfo.startTime,
        endTime: searchInfo.endTime,
        status: searchInfo.status,
    };
    const res = await searchRecords(params);
    records.value = res.data.items;
    searchInfo.recordTotal = res.data.total;
    hasRecords.value = searchInfo.recordTotal !== 0;
    if (!hasRecords.value) {
        return;
    }
    if (!currentRecord.value) {
        currentRecord.value = records.value[0];
    } else {
        let beDelete = true;
        for (const item of records.value) {
            if (item.id === currentRecord.value.id) {
                beDelete = false;
                currentRecord.value = item;
                break;
            }
        }
        if (beDelete) {
            currentRecord.value = records.value[0];
        }
    }
    if (currentRecord.value?.records) {
        loadRecord(currentRecord.value);
    }
};

const onDownload = async (record: any, backupID: number) => {
    if (dialogData.value.rowData.dbName === 'all') {
        MsgInfo(i18n.global.t('cronjob.allOptionHelper', [i18n.global.t('database.database')]));
        return;
    }
    if (dialogData.value.rowData.website === 'all') {
        MsgInfo(i18n.global.t('cronjob.allOptionHelper', [i18n.global.t('website.website')]));
        return;
    }
    if (!record.file || record.file.indexOf('/') === -1) {
        MsgError(i18n.global.t('cronjob.errPath', [record.file]));
        return;
    }
    let params = {
        recordID: record.id,
        backupAccountID: backupID,
    };
    await download(params).then(async (res) => {
        const file = await DownloadByPath(res.data);
        const downloadUrl = window.URL.createObjectURL(new Blob([file]));
        const a = document.createElement('a');
        a.style.display = 'none';
        a.href = downloadUrl;
        if (record.file && record.file.indexOf('/') !== -1) {
            let pathItem = record.file.split('/');
            a.download = pathItem[pathItem.length - 1];
        }
        const event = new MouseEvent('click');
        a.dispatchEvent(event);
    });
};

const forDetail = async (row: Cronjob.Record) => {
    currentRecord.value = row;
    loadRecord(row);
};
const loadRecord = async (row: Cronjob.Record) => {
    if (row.status === 'Failed') {
        currentRecordDetail.value = row.records;
        return;
    }
    if (row.records) {
        const res = await LoadFile({ path: row.records });
        currentRecordDetail.value = res.data;
    }
};

const onClean = async () => {
    if (!isBackup()) {
        ElMessageBox.confirm(i18n.global.t('commons.msg.clean'), i18n.global.t('commons.msg.deleteTitle'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'warning',
        }).then(async () => {
            await cleanRecords(dialogData.value.rowData.id, cleanData.value)
                .then(() => {
                    delLoading.value = false;
                    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                    search();
                })
                .catch(() => {
                    delLoading.value = false;
                });
        });
    } else {
        deleteVisiable.value = true;
    }
};

const cleanRecord = async () => {
    delLoading.value = true;
    await cleanRecords(dialogData.value.rowData.id, cleanData.value)
        .then(() => {
            delLoading.value = false;
            deleteVisiable.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .catch(() => {
            delLoading.value = false;
        });
};

function isBackup() {
    return (
        dialogData.value.rowData!.type === 'website' ||
        dialogData.value.rowData!.type === 'database' ||
        dialogData.value.rowData!.type === 'directory'
    );
}
function loadWeek(i: number) {
    for (const week of weekOptions) {
        if (week.value === i) {
            return week.label;
        }
    }
    return '';
}

onBeforeUnmount(() => {
    clearInterval(Number(timer));
    timer = null;
});

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.infinite-list {
    height: calc(100vh - 435px);
    padding: 0;
    margin: 0;
}
.infinite-list .infinite-list-item {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 30px;
    background: var(--el-color-primary-light-9);
    margin: 10px;
    color: var(--el-color-primary);
    cursor: pointer;
    &:hover {
        display: flex;
        align-items: center;
        justify-content: center;
        height: 30px;
        background: var(--el-color-primary-light-9);
        margin: 10px;
        font-weight: 500;
        color: red;
    }
}
.a-card {
    font-size: 17px;
    .el-card {
        --el-card-padding: 12px;
        .buttons {
            margin-left: 100px;
        }
    }
}
.status-content {
    float: left;
    margin-left: 50px;
}

.app-warn {
    text-align: center;
    margin-top: 100px;
    span:first-child {
        color: #bbbfc4;
    }

    span:nth-child(2) {
        color: $primary-color;
        cursor: pointer;
    }

    span:nth-child(2):hover {
        color: #74a4f3;
    }

    img {
        width: 300px;
        height: 300px;
    }
}
.descriptionWide {
    width: 40%;
}
.description {
    width: 30%;
}

@media only screen and (max-width: 1000px) {
    .mainClass {
        overflow: auto;
    }
    .mainRowClass {
        min-width: 900px;
    }
}
</style>
