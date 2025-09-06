<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('terminal.host', 2)">
            <template #leftToolBar>
                <el-button type="primary" @click="onOpenDialog('create')">
                    {{ $t('terminal.addHost') }}
                </el-button>
                <el-button type="primary" plain @click="onOpenGroupDialog()">
                    {{ $t('commons.table.group') }}
                </el-button>
                <el-button type="primary" plain :disabled="selects.length === 0" @click="onBatchDelete(null)">
                    {{ $t('commons.button.delete') }}
                </el-button>
            </template>
            <template #rightToolBar>
                <el-select v-model="group" @change="search()" clearable class="p-w-200 mr-5">
                    <template #prefix>{{ $t('commons.table.group') }}</template>
                    <el-option :label="$t('commons.table.all')" value=""></el-option>
                    <div v-for="item in groupList" :key="item.id">
                        <el-option
                            v-if="item.name === 'Default'"
                            :label="$t('commons.table.default')"
                            :value="item.id"
                        />
                        <el-option v-else :label="item.name" :value="item.id" />
                    </div>
                </el-select>
                <TableSearch @search="search()" v-model:searchName="info" />
            </template>
            <template #main>
                <ComplexTable
                    :pagination-config="paginationConfig"
                    v-model:selects="selects"
                    :data="data"
                    @search="search"
                >
                    <el-table-column type="selection" fix />
                    <el-table-column :label="$t('terminal.ip')" prop="addr" fix />
                    <el-table-column :label="$t('commons.login.username')" show-overflow-tooltip prop="user" />
                    <el-table-column :label="$t('commons.table.port')" prop="port" />
                    <el-table-column
                        :label="$t('commons.table.group')"
                        prop="description"
                        min-width="80"
                        show-overflow-tooltip
                    >
                        <template #default="{ row }">
                            <fu-select-rw-switch v-model="row.groupID" @change="updateGroup(row)">
                                <template #read>
                                    {{ row.groupBelong === 'Default' ? $t('commons.table.default') : row.groupBelong }}
                                </template>
                                <div v-for="item in groupList" :key="item.id">
                                    <el-option
                                        v-if="item.name === 'Default'"
                                        :label="$t('commons.table.default')"
                                        :value="item.id"
                                    />
                                    <el-option v-else :label="item.name" :value="item.id" />
                                </div>
                            </fu-select-rw-switch>
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
    </div>
</template>

<script setup lang="ts">
import GroupDialog from '@/components/group/index.vue';
import OperateDialog from '@/views/terminal/host/operate/index.vue';
import { deleteHost, editHostGroup, searchHosts } from '@/api/modules/terminal';
import { getGroupList } from '@/api/modules/group';
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
    pageSize: Number(localStorage.getItem('terminal-host-page-size')) || 10,
    total: 0,
});
const info = ref();
const group = ref<string>('');

const opRef = ref();

const acceptParams = () => {
    search();
};

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
    const res = await getGroupList('host');
    groupList.value = res.data;
};

const updateGroup = async (row: any) => {
    await editHostGroup({ id: row.id, groupID: row.groupID });
    search();
    MsgSuccess(i18n.global.t('commons.msg.operationSuccess'));
};

const buttons = [
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
