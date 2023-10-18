<template>
    <el-drawer v-model="open" :before-close="handleClose" :close-on-click-modal="false" size="50%">
        <template #header>
            <DrawerHeader :header="$t('file.recycleBin')" :back="handleClose" />
        </template>
        <el-button @click="clear" type="primary" :disabled="data == null || data.length == 0">
            {{ $t('file.clearRecycleBin') }}
        </el-button>
        <ComplexTable
            :pagination-config="paginationConfig"
            v-model:selects="selects"
            :data="data"
            @search="search"
            class="mt-5"
        >
            <el-table-column type="selection" fix />
            <el-table-column
                :label="$t('commons.table.name')"
                min-width="100"
                fix
                show-overflow-tooltip
                prop="name"
            ></el-table-column>
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
            ></el-table-column>
            <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fix />
        </ComplexTable>
    </el-drawer>
</template>

<script lang="ts" setup>
import { DeleteFile, clearRecycle, getRecycleList, reduceFile } from '@/api/modules/files';
import { reactive, ref } from 'vue';
import { dateFormat, computeSize } from '@/utils/util';
import i18n from '@/lang';
import { MsgSuccess } from '@/utils/message';

const open = ref(false);
const req = reactive({
    page: 1,
    pageSize: 100,
});
const data = ref([]);
const em = defineEmits(['close']);
const selects = ref([]);
const loading = ref(false);
const files = ref([]);

const paginationConfig = reactive({
    cacheSizeKey: 'recycle-page-size',
    currentPage: 1,
    pageSize: 100,
    total: 0,
});

const handleClose = () => {
    open.value = false;
    em('close', false);
};

const getFileSize = (size: number) => {
    return computeSize(size);
};

const acceptParams = () => {
    search();
};

const search = async () => {
    try {
        const res = await getRecycleList(req);
        data.value = res.data.items;
        paginationConfig.total = res.data.total;
        open.value = true;
    } catch (error) {}
};

const singleDel = (row: any) => {
    ElMessageBox.confirm(i18n.global.t('commons.msg.delete'), i18n.global.t('commons.button.delete'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        files.value = [];
        files.value.push(row);
        deleteFile();
    });
};

const deleteFile = async () => {
    const pros = [];
    for (const s of files.value) {
        pros.push(DeleteFile({ path: s.from + '/' + s.rName, isDir: s.isDir, forceDelete: true }));
    }
    loading.value = true;
    Promise.all(pros)
        .then(() => {
            MsgSuccess(i18n.global.t('commons.msg.deleteSuccess'));
            search();
        })
        .finally(() => {
            loading.value = false;
        });
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
