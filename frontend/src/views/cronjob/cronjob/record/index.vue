<template>
    <div v-if="recordShow" v-loading="loading">
        <div class="app-status card-interval">
            <el-card>
                <div class="flex w-full flex-col gap-4 md:flex-row">
                    <div class="flex flex-wrap gap-4 ml-3">
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
                        <el-tag
                            v-if="dialogData.rowData.name.length < 15"
                            class="float-left"
                            effect="dark"
                            type="success"
                        >
                            {{ $t('cronjob.' + dialogData.rowData.type) }} - {{ dialogData.rowData.name }}
                        </el-tag>

                        <Status class="mt-0.5" :status="dialogData.rowData.status" />
                    </div>
                    <div class="mt-0.5">
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
                    </div>
                </div>
            </el-card>
        </div>

        <LayoutContent :title="$t('cronjob.record')" :reload="true">
            <template #rightToolBar>
                <el-date-picker
                    class="mr-2.5"
                    @change="search()"
                    v-model="timeRangeLoad"
                    type="datetimerange"
                    range-separator="-"
                    :start-placeholder="$t('commons.search.timeStart')"
                    :end-placeholder="$t('commons.search.timeEnd')"
                    :shortcuts="shortcuts"
                ></el-date-picker>
                <el-select @change="search()" v-model="searchInfo.status" class="p-w-200">
                    <template #prefix>{{ $t('commons.table.status') }}</template>
                    <el-option :label="$t('commons.table.all')" value="" />
                    <el-option :label="$t('commons.status.success')" value="Success" />
                    <el-option :label="$t('commons.status.waiting')" value="Waiting" />
                    <el-option :label="$t('commons.status.failed')" value="Failed" />
                </el-select>
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
                                    @row-click="forDetail"
                                >
                                    <el-table-column>
                                        <template #default="{ row }">
                                            <span v-if="row.id === currentRecord.id" class="select-sign"></span>
                                            <Status class="mr-2 ml-1 float-left" :status="row.status" />
                                            <div class="mt-0.5">
                                                <span>
                                                    {{ row.startTime }}
                                                </span>
                                            </div>
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
                                        <el-button link v-if="!currentRecord?.interval" :loading="true" />
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
                                        <Status :status="currentRecord?.status" />
                                    </el-form-item>
                                </el-row>
                                <el-row v-if="currentRecord?.status === 'Failed'">
                                    <el-form-item class="w-full">
                                        <template #label>
                                            <span class="status-label">{{ $t('commons.table.message') }}</span>
                                        </template>
                                        {{ currentRecord?.message }}
                                    </el-form-item>
                                </el-row>
                                <el-row v-if="currentRecord?.records">
                                    <span>{{ $t('commons.table.records') }}</span>
                                    <div class="editor-main">
                                        <highlightjs
                                            ref="mymirror"
                                            language="JavaScript"
                                            :autodetect="false"
                                            :code="currentRecordDetail"
                                        ></highlightjs>
                                    </div>
                                </el-row>
                                <el-row v-if="currentRecord?.taskID && currentRecord?.taskID != ''">
                                    <LogFile
                                        :defaultButton="false"
                                        class="w-full"
                                        :key="currentRecord?.taskID"
                                        :heightDiff="410"
                                        :config="{ type: 'task', taskID: currentRecord?.taskID, tail: true }"
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

        <DialogPro v-model="open" :title="$t('commons.button.clean')" size="small">
            <el-form ref="deleteForm" label-position="left" v-loading="delLoading">
                <el-form-item>
                    <el-checkbox v-model="cleanData" :label="$t('cronjob.cleanData')" />
                    <el-checkbox v-if="cleanData" v-model="cleanRemoteData" :label="$t('cronjob.cleanRemoteData')" />
                    <span class="input-help">
                        {{ $t('cronjob.cleanDataHelper') }}
                    </span>
                </el-form-item>
            </el-form>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="open = false" :disabled="delLoading">
                        {{ $t('commons.button.cancel') }}
                    </el-button>
                    <el-button type="primary" @click="cleanRecord">
                        {{ $t('commons.button.confirm') }}
                    </el-button>
                </span>
            </template>
        </DialogPro>
    </div>
</template>

