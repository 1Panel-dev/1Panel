<template>
    <DrawerPro v-model="open" :header="$t('file.favorite')" @close="handleClose" size="large">
        <template #content>
            <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search">
                <el-table-column :label="$t('file.path')" show-overflow-tooltip prop="path">
                    <template #default="{ row }">
                        <el-tooltip class="box-item" effect="dark" :content="row.path" placement="top">
                            <span class="table-link text-ellipsis" @click="toFavorite(row)" type="primary">
                                <svg-icon v-if="row.isDir" className="table-icon" iconName="p-file-folder"></svg-icon>
                                <svg-icon v-else className="table-icon" iconName="p-file-normal"></svg-icon>
                                {{ row.name }}
                            </span>
                        </el-tooltip>
                    </template>
                </el-table-column>
                <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fix />
            </ComplexTable>
        </template>
    </DrawerPro>
</template>

<script setup lang="ts">
import { searchFavorite, removeFavorite } from '@/api/modules/files';
import i18n from '@/lang';
import { reactive, ref } from 'vue';
import { File } from '@/api/interface/file';

const paginationConfig = reactive({
    cacheSizeKey: 'favorite-page-size',
    currentPage: 1,
    pageSize: Number(localStorage.getItem('favorite-page-size')) || 20,
    total: 0,
});
const req = reactive({
    page: 1,
    pageSize: 20,
});
const open = ref(false);
const data = ref([]);
const em = defineEmits(['close', 'jump', 'to-favorite']);

const handleClose = () => {
    open.value = false;
    em('close', false);
};

const toFavorite = (row: File.Favorite) => {
    open.value = false;
    em('to-favorite', row);
};
const openDir = (url: string) => {
    open.value = false;
    em('jump', url);
};

const acceptParams = () => {
    search();
};

const search = async () => {
    try {
        req.page = paginationConfig.currentPage;
        req.pageSize = paginationConfig.pageSize;
        const res = await searchFavorite(req);
        data.value = res.data.items;
        paginationConfig.total = res.data.total;
        open.value = true;
    } catch (error) {}
};

const singleDel = async (id: number) => {
    ElMessageBox.confirm(i18n.global.t('file.removeFavorite'), i18n.global.t('commons.msg.remove'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
    }).then(async () => {
        try {
            await removeFavorite(id);
            await search();
        } catch (error) {}
    });
};

const buttons = [
    {
        label: i18n.global.t('commons.button.open'),
        click: (row: any) => {
            openDir(row.path);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: any) => {
            singleDel(row.id);
        },
    },
];

defineExpose({ acceptParams });
</script>
