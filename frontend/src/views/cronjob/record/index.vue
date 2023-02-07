<template>
    <div v-if="recordShow" v-loading="loading">
        <div class="a-card" style="margin-top: 20px">
            <el-card>
                <div>
                    <el-tag style="float: left" effect="dark" type="success">{{ dialogData.rowData.name }}</el-tag>
                    <el-tag round class="status-content" type="success">
                        {{ $t('cronjob.' + dialogData.rowData.type) }}
                    </el-tag>
                    <el-tag class="status-content" type="info">
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
                    </el-tag>
                    <span class="buttons">
                        <el-button type="primary" @click="onHandle(dialogData.rowData)" link>
                            {{ $t('commons.button.handle') }}
                        </el-button>
                    </span>
                </div>
            </el-card>
        </div>

        <LayoutContent :title="$t('cronjob.record')" :reload="true">
            <template #main>
                <el-row :gutter="20">
                    <el-col :span="6">
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
                        <el-card style="margin-top: 20px">
                            <ul v-infinite-scroll="nextPage" class="infinite-list" style="overflow: auto">
                                <li
                                    v-for="(item, index) in records"
                                    :key="index"
                                    @click="forDetail(item, index)"
                                    class="infinite-list-item"
                                >
                                    <el-icon v-if="item.status === 'Success'"><Select /></el-icon>
                                    <el-icon v-if="item.status === 'Failed'"><CloseBold /></el-icon>
                                    <span v-if="index === currentRecordIndex" style="color: red">
                                        {{ dateFormat(0, 0, item.startTime) }}
                                    </span>
                                    <span v-else>{{ dateFormat(0, 0, item.startTime) }}</span>
                                </li>
                            </ul>
                            <div style="margin-top: 10px; margin-bottom: 5px; font-size: 12px; float: right">
                                <span>{{ $t('commons.table.total', [searchInfo.recordTotal]) }}</span>
                            </div>
                        </el-card>
                    </el-col>
                    <el-col :span="18">
                        <el-select @change="search()" v-model="searchInfo.status">
                            <el-option :label="$t('cronjob.all')" value="" />
                            <el-option :label="$t('cronjob.failedRecord')" value="Failed" />
                            <el-option :label="$t('cronjob.successRecord')" value="Success" />
                        </el-select>
                        <el-card style="height: 382px; margin-top: 20px">
                            <el-form>
                                <el-row v-if="hasScript()">
                                    <span>{{ $t('cronjob.shellContent') }}</span>
                                    <codemirror
                                        ref="mymirror"
                                        :autofocus="true"
                                        placeholder="None data"
                                        :indent-with-tab="true"
                                        :tabSize="4"
                                        style="height: 120px; width: 100%; margin-top: 5px"
                                        :lineWrapping="true"
                                        :matchBrackets="true"
                                        theme="cobalt"
                                        :styleActiveLine="true"
                                        :extensions="extensions"
                                        v-model="dialogData.rowData!.script"
                                        :readOnly="true"
                                    />
                                </el-row>
                                <el-row>
                                    <el-col :span="8" v-if="dialogData.rowData!.type === 'website'">
                                        <el-form-item :label="$t('cronjob.website')">
                                            {{ dialogData.rowData!.website }}
                                        </el-form-item>
                                    </el-col>
                                    <el-col :span="8" v-if="dialogData.rowData!.type === 'database'">
                                        <el-form-item :label="$t('cronjob.database')">
                                            {{ dialogData.rowData!.dbName }}
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
                                                <div
                                                    v-for="item in dialogData.rowData!.exclusionRules.split(';')"
                                                    :key="item"
                                                >
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
                                            {{ dateFormat(0, 0, currentRecord?.startTime) }}
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
                                                <el-tag type="danger">{{ $t('commons.table.statusFailed') }}</el-tag>
                                            </el-tooltip>
                                            <el-tag type="success" v-else>
                                                {{ $t('commons.table.statusSuccess') }}
                                            </el-tag>
                                        </el-form-item>
                                    </el-col>
                                </el-row>
                                <el-row v-if="currentRecord?.records">
                                    <span>{{ $t('commons.table.records') }}</span>
                                    <codemirror
                                        ref="mymirror"
                                        :autofocus="true"
                                        placeholder="None data"
                                        :indent-with-tab="true"
                                        :tabSize="4"
                                        style="height: 130px; width: 100%; margin-top: 5px"
                                        :lineWrapping="true"
                                        :matchBrackets="true"
                                        theme="cobalt"
                                        :styleActiveLine="true"
                                        :extensions="extensions"
                                        v-model="currentRecordDetail"
                                        :readOnly="true"
                                    />
                                </el-row>
                            </el-form>
                        </el-card>
                    </el-col>
                </el-row>
            </template>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import { reactive, ref } from 'vue';
