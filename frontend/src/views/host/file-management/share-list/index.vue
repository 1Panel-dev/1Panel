<template>
    <DrawerPro v-model="open" :header="$t('file.shareList')" @close="handleClose" size="large">
        <template #content>
            <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search">
                <el-table-column :label="$t('file.path')" show-overflow-tooltip prop="path" min-width="250">
                    <template #default="{ row }">
                        <el-tooltip class="box-item" effect="dark" :content="row.path" placement="top">
                            <span class="table-link text-ellipsis" @click="viewDetail(row)" type="primary">
                                <svg-icon className="table-icon" iconName="p-file-normal"></svg-icon>
                                {{ row.fileName }}
                            </span>
                        </el-tooltip>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('file.shareExpiresAt')" min-width="140">
                    <template #default="{ row }">
                        {{ formatExpire(row) }}
                    </template>
                </el-table-column>
                <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fix />
            </ComplexTable>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { removeFileShare, searchFileShare } from '@/api/modules/files';
import { File } from '@/api/interface/file';
import i18n from '@/lang';
import { dateFormat as formatDateTime } from '@/utils/date';
import { computed, reactive, ref } from 'vue';

const paginationConfig = reactive({
    cacheSizeKey: 'share-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('share-page-size')) || 20,
    total: 0,
});
const req = reactive({
    page: 1,
    pageSize: 20,
});
const open = ref(false);
const data = ref<File.FileShareInfo[]>([]);
const permanentText = computed(() => String(i18n.global.t('website.ever')));
const emit = defineEmits(['close', 'detail']);

const handleClose = () => {
    open.value = false;
    emit('close', false);
};

const formatExpire = (row: File.FileShareInfo) => {
    if (row.permanent || row.expiresAt === 0) {
        return permanentText.value;
    }
    return formatDateTime(null, null, row.expiresAt * 1000);
};

const acceptParams = async () => {
    await search();
};

const search = async () => {
    req.page = paginationConfig.currentPage;
    req.pageSize = paginationConfig.pageSize;
    const res = await searchFileShare(req);
    data.value = res.data.items;
    paginationConfig.total = res.data.total;
    open.value = true;
};

const closeShare = async (path: string) => {
    await removeFileShare(path);
    emit('close', true);
    await search();
};

const viewDetail = (row: File.FileShareInfo) => {
    emit('detail', row.path);
};

const buttons = [
    {
        label: i18n.global.t('file.shareDetail'),
        click: (row: File.FileShareInfo) => {
            viewDetail(row);
        },
    },
    {
        label: i18n.global.t('file.shareClose'),
        permission: true,
        nodeAdmin: true,
        click: (row: File.FileShareInfo) => {
            closeShare(row.path);
        },
    },
];

defineExpose({ acceptParams });
</script>
