<template>
    <div>
        <Submenu activeName="network" />
        <el-card style="margin-top: 20px">
            <ComplexTable :pagination-config="paginationConfig" v-model:selects="selects" :data="data" @search="search">
                <template #toolbar>
                    <el-button style="margin-left: 10px" @click="onCreate()">
                        {{ $t('commons.button.create') }}
                    </el-button>
                    <el-button type="danger" plain :disabled="selects.length === 0" @click="batchDelete(null)">
                        {{ $t('commons.button.delete') }}
                    </el-button>
                </template>
                <el-table-column type="selection" fix />
                <el-table-column :label="$t('commons.table.name')" show-overflow-tooltip min-width="80" prop="name" fix>
                    <template #default="{ row }">
                        <el-link @click="onInspect(row.id)" type="primary">{{ row.name }}</el-link>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('container.driver')" show-overflow-tooltip min-width="40" prop="driver" />
                <el-table-column :label="$t('container.attachable')" min-width="40" prop="attachable" fix>
                    <template #default="{ row }">
                        <el-icon color="green" v-if="row.attachable"><Select /></el-icon>
                        <el-icon color="red" v-if="!row.attachable"><CloseBold /></el-icon>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('container.subnet')" min-width="80" prop="subnet" fix />
                <el-table-column :label="$t('container.gateway')" min-width="80" prop="gateway" fix />
                <el-table-column :label="$t('container.tag')" min-width="140" fix>
                    <template #default="{ row }">
                        <div v-for="(item, index) in row.labels" :key="index">
                            <div v-if="row.expand || (!row.expand && index < 3)">
                                <el-tag>{{ item }}</el-tag>
                            </div>
                        </div>
                        <div v-if="!row.expand && row.labels.length > 3">
                            <el-button type="primary" link @click="row.expand = true">
                                {{ $t('commons.button.expand') }}...
                            </el-button>
                        </div>
                    </template>
                </el-table-column>
                <el-table-column
                    prop="createdAt"
                    min-width="90"
                    :label="$t('commons.table.date')"
                    :formatter="dateFromat"
                />
                <fu-table-operations :buttons="buttons" :label="$t('commons.table.operate')" fix />
            </ComplexTable>
        </el-card>

        <CodemirrorDialog ref="codemirror" />
        <CreateDialog @search="search" ref="dialogCreateRef" />
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import CreateDialog from '@/views/container/network/create/index.vue';
import CodemirrorDialog from '@/components/codemirror-dialog/codemirror.vue';
import Submenu from '@/views/container/index.vue';
import { reactive, onMounted, ref } from 'vue';
import { dateFromat } from '@/utils/util';
import { deleteNetwork, searchNetwork, inspect } from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import i18n from '@/lang';
import { useDeleteData } from '@/hooks/use-delete-data';

const detailInfo = ref();
const codemirror = ref();

const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    page: 1,
    pageSize: 10,
    total: 0,
});

const dialogCreateRef = ref<DialogExpose>();

interface DialogExpose {
    acceptParams: () => void;
}
const onCreate = async () => {
    dialogCreateRef.value!.acceptParams();
};

const search = async () => {
    const params = {
        page: paginationConfig.page,
        pageSize: paginationConfig.pageSize,
    };
    await searchNetwork(params).then((res) => {
        if (res.data) {
            data.value = res.data.items;
        }
        paginationConfig.total = res.data.total;
    });
};

const batchDelete = async (row: Container.NetworkInfo | null) => {
    let ids: Array<string> = [];
    if (row === null) {
        selects.value.forEach((item: Container.NetworkInfo) => {
            ids.push(item.id);
        });
    } else {
        ids.push(row.id);
    }
    await useDeleteData(deleteNetwork, { ids: ids }, 'commons.msg.delete', true);
    search();
};

const onInspect = async (id: string) => {
    const res = await inspect({ id: id, type: 'network' });
    detailInfo.value = JSON.stringify(JSON.parse(res.data), null, 2);
    let param = {
        header: i18n.global.t('commons.button.view'),
        detailInfo: detailInfo.value,
    };
    codemirror.value!.acceptParams(param);
};

const buttons = [
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Container.NetworkInfo) => {
            batchDelete(row);
        },
    },
];

onMounted(() => {
    search();
});
</script>
