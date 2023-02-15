<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('logs.operation')">
            <template #toolbar>
                <el-row>
                    <el-col :span="16">
                        <el-button type="primary" plain @click="onClean()">
                            {{ $t('logs.deleteLogs') }}
                        </el-button>
                    </el-col>
                    <el-col :span="8">
                        <TableSetting @search="search()" />
                        <div class="search-button">
                            <el-input
                                v-model="searchName"
                                clearable
                                @clear="search()"
                                suffix-icon="Search"
                                @keyup.enter="search()"
                                @blur="search()"
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
                    <el-option :label="$t('logs.detail.databases')" value="databases"></el-option>
                    <el-option :label="$t('logs.detail.containers')" value="containers"></el-option>
                    <el-option :label="$t('logs.detail.cronjobs')" value="cronjobs"></el-option>
                    <el-option :label="$t('logs.detail.files')" value="files"></el-option>
                    <el-option :label="$t('logs.detail.hosts')" value="hosts"></el-option>
                    <el-option :label="$t('logs.detail.logs')" value="logs"></el-option>
                    <el-option :label="$t('logs.detail.settings')" value="settings"></el-option>
                </el-select>
                <el-select v-model="searchStatus" @change="search()" clearable style="margin-left: 10px">
                    <template #prefix>{{ $t('commons.table.status') }}</template>
                    <el-option :label="$t('commons.table.all')" value=""></el-option>
                    <el-option :label="$t('commons.status.success')" value="Success"></el-option>
                    <el-option :label="$t('commons.status.failed')" value="Failed"></el-option>
                </el-select>
            </template>
            <template #main>
                <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search">
                    <el-table-column :label="$t('logs.resource')" prop="group" fix>
                        <template #default="{ row }">
                            {{ $t('logs.detail.' + row.source) }}
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('logs.operate')" min-width="150px" prop="detailZH">
                        <template #default="{ row }">
                            <span v-if="globalStore.language === 'zh'">{{ row.detailZH }}</span>
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
import ComplexTable from '@/components/complex-table/index.vue';
import TableSetting from '@/components/table-setting/index.vue';
import ConfirmDialog from '@/components/confirm-dialog/index.vue';
import { dateFormat } from '@/utils/util';
import LayoutContent from '@/layout/layout-content.vue';
import { cleanLogs, getOperationLogs } from '@/api/modules/log';
import { onMounted, reactive, ref } from '@vue/runtime-core';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';
import { GlobalStore } from '@/store';

const loading = ref();
const data = ref();
const confirmDialogRef = ref();
const paginationConfig = reactive({
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
            data.value = res.data.items;
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

const onSubmitClean = async () => {
    await cleanLogs({ logType: 'operation' });
    search();
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
};

onMounted(() => {
    search();
});
</script>
