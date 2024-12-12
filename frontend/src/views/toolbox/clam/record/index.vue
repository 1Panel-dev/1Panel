<template>
    <div v-if="recordShow" v-loading="loading">
        <div class="app-status p-mt-20">
            <el-card>
                <div>
                    <el-tag class="float-left" effect="dark" type="success">
                        {{ $t('commons.table.name') }}: {{ dialogData.rowData.name }}
                    </el-tag>
                    <el-popover
                        v-if="dialogData.rowData.path.length >= 35"
                        placement="top-start"
                        trigger="hover"
                        width="250"
                        :content="dialogData.rowData.path"
                    >
                        <template #reference>
                            <el-tag style="float: left" effect="dark" type="success">
                                {{ $t('file.path') }}: {{ dialogData.rowData.path.substring(0, 20) }}...
                            </el-tag>
                        </template>
                    </el-popover>
                    <el-tag
                        v-if="dialogData.rowData.path.length < 35"
                        class="float-left ml-5"
                        effect="dark"
                        type="success"
                    >
                        {{ $t('toolbox.clam.scanDir') }}: {{ dialogData.rowData.path }}
                    </el-tag>

                    <span class="buttons">
                        <el-button type="primary" @click="onHandle(dialogData.rowData)" link>
                            {{ $t('commons.button.handle') }}
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
                </el-row>
            </template>
            <template #main>
                <div class="mainClass">
                    <el-row :gutter="20" v-show="hasRecords" class="mainRowClass">
                        <el-col :span="7">
                            <div class="infinite-list" style="overflow: auto">
                                <el-table
                                    style="cursor: pointer"
                                    :data="records"
                                    border
                                    :show-header="false"
                                    @row-click="clickRow"
                                >
                                    <el-table-column>
                                        <template #default="{ row }">
                                            <span v-if="row.name === currentRecord.name" class="select-sign"></span>
                                            <el-tag v-if="row.status === 'Done'" type="success">
                                                {{ $t('commons.status.done') }}
                                            </el-tag>
                                            <el-tag v-if="row.status === 'Waiting'" type="info">
                                                {{ $t('commons.status.scanFailed') }}
                                            </el-tag>
                                            <span>
                                                {{ row.name }}
                                            </span>
                                        </template>
                                    </el-table-column>
                                </el-table>
                            </div>
                            <div class="page-item">
                                <el-pagination
                                    :page-size="searchInfo.pageSize"
                                    :current-page="searchInfo.page"
                                    @current-change="handleCurrentChange"
                                    @size-change="handleSizeChange"
                                    :pager-count="3"
                                    :page-sizes="[6, 8, 10, 12, 14]"
                                    small
                                    layout="total, sizes, prev, pager, next"
                                    :total="searchInfo.recordTotal"
                                />
                            </div>
                        </el-col>
                        <el-col :span="17">
                            <el-form label-position="top" :v-key="refresh">
                                <el-row>
                                    <el-form-item class="descriptionWide">
                                        <template #label>
                                            <span class="status-label">{{ $t('toolbox.clam.scanTime') }}</span>
                                        </template>
                                        <span class="status-count">
                                            {{ currentRecord?.status === 'Done' ? currentRecord?.scanTime : '-' }}
                                        </span>
                                    </el-form-item>
                                    <el-form-item class="descriptionWide">
                                        <template #label>
                                            <span class="status-label">{{ $t('toolbox.clam.infectedFiles') }}</span>
                                        </template>
                                        <span class="status-count" v-if="!hasInfectedDir()">
                                            {{ currentRecord?.status === 'Done' ? currentRecord?.infectedFiles : '-' }}
                                        </span>
                                        <div class="count" v-else>
                                            <span @click="toFolder(currentRecord?.name)">
                                                {{
                                                    currentRecord?.status === 'Done'
                                                        ? currentRecord?.infectedFiles
                                                        : '-'
                                                }}
                                            </span>
                                        </div>
                                    </el-form-item>
                                </el-row>
                                <el-row>
                                    <el-select
                                        class="descriptionWide"
                                        @change="search"
                                        v-model.number="searchInfo.tail"
                                    >
                                        <template #prefix>{{ $t('toolbox.clam.scanResult') }}</template>
                                        <el-option :value="0" :label="$t('commons.table.all')" />
                                        <el-option :value="10" :label="10" />
                                        <el-option :value="100" :label="100" />
                                        <el-option :value="200" :label="200" />
                                        <el-option :value="500" :label="500" />
                                        <el-option :value="1000" :label="1000" />
                                    </el-select>
                                    <codemirror
                                        ref="mymirror"
                                        :autofocus="true"
                                        :placeholder="$t('cronjob.noLogs')"
                                        :indent-with-tab="true"
                                        :tabSize="4"
                                        style="height: calc(100vh - 498px); width: 100%; margin-top: 5px"
                                        :lineWrapping="true"
                                        :matchBrackets="true"
                                        theme="cobalt"
                                        :styleActiveLine="true"
                                        :extensions="extensions"
                                        @ready="handleReady"
                                        v-model="logContent"
                                        :disabled="true"
                                    />
                                </el-row>
                            </el-form>
                        </el-col>
                    </el-row>
                </div>
                <div class="app-warn" v-show="!hasRecords">
                    <div>
                        <span>{{ $t('toolbox.clam.noRecords') }}</span>
                        <div>
                            <img src="@/assets/images/no_app.svg" />
                        </div>
                    </div>
                </div>
            </template>
        </LayoutContent>
    </div>
