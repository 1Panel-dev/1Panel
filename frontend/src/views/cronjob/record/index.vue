<template>
    <el-dialog v-model="cronjobVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="70%">
        <template #header>
            <div class="card-header">
                <span>{{ title }}{{ $t('cronjob.cronTask') }}</span>
            </div>
        </template>
        <el-date-picker
            @change="search()"
            v-model="timeRangeLoad"
            type="datetimerange"
            :range-separator="$t('commons.search.timeRange')"
            :start-placeholder="$t('commons.search.timeStart')"
            :end-placeholder="$t('commons.search.timeEnd')"
            :shortcuts="shortcuts"
        ></el-date-picker>
        <el-checkbox style="margin-left: 20px" @change="search()" v-model="searchInfo.status">
            {{ $t('cronjob.failedFilter') }}
        </el-checkbox>
        <el-row :gutter="20" style="margin-top: 20px">
            <el-col :span="6">
                <el-card>
                    <ul v-infinite-scroll="nextPage" class="infinite-list" style="overflow: auto">
                        <li
                            v-for="(item, index) in records"
                            :key="index"
                            @click="forDetail(item)"
                            class="infinite-list-item"
                        >
                            <el-icon v-if="item.status === 'Success'"><Select /></el-icon>
                            <el-icon v-if="item.status === 'Failed'"><CloseBold /></el-icon>
                            {{ dateFromat(0, 0, item.startTime) }}
                        </li>
                    </ul>
                    <div style="margin-top: 10px; margin-bottom: 5px; font-size: 12px; float: right">
                        <span>{{ $t('commons.table.total', [searchInfo.recordTotal]) }}</span>
                    </div>
                </el-card>
            </el-col>
            <el-col :span="18">
                <el-card style="height: 352px">
                    <el-form>
                        <el-row>
                            <el-col :span="8">
                                <el-form-item :label="$t('cronjob.taskType')">
                                    {{ dialogData.rowData?.type }}
                                </el-form-item>
                            </el-col>
                            <el-col :span="8">
                                <el-form-item :label="$t('cronjob.taskName')">
                                    {{ dialogData.rowData?.name }}
                                </el-form-item>
                            </el-col>
                            <el-col :span="8">
                                <el-form-item :label="$t('cronjob.cronSpec')">
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
                                        {{ loadWeek(dialogData.rowData?.week) }}&nbsp;
                                        {{ loadZero(dialogData.rowData?.hour) }} :
                                        {{ loadZero(dialogData.rowData?.minute) }}
                                    </span>
                                    <span v-if="dialogData.rowData?.specType === 'perNDay'">
                                        {{ dialogData.rowData?.day }}{{ $t('cronjob.day1') }},&nbsp;
                                        {{ loadZero(dialogData.rowData?.hour) }} :
                                        {{ loadZero(dialogData.rowData?.minute) }}
                                    </span>
                                    <span v-if="dialogData.rowData?.specType === 'perNHour'">
                                        {{ dialogData.rowData?.hour }}{{ $t('cronjob.hour') }},&nbsp;
                                        {{ loadZero(dialogData.rowData?.minute) }}
                                    </span>
                                    <span v-if="dialogData.rowData?.specType === 'perHour'">
                                        &nbsp;{{ loadZero(dialogData.rowData?.minute) }}
                                    </span>
                                    <span v-if="dialogData.rowData?.specType === 'perNMinute'">
                                        &nbsp;{{ dialogData.rowData?.minute }}{{ $t('cronjob.minute') }}
                                    </span>
                                    &nbsp;{{ $t('cronjob.handle') }}
                                </el-form-item>
                            </el-col>
                            <el-col :span="8" v-if="hasScript()">
                                <el-form-item :label="$t('cronjob.shellContent')">
                                    <el-popover
                                        placement="right"
                                        :width="600"
                                        trigger="click"
                                        style="white-space: pre-wrap"
                                    >
                                        <div style="margin-left: 20px; max-height: 400px; overflow: auto">
                                            <span style="white-space: pre-wrap">{{ dialogData.rowData!.script }}</span>
                                        </div>
                                        <template #reference>
                                            <el-button type="primary" link>{{ $t('commons.button.expand') }}</el-button>
                                        </template>
                                    </el-popover>
                                </el-form-item>
                            </el-col>
                            <el-col :span="8" v-if="dialogData.rowData!.type === 'website'">
                                <el-form-item :label="$t('cronjob.website')">
                                    {{ dialogData.rowData!.website }}
                                </el-form-item>
                            </el-col>
                            <el-col :span="8" v-if="dialogData.rowData!.type === 'database'">
                                <el-form-item :label="$t('cronjob.database')">
                                    {{ dialogData.rowData!.database }}
                                </el-form-item>
                            </el-col>
                            <el-col :span="8" v-if="dialogData.rowData!.type === 'directory'">
                                <el-form-item :label="$t('cronjob.directory')">
                                    {{ dialogData.rowData!.sourceDir }}
                                </el-form-item>
                            </el-col>
                            <el-col :span="8" v-if="isBackup()">
                                <el-form-item :label="$t('cronjob.target')">
                                    {{ loadBackupName(dialogData.rowData!.targetDir) }}
                                    <el-button
                                        v-if="currentRecord?.status! !== 'Failed'"
                                        type="primary"
                                        style="margin-left: 10px"
                                        link
                                        icon="Download"
                                        @click="onDownload(currentRecord!.id, dialogData.rowData!.targetDirID)"
                                    >
                                        {{ $t('file.download') }}
                                    </el-button>
                                </el-form-item>
                            </el-col>
                            <el-col :span="8" v-if="isBackup()">
                                <el-form-item :label="$t('cronjob.retainCopies')">
                                    {{ dialogData.rowData!.retainCopies }}
                                </el-form-item>
                            </el-col>
                            <el-col :span="8" v-if="dialogData.rowData!.type === 'curl'">
                                <el-form-item :label="$t('cronjob.url')">
                                    {{ dialogData.rowData!.url }}
                                </el-form-item>
                            </el-col>
                            <el-col
                                :span="8"
                                v-if="dialogData.rowData!.type === 'website' || dialogData.rowData!.type === 'directory'"
                            >
                                <el-form-item :label="$t('cronjob.exclusionRules')">
                                    <div v-if="dialogData.rowData!.exclusionRules">
                                        <div v-for="item in dialogData.rowData!.exclusionRules.split(';')" :key="item">
                                            <el-tag>{{ item }}</el-tag>
                                        </div>
                                    </div>
                                    <span v-else>-</span>
                                </el-form-item>
                            </el-col>
                        </el-row>
                        <el-row>
                            <el-col :span="8">
                                <el-form-item :label="$t('commons.search.timeStart')">
                                    {{ dateFromat(0, 0, currentRecord?.startTime) }}
                                </el-form-item>
                            </el-col>
                            <el-col :span="8">
                                <el-form-item :label="$t('commons.table.interval')">
                                    <span v-if="currentRecord?.interval! <= 1000">
                                        {{ currentRecord?.interval }} ms
                                    </span>
                                    <span v-if="currentRecord?.interval! > 1000">
                                        {{ currentRecord?.interval! / 1000 }} s
                                    </span>
                                </el-form-item>
                            </el-col>
                            <el-col :span="8">
                                <el-form-item :label="$t('commons.table.status')">
                                    <el-tooltip
                                        v-if="currentRecord?.status === 'Failed'"
                                        class="box-item"
                                        :content="currentRecord?.message"
                                        placement="top"
                                    >
                                        {{ $t('commons.table.statusFailed') }}
                                    </el-tooltip>
                                    <span v-else>{{ $t('commons.table.statusSuccess') }}</span>
                                </el-form-item>
                            </el-col>
                        </el-row>
                        <el-row>
                            <el-col :span="24">
                                <el-form-item :label="$t('commons.table.records')">
                                    <span style="color: red" v-if="currentRecord?.status! === 'Failed'">
                                        {{ currentRecord?.message }}
                                    </span>
                                    <div v-else>
                                        <el-popover
                                            placement="right"
                                            :width="600"
                                            trigger="click"
                                            style="white-space: pre-wrap"
                                        >
                                            <div style="margin-left: 20px; max-height: 400px; overflow: auto">
                                                <span style="white-space: pre-wrap">
                                                    {{ currentRecordDetail }}
                                                </span>
                                            </div>
                                            <template #reference>
                                                <el-button
                                                    type="primary"
                                                    link
                                                    @click="loadRecord(currentRecord?.records!)"
                                                >
                                                    {{ $t('commons.button.expand') }}
                                                </el-button>
                                            </template>
                                        </el-popover>
                                    </div>
                                </el-form-item>
                            </el-col>
                        </el-row>
                    </el-form>
                </el-card>
            </el-col>
        </el-row>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="cronjobVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
            </span>
        </template>
    </el-dialog>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Cronjob } from '@/api/interface/cronjob';
