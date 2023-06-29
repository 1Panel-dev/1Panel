<template>
    <div>
        <LayoutContent v-loading="loading" :title="$t('terminal.host')">
            <template #toolbar>
                <el-row>
                    <el-col :span="20">
                        <el-button type="primary" @click="onOpenDialog('create')">
                            {{ $t('terminal.addHost') }}
                        </el-button>
                        <el-button type="primary" plain @click="onOpenGroupDialog()">
                            {{ $t('terminal.group') }}
                        </el-button>
                        <el-button type="primary" plain :disabled="selects.length === 0" @click="onBatchDelete(null)">
                            {{ $t('commons.button.delete') }}
                        </el-button>
                    </el-col>
                    <el-col :span="4">
                        <div class="search-button">
                            <el-input
                                v-model="info"
                                clearable
                                @clear="search()"
                                suffix-icon="Search"
                                @keyup.enter="search()"
                                @change="search()"
                                :placeholder="$t('commons.button.search')"
                            ></el-input>
                        </div>
                    </el-col>
                </el-row>
            </template>
            <template #search>
                <el-select v-model="group" @change="search()" clearable>
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
                    <el-table-column :label="$t('commons.table.title')" show-overflow-tooltip prop="name">
                        <template #default="{ row }">
                            <span v-if="row.addr === '127.0.0.1'">{{ $t('terminal.localhost') }}</span>
                            <span v-else>{{ row.name }}</span>
                        </template>
                    </el-table-column>
                    <el-table-column
                        :label="$t('commons.table.description')"
                        show-overflow-tooltip
                        prop="description"
                    />
                    <fu-table-operations width="200px" :buttons="buttons" :label="$t('commons.table.operate')" fix />
                </ComplexTable>
            </template>
        </LayoutContent>

        <OperateDialog @search="search" ref="dialogRef" />
        <GroupDialog @search="search" ref="dialogGroupRef" />
        <GroupChangeDialog @search="search" ref="dialogGroupChangeRef" />
    </div>
</template>

<script setup lang="ts">
import GroupDialog from '@/components/group/index.vue';
import GroupChangeDialog from '@/views/host/terminal/host/change-group/index.vue';
import OperateDialog from '@/views/host/terminal/host/operate/index.vue';
import { deleteHost, searchHosts } from '@/api/modules/host';
import { GetGroupList } from '@/api/modules/group';
import { reactive, ref } from 'vue';
import i18n from '@/lang';
import { Host } from '@/api/interface/host';
import { useDeleteData } from '@/hooks/use-delete-data';

const loading = ref();
const data = ref();
const groupList = ref();
const selects = ref<any>([]);
const paginationConfig = reactive({
    currentPage: 1,
    pageSize: 10,
    total: 0,
});
const info = ref();
const group = ref<string>('');
const dialogGroupChangeRef = ref();

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
    let ids: Array<number> = [];
    if (row) {
        ids.push(row.id);
    } else {
        selects.value.forEach((item: Host.Host) => {
            ids.push(item.id);
        });
    }
    await useDeleteData(deleteHost, { ids: ids }, 'commons.msg.delete');
    search();
};

const loadGroups = async () => {
    const res = await GetGroupList({ type: 'host' });
    groupList.value = res.data;
};

const buttons = [
    {
        label: i18n.global.t('terminal.groupChange'),
        click: (row: any) => {
            dialogGroupChangeRef.value!.acceptParams({ id: row.id, group: row.groupBelong });
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