<script lang="ts" setup>
import { nextTick, onBeforeUnmount, reactive, ref } from 'vue';
import { Cronjob } from '@/api/interface/cronjob';
import { searchRecords, handleOnce, updateStatus, cleanRecords, getRecordLog } from '@/api/modules/cronjob';
import { dateFormat } from '@/utils/util';
import LogFile from '@/components/log/file/index.vue';
import i18n from '@/lang';
import { ElMessageBox } from 'element-plus';
import { MsgSuccess } from '@/utils/message';
import { listDbItems } from '@/api/modules/database';
import { listAppInstalled } from '@/api/modules/app';
import { shortcuts } from '@/utils/shortcuts';
import { hasBackup } from '../helper';

const loading = ref();
const refresh = ref(false);
const hasRecords = ref();

let timer: NodeJS.Timer | null = null;

const mymirror = ref();
const scrollerElement = ref<HTMLElement | null>(null);

interface DialogProps {
    rowData: Cronjob.CronjobInfo;
}
const recordShow = ref(false);
const dialogData = ref();
const records = ref<Array<Cronjob.Record>>([]);
const currentRecord = ref<Cronjob.Record>();
const currentRecordDetail = ref<string>('');

const open = ref();
const delLoading = ref();
const cleanData = ref();
const cleanRemoteData = ref();

const acceptParams = async (params: DialogProps): Promise<void> => {
    let itemSize = Number(localStorage.getItem(searchInfo.cacheSizeKey));
    if (itemSize) {
        searchInfo.pageSize = itemSize;
    }

    recordShow.value = true;
    dialogData.value = params;
    if (dialogData.value.rowData.type === 'database') {
        const data = await listDbItems('mysql,mariadb,postgresql');
        let itemDBs = data.data || [];
        for (const item of itemDBs) {
            if (item.id == dialogData.value.rowData.dbName) {
                dialogData.value.rowData.dbName = item.database + ' [' + item.name + ']';
                break;
            }
        }
    }
    if (dialogData.value.rowData.type === 'app') {
        const res = await listAppInstalled();
        let itemApps = res.data || [];
        for (const item of itemApps) {
            if (item.id == dialogData.value.rowData.appID) {
                dialogData.value.rowData.appID = item.key + ' [' + item.name + ']';
                break;
            }
        }
    }
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

const timeRangeLoad = ref<[Date, Date]>([
    new Date(new Date(new Date().getTime() - 3600 * 1000 * 24 * 7).setHours(0, 0, 0, 0)),
    new Date(new Date().setHours(23, 59, 59, 999)),
]);
const searchInfo = reactive({
    cacheSizeKey: 'cronjob-record-page-size',
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

const forDetail = async (row: Cronjob.Record) => {
    currentRecord.value = row;
    loadRecord(row);
};
const loadRecord = async (row: Cronjob.Record) => {
    currentRecord.value = row;
    if (row.records) {
        const res = await getRecordLog(row.id);
        let log = res.data.replace(/\x1B\[[0-?]*[ -/]*[@-~]/g, '');
        if (currentRecordDetail.value === log) {
            return;
        }
        currentRecordDetail.value = log;
        nextTick(() => {
            initLog();
            scrollerElement.value.scrollTop = scrollerElement.value.scrollHeight;
        });
    }
};

const onClean = async () => {
    if (!hasBackup(dialogData.value.rowData.type)) {
        ElMessageBox.confirm(i18n.global.t('commons.msg.clean'), i18n.global.t('commons.button.clean'), {
            confirmButtonText: i18n.global.t('commons.button.confirm'),
            cancelButtonText: i18n.global.t('commons.button.cancel'),
            type: 'warning',
        }).then(async () => {
            await cleanRecords(dialogData.value.rowData.id, cleanData.value, cleanRemoteData.value)
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
        open.value = true;
    }
};

const cleanRecord = async () => {
    delLoading.value = true;
    await cleanRecords(dialogData.value.rowData.id, cleanData.value, false)
        .then(() => {
            delLoading.value = false;
            open.value = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .catch(() => {
            delLoading.value = false;
        });
};

const initLog = () => {
    if (mymirror.value && scrollerElement.value == undefined) {
        scrollerElement.value = mymirror.value.$el as HTMLElement;
        let hljsDom = scrollerElement.value.querySelector('.hljs') as HTMLElement;
        hljsDom.style['min-height'] = '500px';
    }
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

@media only screen and (max-width: 1400px) {
    .mainClass {
        overflow: auto;
    }
    .mainRowClass {
        min-width: 1200px;
    }
}
.editor-main {
    height: calc(100vh - 488px);
    width: 100%;
    margin-top: 5px;
    overflow-x: auto;
}
</style>
