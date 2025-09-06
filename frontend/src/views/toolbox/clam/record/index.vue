<template>
    <div v-if="recordShow" v-loading="loading">
        <div class="app-status p-mt-20">
            <el-card>
                <div class="flex w-full flex-col gap-4 md:flex-row">
                    <div class="flex flex-wrap gap-4 ml-3">
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
                            class="float-left"
                            effect="dark"
                            type="success"
                        >
                            {{ $t('toolbox.clam.scanDir') }}: {{ dialogData.rowData.path }}
                        </el-tag>

                        <span class="mt-0.5">
                            <el-button type="primary" @click="onHandle(dialogData.rowData)" link>
                                {{ $t('commons.button.handle') }}
                            </el-button>
                            <el-divider direction="vertical" />
                            <el-button :disabled="!hasRecords" type="primary" @click="onClean" link>
                                {{ $t('commons.button.clean') }}
                            </el-button>
                        </span>
                    </div>
                </div>
            </el-card>
        </div>

        <LayoutContent :title="$t('cronjob.record')" :reload="true">
            <template #rightToolBar>
                <el-date-picker
                    class="mr-2.5"
                    @change="search(true)"
                    v-model="timeRangeLoad"
                    type="datetimerange"
                    range-separator="-"
                    :start-placeholder="$t('commons.search.timeStart')"
                    :end-placeholder="$t('commons.search.timeEnd')"
                    :shortcuts="shortcuts"
                ></el-date-picker>
                <el-select @change="search(true)" v-model="searchInfo.status" class="p-w-200">
                    <template #prefix>{{ $t('commons.table.status') }}</template>
                    <el-option :label="$t('commons.table.all')" value="" />
                    <el-option :label="$t('commons.status.done')" value="Done" />
                    <el-option :label="$t('commons.status.waiting')" value="Waiting" />
                    <el-option :label="$t('commons.status.failed')" value="Failed" />
                </el-select>
                <TableRefresh @search="search(false)" />
            </template>
            <template #main>
                <div class="mainClass">
                    <el-row :gutter="20" v-show="hasRecords" class="mainRowClass row-box">
                        <el-col :span="7">
                            <el-card class="el-card">
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
                                                <div class="mt-0.5 float-left">
                                                    <span>
                                                        {{ dateFormat(0, 0, row.startTime) }}
                                                    </span>
                                                </div>
                                                <el-button
                                                    class="mt-0.5 float-right"
                                                    type="danger"
                                                    icon="Warning"
                                                    link
                                                    v-if="row.infectedFiles && row.infectedFiles !== '0'"
                                                />
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
                                        :pager-count="5"
                                        :page-sizes="[5, 10, 20, 50, 100, 200, 500, 1000]"
                                        small
                                        layout="total, sizes, prev, pager, next"
                                        :total="searchInfo.recordTotal"
                                    />
                                </div>
                            </el-card>
                        </el-col>
                        <el-col :span="17">
                            <el-card class="el-card">
                                <el-form label-position="top" :v-key="refresh">
                                    <el-row>
                                        <el-form-item class="descriptionWide">
                                            <template #label>
                                                <span class="status-label">{{ $t('commons.table.interval') }}</span>
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
                                                {{
                                                    currentRecord?.status === 'Done'
                                                        ? currentRecord?.infectedFiles
                                                        : '-'
                                                }}
                                            </span>
                                            <div class="count" v-else>
                                                <span @click="toFolder(currentRecord)">
                                                    {{
                                                        currentRecord?.status === 'Done'
                                                            ? currentRecord?.infectedFiles
                                                            : '-'
                                                    }}
                                                </span>
                                            </div>
                                        </el-form-item>
                                    </el-row>
                                    <el-row v-if="currentRecord?.taskID && currentRecord?.taskID != ''">
                                        <LogFile
                                            :defaultButton="true"
                                            class="w-full"
                                            :key="currentRecord?.taskID"
                                            @stop-reading="search(false)"
                                            :heightDiff="430"
                                            :config="{
                                                type: 'task',
                                                colorMode: 'task',
                                                taskID: currentRecord?.taskID,
                                                tail: true,
                                            }"
                                        />
                                    </el-row>
                                </el-form>
                            </el-card>
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
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { ElMessageBox } from 'element-plus';
import { MsgSuccess } from '@/utils/message';
import { shortcuts } from '@/utils/shortcuts';
import { dateFormat, dateFormatForName } from '@/utils/util';
import { Toolbox } from '@/api/interface/toolbox';
import LogFile from '@/components/log/file/index.vue';
import { cleanClamRecord, handleClamScan, searchClamRecord } from '@/api/modules/toolbox';
import { routerToFileWithPath } from '@/utils/router';

const loading = ref();
const refresh = ref(false);
const hasRecords = ref();

const recordShow = ref(false);
interface DialogProps {
    rowData: Toolbox.ClamInfo;
}
const dialogData = ref();
const records = ref<Array<Toolbox.ClamRecord>>([]);
const currentRecord = ref<Toolbox.ClamRecord>();

const acceptParams = async (params: DialogProps): Promise<void> => {
    let itemSize = Number(localStorage.getItem(searchInfo.cacheSizeKey));
    if (itemSize) {
        searchInfo.pageSize = itemSize;
    }

    recordShow.value = true;
    dialogData.value = params;
    search(true);
};

const handleSizeChange = (val: number) => {
    searchInfo.pageSize = val;
    localStorage.setItem(searchInfo.cacheSizeKey, val + '');
    search(true);
};
const handleCurrentChange = (val: number) => {
    searchInfo.page = val;
    search(false);
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
    pageSize: 10,
    status: '',
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
            search(true);
        })
        .catch(() => {
            loading.value = false;
        });
};
const toFolder = async (row: any) => {
    let folder =
        dialogData.value.rowData!.infectedDir +
        '/1panel-infected/' +
        dialogData.value.rowData!.name +
        '/' +
        dateFormatForName(row.startTime);
    routerToFileWithPath(folder);
};

const search = async (changeToLatest: boolean) => {
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
        status: searchInfo.status,
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
    if (changeToLatest) {
        currentRecord.value = records.value[0];
        return;
    }
    for (const item of records.value) {
        if (item.id === currentRecord.value.id) {
            currentRecord.value = item;
            break;
        }
    }
};

const forDetail = async (row: Toolbox.ClamRecord) => {
    currentRecord.value = row;
};

const onClean = async () => {
    ElMessageBox.confirm(i18n.global.t('commons.msg.clean'), i18n.global.t('commons.button.delete'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'warning',
    }).then(async () => {
        loading.value = true;
        cleanClamRecord(dialogData.value.rowData.id)
            .then(() => {
                loading.value = false;
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                search(false);
            })
            .catch(() => {
                loading.value = false;
            });
    });
};

defineExpose({
    acceptParams,
});
</script>

<style lang="scss" scoped>
.infinite-list {
    height: calc(100vh - 320px);
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