import { loadZero } from '@/utils/util';
import { loadBackupName } from '@/views/setting/helper';
import { searchRecords, download } from '@/api/modules/cronjob';
import { dateFromat, dateFromatForName } from '@/utils/util';
import i18n from '@/lang';
import { ElMessage } from 'element-plus';
import { LoadFile } from '@/api/modules/files';

interface DialogProps {
    rowData?: Cronjob.CronjobInfo;
}
const title = ref<string>('');
const cronjobVisiable = ref(false);
const dialogData = ref<DialogProps>({});
const records = ref<Array<Cronjob.Record>>();
const currentRecord = ref<Cronjob.Record>();
const currentRecordDetail = ref<string>('');

const acceptParams = async (params: DialogProps): Promise<void> => {
    dialogData.value = params;
    let itemSearch = {
        page: searchInfo.page,
        pageSize: searchInfo.pageSize,
        cronjobID: dialogData.value.rowData!.id,
        startTime: searchInfo.startTime,
        endTime: searchInfo.endTime,
        status: searchInfo.status ? 'Stoped' : '',
    };
    const res = await searchRecords(itemSearch);
    records.value = res.data.items || [];
    if (records.value.length === 0) {
        ElMessage.info(i18n.global.t('commons.msg.notRecords'));
        return;
    }
    currentRecord.value = records.value[0];
    searchInfo.recordTotal = res.data.total;
    cronjobVisiable.value = true;
};