import { Cronjob } from '@/api/interface/cronjob';
import { loadZero } from '@/utils/util';
import { loadBackupName } from '@/views/setting/helper';
import { searchRecords, download, handleOnce } from '@/api/modules/cronjob';
import { dateFormat, dateFormatForName } from '@/utils/util';
import i18n from '@/lang';
import { ElMessage } from 'element-plus';
import { LoadFile } from '@/api/modules/files';
import LayoutContent from '@/layout/layout-content.vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';

const loading = ref();

const mymirror = ref();
const extensions = [javascript(), oneDark];

interface DialogProps {
    rowData: Cronjob.CronjobInfo;
}
const recordShow = ref(false);
const dialogData = ref();
const records = ref<Array<Cronjob.Record>>();
const currentRecord = ref<Cronjob.Record>();
const currentRecordDetail = ref<string>('');
const currentRecordIndex = ref();

const acceptParams = async (params: DialogProps): Promise<void> => {
    dialogData.value = params;
    let itemSearch = {
        page: searchInfo.page,
        pageSize: searchInfo.pageSize,
        cronjobID: dialogData.value.rowData!.id,
        startTime: new Date(new Date().setHours(0, 0, 0, 0)),
        endTime: new Date(),
        status: searchInfo.status,
    };
    const res = await searchRecords(itemSearch);
    records.value = res.data.items || [];
    if (records.value.length === 0) {
        ElMessage.info(i18n.global.t('commons.msg.notRecords'));
        return;
    }
    currentRecord.value = records.value[0];
    currentRecordIndex.value = 0;
    loadRecord(currentRecord.value.records);
    searchInfo.recordTotal = res.data.total;
    recordShow.value = true;
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
const timeRangeLoad = ref<[Date, Date]>([new Date(new Date().setHours(0, 0, 0, 0)), new Date()]);
const searchInfo = reactive({
    page: 1,
    pageSize: 8,
    recordTotal: 0,
    cronjobID: 0,
    startTime: new Date(new Date().setHours(0, 0, 0, 0)),
    endTime: new Date(),
    status: '',
});

const onHandle = async (row: Cronjob.CronjobInfo) => {
    loading.value = true;
    await handleOnce(row.id)
        .then(() => {
            loading.value = false;
            ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
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
        a.download = dateFormatForName(currentRecord.value?.startTime) + '.sql.gz';
    } else {
        a.download = dateFormatForName(currentRecord.value?.startTime) + '.tar.gz';
    }
    const event = new MouseEvent('click');
    a.dispatchEvent(event);
};

const nextPage = async () => {
    if (searchInfo.pageSize >= searchInfo.recordTotal) {
        return;
    }
    searchInfo.pageSize = searchInfo.pageSize + 5;
    search();
};
const forDetail = async (row: Cronjob.Record, index: number) => {
    currentRecord.value = row;
    currentRecordIndex.value = index;
    loadRecord(row.records);
};
const loadRecord = async (path: string) => {
    if (path) {
        const res = await LoadFile({ path: path });
        currentRecordDetail.value = res.data;
    }
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

<style lang="scss" scoped>
.infinite-list {
    height: 330px;
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
</style>