</template>

<script lang="ts" setup>
import { nextTick, onBeforeUnmount, reactive, ref, shallowRef } from 'vue';
import i18n from '@/lang';
import { ElMessageBox } from 'element-plus';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { MsgSuccess } from '@/utils/message';
import { shortcuts } from '@/utils/shortcuts';
import { Toolbox } from '@/api/interface/toolbox';
import { cleanClamRecord, getClamRecordLog, handleClamScan, searchClamRecord } from '@/api/modules/toolbox';
import { useRouter } from 'vue-router';
const router = useRouter();

const loading = ref();
const refresh = ref(false);
const hasRecords = ref();

let timer: NodeJS.Timer | null = null;

const mymirror = ref();
const extensions = [javascript(), oneDark];
const view = shallowRef();
const handleReady = (payload) => {
    view.value = payload.view;
};

const recordShow = ref(false);
interface DialogProps {
    rowData: Toolbox.ClamInfo;
}
const dialogData = ref();
const records = ref<Array<Toolbox.ClamLog>>([]);
const currentRecord = ref<Toolbox.ClamLog>();
const logContent = ref();

const acceptParams = async (params: DialogProps): Promise<void> => {
    let itemSize = Number(localStorage.getItem(searchInfo.cacheSizeKey));
    if (itemSize) {
        searchInfo.pageSize = itemSize;
    }

    recordShow.value = true;
    dialogData.value = params;
    search();
    timer = setInterval(() => {
        search();
    }, 1000 * 5);
};

const handleSizeChange = (val: number) => {
    searchInfo.pageSize = val;
    localStorage.setItem(searchInfo.cacheSizeKey, val + '');
    search();
};
const handleCurrentChange = (val: number) => {
    searchInfo.page = val;
    search();
};
const hasInfectedDir = () => {
    return (
        dialogData.value.rowData!.infectedStrategy === 'move' || dialogData.value.rowData!.infectedStrategy === 'copy'
    );
};

const timeRangeLoad = ref<[Date, Date]>([
    new Date(new Date(new Date().getTime() - 3600 * 1000 * 24 * 7).setHours(0, 0, 0, 0)),
    new Date(new Date().setHours(23, 59, 59, 999)),
]);
const searchInfo = reactive({
    cacheSizeKey: 'clam-record-page-size',
    page: 1,
    pageSize: 8,
    tail: '100',
    recordTotal: 0,
    startTime: new Date(),
    endTime: new Date(),
});

const onHandle = async (row: Toolbox.ClamInfo) => {
    loading.value = true;
    await handleClamScan(row.id)
        .then(() => {
            loading.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};
const toFolder = async (path: string) => {
    let folder = dialogData.value.rowData!.infectedDir + '/1panel-infected/' + path;
    router.push({ path: '/hosts/files', query: { path: folder } });
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
        clamID: dialogData.value.rowData!.id,
        tail: searchInfo.tail,
        startTime: searchInfo.startTime,
        endTime: searchInfo.endTime,
    };
    const res = await searchClamRecord(params);
    records.value = res.data.items;
    searchInfo.recordTotal = res.data.total;
    hasRecords.value = searchInfo.recordTotal !== 0;
    if (!hasRecords.value) {
        return;
    }
    if (!currentRecord.value) {
        currentRecord.value = records.value[0];
    }
    loadRecordLog();
};

const clickRow = async (row: Toolbox.ClamLog) => {
    currentRecord.value = row;
    loadRecordLog();
};

const loadRecordLog = async () => {
    let param = {
        tail: searchInfo.tail + '',
        clamName: dialogData.value.rowData?.name,
        recordName: currentRecord.value.name,
    };
    const res = await getClamRecordLog(param);
    if (logContent.value === res.data) {
        return;
    }
    logContent.value = res.data;
    nextTick(() => {
        const state = view.value.state;
        view.value.dispatch({
            selection: { anchor: state.doc.length, head: state.doc.length },
            scrollIntoView: true,
        });
    });
};

const onClean = async () => {
    ElMessageBox.confirm(i18n.global.t('commons.msg.clean'), i18n.global.t('commons.msg.deleteTitle'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'warning',
    }).then(async () => {
        loading.value = true;
        cleanClamRecord(dialogData.value.rowData.id)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                search();
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

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
    height: calc(100vh - 420px);
    .select-sign {
        &::before {
            float: left;
            margin-left: -3px;
            position: relative;
            width: 3px;
            height: 24px;
            content: '';
            background: $primary-color;
            border-radius: 20px;
        }
    }
    .el-tag {
        margin-left: 20px;
        margin-right: 20px;
    }
}

.descriptionWide {
    width: 40%;
}
.description {
    width: 30%;
}
.page-item {
    margin-top: 10px;
    font-size: 12px;
    float: right;
}

.count {
    span {
        font-size: 25px;
        color: $primary-color;
        font-weight: 500;
        line-height: 32px;
        cursor: pointer;
    }
}

@media only screen and (max-width: 1400px) {
    .mainClass {
        overflow: auto;
    }
    .mainRowClass {
        min-width: 1200px;
    }
}
</style>
