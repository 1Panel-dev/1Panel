<template>
    <div v-loading="loading">
        <LayoutContent>
            <template #title>
                <back-button name="MySQL" :header="$t('database.remoteDB')" />
            </template>
            <template #toolbar>
                <el-row>
                    <el-col :xs="24" :sm="20" :md="20" :lg="20" :xl="20">
                        <el-button type="primary" @click="onOpenDialog('create')">
                            {{ $t('database.createRemoteDB') }}
                        </el-button>
                    </el-col>
                    <el-col :xs="24" :sm="4" :md="4" :lg="4" :xl="4">
                        <div class="search-button">
                            <el-input
                                v-model="searchName"
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
            <template #main>
                <ComplexTable :pagination-config="paginationConfig" @sort-change="search" @search="search" :data="data">
                    <el-table-column show-overflow-tooltip :label="$t('commons.table.name')" prop="name" sortable />
                    <el-table-column show-overflow-tooltip :label="$t('database.address')" prop="address" />
                    <el-table-column :label="$t('commons.login.username')" prop="username" />
                    <el-table-column :label="$t('commons.login.password')" prop="password">
                        <template #default="{ row }">
                            <div>
                                <span style="float: left; line-height: 25px" v-if="!row.showPassword">***********</span>
                                <div style="cursor: pointer; float: left" v-if="!row.showPassword">
                                    <el-icon
                                        style="margin-left: 5px; margin-top: 3px"
                                        @click="row.showPassword = true"
                                        :size="16"
                                    >
                                        <View />
                                    </el-icon>
                                </div>
                                <span style="float: left" v-if="row.showPassword">{{ row.password }}</span>
                                <div style="cursor: pointer; float: left" v-if="row.showPassword">
                                    <el-icon
                                        style="margin-left: 5px; margin-top: 3px"
                                        @click="row.showPassword = false"
                                        :size="16"
                                    >
                                        <Hide />
                                    </el-icon>
                                </div>
                                <div style="cursor: pointer; float: left">
                                    <el-icon style="margin-left: 5px; margin-top: 3px" :size="16" @click="onCopy(row)">
                                        <DocumentCopy />
                                    </el-icon>
                                </div>
                            </div>
                        </template>
                    </el-table-column>
                    <el-table-column
                        prop="description"
                        :label="$t('commons.table.description')"
                        show-overflow-tooltip
                    />
                    <el-table-column
                        prop="createdAt"
                        :label="$t('commons.table.date')"
                        :formatter="dateFormat"
                        show-overflow-tooltip
                    />
                    <fu-table-operations
                        width="170px"
                        :buttons="buttons"
                        :ellipsis="10"
                        :label="$t('commons.table.operate')"
                        fix
                    />
                </ComplexTable>
            </template>
        </LayoutContent>

        <OperateDialog ref="dialogRef" @search="search" />
    </div>
</template>

<script lang="ts" setup>
import { dateFormat } from '@/utils/util';
import { onMounted, reactive, ref } from 'vue';
import { deleteDatabase, searchDatabases } from '@/api/modules/database';
import OperateDialog from '@/views/database/mysql/remote/operate/index.vue';
import i18n from '@/lang';
import { MsgError, MsgSuccess } from '@/utils/message';
import useClipboard from 'vue-clipboard3';
import { Database } from '@/api/interface/database';
import { useDeleteData } from '@/hooks/use-delete-data';
const { toClipboard } = useClipboard();

const loading = ref(false);

const dialogRef = ref();

const data = ref();
const paginationConfig = reactive({
    cacheSizeKey: 'mysql-remote-page-size',
    currentPage: 1,
    pageSize: 10,
    total: 0,
    orderBy: 'created_at',
    order: 'null',
});
const searchName = ref();

const search = async (column?: any) => {
    paginationConfig.orderBy = column?.order ? column.prop : paginationConfig.orderBy;
    paginationConfig.order = column?.order ? column.order : paginationConfig.order;
    let params = {
        page: paginationConfig.currentPage,
        pageSize: paginationConfig.pageSize,
        info: searchName.value,
        type: 'mysql,mariadb',
        orderBy: paginationConfig.orderBy,
        order: paginationConfig.order,
    };
    const res = await searchDatabases(params);
    data.value = res.data.items || [];
    paginationConfig.total = res.data.total;
};

const onOpenDialog = async (
    title: string,
    rowData: Partial<Database.DatabaseInfo> = {
        name: '',
        type: 'mysql',
        version: '8.x',
        address: '',
        port: 3306,
        username: 'root',
        password: '',
        description: '',
    },
) => {
    let params = {
        title,
        rowData: { ...rowData },
    };
    dialogRef.value!.acceptParams(params);
};

const onCopy = async (row: any) => {
    try {
        await toClipboard(row.password);
        MsgSuccess(i18n.global.t('commons.msg.copySuccess'));
    } catch (e) {
        MsgError(i18n.global.t('commons.msg.copyfailed'));
    }
};

const onDelete = async (row: Database.DatabaseInfo) => {
    await useDeleteData(deleteDatabase, row.id, 'commons.msg.delete');
    search();
};

const buttons = [
    {
        label: i18n.global.t('commons.button.edit'),
        click: (row: Database.DatabaseInfo) => {
            onOpenDialog('edit', row);
        },
    },
    {
        label: i18n.global.t('commons.button.delete'),
        click: (row: Database.DatabaseInfo) => {
            onDelete(row);
        },
    },
];

onMounted(() => {
    search();
});
</script>
