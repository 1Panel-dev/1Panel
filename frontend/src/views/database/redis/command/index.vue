<template>
    <div>
        <el-drawer
            v-model="drawerVisible"
            :destroy-on-close="true"
            @close="handleClose"
            :close-on-click-modal="false"
            :close-on-press-escape="false"
            size="50%"
        >
            <template #header>
                <DrawerHeader :header="$t('database.redisQuickCmd')" :back="handleClose" />
            </template>
            <el-button type="primary" @click="handleCmdAdd()">
                {{ $t('commons.button.add') }}
            </el-button>
            <el-button @click="batchDelete(null)">
                {{ $t('commons.button.delete') }}
            </el-button>
            <el-table :data="data" class="mt-5" @selection-change="handleSelectionChange">
                <el-table-column type="selection" fix />
                <el-table-column :label="$t('commons.table.name')" min-width="50">
                    <template #default="{ row }">
                        <el-input v-if="row.isOn" v-model="row.name" />
                        <span v-else>{{ row.name }}</span>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('terminal.quickCommand')" min-width="120">
                    <template #default="{ row }">
                        <el-input v-if="row.isOn" v-model="row.command" />
                        <span v-else>{{ row.name }}</span>
                    </template>
                </el-table-column>
                <el-table-column min-width="40">
                    <template #default="scope">
                        <el-button link type="primary" @click="handleCmdCreate(scope.row)">
                            {{ $t('commons.button.save') }}
                        </el-button>
                        <el-button link type="primary" @click="handleCmdDelete(scope.$index)">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </template>
                </el-table-column>
            </el-table>
            <template #footer>
                <span class="dialog-footer">
                    <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
                </span>
            </template>
        </el-drawer>

        <OpDialog ref="opRef" @search="search" />
    </div>
</template>

<script setup lang="ts">
import { Command } from '@/api/interface/command';
import { addRedisCommand, deleteRedisCommand, getRedisCommandPage } from '@/api/modules/host';
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import DrawerHeader from '@/components/drawer-header/index.vue';
import { MsgSuccess } from '@/utils/message';

const drawerVisible = ref();
const loading = ref();
const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const opRef = ref();

const emit = defineEmits(['reload']);

const acceptParams = () => {
    drawerVisible.value = true;
    search();
};

const handleSelectionChange = (val: any[]) => {
    selects.value = val;
};

const handleCmdAdd = () => {
    let item = {
        name: '',
        command: '',
        isOn: true,
    };
    data.value.push(item);
};
const handleCmdDelete = (index: number) => {
    batchDelete(data.value[index]);
    data.value.splice(index, 1);
};

const handleCmdCreate = async (row: any) => {
    loading.value = true;
    await addRedisCommand(row)
        .then(() => {
            loading.value = false;
            row.isOn = false;
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
    return;
};

const batchDelete = async (row: Command.RedisCommand | null) => {
    let names = [];
    let ids = [];
    if (row) {
        ids = [row.id];
        names = [row.name];
    } else {
        selects.value.forEach((item: Command.CommandInfo) => {
            ids.push(item.id);
            names.push(item.name);
        });
    }
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('terminal.quickCommand'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: deleteRedisCommand,
        params: { ids: ids },
    });
};

const search = async () => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        info: '',
    };
    loading.value = true;
    await getRedisCommandPage(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

const handleClose = () => {
    drawerVisible.value = false;
    emit('reload');
};

defineExpose({
    acceptParams,
});
</script>