const shortcuts = [
    {
        text: i18n.global.t('monitor.today'),
        value: () => {
            const end = new Date();
            const start = new Date(new Date().setHours(0, 0, 0, 0));
            return [start, end];
        },
    },
    {
        text: i18n.global.t('monitor.yestoday'),
        value: () => {
            const yestoday = new Date(new Date().getTime() - 3600 * 1000 * 24 * 1);
            const end = new Date(yestoday.setHours(23, 59, 59, 999));
            const start = new Date(yestoday.setHours(0, 0, 0, 0));
            return [start, end];
        },
    },
    {
        text: i18n.global.t('monitor.lastNDay', [3]),
        value: () => {
            const start = new Date(new Date().getTime() - 3600 * 1000 * 24 * 3);
            const end = new Date();
            return [start, end];
        },
    },
    {
        text: i18n.global.t('monitor.lastNDay', [7]),
        value: () => {
            const start = new Date(new Date().getTime() - 3600 * 1000 * 24 * 7);
            const end = new Date();
            return [start, end];
        },
    },
    {
        text: i18n.global.t('monitor.lastNDay', [30]),
        value: () => {
            const start = new Date(new Date().getTime() - 3600 * 1000 * 24 * 30);
            const end = new Date();
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
    { label: i18n.global.t('cronjob.sunday'), value: 7 },
];
const timeRangeLoad = ref<Array<any>>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
const searchInfo = reactive({
    page: 1,
    pageSize: 5,
    recordTotal: 0,
    cronjobID: 0,
    startTime: new Date(new Date().setHours(0, 0, 0, 0)),
    endTime: new Date(),
    status: false,
});

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
        status: searchInfo.status ? 'Failed' : '',
    };
    const res = await searchRecords(params);
    records.value = res.data.items || [];
    searchInfo.recordTotal = res.data.total;
};
const onDownload = async (recordID: number, backupID: number) => {
    let params = {
        recordID: recordID,
        backupAccountID: backupID,
    };
    const res = await download(params);
    const downloadUrl = window.URL.createObjectURL(new Blob([res]));
    const a = document.createElement('a');
    a.style.display = 'none';
    a.href = downloadUrl;
    if (dialogData.value.rowData!.type === 'database') {
        a.download = dateFromatForName(currentRecord.value?.startTime) + '.sql.gz';
    } else {
        a.download = dateFromatForName(currentRecord.value?.startTime) + '.tar.gz';
    }
    const event = new MouseEvent('click');
    a.dispatchEvent(event);
};

const nextPage = async () => {
    if (searchInfo.pageSize >= searchInfo.recordTotal) {
        return;
    }
    searchInfo.pageSize = searchInfo.pageSize + 3;
    search();
};
const forDetail = async (row: Cronjob.Record) => {
    currentRecord.value = row;
};
const loadRecord = async (path: string) => {
    const res = await LoadFile({ path: path });
    currentRecordDetail.value = res.data;
};
function isBackup() {
    return (
        dialogData.value.rowData!.type === 'website' ||
        dialogData.value.rowData!.type === 'database' ||
        dialogData.value.rowData!.type === 'directory'
    );
}
function hasScript() {
    return dialogData.value.rowData!.type === 'shell' || dialogData.value.rowData!.type === 'sync';
}
function loadWeek(i: number) {
    for (const week of weekOptions) {
        if (week.value === i) {
            return week.label;
        }
    }
    return '';
}

defineExpose({
    acceptParams,
});
</script>

<style scoped>
.infinite-list {
    height: 300px;
    padding: 0;
    margin: 0;
    list-style: none;
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
}
.infinite-list .infinite-list-item:hover {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 30px;
    background: var(--el-color-primary-light-9);
    margin: 10px;
    font-weight: 500;
    color: red;
}
</style>
