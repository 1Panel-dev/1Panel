<template>
    <div>
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
                        <div v-for="(item, index) of row.labels" :key="index">
                            <el-tooltip class="item" :content="item" placement="top">
                                <el-tag>{{ item }}</el-tag>
                            </el-tooltip>
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

        <el-dialog v-model="detailVisiable" :destroy-on-close="true" :close-on-click-modal="false" width="70%">
            <template #header>
                <div class="card-header">
                    <span>{{ $t('commons.button.view') }}</span>
                </div>
            </template>
            <codemirror
                :autofocus="true"
                placeholder="None data"
                :indent-with-tab="true"
                :tabSize="4"
                style="max-height: 500px"
                :lineWrapping="true"
                :matchBrackets="true"
                theme="cobalt"
                :styleActiveLine="true"
                :extensions="extensions"
                v-model="detailInfo"
                :readOnly="true"
            />
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="detailVisiable = false">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </el-dialog>

        <CreateDialog @search="search" ref="dialogCreateRef" />
    </div>
</template>

<script lang="ts" setup>
import ComplexTable from '@/components/complex-table/index.vue';
import CreateDialog from '@/views/container/network/create/index.vue';
import { Codemirror } from 'vue-codemirror';
import { javascript } from '@codemirror/lang-javascript';
import { oneDark } from '@codemirror/theme-one-dark';
import { reactive, onMounted, ref } from 'vue';
import { dateFromat } from '@/utils/util';
import { deleteNetwork, searchNetwork, inspect } from '@/api/modules/container';
import { Container } from '@/api/interface/container';
import i18n from '@/lang';
import { useDeleteData } from '@/hooks/use-delete-data';

const detailVisiable = ref<boolean>(false);
const detailInfo = ref();
const extensions = [javascript(), oneDark];

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
    detailVisiable.value = true;
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
