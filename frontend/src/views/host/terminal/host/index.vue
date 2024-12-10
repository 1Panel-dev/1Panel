<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('terminal.host', 2)">
            <template #toolbar>
                <div class="flex w-full flex-col gap-4 md:justify-between md:flex-row">
                    <div class="flex flex-wrap gap-4">
                        <el-button type="primary" @click="onOpenDialog('create')">
                            {{ $t('terminal.addHost') }}
                        </el-button>
                        <el-button type="primary" plain @click="onOpenGroupDialog()">
                            {{ $t('terminal.manageGroup') }}
                        </el-button>
                        <el-button type="primary" plain :disabled="selects.length === 0" @click="onBatchDelete(null)">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </div>
                    <div class="flex flex-wrap gap-3">
                        <TableSearch @search="search()" v-model:searchName="info" />
                    </div>
                </div>
            </template>
            <template #search>
                <el-select v-model="group" @change="search()" clearable class="p-w-200">
                    <template #prefix>{{ $t('terminal.group') }}</template>
                    <el-option :label="$t('commons.table.all')" value=""></el-option>
                    <div v-for="item in groupList" :key="item.name">
                        <el-option :value="item.id" :label="item.name" />
                    </div>
                </el-select>
            </template>
            <template #main>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    :data="data"
                    @search="search"
                >
                    <el-table-column type="selection" :selectable="selectable" fix />
                    <el-table-column :label="$t('terminal.ip')" prop="addr" fix />
                    <el-table-column :label="$t('commons.login.username')" show-overflow-tooltip prop="user" />
                    <el-table-column :label="$t('commons.table.port')" prop="port" />
                    <el-table-column :label="$t('commons.table.group')" show-overflow-tooltip prop="groupBelong">
                        <template #default="{ row }">
                            <span v-if="row.groupBelong === 'default'">{{ $t('website.default') }}</span>
                            <span v-else>{{ row.groupBelong }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column :label="$t('commons.table.title')" show-overflow-tooltip prop="name" />
                    <el-table-column
                        :label="$t('commons.table.description')"
                        show-overflow-tooltip
                        prop="description"
                    />
                    <fu-table-operations width="200px" :buttons="buttons" :label="$t('commons.table.operate')" fix />
                </ComplexTable>
            </template>
        </LayoutContent>

        <OpDialog ref="opRef" @search="search" />
        <OperateDialog @search="search" ref="dialogRef" />
        <GroupDialog @search="search" ref="dialogGroupRef" />
        <GroupChangeDialog @search="search" @change="onChangeGroup" ref="dialogGroupChangeRef" />
    </div>
</template>

<script setup lang="ts">
import GroupDialog from '@/components/group/index.vue';
import GroupChangeDialog from '@/components/group/change.vue';
import OperateDialog from '@/views/host/terminal/host/operate/index.vue';
import { deleteHost, editHostGroup, searchHosts } from '@/api/modules/host';
import { GetGroupList } from '@/api/modules/group';
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { Host } from '@/api/interface/host';
import { MsgSuccess } from '@/utils/message';

const loading = ref();
const data = ref();
const groupList = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    cacheSizeKey: 'terminal-host-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const info = ref();
const group = ref<string>('');
const dialogGroupChangeRef = ref();
const currentID = ref();

const opRef = ref();

const acceptParams = () => {
    search();
};

function selectable(row) {
    return row.addr !== '127.0.0.1';
}
const dialogRef = ref();
const onOpenDialog = async (
    title: string,
    rowData: Partial<Host.Host> = {
        port: 22,
        user: 'root',
        authMode: 'password',
    },
) => {
    let params = {
        title,
        rowData: { ...rowData },
    };
    dialogRef.value!.acceptParams(params);
};

const dialogGroupRef = ref();
const onOpenGroupDialog = () => {
    dialogGroupRef.value!.acceptParams({ type: 'host' });
};

const onBatchDelete = async (row: Host.Host | null) => {
    let names = [];
    let ids = [];
    if (row) {
        names = [row.name + '[' + row.addr + ']'];
        ids = [row.id];
    } else {
        selects.value.forEach((item: Host.Host) => {
            names.push(item.name + '[' + item.addr + ']');
            ids.push(item.id);
        });
    }
    opRef.value.acceptParams({
        title: i18n.global.t('commons.button.delete'),
        names: names,
        msg: i18n.global.t('commons.msg.operatorHelper', [
            i18n.global.t('terminal.host'),
            i18n.global.t('commons.button.delete'),
        ]),
        api: deleteHost,
        params: { ids: ids },
    });
};

const loadGroups = async () => {
    const res = await GetGroupList({ type: 'host' });
    groupList.value = res.data;
};

const onChangeGroup = async (groupID: number) => {
    let param = {
        id: currentID.value,
        groupID: groupID,
    };
    await editHostGroup(param);
    search();
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
};

const buttons = [
    {
        label: i18n.global.t('terminal.groupChange'),
        click: (row: any) => {
            currentID.value = row.id;
            dialogGroupChangeRef.value!.acceptParams({ group: row.groupBelong, groupType: 'host' });
        },
    },
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: any) => {
            onOpenDialog('edit', row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Host.Host) => {
            onBatchDelete(row);
        },
        disabled: (row: any) => {
            return row.addr === '127.0.0.1';
        },
    },
];

const search = async () => {
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        groupID: Number(group.value),
        info: info.value,
    };
    loadGroups();
    loading.value = true;
    await searchHosts(params)
        .then((res) => {
            loading.value = false;
            data.value = res.data.items || [];
            paginationConfig.total = res.data.total;
        })
        .catch(() => {
            loading.value = false;
        });
};

defineExpose({
    acceptParams,
});
</script>
