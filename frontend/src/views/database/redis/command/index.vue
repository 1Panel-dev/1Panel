<template>
    <DrawerPro v-model="drawerVisible" :header="$t('database.redisQuickCmd')" @close="handleClose" size="large">
        <template #content>
            <el-button type="primary" @click="handleCmdAdd()">
                {{ $t('commons.button.add') }}
            </el-button>
            <el-button @click="batchDelete(null)">
                {{ $t('commons.button.delete') }}
            </el-button>
            <el-table :data="data" class="mt-5" @selection-change="handleSelectionChange">
                <el-table-column type="selection" fix />
                <el-table-column :label="$t('commons.table.name')" min-width="50" show-overflow-tooltip>
                    <template #default="{ row }">
                        <el-input v-if="row.lineStatus === 'create' || row.lineStatus === 'edit'" v-model="row.name" />
                        <span v-else>{{ row.name }}</span>
                    </template>
                </el-table-column>
                <el-table-column :label="$t('terminal.quickCommand')" min-width="120" show-overflow-tooltip>
                    <template #default="{ row }">
                        <el-input
                            v-if="row.lineStatus === 'create' || row.lineStatus === 'edit'"
                            v-model="row.command"
                        />
                        <span v-else>{{ row.command }}</span>
                    </template>
                </el-table-column>
                <el-table-column min-width="40">
                    <template #default="scope">
                        <el-button
                            v-if="scope.row.lineStatus === 'create' || scope.row.lineStatus === 'edit'"
                            link
                            type="primary"
                            @click="handleCmdSave(scope.row)"
                        >
                            {{ $t('commons.button.save') }}
                        </el-button>
                        <el-button
                            v-if="!scope.row.lineStatus || scope.row.lineStatus === 'saved'"
                            link
                            type="primary"
                            @click="scope.row.lineStatus = 'edit'"
                        >
                            {{ $t('commons.button.edit') }}
                        </el-button>
                        <el-button
                            v-if="scope.row.lineStatus === 'create' || scope.row.lineStatus === 'edit'"
                            link
                            type="primary"
                            @click="search()"
                        >
                            {{ $t('commons.button.cancel') }}
                        </el-button>
                        <el-button
                            v-if="scope.row.lineStatus !== 'create' && scope.row.lineStatus !== 'edit'"
                            link
                            type="primary"
                            @click="handleCmdDelete(scope.$index)"
                        >
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </template>
                </el-table-column>
            </el-table>
        </template>
        <template #footer>
            <span class="dialog-footer">
                <el-button @click="handleClose">{{ $t('commons.button.cancel') }}</el-button>
            </span>
        </template>
    </DrawerPro>
    <OpDialog ref="opRef" @search="search" />
</template>

<script setup lang="ts">
import { Command } from '@/api/interface/command';
import { deleteCommand, getCommandPage, addCommand, editCommand } from '@/api/modules/command';
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';

const drawerVisible = ref();
const loading = ref();
const data = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
    orderBy: 'name',
    order: 'ascending',
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
        type: 'redis',
        lineStatus: 'create',
    };
    data.value.push(item);
};
const handleCmdDelete = (index: number) => {
    batchDelete(data.value[index]);
};

const handleCmdSave = async (row: any) => {
    if (row.name === '' || row.command === '') {
        MsgError(i18n.global.t('toolbox.device.hostHelper'));
        return;
    }
    loading.value = true;
    if (row.lineStatus === 'create') {
        await addCommand(row)
            .then(() => {
                loading.value = false;
                row.lineStatus = 'saved';
                MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
                search();
            })
            .catch(() => {
                loading.value = false;
            });
        return;
    }
    await editCommand(row)
        .then(() => {
            loading.value = false;
            row.lineStatus = 'saved';
            MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
            search();
        })
        .catch(() => {
            loading.value = false;
        });
};

const batchDelete = async (row: Command.CommandInfo | null) => {
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
        api: deleteCommand,
        params: { ids: ids },
    });
};

const search = async (column?: any) => {
    paginationConfig.orderBy = column?.order ? column.prop : paginationConfig.orderBy;
    paginationConfig.order = column?.order ? column.order : paginationConfig.order;
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        orderBy: paginationConfig.orderBy,
        order: paginationConfig.order,
        info: '',
        type: 'redis',
    };
    loading.value = true;
    await getCommandPage(params)
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
