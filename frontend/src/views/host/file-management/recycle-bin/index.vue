<template>
    <el-drawer
        v-model="open"
        :before-close="handleClose"
        :close-on-click-modal="false"
        :close-on-press-escape="false"
        size="50%"
    >
        <template #header>
            <DrawerHeader :header="$t('file.recycleBin')" :back="handleClose" />
        </template>
        <div class="flex space-x-4">
            <el-button @click="clear" type="primary" :disabled="data == null || data.length == 0">
                {{ $t('file.clearRecycleBin') }}
            </el-button>
            <el-button @click="patchDelete" :disabled="data == null || selects.length == 0">
                {{ $t('commons.button.delete') }}
            </el-button>
            <el-button @click="patchReduce" :disabled="data == null || selects.length == 0">
                {{ $t('file.reduce') }}
            </el-button>
            <el-form-item :label="$t('file.fileRecycleBin')">
                <el-switch v-model="status" active-value="enable" inactive-value="disable" @change="changeStatus" />
            </el-form-item>
        </div>
        <ComplexTable
            :pagination-config="paginationConfig"
            v-model:selects="selects"
            :data="data"
            @search="search"
            class="mt-5"
        >
            <el-table-column type="selection" fix />
            <el-table-column prop="name" :label="$t('commons.table.name')" show-overflow-tooltip>
                <template #default="{ row }">
                    <span class="text-ellipsis" type="primary">
                        <svg-icon v-if="row.isDir" className="table-icon" iconName="p-file-folder"></svg-icon>
                        <svg-icon v-else className="table-icon" iconName="p-file-normal"></svg-icon>
                        {{ row.name }}
                    </span>
                </template>
            </el-table-column>

            <el-table-column :label="$t('file.sourcePath')" show-overflow-tooltip prop="sourcePath"></el-table-column>
            <el-table-column :label="$t('file.size')" prop="size" max-width="50">
                <template #default="{ row }">
                    {{ getFileSize(row.size) }}
                </template>
            </el-table-column>
            <el-table-column
                :label="$t('file.deleteTime')"
                prop="deleteTime"
                :formatter="dateFormat"
                show-overflow-tooltip
                sortable
            ></el-table-column>
            <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fix />
        </ComplexTable>
        <Delete ref="deleteRef" @close="search" />
        <Reduce ref="reduceRef" @close="search" />
    </el-drawer>
</template>

<script lang="ts" setup>
import { GetRecycleStatus, clearRecycle, getRecycleList, reduceFile } from '@/api/modules/files';
import { reactive, ref } from 'vue';
import { dateFormat, computeSize } from '@/utils/util';
import i18n from '@/lang';
import Delete from './delete/index.vue';
import Reduce from './reduce/index.vue';
import { updateSetting } from '@/api/modules/setting';
import { MsgSuccess } from '@/utils/message';

const open = ref(false);
const req = reactive({
    page: 1,
    pageSize: 20,
});
const data = ref([]);
const em = defineEmits(['close']);
const selects = ref([]);
const loading = ref(false);
const files = ref([]);
const status = ref('enable');

const paginationConfig = reactive({
    cacheSizeKey: 'recycle-page-size',
    currentPage: 1,
    pageSize: 20,
    total: 0,
});

const deleteRef = ref();
const reduceRef = ref();

const handleClose = () => {
    open.value = false;
    em('close', false);
};

const getFileSize = (size: number) => {
    return computeSize(size);
};

const acceptParams = () => {
    search();
    getStatus();
};

const getStatus = async () => {
    try {
        const res = await GetRecycleStatus();
        status.value = res.data;
    } catch (error) {}
};

const changeStatus = async () => {
    try {
        loading.value = true;
        await updateSetting({ key: 'FileRecycleBin', value: status.value });
        MsgSuccess(i18n.global.t('file.fileRecycleBinMsg', [i18n.global.t('commons.button.' + status.value)]));
        loading.value = false;
    } catch (error) {}
};

const search = async () => {
    try {
        req.page = paginationConfig.currentPage;
        req.pageSize = paginationConfig.pageSize;
        const res = await getRecycleList(req);
        data.value = res.data.items;
        paginationConfig.total = res.data.total;
        open.value = true;
    } catch (error) {}
};

const singleDel = (row: any) => {
    files.value = [];
    files.value.push(row);
    deleteRef.value.acceptParams(files.value);
};

const patchDelete = () => {
    files.value = selects.value;
    deleteRef.value.acceptParams(files.value);
};

const patchReduce = () => {
    files.value = selects.value;
    reduceRef.value.acceptParams(files.value);
};

const rdFile = async (row: any) => {
    ElMessageBox.confirm(i18n.global.t('file.reduceHelper'), i18n.global.t('file.reduce'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        try {
            loading.value = true;
            await reduceFile({ from: row.from, rName: row.rName, name: row.name });
            loading.value = false;
            search();
        } catch (error) {}
    });
};

const clear = async () => {
    ElMessageBox.confirm(i18n.global.t('commons.msg.delete'), i18n.global.t('file.clearRecycleBinHelper'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        try {
            loading.value = true;
            await clearRecycle();
            loading.value = false;
            search();
        } catch (error) {}
    });
};

const buttons = [
    {
        label: i18n.global.t('file.reduce'),
        click: rdFile,
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: singleDel,
    },
];

defineExpose({ acceptParams });
</script>
