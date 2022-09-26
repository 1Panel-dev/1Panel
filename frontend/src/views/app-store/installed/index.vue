<template>
    <ComplexTable :pagination-config="paginationConfig" :data="data" @search="search" v-loading="loading">
        <el-table-column :label="$t('app.name')" prop="name"></el-table-column>
        <!-- <el-table-column :label="$t('app.description')" prop="description"></el-table-column> -->
        <el-table-column :label="$t('app.appName')" prop="appName"></el-table-column>
        <el-table-column :label="$t('app.version')" prop="version"></el-table-column>
        <el-table-column :label="$t('app.container')">
            <template #default="{ row }">
                {{ row.ready / row.total }}
            </template>
        </el-table-column>
        <el-table-column :label="$t('app.status')">
            <template #default="{ row }">
                <el-popover
                    v-if="row.status === 'Error'"
                    placement="top-start"
                    :width="400"
                    trigger="hover"
                    :content="row.message"
                >
                    <template #reference>
                        <el-tag type="error">{{ row.status }}</el-tag>
                    </template>
                </el-popover>
                <el-tag v-else>{{ row.status }}</el-tag>
            </template>
        </el-table-column>
        <el-table-column
            prop="createdAt"
            :label="$t('commons.table.date')"
            :formatter="dateFromat"
            show-overflow-tooltip
        />
        <fu-table-operations :ellipsis="10" :buttons="buttons" :label="$t('commons.table.operate')" fixed="right" fix />
    </ComplexTable>
</template>

<script lang="ts" setup>
import { GetAppInstalled, InstalledOp } from '@/api/modules/app';
import { onMounted, reactive, ref } from 'vue';
import ComplexTable from '@/components/complex-table/index.vue';
import { dateFromat } from '@/utils/util';
import i18n from '@/lang';
import { ElMessage, ElMessageBox } from 'element-plus';

let data = ref<any>();
let loading = ref(false);
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 20,
    total: 0,
});

const search = () => {
    const req = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
    };

    GetAppInstalled(req).then((res) => {
        data.value = res.data.items;
        paginationConfig.total = res.data.total;
    });
};

const operate = async (row: any, op: string) => {
    const req = {
        installId: row.id,
        operate: op,
    };

    ElMessageBox.confirm(i18n.global.t(`${'app.' + op}`) + '?', i18n.global.t('commons.msg.operate'), {
        confirmButtonText: i18n.global.t('commons.button.confirm'),
        cancelButtonText: i18n.global.t('commons.button.cancel'),
        type: 'warning',
        draggable: true,
    }).then(async () => {
        loading.value = true;
        InstalledOp(req)
            .then(() => {
                ElMessage.success(i18n.global.t('commons.msg.operationSuccess'));
                search();
            })
            .finally(() => {
                loading.value = false;
            });
    });
};

const buttons = [
    {
        label: i18n.global.t('app.restart'),
        click: (row: any) => {
            operate(row, 'restart');
        },
    },
    {
        label: i18n.global.t('app.up'),
        click: (row: any) => {
            operate(row, 'up');
        },
    },
    {
        label: i18n.global.t('app.down'),
        click: (row: any) => {
            operate(row, 'down');
        },
    },
    {
        label: i18n.global.t('app.delete'),
        click: (row: any) => {
            operate(row, 'delete');
        },
    },
];

onMounted(() => {
    search();
});
</script>

<style lang="scss">
.i-card {
    height: 60px;
    cursor: pointer;
    .content {
        .image {
            width: auto;
            height: auto;
        }
    }
}
.i-card:hover {
    border: 1px solid;
    border-color: $primary-color;
    z-index: 1;
}
</style>
